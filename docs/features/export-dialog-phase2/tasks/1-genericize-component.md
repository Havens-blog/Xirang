---
id: "1"
title: "AssetExportDialog 行类型泛型化（generic=T）"
priority: "P0"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: AssetExportDialog 行类型泛型化（generic=T）

## Description
**代码仓：D:/Haven/e-cam-web**。共享组件 `src/components/AssetExportDialog.vue` 行类型锁死 `Asset`，
无法承接 InstanceVO/CloudUser 行类型。将行类型泛型化为组件级泛型 `T`（Vue 3.5 SFC `generic` 属性），
**既有 15 页用法与 9 单测语义零改动**。本任务只动共享组件与测试类型层，不迁任何页面（2/3/4 迁移）。

## Reference Files
- `docs/proposals/export-dialog-phase2/proposal.md` — Proposed Solution, In Scope, Risks, Success Criteria
- `D:/Haven/e-cam-web/src/components/AssetExportDialog.vue` — 现 props：`instances: Asset[]/selectedIds/total/fetchAllRows?: ...Promise<Asset[]>` + `config`（getValue: `(row: Asset, key) => string`）
- `D:/Haven/e-cam-web/src/components/__tests__/AssetExportDialog.test.ts` — 9 单测（语义不变通过）
- 阶段 1 提案 `docs/proposals/export-dialog-consolidation/proposal.md` — 既有 API 契约

## Acceptance Criteria
- [ ] 组件声明 `generic="T"`，props 变为 `instances: T[]`、`fetchAllRows?: (onProgress?) => Promise<T[]>`、`config.getValue: (row: T, key: string) => string`；`ExportFieldConfig` 同步泛型化并保持导出
- [ ] 选中过滤逻辑（`selectedIds.includes(row.id)`）语义不变（Asset/InstanceVO/CloudUser 均有 `id: number`）
- [ ] 既有 15 页用法**零改动**通过 typecheck（T 推断为 Asset）；如推断不达，T 兜底默认 `= any` 并在注释说明
- [ ] 9 单测语义不变通过（允许仅类型层适配，断言不改）
- [ ] `pnpm typecheck` + vitest 全量通过（logs 在途错误除外）；`git diff` 仅共享组件 + 测试文件

## Hard Rules
仅改：`src/components/AssetExportDialog.vue` + `src/components/__tests__/AssetExportDialog.test.ts`。
**不改**任何页面、不迁任何 ExportDialog（2/3/4 负责）、`accounts/index.vue`（并发脏）、logs、`components.d.ts`（并发脏勿提交）。
提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- Vue 3.5 支持 `<script setup lang="ts" generic="T">`；与现有导出 `ExportFieldConfig` 的独立 `<script lang="ts">` 块共存无冲突。
- 泛型化后 `ExportFieldConfig` 变为 `ExportFieldConfig<T>`；既有 15 页传 config 时 T 自动推断，无需显式标注。
- 测试 mount 处如需显式类型参数，仅做类型适配不改断言。
- 组件名不改（AssetExportDialog 保持，避免 15 页 import 变动；语义上它已不只服务 Asset，在文件头注释说明）。
