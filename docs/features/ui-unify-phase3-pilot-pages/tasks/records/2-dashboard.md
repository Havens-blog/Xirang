---
status: "completed"
started: "2026-09-21 16:57"
completed: "2026-09-21 17:57"
time_spent: "~1h"
---

# Task Record: 2 dashboard 仪表盘等价重写（组件化 + 浏览器实测）

## Summary
dashboard 仪表盘等价重写（PageContainer + StatCard + StateBlock 组件化）：旧页 13 项用户路径全保留（四统计卡含 k-w 缩写与四条点击路径、四图表区 option 逐字段等价、过期表 7 列/tag 分级/过期着色/7-30-90 天筛选重拉、5 处失败 toast、resize+dispose）。scoped 样式 106→65 行（<100 AC），页面色字面量 12 个全部 ∈ 旧页 22 项基线（零新增，较旧页移除 6 个 → accent/text 令牌化）。浏览器实测（chrome-devtools MCP，嘉立创空间真实数据 5.0k 资产/5 厂商）：panel-card 计算样式双主题命中 --bg-surface/--border-subtle，StatCard 命中 glass 令牌族，过期着色==--accent-red；发现并修复继承缺陷——图表文字层（环心数字/轴标签/外置标签）原为暗色硬编码致浅色主题不可读，改为初始化解析令牌 + 主题切换 watcher 重渲四图（系列色保持 fc0b2e4 迁移色不动），实测浅色环心「4992」清晰、实时切换重渲正确；StatCard hover 三态实测（bg/border/transform）；7 天内切窗实测 8 行全 danger tag。8 张截图归档 docs/features/ui-unify-phase3-pilot-pages/screenshots/2-dashboard/（docs 不进 git）。新增 22 测（含 ?raw 源码守卫：样式 <100 行 + 色字面量基线比对 + 零 #3b82f6），全量 800/800 绿（基线 778+22），vue-tsc -b 零错误，pnpm build EXIT 0，dashboard 目录 lint 零问题（全仓基线 3 错 16 警不变）。

## Changes

### Files Created
- e-cam-web/src/views/dashboard/index.test.ts

### Files Modified
- e-cam-web/src/views/dashboard/index.vue

### Key Decisions
- 统计卡用既有 StatCard 组件（proposal 指定 StatCard 排布）；其 iconMap 无 Warning 图标，即将过期卡取语义等价 Clock，不改共享组件源码（等价重写边界）
- StatCard formatNumber（千分位）与旧页 k-w 缩写不同 → 资产总数卡传预格式化字符串、其余卡 :format-value=false，保持逐字等价显示
- 图表三态由 StateBlock 承载（proposal Key Scenarios：错误态 + 重试由组件内置能力承载），provider 图旧页无错误提示补齐为同款三态（规格内统一）；v-loading 语义升级为骨架屏（DetailDrawer 2.0 同前例），4 张图表 loading 初始 true 消除首帧空态闪烁
- 图表容器（panel-card）用 --bg-surface/--border-subtle（AC3 计算样式断言的令牌对象），StatCard 保持组件自带 glass 令牌族（组件契约）
- 过期时间着色 #ef4444/#f59e0b 与 --accent-red/--accent-yellow 值完全相同 → 改用令牌（零字面量且逐像素等值）
- 修复继承缺陷：图表文字层（#fafafa/#d4d4d8/#a1a1aa/#71717a/labelLine）原为暗色固定值，浅色下环心数字不可读——按 AC4「双主题切换正确」令牌化（--text-primary/--text-regular/--text-secondary/--text-tertiary/--border-strong）并加主题 watcher 重渲；系列色（#7170ff 打头色板）按 Implementation Notes 保持 fc0b2e4 迁移色不动；tooltip 保持暗色固定（与 EP 全局一致）
- canvas 内 tooltip/emphasis 无法用 MCP 合成事件实测（非受信事件不驱动 zrender），以 option 配置断言（单测）+ 双主题截图为证；真实指针 hover 实测落在 StatCard（三态数值前后对照）
- 等价重写中 5 处 ElMessage.error toast 文案、API 调用参数、attributes.expired_time 契约注释逐字保留

## Test Results
- **Tests Executed**: Yes
- **Passed**: 800
- **Failed**: 0
- **Coverage**: 95.2%

## Acceptance Criteria
- [x] 等价重写：重写前盘点旧页全部用户路径（统计卡/图表区/即将过期资源表/时间筛选）逐项列入实现清单，重写后全部保留
- [x] 页面 scoped 样式 <100 行（仅存布局微调），零新增硬编码色字面量
- [x] chrome-devtools-mcp 实测：登录态导航 /cam/dashboard、截图归档、统计卡与图表容器计算样式断言（--bg-surface/--border-* 令牌命中）
- [x] chrome-devtools-mcp 实测：图表配色无旧蓝 #3b82f6 系（与靛紫主题协调）、双主题切换正确、hover 态正常
- [x] 相关测试更新后全量绿，pnpm build + pnpm typecheck 通过

## Notes
覆盖率为 dashboard 模块（--coverage.include=src/views/dashboard/**）：95.18% stmts / 82.4% branch / 88.23% funcs / 96.91% lines，超 80% 目标停止加测。页面 637 行（旧 628）。浏览器实测证据：8 张截图在 D:\Haven\docs\features\ui-unify-phase3-pilot-pages\screenshots\2-dashboard\（计算样式断言数值记录于会话输出；docs 目录不进 git）。实测后现场已恢复（主题 dark、导航回 /cam/assets、天数窗口随重挂载回默认 30）。
