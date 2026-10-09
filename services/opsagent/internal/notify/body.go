package notify

import "github.com/Havens-blog/opsagent/internal/domain"

// buildBody 按渠道组装通知载荷（契约 B4 内容约束：只携带标题 + 摘要 + 诊断入口链接，
// 不含跨租户明细）。
//
// 渠道载荷形状对齐底座直连 webhook 约定：钉钉/企微 markdown，飞书 text，邮件 subject/body。
func buildBody(p domain.Notification) any {
	text := "**" + p.Title + "**\n" + p.Summary
	if p.DiagLinkURL != "" {
		text += "\n[查看诊断](" + p.DiagLinkURL + ")"
	}

	switch p.Channel {
	case ChannelWecom:
		return map[string]any{"msgtype": "markdown", "markdown": map[string]any{"content": text}}
	case ChannelFeishu:
		return map[string]any{"msg_type": "text", "content": map[string]any{"text": text}}
	case ChannelEmail:
		return map[string]any{"subject": p.Title, "body": text}
	default: // 钉钉及兜底
		return map[string]any{"msgtype": "markdown", "markdown": map[string]any{"title": p.Title, "text": text}}
	}
}
