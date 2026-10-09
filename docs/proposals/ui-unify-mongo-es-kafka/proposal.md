---
created: "2026-09-23"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: 云产品页对齐主机标准 — 第二批 MongoDB + Elasticsearch + Kafka

## Problem

第二批延续 `ui-unify-rds-redis-pilot` 已锁定的「主机页紧凑标准」：mongodb（252 行）/ elasticsearch（221 行）/ kafka（224 行）三页仍共享早年 `storage.scss` 版薄骨架——3 状态标、关键词搜索塞操作栏、无字段搜索栏、常驻多选列，观感与交互薄于主机页。

### Evidence

- 实测三页同为 `page-top + stats-badges + action-bar + table-wrapper + col-checkbox`（6 个结构标记），**均无字段搜索栏**（grep `search-bar/searchFields` = 0）
- 行数 252/221/224，量级与 RDS(356→775)/Redis(253→860) 同源
- 设计标准已在前序 proposal `ui-unify-rds-redis-pilot` 与用户三轮原型评审中锁定，无新增决策

### Urgency

同一迁移程序的第二块，样板成本低，逐批闭环。

## Proposed Solution

按已锁定的「主机页紧凑标准」（见 `ui-unify-rds-redis-pilot/proposal.md` 的 Proposed Solution 设计标准表，锁死不变）重写三页：

| 层 | 规则 |
|---|---|
| 页头 | 标题 + tab + 内联状态标（竖分隔，无大卡片） |
| 操作栏 | 按产品领域给动作（MongoDB/Kafka/ES 各自派生，勿照搬主机电源操作） |
| 搜索 | 字段搜索栏（回车加条件 + chips） |
| 表格 | 无默认多选列；状态点 + 平台色标 + 等宽 ID；列可自定义 |
| 详情 | 行点击/详情开右侧 Drawer |
| 主题 | 深/浅一致，走 Linear 靛紫令牌 |

## Requirements Analysis

### Key Scenarios / NFR / Constraints

同前序 proposal（令牌已就绪、不改令牌层、不引入大卡片/PageContainer、浏览器双主题验收）。各页领域动作与列从本页实际字段派生。

## Alternatives & Industry Benchmarking

沿用前序 proposal 结论；本轮为纯展开，无新增权衡。

## Scope

### In Scope

- MongoDB 页重写对齐标准
- Elasticsearch 页重写对齐标准
- Kafka 页重写对齐标准

### Out of Scope

- disk / nas / oss（第三批）
- 大卡片 / PageContainer / 令牌改动

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 各页领域动作梳理不准 | M | M | 以各页现有字段派生；Kafka=Topic/分区，ES=索引/节点 |
| 去多选列连带清批量逻辑 | M | L | 与 RDS 先例一致，勾选逻辑一并清 |

## Success Criteria

- [ ] 三页均为紧凑标准骨架（内联状态标、无大卡片、无默认多选列）
- [ ] 操作栏按各自领域给动作
- [ ] 字段搜索栏 + chips 可用
- [ ] `pnpm build` / `pnpm typecheck` 通过；`pnpm lint` 相对基线零新增
- [ ] 浏览器深/浅两主题与标准一致、无控制台错误

consistency_check_result:
  status: pass
  pairs_checked: 5
  conflicts_found: 0

## Next Steps

- Proceed to `/quick-tasks` to generate tasks