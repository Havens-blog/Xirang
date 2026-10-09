package middleware

import "strings"

// SanitizeUntrustedText 不可信文本（告警原文/日志正文/用户输入）进入 LLM 前的
// 注入清洗：剥离指令性标记（「忽略以上指令」「system:」等，大小写不敏感）。
//
// 与 internal/orchestrator 报告链路的 sanitizeForLLM 规则一致（tech-design › Security ①）；
// 本挂载点供 chat handler（4.4）在意图分类前清洗用户输入，告警侧由编排层报告链路清洗证据。
func SanitizeUntrustedText(text string) string {
	out := text
	for _, marker := range directiveMarkers {
		lowerMarker := strings.ToLower(marker)
		for {
			lower := strings.ToLower(out)
			i := strings.Index(lower, lowerMarker)
			if i < 0 {
				break
			}
			out = out[:i] + out[i+len(marker):]
		}
	}
	return out
}

// directiveMarkers 指令性标记序列（与 3.4 orchestrator 三防线清洗规则一致）。
var directiveMarkers = []string{
	"忽略以上的指令",
	"忽略以上指令",
	"忽略上述指令",
	"无视之前的指令",
	"ignore previous instructions",
	"ignore all previous instructions",
	"ignore the above instructions",
	"disregard all prior instructions",
	"you are now",
	"system:",
	"system：",
	"<|im_start|>system",
	"### instruction",
	"### system",
}
