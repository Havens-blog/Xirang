# Evaluation Report — dashboard-cert-granularity proposal (Iteration 4)

Scorer: Adversarial CTO persona. Rubric: proposal.md (1000 pts, target 900).
Document: `D:\Haven\docs\proposals\dashboard-cert-granularity\proposal.md`
No `<!-- pre-revised -->` markers found → plain scoring.

---

## Phase 1 — Reasoning Audit

**Problem → Solution → Evidence → SC trace**

- Problem: card counts are certificate-granular but the main table is domain-granular; same-name new cert masks expired old certs; workaround is jump-to-ledger.
- Solution: recast main table to certificate granularity, card click → in-page filter, worst-first row badges, server-side orphan hiding mirroring ledger.
- Evidence: code comment citation (`index.vue:158-165`, 7cdd488), static copy (`OverviewCards.vue:13`), one user feedback quote, adjacent ledger-hide feature.
- SC: 9 checkbox clusters, heavily test-anchored (unit + component + grep + relation assertions).

Trace is coherent: the solution removes the stated root cause (granularity mismatch), and the SCs verify closure (rows == counts).

**Self-contradiction check / SC Consistency Deep-Dive**

Cluster by affected area: (1) `GET /certs/dashboard` payload & hiding, (2) OverviewCards, (3) DashboardTable (rows, badges, toggle, sorting), (4) special cards, (5) tests/gates.

- CONTRADICTION (exemption rule): Key Scenarios states the master rule as "豁免仅由到期风险维度搜索触发", but the same sentence extends exemption to probe-state cards: "分级卡或 diff/exempt 特殊卡激活即豁免孤儿隐藏". diff/exempt are probe-dimension cards, not "到期风险维度". A reader cannot determine whether, e.g., an issuer or keyword search exempts orphan hiding. Ambiguous — requires author clarification.
- CONTRADICTION (hiding scope): In Scope says "孤儿仅在无筛选默认视图隐藏", yet Key Scenarios says attribute filters "不豁免，仅在当前可见行集上再过滤". Under a cloud filter, are orphans excluded (hiding active outside the no-filter view, contradicting the first clause) or visible (which is an exemption, contradicting "不豁免")? The two readings are mutually exclusive. Ambiguous — requires author clarification.
- SC↔SC internal: SC2 (`rows.length === countsByLevel[tier].total` under exemption refetch) and SC3 (default view excludes orphans; toggle-on rows == visible+hidden) are jointly satisfiable because countsByLevel totals include orphans and the exemption refetch uses `includeHidden=true`. No numeric contradiction found.
- Intra-SC drift: SC5 badge order ends "…> exempt > 健康" and omits 未探测 from the sequence, while Key Scenarios defines "…> exempt > 健康/未探测" plus "未探测 SAN 排序尾". Which tail element is last is unspecified in the SC.
- Shape drift: Proposed Solution says "countsByLevel 每档 visible/hidden"; In Scope says "{total, visible, hidden}"; SC3 asserts against `total`. Satisfiable but the payload contract is stated in two different shapes.
- Cross-section: the drawer's "SAN >20 折叠，默认最差优先前 20" appears only in SC5 — no corresponding In Scope or Solution item defines the 20 threshold (magic number introduced solely in Success Criteria). SC↔Scope misalignment.

---

## Phase 2 — Rubric Scoring

### 1. Problem Definition — 103/110

- Problem stated clearly: 40/40. Unambiguous; two readers would describe it identically (granularity mismatch, masking).
- Evidence: 36/40. File:line citations, commit reference, verbatim user feedback ("到期看板有的点击按分级过滤为空") are strong. Deduction: actual harm is not quantified — the document's own data ("当前 jlc 实测 94 证 / 93 域，+1 行") implies exactly one masked domain today; the "风险盲区" breadth ("同域已存在过期旧证（可能仍被引用或裸奔）") is asserted with no count of affected certificates.
- Urgency justified: 27/30. Divergence-with-ledger and read-only-default-page arguments are concrete; "不一致窗口越久，用户信任损耗越大" is qualitative with no measurable decay or trigger.

### 2. Solution Clarity — 114/120

- Approach concrete: 39/40. The one-paragraph solution is dense but decodable; specific files, params (`includeHidden`), and predicates (`expired AND no_refs_scanned`) are named.
- User-facing behavior: 43/45. Click→filter, badge reading, drawer drill-down, "可见 N · 隐藏 M" all described. Deduction: the forced-toggle state machine ("分级卡激活时强制开启并禁用，清除筛选恢复原开关态") is only fully specified inside SC3, not in the user-facing solution/scenario text — the observable behavior must be assembled from two sections.
- Technical direction: 32/35. "探测经 `probeByDomain[san]` 聚合" and the three-file backend change surface are clear. Deduction: "探测管线已有产出（见项目记录）" defers to an external, unnamed record — a reader of this document alone cannot verify it.

### 3. Industry Benchmarking — 111/120

- Industry solutions referenced: 37/40. cert-manager (`kubectl get certificate`) and AWS ACM console named with the relevant pattern (cert-as-row). No third-party reference exists for orphan handling — acknowledged honestly ("业界孤儿处理多为…罕见'无引用过期证默认隐藏'").
- ≥3 meaningful alternatives: 27/30. Four alternatives including do-nothing, each genuinely different (not straw men — 方案 B's "纯前端小改" pro is stated fairly). Deduction: only the chosen approach is industry-validated; the two non-trivial alternatives (B, and the no-hiding variant) are self-generated ("画布备选"), so the alternative set lacks an industry-sourced challenger.
- Honest trade-offs: 23/25. Costs admitted on both sides ("后端+前端中等改造", "代价是引入隐藏状态机"); the +1-row honesty is notable.
- Justified against benchmarks: 24/25. Explicit differentiation ("业界证书列表无此合并") and adoption rationale (align with cert-manager/ACM row model).

### 4. Requirements Completeness — 101/110

- Scenario coverage: 35/40. Happy path, four edge cases (same-name, probe conflict, special cards, orphans) and one error path. Deductions: (a) error path is one line — API failure, loading, and partial-probe-data degradation beyond "未探测" are unaddressed; (b) no scenario covers navigation after the ledger jump is removed (see blindspot 1): with "处置流程（仍由台账承载）", the user's path from dashboard to ledger is severed with no replacement described.
- NFRs: 38/40. Performance (5000-row threshold with named trigger), sorting, accessibility (aria-live total basis), compatibility (single consumer, URL deep-link), visibility boundary all present. Deduction: no data-freshness NFR — hiding decisions derive from "一次拉取 refs + 最新扫描快照", and stale-scan behavior (a referenced cert briefly appearing as `no_refs_scanned`) is unspecified.
- Constraints & dependencies: 28/30. ledger-hide semantics, shared-helper no-import-cycle constraint, reuse list all named. Deduction: "见项目记录" is an unresolvable external pointer within this document.

### 5. Solution Creativity — 90/100

- Novelty over baseline: 36/40. Differentiation from cert-manager/ACM articulated ("探测态与到期态并置于一张证书行——业界证书列表无此合并") — but this is a self-asserted claim with no cited counter-example search.
- Cross-domain inspiration: 32/35. "借鉴 CI 与监控面板的'最差优先'reducer" is a genuine, well-applied borrow.
- Simplicity of insight: 22/25. The core leap (fix granularity, counts close naturally) is elegant; but the surrounding machinery (dual counting, `includeHidden`, forced-toggle state machine, localStorage persistence, banner) has grown heavy relative to the insight — the comparison table itself concedes "代价是引入隐藏状态机".

### 6. Feasibility — 92/100

- Technical feasibility: 38/40. Three backend + four frontend touchpoints named; "无新依赖、无 showstopper" is plausible given reuse of ledger-hide's three-state logic and `LatestPerDomain`.
- Resource & timeline: 26/30. "约 1 个 feature 周期（2~3 任务…8 组测试门禁随对应任务内建）" — the estimate is asserted without breakdown, and it strains against the breadth the document itself enumerates: 10 risk mitigations, 9 SC gate clusters spanning go tests, vitest, vue-tsc, grep gates, and component tests. The "8 组" count does not match the enumerable SC/regression items.
- Dependency readiness: 28/30. ledger-hide merged with commit hashes ("1f4ecd5、6393454、fbf31aa") is verifiable-grade evidence; the probe-pipeline readiness rests on the vague "见项目记录".

### 7. Scope Definition — 77/80

- In-scope concrete: 29/30. Every item is a deliverable with file names, params, field lists, and UI behaviors.
- Out-of-scope explicit: 24/25. Named items including the coverage cards' semantic freeze; minor: "探测能力/巡检流程…（仅呈现方式调整）" boundary is stated but the drawer extension ("探测抽屉组件（扩展为证书多 SAN 视图）") arguably brushes against it without an explicit carve-out.
- Bounded: 24/25. 2–3 tasks within one feature cycle is bounded; residual openness comes from the deferred pagination trigger ("届时加装分页/虚拟化") being a promise without an owner or checkpoint.

### 8. Risk Assessment — 87/90

- Risks identified: 29/30. Ten risks, several non-trivial (row-key breakage rated H honestly; wildcard dedup; observability of hidden set). Missing: stale-scan risk — hiding a still-referenced cert because its reference scan is outdated is the failure mode of the whole hiding mechanism and is not listed.
- Likelihood + impact rated: 29/30. Honest spread (H/M/L mix, e.g., row-key risk H/M); not the defensive all-low pattern.
- Mitigations actionable: 29/30. Mostly executable ("grep 全仓 `getCertDashboardApi` 核验单消费端", "单测断言两证计数 1"); a few are deferrals ("届时加装分页/虚拟化，首期不实现分页") which are at least threshold-gated.

### 9. Success Criteria — 69/80

- Measurable/testable: 28/30. Nearly every SC is a test or relation assertion ("纯关系断言 `rows.length === countsByLevel[tier].total`"). Deduction: "SAN >20 折叠，默认最差优先前 20" is a threshold appearing only here, unexplained and unanchored to any requirement.
- Coverage complete: 22/25. All major in-scope items map to SCs. Gaps: empty-state UI ("空态处理", "空态提示 + '查看全部'入口") has no dedicated SC assertion; the toolbar banner itself (copy, behavior) is unspecified anywhere and thus untestable.
- Internal consistency: 19/25. Badge-order mismatch (SC5 omits 未探测 from the ordered sequence vs Key Scenarios' "健康/未探测" tail + "排序尾") and the countsByLevel shape drift ({visible,hidden} vs {total,visible,hidden}) are internal incoherences an implementer must guess at.

### 10. Logical Consistency — 80/90

- Solution addresses problem: 33/35. Directly removes the root cause; count⇔row closure is verifiable. Minor residue: the masking harm's present-day magnitude (+1 row per own data) vs the cost of the state machine is not weighed explicitly.
- Scope ↔ Solution ↔ SC aligned: 25/30. Two tagged contradictions live here: the 总规则 trigger wording ("豁免仅由到期风险维度搜索触发" vs diff/exempt exemption) and the hiding-scope clause ("孤儿仅在无筛选默认视图隐藏" vs "仅在当前可见行集上再过滤" under non-exempt attribute filters). Plus the SAN>20 threshold existing only in SC5.
- Requirements ↔ Solution coherent: 23/25. Clean mapping of special-card domain-count requirement, coverage-card freeze, read-only boundary; orphan requirement maps to hiding + counts. Small orphan-requirement residue: no requirement or solution element restores any dashboard→ledger path despite removing the only one.

**Cross-dimension coherence check**: The contradictions found in D9/D10 both trace to the same root — the 总规则 and hiding-scope semantics are stated three times (Key Scenarios, In Scope, SC3) in non-identical terms. Fixing the rule once upstream resolves D9-internal, D10, and the user-facing assembly gap in D2.

**Total: 103 + 114 + 111 + 101 + 90 + 92 + 77 + 87 + 69 + 80 = 924/1000**

---

## Phase 3 — Blindspot Hunt

1. **[blindspot] Navigation severance for the primary user group.** The document establishes "看板是**只读角色的默认页与唯一菜单入口**（`index.vue:75`）" and then removes the only cross-page path ("移除跳台账逻辑与静态文案"; "点击改为页内筛选…`router.push` 移除"), while declaring "变更单、发起更换等处置流程（仍由台账承载）". For a role whose only menu entry is the dashboard, deleting the ledger jump without adding any dashboard→ledger entry point cuts off the sole route to cert details and disposal workflows. The proposal specifies row-click → probe drawer but never a path to the ledger. Must improve: add an explicit navigation requirement (e.g., row context link to ledger cert detail) or demonstrate the read-only role has another ledger entry.

2. **[blindspot] Stale-reference-scan hiding hazard.** Hiding is derived from "一次拉取 refs + 最新扫描快照后内存归约" and the predicate "expired AND `no_refs_scanned`". Nothing addresses scan freshness: a certificate whose references were removed *or added* since the last scan is classified from stale data, so a still-referenced cert can be silently hidden (the exact "静默少行" the risk table fears from a different angle) until the next scan. Must improve: state the freshness contract (max scan age, re-scan trigger, or last-scanned timestamp surfaced in UI/observability).

3. **[blindspot] "First reaction" user-intent claim rests on author introspection.** "用户点击卡片的第一反应是**就地看到该档具体证书**（依据见 Assumptions Challenged 表）" — the cited basis is "5 Whys", i.e., the author's own reasoning chain, not user research; the only actual user feedback quoted ("到期看板有的点击按分级过滤为空") evidences a filter/count mismatch, not a preference for in-page filtering. Must improve: label the intent claim as hypothesis, or cite observed user behavior.

4. **[blindspot] LocalStorage persistence has no invalidation story.** "开关状态按用户持久化于 localStorage，对账工作流刷新不丢" — no mention of key namespacing per deployment/tenant or behavior when the persisted flag outlives a rollback to the pre-feature frontend (the risk table's own "回退窗口内…短暂显示旧域名行视图" scenario). Low severity but unstated.

---

## Verdict

924/1000 — above the 900 target. The document is unusually strong on evidence discipline, test-anchored success criteria, and honest trade-offs. The remaining distance to full marks is concentrated in: (a) the 总规则/hiding-scope semantics stated inconsistently across three sections, (b) the severed dashboard→ledger path for the read-only role, (c) stale-scan freshness of the hiding predicate, and (d) a handful of thresholds/magic numbers (SAN >20, "8 组测试门禁") introduced without anchors.
