---
status: "completed"
started: "2026-09-20 12:17"
completed: "2026-09-20 12:21"
time_spent: "~4m"
---

# Task Record: 2 ecs 页迁移共享组件（473 行零漂移收敛）

## Summary
ecs 页导出迁移共享 AssetExportDialog（零漂移）：index.vue 新增 exportConfig（22 字段逐字、getFieldValue 七特例闭包原样搬运、filename=虚拟机列表、defaultFields=原 10 项），fetchAllExportRows 闭包衔接不变；删除本地 components/ExportDialog.vue（473 行）

## Changes

### Files Created
无

### Files Modified
- src/views/compute/ecs/index.vue

### Key Decisions
- getValue 闭包逐字搬运，仅缩进适配，不做任何顺手优化（零漂移铁律）
- 原组件「全部数据」按钮文案为「当前过滤条件全部数据」而共享组件为「全部数据」，属纯文本差异非 UI 能力缺失，共享组件不在本任务范围，按共享组件现状保留并记录
- 并发会话脏文件（.env.development / element-theme.scss / components.d.ts / accounts/index.vue）未触碰，提交仅含 ecs 两个文件的显式 pathspec

## Test Results
- **Tests Executed**: Yes
- **Passed**: 512
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] ecs/index.vue 改用共享组件 :config，fields/getValue/filename/defaultFields 与原组件逐项一致
- [x] fetchAllRows 闭包衔接不变；进度/失败降级语义不变（由共享组件承载，diff 确认语义一致）
- [x] 本地 components/ExportDialog.vue 已删除；全前端无该文件引用
- [x] pnpm typecheck 通过（仅剩 cert Step1Select 与 logs 并发会话在途错误，本任务文件零错误）+ vitest 512/512 通过；git diff 仅 ecs/index.vue + 删 1 组件

## Notes
commit f103302 (e-cam-web)。typecheck 剩余 3 个错误均为并发会话在途文件（cert/changes/new/steps/Step1Select.vue、src/views/logs/ 两个测试），与本任务无关。共享组件泛型化（07f1529）使用前已验证。未运行 coverage（项目无 coverage 报告配置基线，指标为空）。
