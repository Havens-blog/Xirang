---
status: "completed"
started: "2026-09-17 13:48"
completed: "2026-09-17 13:56"
time_spent: "~8m"
---

# Task Record: 3 rule_engine.compareValue 委托 MatchValue（Contains 放宽不敏感）

## Summary
rule_engine.compareValue 的 Eq/Contains/Regex 三分支委托 cam/domain.MatchValue（映射 eq→equals、contains→contains、regex→regex），Contains 语义由大小写敏感放宽为不敏感（唯一行为变化，已确认）；Ne/In/NotIn/Exists 原样保留，field mapping 未触碰；移除不再使用的 regexp import；新增 rule_engine_test.go 以 17 个子用例锁定各操作符语义并显式断言行为变化（"Prod" 模式匹配 "prod-01"）

## Changes

### Files Created
- e-cam-service/internal/cam/servicetree/service/rule_engine_test.go

### Files Modified
- e-cam-service/internal/cam/servicetree/service/rule_engine.go

### Key Decisions
- TDD：新增不敏感 contains 断言先 RED（旧实现下仅 2 个 contains 用例失败），委托后 GREEN，证明行为变化点被测试锁定
- Contains 放宽只增匹配不漏（超集语义），与提案 Success Criteria 一致
- gofmt 对 rule_engine.go 报 CRLF 行尾差异为包内既有条件（4 个未触碰文件同样被标记），未做整文件 LF 转换以保持 diff 干净、避免与并发会话冲突

## Test Results
- **Tests Executed**: Yes
- **Passed**: 17
- **Failed**: 0
- **Coverage**: 4.8%

## Acceptance Criteria
- [x] compareValue 的 Eq/Contains/Regex 分支改委托 camdomain.MatchValue（映射 eq→equals、contains→contains、regex→regex）
- [x] Ne/In/NotIn/Exists 分支原样保留（不迁移），In/NotIn 的 TrimSpace 语义保留原处
- [x] Contains 放宽不敏感——新增断言 "Prod" 模式匹配 "prod-01" 的 contains（先 RED 后 GREEN）
- [x] 既有 servicetree 测试回归全绿（go test ./internal/cam/servicetree/...，另含 tag/domain 回归）
- [x] go build ./internal/cam/servicetree/ 通过（另 go vet 通过）

## Notes
coverage 4.8% 为包级基线（runner 输出）；被改函数 compareValue 经 go tool cover -func 实测 100%。60% 包级目标受 Hard Rule（仅限 rule_engine.go + 其测试）约束不可达——包内 repository/web 等无测试层在范围外。regex 语义统一为 err==nil && matched（非法表达式返回 false，与 MatchValue 单点定义一致）；tag 46 用例与 domain 24 用例回归 ok。
