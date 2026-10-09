# Journey Eval — iteration 1

**Journey**: readonly-audit-access
**Rubric**: journey.md (1000 pts, target 850)
**Scorer persona**: Senior QA Engineer (adversarial)
**Date**: 2026-10-08

## SCORE: 570/1000

**Pass condition**: FAIL (total 570 < 850; multiple dimensions below min threshold)

## DIMENSIONS

| Dimension | Score | Max | Min | Pass? |
|-----------|-------|-----|-----|-------|
| 完整性 (Completeness) | 115 | 175 | 105 | PASS (barely) |
| 语义纯度 (Semantic Purity) | 140 | 175 | 105 | PASS |
| 前置条件互斥性 (Precondition Exclusivity) | 75 | 130 | 80 | FAIL |
| 事实依据 (Fact Alignment) | 55 | 130 | 80 | FAIL |
| Surface 适配 (Surface Fitness) | 70 | 130 | 80 | FAIL |
| 一致性 (Internal Consistency) | 115 | 130 | 80 | PASS |
| 工作流覆盖度 (Workflow Coverage) | 0 | 130 | 80 | FAIL (veto) |
| **Total** | **570** | **1000** | **850** | **FAIL** |

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Golden Path veto self-declared.** Frontmatter `golden_path: false` plus only 2 Happy Path steps (< 3) — the Golden Path veto is structurally triggered regardless of content quality. The document opts out of the dimension's non-negotiable requirement.

2. **No Fact Table exists in DOC_DIR.** `find` for `*fact*` under `platform-console` returns nothing. The journey makes factual claims ("eiam 返回 403", "菜单入口不渲染", "403 原地收敛不跳转登录页") but provides no `fact_id` traceability and no `UNKNOWN` / `source: inferred` annotations.

3. **Asymmetric surface coverage.** `surface_types: ["web","api"]` declares both surfaces, but mandatory Web derived outcomes (`validation-error`, `session-expired`) are absent while API's `unauthorized` (the 403 path) is extensively covered. The invariant even references session-expired semantics ("与「会话过期跳登录」语义区分") yet provides no Outcome testing it.

4. **Structural inconsistency: Happy Path steps lack Preconditions.** Edge Cases (1b/2b/1c) carry a `Precondition` field; Happy Path Steps 1–2 do not, relying only on `Setup`. This breaks the rubric's per-step precondition model and weakens Precondition Exclusivity scoring.

## Phase 2 — Dimension Scoring

### 1. Completeness — 115/175
- Metadata (40/45): name `readonly-audit-access` kebab-case ✓; risk_level `Low` justified by read-only observational content ✓. Minor: `golden_path: false` flag actively disqualifies a required element.
- Steps complete (50/70): Each step has User Action + Expected Result. But each step carries exactly ONE outcome — no branching outcomes per step, so the multi-outcome model the rubric expects is absent. Steps form a coherent ordered sequence.
- Outcomes cover happy + derived (25/60): Happy path covered. Mandatory Web derived outcomes `validation-error` and `session-expired` (per surface-web.md) are absent. API `unauthorized` is present as 403. Two of three mandatory derived outcomes missing → heavy deduction.

### 2. Semantic Purity — 140/175
- Natural language (60/70): Outcomes in Chinese natural language; no regex/CSS/XPath/framework assertion calls. `403` is a status code but acceptable as system-level observable behavior.
- Preconditions declarative (45/55): Edge-case preconditions are declarative state ("只读角色绕过前端裁剪，直接调用任意写接口"). OK.
- No implementation coupling (35/50): Step/action text references the specific service `eiam` and "写接口" repeatedly. `eiam` is an implementation/service name, not a user-level concept; "调用写接口" is API-level rather than domain-level. Mild-to-moderate coupling.

### 3. Precondition Exclusivity — 75/130
- Distinct across Outcomes (30/50): Each step has only one outcome, so exclusivity is vacuously satisfied — but the structure never demonstrates multi-outcome exclusivity, which is what the dimension tests.
- Sufficient to uniquely select (30/45): Single-outcome steps trivially select; no ambiguous pairs. But no real exercise of the criterion.
- Missing Preconditions for error/boundary (15/35): Happy Path Steps 1–2 have NO `Precondition` field at all (only global Setup). Error/boundary edge cases do state preconditions, but the happy path's missing preconditions is a structural gap.

### 4. Fact Alignment — 55/130
- Factual claims traceable to fact_id or UNKNOWN (20/50): No Fact Table document exists in DOC_DIR; no `fact_id` references in the journey; no `UNKNOWN` markings on claims like "eiam 返回 403", "数据不变", "不跳转登录页". The frontmatter `sources` lists PRD files (verified: prd-user-stories.md Story 5 AC matches), but that is general grounding, not per-claim fact_id traceability.
- Inferred claims have rule support + `source: inferred` (15/45): The 403/unauthorized outcomes are derivable from surface-api.md `required_outcomes` (unauthorized) and PRD AC, but none carry `source: inferred` or cite the triggering rule. Boundary outcomes (1c "查询不触发状态变更") are reasonable inferences with no annotation.
- No hallucinated unclassified claims (20/35): Claims are not hallucinations (PRD-anchored) but they are unclassified — neither factual-with-fact_id nor inferred-with-annotation. Partial credit only.

### 5. Surface Fitness — 70/130
- Mandatory derived Outcomes present (20/50): API `unauthorized` present. Web `validation-error` ABSENT. Web `session-expired` ABSENT (referenced only as an invariant contrast, no Outcome). Two of three mandatory derived outcomes missing → near-zero on this sub-criterion.
- Test strategy proportions (25/45): surface-web.md and surface-api.md both mandate balanced 50/50 Contract/Journey. The journey is pure narrative with no indication of contract-level coverage split; impossible to verify the 50/50 balance from the document.
- Surface-specific environment realistic (25/35): Mixes browser-URL access (Web) and 403 responses (API) realistically; invariant correctly distinguishes 403-stay from session-expired-redirect. Reasonable but shallow.

### 6. Internal Consistency — 115/130
- Invariants hold in every Step (45/50): Four invariants (no write entries rendered, 403 final boundary, zero side-effects, 403-stays-not-login) are not violated by any step. Invariant 4 ("403 原地收敛不跳转登录页") is reinforced across steps 1b/2b.
- Cross-Step references consistent (35/45): Steps do not heavily reference each other; no dangling references. Minimal but consistent.
- Risk level consistent (35/35): Low risk + read-only workflow — fully consistent.

### 7. Workflow Coverage — 0/130 (VETO TRIGGERED)
- Golden Path existence (0/50 → veto): Frontmatter `golden_path: false`. Happy Path has only 2 steps (< 3 required). The Golden Path veto is explicitly and structurally triggered. Per rubric: "If the Golden Path existence criterion scores 0, the entire Workflow Coverage dimension scores 0 regardless of other sub-scores."
- Multi-step coverage depth (n/a, 0): Vetoed.
- Workflow completeness against PRD/Design (n/a, 0): Vetoed. (Note: Story 5 is covered, but the veto overrides.)

## Phase 3 — Blindspot Hunt

- **[blindspot]** Golden Path veto is self-inflicted: frontmatter `golden_path: false` — "golden_path: false" — the author explicitly declares no golden path, which is the single non-negotiable veto item. Reviser must either construct a 3+ step contiguous Golden Path tied to Story 5 or justify why this journey is exempt (it is not — Story 5 is a primary user workflow).

- **[blindspot]** No Fact Table artifact exists anywhere under `platform-console` — DOC_DIR has only `journey.md` and an empty `contracts/` dir. The rubric's Fact Alignment dimension assumes a Fact Table; its absence makes every factual claim structurally untraceable. Reviser must either add a Fact Table or mark all claims `UNKNOWN`/`source: inferred` with rule citations.

- **[blindspot]** Web surface declared but Web-mandatory outcomes absent while invariant invokes session-expired semantics — "403 原地收敛不跳转登录页——与「会话过期跳登录」语义区分；只读越权不丢失登录态" — the invariant explicitly contrasts against session-expired behavior, proving the author was aware of the `session-expired` required_outcome rule (surface-web.md) but did not generate a corresponding Outcome. This is a deliberate omission of a mandatory derived outcome, not an oversight.

- **[blindspot]** Reasoning audit flagged independently: the Journey solves the "403 enforcement" half of Story 5 thoroughly but only superficially addresses the "查看身份与权限页面以完成合规审计" half — Step 1's Expected Result is one line ("增删改入口不渲染...可正常浏览、搜索、分页查看") with no Outcome verifying the audit-completeness (are all required pages readable? is the data complete enough for compliance?). The audit-purpose of the role is asserted in Overview but not tested in any Outcome.

## ATTACKS (summary)

1. [工作流覆盖度]: Golden Path veto triggered by self-declared `golden_path: false` and only 2 Happy Path steps — "golden_path: false" + "### Step 1... ### Step 2" (no Step 3) — must add a contiguous 3+ step Golden Path grounded in Story 5's audit workflow, or the dimension stays 0 and the journey cannot pass.

2. [事实依据]: No Fact Table exists and no claim carries `fact_id`/`UNKNOWN`/`source: inferred` — "eiam 返回 403，前端提示无权限且数据不变；写操作不生效" — every factual claim (403 behavior, data-unchanged, no-login-redirect) must be traceable to a fact_id or annotated `UNKNOWN`/`source: inferred` with the triggering required_outcomes rule.

3. [Surface 适配]: Web mandatory derived outcomes `validation-error` and `session-expired` absent despite `surface_types: ["web","api"]` — "surface_types: [\"web\", \"api\"]" vs. missing outcomes — must add Web `validation-error` and `session-expired` Outcomes per surface-web.md `required_outcomes`.

4. [前置条件互斥性]: Happy Path steps have no `Precondition` field while Edge Cases do — "### Step 1... **User Action**..." (no Precondition) vs "### Step 1b... **Precondition**: ..." — must add explicit Preconditions to Happy Path Steps 1–2 so per-step exclusivity is evaluable.

5. [完整性]: Each step has exactly one Outcome; no step exercises multi-outcome branching — "**Expected Result**: 增删改入口不渲染..." (single outcome) — must add boundary/error Outcomes per step (e.g., Step 1 should branch on data-present vs data-empty, network-error on search) so the multi-outcome model is exercised.

6. [一致性]: `golden_path: false` frontmatter contradicts the Journey rubric's Golden Path requirement and the journey's own Overview which describes a complete audit workflow — "golden_path: false" vs "Overview: 只读/审计角色登录后查看...以完成合规审计" — must resolve: either set `golden_path: true` with a real 3+ step path or restructure.

7. [语义纯度]: Steps repeatedly name the implementation service `eiam` rather than domain-level concepts — "eiam 返回 403", "eiam 服务可用", "eiam 对所有写端点做服务端权限校验" — Steps should describe system/user-level behavior; service names belong in Setup/invariants, not in every action/expected-result.

8. [blindspot]: Audit-purpose untested — Overview claims "以完成合规审计" but no Outcome verifies audit-completeness (all required pages readable, data sufficient) — "以完成合规审计，但不渲染任何增删改入口" — must add an Outcome asserting the audit goal is achievable (readable across user/role/auth pages with complete data), not just that writes are blocked.

9. [blindspot]: No Fact Table artifact exists in DOC_DIR (only `journey.md` + empty `contracts/`) — the rubric's Fact Alignment dimension is structurally unevaluable — must add a Fact Table file or explicitly mark all claims as inferred with rule citations.

10. [blindspot]: Reasoning audit — journey over-solves the "write-denial" half of Story 5 and under-solves the "view-for-compliance" half; success criteria proxy (403 + no-write-entries) tests the negative constraint but not the positive audit capability — "查看身份与权限页面以完成合规审计" vs Outcomes that only assert writes are blocked — must add positive Outcomes verifying read/audit capability is complete.
