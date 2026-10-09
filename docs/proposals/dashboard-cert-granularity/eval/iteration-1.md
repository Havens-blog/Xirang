# Scorer Report — iteration 1

- Target: `D:\Haven\docs\proposals\dashboard-cert-granularity\proposal.md`
- Rubric: forge 3.0.2 `skills/eval/rubrics/proposal.md` (1000-scale, target 900)
- Verification stance: every code/data claim treated as unverified unless the document itself supplies evidence. Codebase was deliberately not consulted.

**TOTAL: 734/1000**

---

## Phase 1 — Reasoning Audit

Chain traced: Problem (粒度矛盾 → 跳台账妥协) → Solution (证书粒度主表 + 页内筛选 + 行级徽标 + 服务端孤儿隐藏 + 双口径计数) → Evidence (code comment 7cdd488、静态文案、用户反馈、相邻特性) → SC (8 条门禁式断言). The spine holds; the failures are at the seams.

### Self-contradiction: does the solution reintroduce what it eliminates?

Partially yes. The Problem is defined as *"分级卡计数是证书粒度，主表却是域名粒度"*, and Urgency escalates on *"两个页面计数对不上……用户信任损耗越大"*. The solution then deliberately preserves the inverse mismatch for special cards: *"`diffAlertCount`/`exemptCount` 保持域名级计数；点击后主表行集合 = 「存在任一 SAN 为该状态的证书行」"*, with the worked example *"一个 diff 域名挂 2 张证 → 卡计数 1、行数 2"*. A user clicking "diff 1" lands on 2 rows. The document redefines success as "映射闭合（非数值相等）" rather than resolving it. That is a defensible engineering choice but it is not the elimination the Problem section promises, and no user-facing explanation of the 1→2 discrepancy is specified.

### SC Consistency Deep-Dive

Clusters by affected area:
- **C1 backend `dashboard_service.go` / `GET /certs/dashboard` payload**: SC-1, SC-3, SC-4(aggregation), SC-6, SC-7, In-Scope#1
- **C2 frontend `DashboardTable`**: SC-1(referenceStatus per row), SC-3(toggle), SC-4(drawer), In-Scope#2, Risk#7
- **C3 `OverviewCards`**: SC-2, SC-6, In-Scope#3, In-Scope#4
- **C4 gates**: SC-8

Bidirectional derivations run within clusters. Confirmed findings:

```
CONTRADICTION: SC-3 "默认视图（无分级筛选）不含"过期且 no_refs_scanned"行"
            ↔ SC-6 "点击后主表行集合 == 「存在任一 SAN 为该状态的证书行」集合"
Type: mutual-exclusion
Evidence: The orphan exemption is scoped exclusively to 分级卡 — In Scope #3 says
  "激活即豁免孤儿隐藏", and SC-2 repeats "激活该档即豁免孤儿隐藏". diff/exempt are
  explicitly 特殊卡, not 分级卡 (Out of Scope: "2 张特殊卡"). So clicking exempt is a
  no-分级-filter state → SC-3 keeps orphans hidden. Assume SC-6 satisfied: the row set
  equals ALL certs with a matching SAN, including an expired no_refs_scanned cert whose
  SAN is exempt → SC-3 violated. Assume SC-3 satisfied: that cert is absent → the set
  equality in SC-6 fails. Neither can hold with the other for any dataset containing an
  orphan cert with a diff/exempt SAN.
Resolution required: state whether 特殊卡 activation also exempts orphan hiding, or
  restrict SC-6's set definition to "可见行" and add the corresponding 单测.
```

```
CONTRADICTION: Risk#5 mitigation "部署序：后端先上（旧前端读不到新字段仅降级为不显示隐藏提示，不白屏）"
            ↔ Risk#7 "DashboardTable 现用 row-key="domain"，证书粒度化后域名不再唯一，Vue 行复用错乱"
Type: direction-clash
Evidence: Backend-first delivery sends cert-granularity items to the unmodified frontend.
  The unmodified frontend is, by the document's own Risk#7, the exact code that breaks on
  duplicate domains. The doc even supplies the data proving duplicates exist today:
  "94 证 / 93 域". So the deploy-order mitigation guarantees the failure mode the next row
  identifies, and the claim "不白屏" is asserted with zero supporting evidence. The same
  window also silently violates SC-3's "不静默" property (old UI shows server-hidden row
  counts with no toggle or hidden-count affordance).
Resolution required: either frontend-first behind a payload-shape flag, a versioned field
  (items_v2 / Accept header), or an explicit admission that a short breakage window is
  accepted, with the row-key fix shipped in the same deploy.
```

```
AMBIGUOUS — requires author clarification: SC-2 "激活该档即豁免孤儿隐藏" ↔ SC-3 "切换"查看全部"后行数 == visible + hidden"
Type: resource-contention (two controls own the same visibility state)
Evidence: With a 分级卡 active, the "查看全部" toggle is already effectively on for that
  level. The document never states whether the toggle is disabled, forced-on, visually
  reconciled, or silently inconsistent, nor what the card itself renders while active
  (In Scope only specifies "未激活时该档显示"可见 N · 隐藏 M""). Two SC entries write to
  one piece of state with no precedence rule.
```

Also unresolved: the a11y NFR fixes the `aria-live` total to *"可见行数（不含被隐藏孤儿）"*, but in the 分级卡-active state there are no hidden orphans by SC-2 — consistent, yet the announcement text must therefore change meaning between states, which is unspecified.

---

## Phase 2 — Rubric Scoring

### 1. Problem Definition — 91/110

| Criterion | Score |
|---|---|
| Problem stated clearly | 38/40 |
| Evidence provided | 28/40 |
| Urgency justified | 25/30 |

Strong: the root cause is named precisely and non-ambiguously — *"分级卡计数是证书粒度，主表却是域名粒度（同域多张证书仅展示 `notAfter` 最新者独占一行）"*. Evidence is uncommonly good for citing a commit and file:line: *"`e-cam-web/src/views/cert/dashboard/index.vue:158-165`（commit 7cdd488 决策注释自陈）"*.

Deductions:
- **Own data undercuts the headline harm.** The Problem's core damage is *"过期/临期旧证被同名新证掩盖"*, and Urgency escalates it to *"真实风险盲区"*. But the document's only quantification, in Risk#1, is *"当前 jlc 实测 94 证 / 93 域，+1 行"* — i.e. at most **one** certificate is currently being masked. The proposal never reconciles the severity rhetoric with its own cardinality datum, and never states how many expired-and-masked certs exist today. −8 on evidence.
- **Unsourced anecdote.** *"用户此前实测反馈："到期看板有的点击按分级过滤为空""* — no date, no reporter, no ticket, and no reproduction. −4.
- **Internal artifacts cited as proof.** *"（brainstorm Q1 实测确认）"*, *"brainstorm Q2/Q3/Q4/Q5 确认"* — a brainstorm record is the author's own prior output, not independent evidence of user intent. Used five times as the backing for the central behavioral claim. It is not quotable by a reviewer. −0 additional beyond the above, but flagged.
- Urgency's cost-of-delay is directional only: *"不一致窗口越久，用户信任损耗越大"* — no measure, no deadline, no dependent work blocked. −5.

### 2. Solution Clarity — 111/120

| Criterion | Score |
|---|---|
| Approach is concrete | 38/40 |
| User-facing behavior described | 40/45 |
| Technical direction clear | 33/35 |

The solution is reproducible from the text: a reader can state the five changes. The happy path is genuinely observable-behavior-first: *"点击 → 主表就地筛出 11 张过期证书行（含同名多证并立）"*.

Deductions:
- The entire solution is **one 300-character sentence** packing five independent changes separated by semicolons (line 29). Load-bearing, but the single-sentence form makes the boundary between "must" and "how" unreadable for an approver.
- **Two user-visible states are unspecified**: what the 分级卡 renders while active (only the inactive label *"可见 N · 隐藏 M"* is defined), and what happens to the *"查看全部"* toggle while a level filter is active. Both are things a QA engineer must guess. −5.

### 3. Industry Benchmarking — 68/120

| Criterion | Score |
|---|---|
| Industry solutions referenced | 22/40 |
| At least 3 meaningful alternatives | 18/30 |
| Honest trade-off comparison | 14/25 |
| Chosen approach justified against benchmarks | 14/25 |

This is the weakest dimension and it is not a length problem — it is a substance problem.

- **Benchmarks are named but never examined.** *"业界证书/资产管理系统（补天、NJR 类证书台账、云厂商 Certificate Manager 仪表盘）普遍以证书行为载体呈现到期"* — no link, no screenshot, no version, no specific behavior described. *"NJR 类证书台账"* is not a recognizable product identifier. The generic *"各类 Cert Manager 面板"* and *"监控产品标准做法"* are assertions of consensus with zero citation. Notably, the most relevant public comparables for this exact UI problem — Venafi/CyberArk, Keyfactor Command, DigiCert CertCentral, AWS/Azure/Aliyun Certificate Manager consoles, or OSS like certwatch/cert-manager dashboards — are not examined even though the doc claims to follow them. −18.
- **Only 3 alternatives, one of which self-destructs.** *"Do nothing（维持跳台账） … Rejected: 正是本 proposal 要消除的"* is a circular rejection: the alternative is dismissed because the proposal exists, not on cost/benefit. No estimate of what *not* doing this costs versus the change. −12 across alternatives + justification.
- **Trade-offs are one phrase per cell.** The selected option's only cost is *"后端+前端中等改造；聚合语义需明确定义"* — no file count, no test-rebuild cost, no breaking-API coordination cost, despite Risk#3 and Risk#5 establishing both. Cherry-picked in the approving direction. −11.
- **Overstated verdict.** *"**Selected**: 唯一根治"点击→就地看证书"的方案"* — "唯一" is false on the document's own terms: cert-granularity rows plus in-page filtering alone achieves "点击→就地看证书"; the orphan-hiding and badge-aggregation work are separable additions bundled into the same verdict. −11.

Note on straw-man rule: 方案 B (*"域名主表 + 就地证书抽屉 … 纯前端小改"*) is a **genuine** alternative with a real advantage, and I do **not** apply the −20 straw-man deduction. Its rejection reason *"修标不修本"* is substantive.

### 4. Requirements Completeness — 83/110

| Criterion | Score |
|---|---|
| Scenario coverage | 30/40 |
| Non-functional requirements | 27/40 |
| Constraints & dependencies | 26/30 |

Edge cases are the document's strongest work: 同名多证, 探测态冲突 with a full precedence order, 特殊卡筛选 谓词, 孤儿 三态 — all specified at implementation grain, which is load-bearing, not padding.

Deductions:
- **The error path is one line for a breaking backend change**: *"错误路径：探测数据缺失（行上探测徽标为空"未探测"）；看板 API 载荷更新期间仅前端消费，无第三方调用者。"* Nothing on: dashboard API 5xx/timeout, partial data (refs fetched but scan snapshot stale), what the UI shows if `referenceStatus` is absent on a row, or behavior when the reused ledger 三态判定 errors. For a page that is *"只读角色的默认页与唯一菜单入口"*, degraded-mode behavior is a requirement, not a nicety. −10.
- **No security/authorization requirement.** The change moves the table from domain rows to certificate rows carrying *"CN、SANS 折叠计数、issuer、daysLeft、托管、引用云、探测徽标"* plus 证书指纹 in sort logic, on the default page of the least-privileged role. Whether read-only roles may see per-certificate fingerprints and full SAN lists that were previously collapsed away is never asked. −8.
- **Performance NFR rests on an unverified single-tenant sample.** *"主表行数由"去重域名数"增至"证书数"（同量级，当前 jlc 数据 94 证 / 93 域）"* — the "同量级" invariant is an extrapolation from one tenant with an essentially 1:1 ratio. No pagination/virtualization requirement, no stated ceiling at which the assumption breaks, no source for the 94/93 figure. −5.
- Missing: i18n of the new badge copy (which the doc pins to exact Chinese strings, e.g. *"「该证 SAN 现网一致（生效归属另一证书）」"*).

### 5. Solution Creativity — 68/100

| Criterion | Score |
|---|---|
| Novelty over industry baseline | 20/40 |
| Cross-domain inspiration | 26/35 |
| Simplicity of insight | 22/25 |

- The document **self-reports minimal novelty**: *"直接采用该行业的基础范式"*. Honest, and honesty is worth something — but the rubric asks for differentiation from the benchmark, and the only differentiator offered is *"与相邻特性共用同一引用三态判定函数"*, which is internal hygiene rather than a capability the benchmark lacks. −20.
- Cross-domain borrowing is real and named: *"借鉴 CI 状态徽标与监控 dashboard 的"最差优先"reducer 模式"*. It is one borrowing, applied once, with no second source. −9.
- The insight is elegant: reframing "counts can never match the table, so navigate away" into "the table is at the wrong granularity" is a clean Occam move and the `Assumptions Challenged` table documents it well (*"**Overturned**: 证书粒度化后两者天然一致"*). −3 only because the elegance is then compromised by the retained domain-level special-card counts.

### 6. Feasibility — 75/100, **−20 vague language → 55/100**

| Criterion | Score |
|---|---|
| Technical feasibility | 33/40 |
| Resource & timeline | 20/30 |
| Dependency readiness | 22/30 |
| Deduction: vague language w/o quantification | −20 |

- **Vague-language deduction (−20)**: *"改动集中、无新数据源"* — the central feasibility claim for the backend rewrite, with no file count, no LOC estimate, no function inventory. Rubric rule: "Vague language without quantification: −20 pts per instance."
- **Cross-service reuse asserted, never shown callable.** The plan depends on *"必须复用同一实现，不重写"* of `ledgerService.deriveReferenceStatus`, and on the dashboard service calling it. The document never establishes that this symbol is exported, that `dashboard_service.go` may depend on the ledger service without a cycle, or that its signature accepts the batch shape the NFR demands (*"一次拉取 refs + 最新扫描快照后内存归约"*). This is the single hardest technical dependency in the proposal and it is one unverified sentence. −7.
- **Estimate has no derivation.** *"估算约 1 个 feature 周期（2~3 任务：后端载荷+前端交互+联调验证）"* covers a breaking API change, a table rewrite, a drawer extension, an aggregation reducer, orphan-hiding reuse, dual-count plumbing, and a full test rebuild (*"按新语义重建 `dashboard_service_test.go`"*). Prior-experience appeal — *"团队已完成两轮看板工程（UI audit 35 任务、ledger-hide 双口径）"* — is credibility, not sizing. −10.
- **Dependency readiness is unsourced where the rest of the doc is sourced.** *"ledger-hide 已合入 main（三个 commit）"* gives no SHAs, despite the Evidence section demonstrating the author can cite them (`7cdd488`). *"探测管线稳定（本轮验证期间正常出探测记录）"* — no window, no volume, no error rate. −8.

### 7. Scope Definition — 61/80

| Criterion | Score |
|---|---|
| In-scope items are concrete | 24/30 |
| Out-of-scope explicitly listed | 18/25 |
| Scope is bounded | 19/25 |

In-scope entries are deliverable-shaped and enumerate fields, which is real specification.

Deductions:
- **An in-scope deliverable is buried inside Out of Scope.** *"证书循环直出时通配符跳过计数按 domain 去重，防止共享通配符 SAN 的多张证书重复计数"* appears under `### Out of Scope`. It is code that must be written (Risk#8 even assigns it a unit test: *"单测构造两张共享 `*.example.com` 的证书断言计数为 1"*). Putting work in the Out-of-Scope list is a boundary defect that will cause it to be dropped at task-decomposition time. −7.
- **A behavior change with an SC has no In-Scope entry.** The sort order flip — *"由现状域名字典序**改为**到期优先（notAfter asc，是变更而非维持）"* — is asserted in NFR and gated by SC-7, but no In Scope bullet claims it. −6.
- **A new UI control is named but not specified.** *"孤儿默认隐藏 + 切换入口 + 空态处理"* — "切换入口" has no placement, no label, no persistence rule, and no defined interaction with an active level filter. −4.
- Boundedness rests entirely on the unbacked "1 个 feature 周期"; there is no exit condition other than the gates. −6 spread.

### 8. Risk Assessment — 69/90

| Criterion | Score |
|---|---|
| Risks identified | 25/30 |
| Likelihood + impact rated | 20/30 |
| Mitigations are actionable | 24/30 |

Eight risks, ratings genuinely varied (M/L, H/M, L/L, M/H, L/M, M/M) — this is not the all-low-likelihood-high-impact pattern. Mitigations are mostly executable, e.g. *"grep 全仓 `getCertDashboardApi` 核验单消费端"* and *"单测构造两张共享 `*.example.com` 的证书断言计数为 1"*. A rollback plan exists and is specific: *"两侧各自 revert 单 commit 即恢复，无数据迁移、无持久化状态变更"* — credited.

Deductions:
- **The rollback/deploy mitigation is contradicted by the next risk row** (see Phase 1). *"旧前端读不到新字段仅降级为不显示隐藏提示，不白屏"* is an unverified behavioral claim about code the document elsewhere says breaks on duplicate domain keys. −8 (mitigation actionability + honesty).
- **Self-serving rating.** *"`GET /certs/dashboard` 破坏性变更漏网调用者 | L | L"* — the impact of a missed consumer of a breaking payload change is a broken page, not "L". Rating L/L on the one item that would be user-visible in production reads as approval-seeking. −6.
- **Two identified exposures carry no risk row**: the deploy-window inconsistency it creates for its own "不静默" guarantee, and performance at a tenant whose cert:domain ratio is not ~1:1. −4.
- No risk for read-only-role data exposure (see D4). −3.

### 9. Success Criteria — 59/80

| Criterion | Score |
|---|---|
| Criteria are measurable and testable | 26/30 |
| Coverage is complete | 18/25 |
| SC internal consistency | 15/25 |

These are unusually good SC: assertion-level, with test targets and even a grep. *"`grep -rn "router.push" src/views/cert/dashboard/` 无 `/certs` 跳转残留"* is exactly the falsifiable form the rubric wants, and SC-4's aggregation ordering is fully testable.

Deductions:
- **SC-3 ↔ SC-6 mutual exclusion** (derivation in Phase 1). −10 on internal consistency.
- **SC-2 ↔ SC-3 toggle-state ambiguity** — two SC write the same visibility state with no precedence rule; flagged *ambiguous — requires author clarification*. −5.
- **SC-2 contradicts itself in one clause**: *"（动态相等，不写死数值；"已过期"档以实测 11=11 断言）"* — it forbids hardcoded numbers and then hardcodes 11, tying a gate to one tenant's live data. An assertion of `rows.length === countsByLevel.expired.total` is the testable form; `11 === 11` is environment-coupled and will go red on any data change. −4.
- **Coverage gaps**: no SC for the `wildcardSkippedCount` domain-dedup (Risk#8 promises a test; no gate claims it), no SC for the *"查看全部"* toggle behavior itself, no SC asserting the 3 coverage cards and `wildcardSkippedCount` remained at domain semantics after the refactor (i.e. no regression gate on the Out-of-Scope invariants). −7.
- SC-4 pins gate conditions to exact Chinese UI strings (*"「该证 SAN 现网一致（生效归属另一证书）」"*), making the gate brittle against copy review or i18n.

### 10. Logical Consistency — 69/90

| Criterion | Score |
|---|---|
| Solution addresses the stated problem | 30/35 |
| Scope ↔ Solution ↔ SC aligned | 20/30 |
| Requirements ↔ Solution coherent | 19/25 |

- The solution does attack the stated root cause directly — but it **preserves a granularity mismatch for special cards** while the Problem section defines granularity mismatch as the thing being eliminated (*"卡计数 1、行数 2"*). Redefining closure as *"映射闭合（非数值相等）"* is a scope decision, not a resolution, and the user-facing explanation of the 1→2 jump is absent. −5.
- Scope↔SC misalignments: sorting change gated but not scoped; wildcard dedup scoped into Out of Scope but tested in Risk; the deploy-order ↔ row-key clash. −10.
- Requirements↔Solution: the orphan-exemption rule is written for 分级卡 only, while the 特殊卡 requirement defines a set over all matching certs — an orphan requirement with no solution rule. −6.

### Cross-dimension coherence check

The document's quality is bimodal. Everything downstream of "what exactly gets built" (D2, D4, D7, D9) is specified at a grain most proposals never reach, and that detail is load-bearing — the aggregation precedence order, the per-SAN predicate, and the dual-count contract are all things a task decomposer would otherwise invent wrongly. **It is not padding.** The weakness is concentrated upstream and outward: the industry case (D3) is name-dropped rather than examined, the effort/dependency claims (D6) are the only place the author stops citing sources, and the two hardest seams (deploy ordering, special-card orphan semantics) are exactly where two well-specified sections contradict each other. Net: an execution-ready spec resting on an under-argued business case.

---

## Bias Detection Report

Region counting: a "paragraph" = one bullet, table row, or standalone paragraph. Marker-bearing regions (`<!-- pre-revised: * -->`, including inline row markers): **27**. Non-marked substantive regions: **35**.

| Region class | Attack points | Paragraphs | Density |
|---|---|---|---|
| Annotated | 7 | 27 | 0.26 |
| Unannotated | 8 | 35 | 0.23 |
| **Ratio (annotated / unannotated)** | | | **1.13** |

Attacks located in annotated regions: deploy-order ↔ row-key contradiction (lines 143, 145), SC-3 ↔ SC-6 (153, 159), SC-2/SC-3 toggle ambiguity (120, 151, 153), 11=11 hardcode (151), special-card mismatch retained (47, 122), brittle string gate (157), 94/93 vs problem severity (139).

Attacks in unannotated regions: industry references unexamined (76, 34), do-nothing circular rejection (82), "唯一根治" overstatement (84), "改动集中" vague (90), estimate underived (96), dependency SHAs missing (100), error-path thinness (51), missing security NFR (53–62 block).

Interpretation: densities are within 13% of each other, so no material annotation bias. The most severe finding (deploy-order contradiction) does sit in a `high`-marked pair, and it is a **revision-introduced** defect: the detailed deploy-order/rollback text in Risk#5 and the detailed row-key text in Risk#7 are each individually good additions that were evidently not cross-checked against one another. Tagged accordingly in the attack list.

---

## Attack List

1. **[logical-consistency]** `conflict-with-pre-revision` Deploy-order mitigation guarantees the failure mode the very next risk row identifies — *"部署序：后端先上（旧前端读不到新字段仅降级为不显示隐藏提示，不白屏）"* vs *"`DashboardTable` 现用 `row-key="domain"`，证书粒度化后域名不再唯一，Vue 行复用错乱"* — must resolve: frontend-first, versioned payload field, or explicit acceptance of a breakage window with the row-key fix in the same deploy; and drop the unverified "不白屏" claim.
2. **[success-criteria]** SC-3 and SC-6 are mutually unsatisfiable — *"点击后主表行集合 == 「存在任一 SAN 为该状态的证书行」集合"* vs *"默认视图（无分级筛选）不含"过期且 `no_refs_scanned`"行"* — must state whether 特殊卡 activation also exempts orphan hiding, or scope SC-6's set to visible rows with a matching unit test.
3. **[success-criteria]** SC-2 and SC-3 both own the orphan-visibility state with no precedence rule — *"激活该档即豁免孤儿隐藏"* vs *"切换"查看全部"后行数 == visible + hidden"* — must specify the toggle's state (disabled / forced-on / reconciled) and the card's own label while a level filter is active.
4. **[industry-benchmarking]** Benchmarks are named, never examined — *"业界证书/资产管理系统（补天、NJR 类证书台账、云厂商 Certificate Manager 仪表盘）普遍以证书行为载体呈现到期"* — must cite at least two identifiable products with the specific observed behavior (row entity, count semantics, how each handles multi-cert-per-domain), with links or screenshots.
5. **[industry-benchmarking]** Do-nothing is rejected circularly — *"Rejected: 正是本 proposal 要消除的"* — must give the actual cost of inaction (how many masked certs, what the ledger/dashboard count gap will be, who reports it) rather than dismissing it by reference to the proposal's existence.
6. **[industry-benchmarking]** Verdict overstates exclusivity — *"**Selected**: 唯一根治"点击→就地看证书"的方案"* — cert rows + in-page filtering alone satisfies that goal; must justify the bundled orphan-hiding and badge-aggregation work on their own merits or unbundle them.
7. **[problem-definition]** The document's own cardinality datum undercuts its headline harm — *"当前 jlc 实测 94 证 / 93 域，+1 行，非量级变化"* vs *"过期/临期旧证被同名新证掩盖"* / *"真实风险盲区"* — must state how many certs are actually masked today and how many of those are expired-with-refs, or downgrade the severity framing.
8. **[problem-definition]** Central user-intent evidence is the author's own prior artifact — *"（brainstorm Q1 实测确认）"* and *"brainstorm Q2/Q3/Q4/Q5 确认"* — must point to a reviewable source (ticket, session transcript, named user) for the claim that click intent is in-page viewing.
9. **[feasibility]** Vague language, −20 — *"改动集中、无新数据源"* — must quantify: which files, roughly how many call sites, what the test-rebuild surface is.
10. **[feasibility]** The hardest dependency is one unverified sentence — *"必须复用同一实现，不重写"* / *"`ledgerService.deriveReferenceStatus`"* — must confirm the symbol is callable from `dashboard_service.go` without an import cycle and that its signature supports the batch shape the NFR requires (*"一次拉取 refs + 最新扫描快照后内存归约"*).
11. **[feasibility]** Estimate has no derivation and the readiness claims drop the citation discipline used elsewhere — *"估算约 1 个 feature 周期（2~3 任务）"*, *"ledger-hide 已合入 main（三个 commit）"* — must give per-task sizing and the commit SHAs (the Evidence section proves SHAs are available).
12. **[requirements-completeness]** Error path is one line for a breaking backend change on the default page of the least-privileged role — *"错误路径：探测数据缺失（行上探测徽标为空"未探测"）；看板 API 载荷更新期间仅前端消费，无第三方调用者。"* — must specify degraded behavior for API failure, missing `referenceStatus` on a row, stale scan snapshot, and an error from the reused ledger predicate.
13. **[requirements-completeness]** No security/authorization requirement despite expanding per-row data on a read-only role's landing page — *"行渲染（CN、SANS 折叠计数、issuer、daysLeft、托管、引用云、探测徽标）"* plus fingerprint-based sorting — must state whether read-only roles may see per-certificate fingerprints and full SAN lists.
14. **[requirements-completeness]** Performance invariant extrapolated from one near-1:1 tenant — *"（同量级，当前 jlc 数据 94 证 / 93 域）"* — must state the cert:domain ratio at which the assumption fails and whether pagination/virtualization is required.
15. **[scope-definition]** A required code change sits inside the Out of Scope list — *"证书循环直出时通配符跳过计数按 domain 去重，防止共享通配符 SAN 的多张证书重复计数"* — must move to In Scope (Risk#8 already assigns it a unit test).
16. **[scope-definition]** A gated behavior change has no In-Scope entry — *"由现状域名字典序**改为**到期优先（notAfter asc，是变更而非维持）"* gated by SC-7 — must appear as a backend deliverable.
17. **[scope-definition]** A new UI control is named without specification — *"孤儿默认隐藏 + 切换入口 + 空态处理"* — must define placement, label, persistence, and behavior under an active level filter.
18. **[risk-assessment]** Self-serving rating on the one production-visible item — *"`GET /certs/dashboard` 破坏性变更漏网调用者 | L | L"* — impact of a missed consumer is a broken page; must re-rate and state the detection mechanism beyond a one-time grep.
19. **[risk-assessment]** Two exposures the document itself creates have no risk row — the deploy window that violates its own *"孤儿隐藏由服务端执行且不静默"* guarantee, and non-1:1 tenants — must add rows with mitigations.
20. **[logical-consistency]** The solution retains the mismatch the Problem defines as the defect — *"`diffAlertCount`/`exemptCount` 保持域名级计数"* with *"一个 diff 域名挂 2 张证 → 卡计数 1、行数 2"* vs *"两个页面计数对不上……用户信任损耗越大"* — must specify the user-facing explanation of the 1→2 discrepancy (label, tooltip, or count footnote), not only the test-level mapping.
21. **[success-criteria]** SC-2 forbids and then commits the same sin in one clause — *"（动态相等，不写死数值；"已过期"档以实测 11=11 断言）"* — must express the gate as `rows.length === countsByLevel.expired.total` and drop the environment-coupled literal.
22. **[success-criteria]** Coverage gaps against promised work — no gate for the `wildcardSkippedCount` domain-dedup (only Risk#8's *"单测构造两张共享 `*.example.com` 的证书断言计数为 1"*), none for the toggle, and none asserting the Out-of-Scope invariants (3 coverage cards + domain-level special counts) survive the refactor — must add regression gates.
23. **[blindspot]** *Gate brittleness from literal UI copy in acceptance criteria.* The rubric scores whether SC are measurable, not whether they are maintainable. SC-4 pins a gate to an exact Chinese string — *"过期旧证行的"一致"徽标文案为「该证 SAN 现网一致（生效归属另一证书）」"* — so any copy review, wording polish, or future i18n turns a green gate red for a non-defect. Must assert the semantic distinction via a stable key/test-id rather than the rendered string.
24. **[blindspot]** *No observability or post-ship verification plan.* The rubric covers risks and success criteria but not how the change is confirmed healthy in production. The document's verification stops at *"全部门禁绿：e-cam-service `go test ./...` … `vitest` + `vue-tsc` 通过"* — all pre-merge. For a change that silently hides rows server-side, there is no logged `hiddenCount`, no metric, and no way to detect that the dashboard is under-reporting after deploy. Must add a server-side log/metric for the hidden set size.
25. **[blindspot]** *Decision-reversal precedent is set without recording why the original was accepted.* The proposal overturns a deliberate prior decision — *"（commit 7cdd488 决策注释自陈）"* … *"跳台账是当时对粒度矛盾的妥协，非设计目标"* — but never checks with the original decider whether a constraint behind that compromise still holds (the doc only re-reads the comment). Must confirm no unrecorded constraint drove 7cdd488, and plan to replace that code comment so the next reader does not re-litigate this.
26. **[blindspot]** *Next Steps authorizes automated execution before any approval gate.* — *"经 `/quick` 管线生成任务并执行（intent=enhancement，auto.runTasks.quick=true 自动串行）"* — a proposal carrying `status: Draft` with an unresolved breaking-API deploy order should not hand itself to an auto-serial runner. Must gate execution on resolution of attacks 1 and 2 and on an explicit approval, and state who approves.
