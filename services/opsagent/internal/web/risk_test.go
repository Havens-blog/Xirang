package web

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"github.com/Havens-blog/opsagent/internal/repository/dao"
	"go.mongodb.org/mongo-driver/mongo"
)

// ---- fakes ----

type fakeRiskStore struct {
	entries []domain.RiskEntry
	total   int64
	stats   *dao.RiskStats
	listErr error

	updateEntry *domain.RiskEntry
	updateErr   error

	batchRes *dao.BatchStatusResult
	batchErr error
}

func (f *fakeRiskStore) List(_ context.Context, _ string, _ dao.RiskEntryFilter, _, _ int) ([]domain.RiskEntry, int64, error) {
	return f.entries, f.total, f.listErr
}

func (f *fakeRiskStore) UpdateStatus(_ context.Context, _, _ string, _ domain.RiskEntryStatus, _ int, _ string) (*domain.RiskEntry, error) {
	return f.updateEntry, f.updateErr
}

func (f *fakeRiskStore) BatchUpdateStatus(_ context.Context, _ string, _ domain.RiskEntryStatus, _ []dao.BatchStatusItem) (*dao.BatchStatusResult, error) {
	return f.batchRes, f.batchErr
}

func (f *fakeRiskStore) Stats(_ context.Context, _ string) (*dao.RiskStats, error) {
	return f.stats, nil
}

func newRiskRouter(h *RiskCenterHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.TenantIDKey, int64(1))
		c.Set(middleware.CtxUidKey, int64(99))
		c.Next()
	})
	h.RegisterRoutes(r.Group("/api/v1/opsagent"))
	return r
}

func testEntry() domain.RiskEntry {
	return domain.RiskEntry{
		ID:          "re-1",
		Tenant:      "1",
		ServiceName: "order-service",
		Severity:    domain.SeverityP1,
		Status:      domain.RiskStatusPendingView,
		Version:     2,
	}
}

// ---- tests ----

func TestRiskList(t *testing.T) {
	store := &fakeRiskStore{
		entries: []domain.RiskEntry{testEntry()},
		total:   1,
		stats:   &dao.RiskStats{PendingView: 1, TodayNew: 1, HighRisk: 1},
	}
	h := NewRiskCenterHandler(store)
	r := newRiskRouter(h)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/risk-center?serviceName=order&status=pending_view", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var d riskListData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(d.Items) != 1 || d.Total != 1 {
		t.Fatalf("items=%d total=%d", len(d.Items), d.Total)
	}
	if d.Stats == nil || d.Stats.PendingView != 1 {
		t.Fatalf("stats = %+v", d.Stats)
	}
	if d.Page != 1 || d.Limit != 20 {
		t.Fatalf("page=%d limit=%d", d.Page, d.Limit)
	}
}

func TestRiskList_InvalidTime(t *testing.T) {
	h := NewRiskCenterHandler(&fakeRiskStore{})
	r := newRiskRouter(h)

	rec, _ := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/risk-center?startTime=not-a-time", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestRiskUpdateStatus(t *testing.T) {
	store := &fakeRiskStore{updateEntry: &domain.RiskEntry{ID: "re-1", Version: 3, Status: domain.RiskStatusViewed}}
	h := NewRiskCenterHandler(store)
	r := newRiskRouter(h)

	rec, env := doRequest(t, r, http.MethodPost, "/api/v1/opsagent/risk-center/re-1/status", `{"status":"viewed","expectedVersion":2}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var entry domain.RiskEntry
	if err := json.Unmarshal(env.Data, &entry); err != nil {
		t.Fatalf("decode entry: %v", err)
	}
	if entry.Version != 3 || entry.Status != domain.RiskStatusViewed {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestRiskUpdateStatus_Conflict(t *testing.T) {
	latest := &domain.RiskEntry{ID: "re-1", Version: 5, Status: domain.RiskStatusViewed}
	store := &fakeRiskStore{updateErr: &dao.ConflictError{Current: latest}}
	h := NewRiskCenterHandler(store)
	r := newRiskRouter(h)

	rec, env := doRequest(t, r, http.MethodPost, "/api/v1/opsagent/risk-center/re-1/status", `{"status":"viewed","expectedVersion":2}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	if env.Code != middleware.CodeConflict {
		t.Fatalf("code = %v", env.Code)
	}
	var entry domain.RiskEntry
	if err := json.Unmarshal(env.Data, &entry); err != nil {
		t.Fatalf("decode conflict entry: %v", err)
	}
	if entry.Version != 5 {
		t.Fatalf("conflict 应回显最新条目, version=%d", entry.Version)
	}
}

func TestRiskUpdateStatus_InvalidStatus(t *testing.T) {
	h := NewRiskCenterHandler(&fakeRiskStore{})
	r := newRiskRouter(h)

	rec, _ := doRequest(t, r, http.MethodPost, "/api/v1/opsagent/risk-center/re-1/status", `{"status":"foo"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestRiskBatch(t *testing.T) {
	store := &fakeRiskStore{batchRes: &dao.BatchStatusResult{
		Succeeded: []dao.BatchStatusSuccess{{ID: "re-1", Version: 3}},
		Failed:    []dao.BatchStatusFailure{{ID: "re-2", ReasonCode: "conflict", CurrentVersion: 4}},
	}}
	h := NewRiskCenterHandler(store)
	r := newRiskRouter(h)

	rec, env := doRequest(t, r, http.MethodPost, "/api/v1/opsagent/risk-center/batch-status",
		`{"status":"done","items":[{"id":"re-1","expectedVersion":2},{"id":"re-2","expectedVersion":1}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var d batchStatusData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode batch: %v", err)
	}
	if len(d.Succeeded) != 1 || d.Succeeded[0].Version != 3 {
		t.Fatalf("succeeded = %+v", d.Succeeded)
	}
	if len(d.Failed) != 1 || d.Failed[0].ReasonCode != "conflict" || d.Failed[0].CurrentVersion != 4 {
		t.Fatalf("failed = %+v", d.Failed)
	}
	if d.Failed[0].Reason == "" {
		t.Fatal("failed reason should not be empty")
	}
}

func TestRiskBatch_TooMany(t *testing.T) {
	h := NewRiskCenterHandler(&fakeRiskStore{})
	r := newRiskRouter(h)

	body := `{"status":"done","items":[`
	for i := 0; i < 51; i++ {
		if i > 0 {
			body += ","
		}
		body += `{"id":"re-1","expectedVersion":1}`
	}
	body += `]}`
	rec, _ := doRequest(t, r, http.MethodPost, "/api/v1/opsagent/risk-center/batch-status", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestRiskUpdateStatus_NotFound(t *testing.T) {
	store := &fakeRiskStore{updateErr: mongo.ErrNoDocuments}
	h := NewRiskCenterHandler(store)
	r := newRiskRouter(h)

	rec, env := doRequest(t, r, http.MethodPost, "/api/v1/opsagent/risk-center/re-x/status", `{"status":"viewed"}`)
	if rec.Code != http.StatusNotFound || env.Code != middleware.CodeNotFound {
		t.Fatalf("status=%d code=%v, want 404 ERR_NOT_FOUND", rec.Code, env.Code)
	}
}

func TestRiskList_EndTimeAndClamp(t *testing.T) {
	h := NewRiskCenterHandler(&fakeRiskStore{stats: &dao.RiskStats{}})
	r := newRiskRouter(h)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/risk-center?endTime=2026-10-08T00:00:00Z&limit=999", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var d riskListData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if d.Limit != 100 {
		t.Fatalf("limit should clamp to 100, got %d", d.Limit)
	}
}

func TestReasonText(t *testing.T) {
	for _, code := range []string{"conflict", "not_found", "invalid"} {
		if reasonText(code) == "" {
			t.Fatalf("reasonText(%q) should not be empty", code)
		}
	}
}

func TestParsePositiveInt(t *testing.T) {
	if parsePositiveInt("", 20) != 20 || parsePositiveInt("0", 20) != 20 ||
		parsePositiveInt("-1", 20) != 20 || parsePositiveInt("abc", 20) != 20 {
		t.Fatal("parsePositiveInt should fall back for invalid input")
	}
	if parsePositiveInt("5", 20) != 5 {
		t.Fatal("parsePositiveInt(5) should be 5")
	}
}
