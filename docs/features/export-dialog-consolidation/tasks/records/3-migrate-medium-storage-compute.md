---
status: "completed"
started: "2026-09-20 10:35"
completed: "2026-09-20 10:35"
time_spent: ""
---

# Task Record: 3 中档存储/计算 5 页零漂移迁移共享组件

## Summary
中档存储/计算 5 页（oss/nas/disk/snapshot/security-group）零漂移迁移到共享 AssetExportDialog：每页 exportConfig 闭包逐字复制原 getFieldValue/availableFields/filename/defaultFields；nas 的 :metric-map 扩展经 getValue 闭包捕获 fsMetricMap；disk/snapshot/security-group 显式传 :selected-ids="[]" 保持「已选中」禁用；删除 5 个本地 ExportDialog.vue。typecheck 零错误、vitest 512/512、eslint 5 文件零告警。

## Changes

### Files Created
无

### Files Modified
- src/views/storage/oss/index.vue
- src/views/storage/nas/index.vue
- src/views/compute/disk/index.vue
- src/views/compute/snapshot/index.vue
- src/views/compute/security-group/index.vue

### Key Decisions
- nas metric-map 扩展：getValue 闭包捕获 fsMetricMap，行为等价，未新增共享组件 prop
- 共享组件 selectedIds 必填且 watch immediate 解引用 .length：disk/snapshot/security-group 显式传 :selected-ids="[]" 保持「已选中」禁用（不传会运行时崩溃）
- 零漂移铁律：getValue 逐字搬运原函数（oss storage_size/acl/versioning 格式化等），迁移中无行为改动

## Test Results
- **Tests Executed**: Yes
- **Passed**: 512
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 5 页改用共享 AssetExportDialog 并新增 :config prop
- [x] 每页 fields/getValue/filename/defaultFields 与本地组件逐项对照一致
- [x] nas metric-map 经闭包等价处理，未新增共享组件 prop
- [x] 5 个本地 ExportDialog.vue 已删除，fetchAllRows/进度语义保留
- [x] pnpm typecheck + vitest 通过，git diff 仅 5 页 index.vue + 删 5 个 ExportDialog.vue

## Notes
提交 bc28285（显式 pathspec，10 文件 +136/-653）。首次 submit 被质量门 just lint 拦截（e-cam-service 并发会话 EnvelopeCrypto 重构中间态，NewDiscoveryImportService 缺参），后端已收敛、go build ./... 通过后重建本记录重提；gofmt 5 文件为并发会话文件（logquery/cert）非本任务范围。
