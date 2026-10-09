package web

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
)

// maxHistoryWindow 历史检索时间窗上限（≤24h，可配）。
const maxHistoryWindow = 24 * time.Hour

// diagnosisReader 诊断读取窄接口。
type diagnosisReader interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Diagnosis, error)
	GetBySession(ctx context.Context, tenantID, sessionID string) (*domain.Diagnosis, error)
}

// citationReader 引用读取窄接口。
type citationReader interface {
	ListByDiagnosis(ctx context.Context, tenantID, diagnosisID string) ([]domain.Citation, error)
}

// riskViewer 风险条目窄状态迁移窄接口（诊断详情打开触发 pending_view→viewed）。
type riskViewer interface {
	MarkViewed(ctx context.Context, tenantID, id, diagnosisID, by string) (*domain.RiskEntry, bool, error)
}

// historyReader 历史会话检索窄接口。
type historyReader interface {
	GetByID(ctx context.Context, tenantID, id string) (domain.Session, error)
	ListByTime(ctx context.Context, tenantID string, start, end time.Time, limit, offset int) ([]domain.Session, error)
	ListByService(ctx context.Context, tenantID, serviceName string, limit, offset int) ([]domain.Session, error)
}

// historyData 历史检索响应 data（api-handbook §7 同形）。
type historyData struct {
	Items []domain.SessionSummary `json:"items"`
	Total int                     `json:"total"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
}

// DiagnosisHandler 诊断详情 + 历史回溯 handler。
type DiagnosisHandler struct {
	diag   diagnosisReader
	cites  citationReader
	risk   riskViewer
	hist   historyReader
	maxWin time.Duration
}

// NewDiagnosisHandler 构造。
func NewDiagnosisHandler(diag diagnosisReader, cites citationReader, risk riskViewer, hist historyReader) *DiagnosisHandler {
	return &DiagnosisHandler{diag: diag, cites: cites, risk: risk, hist: hist, maxWin: maxHistoryWindow}
}

// RegisterRoutes 注册详情 + 历史路由（Module 接口）。
func (h *DiagnosisHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/diagnosis/:id", h.Detail)
	group.GET("/history", h.History)
}

// Detail 诊断详情：完整 Diagnosis（根因/结论/引用/调用链/处置），带 riskEntryId 惰性已读。
func (h *DiagnosisHandler) Detail(c *gin.Context) {
	tenant := middleware.TenantIDString(c)
	id := c.Param("id")

	diag, err := h.diag.GetByID(c.Request.Context(), tenant, id)
	if err != nil {
		FailErr(c, err) // mongo.ErrNoDocuments → 404（跨租户/不存在不泄露）
		return
	}
	// 引用独立集合关联组装。
	if h.cites != nil {
		if cites, err := h.cites.ListByDiagnosis(c.Request.Context(), tenant, id); err == nil {
			diag.Citations = cites
		}
	}

	// 惰性自动已读：仅携带 riskEntryId 且同租户关联本诊断时才迁移，无副作用。
	if rid := c.Query("riskEntryId"); rid != "" && h.risk != nil {
		_, _, _ = h.risk.MarkViewed(c.Request.Context(), tenant, rid, id, strconv.FormatInt(middleware.GetUid(c), 10))
	}

	OK(c, diag)
}

// History 历史会话检索（时间/服务名筛选 + 精确 sessionId）。
func (h *DiagnosisHandler) History(c *gin.Context) {
	tenant := middleware.TenantIDString(c)
	page := parsePositiveInt(c.Query("page"), 1)
	limit := parsePositiveInt(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// 精确会话（202 异步轮询路径）。
	if sid := c.Query("sessionId"); sid != "" {
		sess, err := h.hist.GetByID(c.Request.Context(), tenant, sid)
		if err != nil {
			FailErr(c, err)
			return
		}
		sum := toSummary(sess)
		if h.diag != nil {
			if d, err := h.diag.GetBySession(c.Request.Context(), tenant, sid); err == nil {
				sum.DiagnosisID = d.ID
			}
		}
		OK(c, historyData{Items: []domain.SessionSummary{sum}, Total: 1, Page: 1, Limit: 1})
		return
	}

	start := parseRFC3339OrZero(c.Query("startTime"))
	end := parseRFC3339OrZero(c.Query("endTime"))
	if !start.IsZero() || !end.IsZero() {
		if start.IsZero() {
			start = time.Now().Add(-h.maxWin)
		}
		if end.IsZero() {
			end = time.Now()
		}
		if end.Sub(start) > h.maxWin || end.Before(start) {
			FailErr(c, &middleware.ParamInvalidError{Msg: "时间窗不得超过 24 小时"})
			return
		}
	}

	var (
		sessions []domain.Session
		err      error
	)
	if svc := c.Query("serviceName"); svc != "" {
		sessions, err = h.hist.ListByService(c.Request.Context(), tenant, svc, limit, offset)
	} else {
		sessions, err = h.hist.ListByTime(c.Request.Context(), tenant, start, end, limit, offset)
	}
	if err != nil {
		FailErr(c, err)
		return
	}

	items := make([]domain.SessionSummary, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, toSummary(s))
	}
	// P1 注：ListByTime/ListByService 未返回 total，以本页条数近似（≤24h 窗口通常单页）。
	OK(c, historyData{Items: items, Total: len(items), Page: page, Limit: limit})
}

// toSummary Session → SessionSummary（历史列表项）。
func toSummary(s domain.Session) domain.SessionSummary {
	return domain.SessionSummary{
		ID:          s.ID,
		Type:        string(s.Type),
		Query:       s.Query,
		ServiceName: s.ServiceName,
		IntentType:  string(s.IntentType),
		Status:      string(s.Status),
		CreatedAt:   s.CreatedAt,
	}
}

// parseRFC3339OrZero 解析时间，非法/空返零值。
func parseRFC3339OrZero(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
