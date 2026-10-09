# Scorer Report — iteration-2

- Proposal: `D:\Haven\docs\proposals\dashboard-cert-granularity\proposal.md`
- Rubric: `forge/3.0.2/skills/eval/rubrics/proposal.md` (1000 pts, target 900)
- Scorer stance: adversarial (CTO persona), verification stance: code claims not substantiated by the document itself count as findings.

---

## Phase 1 — Reasoning Audit

**Chain**: Problem (cert-granularity card counts vs domain-granularity table → stale certs masked) → Solution (re-cast table to cert granularity, in-page filter, server-side orphan hiding, dual-count API) → Evidence (code refs, commit 7cdd488 self-comment, user complaint quote) → Success Criteria (8 testable items). The chain is intact: the solution directly eliminates the named root cause (domain dedup), and SC #1 asserts the non-dedup property.

**Self-contradiction check**: The proposal rejects 方案 B partly because it leaves "卡计数与抽屉行数两来源" (two sources for counts), yet the design intentionally keeps diff/exempt special-card counts at domain level so "卡值小于筛出行数…为预期而非缺陷". This is NOT a reintroduction of the eliminated problem — the doc explicitly re-labels the cards 「域名计数」and reframes the invariant as set-mapping closure (SC #6), not numeric equality. Handled; noted as tension but not scored as contradiction.

**SC Consistency Deep-Dive (clustered by affected area)**

- Cluster `GET /certs/dashboard` (SC#1, SC#3, SC#4, In Scope backend item): SC#1 asserts an `includeHidden=true` query parameter ("`GET /certs/dashboard` 返回证书行：`includeHidden=true` 时 items 长度 == 全部证书数"). This parameter appears **nowhere** in Proposed Solution, In Scope, Compatibility, or Risks — the API contract elsewhere says only "API 返回双口径计数（全局 hiddenCount + countsByLevel 每档 visible/hidden）" and "前端只消费 visible 行 + hiddenCount". Whether hiding is unconditional-server-side or toggle-driven via a parameter is therefore **ambiguous — requires author clarification**. Internally the SC set is satisfiable (SC#3's "开启后行数 == visible + hidden" implies the toggle passes includeHidden=true), but the parameter is defined only inside a success criterion, not in the design it verifies.
- Cluster `OverviewCards` (SC#2, SC#6, In Scope card items): "rows.length === countsByLevel[tier].total" requires countsByLevel[tier].total = visible+hidden. Key Scenario states "countsByLevel 含孤儿，否则行数必小于卡计数", so total includes orphans — satisfiable with SC#3's {visible, hidden} payload if total = visible+hidden (implicit but coherent).
- Cluster `DashboardTable` (SC#3 toggle state machine, SC#5 badge, SC#7 sort, In Scope table items): mutually satisfiable. Toggle "分级卡激活时强制开启并禁用" is redundant with the total-rule exemption (both achieve orphan visibility during filtering) — over-determined but not contradictory.
- Coverage-card cluster: the 3 coverage cards (`registrationRate`/`replaceableRate`/`fingerprintOnlyRate`) are declared out of scope for their "域名级计数语义", but their **click behavior** under the new in-page-filter model is unspecified — are they filter triggers subject to the total rule or inert? **Ambiguous — requires author clarification.**
- SC#5 badge order matches Key Scenario line 45 exactly (incl. "未探测 SAN 排序尾") — consistent.

## Phase 2 — Rubric Scoring

### 1. Problem Definition — 102/110
- Problem stated clearly: 40/40. The granularity mismatch is stated with precision and the failure mode ("过期/临期旧证被同名新证掩盖") is unambiguous.
- Evidence provided: 36/40. Code refs, a commit self-quote, and a user complaint are cited. Deduction: the pivotal behavioral claim "用户点击卡片的第一反应是**就地看到该档具体证书**（依据见 Assumptions Challenged 表）" is backed only by the team's own 5-Whys exercise, not by user data; the only direct user evidence is a single complaint about empty filtering.
- Urgency justified: 26/30. Cost-of-delay is concrete ("看板与刚验收的台账「已过期」计数立即对不上"). Deduction: the doc does not quantify the current mismatch (how many orphans exist today in jlc), so the size of the trust-loss window is asserted, not measured.

### 2. Solution Clarity — 109/120
- Approach is concrete: 38/40. A reader can restate the design. Deduction: the `includeHidden=true` parameter exists only in SC#1 and is never specified in the API contract ("API 返回双口径计数（全局 hiddenCount + countsByLevel 每档 visible/hidden）").
- User-facing behavior described: 38/45. Happy path, toggle states, aria-pressed, empty state are all given. Deductions: (a) the "一致" badge attribution problem is acknowledged ("须区分「本证线上生效」与「生效归属另一证书」") but the observable display semantics — what the user actually sees on a stale-cert row — are never defined; only "以稳定键区分归属语义" in the risk table, deferring the user-visible design; (b) coverage-card click behavior unspecified.
- Technical direction clear: 33/35. Files, reuse list, and payload shape are named; "可被看板导入且无 import cycle" is asserted without derivation, but the direction is clear.

### 3. Industry Benchmarking — 102/120
- Industry solutions referenced: 36/40. cert-manager and AWS ACM are named, but the dedicated "Industry Solutions" section is a single unattributed sentence ("证书到期仪表盘主流形态：证书行主表 + 分级计数卡 + 到期筛选") with no source.
- At least 3 meaningful alternatives: 24/30. Exactly three are given (do-nothing, 方案 B, 方案 A). Do-nothing is required, B is genuine — so the field is effectively **one** real alternative beyond status quo. The rejected 方案 B is characterized in two clauses ("纯前端小改"/"不消除根因") without exploring its actual trade-off envelope (e.g., hybrid: cert-granularity table without orphan hiding, or pagination/virtualization-first, which the risk table itself mentions as a fallback but never evaluates as an alternative).
- Honest trade-off comparison: 20/25. 方案 A's own con ("后端+前端中等改造") is stated; but 方案 B's cons are dismissive one-liners rather than an explored comparison.
- Chosen approach justified against benchmarks: 22/25. Rationale ("唯一消除同名掩盖根因") is clear and benchmark-aligned (adopts the cert-manager standard rather than inventing).

### 4. Requirements Completeness — 101/110
- Scenario coverage: 38/40. Happy path, 同名多证, probe-conflict, special cards, orphans, error path all present. Minor: error path covers probe-data-missing only; no API-failure/loading degradation is described beyond inherited skeleton screens.
- NFRs: 35/40. Performance (batch in-memory three-state evaluation, "逐证 N+1 为反例"), sorting change, accessibility, compatibility, read-only boundary are all called out. Deduction: performance target is qualitative — no row-count or latency budget is stated for the new table (the tenant-scale risk defers entirely: "先加行数日志再决定实现").
- Constraints & dependencies: 28/30. Dependency on ledger-hide's three-state judgment with single-source mandate is explicit; minor deduction for the unsubstantiated "无 import cycle" claim.

### 5. Solution Creativity — 78/100
- Novelty over industry baseline: 28/40. The core move is self-admittedly the industry baseline ("直接采用该行业的基础范式：主表对齐真实业务实体（证书）粒度"). Differentiation is the orphan-hiding isomorphism with the ledger, which is consistency engineering, not novelty.
- Cross-domain inspiration: 28/35. CI-badge / monitoring "worst-first" reducer is a genuine cross-domain borrow, well-applied to the 1:N probe-to-cert attachment problem.
- Simplicity of insight: 22/25. The core insight ("根因就是粒度错配") is elegant; deduction: the surrounding machinery (dual-count payload, toggle state machine with forced-open-and-disabled, total-rule exemption) is heavier than the insight — two mechanisms (exemption + forced toggle) achieve the same visibility outcome.

### 6. Feasibility — 88/100
- Technical feasibility: 38/40. Concrete three-file backend / four-file frontend surface; "无新依赖、无 showstopper". Deduction: "可被看板导入且无 import cycle" is asserted without derivation (package topology not shown).
- Resource & timeline: 22/30. "约 1 个 feature 周期（2~3 任务）" is quantified but thin on bandwidth assumptions. Deduction: vague unquantified claim "探测管线本轮验证稳定" — stable by what measure? No metric, test, or observation is cited.
- Dependency readiness: 28/30. ledger-hide merge commits (1f4ecd5、6393454、fbf31aa) are cited; minor deduction because "探测管线本轮验证稳定" is doing dependency-readiness work while being unverified.

### 7. Scope Definition — 75/80
- In-scope items concrete: 28/30. Each item names files, behaviors, and predicates ("谓词严格为 expired AND `no_refs_scanned`"). Minor: coverage-card click behavior leaves one UI surface undefined.
- Out-of-scope explicitly listed: 24/25. Five named exclusions including count-semantics preservation.
- Scope bounded: 23/25. Bounded to 2~3 tasks; minor deduction: the tenant-scale mitigation ("超阈值租户需分页/虚拟化，先加行数日志再决定实现") defers a potentially scope-expanding implementation decision past approval.

### 8. Risk Assessment — 83/90
- Risks identified: 29/30. Ten risks including non-obvious ones (row-key breakage, wildcardSkippedCount double counting, hidden-set observability).
- Likelihood + impact rated: 27/30. Varied and honest (two H-likelihood entries; includes L×L and L×M). Minor: two H-likelihood items (multi-SAN ambiguity, row-key breakage) suggest design fragility that is mitigated rather than simplified away.
- Mitigations actionable: 27/30. Most are test-anchored ("单测断言两证计数 1"). Deduction: the deploy/rollback mitigation "前后端同一 release/PR 协同上线" asserts a release discipline with no verification step for "grep 全仓 `getCertDashboardApi` 核验单消费端" beyond the grep itself (no CI enforcement named).

### 9. Success Criteria — 72/80
- Measurable and testable: 29/30. Relation assertions, grep gates, and state-machine behavior tests are objectively verifiable.
- Coverage complete: 20/25. Gaps: (a) cloud-filter-option derivation ("云选项来源改为「当前可见行」派生…隐藏态下云选项不含孤儿云") lives only in Risks — no SC asserts it; (b) `includeHidden=true` (SC#1) has no corresponding In-Scope deliverable defining it; (c) tooltip conflict-SAN detail has no explicit SC.
- SC internal consistency: 23/25. The set is satisfiable as a whole (verified in Phase 1); minor deduction for the redundant toggle-forcing vs exemption rule and the implicit definition of `countsByLevel[tier].total` (= visible+hidden, never stated).

### 10. Logical Consistency — 79/90
- Solution addresses the stated problem: 33/35. Directly removes the domain-dedup root cause; special-card count divergence is explicitly reframed, not ignored.
- Scope ↔ Solution ↔ SC aligned: 24/30. The `includeHidden=true` parameter appears only in SC#1 and nowhere in the Solution or In Scope API contract ("items 为破坏性变更；单消费端…同步改造" describes the change but never the parameter) — a cross-section misalignment between the acceptance gate and the described design.
- Requirements ↔ Solution coherent: 22/25. Orphan requirement maps to hiding mechanism; probe requirement maps to badge aggregation. Deduction: coverage-card click behavior is an orphaned surface — In Scope names their counts as out of scope but the solution's in-page-filter model gives them no defined interaction; and the header claim "计数与明细天然闭合" holds only for tier cards, a qualifier the doc applies only later (「域名计数」subtitle), not in Innovation Highlights.

## Phase 3 — Blindspot Hunt

1. [blindspot]: **Loss of deep-linkability is never addressed.** The design removes URL navigation ("点击改为页内筛选（`router.push` 移除）") but never discusses that the previous jump-to-ledger produced a shareable/bookmarkable filtered URL, and the new in-page filter state is not stated to sync to the query string. For a read-only default page used for triage, losing URL-reproducible filtered views is a real regression that no risk, NFR, or SC covers. Must improve: either document filter-state-to-URL sync or explicitly accept and record the deep-link loss.
2. [blindspot]: **"查看全部" toggle is "不持久化" — refresh resets the operator's context.** Quote: "孤儿默认隐藏 + 切换入口（工具栏内联横幅+开关，同台账模式，不持久化）". For an on-call operator reconciling counts, a page refresh silently returns to the hidden view; no risk row covers accidental context loss. Must improve: justify non-persistence against the reconciliation use case, or persist per-user.
3. [blindspot]: **No data-freshness/consistency story between countsByLevel and items within one response.** The dual-count contract ("全局 hiddenCount + countsByLevel 每档 visible/hidden") plus client-side consumption ("前端只消费 visible 行 + hiddenCount，不派生三态") creates two independently-consumed figures in one payload; nothing states they are computed in the same transaction/snapshot, and no SC locks countsByLevel total == len(items under includeHidden). Minor, but the doc's whole premise is count⇔rows closure. Must improve: one snapshot/derivation assertion in SC#1 or #3.
4. [blindspot]: **Commit-hash evidence is uncorroborated within the document.** Dependency readiness rests on "ledger-hide 已合入 main（1f4ecd5、6393454、fbf31aa）" and the problem rests on "7cdd488 注释自陈" — the document offers no excerpt or test name from these commits beyond the one quote it does include; under verification stance these are asserted, not shown. Must improve: inline the decisive ledger-hide helper signature once, since the entire single-source mandate ("必须复用同一实现，不重写") hinges on its reusability.

## Bias Detection Report

- 12 quoted findings total; 8 target unannotated regions, 4 target `<!-- pre-revised -->`-marked regions. Unannotated:annotated attack ratio ≈ 2:1.
- The severity labels did not skew attention: the two highest-impact findings (`includeHidden=true` orphan parameter, deep-link loss) both sit in **unannotated** text, and no pre-revised:high region produced a new-issue finding beyond the coverage-card ambiguity adjacent to it. Annotated regions read as tighter than the surrounding unannotated prose (e.g., SC#2–#6 are mechanically tighter than SC#1).
- No finding contradicts a revision direction; no `conflict-with-pre-revision` tags issued.

## Cross-Dimension Coherence Check

Scores are mutually consistent: the document is strongest where it was revised (SCs, risk table, in-scope predicates) and weakest in the untouched seams between sections (API contract ↔ SC#1 parameter, industry section thinness, unaddressed UX regressions from removing navigation). No dimension's deduction depends on another's generosity; the 889 total reflects real integration gaps rather than isolated nits.

**Total: 889/1000** (target 900 — not met; primary blockers: includeHidden contract gap, SC coverage gaps, thin alternative field).
