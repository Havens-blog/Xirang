package web

import (
	"time"

	"github.com/gin-gonic/gin"
)

// agentStat 逻辑 Agent 观测（api-handbook §10 agents）。
type agentStat struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	TasksTotal    int64  `json:"tasksTotal"`
	TasksFailed   int64  `json:"tasksFailed"`
	AvgDurationMs int64  `json:"avgDurationMs"`
}

// queueStat 告警 worker 队列观测（深度/容量/在途/并发/溢出总量）。
type queueStat struct {
	Depth         int   `json:"depth"`
	Capacity      int   `json:"capacity"`
	Inflight      int   `json:"inflight"`
	Concurrency   int   `json:"concurrency"`
	OverflowTotal int64 `json:"overflowTotal"`
}

// llmBudgetStat LLM 滚动 1h 预算观测（llm_usage 当前窗口聚合）。
type llmBudgetStat struct {
	WindowStart time.Time `json:"windowStart"`
	Calls       int64     `json:"calls"`
	BudgetLimit int64     `json:"budgetLimit"`
	Exceeded    bool      `json:"exceeded"`
}

// loadPoint 负载时间序列点（内存环形缓冲，最近 1h）。
type loadPoint struct {
	Ts       int64 `json:"ts"`
	Depth    int   `json:"depth"`
	Inflight int   `json:"inflight"`
}

// observabilityData GET /agents/observability 响应。
type observabilityData struct {
	Agents      []agentStat   `json:"agents"`
	Queue       queueStat     `json:"queue"`
	LLMBudget   llmBudgetStat `json:"llmBudget"`
	LoadHistory []loadPoint   `json:"loadHistory"`
}

// observabilityProvider 观测数据源窄接口（由 agent/worker 指标暴露）。
type observabilityProvider interface {
	Agents() []agentStat
	Queue() queueStat
	LLMBudget() llmBudgetStat
	LoadHistory() []loadPoint
}

// ObservabilityHandler Agent 观测 handler（GET /agents/observability）。
type ObservabilityHandler struct {
	prov observabilityProvider
}

// NewObservabilityHandler 构造。
func NewObservabilityHandler(prov observabilityProvider) *ObservabilityHandler {
	return &ObservabilityHandler{prov: prov}
}

// RegisterRoutes 注册观测路由（Module 接口）。
func (h *ObservabilityHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/agents/observability", h.Observe)
}

// Observe 返回 4 逻辑 Agent 状态/负载 + 队列 + LLM 预算 + 负载历史。
func (h *ObservabilityHandler) Observe(c *gin.Context) {
	if h.prov == nil {
		OK(c, observabilityData{})
		return
	}
	OK(c, observabilityData{
		Agents:      h.prov.Agents(),
		Queue:       h.prov.Queue(),
		LLMBudget:   h.prov.LLMBudget(),
		LoadHistory: h.prov.LoadHistory(),
	})
}
