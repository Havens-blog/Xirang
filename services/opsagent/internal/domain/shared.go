// Package domain 跨模块共享的领域类型（Ground Truth）。
//
// 本包是 Shared Types 的唯一权威定义处（tech-design › Shared Types），
// json tag 与 api-handbook「Data Contracts」逐字段一致，前端 TS 类型据此生成；
// 存储字段名（bson tag）与 er-diagram 各集合字段一致。
// 后续任务不得复制重定义，一律 import 本包。
package domain

import "time"

// ---- 意图（tech-design › Interface 1）----

// Timeframe 时间窗（RFC3339；End-Start ≤ 24h，可配）。
type Timeframe struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// DiagnoseParams 意图抽取出的排障参数（IntentClassifier 输出 / L2 预置查询预填）。
type DiagnoseParams struct {
	ServiceName string    `json:"serviceName"`        // 必填（抽取不出 → 触发 L2 澄清）
	Metric      string    `json:"metric,omitempty"`   // 错误率/延迟/QPS…，可空
	Timeframe   Timeframe `json:"timeframe"`          // 缺省最近 1h
	LogTypes    []string  `json:"logTypes,omitempty"` // cdn|waf|lb 子集；空 = 全部
}

// IntentCandidate 备选意图（供一键纠正）。
type IntentCandidate struct {
	Type       IntentType     `json:"type"`
	Confidence float64        `json:"confidence"` // 0~1
	Params     DiagnoseParams `json:"params"`
}

// IntentResult 意图分类结果。
type IntentResult struct {
	Type         IntentType
	Confidence   float64           // 0~1；低于阈值触发澄清
	Params       DiagnoseParams    // 抽取的服务名/时间窗/指标
	Alternatives []IntentCandidate // 备选意图（供一键纠正）
	NeedsClarify bool              // 置信度低，需追问
}

// ---- 底座契约（tech-design › Interface 2）----

// LogQueryRequest logquery 日志联邦查询请求（契约 B1）。
type LogQueryRequest struct {
	Tenant    string    `json:"tenant"`
	LogType   string    `json:"logType"`  // cdn|waf|lb
	Cloud     string    `json:"cloud"`    // aliyun|huawei|aws|tencent
	Resource  string    `json:"resource"` // 服务名/资产标识
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Limit     int       `json:"limit"` // 默认 1000，超出截断
}

// LogQueryResult logquery 查询结果（契约 B1）。
type LogQueryResult struct {
	Total     int64      `json:"total"`
	Truncated bool       `json:"truncated"`
	Entries   []LogEntry `json:"entries"`
	RiskScore *float64   `json:"riskScore,omitempty"` // diagnose 规则引擎
}

// LogEntry 联邦日志条目（契约 B1 最小稳定子集；精确字段名 M1 冻结）。
type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Cloud     string         `json:"cloud"`    // aliyun|huawei|aws|tencent
	LogType   string         `json:"logType"`  // cdn|waf|lb
	Resource  string         `json:"resource"` // 服务名/资产标识
	Level     string         `json:"level"`    // error|warn|info
	Message   string         `json:"message"`  // 进入 LLM 前必须过注入清洗（Security ②）
	Fields    map[string]any `json:"fields,omitempty"`
}

// AssetRecord 资产查询结果（契约 B2；tenant 必填供引用校验）。
type AssetRecord struct {
	AssetID string         `json:"assetId"`
	Type    string         `json:"type"`   // ecs|rds|redis|eip|nas|oss
	Tenant  string         `json:"tenant"` // 引用校验：与查询租户不一致即丢弃
	Name    string         `json:"name"`
	Region  string         `json:"region,omitempty"`
	Status  string         `json:"status,omitempty"`
	Raw     map[string]any `json:"raw,omitempty"`
}

// Identity eiam 令牌验证结果（契约 B3 verifyToken 返回）。
type Identity struct {
	UserID string   `json:"userId"`
	Tenant string   `json:"tenant"`
	Roles  []string `json:"roles"`
}

// Notification 通知载荷（契约 B4；只携带摘要与入口链接，不含跨租户明细）。
type Notification struct {
	Tenant      string `json:"tenant"`
	Channel     string `json:"channel"` // dingtalk|feishu|wecom|email
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	DiagLinkURL string `json:"diagLinkUrl"` // 诊断入口链接（带鉴权跳转）
}

// NotifyResult 通知发送结果（SC-2 成功率埋点；risk_entries.notified 同形）。
type NotifyResult struct {
	Channel   string    `json:"channel"`
	OK        bool      `json:"ok"`
	Attempts  int       `json:"attempts"` // 含重试，上限 3
	LastError string    `json:"lastError,omitempty"`
	SentAt    time.Time `json:"sentAt,omitempty"`
}

// ---- 编排（tech-design › Interface 3）----

// AlertEvent 告警入站事件（webhook 契约 → Orchestrator.RunAlert 输入）。
type AlertEvent struct {
	Tenant      string         `json:"tenant"` // 服务端校验，与 webhook 密钥租户一致
	ServiceName string         `json:"serviceName"`
	Metric      string         `json:"metric"`
	Level       string         `json:"level"` // 参与指纹
	Timeframe   Timeframe      `json:"timeframe"`
	Raw         map[string]any `json:"raw"` // 告警原文（落 alerts 留档：正常记账与溢出丢弃均保留）
}

// ---- LLM（tech-design › Interface 4）----

// LLMMessage LLM 消息（不可信文本须包裹分隔符）。
type LLMMessage struct {
	Role string `json:"role"` // system|user|assistant
	Text string `json:"text"`
}

// LLMRequest LLM 请求。
type LLMRequest struct {
	System    string        `json:"system"`
	Messages  []LLMMessage  `json:"messages"`
	MaxTokens int           `json:"maxTokens,omitempty"`
	Timeout   time.Duration `json:"-"` // 单次超时 > 12s 触发降级谓词
}

// ---- 三级降级（tech-design › Error Handling › LLM 三级降级设计）----

// PresetQuery 预置查询条目（L2 响应；settings.preset_queries 持久化，管理员可配）。
type PresetQuery struct {
	ID     string         `json:"id"`
	Label  string         `json:"label"`  // 展示名，如「order-service 近1h错误率」
	Params DiagnoseParams `json:"params"` // 预填完整参数，点击即确定性重入流水线
}

// ---- 系统配置（GET/PUT /opsagent/settings 的 data 元素）----

// Datasource 数据源连接状态。
type Datasource struct {
	Name      string    `json:"name"` // cdn|waf|lb|metric|asset
	Cloud     string    `json:"cloud,omitempty"`
	OK        bool      `json:"ok"` // 最近一次连通性探测结果
	CheckedAt time.Time `json:"checkedAt"`
}

// LLMProviderConfig LLM 提供商配置（避免与 LLMProvider 接口重名；Key 只写不读，GET 返回掩码）。
type LLMProviderConfig struct {
	Name      string `json:"name"` // qwen|deepseek|openai
	Default   bool   `json:"default"`
	Model     string `json:"model"`
	KeyMasked string `json:"keyMasked,omitempty"` // API Key 只写不读，GET 返回掩码（sk-****last4）
}

// NotifyChannel 通知渠道开关。
type NotifyChannel struct {
	Channel string `json:"channel"` // dingtalk|feishu|wecom|email
	Enabled bool   `json:"enabled"`
}

// RiskWhitelistEntry 低危白名单条目（tool 必须命中风险注册表，未注册 → 拒绝保存 400）。
type RiskWhitelistEntry struct {
	Tool      string    `json:"tool"`
	RiskLevel RiskLevel `json:"riskLevel"` // read|low（high 不允许入白名单）
}
