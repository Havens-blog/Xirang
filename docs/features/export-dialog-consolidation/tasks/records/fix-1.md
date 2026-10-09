---
status: "completed"
started: "2026-09-20 10:07"
completed: "2026-09-20 10:14"
time_spent: "~7m"
---

# Task Record: fix-1 Fix: 处理 3 处内联导出 stub（禁用+诚实提示）

## Summary
Disabled 3 inline export button stubs with honest el-tooltip「暂未开放」: compute/image 批量导出, resource/tag 导出 (button lives in TagFilters.vue), iam/compare 对比报告导出. Removed the fake ElMessage stub handlers (including the misleading ElMessage.success in compare). No real export functionality implemented, per SC2.

## Changes

### Files Created
无

### Files Modified
- src/views/compute/image/index.vue
- src/views/resource/tag/index.vue
- src/views/resource/tag/components/TagFilters.vue
- src/views/iam/compare/index.vue

### Key Decisions
- TagFilters.vue edited in addition to the 3 listed source files because the resource/tag export button markup lives there; the stub handler and @export binding in index.vue were removed instead
- Buttons wrapped in <el-tooltip content="暂未开放"><span>…</span></el-tooltip> so the tooltip still shows over disabled buttons
- Removed now-unused stub handlers and the unused 'export' emit declaration to avoid dead code and TS noUnusedLocals errors

## Test Results
- **Tests Executed**: Yes
- **Passed**: 9
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] compute/image 批量导出 button disabled with honest 暂未开放 tooltip
- [x] resource/tag 导出 button disabled with honest 暂未开放 tooltip
- [x] iam/compare 对比报告导出 button disabled, fake ElMessage.success removed
- [x] No「导出功能开发中」residue remains in the 3 target sites
- [x] pnpm typecheck passes
- [x] Targeted vitest run (AssetExportDialog.test.ts, same export feature family) passes 9/9
- [x] Hard Rules respected: accounts/index.vue, src/views/logs/, backend files, components.d.ts untouched; no git add -A; explicit pathspec commit d856049

## Notes
Verification: pnpm typecheck (0 errors) + targeted vitest run of src/components/__tests__/AssetExportDialog.test.ts (9 passed, 0 failed) — the changed view files have no dedicated unit tests; full regression is dispatcher-verified per fix-task boundaries. Commit: d856049 in e-cam-web (main), 4 files, 23 insertions / 26 deletions.
