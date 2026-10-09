package web

import "time"

// inProcObserver 进程内观测数据源（P1 简版：结构完整、指标待采集）。
//
// Agent 任务计数 / 负载时间序列 / LLM 预算实时聚合的采集点尚未接入编排层与
// worker（组合根补全时以零值/静态结构交付），页面可正常渲染、不报错；后续接入
// 采集点后仅需替换本实现。
type inProcObserver struct {
	concurrency int
	capacity    int
	budgetLimit int64
}

// NewInProcObserver 构造进程内观测数据源（capacity/concurrency 与 worker 阈值同源）。
func NewInProcObserver(concurrency, capacity int, budgetLimit int64) observabilityProvider {
	return &inProcObserver{concurrency: concurrency, capacity: capacity, budgetLimit: budgetLimit}
}

// Agents 4 逻辑 Agent 静态状态（任务计数待编排层埋点接入）。
func (o *inProcObserver) Agents() []agentStat {
	return []agentStat{
		{Name: "coordinator", Status: "ok"},
		{Name: "log_analyst", Status: "ok"},
		{Name: "monitor", Status: "ok"},
		{Name: "inspector", Status: "ok"},
	}
}

// Queue 队列容量/并发（与告警 worker 同源阈值；depth/inflight/overflow 待采集）。
func (o *inProcObserver) Queue() queueStat {
	return queueStat{Capacity: o.capacity, Concurrency: o.concurrency}
}

// LLMBudget LLM 滚动 1h 预算（预算上限随配置；实时调用次数待聚合）。
func (o *inProcObserver) LLMBudget() llmBudgetStat {
	return llmBudgetStat{WindowStart: time.Now().Truncate(time.Hour), BudgetLimit: o.budgetLimit}
}

// LoadHistory 负载时间序列（内存环形缓冲待接入，当前空）。
func (o *inProcObserver) LoadHistory() []loadPoint { return nil }
