package config

import (
	"path/filepath"
	"strings"
	"testing"
)

// writeConfigFiles 在临时目录写入配置文件集合，返回目录路径。
func writeConfigFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		full := filepath.Join(dir, name)
		if err := writeFile(full, content); err != nil {
			t.Fatalf("写入配置文件 %s 失败: %v", full, err)
		}
	}
	return dir
}

// validBase 返回一份包含全部必需项的合法配置。
const validBase = `
server:
  port: 8081
redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 1
session:
  encrypted_key: "unit-test-session-key"
  cookie_name: "opsagent-token"
  cookie_domain: "localhost"
  whitelist:
    - /api/v1/opsagent/health
notify:
  alert_webhook_key: "unit-test-webhook-key"
  channels:
    - channel: dingtalk
      enabled: true
llm:
  provider: qwen
  model: qwen-plus
  api_key: "sk-unit-test"
thresholds:
  worker_concurrency: 5
  queue_capacity: 100
whitelist:
  entries:
    - tool: query_ecs
      risk_level: read
`

func TestLoad_ParsesAllSections(t *testing.T) {
	dir := writeConfigFiles(t, map[string]string{"config.yaml": validBase})

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}

	if cfg.Server.Port != 8081 {
		t.Errorf("server.port = %d, 期望 8081", cfg.Server.Port)
	}
	if cfg.Redis.Addr != "127.0.0.1:6379" || cfg.Redis.DB != 1 {
		t.Errorf("redis 配置解析错误: %+v", cfg.Redis)
	}
	if cfg.Session.EncryptedKey != "unit-test-session-key" || cfg.Session.CookieName != "opsagent-token" {
		t.Errorf("session 配置解析错误: %+v", cfg.Session)
	}
	if len(cfg.Session.Whitelist) != 1 || cfg.Session.Whitelist[0] != "/api/v1/opsagent/health" {
		t.Errorf("session 白名单解析错误: %v", cfg.Session.Whitelist)
	}
	if cfg.Notify.AlertWebhookKey != "unit-test-webhook-key" {
		t.Errorf("notify.alert_webhook_key 解析错误: %q", cfg.Notify.AlertWebhookKey)
	}
	if len(cfg.Notify.Channels) != 1 || cfg.Notify.Channels[0].Channel != "dingtalk" || !cfg.Notify.Channels[0].Enabled {
		t.Errorf("notify.channels 解析错误: %+v", cfg.Notify.Channels)
	}
	if cfg.LLM.Provider != "qwen" || cfg.LLM.APIKey != "sk-unit-test" {
		t.Errorf("llm 配置解析错误: %+v", cfg.LLM)
	}
	if len(cfg.Whitelist.Entries) != 1 || cfg.Whitelist.Entries[0].Tool != "query_ecs" || cfg.Whitelist.Entries[0].RiskLevel != "read" {
		t.Errorf("risk whitelist 解析错误: %+v", cfg.Whitelist.Entries)
	}
}

func TestLoad_EnvOverridesSensitiveItems(t *testing.T) {
	dir := writeConfigFiles(t, map[string]string{"config.yaml": validBase})

	t.Setenv("OPSAGENT_ALERT_WEBHOOK_KEY", "env-webhook-key")
	t.Setenv("OPSAGENT_LLM_API_KEY", "env-llm-key")
	t.Setenv("OPSAGENT_REDIS_PASSWORD", "env-redis-pass")
	t.Setenv("OPSAGENT_SESSION_ENCRYPTED_KEY", "env-session-key")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}

	if cfg.Notify.AlertWebhookKey != "env-webhook-key" {
		t.Errorf("alert_webhook_key 未被环境变量覆盖: %q", cfg.Notify.AlertWebhookKey)
	}
	if cfg.LLM.APIKey != "env-llm-key" {
		t.Errorf("llm.api_key 未被环境变量覆盖: %q", cfg.LLM.APIKey)
	}
	if cfg.Redis.Password != "env-redis-pass" {
		t.Errorf("redis.password 未被环境变量覆盖: %q", cfg.Redis.Password)
	}
	if cfg.Session.EncryptedKey != "env-session-key" {
		t.Errorf("session.encrypted_key 未被环境变量覆盖: %q", cfg.Session.EncryptedKey)
	}
}

func TestLoad_MissingRequiredFailsWithClearError(t *testing.T) {
	missing := `
server:
  port: 8081
redis:
  addr: "127.0.0.1:6379"
`
	dir := writeConfigFiles(t, map[string]string{"config.yaml": missing})

	_, err := Load(dir)
	if err == nil {
		t.Fatal("缺失必需项时应返回错误，实际为 nil")
	}
	for _, key := range []string{"session.encrypted_key", "session.cookie_name", "session.cookie_domain", "notify.alert_webhook_key"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("错误信息应包含缺失项 %q，实际: %v", key, err)
		}
	}
}

func TestLoad_InvalidPortFails(t *testing.T) {
	invalid := strings.Replace(validBase, "port: 8081", "port: 99999", 1)
	dir := writeConfigFiles(t, map[string]string{"config.yaml": invalid})

	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "server.port") {
		t.Fatalf("非法端口应报 server.port 错误，实际: %v", err)
	}
}

func TestLoad_NoYAMLFilesFails(t *testing.T) {
	dir := writeConfigFiles(t, map[string]string{"readme.txt": "not yaml"})

	_, err := Load(dir)
	if err == nil {
		t.Fatal("目录中无 YAML 文件时应返回错误，实际为 nil")
	}
}

func TestLoad_MissingDirFails(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "not-exist"))
	if err == nil {
		t.Fatal("配置目录不存在时应返回错误，实际为 nil")
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	noThresholds := `
server:
  port: 8081
redis:
  addr: "127.0.0.1:6379"
session:
  encrypted_key: "k"
  cookie_name: "c"
  cookie_domain: "d"
notify:
  alert_webhook_key: "w"
`
	dir := writeConfigFiles(t, map[string]string{"config.yaml": noThresholds})

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}

	if cfg.Thresholds.WorkerConcurrency != 5 {
		t.Errorf("worker_concurrency 默认值 = %d, 期望 5", cfg.Thresholds.WorkerConcurrency)
	}
	if cfg.Thresholds.QueueCapacity != 100 {
		t.Errorf("queue_capacity 默认值 = %d, 期望 100", cfg.Thresholds.QueueCapacity)
	}
	if cfg.Thresholds.DedupWindowMinutes != 10 {
		t.Errorf("dedup_window_minutes 默认值 = %d, 期望 10", cfg.Thresholds.DedupWindowMinutes)
	}
	if cfg.Thresholds.MaxTimeframeHours != 24 {
		t.Errorf("max_timeframe_hours 默认值 = %d, 期望 24", cfg.Thresholds.MaxTimeframeHours)
	}
	if cfg.Thresholds.ClarifyConfidence != 0.6 {
		t.Errorf("clarify_confidence 默认值 = %v, 期望 0.6", cfg.Thresholds.ClarifyConfidence)
	}
	if cfg.LLM.TimeoutSeconds != 12 {
		t.Errorf("llm.timeout_seconds 默认值 = %d, 期望 12", cfg.LLM.TimeoutSeconds)
	}
	if cfg.LLM.BudgetPerHour != 100 {
		t.Errorf("llm.budget_per_hour 默认值 = %d, 期望 100", cfg.LLM.BudgetPerHour)
	}
	if cfg.Notify.MaxAttempts != 3 {
		t.Errorf("notify.max_attempts 默认值 = %d, 期望 3", cfg.Notify.MaxAttempts)
	}
	if cfg.Server.ReadTimeoutSeconds != 30 || cfg.Server.WriteTimeoutSeconds != 30 {
		t.Errorf("server 超时默认值 = %d/%d, 期望 30/30", cfg.Server.ReadTimeoutSeconds, cfg.Server.WriteTimeoutSeconds)
	}
}

func TestLoad_MergesMultipleYAMLFiles(t *testing.T) {
	base := strings.Replace(validBase, "port: 8081", "port: 8090", 1)
	dir := writeConfigFiles(t, map[string]string{
		"config.yaml": base,
		"llm.yaml":    "llm:\n  provider: deepseek\n",
	})

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}
	if cfg.LLM.Provider != "deepseek" {
		t.Errorf("多文件合并后 llm.provider = %q, 期望 deepseek", cfg.LLM.Provider)
	}
	if cfg.Server.Port != 8090 {
		t.Errorf("多文件合并后 server.port = %d, 期望 8090", cfg.Server.Port)
	}
}
