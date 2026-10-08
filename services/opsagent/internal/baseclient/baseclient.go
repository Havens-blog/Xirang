// Package baseclient 底座契约客户端（超时/重试/降级语义在客户端内闭环）。
//
// 以 HTTP/MCP 契约调用 Haven 底座：logquery（日志联邦 + diagnose 风险分）、
// mcp/cam（资产查询，stdio JSON-RPC）、eiam（令牌本地校验）、通知（直连渠道 webhook）。
// 接口签名见 tech-design › Interface 2 BaseModuleClient；响应结构与冻结契约不符时返回
// ContractDriftError（该步降级「无数据（契约异常）」并递增 opsagent_contract_drift_total
// {module}，10min 内同模块 ≥3 次升级健康告警），绝不向用户抛裸 500。
//
// 契约精确路径/字段来自 contracts/README.md（M1 冻结）+ golden 快照（testdata/contracts）。
package baseclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// 单步预算与重试（与 tech-design › 同步编排延迟预算对齐）。
const (
	logqueryTimeout = 8 * time.Second // 日志联邦单步超时（含 1 次重试计入）
	assetTimeout    = 5 * time.Second // 资产查询 5s 预算（B2 无 per-call 超时，opsagent 自行实现）
	notifyTimeout   = 10 * time.Second
	defaultRetries  = 1
)

// ---------------------------------------------------------------------------
// 错误类型（客户端内闭环：超时/失败/漂移均转为类型化错误，不抛裸网络错误）
// ---------------------------------------------------------------------------

// ContractDriftError 底座响应与冻结契约不符（结构漂移）。
type ContractDriftError struct {
	Module string
	Err    error
}

func (e *ContractDriftError) Error() string {
	return fmt.Sprintf("baseclient: contract drift in module %q: %v", e.Module, e.Err)
}

func (e *ContractDriftError) Unwrap() error { return e.Err }

// ModuleError 底座返回业务错误（非 200 / 信封 code≠0 / 超时耗尽）。上层据此
// 降级「无数据」证据，不向用户透传原文。
type ModuleError struct {
	Module string
	Status int
	Msg    string
}

func (e *ModuleError) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("baseclient: module %q http %d: %s", e.Module, e.Status, e.Msg)
	}
	return fmt.Sprintf("baseclient: module %q: %s", e.Module, e.Msg)
}

// ---------------------------------------------------------------------------
// 契约漂移运行时记账
// ---------------------------------------------------------------------------

const (
	driftAlertWindow    = 10 * time.Minute
	driftAlertThreshold = 3
)

// driftTracker 契约漂移计数：累计计数（供 opsagent_contract_drift_total{module}
// 指标导出）+ 10min 滚动窗口健康告警判定。
type driftTracker struct {
	mu     sync.Mutex
	totals map[string]int64
	recent map[string][]time.Time
}

func newDriftTracker() *driftTracker {
	return &driftTracker{totals: map[string]int64{}, recent: map[string][]time.Time{}}
}

// record 递增模块漂移计数，返回 (累计计数, 是否触发健康告警)。
func (t *driftTracker) record(module string) (total int64, alert bool) {
	now := time.Now()
	cutoff := now.Add(-driftAlertWindow)

	t.mu.Lock()
	defer t.mu.Unlock()
	t.totals[module]++

	win := t.recent[module]
	kept := win[:0]
	for _, ts := range win {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	kept = append(kept, now)
	t.recent[module] = kept
	return t.totals[module], len(kept) >= driftAlertThreshold
}

// total 返回模块累计漂移计数（指标导出用）。
func (t *driftTracker) total(module string) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.totals[module]
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Config 底座客户端配置。
type Config struct {
	LogqueryBaseURL string              // e-cam-service 前缀（如 http://localhost:8001）
	McpServerPath   string              // mcp-server.exe 路径（资产查询 stdio）
	SessionProvider TokenVerifier       // eiam 本地令牌校验（共享 Redis + HMAC；nil 则 VerifyToken 报未配置）
	ChannelWebhooks map[string]string   // 通知渠道 → webhook URL（dingtalk/wecom/feishu）
	Timeout         time.Duration       // 单步超时；≤0 取 logqueryTimeout
	Retries         int                 // 重试次数（不含首次）；<0 取 defaultRetries
	OnHealthAlert   func(module string) // 契约漂移健康告警回调（10min ≥3 次触发）
}

// Client BaseModuleClient 实现（domain.BaseModuleClient）。
type Client struct {
	hc       *http.Client
	logBase  string
	mcpPath  string
	verify   TokenVerifier
	webhooks map[string]string
	timeout  time.Duration
	retries  int
	drift    *driftTracker
	onAlert  func(string)
}

// New 构造 Client。
func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = logqueryTimeout
	}
	retries := cfg.Retries
	if retries < 0 {
		retries = defaultRetries
	}
	return &Client{
		hc:       &http.Client{},
		logBase:  cfg.LogqueryBaseURL,
		mcpPath:  cfg.McpServerPath,
		verify:   cfg.SessionProvider,
		webhooks: cfg.ChannelWebhooks,
		timeout:  timeout,
		retries:  retries,
		drift:    newDriftTracker(),
		onAlert:  cfg.OnHealthAlert,
	}
}

// recordDrift 记账 + 触发健康告警回调。
func (c *Client) recordDrift(module string) {
	_, alert := c.drift.record(module)
	if alert && c.onAlert != nil {
		c.onAlert(module)
	}
}

// envelope 底座统一响应信封（实测：logquery/资产用 msg；eiam/alert 部分用 message，
// 兼容两种键位；data 为 json.RawMessage 保留原始字节供逐字段解码）。
type envelope struct {
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// errMsg 信封错误文案（兼容 msg/message 两种键位）。
func (e envelope) errMsg() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Message
}

// isRetryable 报告错误是否值得重试（超时/连接错误/5xx）。结构漂移与业务错误不重试。
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var me *ModuleError
	if errors.As(err, &me) && me.Status >= 500 {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}

// do 执行一次带重试、信封解析、漂移检测的 HTTP 调用，将 data 解码到 out。
func (c *Client) do(ctx context.Context, module, method, url string, reqBody, out any) error {
	var body io.Reader
	if reqBody != nil {
		raw, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("baseclient: marshal request: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		lastErr = c.once(ctx, module, method, url, body, out)
		if lastErr == nil {
			return nil
		}
		if !isRetryable(lastErr) {
			break
		}
	}
	return lastErr
}

// once 单次调用：超时 ctx + 信封解析 + 漂移检测。
func (c *Client) once(ctx context.Context, module, method, url string, body io.Reader, out any) error {
	cctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, method, url, body)
	if err != nil {
		return &ModuleError{Module: module, Msg: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return &ModuleError{Module: module, Msg: "timeout"}
		}
		return &ModuleError{Module: module, Msg: err.Error()}
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ModuleError{Module: module, Msg: err.Error()}
	}

	if resp.StatusCode != http.StatusOK {
		return &ModuleError{Module: module, Status: resp.StatusCode, Msg: truncate(data)}
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		c.recordDrift(module)
		return &ContractDriftError{Module: module, Err: err}
	}
	if env.Code != 0 {
		return &ModuleError{Module: module, Status: env.Code, Msg: env.errMsg()}
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			c.recordDrift(module)
			return &ContractDriftError{Module: module, Err: err}
		}
	}
	return nil
}

// truncate 截断正文（避免错误日志刷屏）。
func truncate(b []byte) string {
	const max = 200
	if len(b) > max {
		return string(b[:max])
	}
	return string(b)
}

// postRaw 单次原始 HTTP POST（通知 webhook 等非底座信封场景）；返回 nil 表示 200。
func (c *Client) postRaw(ctx context.Context, module, url string, body any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	cctx, cancel := context.WithTimeout(ctx, notifyTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return &ModuleError{Module: module, Msg: "timeout"}
		}
		return &ModuleError{Module: module, Msg: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return &ModuleError{Module: module, Status: resp.StatusCode, Msg: truncate(data)}
	}
	return nil
}
