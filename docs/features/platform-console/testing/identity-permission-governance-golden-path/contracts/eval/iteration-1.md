# Contract Evaluation Report — Iteration 1

- **Journey**: identity-permission-governance-golden-path
- **Surfaces**: web (haven-console) + api (eiam)
- **Rubric**: contract.md (1000 pts, target 850, per-dimension thresholds)
- **Scorer persona**: Senior QA Engineer (adversarial)
- **Iteration**: 1
- **Fact Table**: D:\Haven\.forge\fact-table.json (23 facts, IAM_* IDs)
- **Handbook**: absent → Anchor Integrity scored full marks (backward-compatible)

---

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem→Solution**: The Journey promises end-to-end identity+permission governance (create user → configure source → create role+bind policy → assign → revoke+converge). The 5 Contracts map 1:1 to these steps and carry the core domain actions. Solution addresses the stated problem. ✓
2. **Solution→Evidence**: Contracts cite both PRD-level facts (F1–F12 in journey.md) and code-level facts (IAM_* in fact-table.json). However the two ID schemes are not reconciled — several Contract Invariants cite F-* IDs that do not exist in the canonical Fact Table (which uses IAM_* IDs). This is a traceability defect, not a reasoning break.
3. **Evidence→Success Criteria**: Happy-path Outcomes assert behaviors whose code backing is marked UNKNOWN in the Fact Table — most critically Step 1 success bundles a password-reset side-effect whose endpoint "代码未实现" (IAM_USER_PASSWORD_RESET = UNKNOWN). A happy-path Outcome asserting an unimplemented operation as part of its success contract is a reasoning flaw (captured as a blindspot).
4. **Self-contradiction check**:
   - Cache-convergence value contradicts across documents: journey invariant "≤ 30 秒 (UNKNOWN)" vs Contract invariants "上限 60 秒 (casbin.go:85)". The Contract corrects the journey with code evidence, but the two artifacts disagree.
   - Forward cross-Contract reference: Step 3 `dependency-delete-blocked` Outcome preconditions on "Step 4 已建立授权关系" — a state created by a LATER step. This breaks linear executability.
   - Action-mixing: Step 3 (step-action `create-role-and-bind-policy`) bundles a delete-blocked Outcome for a different action (deleting role/policy); Step 4 (step-action `assign-role-to-new-user`) bundles a `cross-tenant-policy-denied` Outcome about policy-definition CRUD — a different action surface. This creates Outcome-selection ambiguity within a single Contract.

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 130/140

| Criterion | Score | Notes |
|---|---|---|
| Every Outcome has all four mandatory dimensions non-empty (Preconditions, Input, Output, State) | 50/50 | All 5 Contracts × all Outcomes carry non-empty Preconditions/Input/Output/State. Side-effect defaults to "none" where omitted. |
| Journey Invariants section present with ≥1 entry in every Contract | 50/50 | All 5 Contracts have `## Journey Invariants` with multiple entries. |
| Outcomes cover happy path + required derived scenarios per surface type | 30/40 | Journey-level mandatory coverage exists: `validation-error` (Step 1), `session-expired` (Step 4), `unauthorized` (Step 4). But API rule mandates `unauthorized` **per authenticated endpoint** — Steps 1, 2, 3, 5 each call an authenticated eiam endpoint (POST /api/user/create, /api/identity_source/save, /api/role/create, /api/role/unassign) yet none carries an `unauthorized` Outcome. Web `validation-error`/`session-expired` appear in only one step each. −10. |

### 2. Semantic Purity (语义纯度) — 128/170

| Criterion | Score | Notes |
|---|---|---|
| Dimension values use natural language, not code/regex | 48/70 | No regex/CSS/XPath/assertion calls present. But dimension values embed code symbols: API paths (`POST /api/role/batch_assign`), file:line refs (`role/handler.go:66`, `permission.go:24-26`, `checker.go:45-46`, `casbin.go:85`, `identity_source/handler.go:32`), internal identifiers (`middleware.CheckPermission`, `errs.ErrForbidden`, `casbin`). These appear in Input, State, Side-effect, Invariants. |
| Preconditions are declarative state descriptions, not procedural steps | 48/50 | Preconditions are uniformly declarative state ("管理员已登录且会话有效；…"). One mild procedural tint in Step 1 concurrent ("先提交者的改动已生效，后提交者基于过期版本提交") but reads as state. |
| No implementation coupling in dimension values | 32/50 | Mandatory dimensions are largely clean, but: Input carries API endpoint paths (Steps 2,3,4,5); State names internal library "casbin 策略重载" (Step 4); Side-effect (optional) is heavily coupled with file:line + function names. API paths are borderline-acceptable for api surface, but file:line + internal function names are implementation coupling in any surface. |

### 3. Precondition Exclusivity (前置条件互斥性) — 120/140

| Criterion | Score | Notes |
|---|---|---|
| Preconditions are distinct across Outcomes within each Step | 46/50 | Within each Contract, Outcome Preconditions are semantically distinct. Minor risk: Step 4 `session-expired` ("会话已超时过期") vs `unauthorized-401` ("令牌已过期/无效") overlap on "no valid session" but are distinguishable by surface (web vs api) and context. |
| Preconditions sufficient to uniquely select an Outcome | 36/50 | Action-mixing breaks selection: Step 3 (step-action=create) includes a delete-blocked Outcome whose Input is a different action; Step 4 (step-action=assign) includes a policy-CRUD-denied Outcome. Given the step's stated action, the extra Outcomes cannot be selected from the same Input — selection is ambiguous without an action discriminator. |
| Error/boundary Outcomes state triggering conditions explicitly | 38/40 | All boundary Outcomes state triggers (format violation, duplicate, concurrent stale version, backend unavailable, cache window, last-page sole record, mid-workflow expiry, missing creds, cross-tenant). Strong. |

### 4. Fact Alignment (事实依据) — 115/140

| Criterion | Score | Notes |
|---|---|---|
| Factual claims traceable to fact_id or marked UNKNOWN | 40/50 | Most Contract Invariants and Side-effects cite canonical IAM_* IDs (IAM_CONSOLE_DIRECT_EIAM_NO_BFF, IAM_USER_PASSWORD_RESET, IAM_POLICY_DEFINITION_WRITE_SCOPE, IAM_AUTH_CACHE, IAM_UNAUTHORIZED_401, etc.) and UNKNOWNs are explicitly marked. BUT Step 2 Invariants cite `fact: F11` and Step 3 Invariants cite `fact: F5, F12` — F-* IDs are NOT present in the canonical fact-table.json (which uses IAM_* IDs only). These citations are untraceable to the Fact Table. |
| Inferred claims have required_outcomes rule support and source: inferred | 40/50 | Derived Outcomes carry `<!-- source: inferred -->` and reasoning citing surface rules (validation-error, conflict, concurrent-edit, network-error, internal-error, session-expired, unauthorized, pagination) — good. However `dependency-delete-blocked` (Step 3) and `cross-tenant-policy-denied` (Step 4) are marked `source: inferred` but their reasoning cites a **fact** (F8, F12), not a `required_outcomes` rule — these are fact-derived, not rule-derived, and the classification is mis-tagged. |
| No hallucinated claims without classification | 35/40 | No unclassified hallucinations found; UNKNOWN marking is consistent (U1–U4, IAM_* UNKNOWN facts). Minor: the "casbin Redis watcher 近实时生效" claim in Step 4 Side-effect is backed by IAM_ROLE_ASSIGN but not explicitly cited there. |

### 5. Surface Fitness (Surface 适配) — 75/90

| Criterion | Score | Notes |
|---|---|---|
| Mandatory derived Outcomes from surface required_outcomes present | 25/35 | Journey-level: all three mandatory Outcomes (web validation-error, web session-expired, api unauthorized) appear at least once. But api `unauthorized` is required **per authenticated endpoint** — missing on Steps 1, 2, 3, 5. −10. |
| Dimension descriptions use surface-appropriate language | 25/30 | Web language (用户管理页, 表单, 标红提示, 弹窗, 列表) and api language (POST /api/…, 401/403) are used appropriately for the dual surface. Deduction for file:line references (`role/handler.go:66`, `permission.go:24-26`) which are surface-inappropriate for both web and api — they belong in anchor fields, not dimension prose. |
| TUI async timeout Outcomes | 25/25 | N/A (non-TUI surface) — full marks. |

### 6. Internal Consistency (一致性) — 112/140

| Criterion | Score | Notes |
|---|---|---|
| Invariants hold in every Step Contract | 44/50 | Journey invariant "撤销生效延迟 ≤ 30 秒 (UNKNOWN — see U2)" contradicts Contract invariants "上限 60 秒 (casbin.go:85，PRD ≤30 秒 UNKNOWN)". The Contract's 60s is code-backed and the journey's 30s is PRD-claimed-UNKNOWN — the two artifacts disagree on the canonical convergence ceiling. |
| Cross-Contract state references are consistent | 34/50 | Forward reference: Step 3 `dependency-delete-blocked` precondition "Step 3 创建的角色仍已分配给至少一个用户（Step 4 已建立授权关系）" depends on state created in a LATER step — breaks linear execution. Step 4→Step 5 references are consistent (授权关系 from Step 4 used in Step 5 ✓). |
| Outcome Preconditions consistent with preceding Steps' State changes | 34/40 | Step 5 success precondition "该用户当前具备该角色对应的权限" is achievable from Step 4's State ("授权关系写入 eiam；casbin 策略重载") ✓. But Step 5 `cache-delay-boundary` Outcome and Step 5 success both reference the same revoked state with different cache windows — the boundary between them is described in prose, not in a measurable precondition threshold. |

### 7. Anchor Integrity (锚点完整性) — 90/90

Handbook absent (no api-handbook.md / page-map.md). Per rubric activation condition, all criteria score full marks (backward-compatible). Anchor frontmatter fields are present but empty per graceful-degradation rule — no penalty.

### 8. Fixture Specification (前置数据声明) — 70/90

| Criterion | Score | Notes |
|---|---|---|
| Entity completeness (veto item) | 32/35 | Inline per-Outcome fixture_specs declare all operated entities (PlatformAdmin, PlatformUser, LdapServer, PolicyDefinition, Role, TenantAdmin, AuthorizationRecord). Veto NOT triggered. Minor: Step 5 bottom-level fixture_spec omits `AuthorizationRecord` (used only in the pagination Outcome's inline spec) — the bottom summary is incomplete but inline coverage exists. Entity types are plausible domain entities. |
| Relationship and constraint coverage | 16/30 | No Contract declares `relationship_type` or `parent_entity` for multi-entity relationships. Step 3 `dependency-delete-blocked` (Role + PolicyDefinition + PlatformAdmin with implied role↔policy↔user↔role references) and Step 4 success (PlatformUser + Role + bound_policy) all lack explicit relationship declarations — only `field_constraints` capture relationship hints implicitly. −14. |
| Minimum data quantity declarations | 22/25 | min_count values are adequate (Step 1 concurrent PlatformAdmin min 2 ✓; Step 5 pagination AuthorizationRecord min 1 ✓; Step 3 dependency Role/PolicyDefinition min 1 ✓). Minor: Step 1 bottom PlatformUser min_count 0 under-declares vs the conflict/concurrent Outcomes needing min 1, but inline per-Outcome specs compensate. |

---

## Phase 3 — Blindspot Hunt

- **[blindspot] Happy-path Outcome bundles an unimplemented operation.** Step 1 `success` Side-effect asserts "后续密码重置生成一次性凭据（重置端点代码未实现，UNKNOWN — see fact IAM_USER_PASSWORD_RESET)" — a happy-path Outcome should not assert a side-effect whose backing endpoint is admitted unimplemented. The password-reset sub-action should be split into its own step or the entire reset clause marked UNKNOWN, not folded into the success contract as if it executes.
- **[blindspot] Step 3 delete-blocked Outcome tests a different action than the step-action.** The Contract's step-action is `create-role-and-bind-policy`, yet `dependency-delete-blocked` exercises deletion of a role/policy. A QA engineer downstream cannot determine from the step-action which Input action to drive for that Outcome — this is a structural misalignment between scope (create) and success criteria (delete-blocked), not captured cleanly by any single dimension.

---

## Summary

| Dimension | Score | Max | Threshold | Pass |
|---|---|---|---|---|
| 完整性 | 130 | 140 | 85 | ✓ |
| 语义纯度 | 128 | 170 | 100 | ✓ |
| 前置条件互斥性 | 120 | 140 | 85 | ✓ |
| 事实依据 | 115 | 140 | 85 | ✓ |
| Surface 适配 | 75 | 90 | 55 | ✓ |
| 一致性 | 112 | 140 | 85 | ✓ |
| 锚点完整性 | 90 | 90 | 55 | ✓ |
| 前置数据声明 | 70 | 90 | 55 | ✓ |
| **Total** | **840** | **1000** | **850** | **✗ (below target)** |

All per-dimension thresholds met; total 840 < 850 target → **FAIL (iteration 1)**.
