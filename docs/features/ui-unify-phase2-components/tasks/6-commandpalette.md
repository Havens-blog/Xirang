---
id: "6"
title: "CommandPalette ⌘K 全局命令面板组件"
priority: "P1"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 6: CommandPalette ⌘K 全局命令面板组件

## Description
Phase 2 交互层核心组件。⌘K/Ctrl+K 唤起的全局命令面板：路由菜单跳转（23 模块）、常用动作注册与执行（切主题等）、最近访问记录（pinia-plugin-persistedstate 持久化）���资产快捷搜索（复用现有 searchAssetsApi 及其防抖契约）。完整键盘导航。组件本体不含全局挂载逻辑（接入由 Task 7 承载）。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Proposed Solution / Non-Functional Requirements / Key Risks
- `src/layouts/MainLayout.vue` — 现有 header 简易搜索（searchKeyword/searchAssetsApi 调用契约，本组件复用其 API 不复用其实现）
- `src/api/asset.ts` — searchAssetsApi 契约

## Acceptance Criteria
- [ ] ⌘K/Ctrl+K 唤起、Esc 关闭，完整键盘导航（↑↓ 移动、Enter 执行、鼠标可选），焦点管理正确（开时捕获、关时归还）
- [ ] 路由菜单数据源（来自 router 配置的 23 模块）关键词过滤跳转 + 动作注册/执行 API（至少内置"切换主题"动作）
- [ ] 最近访问：执行跳转后记录并按频次/时间排序展示，经 pinia-plugin-persistedstate 持久化刷新不丢
- [ ] 资产快捷搜索：复用 searchAssetsApi（防抖），结果项可直达资产详情
- [ ] TSDoc 覆盖 100% 对外 API；键盘导航状态机收敛在 composable 内单测，行覆盖率 ≥80%
- [ ] 组件源码零硬编码颜色字面量（仅 var(--*)）；`pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 不得引入新运行时依赖
- 零硬编码颜色字面量（只用 var(--*) 令牌）

## Implementation Notes
- 风险缓解（提案 Key Risks）：searchAssetsApi 调用契约先写单测再实现面板
- 面板视觉对齐 Linear 风格：居中浮层、--bg-overlay 底、--border-base 边框、结果分组（页面/动作/资产）
