package intent

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 模板/正则意图匹配（L1 降级内置于 Classify，不依赖 LLM）。词表与规则按 PRD
// 主流程 A/C 的示例与超能力意图清单设计，覆盖「排障诊断 / 资源查询 / 超能力引导」三态。

var (
	// serviceKebabRe 提取 kebab/snake-case 服务名（order-service / cdn_edge / cache-svc…），
	// 要求至少一个 -/_ 连字符以降低误报。
	serviceKebabRe = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9]*(?:[-_][a-zA-Z0-9]+)+`)
	// serviceZhRe 提取「xx服务」中文服务名。
	serviceZhRe = regexp.MustCompile(`([\p{Han}A-Za-z0-9_-]+?)服务`)

	// timeWindowRe 提取「最近/近/过去 N 分钟/小时/天/周」时间窗。
	timeWindowRe = regexp.MustCompile(`(?:最近|近|过去|last|past)\s*(\d+)\s*(分钟|小时|天|周|min|hour|hr|h|day|d|w|m)`)

	metricKeywords = []string{
		"错误率", "错误", "error", "失败率", "失败", "延迟", "latency", "qps",
		"流量", "超时", "timeout", "cpu", "内存", "memory", "磁盘", "disk", "连接数",
	}

	outOfScopeKeywords = []string{
		"开发咨询", "开发", "写代码", "代码", "bug", "提单", "工单", "拓扑", "架构", "重构",
		"闲聊", "你好", "谢谢", "再见", "采购", "报销", "帮我写", "review", "审查", "文档",
	}

	resourceKeywords = []string{
		"资产", "配置", "资源", "有几台", "多少台", "实例", "账号", "区域", "清单",
		"ecs", "rds", "redis", "eip", "nas", "oss", "slb", "域名", "在哪里", "在哪",
	}
)

// extractServiceName 抽取服务名（优先级：kebab/snake 服务名 > 「xx服务」> 无）。
func extractServiceName(input string) string {
	if m := serviceKebabRe.FindString(input); m != "" {
		return strings.ToLower(m)
	}
	if m := serviceZhRe.FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}
	return ""
}

// extractMetric 抽取指标关键词；抽取不出返回空串。
func extractMetric(input string) string {
	lower := strings.ToLower(input)
	for _, k := range metricKeywords {
		if strings.Contains(lower, k) {
			return k
		}
	}
	return ""
}

// extractTimeframe 抽取时间窗（缺省最近 1h）。返回 (start, end, 是否显式命中)。
func extractTimeframe(input string, now time.Time) (time.Time, time.Time, bool) {
	m := timeWindowRe.FindStringSubmatch(strings.ToLower(input))
	if len(m) != 3 {
		return now.Add(-time.Hour), now, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return now.Add(-time.Hour), now, false
	}
	var d time.Duration
	switch m[2] {
	case "分钟", "min", "m":
		d = time.Duration(n) * time.Minute
	case "小时", "hour", "hr", "h":
		d = time.Duration(n) * time.Hour
	case "天", "day", "d":
		d = time.Duration(n) * 24 * time.Hour
	case "周", "w":
		d = time.Duration(n) * 7 * 24 * time.Hour
	default:
		d = time.Duration(n) * time.Hour
	}
	return now.Add(-d), now, true
}

// detectOutOfScope 报告输入是否为超能力意图。
func detectOutOfScope(input string) bool {
	lower := strings.ToLower(input)
	for _, k := range outOfScopeKeywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

// detectResource 报告输入是否为资源查询意图。
func detectResource(input string) bool {
	lower := strings.ToLower(input)
	for _, k := range resourceKeywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}
