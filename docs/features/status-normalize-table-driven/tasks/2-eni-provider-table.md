---
id: "2"
title: "eni.go NormalizeENIStatus 表驱动化"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: [1]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 2: eni.go NormalizeENIStatus 表驱动化

## Description
`internal/shared/cloudx/types/eni.go` 的 `NormalizeENIStatus(provider, status)` 的 `switch provider` 派发
（85-104 行）+ 五个厂商函数（aliyun/aws/huawei/tencent/volcano，106-195+，各为纯 case→枚举 switch）
收敛为 `map[string]map[string]string`（provider → 原始状态串 → 规范状态）。
volcano/volcengine 别名保留（两 key 路由同一映射）；`default: return status` 保留。
**大小写敏感保留**（匹配原始厂商状态串 Available/ACTIVE/in-use，绝不做 ToLower 归一）。

## Reference Files
- `docs/proposals/status-normalize-table-driven/proposal.md` — Proposed Solution, Key Scenarios, Success Criteria
- `internal/shared/cloudx/types/eni.go` — NormalizeENIStatus 派发（85-104）+ 五厂商函数（106-195）

## Acceptance Criteria
- [ ] `switch provider` 派发替换为 `map[provider]map[string]string` 查找；五厂商函数 case 清零（合并进映射）
- [ ] 逐字符对照原 switch 抄表，怪例原样保留（如 tencent `"BINDbindingd"`/`"BINDbindingd "` 带尾空格、`"BINDUNBOUND"`→Available、huawei `"ACTIVE"`→InUse 等）
- [ ] volcano 与 volcengine 两个 provider key 路由同一映射；未知 provider `default: return status` 保留
- [ ] `status == ""` 提前返回 ENIStatusUnknown 保留；未匹配返回原值保留
- [ ] 既有 `eni_test.go` 零改动通过；新增表结构一致性测试（抽样输入输出、别名路由）
- [ ] `go test ./internal/shared/cloudx/types/...` 全绿；`go build ./internal/shared/cloudx/types/` 通过

## Hard Rules
仅修改 `internal/shared/cloudx/types/eni.go`；新增一致性测试追加到现有 `eni_test.go` 或独立 `eni_table_test.go`。依赖 Task 1（同包顺序执行避免测试冲突）。

## Implementation Notes
- 表命名 `var eniStatusMaps = map[string]map[string]string{ "aliyun": {...}, "aws": {...}, ... }`。
- 五个厂商函数删除后，注意 ENI 状态常量（ENIStatusAvailable 等）仍被表值与其它代码使用，保留。
- tencent 的 `"BINDbindingd "`（尾空格）必须原样出现在 key 中——这是零漂移硬约束。
- 与并发会话（cert/logquery）文件无重叠；如工作树脏，仅提交 eni.go 相关改动。
