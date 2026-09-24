// Package domain 跨模块共享的领域类型骨架。
//
// 完整 Shared Types 以 docs/features/haven-opsagent/design/tech-design.md ›
// Shared Types 为唯一定义处，后续任务按需迁入并扩展；本文件仅包含装配期
// 即被引用的枚举与身份类型。
package domain

// Severity 告警/诊断严重性（全局枚举）。
type Severity string

const (
	SeverityP0 Severity = "P0"
	SeverityP1 Severity = "P1"
	SeverityP2 Severity = "P2"
	SeverityP3 Severity = "P3"
)

// RiskLevel 工具风险档位（风险注册表静态声明，LLM 不得改变）。
type RiskLevel string

const (
	RiskRead RiskLevel = "read"
	RiskLow  RiskLevel = "low"
	RiskHigh RiskLevel = "high"
)

// IntentType 意图分类结果类型。
type IntentType string

const (
	IntentDiagnose   IntentType = "diagnose"     // 排障诊断
	IntentResource   IntentType = "resource"     // 资源查询
	IntentOutOfScope IntentType = "out_of_scope" // 超能力意图 → 引导式回应
)

// Identity eiam 令牌验证结果（tech-design 契约 B3）。
type Identity struct {
	UserID string   `json:"userId"`
	Tenant string   `json:"tenant"`
	Roles  []string `json:"roles"`
}
