package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
)

// RequireTenant 要求会话已选定租户（tenant != 0）。
//
// eiam 以 tenant_id=0 表示等待选择租户的临时凭证；0 不得视为「全部租户」通配，
// 否则会导致跨租户数据泄露。租户缺失属于权限问题，返回 403（与 e-cam-service
// 一致；401 + ERR_TENANT_DENIED 仅用于未登录/令牌无效）。
func RequireTenant(logger *elog.Component) gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetTenantID(c) == 0 {
			if logger != nil {
				logger.Warn("会话未选定租户，拒绝访问",
					elog.String("path", c.Request.URL.Path))
			}
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "当前会话未选定租户空间",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetTenantID 从上下文取出租户 ID；0 表示会话未选定租户。
func GetTenantID(c *gin.Context) int64 {
	if v, exists := c.Get(TenantIDKey); exists {
		if tid, ok := v.(int64); ok {
			return tid
		}
	}
	return 0
}

// GetUid 从上下文取出租户内用户 ID。
func GetUid(c *gin.Context) int64 {
	if v, exists := c.Get(CtxUidKey); exists {
		if uid, ok := v.(int64); ok {
			return uid
		}
	}
	return 0
}

// GetUsername 从上下文取出用户名。
func GetUsername(c *gin.Context) string {
	if v, exists := c.Get(CtxUsernameKey); exists {
		if username, ok := v.(string); ok {
			return username
		}
	}
	return ""
}

// TenantIDString 便捷方法：租户 ID 字符串形式（日志/落库键使用）。
func TenantIDString(c *gin.Context) string {
	return strconv.FormatInt(GetTenantID(c), 10)
}
