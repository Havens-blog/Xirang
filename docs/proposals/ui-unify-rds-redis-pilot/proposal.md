---
created: "2026-09-22"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: 云产品页对齐主机页标准 — 首批 RDS + Redis

## Problem

除主机(ECS)页与资产列表(assets)页外，其余云产品页（rds/redis/mongodb/elasticsearch/kafka/disk/nas/oss）共享一套早年的 `storage.scss`，观感与交互显著「薄」于主机页：只有 3 个状态标（总数/运行中/已停止）、关键词搜索塞在操作栏、无字段搜索栏、无标签筛选、表格常驻一列多选框。主机页是唯一拥有独立 `instances.scss` + 富功能（5 状态标、字段搜索栏、标签筛选、密集表格）的页——它已是用户认可的标准，但其余页没跟上；且主机页本身「常驻多选列」也需一并收掉。

### Evidence

- 代码实测：rds/redis/mongodb/elasticsearch/kafka/disk/nas 均 `@import '@/views/storage/styles/storage.scss'`；仅 `compute/ecs` 有独立 `styles/instances.scss`
- 状态标：主机 5 个（总数/运行中/关机/监控异常/异常）vs 其他页 3 个
- 搜索：主机页独立字段搜索栏（回车加条件 + 条件 chips）+ 标签 popover；其他页仅操作栏内关键词 input + 云平台/状态 select
- 表格：主机页与其余页均常驻 `col-checkbox` 多选列（用户反馈需去掉，见 Assumptions Challenged）
- 原型已由用户三轮评审定稿（`F:\front_design\prototype-host-disk.html`，深/浅主题）

### Urgency

「全站统一现代观感」的真正落点是这批云产品页；主机页是现成标杆，样板成本低（2 页试点即可跑通模式），宜尽快闭环再逐批铺开。

## Proposed Solution

把「主机页紧凑标准」固化为统一目标，首批在 **RDS + Redis**（同属数据库域、结构最像、量最小）落地，主机页仅做「去常驻多选列」一处微调。设计标准（已与用户三轮定稿，锁死不再变）：

| 层 | 规则 |
|---|---|
| 页头 | 标题 + tab + **内联状态标**（竖分隔，无大卡片） |
| 操作栏 | **按产品领域**给动作（RDS=新建/重启/暂停/变更配置/扩容；Redis=新建/重启/密码重置/参数配置 等，从各页实际派生，勿照搬主机电源操作） |
| 搜索 | 字段搜索栏（回车加条件 + 条件 chips） |
| 表格 | **无默认多选列**；状态点 + 平台色标 + 等宽 ID/IP；列可自定义 |
| 详情 | 行点击开右侧 Drawer |
| 主题 | 深/浅一致，走已落地 Linear 靛紫令牌（Phase 1 已交付） |

### Innovation Highlights

无新概念——把主机页已验证的既有标准**忠实铺开**，配一个可视化验收：每页重写任务内嵌 chrome-devtools-mcp 浏览器实测（真实登录态、深/浅双主题截图归档、关键元素计算样式断言），延续 Phase 1/3 抓真 bug 的惯例。

## Requirements Analysis

### Key Scenarios

- 正常路径：RDS/Redis 列表页渲染为紧凑标准（内联状态标、字段搜索、密集表格），动作按各自领域
- 边界：深/浅两主题一致；无默认多选列（行点击开抽屉而非多选）
- 异常：改错 scss → `pnpm build` 暴露；改错值能编译 → 浏览器计算样式断言兜底

### Non-Functional Requirements

- 一致性：两页与主机页同骨架、同令牌、同交互节奏；其余 NFR 不涉及

### Constraints & Dependencies

- 依赖：Phase 1 令牌层（已合并 main）、主机页 `compute/ecs` 现有实现（参考标准）
- 约束：不改令牌层；不引入大统计卡/PageContainer 大改；每页动作/列/状态标从其实际数据派生，不得照搬主机

## Alternatives & Industry Benchmarking

### Industry Solutions

多云管理控制台（阿里云/腾讯云/火山引擎）对同类「资源列表页」均采用紧凑列表 + 领域化动作 + 字段搜索，本方案即此标准做法。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零成本 | 其他云产品页持续「薄」于主机页，观感不统一 | Rejected：样板成本低收益确定 |
| 全盘照搬主机页（含电源操作、多选列） | — | 最快 | 动作语义错误（RDS 无「开机」）、多选列被用户否决 | Rejected：用户明确反对 |
| **主机紧凑骨架 + 领域化动作 + 去多选列** | 用户三轮定稿 | 统一 + 语义正确 + 更干净 | 每页动作需单独梳理 | **Selected：用户已确认** |

## Feasibility Assessment

### Technical Feasibility

纯 Vue SFC 重写 + 复用现有令牌；无技术障碍。RDS/Redis 现有实现（356/253 行）结构可参照主机页（2019 行）裁剪。

### Resource & Timeline

2 页 + 1 处微调，单批可完成；后续批（mongo/es/kafka/disk/nas/oss）复用同一模式。

### Dependency Readiness

主机页可作参考（已线上）；令牌已就绪。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 用大统计卡体现「现代化」 | 用户反馈 | Overturned：统计卡太占页面，主机页内联状态标才是标准 |
| 动作/多选可照搬主机页 | 用户反馈 | Overturned：各产品侧重点不同，动作须领域化；常驻多选列要去掉（主机页一并） |
| 主机页要重写 | 用户反馈 | Overturned：主机页已是标准，仅去多选列微调，不重写 |
| 先做 compute 主机+磁盘 | 用户反馈 | Refined：主机不重写，首批改 RDS + Redis |

## Scope

### In Scope

- RDS 页重写对齐主机紧凑标准（领域动作/列/状态标，含字段搜索栏、去多选列）
- Redis 页重写对齐主机紧凑标准（同上）
- 主机页微调：去常驻多选列
- 每页内嵌浏览器双主题视觉验收

### Out of Scope

- 其余云产品页（mongodb/elasticsearch/kafka/disk/nas/oss）——后续批
- 大统计卡、PageContainer 大改、令牌/主题改动
- 移动端响应式

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Redis/RDS 动作语义梳理不准 | M | M | 实现时以各页现有字段/操作派生，原型 RDS 为准，Redis 对照其现有能力 |
| 去多选列连带去掉批量操作能力 | M | L | 批量操作改为「勾选模式开关」或行内操作，实现时确认 |
| 每页域名列不一致导致观感再漂移 | M | M | 字段搜索/状态标/表格样式统一走共享样式，列仅内容不同 |

## Success Criteria

- [ ] RDS、Redis 两页均为紧凑标准骨架（无大卡片、内联状态标）
- [ ] 操作栏按各自领域给出动作，无「开机/重启/暂停」之类照搬
- [ ] 表格无默认多选列；主机页多选列已移除
- [ ] 字段搜索栏 + 条件 chips 可用
- [ ] `pnpm build` / `pnpm typecheck` 通过；`pnpm lint` 相对基线零新增
- [ ] 浏览器深/浅两主题：两页渲染与原型一致，无控制台错误

consistency_check_result:
  status: pass
  pairs_checked: 6
  conflicts_found: 0

## Next Steps

- Proceed to `/quick-tasks` to generate tasks