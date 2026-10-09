---
status: "completed"
started: "2026-09-09 12:25"
completed: "2026-09-09 12:35"
time_spent: "~10m"
---

# Task Record: 1 记录清理前基线（build/vet/test 失败集合留档）

## Summary
记录 dead-modules-cleanup 清理前基线：在 D:\Haven\e-cam-service 运行 go build ./... 与 go vet ./...（均通过），全量 go test -count=1 ./... 单次连续运行得到权威失败集合（4 包 6 测试），连同运行日期、脏工作区（79 项在途变更）说明与首跑 OOM 作废说明一并留档至 docs/features/dead-modules-cleanup/baseline-record.md（中文）。未修改任何代码。

## Changes

### Files Created
- docs/features/dead-modules-cleanup/baseline-record.md

### Files Modified
无

### Key Decisions
无

## Document Metrics
~70 lines across 6 sections

## Referenced Documents
- docs/proposals/dead-modules-cleanup/proposal.md

## Review Status
final

## Acceptance Criteria
- [x] 在 e-cam-service 运行 go build ./... 与 go vet ./... 并留档结果
- [x] 运行 go test -count=1 ./... 单次连续运行，失败测试清单逐条留档（含 TestBsonTagsMatchSchema/K8sCredential）
- [x] 留档文件记录执行时间（2026-09-09）与工作区状态说明（79 项在途变更脏工作区）

## Notes
首跑 go test（默认并行度）因 Windows 工具链 OOM（errno=1455）致 8 个包 [build failed]，测试未实际执行，按瞬态环境失败作废不作基线；以 go test -count=1 -p 2 ./... 重跑为权威基线（全量范围、无 -run 过滤）。失败集合：internal/cam/cost/analysis TestGetCostSummary_DateRanges；internal/cert/domain TestBsonTagsMatchSchema/K8sCredential（提案已登记存量红灯）；tests/no-snapshot-guidance TestNoSnapshotGuidance_Smoke、TestStep2_TriggerScan_Success（200 vs 202）；tests/placeholder-fingerprint-backfill TestStep4_BackfillReferences_WrongBackfillRescanRecoverable、TestStep5_VerifyReferenceList_PartialFailureRescanRefresh（200 vs 202）。两次运行共同执行项结果一致，无 flaky 判定。除本次 4 包外，其余失败均为 OOM 作废所致、第 2 次运行已通过。任务 6 验收比对以该失败集合为子集判据基准。
