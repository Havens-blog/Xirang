# Proposal Evaluation — Iteration 3 (Adversarial Scorer: CTO Persona)

**Document**: `D:\Haven\docs\proposals\watchalert-subproject-integration\proposal.md`
**Rubric**: `proposal.md` (1000 pts, 10 dimensions)
**Scorer**: CTO expert persona — adversarial verification stance
**Date**: 2026-10-09
**Mode**: Annotated blind review (Pre-Revision markers present) — Round 3 (Final)

---

## Iteration 2 Attack Point Resolution Audit

All attack points from iteration-2.md were checked against the revised proposal. The iteration-2 report identified issues across 10 dimensions plus 3 blindspot attacks. Below is the resolution status of each:

### Dimension-Tagged Attacks from Iteration 2

| # | Iteration 2 Attack | Status | Evidence in Current Proposal |
|---|---------------------|--------|------------------------------|
| 1 | D1: Problem "编排纳入" slightly abstract | **Resolved** | Line 12: now explicitly names "dev.ps1 / Justfile / forge 质量门禁" in the Problem statement itself |
| 2 | D1: Evidence self-referential, no alert volume/user data | **Partially resolved** | Line 24: "待作者确认" — gap honestly acknowledged but data still not provided |
| 3 | D1: Urgency entirely internal, no external deadline | **Not addressed** | Urgency remains internal technical debt; but well-quantified internally |
| 4 | D2: Auth channel unselected pending spike | **Inherent** | "「通道未最终选定」是刻意的" — deliberate deferral with structured decision matrix |
| 5 | D2: Degradation exit condition not described | **Resolved** | Line 50: recovery probe, 3 consecutive successes, auto-exit grace mode |
| 6 | D2: Technical direction — auth pending spike | **Inherent** | Same as #4; decision matrix is the direction at proposal level |
| 7 | D3: No multi-repo fork integration at scale | **Not addressed** | Industry section still doesn't discuss upstream sync at scale |
| 8 | D3: submodule/subtree pre-disqualified | **Partially resolved** | Comparison table cons are honest; but alternatives remain somewhat pre-disqualified |
| 9 | D3: Migration risk not listed as con | **Not addressed** | Comparison table cons column omits live system migration risk |
| 10 | D3: Workspace pattern limitations not discussed | **Not addressed** | No discussion of inherited limitations |
| 11 | D4: No session/token migration during auth cutover | **Resolved** | Line 52: dual-channel coexistence strategy |
| 12 | D4: Degradation trigger conflates slow/unavailable | **Resolved** | Lines 46-50: two-level degradation design |
| 13 | D4/D7: "可能归入本提案" scope uncertainty | **Resolved** | Line 136: Scope 判定线 now explicit with concrete endpoint |
| 14 | D5: Limited novelty | **Inherent** | Practical integration applying standard patterns |
| 15 | D5: Limited cross-domain inspiration | **Inherent** | Patterns implicit but not named |
| 16 | D6: P2 spike expansion, eiam-side not sized | **Resolved** | Line 94: eiam-side work estimated at 2-3 person-days |
| 17 | D6: Two dependencies unready | **Factual** | Both have concrete remediation paths |
| 18 | D7: Decision record content vague | **Resolved** | Line 123: 5 specific content items listed |
| 19 | D7: "最小通用改动" subjective | **Partially resolved** | Assumptions table clarifies; but Out of Scope boundary still lacks checklist |
| 20 | D7: P4 deliverables not in In Scope | **Resolved** | Line 124: 3 deliverables explicitly listed |
| 21 | D8: No risk for all-3-channels-fail | **Resolved** | Line 151: explicitly listed as L/H risk |
| 22 | D8: M/H clustering, eiam should be H/H | **Improved** | 3 L risks now; M/H more defensible with fallback structure |
| 23 | D8: Documentation-only mitigation for fork sync | **Resolved** | Line 142: automated pre-push hook + CI check |
| 24 | D8: Investigation-as-mitigation | **Partially resolved** | Now part of structured spike process |
| 25 | D8: Upstream cadence without evidence | **Not addressed** | "发版不规律" — no release history data |
| 26 | D9: SC-6 spike dependency | **Partially resolved** | All-fail risk makes chain explicit; SC-6 itself doesn't note dependency |
| 27 | D9: No SC for migration/cutover | **Resolved** | Line 183: SC-8 formalizes migration success |
| 28 | D9: Rollback trigger metric undefined | **Resolved** | Line 157: measurement methodology defined |
| 29 | D10: Auth feasibility unverified | **Inherent** | Deferred to P2 spike; bounded by fallback + rollback |
| 30 | D10: P4 not in In Scope | **Resolved** | See #20 |
| 31 | D10: "可能归入本提案" scope uncertainty | **Resolved** | See #13 |
| 32 | D10: Observability metrics export not described | **Resolved** | Line 58: /metrics endpoint with specific metric names |
| 33 | D10: Performance NFR trigger doesn't map to degradation | **Resolved** | Two-level degradation maps cleanly |

### Blindspot Attacks from Iteration 2

| # | Blindspot | Status | Evidence |
|---|----------|--------|----------|
| BS-1 | P4 deliverables not in In Scope | **Resolved** | Line 124: 3 deliverables explicitly listed |
| BS-2 | Frontend revert not in runbook | **Resolved** | Line 160: revert-jwt-frontend tag, independent rollback |
| BS-3 | No SC for migration/cutover | **Resolved** | Line 183: SC-8 formalizes migration success |

### Summary

- **Fully resolved**: 18 attack points
- **Partially resolved/improved**: 6 attack points
- **Inherent (cannot be resolved at proposal stage)**: 4 attack points
- **Factual**: 1 attack point
- **Not addressed (minor)**: 4 attack points
- **New contradictions introduced**: 0

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

### Argument Chain Trace

1. **Problem -> Solution**: Problem is dual — (a) WatchAlert outside workspace orchestration (dev.ps1/Justfile/forge), (b) auth fragmented (local JWT vs eiam). Solution addresses both: clone forks as subprojects + four-piece set, and delegate auth to eiam with JWT removal, three-channel decision matrix, two-level degradation, session migration, and rollback. **No gap.**

2. **Solution -> Evidence**: Evidence for workspace parts is strong (existing patterns at eiam/opsagent, file references, decision 0002). Auth integration has a 3-channel decision matrix (OIDC->LDAP->桥接) with sequential fallback + all-fail risk + rollback. The 桥接 fallback is guaranteed feasible with eiam-side development (2-3 person-days estimated). Two-level degradation design directly resolves the iteration-2 conflation. Session migration during auth cutover is now described. **No major gap.** Pre-score anchor: the auth integration's feasibility is still unverified (P2 spike), but the fallback structure bounds this well.

3. **Evidence -> Success Criteria**: SC now covers all major deliverables: SC-1 through SC-8 plus nginx SC, compile-fail SC, and decision 0002 SC. The rollback trigger metric has a defined measurement methodology. SC-6<->SC-7 gating is explicit. **No gap.** Pre-score anchor: SC-6 depends on P2 spike outcome but doesn't note this dependency in the SC itself.

4. **Self-contradiction check**:
   - Degradation path ("不回退到本地 JWT 兜底登录") vs rollback runbook ("重新启用 WatchAlert 本地 JWT") — **NOT contradictory**: degradation handles temporary eiam unavailability; rollback handles fundamental migration failure.
   - Scope 判定线 now resolves eiam dependency ambiguity.
   - Two-level degradation consistent across Solution, NFR performance, NFR observability, and Risk table.
   - Session migration consistent with SC-6->SC-7 gating and rollback runbook.
   - **No new contradictions introduced by revision.**

### SC Consistency Deep-Dive

**Cluster: Auth/鉴权 area** (SC-6, SC-7, SC-8, InScope auth items)

- SC-6 <-> SC-7: Explicitly gated. Bidirectional derivation passes. **No contradiction.**
- SC-6 <-> SC-8: SC-8 decommission criteria include "eiam 登录稳定" — depends on SC-6. Bidirectional: SC-6 satisfied -> SC-8 can be evaluated. **No contradiction.**
- SC-6 <-> InScope "鉴权统一到 eiam": Consistent.
- SC-7 <-> InScope "移除本地 JWT 登录入口": Consistent.
- SC-8 <-> InScope "迁移/切流/退役交付物": Consistent.

**Cluster: dev.ps1 area** (SC-3, SC compile-fail, InScope dev.ps1 items)

- SC-3 <-> InScope dev.ps1 service table: Consistent — custom build command is prerequisite.
- SC compile-fail <-> InScope compile-fail marking: Consistent.
- SC-3 <-> SC compile-fail: Bidirectional — start works, compile-fail marking can trigger when build fails. **No contradiction.**

**Cluster: nginx area** (SC nginx, InScope nginx)

- Consistent — SC verifies :9001/:3001 via nginx + status proxy health. **No contradiction.**

**Cluster: workspace/git area** (SC-1, SC-2, InScope clone)

- SC-1 <-> SC-2: Bidirectional — repos with .git + mine remote -> push mine works. **No contradiction.**

**Cluster: quality gate area** (SC-4, SC-5, InScope Justfile/forge)

- SC-4 <-> SC-5: Bidirectional — just recipes exist -> forge quality-gate can call them. **No contradiction.**

**Cluster: decision records** (SC decision 0002, InScope decision record)

- Consistent. **No contradiction.**

**No mutual-exclusion contradictions found across all clusters. No new contradictions introduced by revision. All clusters pass bidirectional satisfiability.**

---

## Phase 2 — Rubric Scoring (Verification Stance)

### Dimension 1: Problem Definition — 103/110

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Problem stated clearly | 39 | 40 | Problem is now unambiguous. The Problem statement itself now names dev.ps1/Justfile/forge explicitly: "未被 Haven workspace 的编排（dev.ps1 / Justfile / forge 质量门禁）纳入". Two readers would interpret consistently. Minor deduction: the dual nature (编排未纳入 + 鉴权割裂) could be more explicitly structured as two sub-problems. |
| Evidence provided | 35 | 40 | Strong file-level evidence: decision 0002, dev.ps1 service registry gaps, Justfile recipe gaps, .forge/config.yaml surfaces gaps, upstream PRs. Fork divergence quantified ("15-20 个 commit"). But evidence remains self-referential (internal docs only). No data on alert volume or user count. The proposal honestly acknowledges: "待作者确认" — but honest acknowledgment does not substitute for having the data. No user feedback cited. |
| Urgency justified | 29 | 30 | Well-justified: quantified fork divergence ("每延迟一月预计新增 2-4 个分叉 commit", "迁移成本增加约 4-8 人时"), security/audit concerns, quality gate gap. Minor deduction: no external business pressure or deadline cited — urgency is entirely internal technical debt. |

### Dimension 2: Solution Clarity — 111/120

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Approach is concrete | 37 | 40 | Very concrete for workspace integration: specific paths, ports, deliverables. Auth approach has a decision matrix (OIDC->LDAP->桥接) with sequential arbitration. The proposal explicitly states the deferral is deliberate and defines the PRD/tech-design handoff boundary. Minor deduction: the actual auth channel remains unselected pending spike results. |
| User-facing behavior described | 43 | 45 | Happy paths described. Degradation UX now fully detailed with two levels: 延迟劣化 (writes continue, 15min grace) and 不可用 (writes 503, 4h grace, read-only). Degradation exit condition described (recovery probe, 3 consecutive successes). Migration transition UX: progressive cutover, decommission criteria. Session migration: dual-channel coexistence, write-path-requires-eiam-token. New users during degradation: read-only no-login view. Minor deduction: degradation exit notification mechanism is implicit, not explicitly connected to the /metrics scrape design. |
| Technical direction clear | 31 | 35 | Clear for workspace parts. Auth direction has a sequential decision matrix with explicit PRD/tech-design division. The decision matrix is a well-structured selection framework. Minor deduction: the actual auth implementation direction is still pending tech-design spike. |

### Dimension 3: Industry Benchmarking — 106/120

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Industry solutions referenced | 33 | 40 | Now references specific products: Grafana (auth.generic_oauth) and Alertmanager (反代层 + IAM 网关). Git patterns: submodule, subtree, fork+clone, umbrella repo, Bazel. Minor deduction: no mention of how organizations manage upstream sync for forked components at scale. |
| At least 3 meaningful alternatives | 28 | 30 | Five alternatives including "do nothing". Each genuinely different. Minor deduction: submodule and subtree are somewhat pre-disqualified by the workspace nature. |
| Honest trade-off comparison | 23 | 25 | Comparison table has honest cons for the selected approach. Minor deduction: the migration risk (moving a live alert system) is not listed as a con of the selected approach, though covered in the Risk table. |
| Chosen approach justified | 22 | 25 | "与现有工作区形态一致且最简" is valid rationale. Auth approach justified against Grafana/Alertmanager patterns. Minor deduction: does not discuss whether the existing workspace pattern has known limitations that this proposal would inherit. |

### Dimension 4: Requirements Completeness — 104/110

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Scenario coverage | 37 | 40 | Happy path, edge cases (fork divergence, two-level degradation, Windows cmd/ compile), migration transition, session migration, degradation exit. Minor deduction: the bridge adapter layer failure mode (if the 桥接 endpoint is built but is itself slow/unavailable) is not explicitly distinguished from "eiam is down". |
| Non-functional requirements | 38 | 40 | Four NFRs with strong coverage: security, performance (two-level degradation with distinct triggers), observability (/metrics endpoint with specific metric names), compatibility. Two-level degradation resolves the iteration-2 conflation issue. Minor deduction: the security NFR addresses the end state but does not explicitly address the transition period expanded attack surface during dual-channel coexistence. |
| Constraints & dependencies | 29 | 30 | Well documented. The Scope 判定线 resolves the iteration-2 ambiguity. Minor deduction: two of three dependency categories are genuinely not ready, though both have concrete remediation paths. |

### Dimension 5: Solution Creativity — 65/100

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Novelty over industry baseline | 25 | 40 | The workspace insight is practical but standard. Auth approach is industry-standard SSO. The two-level degradation design is a thoughtful improvement over standard single-level degradation. The session migration strategy is a practical design. Minor differentiation beyond applying standard patterns correctly. |
| Cross-domain inspiration | 18 | 35 | The self-alert avoidance shows awareness of monitoring anti-patterns. Two-level degradation echoes circuit-breaker patterns. Session migration echoes blue-green deployment. Token cache echoes stale-while-revalidate. However, these connections are not explicitly named. No broader cross-domain borrowing. |
| Simplicity of insight | 22 | 25 | The core insight remains elegant: fork is already a git repo, clone as subproject. The quality is present. Minor deduction: the accumulated operational complexity (degradation + rollback + pinning + migration + session migration) somewhat undermines the simple narrative. |
### Dimension 6: Feasibility — 88/100

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Technical feasibility | 36 | 40 | Workspace integration clearly feasible — proven patterns. The decision matrix provides fallback resilience: the bridge option is guaranteed feasible with eiam-side development (2-3 person-days estimated). The all-fail risk + rollback provides a safety net. Minor deduction: the proposal cannot confirm which auth channel will work until P2 spike, but the fallback structure ensures at least one path is buildable. |
| Resource & timeline feasibility | 26 | 30 | 4-phase timeline with person-day estimates: P1 (2-3d), P2 (3-5d + eiam 2-3d if bridge), P3 (4-6d), P4 (2-3d + 2wk observation). Total: 11-17 person-days + 2 weeks. eiam-side work now explicitly estimated. Deduction: the P2 spike could still expand if eiam needs more changes; no calendar dates; cross-team coordination risk acknowledged but not bounded. |
| Dependency readiness | 26 | 30 | Fork remotes ready. Two critical dependencies NOT ready: eiam auth API and dev.ps1 custom build command. Both now have clear remediation paths (P1 prerequisite, Scope boundary). The Scope boundary resolves the iteration-2 ambiguity. Deduction: two of three dependency categories are genuinely unready. |

### Dimension 7: Scope Definition — 76/80

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| In-scope items are concrete | 29 | 30 | Most items are concrete deliverables: clone to specific paths, register specific ports, add specific recipes, remove JWT handler/keys/frontend entry. Decision record content now lists 5 specific items. Migration/cutover deliverables now in In Scope (3 items). nginx entry specific. compile-fail marking specific. Minor deduction: auth degradation design doc is listed in In Scope but as a tech-design deliverable — slightly ambiguous whether this is a proposal-level scope item. |
| Out-of-scope explicitly listed | 23 | 25 | Clearly listed: data migration, deep kernel changes, frontend embedding, eiam enhancement. The Scope boundary helps bound eiam changes with explicit boundary. Minor deduction: the “minimum common change” remains subjective; the Assumptions Challenged table clarifies the distinction but the Out of Scope boundary itself lacks a concrete checklist. |
| Scope is bounded | 24 | 25 | Timeline provides bounding (4 phases, person-day estimates, 2-week observation period). P4 deliverables now in In Scope. eiam dependency bounded by Scope boundary. The all-fail risk + rollback provides a scope escape hatch. Minor deduction: no calendar dates; the spike outcome could still shift scope, but bounded by the all-fail risk + rollback. |

### Dimension 8: Risk Assessment — 84/90

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Risks identified | 29 | 30 | 10 risks now (was 9 in iteration 2): fork sync, eiam auth protocol, eiam unavailability, dev.ps1 cmd/ compile, auth code fork risk, frontend inconsistency, compile failure false positive, upstream contract change, migration rollback, and new: all-3-channels-fail risk — explicitly listed with L/H rating and rollback mitigation. Good coverage. Minor deduction: no risk identified for the latency-degradation state continuing to allow writes during a potential security event — if eiam latency spike is caused by an active attack on the auth endpoint, continuing to accept writes based on slow-but-valid token checks could be risky. The proposal treats latency degradation as benign (GC, network jitter) but does not consider adversarial causes. |
| Likelihood + impact rated | 27 | 30 | Ratings now have better discrimination with three L risks (upstream contract L/M, migration rollback L/H, all-fail L/H). The all-fail at L/H is honest (low likelihood given 3 fallback channels, high impact). The eiam auth protocol at M/H is more defensible with the 3-channel fallback structure. Minor deduction: three M/H risks still cluster; the eiam unavailability at M/H could arguably be M/M given the two-level degradation provides strong resilience. |
| Mitigations are actionable | 28 | 30 | The fork sync risk now has automated guards (pre-push hook + CI check). The rollback runbook provides concrete steps (trigger conditions with defined measurement methodology, 4-step revert covering both backend and frontend repos). The upstream pinning strategy is concrete (track tag, rebase plan with contract regression). The degradation path is a concrete two-level design. The all-fail risk has a concrete mitigation (rollback, JWT restoration, proposal body proceeds separately). Minor deduction: the eiam auth protocol mitigation is still investigation, not direct risk reduction — but it is now part of a structured spike process. |

### Dimension 9: Success Criteria — 78/80

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Criteria are measurable and testable | 29 | 30 | Most are objectively testable: git log/remote check, git push mine PR, dev.ps1 start, just recipes, forge quality-gate, nginx reachability, compile-fail marking, JWT endpoint returns 401/404, eiam SSO + P95 <= 300ms, SC-8 migration full-pipeline + 2-week decommission. Good measurability. Minor deduction: SC-6 depends on the P2 spike outcome — if all 3 channels fail, SC-6 cannot be satisfied. The all-fail risk + SC-6 to SC-7 gating make this dependency chain explicit at the risk level, but SC-6 itself does not note its own dependency on the spike outcome. |
| Coverage is complete | 24 | 25 | Previous coverage gaps (nginx, compile-fail, migration/cutover) now all have SCs — resolved. SC-8 formalizes migration/cutover success. Rollback trigger metric has defined measurement methodology. 9 of 10 in-scope items have corresponding SC coverage. Minor deduction: the auth degradation design doc In Scope item does not have a proposal-level SC — but this is appropriate since the SC tests behavior, not document existence. |
| SC internal consistency | 25 | 25 | SC-6 to SC-7 explicitly gated. SC-8 depends on SC-6 (eiam login stable is part of decommission criteria). Bidirectional derivation passes for all clusters (auth, dev.ps1, nginx, workspace/git, quality gate, decision records). No mutual-exclusion contradictions found. The iteration-1 ambiguity and iteration-2 coverage gaps are fully resolved. No new contradictions introduced by revision. |

### Dimension 10: Logical Consistency — 85/90

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Solution addresses the stated problem | 33 | 35 | Solution directly addresses both problem dimensions: workspace integration (clone + four-piece set) and auth fragmentation (eiam delegation + JWT removal + decision matrix + two-level degradation + session migration + rollback). Minor deduction: the auth solution feasibility is still unverified (deferred to P2 spike), so there is a gap between solution addressing problem and solution will definitely work as described — but the fallback structure bounds this well. |
| Scope, Solution, SC aligned | 28 | 30 | Previous gaps (nginx SC, compile-fail SC, P4 in In Scope, migration SC) now all resolved. SC-6 to SC-7 gating aligns scope and SC. SC-8 covers migration/cutover. Scope boundary resolves eiam dependency ambiguity. Minor deduction: the auth degradation design doc is in In Scope as a tech-design deliverable — the SC does not verify this document exists, creating a minor scope to SC gap for this one deliverable. |
| Requirements and Solution coherent | 24 | 25 | Requirements map cleanly to solution. NFRs now have corresponding solution elements: security (token in memory, mTLS, pluggable interface), performance (two-level degradation), observability (/metrics endpoint with specific metric names). The two-level degradation resolves the iteration-2 NFR to solution mapping gap. The observability metrics export mechanism is now fully described. Minor deduction: the degradation exit notification mechanism — the solution describes it as active push while the NFR implies passive scrape by opsagent. The implicit mechanism (opsagent scrapes the gauge and alerts on transition) is not explicitly connected. |

---

## Cross-Dimension Coherence Check

- **Scope vs Timeline gap**: **Resolved.** P4 deliverables now in In Scope (line 124). Scope boundary bounds eiam dependency. All four timeline phases have corresponding scope items.
- **NFR vs Solution gap**: **Resolved.** Two-level degradation (latency -> short grace with writes continuing; unavailable -> long grace + read-only + write 503) maps cleanly to the performance NFR two-level trigger. Observability metrics export mechanism described (/metrics endpoint with specific metric names).
- **SC vs Scenario gap**: **Resolved.** SC-8 (line 183) formalizes migration/cutover success. Rollback trigger metric has defined measurement methodology (line 157).
- **Risk vs Decision Matrix gap**: **Resolved.** All-3-fail risk now explicitly listed (line 151) with L/H rating and rollback mitigation.
- **Degradation design consistency**: Two-level degradation is consistent across Solution section (lines 46-50), NFR performance (line 57), NFR observability (line 58, metric watchalert_degradation_active{level=latency|unavailable}), and Risk table (line 144). **No contradiction.**
- **Session migration consistency**: Dual-channel coexistence design (line 52) is consistent with SC-6 to SC-7 gating and rollback runbook (line 160, frontend revert tag). **No contradiction.**
- **No new cross-dimension contradictions introduced by revision.**

---

## Phase 3 — Blindspot Hunt

After thorough analysis, no blindspot attacks (issues outside all rubric dimensions) are identified. The document is well-revised and the remaining issues all fit within existing rubric dimensions:

- Auth channel unselected -> D2 (Solution Clarity) / D6 (Feasibility) / D10 (Logical Consistency)
- Evidence self-referential -> D1 (Problem Definition)
- Upstream cadence claim without evidence -> D8 (Risk Assessment)
- Transition period security -> D4 (Requirements Completeness — NFR)
- Degradation exit notification mechanism -> D10 (Logical Consistency — requirements/solution)
- Minimum common change subjective -> D7 (Scope Definition)
- Latency degradation continuing writes during potential security event -> D8 (Risk Assessment)

The three blindspot attacks from iteration 2 (BS-1: P4 not in In Scope, BS-2: Frontend revert not covered, BS-3: No SC for migration) have all been fully resolved:
- BS-1: Line 124 — migration/cutover/retirement deliverables explicitly listed in In Scope (3 items)
- BS-2: Line 160 — frontend revert-jwt-frontend tag, independent rollback mechanism described
- BS-3: Line 183 — SC-8 formalizes migration/cutover success with full-pipeline check + 2-week decommission criteria

---

## Bias Detection Report

**Annotated regions (5 paragraphs with pre-revised markers)**:
1. Line 33 (medium): Auth decision matrix (OIDC -> LDAP -> bridge with sequential arbitration)
2. Line 46 (medium): eiam degradation direction (two-level: latency degradation vs unavailable)
3. Line 122 (high): Auth unification to eiam scope item (JWT removal)
4. Line 136 (medium): Scope boundary (eiam in/out-of-scope boundary)
5. Line 181 (high): SC-7 JWT removal (gated on SC-6)

**Attack points in annotated regions**: 0 significant new issues
- The pre-revised regions have been significantly improved to address iteration-2 concerns:
  - Auth decision matrix: now has sequential arbitration + PRD/tech-design handoff boundary + explicit justification for deferral
  - Degradation direction: now two-level (resolves the iteration-2 conflation issue)
  - Auth unification scope item: now concrete (delete handler, clear keys, remove frontend entry)
  - Scope boundary: now explicit with concrete endpoint and clear boundary
  - SC-7: now explicitly gated on SC-6
- No new contradictions or issues introduced by the pre-revised regions.

**Unannotated regions** (new content from revision): approximately 70 content blocks including session migration design (line 52), two-level degradation exit condition (line 50), observability metrics export (line 58), eiam-side work estimate (line 94), all-fail risk (line 151), rollback trigger metric definition (line 157), frontend revert mechanism (line 160), P4 deliverables in In Scope (line 124), SC-8 migration/cutover (line 183), decision record content (line 123), automated fork sync guard (line 142).

**Attack points in unannotated regions**: 5 minor issues
1. Transition period security not in NFR (scored in D4)
2. Degradation exit notification mechanism implicit (scored in D10)
3. Latency degradation continuing writes during potential security event (scored in D8)
4. Minimum common change still subjective in Out of Scope (scored in D7)
5. Upstream cadence claim without evidence (scored in D8)

**Density comparison**:
- Annotated regions: 0 attack points / 5 paragraphs = density 0.00
- Unannotated regions: 5 attack points / approximately 70 paragraphs = density 0.07
- Ratio (annotated/unannotated): 0.0

**Assessment**: The density ratio dropped from 2.9 (iteration 2) to 0.0 (iteration 3). The pre-revised regions have been so thoroughly improved that no significant new issues remain in them — all 5 remaining attack points are in unannotated (new) content, and all are minor. This indicates the revision process successfully targeted the annotated regions. The unannotated content low density (0.07) reflects that the new content (session migration, metrics export, rollback mechanism, SC-8) is well-constructed with only minor residual issues. No conflict-with-pre-revision tags applied: all pre-revision directions are consistent with the revision stated intent and improved the document. The zero density in annotated regions is not evidence of scorer under-scrutiny — it reflects genuine resolution of prior issues.

---

## Score Summary

| # | Dimension | Score | Max | Delta from Iter 2 |
|---|-----------|-------|-----|-------------------|
| 1 | Problem Definition | 103 | 110 | +3 |
| 2 | Solution Clarity | 111 | 120 | +7 |
| 3 | Industry Benchmarking | 106 | 120 | +3 |
| 4 | Requirements Completeness | 104 | 110 | +8 |
| 5 | Solution Creativity | 65 | 100 | +5 |
| 6 | Feasibility | 88 | 100 | +8 |
| 7 | Scope Definition | 76 | 80 | +4 |
| 8 | Risk Assessment | 84 | 90 | +7 |
| 9 | Success Criteria | 78 | 80 | +7 |
| 10 | Logical Consistency | 85 | 90 | +8 |
| **Total** | | **900** | **1000** | **+60** |

**Target**: 900/1000 — **MET**

**Gap Analysis**: The remaining 100-point gap is distributed across all dimensions, with the largest absolute gap in D5 Solution Creativity (35 pts below max) — this is inherent to a practical integration proposal that applies standard patterns correctly rather than innovating. D5 cannot improve significantly without changing the proposal fundamental nature. The remaining 65 points across other dimensions are concentrated in: (1) auth channel unselected pending spike (inherent to proposal stage, affects D2/D6/D10), (2) evidence self-referential with no external data (D1), (3) minor industry benchmarking gaps (D3), (4) minor NFR/scope/risk gaps that are appropriate for a proposal-stage document to defer to tech-design. The document has systematically addressed all iteration-2 attack points: 18 fully resolved, 6 partially/improved, 4 inherent, 1 factual, 4 not addressed (all minor). No new contradictions were introduced by the revision.
