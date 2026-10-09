---
created: "2026-09-21"
author: "Haven"
status: Approved
intent: "new-feature"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: UI Unify Phase 3 — 标杆页面重写 + MainLayout 主题适配（浏览器实测验收）

## Problem

Phase 2 已交付 6 个标准模板组件与 /dev/ui-kit 样板页，但 225 个业务页面尚未有一个真正用新体系重写——组件体系缺乏真实页面背书，"迁移后长什么样"对后续 220+ 页仍是纸面答案；同时 MainLayout 侧栏在浅色主题下渲染异常（Phase 1 验收发现），主题体系存在最后一处系统性破绽。

### Evidence

- 四个标杆页现状：dashboard 628 行 / assets 617 行 / alert-events 239 行 / dictionary 330 行，均未使用任何 Phase 2 组件
- MainLayout.vue 1423 行，含 14 处硬编码色（Logo SVG `#3b82f6`、菜单激活态 `rgba(59,130,246,.15)` 等）；Phase 1 验收实测浅色模式下侧栏渲染与计算样式不符
- 用户明确要求：本轮验收必须用 chrome-devtools-mcp 在真实登录态下进行（Phase 1 的经验证明 diff 审查发现不了运行时问题，浏览器实测抓到过 2 个真 bug）

### Urgency

标杆页是 220+ 页迁移的"第一批活样板"——没有它们，demo 页的样板说服力不足，迁移规范（Phase 4）无法锚定真实页面形态。MainLayout 破绽则让双主题体系始终带着可见缺陷。

## Proposed Solution

用 Phase 2 组件体系**等价重写** 4 个标杆页面（功能不丢 + 规格 §6 指定的交互增强），MainLayout 做最小主题适配（清 14 处硬编码色 + 浅色正确渲染，不动结构），**每个页面重写任务内含 chrome-devtools-mcp 浏览器实测验收**：真实登录态���航、截图存档、关键元素计算样式断言、核心交互链路走通、双主题 + hover 态。

四个标杆页与规格 §6 一致：

| 页面 | 模板类型 | 重写要点 |
|------|---------|---------|
| `views/dashboard/index.vue` | 仪表盘 | PageContainer 骨架 + StatCard 排布 + 图表配色靛紫协调 |
| `views/assets/index.vue` | 列表页 | PageContainer + FilterBar + DataTable + 批量操作 + **行点击开 DetailDrawer** |
| `views/alert/events/index.vue` | 监控高频页 | PageContainer + DataTable + 语义状态色 + 密度切换 |
| `views/system/dictionary/index.vue` | 表单页 | 列表 + 弹窗表单统一布局与校验反馈 |

### Innovation Highlights

把"浏览器实测"从人工验收升级为**每页任务的内嵌验收标准**（可执行、可归档截图、可断言计算样式），延续 Phase 1 验收抓到 2 个真 bug 的方法论；等价重写边界严格——顺带优化被明确禁止，防 scope creep。

## Requirements Analysis

### Key Scenarios

- dashboard：打开即见统计卡与四图表区，数据来自嘉立创空间（5.0k 资产），图表配色与主题协调，切浅色不破
- assets：FilterBar 筛选 → DataTable 加载三态 → 行点击开 DetailDrawer 看详情 → 批量选择出现操作条 → 分页正常；密度切换生效
- alert/events：事件列表查询、状态徽章语义色正确、⌘K 可跳转到本页
- dictionary：类型列表 + 新增/编辑弹窗表单布局统一、校验反馈样式统一
- 双主题：每页深/浅切换 + hover 态浏览器实测
- 错误场景：接口失败时页面呈现 StateBlock 错误态 + 重试（由组件内置能力承载）

### Non-Functional Requirements

- 迁移后每页 scoped 样式显著缩减（目标 <100 行，仅存组件无法承载的布局微调）
- 页面源码零新增硬编码色字面量
- 全量测试无回归（≥ 既有 775 基线）
- 浏览器实测在真实登录态进行（dev server :5175，嘉立创空间有数据；登录态丢失时经 ECMDB :8888 以开发账号重登）

### Constraints & Dependencies

- 依赖：Phase 1 令牌层（main）、Phase 2 六组件（main）
- 浏览器验收工具：chrome-devtools MCP（会话级浏览器实例，登录态可复用）
- MainLayout 结构不动（1423 行拆分不在本期）

## Alternatives & Industry Benchmarking

### Industry Solutions

业界页面迁移惯例：设计系统落地时选 2-4 个代表性页面做 "lighthouse pages"（标杆页），验证组件体系并沉淀迁移模式，再批量推广（Ant Design Pro / Element Plus Pro 均此路径）。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零成本 | 组件体系无真实背书，220 页迁移无样板，浅色破绽延续 | Rejected: 阻塞 Phase 3/4 目标 |
| 全量迁移 225 页 | — | 一步到位 | 工量数周级、回���风险不可控 | Rejected: 违反渐进策略（规格 §7） |
| **4 标杆页 + MainLayout 适配（本提案）** | 规格 §6 | 样板可信度最大化、风险可控、浏览器实测前移 | 其余页面仍待迁移 | **Selected: 规格既定路径** |

## Feasibility Assessment

### Technical Feasibility

组件体系已在 demo 页验证（双主题浏览器实测通过）；四页现有功能均为组件能力覆盖范围内（筛选/表格/三态/抽屉/表单）���无技术盲区。

### Resource & Timeline

每页独立任务独立验收，MainLayout 适配独立任务；任一页受阻不阻塞其他页。

### Dependency Readiness

全部就绪（组件已合 main、dev server 运行中、浏览器会话已有登录态与数据空间）。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "重写应该顺带修各页交互问题" | Stress Test | Overturned：顺带优化使工量不可控且回归面扩大——等价重写 + 规格指定增强，其余记录待 Phase 4 |
| "MainLayout 需要借机拆分重构" | Occam's Razor | Overturned：1423 行拆分与标杆页目标无关，主题适配只需清色 + 令牌化，拆分另行立项 |
| "浏览器实测放最后统一做更省" | Assumption Flip | Overturned：Phase 1 经验证明运行时问题早暴露早修，每页即验避免四页返工叠加 |

## Scope

### In Scope

- `views/dashboard/index.vue` 等价重写（组件化 + 图表配色协调）
- `views/assets/index.vue` 等价重写（FilterBar/DataTable/批量操作/行点击 DetailDrawer）
- `views/alert/events/index.vue` 等价重写（DataTable/语义状态色/密度切换）
- `views/system/dictionary/index.vue` 等价重写（列表 + 弹窗表单统一）
- MainLayout 主题适配：14 处硬编码色清理（Logo 靛紫化、激活态令牌化），浅色侧栏正确渲染
- 每页任务内嵌 chrome-devtools-mcp 浏览器实测 AC：登录态导航、截图存档、计算样式断言、交互链路、双主题 + hover
- 四页完成后一次全站双主题浏览器巡检（回归确认）

### Out of Scope

- 其余 220+ 页面迁移（随日常迭代，样板已立）
- MainLayout 结构拆分（另行立项）
- UI-GUIDELINES / stylelint 防回潮（Phase 4）
- 任何新功能特性（等价重写边界）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 等价重写漏功能（旧页隐藏能力丢失） | M | M | 重写前盘点旧页全部用户路径并列入 AC；浏览器实测逐链路走通 |
| 浏览器实测依赖登录态/数据，会话失效或数据变动 | M | L | 实测步骤含登录态自检与重登路径（ECMDB :8888 开发账号）；数据断言用"存在性/样式"而非具体数值 |
| MainLayout 清色影响深色主题现观感 | L | M | 清色一律映射到等价令牌；双主题截图前后对照 |
| chrome-devtools MCP 会话不稳定（截图超时等） | M | L | 计算样式断言为主要证据、截图为辅；超时重试；必要时重 navigations |
| 并发会话干扰 forge 队列 | M | L | 每次领取前 `forge feature set` + 核对返回 FEATURE 字段（Phase 2 已验证的防御模式） |

## Success Criteria

- [ ] 4 个标杆页面全部基于 Phase 2 组件体系重写，每页 scoped 样式 <100 行（仅存布局微调），页面源码零新增硬编码色
- [ ] 功能等价：dashboard 统计与图表可达、assets 筛选/分页/详情可达、alert-events 查询可达、dictionary 增删改查可达（浏览器实测逐链路确认）
- [ ] 规格指定交互增强生效：assets 行点击开 DetailDrawer、DataTable 密度切换、⌘K 跳转落点到标杆页
- [ ] MainLayout 硬编码色清零（grep 断言 14→0），浅色主题侧栏浏览器实测正确
- [ ] 每页浏览器实测 AC 通过且截图归档：登录态导航、计算样式断言（令牌值命中）、核心交互链路、双主题 + hover
- [ ] 全量测试无回归（≥775 基线全绿），`pnpm build` / `pnpm typecheck` 零新增错误
- [ ] 最终全站双主题浏览器巡检通过（四页 + demo 页 + 主题切换）

## Next Steps

- Proceed to `/quick-tasks`（/quick 管线）
