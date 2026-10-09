---
id: "4"
title: "alert/events 告警事件等价重写（监控页样板 + 浏览器实测）"
priority: "P1"
estimated_time: "1.5h"
complexity: "high"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 4: alert/events 告警事件等价重写（监控页样板 + 浏览器实测）

## Description
监控高频页标杆重写（239 行）：PageContainer + DataTable 承载事件查询列表，语义状态色徽章令牌化，密度切换生效，⌘K 跳转落点验证（CommandPalette 菜单项应能直达本页）。

## Reference Files
- `docs/proposals/ui-unify-phase3-pilot-pages/proposal.md` — Proposed Solution / Key Scenarios
- `src/views/alert/events/index.vue` — 重写对象（查询条件/列/状态徽章盘点入 AC）
- `src/components/{PageContainer,DataTable}/index.vue` — Phase 2 组件

## Acceptance Criteria
- [ ] 等价重写：旧页用户路径盘点（事件查询条件、列结构、状态展示、分页）逐项保留
- [ ] 语义状态色徽章全部令牌化（success/warning/danger 对应 --color-* 令牌），DataTable 密度切换生效
- [ ] 页面 scoped 样式 <100 行，零新增硬编码色字面量
- [ ] chrome-devtools-mcp 实测：导航本页、截图归档、查询链路走通、⌘K 面板搜索并跳转到本页成功、计算样式断言
- [ ] chrome-devtools-mcp 实测：双主题 + hover 正常
- [ ] 相关测试更新后全量绿，`pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 功能等价边界：不顺手新增/优化规格外功能
- 零新增硬编码颜色字面量

## Implementation Notes
- 告警事件数据可能为空：空态由 DataTable 内置 StateBlock 承载，实测时空态本身也是验收点之一
