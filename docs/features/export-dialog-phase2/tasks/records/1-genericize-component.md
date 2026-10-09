---
status: "completed"
started: "2026-09-20 12:02"
completed: "2026-09-20 12:15"
time_spent: "~13m"
---

# Task Record: 1 AssetExportDialog 行类型泛型化（generic=T）

## Summary
AssetExportDialog 行类型泛型化：组件声明 generic="T extends { id: number } = any"，props 变为 instances: T[] / fetchAllRows?: (onProgress?) => Promise<T[]> / config: ExportFieldConfig<T>（getValue: (row: T, key: string) => string）；独立 <script> 块的 ExportFieldConfig 同步泛型化为 ExportFieldConfig<T = any> 并保持导出（默认 any 为既有 15 页裸标注兜底，新接入页建议显式标注）。既有 15 页用法零改动通过 typecheck（T 推断为 Asset），9 单测零改动语义不变通过，全量 vitest 512/512。未迁任何页面（任务 2/3/4 负责），未触碰 accounts/index.vue、logs、components.d.ts 等并发脏文件。

## Changes

### Files Created
无

### Files Modified
- src/components/AssetExportDialog.vue

### Key Decisions
- 组件级泛型加约束 T extends { id: number }：selectedIds.includes(row.id) 过滤逻辑零改动且类型安全（已核实 InstanceVO/InstanceVO cmdb.ts:205、CloudUser iam.ts:14 均有 id: number）
- ExportFieldConfig 泛型默认 T = any 而非 Asset：15 页既有裸 ExportFieldConfig 标注（含 ExportFieldConfig['getValue'] 与独立 const exportConfig 注解）零改动编译，且对后续 InstanceVO/CloudUser 页迁移最大宽容
- 组件名保持 AssetExportDialog 不变，文件头注释说明其已不只服务 Asset 行类型，避免 15 页 import 变动
- 测试文件完全零改动（9/9 通过），git diff 仅共享组件一个文件（commit 07f1529）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 512
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 组件声明 generic="T"，props 变为 instances: T[]、fetchAllRows?: (onProgress?) => Promise<T[]>、config.getValue: (row: T, key: string) => string；ExportFieldConfig 同步泛型化并保持导出
- [x] 选中过滤逻辑（selectedIds.includes(row.id)）语义不变（Asset/InstanceVO/CloudUser 均有 id: number）
- [x] 既有 15 页用法零改动通过 typecheck（T 推断为 Asset）
- [x] 9 单测语义不变通过（零改动，断言未动）
- [x] pnpm typecheck + vitest 全量通过（logs/cert 在途错误除外）；git diff 仅共享组件

## Notes
pnpm typecheck 仅剩 3 处并发在途错误（src/views/cert/changes/new/steps/Step1Select.vue + src/views/logs 两测试文件），经 git stash 对照验证与本改动无关（stash 后 HEAD 同样报错）；本次改动文件零类型错误。eslint 单文件通过。coverage 未由 runner 输出（vitest 未配置覆盖率门禁，重构任务维持既有覆盖不新增测试），故不填实际数值。并发会话脏文件（.env.development、element-theme.scss、components.d.ts、accounts/index.vue）未纳入提交。
