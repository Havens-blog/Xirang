package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/llm"
	"github.com/Havens-blog/opsagent/internal/repository/dao"
)

// ---- fakes ----

type fakeIntent struct {
	result domain.IntentResult
	err    error
}

func (f *fakeIntent) Classify(context.Context, string, string) (domain.IntentResult, error) {
	return f.result, f.err
}
func (f *fakeIntent) Correct(context.Context, string, string, domain.IntentType, domain.DiagnoseParams, string) (domain.IntentResult, error) {
	return domain.IntentResult{}, nil
}

type fakeBase struct {
	logs     domain.LogQueryResult
	logErr   error
	assets   []domain.AssetRecord
	assetErr error
	notified []domain.Notification
}

func (f *fakeBase) QueryLogs(context.Context, domain.LogQueryRequest) (domain.LogQueryResult, error) {
	return f.logs, f.logErr
}
func (f *fakeBase) QueryAssets(context.Context, string, string, map[string]any) ([]domain.AssetRecord, error) {
	return f.assets, f.assetErr
}
func (f *fakeBase) VerifyToken(context.Context, string) (domain.Identity, error) {
	return domain.Identity{}, nil
}
func (f *fakeBase) Notify(_ context.Context, _ string, p domain.Notification) (domain.NotifyResult, error) {
	f.notified = append(f.notified, p)
	return domain.NotifyResult{OK: true}, nil
}

type fakeLLMProvider struct {
	fn func(ctx context.Context, req domain.LLMRequest) (string, error)
}

func (f fakeLLMProvider) Complete(ctx context.Context, req domain.LLMRequest) (string, error) {
	return f.fn(ctx, req)
}

type fakeDiagnoses struct {
	inserted *dao.DiagnosisRecord
}

func (f *fakeDiagnoses) Insert(_ context.Context, rec *dao.DiagnosisRecord) error {
	f.inserted = rec
	rec.Diagnosis.ID = "dx-1"
	return nil
}

type fakeCitations struct{ inserted []domain.Citation }

func (f *fakeCitations) InsertBatch(_ context.Context, _, _ string, citations []domain.Citation) error {
	f.inserted = citations
	return nil
}

type fakeSessions struct{ doneStatus domain.SessionStatus }

func (f *fakeSessions) UpdateStatus(_ context.Context, _, _ string, status domain.SessionStatus) error {
	f.doneStatus = status
	return nil
}

type fakeRiskEntries struct{ inserted *domain.RiskEntry }

func (f *fakeRiskEntries) Insert(_ context.Context, e *domain.RiskEntry) error {
	f.inserted = e
	e.ID = "re-1"
	return nil
}

type fakeAlerts struct {
	dedup  domain.DedupConclusion
	linked string
}

func (f *fakeAlerts) RecordArrival(context.Context, string, string, map[string]any, time.Time) (domain.DedupConclusion, int64, error) {
	return f.dedup, 1, nil
}
func (f *fakeAlerts) LinkDiagnosis(_ context.Context, _, _ string, _ time.Time, diagID string) error {
	f.linked = diagID
	return nil
}

func testNow() time.Time { return time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC) }

func testParams() domain.DiagnoseParams {
	return domain.DiagnoseParams{
		ServiceName: "order-service",
		Metric:      "错误率",
		Timeframe:   domain.Timeframe{StartTime: testNow().Add(-time.Hour), EndTime: testNow()},
	}
}

func endToEndDeps(b *fakeBase, llmG *llm.GuardedLLM, d *fakeDiagnoses, c *fakeCitations, s *fakeSessions, r *fakeRiskEntries, a *fakeAlerts) Deps {
	return Deps{
		Intent:      &fakeIntent{result: domain.IntentResult{Type: domain.IntentDiagnose, Confidence: 0.9, Params: testParams()}},
		Base:        b,
		LLM:         llmG,
		Sessions:    s,
		Diagnoses:   d,
		Citations:   c,
		RiskEntries: r,
		Alerts:      a,
	}
}

// ---- tests ----

// TestRunChat_EndToEnd 端到端：意图 → 编排 → 报告组装 → 落库；TraceStep 覆盖 5 步。
func TestRunChat_EndToEnd(t *testing.T) {
	d := &fakeDiagnoses{}
	c := &fakeCitations{}
	s := &fakeSessions{}
	deps := endToEndDeps(&fakeBase{logs: domain.LogQueryResult{Total: 3, Entries: []domain.LogEntry{{Cloud: "aliyun", Resource: "order-service"}}}},
		nil, d, c, s, nil, nil)

	o := New(deps)
	diag, err := o.RunChat(context.Background(), &domain.Session{TenantID: "t1", ID: "s1"}, "order-service 错误率")
	if err != nil {
		t.Fatalf("RunChat: %v", err)
	}
	if diag.RootCause == "" {
		t.Fatal("root cause should be populated")
	}
	if len(diag.Citations) == 0 {
		t.Fatal("citations should back to sources")
	}
	if len(diag.Trace) != 5 {
		t.Fatalf("trace steps = %d, want 5", len(diag.Trace))
	}
	// 落库：诊断 + 引用 + 会话 done。
	if d.inserted == nil || d.inserted.Diagnosis.ID != "dx-1" {
		t.Fatal("diagnosis should be persisted")
	}
	if len(c.inserted) == 0 {
		t.Fatal("citations should be persisted")
	}
	if s.doneStatus != domain.SessionStatusDone {
		t.Fatalf("session status = %q, want done", s.doneStatus)
	}
}

// TestRunChat_OutOfScope 超能力意图短路，返回引导错误不调用底座。
func TestRunChat_OutOfScope(t *testing.T) {
	deps := Deps{
		Intent: &fakeIntent{result: domain.IntentResult{Type: domain.IntentOutOfScope, Confidence: 1}},
		Base:   &fakeBase{},
	}
	o := New(deps)
	_, err := o.RunChat(context.Background(), &domain.Session{TenantID: "t1"}, "帮我写脚本")
	var oos *OutOfScopeError
	if !errors.As(err, &oos) {
		t.Fatalf("err = %v, want OutOfScopeError", err)
	}
}

// TestRunChat_Clarify 澄清意图短路。
func TestRunChat_Clarify(t *testing.T) {
	deps := Deps{
		Intent: &fakeIntent{result: domain.IntentResult{Type: domain.IntentDiagnose, NeedsClarify: true, Confidence: 0.4}},
		Base:   &fakeBase{},
	}
	o := New(deps)
	_, err := o.RunChat(context.Background(), &domain.Session{TenantID: "t1"}, "最近错误率有点高")
	var ce *ClarifyError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want ClarifyError", err)
	}
}

// TestRunChat_EmptyLogs logquery 空结果 → 报告明示「无匹配日志」，不编造。
func TestRunChat_EmptyLogs(t *testing.T) {
	deps := endToEndDeps(&fakeBase{logs: domain.LogQueryResult{Total: 0}}, nil,
		&fakeDiagnoses{}, &fakeCitations{}, &fakeSessions{}, nil, nil)
	o := New(deps)
	diag, err := o.RunChat(context.Background(), &domain.Session{TenantID: "t1", ID: "s1"}, "order-service 错误率")
	if err != nil {
		t.Fatalf("RunChat: %v", err)
	}
	if diag.RootCause != "时间窗内无匹配日志（服务=order-service），无法定位到可用证据" {
		t.Fatalf("root cause = %q, want 无匹配日志", diag.RootCause)
	}
}

// TestRunChat_StepFailure 单步查询失败 → 该步「未取得数据」并继续，报告仍返回。
func TestRunChat_StepFailure(t *testing.T) {
	deps := endToEndDeps(&fakeBase{logErr: errors.New("timeout")}, nil,
		&fakeDiagnoses{}, &fakeCitations{}, &fakeSessions{}, nil, nil)
	o := New(deps)
	diag, err := o.RunChat(context.Background(), &domain.Session{TenantID: "t1", ID: "s1"}, "order-service 错误率")
	if err != nil {
		t.Fatalf("RunChat should not fail on step error: %v", err)
	}
	// 找出 log_analyst 步骤，应标注「未取得数据」。
	found := false
	for _, st := range diag.Trace {
		if st.Agent == agentLogAnalyst && st.Summary == "未取得数据" {
			found = true
		}
	}
	if !found {
		t.Fatalf("log_analyst step should be degraded no-data, trace = %+v", diag.Trace)
	}
}

// TestRunChat_LLMReport LLM 可用时用 LLM 解读结论。
func TestRunChat_LLMReport(t *testing.T) {
	g := llm.NewGuardedLLM(fakeLLMProvider{fn: func(context.Context, domain.LLMRequest) (string, error) {
		return `{"rootCause":"数据库连接池耗尽","confidence":0.85,"conclusions":[{"text":"连接池打满","citations":["order-service"]}],"disposition":[{"step":"扩容连接池","risk":"low","action":"扩到300"}]}`, nil
	}}, time.Second)
	deps := endToEndDeps(&fakeBase{logs: domain.LogQueryResult{Total: 1, Entries: []domain.LogEntry{{Cloud: "aliyun", Resource: "order-service"}}}}, g,
		&fakeDiagnoses{}, &fakeCitations{}, &fakeSessions{}, nil, nil)
	o := New(deps)
	diag, err := o.RunChat(context.Background(), &domain.Session{TenantID: "t1", ID: "s1"}, "order-service 错误率")
	if err != nil {
		t.Fatalf("RunChat: %v", err)
	}
	if diag.RootCause != "数据库连接池耗尽" {
		t.Fatalf("root cause = %q, want LLM 解读结论", diag.RootCause)
	}
	if diag.Degraded {
		t.Fatal("LLM available should not be degraded")
	}
	if len(diag.Disposition) != 1 || diag.Disposition[0].Executed {
		t.Fatalf("disposition = %+v, want 1 read-only step", diag.Disposition)
	}
}

// TestRunChat_LLMOutOfBoundCitation LLM 引用越界 SourceKey → 整段丢弃并降级规则结论。
func TestRunChat_LLMOutOfBoundCitation(t *testing.T) {
	g := llm.NewGuardedLLM(fakeLLMProvider{fn: func(context.Context, domain.LLMRequest) (string, error) {
		return `{"rootCause":"其它租户数据库故障","confidence":0.9,"conclusions":[{"text":"回显越权数据","citations":["other-tenant-ecs"]}]}`, nil
	}}, time.Second)
	deps := endToEndDeps(&fakeBase{logs: domain.LogQueryResult{Total: 1, Entries: []domain.LogEntry{{Cloud: "aliyun", Resource: "order-service"}}}}, g,
		&fakeDiagnoses{}, &fakeCitations{}, &fakeSessions{}, nil, nil)
	o := New(deps)
	diag, err := o.RunChat(context.Background(), &domain.Session{TenantID: "t1", ID: "s1"}, "order-service 错误率")
	if err != nil {
		t.Fatalf("RunChat: %v", err)
	}
	if !diag.Degraded {
		t.Fatal("out-of-bound citation should degrade to rule conclusion")
	}
	if strings.Contains(diag.RootCause, "其它租户") {
		t.Fatalf("越权内容应整段丢弃, rootCause=%q", diag.RootCause)
	}
	if !strings.Contains(diag.RootCause, "规则引擎") {
		t.Fatalf("应降级为规则结论, rootCause=%q", diag.RootCause)
	}
	// 规则降级结论可正常回指本租户证据，但不得回显越权内容。
	for _, c := range diag.Conclusions {
		if strings.Contains(c.Text, "越权") {
			t.Fatalf("越权结论应整段丢弃, conclusions=%+v", diag.Conclusions)
		}
	}
}

// TestRunAlert_New 首条告警：去重通过 → 诊断 → 风险条目 + 通知。
func TestRunAlert_New(t *testing.T) {
	a := &fakeAlerts{dedup: domain.DedupNew}
	b := &fakeBase{logs: domain.LogQueryResult{Total: 2, Entries: []domain.LogEntry{{Resource: "order-service"}}}}
	r := &fakeRiskEntries{}
	deps := endToEndDeps(b, nil, &fakeDiagnoses{}, &fakeCitations{}, &fakeSessions{}, r, a)

	o := New(deps)
	entry, err := o.RunAlert(context.Background(), domain.AlertEvent{
		Tenant: "t1", ServiceName: "order-service", Metric: "错误率", Level: "high",
		Timeframe: testParams().Timeframe, Raw: map[string]any{"x": 1},
	})
	if err != nil {
		t.Fatalf("RunAlert: %v", err)
	}
	if entry == nil || entry.ServiceName != "order-service" {
		t.Fatalf("entry = %+v, want risk entry", entry)
	}
	if r.inserted == nil || r.inserted.Fingerprint == "" {
		t.Fatal("risk entry should be persisted with fingerprint")
	}
	if a.linked != "dx-1" {
		t.Fatalf("alert should link diagnosis, got %q", a.linked)
	}
	if len(b.notified) != 1 {
		t.Fatal("notification should be sent")
	}
}

// TestRunAlert_Dedup 归并告警：窗口内已排查 → (nil, nil) 不新增风险条目。
func TestRunAlert_Dedup(t *testing.T) {
	a := &fakeAlerts{dedup: domain.DedupMerged}
	r := &fakeRiskEntries{}
	deps := endToEndDeps(&fakeBase{}, nil, &fakeDiagnoses{}, &fakeCitations{}, &fakeSessions{}, r, a)

	o := New(deps)
	entry, err := o.RunAlert(context.Background(), domain.AlertEvent{
		Tenant: "t1", ServiceName: "order-service", Metric: "错误率", Level: "high", Raw: map[string]any{},
	})
	if err != nil {
		t.Fatalf("RunAlert: %v", err)
	}
	if entry != nil {
		t.Fatal("merged alert should return nil entry")
	}
	if r.inserted != nil {
		t.Fatal("merged alert should not create risk entry")
	}
}
