---
id: "4"
title: "前端：规则条件编辑器升级 + dry-run 预览交互"
priority: "P1"
estimated_time: "2.5h"
complexity: "high"
dependencies:
  - "1"
  - "2"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.feature"
mainSession: false
---

# 4: 前端：规则条件编辑器升级 + dry-run 预览交互

## Description
**代码仓：D:/Haven/e-cam-web**。`rules/components/RuleFormDialog.vue` 条件编辑器升级：
① 字段选项改为「基础（name/asset_id/model_uid/region——region 现后端已支持）+ 自由 tag key 输入
（el-select allow-create filterable，保留 env/service/team 常用建议）+ attributes.xxx 自由输入 + 资源组
`attributes.project_id` 预设项」；② 保存前「试运行」按钮调 dry-run API（任务 2）弹层预览命中资产清单
（表格：资产/云平台/地域/当前绑定状态），命中 0 条给显性警告。规则列表页状态/匹配数/执行时间/
node_name 真实呈现（任务 1 后端已透出，删「-」兜底的遮掩）。

## Reference Files
- `docs/proposals/servicetree-overhaul-p1p2/proposal.md` — 一期方案 2/3、SC2/SC3
- `D:/Haven/e-cam-web/src/views/service-tree/rules/components/RuleFormDialog.vue` — 条件编辑器（:78-92 字段选项）
- `D:/Haven/e-cam-web/src/views/service-tree/rules/index.vue` — 列表（:61 node_name 列）
- dry-run API：任务 2 产物 `POST /service-tree/rules/dry-run`（前端补 api 封装）

## Acceptance Criteria
- [ ] 字段下拉：基础 4 项 + 自由 tag key（allow-create，env/service/team 为建议项）+ attributes.xxx 自由输入 + 资源组预设项
- [ ] region 选项保留且后端可匹配（任务 1 已支持）
- [ ] 「试运行」按钮：调 dry-run 弹层展示命中资产表（资产/云平台/地域/当前绑定状态），0 命中显性警告文案
- [ ] 规则列表：目标节点名真实、状态开关真实、匹配数/执行时间真实呈现（无「-」遮掩）
- [ ] `pnpm typecheck` + vitest 全绿（并发在途错误除外）；`git diff` 仅 service-tree 前端文件

## Hard Rules
仅改 `src/views/service-tree/rules/` 内文件（+ `src/api/` 补 dry-run 封装若需）。
**不碰** 总览页（任务 5/7）、`accounts/index.vue`（并发脏）、logs、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径；**提交前确认当前分支**（HEAD 可能不在 main，
在并发会话分支时停止并报告）。

## Implementation Notes
- 自由输入用 el-select 的 allow-create + filterable（Element Plus 原生支持，无需新组件）。
- dry-run 弹层复用 el-dialog + el-table 简表即可；结果为空时 warning 而非 error。
- 匹配数列直接绑任务 1 透出的统计字段。
