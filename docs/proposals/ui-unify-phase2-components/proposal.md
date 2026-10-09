---
created: "2026-09-21"
author: "Haven"
status: Approved
intent: "new-feature"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: UI Unify Phase 2 — 统一页面模板组件体系（ui-kit）

## Problem

e-cam-web 的 225 个视图页面没有可复用的标准页面模板体系，各页面自行拼装布局/筛选/表格/状态展示，导致风格不统一、交互不一致、迁移无法规模化。

### Evidence

- PageContainer（v1 仅是空壳 div）只有 29/225 页面使用（13%）；PageHeader 仅 1 页使用；StatCard/ChartCard 仅 8 页
- 216 个页面手写 scoped style；全仓 1237 处内联样式/硬编码颜色
- 各页手写筛选栏（如 `asset-filters`）、表格工具栏、加载/空/错三态，结构各异
- Phase 1（已合并 main，08904e9）已交付 Linear 规格令牌层，但缺少**消费令牌的标准组件**——令牌无法自动变成统一页面

### Urgency

Phase 3（4 个标杆页面重写）与后续 220+ 页迁移全部依赖本组件体系；没有它，每次页面改动都在继续产生新的风格漂移，Phase 4 的 stylelint 防回潮也无从锚定。

## Proposed Solution

交付 6 个 TypeScript 模板组件（分两批），全部只消费 Phase 1 的 CSS 变量令牌（零硬编码色），配套 TSDoc、单测（覆盖率 ≥80%）和一个 `/dev/ui-kit` 内部 demo 路由页（集中展示典型用法与状态组合，作为 220+ 页迁移的"照抄样板"和可视化验收靶场）。

**第一批（布局层，互相配套）**：
1. `StateBlock` — 统一三态块：skeleton 加载 / 空状态（合并 EmptyState）/ 错误态含重试（合并 ErrorDisplay）
2. `FilterBar` — 配置化筛选栏：声明字段数组自动渲染 input/select/date 等，字段过多响应式折叠
3. `DataTable` — el-table 封装：columns 声明式配置、密度切换、冻结表头、多选批量操作条、内置三态
4. `PageContainer` 2.0 — 页头（标题+操作区）/筛选区/内容区/底部分页区标准骨架，默认 slot 向后兼容 v1

**第二批（交互层，独立性强）**：

5. `DetailDrawer` 2.0 — 统一详情���屉：描述列表（label-value 网格）、分区块、内置加载态，v1 API 向后兼容
6. `CommandPalette` — ⌘K 全局命令面板：路由菜单跳转（23 模块）+ 常用动作 + 最近访问 + 资产快捷搜索（复用 searchAssetsApi），**替换** MainLayout header 现有简易搜索

### Innovation Highlights

属于标准的"内部 UI Kit"做法，创新点克制而务实：(1) demo 页即迁移样板——每个典型用法可直接复制进业务页；(2) 三态内置于 DataTable/StateBlock，业务页从此不再手写 v-if 加载分支；(3) 双强调色语义（--accent-primary=CTA / --accent-blue=交互）在组件层强制贯彻，杜绝混用。

## Requirements Analysis

### Key Scenarios

- 列表页骨架：`<PageContainer title>` + `<FilterBar :fields>` + `<DataTable :columns :fetch>` + 批量操作条 + 分页，一行不写 scoped style
- 三态覆盖：DataTable 加载中自动 skeleton、空数据显示空态、请求失败显示错误态+重试按钮
- 详情抽屉：列表行点击 → DetailDrawer 打开，描述列表分区块展示，加载态内置
- ⌘K：任意页面按键唤起 → 输入"主机"跳转 / 输入资产 ID 直达详情 / 执行"切换主题" / 最近访问回跳
- 兼容：现有 29 个 PageContainer 使用方与 DetailDrawer 调用方**零改动**通过 build
- 错误场景：DataTable fetch 抛错 → StateBlock 错误态 + 重试；CommandPalette 无匹配结果 → 空态提示

### Non-Functional Requirements

- 主题：组件只允许使用 `var(--*)` 令牌，深浅主题均需视觉正确（demo 页双主题验收）
- 无障碍：CommandPalette 完整键盘导航（���↓ 选择、Enter 执行、Esc 关闭）、焦点管理
- 性能：CommandPalette 最近访问本地持久化（复用 pinia-plugin-persistedstate 既有机制）
- 兼容：Vue 3.5 + Element Plus 2.10 + TS strict；不引入任何新运行时依赖

### Constraints & Dependencies

- 依赖：Phase 1 令牌层（已合并）、现有 searchAssetsApi、pinia persistedstate
- 约束：docs/ 在 .gitignore（仓库惯例，文档本地保存）；MainLayout 重构属 Phase 3，本期只替换 header 搜索接入点
- 测试栈沿用项目现有 vitest + @vue/test-utils + happy-dom

## Alternatives & Industry Benchmarking

### Industry Solutions

业界同类：Element Plus Pro 的 ProTable/ProForm 模式、Ant Design Pro ProComponents、Naive UI Admin 的页面骨架封装。共同点是"配置驱动 + 声明式列/字段 + 内置三态"。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零成本 | Phase 3/迁移无锚点，漂移继续扩大 | Rejected: 与整体重构目标冲突 |
| 引入第三方 ProComponents | vue-element-plus-admin 等 | 现成功能多 | 重度定制难、引入新依赖（规格禁止）、风格难对齐 Linear 令牌 | Rejected: 依赖与定制性冲突 |
| 自建 6 组件（本提案） | 规格 §5 + 业界 Pro 模式 | 完全贴合令牌与业务语义、无新依赖、API 可按 29 个现有使用方反推 | 开发工作量自担 | **Selected: 规格已定，唯一合规路径** |

## Feasibility Assessment

### Technical Feasibility

全部基于既有栈能力：el-table/el-drawer 封装、组合式 API、vitest 单测均为项目已验证模式（现有 28 个测试文件）。无技术盲区。

### Resource & Timeline

两批交付，每批独立可验收；API 以现有页面真实用法为锚（29 个 PageContainer 用户 + assets/alert 现状），避免过度设计。

### Dependency Readiness

Phase 1 已合并；searchAssetsApi 已在线上使用（现 header 搜索即用它）；无外部依赖。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "PageContainer 2.0 需要迁移 29 个使用方" | 代码核查 | Overturned：v1 是空壳 div，2.0 保留默认 slot 即天然兼容，零迁移成本 |
| "CommandPalette 需要自建搜索" | 代码核查 | Refined：资产搜索直接复用 searchAssetsApi 与现有防抖逻辑，只做面板壳 |
| "需要更多组件（如 StatCard 2.0）" | Occam's Razor | Overturned：规格 §5 只定 6 个，StatCard 已有且 8 页在用，YAGNI 缓期 |

## Scope

### In Scope

- 6 个组件（两批交付）：StateBlock、FilterBar、DataTable、PageContainer 2.0、DetailDrawer 2.0、CommandPalette
- 全部组件 TypeScript + TSDoc（覆盖对外 props/slots/events/expose）
- 每组件单测，整体覆盖率 ≥80%（vitest --coverage 口径）
- `/dev/ui-kit` 内部 demo 路由页：6 组件典型用法 + 加载/空/错三态组合 + 深浅双主题展示
- CommandPalette 替换 MainLayout header 简易搜索（含旧搜索入口移除）
- 最近访问持久化（pinia persistedstate）

### Out of Scope

- 任何业务页面的迁移改造（Phase 3：4 个标杆页 → 其余页面随迭代）
- MainLayout 整体重构/侧栏浅色适配（Phase 3；本期仅动 header 搜索接入点）
- UI-GUIDELINES 规范文档、stylelint 防回潮规则、迁移清单（Phase 4）
- 引入新运行时依赖（虚拟滚动等如需，用原生实现）
- StatCard/ChartCard 2.0（已有组件缓期评估）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 组件 API 与页面真实需求错位（闭门造车） | M | H | API 以 29 个 PageContainer 使用方 + assets/alert 页现状反推；demo 页先行验证；Phase 3 标杆页迁移时回收反馈修订 |
| DataTable 封装功能面失控（scope creep） | M | M | 严格只做规格承诺的 5 项能力；行内编辑等明确缓期；新需求走 Phase 4 评估 |
| CommandPalette 替换 header 搜索引发回归 | L | M | 替换前用现有 searchAssetsApi 的调用契约写单测；header 接入点改动集中在一处，可快速回退 |
| CommandPalette 键盘导航 80% 覆盖成本高 | M | L | @vue/test-utils 键盘事件模拟成熟；导航状态机收敛在 composable 内单测 |

## Success Criteria

- [ ] 6 个组件全部落地 `src/components/`，TypeScript 严格模式通过，TSDoc 覆盖 100% 对外 API（props/slots/events/expose）
- [ ] 组件单测整体行覆盖率 ≥80%（vitest --coverage），全部测试通过、输出无噪音
- [ ] 现有 29 个 PageContainer 使用方与全部 DetailDrawer v1 调用方零改动通过 `pnpm build` + `pnpm typecheck`
- [ ] `/dev/ui-kit` 路由可达：6 组件典型用法 + 三态组合齐全，深/浅主题下浏览器实测视觉正确（含 hover 态）
- [ ] CommandPalette：⌘K 唤起，可搜索跳转 23 模块路由、执行常用动作（切主题等）、展示最近访问、资产关键词搜索直达；header 旧搜索入口无残留
- [ ] 组件源码硬编码颜色字面量为 0（`grep -rE "#[0-9a-fA-F]{6}" src/components/{六组件}` 仅允许命中 TSDoc 示例文本）
- [ ] `pnpm build` / `pnpm typecheck` 通过，无新增 typecheck/lint 错误（相对 main 基线）

## Next Steps

- Proceed to `/quick-tasks`（本提案走 /quick 管线，跳过 write-prd）
