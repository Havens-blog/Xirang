---
created: "2026-09-23"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: 云产品页对齐主机标准 — 第三批 Disk + NAS + OSS

## Problem

第三批延续已锁定的「主机页紧凑标准」：disk（598 行）/ nas（544 行）/ oss（532 行）三页当前为早年薄骨架（3 状态标、关键词搜索塞操作栏、无字段搜索、常驻多选列），且带若干运营指标卡，观感仍薄于主机页。本批是迁移程序收尾三页（此后仅剩零散专项）。

### Evidence

- 实测三页同为 `page-top + stats-badges + action-bar + table-wrapper`，均无字段搜索栏，`col-checkbox` 存在（disk=3、nas/oss=2）
- 现状另有运营指标卡（磁盘数/平均使用率/IO 繁忙），数据源为 `ecam_disk_metric` 等指标表
- 设计标准已锁定于 `ui-unify-rds-redis-pilot`，本批唯一新增决策：**指标卡保留为紧凑指标条**（用户已确认）

### Urgency

迁移程序最后一批，闭环后全站云产品页观感统一。

## Proposed Solution

按锁定的「主机页紧凑标准」重写三页，并在其上叠加一条：

- 页头：标题 + tab + 内联状态标（5 个，竖分隔，无大卡片）
- 操作栏：按产品领域给动作（Disk/NAS/OSS 各自派生，勿照搬主机电源操作）
- **指标条**：保留运营指标卡（磁盘数/平均使用率/IO），改为与标准同源的紧凑横条（同令牌、同间距、保留 warn/empty 态）
- 搜索：字段搜索栏（回车加条件 + chips）
- 表格：无默认多选列；状态点 + 平台色标 + 等宽 ID；首/末列 20px 对齐
- 详情：行点击/详情开右侧 Drawer
- 主题：深/浅一致，Linear 靛紫令牌

## Requirements Analysis

同前序 proposal（令牌已就绪、不改令牌层、不引入大卡片/PageContainer、浏览器双主题验收）；指标条为紧凑横条而非整卡，避免重蹈「统计卡太占页面」。

## Alternatives & Industry Benchmarking

纯展开，沿用前序结论。

## Scope

### In Scope

- Disk 页重写对齐标准（保留紧凑指标条）
- NAS 页重写对齐标准（保留紧凑指标条）
- OSS 页重写对齐标准（保留紧凑指标条）

### Out of Scope

- 令牌/主题/大卡片/PageContainer 改动

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 指标条过宽仍显「占页」 | M | M | 三卡并排、紧凑 padding（≤16px 纵向），与页头 20px 同节奏 |
| 各页领域动作/列梳理不准 | M | M | 以各页现有字段派生 |

## Success Criteria

- [ ] 三页为紧凑标准骨架（5 内联状态标、无大卡片、无默认多选列）
- [ ] 运营指标卡保留为紧凑指标条（warn/empty 态不丢）
- [ ] 操作栏按各自领域给动作；字段搜索栏 + chips 可用
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题与标准一致、无控制台错误

consistency_check_result:
  status: pass
  pairs_checked: 5
  conflicts_found: 0

## Next Steps

- Proceed to `/quick-tasks` to generate tasks