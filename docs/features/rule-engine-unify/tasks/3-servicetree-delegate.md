---
id: "3"
title: "rule_engine.compareValue 委托 MatchValue（Contains 放宽不敏感）"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: [2]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 3: rule_engine.compareValue 委托 MatchValue（Contains 放宽不敏感）

## Description
`internal/cam/servicetree/service/rule_engine.go` 的 `compareValue` 中 **Eq/Contains/Regex** 三个分支改为委托
`cam/domain.MatchValue`（操作符名映射：`stdomain.OperatorEq("eq")`→"equals"、`OperatorContains("contains")`→"contains"、
`OperatorRegex("regex")`→"regex"）。**Contains 语义由大小写敏感放宽为不敏感**（唯一行为变化，已确认）。
Ne/In/NotIn/Exists 分支原样保留。

## Reference Files
- `docs/proposals/rule-engine-unify/proposal.md` — Proposed Solution, Success Criteria
- `internal/cam/servicetree/service/rule_engine.go` — compareValue（268-300）
- `internal/cam/servicetree/domain/rule.go` — 操作符常量（eq/ne/contains/regex/in/not_in/exists）

## Acceptance Criteria
- [ ] compareValue 的 Eq/Contains/Regex 分支改委托 `camdomain.MatchValue`（映射 eq→equals、contains→contains、regex→regex）
- [ ] Ne/In/NotIn/Exists 分支原样保留（不迁移）
- [ ] Contains 放宽不敏感——`rule_engine` 新增/更新断言：`"Prod"` 模式匹配 `"prod-01"` 的 contains
- [ ] 既有 servicetree 测试回归全绿（`go test ./internal/cam/servicetree/...`）
- [ ] `go build ./internal/cam/servicetree/` 通过

## Hard Rules
仅修改 `internal/cam/servicetree/service/rule_engine.go`（+ 必要时 rule_engine 测试）。不触碰 field mapping（rule_engine.go:236）与 Ne/In/NotIn/Exists。工作树并发会话脏文件无关。

## Implementation Notes
- 委托前确认 servicetree 的 contains 目前是否被分类规则实际使用、放宽影响面（只增匹配不漏）。
- In/NotIn 的 TrimSpace 语义保留原处（不经 MatchValue）。
- 行为变化点必须在测试中显式断言，防止回归。
