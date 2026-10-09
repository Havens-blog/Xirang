---
id: "1"
title: "新建共享 AssetExportDialog 配置驱动组件 + 单测"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: 新建共享 AssetExportDialog 配置驱动组件 + 单测

## Description
**代码仓：D:/Haven/e-cam-web**。新建 `src/components/AssetExportDialog.vue`：把现 ExportDialog 家族最成熟的
全功能卡样式（范围按钮 + 格式卡 + 字段列表 + 全量拉取进度）收敛为一个配置驱动共享组件，导出语义逐字保留
（范围解析/csv|xlsx/文件名/进度/失败降级）。**本任务只建共享组件 + 单测，不改任何现有 ExportDialog**（迁移在 2/3/4）。

## Reference Files
- `docs/proposals/export-dialog-consolidation/proposal.md` — Proposed Solution, In Scope, Risks, Success Criteria
- `D:/Haven/e-cam-web/src/views/compute/ecs/components/ExportDialog.vue` — 全功能卡 UI + 导出逻辑模板（resolveExportRows/scope/csv|xlsx/downloadFile/formatDate）
- `D:/Haven/e-cam-web/src/views/storage/oss/components/ExportDialog.vue` — 中档精简版参考（props API 同形）
- `D:/Haven/e-cam-web/src/api/types/asset.ts` — Asset 类型
- `D:/Haven/e-cam-web/src/utils/constants.ts` — getProviderLabel
- `D:/Haven/e-cam-web/src/utils/exportAll.ts` — 全量拉取工具（不动，仅了解）

## Acceptance Criteria
- [ ] 新建 `src/components/AssetExportDialog.vue`，props 与现中档一致：`visible/instances: Asset[]/selectedIds: number[]/total: number/fetchAllRows?: (onProgress?)=>(Promise<Asset[]>)` + 新增 `config` prop（`ExportFieldConfig`：`fields: {key,label}[]`、`getValue(row,key):string`、`filename:string`、`defaultFields:string[]`）；emit `update:visible`
- [ ] UI 含：范围按钮（全部/当前页/已选中，已选中无数据禁用）、格式卡（xlsx/csv）、字段勾选（全选/取消全选）、「全部数据」分页拉取行内进度（`fetchingAll` 禁用按钮）
- [ ] 导出语义逐字保留：范围解析（current=instances/selected=按 selectedIds 过滤/all=fetchAllRows 闭包或退 current）、csv|xlsx（BOM、引号转义、制表符/逗号分隔）、文件名 `{config.filename}_{YYYYMMDD}.{ext}`、失败 ElMessage.error 且返回 null 不关闭、成功关闭弹窗
- [ ] 单测 `src/components/__tests__/AssetExportDialog.test.ts` 覆盖：范围解析三态、csv/xlsx 内容生成（BOM+转义）、fetchAllRows 进度回调、fetchAllRows 失败降级、无 fetchAllRows 时 all 退 current
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅新增共享组件 + 测试文件

## Hard Rules
仅新增：`src/components/AssetExportDialog.vue` + `src/components/__tests__/AssetExportDialog.test.ts`。
**不改**任何现有 `ExportDialog.vue`/`index.vue`（2/3/4 迁移任务负责）、`accounts/index.vue`（并发会话脏文件）、
logs、后端、`components.d.ts`（并发会话脏文件，勿提交）。提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- 组件内 `export interface ExportFieldConfig` 并导出，供迁移页 import 类型。
- `getValue` 为纯函数入参（row, key），由各页提供闭包（nas 的 metric-map 等页面扩展数据经闭包捕获，不做特殊 prop）。
- 复制 ecs ExportDialog 的 resolveExportRows/handleExport/csv|xlsx 逻辑（含 scope 顺序：current→selected→all）；fetchingAll/fetchedCount 进度态保留。
- 样式：沿用 ecs 内联 scss 全功能卡样式（680px）；不依赖 `@/views/storage/styles/export-dialog.scss`（共享组件自带样式，避免跨目录 import）。
- 测试参考 `src/views/cert/ledger/index.test.ts` 的 vitest 模式；组件用 mount + mock fetchAllRows。
