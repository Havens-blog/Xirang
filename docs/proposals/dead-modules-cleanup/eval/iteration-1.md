# Proposal Eval — Iteration 1

- **Document**: D:\Haven\docs\proposals\dead-modules-cleanup\proposal.md (current revision)
- **Scorer persona**: CTO (approved/rejected hundreds of proposals)
- **Scale**: 1000 (target 900) | **Iteration**: 1
- **Verification stance**: every factual assertion independently re-verified against D:\Haven\e-cam-service on 2026-09-09

## Independent Code Verification Results

All load-bearing factual claims were re-checked and **confirmed accurate**:

| Claim | Verdict |
|---|---|
| `internal/task` external importers = 0; comment mentions = 10 (cam 2 + cert 8, exact lines 25/26/114/962/968, 1208, 231, 3) | ✅ exact match, all line numbers correct |
| `internal/sync`, `internal/servicetree` external importers = 0 | ✅ |
| `camModule.ServiceTreeModule` wiring (ioc/gin.go:144) | ✅ |
| account consumers ≈20 (cert module+3 services+tests, logquery, mcp, cam-cost, cam-task module.go:6 + sync_billing.go:11, ioc/{cert,logquery}, cmd/tmp-k8s-cred) | ✅ measured 22 import lines incl. dead task module |
| `SyncAccount` → `taskQueue.Submit` at account/service/account.go:338 | ✅ |
| mcp/deps.go:59 injects nil queue → `Queue.Submit` (pkg/taskx/queue.go:224) dereferences `q.mu` → panic on sync_account (mcp/tools_sync.go:85) | ✅ |
| `internal/account/module.go`, `internal/asset/module.go` 0 importers (root package) | ✅ |
| 3 tmp tools not in git (`git ls-files` empty), exist on disk | ✅ |
| ioc/, Makefile, main.go zero refs to the 3 dead modules | ✅ (grep exit 1) |
| Baseline red test `TestBsonTagsMatchSchema/K8sCredential` (bson `displayName` vs schema.sql) | ✅ actually fails, exactly as described (`got displayName`, `want` without it) |
| Decision doc 0001 exists at workspace root; Phase 4 = 退位删除 `internal/cmdb`; currently no asset 归属 record | ✅ |
| account/asset doc.go contain the outdated migration plans | ✅ |
| cam-side counterparts `internal/cam/{task,sync,servicetree}` all exist and are active | ✅ |
| tmp tools contain no hardcoded secrets (creds read from config via `v.GetString`) | ✅ |

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution**: direct match. Problem = dangling zero-consumer modules + misleading docs; solution = delete them + rewrite docs. No scope-dodging.
2. **Solution → Evidence**: evidence is import-level measurement, and it is *correct* (verified above). Notably the doc.go rewrite spec in Solution item 2 already excludes `cmd/tmp-k8s-cred` from the consumer list, correctly anticipating its deletion — internally anticipatory, not contradictory.
3. **Evidence → Success Criteria**: SC-1..SC-5 are mechanically checkable and map 1:1 onto In-Scope items — except the "先提交入库再删" archival step (see D9).
4. **Self-contradiction check (SC Consistency Deep-Dive)**: clustered by affected area —
   - Cluster A (internal/{task,sync,servicetree} + tmp tools): SC-1 ↔ SC-2 ↔ In-Scope 1/2 — bidirectionally satisfiable (0 importers verified). No contradiction.
   - Cluster B (internal/account): SC-3 ↔ Solution 2 ↔ In-Scope 3/4 — consistent **provided** SC-3's consumer-list grep is run *post*-deletion; the SC does not state the timing. Marked **ambiguous — requires author clarification** (pre-deletion grep would list tmp-k8s-cred and fail SC-3's one-to-one check against the doc.go spec, which omits it).
   - Cluster C (internal/asset), Cluster D (decision doc), Cluster E (build/test): no contradictions.
   - **No confirmed SC↔SC or SC↔InScope contradictions.** The author's own SC Consistency Check section is sound; I verified its claims independently.
5. Pre-score anchor (channelled to D10/blindspot): Requirements bullet states two different acceptance thresholds for the same verification — "通过/失败集合**一致**" vs "只允许减少、不允许新增".

## Phase 2 — Rubric Scoring

### 1. Problem Definition — 104/110
- **Problem stated clearly: 38/40.** Unambiguous; two readers could not diverge on what a "悬空模块/迁移半成品" means here.
- **Evidence provided: 39/40.** Exceptional — measured dependency analysis with exact file:line citations, all independently verified. "约 20 处" (actual 22) and "约 25 个 import 站点" are approximations, flagged with 约.
- **Urgency justified: 27/30.** Three concrete drivers, incl. a dated cross-link to decision 0001 Phase 4 with a real overlap argument.

### 2. Solution Clarity — 111/120
- **Approach concrete: 38/40.** Four numbered actions with exact paths/files.
- **User-facing behavior: 40/45.** Honestly answers "用户体验：无（纯内部清理）" — the correct answer for this proposal type. Deduct 5: "模块图与文档第一次与运行时事实一致" (line 41) is **overstated on this page** — the proposal's own Out-of-Scope item defers cleanup of 10 live-code comment references that dangle after deletion ("删除后成悬空注释，与本提案'文档与事实一致'收益相悖"). Fact-consistency is achieved only for the rewritten doc.go files, not the module graph's documentation. `conflict-with-pre-revision`: the pre-revision correctly *added* the deferral but left this unannotated claim untouched, creating a cross-section contradiction the revision did not resolve.
- **Technical direction: 33/35.** Verification commands, commit discipline, and doc-authoring methodology all specified.

### 3. Industry Benchmarking — 87/120
- **Industry solutions referenced: 20/40.** The entire Industry Solutions section is: "常规做法即'删除 + 编译验证'…更进一步的做法是先收编再删". Zero named tools or published patterns. For dead-code work the obvious citations exist and are industry-standard: `staticcheck` U1000, `golang.org/x/tools/cmd/deadcode`, golangci-lint `unused`, and the deprecate-then-delete pattern. This is the weakest section of an otherwise evidence-dense document.
- **≥3 meaningful alternatives: 26/30.** Three present including do-nothing; the consolidation alternative is genuine (source-cited to account/doc.go's original plan, real cost "约 25 个 import 站点改动"), not a straw man. Not considering the industry-common deprecation-window variant is defensible given 0 importers, but costs a few points.
- **Honest trade-offs: 22/25.** The adopted approach's con is self-acknowledged ("account/asset 仍位于独立顶层目录"). Deduct 3: "约 25 个 import 站点改动" undercounts — I measure ~32 import lines after excluding the to-be-deleted task module (~41 total today); the error flatters the chosen option.
- **Justified against benchmarks: 19/25.** Do-nothing rejection is well-grounded; the pivotal rejection of consolidation rests on the unquantified verdict "收益不成比例" — no benefit-side accounting at all.

### 4. Requirements Completeness — 92/110
- **Scenario coverage: 30/40.** Safety scenarios are strong (baseline-equivalence with fail-set recording; same-name package protection). Missing edge scenarios: concurrent in-flight branches touching deleted paths (the doc itself acknowledges "在途变更" exist yet only addresses commit-mixing), and partial-deletion failure handling.
- **NFRs: 34/40.** 零行为变化 + build artifacts covered. Missing NFR: the archival strategy permanently commits credential-handling tools into repo history (see D8) — a security/surface NFR never weighed.
- **Constraints & dependencies: 28/30.** Decision-0001 dependency, baseline recording, commit isolation, and explicit deferral of收编迁移 all named.

### 5. Solution Creativity — 46/100
- **Novelty: 16/40.** The document declares "无创新" honestly; the one methodological point — import-statement-only death criterion, excluding comment mentions — is a real, articulated refinement over naive grep-based detection, but small.
- **Cross-domain inspiration: 8/35.** None claimed; none present.
- **Simplicity of insight: 22/25.** The leap is restraint itself: pure deletion with a precise kill criterion. Elegant, not forced.

### 6. Feasibility — 82/100
- **Technical feasibility: 39/40.** Verified end-to-end. Unstated but true: Go's `internal/` visibility rule guarantees no out-of-module importers — the evidence methodology is actually *complete*, though the doc never states this defense-in-depth argument (unstated assumption that happens to hold).
- **Resource & timeline feasibility: 16/30.** **No timeline, no effort estimate, no bandwidth statement anywhere in the document.** Even for a small cleanup, the rubric criterion cannot be satisfied by silence.
- **Dependency readiness: 27/30.** Decision 0001 exists (verified); archival prerequisite for tmp tools is identified and remediated.

### 7. Scope Definition — 78/80
- **In-scope concrete: 29/30.** Five precise deliverables.
- **Out-of-scope explicit: 25/25.** Five named items, including the subtle comment-cleanup deferral *with* reasoning — model out-of-scope writing.
- **Bounded: 24/25.** Finite file list; doc content bounded by SC-3's mechanical criterion.

### 8. Risk Assessment — 59/90
- **Risks identified: 24/30.** Four meaningful, non-trivial risks. Missed: (a) once committed per the Risk-4 mitigation, `cmd/tmp-k8s-cred` — a tool that pulls cluster kubeconfigs using account AccessKeys read from the DB — lives in **git history permanently**, expanding the repo's credential-adjacent surface forever (no history-rewrite plan exists); (b) concurrent in-flight branches referencing deleted paths.
- **Likelihood + impact rated: 7/30.** **The table has no likelihood or impact ratings at all** — only `# | Risk | Mitigation`. The rubric criterion is explicitly unmet. Partial credit only for implicit likelihood reasoning embedded in mitigation prose ("已实测零外部 importer（无 import 即无 init 触发）").
- **Mitigations actionable: 28/30.** All four are executable; Risk 4's commit-first archive is concrete (though its either/or tail "或摘录关键诊断逻辑留档后再删" lacks a selection criterion).

### 9. Success Criteria — 67/80
- **Measurable and testable: 27/30.** SC-3's mechanical one-to-one criterion ("多列、漏列均判失败") is excellent. Deduct: SC-5's "通过/失败集合与清理前基线一致" is ill-defined as a mechanical check — deleting code changes the test universe, and the document's own Requirements formulation ("只允许减少、不允许新增") is the well-defined one; the two are not the same predicate. Also "import 语句级 grep" does not specify how comments are mechanically excluded.
- **Coverage complete: 17/25.** Gap: the **archival step has no SC**. In-Scope says tmp tools "先提交入库再删" and Risk 4's mitigation depends on it — yet SC-1 only verifies the directories are gone. An executor who skips the archival commit and deletes outright passes all five SCs while silently voiding the Risk-4 mitigation. Minor: the "删除单独成 commit" constraint is likewise unverified.
- **SC internal consistency: 23/25.** No SC↔SC or SC↔InScope contradictions (author's self-check independently confirmed). SC-3's grep timing (pre- vs post-deletion) is **ambiguous — requires author clarification**: only the post-deletion reading is consistent with SC-1 deleting tmp-k8s-cred, an account consumer.

### 10. Logical Consistency — 81/90
- **Solution addresses the stated problem: 31/35.** Direct match. Deduct for the line-41 benefit claim partially undercut by the proposal's own deferral (see D2).
- **Scope ↔ Solution ↔ SC aligned: 27/30.** Aligned except the unanchored archival step (no SC) and the stale line-41 claim.
- **Requirements ↔ Solution coherent: 23/25.** The Requirements bullet simultaneously demands "通过/失败集合一致" and "只允许减少、不允许新增" — two different acceptance thresholds for the same verification, one stricter than the other.

**Cross-dimension coherence check**: scope, solution, and SC are aligned on everything except the archival-step verification gap (D9) and the fail-set predicate mismatch (D9/D10). No other cross-section contradictions found.

## Phase 3 — Blindspot Hunt

1. **[blindspot] Deferral-registration asymmetry (Reasoning audit flagged this independently of dimension scoring).** The proposal invents a linkage mechanism precisely to prevent its own deferrals from being forgotten — "Phase 4 补记 asset 归属决断项，防止遗忘" — yet its largest deferral, the 10 dangling comment references, is registered nowhere: "登记为 follow-up（随下轮文档一致性处理）" names no tracking artifact, no issue, no doc anchor. By the proposal's own method, an unanchored follow-up is a forgotten follow-up; the next round of "文档一致性处理" has no pointer back to this list of 10 citations.
2. **[blindspot] Archival creates permanent, one-way history.** "缓解改为先提交、再删除——删除前以独立 commit 入库存档" is the correct fix for recoverability, but it is irreversible in effect: `cmd/tmp-k8s-cred` (which fetches kubeconfigs using account AccessKeys) enters repo history permanently, with no stated policy note (e.g., archive in a separate/private repo, or strip before archival). No rubric dimension covers the permanence direction of a mitigation; scored contextually under D8, flagged here because the *irreversibility* itself is never acknowledged.

## Bias Detection Report

- Annotated regions: 4 attack points / 17 paragraphs = density 0.24
- Unannotated regions: 6 attack points / 37 paragraphs = density 0.16
- Ratio (annotated/unannotated): 1.5

Annotated density is modestly higher, which is expected: pre-revision regions carry the newly added content (tmp-k8s-diag, archival mitigation, mechanical SC criteria) where new gaps (missing archival SC, grep-timing ambiguity, fail-set predicate) surfaced. No systematic leniency toward revised paragraphs was observed; two pre-revision additions (Evidence expansion, baseline-red correction) were verified as strict factual improvements.

## Result

```
SCORE: 807/1000
```
