// Package intent 意图识别与能力路由（tech-design › Interface 1）。
//
// Classify 将自然语言输入分类到 diagnose / resource / out_of_scope，抽取
// DiagnoseParams（服务名/时间窗/指标），返回 IntentResult（含置信度/备选意图/澄清标记）；
// Correct 支持 target+params XOR message 两条纠正路径并追加 correction_log。
// LLM 不可用时回退模板/正则匹配（L1 降级内置于 Classify，不依赖 LLM）；
// 超能力意图不调用底座、不猜测、不真实转发（引导式回应话术见 Guided）。
package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// clarifyThreshold 澄清阈值：置信度低于此值触发 NeedsClarify（供一键纠正）。
const clarifyThreshold = 0.6

// llmTimeout 单次意图分类 LLM 调用预算（tech-design：意图识别 2s，模板 <100ms）。
const llmTimeout = 2 * time.Second

// GuidedTemplate 引导式回应话术（settings.guided_templates 条目，Flow C）。
type GuidedTemplate struct {
	Template string `json:"template"` // 话术模板，如「这是开发咨询类问题，请走 X 渠道」
	Channel  string `json:"channel"`  // 正确渠道/系统入口
}

// SessionStore 会话存取窄接口（Correct 读旧意图 + 追加纠正记录；由 orchestrator 注入）。
type SessionStore interface {
	GetByID(ctx context.Context, tenantID, id string) (domain.Session, error)
	AppendCorrection(ctx context.Context, tenantID, id string, rec domain.CorrectionRecord) error
}

// Classifier 意图识别与能力路由实现（domain.IntentClassifier）。
type Classifier struct {
	llm     domain.LLMProvider
	store   SessionStore
	guided  map[string]GuidedTemplate // 引导话术（key=超能力意图类别，Flow C）
	timeout time.Duration
}

// New 构造 Classifier。llm 为 nil 时恒走模板/正则匹配；store 为 nil 时 Correct 只
// 重路由不落 correction_log。
func New(llm domain.LLMProvider, store SessionStore, guided map[string]GuidedTemplate) *Classifier {
	return &Classifier{llm: llm, store: store, guided: guided, timeout: llmTimeout}
}

// Classify 分类 + 参数抽取。LLM 可用时优先 LLM，失败/解析失败回退模板/正则（L1）。
// 排障意图抽取不出服务名或置信度过低 → NeedsClarify=true + 候选意图（供一键纠正）。
func (c *Classifier) Classify(ctx context.Context, input, tenant string) (domain.IntentResult, error) {
	var res domain.IntentResult
	if c.llm != nil {
		if r, err := c.classifyByLLM(ctx, input); err == nil {
			res = r
		} else {
			res = c.classifyByRule(input, time.Now()) // L1 回退（不依赖 LLM）
		}
	} else {
		res = c.classifyByRule(input, time.Now())
	}

	if res.Type == domain.IntentDiagnose && res.Params.ServiceName == "" {
		res.NeedsClarify = true
	}
	if res.Confidence < clarifyThreshold {
		res.NeedsClarify = true
	}
	if res.NeedsClarify {
		res.Alternatives = c.ruleAlternatives(input, time.Now())
	}
	return res, nil
}

// classifyByLLM LLM 分类路径；解析结构化输出失败返回错误 → 触发 L1 回退。
func (c *Classifier) classifyByLLM(ctx context.Context, input string) (domain.IntentResult, error) {
	cctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := domain.LLMRequest{
		System:   "你是运维聊天机器人的意图分类器。将用户输入分类为 diagnose（排障诊断）/ resource（资源查询）/ out_of_scope（超出能力：开发咨询/提单/拓扑等）。只输出 JSON：{\"type\":\"...\",\"serviceName\":\"...\",\"metric\":\"...\",\"timeWindowMinutes\":60,\"confidence\":0.9}",
		Messages: []domain.LLMMessage{{Role: "user", Text: input}},
	}
	text, err := c.llm.Complete(cctx, req)
	if err != nil {
		return domain.IntentResult{}, err
	}

	var out llmIntentOutput
	if err := json.Unmarshal([]byte(extractJSON(text)), &out); err != nil {
		return domain.IntentResult{}, fmt.Errorf("intent: parse llm output: %w", err)
	}
	t := domain.IntentType(out.Type)
	if !t.IsValid() {
		return domain.IntentResult{}, fmt.Errorf("intent: llm returned invalid type %q", out.Type)
	}
	return out.toResult(time.Now()), nil
}

// llmIntentOutput LLM 分类结构化输出。
type llmIntentOutput struct {
	Type              string   `json:"type"`
	ServiceName       string   `json:"serviceName"`
	Metric            string   `json:"metric"`
	TimeWindowMinutes int      `json:"timeWindowMinutes"`
	Confidence        float64  `json:"confidence"`
	LogTypes          []string `json:"logTypes"`
}

// toResult 映射为 IntentResult。
func (o llmIntentOutput) toResult(now time.Time) domain.IntentResult {
	start, end := now.Add(-time.Hour), now
	if o.TimeWindowMinutes > 0 {
		start = now.Add(-time.Duration(o.TimeWindowMinutes) * time.Minute)
	}
	return domain.IntentResult{
		Type:       domain.IntentType(o.Type),
		Confidence: o.Confidence,
		Params: domain.DiagnoseParams{
			ServiceName: o.ServiceName,
			Metric:      o.Metric,
			Timeframe:   domain.Timeframe{StartTime: start, EndTime: end},
			LogTypes:    o.LogTypes,
		},
	}
}

// classifyByRule 模板/正则意图分类（L1 降级，确定性）。
func (c *Classifier) classifyByRule(input string, now time.Time) domain.IntentResult {
	// 1. 超能力意图：不调用底座、不猜测执行。
	if detectOutOfScope(input) {
		return domain.IntentResult{Type: domain.IntentOutOfScope, Confidence: 1}
	}
	// 2. 资源查询。
	if detectResource(input) {
		start, end, _ := extractTimeframe(input, now)
		return domain.IntentResult{
			Type:       domain.IntentResource,
			Confidence: 1,
			Params: domain.DiagnoseParams{
				ServiceName: extractServiceName(input),
				Timeframe:   domain.Timeframe{StartTime: start, EndTime: end},
			},
		}
	}
	// 3. 排障诊断（缺省）。
	metric := extractMetric(input)
	conf := 0.4
	if svc := extractServiceName(input); svc != "" {
		conf = 0.8
	}
	start, end, _ := extractTimeframe(input, now)
	return domain.IntentResult{
		Type:       domain.IntentDiagnose,
		Confidence: conf,
		Params: domain.DiagnoseParams{
			ServiceName: extractServiceName(input),
			Metric:      metric,
			Timeframe:   domain.Timeframe{StartTime: start, EndTime: end},
		},
	}
}

// ruleAlternatives 依据规则给出候选意图（供澄清时一键纠正）。
func (c *Classifier) ruleAlternatives(input string, now time.Time) []domain.IntentCandidate {
	var out []domain.IntentCandidate
	if svc := extractServiceName(input); svc != "" {
		start, end, _ := extractTimeframe(input, now)
		out = append(out, domain.IntentCandidate{
			Type:       domain.IntentDiagnose,
			Confidence: 0.8,
			Params: domain.DiagnoseParams{
				ServiceName: svc,
				Metric:      extractMetric(input),
				Timeframe:   domain.Timeframe{StartTime: start, EndTime: end},
			},
		})
	}
	out = append(out, domain.IntentCandidate{Type: domain.IntentResource, Confidence: 0.5})
	return out
}

// Correct 依据用户纠正重路由（api-handbook §2）：target+params 与 message 二选一
// （XOR）。target+params 为一键纠正；message 为自然语言纠正（服务端重新分类）。
// 附带回填 correction_log（store 非 nil 时）。
func (c *Classifier) Correct(ctx context.Context, sessionID string, target domain.IntentType, params domain.DiagnoseParams, message string) (domain.IntentResult, error) {
	hasTarget := target != ""
	hasMessage := strings.TrimSpace(message) != ""
	if hasTarget == hasMessage {
		return domain.IntentResult{}, fmt.Errorf("intent: Correct requires exactly one of target+params or message (XOR)")
	}

	var res domain.IntentResult
	if hasTarget {
		if !target.IsValid() {
			return domain.IntentResult{}, fmt.Errorf("intent: invalid target intent %q", target)
		}
		res = domain.IntentResult{Type: target, Params: params, Confidence: 1}
	} else {
		var err error
		res, err = c.Classify(ctx, message, "")
		if err != nil {
			return domain.IntentResult{}, err
		}
	}

	if c.store != nil {
		if err := c.appendCorrection(ctx, sessionID, res, message); err != nil {
			return domain.IntentResult{}, err
		}
	}
	return res, nil
}

// appendCorrection 读取会话旧意图并追加纠正记录（误分类 + 纠正）。
func (c *Classifier) appendCorrection(ctx context.Context, sessionID string, res domain.IntentResult, message string) error {
	sess, err := c.store.GetByID(ctx, "", sessionID)
	if err != nil {
		return err
	}
	rec := domain.CorrectionRecord{
		From:    sess.IntentType,
		To:      res.Type,
		At:      time.Now(),
		Message: message,
	}
	return c.store.AppendCorrection(ctx, "", sessionID, rec)
}

// Guided 返回某超能力意图类别的引导话术（Flow C）。
func (c *Classifier) Guided(category string) (GuidedTemplate, bool) {
	g, ok := c.guided[category]
	return g, ok
}

// extractJSON 从 LLM 输出中提取 JSON 片段（容忍 markdown 代码围栏与前后缀文本）。
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 {
		s = s[:j+1]
	}
	return s
}
