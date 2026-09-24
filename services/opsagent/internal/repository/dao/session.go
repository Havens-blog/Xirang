// Package dao sessions / diagnoses / citations 集合的 Mongo DAO。
//
// 硬约束：所有查询强制携带 tenant 谓词（SC-6 租户隔离）；跨租户/不存在一律
// 以 mongo.ErrNoDocuments 或空切片呈现，由 service 层映射为 404 ERR_NOT_FOUND
// （不泄露资源存在性）。索引名与 schema.mongo.js 逐条一致。
// 落库失败补偿路径由 internal/worker（任务 4.2）处理，本包只返回原始错误。
package dao

import (
	"context"
	"errors"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 集合名（schema.mongo.js）。
const (
	SessionsCollection  = "sessions"
	DiagnosesCollection = "diagnoses"
	CitationsCollection = "citations"
)

// 分页收敛：缺省 20、上限 100（防全量扫）。
const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// clampLimit 收敛分页 limit。
func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

// objectIDOrErr 校验十六进制 ID 形状；非法 ID 与「不存在」同语义
// （返回 mongo.ErrNoDocuments，避免向调用方区分形状错误与跨租户缺失）。
// 说明：domain bson tag 将 _id 冻结为 string（ObjectID.Hex），存储与查询均用
// 字符串形态，可无歧义回转为 ObjectId。
func objectIDOrErr(id string) (string, error) {
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return "", mongo.ErrNoDocuments
	}
	return id, nil
}

// ---------------------------------------------------------------------------
// SessionDAO
// ---------------------------------------------------------------------------

// SessionDAO 排障会话 DAO（sessions 集合；er-diagram sessions 字段与 domain.Session 一致）。
type SessionDAO interface {
	// Insert 创建会话；ID 为空时生成 ObjectID.Hex 并回写 s.ID。
	Insert(ctx context.Context, s *domain.Session) error
	// GetByID 按 (_id, tenant_id) 取回会话；跨租户/不存在 → mongo.ErrNoDocuments。
	GetByID(ctx context.Context, tenantID, id string) (domain.Session, error)
	// ListByTime 按 (tenant_id, created_at) 历史检索（idx_sessions_tenant_time），
	// created_at 倒序；时间窗为闭开区间 [start, end)。
	ListByTime(ctx context.Context, tenantID string, start, end time.Time, limit, offset int) ([]domain.Session, error)
	// ListByService 按 (tenant_id, service_name) 检索（idx_sessions_tenant_service），
	// created_at 倒序。
	ListByService(ctx context.Context, tenantID, serviceName string, limit, offset int) ([]domain.Session, error)
	// UpdateStatus 流转会话状态（running|done|failed），同部刷新 updated_at。
	UpdateStatus(ctx context.Context, tenantID, id string, status domain.SessionStatus) error
	// InitIndexes 初始化 schema.mongo.js 定义的 sessions 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type sessionDAO struct {
	db *mongox.Mongo
}

// NewSessionDAO 构造 SessionDAO。
func NewSessionDAO(db *mongox.Mongo) SessionDAO {
	return &sessionDAO{db: db}
}

// InitIndexes 创建 idx_sessions_tenant_time / idx_sessions_tenant_service /
// idx_sessions_tenant_type（与 schema.mongo.js 一致，全部 tenant_id 前缀）。
func (d *sessionDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_sessions_tenant_time"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "service_name", Value: 1}},
			Options: options.Index().SetName("idx_sessions_tenant_service"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "type", Value: 1}},
			Options: options.Index().SetName("idx_sessions_tenant_type"),
		},
	}
	_, err := d.db.Collection(SessionsCollection).Indexes().CreateMany(ctx, models)
	return err
}

// Insert 创建会话；缺省补齐 ID/status/时间戳（orchestration_trace 内嵌随文档写入）。
func (d *sessionDAO) Insert(ctx context.Context, s *domain.Session) error {
	if s.ID == "" {
		s.ID = primitive.NewObjectID().Hex()
	}
	if s.Status == "" {
		s.Status = domain.SessionStatusRunning
	}
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	_, err := d.db.Collection(SessionsCollection).InsertOne(ctx, s)
	return err
}

// GetByID 按 (_id, tenant_id) 取回会话。
func (d *sessionDAO) GetByID(ctx context.Context, tenantID, id string) (domain.Session, error) {
	var out domain.Session
	oid, err := objectIDOrErr(id)
	if err != nil {
		return out, err
	}
	filter := bson.M{"_id": oid, "tenant_id": tenantID}
	err = d.db.Collection(SessionsCollection).FindOne(ctx, filter).Decode(&out)
	return out, err
}

// listSessions 通用列表查询：tenant 谓词 + 附加过滤 + created_at 倒序分页。
func (d *sessionDAO) listSessions(ctx context.Context, tenantID string, extra bson.M, limit, offset int) ([]domain.Session, error) {
	filter := bson.M{"tenant_id": tenantID}
	for k, v := range extra {
		filter[k] = v
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(clampLimit(limit))).
		SetSkip(int64(max(offset, 0)))
	cur, err := d.db.Collection(SessionsCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []domain.Session{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListByTime 按 (tenant_id, created_at) 检索历史会话。
func (d *sessionDAO) ListByTime(ctx context.Context, tenantID string, start, end time.Time, limit, offset int) ([]domain.Session, error) {
	return d.listSessions(ctx, tenantID, bson.M{
		"created_at": bson.M{"$gte": start, "$lt": end},
	}, limit, offset)
}

// ListByService 按 (tenant_id, service_name) 检索历史会话。
func (d *sessionDAO) ListByService(ctx context.Context, tenantID, serviceName string, limit, offset int) ([]domain.Session, error) {
	return d.listSessions(ctx, tenantID, bson.M{"service_name": serviceName}, limit, offset)
}

// errInvalidSessionStatus 非法会话状态（域层枚举校验失败）。
var errInvalidSessionStatus = errors.New("invalid session status")

// UpdateStatus 流转会话状态；过滤 {_id, tenant_id} 强制租户谓词。
func (d *sessionDAO) UpdateStatus(ctx context.Context, tenantID, id string, status domain.SessionStatus) error {
	if !status.IsValid() {
		return errInvalidSessionStatus
	}
	oid, err := objectIDOrErr(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid, "tenant_id": tenantID}
	update := bson.M{"$set": bson.M{"status": string(status), "updated_at": time.Now()}}
	res, err := d.db.Collection(SessionsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
