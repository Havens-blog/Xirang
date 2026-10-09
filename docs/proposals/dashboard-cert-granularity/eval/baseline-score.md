# Baseline Score — dashboard-cert-granularity (Iteration 1)

- Subject: `D:\Haven\docs\proposals\dashboard-cert-granularity\proposal.md`
- Rubric: `forge/3.0.2/skills/eval/rubrics/proposal.md` (1000 pts, target 900)
- Scorer stance: adversarial CTO; every claim treated as unverified until supported. Code-level assertions were spot-checked against `D:\Haven\e-cam-web` and `D:\Haven\e-cam-service`.
- Verification notes (claims that CHECKED OUT, credit where due):
  - `index.vue:158-165` decision comment exists verbatim ("证书粒度计数 → 台账 daysLeft 过滤可见具体证书……"). ✓
  - `OverviewCards.vue:13` static copy "点击跳转台账按档过滤" — exact line. ✓
  - `index.vue:73-74` "只读查看者的默认页与唯一菜单入口" (doc cites :75, actual :74 — trivial off-by-one). ✓
  - `dashboard_service.go`: `CountsByLevel` cert-loop, `owner` map collapsing same-domain certs to newest `notAfter` (`preferOwner`), `probes.LatestPerDomain` as `diffAlertCount` source — all real. ✓
  - ledger-hide: `ledger_service.go:217 listCertsHideExpiredNoRefs`, `hiddenCount` dual-count, three-state `no_refs_scanned ≠ blind_spot` (`reference_handler_test.go:197`), orphan definition "已过期且 no_refs_scanned" matches ledger exactly. ✓
  - `WildcardProbeOverrides` exists in settings handler. ✓

**TOTAL: 815/1000** (target 900 — below gate)

Per-instance deduction-rule note: vague-language instances ("行高密度化", "中等改造") are folded into the affected criterion scores below rather than double-applied as flat -20s on top.

---

## 1. Problem Definition — 103/110

| Criterion | Score | Notes |
|---|---|---|
| Problem stated clearly | 40/40 | Root cause precisely named and mechanistic: "分级卡计数是证书粒度，主表却是域名粒度（同域多张证书仅展示 notAfter 最新者独占一行）". Two readers cannot reasonably diverge. Verified against `dashboard_service.go:179-195`. |
| Evidence provided | 35/40 | Three of the evidence items are code-anchored and verified. Deducted for unverifiable evidence: "（brainstorm Q1 实测确认）" — the brainstorm artifact is not linked; "用户此前实测反馈：'到期看板有的点击按分级过滤为空'" — no ticket/screenshot reference; "当前 jlc 数据 94 证 / 93 域" — data point with no source or query. |
| Urgency justified | 28/30 | Concrete cost-of-delay: "与 ledger-hide 的隐藏语义产生两个页面计数对不上……不一致窗口越久，用户信任损耗越大" plus the risk blind spot. Deducted: the blind-spot claim ("可能仍在被引用或裸奔") is asserted without any count of how many such orphan certs actually exist in jlc data — the single strongest urgency lever is left unquantified. |

## 2. Solution Clarity — 108/120

| Criterion | Score | Notes |
|---|---|---|
| Approach is concrete | 36/40 | Very concrete: cert-granular rows, in-page filter, worst-first badge aggregation, orphan dual-count. Deducted: "行高密度化" is vague and unquantified (density of what, target row count?); the payload spec never states whether `hiddenCount` is global or per-tier, yet both card and table behavior depend on it. |
| User-facing behavior described | 39/45 | Five scenarios including boundaries and error path ("探测数据缺失（行上探测徽标为空'未探测'）"). Deducted: the card number under orphan hiding is ambiguous — the happy path says "见'已过期 11'卡，点击 → 主表就地筛出 11 张过期证书行" while the default table hides orphans, so 11 cannot be simultaneously the full count and the visible-row count unless card semantics change, which In Scope denies (see D9/D10). Combined state (orphan toggle × tier filter × special-card filter) is never described. |
| Technical direction clear | 33/35 | Names real files/functions (`dashboard_service.go` `CountsByLevel`, `probeByDomain[san]` aggregation, `router.push` removal) — all verified to exist. Deducted: the reuse target for the three-state judgment is a private method on `ledgerService` (`ledger_service.go:217,451`, unexported); reuse requires an extraction/refactor step the doc does not mention ("必须复用同一实现，不重写" glosses over how). |

## 3. Industry Benchmarking — 85/120

| Criterion | Score | Notes |
|---|---|---|
| Industry solutions referenced | 22/40 | Innovation Highlights names "补天、NJR 类证书台账、云厂商 Certificate Manager 仪表盘" — no links, no product versions, no screenshots. "补天" is best known as a crowdsourced vulnerability platform (奇安信), not a certificate-expiry dashboard; "NJR 类" is unexplained jargon. The dedicated Industry Solutions section is entirely generic: "证书到期仪表盘主流形态：证书行主表 + 分级计数卡 + 到期筛选（各类 Cert Manager 面板）" — zero named products. As the adversary: I cannot verify a single named benchmark. |
| At least 3 meaningful alternatives | 22/30 | Exactly 3 including do-nothing; each genuinely different; do-nothing is rejected on problem-tied grounds, not straw-manned. Deducted: 方案B's source is "画布备选" (internal brainstorm canvas), so no industry-validated solution appears *as an alternative* — the industry alignment claim is attached only to the selected option. |
| Honest trade-off comparison | 21/25 | Selected option's cons stated ("后端+前端中等改造；聚合语义需明确定义"); rejected option's pros stated ("纯前端小改"). Deducted: "聚合语义需明确定义" is an odd con since the order IS defined in Requirements — the trade-off cell under-sells what was already solved; "中等改造" is unquantified. |
| Chosen approach justified against benchmarks | 20/25 | Rationale exists ("业界……普遍以证书行为载体呈现到期，域名仅作为字段而非行主体"), but it inherits the credibility problem of the unnamed benchmarks above. |

## 4. Requirements Completeness — 92/110

| Criterion | Score | Notes |
|---|---|---|
| Scenario coverage | 34/40 | Happy path + 同名多证 + 探测态冲突 + 孤儿 + 错误路径. Deducted: no combined-state scenario (orphan toggle active while a tier card is selected); no scenario for a cert whose SANs include both wildcard-skipped and probed results beyond the one given pair; no concurrent-refresh behavior ("看板 API 载荷更新期间仅前端消费" addresses callers, not the user mid-toggle during refresh). |
| Non-functional requirements | 33/40 | Performance quantified with data ("94 证 / 93 域"), accessibility explicit (`aria-live`, `aria-pressed`, 骨架屏), compatibility explicit breaking-change call, read-only Hard Rule restated. Deducted: no performance budget (max rows / render latency target) even though the doc admits row count grows; no browser/viewport note for denser rows. |
| Constraints & dependencies | 25/30 | Dependencies named and verified real (`probes.LatestPerDomain`, drawer, `WildcardProbeOverrides`, `countsByLevel`, ledger three-state, `daysLeft` semantics). Deducted: helper-export refactor unstated (see D2); no statement on whether the breaking payload change needs a coordinated deploy window or feature flag. |

## 5. Solution Creativity — 69/100

| Criterion | Score | Notes |
|---|---|---|
| Novelty over industry baseline | 24/40 | The doc is honest: "直接采用该行业的基础范式：仪表板主表对齐真实业务实体（证书）粒度" — this is adoption, not innovation. The one genuinely differentiating element is "两页面隐藏口径单源不漂移——这是本次设计的工程核心", which is process discipline more than novel design. |
| Cross-domain inspiration | 24/35 | "借鉴 CI 状态徽标与监控 dashboard 的'最差优先'reducer 模式" — a real, apt cross-domain borrow for the 1:N probe→cert attachment problem, but stated generically without a named exemplar (e.g., GitHub commit status / Prometheus worst-state aggregation). |
| Simplicity of insight | 21/25 | The core insight is elegant — "卡片计数与表行数'不可能一致，所以必须跳走'……根因就是粒度错配" — a genuine "why didn't we see this" reversal, correctly traced to commit 7cdd488's compromise. Not overengineered; no speculative machinery. |

## 6. Feasibility — 85/100

| Criterion | Score | Notes |
|---|---|---|
| Technical feasibility | 35/40 | Verified against code: the backend change is indeed a loop-shape change (`owner` map → cert loop), no new data source, probes already per-domain. "无新依赖、无 showstopper" holds. Deducted: unexported-helper extraction not accounted (D2), and `probeByDomain[san]` assumes SAN→domain key alignment with `LatestPerDomain` output that is asserted, not demonstrated (a wildcard cert's SAN `*.x` vs probe key forms). |
| Resource & timeline feasibility | 24/30 | "估算约 1 个 feature 周期（2~3 任务：后端载荷+前端交互+联调验证）" is plausible given verified prior deliveries (ledger-hide, UI audit) and matching stack. Deducted: no estimate basis; the risk table itself predicts "既有 dashboard 测试断言……大面积失效" — test rebuild effort is acknowledged as a risk but not reflected in the 2~3 task estimate. |
| Dependency readiness | 26/30 | "ledger-hide 已合入 main（三个 commit），三态判定函数可直接复用" — implementation and tests verified present. Deducted: "探测管线稳定（本轮验证期间正常出探测记录）" is anecdotal, and "可直接复用" contradicts the helper's current unexported status. |

## 7. Scope Definition — 74/80

| Criterion | Score | Notes |
|---|---|---|
| In-scope items are concrete | 28/30 | Five deliverables with file-level targets and observable behaviors ("卡计数⇔表行数动态相等；隐藏 N 张提示"). Deducted: item 1's "含孤儿" plus "过期豁免恒显示" introduces an undefined exemption rule (what exactly is 恒显示 — protectUntil-protected? probe-exempt?) into an otherwise concrete item. |
| Out-of-scope explicitly listed | 22/25 | Five named exclusions including the smart freeze "diffAlertCount/exemptCount 保持域名级计数不动". Deducted: "3 张覆盖率卡" is referenced as if defined, but the doc never introduces the three coverage cards anywhere — orphan reference requiring external context. |
| Scope is bounded | 24/25 | Single API + single page + tests, 2~3 tasks, explicit non-goals including the read-only Hard Rule. Bounded. Minor deduction: the unstated helper-extraction refactor could expand the blast radius beyond the stated boundary. |

## 8. Risk Assessment — 78/90

| Criterion | Score | Notes |
|---|---|---|
| Risks identified | 28/30 | Six non-trivial risks, including the self-aware "看板与台账隐藏判定实现漂移" and "空态/全隐藏边界". Not padded. |
| Likelihood + impact rated | 27/30 | Mixed and honest — includes H/M ("多 SAN 探测态聚合歧义") rather than uniformly low. Deducted: "同名多证使主表行数增多" rated M/L sits awkwardly against the NFR admission that rows grow and the doc's own evidence that same-domain multi-cert is the core case. |
| Mitigations are actionable | 23/30 | Most are genuinely actionable: "显式最差优先序写入实现+单测"; "grep 全仓 getCertDashboardApi 引用核验" (I verified the symbol exists in exactly 2 source files, so the check is cheap and real); "服务端单测锁定两处一致性". Deducted: "行高密度化" is not actionable as written; and — CTO-perspective miss — **there is no rollback plan for a deliberately breaking API change** ("载荷 items 结构为破坏性变更……同步改造即可"): no deploy-ordering, no flag, no revert story if frontend and backend ship out of sync. |

## 9. Success Criteria — 51/80

| Criterion | Score | Notes |
|---|---|---|
| Criteria are measurable and testable | 21/30 | Mostly written as assertions ("items 长度 == 全部证书数", "无 router.push('/certs'...) 残留", "全部门禁绿"). Deducted: SC #2 and SC #5 hinge on equalities that are ambiguous or unsatisfiable (below); SC #3's ledger cross-check wording is ambiguous (see blindspot 3). |
| Coverage is complete | 20/25 | Covers API shape, cards, orphans, aggregation order, special cards, CI gates. Deducted: the accessibility NFR has no corresponding SC; the "过期豁免恒显示" rule has no SC; empty-state criterion lives only in the risk table. |
| SC internal consistency | 10/25 | **Two contradiction clusters (the heaviest deductions in this review):** |
| | | **(a) SC #5 vs frozen domain-level counts.** SC #5: "diff/exempt 卡点击 → 主表筛出含对应探测徽标的证书行，计数与卡片值一致" — while Out of Scope freezes "diffAlertCount/exemptCount 保持域名级计数不动", and the backend confirms the count source is per-domain (`dashboard_service.go:108`: "LatestPerDomain（items.probeStatus 与 diffAlertCount 数据源）"). When a diff-domain is covered by ≥2 certs (exactly the 同名多证 case the proposal centralizes), every covering cert row carries the diff badge, so filtered rows > domain-level card value. Satisfying one clause makes the other impossible. |
| | | **(b) SC #2 vs countsByLevel 语义不动 + default orphan hiding.** SC #2: "选中卡 → 主表行数 == 该档卡片计数（动态相等）"; In Scope: "保持 summary/countsByLevel 语义不动（本就是证书粒度）"; SC #3: "默认视图不含'过期且 no_refs_scanned'行". The verified `CountsByLevel` loop counts **all** certs including orphans; the default table hides them; therefore the expired tier's rows < card count whenever an orphan exists — the proposal reintroduces the exact "点击按分级过滤对不上" defect it lists as evidence. The "可见 N · 隐藏 M" card treatment cannot rescue this while countsByLevel semantics are frozen and `hiddenCount` is not defined per-tier. |

## 10. Logical Consistency — 70/90

| Criterion | Score | Notes |
|---|---|---|
| Solution addresses the stated problem | 33/35 | Direct hit: granularity mismatch → cert-granular rows → "计数⇔表行数天然一致". The Assumptions Challenged table shows the reasoning path from 7cdd488's compromise to the reversal. Minor: the "语义等价" claim for special-card migration is asserted, not shown (and is false under frozen counts — see D9a). |
| Scope ↔ Solution ↔ SC aligned | 17/30 | Same two contradictions, now cross-section: Out-of-Scope freezes (domain-level diff/exempt counts; unchanged countsByLevel) are load-bearing inputs to SC #2/#5 equalities that require cert-level counts. This is the rubric's exact D10 territory. |
| Requirements ↔ Solution coherent | 20/25 | Requirements map cleanly to solution elements (orphan requirement ↔ dual-count hiding; 1:N probe ↔ worst-first badge; filter single-source ↔ in-page selection). Deducted: undefined "过期豁免恒显示" and the unexamined SAN-key alignment between cert SANs and probe domain keys. |

---

## Cross-dimension coherence check

D9 (51) and D10 (70) are low for the same root cause — the two count-semantics contradictions — which is consistent. D3's weakness (unverifiable benchmarks) also suppresses D5-Novelty (the doc admits it adopts the baseline paradigm), which is coherent, not accidental. No dimension score is inconsistent with the evidence trail.

## Attack list

1. **[Success Criteria/Logical Consistency]** SC #5 equality is unsatisfiable under frozen domain-level counts — "diff/exempt 卡点击 → 主表筛出含对应探测徽标的证书行，计数与卡片值一致" vs "diffAlertCount/exemptCount 保持域名级计数不动" — a diff domain covered by 2+ certs yields more rows than the card value. Must migrate special-card counts to cert-row semantics (or specify which single row carries the badge) and delete the false "语义等价" claim.
2. **[Success Criteria/Logical Consistency]** SC #2 "主表行数 == 该档卡片计数（动态相等）" contradicts "保持 summary/countsByLevel 语义不动" plus default orphan hiding — the expired card will again outnumber visible rows, recreating the "点击按分级过滤对不上" defect cited as Problem evidence. Must define per-tier visible counts (or move card semantics to visible-only) and specify whether hiddenCount is global or per-tier.
3. **[Industry Benchmarking]** Benchmarks are unverifiable and partly dubious — "业界证书/资产管理系统（补天、NJR 类证书台账、云厂商 Certificate Manager 仪表盘）普遍以证书行为载体" — 补天 is a crowdsourced vuln platform, "NJR 类" unexplained, and the Industry Solutions section names zero products ("各类 Cert Manager 面板"). Must cite 3+ verifiable named products with sources.
4. **[Feasibility/Requirements]** The single-source reuse premise glosses over an unstated refactor: "必须复用同一实现，不重写" targets `deriveReferenceStatus` / `listCertsHideExpiredNoRefs` which are unexported methods on `ledgerService` (`ledger_service.go:217,451`). Must scope the extraction into a shared package.
5. **[Risk Assessment]** No rollback/deploy-ordering plan for a deliberately breaking change — "GET /certs/dashboard 载荷 items 结构为破坏性变更；已确认单消费端……同步改造即可". Must state deploy ordering/flag/revert story.
6. **[Success Criteria]** SC #3's ledger cross-check is ambiguous — "与台账 daysLeft=expired 经 includeExpiredNoRefs 的行数口径一致" — while the ledger's documented behavior is "daysLeft=expired 后端强制豁免" (`ledger_handler_test.go:214-215`), making the double qualification redundant or contradictory as written. Must name the exact reference view.
7. **[Solution Clarity/Requirements]** Undefined term "过期豁免恒显示" — "孤儿隐藏双口径（visible + hiddenCount，复用 ledger 三态判定，过期豁免恒显示）" — "豁免" collides with the probe-exempt card semantics used elsewhere in the same doc, and the ledger's protection-period dimension is never mentioned. Must define the exemption rule precisely.
8. **[Problem Definition]** Evidence items without artifacts — "（brainstorm Q1 实测确认）", "用户此前实测反馈：'到期看板有的点击按分级过滤为空'", "当前 jlc 数据 94 证 / 93 域". Must link the brainstorm record, feedback ticket, and the data query.
9. **[Scope Definition]** Orphan reference — "3 张覆盖率卡与 2 张特殊卡的计数语义" assumes reader knowledge of "3 张覆盖率卡" that the document never defines. Must define or link.
10. **[Solution Clarity/Risk Assessment]** Vague mitigation and missing combined-state behavior — "行高密度化" is unquantified; no scenario defines behavior when the orphan toggle, a tier card, and a special card are active simultaneously. Must quantify density targets and specify filter composition semantics.

## Blindspot hunt (outside rubric dimensions)

1. **[blindspot] Self-reintroduction of the solved defect.** The proposal's happy path "见'已过期 11'卡，点击 → 主表就地筛出 11 张过期证书行" cannot survive its own orphan-hiding rule "默认视图不含'过期且 no_refs_scanned'行" unless card counts change — yet the doc freezes them. The document cites "点击按分级过滤为空" as evidence of the current defect and then builds a configuration that reproduces it. Quote pair: "见'已过期 11'卡，点击 → 主表就地筛出 11 张过期证书行" + "保持 summary/countsByLevel 语义不动（本就是证书粒度）".
2. **[blindspot] Badge duplication changes risk perception, not just counts.** With cert rows, one sick domain covered by N certs renders the worst-first badge N times, visually amplifying a single domain problem ("探测状态聚合为行级徽标"). The doc treats this only as a counting question, never as an operator-trust/UX question.
3. **[blindspot] Ledger force-include interaction.** The ledger's "daysLeft=expired 后端强制豁免" behavior (`ledger_handler_test.go:214-215`) means the reference view SC #3 points at differs from the ledger's default view; the doc never reconciles the three views (dashboard default / dashboard show-all / ledger expired). Quote: "与台账 daysLeft=expired 经 includeExpiredNoRefs 的行数口径一致".
4. **[blindspot] SAN-key alignment assumption.** "探测经 probeByDomain[san] 聚合" assumes cert SAN strings match `LatestPerDomain` domain keys byte-for-byte (wildcard forms, case, trailing dots). No requirement or test asserts this mapping; a mismatch silently yields "未探测" rows. Quote: "探测经 `probeByDomain[san]` 聚合、孤儿经 ledger 复用的三态判定过滤".
5. **[blindspot] Sort-direction claim unverified.** Mitigation says "后端按 notAfter asc 排序到期优先" — the current service's sort contract is not cited; if the existing output is domain-ordered, the new ordering needs its own SC, which the SC list lacks (no ordering criterion at all).
