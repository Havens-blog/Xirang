---
status: "completed"
started: "2026-09-21 14:05"
completed: "2026-09-21 14:35"
time_spent: "~30m"
---

# Task Record: 3 DataTable 表格封装（声明式列/内置三态/批量操作）

## Summary
DataTable 表格封装落地 src/components/DataTable/：columns 声明式列配置（prop/label/width/minWidth/align/fixed/sortable/showOverflowTooltip/formatter/自定义单元格 slot，锚定 assets 页列结构可无损表达）、内置三态（loading 骨架屏/空态/fetch 失败错误态 + retry，复用 StateBlock，业务页不再手写 v-if 加载分支）、可选多选列 + 底部批量操作条（batchActions 配置 + selection-change 透传 + batch-action(key, rows)）、密度切换（default/compact 经 el-table size 透传）与冻结表头（max-height 透传）。数据注入双模式：data props 或 fetch 函数（挂载即加载、函数标识变化重载、fetch 优先）。泛型 SFC（generic="T = any"），$attrs 原样透传 el-table（stripe/row-key 零包装成本），分页不内置（由 PageContainer 底部区承载）。21 个新单测全绿，新模块行覆盖率 100%；全量 739 测通过；vue-tsc -b 与 vite build 通过；零硬编码颜色字面量、零新运行时依赖；未做行内编辑/列拖拽（严守 Hard Rules scope）

## Changes

### Files Created
- e-cam-web/src/components/DataTable/index.vue
- e-cam-web/src/components/DataTable/types.ts
- e-cam-web/src/components/__tests__/DataTable.test.ts

### Files Modified
- e-cam-web/src/components.d.ts

### Key Decisions
- fetch 注入签名定为 () => Promise<T[]>：分页由 PageContainer 底部区承载 el-pagination（规格 §5 单一职责），组件只管行数组与三态
- 密度切换经 el-table size prop 透传（default→default、compact→small），冻结表头经 max-height 透传，不重造样式
- $attrs 原样透传 el-table（inheritAttrs: false）：stripe/row-key 等 el-table 既有能力零包装成本
- 泛型 SFC generic="T = any"（沿用 AssetExportDialog 先例）；VTU mount 对泛型组件 props 类型退化为 unknown，测试侧以类型擦除挂载并在注释说明
- exposed clearSelection 在 tableRef.clearSelection() 之外同步清空本地 selectedRows：不依赖 EP clearSelection 是否回发 selection-change，批量条隐藏行为确定
- el-table 在 happy-dom 下触发 MutationObserver realm 私有字段报错，测试统一以 no-op 桩替换（仅用于 DOM 变更后强制重渲染，不影响断言）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 739
- **Failed**: 0
- **Coverage**: 100.0%

## Acceptance Criteria
- [x] columns 声明式配置渲染（prop/label/width/formatter/自定义单元格 slot），数据经 props 传入或 fetch 函数注入
- [x] 内置三态：loading 自动 skeleton、空数据空态、fetch 失败错误态 + retry（复用 StateBlock）
- [x] 多选列可选开启，选中项聚合为底部批量操作条（actions 配置 + selection-change 事件）
- [x] 密度切换（默认/紧凑）+ 冻结表头（sticky header）
- [x] TSDoc 覆盖 100% 对外 props/slots/events/expose
- [x] 单测覆盖列渲染/三态切换/选择事件/密度切换，行覆盖率 ≥80%；零硬编码颜色字面量；pnpm build + pnpm typecheck 通过

## Notes
覆盖率口径 vitest --coverage.include='src/components/DataTable/**'：行 100%、语句 100%、函数 100%、分支 86.84%（均 ≥80% 行覆盖门槛）。全量测试 39 文件 739 通过（基线 718 + 新增 21）。just e-cam-web-compile/fmt/lint：compile 绿；lint 基线 3 错 16 警不变（均不在本任务文件，targeted lint 新文件 0 问题）；fmt 为 advisory（仓库无 .prettierrc，justfile || true）。components.d.ts 新增 DataTable 自动生成条目（连同已提交组件 CommandPalette 的滞后条目一并入库）。el-table happy-dom 兼容（MutationObserver 桩）与 VTU 字符串 scoped slot 需 <template #name="params"> 包裹两点已记录在测试文件头注释。
