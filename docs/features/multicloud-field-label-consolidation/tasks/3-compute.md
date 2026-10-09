---
id: "3"
title: "计算域(ecs/image/disk)状态+枚举单源收口"
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

# 3: 计算域(ecs/image/disk)状态+枚举单源收口

## Description
**代码仓：D:/Haven/e-cam-web**。收敛计算域 ecs/image/disk 三个产品的状态与产品枚举词表:ecs 在 index 多处内联 map(如 `running→运行中/stopped→已关机`)、image 有 `IMAGE_STATUS_LABELS`+`IMAGE_TYPE_LABELS`+`OS_TYPE_LABELS` 且抽屉手写 `statusLabels`、disk 的 `getStatusText`+`getDiskCategory` 在 index 与抽屉**双份重复**。统一为单一共享来源,index 与抽屉 import 同一份。

## Reference Files
- `docs/proposals/multicloud-field-label-consolidation/proposal.md` — Proposed Solution, Scope(In Scope), Success Criteria
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts` — 单一来源模块
- `D:/Haven/e-cam-web/src/views/compute/ecs/index.vue` — 多处内联 status/display map(约 :896/:912/:924/:1108)
- `D:/Haven/e-cam-web/src/views/compute/ecs/components/InstanceDetailDrawer.vue` — 内联 map(:664/:676/:742/:760/:774)
- `D:/Haven/e-cam-web/src/views/compute/image/index.vue` — IMAGE_STATUS_LABELS(:314)/IMAGE_TYPE_LABELS(:327)/OS_TYPE_LABELS(:344)
- `D:/Haven/e-cam-web/src/views/compute/image/components/ImageDetailDrawer.vue` — statusLabels(:182)
- `D:/Haven/e-cam-web/src/views/compute/disk/index.vue` — getStatusText(:649)/getDiskCategory(:650)
- `D:/Haven/e-cam-web/src/views/compute/disk/components/DiskDetailDrawer.vue` — getStatusText(:116)/getDiskCategory(:117)

## Acceptance Criteria
- [ ] ecs/image/disk 状态 + 产品枚举(image type/OS type、disk 磁盘类型)收敛到 `src/utils/fieldLabels.ts` 命名导出
- [ ] 三产品 index 内联 map、抽屉手写 map、disk 的 index↔drawer 双份 `getDiskCategory`/`getStatusText` 全部改为 import 共享词表、删除本地定义
- [ ] 语义保留:ecs `stopped`→已关机(抽屉若写「已停止」须对齐为「已关机」)、`running`→运行中;disk `in_use`/`in-use`→使用中、`attached`→已挂载;磁盘类型高效云盘/SSD 云盘/ESSD 云盘/普通云盘不丢;image OS linux/windows 不丢
- [ ] 该域 grep 不再有本地手写 map 定义(仅余 shared import)
- [ ] `pnpm typecheck` 通过、既有单测不回归

## Hard Rules
仅修改以下文件:
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts`
- `D:/Haven/e-cam-web/src/views/compute/ecs/index.vue`
- `D:/Haven/e-cam-web/src/views/compute/ecs/components/InstanceDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/compute/image/index.vue`
- `D:/Haven/e-cam-web/src/views/compute/image/components/ImageDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/compute/disk/index.vue`
- `D:/Haven/e-cam-web/src/views/compute/disk/components/DiskDetailDrawer.vue`
不改字段键、不改后端、不改数据契约;不碰清单外文件。

## Implementation Notes
- 单一来源位置统一:`src/utils/fieldLabels.ts`;词表命名沿用现有 `IMAGE_STATUS_LABELS` 等,disk 磁盘类型可命名 `DISK_CATEGORY_LABELS`。
- ecs 内联 `displayMap`(如 `{ running: '运行中', stopped: '已关机' }`)与抽屉内联 map 一并收敛,列表页词汇 canonical、抽屉对齐(重点核对 ecs `stopped` 是否两处写法不一致)。
- 复杂度 medium 理由:6 文件、无并发 WIP、漂移集中在 ecs stopped 与 disk 磁盘类型双份。
- 提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径,禁 `git add -A`。