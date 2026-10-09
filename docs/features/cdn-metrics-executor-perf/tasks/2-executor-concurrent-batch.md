---
id: "2"
title: "执行器域名有界并发 + 攒批批量写"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: [1]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 2: 执行器域名有界并发 + 攒批批量写

## Description
`internal/cam/task/executor/sync_cdn_metrics.go` 主循环（230-275）两处优化：
1. 域名 `for` 循环改**有界并发**（semaphore 上限常量默认 5，复用仓内先例），每域名独立采集，错误仍按域名记录、`continue` 语义保留。
2. 内层逐指标 `UpsertMetric` 改：每域名采集结果攒批后调 `metricDAO.BulkUpsertMetrics`（每批 ~31 条），AccountID/Provider 注入、无数据跳过（Bytes==0 && Bandwidth==0 && HitRate<0）、错误日志语义全保留。

## Reference Files
- `docs/proposals/cdn-metrics-executor-perf/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `internal/cam/task/executor/sync_cdn_metrics.go` — 主循环（230-275）
- `internal/cam/task/executor/probe_service.go`（若存在）— 有界并发先例（semaphore 279 行附近）或 `sync_billing.go` WaitGroup 先例
- `internal/cam/task/executor/sync_cdn_metrics_test.go` — 既有测试（零改动为锁定）

## Acceptance Criteria
- [ ] 域名循环经 semaphore 有界并发，上限为包级常量（默认 5，可调）；并发数不超上限（测试断言）
- [ ] 逐条 `UpsertMetric` 调用清零，改为攒批 `BulkUpsertMetrics`（每域名一批）
- [ ] 行为等价：AccountID/Provider 注入、无数据跳过、单域名错误 continue + 日志保留；既有 sync_cdn_metrics 测试零改动通过
- [ ] `go test -race ./internal/cam/task/executor/` 全绿
- [ ] `go build ./internal/cam/task/executor/` 通过

## Hard Rules
仅修改 `internal/cam/task/executor/sync_cdn_metrics.go`。不触碰四厂商 `cdn_metrics.go` 适配器（api 逐日调用保留）。工作树并发会话（若在 CDN 相关文件活跃）与本作用域分文件提交，勿卷入。

## Implementation Notes
- 并发语义：`WaitGroup`/semaphore 组合或 errgroup（若包已依赖），上限 5；每域名 goroutine 内采集+攒批+批量写。
- 批量写错误：单域名批量写失败记日志 continue，不影响其它域名。
- `written` 计数在并发下需原子或互斥累加（sync/atomic 或复用现有计数方式）。
- 行为等价由既有 sync_cdn_metrics 测试锁定（runner 层 mock adapter/dao）。