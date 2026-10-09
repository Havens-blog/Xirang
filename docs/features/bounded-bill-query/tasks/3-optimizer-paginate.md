---
id: "3"
title: "optimizer 四处账单获取改分页遍历"
priority: "P1"
estimated_time: "1h"
complexity: "medium"
dependencies: [2]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 3: optimizer 四处账单获取改分页遍历

## Description
`internal/cam/cost/optimizer/service.go` 的 4 处 `ListUnifiedBills` 直接调用（150/220/231/304 行）改为经 Task 1 分页 helper 逐页处理，保持每账单恰处理一次、结果与全量加载等价、内存有界。

## Reference Files
- `docs/proposals/bounded-bill-query/proposal.md` — Proposed Solution, In Scope
- `internal/cam/cost/optimizer/service.go` — 4 处账单获取（150 总账单、220/231 storage/compute、304 优化扫描）

## Acceptance Criteria
- [ ] optimizer 4 处 `ListUnifiedBills` 直接调用全部替换为分页 helper（回调内执行原账单处理逻辑）
- [ ] 每账单恰处理一次，结果与全量等价——`optimizer/service_test.go` 既有测试零改动通过
- [ ] 各调用点的聚合/落库逻辑保持不动
- [ ] `go test ./internal/cam/cost/optimizer/...` 全绿
- [ ] `go build ./internal/cam/cost/optimizer/` 通过

## Hard Rules
仅修改 `internal/cam/cost/optimizer/service.go`。工作树并发会话脏文件与本作用域无关。

## Implementation Notes
- 4 处回调闭包各自捕获对应上下文（recommendations 积累、storage/compute 分流等），复用原循环体。
- 150 与 304 两处可能是同类「扫描全部账单」逻辑，回调可共享，但避免过度抽象——以行为等价为先。
- 行为等价由既有 optimizer 测试锁定。
