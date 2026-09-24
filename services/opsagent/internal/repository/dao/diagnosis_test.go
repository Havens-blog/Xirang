package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

// newTestDiagnosis 构造一条合法诊断记录（AC-1）。
func newTestDiagnosis(tenant, sessionID, serviceName string, createdAt time.Time) *DiagnosisRecord {
	return &DiagnosisRecord{
		Diagnosis: domain.Diagnosis{
			SessionID:  sessionID,
			Tenant:     tenant,
			RootCause:  "上游 CDN 回源超时导致错误率上升",
			Confidence: 0.87,
			Severity:   domain.SeverityP1,
			RiskLevel:  domain.RiskRead,
			Conclusions: []domain.Conclusion{
				{Text: "回源超时集中在华东节点", Citation: []string{"log-abc-123"}},
			},
			Disposition: []domain.DispositionStep{
				{Step: "检查华东节点回源健康度", Risk: "read", Action: "查看回源 P95 指标"},
			},
			Degraded:  false,
			Truncated: false,
			Trace: []domain.TraceStep{
				{Step: 1, Agent: "coordinator", Action: "intent", DurationMs: 10, Summary: "diagnose"},
			},
		},
		ServiceName: serviceName,
		CreatedAt:   createdAt,
	}
}

// TestDiagnosisDAO_InsertAndGetByID AC-1：诊断落库（含 created_at）并按租户取回。
func TestDiagnosisDAO_InsertAndGetByID(t *testing.T) {
	db := newTestMongo(t)
	d := NewDiagnosisDAO(db)
	ctx := context.Background()

	in := newTestDiagnosis("tenant-a", "sess-1", "order-service", mustUTCMillis(time.Now()))
	if err := d.Insert(ctx, in); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if in.Diagnosis.ID == "" {
		t.Fatal("insert should generate diagnosis ID")
	}

	out, err := d.GetByID(ctx, "tenant-a", in.Diagnosis.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if out.ID != in.Diagnosis.ID || out.SessionID != "sess-1" {
		t.Fatalf("mismatched diagnosis: %+v", out)
	}
	if out.RootCause != in.Diagnosis.RootCause || out.Severity != domain.SeverityP1 {
		t.Fatalf("fields not persisted: %+v", out)
	}
	if len(out.Conclusions) != 1 || len(out.Disposition) != 1 || len(out.Trace) != 1 {
		t.Fatalf("embedded arrays not persisted: %+v", out)
	}
	if out.Confidence != 0.87 {
		t.Fatalf("confidence = %v, want 0.87", out.Confidence)
	}
}

// TestDiagnosisDAO_GetBySession AC-4：按 session_id 关联查询诊断（idx_dx_session）。
func TestDiagnosisDAO_GetBySession(t *testing.T) {
	db := newTestMongo(t)
	d := NewDiagnosisDAO(db)
	ctx := context.Background()

	now := mustUTCMillis(time.Now())
	first := newTestDiagnosis("tenant-a", "sess-1", "svc", now)
	if err := d.Insert(ctx, first); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := d.Insert(ctx, newTestDiagnosis("tenant-a", "sess-2", "svc", now)); err != nil {
		t.Fatalf("insert 2: %v", err)
	}

	out, err := d.GetBySession(ctx, "tenant-a", "sess-1")
	if err != nil {
		t.Fatalf("get by session: %v", err)
	}
	if out.SessionID != "sess-1" || out.ID != first.Diagnosis.ID {
		t.Fatalf("association broken: %+v", out)
	}

	if _, err := d.GetBySession(ctx, "tenant-a", "sess-missing"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("missing session err = %v, want ErrNoDocuments", err)
	}
}

// TestDiagnosisDAO_TenantIsolation AC-3（SC-6）：跨租户读取诊断一律 NotFound 语义。
func TestDiagnosisDAO_TenantIsolation(t *testing.T) {
	db := newTestMongo(t)
	d := NewDiagnosisDAO(db)
	ctx := context.Background()

	in := newTestDiagnosis("tenant-a", "sess-1", "svc", mustUTCMillis(time.Now()))
	if err := d.Insert(ctx, in); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if _, err := d.GetByID(ctx, "tenant-b", in.Diagnosis.ID); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant get err = %v, want ErrNoDocuments", err)
	}
	if _, err := d.GetBySession(ctx, "tenant-b", "sess-1"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant get-by-session err = %v, want ErrNoDocuments", err)
	}
}
