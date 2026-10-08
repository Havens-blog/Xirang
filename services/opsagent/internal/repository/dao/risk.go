package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// RiskEntryDAO — risk_entries 集合（CAS 乐观锁状态机）
// ---------------------------------------------------------------------------

// RiskEntriesCollection risk_entries 集合名（schema.mongo.js）。
const RiskEntriesCollection = "risk_entries"

// riskEntryDoc risk_entries 集合存储文档。与 domain.RiskEntry 的差异：存储侧
// 额外携带 updated_at（er-diagram 为 NOT NULL 存储字段，API 形状不暴露）。
// status_history 内嵌 StatusChange（from/to/by/at）。
type riskEntryDoc struct {
	ID            string                 `bson:"_id"`
	TenantID      string                 `bson:"tenant_id"`
	Fingerprint   string                 `bson:"fingerprint"`
	DiagnosisID   string                 `bson:"diagnosis_id"`
	ServiceName   string                 `bson:"service_name"`
	Severity      domain.Severity        `bson:"severity"`
	Status        domain.RiskEntryStatus `bson:"status"`
	Version       int                    `bson:"version"`
	StatusHistory []domain.StatusChange  `bson:"status_history"`
	Notified      domain.NotifyResult    `bson:"notified"`
	CreatedAt     time.Time              `bson:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at"`
}

// riskEntryDocFrom 领域条目 → 存储文档：补齐 ID/status/version/时间戳缺省值
// （ID 空生成 ObjectID.Hex；status 空 → pending_view；version 0 → 1）。
func riskEntryDocFrom(e *domain.RiskEntry) riskEntryDoc {
	id := e.ID
	if id == "" {
		id = primitive.NewObjectID().Hex()
	}
	status := e.Status
	if status == "" {
		status = domain.RiskStatusPendingView
	}
	version := e.Version
	if version == 0 {
		version = 1
	}
	now := time.Now()
	createdAt := e.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	// status_history 初始化为空数组（nil 切片会编码为 null，$push 时报
	// "must be an array but is of type null"）。
	history := e.StatusHistory
	if history == nil {
		history = []domain.StatusChange{}
	}
	return riskEntryDoc{
		ID:            id,
		TenantID:      e.Tenant,
		Fingerprint:   e.Fingerprint,
		DiagnosisID:   e.DiagnosisID,
		ServiceName:   e.ServiceName,
		Severity:      e.Severity,
		Status:        status,
		Version:       version,
		StatusHistory: history,
		Notified:      e.Notified,
		CreatedAt:     createdAt,
		UpdatedAt:     now,
	}
}

// toDomain 存储文档 → 领域条目（updated_at 为存储侧字段，不进领域形状）。
func (doc riskEntryDoc) toDomain() domain.RiskEntry {
	return domain.RiskEntry{
		ID:            doc.ID,
		Tenant:        doc.TenantID,
		Fingerprint:   doc.Fingerprint,
		DiagnosisID:   doc.DiagnosisID,
		ServiceName:   doc.ServiceName,
		Severity:      doc.Severity,
		Status:        doc.Status,
		Version:       doc.Version,
		StatusHistory: doc.StatusHistory,
		Notified:      doc.Notified,
		CreatedAt:     doc.CreatedAt,
	}
}

// ---------------------------------------------------------------------------
// 列表筛选 / 批量结果
// ---------------------------------------------------------------------------

// RiskEntryFilter 风险中心列表筛选（api-handbook §3 列表语义）。零值字段不参与
// 过滤；ServiceName 为子串模糊匹配。
type RiskEntryFilter struct {
	Status      domain.RiskEntryStatus
	Severity    domain.Severity
	ServiceName string
}

// BatchStatusItem 批量标记请求条目（api-handbook §5：逐条 CAS）。
type BatchStatusItem struct {
	ID              string
	ExpectedVersion int
}

// BatchStatusSuccess 批量标记成功项（返回自增后的最新 version）。
type BatchStatusSuccess struct {
	ID      string
	Version int
}

// BatchStatusFailure 批量标记失败项（ReasonCode：conflict|not_found）。
type BatchStatusFailure struct {
	ID             string
	ReasonCode     string
	CurrentVersion int
}

// BatchStatusResult 批量标记聚合结果（成功/失败分别聚合，冲突不覆盖并发单标）。
type BatchStatusResult struct {
	Succeeded []BatchStatusSuccess
	Failed    []BatchStatusFailure
}

// ---------------------------------------------------------------------------
// 错误
// ---------------------------------------------------------------------------

// ConflictError 版本冲突（CAS version 不匹配）——携带库内最新条目，供 service
// 层映射 409 ERR_CONFLICT 并回显「已被其他值班更新为 X」。
type ConflictError struct {
	Current *domain.RiskEntry
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("risk entry %q version conflict (latest status=%s version=%d)",
		e.Current.ID, e.Current.Status, e.Current.Version)
}

// errInvalidRiskTransition 非法状态流转（done 回退 / 目标 pending_view / 非法目标）。
var errInvalidRiskTransition = errors.New("invalid risk entry status transition")

// canTransition 状态机合法性：pending_view → viewed/done；viewed → done；
// done 为终态不可回退。同状态幂等由调用方在 assembleCAS 前短路。
func canTransition(from, to domain.RiskEntryStatus) bool {
	switch from {
	case domain.RiskStatusPendingView:
		return to == domain.RiskStatusViewed || to == domain.RiskStatusDone
	case domain.RiskStatusViewed:
		return to == domain.RiskStatusDone
	default: // done 或非法 from
		return false
	}
}

// ---------------------------------------------------------------------------
// RiskEntryDAO
// ---------------------------------------------------------------------------

// RiskEntryDAO 风险中心条目 DAO（risk_entries 集合；CAS 乐观锁状态机）。
type RiskEntryDAO interface {
	// Insert 落库条目；回写 ID/status/version/created_at 缺省值（version 初始 1）。
	Insert(ctx context.Context, e *domain.RiskEntry) error
	// GetByID 按 (_id, tenant_id) 取回；跨租户/不存在 → mongo.ErrNoDocuments。
	GetByID(ctx context.Context, tenantID, id string) (*domain.RiskEntry, error)
	// UpdateStatus 状态流转（CAS findAndModify）：过滤 {_id, tenant_id, version}，
	// $inc version 并 $push status_history。expectedVersion 为调用方 CAS 令牌
	// （0 = 不做预校验）；version 不匹配返回 *ConflictError（携带最新条目）。
	UpdateStatus(ctx context.Context, tenantID, id string, status domain.RiskEntryStatus, expectedVersion int, by string) (*domain.RiskEntry, error)
	// MarkViewed 惰性自动已读（诊断详情打开触发）：仅当过滤
	// {_id, tenant_id, diagnosis_id, status:pending_view} 命中时才流转为 viewed，
	// 幂等；不关联/已流转时不流转（changed=false），不存在/跨租户 ErrNoDocuments。
	MarkViewed(ctx context.Context, tenantID, id, diagnosisID, by string) (*domain.RiskEntry, bool, error)
	// BatchUpdateStatus 批量逐条独立 CAS（api-handbook §5）：每条过滤 version，
	// 成功/失败分别聚合；冲突与不存在分别记账 ReasonCode=conflict/not_found，
	// 冲突条目不覆盖并发单标。
	BatchUpdateStatus(ctx context.Context, tenantID string, status domain.RiskEntryStatus, items []BatchStatusItem) (*BatchStatusResult, error)
	// List 风险中心列表：created_at 倒序，status/severity/serviceName(模糊) 筛选
	// + 分页；返回命中的 total（当前筛选条件总条数）。
	List(ctx context.Context, tenantID string, filter RiskEntryFilter, limit, offset int) ([]domain.RiskEntry, int64, error)
	// InitIndexes 初始化 schema.mongo.js 定义的 risk_entries 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type riskEntryDAO struct {
	db *mongox.Mongo
}

// NewRiskEntryDAO 构造 RiskEntryDAO。
func NewRiskEntryDAO(db *mongox.Mongo) RiskEntryDAO {
	return &riskEntryDAO{db: db}
}

// InitIndexes 创建 idx_re_tenant_status / idx_re_tenant_service /
// idx_re_tenant_fingerprint / idx_re_dx（与 schema.mongo.js 一致）。
func (d *riskEntryDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_re_tenant_status"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "service_name", Value: 1}},
			Options: options.Index().SetName("idx_re_tenant_service"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "fingerprint", Value: 1}},
			Options: options.Index().SetName("idx_re_tenant_fingerprint"),
		},
		{
			Keys:    bson.D{{Key: "diagnosis_id", Value: 1}},
			Options: options.Index().SetName("idx_re_dx"),
		},
	}
	_, err := d.db.Collection(RiskEntriesCollection).Indexes().CreateMany(ctx, models)
	return err
}

// Insert 落库条目并回写缺省值。
func (d *riskEntryDAO) Insert(ctx context.Context, e *domain.RiskEntry) error {
	doc := riskEntryDocFrom(e)
	if _, err := d.db.Collection(RiskEntriesCollection).InsertOne(ctx, doc); err != nil {
		return err
	}
	e.ID = doc.ID
	e.Status = doc.Status
	e.Version = doc.Version
	e.CreatedAt = doc.CreatedAt
	return nil
}

// getByOID 按 (已校验的 _id, tenant_id) 取回条目。
func (d *riskEntryDAO) getByOID(ctx context.Context, oid, tenantID string) (*domain.RiskEntry, error) {
	var doc riskEntryDoc
	err := d.db.Collection(RiskEntriesCollection).FindOne(ctx, bson.M{"_id": oid, "tenant_id": tenantID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	e := doc.toDomain()
	return &e, nil
}

// GetByID 按 (_id, tenant_id) 取回条目。
func (d *riskEntryDAO) GetByID(ctx context.Context, tenantID, id string) (*domain.RiskEntry, error) {
	oid, err := objectIDOrErr(id)
	if err != nil {
		return nil, err
	}
	return d.getByOID(ctx, oid, tenantID)
}

// UpdateStatus 状态流转：预读校验（状态机 + 幂等 + 版本）后走 findAndModify CAS。
func (d *riskEntryDAO) UpdateStatus(ctx context.Context, tenantID, id string, status domain.RiskEntryStatus, expectedVersion int, by string) (*domain.RiskEntry, error) {
	oid, err := objectIDOrErr(id)
	if err != nil {
		return nil, err
	}
	current, err := d.getByOID(ctx, oid, tenantID)
	if err != nil {
		return nil, err
	}
	// 同状态幂等 no-op：不推进 version、不追加 history。
	if current.Status == status {
		return current, nil
	}
	if !canTransition(current.Status, status) {
		return nil, errInvalidRiskTransition
	}
	if expectedVersion > 0 && current.Version != expectedVersion {
		return current, &ConflictError{Current: current}
	}
	return d.casUpdateStatus(ctx, oid, tenantID, current, status, by)
}

// casUpdateStatus findAndModify 原子 CAS：过滤 {_id, tenant_id, version}（version
// 为本次读到的乐观锁令牌），成功后 $inc version + $push status_history。未命中
// （含并发推进 version）→ 重读最新条目并以 ConflictError 返回，绝不覆盖。
func (d *riskEntryDAO) casUpdateStatus(ctx context.Context, oid, tenantID string, current *domain.RiskEntry, to domain.RiskEntryStatus, by string) (*domain.RiskEntry, error) {
	now := time.Now()
	filter := bson.M{"_id": oid, "tenant_id": tenantID, "version": current.Version}
	update := bson.M{
		"$set": bson.M{"status": string(to), "updated_at": now},
		"$inc": bson.M{"version": 1},
		"$push": bson.M{"status_history": domain.StatusChange{
			From: string(current.Status),
			To:   string(to),
			By:   by,
			At:   now,
		}},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc riskEntryDoc
	err := d.db.Collection(RiskEntriesCollection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		// 过滤含 version：未命中只可能是版本被并发推进（被抢先更新），
		// 或条目在此期间被删除；两者都不得覆盖，映射为冲突。
		if errors.Is(err, mongo.ErrNoDocuments) {
			latest, gerr := d.getByOID(ctx, oid, tenantID)
			if gerr != nil {
				return nil, gerr
			}
			return latest, &ConflictError{Current: latest}
		}
		return nil, err
	}
	e := doc.toDomain()
	return &e, nil
}

// MarkViewed 惰性自动已读（诊断详情打开触发）。仅当条目确实处于 pending_view 且
// 关联本诊断时才流转；不满足条件时返回 (当前条目, false, nil)；不存在/跨租户
// 返回 ErrNoDocuments（上层按无参数处理，不泄露资源存在性）。
func (d *riskEntryDAO) MarkViewed(ctx context.Context, tenantID, id, diagnosisID, by string) (*domain.RiskEntry, bool, error) {
	oid, err := objectIDOrErr(id)
	if err != nil {
		return nil, false, err
	}
	now := time.Now()
	filter := bson.M{
		"_id":          oid,
		"tenant_id":    tenantID,
		"diagnosis_id": diagnosisID,
		"status":       string(domain.RiskStatusPendingView),
	}
	update := bson.M{
		"$set": bson.M{"status": string(domain.RiskStatusViewed), "updated_at": now},
		"$inc": bson.M{"version": 1},
		"$push": bson.M{"status_history": domain.StatusChange{
			From: string(domain.RiskStatusPendingView),
			To:   string(domain.RiskStatusViewed),
			By:   by,
			At:   now,
		}},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc riskEntryDoc
	err = d.db.Collection(RiskEntriesCollection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err == nil {
		e := doc.toDomain()
		return &e, true, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, false, err
	}
	// 过滤未命中：区分「条目不存在/跨租户」（ErrNoDocuments）与
	// 「存在但不满足条件」（诊断不匹配 / 已 viewed/done → 不流转 changed=false）。
	cur, gerr := d.GetByID(ctx, tenantID, id)
	if gerr != nil {
		return nil, false, gerr
	}
	return cur, false, nil
}

// BatchUpdateStatus 批量逐条独立 CAS，成功/失败分别聚合。冲突与不存在分别记账，
// 不覆盖并发单标的推进结果。
func (d *riskEntryDAO) BatchUpdateStatus(ctx context.Context, tenantID string, status domain.RiskEntryStatus, items []BatchStatusItem) (*BatchStatusResult, error) {
	res := &BatchStatusResult{
		Succeeded: []BatchStatusSuccess{},
		Failed:    []BatchStatusFailure{},
	}
	for _, it := range items {
		// by 传空串：批量 DAO 不承载操作者身份（由 web 层 task 4.5 经单条 UpdateStatus
		// 注入 eiam user；此处仅推进状态机与版本）。
		got, err := d.UpdateStatus(ctx, tenantID, it.ID, status, it.ExpectedVersion, "")
		var conflict *ConflictError
		switch {
		case err == nil:
			res.Succeeded = append(res.Succeeded, BatchStatusSuccess{ID: got.ID, Version: got.Version})
		case errors.As(err, &conflict):
			res.Failed = append(res.Failed, BatchStatusFailure{ID: it.ID, ReasonCode: "conflict", CurrentVersion: conflict.Current.Version})
		case errors.Is(err, mongo.ErrNoDocuments):
			res.Failed = append(res.Failed, BatchStatusFailure{ID: it.ID, ReasonCode: "not_found"})
		default:
			res.Failed = append(res.Failed, BatchStatusFailure{ID: it.ID, ReasonCode: "invalid"})
		}
	}
	return res, nil
}

// List created_at 倒序列表 + 分页 + total（当前筛选条件命中总条数）。
func (d *riskEntryDAO) List(ctx context.Context, tenantID string, filter RiskEntryFilter, limit, offset int) ([]domain.RiskEntry, int64, error) {
	f := bson.M{"tenant_id": tenantID}
	if filter.Status.IsValid() {
		f["status"] = string(filter.Status)
	}
	if filter.Severity.IsValid() {
		f["severity"] = string(filter.Severity)
	}
	if filter.ServiceName != "" {
		f["service_name"] = bson.M{"$regex": filter.ServiceName}
	}

	total, err := d.db.Collection(RiskEntriesCollection).CountDocuments(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(clampLimit(limit))).
		SetSkip(int64(max(offset, 0)))
	cur, err := d.db.Collection(RiskEntriesCollection).Find(ctx, f, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	docs := []riskEntryDoc{}
	if err := cur.All(ctx, &docs); err != nil {
		return nil, 0, err
	}
	out := make([]domain.RiskEntry, 0, len(docs))
	for _, doc := range docs {
		out = append(out, doc.toDomain())
	}
	return out, total, nil
}
