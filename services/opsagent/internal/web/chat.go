package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/baseclient"
	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"github.com/Havens-blog/opsagent/internal/orchestrator"
)

// orientRequestContext 为编排链挂载请求级上下文：
//  1. 用户会话令牌原文（baseclient 透传底座 eiam 鉴权，契约 B1）；
//  2. 会话用户名（correction_log 的 By 字段审计）。
//
// 鉴权结论仍只认 EiamAuth 中间件注入的会话主体，此处仅透传面向下游的数据。
func orientRequestContext(c *gin.Context) context.Context {
	ctx := baseclient.WithUserToken(c.Request.Context(), middleware.RawToken(c))
	return domain.WithUserAttr(ctx, middleware.GetUsername(c))
}

// chat 响应 type 枚举（api-handbook §1 严格一致）。
const (
	chatTypeReport         = "report"
	chatTypePresetEntries  = "preset_entries"
	chatTypeDegradedNotice = "degraded_notice"
	chatTypeGuided         = "guided"
	chatTypeClarify        = "clarify"
)

// maxMessageLen 自然语言提问长度上限（过长 → 400）。
const maxMessageLen = 4096

// defaultChatBudget 同步编排硬预算（tech-design ≤25s）。
const defaultChatBudget = 25 * time.Second

// chatStore 会话存取窄接口（会话创建 + 归属租户读取）。
type chatStore interface {
	Insert(ctx context.Context, s *domain.Session) error
	GetByID(ctx context.Context, tenantID, id string) (domain.Session, error)
}

// chatRunner 编排执行窄接口（Orchestrator.RunChat 适配）。
type chatRunner interface {
	RunChat(ctx context.Context, session *domain.Session, input string) (*domain.Diagnosis, error)
}

// guidedResolver 超能力意图引导话术解析（settings.guided_templates 按租户，Flow C）。
type guidedResolver func(tenantID, category string) (string, bool)

// presetsLoader L2 预置查询目录（settings.preset_queries，按租户读取）。
type presetsLoader func(tenantID string) []domain.PresetQuery

// chatRequest POST /chat 请求体。
type chatRequest struct {
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

// correctRequest POST /chat/:sessionId/correct 请求体（targetIntent+params XOR message）。
type correctRequest struct {
	TargetIntent string                `json:"targetIntent"`
	Params       domain.DiagnoseParams `json:"params"`
	Message      string                `json:"message"`
}

// chatData chat 响应 data（api-handbook §1 同形）。
type chatData struct {
	SessionID    string                   `json:"sessionId"`
	Diagnosis    *domain.Diagnosis        `json:"diagnosis"`
	Report       string                   `json:"report"`
	Type         string                   `json:"type"`
	NeedsClarify bool                     `json:"needsClarify"`
	Candidates   []domain.IntentCandidate `json:"candidates,omitempty"`
	DegradeLevel int                      `json:"degradeLevel"`
	Presets      []domain.PresetQuery     `json:"presets,omitempty"`
}

// ChatHandler 对话排障 handler（chat + correct）。
type ChatHandler struct {
	store   chatStore
	intent  domain.IntentClassifier
	runner  chatRunner
	degrade *orchestrator.DegradationController
	llmOK   func() bool
	presets presetsLoader
	guided  guidedResolver
	budget  time.Duration
}

// NewChatHandler 构造 ChatHandler（budget≤0 取默认 25s）。
func NewChatHandler(store chatStore, intent domain.IntentClassifier, runner chatRunner,
	degrade *orchestrator.DegradationController, llmOK func() bool,
	presets presetsLoader, guided guidedResolver, budget time.Duration) *ChatHandler {
	if budget <= 0 {
		budget = defaultChatBudget
	}
	return &ChatHandler{
		store: store, intent: intent, runner: runner, degrade: degrade,
		llmOK: llmOK, presets: presets, guided: guided, budget: budget,
	}
}

// RegisterRoutes 注册 chat + correct 路由（Module 接口）。
func (h *ChatHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/chat", h.Chat)
	group.POST("/chat/:sessionId/correct", h.Correct)
}

// Chat 发起对话排障：创建会话 + 意图分类 + 编排。
func (h *ChatHandler) Chat(c *gin.Context) {
	c.Request = c.Request.WithContext(orientRequestContext(c))
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, &middleware.ParamInvalidError{Msg: "请求体非法"})
		return
	}
	msg := strings.TrimSpace(middleware.SanitizeUntrustedText(req.Message))
	if msg == "" || len(msg) > maxMessageLen {
		FailErr(c, &middleware.ParamInvalidError{Msg: "message 为空或超长"})
		return
	}

	tenant := middleware.TenantIDString(c)
	sess := &domain.Session{
		TenantID: tenant,
		Type:     domain.SessionTypeChat,
		UserID:   strconv.FormatInt(middleware.GetUid(c), 10),
		Query:    msg,
		Status:   domain.SessionStatusRunning,
	}
	if err := h.store.Insert(c.Request.Context(), sess); err != nil {
		FailErr(c, &middleware.PersistFailedError{Err: err})
		return
	}

	intentRes, err := h.intent.Classify(c.Request.Context(), msg, tenant)
	if err != nil {
		FailErr(c, err)
		return
	}
	h.execute(c, sess, msg, intentRes)
}

// Correct 会话内纠正（一键纠正 XOR 自然语言纠正重路由）。
func (h *ChatHandler) Correct(c *gin.Context) {
	c.Request = c.Request.WithContext(orientRequestContext(c))
	sessionID := c.Param("sessionId")
	tenant := middleware.TenantIDString(c)

	// 归属租户校验：跨租户/不存在 → mongo.ErrNoDocuments → 404。
	sess, err := h.store.GetByID(c.Request.Context(), tenant, sessionID)
	if err != nil {
		FailErr(c, err)
		return
	}

	var req correctRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, &middleware.ParamInvalidError{Msg: "请求体非法"})
		return
	}
	hasTarget := req.TargetIntent != ""
	hasMessage := strings.TrimSpace(req.Message) != ""
	if hasTarget == hasMessage {
		FailErr(c, &middleware.ParamInvalidError{Msg: "targetIntent 与 message 必须二选一"})
		return
	}

	res, err := h.intent.Correct(c.Request.Context(), tenant, sessionID, domain.IntentType(req.TargetIntent), req.Params, req.Message)
	if err != nil {
		FailErr(c, err)
		return
	}

	input := strings.TrimSpace(req.Message)
	if hasTarget {
		input = req.Params.ServiceName // 一键纠正：以服务名重入编排
	}
	h.execute(c, &sess, input, res)
}

// execute 统一编排路由：意图 → 引导/澄清/降级/编排。
func (h *ChatHandler) execute(c *gin.Context, sess *domain.Session, input string, intentRes domain.IntentResult) {
	// 超能力意图 → 引导式回应，不调用底座。
	if intentRes.Type == domain.IntentOutOfScope {
		h.writeGuided(c, sess, intentRes)
		return
	}
	// LLM 可用但置信度低 → 澄清追问（候选意图）。
	if intentRes.NeedsClarify && h.llmOK() {
		h.writeClarify(c, sess, intentRes)
		return
	}

	// 降级决策（LLM 可用 → L0；不可用 → L1 模板 / L2 预置入口 / L3 明示降级）。
	presets := h.presets(sess.TenantID)
	decision := h.degrade.Decide(sess.ID, h.llmOK, intentRes, presets)
	switch decision.Level {
	case orchestrator.LevelPreset:
		c.JSON(http.StatusOK, Response{Code: "0", Message: "ok", Data: chatData{
			SessionID:    sess.ID,
			Type:         chatTypePresetEntries,
			DegradeLevel: int(decision.Level),
			Presets:      presets,
		}})
		return
	case orchestrator.LevelNotice:
		c.JSON(http.StatusOK, Response{Code: "0", Message: "ok", Data: chatData{
			SessionID:    sess.ID,
			Type:         chatTypeDegradedNotice,
			DegradeLevel: int(decision.Level),
			Report:       "当前处于降级模式，暂无法提供诊断，请稍后重试或走人工通道。",
		}})
		return
	}

	// L0/L1：同步编排（预算 context；超时转 202 异步回退）。
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.budget)
	defer cancel()
	diag, err := h.runner.RunChat(ctx, sess, input)
	if ctx.Err() == context.DeadlineExceeded {
		go h.continueAsync(sess, input)
		c.JSON(http.StatusAccepted, Response{Code: "0", Message: "ok",
			Data: gin.H{"sessionId": sess.ID, "status": "running"}})
		return
	}
	if err != nil && diag == nil {
		FailErr(c, err)
		return
	}

	data := chatData{
		SessionID:    sess.ID,
		Diagnosis:    diag,
		Report:       reportBody(diag),
		Type:         chatTypeReport,
		DegradeLevel: int(decision.Level),
	}
	// 落库失败：HTTP 200 + code=ERR_PERSIST_FAILED，data.report 仍返回正文。
	if err != nil {
		c.JSON(http.StatusOK, Response{Code: middleware.CodePersistFailed, Message: "结果暂未能保存，请稍后重试", Data: data})
		return
	}
	c.JSON(http.StatusOK, Response{Code: "0", Message: "ok", Data: data})
}

// continueAsync 同步编排超预算后的后台继续（落库 + 会话状态 done）。
func (h *ChatHandler) continueAsync(sess *domain.Session, input string) {
	ctx, cancel := context.WithTimeout(context.Background(), h.budget)
	defer cancel()
	_, _ = h.runner.RunChat(ctx, sess, input)
}

// writeGuided 超能力意图引导式回应。
func (h *ChatHandler) writeGuided(c *gin.Context, sess *domain.Session, intentRes domain.IntentResult) {
	report := "这是开发咨询类问题，超出运维排障能力范围，请走对应渠道处理。"
	if h.guided != nil {
		if tmpl, ok := h.guided(sess.TenantID, string(intentRes.Type)); ok && tmpl != "" {
			report = tmpl
		}
	}
	c.JSON(http.StatusOK, Response{Code: "0", Message: "ok", Data: chatData{
		SessionID: sess.ID,
		Type:      chatTypeGuided,
		Report:    report,
	}})
}

// writeClarify 澄清追问（候选意图供一键纠正）。
func (h *ChatHandler) writeClarify(c *gin.Context, sess *domain.Session, intentRes domain.IntentResult) {
	c.JSON(http.StatusOK, Response{Code: "0", Message: "ok", Data: chatData{
		SessionID:    sess.ID,
		Type:         chatTypeClarify,
		NeedsClarify: true,
		Candidates:   intentRes.Alternatives,
		Report:       "没太确定您的排障目标，请选择或补充服务名。",
	}})
}

// reportBody 面向用户报告正文（根因 + 结论 + 建议 + 降级提示）。
func reportBody(d *domain.Diagnosis) string {
	if d == nil {
		return ""
	}
	var b strings.Builder
	if d.RootCause != "" {
		b.WriteString("根因：" + d.RootCause + "\n")
	}
	for _, c := range d.Conclusions {
		b.WriteString("· " + c.Text + "\n")
	}
	for _, dp := range d.Disposition {
		b.WriteString("建议：" + dp.Step + "（" + dp.Risk + "）\n")
	}
	if d.Degraded {
		b.WriteString("（提示：LLM 降级，以上为规则结论）\n")
	}
	return strings.TrimSpace(b.String())
}
