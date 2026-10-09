// 组合根（composition root）：装配 Mongo → DAO → LLM/意图/风险/底座客户端 →
// 编排层 → 6 个 web handler + 告警 worker，返回可注册模块与后台任务。
//
// 原 main.go 的 opsagentModules() 仅有 P0 空壳（返回 nil）；本文件补齐「模块装配」——
// 服务级「可跑」的最后一环（feature haven-opsagent 的集成补全）。
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Havens-blog/opsagent/internal/baseclient"
	"github.com/Havens-blog/opsagent/internal/config"
	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/llm"
	"github.com/Havens-blog/opsagent/internal/orchestrator"
	"github.com/Havens-blog/opsagent/internal/orchestrator/intent"
	"github.com/Havens-blog/opsagent/internal/repository/dao"
	"github.com/Havens-blog/opsagent/internal/risk"
	"github.com/Havens-blog/opsagent/internal/web"
	"github.com/Havens-blog/opsagent/internal/worker"
	"github.com/Havens-blog/opsagent/pkg/mongox"
	goRedis "github.com/redis/go-redis/v9"
)

// syncBudget 同步编排硬预算（tech-design ≤25s，与 chat 默认一致）。
const syncBudget = 25 * time.Second

// mongoDialTimeout Mongo 连接/建库超时。
const mongoDialTimeout = 10 * time.Second

// app 组合根产物：用户态模块 + 开放态(告警 webhook)模块 + 后台 worker + 资源句柄。
type app struct {
	modules     []web.Module
	alertModule web.Module
	worker      *worker.Worker
	mongoClient *mongo.Client
}

// buildApp 装配全部业务组件并返回可注册的模块集合。失败时返回明确错误（上层据此终止启动）。
func buildApp(cfg *config.Config, rdb *goRedis.Client, logger *elog.Component) (*app, error) {
	// —— Mongo ——
	mctx, cancel := context.WithTimeout(context.Background(), mongoDialTimeout)
	defer cancel()
	mongoClient, err := mongo.Connect(mctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, fmt.Errorf("连接 Mongo(%s) 失败: %w", cfg.Mongo.URI, err)
	}
	if err := mongoClient.Ping(mctx, nil); err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, fmt.Errorf("Ping Mongo(%s) 失败: %w", cfg.Mongo.URI, err)
	}
	db := mongox.NewMongo(mongoClient, cfg.Mongo.DBName)

	// —— DAO ——
	sessionDAO := dao.NewSessionDAO(db)
	diagnosisDAO := dao.NewDiagnosisDAO(db)
	citationDAO := dao.NewCitationDAO(db)
	riskEntryDAO := dao.NewRiskEntryDAO(db)
	alertDAO := dao.NewAlertDAO(db)
	settingsDAO := dao.NewSettingsDAO(db)
	llmUsageDAO := dao.NewLLMUsageDAO(db)
	compensationDAO := dao.NewCompensationDAO(db)

	// 索引初始化（幂等；失败不阻塞启动——索引亦可用 schema.mongo.js 预建，此处兜底）。
	if err := initIndexes(context.Background(),
		sessionDAO, diagnosisDAO, citationDAO, riskEntryDAO,
		alertDAO, settingsDAO, llmUsageDAO, compensationDAO,
	); err != nil {
		logger.Warn("初始化 Mongo 索引失败（降级继续启动）", elog.FieldErr(err))
	}

	// —— LLM（未配置 apiKey → disabledProvider，编排走三级降级，不阻塞启动）——
	provider := llm.NewProvider(cfg.LLM)
	guarded := llm.NewGuardedLLM(provider, time.Duration(cfg.LLM.TimeoutSeconds)*time.Second)

	// —— 意图分类（store=sessionDAO：Correct 落 correction_log；内置引导话术兜底 Flow C）——
	classifier := intent.New(provider, sessionDAO, intent.DefaultGuidedTemplates())

	// —— 风险注册表（确定性查表，Go 代码非 DB）——
	registry := risk.NewRegistry()

	// —— 底座客户端 ——
	webhooks := map[string]string{}
	for _, ch := range cfg.Notify.Channels {
		if ch.Enabled && ch.Webhook != "" {
			webhooks[ch.Channel] = ch.Webhook
		}
	}
	baseClient := baseclient.New(baseclient.Config{
		LogqueryBaseURL: cfg.Base.LogqueryBaseURL,
		McpServerPath:   cfg.Base.McpServerPath,
		ChannelWebhooks: webhooks,
		SessionProvider: nil, // VerifyToken 未被编排链路调用；会话校验由 EiamAuth 中间件承担
		OnHealthAlert: func(module string) {
			logger.Error("底座契约漂移触发健康告警", elog.String("module", module))
		},
	})

	// —— 编排层 ——
	orc := orchestrator.New(orchestrator.Deps{
		Intent:      classifier,
		Base:        baseClient,
		LLM:         guarded,
		Sessions:    sessionDAO,
		Diagnoses:   diagnosisDAO,
		Citations:   citationDAO,
		RiskEntries: riskEntryDAO,
		Alerts:      alertDAO,
		Timeout:     syncBudget,
	})

	// —— 三级降级控制器 ——
	degrade := orchestrator.NewDegradationController(0, cfg.Thresholds.ClarifyConfidence, nil)

	// —— 告警 worker（指纹去重 + 1h 预算 + 并发池）——
	alertWorker := worker.New(
		worker.NewRedisDeduper(rdb),
		worker.NewRedisBudget(rdb),
		alertDAO,
		alertInvestigator{orc: orc},
		worker.Config{
			Concurrency:   cfg.Thresholds.WorkerConcurrency,
			QueueCapacity: cfg.Thresholds.QueueCapacity,
			DedupWindow:   time.Duration(cfg.Thresholds.DedupWindowMinutes) * time.Minute,
			BudgetLimit:   int64(cfg.LLM.BudgetPerHour),
		},
		func(err error) { logger.Error("告警 worker 后台任务失败", elog.FieldErr(err)) },
	)

	// —— web handlers ——
	llmOK := func() bool { return guarded.IsAvailable(context.Background()) }

	// L2 预置目录：租户级 settings.preset_queries，缺失回退全局（tenant=""）。
	presetsLoader := func(tenantID string) []domain.PresetQuery {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var out []domain.PresetQuery
		if err := settingsDAO.Get(ctx, tenantID, domain.ScopePresetQueries, &out); err != nil || len(out) == 0 {
			_ = settingsDAO.Get(ctx, "", domain.ScopePresetQueries, &out)
		}
		return out
	}
	// Flow C 引导话术：租户级 settings.guided_templates → 全局 → 内置默认。
	guidedResolver := func(tenantID, category string) (string, bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		lookup := func(tid string) (string, bool) {
			var m map[string]intent.GuidedTemplate
			if err := settingsDAO.Get(ctx, tid, domain.ScopeGuidedTemplates, &m); err != nil {
				return "", false
			}
			g, ok := m[category]
			return g.Template, ok && g.Template != ""
		}
		if tmpl, ok := lookup(tenantID); ok {
			return tmpl, true
		}
		if tmpl, ok := lookup(""); ok {
			return tmpl, true
		}
		if g, ok := intent.DefaultGuidedTemplates()[category]; ok {
			return g.Template, true
		}
		return "", false
	}

	chatHandler := web.NewChatHandler(
		sessionDAO, classifier, orc, degrade, llmOK,
		presetsLoader,
		guidedResolver,
		syncBudget,
	)
	riskHandler := web.NewRiskCenterHandler(riskEntryDAO)
	diagnosisHandler := web.NewDiagnosisHandler(diagnosisDAO, citationDAO, riskEntryDAO, sessionDAO)
	settingsHandler := web.NewSettingsHandler(settingsDAO, registry, isAdmin)
	obsHandler := web.NewObservabilityHandler(web.NewInProcObserver(
		cfg.Thresholds.WorkerConcurrency, cfg.Thresholds.QueueCapacity, int64(cfg.LLM.BudgetPerHour)))
	alertHandler := web.NewAlertWebhookHandler([]string{cfg.Notify.AlertWebhookKey}, alertWorker)

	return &app{
		modules:     []web.Module{chatHandler, riskHandler, diagnosisHandler, settingsHandler, obsHandler},
		alertModule: alertHandler,
		worker:      alertWorker,
		mongoClient: mongoClient,
	}, nil
}

// alertInvestigator 适配 Orchestrator.RunAlert → worker.Investigator。
// degrade 由编排层依据 GuardedLLM 可用性自持，此处忽略该标志。
type alertInvestigator struct{ orc *orchestrator.Orchestrator }

func (a alertInvestigator) Investigate(ctx context.Context, alert domain.AlertEvent, _ bool) (*domain.RiskEntry, error) {
	return a.orc.RunAlert(ctx, alert)
}

// isAdmin 从 eiam 会话 claims 判定管理员（settings PUT 权限；契约 B3 Data.is_admin）。
func isAdmin(c *gin.Context) bool {
	v, ok := c.Get(session.CtxSessionKey)
	if !ok {
		return false
	}
	sess, ok := v.(session.Session)
	if !ok {
		return false
	}
	return sess.Claims().Data["is_admin"] == "true"
}

// indexInitializer 集合索引初始化窄接口（各 DAO 均满足）。
type indexInitializer interface {
	InitIndexes(ctx context.Context) error
}

// initIndexes 依次初始化各集合索引，首个失败即返回。
func initIndexes(ctx context.Context, inits ...indexInitializer) error {
	for _, in := range inits {
		if err := in.InitIndexes(ctx); err != nil {
			return err
		}
	}
	return nil
}
