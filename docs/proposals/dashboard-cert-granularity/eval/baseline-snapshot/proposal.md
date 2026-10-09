---
created: "2026-09-21"
author: "Haven"
status: Draft
intent: "enhancement"
---

# Proposal: 到期看板分级卡页内下钻（证书粒度主表，弃跳台账）

## Problem

到期看板分级卡的点击交互是"跳转台账按档过滤"，而非页面内呈现该档证书——根因是看板存在粒度矛盾：**分级卡计数是证书粒度，主表却是域名粒度**（同域多张证书仅展示 `notAfter` 最新者独占一行），导致过期/临期旧证被同名新证掩盖，卡内无法呈现证书明细，只能跳去证书粒度的台账看。

### Evidence

- `e-cam-web/src/views/cert/dashboard/index.vue:158-165`（commit 7cdd488 决策注释自陈）：*"证书粒度计数 → 台账 daysLeft 过滤可见具体证书——看板表格为域名粒度，过期/临期证书被同名新证掩盖无法页内表达"*——跳台账是当时对粒度矛盾的妥协，非设计目标。
- `OverviewCards.vue:13` 卡片静态文案"点击跳转台账按档过滤"，用户点击卡片的第一反应是**就地看到该档具体证书**（brainstorm Q1 实测确认），跳转造成页面断裂感。
- 用户此前实测反馈："到期看板有的点击按分级过滤为空"——即域名粒度下行内筛选与卡片计数对不上。
- 相邻特性 `ledger-hide-expired-orphans` 刚交付"台账默认隐藏过期且无引用证书"；看板计数不过滤将与其**口径分裂**（台账说不重要的证，看板全量展示）。

### Urgency

- 看板是**只读角色的默认页与唯一菜单入口**（`index.vue:75`），其粒度矛盾直接影响最大众的用户群。
- 域名粒度造成**真实风险盲区**：域名行显示健康（最新证有效），但同域已存在过期旧证（可能仍在被引用或裸奔），运营误判"无风险"。
- 若现在不做，立即与 ledger-hide 的隐藏语义产生两个页面计数对不上（用户刚验收的台账"已过期"数与看板卡片数不一致）——不一致窗口越久，用户信任损耗越大。

## Proposed Solution

将到期看板主表从**域名粒度重铸为证书粒度**，分级卡点击改为**页内筛选**（复用现有 filter 状态源与选中高亮交互），彻底移除跳台账；探测维度以**行级聚合徽标**附着在证书行（行点击仍进入现成的域名级探测详情抽屉）；"已过期且无引用"孤儿证书**与台账同构默认隐藏**，API 返回双口径计数。

### Innovation Highlights

- 直接采用该行业的基础范式：**仪表板主表对齐真实业务实体（证书）粒度**，计数与明细天然闭合。业界证书/资产管理系统（补天、NJR 类证书台账、云厂商 Certificate Manager 仪表盘）普遍以证书行为载体呈现到期，域名仅作为字段而非行主体。
- 探测状态**聚合为行级徽标**（借鉴 CI 状态徽标与监控 dashboard 的"最差优先"reducer 模式），解决"探测按域名进行、证书按张计数"的 1:N 附着问题。
- 与相邻特性共用**同一引用三态判定函数**，两页面隐藏口径单源不漂移——这是本次设计的工程核心（避免出现本 proposal 治理的双口径漂移）。

## Requirements Analysis

### Key Scenarios

- **快乐路径**：操作员打开看板，见"已过期 11"卡，点击 → 主表就地筛出 11 张过期证书行（含同名多证并立），从行上引用/托管/探测徽标即可判断处置优先级，行点击看探测详情 — 全程未离开看板。
- **边界：同名多证**：`*.jlc-robot.com` 过期旧证与有效新证同时在主表成两行（不再互相掩盖），daysLeft 独立分档。
- **边界：探测态冲突**：一张证书多个 SAN 探测结果不一（一 diff 一 consistent）→ 聚合取最差优先序 `diff > change_linked_diff > unreachable > wildcard_skipped > exempt > 健康/未探测`。
- **边界：孤儿**：过期且 `referenceStatus=no_refs_scanned` 默认隐藏，卡片与表均以"可见 N · 隐藏 M"双口径呈现，用户可切换查看全部。
- **错误路径**：探测数据缺失（行上探测徽标为空"未探测"）；看板 API 载荷更新期间仅前端消费，无第三方调用者。

### Non-Functional Requirements

- **性能**：主表行数由"去重域名数"增至"证书数"（同量级，当前 jlc 数据 94 证 / 93 域），行内 SAN 折叠展示；后端维持排序输出（到期优先）。
- **可访问性**：沿用现有 `aria-live` 筛选通告与骨架屏；卡片选中态保留 `aria-pressed`。
- **兼容性**：`GET /certs/dashboard` 载荷 items 结构为破坏性变更；已确认单消费端（e-cam-web dashboard 页），同步改造即可，无外部消费者。
- **只读边界**：看板保持无任何变更/扫描/发起更换入口（Hard Rule 不破）。

### Constraints & Dependencies

- 依赖 ledger-hide 的引用三态判定（`referenceStatus∈{has_refs,no_refs_scanned,blind_spot}`）与 `daysLeft` expired 豁免语义——**必须复用同一实现，不重写**。
- 复用：`probes.LatestPerDomain`（探测数据源）、探测详情抽屉、`WildcardProbeOverrides`（通配符跳过判定）、`countsByLevel` 现有证书粒度计数。
- 前端 filter 单一状态源机制沿用（分级选中=页内筛选，与特殊卡交互一致）。

## Alternatives & Industry Benchmarking

### Industry Solutions

证书到期仪表盘主流形态：证书行主表 + 分级计数卡 + 到期筛选（各类 Cert Manager 面板）；探测/合规状态以行级标签呈现（监控产品标准做法）。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing（维持跳台账） | 现状 7cdd488 | 零成本 | 粒度矛盾与风险盲区永存；与 ledger-hide 口径分裂 | Rejected: 正是本 proposal 要消除的 |
| 域名主表 + 就地证书抽屉（方案 B） | 画布备选 | 纯前端小改 | 不消除"同名新证掩盖"根因；卡计数与抽屉行数两来源 | Rejected: 修标不修本 |
| **证书粒度主表 + 页内筛选 + 探测徽标 + 孤儿同构隐藏（方案 A）** | 本 proposal（brainstorm Q2/Q3/Q4/Q5 确认） | 计数⇔表行数闭合、消除掩盖、与台账口径单源 | 后端+前端中等改造；聚合语义需明确定义 | **Selected**: 唯一根治"点击→就地看证书"的方案 |

## Feasibility Assessment

### Technical Feasibility

- 后端：`dashboard_service.go` 现有 `CountsByLevel` 循环即证书粒度；将 `owner` 域名收敛改为**证书循环直出**（每证一行），探测经 `probeByDomain[san]` 聚合、孤儿经 ledger 复用的三态判定过滤——改动集中、无新数据源。
- 前端：`DashboardTable` 行渲染改造 + `OverviewCards` 卡点击改页内 filter（`router.push` 移除）；可无障碍复用现有抽屉/徽标/通告组件。
- 无新依赖、无 showstopper。

### Resource & Timeline

团队已完成两轮看板工程（UI audit 35 任务、ledger-hide 双口径），后端 Go / 前端 Vue 技能齐备；估算约 1 个 feature 周期（2~3 任务：后端载荷+前端交互+联调验证）。

### Dependency Readiness

ledger-hide 已合入 main（三个 commit），三态判定函数可直接复用；探测管线稳定（本轮验证期间正常出探测记录）。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 点击分级卡"就应该跳台账" | 5 Whys | **Overturned**: 真实意图是"就地看到该档证书"（Q1 实测）；跳转只是粒度矛盾下的妥协产物 |
| 看板表格必须是域名粒度（才能表达探测） | Assumption Flip | **Refined**: 探测可聚合为证书行徽标，行点击下钻域名明细（Q3 确认）——探测洞察不丢 |
| 卡片计数与表行数"不可能一致，所以必须跳走" | Occam's Razor | **Overturned**: 证书粒度化后两者天然一致；根因就是粒度错配 |
| 台账隐藏孤儿、看板可全量显示 | Stress Test（一致性） | **Refined**: 会造成两页面计数口径分裂；判定同构复用（Q4 确认） |
| diff/exempt 卡语义必须筛域名行 | Assumption Flip | **Refined**: 转为"筛出含对应徽标的证书行"，语义等价且随主表粒度迁移 |

## Scope

### In Scope

- 后端 `GET /certs/dashboard`：items 改为证书行（每证一行，含孤儿）；行内聚合探测徽标判定（最差优先序）；孤儿隐藏双口径（visible + hiddenCount，复用 ledger 三态判定，过期豁免恒显示）；保持 summary/countsByLevel 语义不动（本就是证书粒度）。
- 前端 `DashboardTable`：证书行渲染（CN、SANS 折叠计数、issuer、daysLeft、托管、引用云、探测徽标）；孤儿默认隐藏 + 切换入口 + 空态处理；行点击 → 现成探测详情抽屉。
- `OverviewCards` 分级卡：点击改为页内筛选（选中高亮、再点取消），移除跳台账逻辑与"点击跳转台账"文案；卡计数⇔表行数动态相等；隐藏 N 张提示。
- 特殊卡 diff/exempt：语义迁移为筛选含对应探测徽标的证书行。
- 配套测试更新：`dashboard_service_test.go`（多证并立/聚合序/隐藏双口径）、前端 format/组件测试（粒度、卡交互、隐藏切换）。

### Out of Scope

- 台账 ledger 端隐藏逻辑本身（ledger-hide-expired-orphans 已交付，本 feature 只读复用，不改动）。
- 探测能力 / 巡检流程 / 通配符语义（仅呈现方式调整）。
- 3 张覆盖率卡与 2 张特殊卡的**计数语义**（diffAlertCount/exemptCount 保持域名级计数不动）。
- 看板只读边界（不增加任何变更/扫描/更换入口）。
- 变更单、发起更换等处置流程（仍由台账承载）。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 同名多证使主表行数增多、拖长页面 | M | L | 行内 SAN 折叠为数量徽标；后端按 notAfter asc 排序到期优先；行高密度化 |
| 多 SAN 探测态聚合歧义（用户对"这一行到底健康不健康"有疑问） | H | M | 显式最差优先序写入实现+单测；徽标带 tooltip 注明冲突 SAN 明细，抽屉保留域名级真相 |
| 既有 dashboard 测试断言（域名行语义）大面积失效 | M | L | 按新语义重建 `dashboard_service_test.go` + 前端 format 测试，随行语义一次性更新 |
| 看板与台账隐藏判定实现漂移（两处同时改判断逻辑） | M | H | 复用**同一个** ledger 三态判定 helper，不许看板另写判定；服务端单测锁定两处一致性 |
| `GET /certs/dashboard` 破坏性变更漏网调用者 | L | L | 确认单消费端（前端 dashboard 页）；grep 全仓 `getCertDashboardApi` 引用核验 |
| 空态/全隐藏边界（全部证书被隐藏时看板表空） | L | M | 与台账同构：空态提示 + "查看全部"入口保留，不得静默吞行 |

## Success Criteria

- [ ] `GET /certs/dashboard` 返回证书行：items 长度 == 全部证书数（含同域多证，不再域名去重）；单测断言同域新旧两证同时成行、独立分档
- [ ] 分级卡点击在**页内筛选**：选中卡 → 主表行数 == 该档卡片计数（动态相等）；再点取消；代码中无 `router.push('/certs'...)` 残留
- [ ] 孤儿同构隐藏：默认视图不含"过期且 no_refs_scanned"行；API 返回 hiddenCount，切换后行数 == 可见+隐藏；与台账 `daysLeft=expired` 经 `includeExpiredNoRefs` 的行数口径一致
- [ ] 探测聚合：构造多 SAN 冲突样例（diff+consistent / unreachable+diff 等）→ 聚合徽标按 `diff > change_linked_diff > unreachable > wildcard_skipped > exempt > 健康`；行点击进入域名级探测详情抽屉
- [ ] diff/exempt 卡点击 → 主表筛出含对应探测徽标的证书行，计数与卡片值一致
- [ ] 全部门禁绿：e-cam-service `go test ./...`（含新的 dashboard_service 单测）、e-cam-web `vitest` + `vue-tsc` 通过

## Next Steps

- 经 `/quick` 管线生成任务并执行（intent=enhancement，auto.runTasks.quick=true 自动串行）；或先进 `/write-prd` 正式化需求。