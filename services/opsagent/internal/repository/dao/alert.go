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

// AlertsCollection alerts 集合名（schema.mongo.js）。
const AlertsCollection = "alerts"

// alertDoc alerts 集合存储文档（bson 字段名与 er-diagram alerts 一致）。
// 一个文档代表一个 (tenant, fingerprint, window_start) 记账窗口：首条 new/count=1，
// 后续同指纹 $inc count 置 merged；溢出丢弃各写一条 overflow_dropped。
type alertDoc struct {
	ID              string         `bson:"_id"`
	TenantID        string         `bson:"tenant_id"`
	Fingerprint     string         `bson:"fingerprint"`
	RawPayload      map[string]any `bson:"raw_payload"`
	DedupConclusion string         `bson:"dedup_conclusion"`
	DiagnosisRef    string         `bson:"diagnosis_ref,omitempty"`
	Count           int64          `bson:"count"`
	WindowStart     time.Time      `bson:"window_start"`
	CreatedAt       time.Time      `bson:"created_at"`
}

// AlertDAO 告警记账 DAO（alerts 集合；按指纹更短窗口记账留档）。
//
// 说明：Redis SET NX 为窗口去重的**主决策**（worker 侧），本集合为记账留档
// （审计/SC 口径），故首条/归并判定不依赖跨实例强一致；P1 单实例部署下无竞态。
type AlertDAO interface {
	// RecordArrival 按 (tenant_id, fingerprint, window_start) 记账：窗口内首条落地
	// （new, count=1），后续同指纹 $inc count 置 merged。返回记账结论与该窗口累计 count。
	RecordArrival(ctx context.Context, tenantID, fingerprint string, payload map[string]any, windowStart time.Time) (domain.DedupConclusion, int64, error)
	// RecordOverflow 队列溢出丢弃的低级别告警留档（overflow_dropped，保留 raw_payload；
	// 各写一条，不参与指纹归并计数）。
	RecordOverflow(ctx context.Context, tenantID, fingerprint string, payload map[string]any, windowStart time.Time) error
	// LinkDiagnosis 回填指纹窗口的关联诊断（worker 排查完成后调用）。
	LinkDiagnosis(ctx context.Context, tenantID, fingerprint string, windowStart time.Time, diagnosisID string) error
	// InitIndexes 初始化 schema.mongo.js 定义的 alerts 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type alertDAO struct {
	db *mongox.Mongo
}

// NewAlertDAO 构造 AlertDAO。
func NewAlertDAO(db *mongox.Mongo) AlertDAO {
	return &alertDAO{db: db}
}

// InitIndexes 创建 idx_alerts_fingerprint_window / idx_alerts_tenant_time
// （与 schema.mongo.js 一致）。
func (d *alertDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "fingerprint", Value: 1},
				{Key: "window_start", Value: 1},
			},
			Options: options.Index().SetName("idx_alerts_fingerprint_window"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_alerts_tenant_time"),
		},
	}
	_, err := d.db.Collection(AlertsCollection).Indexes().CreateMany(ctx, models)
	return err
}

// RecordArrival 记账：先尝试归并（findAndModify），未命中再落地首条。
func (d *alertDAO) RecordArrival(ctx context.Context, tenantID, fingerprint string, payload map[string]any, windowStart time.Time) (domain.DedupConclusion, int64, error) {
	now := time.Now()
	filter := bson.M{"tenant_id": tenantID, "fingerprint": fingerprint, "window_start": windowStart}
	update := bson.M{
		"$inc": bson.M{"count": 1},
		"$set": bson.M{"dedup_conclusion": string(domain.DedupMerged), "created_at": now},
	}
	var doc alertDoc
	err := d.db.Collection(AlertsCollection).
		FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).
		Decode(&doc)
	if err == nil {
		return domain.DedupMerged, doc.Count, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return "", 0, err
	}

	// 窗口首条：落地 new,count=1。并发下极小概率双写两条「new」，按上述单实例前提忽略。
	doc = alertDoc{
		ID:              primitive.NewObjectID().Hex(),
		TenantID:        tenantID,
		Fingerprint:     fingerprint,
		RawPayload:      payload,
		DedupConclusion: string(domain.DedupNew),
		Count:           1,
		WindowStart:     windowStart,
		CreatedAt:       now,
	}
	if _, err := d.db.Collection(AlertsCollection).InsertOne(ctx, doc); err != nil {
		return "", 0, err
	}
	return domain.DedupNew, 1, nil
}

// RecordOverflow 溢出丢弃留档（overflow_dropped）。
func (d *alertDAO) RecordOverflow(ctx context.Context, tenantID, fingerprint string, payload map[string]any, windowStart time.Time) error {
	_, err := d.db.Collection(AlertsCollection).InsertOne(ctx, alertDoc{
		ID:              primitive.NewObjectID().Hex(),
		TenantID:        tenantID,
		Fingerprint:     fingerprint,
		RawPayload:      payload,
		DedupConclusion: string(domain.DedupOverflowDropped),
		Count:           1,
		WindowStart:     windowStart,
		CreatedAt:       time.Now(),
	})
	return err
}

// LinkDiagnosis 回填指纹窗口的关联诊断。
func (d *alertDAO) LinkDiagnosis(ctx context.Context, tenantID, fingerprint string, windowStart time.Time, diagnosisID string) error {
	res, err := d.db.Collection(AlertsCollection).UpdateOne(ctx,
		bson.M{"tenant_id": tenantID, "fingerprint": fingerprint, "window_start": windowStart},
		bson.M{"$set": bson.M{"diagnosis_ref": diagnosisID}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
