package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

// newTestSession 构造一条合法会话（AC-1：含 orchestration_trace / status / created_at）。
func newTestSession(tenant, service, query string, createdAt time.Time) *domain.Session {
	return &domain.Session{
		TenantID:    tenant,
		Type:        domain.SessionTypeChat,
		UserID:      "user-1",
		Query:       query,
		ServiceName: service,
		IntentType:  domain.IntentDiagnose,
		Status:      domain.SessionStatusRunning,
		OrchestrationTrace: []domain.TraceStep{
			{Step: 1, Agent: "coordinator", Action: "intent", DurationMs: 12, Summary: "diagnose"},
			{Step: 2, Agent: "log_analyst", Action: "query", Source: "logquery", DurationMs: 800, Summary: "100 entries"},
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}

// TestSessionDAO_InsertAndGetByID AC-1：会话写入正确落库（含 orchestration_trace、
// status、created_at），并按 (tenant_id, _id) 取回。
func TestSessionDAO_InsertAndGetByID(t *testing.T) {
	db := newTestMongo(t)
	d := NewSessionDAO(db)
	ctx := context.Background()

	in := newTestSession("tenant-a", "order-service", "order-service 最近一小时错误率上升", mustUTCMillis(time.Now()))
	if err := d.Insert(ctx, in); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if in.ID == "" {
		t.Fatal("insert should generate session ID")
	}

	out, err := d.GetByID(ctx, "tenant-a", in.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if out.ID != in.ID || out.TenantID != "tenant-a" || out.ServiceName != "order-service" {
		t.Fatalf("mismatched session: %+v", out)
	}
	if out.Status != domain.SessionStatusRunning {
		t.Fatalf("status = %q, want running", out.Status)
	}
	if len(out.OrchestrationTrace) != 2 || out.OrchestrationTrace[1].Agent != "log_analyst" {
		t.Fatalf("orchestration_trace not persisted: %+v", out.OrchestrationTrace)
	}
	if !out.CreatedAt.Equal(in.CreatedAt) {
		t.Fatalf("created_at = %v, want %v", out.CreatedAt, in.CreatedAt)
	}
}

// TestSessionDAO_ListByTime AC-2（SC-5）：按 (tenant_id, created_at) 历史检索，
// created_at 倒序、分页生效、只命中本租户。
func TestSessionDAO_ListByTime(t *testing.T) {
	db := newTestMongo(t)
	d := NewSessionDAO(db)
	ctx := context.Background()

	base := mustUTCMillis(time.Now().Add(-time.Hour))
	for i, at := range []time.Time{base, base.Add(time.Minute), base.Add(2 * time.Minute)} {
		s := newTestSession("tenant-a", "svc", "q", at)
		if err := d.Insert(ctx, s); err != nil {
			t.Fatalf("insert #%d: %v", i, err)
		}
	}
	other := newTestSession("tenant-b", "svc", "q", base.Add(3*time.Minute))
	if err := d.Insert(ctx, other); err != nil {
		t.Fatalf("insert other tenant: %v", err)
	}

	got, err := d.ListByTime(ctx, "tenant-a", base.Add(-time.Minute), base.Add(time.Hour), 0, 0)
	if err != nil {
		t.Fatalf("list by time: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3 (tenant-a only)", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].CreatedAt.Before(got[i].CreatedAt) {
			t.Fatalf("not sorted desc by created_at: %v before %v", got[i-1].CreatedAt, got[i].CreatedAt)
		}
	}

	page, err := d.ListByTime(ctx, "tenant-a", base.Add(-time.Minute), base.Add(time.Hour), 2, 0)
	if err != nil {
		t.Fatalf("list page: %v", err)
	}
	if len(page) != 2 || page[0].CreatedAt.Before(page[1].CreatedAt) {
		t.Fatalf("pagination broken: %+v", page)
	}
}

// TestSessionDAO_ListByService AC-2（SC-5）：按 (tenant_id, service_name) 检索。
func TestSessionDAO_ListByService(t *testing.T) {
	db := newTestMongo(t)
	d := NewSessionDAO(db)
	ctx := context.Background()

	now := mustUTCMillis(time.Now())
	for _, svc := range []string{"order-service", "order-service", "cdn-service"} {
		if err := d.Insert(ctx, newTestSession("tenant-a", svc, "q", now)); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	if err := d.Insert(ctx, newTestSession("tenant-b", "order-service", "q", now)); err != nil {
		t.Fatalf("insert other tenant: %v", err)
	}

	got, err := d.ListByService(ctx, "tenant-a", "order-service", 0, 0)
	if err != nil {
		t.Fatalf("list by service: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for _, s := range got {
		if s.ServiceName != "order-service" || s.TenantID != "tenant-a" {
			t.Fatalf("unexpected row: %+v", s)
		}
	}
}

// TestSessionDAO_UpdateStatus 会话状态流转（running→done），且带租户谓词。
func TestSessionDAO_UpdateStatus(t *testing.T) {
	db := newTestMongo(t)
	d := NewSessionDAO(db)
	ctx := context.Background()

	s := newTestSession("tenant-a", "svc", "q", mustUTCMillis(time.Now()))
	if err := d.Insert(ctx, s); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := d.UpdateStatus(ctx, "tenant-a", s.ID, domain.SessionStatusDone); err != nil {
		t.Fatalf("update status: %v", err)
	}
	out, err := d.GetByID(ctx, "tenant-a", s.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out.Status != domain.SessionStatusDone {
		t.Fatalf("status = %q, want done", out.Status)
	}
	if out.UpdatedAt.Before(out.CreatedAt) {
		t.Fatalf("updated_at should advance: %v < %v", out.UpdatedAt, out.CreatedAt)
	}

	// 跨租户更新必须不命中（SC-6）
	if err := d.UpdateStatus(ctx, "tenant-x", s.ID, domain.SessionStatusFailed); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant update err = %v, want ErrNoDocuments", err)
	}
}

// TestSessionDAO_AppendCorrection 纠正记录 $push 落库 + 强制租户谓词：
// 跨租户/不存在 → ErrNoDocuments；归属租户追加后回读可见。
func TestSessionDAO_AppendCorrection(t *testing.T) {
	db := newTestMongo(t)
	d := NewSessionDAO(db)
	ctx := context.Background()

	s := newTestSession("tenant-a", "svc", "q", mustUTCMillis(time.Now()))
	if err := d.Insert(ctx, s); err != nil {
		t.Fatalf("insert: %v", err)
	}

	rec := domain.CorrectionRecord{
		From:    domain.IntentResource,
		To:      domain.IntentDiagnose,
		By:      "u1",
		At:      time.Now(),
		Message: "不对，我要排障",
	}
	if err := d.AppendCorrection(ctx, "tenant-a", s.ID, rec); err != nil {
		t.Fatalf("append correction: %v", err)
	}
	out, err := d.GetByID(ctx, "tenant-a", s.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(out.CorrectionLog) != 1 {
		t.Fatalf("correction_log len = %d, want 1", len(out.CorrectionLog))
	}
	got := out.CorrectionLog[0]
	if got.From != domain.IntentResource || got.To != domain.IntentDiagnose || got.By != "u1" {
		t.Fatalf("correction = %+v, want resource→diagnose by u1", got)
	}

	// 跨租户追加必须不命中（SC-6），也不得写入文档。
	if err := d.AppendCorrection(ctx, "tenant-x", s.ID, rec); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant append err = %v, want ErrNoDocuments", err)
	}
	out, err = d.GetByID(ctx, "tenant-a", s.ID)
	if err != nil {
		t.Fatalf("get after cross-tenant attempt: %v", err)
	}
	if len(out.CorrectionLog) != 1 {
		t.Fatalf("cross-tenant attempt leaked a write: len = %d, want 1", len(out.CorrectionLog))
	}
}

// TestSessionDAO_TenantIsolation AC-3（SC-6）：跨租户 GetByID 返回 NotFound 语义、
// 检索返回空；非法 ID 与不存在等价（不泄露资源存在性，service 层映射 404）。
func TestSessionDAO_TenantIsolation(t *testing.T) {
	db := newTestMongo(t)
	d := NewSessionDAO(db)
	ctx := context.Background()

	s := newTestSession("tenant-a", "svc", "q", mustUTCMillis(time.Now()))
	if err := d.Insert(ctx, s); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if _, err := d.GetByID(ctx, "tenant-b", s.ID); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant get err = %v, want ErrNoDocuments", err)
	}
	if _, err := d.GetByID(ctx, "tenant-a", "not-a-hex-id"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("invalid id err = %v, want ErrNoDocuments", err)
	}

	list, err := d.ListByService(ctx, "tenant-b", "svc", 0, 0)
	if err != nil {
		t.Fatalf("cross-tenant list: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("cross-tenant list should be empty, got %d rows", len(list))
	}
}
