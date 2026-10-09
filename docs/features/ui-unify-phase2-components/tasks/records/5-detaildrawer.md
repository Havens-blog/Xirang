---
status: "completed"
started: "2026-09-21 14:37"
completed: "2026-09-21 14:58"
time_spent: "~21m"
---

# Task Record: 5 DetailDrawer 2.0 统一详情抽屉（向后兼容 v1）

## Summary
DetailDrawer 2.0 统一详情抽屉：v1 全部对外 API 逐项保留（10 props/3 events/4 slots），新增 items/sections 描述列表（label-value 网格，支持 formatter/span/columns，sections 分区块带可选标题）与基于 StateBlock 的内置骨架加载态；TSDoc 覆盖 100% 对外 API（4 导出类型 + 13 props + 3 emits + 4 slots）；21 个新单测全绿（RED 10 失败 → GREEN 21 通过），组件行覆盖率 95.83%、分支 97.56%；现有调用方（src/views/assets/index.vue）零改动，vue-tsc -b 与 vite build 双绿。

## Changes

### Files Created
- src/components/__tests__/DetailDrawer.test.ts

### Files Modified
- src/components/DetailDrawer/index.vue

### Key Decisions
- loading prop 语义保留（'显示加载态'）但实现由 v1 的 v-loading 遮罩升级为 StateBlock 骨架屏——proposal 明确要求'内置加载态，复用 Task 1 StateBlock'；唯一真实调用方（assets/index.vue）未传 loading，零回归
- 内容区渲染优先级：loading > tabs（v1 分支先行，tabs+items 同传时 tabs 优先且 items 忽略）> 描述列表 + 默认插槽
- 扁平 items 归一为单个无标题区块（descriptionGroups computed）与 sections 共用同一渲染路径，模板零重复
- value 渲染规则：formatter 优先 → null/undefined/空串渲染 '-' 占位 → 其余 String(value)（0 正常渲染 '0'）
- columns ≥1 校验（非法值回退默认 2 列，对齐存量详情页 info-grid 两列布局），列数经 --detail-columns CSS 自定义属性下发
- 样式仅消费 Phase 1 Linear 令牌（--bg-surface/--bg-elevated/--bg-base/--border-subtle/--text-primary/--text-secondary/--text-tertiary，深浅主题均已定义）；组件与测试保持仓库风格（与 StateBlock/DataTable/PageContainer 一致）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 21
- **Failed**: 0
- **Coverage**: 95.8%

## Acceptance Criteria
- [x] v1 对外 API（props/slots/events）逐项保留，现有调用方零改动通过 pnpm build + pnpm typecheck
- [x] 新增描述列表模式：items 配置（label/value/formatter/span）渲染 label-value 网格，支持 sections 分区块
- [x] 内置加载态（复用 StateBlock loading）
- [x] TSDoc 覆盖 100% 对外 API；单测覆盖兼容契约 + 新增能力，行覆盖率 ≥80%
- [x] 组件源码零硬编码颜色字面量（仅 var(--*)）
- [x] 深浅主题下视觉正确（demo 页联验前先自测）

## Notes
验证链：just e-cam-web-compile（vue-tsc -b）exit 0 → just e-cam-web-fmt advisory → just e-cam-web-lint exit 0（基线 3 错 17 警不变，本任务测试文件 1 个 unused-var 警告已自修复清零）→ 目标测试 21/21 绿 → npx vite build exit 0（chunk 体积提示为全仓既有）。Hard Rule 核验：git status 确认未修改任何 DetailDrawer 调用方文件，仅 components.d.ts 为其他任务遗留的自动生成脏区（未纳入本任务）。fmt 警告（本任务两文件非 prettier 达标）为仓库既定 advisory 惯例——全部兄弟 Phase 2 组件（StateBlock/DataTable/PageContainer）同样保持仓库风格不重排。AC6 自测口径：组件样式零颜色字面量、所消费令牌在 theme-variables.scss 深色块（L20-46）与浅色块（L98-124）均有定义；/dev/ui-kit demo 页联验属 feature 级后续步骤。src/views/logs/index.test.ts 中的 LogDetailDrawer 为视图本地组件，与本组件无关。未覆盖行 48/214 为模板渲染函数边缘行，行覆盖 95.83% 已超 80% 目标。
