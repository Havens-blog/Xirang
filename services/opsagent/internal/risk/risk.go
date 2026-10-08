// Package risk 风险分级注册表（确定性代码，非 DB 表）。
//
// 每个工具在注册表静态声明风险档（只读/低危/高危）；LLM 输出不得改变档位、
// 不得引入未注册工具。运行时低危白名单覆盖须命中注册表，否则拒绝保存。
//
// 类型唯一定义处在 internal/domain（tech-design › Interface 5）；
// 本包以类型别名复用，保持既有 `risk.ToolMeta` / `risk.RiskRegistry` 引用不变。
package risk

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// ToolMeta 工具风险元数据（别名 → domain.ToolMeta，勿在此重定义字段）。
type ToolMeta = domain.ToolMeta

// RiskRegistry 风险分级查表接口（别名 → domain.RiskRegistry）。
type RiskRegistry = domain.RiskRegistry

// 错误哨兵。
var (
	// ErrUnregistered 工具未注册（Classify 未命中 / 白名单引入未注册工具）。
	ErrUnregistered = errors.New("risk: unregistered tool")
	// ErrHighNotAllowlisted 高危工具不允许入白名单。
	ErrHighNotAllowlisted = errors.New("risk: high-risk tool cannot be allowlisted")
	// ErrInvalidWhitelistLevel 白名单条目 riskLevel 非法（非 read|low）。
	ErrInvalidWhitelistLevel = errors.New("risk: whitelist entry level must be read or low")
)

// staticLevels 工具风险档静态声明（确定性代码，唯一档位来源）。
//
// P1 排障主链路全部只读；低危/高危为 P3「可控执行」预留档，白名单仅覆盖低危
// （high 永远不允许入白名单）。粒度与底座能力对齐：logquery/诊断/告警/资产查询为
// 只读；通知发送为低危；重启/回滚/封禁/删除为高危。
var staticLevels = map[string]domain.RiskLevel{
	// ---- 只读（P1 排障主链路，无写/外部副作用）----
	"query_log":   domain.RiskRead, // 日志联邦查询（契约 B1）
	"diagnose":    domain.RiskRead, // 规则引擎风险分级（契约 B1）
	"query_alert": domain.RiskRead, // 告警查询
	"query_asset": domain.RiskRead, // 资产查询（契约 B2 MCP）

	// ---- 低危（P3 可控执行；白名单候选）----
	"notify": domain.RiskLow, // 通知发送（外部副作用但可控）

	// ---- 高危（P3 可控执行；永不允许入白名单 / 不渲染确认控件）----
	"restart_service": domain.RiskHigh, // 重启服务
	"rollback_deploy": domain.RiskHigh, // 回滚部署
	"ban_ip":          domain.RiskHigh, // 封禁 IP
	"delete_asset":    domain.RiskHigh, // 删除资产
}

// Registry 风险分级注册表。levels 为不可变静态档位（进程内共享只读）；allow 为
// 运行时白名单（settings 运维覆盖，随 ApplyWhitelist 整体替换）。
type Registry struct {
	levels map[string]domain.RiskLevel

	mu    sync.RWMutex // 保护 allow
	allow map[string]bool
}

// NewRegistry 构造注册表（内置默认工具档位，白名单初始为空）。
func NewRegistry() *Registry {
	return &Registry{
		levels: staticLevels,
		allow:  map[string]bool{},
	}
}

// Classify 查表判定；未注册返回 ErrUnregistered（不默认放行）。
func (r *Registry) Classify(tool string) (domain.ToolMeta, error) {
	lvl, ok := r.levels[tool]
	if !ok {
		return domain.ToolMeta{}, ErrUnregistered
	}
	r.mu.RLock()
	allow := r.allow[tool]
	r.mu.RUnlock()
	return domain.ToolMeta{Name: tool, RiskLevel: lvl, Allowlist: allow}, nil
}

// ApplyWhitelist 整体替换白名单（api-handbook §8 写语义）；任一条目未注册、目标
// 高危或条目 riskLevel 非法 → 整体拒绝（返回错误，不部分生效）。
func (r *Registry) ApplyWhitelist(entries []domain.RiskWhitelistEntry) error {
	next := make(map[string]bool, len(entries))
	for _, e := range entries {
		lvl, ok := r.levels[e.Tool]
		if !ok {
			return fmt.Errorf("%w: %q", ErrUnregistered, e.Tool)
		}
		if lvl == domain.RiskHigh {
			return fmt.Errorf("%w: %q", ErrHighNotAllowlisted, e.Tool)
		}
		if e.RiskLevel != domain.RiskRead && e.RiskLevel != domain.RiskLow {
			return fmt.Errorf("%w: %q=%q", ErrInvalidWhitelistLevel, e.Tool, e.RiskLevel)
		}
		next[e.Tool] = true
	}
	r.mu.Lock()
	r.allow = next
	r.mu.Unlock()
	return nil
}
