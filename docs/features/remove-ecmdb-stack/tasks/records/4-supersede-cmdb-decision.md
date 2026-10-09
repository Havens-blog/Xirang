---
status: "completed"
started: "2026-10-08 21:33"
completed: "2026-10-08 21:38"
time_spent: "~5m"
---

# Task Record: 4 决策 0001 标废弃并登记移除 ecmdb 的新决策

## Summary
标废弃决策 0001（顶部横幅 + 状态字段指向 0003），新增替代决策 0003 记录移除 ecmdb+ecmdb-web 及依赖安全判定（能力已废弃 / eiam 走已发布 v1.11.0 / e-cam-service 零依赖 / nginx 根入口 301 /console/）

## Changes

### Files Created
- docs/decisions/0003-remove-ecmdb.md

### Files Modified
- docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md

### Key Decisions
无

## Document Metrics
0003 ~70 行含 7 节；0001 仅追加横幅与状态字段，原文保留

## Referenced Documents
- docs/proposals/remove-ecmdb-stack/proposal.md
- docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md
- docs/decisions/0002-watchalert-integration.md

## Review Status
final

## Acceptance Criteria
- [x] 决策 0001 的状态由「已决策」改为「已废弃」，并引用新决策 0003
- [x] 新决策 0003 记录：能力已废弃（用户 2026-10-08 确认）、eiam 走已发布模块 v1.11.0、e-cam-service 零依赖、nginx 根入口 301 /console/
- [x] docs/decisions/ 下无其他仍声明「保留 ecmdb」的落地决策

## Notes
0001 原文完整保留供审计，仅顶部加废弃横幅并将状态字段指向 0003；0002 仅以文件名引用 0001，非声明保留 ecmdb。
