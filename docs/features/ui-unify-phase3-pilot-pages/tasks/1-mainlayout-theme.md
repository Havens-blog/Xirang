---
id: "1"
title: "MainLayout 主题适配：清 14 处硬编码色 + 浅色侧栏修正"
priority: "P0"
estimated_time: "1.5h"
complexity: "high"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 1: MainLayout 主题适配：清 14 处硬编码色 + 浅色侧栏修正

## Description
Phase 1 验收发现 MainLayout 侧栏在浅色主题下渲染异常，且全文件含 14 处硬编码色（Logo SVG `#3b82f6`、菜单激活态 `rgba(59,130,246,.15)` 等）。本任务做最小主题适配：全部硬编码色映射到 Phase 1 等价令牌（Logo 靛紫化、激活态/悬浮态改 var(--*)），使浅色侧栏正确渲染——**布局结构与业务逻辑零改动**（1423 行拆分另行立项）。

## Reference Files
- `docs/proposals/ui-unify-phase3-pilot-pages/proposal.md` — Proposed Solution / Key Scenarios / Key Risks
- `src/layouts/MainLayout.vue` — 唯一修改目标（14 处硬编码色分布）

## Acceptance Criteria
- [ ] 硬编码色清零：`grep -cE "#[0-9a-fA-F]{6}|rgba\(59, ?130, ?246" src/layouts/MainLayout.vue` 返回 0（Logo SVG 改 #5e6ad2/#7170ff 族属映射后保留，计入白名单并在报告中列明）
- [ ] 菜单激活态/悬浮态/滚动条等改用 var(--*) 令牌，深浅主题语义正确
- [ ] chrome-devtools-mcp 浏览器实测：浅色主题下侧栏/菜单渲染正确（截图归档 + 侧栏元素计算样式断言），深色主题视觉零回归（改前后截图对照）
- [ ] MainLayout 结构与业务逻辑零改动（diff 仅限颜色/样式相关行）
- [ ] `pnpm build` + `pnpm typecheck` 通过，涉及测试全绿

## Hard Rules
- 仅修改 src/layouts/MainLayout.vue（及其直接关联测试）
- 不做结构拆分、不动业务逻辑

## Implementation Notes
- 浏览器环境：dev server :5175（若不在跑先 `pnpm dev`）；登录态经 ECMDB :8888（开发账号 admin/12345678），工作空间切嘉立创（jlc）有数据
- 浅色切换可直接 classList 操作 html.light 验证（与 appStore.applyTheme 等效），不必依赖 UI 按钮
- 计算样式断言为主证据、截图为辅；MCP 截图偶发超时属已知现象，重试即可
