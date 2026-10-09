---
status: "completed"
started: "2026-09-17 11:44"
completed: "2026-09-17 11:49"
time_spent: "~5m"
---

# Task Record: 3 optimizer 四处账单获取改分页遍历

## Summary
optimizer/service.go 的 4 处 ListUnifiedBills 直接全量调用（原 150/220/231/304 行）全部替换为经 repository.PaginateUnifiedBills 分页遍历：detectLowCPUInstances、detectUnattachedDisks（storage+compute 两处）、detectOnDemandConvert 的原账单循环体逐段移入分页回调闭包，聚合 map 与建议构建逻辑保持不动，错误包装文案不变，每账单恰处理一次、内存有界（每页 500）。

## Changes

### Files Created
无

### Files Modified
- e-cam-service/internal/cam/cost/optimizer/service.go

### Key Decisions
- 复用 allocation/service.go 同款模式：PaginateUnifiedBills(ctx, dao, filter, 0, closure)，pageSize=0 走默认 500
- detectUnattachedDisks 保持原有 DAO 查询顺序（storage 先于 compute），仅将循环体移入各自闭包，避免引入额外行为差异
- 行为等价由既有 service_test.go 零改动锁定（mock 返回量 < 页大小，单页早停语义与全量等价）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 7
- **Failed**: 0
- **Coverage**: 89.7%

## Acceptance Criteria
- [x] optimizer 4 处 ListUnifiedBills 直接调用全部替换为分页 helper（回调内执行原账单处理逻辑）
- [x] 每账单恰处理一次，结果与全量等价——optimizer/service_test.go 既有测试零改动通过
- [x] 各调用点的聚合/落库逻辑保持不动
- [x] go test ./internal/cam/cost/optimizer/... 全绿
- [x] go build ./internal/cam/cost/optimizer/ 通过

## Notes
仓库无 justfile，静态检查用等效命令：go build + gofmt + go vet（golangci-lint 未安装）。gofmt -l 标记的 3 个 optimizer 文件为预存在 CRLF 行尾差异（含未改动的测试文件），非本次引入。硬规则遵守：仅修改 optimizer/service.go。测试命令 go test ./internal/cam/cost/optimizer/... -count=1 -cover，覆盖率 89.7%（目标 60%）。
