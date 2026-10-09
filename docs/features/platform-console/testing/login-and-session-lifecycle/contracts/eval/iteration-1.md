# Contract Eval Report — Iteration 1

**Journey**: login-and-session-lifecycle
**Surface**: web + api
**Iteration**: 1
**Scorer**: Senior QA Engineer (adversarial)
**Date**: 2026-10-08

## Summary

| Dimension | Score | Max | Threshold | Pass |
|-----------|-------|-----|-----------|------|
| 1. 完整性 (Completeness) | 125 | 140 | 85 | ✓ |
| 2. 语义纯度 (Semantic Purity) | 114 | 170 | 100 | ✓ |
| 3. 前置条件互斥性 (Precondition Exclusivity) | 123 | 140 | 85 | ✓ |
| 4. 事实依据 (Fact Alignment) | 95 | 140 | 85 | ✓ |
| 5. Surface 适配 (Surface Fitness) | 70 | 90 | 55 | ✓ |
| 6. 一致性 (Internal Consistency) | 131 | 140 | 85 | ✓ |
| 7. 锚点完整性 (Anchor Integrity) | 90 | 90 | 55 | ✓ |
| 8. 前置数据声明 (Fixture Specification) | 65 | 90 | 55 | ✓ |
| **Total** | **813** | **1000** | **850** | **FAIL** |

All dimension thresholds pass individually, but the total (813) falls below the 850 target. The dominant drag is Semantic Purity (implementation coupling in Side-effect/Input dimensions) and Fact Alignment (one unclassified security claim).

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

**Problem → Solution**: The 4-step contract chain (enter login page → submit auth + tenant routing → SSO jump → dual-side logout) covers the stated journey lifecycle completely. The solution addresses the stated problem.

**Solution → Evidence**: Most outcomes trace to fact-table entries (F1–F7, IAM_*). F7 (tenant_id write semantics) is correctly marked UNKNOWN. MFA lockout correctly supersedes F2's hardcoded "5/15" with config-driven reality (IAM_LOGIN_ENDPOINT). Evidence support is strong.

**Evidence → Success Criteria**: Outcomes describe testable behavior. However, Step 4 (logout) has only 2 happy-path variants and zero error/boundary outcomes — a downstream QA agent cannot test "logout when session already expired" or "network error during logout" because those scenarios are absent.

**Self-contradiction check**:
- Step 2 `multi-tenant-switch` Outcome describes a post-login tenant-switch action ("用户经顶部租户切换器重选租户") stuffed into Step 2 whose action is "提交认证并完成租户分流". Switching tenants is a separate workflow action, not an auth-submission branch. This is a scope-vs-action mismatch, not a logical contradiction.
- Step 1 `validation-error` Outcome tests form-submission validation, but Step 1's action is "进入登录页并选择登录方式" (enter page). Form submission belongs to Step 2. The contract mirrors the journey's placement, but the action-outcome alignment is off.
- No confirmed mutual-exclusion contradictions across SC/Outcomes were found. Precondition branching by tenant_count (0 / 1 / ≥2) is cleanly partitioned.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 125/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Every Outcome has all four mandatory dimensions non-empty | 50/50 | All 13 Outcomes across 4 files have non-empty Preconditions, Input, Output, State. Side-effect present or "none". |
| Journey Invariants section present with ≥1 entry in every file | 50/50 | All 4 files have `## Journey Invariants` with 4–5 entries each. |
| Outcomes cover happy path + required derived scenarios | 25/40 | Steps 1–3 have rich boundary coverage. **Step 4 has zero error/boundary Outcomes** — only `console-logout-success` + `ecam-web-logout-reverse` (both happy-path variants). No derived outcomes for logout-when-session-already-expired, network-error-during-logout, or unauthorized-logout-call. Step 2 `multi-tenant-switch` is a different action (tenant switching) crammed into the auth-submission step; `validation-error` is placed in Step 1 (enter page) rather than Step 2 (submit form). |

### 2. Semantic Purity (语义纯度) — 114/170

| Criterion | Score | Notes |
|-----------|-------|-------|
| Dimension values use natural language, not code/regex | 45/70 | Primary dimensions (Preconditions, Output, State) are clean natural language. **Side-effect dimensions contain file:line references and internal function names** across Steps 2/3/4: `user.go:67`, `permission.go:24-26`, `identity_source.go:105-106`, `middleware.CheckPermission`, `LocalConfig.MaxFailedAttempts/LockoutDuration`. These are code artifacts in dimension values. No regex/CSS/XPath found. |
| Preconditions are declarative state descriptions | 47/50 | Preconditions are declarative ("账号所属租户数 = 1", "用户双侧会话均处于登录态"). Minor procedural hint in Step 1 passkey-unavailable Input but Preconditions themselves are state-based. |
| No implementation coupling in dimension values | 22/50 | API endpoint paths in Input: "直调 eiam POST /api/user/system/login" (Step 2), "直调 eiam POST /api/user/logout" (Step 4). File paths + function names in Side-effect: 6+ instances across Steps 2/3/4. Fact-id references (`PRD DF002`, `see F7`) embedded in State dimension. |

### 3. Precondition Exclusivity (前置条件互斥性) — 123/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes within each Step | 45/50 | Step 1 `success-render` does not encode "browser supports WebAuthn" while `passkey-unavailable` encodes "浏览器不支持 WebAuthn" — exclusivity is asymmetric (only one side declares the distinguishing condition). Steps 2/3/4 are cleanly partitioned. |
| Preconditions sufficient to uniquely select an Outcome | 42/50 | Step 1 `success-render` Preconditions ("无有效会话；eiam 支持密码/LDAP/passkey") could match a browser without WebAuthn support — the page still renders, creating overlap with `passkey-unavailable`. Step 2 branching by tenant_count is sufficient. |
| Error/boundary Outcomes state triggering conditions explicitly | 36/40 | All error/boundary Outcomes in Steps 1–3 state explicit triggers. Step 4 has no error/boundary Outcomes to evaluate. |

### 4. Fact Alignment (事实依据) — 95/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 40/50 | Strong traceability overall (F1–F7, IAM_SESSION_SHARED, IAM_UNAUTHORIZED_401, IAM_LOGIN_ENDPOINT, IAM_TENANT_COUNT_ROUTING all cited). **Step 1 `success-render` Output asserts "页面安全响应头存在，页面不可被 iframe 嵌套" with no fact_id and no UNKNOWN marking** — this is a specific security claim (security headers, iframe blocking) absent from the Fact Table. |
| Inferred claims have required_outcomes rule support and source: inferred | 45/50 | Step 1 `validation-error`, Step 2 `network-error`, Step 3 `session-expired` + `unauthorized` all correctly annotated `source: inferred` with surface rule citation. **Step 2 `mfa-lockout` is marked `source: inferred` but is actually fact-backed (F2 verified)** — it's a factual claim, not a surface-rule derivation. Misclassification. |
| No hallucinated claims without classification | 10/40 | The iframe/security-headers claim is neither factual (no fact_id) nor inferred (no required_outcomes rule support) — it is an unclassified claim. Per rubric: -30 pts per instance. "无 JS 报错" is a generic quality expectation (borderline). "不做实时广播" (Step 4) is inferable from F6 but uncited. |

### 5. Surface Fitness (Surface 适配) — 70/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes from surface required_outcomes present | 25/35 | Web mandatory: `validation-error` (Step 1) + `session-expired` (Step 3) present. API mandatory: `unauthorized` (Step 3) present. **Step 2 (web+api, the form-submission step) lacks `validation-error`** — it was placed in Step 1 (enter page) instead. **Step 4 (web+api, authenticated logout) has zero derived Outcomes** — no `unauthorized` for logout-without-session scenario. |
| Dimension descriptions use surface-appropriate language | 20/30 | Web language is appropriate ("用户访问控制台", "点击「云管」应用入口", "登录页正常渲染"). API language (POST endpoints, 401 status) appears in Side-effect for dual-surface steps — acceptable. **File:line references (`user.go:67`, `permission.go:24-26`) are surface-inappropriate for both web and api** — they are implementation debug artifacts, not surface interaction language. |
| TUI async Contracts include timeout Outcomes | 25/25 | Non-TUI surface — full marks. |

### 6. Internal Consistency (一致性) — 131/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step Contract | 45/50 | No invariant is violated by any Outcome. **Step 3 omits "平台管理员豁免零租户限制" from its Journey Invariants list (4 entries vs 5 in journey.md); Step 4 omits "登录页安全面不泄露用户名是否存在"** — these are list omissions, not violations, but break journey-level invariant uniformity. Step 1 `success-render` Outcome-level Invariant about credential-error no-leak is misplaced (no credential error occurs in the happy render). |
| Cross-Contract state references consistent | 48/50 | Step 3 Preconditions ("用户已在控制台工作台登录态") ← Step 2 State ("共享会话建立"). Step 4 Preconditions ("双侧会话均处于登录态") ← Step 3 State ("e-cam-web 复用控制台共享会话"). Step 3 `unauthorized` references "已被另一侧登出销毁" → Step 4 covers this. All cross-references resolve. |
| Outcome Preconditions consistent with preceding Steps' State changes | 38/40 | Step 2→3 and Step 3→4 state transitions are consistent. Step 2 `multi-tenant-switch` precondition "已进入租户选择页或工作台" is self-referential within Step 2 (achievable from Step 2's own outcomes) — structurally odd but not contradictory. |

### 7. Anchor Integrity (锚点完整性) — 90/90

No handbook file exists in the design directory (`design/` contains only `tech-design.md`, no `api-handbook.md` / `page-map.md` / `cli-handbook.md`). Per rubric activation condition: score full marks. Anchor fields in all 4 contract frontmatters are empty with graceful-degradation comment `<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->`.

### 8. Fixture Specification (前置数据声明) — 65/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Entity completeness (veto item) | 23/35 | Veto NOT triggered — no entity referenced in operations is missing from `fixture_spec.entities`. **Semantic mismatch**: `LoginProvider` (kind: password/ldap/passkey) does not match design's `IdentitySource` (type: 'ldap' only — password and passkey are login endpoints, not identity sources). `EiamService` is an infrastructure dependency, not a domain model entity (design defines User, Tenant, IdentitySource, Role, SessionClaims). `UserAccount` maps to design's `User` (naming variant, acceptable). `Tenant` matches. |
| Relationship and constraint coverage | 20/30 | `field_constraints` present where needed (credentials, tenant_count, mfa_enabled, session state, kind). **No `relationship_type` / `parent_entity` declared for multi-entity contracts** (Step 2 has UserAccount + Tenant association; Step 2 multi-tenant has UserAccount + 2 Tenants). Per rubric: -10 per missing relationship. |
| Minimum data quantity declarations | 22/25 | Per-Outcome min_counts are correct (multi-tenant Tenant min_count=2, zero-tenant has no Tenant). **Top-level `## Fixture Specification` summary in Step 2 declares Tenant min_count=1** — insufficient for the multi-tenant-switch scenario (needs ≥2). Per-outcome specs compensate but the top-level summary is under-declared. |

---

## Phase 3 — Blindspot Hunt

### [blindspot] Step 2 `multi-tenant-switch` is a different action stuffed into the auth-submission step

> "用户经顶部租户切换器重选租户"

Step 2's step-action is "submit-auth-and-route-by-tenant-count". Tenant switching is a post-login workspace action, not an auth-submission branch. This Outcome tests a different user action than the step's stated purpose. A downstream QA agent executing Step 2 would be confused: is this testing auth submission or tenant switching? This is a structural misalignment between scope (step action) and success criteria (outcome) that no single rubric dimension fully captures — it manifests partially in Completeness and Surface Fitness but the root cause is action-outcome scope mismatch.

### [blindspot] Per-Outcome `fixture_spec` blocks vs top-level `## Fixture Specification` section diverge

Step 2 per-Outcome `network-error` declares `EiamService` entity; the top-level `## Fixture Specification` section omits it. Step 2 per-Outcome `multi-tenant-switch` declares Tenant min_count=2; top-level says min_count=1. Step 3 per-Outcome `session-expired` declares session=expired; top-level says session="valid shared session". Two parallel fixture declarations with different content creates ambiguity about which is authoritative for test data setup. This falls between Fixture Specification and Internal Consistency dimensions.

### [blindspot] Step 1 `success-render` Outcome-level Invariant is semantically inapplicable

> Invariants: "凭据错误文案不泄露用户名是否存在（fact: F1, IAM_LOGIN_ERROR_NO_LEAK）"

This Invariant is attached to the happy-path render Outcome where no credential error occurs. The no-leak invariant is relevant to error Outcomes (validation-error, credential failure), not to successful page rendering. Attaching it to `success-render` is semantically misplaced — it doesn't cause a test failure but misleads a downstream agent about which Outcome validates this invariant.

---

## Attacks

1. [语义纯度]: Side-effect dimensions embed file:line references and internal function names as implementation coupling — "直调 eiam POST /api/user/system/login（user.go:67）" / "eiam middleware.CheckPermission 返回 401（permission.go:24-26）" — Move all code traceability (file paths, function names, line numbers) out of dimension values into HTML comments or a dedicated traceability annotation block; dimension values must describe system-level behavior only.

2. [事实依据]: Unclassified security claim with no fact_id and no inferred-rule basis — "页面安全响应头存在，页面不可被 iframe 嵌套" — Either cite a fact_id from the Fact Table (or add one via code reconnaissance) or mark the claim `UNKNOWN`; unclassified behavioral assertions violate zero-tolerance for hallucinated claims.

3. [完整性]: Step 4 has zero error/boundary Outcomes — only "console-logout-success" and "ecam-web-logout-reverse" happy-path variants exist — Add derived Outcomes: logout when session already expired (no-op or redirect), network-error during logout call, unauthorized logout endpoint call without valid session.

4. [前置数据声明]: LoginProvider and EiamService entity_types do not match design domain model entities — "entity_type: 'LoginProvider'" (design has `IdentitySource` with type 'ldap' only) and "entity_type: 'EiamService'" (infrastructure, not a domain entity) — Align entity_type values with the design's domain model; use `IdentitySource` for login-method entities and model eiam reachability as an environment precondition rather than a fixture entity.

5. [Surface 适配]: validation-error Outcome placed in Step 1 (enter login page) instead of Step 2 (submit auth) — "用户提交登录表单时必填字段为空或字段格式非法" is a form-submission behavior — Move validation-error to Step 2 whose action is form submission, or add a justification for why form validation is tested under the page-entry step.

6. [前置条件互斥性]: Step 1 success-render does not encode the WebAuthn-support condition that distinguishes it from passkey-unavailable — success-render Preconditions say "eiam 支持密码/LDAP/passkey 登录方式" (eiam capability) but omit browser WebAuthn support; passkey-unavailable says "浏览器不支持 WebAuthn（区别于支持 WebAuthn 的成功路径）" — Add "浏览器支持 WebAuthn" to success-render Preconditions to make the pair bidirectionally exclusive.

7. [一致性]: Journey Invariants lists differ across contract files — Step 3 has 4 invariants (drops "平台管理员豁免零租户限制"), Step 4 has 4 (drops "登录页安全面不泄露用户名是否存在") — All 4 contract files should declare the full journey-level invariant set (5 entries from journey.md) for uniformity, even if some are contextually less relevant to a given step.

8. [语义纯度]: API endpoint paths leak into Input dimension — "用户输入有效账号密码提交登录（直调 eiam POST /api/user/system/login）" — Input should describe the user interaction ("用户输入有效账号密码并提交登录"); move the API endpoint detail to Side-effect or a separate API-surface annotation.
