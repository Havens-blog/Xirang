---
id: "3"
title: "H-03 接收方：name 系 2 页 + searchKeyword 系 2 页 + ecs"
priority: "P1"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.fix"
mainSession: false
---

# 3: H-03 接收方：name 系 2 页 + searchKeyword 系 2 页 + ecs

## Description
**代码仓：D:/Haven/e-cam-web**。其余 5 个目标页接线：vpc/eip 读 `query.search` → `filters.name`
（两页已有 name 过滤字段）；nas/oss 读 → `searchKeyword` ref（既有搜索框绑定）；ecs 读 → `filters.keyword`
（与既有 `route.query.uid` 读取并存，uid 精确/keyword 模糊不同维度，共存无害）。

## Reference Files
- `docs/proposals/cmdb-jump-wiring/proposal.md` — Proposed Solution, Success Criteria
- `D:/Haven/e-cam-web/src/views/network/vpc/index.vue`（filters.name:299）、`network/eip/index.vue`（:289）
- `D:/Haven/e-cam-web/src/views/storage/nas/index.vue`（searchKeyword:285）、`storage/oss/index.vue`（searchKeyword、列表 name:394）
- `D:/Haven/e-cam-web/src/views/compute/ecs/index.vue`（onMounted:1297 已读 query.uid；filters.keyword:588）

## Acceptance Criteria
- [ ] vpc/eip：首载前读 `query.search` 赋 `filters.name`；nas/oss：赋 `searchKeyword`
- [ ] ecs：`query.search` 赋 `filters.keyword`（保留既有 uid 读取逻辑不动）
- [ ] 各页带 search 进入首屏即过滤结果；无 query 直达零改动
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅这 5 个 index.vue

## Hard Rules
仅改 5 个 index.vue（network/vpc、network/eip、storage/nas、storage/oss、compute/ecs）。
**不碰** keyword 系 5 页（任务 2）、发送方（任务 4）、`accounts/index.vue`（并发脏）、logs、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- 统一三行模式同任务 2；nas/oss 的 searchKeyword 是搜索框绑定 ref，赋值后首载即生效。
- ecs 的读取加在现有 onMounted 内 uid 读取之后，互不影响。
