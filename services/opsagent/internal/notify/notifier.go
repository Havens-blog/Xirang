// Package notify 通知渠道发送（钉钉/飞书/企微/邮件）+ 重试埋点。
//
// 只携带摘要与诊断入口链接，不含跨租户明细；发送结果 NotifyResult 用于
// SC-2 成功率埋点，重试计入失败率但不影响落库。具体实现在后续任务交付。
package notify

import (
	"context"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// 通知渠道常量（契约 B4 + api-handbook 通知渠道）。
const (
	ChannelDingTalk = "dingtalk"
	ChannelFeishu   = "feishu"
	ChannelWecom    = "wecom"
	ChannelEmail    = "email"
)

// maxAttempts 发送重试上限（契约 B4：含重试上限 3）。
const maxAttempts = 3

// Sender 单次渠道发送窄接口（依赖倒置：实现侧可接直连渠道 webhook / SMTP / 测试 fake）。
// 只负责一次发送，重试与埋点由 Notifier 承担。
type Sender interface {
	Send(ctx context.Context, channel string, body any) error
}

// Metrics 推送成功率埋点窄接口（opsagent_notify_success_total / opsagent_notify_failure_total{channel}）。
type Metrics interface {
	IncSuccess(channel string)
	IncFailure(channel string)
}

// Notifier 通知发送器：重试（上限 3）+ NotifyResult 塑造 + 成功率埋点。
//
// 语义与 BaseModuleClient.Notify 一致：发送失败不返回 error（计入 NotifyResult.OK=false），
// 不阻断上层落库与展示。
type Notifier struct {
	sender  Sender
	metrics Metrics
}

// New 构造 Notifier。metrics 为 nil 时埋点禁用（单测/降级场景）。
func New(sender Sender, metrics Metrics) *Notifier {
	return &Notifier{sender: sender, metrics: metrics}
}

// Send 发送告警提醒（摘要 + 诊断入口链接），最多重试 maxAttempts 次。
// 成功时清空失败计数语义（外层无状态）；终失败计入失败率埋点。
func (n *Notifier) Send(ctx context.Context, p domain.Notification) domain.NotifyResult {
	if n.sender == nil {
		return domain.NotifyResult{Channel: p.Channel, OK: false, LastError: "notify: sender nil"}
	}
	body := buildBody(p)
	var lastErr string
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := n.sender.Send(ctx, p.Channel, body)
		if err == nil {
			n.incSuccess(p.Channel)
			return domain.NotifyResult{
				Channel:  p.Channel,
				OK:       true,
				Attempts: attempt,
				SentAt:   time.Now(),
			}
		}
		lastErr = err.Error()
	}
	n.incFailure(p.Channel)
	return domain.NotifyResult{
		Channel:   p.Channel,
		OK:        false,
		Attempts:  maxAttempts,
		LastError: lastErr,
	}
}

// incSuccess 推送成功埋点（metrics 为 nil 时禁用）。
func (n *Notifier) incSuccess(channel string) {
	if n.metrics != nil {
		n.metrics.IncSuccess(channel)
	}
}

// incFailure 推送终失败埋点（metrics 为 nil 时禁用）。
func (n *Notifier) incFailure(channel string) {
	if n.metrics != nil {
		n.metrics.IncFailure(channel)
	}
}
