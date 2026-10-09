---
status: "completed"
started: "2026-09-21 12:13"
completed: "2026-09-21 12:59"
time_spent: "~46m"
---

# Task Record: 2 后端接口 POST /logs/cache-analyze(当前窗并发聚合+前窗 1 帧,调规则引擎,降级标注,预留 summary)

## Summary
新增联邦缓存分析接口 POST /logs/cache-analyze(仅 CDN,镜像 WAF 诊断编排骨架):当前窗 5 个维度组并发聚合(cache_hit 计数分布主帧 / cache_hit×sum_bytes 字节口径 / status 状态码 / host×nonhit_count 域名未命中 / url×nonhit_count URI 未命中,order by 未命中数天然 miss 预筛),前一等长窗口仅 1 帧 cache_hit 分布对比,拼装 cdncache.CacheAnalyzeInput 调任务 1 纯函数引擎 → 输出双命中率(全请求/可缓存)+ 字节双口径、域名命中率排行(未命中 Top N 域名,双口径分列)、未命中 URI TOP(查询串归一/变体归并/归属域名)、状态码 4xx/5xx 占比、结构化优化项、summary 占位(恒空,不引入 LLM)、per-source 状态与 dimension_notes 覆盖范围标注。成本护栏:窗口上限 24h、执行前预估扫描量(>6h 默认拦截提示 confirm)、独立 feature flag LOGQUERY_CACHE_ANALYZE_ENABLED 默认关(关闭明确报错)、同参 SWR 缓存、AggregateFrames=2 成本标注。新增 nonhit_count 聚合指标(单维度聚合内 case-when 过滤 cache_hit=miss+error,不依赖二维交叉、不做 host×uri 交叉;DCDN hit_info 复合值以 regexp_extract 首段镜像 mapper firstSegment 语义,归一仍走明细层共用 NormalizeCacheHit 单一映射源);缺失源降级标注、非主维度失败仅退化不白屏、前窗失败/无数据降级为当前窗绝对量判定。16 个新单测全绿,新文件逐函数覆盖率 96.7%。

## Changes

### Files Created
- e-cam-service/internal/logquery/service/cache_analyze.go
- e-cam-service/internal/logquery/service/cache_analyze_test.go
- e-cam-service/internal/logquery/web/handler_cache_analyze_test.go
- e-cam-service/internal/shared/cloudx/logquery/aliyun/nonhit_metric_test.go
- e-cam-service/internal/shared/cloudx/logquery/tencent/nonhit_metric_test.go

### Files Modified
- e-cam-service/internal/logquery/web/handler.go
- e-cam-service/internal/shared/cloudx/logquery/aggregate.go
- e-cam-service/internal/shared/cloudx/logquery/aliyun/filter.go
- e-cam-service/internal/shared/cloudx/logquery/tencent/eo.go

### Key Decisions
- 域名/URI 未命中以新增 nonhit_count 指标实现(单维度聚合内 SQL case-when 过滤,order by v=未命中数):规避复用 host|state 复合分组键时联邦 TopN 截断(每源/全局 limit 10)导致的命中率偏置(hit 组挤掉 miss 组 → 误判 100% 命中),并天然满足 URI miss 预筛;域名排行取未命中 Top N 域名(Count=该域名全请求、Value=未命中数,口径在响应注释与 notes 标注)
- cache_hit 归一单源化:编排层 normalizeCacheHitDist 将聚合组键(源原始值如 TCP_HIT/hit_info 复合值/EdgeCacheStatus)经明细层共用 logquery.NormalizeCacheHit 归一并同态合并,不另起归一路径;DCDN hit_info 复合值 SQL 侧 regexp_extract('^[^|,]+') 镜像 mapper firstSegment(firstSegment(x,'|'),',') 语义并有注释锚点
- URI 归属域名不做 host×uri 交叉,仅从含主机的完整 URL 组键(离线转存 RequestURL 等)解析,取不到置空由 notes 说明
- 共享映射文件仅做纯增量扩展(AggregateMetrics/MetricIsWeighted 增加 nonhit_count、metricExpr 三个 CDN kind、eoMetricExpr),既有指标/维度行为零改动,符合'只新增不改既有日志查询路径'硬规则
- 预估扫描量拦截以窗口时长为确定性代理(>6h 默认拦截 + confirm 放行,24h 硬上限),拦截发生于任何聚合调用之前

## Test Results
- **Tests Executed**: Yes
- **Passed**: 137
- **Failed**: 0
- **Coverage**: 96.7%

## Acceptance Criteria
- [x] POST /logs/cache-analyze(仅 CDN)返回双命中率(全请求/可缓存)+ 字节双口径 + 域名命中率 Top 8(双口径分列,总览标注已覆盖源)+ 未命中 URI TOP(查询串归一,Top 5,归属域名)+ 状态码 4xx/5xx 占比 + 结构化优化项(Top 5)+ summary 占位 + per-source 状态
- [x] 一次分析 ≤2 帧窗口聚合:当前窗维度聚合 N=5(host/cache_hit 计数/cache_hit 字节/status/url),前窗仅 1 帧;窗口上限 24h;预估扫描量超限默认拦截并提示确认(拦截先于任何聚合调用)
- [x] URI 维度查询串归一后聚合(变体归并);URI 高基数有分组上限(每源/联邦 TopN limit 10)+ miss 预筛(order by 未命中数)+ 超限降级(聚合失败仅 dimension_notes 标注,不整卡失败)
- [x] 缺失源降级:dimension_notes 标注覆盖范围(覆盖 X/Y 源及缺失明细),已覆盖源照常计算不白屏;前窗无数据/失败 → 降级当前窗绝对量判定不报错;单源聚合失败 → per-source 状态标注
- [x] 命中率口径与 cache_hit 分布可复现(同一时间边界 + 同批并发 + golden 断言调用窗口精确;同参 SWR 缓存);summary 占位恒空;feature flag 默认关、只新增不改既有日志查询路径(关闭时明确报错)
- [x] 回归交付:internal/logquery/... 与 shared/cloudx/logquery/... 既有测试全绿(137 passed/0 failed),go build ./internal/logquery/... 绿;跨源/跨层口径对照(aliyun DCDN/Akamai/离线转存/EdgeOne)与 golden 回放锁定(固定日志样本、窗口口径一致)

## Notes
coverage 96.7% 为本任务新文件 cache_analyze.go 逐函数覆盖率均值(go tool cover -func 实测:88.9%~100%,8 函数);所在 service 包整体(含本任务未触碰的 federation/diagnose/result_cache 既有路径)为 75.0%。测试统计为 go test -count=1 ./internal/logquery/... ./internal/shared/cloudx/logquery/... 全量输出(137 pass/0 fail,含本任务 16 个新用例)。环境限制:-race 需 cgo,本机无 C 工具链无法启用(与 justfile 既有测试配方一致,非本任务引入)。静态检查:go vet ./internal/logquery/... ./internal/shared/cloudx/logquery/... 绿;gofmt 本任务文件已格式化(gofmt -l 另列出的均为本任务未触碰的既有未格式化文件);全仓 go build ./... 仅 tests/osstest/harness.go 失败(web.NewAssetHandler 签名漂移,系并发特性未完成更新,先于本任务存在,不属本任务面与 AC 构建门 'go build ./internal/logquery/...')。仓库存在大量并发特性脏文件,提交须显式 pathspec 仅含本任务 9 个文件。
