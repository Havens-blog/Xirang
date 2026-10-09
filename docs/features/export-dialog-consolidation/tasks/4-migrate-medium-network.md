---
id: "4"
title: "中档网络 5 页零漂移迁移共享组件"
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

# 4: 中档网络 5 页零漂移迁移共享组件

## Description
**代码仓：D:/Haven/e-cam-web**。5 个中档真实导出页（`network/{eip,eni,waf,cdn,vpc}`）
从本地 `ExportDialog.vue`/`EniExportDialog.vue`/`WafExportDialog.vue` 迁移到共享 `AssetExportDialog`，
**字段/取值/文件名/默认字段逐项一致（零行为漂移）**。删除 5 个本地导出组件（含变体名）。

## Reference Files
- `docs/proposals/export-dialog-consolidation/proposal.md` — Success Criteria（SC3/SC4）
- `D:/Haven/e-cam-web/src/views/network/eip/index.vue` + `components/ExportDialog.vue`（138 行）
- `D:/Haven/e-cam-web/src/views/network/eni/index.vue` + `components/EniExportDialog.vue`（**变体名**）
- `D:/Haven/e-cam-web/src/views/network/waf/index.vue` + `components/WafExportDialog.vue`（**变体名**）
- `D:/Haven/e-cam-web/src/views/network/cdn/index.vue` + `components/ExportDialog.vue`（145 行）
- `D:/Haven/e-cam-web/src/views/network/vpc/index.vue` + `components/ExportDialog.vue`（183 行）
- `D:/Haven/e-cam-web/src/components/AssetExportDialog.vue` — 共享组件（任务 1 产物）

## Acceptance Criteria
- [ ] 5 页 `index.vue` 改用共享 `AssetExportDialog`：import 改 `@/components/AssetExportDialog.vue`（eni/waf 的原组件名引用同步改），新增 `:config` prop
- [ ] 每页 config：`fields`=原 availableFields 逐字复制、`getValue`=原 getFieldValue 函数整体搬入闭包、`filename`=原下载文件名前缀、`defaultFields`=原 exportForm.fields 初始值——**逐项与本地组件对照一致**
- [ ] 5 个本地 `ExportDialog.vue`/`EniExportDialog.vue`/`WafExportDialog.vue` 已删除；`fetchAllRows` 闭包/进度语义保留
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅这 5 页 index.vue + 删除 5 个本地导出组件

## Hard Rules
仅改：`src/views/network/{eip,eni,waf,cdn,vpc}/index.vue`；删除对应本地 `components/ExportDialog.vue`/`EniExportDialog.vue`/`WafExportDialog.vue`。
**不触碰** 其它任何页、`accounts/index.vue`（并发会话脏文件）、logs、后端、`components.d.ts`。
提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- **零漂移铁律**：getValue 逐字搬运原函数（含特殊格式化与 label 映射），不在迁移中"顺手优化"。
- eni/waf 页原组件命名 `EniExportDialog`/`WafExportDialog`，index.vue 的 import 与模板标签都要改为共享组件名。
- 文件名前缀从原 downloadFile 提取（如「弹性IP」/「VPC」/「CDN」等按各页实际）。
- 若某页 getFieldValue 引用页面级常量/label map，闭包从 index.vue 作用域捕获或内联常量。
