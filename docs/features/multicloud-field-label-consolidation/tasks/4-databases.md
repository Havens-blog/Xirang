---
id: "4"
title: "数据库域(rds/redis/mongodb)状态单源收口"
priority: "P1"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 4: 数据库域(rds/redis/mongodb)状态单源收口

## Description
**代码仓：D:/Haven/e-cam-web**。收敛数据库域 rds/redis/mongodb 三产品的状态词表:每个产品的 index.vue 与详情抽屉**各写一份内联 `getStatusText`**(键集还不一致,mongodb 只认 `running/stopped`,redis 多认 `normal`)。统一为单一共享词表,index 与抽屉 import 同一份。

## Reference Files
- `docs/proposals/multicloud-field-label-consolidation/proposal.md` — Proposed Solution, Scope(In Scope), Success Criteria
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts` — 单一来源模块
- `D:/Haven/e-cam-web/src/views/databases/rds/index.vue` — getStatusText(:549)
- `D:/Haven/e-cam-web/src/views/databases/rds/components/RdsDetailDrawer.vue` — getStatusText(:112)
- `D:/Haven/e-cam-web/src/views/databases/redis/index.vue` — getStatusText(:546)
- `D:/Haven/e-cam-web/src/views/databases/redis/components/RedisDetailDrawer.vue` — getStatusText(:95)
- `D:/Haven/e-cam-web/src/views/databases/mongodb/index.vue` — getStatusText(:545)
- `D:/Haven/e-cam-web/src/views/databases/mongodb/components/MongodbDetailDrawer.vue` — getStatusText(:96)

## Acceptance Criteria
- [ ] rds/redis/mongodb 状态词表收敛到 `src/utils/fieldLabels.ts` 命名导出(如 `RDS_STATUS_LABELS`/`REDIS_STATUS_LABELS`/`MONGODB_STATUS_LABELS`)
- [ ] 六个文件(3 index + 3 drawer)删除本地内联 `getStatusText` map、改为 import 共享词表 + `labelOf`/`labelOfLenient`
- [ ] 语义保留:`running`→运行中、`stopped`→已停止、redis `normal`→运行中;列表页词汇 canonical(键集以覆盖最全者为准)、抽屉对齐
- [ ] 该域 grep 不再有本地内联 `getStatusText` map 定义(仅余 shared import)
- [ ] `pnpm typecheck` 通过、既有单测不回归

## Hard Rules
仅修改以下文件:
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts`
- `D:/Haven/e-cam-web/src/views/databases/rds/index.vue`
- `D:/Haven/e-cam-web/src/views/databases/rds/components/RdsDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/databases/redis/index.vue`
- `D:/Haven/e-cam-web/src/views/databases/redis/components/RedisDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/databases/mongodb/index.vue`
- `D:/Haven/e-cam-web/src/views/databases/mongodb/components/MongodbDetailDrawer.vue`
不改字段键、不改后端、不改数据契约;不碰清单外文件。

## Implementation Notes
- 单一来源位置统一:`src/utils/fieldLabels.ts`。
- 三产品价态高度同构(running/stopped 为主、redis 多 normal),收敛时注意抽屉若用 `getStatusClass`(色调用)只收相应样式,不复制文案。
- 复杂度 medium 理由:6 文件、同构模式、无并发 WIP、漂移低。
- 提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径,禁 `git add -A`。