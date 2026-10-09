---
status: "completed"
started: "2026-09-17 11:03"
completed: "2026-09-17 11:10"
time_spent: "~7m"
---

# Task Record: 1 status.go 八个 NormalizeXStatus 表驱动化

## Summary
status.go 剩余 4 个 NormalizeXStatus switch（SecurityGroup/Image/Disk/Snapshot）零漂移收敛为包级 map[string]string 查找（securityGroupStatusMap/imageStatusMap/diskStatusMap/snapshotStatusMap），每函数保留 strings.ToLower 与 default: return lower 兜底；谓词函数不动；既有 status_test.go 零改动通过，新增 status_table_test.go 表结构一致性测试（5 个测试）。注：NormalizeStatus/RDS/Redis/MongoDB 在当前代码中已是 map 实现（与 proposal 证据面不符），零改动保留。

## Changes

### Files Created
- internal/shared/cloudx/types/status_table_test.go

### Files Modified
- internal/shared/cloudx/types/status.go

### Key Decisions
- 仅转换仍为 switch 的 4 个 normalizer；已表驱动的 4 个（Status/RDS/Redis/MongoDB）按零漂移与 surgical 原则保持不动
- gofmt 报告的整文件差异为仓库级 CRLF 预存现象，未重写行尾以避免污染 diff
- golangci-lint 未安装、make lint 存在预存 shell 引号问题，以 go vet（通过）替代

## Test Results
- **Tests Executed**: Yes
- **Passed**: 40
- **Failed**: 0
- **Coverage**: 93.3%

## Acceptance Criteria
- [x] 8 个 NormalizeXStatus 全部 map 查找（statusXxxMap），保留 ToLower 与 default: return lower
- [x] 逐 case 对照抄表（含 ""→unknown、多值 case），无遗漏
- [x] 谓词函数保持不动
- [x] status_test.go 零改动通过；新增表结构一致性测试
- [x] go test ./internal/shared/cloudx/types/... 全绿；go build 通过

## Notes
coverage 93.3%（types 包）。工作树含并发会话（cert/logquery）脏文件，与本任务作用域不相交，仅提交 status.go 相关改动。
