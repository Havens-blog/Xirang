---
id: "2"
title: "H-03 接收方：keyword 系 5 页接 query.search"
priority: "P0"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.fix"
mainSession: false
---

# 2: H-03 接收方：keyword 系 5 页接 query.search

## Description
**代码仓：D:/Haven/e-cam-web**。实例详情「查看资产」带 `query.search` 跳转，但 5 个 keyword 系目标页
（rds/redis/mongodb/kafka/elasticsearch）从不读取。统一三行模式接线：import useRoute → onMounted 首载前
读 `route.query.search` 赋 `filters.keyword` → 依赖既有首载发出过滤请求。

## Reference Files
- `docs/proposals/cmdb-jump-wiring/proposal.md` — Proposed Solution, Success Criteria
- `D:/Haven/e-cam-web/src/views/databases/rds/index.vue`（filters.keyword:232、handleSearch:291）
- `D:/Haven/e-cam-web/src/views/databases/redis/index.vue`、`databases/mongodb/index.vue`
- `D:/Haven/e-cam-web/src/views/middleware/kafka/index.vue`、`middleware/elasticsearch/index.vue`

## Acceptance Criteria
- [ ] 5 页各自：import useRoute；onMounted 首次 fetch 前读 `route.query.search`（string 化、空值安全）赋 `filters.keyword`
- [ ] 带 search 进入时首屏即为过滤结果（无二次全量请求）；无 query 直达行为零改动
- [ ] `pnpm typecheck` + vitest 通过；`git diff` 仅这 5 个 index.vue

## Hard Rules
仅改 5 个 index.vue（databases/rds、databases/redis、databases/mongodb、middleware/kafka、middleware/elasticsearch）。
**不碰** vpc/eip/nas/oss/ecs（任务 3）、发送方（任务 4）、`accounts/index.vue`（并发脏）、logs、`components.d.ts`。
提交 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- 统一模式：`const route = useRoute()`；在现有 onMounted（或首次 fetch 调用前）加
  `const s = route.query.search; if (typeof s === 'string' && s) filters.keyword = s`。
- 若某页 onMounted 不直接调 fetch（如先 loadXxx 包装），确保赋值在最终 fetch 前执行即可。
