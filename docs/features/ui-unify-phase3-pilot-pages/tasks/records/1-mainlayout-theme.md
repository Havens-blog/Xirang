---
status: "completed"
started: "2026-09-21 16:10"
completed: "2026-09-21 16:53"
time_spent: "~43m"
---

# Task Record: 1 MainLayout 主题适配：清 14 处硬编码色 + 浅色侧栏修正

## Summary
MainLayout 主题适配：14 处硬编码色全部映射到 Phase 1 等价令牌（Logo 靛紫化 #3b82f6→#5e6ad2/#7170ff 白名单保留；8 处 rgba(59,130,246) → color-mix(in srgb, var(--accent-primary) N%, transparent)；3 处内联 #333333 → var(--text-regular)），浅色侧栏浏览器实测渲染正确、深色零回归，结构与业务逻辑零改动（diff 仅 14 处颜色/样式行）。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/layouts/MainLayout.vue
- e-cam-web/src/layouts/__tests__/MainLayout.test.ts

### Key Decisions
- 激活态/悬浮态/平台导航入口 8 处 rgba(59,130,246,x) 统一映射为 color-mix(in srgb, var(--accent-primary) 等比%, transparent)——采用 cert 视图既有写法，颜色随令牌联动且深浅主题语义一致（accent 两主题同值）
- Logo SVG 按 AC 白名单保留字面量：实面 #5e6ad2（--accent-primary 品牌锚点）+ 双描边 #7170ff（--accent-blue）
- 主题/全屏/通知按钮内联色浅色分支 #333333 → var(--text-regular)（浅色主题下 --text-regular 恰为 #333333，两主题渲染逐像素不变）；保留 isDarkTheme 三元结构不动业务逻辑
- flyout 门户阴影 rgba(0,0,0,0.25) 不在 14 处清零范围且改 --shadow-lg 会加深深色阴影（违背深色零回归），有意保留并在 notes 说明
- 新增 3 个源码级测试（复用 ?raw 导入）：清零+白名单断言 / Logo 靛紫化 / color-mix 令牌化计数

## Test Results
- **Tests Executed**: Yes
- **Passed**: 778
- **Failed**: 0
- **Coverage**: 60.4%

## Acceptance Criteria
- [x] 硬编码色清零：grep -cE "#[0-9a-fA-F]{6}|rgba\(59, ?130, ?246" 返回 0（Logo 白名单除外）
- [x] 菜单激活态/悬浮态/滚动条改用 var(--*) 令牌，深浅主题语义正确
- [x] chrome-devtools-mcp 浏览器实测：浅色侧栏/菜单渲染正确（截图归档+计算样式断言），深色零回归（改前后对照）
- [x] MainLayout 结构与业务逻辑零改动（diff 仅限颜色/样式相关行）
- [x] pnpm build + pnpm typecheck 通过，涉及测试全绿

## Notes
AC1 实测 grep 返回 3，全部为白名单 Logo SVG 行：480 fill="#5e6ad2"、481-482 stroke="#7170ff"；旧蓝 rgba 与 #333333 均为 0。AC3 证据归档于 docs/features/ui-unify-phase3-pilot-pages/screenshots/task-1-mainlayout/（before/after 双主题截图 4 张 + 计算样式快照 JSON 2 份）：浅色 sidebarBg #ffffff、activeBg #5e6ad2@15%、经真实 UI 切换（appStore.toggleTheme）主题按钮 #333333；深色与改前基线逐项一致（仅激活/入口底色按预期蓝转靛）。浏览器状态已还原（/cam/assets + dark）。AC5：pnpm typecheck/build exit 0，MainLayout 测试 8/8，全量 778/778（基线 775+新增 3）。静态检查：compile exit 0；lint 3 错 16 警为既有基线（本任务两文件 eslint 直查 0 告）；prettier 两文件与 HEAD 同状态（本就不达标，维持仓库风格，fmt 为 advisory）。coverage 60.42% 为 MainLayout.vue 模块行覆盖，未覆盖行为本任务禁止触碰的既有业务逻辑（菜单配置/面包屑/flyout/登出），本任务验证以源码断言+浏览器计算样式为主（AC3 指定方法）。
