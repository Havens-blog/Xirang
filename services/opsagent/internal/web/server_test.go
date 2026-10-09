package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"

	"github.com/Havens-blog/opsagent/internal/config"
)

// ---- 测试替身（与 middleware 包内一致的手工 fake，测试代码不外溢到生产包）----

type fakeSession struct {
	claims session.Claims
}

func (f *fakeSession) Set(_ context.Context, _ string, _ any) error { return nil }
func (f *fakeSession) Get(_ context.Context, _ string) ekit.AnyValue {
	return ekit.AnyValue{}
}
func (f *fakeSession) Del(_ context.Context, _ string) error { return nil }
func (f *fakeSession) Destroy(_ context.Context) error       { return nil }
func (f *fakeSession) Claims() session.Claims                { return f.claims }

type fakeProvider struct {
	sess session.Session
	err  error
}

func (f *fakeProvider) NewSession(_ *gctx.Context, _ int64, _ map[string]string, _ map[string]any) (session.Session, error) {
	return f.sess, f.err
}
func (f *fakeProvider) Get(_ *gctx.Context) (session.Session, error)         { return f.sess, f.err }
func (f *fakeProvider) Destroy(_ *gctx.Context) error                        { return nil }
func (f *fakeProvider) UpdateClaims(_ *gctx.Context, _ session.Claims) error { return nil }
func (f *fakeProvider) RenewAccessToken(_ *gctx.Context) error               { return nil }

// echoModule 测试用模块：回显当前上下文租户。
type echoModule struct{}

func (echoModule) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/echo-tenant", func(c *gin.Context) {
		OK(c, gin.H{"tenant": c.GetInt64("tenant_id")})
	})
}

func testConfig() *config.Config {
	return &config.Config{
		Session: config.SessionConfig{
			Whitelist: []string{"/api/v1/opsagent/alerts"},
		},
	}
}

func newTestServer(sp session.Provider) *Server {
	gin.SetMode(gin.TestMode)
	return NewServer(testConfig(), sp, elog.DefaultLogger, echoModule{})
}

func TestNewServer_HealthEndpointOpenAccess(t *testing.T) {
	s := newTestServer(&fakeProvider{err: context.DeadlineExceeded})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opsagent/health", nil)
	s.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health 应 200, 实际 %d", w.Code)
	}
	var body Response
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v", err)
	}
	if body.Code != float64(0) || body.Message != "ok" {
		t.Errorf("health 信封错误: %+v", body)
	}
}

func TestNewServer_UnauthenticatedReturns401TenantDenied(t *testing.T) {
	s := newTestServer(&fakeProvider{err: context.DeadlineExceeded})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opsagent/echo-tenant", nil)
	s.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("未认证应 401, 实际 %d", w.Code)
	}
	var body Response
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Code != "ERR_TENANT_DENIED" {
		t.Errorf("错误码 = %v, 期望 ERR_TENANT_DENIED", body.Code)
	}
}

func TestNewServer_AuthenticatedRequestReachesModule(t *testing.T) {
	sp := &fakeProvider{sess: &fakeSession{claims: session.Claims{
		Uid:  7,
		Data: map[string]string{"tenant_id": "42"},
	}}}
	s := newTestServer(sp)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opsagent/echo-tenant", nil)
	s.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("已认证请求应 200, 实际 %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Code any `json:"code"`
		Data struct {
			Tenant int64 `json:"tenant"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应体解析失败: %v", err)
	}
	if body.Code != float64(0) || body.Data.Tenant != 42 {
		t.Errorf("信封或租户注入错误: %+v", body)
	}
}

func TestNewServer_SessionWithoutTenantRejected(t *testing.T) {
	sp := &fakeProvider{sess: &fakeSession{claims: session.Claims{Uid: 7}}}
	s := newTestServer(sp)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opsagent/echo-tenant", nil)
	s.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("会话未选定租户应 403, 实际 %d", w.Code)
	}
}

func TestFail_EnvelopeShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/boom", func(c *gin.Context) {
		Fail(c, http.StatusBadRequest, "ERR_PARAM_INVALID", "参数非法")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	r.ServeHTTP(w, req)

	var body Response
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v", err)
	}
	if body.Code != "ERR_PARAM_INVALID" || body.Message != "参数非法" {
		t.Errorf("Fail 信封错误: %+v", body)
	}
}
