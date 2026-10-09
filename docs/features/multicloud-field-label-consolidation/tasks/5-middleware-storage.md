---
id: "5"
title: "中间件+存储域(es/kafka/nas)状态单源收口"
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

# 5: 中间件+存储域(es/kafka/nas)状态单源收口

## Description
**代码仓：D:/Haven/e-cam-web**。收敛中间件 es/kafka 与存储 nas 三产品的状态词表:每个产品的 index.vue 与详情抽屉**各写一份内联 `getStatusText`**(es 多认 `active`、kafka 多认 `serving`、nas 多认 `pending` 且含大写 Running/Stopped 变体)。统一为单一共享词表,index 与抽屉 import 同一份。

## Reference Files
- `docs/proposals/multicloud-field-label-consolidation/proposal.md` — Proposed Solution, Scope(In Scope), Success Criteria
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts` — 单一来源模块
- `D:/Haven/e-cam-web/src/views/middleware/elasticsearch/index.vue` — getStatusText(:540)
- `D:/Haven/e-cam-web/src/views/middleware/elasticsearch/components/EsDetailDrawer.vue` — getStatusText(:79)
- `D:/Haven/e-cam-web/src/views/middleware/kafka/index.vue` — getStatusText(:545)
- `D:/Haven/e-cam-web/src/views/middleware/kafka/components/KafkaDetailDrawer.vue` — getStatusText(:78)
- `D:/Haven/e-cam-web/src/views/storage/nas/index.vue` — getStatusText(:673)
- `D:/Haven/e-cam-web/src/views/storage/nas/components/NasDetailDrawer.vue` — getStatusText(:190)

## Acceptance Criteria
- [ ] es/kafka/nas 状态词表收敛到 `src/utils/fieldLabels.ts` 命名导出(如 `ES_STATUS_LABELS`/`KAFKA_STATUS_LABELS`/`NAS_STATUS_LABELS`)
- [ ] 六个文件(3 index + 3 drawer)删除本地内联 `getStatusText` map、改为 import 共享词表 + `labelOf`/`labelOfLenient`
- [ ] 语义保留:es `active`→运行中、kafka `serving`→运行中、`running`→运行中、`stopped`→已停止、nas `pending`→创建中;列表页词汇 canonical(键集覆盖最全为准)、抽屉对齐
- [ ] 该域 grep 不再有本地内联 `getStatusText` map 定义(仅余 shared import)
- [ ] `pnpm typecheck` 通过、既有单测不回归

## Hard Rules
仅修改以下文件:
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts`
- `D:/Haven/e-cam-web/src/views/middleware/elasticsearch/index.vue`
- `D:/Haven/e-cam-web/src/views/middleware/elasticsearch/components/EsDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/middleware/kafka/index.vue`
- `D:/Haven/e-cam-web/src/views/middleware/kafka/components/KafkaDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/storage/nas/index.vue`
- `D:/Haven/e-cam-web/src/views/storage/nas/components/NasDetailDrawer.vue`
不改字段键、不改后端、不改数据契约;不碰清单外文件。

## Implementation Notes
- 单一来源位置统一:`src/utils/fieldLabels.ts`。
- 三产品价态同构(running/stopped 为主+各自少量扩展键),收敛时把 es `active`/kafka `serving`/nas `pending` 归入各自词表,抽屉 `getStatusClass` 色调用只收样式不复制文案。
- 复杂度 medium 理由:6 文件、同构模式、无并发 WIP、漂移低。
- 提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径,禁 `git add -A`。