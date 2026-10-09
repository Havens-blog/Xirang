---
id: "2"
title: "后端接口 POST /logs/cache-analyze(当前窗并发聚合+前窗 1 帧,调规则引擎,降级标注,预留 summary)"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: [1]
surface-key: "e-cam-service"
surface-type: "api"
breaking: false
type: "coding.feature"
mainSession: false
---

# 2: 后端接口 POST /logs/cache-analyze(当前窗并发聚合+前窗 1 帧,调规则引擎,降级标注,预留 summary)

## Description
新增联邦缓存分析接口 `POST /logs/cache-analyze`(仅 CDN 类型),镜像 WAF 诊断编排骨架:对**当前窗口**并发聚合 cache_hit×host(域名)、×uri(查询串归一后的 URI 未命中,单维度聚合、不做 host×uri 交叉)、×status(状态码分布,4xx/5xx 占比;状态×cache_hit 二维交叉为后续增强)+ sum_bytes(全请求与可缓存双字节口径),对**前一等长窗口**再跑 1 帧聚合(cache_hit 分布对比)——≤2 帧窗口聚合;拼成输入调任务 1 规则引擎 → 输出 CacheAnalyzeResponse(双命中率、域名排行 Top 8、未命中 URI TOP(查询串归一,全局 Top N 默认 5,标注归属域名)、状态码分布、结构化优化项、summary 占位、per-source 状态)。缺失源(尚未映射 cache_hit 索引)降级标注;窗口时长上限 24h + 执行前预估扫描量超限拦截;独立 feature flag 默认关,只新增不改既有日志查询路径。

## Reference Files
- `docs/proposals/cdn-cache-analysis/proposal.md` — Source proposal;«In Scope» «Success Criteria» «Key Risks» «Constraints & Dependencies»
- `e-cam-service/internal/logquery/service/diagnose.go`: WAF 诊断编排骨架复用(聚合并发通道/SWR 缓存/per-source 降级/dimension_notes/前窗对比)
- `e-cam-service/internal/logquery/web/handler.go`: 路由注册与鉴权(RequireTenant)/tenantID(诊断接口同构)
- `e-cam-service/internal/shared/cloudx/logquery/types.go` 及 `tencent/eo.go`: cache_hit 归一映射单一来源(aggregatable/cache_hit 维度)
- `e-cam-service/internal/shared/cloudx/logquery/aliyun/filter.go` 及 `tencent/eo.go`: host/uri/status 维度可聚合映射(dimensionExpr/dimGroupExpr)

## Acceptance Criteria
- [ ] `POST /logs/cache-analyze`(仅 CDN)返回:请求命中率(全请求/可缓存两档)+ 字节命中率(全请求/可缓存双口径)、域名命中率 Top 8(按口径分列或加角标,总览明确"已覆盖源命中率")、未命中 URI TOP(查询串归一后,全局 Top N 默认 5,标注归属域名)、状态码分布(4xx/5xx 占比)、结构化优化项(非优档 Top 5,优档允许 0 条)、`summary` 占位字段(AI 解读后置,env 开关)、per-source 状态
- [ ] 一次分析 ≤2 帧窗口聚合:当前窗维度聚合次数 N≤5(host/uri/status/字节口径等维度组),前窗仅 1 帧(cache_hit 分布对比);窗口时长上限 24h;执行前预估扫描量,超限默认拦截并提示确认
- [ ] URI 维度查询串归一(剥离/排序 query 参数)后聚合;URI 高基数(百万级)有分组上限 + miss 预筛 + 超限降级(不导致整卡失败)(eval residual issue 3)
- [ ] 缺失源降级:尚未映射 cache_hit 索引的源在 dimension_notes 标注覆盖范围,已覆盖源照常计算,不白屏;前窗无数据 → 只给当前窗绝对量判定,不报错;单源聚合失败 → per-source 状态标注
- [ ] 命中率口径与 cache_hit 分布聚合可复现(同一时间边界 + 同一执行批次并发查询);响应预留 `summary` 字段;feature flag 默认关、只新增不改既有日志查询路径(开关关闭时接口明确报错)
- [ ] 回归交付:明细/聚合既有行为回归通过(新端点不改既有路径);跨源/跨层口径对照测试(aliyun DCDN/Akamai/离线转存/EdgeOne)+ golden 回放锁定(固定日志样本,窗口口径一致);`go build ./internal/logquery/...` 与相关单测全绿

## Hard Rules
仅修改 e-cam-service 仓;仅 CDN 类型开放(SLB/WAF 明确报错);≤2 帧窗口聚合 + 手动触发,不得放大 SLS 扫描量;不得引入 LLM(只预留 summary 字段);cache_hit 归一沿用任务 1/明细共用单一映射源,不得另起独立归一路径;新端点独立 feature flag,只新增不改既有日志查询代码路径。

## Implementation Notes
- 路由:`g.POST("/cache-analyze", h.CacheAnalyze)` 挂 /api/v1/cam/logs 组(与 search/aggregate/diagnose 同组,RequireTenant 已有)。请求体复用 aggregate 的 log_type/start_time/end_time/query/clouds/resources/filters;服务层 `CacheAnalyze(ctx, tenantID, req)` 镜像 `Diagnose`。
- 复用:`s.Aggregate` 内部 cachedCall;当前窗维度聚合直接并发调用(host/uri/status/sum_bytes),前窗用 `end-(start-end)` 窗口再调 1 帧 cache_hit 分布;Top host/uri/status 靠 `Dimension: "host"/"uri"/"status"` 的聚合,字节口径靠 `Metric: "sum_bytes"`。
- **可缓存口径剔除机制(eval residual issue 1)**:与任务 1 保持一致——若聚合例外不支持 cache_hit×status 交叉剔除,则近似剔除(按 cache_hit=error + 4xx 状态分布近似),不得出现"剔除依赖交叉、交叉又不承诺"的矛盾。
- **URI 高基数(eval residual issue 3)**:定义分组上限(如 GROUP BY 前按 miss 预筛或限制 Top N),SLS 聚合超限时降级为"仅返回域名级 + 全局 Top N 部分",标注截断,不整卡失败。
- **AI 外发隐私与回滚(eval residual issue 4)**:外发仅限聚合统计(域名/URI 前缀命中率、优化项、占比数值);URI 级聚合含查询串时外发前剥离 query 参数与可能的 PII,只发查询串归一后的路径与统计数值,不发原始日志/请求明细;外发由 env 开关控制、失败降级;新端点 feature flag 默认关,异常时关闭开关即回滚,灰度顺序内部环境 → 放开手动触发入口 → 全量。
- 响应含 `Cached`/`CacheStale` 标注(与 search/aggregate/diagnose 一致);`dimension_notes` 风格说明覆盖源范围。
