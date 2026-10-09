---
created: "2026-09-21"
author: "Haven"
status: Draft
intent: "new-feature"
---

# Proposal: CDN 缓存分析(命中率与缓存优化)

## Problem

运维/优化需要快速回答 **CDN 缓存健康度**:当前命中率多高(请求 / 字节)、哪些**域名 / URI** 缓存失效最重、**回源流量浪费**多少、该怎么做缓存优化;当前日志查询只有原始聚合(Total / TopN / 趋势),没有"命中率判定 / 未命中定位 / 优化建议"的结论性输出,优化依赖人工逐域名 / 逐 URI 排查。

### Evidence

- CDN 统一模型已含 `cache_hit`(hit/miss/partial/error)与 `bytes_sent`,`/aggregate` 已支持 `cache_hit` 维度与 `sum_bytes` 指标(实测 aliyun DCDN 聚合白名单含 cache_hit;Akamai cacheStatus 已归一;离线转存 HitInfo 已归一)——**数据面零新增采集**。
- 线上实测(2026-09-20):CDN aliyun 1h 全窗 **1453 万请求**,域名 datasink.jlc.com 单域名 **286 万次**;www.jlc-smt.com 等 WAF 站点 2.2M/h —— 域名/URI 级缓存失效面大,人工逐项看成本高。
- 请求命中率与字节命中率的差距直接对应**回源带宽 / 流量费用浪费**(大文件未命中 = 高字节回源),可量化到金额级,是 CDN 优化最实的抓手。

### Urgency

- 缓存优化是持续成本项:命中率每低 1 个点,热点域名的回源带宽/流量费都会放大;延迟 = 费用 + 源站压力双损失。
- 现有功能已具备全部数据基础,缺的是一次"结论性 + 可执行"的结构化分析;做出来即日可用(手动触发,成本可控)。

## Proposed Solution

在日志查询 **CDN Tab 内嵌「缓存分析」卡**(手动触发,镜像 WAF 流量诊断卡交互):后端对**当前窗口 + 前一等长窗口**做聚合(复用现有联邦聚合通道),规则引擎输出 **请求命中率 + 字节命中率(双口径)、域名命中率排行、未命中 URI TOP、状态码命中率、结构化缓存优化项(域名/URI 前缀 + 建议动作 + 未命中流量占比)**;AI 解读**后置**(复用已落地的 `llm` 包,env 开启才调,失败降级)。

### Innovation Highlights

- **双命中率口径**:请求命中率(hit+partial/总请求)+ 字节命中率(hit+partial 字节/总字节);两者 gap 大 = 大文件未命中 = 回源带宽浪费,直接量化到流量金额。
- **未命中归因到 URI 前缀**:URI 级 miss 集中度是缓存规则优化的直接输入,优化项**结构化**(域名 + URI 前缀 + 建议动作:加缓存规则 / 调 TTL / 忽略查询串 / 强制回源校验),可执行而非纯文案。
- **复用 WAF 诊断骨架**:聚合并发通道 / 前窗对比 / SWR 缓存 / per-source 降级 / AI 后置全套复用,交互一致、成本可控(一次分析 ≤ 2 帧聚合)。
- **缺失源降级**:EdgeOne 等未建 cache_hit 索引的源显式标注(dimension_notes 风格),已覆盖源照常计算,不白屏。

## Requirements Analysis

### Key Scenarios

- **缓存健康**:命中率 ≥90%(请求)且字节命中率接近 → "优",无优化项或少量提示。
- **未命中集中**:某 URI 前缀 miss 占比高(如动态接口误开缓存 / 查询串不忽略)→ 优化项:该前缀加/调缓存规则、忽略查询参数。
- **大文件未命中**:请求命中率 OK 但字节命中率明显低(gap 大)→ 大文件回源浪费,优化项指向大文件 URI + TTL/分片建议。
- **4xx/5xx 高**:状态码命中率低(4xx/5xx 占比高)→ 缓存对错误响应无效,建议回源校验/缓存规则排除错误码。
- **前窗对比**:命中率较前窗上升 / 下降,标注趋势(↑↓),下降 > 阈值时提示关注。
- **错误场景**:EdgeOne / 未开 cache_hit 索引的源 → 命中率按已覆盖源计算并标注覆盖范围;前窗无数据 → 只给当前窗绝对量判定;单源聚合失败 → per-source 状态标注。

### Non-Functional Requirements

- 一次分析 ≤ 2 次聚合查询(当前窗 + 前窗),手动触发控制 SLS 扫描成本。
- 规则引擎为纯函数(输入聚合行 → 输出判定 + 结构化优化项),单测覆盖率 ≥ 80%。
- 命中率口径与 `cache_hit` 分布聚合**可复现**(同一窗口口径一致),partial/error 归类文档化。
- 响应携带 per-source 状态与 `summary` 占位字段(AI 解读后置);前端卡交互镜像 WAF 诊断卡。

### Constraints & Dependencies

- 依赖现有 `/aggregate` 联邦能力(cache_hit 维度 + sum_bytes 指标 + 分桶),不新增采集/存储。
- 各源 cache_hit 语义差异(Akamai cacheStatus、DCDN hit_info、离线转存 HitInfo、EdgeOne 无)→ 归一映射已有,需补测试;EdgeOne 命中率部分缺失走降级。
- SLS 按扫描量计费:必须手动触发而非随查询自动(与 WAF 诊断同约束)。

## Alternatives & Industry Benchmarking

### Industry Solutions

- **云厂商 CDN 控制台命中率报表**(阿里云 CDN / Akamai 报表):权威但只在各自控制台、按域名/报表粒度,与自建统一日志查询隔离,无法跨云横向比。
- **开源时序异常/命中率监控**(Grafana + 云指标):能看趋势但不下沉到 URI 级未命中归因,不产出可执行优化项。
- **纯 LLM 分析**:喂统计给大模型生成结论,灵活但不可审计、成本/延迟、结论不稳定,不适合作为优化依据唯一来源。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing(人工看聚合) | — | 零开发 | 依赖经验、慢、每次多查询多计费 | Rejected: 无法规模化定位未命中与优化 |
| 云厂商控制台命中率报表 | 阿里云 CDN / Akamai | 权威 | 跨云割裂、不下沉 URI、只该云 | Rejected: 不符合统一日志平台 |
| 纯 LLM 分析 | 大模型 API | 灵活、自然语言 | 不可审计、成本/延迟、结论不稳定 | Rejected: 不适合作为优化依据唯一来源 |
| **规则为主 + 可选 AI 解读 + 结构化优化项** | 本项目方案(复用 WAF 诊断骨架) | 可复现/可审计/零新采集,URI 级可执行建议,跨云统一 | 规则阈值需校准 | **Selected: 判据确定可靠,AI 解读后置** |

## Feasibility Assessment

### Technical Feasibility

现有聚合通道(cache_hit 维度、sum_bytes、分桶)已在生产可用(实测 aliyun DCDN cache_hit/sum_bytes 可聚合);新增 `POST /logs/cache-analyze`(复用 diagnose 编排骨架)+ 纯函数规则引擎 + 前端卡,无 showstopper 依赖。

### Resource & Timeline

团队已完整实现并验证 WAF 诊断同构功能(规则引擎/聚合并发/前窗对比/前端卡/AI 后置),同构复用估 1-1.5 天。

### Dependency Readiness

CDN 数据流动正常(实测 CDN aliyun 1h 1453 万请求);cache_hit 聚合白名单已覆盖 DCDN/Akamai/离线转存;EdgeOne cache_hit 未映射(降级标注,不阻塞)。无上游阻塞。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "缓存分析要独立页面" | Need Gate | Refined: CDN Tab 内嵌卡片,复用查询上下文与 WAF 交互;用户确认内嵌 |
| "只算请求命中率即可" | Assumption Flip | Refined: 请求+字节双命中率,字节口径直击回源带宽/费用浪费;用户确认双口径 |
| "优化建议给文案清单" | Assumption Flip | Refined: 结构化优化项(域名/URI 前缀+动作+未命中流量占比)更可执行;用户确认结构化 |
| "命中率可随查询自动算" | Need Gate(成本) | Overturned: 每查询多帧聚合、SLS 计费放大;用户确认手动触发+前窗对比 |

## Scope

### In Scope

- 后端缓存分析接口 `POST /logs/cache-analyze`:当前窗并发聚合 cache_hit×host(域名)、×uri(URI 未命中)、×status(状态码命中率)+ sum_bytes(字节命中率)+ 前窗 1 帧(cache_hit 分布对比)。
- 规则引擎 cdn-cache(纯函数、可单测):双命中率计算、健康档位(优/中/差)、请求-字节 gap(大文件回源浪费)、未命中 URI 集中度、4xx/5xx 占比 → **结构化优化项 ≥3**(域名/URI 前缀 + 建议动作 + 未命中流量占比)。
- 前端:CDN Tab 内嵌"缓存分析"卡(折叠/手动触发/前窗对比/双命中率总览/域名排行/未命中 URI/状态码命中率/优化项可复制/AI 解读区)。
- 缺失源降级标注(EdgeOne 等无 cache_hit 索引 → dimension_notes 风格说明覆盖范围)。
- 响应预留 `summary` 字段(AI 解读后置,复用 `llm` 包,env 开启)。

### Out of Scope

- 定时看护(CDN 版 diagwatch)——结构可复用,后续立项。
- 缓存规则一键下发到云(只给建议,不碰云 API)。
- 独立缓存分析页面 / 多窗口历史趋势(只对比前一等长窗口)。
- UA / 边缘节点粒度命中率(维度集本期固定为域名/URI/状态码)。
- WAF/SLB 命中率分析(本期仅 CDN)。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| EdgeOne 等源无 cache_hit 索引,命中率部分缺失 | H | M | 按已覆盖源计算并标注覆盖范围;不白屏;后续补 EdgeOne 归一映射 |
| partial/error 归入口径与云侧报表不一致引发质疑 | M | M | 命中=hit+partial、未命中=miss+error 文档化;口径以 cache_hit 分布可复现为准 |
| 大窗口多域名扫描成本 | M | L | 复用聚合并发通道 + SWR 缓存;手动触发 ≤2 帧 |
| 各源 cache_hit 语义差异导致命中率口径漂移 | M | M | 已有归一映射(Akamai cacheStatus / hit_info / HitInfo),补跨源口径测试 |
| 字节命中率依赖 bytes_sent 归一批量 | L | M | 各源量纲已映射(response_size 等);缺失源降级为仅请求命中率并标注 |

## Success Criteria

- [ ] CDN Tab 出现「缓存分析」入口,手动触发后 ≤ 2 次聚合查询返回分析结果。
- [ ] 结果包含:请求命中率 + 字节命中率(双口径)、域名命中率 Top 8、未命中 URI TOP(每域名前几)、状态码命中率、结构化优化项 ≥3(含域名/URI 前缀 + 建议动作 + 未命中流量占比)。
- [ ] 命中率口径与 `cache_hit` 分布聚合一致(同一窗口可复现)。
- [ ] 前窗对比给出命中率变化(↑↓);EdgeOne/缺失源降级标注,页面不白屏;明细/聚合既有行为回归通过。
- [ ] 规则引擎纯函数单测覆盖:双命中率计算、档位判定、gap 检测、优化项生成;覆盖率 ≥ 80%;响应预留 `summary` 字段。

## Next Steps

- Proceed to `/quick-tasks` 生成可执行任务。
