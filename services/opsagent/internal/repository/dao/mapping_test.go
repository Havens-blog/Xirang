package dao

import (
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// 本文件为纯映射单元测试（无需 Mongo，本地/CI 均执行）：
// 验证存储文档结构与领域类型之间的往返一致性（bson 字段名与 er-diagram 对齐）。

// TestDiagnosisDocRoundTrip DiagnosisRecord → diagnosisDoc → domain.Diagnosis 往返无损。
func TestDiagnosisDocRoundTrip(t *testing.T) {
	createdAt := mustUTCMillis(time.Now())
	rec := newTestDiagnosis("tenant-a", "sess-9", "order-service", createdAt)
	rec.Diagnosis.ID = "507f1f77bcf86cd799439011"

	doc := diagnosisDocFromRecord(rec)
	if doc.ID != rec.Diagnosis.ID {
		t.Fatalf("doc id = %q, want %q", doc.ID, rec.Diagnosis.ID)
	}
	if doc.ServiceName != "order-service" || !doc.CreatedAt.Equal(createdAt) {
		t.Fatalf("storage-side fields lost: %+v", doc)
	}

	out := doc.toDomain()
	if out.ID != rec.Diagnosis.ID || out.SessionID != rec.Diagnosis.SessionID ||
		out.Tenant != rec.Diagnosis.Tenant || out.RootCause != rec.Diagnosis.RootCause {
		t.Fatalf("round-trip field mismatch: %+v", out)
	}
	if out.Severity != domain.SeverityP1 || out.RiskLevel != domain.RiskRead {
		t.Fatalf("enum mismatch: %+v", out)
	}
	if len(out.Conclusions) != 1 || out.Conclusions[0].Citation[0] != "log-abc-123" {
		t.Fatalf("conclusions lost: %+v", out.Conclusions)
	}
	if len(out.Disposition) != 1 || out.Disposition[0].Step != rec.Diagnosis.Disposition[0].Step {
		t.Fatalf("disposition lost: %+v", out.Disposition)
	}
	if len(out.Trace) != 1 || out.Trace[0].Agent != rec.Diagnosis.Trace[0].Agent {
		t.Fatalf("trace lost: %+v", out.Trace)
	}
	if out.Degraded || out.Truncated {
		t.Fatalf("flags should be false: %+v", out)
	}
}

// TestCitationDocMapping domain.Citation → citationDoc 字段对齐（er-diagram citations）。
func TestCitationDocMapping(t *testing.T) {
	fetchedAt := mustUTCMillis(time.Now())
	in := domain.Citation{SourceType: "log", SourceKey: "log-1", Snippet: "snip"}

	doc := citationDocFrom("tenant-a", "diag-9", in, fetchedAt)
	if doc.TenantID != "tenant-a" || doc.DiagnosisID != "diag-9" {
		t.Fatalf("keys not set: %+v", doc)
	}
	if doc.SourceType != "log" || doc.SourceKey != "log-1" || doc.Snippet != "snip" {
		t.Fatalf("payload mismatch: %+v", doc)
	}
	if !doc.FetchedAt.Equal(fetchedAt) {
		t.Fatalf("fetched_at = %v, want %v", doc.FetchedAt, fetchedAt)
	}
	if doc.ID == "" {
		t.Fatal("doc id should be generated")
	}

	out := doc.toDomain()
	if out != in {
		t.Fatalf("round-trip mismatch: got %+v want %+v", out, in)
	}
}

// TestClampLimit 分页 limit 收敛：缺省 20、上限 100（防全表扫）。
func TestClampLimit(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{in: 0, want: defaultListLimit},
		{in: -5, want: defaultListLimit},
		{in: 10, want: 10},
		{in: 500, want: maxListLimit},
	}
	for _, c := range cases {
		if got := clampLimit(c.in); got != c.want {
			t.Fatalf("clampLimit(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}
