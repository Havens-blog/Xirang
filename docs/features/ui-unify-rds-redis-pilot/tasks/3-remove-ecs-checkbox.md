---
id: "3"
title: "主机页移除常驻多选列"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 3: 主机页移除常驻多选列

## Description

主机页表格当前常驻一列 `col-checkbox`（多选勾选框）与「选中 N 条」字样，用户反馈不需要默认多选状态。本任务移除该列与相关多选状态，交互相应简化。

## Reference Files

- `docs/proposals/ui-unify-rds-redis-pilot/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope）
- `src/views/compute/ecs/index.vue` — 待微调页面（2019 行，表格 `col-checkbox` 列）

## Acceptance Criteria

- [ ] 主机页表格移除常驻 `col-checkbox` 列（表头 + 每行勾选框）
- [ ] 移除「选中 N 条」相关只读状态与无作用的「批量操作/更多状态」禁用占位（如有）
- [ ] `pnpm build` 通过（`pnpm typecheck` 相对基线零新增）
- [ ] 浏览器深/浅两主题：主机页无多选列、无回归、无控制台错误

## Hard Rules

- 仅修改 `src/views/compute/ecs/**` 文件
- 不改主机页既有布局/字段搜索/标签筛选等其余结构

## Implementation Notes

**互助边界**：本任务只做「去多选列」这一处微调，不重写主机页其它部分（用户明确：主机页已是标准）。若移除勾选框后 `selectedIds`/批量相关变量成为死代码，一并清理对应选择逻辑，但**不得**改动主机页的页面骨架、统计状态标、字段搜索栏、表格列定义。

**验收注意**：`pnpm lint` 历史 3 error 判据「相对基线零新增」；浏览器实测经 nginx `http://localhost:8888/cam/compute/ecs`。