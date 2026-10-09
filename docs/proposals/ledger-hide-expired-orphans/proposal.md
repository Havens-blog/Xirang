---
created: "2026-09-21"
author: "Haven"
status: Approved
intent: "enhancement"
---

# Proposal: 台账默认隐藏无引用过期证书（可切换查看）

## Problem

证书台账默认平铺展示全部证书，其中「已过期且经成功引用扫描确认零引用（referenceStatus=no_refs_scanned）」的证书是纯噪音——不可操作、无风险信号，占据列表首屏位置。

### Evidence

- 当前台账 93 张证书，11 张已过期：其中 **1 张 refCount=0**（`*.jlc-robot.com`，-11 天，fingerprint_only，非保护期，可删除）；**其余 10 张 refCount>0**（如 `*.jlcpcb.com` 引用 21、`*.jlccam.com` 引用 11）——「仍被线上引用 = 线上在用过期证书」，是探测/差异告警要追踪的风险信号，**不能隐藏**。注意：裸 refCount 数字无法区分「明确零引用」与「扫描未覆盖」（refCount=0 可能同时来自 no_refs_scanned 与 blind_spot），因此判定一律以引用三态 `referenceStatus`（has_refs / no_refs_scanned / blind_spot）为准。
- 变更换证后旧证书引用清零（refCount→0）会累积（当前 1 张，未来随批量换证增多），噪音只会变多。
- 台账已有「剩余天数=已过期」筛选项，但默认视图无区分，噪音与风险混排。

### Urgency

- 当前噪音仅 1 张，但换证操作正在持续进行（近 2 天手动换证/导入活动频繁），旧证引用清零后就会变成噪音。
- 不解决则台账首屏被无价值行占用，可操作性（可换证/可观察风险）被稀释。

## Proposed Solution

台账默认列表**不显示「已过期且 referenceStatus=no_refs_scanned（已成功扫描且明确零引用）」的证书**（服务端过滤，保证分页/total 一致），工具栏显示行内提示「已隐藏 N 张无引用过期证书 · 查看全部」，点击切换为显示全部（提示变为「再次点击隐藏无引用过期证书」，再点恢复隐藏）。**引用三态为 `has_refs` / `no_refs_scanned` / `blind_spot`：仅 `no_refs_scanned` 判定为可隐藏；`blind_spot`（无成功引用快照或扫描范围未覆盖）与引用状态未知一律保守显示，绝不隐藏**。**被线上引用的过期证书始终显示**（风险信号）。**主动筛选「已过期」（含看板「已过期」卡跳转 `/certs?daysLeft=expired`）时显示全部过期证书**，与看板「已过期」卡计数保持一致——该豁免由**后端强制**：exclude 参数不改变 `daysLeft=expired` 的语义（`daysLeft=expired` 查询恒返回全部过期证书），前端无任何豁免逻辑，看板跳转的「行数=卡计数」一致性由此天然保证。隐藏计数 N 由服务端在**同一响应**内按**当前其余过滤条件（search/hostingStatus）上下文**统计返回（`visibleTotal + hiddenCount` 双口径），禁止前端推算。

**切换态 × 筛选态状态机**：用户切换态（展开/隐藏）仅作用于默认视图；进入 `daysLeft=expired` 筛选时切换态**挂起**（显示全部，不显示隐藏提示）；清除筛选后回到切换态；切换导致 total 变化时**重置到第 1 页**。

### Innovation Highlights

- 以引用三态（是否被线上资源引用）区分「噪音」与「风险信号」，而非一刀切隐藏所有过期证书——普通做法是「默认隐藏已过期」，会藏掉真实漂移；本方案按 `no_refs_scanned`（已成功扫描且明确零引用）判定噪音，`blind_spot`/未知引用保守显示，保留可操作的风险可见性。
- 隐藏计数由服务端返回（`隐藏 N 张`），切换与计数同步，避免前端过滤导致分页/total 漂移。
- 主动筛选「已过期」时隐藏不生效，保持看板跳转与卡计数口径一致（呼应前期 7cdd488 看板点卡跳台账修复）。

## Requirements Analysis

### Key Scenarios

- 默认进台账：无引用过期证书（referenceStatus=no_refs_scanned）不出现（当前 1 张隐藏），total 与行一致，工具栏提示「已隐藏 1 张无引用过期证书 · 查看全部」；计数为当前其余过滤条件上下文内的服务端 hiddenCount。
- 点击「查看全部」：无引用过期证书显示出来，提示变「再次点击隐藏无引用过期证书」；再点恢复隐藏；切换导致 total 变化时重置到第 1 页。
- 有引用过期证（10 张）默认视图始终可见，徽章红色「已过期 N 天」。
- blind_spot / 引用状态未知的过期证书：保守显示，任何情况下不隐藏。
- 筛选「剩余天数=已过期」或 `/certs?daysLeft=expired`：全部过期可见（含无引用），行数与看板「已过期」卡计数一致（由后端豁免强制保证）；此时切换态挂起、不显示隐藏提示；清除筛选后回到切换态。
- 结合托管状态/搜索过滤：隐藏逻辑与其它过滤正交（先按其余条件过滤，再隐藏无引用过期行）；隐藏计数也随该上下文重算。
- 空结果：默认隐藏后无可操作证书 → 空态视图仍显示「已隐藏 N 张无引用过期证书 · 查看全部」提示条与「查看全部」入口（提示条不随表格主体消失），并提供空态文案指引（调整筛选）。
- **记录取舍**：被隐藏的无引用过期证书（如 `*.jlc-robot.com`）同时也是「可删除」的清理对象，查看入口（展开）是发现它们的唯一路径——隐藏使清理入口变深，以「提示文案 + 一键展开」换取首屏降噪，属有意取舍。

### Non-Functional Requirements

- **口径一致**：分页 total、行数、隐藏计数三者闭合，且闭合限定在同一过滤上下文（search/hostingStatus）内；与看板「已过期」卡计数一致（后端豁免保证）。
- **可发现性**：隐藏行为必须有可见提示（行内提示切换），不得静默吞行；空态下提示条仍保留。
- 性能：复用 listCertsApi 主查询，仅新增一次同条件统计口径（或单查询双 total 返回 `visibleTotal + hiddenCount`），不新增独立接口。

### Constraints & Dependencies

- 后端 `GET /certs`（ledger_handler.go / ledger_service）现有 search/hostingStatus/daysLeft 服务端过滤——新增「排除无引用过期」参数需与 daysLeft/expired 联动的语义；**豁免落点已定死为后端**：`daysLeft=expired` 时不排除，exclude 参数不改变 `daysLeft=expired` 语义，前端无豁免逻辑（杜绝 URL 直达等入口绕过）。
- 前端台账 `src/views/cert/ledger/`（index.vue / CertTable.vue / format.ts）现有筛选状态源。
- 看板跳转 `/certs?daysLeft=expired`（7cdd488）必须保持显示全部过期。

## Alternatives & Industry Benchmarking

### Industry Solutions

- **cert-manager（K8s 开源证书管理）**：证书续期由 renewal threshold 驱动，过期且未续期的证书仅告警不自动清理，清理对象聚焦「不再被引用的证书资源」——即以「是否仍被 workload 引用」区分处置对象，与本方案「按引用三态判定噪音」思路一致。
- **商业证书管理产品（Venafi TLS Protect / Keyfactor 等类目）**：过期证书普遍采用「归档/折叠出默认列表」的处理（默认隐藏已过期），但均保留检索入口且不对「仍在部署中使用的过期证书」做静默降噪——过期仍被使用属于告警级事件。
- 共同取舍：**默认降噪可以，但不吞风险**。本方案按引用三态区分噪音（`no_refs_scanned`）与风险（`has_refs` / `blind_spot`），并要求隐藏必有可见提示与一键展开入口，对齐该行业共识。
- 按生命周期分桶（进行中/正常/临期/过期归档）是另一种组织方式，但对台账单列表改动大。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 默认隐藏全部过期证书 | 行业常见 | 列表最干净 | 藏掉 10 张「线上在用过期」风险信号 | Rejected: 丢风险 |
| 默认隐藏「过期 + no_refs_scanned」（blind_spot/未知保守显示） | 本方案 | 去噪精准，保留风险可见 | 需后端过滤参数 + 提示切换 | Accepted |
| 折叠/灰化过期 | 轻量展示优化 | 改动小 | 噪音仍占位，可操作性未提升 | Rejected: 治标 |
| Do nothing | — | 零成本 | 噪音累积，首屏被占用 | Rejected: 当前 1 张可先导验证，换证持续将恶化 |

## Scope

### In Scope

- 后端 `GET /certs` 新增过滤：排除「已过期且 referenceStatus=no_refs_scanned（已成功扫描且明确零引用）」的行；`blind_spot` 与引用状态未知的行保守显示；`daysLeft=expired` 时豁免由后端强制（exclude 参数不改变该筛选语义）；同一响应返回 `visibleTotal + hiddenCount` 双口径，hiddenCount 按当前其余过滤条件上下文统计。
- 前端台账默认应用该过滤：工具栏行内提示「已隐藏 N 张无引用过期证书 · 查看全部」，点击切换显示/隐藏（展开态提示「再次点击隐藏无引用过期证书」），计数联动，切换改变 total 时重置到第 1 页；空态视图保留提示条与「查看全部」入口。
- 主动筛选「已过期」（含 `?daysLeft=expired` 跳转）时显示全部过期（后端强制豁免，前端无豁免逻辑；切换态挂起，清除筛选后恢复）。
- 前端 vitest + 后端单测覆盖新增过滤与切换交互，含 blind_spot / refCount=null（引用状态未知）过期证书默认可见的用例；typecheck 通过。

### Out of Scope

- 批量删除/清理无引用过期证书（本需求为隐藏；删除能力已存在，逐行）。
- 探测/差异告警行为变更。
- 证书归档状态（不引入 archive 生命周期）。
- 看板页变更（仅保持跳转即显示全部过期）。

## Risks

| # | Risk | 可能性 | 影响 | Mitigation |
|---|------|--------|------|-----------|
| 1 | 隐藏计数与分页 total 漂移（前端过滤导致） | 低 | 高 | 服务端过滤 + 同一响应返回 `visibleTotal + hiddenCount` 双口径（按当前过滤上下文统计），杜绝前后端口径分裂与前端推算 |
| 2 | 隐藏有引用过期证（误藏风险信号） | 低 | 高 | 过滤条件严格限定 `expired AND referenceStatus=no_refs_scanned`；单测锁定「有引用过期可见」「refCount=null/未知引用过期可见」 |
| 3 | blind_spot / 引用状态未知被误判隐藏（扫描未完成或范围未覆盖的线上在用过期证被静默吞掉） | 中 | 高 | 三态判定：仅 `no_refs_scanned` 可隐藏，`blind_spot`/未知一律保守显示；后端单测覆盖 blind_spot 用例 |
| 4 | 隐藏后证书引用状态漂移（如重新被引用）被漏看，唯一恢复路径是发版回滚 | 低 | 中 | 行内提示常显隐藏计数 + 一键展开入口；判定集中在 ledger_service 一处可快速调整；过滤行为加配置开关可一键关闭隐藏（回归显示全部），后续可按 blind_spot 占比等阈值自动降级 |
| 5 | 看板「已过期」卡跳转后计数对不上（回口径不一致问题） | 低 | 中 | 豁免由后端强制：exclude 参数不改变 daysLeft=expired 语义，天然保证一致；联调动态断言行数=卡计数 |
| 6 | 用户不知道有证书被隐藏（静默吞行） | 中 | 中 | 始终显示「已隐藏 N 张…」行内提示（非静默），空态下提示条仍保留 |
| 7 | 未来 refCount/referenceStatus 语义变化影响判定 | 低 | 中 | 三态语义已在本仓库固化（reference_status.go）；判定集中在 ledger_service 一处，便于调整 |

## Success Criteria

- [ ] 台账默认视图不显示「已过期且 referenceStatus=no_refs_scanned」的证书，分页 total 与行数一致。
- [ ] 工具栏显示「已隐藏 N 张无引用过期证书 · 查看全部」，点击显示全部（提示变「再次点击隐藏无引用过期证书」）、再点恢复隐藏，计数与行实时一致（N 为当前过滤上下文内的服务端 hiddenCount）。
- [ ] `has_refs` 过期证书与 `blind_spot` / 引用状态未知的过期证书在默认视图始终可见（保守显示）。
- [ ] 选「剩余天数=已过期」或访问 `/certs?daysLeft=expired`（含看板卡跳转）：全部过期可见，且 daysLeft=expired 筛选行数 = 看板「已过期」卡计数（动态相等断言，不写死数值）。
- [ ] 隐藏逻辑与搜索/托管筛选正交（组合过滤结果正确，隐藏计数随上下文重算）；空态场景下仍显示「已隐藏 N 张…」提示条与「查看全部」入口。
- [ ] 前端 vitest（format/filter + 标签逻辑）+ 后端 ledger 单测覆盖，含 blind_spot / 引用状态未知过期证书默认可见的用例；`vue-tsc` 通过。