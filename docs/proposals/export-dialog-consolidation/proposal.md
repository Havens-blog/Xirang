---
created: "2026-09-20"
author: "Haven"
status: Approved
intent: "refactor"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: ExportDialog 家族收敛为共享配置驱动组件（分阶段）

## Problem

`e-cam-web/src/views/**/components/ExportDialog.vue` 存在 **17 个同构副本**（45–473 行），对话框壳、范围/格式/字段三段式 UI、CSV/XLSX 下载逻辑逐字重复；差异仅字段清单、取值映射、文件名前缀。且 **5 个最小副本（mongodb/rds/redis/kafka/es）是 no-op 占位**：`handleExport = () => ElMessage.success('导出功能开发中')`——用户选择范围/格式/字段后点「导出」，什么都不导出。

### Evidence

- 17 个 `ExportDialog.vue`（45–473 行）+ 2 个变体名（`WafExportDialog.vue`/`EniExportDialog.vue`），分布于 compute/databases/middleware/network/storage/cmdb/iam。
- 三档复杂度：大自定义件 ecs(473)/cmdb(467)/iam-users(313)/iam-audit(205)；中档真实导出 oss/nas/disk/snapshot/security-group/eip/eni/waf/cdn/vpc（`getFieldValue`+下载逻辑，`fetchAllRows` 全量拉取+进度）；**5 个 stub 页 mongodb(59)/rds(62)/redis(60)/kafka(45)/elasticsearch(45) 的 `handleExport` 均为 `ElMessage.success('导出功能开发中')` no-op**。
- 中档/大件已统一 props API：`visible/instances/selectedIds/total/fetchAllRows?`，全量拉取复用 `src/utils/exportAll.ts`（其注释引用审计 §2.2 导出家族 16 项）。
- `src/components/` 已有共享组件先例（AssetStatusBadge/CloudPlatformTag/EmptyState/ErrorDisplay）。

### Urgency

- **中**：非数据完整性缺陷，但 17 副本 DRY 成本高（改一处导出逻辑需同步 17 处）；5 个 stub no-op 让用户以为能导出却点了无反应（「开发中」toast 至少诚实，严重度低于 IAM/拓扑假数据）。抽取后修复与收敛同时完成。

## Proposed Solution

新建共享 `src/components/AssetExportDialog.vue`（**配置驱动**），分阶段迁移：
- 共享组件 props 与现中档完全一致（`visible/instances/selectedIds/total/fetchAllRows?`），新增 `config` prop：`{ fields: {key,label}[], getValue(row,key), filename, defaultFields }`。UI 用生产最成熟的全功能卡样式（范围按钮+格式卡+字段列表+全量拉取进度），导出语义逐字保留（范围解析/csv|xlsx/文件名/进度/失败降级）。
- **阶段 1（本次）**：5 个 stub 页（mongodb/rds/redis/kafka/es）迁移并**改为真实导出**（no-op → 可用）；中档 10 页（oss/nas/disk/snapshot/security-group/eip/eni/waf/cdn/vpc）迁移，字段/取值/文件名逐项一致（零漂移）。删除被迁移页的本地 `ExportDialog.vue`/`WafExportDialog.vue`/`EniExportDialog.vue`。
- **阶段 2（另排）**：ecs/cmdb/iam-users/iam-audit 大自定义件迁移（含 iam/groups 内联导出）。

### Innovation Highlights

配置驱动单组件收敛 15 副本 + 修复 5 处 no-op，一次到位。价值在消除 DRY 扩散（未来导出逻辑只改一处）与恢复「导出」按钮可信性。

## Requirements Analysis

### Key Scenarios

- 各页点「导出」：弹出共享对话框，范围（当前页/已选中/全部）、格式（xlsx/csv）、字段（勾选）均可用。
- 全部数据：`fetchAllRows` 闭包拉全量 + 行内进度（`fetchingAll` 按钮禁用）；未提供闭包退回当前页。
- 5 个 stub 页：迁移后点「导出」真正生成文件（原 `ElMessage.success('导出功能开发中')` 行为移除）。
- 中档 10 页：迁移后行为与迁移前**逐项一致**（字段顺序/取值映射/文件名/进度文案）。

### Non-Functional Requirements

- 全前端无 `导出功能开发中` 残留（grep 断言）。
- 共享组件 props/语义与现中档 ExportDialog 完全兼容；中档页零行为漂移。
- `src/utils/exportAll.ts` 不动（已共享）。

### Constraints & Dependencies

- 新增 `src/components/AssetExportDialog.vue`；修改迁移页 index.vue（import + config prop）；删除被迁移页的本地 ExportDialog。
- 不触碰 `accounts/index.vue`（并发会话脏文件）、logs、后端、`src/views/iam/`（阶段 2）、`src/views/compute/ecs`、`src/views/cmdb/instances`（阶段 2）。
- `components.d.ts` 为并发会话脏文件——若共享组件需登记，勿提交该文件（依赖自动导入）。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 共享配置驱动组件，分阶段迁移 | 本提案 | 一次收敛 15 副本+修 5 no-op；逐页零漂移 | 触碰 ~15 文件，需逐页验证 | Adopt（用户已确认） |
| 全量收敛 17 副本 | 备选 | 一步到位 | 大自定义件（ecs/cmdb/iam-*）差异化高，零漂移风险大 | Deferred：阶段 2 |
| 仅修 5 stub no-op | 备选 | 改动最小 | DRY 扩散仍在 | Rejected：用户已选抽取 |
| Do nothing | — | 零改动 | 17 副本维护成本+5 no-op 持续 | Rejected |

## Scope

### In Scope

- 新建共享 `src/components/AssetExportDialog.vue`（配置驱动 + 全功能卡 UI + 全部导出逻辑）。
- 5 个 stub 页迁移并真实导出：`databases/{mongodb,rds,redis}`、`middleware/{kafka,elasticsearch}`。
- 中档 10 页零漂移迁移：`storage/{oss,nas}`、`compute/{disk,snapshot,security-group}`、`network/{eip,eni,waf,cdn,vpc}`。
- 删除被迁移页的本地 `ExportDialog.vue`/`WafExportDialog.vue`/`EniExportDialog.vue`。
- 共享组件单测（范围解析/csv/xlsx/进度/失败降级）。

### Out of Scope

- 大自定义件迁移：`compute/ecs`、`cmdb/instances`、`iam/{users,audit,groups}`（阶段 2 另排）。
- 后端导出接口、`accounts/index.vue`、logs、拓扑/证书等其它页面。
- 导出字段清单本身的内容调整（仅原样迁移，不改字段语义）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 中档页 getFieldValue 取值映射有每资源特例（如 ecs memory/security_groups），迁移时漏掉 | Medium | 迁移即把原函数整体搬入 config.getValue 闭包（非重写）；对照原文件 diff 逐页验证 |
| 文件名/格式/默认字段差异被抹平 | Medium | config 含 filename/defaultFields；迁移时从原文件逐字复制 |
| 共享组件 UI 替换 5 个 stub 的简版 el-form 样式（480px→680px 卡样式） | Low | 属可接受升级；5 stub 页本无真实导出，无旧行为可比 |
| 并发会话 `components.d.ts` 脏文件与自动导入冲突 | Low | 不提交该文件；共享组件显式 import 而非依赖全局注册 |
| 迁移页 index.vue 的 import 路径改动遗漏 | Low | 每页 typecheck + 页面级 vitest（若有）兜底 |

## Success Criteria

- [ ] 新建 `src/components/AssetExportDialog.vue`（配置驱动：fields/getValue/filename/defaultFields），props 与现中档一致（visible/instances/selectedIds/total/fetchAllRows），导出语义逐字保留
- [ ] 5 个 stub 页改用共享组件并真实导出；全前端无 `导出功能开发中` 残留（grep 断言）
- [ ] 中档 10 页迁移共享组件，字段/取值/文件名与原 ExportDialog 逐项一致（逐页对照 diff 断言）
- [ ] 已迁移页的本地 `ExportDialog.vue`/`WafExportDialog.vue`/`EniExportDialog.vue` 删除；ecs/cmdb/iam-users/iam-audit 大自定义件与 `components.d.ts` 未动
- [ ] 共享组件单测覆盖：范围解析、csv/xlsx 生成、fetchAllRows 进度、失败降级
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅迁移作用域文件（共享组件 + 迁移页 + 测试）
