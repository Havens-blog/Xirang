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

// ---------------------------------------------------------------------------
// CitationDAO
// ---------------------------------------------------------------------------

// citationDoc citations 集合存储文档（bson 字段名与 er-diagram citations 一致）。
type citationDoc struct {
	ID          string    `bson:"_id"`
	DiagnosisID string    `bson:"diagnosis_id"`
	TenantID    string    `bson:"tenant_id"`
	SourceType  string    `bson:"source_type"`
	SourceKey   string    `bson:"source_key"`
	Snippet     string    `bson:"snippet"`
	FetchedAt   time.Time `bson:"fetched_at"`
}

// citationDocFrom 领域引用 → 存储文档（补齐归属键与 fetched_at）。
func citationDocFrom(tenantID, diagnosisID string, c domain.Citation, fetchedAt time.Time) citationDoc {
	return citationDoc{
		ID:          primitive.NewObjectID().Hex(),
		DiagnosisID: diagnosisID,
		TenantID:    tenantID,
		SourceType:  c.SourceType,
		SourceKey:   c.SourceKey,
		Snippet:     c.Snippet,
		FetchedAt:   fetchedAt,
	}
}

// toDomain 存储文档 → 领域引用（结论回指所需最小形状）。
func (doc citationDoc) toDomain() domain.Citation {
	return domain.Citation{
		SourceType: doc.SourceType,
		SourceKey:  doc.SourceKey,
		Snippet:    doc.Snippet,
	}
}

// CitationDAO 数据源引用 DAO（citations 集合；诊断结论回指日志/指标/资产/告警）。
type CitationDAO interface {
	// InsertBatch 批量落库一次诊断的引用集合（fetched_at 由 DAO 统一落当前时间）；
	// citations 为空时为无操作。
	InsertBatch(ctx context.Context, tenantID, diagnosisID string, citations []domain.Citation) error
	// ListByDiagnosis 按 (diagnosis_id, tenant_id) 关联查询（idx_cit_dx + 租户谓词）。
	ListByDiagnosis(ctx context.Context, tenantID, diagnosisID string) ([]domain.Citation, error)
	// InitIndexes 初始化 schema.mongo.js 定义的 citations 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type citationDAO struct {
	db *mongox.Mongo
}

// NewCitationDAO 构造 CitationDAO。
func NewCitationDAO(db *mongox.Mongo) CitationDAO {
	return &citationDAO{db: db}
}

// InitIndexes 创建 idx_cit_dx / idx_cit_tenant（与 schema.mongo.js 一致）。
func (d *citationDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "diagnosis_id", Value: 1}},
			Options: options.Index().SetName("idx_cit_dx"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}},
			Options: options.Index().SetName("idx_cit_tenant"),
		},
	}
	_, err := d.db.Collection(CitationsCollection).Indexes().CreateMany(ctx, models)
	return err
}

// InsertBatch 批量落库引用。
func (d *citationDAO) InsertBatch(ctx context.Context, tenantID, diagnosisID string, citations []domain.Citation) error {
	if len(citations) == 0 {
		return nil
	}
	fetchedAt := time.Now()
	docs := make([]any, 0, len(citations))
	for _, c := range citations {
		docs = append(docs, citationDocFrom(tenantID, diagnosisID, c, fetchedAt))
	}
	_, err := d.db.Collection(CitationsCollection).InsertMany(ctx, docs)
	return err
}

// ListByDiagnosis 按 (diagnosis_id, tenant_id) 关联查询引用。
func (d *citationDAO) ListByDiagnosis(ctx context.Context, tenantID, diagnosisID string) ([]domain.Citation, error) {
	filter := bson.M{"diagnosis_id": diagnosisID, "tenant_id": tenantID}
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cur, err := d.db.Collection(CitationsCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	docs := []citationDoc{}
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]domain.Citation, 0, len(docs))
	for _, doc := range docs {
		out = append(out, doc.toDomain())
	}
	return out, nil
}
