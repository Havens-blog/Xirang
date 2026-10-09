package baseclient

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// logqueryModule 契约漂移指标模块标识。
const logqueryModule = "logquery"

// logSearchRequest 契约 B1 POST /api/v1/cam/logs/search 请求体。
type logSearchRequest struct {
	LogType   string   `json:"log_type"`
	StartTime int64    `json:"start_time"` // Unix 毫秒 UTC
	EndTime   int64    `json:"end_time"`   // Unix 毫秒 UTC
	Query     string   `json:"query,omitempty"`
	Clouds    []string `json:"clouds,omitempty"`
	Resources []string `json:"resources,omitempty"`
	Limit     int      `json:"limit,omitempty"`
}

// logSearchResponse 契约 B1 /search 的 data（最小稳定子集）。
type logSearchResponse struct {
	LogType   string            `json:"log_type"`
	Total     int64             `json:"total"`
	Truncated bool              `json:"truncated"`
	Entries   []json.RawMessage `json:"entries"`
	RiskScore *float64          `json:"riskScore,omitempty"` // diagnose 才有；search 缺失即为 nil
}

// b1EntryHeader 多态条目的固定字段（契约 B1：meta{cloud,resource_id} + timestamp）。
type b1EntryHeader struct {
	Meta struct {
		Cloud      string `json:"cloud"`
		ResourceID string `json:"resource_id"`
	} `json:"meta"`
	Timestamp int64 `json:"timestamp"` // Unix 毫秒 UTC
}

// QueryLogs 日志联邦查询（契约 B1）。超时/重试/漂移在客户端内闭环；出错返回类型化
// 错误，由编排层降级「无数据」证据。
func (c *Client) QueryLogs(ctx context.Context, req domain.LogQueryRequest) (domain.LogQueryResult, error) {
	body := logSearchRequest{
		LogType:   req.LogType,
		StartTime: req.StartTime.UnixMilli(),
		EndTime:   req.EndTime.UnixMilli(),
		Query:     req.Resource, // 词频：resource 也作为关键字查询兜底
		Limit:     req.Limit,
	}
	if req.Cloud != "" {
		body.Clouds = []string{req.Cloud}
	}
	if req.Resource != "" {
		body.Resources = []string{req.Resource}
	}

	var data logSearchResponse
	url := c.logBase + "/api/v1/cam/logs/search"
	if err := c.do(ctx, logqueryModule, http.MethodPost, url, body, &data); err != nil {
		return domain.LogQueryResult{}, err
	}

	entries := make([]domain.LogEntry, 0, len(data.Entries))
	for _, raw := range data.Entries {
		e, err := mapLogEntry(data.LogType, raw)
		if err != nil {
			// 单条条目裁剪失败视为噪声，跳过不影响整链（契约漂移由信封级检测兜底）。
			continue
		}
		entries = append(entries, e)
	}
	return domain.LogQueryResult{
		Total:     data.Total,
		Truncated: data.Truncated,
		Entries:   entries,
		RiskScore: data.RiskScore,
	}, nil
}

// mapLogEntry 将 B1 多态条目裁剪为 domain.LogEntry 最小稳定子集。
// level/message 在多态日志类型中无统一字段，置空；Fields 保留原始条目供引用追溯。
func mapLogEntry(logType string, raw json.RawMessage) (domain.LogEntry, error) {
	var h b1EntryHeader
	if err := json.Unmarshal(raw, &h); err != nil {
		return domain.LogEntry{}, err
	}
	var fields map[string]any
	_ = json.Unmarshal(raw, &fields) // 二次解码失败已由 header 解码兜底，忽略
	return domain.LogEntry{
		Timestamp: time.UnixMilli(h.Timestamp),
		Cloud:     h.Meta.Cloud,
		LogType:   logType,
		Resource:  h.Meta.ResourceID,
		Fields:    fields,
	}, nil
}
