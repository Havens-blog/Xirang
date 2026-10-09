---
id: "7"
title: "前端：树徽标 + 关联资产聚合语义 + 环境疑异标红"
priority: "P1"
estimated_time: "2.5h"
complexity: "high"
dependencies:
  - "6"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.feature"
mainSession: false
---

# 7: 前端：树徽标 + 关联资产聚合语义 + 环境疑异标红

## Description
**代码仓：D:/Haven/e-cam-web**。服务树总览（`views/service-tree/index.vue`）二期可见性：
① 树每个节点显示子树资产数徽标（el-tag badge，0 则灰显）——加载时批量拉取各节点 summary（或一次聚合接口批量版）；
② 「关联资产」tab 聚合语义修正：根/中间节点展示子树聚合资产（现状空），叶子展示直属绑定；
③ 环境疑异标红：绑定/关联资产列表中 suspicious 资产行标红 + tooltip 疑异原因（如「命名含 prod 却绑定 dev 环境」）。

## Reference Files
- `docs/proposals/servicetree-overhaul-p1p2/proposal.md` — 二期方案 7/8/9、SC5/SC6
- `D:/Haven/e-cam-web/src/views/service-tree/index.vue` — 总览（树 + 绑定/关联资产双 tab）
- 任务 6 产物：`GET /nodes/:id/asset-summary`

## Acceptance Criteria
- [ ] 树节点显示子树资产数徽标（非 0 数字标签，0 灰显或隐藏）；点击节点详情与徽标一致
- [ ] 关联资产 tab：根/中间节点显示子树聚合资产清单（含来源节点列更佳），叶子节点显示直属
- [ ] suspicious 资产行醒目标红（行背景或图标）+ tooltip 展示疑异原因
- [ ] `pnpm typecheck` + vitest 全绿（并发在途错误除外）；`git diff` 仅 service-tree 总览相关前端文件

## Hard Rules
仅改 `src/views/service-tree/index.vue`（+ 必要子组件/`src/api/` 封装）。**不碰** rules 页（任务 4/5）、
`accounts/index.vue`、logs、`components.d.ts`。提交显式 pathspec；**提交前确认当前分支**（不在 main 时停止报告）。

## Implementation Notes
- 徽标数据源：任务 6 summary 是单节点接口——若逐节点拉取过重，评估后端加批量参数（`?ids=1,2,3`）或前端按需懒加载（展开节点时拉）。
- 关联资产聚合复用任务 6 的 summary 明细（或扩展 summary 接口带 items 分页——以任务 6 实际产物为准，必要时补接口参数并在本任务内说明）。
- 环境疑异清单直接消费 summary.suspicious。
