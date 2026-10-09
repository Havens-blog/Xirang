---
id: "3"
title: "DataTable 表格封装（声明式列/内置三态/批量操作）"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: [1]
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 3: DataTable 表格封装（声明式列/内置三态/批量操作）

## Description
Phase 2 布局层核心组件。基于 el-table 封装：columns 声明式配置、密度切换、冻结表头、多选批量操作条，并内置复用 Task 1 的 StateBlock 实现加载/空/错三态——业务页不再手写 v-if 加载分支。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Proposed Solution / Key Risks（功能面失控风险）
- `src/components/StateBlock/index.vue` — Task 1 产出，三态复用
- `src/views/assets/index.vue` — 现有表格用法（列结构/分页锚点）

## Acceptance Criteria
- [ ] columns 声明式配置渲染（prop/label/width/formatter/自定义单元格 slot），数据经 props 传入或 fetch 函数注入
- [ ] 内置三态：loading 自动 skeleton、空数据空态、fetch 失败错误态 + retry（复用 StateBlock）
- [ ] 多选列可选开启，选中项聚合为底部批量操作条（actions 配置 + selection-change 事件）
- [ ] 密度切换（默认/紧凑）+ 冻结表头（sticky header）
- [ ] TSDoc 覆盖 100% 对外 props/slots/events/expose
- [ ] 单测覆盖列渲染/三态切换/选择事件/密度切换，行覆盖率 ≥80%；零硬编码颜色字面量；`pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 严格只做本任务承诺的能力，行内编辑/列拖拽排序等明确缓期（防 scope creep）
- 不得引入新运行时依赖

## Implementation Notes
- 分页不内置在 DataTable 里（由 PageContainer 底部区承载 el-pagination），保持单一职责——与规格 §5 一致
- 密度切换通过 el-table size prop 透传实现，不重造样式
