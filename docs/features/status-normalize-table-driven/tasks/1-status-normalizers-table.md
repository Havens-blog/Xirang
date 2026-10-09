---
id: "1"
title: "status.go 八个 NormalizeXStatus 表驱动化"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: status.go 八个 NormalizeXStatus 表驱动化

## Description
`internal/shared/cloudx/types/status.go` 的 8 个状态归一化函数（NormalizeStatus/RDS/Redis/MongoDB/SecurityGroup/Image/Disk/Snapshot）
的 `strings.ToLower` + switch（26+ case，每 case 单行 return）收敛为 `map[string]string`（小写输入 → 规范状态），
保留 `default: return lower` 备选。行为零漂移（用户已确认口径）。

## Reference Files
- `docs/proposals/status-normalize-table-driven/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `internal/shared/cloudx/types/status.go` — 8 个 normalizer + 状态常量（Disk 示例 462-496、Snapshot 523-541）

## Acceptance Criteria
- [ ] 8 个 NormalizeXStatus 的 switch 全部替换为包级 `map[string]string` 查找（`statusXxxMap`），每函数保留 `strings.ToLower` 与 `default: return lower` 备选
- [ ] 逐 case 对照原 switch 抄表（含 `""`→unknown、多值 case 如 "in_use"/"in-use"/"attached" 全保留），无遗漏
- [ ] 谓词函数（IsRunningStatus/IsDiskAvailable 等）保持不动
- [ ] 既有 `status_test.go` 零改动通过；新增表结构一致性测试（抽样输入输出与 map 一致）
- [ ] `go test ./internal/shared/cloudx/types/...` 全绿；`go build ./internal/shared/cloudx/types/` 通过

## Hard Rules
仅修改 `internal/shared/cloudx/types/status.go`；新增一致性测试追加到现有 `status_test.go` 或独立 `status_table_test.go`。不触碰 eni.go（Task 2 范围）。

## Implementation Notes
- 表命名统一 `var statusXxxMap = map[string]string{...}`（如 disk → `diskStatusMap`）。
- 关键差异保留：各 normalizer 对 `""` 的返回（多为 Unknown）；未匹配一律 `return lower`（非原输入）。
- 腾讯/厂商怪例无（status.go 不涉及），但注意 SecurityGroup/Image 的多值 case 完整性。
- 与并发会话（cert/logquery）文件无重叠；如工作树脏，仅提交 status.go 相关改动。
