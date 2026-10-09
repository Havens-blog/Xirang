---
id: "4"
title: "billing 三份 strVal 收敛为共享 DerefStr"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 4: billing 三份 strVal 收敛为共享 DerefStr

## Description
`internal/shared/cloudx/billing/{huawei,tencent,volcano}/adapter.go` 三份逐字节相同的
私有 `strVal(s *string) string`（nil → ""，否则返回 *s）收敛为 `billing` 包共享的
`DerefStr` helper，三个调用点改用它，删除私有函数。

## Reference Files
- `docs/proposals/cloudx-template-consolidation/proposal.md` — Proposed Solution, Success Criteria
- `internal/shared/cloudx/billing/huawei/adapter.go` — strVal 定义 + 调用点
- `internal/shared/cloudx/billing/tencent/adapter.go` — strVal 定义 + 调用点

## Acceptance Criteria
- [ ] `billing` 包新增 `DerefStr(s *string) string`（nil → ""，语义与旧 strVal 逐字一致）
- [ ] 三处 strVal 调用点改用 `DerefStr`，三个私有 `func strVal` 定义已删除
- [ ] `grep -rn "func strVal" internal/` 无残留
- [ ] `DerefStr` 单测覆盖 nil / 空串 / 非空串
- [ ] 全仓 `go build ./...` + `go test ./internal/shared/cloudx/billing/...` 通过

## Hard Rules
仅修改以下文件：新增 `internal/shared/cloudx/billing/derefstr.go`（+ 单测）；修改 `billing/huawei/adapter.go`、`billing/tencent/adapter.go`、`billing/volcano/adapter.go` 的调用点与删除私有函数。

## Implementation Notes
- `DerefStr` 放 billing 包根（与 adapter 子包同层，注意 billing 根包的既有文件命名）。
- 三个调用点直接替换为 `billing.DerefStr(x)`；若调用处为同包内 strVal（adapter.go 与 billing 根包不同包），注意包限定符。
- 纯行为等价，不触碰其它格式化逻辑。
