// opsagent 服务入口：独立编排层服务（gin + gotomicro/ego）。
//
// 启动流程：加载 config/*.yaml + env → 初始化 Redis/eiam 会话校验 →
// 装配 HTTP 服务（eiam 鉴权 + 租户强制注入）→ 监听配置端口（优雅关停）。
// 缺失必需配置项时启动即失败并报明确错误。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecodeclub/ginx/session"
	"github.com/ecodeclub/ginx/session/cookie"
	"github.com/ecodeclub/ginx/session/header"
	"github.com/ecodeclub/ginx/session/mixin"
	ginRedis "github.com/ecodeclub/ginx/session/redis"
	"github.com/gotomicro/ego/core/elog"
	goRedis "github.com/redis/go-redis/v9"

	"github.com/Havens-blog/opsagent/internal/config"
	"github.com/Havens-blog/opsagent/internal/web"
)

func main() {
	configDir := flag.String("config", "config", "配置目录（包含 *.yaml 配置文件）")
	flag.Parse()

	logger := elog.DefaultLogger

	cfg, err := config.Load(*configDir)
	if err != nil {
		// 先落结构化日志，再向 stderr 显式输出：elog 为异步缓冲写，
		// os.Exit 前不保证刷盘；启动失败的明确报错必须立即可见。
		logger.Error("加载配置失败，服务无法启动", elog.FieldErr(err))
		fmt.Fprintf(os.Stderr, "加载配置失败，服务无法启动: %v\n", err)
		os.Exit(1)
	}

	rdb := goRedis.NewClient(&goRedis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Warn("关闭 Redis 连接失败", elog.FieldErr(err))
		}
	}()

	sp := newSessionProvider(cfg, rdb, logger)

	// 组合根装配全部业务组件（DAO/编排/handler/worker），失败即终止启动。
	ap, err := buildApp(cfg, rdb, logger)
	if err != nil {
		logger.Error("装配业务组件失败，服务无法启动", elog.FieldErr(err))
		fmt.Fprintf(os.Stderr, "装配业务组件失败，服务无法启动: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		ap.worker.Stop()
		if err := ap.mongoClient.Disconnect(context.Background()); err != nil {
			logger.Warn("关闭 Mongo 连接失败", elog.FieldErr(err))
		}
	}()

	server := web.NewServer(cfg, sp, logger, ap.modules...)
	// 告警入站 webhook 挂开放组（不经 EiamAuth/RequireTenant，服务间密钥鉴权替代）。
	ap.alertModule.RegisterRoutes(server.Engine().Group("/api/v1/opsagent"))
	// 告警异步排查 worker（后台并发池）。
	ap.worker.Start(context.Background())

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      server.Engine(),
		ReadTimeout:  cfg.ReadTimeout(),
		WriteTimeout: cfg.WriteTimeout(),
	}

	go func() {
		logger.Info("opsagent HTTP 服务启动",
			elog.Int("port", cfg.Server.Port),
			elog.Int("redis_db", cfg.Redis.DB))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP 服务启动失败", elog.FieldErr(err))
			os.Exit(1)
		}
	}()

	// 优雅关停
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("收到退出信号，开始优雅关停")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("优雅关停失败", elog.FieldErr(err))
	}
	logger.Info("opsagent 已停止")
}

// newSessionProvider 构建 eiam 会话校验 Provider：
// eiam 通过同一 Redis DB 写入会话（ginx redis session），本服务只读校验。
func newSessionProvider(cfg *config.Config, rdb *goRedis.Client, logger *elog.Component) session.Provider {
	const sessionTTL = 30 * 24 * time.Hour
	sp := ginRedis.NewSessionProvider(rdb, cfg.Session.EncryptedKey, sessionTTL)

	// 会话令牌载体：Header 优先（服务间/前端均可用），Cookie 兜底（浏览器）。
	cookieC := &cookie.TokenCarrier{
		MaxAge:   int(sessionTTL.Seconds()),
		Name:     cfg.Session.CookieName,
		Secure:   true,
		HttpOnly: false,
		Domain:   cfg.Session.CookieDomain,
	}
	sp.TokenCarrier = mixin.NewTokenCarrier(header.NewTokenCarrier(), cookieC)

	logger.Info("eiam 会话校验 Provider 初始化完成",
		elog.String("cookie_name", cfg.Session.CookieName),
		elog.Int("redis_db", cfg.Redis.DB))
	return sp
}
