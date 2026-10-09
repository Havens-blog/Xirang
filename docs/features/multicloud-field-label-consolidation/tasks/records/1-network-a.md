---
status: "completed"
started: "2026-09-30 10:30"
completed: "2026-09-30 10:47"
time_spent: "~17m"
---

# Task Record: 1 网络域 A(eip/eni/lb/vpc)状态+枚举单源收口

## Summary
网络域 A(eip/eni/lb/vpc)状态+枚举单源收口：7 张词表(EIP_STATUS_LABELS/EIP_INSTANCE_TYPE_LABELS/ENI_STATUS_LABELS/ENI_TYPE_LABELS/LB_STATUS_LABELS/LB_TYPE_LABELS/VPC_STATUS_LABELS)收敛到 src/utils/fieldLabels.ts 命名导出；4 个 index.vue 删除本地 map 定义改走共享 import+labelOfLenient，4 个详情抽屉删除本地 statusLabels 改传共享词表(statusTones 色调按 Implementation Notes 保留在抽屉)；index 词表为 canonical，抽屉对齐补齐漂移键(LB 抽屉补 Inactive、VPC 抽屉补 OK/ok/创建中、EIP 抽屉实例类型补 ClbInstance/AlbInstance/Nat/EVPN)。不改字段键/后端/数据契约。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/utils/fieldLabels.ts
- e-cam-web/src/views/network/eip/index.vue
- e-cam-web/src/views/network/eip/components/EipDetailDrawer.vue
- e-cam-web/src/views/network/eni/index.vue
- e-cam-web/src/views/network/eni/components/EniDetailDrawer.vue
- e-cam-web/src/views/network/lb/index.vue
- e-cam-web/src/views/network/lb/components/LbDetailDrawer.vue
- e-cam-web/src/views/network/vpc/index.vue
- e-cam-web/src/views/network/vpc/components/VpcDetailDrawer.vue

### Key Decisions
- LB/VPC 抽屉与 index 词表存在键集漂移，按 proposal 裁定规则取列表页(canonical)超集，抽屉对齐
- EIP 实例类型 map index(9 键)与抽屉(5 键)不等，取 index 超集为 EIP_INSTANCE_TYPE_LABELS
- 抽屉 statusTones(色调)保留在抽屉本地，仅文案词表收敛(Implementation Notes 允许)
- LbDetailDrawer/VpcDetailDrawer 用 const statusLabels = XXX_STATUS_LABELS 别名保持模板零改动，最小化 diff
- 并发 WIP 文件 VpcDetailDrawer.vue：编辑前重读最新版(并发改动仅 N2-008 租户行 +2/-2，不重叠)，仅收敛 statusLabels 一处
- 仓库 lint 脚本带 --fix 会波及并发脏文件，改为对 9 个改动文件跑 eslint(无 --fix)，全部通过

## Test Results
- **Tests Executed**: Yes
- **Passed**: 919
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 四产品状态词表+产品枚举收敛到 fieldLabels.ts 命名导出(沿用 XXX_STATUS_LABELS 命名)
- [x] 4 个 index.vue 与 4 个抽屉删除本地 map 定义、改 import 共享词表，index 查表走 labelOf/labelOfLenient
- [x] 语义保留:EIP InUse→已绑定/Available→未绑定/Bindable→可绑定;LB slb/alb/nlb/clb;ENI 类型不丢;列表页 canonical、抽屉对齐
- [x] 该域 grep 无本地手写 statusLabels/XXXX_STATUS_LABELS 定义(仅余 shared import)
- [x] VpcDetailDrawer.vue 并发 WIP:仅收敛本地 statusLabels 一处,未覆盖并发会话其他改动
- [x] pnpm typecheck 通过、既有单测不回归(vitest 52 文件 919/919 通过)

## Notes
coverage 为 null：本仓 vitest 未配 coverage provider(coverage-v8 装不上，历史既定不臆造)，重构任务按 IMPORTANT 仅为信息项不做门槛。EIP index.vue 内 getExportValue/状态族判定(isBoundStatus 等)及 eni/lb/vpc 状态族判定数组保留原样(色调口径，非文案词表，不在收敛范围)。
