// Package web HTTP 服务装配：gin 引擎、响应信封、健康检查与模块路由挂载。
//
// 装配模式与 e-cam-service 一致：模块实现 Module 接口，在 NewServer 中挂载到
// 统一的用户路由组（eiam 鉴权 + 租户强制注入中间件之后）。
package web

import (
	"net/http"

	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"

	"github.com/Havens-blog/opsagent/internal/config"
	"github.com/Havens-blog/opsagent/internal/middleware"
)

// Module 功能模块路由注册接口（对齐 e-cam-service 的 module.go 模式）。
type Module interface {
	// RegisterRoutes 将模块路由注册到已挂载鉴权中间件的用户路由组。
	RegisterRoutes(group *gin.RouterGroup)
}

// Response 统一响应信封：{code, message, data}；code=0 成功，错误时为 ERR_* 字符串码。
type Response struct {
	Code    any    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Server opsagent HTTP 服务。
type Server struct {
	engine *gin.Engine
	group  *gin.RouterGroup
}

// NewServer 装配 gin 引擎：健康检查（开放访问）+ 用户路由组（eiam 鉴权 + 租户注入）+ 模块挂载。
//
// 注意：告警入站 webhook（/api/v1/opsagent/alerts，服务间密钥鉴权）由后续任务
// 以独立路由组注册，不经过用户会话中间件。
func NewServer(cfg *config.Config, sp session.Provider, logger *elog.Component, modules ...Module) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	registerHealth(engine)

	group := engine.Group("/api/v1/opsagent")
	group.Use(middleware.EiamAuth(sp, cfg.Session.Whitelist, logger))
	group.Use(middleware.RequireTenant(logger))
	for _, m := range modules {
		m.RegisterRoutes(group)
	}

	return &Server{engine: engine, group: group}
}

// Engine 返回 gin 引擎（供 http.Server 使用）。
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Group 返回已挂载鉴权中间件的用户路由组（供后续任务注册模块路由）。
func (s *Server) Group() *gin.RouterGroup {
	return s.group
}

// registerHealth 注册健康检查端点（开放访问，供存活探针使用）。
func registerHealth(engine *gin.Engine) {
	engine.GET("/api/v1/opsagent/health", func(c *gin.Context) {
		OK(c, gin.H{"status": "up", "service": "opsagent"})
	})
}

// OK 成功响应：HTTP 200 + 信封 code=0。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// Fail 错误响应：按 api-handbook 错误码表映射 HTTP 状态与 ERR_* 业务码。
func Fail(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Response{Code: code, Message: message})
}

// FailErr 统一错误响应：经 middleware.MapError 映射内部错误 → HTTP 状态 + ERR_* 业务码，
// 跨租户/不存在一律 404 不泄露存在性。
func FailErr(c *gin.Context, err error) {
	status, code, message := middleware.MapError(err)
	Fail(c, status, code, message)
}
