---
id: "3"
title: "前端 CDN Tab 内嵌「缓存分析」卡(手动触发/双命中率总览/域名排行/未命中 URI/状态码分布/优化项复制/AI 解读区/前窗对比降级)"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: [2]
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 3: 前端 CDN Tab 内嵌「缓存分析」卡(手动触发/双命中率总览/域名排行/未命中 URI/状态码分布/优化项复制/AI 解读区/前窗对比降级)

## Description
在日志查询 **CDN Tab** 新增"缓存分析"折叠卡,镜像 WAF 流量诊断卡(LogDiagnoseCard)交互:手动触发按钮(取当前查询的窗口/域名/检索式/筛选),调用 `/logs/cache-analyze`,渲染:双命中率总览(请求全请求/可缓存两档 + 字节全请求/可缓存双口径)、域名命中率排行 Top 8(按口径分列或角标,标注"已覆盖源命中率")、未命中 URI TOP(查询串归一后,标注归属域名)、状态码分布(4xx/5xx 占比,文案与后端"状态码分布"口径一致)、结构化优化项(非优档 Top 5,含域名/URI 前缀 + 建议动作 + 未命中流量占比 + 可信度分级,可复制;低可信度启发式默认折叠;统一标注"建议,执行前请验证")、前窗对比(↑↓ 标注,卡面显示"对比基期:前一 N 分钟"与"基于 X/Y 源")、AI 解读区(summary 非空时展示)。含加载态/错误态/降级提示,不白屏;独立于既有统计视图,不影响明细/聚合行为。

## Reference Files
- `docs/proposals/cdn-cache-analysis/proposal.md` — Source proposal;«Proposed Solution» «In Scope» «Success Criteria»
- `e-cam-web/src/views/logs/components/LogDiagnoseCard.vue`: WAF 诊断卡交互镜像(触发/加载/降级/措施复制/drilldown)
- `e-cam-web/src/views/logs/index.vue`: CDN Tab 现有结构/聚合条/统计卡(缓存分析卡同区块插入)
- `e-cam-web/src/api/logs.ts`: API 封装(cacheAnalyzeApi 新增)
- `e-cam-web/src/api/types/logs.ts`: 响应类型(CacheAnalyzeResponse)

## Acceptance Criteria
- [ ] CDN Tab 出现"缓存分析"入口(折叠卡,仅 logType==='cdn' 显示),点击手动触发后调用 `/logs/cache-analyze` 并渲染结果
- [ ] 展示:双命中率总览(请求全请求/可缓存两档 + 字节全请求/可缓存双口径,partial 上偏标注)、域名命中率排行 Top 8(按口径分列或加角标,标注"已覆盖源命中率")、未命中 URI TOP(查询串归一,标注归属域名)、状态码分布(4xx/5xx 占比)
- [ ] 结构化优化项(非优档 Top 5):含域名/URI 前缀 + 建议动作 + 未命中流量占比 + 可信度分级,可复制;低可信度启发式默认折叠;统一标注"建议,执行前请验证"
- [ ] 前窗对比:命中率较基期 ↑↓ 标注,卡面显示"对比基期:前一 N 分钟"与"基于 X/Y 源";缺失源/前窗无数据/接口降级时给出对应提示(不白屏);加载态/失败可重试
- [ ] AI 解读区:summary 非空时展示(后端生成,前端不调模型);状态码区块文案与后端"状态码分布"口径一致,不叫"状态码命中率"(eval residual issue 2)
- [ ] 缓存分析卡不改变既有明细/统计/聚合行为(纯新增区块);`vue-tsc --noEmit` 通过;既有前端测试回归通过

## Hard Rules
仅修改 e-cam-web 仓;缓存分析卡不得改变既有明细/统计/聚合行为(纯新增区块,无守卫危险操作);不得在前端引入模型调用(AI 解读由后端 summary 占位/填充);仅 CDN 类型显示入口。

## Implementation Notes
- 触发按钮放置:CDN Tab 统计卡/聚合条附近("缓存分析"按钮 + 折叠面板),复用当前 `timeRange`/`keyword`/`selectedClouds`/`selectedResources`/`buildFilters()`。
- 复用 LogDiagnoseCard 的风格与 drilldown 交互:双命中率用徽标/进度条区分口径;优化项清单用纯文本列表 + 复制按钮;低可信度优化项折叠在"更多建议"下。
- `cacheAnalyzeApi` 请求体 = 当前聚合请求字段(不含 dimension/metric);响应类型 `CacheAnalyzeResponse`(request_hit_rate/byte_hit_rate/domain_ranks/uri_misses/status_dist/summary/optimizations/prev_window/per_source)。
- 前端不调模型、不取原始日志;AI 解读区仅在 `summary` 非空时渲染,为空时隐藏(不占位灰块)。
- 前端展示的命中率均为"已覆盖源命中率",覆盖率/缺失源信息来自 dimension_notes/per_source 字段,与总览同时标注。
