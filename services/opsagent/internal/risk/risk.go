// Package risk 风险分级注册表（确定性代码，非 DB 表）。
//
// 每个工具在注册表静态声明风险档（只读/低危/高危）；LLM 输出不得改变档位、
// 不得引入未注册工具。运行时低危白名单覆盖须命中注册表，否则拒绝保存。
package risk

import "github.com/Havens-blog/opsagent/internal/domain"

// ToolMeta 工具风险元数据（tech-design › Interface 5）。
type ToolMeta struct {
	Name      string
	RiskLevel domain.RiskLevel // 注册表静态声明，LLM 不得改变
	Allowlist bool             // 是否在低危白名单内
}

// RiskRegistry 风险分级查表接口。
type RiskRegistry interface {
	// Classify 查表判定；返回错误当且仅当工具未注册。
	Classify(tool string) (ToolMeta, error)
}
