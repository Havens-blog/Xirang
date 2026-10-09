---
id: "1"
title: "cam/domain 新增共享 MatchValue 求值函数"
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

# 1: cam/domain 新增共享 MatchValue 求值函数

## Description
`internal/cam/domain` 新增共享规则操作符求值函数 `MatchValue(operator, actual, expected string) bool`，
覆盖 5 个操作符：**equals**（精确）、**contains**/**prefix**/**suffix**（ToLower 双侧后比较，大小写不敏感）、**regex**（`regexp.MatchString`，err==nil && matched）。
操作符命名以 tag 侧为准（equals/contains/prefix/suffix/regex）。

## Reference Files
- `docs/proposals/rule-engine-unify/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `internal/cam/tag/tag_rules.go` — matchOperator（358，现 contains/prefix/suffix 已不敏感，是 MatchValue 的语义基准）
- `internal/cam/servicetree/domain/rule.go` — servicetree 操作符常量（eq/contains/regex 等）

## Acceptance Criteria
- [ ] `internal/cam/domain/rule_match.go` 存在，导出 `MatchValue(operator, actual, expected string) bool`
- [ ] contains/prefix/suffix 用 `strings.ToLower` 双侧后比较（不敏感）；equals 精确；regex 用 `regexp.MatchString` 且 err==nil 才 true；未知操作符 false
- [ ] `MatchValue` 单测覆盖：各 op 大小写用例（"Prod" 匹配 "prod-01"）、equals 敏感、regex 转义/非法模式、未知 op false
- [ ] `go test ./internal/cam/domain/` 全绿；`go build ./internal/cam/domain/` 通过

## Hard Rules
仅新增 `internal/cam/domain/rule_match.go` + `rule_match_test.go`。工作树并发会话脏文件与本作用域无关。

## Implementation Notes
- 语义基准是 tag_rules.go 现口径（不敏感），不得用 EqualFold 改变 Unicode 语义——保持与现网一致的 ToLower 行为。
- 供 Task 2（tag 委托）与 Task 3（servicetree 委托）使用。
