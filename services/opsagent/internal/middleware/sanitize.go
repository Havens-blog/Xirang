package middleware

// SanitizeUntrustedText 不可信文本（告警原文/日志正文）进入 LLM 前的注入清洗挂载点。
//
// P0 骨架为透传占位；后续任务在此实现 tech-design › Security ② 的清洗规则：
// 剥离指令性标记（「忽略以上指令」「system:」等）。所有进入 LLM 的文本必须经过
// 本函数，清洗 + 结构化包裹 + 输出围栏构成 LLM 租户安全三道防线。
func SanitizeUntrustedText(text string) string {
	// TODO(后续任务): 实现 prompt injection 清洗规则（见 tech-design Security Mitigations）。
	return text
}
