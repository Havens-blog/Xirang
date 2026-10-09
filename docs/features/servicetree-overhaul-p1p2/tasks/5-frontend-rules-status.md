---
id: "5"
title: "前端：规则列表真实状态呈现（去「-」遮掩）"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies:
  - "1"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.fix"
mainSession: false
---

# 5: 前端：规则列表真实状态呈现（去「-」遮掩）

## Description
**代码仓：D:/Haven/e-cam-web**。任务 1 后端透出 node_name/执行统计后，规则列表页（`rules/index.vue`）
把「目标节点 -」「状态空」「匹配数 -」的遮掩呈现替换为真实数据：node_name 直绑、启用状态用
el-switch/tag（含行内启用切换调既有更新接口）、匹配数与最近执行时间列真实渲染（从未执行显示「未执行」
而非「-」）。

## Reference Files
- `docs/proposals/servicetree-overhaul-p1p2/proposal.md` — SC1
- `D:/Haven/e-cam-web/src/views/service-tree/rules/index.vue` — :61 node_name 列、状态列、匹配数列
- 任务 1 透出字段：node_name / enabled / last_match_count / last_executed_at

## Acceptance Criteria
- [ ] 目标节点列显示真实节点名；状态列真实反映 enabled（含行内切换，调既有更新 API 后刷新）
- [ ] 匹配数列：有执行记录显示数字，从未执行显示「未执行」（非「-」）；最近执行时间列新增
- [ ] `pnpm typecheck` + vitest 全绿（并发在途错误除外）；`git diff` 仅 rules/index.vue

## Hard Rules
仅改 `src/views/service-tree/rules/index.vue`。**不碰** RuleFormDialog（任务 4）、总览页（任务 7）、
`accounts/index.vue`、logs、`components.d.ts`。提交显式 pathspec；**提交前确认当前分支**（不在 main 时停止报告）。

## Implementation Notes
- 行内启用切换复用既有更新规则接口（PUT），乐观更新 + 失败回滚。
- 时间格式化用页面既有 formatDateTime 工具（如有）或 `@/utils/format`。
