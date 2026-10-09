package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"github.com/Havens-blog/opsagent/internal/worker"
)

// webhookKeyHeader 告警入站 webhook 服务间密钥请求头（api-handbook §9）。
const webhookKeyHeader = "X-Opsagent-Webhook-Key"

// alertEnqueuer 告警入队窄接口（worker.Enqueue 适配）。
type alertEnqueuer interface {
	Enqueue(ctx context.Context, alert domain.AlertEvent) worker.EnqueueResult
}

// AlertWebhookHandler 告警入站 webhook（服务间密钥鉴权 + 入队异步排查 + 溢出语义）。
//
// 挂载在独立开放路由组（不经 EiamAuth 用户会话中间件），由服务间密钥鉴权替代。
// 双密钥并存滚动轮换：keys 可含多个有效密钥（灰度期新旧并存，之后下线旧密钥）。
type AlertWebhookHandler struct {
	keys    []string
	enqueue alertEnqueuer
}

// NewAlertWebhookHandler 构造（keys 为有效 webhook 密钥列表，env 注入不入码）。
func NewAlertWebhookHandler(keys []string, enqueue alertEnqueuer) *AlertWebhookHandler {
	return &AlertWebhookHandler{keys: keys, enqueue: enqueue}
}

// RegisterRoutes 注册告警入站路由（Module 接口；应挂载于无 EiamAuth 的开放组）。
func (h *AlertWebhookHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/alerts", h.Alert)
}

// Alert 告警入站：密钥鉴权 → tenant 校验 → 入队异步排查。
func (h *AlertWebhookHandler) Alert(c *gin.Context) {
	key := c.GetHeader(webhookKeyHeader)
	if !h.validKey(key) {
		c.JSON(http.StatusUnauthorized, Response{Code: middleware.CodeTenantDenied, Message: "webhook 密钥无效"})
		return
	}

	var alert domain.AlertEvent
	if err := c.ShouldBindJSON(&alert); err != nil {
		FailErr(c, &middleware.ParamInvalidError{Msg: "请求体非法"})
		return
	}
	if alert.Tenant == "" || alert.ServiceName == "" || alert.Metric == "" || alert.Level == "" {
		FailErr(c, &middleware.ParamInvalidError{Msg: "tenant/serviceName/metric/level 必填"})
		return
	}

	res := h.enqueue.Enqueue(c.Request.Context(), alert)
	// 半成功语义：队列满溢出丢弃低级别仍 202，dropped=true 标记（调用方无需重试）。
	c.JSON(http.StatusAccepted, Response{Code: "0", Message: "ok", Data: gin.H{
		"accepted": res.Accepted,
		"dropped":  res.Dropped,
	}})
}

// validKey 校验请求头密钥是否命中任一有效密钥（双密钥轮换）。
func (h *AlertWebhookHandler) validKey(key string) bool {
	if key == "" {
		return false
	}
	for _, k := range h.keys {
		if k != "" && key == k {
			return true
		}
	}
	return false
}
