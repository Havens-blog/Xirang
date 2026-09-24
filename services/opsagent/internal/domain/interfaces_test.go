package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// ---- tech-design Interfaces：编译期签名断言（存根实现必须与定义完全匹配） ----

type stubIntentClassifier struct{}

func (stubIntentClassifier) Classify(ctx context.Context, input string, tenant string) (domain.IntentResult, error) {
	return domain.IntentResult{Type: domain.IntentDiagnose, Confidence: 1, Params: domain.DiagnoseParams{ServiceName: input}}, nil
}

func (stubIntentClassifier) Correct(ctx context.Context, sessionID string, target domain.IntentType, params domain.DiagnoseParams, message string) (domain.IntentResult, error) {
	return domain.IntentResult{Type: target}, nil
}

type stubBaseModuleClient struct{}

func (stubBaseModuleClient) QueryLogs(ctx context.Context, req domain.LogQueryRequest) (domain.LogQueryResult, error) {
	return domain.LogQueryResult{}, nil
}

func (stubBaseModuleClient) QueryAssets(ctx context.Context, tenant string, tool string, args map[string]any) ([]domain.AssetRecord, error) {
	return nil, nil
}

func (stubBaseModuleClient) VerifyToken(ctx context.Context, token string) (domain.Identity, error) {
	return domain.Identity{}, nil
}

func (stubBaseModuleClient) Notify(ctx context.Context, channel string, payload domain.Notification) (domain.NotifyResult, error) {
	return domain.NotifyResult{Channel: channel}, nil
}

type stubOrchestrator struct{}

func (stubOrchestrator) RunChat(ctx context.Context, session *domain.Session, input string) (*domain.Diagnosis, error) {
	return &domain.Diagnosis{SessionID: session.ID}, nil
}

func (stubOrchestrator) RunAlert(ctx context.Context, alert domain.AlertEvent) (*domain.RiskEntry, error) {
	return &domain.RiskEntry{Fingerprint: domain.Fingerprint(alert.Tenant, alert.ServiceName, alert.Metric, alert.Level)}, nil
}

type stubLLMProvider struct{}

func (stubLLMProvider) Complete(ctx context.Context, req domain.LLMRequest) (string, error) {
	return "ok", nil
}

type stubRiskRegistry struct{}

func (stubRiskRegistry) Classify(tool string) (domain.ToolMeta, error) {
	if tool == "unknown" {
		return domain.ToolMeta{}, errors.New("unregistered tool")
	}
	return domain.ToolMeta{Name: tool, RiskLevel: domain.RiskRead, Allowlist: true}, nil
}

func TestInterfaceSignatures(t *testing.T) {
	// 编译期断言：签名不一致即无法通过编译
	var _ domain.IntentClassifier = stubIntentClassifier{}
	var _ domain.BaseModuleClient = stubBaseModuleClient{}
	var _ domain.Orchestrator = stubOrchestrator{}
	var _ domain.LLMProvider = stubLLMProvider{}
	var _ domain.RiskRegistry = stubRiskRegistry{}

	// 运行期冒烟：确认类型可正常流通
	ctx := context.Background()
	res, err := stubIntentClassifier{}.Classify(ctx, "order-service 报错", "t1")
	if err != nil || res.Type != domain.IntentDiagnose {
		t.Fatalf("Classify = %+v, %v", res, err)
	}
	dx, err := stubOrchestrator{}.RunChat(ctx, &domain.Session{ID: "s1"}, "q")
	if err != nil || dx.SessionID != "s1" {
		t.Fatalf("RunChat = %+v, %v", dx, err)
	}
	txt, err := stubLLMProvider{}.Complete(ctx, domain.LLMRequest{Timeout: 12 * time.Second})
	if err != nil || txt != "ok" {
		t.Fatalf("Complete = %q, %v", txt, err)
	}
	if _, err := (stubRiskRegistry{}).Classify("unknown"); err == nil {
		t.Error("Classify(unknown) should error for unregistered tool")
	}
}
