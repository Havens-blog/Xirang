package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"github.com/Havens-blog/opsagent/internal/worker"
)

type fakeEnqueuer struct {
	result worker.EnqueueResult
	alerts []domain.AlertEvent
}

func (f *fakeEnqueuer) Enqueue(_ context.Context, a domain.AlertEvent) worker.EnqueueResult {
	f.alerts = append(f.alerts, a)
	return f.result
}

func newAlertRouter(h *AlertWebhookHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1/opsagent"))
	return r
}

func doWebhook(t *testing.T, r *gin.Engine, key, body string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/opsagent/alerts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set(webhookKeyHeader, key)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return rec, env
}

const validAlertBody = `{"tenant":"t1","serviceName":"order-service","metric":"错误率","level":"high"}`

func TestAlert_InvalidKey(t *testing.T) {
	h := NewAlertWebhookHandler([]string{"k1"}, &fakeEnqueuer{})
	r := newAlertRouter(h)

	rec, env := doWebhook(t, r, "", validAlertBody)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if env.Code != middleware.CodeTenantDenied {
		t.Fatalf("code = %v", env.Code)
	}
}

func TestAlert_WrongKey(t *testing.T) {
	h := NewAlertWebhookHandler([]string{"k1"}, &fakeEnqueuer{})
	r := newAlertRouter(h)

	rec, _ := doWebhook(t, r, "wrong-key", validAlertBody)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAlert_MissingFields(t *testing.T) {
	h := NewAlertWebhookHandler([]string{"k1"}, &fakeEnqueuer{})
	r := newAlertRouter(h)

	rec, _ := doWebhook(t, r, "k1", `{"serviceName":"svc"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAlert_Accepted(t *testing.T) {
	enq := &fakeEnqueuer{result: worker.EnqueueResult{Accepted: true}}
	h := NewAlertWebhookHandler([]string{"k1"}, enq)
	r := newAlertRouter(h)

	rec, env := doWebhook(t, r, "k1", validAlertBody)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", rec.Code)
	}
	var d struct {
		Accepted bool `json:"accepted"`
		Dropped  bool `json:"dropped"`
	}
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if !d.Accepted || d.Dropped {
		t.Fatalf("data = %+v, want accepted=true dropped=false", d)
	}
	if len(enq.alerts) != 1 || enq.alerts[0].Tenant != "t1" {
		t.Fatalf("enqueued alerts = %+v", enq.alerts)
	}
}

func TestAlert_Dropped(t *testing.T) {
	enq := &fakeEnqueuer{result: worker.EnqueueResult{Accepted: true, Dropped: true}}
	h := NewAlertWebhookHandler([]string{"k1"}, enq)
	r := newAlertRouter(h)

	rec, env := doWebhook(t, r, "k1", validAlertBody)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", rec.Code)
	}
	var d struct {
		Accepted bool `json:"accepted"`
		Dropped  bool `json:"dropped"`
	}
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if !d.Accepted || !d.Dropped {
		t.Fatalf("data = %+v, want accepted=true dropped=true", d)
	}
}

// TestAlert_DualKey 双密钥并存滚动轮换：新旧密钥均有效。
func TestAlert_DualKey(t *testing.T) {
	enq := &fakeEnqueuer{result: worker.EnqueueResult{Accepted: true}}
	h := NewAlertWebhookHandler([]string{"old-key", "new-key"}, enq)
	r := newAlertRouter(h)

	if rec, _ := doWebhook(t, r, "old-key", validAlertBody); rec.Code != http.StatusAccepted {
		t.Fatalf("old-key status = %d", rec.Code)
	}
	if rec, _ := doWebhook(t, r, "new-key", validAlertBody); rec.Code != http.StatusAccepted {
		t.Fatalf("new-key status = %d", rec.Code)
	}
}
