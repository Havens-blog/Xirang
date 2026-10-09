# Proposal Eval — Iteration 3 (final)

- **Document**: D:\Haven\docs\proposals\dead-modules-cleanup\proposal.md (current revision, after Pre-Revision + three revision rounds)
- **Scorer persona**: CTO (approved/rejected hundreds of proposals)
- **Scale**: 1000 (target 900) | **Iteration**: 3 | **PREVIOUS_REPORT_PATH**: eval/iteration-2.md
- **Verification stance**: every new/changed assertion re-verified against D:\Haven\e-cam-service on 2026-09-09; carry-over claims spot-re-verified, not assumed.

## Prior Attack Disposition (iteration-2 → iteration-3)

| # | Iteration-2 attack | Status | Evidence |
|---|---|---|---|
| 1 | D3: consolidation rejection's benefit side a single clause ("收益仅目录美观") | **Fixed** | Comparison Verdict now accounts benefits concretely: "收益为单一目录归属（asset 直属 cert 域、account 不再游离顶层）并消除 doc.go 与目录事实再漂移的维护面，但模块图问题已由 doc.go 正名解决，收益不成比例于约 32 个 import 站点改动" — genuine two-sided accounting against a verified cost |
| 2 | D4: no concurrent in-flight-branch or partial-deletion-failure scenarios | **Fixed** | Two new Key Scenarios ("在途分支冲突", "部分删除失败") with recovery paths (revert → rebase / revert → re-initiate); minor new defect introduced, see attacks |
| 3 | D4 NFR: no rollback/recovery statement | **Fixed** | "回滚即 revert 该 commit" + "tmp 工具已按 SC-1 在删除前留档，无内容丢失" |
| 4 | D6: no timeline/effort/bandwidth anywhere | **Fixed** | New "Feasibility Assessment" section: "**半天内单人完成**", the two genuinely time-consuming items named (SC-3 grep comparison, SC-1 archival), operator = maintainer, "可插入任意空档执行，无排期约束" |
| 4b | Blindspot 2: "Feasibility Assessment" heading missing vs Section→Dimension map | **Fixed** | Section now exists with that exact heading; structural non-compliance resolved at zero cost as suggested |
| 5 | D8: self-identified dangling-comment hazard absent from risk register; concurrent branches absent | **Fixed** | Risk 6 (悬空注释主动误导, 中/低, PR-announcement mitigation) + Risk 7 (并发在途分支, 低/中) — both now registered with L/I |
| 6 | D9: SC-3 contaminator parenthetical named only 1 of 3 to-be-deleted importers | **Fixed** | Now names all three: "cmd/tmp-k8s-cred，以及 internal/task/module.go:8 import account、internal/servicetree/service/rule_engine.go:9-10 import asset" — **re-verified exactly**: module.go line 8 is the accountrepo import; rule_engine.go lines 9-10 are the asset domain/repository imports |
| 7 | D9: SC-3 一一对应 granularity unfixed | **Fixed** | "粒度为**文件级**：grep 命中的每个消费方文件须在 doc.go 清单中，多列、漏列均判失败；不要求到行级" — consistent with Solution 2's "消费方列全至文件级" |
| 8 | D9: SC-2 mechanical comment-exclusion method unspecified | **Not fixed** | SC-2 still says only "import 语句级 grep…零命中（internal/cam/* 同名路径除外）" — no command/pattern for how comments are excluded (residual) |
| 9 | D9/D10: no SC for "删除单独成 commit" / baseline archiving | **Not fixed** | Both remain constraints-only (Constraints & Dependencies); no acceptance criterion. SC-5's dependency on the baseline provides implicit enforcement only (residual) |
| 10 | Blindspot 1: subset criterion had no flakiness policy | **Fixed** | Full policy in SC-5: single continuous run (`go test -count=1 ./...`), "禁止反复重跑取绿", 1 repro run, "同一测试两次结果不一致（flaky）按基线失败集合成员处理，不计为新增" — closes the last-run-wins loophole the blindspot described |
| 11 | D2/D4 misc residuals (developer-benefit qualitative) | Unchanged | "体验改善发生在开发者侧" remains qualitative (minor) |

## Independent Code Verification Results (this iteration)

| Claim (new or load-bearing) | Verdict |
|---|---|
| Feasibility: "go vet ./... 在当前树已实测通过（2026-09-09）" | ✅ re-measured this iteration: exit 0, zero output — SC-5's vet gate is a present-tense fact |
| account "以读为主"（SyncAccount 经 taskQueue.Submit, account/service/account.go:338） | ✅ Submit call present at the cited location |
| mcp 注入 nil 队列（internal/mcp/deps.go:59, `NewCloudAccountService(accountRepo, nil, logger)`） | ✅ nil queue literal in the deps constructor; sync_account panic path claim is sound |
| tmp-k8s-cred credential chain（main.go:55: `NewClientWithAccessKey(…acc.AccessKeyID, acc.AccessKeySecret)`） | ✅ exact; also contains a hardcoded clusterID (non-credential, does not undermine Risk 5's claim) |
| Risk 5: tmp-cas-diag/tmp-k8s-diag 无硬编码凭证（均经 v.GetString） | ✅ both read mongodb dsn/user/password via `v.GetString`; no secrets in source |
| 三只 cmd/tmp-* 未入 git（`git ls-files` 为空） | ✅ re-confirmed: `git ls-files cmd/` contains no tmp-* entries; dirs exist on disk |
| internal/task/module.go:8 imports account | ✅ exact line |
| internal/servicetree/service/rule_engine.go:9-10 imports asset | ✅ exact lines (grep -n: 9, 10) |
| account/asset module.go 零引用外壳 | ✅ root package paths `"…/internal/account"` and `"…/internal/asset"` have 0 import sites; subpackage consumers unaffected by shell deletion |
| cam/task/module.go:6 and executor/sync_billing.go:11 import account/service | ✅ exact lines |
| account/asset doc.go current "待迁移/迁移计划" misdirection | ✅ both doc.go files carry the abandoned migration plans as quoted |
| Decision 0001 Phase 4 will delete internal/cmdb（工作区根 docs/decisions/） | ✅ file exists at D:\Haven\docs\decisions\, Phase 4 "退位删除" confirmed |
| Consolidation cost "约 32（剔除待删 task 后）/约 41（现状）" | ✅ carried over from iteration-2 measurement (34/44 raw by import-line rule, within 约 tolerance) |

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution**: direct match, unchanged across all rounds. Pure deletion + doc truth-telling; no scope-dodging.
2. **Solution → Evidence**: every new assertion in round 3 (effort estimate, vet-passed-today, credential chain, per-tool disposal) is either measured or verified exact. No cherry-picking detected.
3. **Evidence → Success Criteria**: SC-1..SC-5 each mechanically checkable, 1:1 with In-Scope items; the flaky policy makes the subset predicate total (defined for every outcome).
4. **SC Consistency Deep-Dive (re-run on the round-3 set)**:
   - Cluster A (internal/{task,sync,servicetree} + 3 tmp tools): SC-1 ↔ SC-2 ↔ In-Scope 1/2 — bidirectionally satisfiable; SC-1's `git log -- cmd/<dir>` check still catches a net-zero same-commit add+delete mistake.
   - Cluster B (account/asset): SC-3 ↔ Solution 2/3 ↔ In-Scope 3/4 — consistent; granularity and timing now fully pinned; contaminator justification now exhaustive and verified.
   - Cluster C (decision doc, SC-4) and Cluster D (build/test, SC-5): no coupling conflicts; flaky policy consistent with the subset predicate in Requirements (identical wording).
   - New cross-reference found: Key Scenarios/Risk 7 point in-flight-branch import cleanup at "followup.md 清单", but followup.md is explicitly the **comment-level** dangling-reference list (10 处, "判定方法: import 语句级 grep 对这些行不命中（均为注释级）") and contains no import entries. The artifact is being repurposed as a tracking anchor without saying so. Operative recovery path (rebase → remove broken imports) is still executable; the pointer is just mislabeled. Channeled to D4/D8 as minor deductions.
   - **No SC↔SC or SC↔InScope contradictions; no new contradictions introduced by round 3.**
5. Pre-score anchor: the revision program is converging — all remaining findings are minor residuals or the one new mislabeled pointer.

## Phase 2 — Rubric Scoring (verification stance; no credit for improvement, page-state only)

### 1. Problem Definition — 106/110
- **Problem stated clearly: 39/40.** "悬空模块（零外部 importer + 误导性'待迁移'文档）" unambiguous; same-name pairs enumerated.
- **Evidence provided: 40/40.** Import-level measurement with exact file:line citations; every spot-check across three iterations has matched exactly, including this round's module.go:8 and rule_engine.go:9-10.
- **Urgency justified: 27/30.** Three concrete drivers incl. the dated Phase-4 overlap argument; still no quantified cost-of-delay.

### 2. Solution Clarity — 116/120
- **Approach concrete: 39/40.** Four numbered actions, exact paths, per-tool disposal.
- **User-facing behavior: 43/45.** "用户体验：无（纯内部清理）" honest; benefit correctly scoped to import level with the deferral registered. Residual: developer-side improvement stays qualitative.
- **Technical direction: 34/35.** Verification commands, commit discipline, doc-authoring method, post-deletion grep rule all specified.

### 3. Industry Benchmarking — 112/120
- **Industry solutions referenced: 37/40.** staticcheck U1000, x/tools/cmd/deadcode, golangci-lint unused, deprecate-then-delete — correctly characterized. Deduct 3: named without external sourcing, and the non-adoption rationale slightly overstates ("工具不处理 internal/task vs cam/task 同名不同包的歧义" — unused/deadcode operate on distinct import paths and would in fact flag internal/task correctly; the honest reason is setup cost vs a 5-minute grep).
- **≥3 meaningful alternatives: 27/30.** Three approaches incl. do-nothing; consolidation alternative is genuine, cost-verified, source-cited.
- **Honest trade-offs: 24/25.** Chosen option's con self-acknowledged; 32/41 counts verified.
- **Justified against benchmarks: 24/25.** Consolidation rejection now has real benefit-side accounting (ownership, top-level hygiene, doc-drift elimination) weighed against measured cost. Residual: no treatment of the intermediate option (consolidate asset only, ~small site count) — arguably a straw-man-free gap but not a required alternative.

### 4. Requirements Completeness — 100/110
- **Scenario coverage: 35/40.** Now covers baseline-subset acceptance, same-name protection, in-flight branch conflict, and partial-deletion failure with concrete recovery. Deduct 5: the in-flight-branch recovery instruction points at the wrong artifact — "在途分支 rebase 后按 followup.md 清单移除失效 import" — followup.md is the comment-level dangling list and holds no import entries (its own header: "internal/task 注释级悬空引用清单（10 处）…判定方法: import 语句级 grep 对这些行不命中"); an in-flight branch's broken imports would be new, unlisted references. Contingency-only ("正常不应出现此情形") so not fatal.
- **NFRs: 38/40.** 零行为变化, build artifacts, archival security surface, and now rollback ("revert 删除 commit 即恢复原状"). Residual: rollback assumes a clean revert with no interleaved commits — not stated.
- **Constraints & dependencies: 27/30.** Decision-0001 dependency, commit isolation, explicit deferral all named. Residual: baseline 留档 location/format still unspecified.

### 5. Solution Creativity — 46/100
- **Novelty: 16/40.** "无创新" declared honestly; import-statement-only death criterion is a small real refinement.
- **Cross-domain inspiration: 8/35.** None claimed, none present.
- **Simplicity of insight: 22/25.** The leap is restraint; elegant, not forced.

### 6. Feasibility — 94/100
- **Structural note**: "Feasibility Assessment" section now exists, matching the Section→Dimension map; the iteration-2 structural gap is closed.
- **Technical feasibility: 39/40.** `go vet ./...` re-measured exit 0 this iteration — the vet-gate claim is a verified present fact, not aspiration; deletion compile-safety backed by 0-importer evidence re-confirmed (including root-package paths for account/asset shells).
- **Resource & timeline feasibility: 27/30.** "**半天内单人完成**" with the two real time sinks named; operator skills and bandwidth stated; "可插入任意空档执行，无排期约束". Residual: no calendar anchor or slot commitment — acceptable for a half-day self-executed cleanup.
- **Dependency readiness: 28/30.** Decision 0001 verified on disk with the Phase 4 that motivates the linkage; archival prerequisites differentiated per tool.

### 7. Scope Definition — 78/80
- **In-scope concrete: 29/30.** Five precise deliverables with per-tool archival rules.
- **Out-of-scope explicit: 25/25.** Five named items; the comment-cleanup deferral carries a working anchor (followup.md verified 1:1 with Evidence in iteration 2, header/trigger/载体 fields present).
- **Bounded: 24/25.** Finite file list; doc content bounded by SC-3's mechanical criterion.

### 8. Risk Assessment — 87/90
- **Risks identified: 29/30.** Seven meaningful risks; both previously missing hazards (self-identified dangling-comment misdirection, concurrent branches) now registered. Nothing further a CTO would demand for a zero-behavior-change deletion.
- **Likelihood + impact rated: 28/30.** All seven rated with honest distribution (中/低 pairs present). Residual: per-cell justification remains implicit in mitigation prose.
- **Mitigations actionable: 29/30.** All executable; Risk 5's differentiated disposal verified (both non-credential tools confirmed `v.GetString`-only). Deduct 1: Risk 7's rebase instruction references "followup.md 移除失效 import" — same mislabeled-pointer defect as Key Scenarios.

### 9. Success Criteria — 77/80
- **Measurable and testable: 28/30.** SC-1 (git log non-empty + excerpt file + dirs absent), SC-3 (post-deletion grep, file-level 1:1, timing + exhaustive contaminator rationale), SC-5 (subset predicate + full flaky policy + concrete red-test example) are strong. Residuals: (a) SC-2's "import 语句级 grep" still lacks the mechanical exclusion command — Evidence states the 口径 ("仅统计 import 语句（排除注释）") but an executor implementing SC-2 must reinvent the pattern; (b) SC-3 does not state the grep invocation that distinguishes import lines from comments.
- **Coverage complete: 24/25.** Archival has SC-1; flaky handling has SC-5 coverage. Residual: "删除单独成 commit" and baseline 留档 remain SC-less — SC-5 is unexecutable without the baseline, giving implicit but not explicit enforcement.
- **SC internal consistency: 25/25.** Full clustering + bidirectional satisfiability re-run: no contradictions, including the new flaky clause (consistent with the Requirements predicate, word-for-word).

### 10. Logical Consistency — 88/90
- **Solution addresses the stated problem: 34/35.** Direct match; benefit claims correctly scoped with deferrals registered.
- **Scope ↔ Solution ↔ SC aligned: 29/30.** Aligned end to end. Residual: commit-isolation constraint unverified by any SC.
- **Requirements ↔ Solution coherent: 25/25.** Acceptance predicate single and identical in both places; consumer lists (Solution 2, Evidence, SC-3) mutually consistent — tmp-k8s-cred correctly excluded from the post-deletion grep list it contaminates pre-deletion.

**Cross-dimension coherence check**: scope, solution, SC, risk register, and now feasibility are mutually consistent. The only cross-section defect is the followup.md pointer misuse (Requirements/Risk-7 ↔ Out-of-Scope anchor), scored in D4/D8.

## Phase 3 — Blindspot Hunt

1. **[blindspot] The archive commit itself is uncommitted content entering a repo whose hygiene this proposal is meant to improve.** "tmp-cas-diag/tmp-k8s-diag 先提交入库再删（提交信息注明存档用途）" creates two history-only commits whose only purpose is archival inside the application repo — the same repo whose CMDB decision (cited by this proposal) is deleting `internal/cmdb` to shrink surface. No alternative archival home was weighed for the two clean tools (e.g., the proposal directory, as already done for tmp-k8s-cred's excerpt — the document itself contains the better pattern). Not a rubric dimension (D8 scored the credential risk; this is a repo-hygiene/archival-location question), and it does not change the verdict — but the asymmetry (one tool excerpted to the proposal dir, two tools committed into the app repo) is decided per-tool by credential exposure only, without weighing archival location at all. Low severity: both tools are verified credential-free.

2. **[blindspot] doc.go rewrite has no review checkpoint.** SC-3 verifies the doc against a post-deletion grep, but the grep is run by the same single operator (the proposal states "操作者为项目维护者（提案作者本人）") who wrote the doc. For a deliverable whose entire purpose is "让文档说真话", the verification and the artifact share one author with no second-reader step, no PR review requirement, and no diff-review instruction — the proposal elsewhere mandates PR announcements for commit isolation but not for the two doc.go rewrites, which are the only content-producing changes in scope. Reasoning audit flagged this independently of dimension scoring; it fits no dimension cleanly (D9 checks the criterion's form, not its enforcement independence).

## Bias Detection Report

- Annotated regions: 28 of 94 non-empty lines carry `<!-- pre-revised: … -->` markers (21 high / 4 medium / 3 low)
- Attack points in annotated regions: 3 (followup.md mislabeled as import list — lines 55/119; tool-ambiguity rationale overstatement — line 79; baseline 留档 format unspecified — line 66) / 28 annotated lines → density 0.107
- Attack points in unannotated regions: 2 (SC-2/SC-3 mechanical grep pattern unspecified — line 124; commit-isolation/baseline SC absence) / 66 unannotated lines → density 0.030
- **Ratio (annotated/unannotated): 3.6**

Annotated density is now clearly higher, and the pattern is the expected late-revision signature: the three annotated-region findings are minor accuracy/pointer nits inside freshly written text (the contingency path in question is explicitly not expected to occur), while the two unannotated findings are small enforcement gaps that have survived all three rounds. No revision introduced a contradiction, regression, or scope drift: every checkable factual addition in round 3 (effort estimate, vet-passed claim, credential chain, per-tool disposal, contaminator file:line list, flaky policy) verified exact against the codebase. No `conflict-with-pre-revision` tags were warranted — every revision direction moved toward more rigor and none contradicted rubric judgment.

## Result

```
SCORE: 904/1000
```

Target (900) met. All six iteration-2 attack points plus both blindspots were addressed with verified-accurate fixes; the residual 96 points are concentrated in items of diminishing severity: the mislabeled followup.md pointer in the in-flight-branch contingency (D4/D8, −2), SC-2/SC-3's unspecified mechanical grep pattern (D9, −2), the SC-less commit-isolation and baseline-archiving constraints (D9/D10, −2), unspecified baseline 留档 format (D4, −3 across constraints/NFR residuals), and the honest-zero creativity ceiling (D5, structural for a cleanup proposal). The two new blindspots are low-severity process observations, not blocking defects.
