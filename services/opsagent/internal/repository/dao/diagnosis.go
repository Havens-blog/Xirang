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
// DiagnosisDAO
// ---------------------------------------------------------------------------

// DiagnosisRecord 诊断落库记录：领域 Diagnosis + 存储侧补充字段。
//
// service_name / created_at 属 er-diagram diagnoses 集合字段（供
// idx_dx_tenant_service / idx_dx_tenant_time 检索），但 API 形状的
// domain.Diagnosis 不携带，由编排层在落库时经本结构补充；
// 引用（citations）为独立集合（见 CitationDAO），不在诊断文档内嵌。
type DiagnosisRecord struct {
	Diagnosis   domain.Diagnosis
	ServiceName string
	CreatedAt   time.Time
}

// diagnosisDoc diagnoses 集合存储文档（bson 字段名与 er-diagram diagnoses 一致；
// trace 为 API 诊断详情所需的编排调用链快照，随文档落库）。
type diagnosisDoc struct {
	ID          string                   `bson:"_id"`
	SessionID   string                   `bson:"session_id"`
	TenantID    string                   `bson:"tenant_id"`
	ServiceName string                   `bson:"service_name"`
	RootCause   string                   `bson:"root_cause"`
	Confidence  float64                  `bson:"confidence"`
	Severity    domain.Severity          `bson:"severity"`
	RiskLevel   domain.RiskLevel         `bson:"risk_level"`
	Conclusions []domain.Conclusion      `bson:"conclusions"`
	Disposition []domain.DispositionStep `bson:"disposition"`
	Trace       []domain.TraceStep       `bson:"trace"`
	Degraded    bool                     `bson:"degraded"`
	Truncated   bool                     `bson:"truncated"`
	CreatedAt   time.Time                `bson:"created_at"`
}

// diagnosisDocFromRecord 落库记录 → 存储文档（ID 为空时生成 ObjectID.Hex）。
func diagnosisDocFromRecord(rec *DiagnosisRecord) diagnosisDoc {
	id := rec.Diagnosis.ID
	if id == "" {
		id = primitive.NewObjectID().Hex()
	}
	createdAt := rec.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return diagnosisDoc{
		ID:          id,
		SessionID:   rec.Diagnosis.SessionID,
		TenantID:    rec.Diagnosis.Tenant,
		ServiceName: rec.ServiceName,
		RootCause:   rec.Diagnosis.RootCause,
		Confidence:  rec.Diagnosis.Confidence,
		Severity:    rec.Diagnosis.Severity,
		RiskLevel:   rec.Diagnosis.RiskLevel,
		Conclusions: rec.Diagnosis.Conclusions,
		Disposition: rec.Diagnosis.Disposition,
		Trace:       rec.Diagnosis.Trace,
		Degraded:    rec.Diagnosis.Degraded,
		Truncated:   rec.Diagnosis.Truncated,
		CreatedAt:   createdAt,
	}
}

// toDomain 存储文档 → 领域 Diagnosis（citations 由调用方经 CitationDAO 关联组装）。
func (doc diagnosisDoc) toDomain() domain.Diagnosis {
	return domain.Diagnosis{
		ID:          doc.ID,
		SessionID:   doc.SessionID,
		Tenant:      doc.TenantID,
		RootCause:   doc.RootCause,
		Confidence:  doc.Confidence,
		Severity:    doc.Severity,
		RiskLevel:   doc.RiskLevel,
		Conclusions: doc.Conclusions,
		Disposition: doc.Disposition,
		Degraded:    doc.Degraded,
		Truncated:   doc.Truncated,
		Trace:       doc.Trace,
	}
}

// DiagnosisDAO 诊断报告 DAO（diagnoses 集合）。
type DiagnosisDAO interface {
	// Insert 落库诊断；生成/回写 rec.Diagnosis.ID。
	Insert(ctx context.Context, rec *DiagnosisRecord) error
	// GetByID 按 (_id, tenant_id) 取回；跨租户/不存在 → mongo.ErrNoDocuments。
	GetByID(ctx context.Context, tenantID, id string) (*domain.Diagnosis, error)
	// GetBySession 按 (session_id, tenant_id) 关联查询（idx_dx_session + 租户谓词）。
	GetBySession(ctx context.Context, tenantID, sessionID string) (*domain.Diagnosis, error)
	// InitIndexes 初始化 schema.mongo.js 定义的 diagnoses 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type diagnosisDAO struct {
	db *mongox.Mongo
}

// NewDiagnosisDAO 构造 DiagnosisDAO。
func NewDiagnosisDAO(db *mongox.Mongo) DiagnosisDAO {
	return &diagnosisDAO{db: db}
}

// InitIndexes 创建 idx_dx_session / idx_dx_tenant_service / idx_dx_tenant_time
// （与 schema.mongo.js 一致）。
func (d *diagnosisDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetName("idx_dx_session"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "service_name", Value: 1}},
			Options: options.Index().SetName("idx_dx_tenant_service"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_dx_tenant_time"),
		},
	}
	_, err := d.db.Collection(DiagnosesCollection).Indexes().CreateMany(ctx, models)
	return err
}

// Insert 落库诊断。
func (d *diagnosisDAO) Insert(ctx context.Context, rec *DiagnosisRecord) error {
	doc := diagnosisDocFromRecord(rec)
	_, err := d.db.Collection(DiagnosesCollection).InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	rec.Diagnosis.ID = doc.ID
	return nil
}

// GetByID 按 (_id, tenant_id) 取回诊断。
func (d *diagnosisDAO) GetByID(ctx context.Context, tenantID, id string) (*domain.Diagnosis, error) {
	oid, err := objectIDOrErr(id)
	if err != nil {
		return nil, err
	}
	return d.findOneDiagnosis(ctx, bson.M{"_id": oid, "tenant_id": tenantID})
}

// GetBySession 按 (session_id, tenant_id) 关联查询。
func (d *diagnosisDAO) GetBySession(ctx context.Context, tenantID, sessionID string) (*domain.Diagnosis, error) {
	return d.findOneDiagnosis(ctx, bson.M{"session_id": sessionID, "tenant_id": tenantID})
}

// findOneDiagnosis 单文档查询 + 文档→领域映射。
func (d *diagnosisDAO) findOneDiagnosis(ctx context.Context, filter bson.M) (*domain.Diagnosis, error) {
	var doc diagnosisDoc
	err := d.db.Collection(DiagnosesCollection).FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, err
	}
	diag := doc.toDomain()
	return &diag, nil
}
