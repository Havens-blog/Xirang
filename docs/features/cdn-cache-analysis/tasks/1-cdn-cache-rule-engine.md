---
id: "1"
title: "cdn-cache 缓存命中率规则引擎(双命中率/健康档位/gap 归因/结构化优化项,纯函数+单测)"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: "e-cam-service"
surface-type: "api"
breaking: false
type: "coding.feature"
mainSession: false
---

# 1: cdn-cache 缓存命中率规则引擎(双命中率/健康档位/gap 归因/结构化优化项,纯函数+单测)

## Description
实现 CDN 缓存分析的确定性判定引擎(纯函数、零 I/O、可单测),镜像 WAF 规则引擎 `internal/logquery/diagnose` 的形态。输入:当前窗聚合行(cache_hit×host 分布、cache_hit×uri 分布、status 分布、sum_bytes 全请求/可缓存双口径)+ 前窗 cache_hit 分布对比。输出:请求命中率(全请求/可缓存两档)+ 字节命中率(全请求/可缓存双口径)、健康档位(优/中/差)、请求-字节 gap 双向归因(大文件回源带宽 + 小文件回源请求)、未命中 URI 集中度(查询串归一后)、4xx/5xx 占比 → 结构化优化项(非优档按未命中流量占比降序 Top N,默认 5;优档允许 0 条)+ 可信度分级(高/中/低)。判据与阈值来自 proposal:命中=hit+partial、未命中=miss+error;优档默认 90% 参数化,按历史命中率分位数初始化。

## Reference Files
- `docs/proposals/cdn-cache-analysis/proposal.md` — Source proposal;«Proposed Solution» «Key Scenarios» «Success Criteria» «Key Risks»
- `e-cam-service/internal/logquery/diagnose/`: WAF 规则引擎形态参考(纯函数输入/输出、config.go 阈值集中配置、diagnose_test.go 单测模式)
- `e-cam-service/internal/logquery/service/federation.go`: 聚合编排/AggregateParams/AggregateResult(诊断输入结构同源)
- `e-cam-service/internal/shared/cloudx/logquery/types.go`: cache_hit 四态归一(hit/miss/partial/error)共用映射源

## Acceptance Criteria
- [ ] 双命中率计算:请求命中率分全请求/可缓存两档;字节命中率分全请求(主口径 = hit+partial 字节/总字节,partial 按全命中计入并显式标注"字节命中率上偏")与可缓存字节口径(与可缓存请求档对齐);命中=hit+partial、未命中=miss+error 归类文档化
- [ ] 健康档位优/中/差阈值参数化(默认 90% 优,集中配置可校准);动态占比高域名按历史命中率分位数初始化档位边界,取不到时回退默认阈值;档位临界分有单测覆盖
- [ ] 请求-字节 gap 检测:gap 大 → 大文件未命中(回源带宽浪费)优化项;请求命中率低而字节尚可 → 小文件大量 miss(回源请求数/QPS 浪费)优化项;请求与字节两个方向归因同时纳入优化项
- [ ] 未命中 URI 集中度(查询串归一后)判定 + 4xx/5xx 占比 → 结构化优化项(域名/URI 前缀 + 建议动作(加缓存规则/调 TTL/忽略查询串/强制回源校验)+ 未命中流量占比);非优档按未命中流量占比降序 Top N(默认 5),优档允许 0 条(不硬凑建议);"忽略查询串"动作须附 miss 集中证据
- [ ] 优化项带可信度分级(高 = URI miss 集中度与 gap 证据充分、中 = 启发式/近似口径、低 = 动态接口类启发式判定);低可信度建议标记供前端默认折叠
- [ ] 纯函数(无网络/DB/LLM),同一输入 → 同一输出(口径可复现);单测覆盖率 ≥80%(双命中率计算/档位判定/gap 检测/优化项生成/URI 归一)

## Hard Rules
仅修改 e-cam-service 仓;纯函数零 I/O,不得引入 LLM/外部调用;命中率口径与 cache_hit 分布聚合可复现(同一时间边界 + 同一执行批次输入 → 同一输出);partial/error 归类文档化,不得静默改变既有归一路径。

## Implementation Notes
- 结构建议:`internal/logquery/cdncache/`(config.go 集中阈值 + engine.go 纯函数 `Evaluate(input) Result` + engine_test.go),镜像 `internal/logquery/diagnose/` 的组织方式;优化项为结构化对象(domain/uriPrefix/action/missTrafficRatio/confidence),非纯文案。
- **可缓存口径剔除机制统一(eval residual issue 1)**:"按 cache_hit×status 剔除 4xx/5xx 与 cache_hit=error"与"状态×cache_hit 二维交叉不承诺"互斥,实现期二选一且必须一致——要么聚合例外支持 cache_hit×status 剔除,要么改用不依赖交叉的近似剔除(按 cache_hit=error + 4xx 状态分布近似);method(POST)/认证/Range 不在单维度 /aggregate 输入中,一律计入可缓存分母并显式标注"可缓存档为聚合层近似(口径偏保守,避免高估命中)"。
- **阈值校准**:上线初期按各域名历史命中率分位数初始化档位边界(历史数据来源为一次性查询现有 /aggregate 按域名分组的 cache_hit 聚合,不新增采集/存储);取不到回退默认;漂移监控不属本期(diagwatch 后续立项),本期仅单次快照 + 参数化。
- 规则输入来自任务 2 的聚合输出结构;与任务 2 协商好 `CacheAnalyzeInput`/`CacheAnalyzeResult` 结构契约(域名排行/URI TOP/状态码分布/双字节口径/前窗对比字段)。
- 单测覆盖边界:两档分子分母、partial 上偏标注、优/中/差临界分、gap 双向归因、URI 归一(带/不带 query、参数排序)、优档 0 条优化项。
