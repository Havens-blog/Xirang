---
id: "1"
title: "网络域 A(eip/eni/lb/vpc)状态+枚举单源收口"
priority: "P1"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: 网络域 A(eip/eni/lb/vpc)状态+枚举单源收口

## Description
**代码仓：D:/Haven/e-cam-web**。把网络域 eip/eni/lb/vpc 四个产品的「状态」词表与产品专属枚举(index 本地 `XXX_STATUS_LABELS`/内联 map、抽屉手写 `statusLabels` 各一份)收敛为单一共享定义,index.vue 与详情抽屉 import 同一份,消除 index↔drawer 漂移与重复。纯前端展示层,不改字段键/后端/数据契约。

## Reference Files
- `docs/proposals/multicloud-field-label-consolidation/proposal.md` — Proposed Solution, Scope(In Scope), Key Risks, Success Criteria
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts` — 单一来源模块(已有 CHARGE_TYPE_LABELS/ASSET_STATUS_LABELS/labelOf/labelOfLenient,词表放这里)
- `D:/Haven/e-cam-web/src/views/network/eip/index.vue` — EIP_STATUS_LABELS(约 :305)
- `D:/Haven/e-cam-web/src/views/network/eip/components/EipDetailDrawer.vue` — statusLabels/statusTones(约 :212/:217)、getInstanceTypeLabel(:251)
- `D:/Haven/e-cam-web/src/views/network/eni/index.vue` — ENI_STATUS_LABELS/ENI_TYPE_LABELS(约 :296/:313)
- `D:/Haven/e-cam-web/src/views/network/eni/components/EniDetailDrawer.vue` — statusLabels(:178)/statusTones(:190)
- `D:/Haven/e-cam-web/src/views/network/lb/index.vue` — LB_STATUS_LABELS/LB_TYPE_LABELS(约 :307/:336)
- `D:/Haven/e-cam-web/src/views/network/lb/components/LbDetailDrawer.vue` — statusLabels(:321)
- `D:/Haven/e-cam-web/src/views/network/vpc/index.vue` — VPC_STATUS_LABELS(约 :297)
- `D:/Haven/e-cam-web/src/views/network/vpc/components/VpcDetailDrawer.vue` — statusLabels(:312)

## Acceptance Criteria
- [ ] eip/eni/lb/vpc 四产品状态词表 + 产品枚举(EIP 实例类型、ENI type、LB type 等已存在的本地 map)收敛到 `src/utils/fieldLabels.ts` 命名导出(沿用 `XXX_STATUS_LABELS` 命名)
- [ ] 该域 4 个 index.vue 与 4 个抽屉均删除本地 map 定义、改为 `import { ... } from '@/utils/fieldLabels'`,查表用 `labelOf`/`labelOfLenient`
- [ ] 语义保留:EIP `InUse`→已绑定、`Available`→未绑定、`Bindable`→可绑定;LB 类型 slb→SLB/alb→ALB/nlb→NLB/clb→CLB;ENI 类型不丢;列表页词汇为 canonical、抽屉对齐
- [ ] 该域 grep 不再有本地手写 `statusLabels`/`XXXX_STATUS_LABELS` 定义(仅余 shared import)
- [ ] `vpc/components/VpcDetailDrawer.vue` 是并发会话 WIP:动它前先读最新版,仅收敛其本地 statusLabels 一处,不得覆盖并发会话的其他改动
- [ ] `pnpm typecheck` 通过、既有单测不回归

## Hard Rules
仅修改以下文件:
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts`
- `D:/Haven/e-cam-web/src/views/network/eip/index.vue`
- `D:/Haven/e-cam-web/src/views/network/eip/components/EipDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/network/eni/index.vue`
- `D:/Haven/e-cam-web/src/views/network/eni/components/EniDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/network/lb/index.vue`
- `D:/Haven/e-cam-web/src/views/network/lb/components/LbDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/network/vpc/index.vue`
- `D:/Haven/e-cam-web/src/views/network/vpc/components/VpcDetailDrawer.vue`
不改字段键、不改后端、不改数据契约(纯展示层);不碰上述清单外的任何文件(尤其并发会话脏文件 `src/api/request/index.ts`、`src/layouts/MainLayout.vue`、`src/stores/user.ts`、`src/views/service-tree/rules/index.vue`、`src/views/network/security-group/...`)。

## Implementation Notes
- 单一来源位置统一:`src/utils/fieldLabels.ts`(与 CHARGE_TYPE_LABELS/ASSET_STATUS_LABELS 同模块),词表命名沿用现有 `XXX_STATUS_LABELS`/`XXX_TYPE_LABELS`。
- 抽屉本地 `statusLabels`/`statusTones`(文案+色调用)收敛:文案走共享词表;色调用(`statusTones`)保留在抽屉或随词表一并导出,但文案键值不得再复制一份。
- 复杂度 high 理由:8 文件 + 并发 WIP 文件(vpc 抽屉)+ EIP 实例类型/ENI/LB 类型多枚举 + index↔drawer 文案对齐风险。
- 提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径,禁 `git add -A`/`git add .`。