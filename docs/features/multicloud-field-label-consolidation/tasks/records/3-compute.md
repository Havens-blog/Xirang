---
status: "completed"
started: "2026-09-30 11:24"
completed: "2026-09-30 11:37"
time_spent: "~13m"
---

# Task Record: 3 计算域(ecs/image/disk)状态+枚举单源收口

## Summary
计算域 ecs/image/disk 状态+产品枚举词表收敛到 src/utils/fieldLabels.ts 单源：新增 ECS_STATUS_LABELS/IMAGE_STATUS_LABELS/IMAGE_TYPE_LABELS/OS_TYPE_LABELS/DISK_STATUS_LABELS/DISK_CATEGORY_LABELS 六个命名导出；ecs index（statBadge displayMap/getStatusText/导出 getValue）、ecs InstanceDetailDrawer（getStatusText/getDiskStatusText）、image index（三个本地词表）、image ImageDetailDrawer（statusLabels/getImageTypeLabel）、disk index 与 DiskDetailDrawer（getStatusText/getDiskCategory 双份重复）全部改为 import 共享词表并删除本地定义。语义保留：ecs stopped→已关机、running→运行中；disk in_use/in-use→使用中、attached→已挂载；磁盘类型四键与 image OS linux/windows 不丢。按「列表页 canonical」裁决修正两处抽屉漂移：ecs 抽屉盘状态 attached 使用中→已挂载、image 抽屉镜像类型 公共镜像等→公共等（与列表同词）。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/utils/fieldLabels.ts
- e-cam-web/src/views/compute/ecs/index.vue
- e-cam-web/src/views/compute/ecs/components/InstanceDetailDrawer.vue
- e-cam-web/src/views/compute/image/index.vue
- e-cam-web/src/views/compute/image/components/ImageDetailDrawer.vue
- e-cam-web/src/views/compute/disk/index.vue
- e-cam-web/src/views/compute/disk/components/DiskDetailDrawer.vue

### Key Decisions
- 词表键集保留各消费方原口径并集（ECS 含大小写变体供 toUpperCase/原始键直查两种用法；DISK_STATUS 供 toLowerCase 归一查表），不改字段键、不改数据契约
- 颜色映射 getStatusClass、searchFieldLabels、getDiskTypeText(system/data 单文件单用无漂移)、getSnapshotStatusText(快照产品属任务4域)均不动，符合 proposal Out of Scope
- image 抽屉 AssetStatusBadge :labels 直接绑 IMAGE_STATUS_LABELS（键集为原 statusLabels 超集，无回归路径）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 27
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] ecs/image/disk 状态+产品枚举收敛到 fieldLabels.ts 命名导出
- [x] 三产品 index/drawer 本地 map 全部改为 import 共享词表并删除本地定义
- [x] 语义保留：ecs stopped→已关机/running→运行中、disk in_use→使用中/attached→已挂载、磁盘类型四键、image OS linux/windows 不丢
- [x] 该域 grep 不再有本地手写状态/枚举 map 定义（仅余共享 import 与范围外 map：颜色/搜索字段名/快照域）
- [x] pnpm typecheck 通过、既有单测不回归（diskMetrics.test.ts 27/27）

## Notes
typecheck (vue-tsc -b) 零错误；eslint 对 7 个改动文件 0 error（1 个 pre-existing warning：InstanceDetailDrawer.vue:548 unused emit，非本次改动区域）。coverage 无 runner 输出（套件未配覆盖率工具），未填报。工作区存在并行会话脏文件（logs/service-tree/security-group），未触碰，提交须显式 pathspec。
