package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"github.com/Havens-blog/opsagent/internal/repository/dao"
)

// maxBatchStatus 批量标记上限（可配，默认 50）。
const maxBatchStatus = 50

// riskStore 风险中心存取窄接口（列表/单条 CAS/批量 CAS/聚合统计）。
type riskStore interface {
	List(ctx context.Context, tenantID string, filter dao.RiskEntryFilter, limit, offset int) ([]domain.RiskEntry, int64, error)
	UpdateStatus(ctx context.Context, tenantID, id string, status domain.RiskEntryStatus, expectedVersion int, by string) (*domain.RiskEntry, error)
	BatchUpdateStatus(ctx context.Context, tenantID string, status domain.RiskEntryStatus, items []dao.BatchStatusItem) (*dao.BatchStatusResult, error)
	Stats(ctx context.Context, tenantID string) (*dao.RiskStats, error)
}

// riskListData GET /risk-center 响应 data。
type riskListData struct {
	Items []domain.RiskEntry `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
	Stats *dao.RiskStats     `json:"stats"`
}

// riskStatusRequest POST /risk-center/:id/status 请求体。
type riskStatusRequest struct {
	Status          string `json:"status"`
	ExpectedVersion int    `json:"expectedVersion"`
}

// batchStatusRequest POST /risk-center/batch-status 请求体。
type batchStatusRequest struct {
	Status string `json:"status"`
	Items  []struct {
		ID              string `json:"id"`
		ExpectedVersion int    `json:"expectedVersion"`
	} `json:"items"`
}

// batchStatusData 批量标记响应 data（成功/失败分别聚合）。
type batchStatusData struct {
	Succeeded []batchSuccess `json:"succeeded"`
	Failed    []batchFailure `json:"failed"`
}

type batchSuccess struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type batchFailure struct {
	ID             string `json:"id"`
	Reason         string `json:"reason"`
	ReasonCode     string `json:"reasonCode"`
	CurrentVersion int    `json:"currentVersion,omitempty"`
}

// RiskCenterHandler 风险中心 handler（list + 单条/批量 CAS + stats）。
type RiskCenterHandler struct {
	store riskStore
}

// NewRiskCenterHandler 构造。
func NewRiskCenterHandler(store riskStore) *RiskCenterHandler {
	return &RiskCenterHandler{store: store}
}

// RegisterRoutes 注册风险中心路由（Module 接口）。
func (h *RiskCenterHandler) RegisterRoutes(group *gin.RouterGroup) {
	g := group.Group("/risk-center")
	g.GET("", h.List)
	g.POST("/:id/status", h.UpdateStatus)
	g.POST("/batch-status", h.BatchUpdateStatus)
}

// List 风险中心列表（租户/时间/服务名/状态/severity 筛选 + 分页 + stats 聚合）。
func (h *RiskCenterHandler) List(c *gin.Context) {
	tenant := middleware.TenantIDString(c)

	filter := dao.RiskEntryFilter{
		Status:      domain.RiskEntryStatus(c.Query("status")),
		Severity:    domain.Severity(c.Query("severity")),
		ServiceName: c.Query("serviceName"),
	}
	if s := c.Query("startTime"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			FailErr(c, &middleware.ParamInvalidError{Msg: "startTime 非法"})
			return
		}
		filter.StartTime = t
	}
	if s := c.Query("endTime"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			FailErr(c, &middleware.ParamInvalidError{Msg: "endTime 非法"})
			return
		}
		filter.EndTime = t
	}

	page := parsePositiveInt(c.Query("page"), 1)
	limit := parsePositiveInt(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100
	}

	items, total, err := h.store.List(c.Request.Context(), tenant, filter, limit, (page-1)*limit)
	if err != nil {
		FailErr(c, err)
		return
	}
	stats, err := h.store.Stats(c.Request.Context(), tenant)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, riskListData{Items: items, Total: total, Page: page, Limit: limit, Stats: stats})
}

// UpdateStatus 单条标记状态（CAS 乐观锁）。
func (h *RiskCenterHandler) UpdateStatus(c *gin.Context) {
	tenant := middleware.TenantIDString(c)
	id := c.Param("id")

	var req riskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, &middleware.ParamInvalidError{Msg: "请求体非法"})
		return
	}
	status := domain.RiskEntryStatus(req.Status)
	if status != domain.RiskStatusViewed && status != domain.RiskStatusDone {
		FailErr(c, &middleware.ParamInvalidError{Msg: "status 仅支持 viewed|done"})
		return
	}

	entry, err := h.store.UpdateStatus(c.Request.Context(), tenant, id, status, req.ExpectedVersion, strconv.FormatInt(middleware.GetUid(c), 10))
	if err != nil {
		var conflict *dao.ConflictError
		if errors.As(err, &conflict) {
			c.JSON(http.StatusConflict, Response{Code: middleware.CodeConflict, Message: "已被其他值班更新，请刷新后重试", Data: conflict.Current})
			return
		}
		FailErr(c, err)
		return
	}
	OK(c, entry)
}

// BatchUpdateStatus 批量逐条 CAS 标记（成功/失败分别聚合）。
func (h *RiskCenterHandler) BatchUpdateStatus(c *gin.Context) {
	tenant := middleware.TenantIDString(c)

	var req batchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, &middleware.ParamInvalidError{Msg: "请求体非法"})
		return
	}
	status := domain.RiskEntryStatus(req.Status)
	if status != domain.RiskStatusViewed && status != domain.RiskStatusDone {
		FailErr(c, &middleware.ParamInvalidError{Msg: "status 仅支持 viewed|done"})
		return
	}
	if len(req.Items) == 0 || len(req.Items) > maxBatchStatus {
		FailErr(c, &middleware.ParamInvalidError{Msg: "items 数量须在 1~50"})
		return
	}

	items := make([]dao.BatchStatusItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, dao.BatchStatusItem{ID: it.ID, ExpectedVersion: it.ExpectedVersion})
	}

	res, err := h.store.BatchUpdateStatus(c.Request.Context(), tenant, status, items)
	if err != nil {
		FailErr(c, err)
		return
	}

	data := batchStatusData{Succeeded: []batchSuccess{}, Failed: []batchFailure{}}
	for _, s := range res.Succeeded {
		data.Succeeded = append(data.Succeeded, batchSuccess{ID: s.ID, Version: s.Version})
	}
	for _, f := range res.Failed {
		data.Failed = append(data.Failed, batchFailure{
			ID:             f.ID,
			ReasonCode:     f.ReasonCode,
			Reason:         reasonText(f.ReasonCode),
			CurrentVersion: f.CurrentVersion,
		})
	}
	OK(c, data)
}

// reasonText 批量失败 reasonCode → 用户可读原因。
func reasonText(code string) string {
	switch code {
	case "conflict":
		return "已被其他值班更新，请刷新后重试"
	case "not_found":
		return "条目不存在或已删除"
	default:
		return "状态流转非法"
	}
}

// parsePositiveInt 解析正整数查询参数，非法/缺省取 fallback。
func parsePositiveInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
