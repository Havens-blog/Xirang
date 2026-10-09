package orchestrator

import (
	"context"
	"strconv"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// 逻辑 Agent 名（TraceStep.Agent；与 domain.TraceStep 文档枚举一致）。
const (
	agentCoordinator = "coordinator"
	agentLogAnalyst  = "log_analyst"
	agentMonitor     = "monitor"
	agentInspector   = "inspector"
)

// noDataSummary 某步未取得数据时的显式标注（tech-design：降级/无数据不中断主链路）。
const noDataSummary = "未取得数据"

// newStep 构造一步编排调用链（统一补 DurationMs，序号由调用方给定）。
func newStep(seq int, agent, action, summary string, elapsed time.Duration) domain.TraceStep {
	return domain.TraceStep{
		Step:       seq,
		Agent:      agent,
		Action:     action,
		DurationMs: elapsed.Milliseconds(),
		Summary:    summary,
	}
}

// CoordinatorAgent 意图识别与能力路由（tech-design › Interface 1）。
// Run 返回一步 TraceStep、分类结果与错误；LLM 失败由 Classifier 内部回退规则
// （L1），此处收到的 err 仅在伪造实现或极端情况下非空。
type CoordinatorAgent struct {
	intent domain.IntentClassifier
}

func (a *CoordinatorAgent) Run(ctx context.Context, seq int, input, tenant string) (domain.TraceStep, domain.IntentResult, error) {
	start := time.Now()
	res, err := a.intent.Classify(ctx, input, tenant)
	step := newStep(seq, agentCoordinator, "intent", "意图识别", time.Since(start))
	if err != nil {
		step.Summary = noDataSummary
		step.DegradeLevel = 1
		return step, res, err
	}
	step.Summary = "意图=" + string(res.Type)
	return step, res, nil
}

// LogAnalystAgent 日志联邦查询（baseclient.QueryLogs，契约 B1）。
// 入口资源引用（Citation）由 assembleReport 依据 logs.Entries 统一构造。
type LogAnalystAgent struct {
	base domain.BaseModuleClient
}

func (a *LogAnalystAgent) Run(ctx context.Context, seq int, params domain.DiagnoseParams) (domain.TraceStep, domain.LogQueryResult, error) {
	start := time.Now()
	req := domain.LogQueryRequest{
		LogType:   firstLogType(params.LogTypes),
		Resource:  params.ServiceName,
		StartTime: params.Timeframe.StartTime,
		EndTime:   params.Timeframe.EndTime,
		Limit:     100,
	}
	logs, err := a.base.QueryLogs(ctx, req)
	step := newStep(seq, agentLogAnalyst, "query", "查询日志", time.Since(start))
	if err != nil {
		step.Summary = noDataSummary
		step.DegradeLevel = 1
		return step, domain.LogQueryResult{}, err
	}
	step.Summary = "日志 total=" + itoa(logs.Total)
	step.Source = "logquery"
	return step, logs, nil
}

// MonitorAgent 资产上下文（baseclient.QueryAssets，契约 B2）。
type MonitorAgent struct {
	base domain.BaseModuleClient
}

func (a *MonitorAgent) Run(ctx context.Context, seq int, tenant string, params domain.DiagnoseParams) (domain.TraceStep, []domain.AssetRecord, error) {
	start := time.Now()
	assets, err := a.base.QueryAssets(ctx, tenant, "query_asset", map[string]any{"serviceName": params.ServiceName})
	step := newStep(seq, agentMonitor, "query", "查询资产上下文", time.Since(start))
	if err != nil {
		step.Summary = noDataSummary
		step.DegradeLevel = 1
		return step, nil, err
	}
	step.Summary = "资产 total=" + itoa(int64(len(assets)))
	step.Source = "asset"
	return step, assets, nil
}

// InspectorAgent 变更/拓扑上下文（P1 stub：留接口，P2 接入）。
type InspectorAgent struct{}

func (a *InspectorAgent) Run(ctx context.Context, seq int, params domain.DiagnoseParams) domain.TraceStep {
	start := time.Now()
	step := newStep(seq, agentInspector, "query", "查询变更/拓扑上下文", time.Since(start))
	step.Summary = noDataSummary
	step.DegradeLevel = 1
	return step
}

// firstLogType 取首个日志类型（空则空串，契约 B1 视为全类型）。
func firstLogType(types []string) string {
	if len(types) == 0 {
		return ""
	}
	return types[0]
}

// itoa int64 → string（规避 strconv 直连噪音）。
func itoa(n int64) string { return strconv.FormatInt(n, 10) }
