---
created: "2026-09-20"
author: "Haven"
status: Approved
intent: "refactor"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: ExportDialog 收敛阶段 2 —— 泛型化共享组件并迁移大自定义件

## Problem

export-dialog-consolidation（阶段 1）已把 15 个同构副本收敛到共享 `AssetExportDialog`，但留了 4 个大自定义件（45–473 行）未迁：ecs(473)/cmdb-instances(467)/iam-users(313)/iam-audit(205)。其中 3 个（ecs/cmdb/iam-users）与共享组件**同族**（行列表 + 范围/格式/字段选择 + 客户端 CSV/XLSX 生成），仅行类型不同；导出逻辑仍三处重复，改一处导出行为需同步三处。共享组件行类型锁死 `Asset[]`，无法承接其余行类型。

### Evidence

- `compute/ecs/components/ExportDialog.vue`(473 行)：`instances: Asset[]`，22 字段，getFieldValue 7 个特例分支（status 映射/charge_type/memory 格式化/security_groups/tags）——与共享组件同族，纯配置差异。
- `cmdb/instances/components/ExportDialog.vue`(467 行)：`instances: InstanceVO[]`（`api/types/cmdb.ts:205`，有 `id: number`），getFieldValue 6 个特例分支——同族。
- `iam/users/components/ExportDialog.vue`(313 行)：`CloudUser[]`（`api/types/iam.ts:14`，有 `id: number`），字段描述用 `{label,value}`（共享组件为 `{key,label}`）、props 名不同（`users`/`selectedUsers`/`totalCount`）、多一个 `success` emit——同族但接口形状有差异。
- `iam/audit/components/ExportDialog.vue`(205 行)：props 只有 `filters`，走 `exportAuditLogsApi(filters)` 服务端 blob 下载（`:82,171-177`）——**服务端过滤导出，非客户端行列表导出，不同族**。
- 共享组件 `src/components/AssetExportDialog.vue` props 锁死 `instances: Asset[]`/`fetchAllRows?: ...Promise<Asset[]>`；项目 Vue 3.5 支持 SFC `generic` 属性。
- 阶段 1 迁移的 15 页 + 9 单测全部依赖现 props API（`visible/instances/selectedIds/total/fetchAllRows?/config`）。

### Urgency

- **中**：非缺陷，但 3 处重复导出逻辑的维护成本持续；阶段 1 已建好共享组件与迁移模式，边际成本低。完成除 audit 外全量收敛。

## Proposed Solution

泛型化共享组件行类型，迁移 3 个同族大件，排除不同族的 iam/audit：
- `AssetExportDialog.vue` 行类型从 `Asset` 泛化为组件级泛型 `T`（`instances: T[]`、`fetchAllRows?: ...Promise<T[]>`、`config.getValue: (row: T, key: string) => string`）；既有 15 页用法不变（T 推断为 Asset）。
- 三页零漂移迁移：每页 config 闭包逐字复制原 `availableFields`/`getFieldValue`/文件名/默认字段；iam/users 页侧做 prop 映射（`:instances="users"`、`:selected-ids="selectedUsers.map(u => u.id)"`、字段描述 `{label,value}`→`{key,label}` 翻译、`success` emit 由既有 `update:visible`+导出成功路径保留）。
- 删除 3 个本地 ExportDialog.vue（约 1250 行 → 约 150 行配置）。
- `iam/audit` **排除**：与服务端导出族语义不同，保持独立（记录于 Out of Scope，不为套壳而扭曲）。

### Innovation Highlights

组件级泛型让单一共享组件承接多行类型，阶段 1 的配置驱动设计自然延伸；无创新实现，价值在消除最后 3 处同族重复并明确家族边界（audit 另立）。

## Requirements Analysis

### Key Scenarios

- ecs/cmdb/iam-users 三页导出：范围（当前页/已选中/全部+进度）、格式（xlsx/csv）、字段勾选行为与迁移前逐项一致。
- 既有 15 页（阶段 1 迁移）：泛型化后行为与类型完全不变（T=Asset）。
- iam/users 的 `success` emit：迁移后导出成功仍触发页面侧刷新逻辑（若有消费）。

### Non-Functional Requirements

- 零行为漂移：三页字段/取值/文件名/默认字段与原组件逐项一致。
- 泛型化不改变既有 15 页的类型推断与运行时行为；共享组件 9 单测不改语义通过。
- 全前端 grep：`views/compute/ecs`、`views/cmdb/instances`、`views/iam/users` 下不再有本地 ExportDialog.vue。

### Constraints & Dependencies

- 改 `src/components/AssetExportDialog.vue`（泛型化）+ 其测试的类型适配；改 3 页 index.vue；删 3 个本地组件。
- 不触碰 `accounts/index.vue`（并发会话脏文件，+349 行在改）、logs 文件（并发在途）、`components.d.ts`。
- `iam/audit` 组件不动。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 泛型化 + 迁 3 大件，audit 排除 | 本提案 | 除 audit 外全量收敛；多行类型单组件 | 泛型化触碰共享组件与既有测试 | Adopt（用户已确认） |
| 只迁同形态 2 个（ecs/cmdb），iam/users 三期 | 备选 | iam/users 接口形状差异规避 | 收敛不彻底，iam/users 313 行重复仍在 | Rejected：差异可控（页侧映射即可） |
| 四个都套共享组件 | 备选 | 数量上全收敛 | audit 是服务端 filter 导出，强行套壳扭曲两者语义 | Rejected |
| 维持现状 | — | 零改动 | 3 处重复导出逻辑持续 | Rejected |

## Scope

### In Scope

- `src/components/AssetExportDialog.vue`：行类型泛型化（`generic="T"`），props/config 类型适配；既有 9 单测语义不变通过（必要时仅类型层适配）。
- `compute/ecs/index.vue` 迁移 + 删本地 `components/ExportDialog.vue`。
- `cmdb/instances/index.vue` 迁移 + 删本地 `components/ExportDialog.vue`。
- `iam/users/index.vue` 迁移（含 prop 映射与字段描述翻译）+ 删本地 `components/ExportDialog.vue`；`success` emit 消费链保留。
- 三页零漂移：字段/取值/文件名/默认字段逐字搬运。

### Out of Scope

- `iam/audit` 导出组件（服务端 filter 导出，不同族，保持独立）。
- `accounts/index.vue`、logs、后端、拓扑/证书等其它页面。
- 导出字段内容调整（仅原样迁移）。
- 已迁移 15 页的任何行为改动。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 泛型化破坏既有 15 页类型推断或运行时 | Medium | `generic="T"` 对无标注用法推断为具体行类型；全量 typecheck + 512 vitest 兜底；必要时 T 默认 `= any` |
| iam/users prop 形状差异映射遗漏（`success` emit、字段描述翻转） | Medium | 迁移前先盘点该页对 `success` 的消费；config 翻译为纯数据映射并在提交说明列出 |
| 三大件 getFieldValue 特例（7+6+IAM 状态映射）搬运走样 | Medium | 逐字复制原函数入 config 闭包；逐页对照原文件 diff 断言 |
| ecs/cmdb 组件有共享组件未覆盖的 UI 细节（如额外列说明/提示） | Low | 迁移前 diff 原 UI 与共享组件能力清单；确有差异的诚实评估是否值得为它加配置项，不值得则记录保留差异 |

## Success Criteria

- [ ] `AssetExportDialog.vue` 行类型泛型化（`instances: T[]`/`fetchAllRows?: ...Promise<T[]>`/`getValue: (row: T, key) => string`）；既有 15 页用法零改动，9 单测语义不变通过
- [ ] ecs/cmdb/iam-users 三页迁到共享组件，字段/取值/文件名/默认字段与原组件逐项一致（逐页 diff 断言）；iam/users 的 `success` 链路保留
- [ ] 3 个本地 ExportDialog.vue 删除；全前端除 `iam/audit` 外无独立行列表导出组件（grep 断言）
- [ ] `iam/audit` 组件零改动
- [ ] `pnpm typecheck` + vitest 全量通过（logs 在途错误除外，属并发会话）；`git diff` 仅共享组件 + 3 页 + 删 3 组件
