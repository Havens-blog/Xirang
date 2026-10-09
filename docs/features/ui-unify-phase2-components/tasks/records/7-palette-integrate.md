---
status: "completed"
started: "2026-09-21 13:21"
completed: "2026-09-21 13:42"
time_spent: "~21m"
---

# Task Record: 7 CommandPalette 接入 MainLayout 并移除旧 header 搜索

## Summary
CommandPalette 接入 MainLayout 并移除旧 header 简易搜索：MainLayout 挂载 <CommandPalette />（非受控模式，⌘K/Ctrl+K 唤起、Esc 关闭由组件自管）；完整移除旧搜索的脚本逻辑（searchKeyword/searchResults/showSearchResults/searchLoading/searchTotal/assetTypeMap/handleSearch/防抖 watch/highlightText/handleResultClick/closeSearchResults/handleKeydown/handleClickOutside）、document 键盘监听注册、.search-box 模板节点及 .search-box/.search-results/.search-highlight/.search-dropdown/@keyframes rotate 全部样式；新增 src/layouts/__tests__/MainLayout.test.ts 五个集成用例（Ctrl+K 唤起、Meta+K 唤起、toggle 奇偶性防双监听冲突、源码 grep 无残留断言、palette 资产搜索直达 /assets/:id）。侧栏/用户菜单/租户选择器零改动（git diff 全部 hunk 均限于搜索区与 palette 挂载点）。

## Changes

### Files Created
- e-cam-web/src/layouts/__tests__/MainLayout.test.ts

### Files Modified
- e-cam-web/src/layouts/MainLayout.vue

### Key Decisions
- palette 以裸挂载（非受控模式）接入：CommandPalette 不传 visible 时自持开关状态并自带 document 级 ⌘K 监听，MainLayout 不再注册任何键盘监听，避免双监听 toggle 冲突（测试以开/关奇偶性断言锁定）
- 资产搜索路径语义按提案执行：旧搜索跳列表页带 ?search= 参数，提案（权威规格）定义为 palette 资产搜索直达 /assets/:id 详情，AC3 的验证口径即 palette 搜索同一资产可达同一详情页，回归用例锁定 /assets/42
- 旧 handleKeydown 的 Escape 分支随之移除：palette 关闭态下 Esc 无全局行为系任务 6 组件设计（Esc 仅在面板输入框内消费）
- git diff 中 platform-nav-trigger 仅有的 3 行尾空白差异已还原为与 HEAD 一致，保证 AC4 '除搜索区与挂载点外零改动' 严格成立

## Test Results
- **Tests Executed**: Yes
- **Passed**: 718
- **Failed**: 0
- **Coverage**: 96.7%

## Acceptance Criteria
- [x] MainLayout 挂载 CommandPalette，⌘K（macOS）/Ctrl+K（Win）全局唤起生效，与既有输入框快捷键无冲突
- [x] 旧 header 简易搜索 UI 与逻辑（searchKeyword/searchResults/showSearchResults/searchLoading 及其模板节点）移除无残留（grep 断言）
- [x] 搜索资产直达详情的用户路径行为等价（回归验证：palette 搜索同一资产可达同一详情页）
- [x] MainLayout 除搜索区与 palette 挂载点外零改动（侧栏/用户菜单/租户选择器 diff 为空）
- [x] pnpm build + pnpm typecheck 通过；涉及文件的相关既有测试更新后全绿

## Notes
coverage 字段为本任务特性模块（src/components/CommandPalette/**）行覆盖 96.72%（语句 93.02%/分支 88.13%）；MainLayout.vue 壳层行覆盖 60.22%（1700 行既有布局骨架，任务前无任何测试，侧栏/flyout/面包屑属 Phase 3 重构范围，未在本任务补测）。全量 38 文件 718 测试全绿（含新增 5 个）。just e-cam-web-compile（vue-tsc -b）通过；pnpm build 退出码 0。lint：触及文件 eslint 0 告警，全仓 3 错 16 警与既定基线完全一致（api/types 空接口 3 错）；prettier：MainLayout.vue 在 HEAD 即不达标（仓库无 .prettierrc，fmt 为 advisory），维持仓库既有风格未做 prettier 重排。
