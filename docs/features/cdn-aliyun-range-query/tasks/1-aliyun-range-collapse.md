---
id: "1"
title: "aliyun CDN 三 fetch 改单次范围调用"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 1: aliyun CDN 三 fetch 改单次范围调用

## Description
`internal/shared/cloudx/aliyun/cdn_metrics.go` 的 `fetchTrafficByDay`/`fetchBpsByDay`/`fetchHitRateByDay`
三个逐日循环（`for _, d := range dates` + `Interval="300"`）改为**单次范围调用**：
`StartTime=dates[0]` 00:00、`EndTime=dates[len-1]` 23:59，保持 `Interval="300"`（5min）。
复用现有按日分桶聚合函数（`aggregateAliyunTraffic/Bps/HitRate`，已 `parseAliyunTimeStamp` 按日分桶）。
每域名 API 90→3 次，聚合语义逐字不变（与三厂商区间模式对齐）。

## Reference Files
- `docs/proposals/cdn-aliyun-range-query/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `internal/shared/cloudx/aliyun/cdn_metrics.go` — fetchTrafficByDay(72)/fetchBpsByDay(101)/fetchHitRateByDay(125) + 聚合(243/256/271) + GetDomainMetrics(18)
- `internal/shared/cloudx/tencent/cdn_metrics.go` — 区间查询先例（fetchCdnMetricRange，Start/End 整段 + 聚合）

## Acceptance Criteria
- [ ] 三个 fetch 的 `for _, d := range dates` 循环清零（改为单次范围调用，区间 Start=dates[0]/End=dates[len-1]），`dayBoundsUTC` 不再用于逐日调用
- [ ] `Interval="300"` 保留；`GetDomainMetrics` 内 3 个 fetch 传整段 dates
- [ ] 单测：构造多日范围响应，断言范围聚合结果与逐日聚合**逐日期相等**；无数据日；边界日期包含
- [ ] 既有 `aliyun/cdn_metrics_test.go` 零改动通过（或新增范围用例）；`go test ./internal/shared/cloudx/aliyun/` 全绿
- [ ] `go build ./internal/shared/cloudx/aliyun/` 通过

## Hard Rules
仅修改 `internal/shared/cloudx/aliyun/cdn_metrics.go`（+ 测试文件）。不触碰 tencent/huawei/volcano 与执行器/DAO。工作树并发会话脏文件与本作用域无关。

## Implementation Notes
- 范围边界语义：沿用 `dayBoundsUTC(d)` 对首/末日构造（Start=dates[0] 00:00、End=dates[last] 23:59），保持时区一致。
- 聚合函数无需改动（已按日期分桶）；若 range 响应含非目标日期模块，聚合天然忽略。
- 风险提示：单次响应 ~8640 模块；若生产遇阿里截断，fallback 为分段范围（本任务不实现，仅代码注释标注）。
