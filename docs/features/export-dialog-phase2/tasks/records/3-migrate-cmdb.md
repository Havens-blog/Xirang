---
status: "completed"
started: "2026-09-20 12:22"
completed: "2026-09-20 12:32"
time_spent: "~10m"
---

# Task Record: 3 cmdb/instances 页迁移共享组件（InstanceVO 行类型）

## Summary
cmdb/instances 页导出弹窗迁到泛型共享组件 AssetExportDialog（InstanceVO 为泛型 T≠Asset 的首个消费者）：页面级 exportConfig: ExportFieldConfig<InstanceVO>，fields(10 项)/getValue(6 特例分支逐字)/filename('资源实例列表')/defaultFields(6 key) 全部逐字自原 components/ExportDialog.vue 迁移，零行为漂移；fetchAllExportRows(Promise<InstanceVO[]>) 衔接不变，进度/失败降级语义不变；删除本地 components/ExportDialog.vue（467 行）。提交 53f8106（git 显式 pathspec，仅 index.vue + 删 1 组件）。

## Changes

### Files Created
无

### Files Modified
- src/views/cmdb/instances/index.vue

### Key Decisions
- getValue 内 provider 分支沿用原弹窗的 utils/constants getProviderLabel 原函数（以别名 getProviderLabelRaw 导入），而非页内带 '-' 兜底的 getProviderLabel 包装——避免空 provider 从 '' 漂移为 '-'
- 原弹窗「当前过滤条件全部数据」按钮文案与共享组件「全部数据」不一致，按任务 2 同一诚实评估原则采用共享组件标准化文案
- commit 注释沿用页内「供 ExportDialog 调用」措辞不动，与任务 2（ecs 页）做法一致

## Test Results
- **Tests Executed**: Yes
- **Passed**: 512
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] cmdb/instances/index.vue 改用共享组件：fields/getValue/filename/defaultFields 与原组件逐项一致（6 特例分支逐字搬运）
- [x] T=InstanceVO 推断正常（泛型首个非 Asset 消费者，ExportFieldConfig<InstanceVO> 显式标注，vue-tsc 通过）
- [x] fetchAllRows（Promise<InstanceVO[]>）衔接不变；进度/失败降级语义不变
- [x] 本地 components/ExportDialog.vue 已删除
- [x] pnpm typecheck + vitest 通过；git diff 仅 cmdb/instances/index.vue + 删 1 组件

## Notes
typecheck 仅剩 src/views/logs/ 2 个在途错误（并发会话，任务允许忽略）；vitest 全量 512/512 通过（25 文件）。coverage 未由 runner 输出（vitest 未启用覆盖率收集），不作为门禁——重构任务维持既有覆盖，无新测试要求。并发会话脏文件（accounts/index.vue、components.d.ts、Step1Select.vue 等）未触碰、未提交。
