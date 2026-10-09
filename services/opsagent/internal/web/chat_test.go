package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"github.com/Havens-blog/opsagent/internal/orchestrator"
	"go.mongodb.org/mongo-driver/mongo"
)

// ---- fakes ----

type fakeChatStore struct {
	inserted *domain.Session
	get      domain.Session
	getErr   error
}

func (f *fakeChatStore) Insert(_ context.Context, s *domain.Session) error {
	s.ID = "sess-1"
	f.inserted = s
	return nil
}

func (f *fakeChatStore) GetByID(_ context.Context, _, _ string) (domain.Session, error) {
	return f.get, f.getErr
}

type fakeIntent struct {
	result        domain.IntentResult
	err           error
	correctResult domain.IntentResult
	corrected     bool
}

func (f *fakeIntent) Classify(context.Context, string, string) (domain.IntentResult, error) {
	return f.result, f.err
}

func (f *fakeIntent) Correct(context.Context, string, string, domain.IntentType, domain.DiagnoseParams, string) (domain.IntentResult, error) {
	f.corrected = true
	return f.correctResult, nil
}

type fakeRunner struct {
	diag           *domain.Diagnosis
	err            error
	blockUntilDone bool
}

func (f *fakeRunner) RunChat(ctx context.Context, _ *domain.Session, _ string) (*domain.Diagnosis, error) {
	if f.blockUntilDone {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return f.diag, f.err
}

// diagnoseIntent 高置信排障意图结果。
func diagnoseIntent() domain.IntentResult {
	return domain.IntentResult{
		Type:       domain.IntentDiagnose,
		Confidence: 0.9,
		Params:     domain.DiagnoseParams{ServiceName: "order-service"},
	}
}

func testDiag() *domain.Diagnosis {
	return &domain.Diagnosis{
		ID:          "dx-1",
		SessionID:   "sess-1",
		RootCause:   "数据库连接池耗尽",
		Severity:    domain.SeverityP1,
		Confidence:  0.85,
		Conclusions: []domain.Conclusion{{Text: "连接池打满", Citation: []string{"order-service"}}},
	}
}

// newChatRouter 组装测试路由（注入 tenant/uid 上下文）。
func newChatRouter(h *ChatHandler) *gin.Engine {
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

func newTestHandler(store *fakeChatStore, intent domain.IntentClassifier, runner chatRunner,
	llmOK bool, presets []domain.PresetQuery, budget time.Duration) *ChatHandler {
	degrade := orchestrator.NewDegradationController(0, 0.6, nil)
	llm := func() bool { return llmOK }
	presetsFn := func(_ string) []domain.PresetQuery { return presets }
	return NewChatHandler(store, intent, runner, degrade, llm, presetsFn, nil, budget)
}

type envelope struct {
	Code    any             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func doRequest(t *testing.T, r *gin.Engine, method, path, body string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	var reqBody *strings.Reader
	if body == "" {
		reqBody = strings.NewReader("")
	} else {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v (body=%s)", err, rec.Body.String())
	}
	return rec, env
}

func decodeData(t *testing.T, raw json.RawMessage) chatData {
	t.Helper()
	var d chatData
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatalf("decode chatData: %v (raw=%s)", err, raw)
	}
	return d
}

// ---- tests ----

func TestChat_Report(t *testing.T) {
	store := &fakeChatStore{}
	runner := &fakeRunner{diag: testDiag()}
	h := newTestHandler(store, &fakeIntent{result: diagnoseIntent()}, runner, true, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":"order-service 错误率"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	d := decodeData(t, env.Data)
	if d.Type != chatTypeReport {
		t.Fatalf("type = %q, want report", d.Type)
	}
	if d.Diagnosis == nil || d.Diagnosis.RootCause != "数据库连接池耗尽" {
		t.Fatalf("diagnosis = %+v", d.Diagnosis)
	}
	if d.DegradeLevel != 0 {
		t.Fatalf("degradeLevel = %d, want 0 (LLM 可用)", d.DegradeLevel)
	}
	if d.SessionID != "sess-1" {
		t.Fatalf("sessionId = %q", d.SessionID)
	}
}

func TestChat_Guided(t *testing.T) {
	h := newTestHandler(&fakeChatStore{}, &fakeIntent{result: domain.IntentResult{Type: domain.IntentOutOfScope, Confidence: 1}}, &fakeRunner{}, true, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":"帮我写脚本"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	d := decodeData(t, env.Data)
	if d.Type != chatTypeGuided {
		t.Fatalf("type = %q, want guided", d.Type)
	}
	if d.Report == "" {
		t.Fatal("guided report should not be empty")
	}
}

func TestChat_Clarify(t *testing.T) {
	intent := diagnoseIntent()
	intent.NeedsClarify = true
	intent.Confidence = 0.4
	intent.Alternatives = []domain.IntentCandidate{{Type: domain.IntentDiagnose, Confidence: 0.8}}
	h := newTestHandler(&fakeChatStore{}, &fakeIntent{result: intent}, &fakeRunner{}, true, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":"最近有点高"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	d := decodeData(t, env.Data)
	if d.Type != chatTypeClarify || !d.NeedsClarify {
		t.Fatalf("type=%q needsClarify=%v, want clarify", d.Type, d.NeedsClarify)
	}
	if len(d.Candidates) != 1 {
		t.Fatalf("candidates = %+v", d.Candidates)
	}
}

func TestChat_PresetEntries(t *testing.T) {
	intent := domain.IntentResult{Type: domain.IntentDiagnose, Confidence: 0.3, Params: domain.DiagnoseParams{ServiceName: "order-service"}}
	presets := []domain.PresetQuery{{ID: "p1", Label: "order-service 近1h错误率"}}
	h := newTestHandler(&fakeChatStore{}, &fakeIntent{result: intent}, &fakeRunner{}, false, presets, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":"错误率"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	d := decodeData(t, env.Data)
	if d.Type != chatTypePresetEntries {
		t.Fatalf("type = %q, want preset_entries", d.Type)
	}
	if d.DegradeLevel != 2 {
		t.Fatalf("degradeLevel = %d, want 2", d.DegradeLevel)
	}
	if len(d.Presets) != 1 {
		t.Fatalf("presets = %+v", d.Presets)
	}
}

func TestChat_DegradedNotice(t *testing.T) {
	intent := domain.IntentResult{Type: domain.IntentDiagnose, Confidence: 0.3}
	h := newTestHandler(&fakeChatStore{}, &fakeIntent{result: intent}, &fakeRunner{}, false, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":"错误率"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	d := decodeData(t, env.Data)
	if d.Type != chatTypeDegradedNotice || d.DegradeLevel != 3 {
		t.Fatalf("type=%q level=%d, want degraded_notice/3", d.Type, d.DegradeLevel)
	}
}

func TestChat_PersistFailed(t *testing.T) {
	runner := &fakeRunner{diag: testDiag(), err: errors.New("mongo down")}
	h := newTestHandler(&fakeChatStore{}, &fakeIntent{result: diagnoseIntent()}, runner, true, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":"order-service 错误率"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (信封语义)", rec.Code)
	}
	if env.Code != middleware.CodePersistFailed {
		t.Fatalf("code = %v, want ERR_PERSIST_FAILED", env.Code)
	}
	d := decodeData(t, env.Data)
	if d.Report == "" || d.Diagnosis == nil {
		t.Fatalf("data.report 应仍返回正文, got report=%q diag=%+v", d.Report, d.Diagnosis)
	}
}

func TestChat_202Async(t *testing.T) {
	runner := &fakeRunner{blockUntilDone: true}
	h := newTestHandler(&fakeChatStore{}, &fakeIntent{result: diagnoseIntent()}, runner, true, nil, 50*time.Millisecond)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":"order-service 错误率"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", rec.Code)
	}
	var async struct {
		SessionID string `json:"sessionId"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(env.Data, &async); err != nil {
		t.Fatalf("decode async: %v", err)
	}
	if async.SessionID != "sess-1" || async.Status != "running" {
		t.Fatalf("async = %+v", async)
	}
}

func TestChat_EmptyMessage(t *testing.T) {
	h := newTestHandler(&fakeChatStore{}, &fakeIntent{}, &fakeRunner{}, true, nil, 0)
	r := newChatRouter(h)

	rec, _ := doRequest(t, r, "POST", "/api/v1/opsagent/chat", `{"message":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCorrect_CrossTenant404(t *testing.T) {
	store := &fakeChatStore{getErr: mongo.ErrNoDocuments}
	h := newTestHandler(store, &fakeIntent{}, &fakeRunner{}, true, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat/sess-x/correct", `{"message":"不对，我要查资产"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if env.Code != middleware.CodeNotFound {
		t.Fatalf("code = %v, want ERR_NOT_FOUND", env.Code)
	}
}

func TestCorrect_SuccessMessage(t *testing.T) {
	store := &fakeChatStore{get: domain.Session{ID: "sess-1", TenantID: "1"}}
	intent := &fakeIntent{correctResult: diagnoseIntent()}
	runner := &fakeRunner{diag: testDiag()}
	h := newTestHandler(store, intent, runner, true, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat/sess-1/correct", `{"message":"不对，是 order-service 的错误率"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !intent.corrected {
		t.Fatal("Correct should be invoked")
	}
	d := decodeData(t, env.Data)
	if d.Type != chatTypeReport || d.Diagnosis == nil {
		t.Fatalf("type=%q diagnosis=%+v, want report", d.Type, d.Diagnosis)
	}
}

func TestCorrect_SuccessTarget(t *testing.T) {
	store := &fakeChatStore{get: domain.Session{ID: "sess-1", TenantID: "1"}}
	intent := &fakeIntent{correctResult: diagnoseIntent()}
	runner := &fakeRunner{diag: testDiag()}
	h := newTestHandler(store, intent, runner, true, nil, 0)
	r := newChatRouter(h)

	rec, env := doRequest(t, r, "POST", "/api/v1/opsagent/chat/sess-1/correct", `{"targetIntent":"diagnose","params":{"serviceName":"order-service"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	d := decodeData(t, env.Data)
	if d.Type != chatTypeReport {
		t.Fatalf("type = %q, want report", d.Type)
	}
}

func TestCorrect_XOR(t *testing.T) {
	store := &fakeChatStore{get: domain.Session{ID: "sess-1", TenantID: "1"}}
	h := newTestHandler(store, &fakeIntent{}, &fakeRunner{}, true, nil, 0)
	r := newChatRouter(h)

	// 两者都提供 → 400。
	rec, _ := doRequest(t, r, "POST", "/api/v1/opsagent/chat/sess-1/correct", `{"targetIntent":"resource","message":"查资产"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("both provided status = %d, want 400", rec.Code)
	}
}
