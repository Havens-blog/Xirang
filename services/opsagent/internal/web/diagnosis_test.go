package web

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

// ---- fakes ----

type fakeDiagReader struct {
	diag      *domain.Diagnosis
	bySession *domain.Diagnosis
	err       error
}

func (f *fakeDiagReader) GetByID(context.Context, string, string) (*domain.Diagnosis, error) {
	return f.diag, f.err
}
func (f *fakeDiagReader) GetBySession(context.Context, string, string) (*domain.Diagnosis, error) {
	return f.bySession, f.err
}

type fakeCiteReader struct {
	cites []domain.Citation
	err   error
}

func (f *fakeCiteReader) ListByDiagnosis(context.Context, string, string) ([]domain.Citation, error) {
	return f.cites, f.err
}

type fakeRiskViewer struct {
	viewed string
	diagID string
}

func (f *fakeRiskViewer) MarkViewed(_ context.Context, _, id, diagnosisID, _ string) (*domain.RiskEntry, bool, error) {
	f.viewed = id
	f.diagID = diagnosisID
	return &domain.RiskEntry{}, true, nil
}

type fakeHistoryReader struct {
	sessions []domain.Session
	sess     domain.Session
	err      error
}

func (f *fakeHistoryReader) GetByID(context.Context, string, string) (domain.Session, error) {
	return f.sess, f.err
}
func (f *fakeHistoryReader) ListByTime(context.Context, string, time.Time, time.Time, int, int) ([]domain.Session, error) {
	return f.sessions, f.err
}
func (f *fakeHistoryReader) ListByService(context.Context, string, string, int, int) ([]domain.Session, error) {
	return f.sessions, f.err
}

func newDiagRouter(h *DiagnosisHandler) *gin.Engine {
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

func decodeDiag(t *testing.T, env envelope) *domain.Diagnosis {
	t.Helper()
	var d domain.Diagnosis
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode diagnosis: %v", err)
	}
	return &d
}

// ---- tests ----

func TestDiagnosisDetail(t *testing.T) {
	diag := &domain.Diagnosis{ID: "dx-1", RootCause: "连接池耗尽"}
	cites := &fakeCiteReader{cites: []domain.Citation{{SourceKey: "order-service"}}}
	h := NewDiagnosisHandler(&fakeDiagReader{diag: diag}, cites, &fakeRiskViewer{}, &fakeHistoryReader{})
	r := newDiagRouter(h)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/diagnosis/dx-1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	got := decodeDiag(t, env)
	if got.ID != "dx-1" || len(got.Citations) != 1 {
		t.Fatalf("diag = %+v", got)
	}
}

func TestDiagnosisDetail_NotFound(t *testing.T) {
	h := NewDiagnosisHandler(&fakeDiagReader{err: mongo.ErrNoDocuments}, &fakeCiteReader{}, &fakeRiskViewer{}, &fakeHistoryReader{})
	r := newDiagRouter(h)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/diagnosis/dx-1", "")
	if rec.Code != http.StatusNotFound || env.Code != middleware.CodeNotFound {
		t.Fatalf("status=%d code=%v", rec.Code, env.Code)
	}
}

func TestDiagnosisDetail_MarkViewed(t *testing.T) {
	viewer := &fakeRiskViewer{}
	h := NewDiagnosisHandler(&fakeDiagReader{diag: &domain.Diagnosis{ID: "dx-1"}}, &fakeCiteReader{}, viewer, &fakeHistoryReader{})
	r := newDiagRouter(h)

	rec, _ := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/diagnosis/dx-1?riskEntryId=re-1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if viewer.viewed != "re-1" || viewer.diagID != "dx-1" {
		t.Fatalf("MarkViewed not invoked correctly = %+v", viewer)
	}
}

func TestHistory_SessionId(t *testing.T) {
	diag := &fakeDiagReader{bySession: &domain.Diagnosis{ID: "dx-1"}}
	hist := &fakeHistoryReader{sess: domain.Session{ID: "s1", Type: domain.SessionTypeChat, Query: "查错误率", Status: domain.SessionStatusDone}}
	h := NewDiagnosisHandler(diag, &fakeCiteReader{}, &fakeRiskViewer{}, hist)
	r := newDiagRouter(h)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/history?sessionId=s1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var d historyData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if len(d.Items) != 1 || d.Items[0].DiagnosisID != "dx-1" {
		t.Fatalf("items = %+v", d.Items)
	}
}

func TestHistory_List(t *testing.T) {
	hist := &fakeHistoryReader{sessions: []domain.Session{
		{ID: "s1", Type: domain.SessionTypeChat, Query: "查错误率", ServiceName: "order-service"},
	}}
	h := NewDiagnosisHandler(&fakeDiagReader{}, &fakeCiteReader{}, &fakeRiskViewer{}, hist)
	r := newDiagRouter(h)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/history?serviceName=order", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var d historyData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if len(d.Items) != 1 || d.Items[0].ServiceName != "order-service" {
		t.Fatalf("items = %+v", d.Items)
	}
}

func TestHistory_WindowTooWide(t *testing.T) {
	h := NewDiagnosisHandler(&fakeDiagReader{}, &fakeCiteReader{}, &fakeRiskViewer{}, &fakeHistoryReader{})
	r := newDiagRouter(h)

	rec, _ := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/history?startTime=2026-10-01T00:00:00Z&endTime=2026-10-03T00:00:00Z", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHistory_CrossTenant(t *testing.T) {
	hist := &fakeHistoryReader{err: mongo.ErrNoDocuments}
	h := NewDiagnosisHandler(&fakeDiagReader{}, &fakeCiteReader{}, &fakeRiskViewer{}, hist)
	r := newDiagRouter(h)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/history?sessionId=s-x", "")
	if rec.Code != http.StatusNotFound || env.Code != middleware.CodeNotFound {
		t.Fatalf("status=%d code=%v", rec.Code, env.Code)
	}
}
