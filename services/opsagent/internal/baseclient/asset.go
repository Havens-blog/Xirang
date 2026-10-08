package baseclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/Havens-blog/opsagent/internal/domain"
)

const assetModule = "asset"

// mcpRequest / mcpToolCallParams / mcpResponse 契约 B2 MCP stdio JSON-RPC 形状。
type mcpRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type mcpToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type mcpResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError"`
}

type mcpResponse struct {
	Result *mcpResult `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// assetResult B2 list/search 类工具的 text JSON 形状（CallToolResult.content[0].text 内）。
type assetResult struct {
	Total int `json:"total"`
	Items []struct {
		ID        int64  `json:"id"`
		AssetID   string `json:"asset_id"`
		AssetName string `json:"asset_name"`
		Region    string `json:"region"`
		Status    string `json:"status"`
	} `json:"items"`
}

// QueryAssets 资产查询（契约 B2：MCP Tools stdio JSON-RPC，5s 预算）。
//
// 冻结注记（contracts/README B2）：底座 mcp-server.exe 存在 stdin nil 缺陷（任何
// 客户端消息即 panic），该路径在底座修复前实测不可用；本实现保持客户端结构完整，
// 运行时错误由编排层降级「无数据」证据。
func (c *Client) QueryAssets(ctx context.Context, tenant, tool string, args map[string]any) ([]domain.AssetRecord, error) {
	if c.mcpPath == "" {
		return nil, &ModuleError{Module: assetModule, Msg: "mcp server not configured"}
	}
	cctx, cancel := context.WithTimeout(ctx, assetTimeout)
	defer cancel()
	return c.queryAssetsStdio(cctx, tenant, tool, args)
}

func (c *Client) queryAssetsStdio(ctx context.Context, tenant, tool string, args map[string]any) ([]domain.AssetRecord, error) {
	cmd := exec.CommandContext(ctx, c.mcpPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, &ModuleError{Module: assetModule, Msg: err.Error()}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, &ModuleError{Module: assetModule, Msg: err.Error()}
	}
	if err := cmd.Start(); err != nil {
		return nil, &ModuleError{Module: assetModule, Msg: fmt.Sprintf("start mcp server: %v", err)}
	}
	defer func() { _ = cmd.Process.Kill() }()

	// 租户注入（冻结口径：恒传会话派生的 tenant，绝不透传用户输入）。
	if args == nil {
		args = map[string]any{}
	}
	if tid, err := strconv.ParseInt(tenant, 10, 64); err == nil {
		args["tenant_id"] = tid
	}

	req := mcpRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call",
		Params: mcpToolCallParams{Name: tool, Arguments: args}}
	if err := json.NewEncoder(stdin).Encode(req); err != nil {
		return nil, &ModuleError{Module: assetModule, Msg: err.Error()}
	}

	var resp mcpResponse
	if err := json.NewDecoder(stdout).Decode(&resp); err != nil {
		return nil, &ModuleError{Module: assetModule, Msg: fmt.Sprintf("read mcp response: %v", err)}
	}
	text, err := decodeMcpContent(resp)
	if err != nil {
		return nil, err
	}
	assetType, _ := args["asset_type"].(string)
	return c.parseAssetText(tenant, assetType, text)
}

// decodeMcpContent 校验 MCP JSON-RPC 响应并取出首条 text 内容（独立成函数以便测试，
// 无需起子进程）。
func decodeMcpContent(resp mcpResponse) (string, error) {
	if resp.Error != nil {
		return "", &ModuleError{Module: assetModule, Msg: "mcp error: " + resp.Error.Message}
	}
	if resp.Result == nil || len(resp.Result.Content) == 0 {
		return "", &ModuleError{Module: assetModule, Msg: "empty mcp result"}
	}
	content := resp.Result.Content[0]
	if resp.Result.IsError || content.Type != "text" {
		return "", &ModuleError{Module: assetModule, Msg: "mcp tool error: " + content.Text}
	}
	return content.Text, nil
}

// parseAssetText 解析 MCP text JSON（CallToolResult.content[0].text 内的业务 JSON）为
// []AssetRecord；结构漂移 → ContractDriftError。独立成方法以便单元测试覆盖（无需起子进程）。
func (c *Client) parseAssetText(tenant, assetType, text string) ([]domain.AssetRecord, error) {
	var ar assetResult
	if err := json.Unmarshal([]byte(text), &ar); err != nil {
		c.recordDrift(assetModule)
		return nil, &ContractDriftError{Module: assetModule, Err: fmt.Errorf("parse asset text: %w", err)}
	}

	records := make([]domain.AssetRecord, 0, len(ar.Items))
	for _, it := range ar.Items {
		assetID := it.AssetID
		if assetID == "" {
			assetID = strconv.FormatInt(it.ID, 10)
		}
		records = append(records, domain.AssetRecord{
			AssetID: assetID,
			Type:    assetType,
			Tenant:  tenant,
			Name:    it.AssetName,
			Region:  it.Region,
			Status:  it.Status,
		})
	}
	return records, nil
}
