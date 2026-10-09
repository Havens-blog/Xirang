package dao

import (
	"context"
	"testing"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// TestCitationDAO_InsertBatchAndListByDiagnosis AC-1 + AC-4：引用落库，
// 按 diagnosis_id 关联查询（idx_cit_dx）。
func TestCitationDAO_InsertBatchAndListByDiagnosis(t *testing.T) {
	db := newTestMongo(t)
	d := NewCitationDAO(db)
	ctx := context.Background()

	in := []domain.Citation{
		{SourceType: "log", SourceKey: "log-abc-123", Snippet: "upstream_timeout ..."},
		{SourceType: "asset", SourceKey: "ecs-i-0e12", Snippet: "ecs us-east-1 running"},
	}
	if err := d.InsertBatch(ctx, "tenant-a", "diag-1", in); err != nil {
		t.Fatalf("insert batch: %v", err)
	}

	got, err := d.ListByDiagnosis(ctx, "tenant-a", "diag-1")
	if err != nil {
		t.Fatalf("list by diagnosis: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	byKey := map[string]domain.Citation{}
	for _, c := range got {
		byKey[c.SourceKey] = c
	}
	if c, ok := byKey["log-abc-123"]; !ok || c.SourceType != "log" || c.Snippet == "" {
		t.Fatalf("citation log-abc-123 not persisted: %+v", got)
	}
	if _, ok := byKey["ecs-i-0e12"]; !ok {
		t.Fatalf("citation ecs-i-0e12 not persisted: %+v", got)
	}
}

// TestCitationDAO_AssociationIsolation AC-4：引用严格按 diagnosis_id 归属，
// 诊断 A 的引用不会出现在诊断 B 的关联查询中。
func TestCitationDAO_AssociationIsolation(t *testing.T) {
	db := newTestMongo(t)
	d := NewCitationDAO(db)
	ctx := context.Background()

	if err := d.InsertBatch(ctx, "tenant-a", "diag-1", []domain.Citation{
		{SourceType: "log", SourceKey: "log-a", Snippet: "s"},
	}); err != nil {
		t.Fatalf("insert diag-1: %v", err)
	}
	if err := d.InsertBatch(ctx, "tenant-a", "diag-2", []domain.Citation{
		{SourceType: "metric", SourceKey: "metric-b", Snippet: "s"},
	}); err != nil {
		t.Fatalf("insert diag-2: %v", err)
	}

	got, err := d.ListByDiagnosis(ctx, "tenant-a", "diag-1")
	if err != nil {
		t.Fatalf("list diag-1: %v", err)
	}
	if len(got) != 1 || got[0].SourceKey != "log-a" {
		t.Fatalf("association leaked: %+v", got)
	}
}

// TestCitationDAO_TenantIsolation AC-3（SC-6）：跨租户查询引用返回空。
func TestCitationDAO_TenantIsolation(t *testing.T) {
	db := newTestMongo(t)
	d := NewCitationDAO(db)
	ctx := context.Background()

	if err := d.InsertBatch(ctx, "tenant-a", "diag-1", []domain.Citation{
		{SourceType: "alert", SourceKey: "alert-1", Snippet: "s"},
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := d.ListByDiagnosis(ctx, "tenant-b", "diag-1")
	if err != nil {
		t.Fatalf("cross-tenant list: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("cross-tenant list should be empty, got %d rows", len(got))
	}
}
