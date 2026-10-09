---
status: "completed"
started: "2026-09-20 16:57"
completed: "2026-09-20 17:04"
time_spent: "~7m"
---

# Task Record: 2 H-03 接收方：keyword 系 5 页接 query.search

## Summary
H-03 接收方：keyword 系 5 页（databases/rds、databases/redis、databases/mongodb、middleware/kafka、middleware/elasticsearch）接 query.search。统一三行模式：import useRoute → 现有 onMounted 内首次 fetchInstances() 之前读 route.query.search（typeof string 且非空）赋 filters.keyword → 依赖既有首载发出过滤请求。无新增请求（零二次全量请求）；无 query 直达时 if 不命中，行为零改动。提交 f4fd016。

## Changes

### Files Created
无

### Files Modified
- src/views/databases/rds/index.vue
- src/views/databases/redis/index.vue
- src/views/databases/mongodb/index.vue
- src/views/middleware/kafka/index.vue
- src/views/middleware/elasticsearch/index.vue

### Key Decisions
- 赋值放在现有 onMounted 内、既有 fetchInstances() 调用之前，不新增请求调用，保证首屏即过滤结果且无二次全量请求
- useRoute 导入风格对齐项目既有惯例（src/views/compute/ecs/index.vue:512-513：vue 导入后单独一行 import { useRoute } from 'vue-router'）
- 读取模式严格按任务 Implementation Notes：const s = route.query.search; if (typeof s === 'string' && s) filters.keyword = s（string 化 + 空值安全）
- 严守 Hard Rules：仅改 5 个 index.vue，未触碰 vpc/eip/nas/oss/ecs、发送方、accounts、logs、components.d.ts；git add 使用显式 pathspec

## Test Results
- **Tests Executed**: Yes
- **Passed**: 548
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 5 页各自：import useRoute；onMounted 首次 fetch 前读 route.query.search（string 化、空值安全）赋 filters.keyword
- [x] 带 search 进入时首屏即为过滤结果（无二次全量请求）；无 query 直达行为零改动
- [x] pnpm typecheck + vitest 通过；git diff 仅这 5 个 index.vue

## Notes
vitest 全量 27 文件 548 用例通过、0 失败（运行命令 pnpm vitest run --exclude src/views/cert/ledger/index.test.ts；排除项为并发在途的 cert/ledger 既有失败，与本任务 5 个文件无关，全量含该文件时为 548 passed / 3 failed，均为 pre-existing）。pnpm typecheck 仅报 src/views/logs/ 两处既有错误（并发在途），5 个改动文件零新增错误。本任务 5 个视图页无既有单测覆盖（仓库惯例 views 层不配单测，targeted tests 位于 components/__tests__ 与 api/*.test.ts），故以全量套件（排除并发损坏文件）作为测试证据。coverage 未配置覆盖率运行器，未上报。git diff 范围核验：仅 5 个 index.vue（+50/-5），其余脏文件（.env.development、element-theme.scss、components.d.ts、accounts/index.vue、.claude/）为并发会话所有，未触碰未暂存。
