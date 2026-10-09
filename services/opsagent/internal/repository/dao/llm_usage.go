package dao

import (
	"context"
	"time"

	"github.com/Havens-blog/opsagent/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// LLMUsageCollection llm_usage 集合名（schema.mongo.js）。
const LLMUsageCollection = "llm_usage"

// llmUsageDoc llm_usage 集合存储文档（租户级滚动 1h 预算账本，每窗口一行）。
type llmUsageDoc struct {
	ID             string    `bson:"_id"`
	TenantID       string    `bson:"tenant_id"`
	WindowStart    time.Time `bson:"window_start"`
	Calls          int64     `bson:"calls"`
	BudgetExceeded bool      `bson:"budget_exceeded"`
}

// LLMUsage llm_usage 窗口账本读模型（DAOD 层；internal，非 API 形状）。
type LLMUsage struct {
	TenantID       string
	WindowStart    time.Time
	Calls          int64
	BudgetExceeded bool
}

// LLMUsageDAO LLM 滚动 1h 预算账本 DAO（llm_usage 集合）。
//
// 预算限额属配置（DegradationController 判定），本 DAO 只负责 per-window 计数与
// budget_exceeded 置位；不存在跨窗口聚合。
type LLMUsageDAO interface {
	// IncrCall 在 (tenant_id, window_start) 窗口 $inc calls（窗口不存在时 upsert），
	// 返回累计 calls 供预算判定。
	IncrCall(ctx context.Context, tenantID string, windowStart time.Time) (int64, error)
	// GetWindow 返回窗口账本；不存在 → mongo.ErrNoDocuments。
	GetWindow(ctx context.Context, tenantID string, windowStart time.Time) (*LLMUsage, error)
	// MarkBudgetExceeded 置该窗口 budget_exceeded=true（预算耗尽，触发三级降级谓词）。
	MarkBudgetExceeded(ctx context.Context, tenantID string, windowStart time.Time) error
	// InitIndexes 初始化 schema.mongo.js 定义的 llm_usage 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type llmUsageDAO struct {
	db *mongox.Mongo
}

// NewLLMUsageDAO 构造 LLMUsageDAO。
func NewLLMUsageDAO(db *mongox.Mongo) LLMUsageDAO {
	return &llmUsageDAO{db: db}
}

// InitIndexes 创建 idx_llm_tenant_window（与 schema.mongo.js 一致）。
func (d *llmUsageDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "window_start", Value: 1}},
			Options: options.Index().SetName("idx_llm_tenant_window"),
		},
	}
	_, err := d.db.Collection(LLMUsageCollection).Indexes().CreateMany(ctx, models)
	return err
}

// IncrCall 原子 upsert $inc calls，返回累计值。
func (d *llmUsageDAO) IncrCall(ctx context.Context, tenantID string, windowStart time.Time) (int64, error) {
	filter := bson.M{"tenant_id": tenantID, "window_start": windowStart}
	update := bson.M{"$inc": bson.M{"calls": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)
	var doc llmUsageDoc
	err := d.db.Collection(LLMUsageCollection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		return 0, err
	}
	return doc.Calls, nil
}

// GetWindow 读窗口账本。
func (d *llmUsageDAO) GetWindow(ctx context.Context, tenantID string, windowStart time.Time) (*LLMUsage, error) {
	var doc llmUsageDoc
	err := d.db.Collection(LLMUsageCollection).
		FindOne(ctx, bson.M{"tenant_id": tenantID, "window_start": windowStart}).
		Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &LLMUsage{
		TenantID:       doc.TenantID,
		WindowStart:    doc.WindowStart,
		Calls:          doc.Calls,
		BudgetExceeded: doc.BudgetExceeded,
	}, nil
}

// MarkBudgetExceeded 置预算耗尽位。
func (d *llmUsageDAO) MarkBudgetExceeded(ctx context.Context, tenantID string, windowStart time.Time) error {
	res, err := d.db.Collection(LLMUsageCollection).UpdateOne(ctx,
		bson.M{"tenant_id": tenantID, "window_start": windowStart},
		bson.M{"$set": bson.M{"budget_exceeded": true}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
