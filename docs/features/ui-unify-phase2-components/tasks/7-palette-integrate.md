---
id: "7"
title: "CommandPalette 接入 MainLayout 并移除旧 header 搜索"
priority: "P1"
estimated_time: "1h"
complexity: "medium"
dependencies: [6]
surface-key: "e-cam-web"
surface-type: "web"
breaking: true
type: "coding.feature"
mainSession: false
---

# 7: CommandPalette 接入 MainLayout 并移除旧 header 搜索

## Description
将 Task 6 的 CommandPalette 挂载到 MainLayout 并接通 ⌘K 快捷键；移除 header 现有简易搜索（searchKeyword/searchResults/showSearchResults 相关 UI 与逻辑）。MainLayout 其余区域（侧栏/顶栏布局/用户菜单/租户选择器）零改动——侧栏浅色适配属 Phase 3。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Scope（替换 header 简易搜索）/ Out of Scope（MainLayout 重构属 P3）
- `src/layouts/MainLayout.vue` — 接入点（header 搜索区 + 快捷键监听）
- `src/components/CommandPalette/index.vue` — Task 6 产出

## Acceptance Criteria
- [ ] MainLayout 挂载 CommandPalette，⌘K（macOS）/Ctrl+K（Win）全局唤起生效，与既有输入框快捷键无冲突
- [ ] 旧 header 简易搜索 UI 与逻辑（searchKeyword/searchResults/showSearchResults/searchLoading 及其模板节点）移除无残留（grep 断言）
- [ ] 搜索资产直达详情的用户路径行为等价（回归验证：palette 搜索同一资产可达同一详情页）
- [ ] MainLayout 除搜索区与 palette 挂载点外零改动（侧栏/用户菜单/租户选择器 diff 为空）
- [ ] `pnpm build` + `pnpm typecheck` 通过；涉及文件的相关既有测试更新后全绿

## Hard Rules
- 仅修改 src/layouts/MainLayout.vue 及其直接关联测试文件
- 侧栏样式/布局样式一律不动（Phase 3 范围）

## Implementation Notes

### Test Impact
- Affected test suite(s): src/layouts/__tests__/（如存在 MainLayout 相关测试）及 MainLayout 内联逻辑的既有断言
- Expected fixture changes: 旧搜索相关的 mock（searchAssetsApi 在 MainLayout 层的 mock）移除，改由 CommandPalette 组件测试承载
- Risk level: low
