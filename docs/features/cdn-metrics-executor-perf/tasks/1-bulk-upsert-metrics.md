---
id: "1"
title: "cdnMetricDAO 新增 BulkUpsertMetrics 批量写"
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

# 1: cdnMetricDAO 新增 BulkUpsertMetrics 批量写

## Description
`internal/cam/repository/dao/cdn_metric.go` 新增 `BulkUpsertMetrics(ctx, metrics []types.CDNMetric) error`：
用 mongo `BulkWrite`（`UpdateOne` + upsert，filter 用唯一键 `account+domain+date`）替代逐条 `UpsertMetric`。
幂等、空批直接返回 nil、错误回传。既有 `UpsertMetric` 保留（兼容其它调用方）。

## Reference Files
- `docs/proposals/cdn-metrics-executor-perf/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `internal/cam/repository/dao/cdn_metric.go` — cdnMetricDAO + UpsertMetric（75 行）
- `internal/cam/task/executor/sync_cdn_metrics_test.go` — executor 测试（如有批量相关可参照）

## Acceptance Criteria
- [ ] `cdnMetricDAO.BulkUpsertMetrics(ctx, []types.CDNMetric) error` 存在：BulkWrite UpdateOne+upsert，filter 用唯一键（account_id+domain+date）
- [ ] 空 metrics 直接返回 nil；BulkWrite 错误回传（含部分失败语义符合 mongo BulkWrite）
- [ ] 既有 `UpsertMetric` 单条方法保留（兼容）
- [ ] 单测覆盖：空批、单批多条、幂等（同键两次写入不重复）、错误注入回传
- [ ] `go test ./internal/cam/repository/dao/` 全绿；`go build ./internal/cam/repository/dao/` 通过

## Hard Rules
仅修改 `internal/cam/repository/dao/cdn_metric.go`（+ 单测文件）。唯一键字段名与现有 upsert 一致（参照 UpsertMetric 现 filter 写法）。工作树并发会话脏文件与本作用域无关。

## Implementation Notes
- 复用 `UpsertMetric` 现有 filter 构造（唯一键 account+domain+date），逐条转 BulkWrite model 数组。
- mongo-driver 的 `mongo.NewWriteModels` 或显式 []mongo.WriteModel 均可。
- 供 Task 2（执行器攒批调用）使用。