---
status: "completed"
started: "2026-09-30 11:51"
completed: "2026-09-30 12:04"
time_spent: "~13m"
---

# Task Record: 5 中间件+存储域(es/kafka/nas)状态单源收口

## Summary
中间件+存储域(es/kafka/nas)状态单源收口：fieldLabels.ts 新增 ES/KAFKA/NAS_STATUS_LABELS 三张命名导出词表(列表页 canonical 键集+大小写变体键)；6 个视图文件(3 index + 3 drawer)删除本地内联 getStatusText map / nas 本地 STATUS_TEXT，改为 import 共享词表 + labelOfLenient 一行式(getStatusText 调用点与 getStatusClass 色调口径不动)。语义保留：es active→运行中、kafka serving→运行中、nas pending→创建中；nas getExportValue 空 status 导出空串行为保留(不落 labelOfLenient '-' 兜底)；kafka 词表补 shutdown→已停止(与域内 getStatusClass 停止态口径一致，同 T4 rds 先例)。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/utils/fieldLabels.ts
- e-cam-web/src/views/middleware/elasticsearch/index.vue
- e-cam-web/src/views/middleware/elasticsearch/components/EsDetailDrawer.vue
- e-cam-web/src/views/middleware/kafka/index.vue
- e-cam-web/src/views/middleware/kafka/components/KafkaDetailDrawer.vue
- e-cam-web/src/views/storage/nas/index.vue
- e-cam-web/src/views/storage/nas/components/NasDetailDrawer.vue

### Key Decisions
- 沿用 T3/T4 已验证模式：词表进 fieldLabels.ts + labelOfLenient 一行式包装，大小写不敏感行为靠词表变体键(running/Running/RUNNING)承载
- nas index 本地 STATUS_TEXT 被 getStatusText 与 getExportValue 双点引用，一并迁移；getExportValue 空 status 仍导出空串(原行为)，不用 '-' 兜底
- kafka 词表补 shutdown→已停止：index/drawer 两处 getStatusClass 均已把 shutdown 归入停止态色调，文案对齐同口径
- getStatusClass/getStatusTones 色调数组保留本地(色调口径非文案词表，不在收敛范围)

## Test Results
- **Tests Executed**: Yes
- **Passed**: 919
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] es/kafka/nas 状态词表收敛到 src/utils/fieldLabels.ts 命名导出(ES/KAFKA/NAS_STATUS_LABELS)
- [x] 六个文件(3 index + 3 drawer)删除本地内联 getStatusText map、改为 import 共享词表 + labelOfLenient
- [x] 语义保留：es active→运行中、kafka serving→运行中、running→运行中、stopped→已停止、nas pending→创建中；列表页 canonical、抽屉对齐
- [x] 该域 grep 不再有本地内联 getStatusText map 定义(仅余 shared import)
- [x] pnpm typecheck 通过、既有单测不回归

## Notes
验证证据：just compile(vue-tsc -b) 零错误；npx vitest run 52 文件 919/919 全绿；eslint 7 个改动文件零告警。prettier --check 7 文件 warn 为既有状态(HEAD 版本与 T4 已入库文件同样 non-conform，just fmt 配方本身 || true 非阻断)，未跑 --write 以免波及并发 WIP 脏文件。coverage 为 null：vitest 无 coverage provider。工作树含并发 WIP(iam/logs 等)未触碰，提交将用显式 pathspec。
