---
id: "2"
title: "后端诊断接口 POST /logs/diagnose(当前窗+前窗聚合,调规则引擎,预留 summary)"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: [1]
surface-key: "e-cam-service"
surface-type: "api"
breaking: false
type: "coding.feature"
mainSession: false
---

# 2: 后端诊断接口 POST /logs/diagnose(当前窗+前窗聚合,调规则引擎,预留 summary)

## Description
新增联邦诊断接口:复用现有 `/aggregate` 通道,对**当前窗口**跑聚合(Total、时间分桶、Top client_ip、Top UA、状态码分布、动作分布),对**前一等长窗口**再跑一帧聚合(Total、Top client_ip → 突增倍数),拼成诊断输入 → 调任务 1 的规则引擎 → 输出 DiagnoseResponse。WAF 类型;其他云无 WAF 数据(华为非结构化/腾讯未投递)时降级标注;前窗无数据时降级为当前窗绝对量判定,不报错。

## Reference Files
- `docs/proposals/waf-flood-diagnosis/proposal.md` — Source proposal;«In Scope» «Success Criteria» «Key Risks»
- `e-cam-service/internal/logquery/service/federation.go`: Aggregate/AggregateRequest/AggregateResponse(复用聚合通道)
- `e-cam-service/internal/logquery/web/handler.go`: 路由注册与鉴权(RequireTenant)/tenantID(诊断接口同构)
- `e-cam-service/internal/shared/cloudx/logquery/aliyun/filter.go`: client_ip/user_agent 维度可聚合映射(dimensionExpr/dimColumnExpr)

## Acceptance Criteria
- [ ] `POST /logs/diagnose`(WAF)返回:风险等级、疑似攻击类型、Top5 攻击源 IP(含请求数与占比)、来源分布、≥3 条措施、`summary` 空字段(AI 解读后置)、per-source 状态
- [ ] 前窗对比:对前一等长窗口执行一次聚合(Total + Top client_ip),输出突增倍数;前窗无数据时降级为仅当前窗绝对量判定,不报错
- [ ] 聚合失败源 per-source 状态标注,已聚合源照常判定
- [ ] 单次诊断聚合查询 ≤2 帧(当前窗 + 前窗),不放大 SLS 扫描量;鉴权/租户边界与现有 /logs/* 一致
- [ ] `go build ./internal/logquery/...` 与相关单测全绿;明细/聚合既有行为回归通过(诊断不改变既有接口)

## Hard Rules
仅修改 e-cam-service 仓;诊断仅 WAF 类型,SLB/CDN 不开放(返回明确错误);不得放大扫描量(前窗仅 1 帧,手动触发);不得引入 LLM(本任务只预留 summary 字段)。

## Implementation Notes
- 路由:`g.POST("/diagnose", h.Diagnose)` 挂 /api/v1/cam/logs 组(与 search/aggregate 同组,RequireTenant 已有)。
- 请求体:复用 aggregate 的 log_type/start_time/end_time/query/clouds/resources/filters;服务层 `Diagnose(ctx, tenantID, req)`。
- 复用:`s.Aggregate` 内部走 cachedCall;诊断当前窗直接调 `aggregateUncached`(或新增聚合调用),前窗用 `end-(start-end)` 窗口再调一次;Top client_ip/user_agent 靠 `Dimension: "client_ip"/"user_agent"` 的聚合。
- 响应含 `Cached`/`CacheStale` 标注(与 search/aggregate 一致)。
- 诊断结果落日志(duration_ms、风险等级),便于后续校准阈值。
