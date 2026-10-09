package domain

import "context"

// 编排层 5 个核心接口签名（tech-design › Interfaces 唯一定义处）。
// 实现分别落位于 internal/orchestrator、internal/baseclient、internal/llm、
// internal/risk；本文件只声明契约，不含实现。

// IntentClassifier 意图识别与能力路由（tech-design › Interface 1）。
type IntentClassifier interface {
	// Classify 分类用户自然语言输入；llm 不可用时走模板/正则降级
	Classify(ctx context.Context, input string, tenant string) (IntentResult, error)
	// Correct 依据用户纠正重路由，返回新意图并保留误分类+纠正记录。
	// target+params 与 message 二选一（XOR，与 api-handbook §2 一致）：
	// target+params = 候选意图一键纠正；message = 自然语言纠正（服务端重新分类）。
	// tenant 用于强制租户谓词：会话归属校验与 correction_log 落库均不可跨租户。
	Correct(ctx context.Context, tenantID, sessionID string, target IntentType, params DiagnoseParams, message string) (IntentResult, error)
}

// BaseModuleClient 底座模块统一调用封装：超时/重试/降级语义在客户端内闭环
// （tech-design › Interface 2；HTTP 契约见 api-handbook › Part B）。
type BaseModuleClient interface {
	QueryLogs(ctx context.Context, req LogQueryRequest) (LogQueryResult, error)
	QueryAssets(ctx context.Context, tenant string, tool string, args map[string]any) ([]AssetRecord, error)
	VerifyToken(ctx context.Context, token string) (Identity, error) // 来自 eiam
	Notify(ctx context.Context, channel string, payload Notification) (NotifyResult, error)
}

// Orchestrator 编排流水线入口（tech-design › Interface 3）。
type Orchestrator interface {
	// RunChat 对话式排障：意图→编排→诊断→落库，返回诊断报告
	RunChat(ctx context.Context, session *Session, input string) (*Diagnosis, error)
	// RunAlert 告警主动触发：去重→排查→落库→风险中心+通知
	RunAlert(ctx context.Context, alert AlertEvent) (*RiskEntry, error)
}

// LLMProvider LLM 抽象（tech-design › Interface 4；降级封装 GuardedLLM 落位 internal/llm）。
type LLMProvider interface {
	// Complete 生成文本；返回 (text, err)。超时/解析失败/未配置 → err
	Complete(ctx context.Context, req LLMRequest) (string, error)
}

// ToolMeta 工具风险元数据（tech-design › Interface 5）。
type ToolMeta struct {
	Name      string
	RiskLevel RiskLevel // 注册表静态声明，LLM 不得改变
	Allowlist bool      // 是否在低危白名单内
}

// RiskRegistry 风险分级确定性查表接口（注册表本体是 Go 代码，非 DB 表）。
type RiskRegistry interface {
	// Classify 查表判定；返回错误当且仅当工具未注册
	Classify(tool string) (ToolMeta, error)
}
