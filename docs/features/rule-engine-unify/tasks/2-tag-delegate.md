---
id: "2"
title: "tag_rules.matchOperator 委托 cam/domain.MatchValue"
priority: "P0"
estimated_time: "30m"
complexity: "low"
dependencies: [1]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 2: tag_rules.matchOperator 委托 cam/domain.MatchValue

## Description
`internal/cam/tag/tag_rules.go` 的 `matchOperator(operator, fieldVal, pattern string)` 改为委托
`cam/domain.MatchValue`（tag 侧操作符名 equals/contains/prefix/suffix/regex 与 MatchValue 规范名一致，语义零变化）。

## Reference Files
- `docs/proposals/rule-engine-unify/proposal.md` — Proposed Solution, In Scope
- `internal/cam/tag/tag_rules.go` — matchOperator（358-373，现 contains/prefix/suffix 已不敏感）

## Acceptance Criteria
- [ ] `matchOperator` 函数体替换为 `return camdomain.MatchValue(operator, fieldVal, pattern)`（或直接删除该函数、调用点改用 MatchValue）
- [ ] tag 行为零变化——既有 tag 测试零改动通过
- [ ] `go test ./internal/cam/tag/...` 全绿；`go build ./internal/cam/tag/` 通过

## Hard Rules
仅修改 `internal/cam/tag/tag_rules.go`。不触碰 conditionToFilter 与 field mapping。工作树并发会话脏文件无关。

## Implementation Notes
- tag 操作符名与 MatchValue 规范名已一致（equals/contains/prefix/suffix/regex），无需映射。
- 删除 matchOperator 私有函数后检查其唯一调用点，直接改调 `camdomain.MatchValue`。
