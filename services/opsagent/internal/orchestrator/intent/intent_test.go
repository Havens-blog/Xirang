package intent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// fakeLLM 测试用 LLM Provider。
type fakeLLM struct {
	fn func(ctx context.Context, req domain.LLMRequest) (string, error)
}

func (f fakeLLM) Complete(ctx context.Context, req domain.LLMRequest) (string, error) {
	return f.fn(ctx, req)
}

// fakeStore 测试用 SessionStore。
type fakeStore struct {
	sess        domain.Session
	corrections []domain.CorrectionRecord
}

func (f *fakeStore) GetByID(context.Context, string, string) (domain.Session, error) {
	return f.sess, nil
}

func (f *fakeStore) AppendCorrection(_ context.Context, _, _ string, rec domain.CorrectionRecord) error {
	f.corrections = append(f.corrections, rec)
	return nil
}

// TestClassify_Diagnose 排障意图：抽取服务名/指标/时间窗，高置信度不澄清。
func TestClassify_Diagnose(t *testing.T) {
	c := New(nil, nil, nil)
	res, err := c.Classify(context.Background(), "order-service 最近一小时错误率上升，帮我看看", "t")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if res.Type != domain.IntentDiagnose {
		t.Fatalf("type = %q, want diagnose", res.Type)
	}
	if res.Params.ServiceName != "order-service" {
		t.Fatalf("service = %q, want order-service", res.Params.ServiceName)
	}
	if res.Params.Metric == "" {
		t.Fatal("metric should be extracted")
	}
	if res.NeedsClarify {
		t.Fatal("high-confidence diagnose should not need clarify")
	}
	if res.Params.Timeframe.StartTime.IsZero() || res.Params.Timeframe.EndTime.IsZero() {
		t.Fatal("timeframe should be filled")
	}
}

// TestClassify_Resource 资源查询意图。
func TestClassify_Resource(t *testing.T) {
	c := New(nil, nil, nil)
	res, _ := c.Classify(context.Background(), "帮我查一下有多少台 ecs 资产", "t")
	if res.Type != domain.IntentResource {
		t.Fatalf("type = %q, want resource", res.Type)
	}
}

// TestClassify_OutOfScope 超能力意图（不调用底座，上层返回引导式回应）。
func TestClassify_OutOfScope(t *testing.T) {
	c := New(nil, nil, nil)
	for _, in := range []string{"帮我写个脚本", "怎么开发一个新功能", "工单交到哪里"} {
		res, _ := c.Classify(context.Background(), in, "t")
		if res.Type != domain.IntentOutOfScope {
			t.Fatalf("%q type = %q, want out_of_scope", in, res.Type)
		}
	}
}

// TestClassify_NeedsClarify 诊断意图抽取不出服务名 → NeedsClarify + 备选意图。
func TestClassify_NeedsClarify(t *testing.T) {
	c := New(nil, nil, nil)
	res, _ := c.Classify(context.Background(), "最近错误率有点高", "t")
	if res.Type != domain.IntentDiagnose {
		t.Fatalf("type = %q, want diagnose", res.Type)
	}
	if !res.NeedsClarify {
		t.Fatal("diagnose without service name should need clarify")
	}
	if len(res.Alternatives) == 0 {
		t.Fatal("needs-clarify should have candidates")
	}
}

// TestClassify_LLMFallback LLM 不可用（失败）时回退规则匹配。
func TestClassify_LLMFallback(t *testing.T) {
	c := New(fakeLLM{fn: func(context.Context, domain.LLMRequest) (string, error) {
		return "", errors.New("boom")
	}}, nil, nil)
	res, err := c.Classify(context.Background(), "order-service 错误率", "t")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if res.Type != domain.IntentDiagnose || res.Params.ServiceName != "order-service" {
		t.Fatalf("rule fallback = %+v, want diagnose/order-service", res)
	}
}

// TestClassify_LLMPath LLM 可用时用语义分类结果。
func TestClassify_LLMPath(t *testing.T) {
	c := New(fakeLLM{fn: func(context.Context, domain.LLMRequest) (string, error) {
		return `{"type":"diagnose","confidence":0.95,"serviceName":"pay-service","metric":"延迟","timeWindowMinutes":30}`, nil
	}}, nil, nil)
	res, err := c.Classify(context.Background(), "任意输入", "t")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if res.Params.ServiceName != "pay-service" || res.Params.Metric != "延迟" {
		t.Fatalf("llm result = %+v, want pay-service/延迟", res)
	}
}

// TestCorrect_XOR target 与 message 互斥。
func TestCorrect_XOR(t *testing.T) {
	c := New(nil, nil, nil)
	ctx := context.Background()
	if _, err := c.Correct(ctx, "t1", "s1", "", domain.DiagnoseParams{}, ""); err == nil {
		t.Fatal("both empty should error (XOR)")
	}
	if _, err := c.Correct(ctx, "t1", "s1", domain.IntentDiagnose, domain.DiagnoseParams{}, "自然语言"); err == nil {
		t.Fatal("both set should error (XOR)")
	}
}

// TestCorrect_Target 一键纠正：target+params 直接重路由。
func TestCorrect_Target(t *testing.T) {
	c := New(nil, nil, nil)
	res, err := c.Correct(context.Background(), "t1", "s1", domain.IntentResource, domain.DiagnoseParams{ServiceName: "rds-1"}, "")
	if err != nil {
		t.Fatalf("correct: %v", err)
	}
	if res.Type != domain.IntentResource || res.Confidence != 1.0 {
		t.Fatalf("correct = %+v, want resource/1.0", res)
	}
}

// TestCorrect_Message 自然语言纠正：message 重新分类。
func TestCorrect_Message(t *testing.T) {
	c := New(nil, nil, nil)
	res, err := c.Correct(context.Background(), "t1", "s1", "", domain.DiagnoseParams{}, "不是排障，帮我查资产有多少 ecs")
	if err != nil {
		t.Fatalf("correct: %v", err)
	}
	if res.Type != domain.IntentResource {
		t.Fatalf("re-classified = %q, want resource", res.Type)
	}
}

// TestCorrect_AppendCorrection 纠正记录追加（store 注入时）。
func TestCorrect_AppendCorrection(t *testing.T) {
	store := &fakeStore{sess: domain.Session{IntentType: domain.IntentDiagnose}}
	c := New(nil, store, nil)
	ctx := domain.WithUserAttr(context.Background(), "admin")
	if _, err := c.Correct(ctx, "t1", "s1", domain.IntentResource, domain.DiagnoseParams{}, ""); err != nil {
		t.Fatalf("correct: %v", err)
	}
	if len(store.corrections) != 1 {
		t.Fatalf("corrections = %d, want 1", len(store.corrections))
	}
	if store.corrections[0].From != domain.IntentDiagnose || store.corrections[0].To != domain.IntentResource {
		t.Fatalf("correction = %+v, want diagnose→resource", store.corrections[0])
	}
	if store.corrections[0].By != "admin" {
		t.Fatalf("correction By = %q, want admin（来自请求上下文用户属性）", store.corrections[0].By)
	}
}

// TestCorrect_InvalidTarget 非法 target 拒绝。
func TestCorrect_InvalidTarget(t *testing.T) {
	c := New(nil, nil, nil)
	if _, err := c.Correct(context.Background(), "t1", "s1", domain.IntentType("bogus"), domain.DiagnoseParams{}, ""); err == nil {
		t.Fatal("invalid target should error")
	}
}

// TestGuided 引导话术查询（Flow C）。
func TestGuided(t *testing.T) {
	c := New(nil, nil, map[string]GuidedTemplate{
		"开发咨询": {Template: "这是开发咨询类问题，请走 X 渠道", Channel: "研发工单系统"},
	})
	if g, ok := c.Guided("开发咨询"); !ok || g.Channel == "" {
		t.Fatalf("guided = %+v ok=%v", g, ok)
	}
	if _, ok := c.Guided("未知类别"); ok {
		t.Fatal("unknown category should not be found")
	}
}

// TestRuleHelpers 规则抽取原子函数。
func TestRuleHelpers(t *testing.T) {
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	if got := extractServiceName("cdn-edge 服务流量突增"); got != "cdn-edge" {
		t.Fatalf("service = %q, want cdn-edge", got)
	}
	if got := extractServiceName("帮我看看"); got != "" {
		t.Fatalf("service = %q, want empty", got)
	}
	if got := extractServiceName("订单服务 故障"); got != "订单" {
		t.Fatalf("zh service = %q, want 订单", got)
	}
	if got := extractMetric("order-service 延迟升高"); got == "" {
		t.Fatal("metric should be extracted")
	}
	start, end, hit := extractTimeframe("最近 30 分钟", now)
	if !hit || end.Sub(start) != 30*time.Minute {
		t.Fatalf("timeframe = %v~%v hit=%v, want 30min", start, end, hit)
	}
	if !detectOutOfScope("帮我开发个接口") || detectOutOfScope("order-service 错误率") {
		t.Fatal("out-of-scope detection wrong")
	}
}

// TestExtractJSON 容忍 markdown 代码围栏。
func TestExtractJSON(t *testing.T) {
	if got := extractJSON("```json\n{\"type\":\"diagnose\"}\n```"); got != `{"type":"diagnose"}` {
		t.Fatalf("extract = %q", got)
	}
}
