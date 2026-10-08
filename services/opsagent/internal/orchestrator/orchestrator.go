package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/llm"
	"github.com/Havens-blog/opsagent/internal/repository/dao"
)

// ---- 窄 DAO 接口（依赖倒置：编排层只依赖所需方法，具体 DAO 结构上满足即可）----

// diagnosisSink 诊断落库窄接口。
type diagnosisSink interface {
	Insert(ctx context.Context, rec *dao.DiagnosisRecord) error
}

// citationSink 引用落库窄接口。
type citationSink interface {
	InsertBatch(ctx context.Context, tenantID, diagnosisID string, citations []domain.Citation) error
}

// sessionStatusUpdater 会话状态流转窄接口。
type sessionStatusUpdater interface {
	UpdateStatus(ctx context.Context, tenantID, id string, status domain.SessionStatus) error
}

// riskEntrySink 风险条目落库窄接口。
type riskEntrySink interface {
	Insert(ctx context.Context, e *domain.RiskEntry) error
}

// alertLedger 告警记账窄接口（去重 + 诊断回填）。
type alertLedger interface {
	RecordArrival(ctx context.Context, tenantID, fingerprint string, payload map[string]any, windowStart time.Time) (domain.DedupConclusion, int64, error)
	LinkDiagnosis(ctx context.Context, tenantID, fingerprint string, windowStart time.Time, diagnosisID string) error
}

// Deps 编排层依赖注入（全部为接口/窄接口，便于测试替换）。
type Deps struct {
	Intent      domain.IntentClassifier
	Base        domain.BaseModuleClient
	LLM         *llm.GuardedLLM
	Sessions    sessionStatusUpdater
	Diagnoses   diagnosisSink
	Citations   citationSink
	RiskEntries riskEntrySink
	Alerts      alertLedger
	Timeout     time.Duration // 同步编排硬预算（tech-design ≤25s）；0 = 不额外限时
}

// Orchestrator 编排流水线（domain.Orchestrator）：Coordinator 意图路由 →
// LogAnalyst/Monitor/Inspector 证据采集 → LLM 解读 → 报告组装 → 落库。
type Orchestrator struct {
	deps        Deps
	coordinator *CoordinatorAgent
	logAnalyst  *LogAnalystAgent
	monitor     *MonitorAgent
	inspector   *InspectorAgent
}

// New 构造 Orchestrator。
func New(deps Deps) *Orchestrator {
	return &Orchestrator{
		deps:        deps,
		coordinator: &CoordinatorAgent{intent: deps.Intent},
		logAnalyst:  &LogAnalystAgent{base: deps.Base},
		monitor:     &MonitorAgent{base: deps.Base},
		inspector:   &InspectorAgent{},
	}
}

// OutOfScopeError 超能力意图（上层返回引导式回应 Flow C，不调用底座）。
type OutOfScopeError struct{ Intent domain.IntentResult }

func (e *OutOfScopeError) Error() string { return "orchestrator: out-of-scope intent" }

// ClarifyError 意图需要澄清（上层返回 candidates 供一键纠正）。
type ClarifyError struct{ Intent domain.IntentResult }

func (e *ClarifyError) Error() string { return "orchestrator: needs clarification" }

// RunChat 对话式排障：意图 → 编排 → 诊断 → 落库，返回含根因/建议/引用的报告。
func (o *Orchestrator) RunChat(ctx context.Context, session *domain.Session, input string) (*domain.Diagnosis, error) {
	if o.deps.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.deps.Timeout)
		defer cancel()
	}
	tenant := session.TenantID

	// Step ① Coordinator：意图识别与能力路由。
	step, intentRes, err := o.coordinator.Run(ctx, 1, input, tenant)
	trace := []domain.TraceStep{step}
	if err == nil {
		if intentRes.Type == domain.IntentOutOfScope {
			return nil, &OutOfScopeError{Intent: intentRes}
		}
		if intentRes.NeedsClarify {
			return nil, &ClarifyError{Intent: intentRes}
		}
	}

	params := intentRes.Params
	if params.ServiceName == "" {
		params.ServiceName = session.ServiceName // 会话顶层 serviceName 兜底
	}
	if params.Timeframe.EndTime.IsZero() {
		now := time.Now()
		params.Timeframe = domain.Timeframe{StartTime: now.Add(-time.Hour), EndTime: now}
	}

	// Step ②③④ 证据采集（LogAnalyst 日志 / Monitor 资产 / Inspector P1 stub）。
	lstep, logs, _ := o.logAnalyst.Run(ctx, 2, params)
	mstep, assets, _ := o.monitor.Run(ctx, 3, tenant, params)
	istep := o.inspector.Run(ctx, 4, params)
	trace = append(trace, lstep, mstep, istep)

	// 组装报告（LLM 解读 + 引用；失败回退规则结论）。
	diag, err := o.assembleReport(ctx, tenant, params, buildEvidence{logs: logs, assets: assets}, 5)
	if err != nil {
		return nil, err
	}
	diag.SessionID = session.ID
	diag.Tenant = tenant
	diag.Severity = domain.SeverityP3 // 对话链路无告警级别，缺省 P3
	diag.Trace = append(trace, diag.Trace...)

	// 落库（失败不阻断响应，错误透传供上层 PersistFailedError 语义处理）。
	if err := o.persist(ctx, session, params.ServiceName, diag); err != nil {
		return diag, err
	}
	return diag, nil
}

// RunAlert 告警主动触发：去重 → 排查 → 诊断 → 落库 → 风险条目 + 通知。
// 返回 (nil, nil) 表示窗口内已排查（归并不新增风险条目）。
func (o *Orchestrator) RunAlert(ctx context.Context, alert domain.AlertEvent) (*domain.RiskEntry, error) {
	tenant := alert.Tenant
	fingerprint := domain.Fingerprint(tenant, alert.ServiceName, alert.Metric, alert.Level)
	windowStart := alert.Timeframe.StartTime
	if windowStart.IsZero() {
		windowStart = time.Now().Add(-time.Hour)
	}

	// 去重：记账并判定窗口内是否已排查。
	dedup, _, err := o.deps.Alerts.RecordArrival(ctx, tenant, fingerprint, alert.Raw, windowStart)
	if err != nil {
		return nil, err
	}
	if dedup == domain.DedupMerged {
		return nil, nil // 归并：不新增风险条目
	}

	params := domain.DiagnoseParams{
		ServiceName: alert.ServiceName,
		Metric:      alert.Metric,
		Timeframe:   alert.Timeframe,
	}
	if params.Timeframe.EndTime.IsZero() {
		now := time.Now()
		params.Timeframe = domain.Timeframe{StartTime: now.Add(-time.Hour), EndTime: now}
	}

	// 编排（告警已明确服务名/时间窗，跳过意图分类）。
	lstep, logs, _ := o.logAnalyst.Run(ctx, 1, params)
	mstep, assets, _ := o.monitor.Run(ctx, 2, tenant, params)
	diag, err := o.assembleReport(ctx, tenant, params, buildEvidence{logs: logs, assets: assets}, 3)
	if err != nil {
		return nil, err
	}
	diag.Tenant = tenant
	diag.Severity = severityOf(alert.Level)
	diag.Trace = append([]domain.TraceStep{lstep, mstep}, diag.Trace...)

	// 落库诊断 + 引用（无 session 模板，仅诊断）。
	if err := o.persistDiagnosis(ctx, params.ServiceName, diag); err != nil {
		return nil, err
	}

	// 回填指纹窗口的关联诊断。
	_ = o.deps.Alerts.LinkDiagnosis(ctx, tenant, fingerprint, windowStart, diag.ID)

	// 创建风险条目 + 通知（通知失败不阻断落库）。
	entry := &domain.RiskEntry{
		Tenant:      tenant,
		Fingerprint: fingerprint,
		DiagnosisID: diag.ID,
		ServiceName: alert.ServiceName,
		Severity:    diag.Severity,
	}
	if err := o.deps.RiskEntries.Insert(ctx, entry); err != nil {
		return nil, err
	}
	o.notifyAlert(ctx, alert, diag)
	return entry, nil
}

// persist 落库诊断 + 引用 + 会话状态流转（对话路径）。
func (o *Orchestrator) persist(ctx context.Context, session *domain.Session, serviceName string, diag *domain.Diagnosis) error {
	if err := o.persistDiagnosis(ctx, serviceName, diag); err != nil {
		return err
	}
	if session != nil && session.ID != "" {
		_ = o.deps.Sessions.UpdateStatus(ctx, session.TenantID, session.ID, domain.SessionStatusDone)
	}
	return nil
}

// persistDiagnosis 落库诊断文档 + 引用集合。
func (o *Orchestrator) persistDiagnosis(ctx context.Context, serviceName string, diag *domain.Diagnosis) error {
	rec := &dao.DiagnosisRecord{
		Diagnosis:   *diag,
		ServiceName: serviceName,
		CreatedAt:   time.Now(),
	}
	if err := o.deps.Diagnoses.Insert(ctx, rec); err != nil {
		return err
	}
	diag.ID = rec.Diagnosis.ID
	if err := o.deps.Citations.InsertBatch(ctx, diag.Tenant, diag.ID, diag.Citations); err != nil {
		return err
	}
	return nil
}

// notifyAlert 告警提醒推送（摘要 + 诊断入口链接；失败不阻断）。
func (o *Orchestrator) notifyAlert(ctx context.Context, alert domain.AlertEvent, diag *domain.Diagnosis) {
	if o.deps.Base == nil {
		return
	}
	_, _ = o.deps.Base.Notify(ctx, "dingtalk", domain.Notification{
		Tenant:      alert.Tenant,
		Channel:     "dingtalk",
		Title:       "运维告警：" + alert.ServiceName,
		Summary:     diag.RootCause,
		DiagLinkURL: "/opsagent/diagnosis/" + diag.ID,
	})
}

// buildEvidence 组装报告所需证据（日志 + 资产）。
type buildEvidence struct {
	logs   domain.LogQueryResult
	assets []domain.AssetRecord
}

// assembleReport 组装诊断报告：引用 + LLM 解读（失败回退规则结论）+ report 步骤。
// reportSeq 为该步在编排调用链中的序号（对话=5，告警=3）。
func (o *Orchestrator) assembleReport(ctx context.Context, tenant string, params domain.DiagnoseParams, ev buildEvidence, reportSeq int) (*domain.Diagnosis, error) {
	start := time.Now()
	diag := &domain.Diagnosis{
		Tenant:    tenant,
		RiskLevel: domain.RiskRead, // P1 只读推断，无处置执行
	}

	// 引用：日志条目 → Citation（结论回指数据源）；资产条目同理。
	citations := make([]domain.Citation, 0, len(ev.logs.Entries)+len(ev.assets))
	for _, e := range ev.logs.Entries {
		citations = append(citations, domain.Citation{
			SourceType: "log",
			SourceKey:  e.Resource,
			Snippet:    e.Message,
		})
	}
	for _, a := range ev.assets {
		citations = append(citations, domain.Citation{
			SourceType: "asset",
			SourceKey:  a.AssetID,
			Snippet:    a.Type,
		})
	}
	diag.Citations = citations

	switch {
	case ev.logs.Total == 0:
		// 空证据 → 明示无匹配日志，不编造。
		diag.RootCause = fmt.Sprintf("时间窗内无匹配日志（服务=%s），无法定位到可用证据", params.ServiceName)
		diag.Confidence = 0.5
	case o.interpretWithLLM(ctx, params, ev, diag):
		// LLM 解读成功：RootCause/Conclusions/Disposition 已产出。
	default:
		// LLM 不可用/失败 → 规则结论（L1 降级）。
		diag.Degraded = true
		diag.RootCause = fmt.Sprintf("时间窗内发现 %d 条日志（服务=%s），规则引擎给出初步结论", ev.logs.Total, params.ServiceName)
		diag.Confidence = 0.6
		diag.Conclusions = []domain.Conclusion{{
			Text:     "日志总量存在异常，建议人工复核（LLM 降级）",
			Citation: citationKeys(citations),
		}}
	}

	reportStep := newStep(reportSeq, agentCoordinator, "report", "报告组装", time.Since(start))
	if diag.Degraded {
		reportStep.Summary = "报告组装（规则降级）"
		reportStep.DegradeLevel = 1
	}
	diag.Trace = []domain.TraceStep{reportStep}
	return diag, nil
}

// interpretWithLLM 用 GuardedLLM 解读证据（经安全三防线：清洗 + 包裹 + 围栏），
// 产出 RootCause/Conclusions/Disposition；返回是否成功（失败 → false，交由规则降级）。
func (o *Orchestrator) interpretWithLLM(ctx context.Context, params domain.DiagnoseParams, ev buildEvidence, diag *domain.Diagnosis) bool {
	if o.deps.LLM == nil {
		return false
	}
	req := domain.LLMRequest{
		System:   reportSystemPrompt(),
		Messages: []domain.LLMMessage{{Role: "user", Text: evidenceForLLM(params, ev)}},
	}
	text, ok := o.deps.LLM.CompleteGuarded(ctx, req)
	if !ok {
		return false
	}
	var out llmReportOutput
	if err := json.Unmarshal([]byte(extractJSON(text)), &out); err != nil {
		return false
	}
	if out.RootCause == "" {
		return false
	}

	conclusions, ok := fenceConclusions(out, validKeySet(diag.Citations))
	if !ok {
		return false // 越权/越租户引用：整段丢弃 → 规则降级
	}

	diag.RootCause = out.RootCause
	diag.Confidence = out.Confidence
	diag.Conclusions = conclusions
	diag.Disposition = fenceDisposition(out)
	return true
}

// llmReportOutput LLM 诊断报告结构化输出。
type llmReportOutput struct {
	RootCause   string  `json:"rootCause"`
	Confidence  float64 `json:"confidence"`
	Conclusions []struct {
		Text      string   `json:"text"`
		Citations []string `json:"citations"`
	} `json:"conclusions"`
	Disposition []struct {
		Step   string `json:"step"`
		Risk   string `json:"risk"`
		Action string `json:"action"`
	} `json:"disposition"`
}

// citationKeys 提取引用 SourceKey 列表（规则结论回指）。
func citationKeys(citations []domain.Citation) []string {
	keys := make([]string, 0, len(citations))
	for _, c := range citations {
		keys = append(keys, c.SourceKey)
	}
	return keys
}

// severityOf 告警级别 → 诊断严重性（P0/P1/P2/P3 映射）。
func severityOf(level string) domain.Severity {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "P0", "CRITICAL":
		return domain.SeverityP0
	case "P1", "HIGH":
		return domain.SeverityP1
	case "P2", "MEDIUM":
		return domain.SeverityP2
	default:
		return domain.SeverityP3
	}
}

// extractJSON 从 LLM 输出中提取 JSON 片段（容忍 markdown 代码围栏与前后缀文本）。
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 {
		s = s[:j+1]
	}
	return s
}
