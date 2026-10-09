---
id: "3"
title: "中档存储/计算 5 页零漂移迁移共享组件"
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

# 3: 中档存储/计算 5 页零漂移迁移共享组件

## Description
**代码仓：D:/Haven/e-cam-web**。5 个中档真实导出页（`storage/{oss,nas}`、`compute/{disk,snapshot,security-group}`）
从本地 `ExportDialog.vue` 迁移到共享 `AssetExportDialog`，**字段/取值/文件名/默认字段逐项一致（零行为漂移）**。
删除 5 个本地 `ExportDialog.vue`。

## Reference Files
- `docs/proposals/export-dialog-consolidation/proposal.md` — Success Criteria（SC3 零漂移逐页对照）
- `D:/Haven/e-cam-web/src/views/storage/oss/index.vue` + `components/ExportDialog.vue`（135 行，getFieldValue+storage_size 格式化）
- `D:/Haven/e-cam-web/src/views/storage/nas/index.vue` + `components/ExportDialog.vue`（143 行，**带 `:metric-map="fsMetricMap"` 额外 prop**）
- `D:/Haven/e-cam-web/src/views/compute/disk/index.vue` + `components/ExportDialog.vue`（119 行）
- `D:/Haven/e-cam-web/src/views/compute/snapshot/index.vue` + `components/ExportDialog.vue`（124 行，index.vue 未传 selectedIds）
- `D:/Haven/e-cam-web/src/views/compute/security-group/index.vue` + `components/ExportDialog.vue`（122 行）
- `D:/Haven/e-cam-web/src/components/AssetExportDialog.vue` — 共享组件（任务 1 产物）

## Acceptance Criteria
- [ ] 5 页 `index.vue` 改用共享 `AssetExportDialog`：import 改 `@/components/AssetExportDialog.vue`，新增 `:config` prop
- [ ] 每页 config：`fields`=原 availableFields 逐字复制、`getValue`=原 getFieldValue 函数整体搬入闭包、`filename`=原下载文件名前缀、`defaultFields`=原 exportForm.fields 初始值——**逐项与本地 ExportDialog.vue 对照一致**
- [ ] nas 页 `metric-map` 扩展：getValue 闭包捕获 `fsMetricMap`，行为与原 `:metric-map` prop 等价（不新增共享组件 prop）
- [ ] 5 个本地 `ExportDialog.vue` 已删除；`fetchAllRows` 闭包/进度语义保留
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅这 5 页 index.vue + 删除 5 个 ExportDialog.vue

## Hard Rules
仅改：`src/views/storage/{oss,nas}/index.vue`、`src/views/compute/{disk,snapshot,security-group}/index.vue`；删除这 5 页的 `components/ExportDialog.vue`。
**不触碰** `compute/ecs`（大件留阶段 2）、其它页、`accounts/index.vue`（并发会话脏文件）、logs、后端、`components.d.ts`。
提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- **零漂移铁律**：getValue 逐字搬运原函数（含特殊格式化：oss storage_size TB/GB/MB、acl/versioning 映射等），不在迁移中"顺手优化"。
- 文件名前缀从原 `downloadFile(blob, \`XXX_${formatDate()}.csv\`)` 提取（如「OSS存储桶」）；共享组件已含 formatDate。
- snapshot 页未传 selectedIds → 不传即默认 `[]`，「已选中」自动禁用（共享组件已处理）。
- 若某页 `getFieldValue` 引用了页面级常量（label map 等），闭包需从 index.vue 作用域捕获或内联常量。
