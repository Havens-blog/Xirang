package orchestrator

import (
	"sync"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// DegradeLevel 三级降级级别（TraceStep.DegradeLevel / chat 响应 degradeLevel）。
type DegradeLevel int

const (
	// LevelNormal L0 正常：LLM 可用，完整 LLM 解读。
	LevelNormal DegradeLevel = 0
	// LevelTemplate L1 模板意图匹配：LLM 不可用，意图回退模板/正则，报告由规则引擎结论填充。
	LevelTemplate DegradeLevel = 1
	// LevelPreset L2 预置查询入口：低置信/无服务名，返回预置查询目录。
	LevelPreset DegradeLevel = 2
	// LevelNotice L3 明示降级：预置目录空或规则结论不可得，明示降级模式。
	LevelNotice DegradeLevel = 3
)

// chat 响应 type 枚举（api-handbook §1，严格一致）。
const (
	respTypeReport         = "report"          // 正常/L1
	respTypePresetEntries  = "preset_entries"  // L2
	respTypeDegradedNotice = "degraded_notice" // L3
)

// DegradeDecision 降级决策结果（chat 响应 200 data 增量字段同形）。
type DegradeDecision struct {
	Level    DegradeLevel // 0/1/2/3
	Degraded bool         // LLM 降级标识
	Type     string       // report | preset_entries | degraded_notice
}

// Stamp 将降级级别计入编排调用链步骤（前端徽标依据）。
func (d DegradeDecision) Stamp(step *domain.TraceStep) {
	step.DegradeLevel = int(d.Level)
}

// 降级谓词缺省参数（tech-design：冷却 5min、模板匹配置信度阈值 0.6 可配）。
const (
	defaultCooldown  = 5 * time.Minute
	defaultThreshold = 0.6
)

// DegradationController 三级降级统一决策器（tech-design › Error Handling › LLM 三级降级设计）。
//
// 逐级尝试、只降不升：进入降级后本会话冷却 5min 内不再探活；冷却后经 probe
// 探活，成功即回升 L0、失败则重置冷却继续降级。降级触发/恢复均随返回值暴露，
// 由调用方落监控。降级谓词（超时>12s/连续2次失败/预算耗尽）封装在 probe 内，
// 本控制器只关心冷却与级别决策，不感知具体谓词实现。
type DegradationController struct {
	cooldown  time.Duration
	threshold float64
	now       func() time.Time

	mu        sync.Mutex
	downgrade map[string]time.Time // sessionKey → 进入降级时刻（冷却起点）
}

// NewDegradationController 构造降级控制器。cooldown≤0 取默认 5min；threshold 不在
// (0,1] 取默认 0.6；now 为 nil 取 time.Now（测试可注入）。
func NewDegradationController(cooldown time.Duration, threshold float64, now func() time.Time) *DegradationController {
	if cooldown <= 0 {
		cooldown = defaultCooldown
	}
	if threshold <= 0 || threshold > 1 {
		threshold = defaultThreshold
	}
	if now == nil {
		now = time.Now
	}
	return &DegradationController{
		cooldown:  cooldown,
		threshold: threshold,
		now:       now,
		downgrade: map[string]time.Time{},
	}
}

// Decide 统一降级决策。probe 仅在需要探活时被调用（冷却期内不调用，避免重复消耗
// LLM 预算）；意图分类回退模板匹配已由 IntentClassifier 完成，此处 intent 为分类结果。
// presets 为 L2 预置查询目录（空则 L2 落 L3）。
func (c *DegradationController) Decide(sessionKey string, probe func() bool, intent domain.IntentResult, presets []domain.PresetQuery) DegradeDecision {
	now := c.now()

	c.mu.Lock()
	enteredAt, degraded := c.downgrade[sessionKey]
	c.mu.Unlock()

	// 冷却窗口内：不探活，维持降级级别（只降不升）。
	if degraded && now.Sub(enteredAt) < c.cooldown {
		return c.decideLevel(intent, presets)
	}

	// 冷却过期 / 从未降级：探活（锁外调用，避免阻塞并发会话）。
	if probe() {
		c.mu.Lock()
		delete(c.downgrade, sessionKey) // 回升 L0，重置降级状态
		c.mu.Unlock()
		return DegradeDecision{Level: LevelNormal, Type: respTypeReport}
	}

	// 探活失败：进入/维持降级，重置冷却起点。
	c.mu.Lock()
	c.downgrade[sessionKey] = now
	c.mu.Unlock()
	return c.decideLevel(intent, presets)
}

// decideLevel LLM 不可用时的级别判定（纯决策、无状态）。
func (c *DegradationController) decideLevel(intent domain.IntentResult, presets []domain.PresetQuery) DegradeDecision {
	// L2 触发：模板匹配置信度 < 阈值，或抽取不出服务名。
	if intent.Confidence < c.threshold || intent.Params.ServiceName == "" {
		if len(presets) == 0 {
			// L3：预置目录空 → 明示降级。
			return DegradeDecision{Level: LevelNotice, Degraded: true, Type: respTypeDegradedNotice}
		}
		return DegradeDecision{Level: LevelPreset, Degraded: true, Type: respTypePresetEntries}
	}
	// L1：模板匹配置信且服务名可得 → 模板报告填充规则结论。
	return DegradeDecision{Level: LevelTemplate, Degraded: true, Type: respTypeReport}
}
