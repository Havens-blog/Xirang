---
status: "completed"
started: "2026-09-21 11:29"
completed: "2026-09-21 12:02"
time_spent: "~33m"
---

# Task Record: 1 cdn-cache 缓存命中率规则引擎(双命中率/健康档位/gap 归因/结构化优化项,纯函数+单测)

## Summary
新增 CDN 缓存命中率规则引擎纯函数包 internal/logquery/cdncache(镜像 diagnose 形态):双命中率(请求/字节 × 全请求/可缓存)、健康档位优/中/差(默认 90%/80% 参数化 + 历史分位数初始化辅助 CalibrateThresholds + 域名级覆盖)、请求-字节 gap 双向归因(大文件回源带宽 tune_ttl + 小文件回源请求 add_cache_rule)、未命中 URI 集中度(查询串剥离/排序归一 NormalizeURIPath/NormalizeURIQuery,变体归并证据)、4xx/5xx 占比 origin_verify、结构化优化项(占比降序 Top 5,优档仅强证据 gap 项或 0 条)、可信度三级(高/中/低 + LowConfidence 折叠标记)、前窗趋势 up/flat/down + 下降关注阈值、口径标注集中 Notes(partial 字节上偏/可缓存聚合层近似/命中归类 hit+partial vs miss+error 文档化)。18 个单测全绿,覆盖率 90.0%(≥80% 目标)。

## Changes

### Files Created
- e-cam-service/internal/logquery/cdncache/config.go
- e-cam-service/internal/logquery/cdncache/engine.go
- e-cam-service/internal/logquery/cdncache/engine_test.go

### Files Modified
无

### Key Decisions
- 可缓存口径剔除机制统一(eval residual issue 1):单维度 /aggregate 无 cache_hit×status 交叉 → 采用不依赖交叉的近似剔除,剔除量取 max(cache_hit=error, 4xx+5xx)(两单维度聚合重叠未知,取大者避免重复剔除高估命中);4xx/5xx 按未命中近似,剔除不触碰命中分子;与任务 2 的口径要求一致
- 契约设计:导出 CacheAnalyzeInput/CacheAnalyzeResult 供任务 2 编排层拼装;HostCacheStat(域名×cache_hit 交叉)与 URIMissItem(URI 未命中)由聚合层产出,引擎纯消费、缺任一维度降级标注不白屏
- gap 双向归因项视为强证据,不受档位限制(优档'允许 0 条'解释为无强证据不硬凑,而非强制清零):请求命中尚可而字节低 → 大文件项;可缓存请求低而字节尚可 → 小文件项;URI 集中度/状态码/域名级项仅非优档生成
- URI 查询串归一双函数:NormalizeURIPath(剥离,集中度分组键)+ NormalizeURIQuery(参数排序归一,变体计数),导出供任务 2 聚合层复用;变体 ≥2 才建议'忽略查询串'并附 miss 集中证据
- 命中归类硬规则文档化:命中=hit+partial、未命中=miss+error、cache_hit '-' 计入分母不计入两侧,经 NoteClassification 常量透出,不改写 shared/cloudx/logquery.NormalizeCacheHit 既有归一路径
- 输出全量确定性排序(占比降序+可信度+字典序兜底/topKey 并列取字典序最小),TestEvaluateDeterministic 以 JSON 全文相等锁定'同一输入→同一输出'

## Test Results
- **Tests Executed**: Yes
- **Passed**: 18
- **Failed**: 0
- **Coverage**: 90.0%

## Acceptance Criteria
- [x] 双命中率计算:请求命中率全请求/可缓存两档 + 字节命中率双口径(partial 全命中计入并标注上偏)+ 命中/未命中归类文档化
- [x] 健康档位优/中/差阈值参数化(默认 90% 优)+ 历史命中率分位数初始化(CalibrateThresholds,取不到回退默认)+ 临界分单测
- [x] 请求-字节 gap 双向归因(大文件回源带宽 + 小文件回源请求)同时纳入优化项
- [x] 未命中 URI 集中度(查询串归一后)+ 4xx/5xx 占比 → 结构化优化项(域名/URI 前缀+动作+未命中流量占比),非优档占比降序 Top 5、优档 0 条不硬凑,忽略查询串附 miss 集中证据
- [x] 优化项可信度分级(高/中/低),低可信度带 LowConfidence 标记供前端默认折叠
- [x] 纯函数零 I/O,同一输入同一输出(确定性单测),单测覆盖率 ≥80%(实测 90.0%)

## Notes
静态检查:go build ./... 全绿;gofmt 已格式化本任务两个文件(gofmt -l 另列出 5 个本任务未触碰的既有未格式化文件:cert/web/reference_handler_test.go、diagwatch/watch.go、watch_test.go、llm/llm.go、llm_test.go,按外科手术原则未改动);go vet ./internal/logquery/... 全绿;justfile 无 e-cam-service 专用配方,按 justfile 委托的 go 命令执行。相邻包回归:diagnose 91.8%、shared/cloudx/logquery 全绿。git 仓库有大量并发脏文件,提交时须显式 pathspec 只加本任务三个新文件。
