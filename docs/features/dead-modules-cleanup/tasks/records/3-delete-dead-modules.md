---
status: "completed"
started: "2026-09-09 12:46"
completed: "2026-09-09 12:53"
time_spent: "~7m"
---

# Task Record: 3 删除三死模块与三只 tmp 工具目录

## Summary
删除零外部 importer 的三个悬空模块（internal/task、internal/sync、internal/servicetree）与三只已留档的 tmp 工具目录（cmd/tmp-cas-diag、cmd/tmp-k8s-cred、cmd/tmp-k8s-diag）。删除单独成 commit dc3eee0（43 files, -7630），surgical staging 未混入在途变更。同名活跃包 internal/cam/{task,sync,servicetree} 完好，go build ./... 与 go vet ./... 全绿。

## Changes

### Files Created
无

### Files Modified
- internal/task (deleted)
- internal/sync (deleted)
- internal/servicetree (deleted)
- cmd/tmp-cas-diag (deleted)
- cmd/tmp-k8s-cred (deleted, untracked dir)
- cmd/tmp-k8s-diag (deleted)

### Key Decisions
- cmd/tmp-k8s-cred was untracked (no git history), so its removal leaves no git trace — removed from disk only
- 10 dangling comment references left untouched per Out of Scope; commit message lists all 10 with pointer to eval/followup.md and 待 follow-up 清除
- No Mongo data or index changes

## Test Results
- **Tests Executed**: Yes
- **Passed**: 62
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 六个目录均不存在 (internal/task, internal/sync, internal/servicetree, cmd/tmp-cas-diag, cmd/tmp-k8s-cred, cmd/tmp-k8s-diag)
- [x] import 语句级 grep 三个死模块路径零命中（排除 internal/cam/*）
- [x] go build ./... 与 go vet ./... 通过
- [x] 删除单独成 commit（仅暂存删除路径，未混在途变更）
- [x] commit 描述列明 10 处悬空注释清单并注明待 follow-up 清除

## Notes
Test metrics: go test -p 2 ./... => 62 packages ok, 4 FAIL are pre-existing baseline reds (internal/cam/cost/analysis, internal/cert/domain, tests/no-snapshot-guidance, tests/placeholder-fingerprint-backfill) — identical to the stated pre-task baseline, none introduced by this deletion; full-suite regression is owned by task 6. go test linker OOM on default parallelism; -p 2 used per project notes.
