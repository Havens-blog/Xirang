// Package config opsagent 配置加载。
//
// 加载顺序：
//  1. 读取配置目录下全部 *.yaml / *.yml 文件（按文件名排序）并合并；
//  2. 环境变量覆盖敏感项（OPSAGENT_ALERT_WEBHOOK_KEY / OPSAGENT_LLM_API_KEY /
//     OPSAGENT_REDIS_PASSWORD / OPSAGENT_SESSION_ENCRYPTED_KEY）；
//  3. 校验必需项，缺失时启动即失败并报明确错误。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// EnvPrefix opsagent 环境变量统一前缀。
const EnvPrefix = "OPSAGENT"

// Config opsagent 全量配置。
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Session    SessionConfig    `mapstructure:"session"`
	Mongo      MongoConfig      `mapstructure:"mongo"`
	Base       BaseConfig       `mapstructure:"base"`
	LLM        LLMConfig        `mapstructure:"llm"`
	Notify     NotifyConfig     `mapstructure:"notify"`
	Thresholds ThresholdsConfig `mapstructure:"thresholds"`
	Whitelist  WhitelistConfig  `mapstructure:"whitelist"`
}

// ServerConfig HTTP 服务配置（超时与 tech-design 同步编排延迟预算对齐，默认 30s）。
type ServerConfig struct {
	Port                int `mapstructure:"port"`
	ReadTimeoutSeconds  int `mapstructure:"read_timeout_seconds"`
	WriteTimeoutSeconds int `mapstructure:"write_timeout_seconds"`
}

// RedisConfig 与 eiam/e-cam-service 共享的 Redis（会话校验、去重、LLM 预算）。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// SessionConfig eiam 会话校验配置（与 eiam 写入同一 Redis DB 的会话对齐）。
type SessionConfig struct {
	EncryptedKey string   `mapstructure:"encrypted_key"`
	CookieName   string   `mapstructure:"cookie_name"`
	CookieDomain string   `mapstructure:"cookie_domain"`
	Whitelist    []string `mapstructure:"whitelist"`
}

// MongoConfig MongoDB 连接配置（DB: opsagent，8 集合）。
type MongoConfig struct {
	URI    string `mapstructure:"uri"`
	DBName string `mapstructure:"dbname"`
}

// BaseConfig 底座客户端配置（地址与本地资产 MCP 服务路径）。
type BaseConfig struct {
	LogqueryBaseURL string `mapstructure:"logquery_base_url"` // e-cam-service 前缀（契约 B1）
	McpServerPath   string `mapstructure:"mcp_server_path"`   // 资产查询 mcp-server 路径（契约 B2，stdio）
}

// LLMConfig LLM Provider 配置（未配置 API Key 时走降级路径，不阻塞启动）。
type LLMConfig struct {
	Provider       string `mapstructure:"provider"`
	Model          string `mapstructure:"model"`
	APIKey         string `mapstructure:"api_key"`
	BaseURL        string `mapstructure:"base_url"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
	BudgetPerHour  int    `mapstructure:"budget_per_hour"`
}

// NotifyChannelConfig 单个通知渠道配置。
type NotifyChannelConfig struct {
	Channel string `mapstructure:"channel"`
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
}

// NotifyConfig 通知渠道 + 告警入站 webhook 服务间密钥。
type NotifyConfig struct {
	AlertWebhookKey string                `mapstructure:"alert_webhook_key"`
	MaxAttempts     int                   `mapstructure:"max_attempts"`
	Channels        []NotifyChannelConfig `mapstructure:"channels"`
}

// ThresholdsConfig 阈值配置（worker 并发/队列/去重窗口等，与技术设计对齐）。
type ThresholdsConfig struct {
	WorkerConcurrency  int     `mapstructure:"worker_concurrency"`
	QueueCapacity      int     `mapstructure:"queue_capacity"`
	DedupWindowMinutes int     `mapstructure:"dedup_window_minutes"`
	MaxTimeframeHours  int     `mapstructure:"max_timeframe_hours"`
	ClarifyConfidence  float64 `mapstructure:"clarify_confidence"`
}

// RiskWhitelistEntry 低危白名单条目。
type RiskWhitelistEntry struct {
	Tool      string `mapstructure:"tool"`
	RiskLevel string `mapstructure:"risk_level"`
}

// WhitelistConfig 风险白名单（运行时覆盖需命中 internal/risk 注册表）。
type WhitelistConfig struct {
	Entries []RiskWhitelistEntry `mapstructure:"entries"`
}

// ReadTimeout HTTP 读超时。
func (c *Config) ReadTimeout() time.Duration {
	return time.Duration(c.Server.ReadTimeoutSeconds) * time.Second
}

// WriteTimeout HTTP 写超时。
func (c *Config) WriteTimeout() time.Duration {
	return time.Duration(c.Server.WriteTimeoutSeconds) * time.Second
}

// Load 从配置目录加载并合并全部 YAML 配置，应用环境变量覆盖与默认值后校验。
func Load(dir string) (*Config, error) {
	if dir == "" {
		dir = "config"
	}

	v := viper.New()
	if err := mergeDir(v, dir); err != nil {
		return nil, err
	}

	applyEnvOverrides(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	applyDefaults(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// mergeDir 合并目录下全部 YAML 配置文件（按文件名排序，后加载覆盖先加载）。
func mergeDir(v *viper.Viper, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取配置目录 %s 失败: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, name)
		}
	}
	if len(files) == 0 {
		return fmt.Errorf("配置目录 %s 中未找到任何 YAML 配置文件", dir)
	}

	sort.Strings(files)
	for _, name := range files {
		path := filepath.Join(dir, name)
		v.SetConfigFile(path)
		if err := v.MergeInConfig(); err != nil {
			return fmt.Errorf("加载配置文件 %s 失败: %w", path, err)
		}
	}
	return nil
}

// envOverrides 环境变量 → 配置键覆盖表（敏感项，不入库不入码）。
var envOverrides = []struct {
	Env string
	Key string
}{
	{"OPSAGENT_ALERT_WEBHOOK_KEY", "notify.alert_webhook_key"},
	{"OPSAGENT_LLM_API_KEY", "llm.api_key"},
	{"OPSAGENT_REDIS_PASSWORD", "redis.password"},
	{"OPSAGENT_SESSION_ENCRYPTED_KEY", "session.encrypted_key"},
}

// applyEnvOverrides 用环境变量覆盖 YAML 中的敏感项（仅当环境变量非空时生效）。
func applyEnvOverrides(v *viper.Viper) {
	for _, o := range envOverrides {
		if val := os.Getenv(o.Env); val != "" {
			v.Set(o.Key, val)
		}
	}
}

// applyDefaults 填充未配置项的默认值（与技术设计对齐）。
func applyDefaults(cfg *Config) {
	if cfg.Server.ReadTimeoutSeconds <= 0 {
		cfg.Server.ReadTimeoutSeconds = 30
	}
	if cfg.Server.WriteTimeoutSeconds <= 0 {
		cfg.Server.WriteTimeoutSeconds = 30
	}
	if cfg.LLM.TimeoutSeconds <= 0 {
		cfg.LLM.TimeoutSeconds = 12 // >12s 触发降级谓词
	}
	if cfg.LLM.BudgetPerHour <= 0 {
		cfg.LLM.BudgetPerHour = 100 // 滚动 1h LLM 预算
	}
	if cfg.Notify.MaxAttempts <= 0 {
		cfg.Notify.MaxAttempts = 3
	}
	if cfg.Thresholds.WorkerConcurrency <= 0 {
		cfg.Thresholds.WorkerConcurrency = 5
	}
	if cfg.Thresholds.QueueCapacity <= 0 {
		cfg.Thresholds.QueueCapacity = 100
	}
	if cfg.Thresholds.DedupWindowMinutes <= 0 {
		cfg.Thresholds.DedupWindowMinutes = 10
	}
	if cfg.Thresholds.MaxTimeframeHours <= 0 {
		cfg.Thresholds.MaxTimeframeHours = 24
	}
	if cfg.Thresholds.ClarifyConfidence <= 0 {
		cfg.Thresholds.ClarifyConfidence = 0.6
	}
}

// validate 校验必需配置项；缺失时返回一次性列出全部缺失项的明确错误。
func validate(cfg *Config) error {
	type required struct {
		key  string
		fail bool
	}

	portOK := cfg.Server.Port > 0 && cfg.Server.Port <= 65535
	checks := []required{
		{"server.port", !portOK},
		{"redis.addr", cfg.Redis.Addr == ""},
		{"session.encrypted_key", cfg.Session.EncryptedKey == ""},
		{"session.cookie_name", cfg.Session.CookieName == ""},
		{"session.cookie_domain", cfg.Session.CookieDomain == ""},
		{"mongo.uri", cfg.Mongo.URI == ""},
		{"mongo.dbname", cfg.Mongo.DBName == ""},
		{"notify.alert_webhook_key", cfg.Notify.AlertWebhookKey == ""},
	}

	var missing []string
	for _, c := range checks {
		if c.fail {
			missing = append(missing, c.key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少必需配置项: %s（可通过 config/*.yaml 或对应环境变量提供）", strings.Join(missing, ", "))
	}
	return nil
}
