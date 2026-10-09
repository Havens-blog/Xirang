// Package middleware opsagent HTTP 中间件：eiam 会话鉴权、租户强制注入、
// prompt injection 清洗挂载点。
//
// 与 e-cam-service/internal/shared/middleware 同一套约定：
// eiam 通过 ginx session 将会话写入共享 Redis，本服务以 session.Provider 校验，
// 租户一律从会话 claims 派生并强制注入上下文，客户端传值丢弃（DF005）。
package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
)

const (
	// CtxUidKey 用户 ID 在上下文中的键。
	CtxUidKey = "uid"
	// CtxUsernameKey 用户名在上下文中的键。
	CtxUsernameKey = "username"
	// TenantIDKey 租户 ID 在上下文中的键。
	// 唯一写入方是 EiamAuth（取自 eiam 会话 claims）；客户端不得指定租户，
	// X-Tenant-ID 头与请求体中的 tenant 字段一律丢弃。
	TenantIDKey = "tenant_id"
)

// CodeTenantDenied 鉴权失败错误码（未登录 / 令牌无效 → 401）。
const CodeTenantDenied = "ERR_TENANT_DENIED"

// EiamAuth eiam 会话鉴权中间件。
//
// 校验失败返回 401 + ERR_TENANT_DENIED；成功后将 uid/username/tenant 强制注入
// 请求上下文。whitelist 中的路径跳过鉴权（精确匹配或 /path/* 前缀匹配）。
func EiamAuth(sp session.Provider, whitelist []string, logger *elog.Component) gin.HandlerFunc {
	return func(c *gin.Context) {
		if matchWhitelist(c.Request.URL.Path, whitelist) {
			c.Next()
			return
		}

		gCtx := &gctx.Context{Context: c}
		sess, err := sp.Get(gCtx)
		if err != nil {
			if logger != nil {
				logger.Debug("eiam 会话鉴权失败",
					elog.FieldErr(err),
					elog.String("path", c.Request.URL.Path),
				)
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    CodeTenantDenied,
				"message": "认证失败：未登录或令牌无效",
			})
			c.Abort()
			return
		}

		claims := sess.Claims()
		c.Set(CtxUidKey, claims.Uid)
		if username, ok := claims.Data["username"]; ok {
			c.Set(CtxUsernameKey, username)
		}

		// 租户强制注入：只取自会话 claims（eiam 侧以字符串写入），
		// 客户端任何渠道传入的租户值一律丢弃。
		// 缺失或解析失败时不注入，由 RequireTenant 拒绝。
		if raw, ok := claims.Data["tenant_id"]; ok && raw != "" {
			if tid, err := strconv.ParseInt(raw, 10, 64); err == nil {
				c.Set(TenantIDKey, tid)
			} else if logger != nil {
				logger.Warn("eiam 会话中的 tenant_id 无法解析为 int64",
					elog.String("raw", raw),
					elog.String("path", c.Request.URL.Path))
			}
		}

		c.Set(session.CtxSessionKey, sess)
		c.Next()
	}
}

// matchWhitelist 白名单路径匹配：/path/* 为前缀匹配，其余为精确匹配。
func matchWhitelist(path string, whitelist []string) bool {
	for _, pattern := range whitelist {
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}
		if path == pattern {
			return true
		}
	}
	return false
}

// RawToken 提取请求携带的 eiam 会话令牌原文（Authorization: Bearer x，或裸令牌）。
// 供编排层向底座（e-cam-service 契约 B1）透传调用方凭证；鉴权本身由 EiamAuth 完成，
// 本函数不做校验、只取原文。告警 webhook 等服务间请求不携带用户令牌，返回空串。
func RawToken(c *gin.Context) string {
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth == "" {
		return ""
	}
	const bearer = "Bearer "
	if strings.HasPrefix(auth, bearer) {
		return strings.TrimSpace(auth[len(bearer):])
	}
	return auth
}
