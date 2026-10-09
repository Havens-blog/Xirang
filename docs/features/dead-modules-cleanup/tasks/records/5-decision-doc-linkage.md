---
status: "completed"
started: "2026-09-09 12:35"
completed: "2026-09-09 12:37"
time_spent: "~2m"
---

# Task Record: 5 决策文档 0001 Phase 4 补记 asset 归属决断项

## Summary
在决策文档 0001（工作区根 docs/decisions/，非 e-cam-service 仓库内）的 Phase 4 节以 append-only 方式补记 internal/asset 归属决断项：登记该目录直写 c_instance 的历史包袱（现仅存 cert 域覆盖率分母只读消费）、module.go 外壳已删与 doc.go 已正名的事实，并明确最终归属（收编 cert 域 / 随收敛删除直写路径后重定位）随 CMDB 收敛 Phase 4 一并决断，与 internal/asset/doc.go 的链接形成双向锚点。未改写决策 0001 任何既有结论。

## Changes

### Files Created
无

### Files Modified
- docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md

### Key Decisions
无

## Document Metrics
Phase 4 节追加 1 个补记块（blockquote，3 段），既有 5 个章节内容零改动

## Referenced Documents
- docs/features/dead-modules-cleanup/tasks/5-decision-doc-linkage.md
- docs/proposals/dead-modules-cleanup/proposal.md

## Review Status
final

## Acceptance Criteria
- [x] 决策文档 Phase 4 节含 internal/asset 归属决断记录（直写 c_instance 包袱 + 收编/删除决断随 Phase 4）
- [x] 记录注明坐标系：该文档位于工作区根 D:\Haven\docs\decisions\，不在 e-cam-service 仓库内

## Notes
纯追加编辑（Edit 定位 Phase 4 节末尾），符合 Implementation Notes 的 append-only 约束。工作区根非 git 仓库，按任务指示未提交。
