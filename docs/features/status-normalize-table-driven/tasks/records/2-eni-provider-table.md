---
status: "completed"
started: "2026-09-17 11:12"
completed: "2026-09-17 11:17"
time_spent: "~5m"
---

# Task Record: 2 eni.go NormalizeENIStatus 表驱动化

## Summary
eni.go NormalizeENIStatus 表驱动化零漂移重构：switch provider 派发 + 五厂商函数（aliyun/aws/huawei/tencent/volcano）收敛为 eniStatusMaps map[string]map[string]string；volcano/volcengine 两 key 路由 eniVolcanoENIStatusMap 同一映射；default: return status 与 status=="" 提前返回 unknown 保留；大小写敏感不做 ToLower；tencent "BINDbindingd "/"BINDUNBOUND"、huawei "ACTIVE" 等怪例逐字符抄表保留。既有 eni_test.go 零改动，新增 eni_table_test.go 一致性护栏测试（表值枚举合法性、provider key 集合、别名路由、抽样输入输出、大小写敏感与 fallback 语义）。

## Changes

### Files Created
- internal/shared/cloudx/types/eni_table_test.go

### Files Modified
- internal/shared/cloudx/types/eni.go

### Key Decisions
- 一致性测试放独立 eni_table_test.go 而非追加 eni_test.go，满足既有测试文件零改动硬约束
- volcano 与 aliyun 映射当前内容相同但保留两张独立表字面量，忠实对应原两个独立函数，避免未来单厂商调整时误联动
- assert.Same 不适用于 map 值，别名同映射断言改为 assert.Equal 深比较 + 行为层抽样
- 工作树含并发会话（cert/logquery）脏文件，与本任务作用域不相交，仅提交 eni 相关改动

## Test Results
- **Tests Executed**: Yes
- **Passed**: 45
- **Failed**: 0
- **Coverage**: 96.2%

## Acceptance Criteria
- [x] switch provider 派发替换为 map[provider]map[string]string 查找；五厂商函数 case 清零（合并进映射）
- [x] 逐字符对照原 switch 抄表，怪例原样保留（tencent BINDbindingd 尾空格/BINDUNBOUND、huawei ACTIVE 等）
- [x] volcano 与 volcengine 两个 provider key 路由同一映射；未知 provider default: return status 保留
- [x] status=="" 提前返回 ENIStatusUnknown 保留；未匹配返回原值保留
- [x] 既有 eni_test.go 零改动通过；新增表结构一致性测试（抽样输入输出、别名路由）
- [x] go test ./internal/shared/cloudx/types/... 全绿；go build ./internal/shared/cloudx/types/ 通过

## Notes
任务文件引用路径 internal/shared/cloudx/types/eni.go 实际位于 e-cam-service/internal/shared/cloudx/types/eni.go（与 proposal 中 4 switch 计数不符处均为引用面偏差，不影响改造目标）。golangci-lint 未安装，以 go vet 替代通过；gofmt 已通过。分支为 main（按规程仅警告）。coverage 96.2% 为 types 包整体（含 Task 1 status.go）。
