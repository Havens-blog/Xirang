---
status: "completed"
started: "2026-09-17 13:37"
completed: "2026-09-17 13:41"
time_spent: "~4m"
---

# Task Record: 1 cam/domain 新增共享 MatchValue 求值函数

## Summary
新增 internal/cam/domain/rule_match.go 共享 MatchValue(operator, actual, expected string) bool，覆盖 equals（精确）/contains/prefix/suffix（双侧 ToLower 不敏感）/regex（regexp.MatchString 且 err==nil 才 true），未知操作符返回 false；语义与 tag_rules.go:357 matchOperator 现口径一致。附 24 个子用例单测。

## Changes

### Files Created
- e-cam-service/internal/cam/domain/rule_match.go
- e-cam-service/internal/cam/domain/rule_match_test.go

### Files Modified
无

### Key Decisions
- contains/prefix/suffix 使用 strings.ToLower 双侧比较而非 EqualFold，保持与现网 tag 侧一致的 Unicode 语义（按 Hard Rule）
- regex 仅 err==nil && matched 为 true，非法模式返回 false
- 项目无 justfile，静态检查回退为 gofmt/go vet/go test

## Test Results
- **Tests Executed**: Yes
- **Passed**: 24
- **Failed**: 0
- **Coverage**: 100.0%

## Acceptance Criteria
- [x] internal/cam/domain/rule_match.go 存在，导出 MatchValue(operator, actual, expected string) bool
- [x] contains/prefix/suffix ToLower 双侧不敏感；equals 精确；regex err==nil 才 true；未知操作符 false
- [x] 单测覆盖各 op 大小写用例、equals 敏感、regex 转义/非法模式、未知 op false
- [x] go test ./internal/cam/domain/ 全绿；go build ./internal/cam/domain/ 通过

## Notes
覆盖率 100.0%（go test -cover，cam/domain 包）。gofmt 列出的格式差异均为 pre-existing 文件，新增两文件格式干净。未触碰 tag/servicetree（Task 2/3 委托范围）。
