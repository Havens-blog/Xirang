package baseclient

import (
	"context"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

const notifyModule = "notify"

// notifyMaxAttempts 通知发送重试上限（契约 B4：含重试上限 3）。
const notifyMaxAttempts = 3

// Notify 发送诊断入口通知（契约 B4：opsagent 直连渠道 webhook）。发送失败不返回错误
// （notify 语义：失败计入 NotifyResult.OK=false，不阻断诊断落库与响应）。
func (c *Client) Notify(ctx context.Context, channel string, payload domain.Notification) (domain.NotifyResult, error) {
	res := domain.NotifyResult{Channel: channel}
	webhook, ok := c.webhooks[channel]
	if !ok || webhook == "" {
		res.LastError = "webhook not configured"
		return res, nil
	}

	body := buildNotifyBody(payload)
	for attempts := 1; attempts <= notifyMaxAttempts; attempts++ {
		if err := c.postRaw(ctx, notifyModule, webhook, body); err == nil {
			res.OK = true
			res.Attempts = attempts
			res.SentAt = time.Now()
			return res, nil
		} else {
			res.LastError = err.Error()
		}
	}
	res.Attempts = notifyMaxAttempts
	return res, nil
}

// buildNotifyBody 按渠道组装载荷（钉钉/企微 markdown、飞书 text）。只携带标题 + 摘要 +
// 诊断入口链接（契约 B4 内容约束，不含跨租户明细）。
func buildNotifyBody(p domain.Notification) any {
	text := "**" + p.Title + "**\n" + p.Summary
	if p.DiagLinkURL != "" {
		text += "\n[查看诊断](" + p.DiagLinkURL + ")"
	}
	switch p.Channel {
	case "wecom":
		return map[string]any{"msgtype": "markdown", "markdown": map[string]any{"content": text}}
	case "feishu":
		return map[string]any{"msg_type": "text", "content": map[string]any{"text": text}}
	default: // dingtalk 及兜底
		return map[string]any{"msgtype": "markdown", "markdown": map[string]any{"title": p.Title, "text": text}}
	}
}
