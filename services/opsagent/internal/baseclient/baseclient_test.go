package baseclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// newTestClient 以 httptest server 作 logquery base 构造客户端。
func newTestClient(t *testing.T, handler http.HandlerFunc, cfg Config) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cfg.LogqueryBaseURL = srv.URL
	return New(cfg)
}

func testLogRequest() domain.LogQueryRequest {
	return domain.LogQueryRequest{
		LogType:   "cdn",
		Cloud:     "aliyun",
		Resource:  "order-service",
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
		Limit:     10,
	}
}

// TestQueryLogs_Success B1 /search 成功：信封解析 + 多态条目裁剪为最小稳定子集。
func TestQueryLogs_Success(t *testing.T) {
	ts := int64(1720000000000)
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cam/logs/search" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(
			`{"code":0,"msg":"ok","data":{"log_type":"cdn","total":1,"truncated":false,` +
				`"entries":[{"timestamp":1720000000000,"meta":{"cloud":"aliyun","resource_id":"order-service"},"client_ip":"1.2.3.4"}],` +
				`"sources":null,"cached":false,"cache_stale":false}}`))
	}, Config{})

	res, err := c.QueryLogs(context.Background(), testLogRequest())
	if err != nil {
		t.Fatalf("QueryLogs: %v", err)
	}
	if res.Total != 1 || res.Truncated {
		t.Fatalf("total/truncated = %d/%v, want 1/false", res.Total, res.Truncated)
	}
	if len(res.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(res.Entries))
	}
	e := res.Entries[0]
	if e.Cloud != "aliyun" || e.Resource != "order-service" || e.LogType != "cdn" {
		t.Fatalf("entry = %+v, want cloud=aliyun resource=order-service", e)
	}
	if e.Timestamp.UnixMilli() != ts {
		t.Fatalf("timestamp = %d, want %d", e.Timestamp.UnixMilli(), ts)
	}
}

// TestQueryLogs_Timeout 单步超时 → 类型化 ModuleError（不 panic、不抛裸错误）。
func TestQueryLogs_Timeout(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"total":0}}`))
	}, Config{Timeout: 40 * time.Millisecond})

	_, err := c.QueryLogs(context.Background(), testLogRequest())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	var me *ModuleError
	if !errors.As(err, &me) {
		t.Fatalf("err = %T, want *ModuleError", err)
	}
	if me.Module != logqueryModule {
		t.Fatalf("module = %q, want logquery", me.Module)
	}
}

// TestQueryLogs_ContractDrift 响应结构漂移 → ContractDriftError + 漂移计数递增。
func TestQueryLogs_ContractDrift(t *testing.T) {
	var alerted string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// data 为字符串而非对象 → 结构漂移
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":"not-an-object"}`))
	}, Config{OnHealthAlert: func(m string) { alerted = m }})

	_, err := c.QueryLogs(context.Background(), testLogRequest())
	var drift *ContractDriftError
	if !errors.As(err, &drift) {
		t.Fatalf("err = %T, want *ContractDriftError", err)
	}
	if drift.Module != logqueryModule {
		t.Fatalf("drift module = %q, want logquery", drift.Module)
	}
	if c.drift.total(logqueryModule) != 1 {
		t.Fatalf("drift total = %d, want 1", c.drift.total(logqueryModule))
	}
	if alerted != "" {
		t.Fatalf("1 drift should not alert (got %q)", alerted)
	}
}

// TestQueryLogs_BusinessError 非 200 / code≠0 → ModuleError 携带底座语义。
func TestQueryLogs_BusinessError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"msg":"log_type 非法","data":null}`))
	}, Config{})

	_, err := c.QueryLogs(context.Background(), testLogRequest())
	var me *ModuleError
	if !errors.As(err, &me) {
		t.Fatalf("err = %T, want *ModuleError", err)
	}
	if me.Status != 400 {
		t.Fatalf("status = %d, want 400", me.Status)
	}
}

// TestDriftTracker_HealthAlert 10min 内同模块 ≥3 次 → 健康告警。
func TestDriftTracker_HealthAlert(t *testing.T) {
	tr := newDriftTracker()
	for i := 1; i < driftAlertThreshold; i++ {
		if total, alert := tr.record("logquery"); total != int64(i) || alert {
			t.Fatalf("record %d = total %d alert %v", i, total, alert)
		}
	}
	total, alert := tr.record("logquery")
	if total != driftAlertThreshold || !alert {
		t.Fatalf("3rd drift should alert (total=%d alert=%v)", total, alert)
	}
	if tr.total("logquery") != driftAlertThreshold {
		t.Fatalf("total = %d, want %d", tr.total("logquery"), driftAlertThreshold)
	}
	// 其他模块独立计数。
	if total, alert := tr.record("asset"); total != 1 || alert {
		t.Fatalf("asset drift = %d/%v, want 1/false", total, alert)
	}
}

// TestNotify_Success 通知发送成功。
func TestNotify_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()

	c := New(Config{ChannelWebhooks: map[string]string{"dingtalk": srv.URL}})
	res, err := c.Notify(context.Background(), "dingtalk", domain.Notification{
		Channel: "dingtalk", Title: "高危告警", Summary: "order-service 错误率上升", DiagLinkURL: "http://x/dx/1",
	})
	if err != nil {
		t.Fatalf("notify err: %v", err)
	}
	if !res.OK || res.Attempts != 1 {
		t.Fatalf("notify = %+v, want ok/attempts=1", res)
	}
}

// TestNotify_NotConfigured 渠道未配置 → 不阻断（OK=false + lastError）。
func TestNotify_NotConfigured(t *testing.T) {
	c := New(Config{})
	res, err := c.Notify(context.Background(), "dingtalk", domain.Notification{Channel: "dingtalk"})
	if err != nil {
		t.Fatalf("unconfigured notify should not return error: %v", err)
	}
	if res.OK || res.LastError == "" {
		t.Fatalf("notify = %+v, want ok=false with lastError", res)
	}
}

// TestNotify_RetriesUntilExhausted 非 200 连续失败 → 重试 3 次耗尽 → ok=false。
func TestNotify_RetriesExhausted(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(Config{ChannelWebhooks: map[string]string{"dingtalk": srv.URL}})
	res, err := c.Notify(context.Background(), "dingtalk", domain.Notification{Channel: "dingtalk", Title: "t", Summary: "s"})
	if err != nil {
		t.Fatalf("notify err: %v", err)
	}
	if res.OK {
		t.Fatalf("notify = %+v, want ok=false", res)
	}
	if attempts != notifyMaxAttempts {
		t.Fatalf("attempts = %d, want %d", attempts, notifyMaxAttempts)
	}
	if res.Attempts != notifyMaxAttempts {
		t.Fatalf("res.attempts = %d, want %d", res.Attempts, notifyMaxAttempts)
	}
}

// TestVerifyToken 委托 TokenVerifier；未注入 → 类型化错误。
func TestVerifyToken(t *testing.T) {
	c := New(Config{SessionProvider: tokenVerifierFunc(func(context.Context, string) (domain.Identity, error) {
		return domain.Identity{UserID: "42", Tenant: "2", Roles: []string{"admin"}}, nil
	})})
	id, err := c.VerifyToken(context.Background(), "tok")
	if err != nil || id.UserID != "42" || id.Tenant != "2" {
		t.Fatalf("verify = %+v err=%v", id, err)
	}

	c2 := New(Config{})
	if _, err := c2.VerifyToken(context.Background(), "tok"); err == nil {
		t.Fatal("unconfigured verifier should return error")
	}
}

// tokenVerifierFunc 函数适配器。
type tokenVerifierFunc func(context.Context, string) (domain.Identity, error)

func (f tokenVerifierFunc) Verify(ctx context.Context, token string) (domain.Identity, error) {
	return f(ctx, token)
}

// TestQueryAssets_NotConfigured 未配置 MCP 路径 → 类型化错误。
func TestQueryAssets_NotConfigured(t *testing.T) {
	c := New(Config{})
	if _, err := c.QueryAssets(context.Background(), "2", "list_instances", nil); err == nil {
		t.Fatal("unconfigured mcp should return error")
	}
}

// TestQueryAssets_StartError mcp-server 路径无效 → 启动失败的类型化错误。
func TestQueryAssets_StartError(t *testing.T) {
	c := New(Config{McpServerPath: "nonexistent-mcp-server.exe"})
	_, err := c.QueryAssets(context.Background(), "2", "list_instances", nil)
	var me *ModuleError
	if !errors.As(err, &me) || me.Module != assetModule {
		t.Fatalf("err = %v, want asset ModuleError", err)
	}
}

// TestDecodeMcpContent JSON-RPC 响应校验各分支：成功/错误/空/isError/非 text。
func TestDecodeMcpContent(t *testing.T) {
	text, err := decodeMcpContent(mcpResponse{Result: &mcpResult{
		Content: []mcpContent{{Type: "text", Text: `{"total":0}`}},
	}})
	if err != nil || text != `{"total":0}` {
		t.Fatalf("success = %q err=%v", text, err)
	}

	cases := []struct {
		name  string
		resp  mcpResponse
		isErr bool
	}{
		{"jsonrpc error", mcpResponse{Error: &struct {
			Message string `json:"message"`
		}{Message: "boom"}}, true},
		{"empty result", mcpResponse{}, true},
		{"isError tool", mcpResponse{Result: &mcpResult{
			IsError: true, Content: []mcpContent{{Type: "text", Text: "错误: x"}},
		}}, true},
		{"non-text content", mcpResponse{Result: &mcpResult{
			Content: []mcpContent{{Type: "image", Text: "base64"}},
		}}, true},
	}
	for _, c := range cases {
		if _, err := decodeMcpContent(c.resp); (err != nil) != c.isErr {
			t.Fatalf("%s: err = %v, want isErr=%v", c.name, err, c.isErr)
		}
	}
}

// TestParseAssetText B2 资产 text JSON 解析：items → []AssetRecord。
func TestParseAssetText(t *testing.T) {
	c := New(Config{})
	recs, err := c.parseAssetText("2", "ecs", `{"total":2,"items":[{"id":1,"asset_id":"","asset_name":"web-1","region":"cn-hangzhou","status":"running"},{"asset_id":"i-abc","asset_name":"db-1"}]}`)
	if err != nil {
		t.Fatalf("parse asset text: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("recs = %d, want 2", len(recs))
	}
	if recs[0].AssetID != "1" || recs[0].Type != "ecs" || recs[0].Tenant != "2" || recs[0].Name != "web-1" {
		t.Fatalf("rec[0] = %+v", recs[0])
	}
	if recs[1].AssetID != "i-abc" {
		t.Fatalf("rec[1].assetID = %q, want i-abc", recs[1].AssetID)
	}
}

// TestParseAssetText_Drift text JSON 非预期形状（items 为对象而非数组）→ ContractDriftError。
func TestParseAssetText_Drift(t *testing.T) {
	c := New(Config{})
	_, err := c.parseAssetText("2", "ecs", `{"total":0,"items":{}}`)
	var drift *ContractDriftError
	if !errors.As(err, &drift) {
		t.Fatalf("err = %T, want *ContractDriftError", err)
	}
	if c.drift.total(assetModule) != 1 {
		t.Fatalf("drift total = %d, want 1", c.drift.total(assetModule))
	}
}

// TestQueryLogs_Retry 5xx 后重试成功（客户端内闭环重试）。
func TestQueryLogs_Retry(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"log_type":"cdn","total":0,"truncated":false,"entries":[]}}`))
	}, Config{Retries: 2})

	res, err := c.QueryLogs(context.Background(), testLogRequest())
	if err != nil {
		t.Fatalf("retry QueryLogs: %v", err)
	}
	if res.Total != 0 || calls != 2 {
		t.Fatalf("total=%d calls=%d, want 0/2", res.Total, calls)
	}
}

// TestQueryLogs_EnvelopeDrift 信封 JSON 整体非法 → ContractDriftError。
func TestQueryLogs_EnvelopeDrift(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok",`)) // 截断的非法 JSON
	}, Config{})

	_, err := c.QueryLogs(context.Background(), testLogRequest())
	var drift *ContractDriftError
	if !errors.As(err, &drift) {
		t.Fatalf("err = %T, want *ContractDriftError", err)
	}
	if c.drift.total(logqueryModule) != 1 {
		t.Fatalf("drift total = %d, want 1", c.drift.total(logqueryModule))
	}
}

// TestQueryLogs_MessageEnvelope 兼容 message 键位（非 msg）的业务错误。
func TestQueryLogs_MessageEnvelope(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":400,"message":"当前会话未选定租户空间","data":null}`))
	}, Config{})

	_, err := c.QueryLogs(context.Background(), testLogRequest())
	var me *ModuleError
	if !errors.As(err, &me) || me.Msg != "当前会话未选定租户空间" {
		t.Fatalf("err = %v, want message-envelope error", err)
	}
}

// TestBuildNotifyBody 渠道载荷组装（dingtalk/wecom/feishu 分支）。
func TestBuildNotifyBody(t *testing.T) {
	p := domain.Notification{Channel: "dingtalk", Title: "T", Summary: "S", DiagLinkURL: "http://x"}
	ding := buildNotifyBody(p).(map[string]any)
	if ding["msgtype"] != "markdown" {
		t.Fatalf("dingtalk = %+v", ding)
	}
	wecom := buildNotifyBody(domain.Notification{Channel: "wecom", Title: "T", Summary: "S"}).(map[string]any)
	if md, _ := wecom["markdown"].(map[string]any); md["content"] == "" {
		t.Fatalf("wecom markdown = %+v", wecom)
	}
	feishu := buildNotifyBody(domain.Notification{Channel: "feishu", Title: "T", Summary: "S"}).(map[string]any)
	if feishu["msg_type"] != "text" {
		t.Fatalf("feishu = %+v", feishu)
	}
}

// TestQueryLogs_TokenForwarding 上下文挂载用户令牌时，底座 HTTP 调用透传
// Authorization: Bearer <token>（契约 B1/B3：底座按 eiam 会话鉴权）。
func TestQueryLogs_TokenForwarding(t *testing.T) {
	const wantToken = "eiam-signed-token-xyz"
	seen := make(chan string, 1)
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"total":0,"truncated":false,"entries":[]}}`))
	}, Config{})

	ctx := WithUserToken(context.Background(), wantToken)
	if _, err := c.QueryLogs(ctx, testLogRequest()); err != nil {
		t.Fatalf("QueryLogs: %v", err)
	}
	if got := <-seen; got != "Bearer "+wantToken {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer "+wantToken)
	}
}

// TestQueryLogs_NoTokenForwarding 未挂载令牌（告警 worker 链路）时不透传该头；
// 底座 401 由调用方按契约降级「无数据」，客户端不伪造凭证。
func TestQueryLogs_NoTokenForwarding(t *testing.T) {
	seen := make(chan string, 1)
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"total":0,"truncated":false,"entries":[]}}`))
	}, Config{})

	if _, err := c.QueryLogs(context.Background(), testLogRequest()); err != nil {
		t.Fatalf("QueryLogs: %v", err)
	}
	if got := <-seen; got != "" {
		t.Fatalf("Authorization = %q, want empty", got)
	}
}

// TestIsRetryable 重试判定：5xx 可重试、业务错误/漂移不重试。
func TestIsRetryable(t *testing.T) {
	if !isRetryable(&ModuleError{Status: 500}) {
		t.Fatal("500 should be retryable")
	}
	if isRetryable(&ModuleError{Status: 400}) {
		t.Fatal("400 should not be retryable")
	}
	if isRetryable(&ContractDriftError{}) {
		t.Fatal("drift should not be retryable")
	}
	if !isRetryable(context.DeadlineExceeded) {
		t.Fatal("deadline should be retryable")
	}
}
