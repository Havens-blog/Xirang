---
created: "2026-09-19"
author: "Haven"
status: Approved
intent: "enhancement"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: aliyun CDN 指标采集区间查询折叠（90→3 API/域名）

## Problem

aliyun CDN 指标采集对每域名按日逐次调 3 个指标 API（`fetchTrafficByDay`/`fetchBpsByDay`/`fetchHitRateByDay` 各循环 `dates`），30 天 × 3 = 90 次 API/域名。夜间任务随域名数增长，API 调用量成主要成本。

### Evidence

- `internal/shared/cloudx/aliyun/cdn_metrics.go:72/101/125`：三个 fetch 各 `for _, d := range dates` 逐日调用，`Interval="300"`（5min）。
- **其余三厂商已用区间查询**（tencent `fetchCdnMetricRange`、huawei `fetchDomainStats` interval=86400、volcano `fetchEdgeMetric` 均一次调用跨整月）——本特征实为 aliyun 单厂商补齐。
- aliyun 聚合函数 `aggregateAliyunTraffic/Bps/HitRate`（243/256/271）已按 `parseAliyunTimeStamp` 按日期分桶，天然支持多日响应。
- 执行器并发 + 批量写已优化（前一 feature），API 调用量是剩余瓶颈。

### Urgency

- 无功能故障；夜间任务可完成，属规模放大优化。
- 不做的代价：域名越多 API 调用线性放大。
- 用户已确认粒度（保持 5min+区间，语义逐字一致）。

## Proposed Solution

aliyun 三个 fetch 的逐日循环改为**单次范围调用**（`StartTime=dates[0]`、`EndTime=dates[len(dates)-1]`、保持 `Interval="300"`），复用现有按日分桶聚合函数。每域名 API 调用从 90 次 → 3 次，聚合语义逐字不变（同为 5min 求和/取峰/均值）。

### Innovation Highlights

直白补齐与三厂商一致的区间模式，无创新点；价值在消除 aliyun 独有的 30× 调用放大。

## Requirements Analysis

### Key Scenarios

- 30 天整月：单次 DescribeDomainTrafficData/BpsData/HttpCodeData 范围调用，返回 ~8640 个 5min 模块，聚合函数按日分桶（与逐日调用结果一致）。
- 某日无数据：该日无模块/聚合为空，不写库（现有语义）。
- 响应体积：单次 ~8640 模块 vs 现 30×288——体积不变（总模块数相同），仅单响应变大。

### Non-Functional Requirements

- **行为等价**：每域名每日的流量求和 / bps 峰值 / 命中率均值与逐日调用逐字一致（同一 API、同一 Interval、同一聚合）。
- API 调用量：每域名 90 → 3。

### Constraints & Dependencies

- 仅改 `internal/shared/cloudx/aliyun/cdn_metrics.go`（3 个 fetch + 调用处）。
- 不动 tencent/huawei/volcano（已区间化）；不动执行器与 DAO（已优化）。
- `dates[0]`/`dates[len-1]` 为区间边界；`dayBoundsUTC` 仅用于逐日调用，范围调用直接用边界日期构造 Start/End。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 5min+区间折叠 | 本提案 | 语义逐字一致、与三厂商对齐 | 单响应 ~8640 模块（体积未变小） | Adopt（用户已确认） |
| 86400 日粒度 | 备选 | 响应小 | 日值厂商口径可能与 5min 聚合有细微差异 | Rejected：用户已选 5min |
| 维持逐日 90 次 | 备选 | 零风险 | 30× 调用放大 | Rejected：用户已确认折叠 |
| Do nothing | — | 零改动 | 调用放大持续 | Rejected |

## Scope

### In Scope

- `fetchTrafficByDay`/`fetchBpsByDay`/`fetchHitRateByDay` 三个函数改单次范围调用（或合并为 fetchTraffic/fetchBps/fetchHitRate），保持 Interval=300 与现有聚合函数。
- `GetDomainMetrics` 内 3 个 fetch 调用传整段 dates 区间。
- 单测：范围响应（多日模块）聚合 == 逐日聚合结果；无数据日；边界日期。

### Out of Scope

- tencent/huawei/volcano（已区间化）。
- 执行器并发/批量写（已完成）。
- 响应上限 fallback（若生产遇到阿里限制，另行处理）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 阿里单次范围响应过大/被截断 | Low-Medium | 单测构造多日响应验证聚合；若生产截断，回退分段范围（记录 fallback 方案） |
| 区间边界时区对齐差异 | Low | Start=dates[0] 00:00、End=dates[last] 23:59（沿用 dayBoundsUTC 边界语义）；单测断言边界日包含 |
| 聚合函数对多日响应的日期分桶回归 | Low | 单测对比逐日 vs 范围聚合结果逐日期相等 |

## Success Criteria

- [ ] aliyun 三个 fetch 无逐日 `for _, d := range dates` 循环（grep 无残留），单次范围调用
- [ ] 范围调用聚合结果与逐日调用**逐日期相等**（单测对比断言）
- [ ] 既有 `aliyun/cdn_metrics_test.go` 零改动通过（或新增用例）；`go test ./internal/shared/cloudx/aliyun/` 全绿
- [ ] 全仓 `go build ./...` 通过