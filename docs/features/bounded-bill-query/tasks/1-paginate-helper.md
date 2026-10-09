---
id: "1"
title: "cost/repository 新增共享分页 helper"
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

# 1: cost/repository 新增共享分页 helper

## Description
`internal/cam/cost/repository` 新增共享分页 helper：封装「按 Offset/Limit 分页遍历 `ListUnifiedBills`，逐页回调处理，取回数 < 页大小即早停」的循环。页大小常量 500。DAO 接口签名不动（复用既有 Offset/Limit）。

## Reference Files
- `docs/proposals/bounded-bill-query/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `internal/cam/cost/repository/types.go` — BillRepository 接口（ListUnifiedBills 已支持 Offset/Limit）
- `internal/cam/cost/repository/dao/bill.go` — ListUnifiedBills 实现（142-162，Limit>0 才 SetLimit）

## Acceptance Criteria
- [ ] `cost/repository/paginate.go` 存在，导出分页函数（如 `PaginateUnifiedBills`）：入参 repository 接口 + filter + 页大小 + 每页回调；页循环用 Offset=页序×页大小、Limit=页大小；取回数 < 页大小即停
- [ ] 页大小常量 500（包级可调）；每页回调返回 error 时中止并回传
- [ ] `cost/repository/paginate_test.go` 单测覆盖：单页 / 多页 / 恰好整除 / 尾页早停 / 空结果 / 回调错误中止；每 bill 恰回调一次（mock DAO 计数断言）
- [ ] 既有 `go test ./internal/cam/cost/repository/...` 全绿
- [ ] `go build ./internal/cam/cost/repository/` 通过

## Hard Rules
仅新增 `internal/cam/cost/repository/paginate.go` + `paginate_test.go`；不改动 DAO 与既有接口。工作树可能的并发会话脏文件与本作用域无关。

## Implementation Notes
- 排序由 DAO 内部既有 `billing_date desc, ctime desc` 保证，分页偏移稳定。
- 回调签名建议 `func([]domain.UnifiedBill) error`，调用方在回调内处理/聚合。
- 零行为漂移：全量语义由「循环取完所有页」保证，回调不落则等价于现状全量。
