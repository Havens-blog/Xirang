---
status: "completed"
started: "2026-09-30 10:51"
completed: "2026-09-30 11:21"
time_spent: "~30m"
---

# Task Record: 2 网络域 B(vswitch/waf/dns)状态+版本+枚举单源收口

## Summary
网络域 B(vswitch/waf/dns) 状态+版本+枚举单源收口：fieldLabels.ts 新增 7 张命名导出词表（VSWITCH_STATUS_LABELS / WAF_STATUS_LABELS / WAF_EDITION_LABELS / WAF_PROTECTION_MODE_LABELS / DNS_DOMAIN_STATUS_LABELS / DNS_LINE_LABELS / DNS_RECORD_SOURCE_LABELS），vswitch/waf/dns 共 8 个文件全部改为 import 单源，消除 WAF index↔drawer 状态漂移与 edition 键集不等。纯展示层重构，不改字段键与数据契约。

## Changes

### Files Created
无

### Files Modified
- src/utils/fieldLabels.ts
- src/views/network/vswitch/index.vue
- src/views/network/vswitch/components/VSwitchDetailDrawer.vue
- src/views/network/waf/index.vue
- src/views/network/waf/components/WafDetailDrawer.vue
- src/views/network/dns/components/DomainTable.vue
- src/views/network/dns/components/RecordTable.vue
- src/views/network/dns/components/RecordFormDialog.vue

### Key Decisions
- WAF edition 键集以列表页为准（含 sparta-waf/EdgeOne，实测值域 EdgeOne(13)/sparta-waf(100)），抽屉 4 键旧表弃用
- WAF 抽屉 statusLabels 组合为 { ...ASSET_STATUS_LABELS, ...WAF_STATUS_LABELS }：WAF 词表 canonical 胜出（active→正常），ASSET 宽词表保留 paused/bypass/expiring 等 WAF 域原生态兜底，行为对齐 AC 且不丢抽屉宽词表能力
- RecordFormDialog 的运营商词表实为模板 el-option 列表（Reference File 行号 :81 实为记录值 placeholder map，grep 核实后按 SPEC 意图处理）：改为 v-for 遍历共享 DNS_LINE_LABELS，选项集与顺序不变
- 颜色映射按 Implementation Notes 留在本地：statusTones(两抽屉)、getProtectionModeType(waf index)、getEditionTagType(waf drawer)；WAF 抽屉 charge map 为 proposal Out of Scope 不动
- 查表函数切换 labelOfLenient 时逐 case 核对与原 || 回退链等价（空值/未知值/命中三分支）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 919
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] vswitch/waf/dns 三产品词表收敛到 fieldLabels.ts 命名导出
- [x] WAF editionMap 收敛单一定义，index 与抽屉仅 import，键集含 sparta-waf/EdgeOne
- [x] WAF 状态以列表页 canonical，抽屉弃用裸 ASSET_STATUS_LABELS 改共享 WAF 词表，消除漂移
- [x] 语义保留：dns paused→暂停、运营商电信/联通/移动、接入源 cdn/waf/slb/ecs/eip、waf edition/protectionMode 各键文案不变
- [x] 该域 grep 无本地手写词表 map 定义（仅余 searchFieldLabels/placeholder 提示等单用非词表 map 与颜色映射）
- [x] pnpm typecheck(vue-tsc -b) 通过、既有单测不回归

## Notes
coverage 为 null：本仓 vitest 未配 coverage provider（与任务 1 同因，不臆造），重构任务按 IMPORTANT 仅为信息项不做门槛。验证证据：npx vue-tsc -b exit 0；npx eslint 8 个改动文件 exit 0；npx vitest run 52 files / 919 tests 全过。Pre-check git 非净但脏文件均为并发会话 WIP（service-tree/logs/security-group），与本次 8 文件清单零交集，Hard Rules 已明示该场景；未触碰任何清单外文件。git diff --stat 确认改动恰为 8 文件。
