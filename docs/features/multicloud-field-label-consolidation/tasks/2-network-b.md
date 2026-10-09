---
id: "2"
title: "网络域 B(vswitch/waf/dns)状态+版本+枚举单源收口"
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

# 2: 网络域 B(vswitch/waf/dns)状态+版本+枚举单源收口

## Description
**代码仓：D:/Haven/e-cam-web**。收敛网络域 vswitch/waf/dns 三个产品的「状态 + 版本(edition)+ 防护模式 + 产品枚举」词表。重点是 WAF 的 `editionMap` 在 index 与抽屉**两份且键集不等**(index 含 sparta-waf/EdgeOne,抽屉仅 basic/pro/business/enterprise)、WAF 状态 index 用本地 `WAF_STATUS_LABELS`(active→正常)而抽屉用 `ASSET_STATUS_LABELS`(active→运行中)已漂移;dns 的运营商 map 在 RecordTable 与 RecordFormDialog 重复。统一为单一共享来源,列表页词汇 canonical、抽屉对齐。

## Reference Files
- `docs/proposals/multicloud-field-label-consolidation/proposal.md` — Proposed Solution, Scope(In Scope), Key Risks, Success Criteria
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts` — 单一来源模块
- `D:/Haven/e-cam-web/src/views/network/vswitch/index.vue` — VSWITCH_STATUS_LABELS(约 :303)
- `D:/Haven/e-cam-web/src/views/network/vswitch/components/VSwitchDetailDrawer.vue` — statusLabels(:156)/statusTones(:161)
- `D:/Haven/e-cam-web/src/views/network/waf/index.vue` — WAF_STATUS_LABELS(:355)、editionMap(:362)、protectionMode map(:378)
- `D:/Haven/e-cam-web/src/views/network/waf/components/WafDetailDrawer.vue` — getEditionLabel/getEditionTagType map(:356~:363)、charge map(:370);status 已用 ASSET_STATUS_LABELS(:305)
- `D:/Haven/e-cam-web/src/views/network/dns/components/DomainTable.vue` — 域名状态 map(:136)
- `D:/Haven/e-cam-web/src/views/network/dns/components/RecordTable.vue` — 运营商 map(:165)、接入源 map(:170)
- `D:/Haven/e-cam-web/src/views/network/dns/components/RecordFormDialog.vue` — 运营商 map(:81)

## Acceptance Criteria
- [ ] vswitch/waf/dns 三产品词表(vswitch 状态、waf 状态+edition+protectionMode、dns 域名状态+运营商+接入源)收敛到 `src/utils/fieldLabels.ts` 命名导出
- [ ] WAF `editionMap` 定义收敛到单一文件,`waf/index.vue` 与 `WafDetailDrawer.vue` 仅 import;键集以列表页为准(含 `sparta-waf`/`EdgeOne`)
- [ ] WAF 状态以列表页 canonical(`active`→正常、`suspended`→已暂停、`pending`→待接入),抽屉弃用 `ASSET_STATUS_LABELS` 改 import 共享 WAF 词表,消除 index↔drawer 漂移
- [ ] 语义保留:dns `paused`→暂停、运营商电信/联通/移动、接入源 cdn/waf/slb/ecs/eip、waf edition basic→基础版/pro→专业版/business→商业版/enterprise→企业版、protectionMode block→拦截/observe→观察/off→关闭
- [ ] 该域 grep 不再有本地手写 map 定义(仅余 shared import)
- [ ] `pnpm typecheck` 通过、既有单测不回归

## Hard Rules
仅修改以下文件:
- `D:/Haven/e-cam-web/src/utils/fieldLabels.ts`
- `D:/Haven/e-cam-web/src/views/network/vswitch/index.vue`
- `D:/Haven/e-cam-web/src/views/network/vswitch/components/VSwitchDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/network/waf/index.vue`
- `D:/Haven/e-cam-web/src/views/network/waf/components/WafDetailDrawer.vue`
- `D:/Haven/e-cam-web/src/views/network/dns/components/DomainTable.vue`
- `D:/Haven/e-cam-web/src/views/network/dns/components/RecordTable.vue`
- `D:/Haven/e-cam-web/src/views/network/dns/components/RecordFormDialog.vue`
不改字段键、不改后端、不改数据契约(纯展示层);不碰清单外文件(尤其并发会话脏文件 `src/views/service-tree/rules/index.vue`、`src/views/network/security-group/...`、`src/views/accounts/...`)。

## Implementation Notes
- 单一来源位置统一:`src/utils/fieldLabels.ts`。词表命名沿用 `XXX_STATUS_LABELS`/`WAF_EDITION_LABELS`/`DNSDomainStatus` 等 UPPER_SNAKE/Camel 对齐现有风格。
- WAF edition 标签色(`getEditionTagType` map basic→info/pro→primary/business→warning/enterprise→danger)为颜色映射,可随词表一起收敛或留在抽屉,但文案键值不得再复制。
- dns 无 index+drawer 对,收敛目标是「RecordTable 与 RecordFormDialog 的运营商 map 去重共享」+「DomainTable 域名状态单源」。
- 复杂度 high 理由:WAF index↔drawer 语义冲突(正常 vs 运行中)+ edition 键集不等 + dns 跨文件重复三处治理点。
- 提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径,禁 `git add -A`。