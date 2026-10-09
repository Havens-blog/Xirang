---
id: "2"
title: "dashboard 仪表盘等价重写（组件化 + 浏览器实测）"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 2: dashboard 仪表盘等价重写（组件化 + 浏览器实测）

## Description
用 PageContainer 骨架重写仪表盘（628 行）：统计卡排布 + 云厂商分布/资产类型分布/地域 TOP10/成本 TOP10 四图表区。功能等价（现有用户路径全保留），图表配色与靛紫主题协调，PageContainer 自承载页面骨架。

## Reference Files
- `docs/proposals/ui-unify-phase3-pilot-pages/proposal.md` — Proposed Solution / Key Scenarios / Success Criteria
- `src/views/dashboard/index.vue` — 重写对象（旧页用户路径盘点入 AC）
- `src/components/PageContainer/index.vue` — 骨架组件
- `docs/features/ui-unify-phase2-components/tasks/8-ui-kit-demo.md` — demo 页样板参照

## Acceptance Criteria
- [ ] 等价重写：重写前盘点旧页全部用户路径（统计卡/图表区/即将过期资源表/时间筛选）逐项列入实现清单，重写后全部保留
- [ ] 页面 scoped 样式 <100 行（仅存布局微调），零新增硬编码色字面量
- [ ] chrome-devtools-mcp 实测：登录态导航 /cam/dashboard、截图归档、统计卡与图表容器计算样式断言（--bg-surface/--border-* 令牌命中）
- [ ] chrome-devtools-mcp 实测：图表配色无旧蓝 #3b82f6 系（与靛紫主题协调）、双主题切换正确、hover 态正常
- [ ] 相关测试更新后全量绿，`pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 功能等价边界：不顺手新增/优化规格外功能
- 零新增硬编码颜色字面量

## Implementation Notes
- 嘉立创空间（jlc）有真实数据（5.0k 资产、5 云厂商）；数据断言用"存在性/样式"而非具体数值（防数据变动）
- ECharts 主题配色保持 fc0b2e4 迁移后的系列色（#7170ff 交互紫打头）
