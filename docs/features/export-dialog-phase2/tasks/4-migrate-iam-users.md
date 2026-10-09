---
id: "4"
title: "iam/users 页迁移共享组件（CloudUser + prop 映射）"
priority: "P1"
estimated_time: "2h"
complexity: "high"
dependencies:
  - "1"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 4: iam/users 页迁移共享组件（CloudUser + prop 映射）

## Description
**代码仓：D:/Haven/e-cam-web**。`iam/users` 的本地 ExportDialog（313 行，`CloudUser[]` 行类型）迁到泛型化
共享组件。与另两页不同，它接口形状有差异需**页侧映射**：props 名（`users`/`selectedUsers`/`totalCount`）、
字段描述 `{label,value}`（共享组件 `{key,label}`）、多一个 `success` emit。删除本地组件。

## Reference Files
- `docs/proposals/export-dialog-phase2/proposal.md` — Success Criteria（SC2）
- `D:/Haven/e-cam-web/src/views/iam/users/index.vue` + `components/ExportDialog.vue`（313 行：Props :74、availableFields :90（10 字段 `{label,value}`）、success emit）
- `D:/Haven/e-cam-web/src/api/types/iam.ts:14` — CloudUser（id: number）
- `D:/Haven/e-cam-web/src/components/AssetExportDialog.vue` — 共享组件（任务 1 泛型化后）

## Acceptance Criteria
- [ ] `iam/users/index.vue` 改用共享组件：`:instances="users"`、`:selected-ids` 由 `selectedUsers` 映射 id、`:total="totalCount"`；config 字段描述 `{label,value}` → `{key,label}` 翻译（10 字段逐字）、getValue 闭包逐字搬运、filename/defaultFields 原样
- [ ] `success` emit 消费链保留：确认页面是否消费该事件；消费则由共享组件导出成功路径触发（共享组件导出成功后 emit `update:visible` false——如需 success 需在页侧以导出完成状态衔接或经配置扩展，诚实评估最小方案并记录）
- [ ] 本地 `components/ExportDialog.vue` 已删除；`fetchAllRows?: Promise<CloudUser[]>` 衔接不变
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅 iam/users/index.vue + 删 1 组件

## Hard Rules
仅改 `src/views/iam/users/index.vue`；删 `src/views/iam/users/components/ExportDialog.vue`。
**不碰** `iam/audit`（排除项，零改动）、ecs/cmdb 页（2/3 负责）、`accounts/index.vue`（并发脏）、logs、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- **零漂移铁律**：10 字段（username/display_name/user_type/status/provider/cloud_account_name/email/permission_groups/create_time/update_time）取值逻辑逐字搬运（permission_groups 数组格式化等特例保留）。
- `success` emit：先 grep `@success` 在 index.vue 的消费；无消费则直接丢弃并记录；有消费则最小方案衔接（如监听导出完成），不为它扭曲共享组件 API。
- 原 `dialogVisible` computed（visible↔update:visible）由共享组件 `v-model:visible` 等价承接。
