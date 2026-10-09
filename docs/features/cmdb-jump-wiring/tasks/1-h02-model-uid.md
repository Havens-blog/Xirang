---
id: "1"
title: "H-02：instances 页接 model_uid 过滤（模型→实例下钻）"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.fix"
mainSession: false
---

# 1: H-02：instances 页接 model_uid 过滤（模型→实例下钻）

## Description
**代码仓：D:/Haven/e-cam-web**。模型列表「查看实例」跳 `/cmdb/instances?model_uid=<uid>`
（`cmdb/models/index.vue:428`），但 instances 页未 import `useRoute`、从不读取——落到未过滤全量列表。
后端 `cmdb/web/instance_handler.go:213` 已支持 `model_uid` 查询参数。接线：读 query → 请求参数 + 可清除过滤标签。

## Reference Files
- `docs/proposals/cmdb-jump-wiring/proposal.md` — Proposed Solution, Success Criteria
- `D:/Haven/e-cam-web/src/views/cmdb/instances/index.vue` — filters(287)/列表请求构造/分页
- `D:/Haven/e-cam-service/internal/cmdb/web/instance_handler.go:213` — 后端已支持 model_uid（只读参考，不改后端）

## Acceptance Criteria
- [ ] instances 页 import `useRoute`，onMounted 首次 fetch 前读 `route.query.model_uid` 存入响应式状态并纳入列表请求参数
- [ ] 存在 model_uid 过滤时显示可清除标签（如 el-tag closable「按模型过滤: <uid>」），清除后回到全量并重新加载
- [ ] 无 query 直达行为与现状完全一致（零改动）
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅 cmdb/instances/index.vue

## Hard Rules
仅改 `src/views/cmdb/instances/index.vue`。**不碰** 10 个资产目标页（任务 2/3）、发送方切换（任务 4）、
`accounts/index.vue`（并发脏）、logs、后端、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- 读取时机必须在首次 fetch 之前（避免先全量后过滤的二次请求）。
- model_uid 直接作为请求参数（后端按它过滤），不混入 keyword。
