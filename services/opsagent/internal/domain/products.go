package domain

import "time"

// ---- 编排产物结构（api-handbook › Data Contracts 同形；存储字段名见 er-diagram）----

// Diagnosis 诊断报告（RunChat 产物；diagnoses 集合主体）。
type Diagnosis struct {
	ID          string            `json:"id" bson:"_id"`
	SessionID   string            `json:"sessionId" bson:"session_id"`
	Tenant      string            `json:"tenant" bson:"tenant_id"`
	RootCause   string            `json:"rootCause" bson:"root_cause"`
	Confidence  float64           `json:"confidence" bson:"confidence"`
	Severity    Severity          `json:"severity" bson:"severity"`    // P0|P1|P2|P3
	RiskLevel   RiskLevel         `json:"riskLevel" bson:"risk_level"` // read|low|high
	Conclusions []Conclusion      `json:"conclusions" bson:"conclusions"`
	Disposition []DispositionStep `json:"disposition" bson:"disposition"` // 处置预案
	Citations   []Citation        `json:"citations" bson:"citations"`     // 数据源引用
	Degraded    bool              `json:"degraded" bson:"degraded"`       // LLM 降级标识
	Truncated   bool              `json:"truncated" bson:"truncated"`
	Trace       []TraceStep       `json:"trace" bson:"trace"` // 编排调用链
}

// RiskEntry 风险中心条目（risk_entries；CAS 乐观锁 version 创建为 1，每次成功更新 +1）。
type RiskEntry struct {
	ID            string          `json:"id" bson:"_id"`
	Tenant        string          `json:"tenant" bson:"tenant_id"`
	Fingerprint   string          `json:"fingerprint" bson:"fingerprint"`
	DiagnosisID   string          `json:"diagnosisId" bson:"diagnosis_id"`
	ServiceName   string          `json:"serviceName" bson:"service_name"`
	Severity      Severity        `json:"severity" bson:"severity"`
	Status        RiskEntryStatus `json:"status" bson:"status"` // pending_view|viewed|done
	Version       int             `json:"version" bson:"version"`
	StatusHistory []StatusChange  `json:"statusHistory" bson:"status_history"`
	Notified      NotifyResult    `json:"notified" bson:"notified"`
	CreatedAt     time.Time       `json:"createdAt" bson:"created_at"`
}

// Citation 数据源引用（结论回指；citations 集合同形）。
type Citation struct {
	SourceType string `json:"sourceType"` // log|metric|asset|alert
	SourceKey  string `json:"sourceKey"`
	Snippet    string `json:"snippet"`
}

// Conclusion 单条结论（Citation 须能回指本 Diagnosis.citations 内条目，校验失败即丢弃）。
type Conclusion struct {
	Text     string   `json:"text"`
	Citation []string `json:"citation"` // 引用 Citation.SourceKey 列表
}

// DispositionStep 处置预案步骤（P1 纯只读，Executed 恒 false）。
type DispositionStep struct {
	Step     string `json:"step"`     // 处置步骤
	Risk     string `json:"risk"`     // read|low|high
	Action   string `json:"action"`   // 建议动作
	Executed bool   `json:"executed"` // P1 恒 false
}

// StatusChange 风险条目状态流转记录（risk_entries.status_history 元素）。
type StatusChange struct {
	From string    `json:"from" bson:"from"` // 旧状态；首次流转为空串
	To   string    `json:"to" bson:"to"`     // pending_view|viewed|done
	By   string    `json:"by" bson:"by"`     // 操作人 userId；系统自动流转为 "system"
	At   time.Time `json:"at" bson:"at"`
}

// TraceStep 编排调用链步骤（orchestration_trace 元素 / Diagnosis.trace）。
type TraceStep struct {
	Step         int    `json:"step" bson:"step"`                         // 序号，1 起
	Agent        string `json:"agent" bson:"agent"`                       // coordinator|log_analyst|monitor|inspector
	Action       string `json:"action" bson:"action"`                     // intent|query|diagnose|report|notify
	Source       string `json:"source,omitempty" bson:"source,omitempty"` // 底座模块 / LLM 标识
	DurationMs   int64  `json:"durationMs" bson:"durationMs"`
	Summary      string `json:"summary" bson:"summary"`                               // 结果摘要；无数据步骤显式标注「未取得数据」
	DegradeLevel int    `json:"degradeLevel,omitempty" bson:"degradeLevel,omitempty"` // 该步降级级别（0=正常）
}

// ---- 会话（er-diagram sessions 集合；Orchestrator.RunChat 参数）----

// CorrectionRecord 误分类 + 纠正记录（sessions.correction_log 元素）。
type CorrectionRecord struct {
	From    IntentType `json:"from" bson:"from"` // 误分类意图
	To      IntentType `json:"to" bson:"to"`     // 纠正后意图
	By      string     `json:"by" bson:"by"`     // 操作人 userId
	At      time.Time  `json:"at" bson:"at"`
	Message string     `json:"message,omitempty" bson:"message,omitempty"` // 自然语言纠正原文（一键纠正为空）
}

// Session 排障会话（字段覆盖 er-diagram sessions 集合）。
type Session struct {
	ID                 string             `json:"id" bson:"_id"`
	TenantID           string             `json:"tenantId" bson:"tenant_id"`       // eiam 派生，服务端强制注入
	Type               SessionType        `json:"type" bson:"type"`                // chat|alert
	UserID             string             `json:"userId" bson:"user_id"`           // 发起人（对话）或告警来源标识
	Query              string             `json:"query" bson:"query"`              // 自然语言提问原文（alert 为触发摘要）
	ServiceName        string             `json:"serviceName" bson:"service_name"` // 顶层字段，供历史检索
	IntentType         IntentType         `json:"intentType,omitempty" bson:"intent_type"`
	IntentConfidence   float64            `json:"intentConfidence,omitempty" bson:"intent_confidence"`
	CorrectionLog      []CorrectionRecord `json:"correctionLog,omitempty" bson:"correction_log"`
	OrchestrationTrace []TraceStep        `json:"orchestrationTrace,omitempty" bson:"orchestration_trace"`
	Status             SessionStatus      `json:"status" bson:"status"` // running|done|failed
	CreatedAt          time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updated_at"`
}

// SessionSummary 历史会话列表项（GET /opsagent/history）。
type SessionSummary struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`  // chat|alert
	Query       string    `json:"query"` // 自然语言提问（alert 为触发摘要）
	ServiceName string    `json:"serviceName"`
	IntentType  string    `json:"intentType,omitempty"`
	Status      string    `json:"status"` // running|done|failed
	DiagnosisID string    `json:"diagnosisId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}
