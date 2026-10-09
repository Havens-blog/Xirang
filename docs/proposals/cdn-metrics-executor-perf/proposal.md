---
created: "2026-09-19"
author: "Haven"
status: Approved
intent: "enhancement"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: CDN 指标采集执行器优化（有界并发 + 批量写）

## Problem

`internal/cam/task/executor/sync_cdn_metrics.go` 的指标采集执行器存在双重 N+1：**按域名串行**采集（每个域名完整跑完才下一个），且**每指标逐条单写** Mongo（`UpsertMetric`）。规模 = 账号 × 域名 × 天数，夜间任务 wall-time 与库写入压力不可控。

### Evidence

- 执行器主循环（`sync_cdn_metrics.go:240-272`）：`for _, inst := range instances` 串行调 `GetDomainMetrics`，失败单域名 `continue`；内层 `for _, m := range metrics` 每指标一次 `e.metricDAO.UpsertMetric`（100 域名 × 31 天 ≈ 3100 次单条 Mongo 写）。
- 执行器层已有先例：`probe_service.go:279` 用 semaphore 做有界并发。
- 唯一键已多账号化（3cea6d7：`account+domain+date`），批量写可幂等。
- 适配器逐日 API（90 次/域名）**不在本提案范围**（用户已确认，涉厂商 API 语义与活跃迭代冲突）。

### Urgency

- 无功能故障，采集慢属规模放大问题。不做的代价：域名/账号越多，夜间任务越长，Mongo 写放大持续。
- 用户已确认仅执行器层范围。

## Proposed Solution

执行器 `sync_cdn_metrics.go` 两处优化，行为等价（同指标、同跳过/错误日志逻辑）：
1. **域名循环有界并发**：semaphore 限制并发域名数（~5），每域名独立采集；错误仍按域名记录。
2. **批量写**：DAO 新增 `BulkUpsertMetrics(ctx, []types.CDNMetric)`（mongo BulkWrite，upsert 语义，幂等），执行器按域名采集结果攒批写入（每批 ~31 条），替代逐条单写。

### Innovation Highlights

直白并发 + 批量，无创新点；价值在复用仓内既有 semaphore 先例与既有唯一键。

## Requirements Analysis

### Key Scenarios

- 100 域名：并发 5 轮流采集，批量写 100 批（每批 ~31 条）替代 3100 次单写。
- 单域名失败：不影响其它域名（现有 continue 语义保留）。
- 无数据日（Bytes==0 && Bandwidth==0 && HitRate<0）：跳过不写（现有语义保留）。

### Non-Functional Requirements

- **行为等价**：写入的指标集合与现状一致（含跳过逻辑、AccountID/Provider 注入）；错误日志逐条保留。
- 并发上限有意义（受厂商限流约束），不可无界并发。

### Constraints & Dependencies

- 仅改 `internal/cam/task/executor/sync_cdn_metrics.go` + `internal/cam/repository/dao/cdn_metric.go`（新增批量方法）。
- 不改 `internal/shared/cloudx/*/cdn_metrics.go`（四厂商适配器逐日 API 保留）。
- 复用 `golang.org/x/sync/errgroup` 或自建 semaphore？——执行器包若未引入 errgroup，用仓内现有模式（`probe_service.go:279` semaphore 或 `sync.WaitGroup`）。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 有界并发 + 批量写（执行器层） | 本提案 | 行为等价、改动聚焦、复用先例 | 适配器 90 次 API 未动（收益打折） | Adopt（用户已确认） |
| 含适配器区间查询折叠（90→3 API） | 备选 | 收益最大 | 涉四厂商 API 语义、响应大小、活跃迭代冲突 | Deferred：用户排除，后续专项 |
| 仅批量写 | 备选 | 最安全 | 串行瓶颈未解 | Rejected：用户已选并发+批量 |
| Do nothing | — | 零改动 | N+1 持续 | Rejected |

## Scope

### In Scope

- 执行器域名循环改有界并发（semaphore 上限常量，默认 5）。
- DAO `cdnMetricDAO` 新增 `BulkUpsertMetrics(ctx, metrics)`（BulkWrite + upsert）。
- 执行器按域名攒批调用批量写（AccountID/Provider 注入、无数据跳过、错误日志保留）。
- 单测：批量写方法（幂等/空批/错误回传）+ 执行器并发路径（有界：并发数不超上限）。

### Out of Scope

- 四厂商适配器逐日 API（区间查询折叠，另行专项）。
- 采集触发频率/调度逻辑、指标口径（hit_flux 等）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 并发引入竞态（共享写入/日志） | Medium | 每域名独立写批、semaphore 上限；`go test -race` 相关包 |
| 批量写改变幂等/跳过语义 | Low | BulkUpsert 用 upsert 语义 + 唯一键；既有 cdn_metric 测试回归 |
| 并发上限过低/过高 | Low | 常量可调；与厂商限流（QPS~5）对齐 |
| 与并发会话冲突（CDN 指标代码活跃区） | Medium | 仅触碰 sync_cdn_metrics.go + cdn_metric.go；pre-check 独立验证；分文件提交 |

## Success Criteria

- [ ] 执行器域名循环经 semaphore 有界并发，上限常量可调（默认 5）。
- [ ] `cdnMetricDAO.BulkUpsertMetrics` 存在（BulkWrite + upsert，幂等），单测覆盖空批/多批/错误回传。
- [ ] 执行器逐条 `UpsertMetric` 调用清零（改批量），无数据跳过与错误日志语义保留。
- [ ] `go test -race ./internal/cam/task/executor/ ./internal/cam/repository/dao/` 全绿；既有 sync_cdn_metrics 测试零改动通过。
- [ ] 全仓 `go build ./...` 通过。