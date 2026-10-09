---
status: "completed"
started: "2026-09-21 12:41"
completed: "2026-09-21 13:19"
time_spent: "~38m"
---

# Task Record: 6 CommandPalette ⌘K 全局命令面板组件

## Summary
CommandPalette ⌘K 全局命令面板组件落地：键盘导航状态机收敛于 useCommandPalette composable（⌘K/Ctrl+K 唤起、Esc 关闭、↑↓ 循环移动、Enter 执行、焦点开时捕获/关时归还）；路由菜单数据源直接派生自 @/router/routes 配置（flattenRouteEntries，hideInMenu/redirect/404 剔除）支持关键词过滤跳转；动作注册/执行 API（actions prop + registerActions expose + 内置「切换主题」走 appStore.toggleTheme）；最近访问经新 recentVisits pinia store（pinia-plugin-persistedstate 持久化，频次/时间排序，容量 8 条淘汰）；资产快捷搜索复用 searchAssetsApi 既有契约（300ms 防抖、{keyword, limit:10}、res.data.items/total、序号防竞态），结果直达 /assets/:id 数字 id 详情路由。组件零硬编码���色（19 处 var(--*) 令牌），TSDoc 覆盖全部对外 props/emits/expose/composable API，未含全局挂载逻辑（Task 7 接入）。

## Changes

### Files Created
- D:/Haven/e-cam-web/src/components/CommandPalette/index.vue
- D:/Haven/e-cam-web/src/components/CommandPalette/useCommandPalette.ts
- D:/Haven/e-cam-web/src/components/CommandPalette/sources.ts
- D:/Haven/e-cam-web/src/components/CommandPalette/types.ts
- D:/Haven/e-cam-web/src/stores/recentVisits.ts
- D:/Haven/e-cam-web/src/components/__tests__/CommandPalette.test.ts
- D:/Haven/e-cam-web/src/components/__tests__/useCommandPalette.test.ts
- D:/Haven/e-cam-web/src/stores/recentVisits.test.ts

### Files Modified
- D:/Haven/e-cam-web/src/stores/index.ts

### Key Decisions
- 路由条目从 @/router/routes 配置递归派生而非复制 MainLayout 菜单配置，保持单一事实来源（新模块路由自动进入面板）
- 键盘导航状态机独立为 useCommandPalette composable（不接触 DOM），键盘事件处理函数由组件绑定，满足提案『状态机收敛在 composable 内单测』的风险缓解
- 资产条目 filterable=false 本地直通（服务端已过滤），空 query 时排除资产组；资产详情跳转用数字 Asset.id 拼 /assets/:id（detail.vue 以 Number(route.params.id) 取参）
- 开关双模式：不传 visible 组件内自持状态（挂载即具备 ⌘K），传 visible 走受控 v-model:visible（emit 驱动），便于 Task 7 任意方式接入
- 遮罩背景用 var(--bg-base)+opacity 而非 rgba 字面量，面板底 var(--bg-overlay)、边框 var(--border-base)、阴影 var(--shadow-lg)，满足 Hard Rule 零颜色字面量
- 测试中发现 pinia.use() 在 app.use(pinia) 之前仅入队（toBeInstalled），纯 store 测试必须 createApp().use(pinia) 才能激活持久化插件

## Test Results
- **Tests Executed**: Yes
- **Passed**: 653
- **Failed**: 0
- **Coverage**: 97.0%

## Acceptance Criteria
- [x] ⌘K/Ctrl+K 唤起、Esc 关闭，完整键盘导航（↑↓ 移动、Enter 执行、鼠标可选），焦点管理正确（开时捕获、关时归还）
- [x] 路由菜单数据源（来自 router 配置的 23 模块）关键词过滤跳转 + 动作注册/执行 API（至少内置切换主题动作）
- [x] 最近访问：执行跳转后记录并按频次/时间排序展示，经 pinia-plugin-persistedstate 持久化刷新不丢
- [x] 资产快捷搜索：复用 searchAssetsApi（防抖），结果项可直达资产详情
- [x] TSDoc 覆盖 100% 对外 API；键盘导航状态机收敛在 composable 内单测，行覆盖率 ≥80%
- [x] 组件源码零硬编码颜色字面量（仅 var(--*)）；pnpm build + pnpm typecheck 通过

## Notes
coverage 97.02% lines（vitest --coverage v8，范围 src/components/CommandPalette/** + src/stores/recentVisits.ts，全部新模块行覆盖 96.7%+）。全量 vitest 653/653 通过（34 文件）。just e-cam-web-compile（vue-tsc -b）通过；vite build 通过（chunk 体积告警为存量 advisory）。just e-cam-web-lint 输出 3 错 16 警均为基线存量（src/api/types/{cost,statistics,task}.ts 空接口等），新增文件零 lint 告警；fmt 为 advisory（仓库无 .prettierrc 惯例，沿用现有风格）。未引入任何新运行时依赖。
