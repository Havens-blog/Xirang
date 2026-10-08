// Package baseclient 底座契约客户端（超时/重试/降级语义在客户端内闭环）。
//
// 以 HTTP/MCP 契约调用 Haven 底座：logquery（日志联邦 + diagnose 风险分）、
// mcp/cam（资产查询，stdio JSON-RPC）、eiam（令牌本地校验）、通知（直连渠道 webhook）。
// 接口签名见 tech-design › Interface 2 BaseModuleClient；响应结构与冻结契约不符时返回
// ContractDriftError（该步降级「无数据（契约异常）」并递增 opsagent_contract_drift_total
// {module}，10min 内同模块 ≥3 次升级健康告警），绝不向用户抛裸 500。
//
// 契约精确路径/字段来自 contracts/README.md（M1 冻结）+ golden 快照。资产查询经 MCP
// stdio（契约 B2）——冻结记录底座 mcp-server.exe 存在 stdin nil 缺陷，该路径在底座
// 修复前实测不可用，客户端结构已就绪。
package baseclient
