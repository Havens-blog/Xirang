---
status: "completed"
started: "2026-09-30 11:40"
completed: "2026-09-30 11:49"
time_spent: "~9m"
---

# Task Record: 4 数据库域(rds/redis/mongodb)状态单源收口

## Summary
数据库域(rds/redis/mongodb)状态单源收口：在 src/utils/fieldLabels.ts 新增 RDS_STATUS_LABELS/REDIS_STATUS_LABELS/MONGODB_STATUS_LABELS 三个命名导出（键集以列表页 canonical 为准，含大小写变体），6 个文件(3 index + 3 drawer)删除本地内联 getStatusText map，改为 import 共享词表 + labelOfLenient 一行包装。语义保留：running→运行中、stopped/shutdown→已停止、redis normal→运行中、creating→创建中、deleting→删除中；抽屉对齐补齐原缺失键(rds/mongodb 抽屉补 shutdown/creating/deleting，redis 抽屉补 shutdown/creating)。getStatusClass 为样式逻辑按 Hard Rules 未动。

## Changes

### Files Created
无

### Files Modified
- src/utils/fieldLabels.ts
- src/views/databases/rds/index.vue
- src/views/databases/rds/components/RdsDetailDrawer.vue
- src/views/databases/redis/index.vue
- src/views/databases/redis/components/RedisDetailDrawer.vue
- src/views/databases/mongodb/index.vue
- src/views/databases/mongodb/components/MongodbDetailDrawer.vue

### Key Decisions
- canonical 键集取 index 与 drawer 的并集（index 为超集）：rds/mongodb = running/stopped/shutdown/creating/deleting，redis = running/normal/stopped/shutdown/creating
- 查表走 labelOfLenient(X_STATUS_LABELS, status, status || '-')：空值→'-'、未知原始值透出，与原 map[toLowerCase()]||status 行为等价；大小写覆盖通过词表内 running/Running/RUNNING 变体键实现，与 vswitch/eni/lb 已收口页模式一致
- getStatusClass(色调用) 不在收口范围，按 Implementation Notes 原样保留
- 工作区存在并发会话脏文件(logs/service-tree/security-group)，与本次范围无交集，仅按 pathspec 提交本 7 文件

## Test Results
- **Tests Executed**: Yes
- **Passed**: 919
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] rds/redis/mongodb 状态词表收敛到 fieldLabels.ts 命名导出
- [x] 六个文件删除本地内联 getStatusText map、改为 import 共享词表 + labelOfLenient
- [x] 语义保留: running→运行中、stopped→已停止、redis normal→运行中；列表页 canonical、抽屉对齐
- [x] 该域 grep 不再有本地内联 getStatusText map 定义
- [x] pnpm typecheck 通过、既有单测不回归

## Notes
验证证据: vue-tsc -b 零错误；npx vitest run 52 文件/919 用例全过；eslint(7 个改动文件) exit 0。coverage 字段无运行器报告值(该 6 视图文件无既有单测)，填 0 表示未测量而非 0% 覆盖。pnpm test 脚本不存在，测试经 npx vitest run 执行。
