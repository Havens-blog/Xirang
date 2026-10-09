---
id: "4"
title: "H-03 发送方：handleViewDetail 改发 asset_name（接收方就绪后切换）"
priority: "P1"
estimated_time: "30min"
complexity: "low"
dependencies:
  - "2"
  - "3"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.fix"
mainSession: false
---

# 4: H-03 发送方：handleViewDetail 改发 asset_name（接收方就绪后切换）

## Description
**代码仓：D:/Haven/e-cam-web**。接收方（任务 2/3）就绪后，发送方 `handleViewDetail`
（`cmdb/instances/index.vue:578`）把 `query.search` 的值从 `row.asset_id` 改为
`row.asset_name || row.asset_id`——后端各列表 `name` 模糊搜索匹配的是展示名；
asset_id（云 ID 如 i-bp1xxx）在 name 搜索里匹配不上。asset_name 可空（`InstanceVO.asset_name?`），空时兜底 asset_id。

## Reference Files
- `docs/proposals/cmdb-jump-wiring/proposal.md` — Proposed Solution, Success Criteria
- `D:/Haven/e-cam-web/src/views/cmdb/instances/index.vue:572-586` — handleViewDetail + assetTypeRouteMap(243)

## Acceptance Criteria
- [ ] `query.search` 值改为 `row.asset_name || row.asset_id`
- [ ] 10 个路由目标与未映射类型的警告分支行为不变
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅 cmdb/instances/index.vue

## Hard Rules
仅改 `src/views/cmdb/instances/index.vue`。**不碰** 10 目标页（任务 2/3 已完成）、
`accounts/index.vue`（并发脏）、logs、后端、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- 依赖任务 2/3 完成（中间态切换会导致发 name 但目标页不读的空窗）。
- asset_name 可能含特殊字符，router.push query 序列化自动处理，无需 encode。
