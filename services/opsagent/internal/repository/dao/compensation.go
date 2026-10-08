package dao

import (
	"context"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CompensationsCollection compensations 集合名（schema.mongo.js）。
const CompensationsCollection = "compensations"

// maxCompensationRetries 补偿重试上限（达上限标 dead + 运维告警；tech-design Error Handling）。
const maxCompensationRetries = 5

// compensationDoc compensations 集合存储文档（payload 自包含原写操作快照，
// 不外键引用业务集合——目标文档可能尚未落成）。
type compensationDoc struct {
	ID          string         `bson:"_id"`
	Kind        string         `bson:"kind"`
	Payload     map[string]any `bson:"payload"`
	RetryCount  int            `bson:"retry_count"`
	Status      string         `bson:"status"`
	NextRetryAt time.Time      `bson:"next_retry_at"`
	CreatedAt   time.Time      `bson:"created_at"`
}

// Compensation 补偿项读模型（DAO 层；唯一属主为 internal/worker 重试循环）。
type Compensation struct {
	ID          string
	Kind        domain.CompensationKind
	Payload     map[string]any
	RetryCount  int
	Status      domain.CompensationStatus
	NextRetryAt time.Time
	CreatedAt   time.Time
}

// CompensationDAO 落库补偿队列 DAO（compensations 集合；Mongo 持久化，Redis 仅触发信号）。
//
// 硬约束：本包只负责读写，不重放——重试循环唯一属主是 internal/worker（任务 4.2），
// 落库失败方只调用 Enqueue 写入。
type CompensationDAO interface {
	// Enqueue 写入补偿项（status=retry, retry_count=0，next_retry_at=now 即到期可扫）。
	Enqueue(ctx context.Context, kind domain.CompensationKind, payload map[string]any) error
	// ListDue 扫描到期待重试项（status=retry 且 next_retry_at ≤ now），next_retry_at 升序。
	ListDue(ctx context.Context, now time.Time, limit int) ([]Compensation, error)
	// Delete 重放成功后删除补偿项。
	Delete(ctx context.Context, id string) error
	// MarkRetried 重放失败：retry_count+1 并推进 next_retry_at；已达上限标 dead。
	MarkRetried(ctx context.Context, id string, nextRetryAt time.Time) error
	// InitIndexes 初始化 schema.mongo.js 定义的 compensations 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type compensationDAO struct {
	db *mongox.Mongo
}

// NewCompensationDAO 构造 CompensationDAO。
func NewCompensationDAO(db *mongox.Mongo) CompensationDAO {
	return &compensationDAO{db: db}
}

// InitIndexes 创建 idx_comp_retry（status + next_retry_at；系统级队列，无 tenant 前缀）。
func (d *compensationDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "next_retry_at", Value: 1}},
			Options: options.Index().SetName("idx_comp_retry"),
		},
	}
	_, err := d.db.Collection(CompensationsCollection).Indexes().CreateMany(ctx, models)
	return err
}

// Enqueue 写入补偿项。
func (d *compensationDAO) Enqueue(ctx context.Context, kind domain.CompensationKind, payload map[string]any) error {
	doc := compensationDoc{
		ID:          primitive.NewObjectID().Hex(),
		Kind:        string(kind),
		Payload:     payload,
		RetryCount:  0,
		Status:      string(domain.CompStatusRetry),
		NextRetryAt: time.Now(),
		CreatedAt:   time.Now(),
	}
	_, err := d.db.Collection(CompensationsCollection).InsertOne(ctx, doc)
	return err
}

// listDueDoc 转为读模型。
func (doc compensationDoc) toCompensation() Compensation {
	return Compensation{
		ID:          doc.ID,
		Kind:        domain.CompensationKind(doc.Kind),
		Payload:     doc.Payload,
		RetryCount:  doc.RetryCount,
		Status:      domain.CompensationStatus(doc.Status),
		NextRetryAt: doc.NextRetryAt,
		CreatedAt:   doc.CreatedAt,
	}
}

// ListDue 扫描到期待重试项。
func (d *compensationDAO) ListDue(ctx context.Context, now time.Time, limit int) ([]Compensation, error) {
	filter := bson.M{"status": string(domain.CompStatusRetry), "next_retry_at": bson.M{"$lte": now}}
	opts := options.Find().
		SetSort(bson.D{{Key: "next_retry_at", Value: 1}}).
		SetLimit(int64(clampLimit(limit)))
	cur, err := d.db.Collection(CompensationsCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	docs := []compensationDoc{}
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]Compensation, 0, len(docs))
	for _, doc := range docs {
		out = append(out, doc.toCompensation())
	}
	return out, nil
}

// Delete 重放成功后删除。
func (d *compensationDAO) Delete(ctx context.Context, id string) error {
	res, err := d.db.Collection(CompensationsCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// MarkRetried 重放失败：retry_count+1 并推进 next_retry_at；达上限标 dead。
func (d *compensationDAO) MarkRetried(ctx context.Context, id string, nextRetryAt time.Time) error {
	res, err := d.db.Collection(CompensationsCollection).UpdateOne(ctx,
		bson.M{"_id": id, "status": string(domain.CompStatusRetry)},
		bson.M{
			"$inc": bson.M{"retry_count": 1},
			"$set": bson.M{"next_retry_at": nextRetryAt},
		})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	// 达上限 → 标 dead。
	var doc compensationDoc
	if err := d.db.Collection(CompensationsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		return err
	}
	if doc.RetryCount >= maxCompensationRetries {
		_, err := d.db.Collection(CompensationsCollection).UpdateOne(ctx,
			bson.M{"_id": id},
			bson.M{"$set": bson.M{"status": string(domain.CompStatusDead)}})
		return err
	}
	return nil
}
