package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/config"
	"github.com/Havens-blog/opsagent/internal/domain"
)

// fakeProvider 测试用可控 provider。
type fakeProvider struct {
	fn func(ctx context.Context, req domain.LLMRequest) (string, error)
}

func (f *fakeProvider) Complete(ctx context.Context, req domain.LLMRequest) (string, error) {
	return f.fn(ctx, req)
}

func testRequest() domain.LLMRequest {
	return domain.LLMRequest{Messages: []domain.LLMMessage{{Role: "user", Text: "hi"}}}
}

// TestGuardedLLM_Disabled 未配置 → IsAvailable=false，CompleteGuarded 返回空串不 panic。
func TestGuardedLLM_Disabled(t *testing.T) {
	ctx := context.Background()
	for _, p := range []domain.LLMProvider{disabledProvider{}, nil} {
		g := NewGuardedLLM(p, time.Second)
		if g.IsAvailable(ctx) {
			t.Fatalf("disabled provider should be unavailable")
		}
		if text, avail := g.CompleteGuarded(ctx, testRequest()); text != "" || avail {
			t.Fatalf("disabled complete = (%q,%v), want (\"\",false)", text, avail)
		}
	}
}

// TestGuardedLLM_ConsecutiveFailures 连续 2 次失败 → IsAvailable=false；1 次失败仍可用。
func TestGuardedLLM_ConsecutiveFailures(t *testing.T) {
	ctx := context.Background()
	g := NewGuardedLLM(&fakeProvider{fn: func(context.Context, domain.LLMRequest) (string, error) {
		return "", errors.New("boom")
	}}, time.Second)

	if _, avail := g.CompleteGuarded(ctx, testRequest()); avail {
		t.Fatal("1st failure should be unavailable result")
	}
	if !g.IsAvailable(ctx) {
		t.Fatal("after 1 non-timeout failure should still be available")
	}
	if _, avail := g.CompleteGuarded(ctx, testRequest()); avail {
		t.Fatal("2nd failure should be unavailable result")
	}
	if g.IsAvailable(ctx) {
		t.Fatal("after 2 consecutive failures should be unavailable")
	}
}

// TestGuardedLLM_SingleTimeout 单次超时即不可用（置为阈值）。
func TestGuardedLLM_SingleTimeout(t *testing.T) {
	ctx := context.Background()
	g := NewGuardedLLM(&fakeProvider{fn: func(context.Context, domain.LLMRequest) (string, error) {
		return "", context.DeadlineExceeded
	}}, time.Second)

	if text, avail := g.CompleteGuarded(ctx, testRequest()); text != "" || avail {
		t.Fatalf("timeout complete = (%q,%v), want (\"\",false)", text, avail)
	}
	if g.IsAvailable(ctx) {
		t.Fatal("single timeout should mark unavailable")
	}
}

// TestGuardedLLM_Recovery 失败后成功即恢复（失败计数清零）。
func TestGuardedLLM_Recovery(t *testing.T) {
	ctx := context.Background()
	calls := 0
	g := NewGuardedLLM(&fakeProvider{fn: func(context.Context, domain.LLMRequest) (string, error) {
		calls++
		if calls <= 2 {
			return "", errors.New("boom")
		}
		return "ok", nil
	}}, time.Second)

	g.CompleteGuarded(ctx, testRequest())
	g.CompleteGuarded(ctx, testRequest())
	if g.IsAvailable(ctx) {
		t.Fatal("after 2 failures should be unavailable")
	}
	if text, avail := g.CompleteGuarded(ctx, testRequest()); text != "ok" || !avail {
		t.Fatalf("recovery complete = (%q,%v), want (\"ok\",true)", text, avail)
	}
	if !g.IsAvailable(ctx) {
		t.Fatal("success should recover availability")
	}
}

// TestGuardedLLM_TimeoutBounded 单次调用 context 超时生效，不阻塞（慢 provider 在超时后返回）。
func TestGuardedLLM_TimeoutBounded(t *testing.T) {
	ctx := context.Background()
	block := make(chan bool, 1)
	g := NewGuardedLLM(&fakeProvider{fn: func(c context.Context, _ domain.LLMRequest) (string, error) {
		<-c.Done()
		block <- true
		return "", c.Err()
	}}, 40*time.Millisecond)

	start := time.Now()
	if _, avail := g.CompleteGuarded(ctx, testRequest()); avail {
		t.Fatal("blocking provider should time out")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("complete blocked for %v, want bounded by timeout", elapsed)
	}
	select {
	case <-block:
	default:
		t.Fatal("provider goroutine should have observed ctx cancellation")
	}
	if g.IsAvailable(ctx) {
		t.Fatal("timeout should mark unavailable")
	}
}

// TestNewProvider_Dispatch 未配置 / 未知 provider → disabled；deepseek 命中默认 baseURL。
func TestNewProvider_Dispatch(t *testing.T) {
	if p := NewProvider(config.LLMConfig{Provider: "qwen"}); !isDisabled(p) {
		t.Fatalf("no apiKey should be disabled, got %T", p)
	}
	if p := NewProvider(config.LLMConfig{Provider: "unknown", APIKey: "sk-x"}); !isDisabled(p) {
		t.Fatalf("unknown provider should be disabled, got %T", p)
	}

	p := NewProvider(config.LLMConfig{Provider: "deepseek", APIKey: "sk-test", Model: "deepseek-chat"})
	hp, ok := p.(*httpProvider)
	if !ok {
		t.Fatalf("deepseek provider = %T, want *httpProvider", p)
	}
	if hp.baseURL != deepseekBaseURL || hp.model != "deepseek-chat" || hp.apiKey != "sk-test" {
		t.Fatalf("deepseek provider config mismatch: %+v", hp)
	}
}

// isDisabled 是否 disabledProvider（值或指针）。
func isDisabled(p domain.LLMProvider) bool {
	switch p.(type) {
	case disabledProvider, *disabledProvider:
		return true
	default:
		return false
	}
}

// TestHTTPProvider_Complete OpenAI-compatible 请求/响应闭环（httptest 验证）。
func TestHTTPProvider_Complete(t *testing.T) {
	var gotAuth, gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body apiRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		gotModel = body.Model
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"你好"}}]}`))
	}))
	defer srv.Close()

	p := newHTTPProvider(srv.URL, "sk-test", "test-model")
	text, err := p.Complete(context.Background(), domain.LLMRequest{
		Messages: []domain.LLMMessage{{Role: "user", Text: "hi"}},
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if text != "你好" {
		t.Fatalf("text = %q, want 你好", text)
	}
	if gotAuth != "Bearer sk-test" || gotModel != "test-model" {
		t.Fatalf("request auth=%q model=%q, want Bearer sk-test / test-model", gotAuth, gotModel)
	}
}

// TestHTTPProvider_Non200 非 200 返回错误。
func TestHTTPProvider_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer srv.Close()

	p := newHTTPProvider(srv.URL, "sk-bad", "m")
	if _, err := p.Complete(context.Background(), testRequest()); err == nil {
		t.Fatal("non-200 should return error")
	}
}

// TestHTTPProvider_EmptyChoices 空 choices 返回错误。
func TestHTTPProvider_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	p := newHTTPProvider(srv.URL, "k", "m")
	if _, err := p.Complete(context.Background(), testRequest()); err == nil {
		t.Fatal("empty choices should return error")
	}
}

// TestProviderConstructors 三个命名构造函数命中各自默认 baseURL。
func TestProviderConstructors(t *testing.T) {
	cases := []struct {
		name  string
		prov  domain.LLMProvider
		base  string
		model string
	}{
		{"qwen", NewQwenProvider("k", "qwen-max"), qwenBaseURL, "qwen-max"},
		{"deepseek", NewDeepseekProvider("k", "deepseek-chat"), deepseekBaseURL, "deepseek-chat"},
		{"openai", NewOpenaiProvider("k", "gpt-4o"), openaiBaseURL, "gpt-4o"},
	}
	for _, c := range cases {
		hp, ok := c.prov.(*httpProvider)
		if !ok {
			t.Fatalf("%s provider = %T, want *httpProvider", c.name, c.prov)
		}
		if hp.baseURL != c.base || hp.model != c.model {
			t.Fatalf("%s config = %q/%q, want %q/%q", c.name, hp.baseURL, hp.model, c.base, c.model)
		}
	}
}

// TestTruncateBody 截断长响应正文。
func TestTruncateBody(t *testing.T) {
	long := make([]byte, 500)
	for i := range long {
		long[i] = 'x'
	}
	if got := truncateBody(long); len(got) != 200 {
		t.Fatalf("truncate len = %d, want 200", len(got))
	}
	if got := truncateBody([]byte("short")); got != "short" {
		t.Fatalf("short = %q, want short", got)
	}
}
