---
id: "2"
title: "5 个 stub 页迁移共享组件并真实导出"
priority: "P0"
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

# 2: 5 个 stub 页迁移共享组件并真实导出

## Description
**代码仓：D:/Haven/e-cam-web**。5 个 no-op stub 页（`databases/{mongodb,rds,redis}`、`middleware/{kafka,elasticsearch}`）
的 `ExportDialog` 目前 `handleExport = () => ElMessage.success('导出功能开发中')`（点了不导出）。迁移到任务 1 的
共享 `AssetExportDialog`，用各页现有 `availableFields` 作为真实导出字段 → 恢复「导出」可信性。删除 5 个本地
`ExportDialog.vue`。

## Reference Files
- `docs/proposals/export-dialog-consolidation/proposal.md` — Success Criteria（SC2 无「导出功能开发中」残留）
- `D:/Haven/e-cam-web/src/views/databases/mongodb/index.vue` + `components/ExportDialog.vue`（59 行 stub，availableFields 9 项）
- `D:/Haven/e-cam-web/src/views/databases/rds/index.vue` + `components/ExportDialog.vue`（62 行 stub）
- `D:/Haven/e-cam-web/src/views/databases/redis/index.vue` + `components/ExportDialog.vue`（60 行 stub）
- `D:/Haven/e-cam-web/src/views/middleware/kafka/index.vue` + `components/ExportDialog.vue`（45 行 stub）
- `D:/Haven/e-cam-web/src/views/middleware/elasticsearch/index.vue` + `components/ExportDialog.vue`（45 行 stub）
- `D:/Haven/e-cam-web/src/components/AssetExportDialog.vue` — 共享组件（任务 1 产物）

## Acceptance Criteria
- [ ] 5 页 `index.vue` 改用共享 `AssetExportDialog`：import 改 `@/components/AssetExportDialog.vue`，新增 `:config` prop（fields=原 availableFields、getValue=通用取值、filename 各页中文名、defaultFields=原 exportFields）
- [ ] 通用 `getValue`：asset_id/asset_name 直取；provider → `getProviderLabel`；其余 `attributes[key] || ''`
- [ ] 5 个本地 `ExportDialog.vue` 已删除；`exportDialogVisible` 等页面状态与 `fetchAllRows`（若有）衔接正常
- [ ] 全前端 `导出功能开发中` 无残留（grep 断言）
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅这 5 页 index.vue + 删除 5 个 ExportDialog.vue

## Hard Rules
仅改：`src/views/databases/{mongodb,rds,redis}/index.vue`、`src/views/middleware/{kafka,elasticsearch}/index.vue`；删除这 5 页的 `components/ExportDialog.vue`。
**不触碰** 其它任何页（中档迁移在 3/4）、`accounts/index.vue`（并发会话脏文件）、logs、后端、`components.d.ts`。
提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- 各页默认字段沿用原 `exportFields` 值（如 `['asset_name','asset_id','status','version','region']`）；无 fetchAllRows 的页省略该 prop（all 范围退 current）。
- filename 建议：mongodb「MongoDB实例」、rds「RDS实例」、redis「Redis实例」、kafka「Kafka实例」、elasticsearch「Elasticsearch实例」。
- 迁移前先读各页 stub 确认 availableFields 与页面数据绑定（instances/selectedIds/total 已由 index.vue 传入，仅换组件）。
- 单测若各页无既有测试，共享组件测试（任务 1）已覆盖导出逻辑；本任务以 typecheck + 手动页面校验为主。
