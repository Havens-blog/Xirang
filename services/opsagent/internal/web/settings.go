package web

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"github.com/Havens-blog/opsagent/internal/orchestrator/intent"
)

// settingsStore 系统配置存取窄接口（六类 scope 泛化读写）。
type settingsStore interface {
	Get(ctx context.Context, tenantID string, scope domain.SettingScope, out any) error
	Upsert(ctx context.Context, tenantID string, scope domain.SettingScope, value any) error
}

// settingsData GET /settings 响应（六类配置）。
type settingsData struct {
	Datasources     []domain.Datasource              `json:"datasources"`
	LLMProviders    []domain.LLMProviderConfig       `json:"llmProviders"`
	NotifyChannels  []domain.NotifyChannel           `json:"notifyChannels"`
	RiskWhitelist   []domain.RiskWhitelistEntry      `json:"riskWhitelist"`
	PresetQueries   []domain.PresetQuery             `json:"presetQueries"`
	GuidedTemplates map[string]intent.GuidedTemplate `json:"guidedTemplates"`
}

// llmProviderUpdate PUT 的 llm_providers 条目（apiKey 只写，GET 返回掩码）。
type llmProviderUpdate struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
	Model   string `json:"model"`
	APIKey  string `json:"apiKey"`
}

// settingsUpdate PUT 请求（指针字段区分「未提供」与「提供空」；datasources 只读忽略）。
type settingsUpdate struct {
	LLMProviders    *[]llmProviderUpdate              `json:"llmProviders"`
	NotifyChannels  *[]domain.NotifyChannel           `json:"notifyChannels"`
	RiskWhitelist   *[]domain.RiskWhitelistEntry      `json:"riskWhitelist"`
	PresetQueries   *[]domain.PresetQuery             `json:"presetQueries"`
	GuidedTemplates *map[string]intent.GuidedTemplate `json:"guidedTemplates"`
}

// SettingsHandler 系统配置 handler（GET/PUT /settings；写需管理员）。
type SettingsHandler struct {
	store    settingsStore
	registry domain.RiskRegistry
	isAdmin  func(*gin.Context) bool
}

// NewSettingsHandler 构造。isAdmin 为 nil 时写操作一律 403（运维只读）。
func NewSettingsHandler(store settingsStore, registry domain.RiskRegistry, isAdmin func(*gin.Context) bool) *SettingsHandler {
	return &SettingsHandler{store: store, registry: registry, isAdmin: isAdmin}
}

// RegisterRoutes 注册 settings 路由（Module 接口）。
func (h *SettingsHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/settings", h.Get)
	group.PUT("/settings", h.Put)
}

// Get 读六类系统配置（各 scope 缺失返回空值）。
func (h *SettingsHandler) Get(c *gin.Context) {
	tenant := middleware.TenantIDString(c)
	ctx := c.Request.Context()

	var llm []domain.LLMProviderConfig
	_ = h.store.Get(ctx, tenant, domain.ScopeLLM, &llm)
	var notify []domain.NotifyChannel
	_ = h.store.Get(ctx, tenant, domain.ScopeNotify, &notify)
	var wl []domain.RiskWhitelistEntry
	_ = h.store.Get(ctx, tenant, domain.ScopeRiskWhitelist, &wl)
	var presets []domain.PresetQuery
	_ = h.store.Get(ctx, tenant, domain.ScopePresetQueries, &presets)
	var guided map[string]intent.GuidedTemplate
	_ = h.store.Get(ctx, tenant, domain.ScopeGuidedTemplates, &guided)
	var ds []domain.Datasource
	_ = h.store.Get(ctx, tenant, domain.ScopeDatasource, &ds)

	OK(c, settingsData{
		Datasources:     ds,
		LLMProviders:    llm,
		NotifyChannels:  notify,
		RiskWhitelist:   wl,
		PresetQueries:   presets,
		GuidedTemplates: guided,
	})
}

// Put 按字段部分更新（仅提供字段覆盖；datasources 只读忽略）。
func (h *SettingsHandler) Put(c *gin.Context) {
	if h.isAdmin == nil || !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, Response{Code: middleware.CodeTenantDenied, Message: "仅管理员可修改系统配置"})
		return
	}
	tenant := middleware.TenantIDString(c)
	ctx := c.Request.Context()

	var req settingsUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, &middleware.ParamInvalidError{Msg: "请求体非法"})
		return
	}

	if req.RiskWhitelist != nil {
		if err := h.validateWhitelist(*req.RiskWhitelist); err != nil {
			FailErr(c, err)
			return
		}
		if err := h.store.Upsert(ctx, tenant, domain.ScopeRiskWhitelist, *req.RiskWhitelist); err != nil {
			FailErr(c, &middleware.PersistFailedError{Err: err})
			return
		}
	}
	if req.PresetQueries != nil {
		if err := validatePresets(*req.PresetQueries); err != nil {
			FailErr(c, err)
			return
		}
		// id 由服务端生成：客户端传入 id 忽略。
		presets := make([]domain.PresetQuery, 0, len(*req.PresetQueries))
		for _, p := range *req.PresetQueries {
			p.ID = ""
			presets = append(presets, p)
		}
		if err := h.store.Upsert(ctx, tenant, domain.ScopePresetQueries, presets); err != nil {
			FailErr(c, &middleware.PersistFailedError{Err: err})
			return
		}
	}
	if req.NotifyChannels != nil {
		if err := h.store.Upsert(ctx, tenant, domain.ScopeNotify, *req.NotifyChannels); err != nil {
			FailErr(c, &middleware.PersistFailedError{Err: err})
			return
		}
	}
	if req.GuidedTemplates != nil {
		if err := h.store.Upsert(ctx, tenant, domain.ScopeGuidedTemplates, *req.GuidedTemplates); err != nil {
			FailErr(c, &middleware.PersistFailedError{Err: err})
			return
		}
	}
	if req.LLMProviders != nil {
		providers := make([]domain.LLMProviderConfig, 0, len(*req.LLMProviders))
		for _, p := range *req.LLMProviders {
			providers = append(providers, domain.LLMProviderConfig{
				Name:      p.Name,
				Default:   p.Default,
				Model:     p.Model,
				KeyMasked: maskAPIKey(p.APIKey), // apiKey 只写不读：落库/回显仅掩码
			})
		}
		if err := h.store.Upsert(ctx, tenant, domain.ScopeLLM, providers); err != nil {
			FailErr(c, &middleware.PersistFailedError{Err: err})
			return
		}
	}

	OK(c, nil)
}

// validateWhitelist 校验白名单条目：工具须命中风险注册表且档位非 high，否则整体拒绝。
func (h *SettingsHandler) validateWhitelist(entries []domain.RiskWhitelistEntry) error {
	if h.registry == nil {
		return &middleware.ParamInvalidError{Msg: "风险注册表未配置"}
	}
	for _, e := range entries {
		meta, err := h.registry.Classify(e.Tool)
		if err != nil {
			return &middleware.ParamInvalidError{Msg: "工具未注册: " + e.Tool}
		}
		if meta.RiskLevel == domain.RiskHigh {
			return &middleware.ParamInvalidError{Msg: "高危工具不允许入白名单: " + e.Tool}
		}
	}
	return nil
}

// validatePresets 校验预置查询目录：每条 serviceName 非空、时间窗 ≤24h，否则整体拒绝。
func validatePresets(presets []domain.PresetQuery) error {
	for _, p := range presets {
		if strings.TrimSpace(p.Params.ServiceName) == "" {
			return &middleware.ParamInvalidError{Msg: "preset 缺少 serviceName"}
		}
		if !p.Params.Timeframe.EndTime.IsZero() && p.Params.Timeframe.EndTime.Sub(p.Params.Timeframe.StartTime) > 24*time.Hour {
			return &middleware.ParamInvalidError{Msg: "preset 时间窗超过 24 小时"}
		}
	}
	return nil
}

// maskAPIKey API Key 掩码：sk-**** + 末 4 位（只写不读语义，GET 永返掩码）。
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 4 {
		return "sk-****"
	}
	return "sk-****" + key[len(key)-4:]
}
