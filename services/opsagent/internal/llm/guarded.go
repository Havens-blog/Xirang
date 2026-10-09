// GuardedLLM 降级封装：隔离「不可用判定谓词」与「空串降级」。
//
// 谓词（tech-design › Error Handling）：单次调用超时 > 12s，或连续 2 次失败，或
// 未配置（disabled）→ IsAvailable=false。三级降级决策（L1/L2/L3）由 3.3
// DegradationController 负责，本类型只提供谓词与空串返回，不中断主链路。
package llm

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

const (
	// maxConsecutiveFailures 连续失败阈值（≥ 2 判定不可用）。
	maxConsecutiveFailures = 2
	// defaultTimeoutMs 单次调用超时默认值（> 12s 判定超时不可用）。
	defaultTimeout = 12 * time.Second
)

// GuardedLLM LLM 降级封装（tech-design › Interface 4）。
//
// CompleteGuarded 始终尝试调用（disabled 除外），成功即清零失败计数以自动恢复；
// 是否进入降级由 DegradationController 依据 IsAvailable + 冷却窗口决策并择机探活。
type GuardedLLM struct {
	provider domain.LLMProvider
	timeout  time.Duration

	mu       sync.Mutex
	failures int // 连续失败计数；单次超时直接置为阈值（单次超时即不可用）
}

// NewGuardedLLM 构造 GuardedLLM；timeout ≤ 0 取默认 12s。
func NewGuardedLLM(provider domain.LLMProvider, timeout time.Duration) *GuardedLLM {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &GuardedLLM{provider: provider, timeout: timeout}
}

// IsAvailable 判定谓词：未配置（disabled）/ 发生过超时 / 连续失败达阈值 → false。
func (g *GuardedLLM) IsAvailable(context.Context) bool {
	if !g.configured() {
		return false
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.failures < maxConsecutiveFailures
}

// configured 报告 provider 是否已配置（非 nil、非 disabledProvider）。
func (g *GuardedLLM) configured() bool {
	if g.provider == nil {
		return false
	}
	switch g.provider.(type) {
	case disabledProvider, *disabledProvider:
		return false
	default:
		return true
	}
}

// CompleteGuarded 带降级语义的生成调用：不可用/失败时返回空串 + available=false，
// 依赖 context 超时保证不阻塞；成功即清零失败计数（自动恢复）。
func (g *GuardedLLM) CompleteGuarded(ctx context.Context, req domain.LLMRequest) (string, bool) {
	if !g.configured() {
		return "", false
	}
	cctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	text, err := g.provider.Complete(cctx, req)

	g.mu.Lock()
	defer g.mu.Unlock()
	if err != nil {
		if isTimeout(err) {
			g.failures = maxConsecutiveFailures // 单次超时即不可用
		} else {
			g.failures++
		}
		return "", false
	}
	g.failures = 0
	return text, true
}

// isTimeout 判定 error 是否为调用超时（context 截至 / 主动取消）。
func isTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}
