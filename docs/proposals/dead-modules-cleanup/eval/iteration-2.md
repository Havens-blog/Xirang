# Proposal Eval — Iteration 2

- **Document**: D:\Haven\docs\proposals\dead-modules-cleanup\proposal.md (current revision, after Pre-Revision + two revision rounds)
- **Scorer persona**: CTO (approved/rejected hundreds of proposals)
- **Scale**: 1000 (target 900) | **Iteration**: 2 | **PREVIOUS_REPORT_PATH**: eval/iteration-1.md
- **Verification stance**: every new/changed assertion re-verified against D:\Haven\e-cam-service on 2026-09-09; carry-over claims spot-re-verified, not assumed.

## Prior Attack Disposition (iteration-1 → iteration-2)

| # | Iteration-1 attack | Status | Evidence |
|---|---|---|---|
| 1 | D2: line-41 benefit overstatement ("模块图与文档第一次与运行时事实一致") undercut by own deferral | **Fixed** | Rewritten as "模块图在 **import 级**与运行时事实一致…口径说明：live 代码中 10 处注释级悬空引用不在本轮…注释级的'文档与事实一致'待该轮完成后成立" — claim now correctly scoped to import level |
| 2 | D3: zero named industry tools/patterns | **Fixed** | staticcheck U1000, golang.org/x/tools/cmd/deadcode, golangci-lint unused, deprecate-then-delete — all named, correctly characterized, with explicit non-adoption rationale |
| 3 | D3: "约 25 个 import 站点" undercount flattered chosen option | **Fixed** | Now "约 32 个 import 站点改动（2026-09-09 实测，剔除待删 task 模块后；全仓现状约 41）" — matches my raw import-line recount (34/44 by a slightly different counting rule; within the stated 约 tolerance) |
| 4 | D3: rejection of consolidation had no benefit-side accounting | **Partially fixed** | "收益仅目录美观（模块图已由 doc.go 正名解决）" — benefit side now stated but still a single clause (residual, see attacks) |
| 5 | D4/D8: concurrent in-flight branches + partial-deletion failure scenarios | **Not fixed** | Still only "删除单独成 commit，不与在途变更混提交" (commit-mixing only) |
| 6 | D4 NFR: archival credential-surface never weighed | **Fixed** | New Risk 5 with explicit irreversibility acknowledgment |
| 7 | D6: no timeline/effort/bandwidth | **Not fixed** | Still absent from the entire document |
| 8 | D8: no L/I ratings | **Fixed** | All five risks now carry 低/中/高 L and I, honestly distributed (Risk 3 中/低, Risk 5 低/中) |
| 9 | D8: Risk-4 either/or archival lacked selection criterion | **Fixed** | Split into Risk 4/Risk 5 with per-tool disposal rules |
| 10 | D9: SC-5 fail-set predicate ill-defined ("一致" vs "子集") | **Fixed** | Unified: "失败集合是清理前基线失败集合的**子集**（统一判据：只允许减少、不允许新增，不要求两集合相同）" in both Requirements and SC-5 |
| 11 | D9: archival step had no SC | **Fixed** | SC-1 now "存档先行"：git log check for cas-diag/k8s-diag + excerpt-file check for tmp-k8s-cred |
| 12 | D9: SC-3 grep timing ambiguous | **Fixed** | "必须在删除后执行" with stated reason |
| 13 | D9/D10: "删除单独成 commit" unverified | **Not fixed** | Still no SC (residual) |
| 14 | Blindspot 1: 10 dangling comment refs registered nowhere | **Fixed** | Out of Scope names tracking anchor eval/followup.md; file exists, all 10 file:line entries match Evidence exactly (verified: cam 2 + cert 8, lines 25/26/114/962/968, 1208, 231, 3) |
| 15 | Blindspot 2: archival irreversibility unacknowledged | **Fixed** (for tmp-k8s-cred) | Risk 5: "永久写进 git 历史——不可逆，且本提案无历史重写计划" |

## Independent Code Verification Results (this iteration)

| Claim (new or load-bearing) | Verdict |
|---|---|
| 10 comment-level `internal/task` refs (cert 8 + cam 2, exact lines) | ✅ grep-verified, exact match incl. line numbers; followup.md entries 1:1 identical |
| `cmd/tmp-{cas-diag,k8s-cred,k8s-diag}` untracked (`git ls-files` empty), dirs exist | ✅ |
| `internal/{task,sync,servicetree,account,asset}` and `internal/cam/{task,sync,servicetree}` all present | ✅ |
| SC-3 timing reason: "删除前会计入待删的 cmd/tmp-k8s-cred" | ⚠️ **incomplete**: `internal/task/module.go:8` imports `internal/account/repository` and `internal/servicetree/service/rule_engine.go:9-10` imports `internal/asset/{domain,repository}` — two more to-be-deleted contaminators the parenthetical does not name (operative post-deletion instruction is nonetheless correct and unambiguous) |
| Consolidation cost "约 32（剔除待删 task 后）/约 41（现状）" | ✅ within tolerance (raw count 34/44 by import-line rule) |
| SC-5 "go vet ./... 通过" achievable pre-deletion | ✅ measured: `go vet ./...` exit 0, zero output on current tree — the claim is not just aspirational |
| account doc.go consumer spec (cert/logquery/mcp/cam-cost/cam-task/ioc) complete post-deletion | ✅ (grep: remaining account importers are exactly those 6 groups) |
| asset consumer claim "删掉 task/servicetree 后仅剩 cert 域 + ioc/cert.go" | ✅ |
| `internal/task` does NOT import account/asset via service-level paths that would survive | ✅ (its account/asset imports are all inside the to-be-deleted tree) |

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution**: direct match, unchanged from iteration 1. Pure deletion + doc correction; no scope-dodging.
2. **Solution → Evidence**: all evidence claims re-verified or carried over verified. The revision's additions (Risk 5 credential-chain description, differentiated archival disposal) are factually accurate.
3. **Evidence → Success Criteria**: SC-1..SC-5 now each mechanically checkable and 1:1 with In-Scope items, including the previously missing archival step.
4. **SC Consistency Deep-Dive (re-run on revised set)**: clustered by affected area —
   - Cluster A (internal/{task,sync,servicetree} + 3 tmp tools): SC-1 ↔ SC-2 ↔ In-Scope 1/2 — bidirectionally satisfiable (0 importers verified; the add-then-delete-in-same-commit failure mode is correctly excluded by Requirements' "删除单独成 commit" plus SC-1's `git log -- cmd/<dir>` 非空 check, which returns empty for a net-zero same-commit add+delete — the SC actually catches that mistake).
   - Cluster B (internal/account + internal/asset): SC-3 ↔ Solution 2/3 ↔ In-Scope 3/4 — consistent; timing ambiguity **resolved** ("必须在删除后执行"). New minor issue: the *justification* parenthetical names only tmp-k8s-cred as the pre-deletion contaminator; internal/task and internal/servicetree also import account/asset (verified above). Operative criterion unaffected.
   - Cluster C (decision doc), Cluster D (build/test): SC-4 doc-only, no coupling with SC-5. SC-5's subset predicate is now well-defined and consistent with the Requirements formulation.
   - **No confirmed SC↔SC or SC↔InScope contradictions introduced by the revisions.** The author's SC Consistency Check section remains accurate; re-verified independently.
5. Pre-score anchor (channelled to D9): one revision-introduced accuracy gap (SC-3 parenthetical) and one pre-existing operational gap (no flakiness policy for the subset criterion → blindspot).

## Phase 2 — Rubric Scoring (verification stance; no credit for improvement, page-state only)

### 1. Problem Definition — 106/110
- **Problem stated clearly: 39/40.** "悬空模块（零外部 importer + 误导性'待迁移'文档）" is unambiguous; two readers cannot diverge.
- **Evidence provided: 40/40.** Import-level measurement with exact file:line citations; every spot-check this iteration matched exactly, including the 10-entry comment list that now mirrors followup.md. Approximations are flagged with 约 and their counting rule stated ("仅统计 import 语句（排除注释）").
- **Urgency justified: 27/30.** Three concrete drivers incl. the dated decision-0001 Phase 4 overlap argument. No quantified cost-of-delay.

### 2. Solution Clarity — 116/120
- **Approach concrete: 39/40.** Four numbered actions with exact paths/files and per-tool archival disposal.
- **User-facing behavior: 43/45.** "用户体验：无（纯内部清理）" remains the honest answer, and the previously overstated benefit is now correctly scoped ("模块图在 **import 级**与运行时事实一致" + explicit 口径说明 that comment-level consistency awaits the follow-up round). Small residual: "体验改善发生在开发者侧" stays qualitative.
- **Technical direction: 34/35.** Verification commands, commit discipline, doc-authoring methodology, and the SC-3 post-deletion grep rule are all specified.

### 3. Industry Benchmarking — 109/120
- **Industry solutions referenced: 37/40.** Three real tools named and correctly characterized (deadcode as call-graph-based is accurate), plus the deprecate-then-delete pattern. Deduct 3: references are named but not externally sourced, and the one-sentence comparison of what U1000/deadcode would add here is thin.
- **≥3 meaningful alternatives: 27/30.** Three approaches incl. do-nothing; the consolidation alternative is genuine (source-cited, real measured cost now matching independent measurement). Deprecate-then-delete variant disposed with a valid rationale (0 importers).
- **Honest trade-offs: 24/25.** Chosen option's con self-acknowledged; the 32/41 site counts now verified within tolerance.
- **Justified against benchmarks: 21/25.** Non-adoption rationale for tools is clear and correct; the pivotal consolidation rejection's benefit side is now stated but remains a single clause ("收益仅目录美观（模块图已由 doc.go 正名解决）") with no accounting of what consolidation would *also* buy (e.g., single-writer-path clarity for asset).

### 4. Requirements Completeness — 93/110
- **Scenario coverage: 30/40.** Safety scenarios strong (baseline-subset criterion, same-name package protection). Unfixed from iteration 1: no scenario for concurrent in-flight branches referencing deleted paths (the doc shows awareness of in-flight work but only handles commit-mixing), and no partial-deletion failure/recovery scenario.
- **NFRs: 36/40.** 零行为变化 + build artifacts + the archival security-surface NFR now covered via Risk 5. Residual: no rollback/recovery statement (implicitly available via archival + git revert, but never stated).
- **Constraints & dependencies: 27/30.** Decision-0001 dependency, baseline recording, commit isolation, and explicit deferral of 收编迁移 all named. Minor: baseline 留档 location/format unspecified.

### 5. Solution Creativity — 46/100
- **Novelty: 16/40.** Document declares "无创新" honestly; the import-statement-only death criterion is a real but small methodological refinement.
- **Cross-domain inspiration: 8/35.** None claimed, none present.
- **Simplicity of insight: 22/25.** The leap is restraint; elegant, not forced.

### 6. Feasibility — 83/100
- **Structural note**: the document has no "Feasibility Assessment" heading; feasibility substance is scattered across Evidence, Constraints & Dependencies, and Risks. This iteration scores substance (consistent with the iteration-1 precedent the reviser relied on) rather than zeroing per a strict reading of the Section→Dimension mapping — but the missing section is a real structural gap an executor would notice.
- **Technical feasibility: 39/40.** End-to-end verified, now including the measured fact that `go vet ./...` already passes — SC-5's build/vet gate is achievable today, not merely asserted.
- **Resource & timeline feasibility: 16/30.** **Still no timeline, effort estimate, or bandwidth statement anywhere in the document** (unfixed from iteration 1). Even a one-line "half-day, single engineer" would satisfy the criterion; silence does not.
- **Dependency readiness: 28/30.** Decision 0001 verified to exist; archival prerequisites identified and now differentiated per tool.

### 7. Scope Definition — 78/80
- **In-scope concrete: 29/30.** Five precise deliverables, incl. per-tool archival rules.
- **Out-of-scope explicit: 25/25.** Five named items; the comment-cleanup deferral now carries a working tracking anchor (followup.md, verified 1:1 with Evidence) — model out-of-scope writing.
- **Bounded: 24/25.** Finite file list; doc content bounded by SC-3's mechanical criterion.

### 8. Risk Assessment — 84/90
- **Risks identified: 27/30.** Five meaningful, non-trivial risks. Still missing: (a) the hazard the proposal itself identifies — post-deletion dangling comments in live code ("删除后成悬空注释；登记为 follow-up") will actively misdirect readers of `cert/service/execute_service.go:26` ("参考 internal/task executor 既有用法") until the CMDB doc-consistency pass — a self-identified hazard that never entered the risk register; (b) concurrent in-flight branches referencing deleted paths.
- **Likelihood + impact rated: 27/30.** All five rated; distribution is honest (mixed L/I, not the degenerate "all low"). Ratings are asserted without per-cell justification, though the mitigation prose carries the reasoning.
- **Mitigations actionable: 30/30.** All executable; Risk 4/Risk 5's differentiated disposal (commit-then-delete for the two secret-free tools vs excerpt-only for tmp-k8s-cred) is concrete, with the secret-free claim verified in iteration 1 (`v.GetString` config reads).

### 9. Success Criteria — 73/80
- **Measurable and testable: 26/30.** SC-1 (git log non-empty + excerpt file exists + dirs absent), SC-2 (import-level grep with cam-path exemption), SC-3 (post-deletion grep, one-to-one, explicit timing), SC-5 (well-defined subset predicate with a concrete red-test example) are strong. Deducts: (a) SC-3's explanatory parenthetical is incomplete — "删除前会计入待删的 cmd/tmp-k8s-cred" names only one of three to-be-deleted contaminators (verified: `internal/task/module.go:8` imports account, `internal/servicetree/service/rule_engine.go:9-10` imports asset); the operative instruction is correct, the stated reason is not exhaustive; (b) SC-3's "一一对应（多列、漏列均判失败）" does not fix granularity — grep yields ~16 per-file import lines across 6 domains while doc.go will list domain groups; at which level "多列" fails is unstated; (c) SC-2's "import 语句级 grep" still leaves the mechanical comment-exclusion method unspecified.

- **Coverage complete: 22/25.** The archival step now has SC-1 coverage. Residual: the "删除单独成 commit" constraint and the baseline-recording prerequisite ("基线失败集合删除前留档") have no SC — though SC-5 is unexecutable without the baseline, which provides implicit enforcement.
- **SC internal consistency: 25/25.** Re-ran the full clustering + bidirectional satisfiability check on the revised set: no SC↔SC or SC↔InScope contradictions; the SC-3 timing ambiguity is resolved; no new contradictions introduced by either revision round.

### 10. Logical Consistency — 88/90
- **Solution addresses the stated problem: 34/35.** Direct match; the benefit claim is now correctly scoped to import-level consistency with the deferral honestly registered.
- **Scope ↔ Solution ↔ SC aligned: 29/30.** Aligned including the previously missing archival SC. Residual: commit-isolation constraint unverified.
- **Requirements ↔ Solution coherent: 25/25.** The acceptance predicate is now single and identical in both places ("只允许减少、不允许新增，不要求两集合相同").

**Cross-dimension coherence check**: scope, solution, SC, and risk register are mutually consistent on everything except the residuals noted above. No cross-section contradictions found in the revised document.

## Phase 3 — Blindspot Hunt

1. **[blindspot] The subset criterion has no flakiness policy.** "失败集合是清理前基线失败集合的**子集**（统一判据：只允许减少、不允许新增）" is a binary mechanical check, but the test universe is non-deterministic at the margins: a flaky test that passed at baseline and fails post-deletion makes the criterion fail spuriously — and the executor's only escape is re-running until green, which quietly converts the subset check into "last run wins". No retry/quarantine rule, no definition of whether a test red at baseline but green post-deletion and then red again counts as a violation. For a proposal whose entire verification edifice rests on this one predicate, the missing flakiness handling is a real operational gap no rubric dimension captures.

2. **[blindspot] Feasibility section absent against the rubric's required-sections map.** The Section→Dimension mapping requires a "Feasibility Assessment" section; missing sections score 0. This iteration scored D6 on substance (consistent with the iteration-1 precedent), but the document would be strictly non-compliant under a literal reading, and the scattered feasibility content (Evidence → technical; Constraints → dependency; nothing → timeline) is exactly what a reviewer reads as "feasibility was never seriously done". Adding the section heading — even 5 lines consolidating existing content plus an effort estimate — removes the ambiguity at near-zero cost. Reasoning-audit flagged this independently of dimension scoring.

## Bias Detection Report

- Annotated regions: 23 of 86 non-empty lines carry `<!-- pre-revised: … -->` markers (16 high / 4 medium / 3 low)
- Attack points in annotated regions: 2 (SC-3 parenthetical incompleteness; SC-3 一一对应 granularity) / 23 annotated lines → density 0.087
- Attack points in unannotated regions: 4 (no timeline; missing concurrent-branch/partial-failure scenarios; risk-register omission of the self-identified dangling-comment hazard; thin benefit-side justification) / 63 unannotated lines → density 0.063
- **Ratio (annotated/unannotated): 1.4**

Annotated density is only slightly higher; the two annotated-region attacks are minor accuracy nits (the revision's operative logic is correct), while the substantive attacks all live in regions the revisions did not touch — i.e., the remaining score is limited by pre-existing omissions, not by revision-introduced damage. No revision introduced a contradiction, regression, or scope drift; every checkable factual addition (Risk 5 credential chain, 32/41 site counts, SC-3 timing rule, followup.md anchor) verified accurate. No `conflict-with-pre-revision` tags were warranted: no revision direction contradicted rubric judgment this round.

## Result

```
SCORE: 876/1000
```

Residual gap to target (900) is concentrated in three fixable items: add an effort/timeline statement (D6 +14 potential), register the self-identified dangling-comment misdirection hazard in the risk table (D8/D4 +6), and close the SC mechanical-spec nits (SC-3 contaminator list, 一一对应 granularity, SC-2 exclusion method) plus an SC for commit isolation (D9/D10 +5).
