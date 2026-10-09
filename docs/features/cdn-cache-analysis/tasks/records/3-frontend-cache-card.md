---
status: "completed"
started: "2026-09-21 13:01"
completed: "2026-09-21 13:21"
time_spent: "~20m"
---

# Task Record: 3 前端 CDN Tab 内嵌「缓存分析」卡(手动触发/双命中率总览/域名排行/未命中 URI/状态码分布/优化项复制/AI 解读区/前窗对比降级)

## Summary
CDN Tab 内嵌「缓存分析」折叠卡落地(镜像 WAF 诊断卡交互):新增 CacheAnalyzeCard 组件 + cachecard 纯函数展示辅助 + cacheAnalyzeApi 与 CacheAnalyzeResponse 契约类型,接入 logs/index.vue CDN Tab(仅 cdn 显示,纯新增区块)。展示:双命中率总览(请求/字节 × 全请求/可缓存四格 + partial 上偏等口径标注)、域名命中率排行 Top 8(双口径分列 + 档位角标 + 已覆盖源命中率标注)、未命中 URI TOP(查询串归一 + 归属域名,空 → 归属未知 + 变体归并标注)、状态码分布(4xx/5xx 计数与占比,口径名与后端一致)、结构化优化项(目标/动作/未命中流量占比/可信度,可复制,低可信默认折叠于「更多建议」,统一标注「建议,执行前请验证」)、前窗对比(对比基期:前一 N + 基于 X/Y 源 + ↑↓ 趋势与告警 + 缺失源列出)、AI 解读区(summary 非空才渲染)。降级不白屏:前窗缺失/prev_error/dimension_notes/per-source 失败均有提示;窗口 >6h 预估拦截在错误态提供「确认并继续分析」(confirm:true 放行);加载态/失败可重试;域名/URI 行支持下钻筛选。

## Changes

### Files Created
- e-cam-web/src/views/logs/cachecard.ts
- e-cam-web/src/views/logs/cachecard.test.ts
- e-cam-web/src/views/logs/components/CacheAnalyzeCard.vue
- e-cam-web/src/views/logs/components/CacheAnalyzeCard.test.ts

### Files Modified
- e-cam-web/src/api/logs.ts
- e-cam-web/src/api/types/logs.ts
- e-cam-web/src/views/logs/index.vue

### Key Decisions
- 响应契约以任务 2 已提交后端实现为准(result 内 domain_ranking/miss_uri_top/status/recommendations/prev_trend/notes + 根级 prev/prev_sources/sources/dimension_notes),而非任务书 Implementation Notes 中的预设字段名(domain_ranks/uri_misses 等)——两者冲突时按已落地接口
- 缓存分析上下文复用 diagnoseContext computed(结构同构:当前查询字段、无 dimension/metric),以别名 cacheAnalyzeContext 引用,避免复制组装逻辑产生漂移
- 预估扫描量拦截(窗口 >6h)采用错误态内联「确认并继续分析」按钮(带 confirm:true 重发)而非模态框:与既有错误态/重试交互同构且可测
- 低可信优化项折叠用 v-if 而非 v-show(v-show 保留 DOM 会使「默认折叠」不可断言);confidence==='low' 兜底 low_confidence 标记缺失的场景
- 前端不引入任何模型调用;AI 解读区仅 summary 非空渲染,为空不占位

## Test Results
- **Tests Executed**: Yes
- **Passed**: 713
- **Failed**: 0
- **Coverage**: 96.3%

## Acceptance Criteria
- [x] CDN Tab 出现「缓存分析」入口(折叠卡,仅 logType==='cdn' 显示),点击手动触发后调用 /logs/cache-analyze 并渲染结果
- [x] 展示双命中率总览(请求全请求/可缓存两档 + 字节双口径,partial 上偏标注)、域名命中率排行 Top 8(双口径分列 + 已覆盖源命中率标注)、未命中 URI TOP(查询串归一,标注归属域名)、状态码分布(4xx/5xx 占比)
- [x] 结构化优化项(非优档 Top 5):含域名/URI 前缀 + 建议动作 + 未命中流量占比 + 可信度分级,可复制;低可信度启发式默认折叠;统一标注「建议,执行前请验证」
- [x] 前窗对比:命中率较基期 ↑↓ 标注,卡面显示「对比基期:前一 N 分钟」与「基于 X/Y 源」;缺失源/前窗无数据/接口降级时给出对应提示(不白屏);加载态/失败可重试
- [x] AI 解读区:summary 非空时展示(后端生成,前端不调模型);状态码区块文案与后端「状态码分布」口径一致,不叫「状态码命中率」
- [x] 缓存分析卡不改变既有明细/统计/聚合行为(纯新增区块);vue-tsc --noEmit 通过;既有前端测试回归通过

## Notes
静态检查:vue-tsc -b 通过;eslint 对本次 7 个文件零告警(仓库存在 3 个与本次无关的既有 error:api/types/cost.ts、statistics.ts、task.ts 空接口);prettier --check 因仓库无 .prettierrc 按默认配置对全仓 414 个文件(含未触碰的 LogDiagnoseCard.vue 等)同报,本次文件与仓库既有风格(4 空格)一致,未做 --write 以免偏离基线。coverage 96.3 为新增源(cachecard.ts + CacheAnalyzeCard.vue)v8 范围口径(语句 96.34%/分支 89.86%/行 98.95%)。后端 feature flag LOGQUERY_CACHE_ANALYZE_ENABLED 默认关,前端联调需先开启(>6h 窗口需 confirm 放行)。
