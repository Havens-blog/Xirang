package domain

// 本文件为跨模块共享的全局枚举（tech-design › Shared Types 唯一定义处）。
// 每个枚举附带合法性校验函数 IsValid，供边界层（web/worker/llm）入参校验复用。

// Severity 告警/诊断严重性（全局枚举）。
type Severity string

const (
	SeverityP0 Severity = "P0"
	SeverityP1 Severity = "P1"
	SeverityP2 Severity = "P2"
	SeverityP3 Severity = "P3"
)

// IsValid 报告 s 是否为合法 Severity（P0|P1|P2|P3）。
func (s Severity) IsValid() bool {
	switch s {
	case SeverityP0, SeverityP1, SeverityP2, SeverityP3:
		return true
	default:
		return false
	}
}

// RiskLevel 工具风险档位（风险注册表静态声明，LLM 不得改变）。
type RiskLevel string

const (
	RiskRead RiskLevel = "read"
	RiskLow  RiskLevel = "low"
	RiskHigh RiskLevel = "high"
)

// IsValid 报告 r 是否为合法 RiskLevel（read|low|high）。
func (r RiskLevel) IsValid() bool {
	switch r {
	case RiskRead, RiskLow, RiskHigh:
		return true
	default:
		return false
	}
}

// IntentType 意图分类结果类型。
type IntentType string

const (
	IntentDiagnose   IntentType = "diagnose"     // 排障诊断
	IntentResource   IntentType = "resource"     // 资源查询
	IntentOutOfScope IntentType = "out_of_scope" // 超能力意图 → 引导式回应
)

// IsValid 报告 t 是否为合法 IntentType（diagnose|resource|out_of_scope）。
func (t IntentType) IsValid() bool {
	switch t {
	case IntentDiagnose, IntentResource, IntentOutOfScope:
		return true
	default:
		return false
	}
}

// SessionType 会话触发来源（er-diagram sessions.type：chat|alert）。
type SessionType string

const (
	SessionTypeChat  SessionType = "chat"  // 对话排障
	SessionTypeAlert SessionType = "alert" // 告警主动触发
)

// SessionStatus 会话状态（er-diagram sessions.status：running|done|failed）。
type SessionStatus string

const (
	SessionStatusRunning SessionStatus = "running"
	SessionStatusDone    SessionStatus = "done"
	SessionStatusFailed  SessionStatus = "failed"
)

// IsValid 报告 s 是否为合法 SessionStatus（running|done|failed）。
func (s SessionStatus) IsValid() bool {
	switch s {
	case SessionStatusRunning, SessionStatusDone, SessionStatusFailed:
		return true
	default:
		return false
	}
}

// RiskEntryStatus 风险条目状态机（risk_entries.status：pending_view|viewed|done）。
type RiskEntryStatus string

const (
	RiskStatusPendingView RiskEntryStatus = "pending_view"
	RiskStatusViewed      RiskEntryStatus = "viewed"
	RiskStatusDone        RiskEntryStatus = "done"
)

// IsValid 报告 s 是否为合法 RiskEntryStatus（pending_view|viewed|done）。
func (s RiskEntryStatus) IsValid() bool {
	switch s {
	case RiskStatusPendingView, RiskStatusViewed, RiskStatusDone:
		return true
	default:
		return false
	}
}

// DedupConclusion 告警记账结论（alerts.dedup_conclusion）。
// new=窗口首条建文档（count=1）；merged=同指纹归并（$inc count）；overflow_dropped=队列溢出丢弃留档。
type DedupConclusion string

const (
	DedupNew             DedupConclusion = "new"
	DedupMerged          DedupConclusion = "merged"
	DedupOverflowDropped DedupConclusion = "overflow_dropped"
)

// IsValid 报告 c 是否为合法 DedupConclusion（new|merged|overflow_dropped）。
func (c DedupConclusion) IsValid() bool {
	switch c {
	case DedupNew, DedupMerged, DedupOverflowDropped:
		return true
	default:
		return false
	}
}

// SettingScope 系统配置作用域（settings.scope）。
type SettingScope string

const (
	ScopeDatasource      SettingScope = "datasource"
	ScopeLLM             SettingScope = "llm"
	ScopeNotify          SettingScope = "notify"
	ScopeRiskWhitelist   SettingScope = "risk_whitelist"
	ScopePresetQueries   SettingScope = "preset_queries"
	ScopeGuidedTemplates SettingScope = "guided_templates"
)

// IsValid 报告 s 是否为合法 SettingScope（六类作用域）。
func (s SettingScope) IsValid() bool {
	switch s {
	case ScopeDatasource, ScopeLLM, ScopeNotify, ScopeRiskWhitelist, ScopePresetQueries, ScopeGuidedTemplates:
		return true
	default:
		return false
	}
}

// CompensationKind 补偿类型（compensations.kind；P1 仅落库失败重放）。
type CompensationKind string

const (
	CompKindPersistFailed CompensationKind = "persist_failed"
)

// CompensationStatus 补偿项状态（compensations.status：retry|dead）。
type CompensationStatus string

const (
	CompStatusRetry CompensationStatus = "retry"
	CompStatusDead  CompensationStatus = "dead"
)

// IsValid 报告 s 是否为合法 CompensationStatus（retry|dead）。
func (s CompensationStatus) IsValid() bool {
	switch s {
	case CompStatusRetry, CompStatusDead:
		return true
	default:
		return false
	}
}
