---
id: "3"
title: "cmdb/instances 页迁移共享组件（InstanceVO 行类型）"
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

# 3: cmdb/instances 页迁移共享组件（InstanceVO 行类型）

## Description
**代码仓：D:/Haven/e-cam-web**。`cmdb/instances` 的本地 ExportDialog（467 行，`InstanceVO[]` 行类型，
getFieldValue 6 个特例分支）迁到泛型化共享组件，**零行为漂移**。删除本地组件。这是泛型 T≠Asset 的首个消费者。

## Reference Files
- `docs/proposals/export-dialog-phase2/proposal.md` — Success Criteria（SC2/SC3）
- `D:/Haven/e-cam-web/src/views/cmdb/instances/index.vue` + `components/ExportDialog.vue`（467 行：InstanceVO、availableFields :139、getFieldValue 6 特例、fetchAllRows）
- `D:/Haven/e-cam-web/src/api/types/cmdb.ts:205` — InstanceVO（id: number）
- `D:/Haven/e-cam-web/src/components/AssetExportDialog.vue` — 共享组件（任务 1 泛型化后）

## Acceptance Criteria
- [ ] `cmdb/instances/index.vue` 改用共享组件：config 的 fields/getValue/filename/defaultFields 与原组件逐项一致（6 特例分支逐字搬运）
- [ ] T=InstanceVO 推断正常（泛型首个非 Asset 消费者，验证任务 1 泛型化实效）
- [ ] `fetchAllRows`（返回 `Promise<InstanceVO[]>`）衔接不变；进度/失败降级语义不变
- [ ] 本地 `components/ExportDialog.vue` 已删除
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅 cmdb/instances/index.vue + 删 1 组件

## Hard Rules
仅改 `src/views/cmdb/instances/index.vue`；删 `src/views/cmdb/instances/components/ExportDialog.vue`。
**不碰** ecs/iam 页（2/4 负责）、`accounts/index.vue`（并发脏）、logs、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- **零漂移铁律**：getFieldValue 逐字搬运；InstanceVO 字段访问（uid/asset_id/attributes 等）原样保留。
- 若原组件 UI/语义与共享组件有出入，按任务 2 的同一诚实评估原则处理。
