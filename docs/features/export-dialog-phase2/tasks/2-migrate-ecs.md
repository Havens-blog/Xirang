---
id: "2"
title: "ecs 页迁移共享组件（473 行零漂移收敛）"
priority: "P1"
estimated_time: "1.5h"
complexity: "high"
dependencies:
  - "1"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 2: ecs 页迁移共享组件（473 行零漂移收敛）

## Description
**代码仓：D:/Haven/e-cam-web**。`compute/ecs` 是阶段 1 共享组件的 UI 原型来源（473 行，22 字段，
getFieldValue 7 个特例分支），迁到泛型化后的共享 `AssetExportDialog`，**零行为漂移**。删除本地 ExportDialog.vue。

## Reference Files
- `docs/proposals/export-dialog-phase2/proposal.md` — Success Criteria（SC2 零漂移）
- `D:/Haven/e-cam-web/src/views/compute/ecs/index.vue` + `components/ExportDialog.vue`（473 行：22 availableFields、getFieldValue 特例 status/charge_type/memory/security_groups/tags、fetchAllRows）
- `D:/Haven/e-cam-web/src/components/AssetExportDialog.vue` — 共享组件（任务 1 泛型化后）

## Acceptance Criteria
- [ ] `ecs/index.vue` 改用共享组件：新增 `:config` prop，fields=原 22 项逐字、getValue=原 getFieldValue 整体闭包搬运（status/charge_type/memory GB/MB/security_groups/tags 特例全保留）、filename=「虚拟机列表」、defaultFields=原初始 10 项
- [ ] `fetchAllRows` 闭包衔接不变；进度/失败降级语义不变
- [ ] 本地 `components/ExportDialog.vue` 已删除；全前端无该文件引用
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅 ecs/index.vue + 删 1 组件

## Hard Rules
仅改 `src/views/compute/ecs/index.vue`；删 `src/views/compute/ecs/components/ExportDialog.vue`。
**不碰** cmdb/iam 页（3/4 负责）、`accounts/index.vue`（并发脏）、logs、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- **零漂移铁律**：getFieldValue 逐字搬运，不"顺手优化"；原组件导出语义（范围解析顺序 current→selected→all、BOM/转义、文件名格式）由共享组件承载——先 diff 确认原组件与共享组件语义一致（阶段 1 即以 ecs 为模板，应完全一致）。
- 若原组件有共享组件没有的 UI 能力，诚实评估：值得加配置项则加（记录），不值得则记录差异保留现状。
