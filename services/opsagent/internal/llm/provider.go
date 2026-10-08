// Provider 抽象与三实现。qwen / deepseek / openai 三者均采用 OpenAI 兼容的
// chat/completions 协议（仅 baseURL 与鉴权/模型命名差异），故共享一个 httpProvider
// 实现，经三个命名构造函数注入各自 baseURL，避免三份重复实现（DRY）。
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Havens-blog/opsagent/internal/config"
	"github.com/Havens-blog/opsagent/internal/domain"
)

// 各 provider 的默认 OpenAI-compatible chat/completions 端点。
const (
	qwenBaseURL     = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	deepseekBaseURL = "https://api.deepseek.com/chat/completions"
	openaiBaseURL   = "https://api.openai.com/v1/chat/completions"
)

// ErrDisabled LLM 未配置（apiKey 为空 / provider 未知）。
var ErrDisabled = errors.New("llm disabled: no api key configured")

// disabledProvider 未配置时的占位实现：不发起网络调用，Complete 恒返回 ErrDisabled。
type disabledProvider struct{}

func (disabledProvider) Complete(context.Context, domain.LLMRequest) (string, error) {
	return "", ErrDisabled
}

// apiMessage / apiRequest / apiResponse OpenAI-compatible 协议形状。
type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiRequest struct {
	Model     string       `json:"model"`
	Messages  []apiMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

type apiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// httpProvider OpenAI-compatible chat/completions 通用 HTTP 实现。
type httpProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// Complete 发起一次 chat/completions 调用；依赖 ctx 控制超时（由 GuardedLLM 注入）。
func (p *httpProvider) Complete(ctx context.Context, req domain.LLMRequest) (string, error) {
	msgs := make([]apiMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, apiMessage{Role: m.Role, Content: m.Text})
	}
	body, err := json.Marshal(apiRequest{Model: p.model, Messages: msgs, MaxTokens: req.MaxTokens})
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", readErr
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm http %d: %s", resp.StatusCode, truncateBody(data))
	}

	var out apiResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("llm: empty choices")
	}
	return out.Choices[0].Message.Content, nil
}

// truncateBody 截断错误响应正文（避免日志刷屏）。
func truncateBody(b []byte) string {
	const max = 200
	if len(b) > max {
		return string(b[:max])
	}
	return string(b)
}

// newHTTPProvider 构造 OpenAI-compatible provider（client 依赖 ctx 控时，无独立 timeout）。
func newHTTPProvider(baseURL, apiKey, model string) domain.LLMProvider {
	return &httpProvider{baseURL: baseURL, apiKey: apiKey, model: model, client: &http.Client{}}
}

// NewQwenProvider 通义千问（DashScope compatible-mode）。
func NewQwenProvider(apiKey, model string) domain.LLMProvider {
	return newHTTPProvider(qwenBaseURL, apiKey, model)
}

// NewDeepseekProvider DeepSeek。
func NewDeepseekProvider(apiKey, model string) domain.LLMProvider {
	return newHTTPProvider(deepseekBaseURL, apiKey, model)
}

// NewOpenaiProvider OpenAI。
func NewOpenaiProvider(apiKey, model string) domain.LLMProvider {
	return newHTTPProvider(openaiBaseURL, apiKey, model)
}

// NewProvider 依据配置构造 provider（apiKey 自 config 注入，不入库不入码）；
// 未配置 apiKey 或 provider 未知时返回 disabledProvider（降级路径，不阻塞启动）。
func NewProvider(cfg config.LLMConfig) domain.LLMProvider {
	if cfg.APIKey == "" {
		return disabledProvider{}
	}
	baseURL := cfg.BaseURL
	switch cfg.Provider {
	case "qwen":
		if baseURL == "" {
			baseURL = qwenBaseURL
		}
	case "deepseek":
		if baseURL == "" {
			baseURL = deepseekBaseURL
		}
	case "openai":
		if baseURL == "" {
			baseURL = openaiBaseURL
		}
	default:
		return disabledProvider{}
	}
	return newHTTPProvider(baseURL, cfg.APIKey, cfg.Model)
}
