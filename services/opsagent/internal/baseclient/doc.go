// Package baseclient 底座契约客户端（超时/重试/降级语义在客户端内闭环）。
//
// 以 HTTP 契约调用 Haven 底座：logquery（日志联邦 + diagnose 风险分）、
// alert（通知发送）、mcp/cam（资产查询）、eiam（令牌验证）。接口签名见
// tech-design › Interface 2 BaseModuleClient；响应结构与契约不符时返回
// ContractDriftError（该步降级「无数据（契约异常）」并推送运维告警，
// 绝不向用户抛裸 500）。具体实现在后续任务交付。
package baseclient
