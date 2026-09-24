// Package risk 风险分级注册表（确定性代码，非 DB 表）。
//
// 每个工具在注册表静态声明风险档（只读/低危/高危）；LLM 输出不得改变档位、
// 不得引入未注册工具。运行时低危白名单覆盖须命中注册表，否则拒绝保存。
//
// 类型唯一定义处在 internal/domain（tech-design › Interface 5）；
// 本包以类型别名复用，保持既有 `risk.ToolMeta` / `risk.RiskRegistry` 引用不变。
package risk

import "github.com/Havens-blog/opsagent/internal/domain"

// ToolMeta 工具风险元数据（别名 → domain.ToolMeta，勿在此重定义字段）。
type ToolMeta = domain.ToolMeta

// RiskRegistry 风险分级查表接口（别名 → domain.RiskRegistry）。
type RiskRegistry = domain.RiskRegistry
