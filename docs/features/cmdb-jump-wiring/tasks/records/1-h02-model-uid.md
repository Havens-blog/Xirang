---
status: "completed"
started: "2026-09-20 16:49"
completed: "2026-09-20 16:56"
time_spent: "~7m"
---

# Task Record: 1 H-02：instances 页接 model_uid 过滤（模型→实例下钻）

## Summary
H-02：cmdb instances 页接入 model_uid 下钻过滤。import useRoute；onMounted 首次 fetch 前读 route.query.model_uid 存入 filters.model_uid；buildListParams 发送 params.model_uid（后端 instance_handler.go List 已支持该参数，只读确认）；激活时显示可清除 el-tag 过滤标签「按模型过滤: <uid>」，清除后回全量并重载（复用 handleSearch：page=1 + refetch）；导出全量拉取共用 buildListParams，过滤口径一致；无 query 直达行为零改动。仅改 src/views/cmdb/instances/index.vue。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/views/cmdb/instances/index.vue

### Key Decisions
- model_uid 加入既有 filters reactive 与 buildListParams（列表与导出共用同一参数构造），独立参数不混入 keyword（Implementation Notes 要求）
- 读取时机：onMounted 内 initColumnSettings 之后、fetchInstances 之前，保证首屏即过滤结果且无先全量后过滤的二次请求
- 过滤标签置于 filters-left 行尾（el-tag closable warning size=small），清除复用 handleSearch 保持既有分页重置语义
- 按 Hard Rules 未触碰：10 个资产目标页（任务 2/3）、发送方 handleViewDetail 切换（任务 4）、accounts/index.vue、logs、后端、components.d.ts

## Test Results
- **Tests Executed**: Yes
- **Passed**: 108
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] instances 页 import useRoute，onMounted 首次 fetch 前读 route.query.model_uid 存入响应式状态并纳入列表请求参数
- [x] 存在 model_uid 过滤时显示可清除标签（el-tag closable「按模型过滤: <uid>」），清除后回到全量并重新加载
- [x] 无 query 直达行为与现状完全一致（零改动）
- [x] pnpm typecheck + vitest 通过；git diff 仅 cmdb/instances/index.vue

## Notes
验证证据：pnpm typecheck 仅 2 个既有错误（src/views/logs/*.test.ts，并发会话在途工作，任务说明明确忽略）；本文件 0 错误。vitest 全量：28 文件/551 用例 → 548 通过、3 失败，3 个失败全部位于 src/views/cert/ledger/index.test.ts（并发 cert 工作，本任务未触碰）；剔除并发 cert/logs 模块后的在范围运行：11 文件/108 用例全部通过（testsPassed=108 取自该次真实 runner 输出）。coverage 未配置 runner 报告，故省略。git 提交 564fe3b 仅含 src/views/cmdb/instances/index.vue（+30/-1）；并发脏文件（accounts/index.vue、components.d.ts、.env.development、element-theme.scss、.claude/）未 staged 未提交，未使用 git add -A。
