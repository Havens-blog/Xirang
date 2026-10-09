---
id: "2"
title: "AllocateCosts 账单获取改分页遍历"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: [1]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 2: AllocateCosts 账单获取改分页遍历

## Description
`internal/cam/cost/allocation/service.go` 的 `AllocateCosts`（135 行）把「一次性 `ListUnifiedBills` 全量加载 + 逐账单×规则嵌套处理」改为经 Task 1 的共享分页 helper 逐页处理，保持每账单恰处理一次、全量语义不变、内存有界。

## Reference Files
- `docs/proposals/bounded-bill-query/proposal.md` — Proposed Solution, Success Criteria
- `internal/cam/cost/allocation/service.go` — AllocateCosts（115-170：取账单 → 双层循环 matchAndAllocate/createUnmatchedAllocation → 积累 allocations）

## Acceptance Criteria
- [ ] AllocateCosts 的 `ListUnifiedBills` 直接调用替换为分页 helper，逐页回调内执行原「for bill × for rule」处理逻辑，accumulate allocations
- [ ] 每账单恰处理一次（无遗漏/重复），结果集与全量加载等价——`allocation/service_test.go` 既有测试零改动通过
- [ ] 处理完成后的落库逻辑（bulk insert allocations）保持不动
- [ ] `go test ./internal/cam/cost/allocation/...` 全绿
- [ ] `go build ./internal/cam/cost/allocation/` 通过

## Hard Rules
仅修改 `internal/cam/cost/allocation/service.go`。不动 matchAndAllocate/createUnmatchedAllocation 规则逻辑。工作树并发会话脏文件与本作用域无关。

## Implementation Notes
- 回调内复用原 `bills` 循环体（rules、defaultPolicy、now 为闭包捕获）。
- `hasDefaultPolicy` 等上下文在外层计算，回调只读。
- 行为等价由既有 allocation 测试锁定。
