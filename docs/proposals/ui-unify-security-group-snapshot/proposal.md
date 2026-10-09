---
created: "2026-09-24"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: 云产品页对齐主机标准 — 第四批 SecurityGroup + Snapshot

## Problem

第三批（disk/nas/oss）收尾后，早年「薄骨架」云产品页还剩两张零散专项：`compute/security-group`（240 行）与 `compute/snapshot`（256 行）。两者与 disk/nas/oss 迁移前同型：`vm-page` + `page-top`（仅 1–2 个状态标）+ `action-bar` 把关键词搜索塞进操作栏 + 手写 `<table>` + 常驻 `col-checkbox` 多选列，无字段搜索栏、无 chips，观感薄于主机页。本批把这两张页对齐「主机页紧凑标准」，补全迁移程序覆盖面。

### Evidence

- 两页同为 `page-top + stats-badges + action-bar(内联关键词搜索) + table-wrapper(手写 table) + col-checkbox`，与 disk/nas/oss 迁移前签名一致
- security-group 仅 1 个状态标（总数）；snapshot 2 个（总数/正常）；均吸盘常驻多选复选框列
- 设计标准已锁定于 `ui-unify-rds-redis-pilot`（三页后续批沿用），本批无新增决策

### Urgency

早年薄骨架剩余最后两张列表页；迁移完即全站云产品列表页观感统一。

## Proposed Solution

按锁定的「主机页紧凑标准」重写两页（同 disk/nas/oss 批）：
- 页头：标题 + tab + 内联状态标（5 个，竖分隔，无大卡片）
- 操作栏：按产品领域给动作（安全组/快照各自派生，勿照搬主机电源操作）
- 搜索：字段搜索栏（回车加条件 + chips）
- 表格：无默认多选列；状态点 + 平台色标 + 等宽 ID；首/末列 20px 对齐
- 详情：行点击/详情开右侧 Drawer
- 主题：深/浅一致，Linear 靛紫令牌

## Requirements Analysis

同前序 proposal（令牌已就绪、不改令牌层、不引入大卡片/PageContainer、浏览器双主题验收）。

## Scope

### In Scope

- SecurityGroup 页重写对齐标准
- Snapshot 页重写对齐标准

### Out of Scope

- 令牌/主题/大卡片/PageContainer 改动
- network 系 PageContainer 页（eip/lb/vpc/vswitch/eni/cdn/waf）与 compute/image —— 另一代骨架，留待独立批次
- compute/provision（开通向导）与 compute/template（模板管理）—— 非列表页

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 领域动作/列梳理不准 | M | M | 以各页现有字段派生（security_group_type/vpc_id/rule_count；snapshot status/容量） |

## Success Criteria

- [ ] 两页为紧凑标准骨架（5 内联状态标、无大卡片、无默认多选列）
- [ ] 操作栏按各自领域给动作；字段搜索栏 + chips 可用
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题与标准一致、无控制台错误

consistency_check_result:
  status: pass
  pairs_checked: 4
  conflicts_found: 0

## Next Steps

- Proceed to `/quick-tasks` to generate tasks