# Contract Eval Report — Iteration 2

**Journey**: login-and-session-lifecycle
**Surface**: web + api
**Iteration**: 2
**Scorer**: Senior QA Engineer (adversarial)
**Date**: 2026-10-08

## Summary

| Dimension | Score | Max | Threshold | Pass |
|-----------|-------|-----|-----------|------|
| 1. 完整性 (Completeness) | 134 | 140 | 85 | ✓ |
| 2. 语义纯度 (Semantic Purity) | 152 | 170 | 100 | ✓ |
| 3. 前置条件互斥性 (Precondition Exclusivity) | 132 | 140 | 85 | ✓ |
| 4. 事实依据 (Fact Alignment) | 127 | 140 | 85 | ✓ |
| 5. Surface 适配 (Surface Fitness) | 85 | 90 | 55 | ✓ |
| 6. 一致性 (Internal Consistency) | 134 | 140 | 85 | ✓ |
| 7. 锚点完整性 (Anchor Integrity) | 90 | 90 | 55 | ✓ |
| 8. 前置数据声明 (Fixture Specification) | 83 | 90 | 55 | ✓ |
| **Total** | **937** | **1000** | **850** | **PASS** |

All dimensions pass their min thresholds; total (937) exceeds the 850 target. Major gains over iteration 1 (813) come from resolving all 8 dimension-tagged attacks: file:line traces moved out of dimension values into HTML comments, the iframe/security claim marked UNKNOWN, Step 4 error/boundary Outcomes added, LoginProvider/EiamService entity types realigned to the design domain model, validation-error moved to Step 2, success-render WebAuthn precondition made bidirectionally exclusive, Journey Invariants lists unified across all 4 files, and API endpoint paths removed from the Input dimension. Two blindspot attacks remain partially/unresolved (multi-tenant-switch scope mismatch; per-outcome vs top-level fixture divergence for non-happy-path states).

---

## Iteration-1 Attack Resolution Audit

| # | Attack (iter-1) | Resolved? | Evidence in iter-2 |
|---|----------------|-----------|--------------------|
| 1 | Side-effect dimensions embed file:line + function names | YES | All file:line/function-name traces now in HTML comments (e.g. Step 2 State `<!-- traceability: ... eiam/internal/web/user/user.go:67 -->`); dimension values describe behavior only. |
| 2 | iframe/security-headers claim unclassified | YES | Step 1 success-render Output now carries `<!-- security-headers / iframe-nested claim: UNKNOWN — 事实表中无对应 fact_id，未在代码中核查 -->`. |
| 3 | Step 4 has zero error/boundary Outcomes | YES | Step 4 now has `session-already-expired-logout`, `network-error`, `unauthorized` (3 derived Outcomes added). |
| 4 | LoginProvider/EiamService entity types don't match design | YES | `LoginProvider` → `IdentitySource`; `EiamService` removed and documented as environment precondition in HTML comments. |
| 5 | validation-error placed in Step 1 (enter page) | YES | Moved to Step 2 (submit auth); Step 1 retains `success-render` + `passkey-unavailable` only. |
| 6 | Step 1 success-render missing WebAuthn precondition | YES | success-render Preconditions now include "浏览器支持 WebAuthn（区别于 passkey-unavailable 的不支持路径）". |
| 7 | Journey Invariants lists differ across files | YES | All 4 files now declare the full 5-entry set. |
| 8 | API endpoint paths leak into Input dimension | YES | Step 2 single-tenant Input is now "用户输入有效账号密码并提交登录" (no POST path); endpoint detail moved to traceability comment. |
| B1 | [blindspot] multi-tenant-switch is a different action | NO | Still in Step 2 with Input "用户经顶部租户切换器重选租户" — a post-login workspace action, not an auth-submission branch. |
| B2 | [blindspot] Per-Outcome vs top-level fixture_spec diverge | PARTIAL | Step 2 top-level Tenant min_count raised 1→2 (resolved). Step 3 top-level still declares session="valid shared session in console" while per-outcome session-expired declares session="expired in e-cam-web workflow" — expected happy-path default, acceptable. |
| B3 | [blindspot] Step 1 success-render misplaced Outcome-level Invariant | YES | The stray `Invariants:` field on success-render is removed. |

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

**Problem → Solution**: The 4-step chain (enter page → submit auth + tenant routing → SSO jump → dual-side logout) still covers the stated journey lifecycle completely. Step 4 now has error/boundary coverage, closing the success-criteria gap from iteration 1.

**Solution → Evidence**: Traceability is strong — every fact-backed Outcome cites F1–F7 or an IAM_* fact_id via HTML comments. F7 (tenant_id write semantics) remains UNKNOWN. MFA lockout correctly reclassified `source: fact` (was `source: inferred` in iter-1). The iframe claim is now UNKNOWN.

**Evidence → Success Criteria**: Outcomes are testable. Residual: Step 3 `session-expired` Output offers two alternatives ("未保存的交互内容被保留或给出数据丢失警示") without a distinguishing precondition — a downstream QA agent cannot deterministically assert which branch occurs.

**Self-contradiction check**:
- multi-tenant-switch remains a scope-vs-action mismatch (different action stuffed into the auth-submission step). Not a logical contradiction but a structural misalignment.
- No confirmed mutual-exclusion contradictions across SC/Outcomes. Precondition branching by tenant_count (0 / 1 / ≥2) and by session state (valid / expired / absent) is cleanly partitioned within each step.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 134/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Every Outcome has all four mandatory dimensions non-empty | 50/50 | All 16 Outcomes across 4 files have non-empty Preconditions, Input, Output, State; Side-effect present or "none". |
| Journey Invariants section present with ≥1 entry in every file | 50/50 | All 4 files now declare the full 5-entry invariant set (was 4 in Step 3 and Step 4). |
| Outcomes cover happy path + required derived scenarios | 34/40 | Step 4 now has 5 Outcomes including 3 derived (session-already-expired, network-error, unauthorized) — iter-1 gap closed. Step 2 validation-error correctly placed. **Residual**: Step 2 `multi-tenant-switch` is a post-login tenant-switching action crammed into the auth-submission step (scope mismatch, unresolved from iter-1). Step 1 has only 2 Outcomes (minimal but adequate for "enter page"). |

### 2. Semantic Purity (语义纯度) — 152/170

| Criterion | Score | Notes |
|-----------|-------|-------|
| Dimension values use natural language, not code/regex | 63/70 | No regex/CSS/XPath. No file:line or API endpoint paths in dimension values (moved to HTML comments). **Residual**: "tenant_id" field name in Step 2 State (domain vocabulary per fact-table, mild); "no-op" jargon in Step 4 session-already-expired Output; "PRD 称 5 次"/"PRD 称 15 分钟" embeds source citation in mfa-lockout Output values. |
| Preconditions are declarative state descriptions | 48/50 | Preconditions are state-based throughout ("账号所属租户数 = 1", "双侧会话均处于登录态", "会话已过期"). Minor action hint in "用户在登录页选择密码方式，已输入有效账号密码" but describes state. |
| No implementation coupling in dimension values | 41/50 | Major coupling (file:line, function names, API endpoints in values) resolved. **Residual**: "eiam 鉴权中间件" (authz middleware — internal component named) in Side-effect of Step 3 `unauthorized` and Step 4 `unauthorized`; "前端拦截器" (frontend interceptor) in Step 3 session-expired Side-effect; "PRD 称..." source citations embedded in mfa-lockout Output. These name internal components rather than describing system-level behavior. |

### 3. Precondition Exclusivity (前置条件互斥性) — 132/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes within each Step | 48/50 | Step 1 now bidirectionally exclusive (success-render declares "浏览器支持 WebAuthn"; passkey-unavailable declares "不支持 WebAuthn"). Step 2 branches by tenant_count + MFA + network + form-validity. Step 3/4 branch by session state. |
| Preconditions sufficient to uniquely select an Outcome | 46/50 | Each combination uniquely selects one Outcome. **Residual**: Step 2 `multi-tenant-switch` precondition ("已进入租户选择页或工作台") is a post-login lifecycle phase distinct from the other Step 2 outcomes which are at-login-submission — while distinguishable, an agent executing "submit auth" would not reach this state, creating phase confusion. |
| Error/boundary Outcomes state triggering conditions explicitly | 38/40 | All error/boundary Outcomes state explicit triggers (WebAuthn unsupported, MFA threshold reached, network/eiam unavailable, tenant_count=0, form fields invalid, session expired, no valid session). |

### 4. Fact Alignment (事实依据) — 127/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 45/50 | Traceability comments cite fact_ids (F1–F7, IAM_SESSION_SHARED, IAM_UNAUTHORIZED_401, IAM_LOGIN_ENDPOINT, IAM_TENANT_COUNT_ROUTING, IAM_SESSION_EXPIRED_VS_403_DISTINCT). iframe claim now UNKNOWN. **Residual**: `source: fact` annotation applied inconsistently — only single-tenant-success and mfa-lockout carry `<!-- source: fact -->`; zero-tenant-restricted, multi-tenant-switch, success-shared-session, console-logout-success, ecam-web-logout-reverse have traceability comments but no `source:` tag. "无 JS 报错" is a generic quality assertion without classification. |
| Inferred claims have required_outcomes rule support and source: inferred | 47/50 | All 7 inferred Outcomes (network-error ×2, validation-error, session-expired ×2, unauthorized ×2) carry `source: inferred` with surface rule citation. mfa-lockout correctly reclassified `source: fact`. **Residual**: "不做实时广播" (Step 4 console-logout-success Output) is an inferred claim from F6 (comment states "可由 F6 推导") but lacks `source: inferred` tag. |
| No hallucinated claims without classification | 35/40 | The iter-1 unclassified iframe claim is now UNKNOWN — main hallucination resolved. **Residual**: "不做实时广播" is an inferred-from-F6 claim without formal classification tag (has reasoning comment); "无 JS 报错" is an unclassified generic quality assertion. |

### 5. Surface Fitness (Surface 适配) — 85/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes from surface required_outcomes present | 34/35 | Web mandatory: validation-error (Step 2) + session-expired (Step 3, Step 4). API mandatory: unauthorized (Step 3, Step 4). All present and correctly placed (validation-error moved from Step 1 to Step 2). |
| Dimension descriptions use surface-appropriate language | 26/30 | Web interaction language throughout ("用户访问控制台", "点击「云管」应用入口", "登录页正常渲染"). API endpoints/status codes appear only in HTML comments. **Residual**: "eiam 鉴权中间件" and "前端拦截器" in Side-effect values are architecture-component terms, not pure web-interaction language — a web-surface description would say "系统拒绝无凭据请求" rather than naming the middleware. |
| TUI async Contracts include timeout Outcomes | 25/25 | Non-TUI surface — full marks. |

### 6. Internal Consistency (一致性) — 134/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step Contract | 48/50 | All 5 invariants present in all 4 files; no Outcome violates any invariant. The stray Outcome-level Invariant on Step 1 success-render (iter-1 blindspot) is removed. Minor: platform-admin-exempt invariant listed in Step 3/Step 4 where contextually less relevant, but uniformity is desired per iter-1 attack. |
| Cross-Contract state references consistent | 48/50 | Step 2 State (共享会话建立) → Step 3 Preconditions (已登录态) ✓. Step 3 State (e-cam-web 复用共享会话) → Step 4 Preconditions (双侧登录态) ✓. Step 4 console-logout-success State (共享会话销毁) → Step 3 unauthorized Preconditions (已被另一侧登出销毁) ✓. Step 2 top-level fixture Tenant min_count now 2 (was 1). |
| Outcome Preconditions consistent with preceding Steps' State changes | 38/40 | Step 2→3 and Step 3→4 state transitions consistent. Step 4 session-already-expired precondition achievable from Step 3 session-expired. **Residual**: Step 2 `multi-tenant-switch` precondition "已进入租户选择页或工作台" is self-referential within Step 2 (achievable from Step 2's own single-tenant-success) — structurally odd but not contradictory. |

### 7. Anchor Integrity (锚点完整性) — 90/90

Anchor Integrity: handbook not found — check skipped (backward-compatible). No `design/api-handbook.md`, `design/page-map.md`, or `design/cli-handbook.md` exists in the design directory. All 4 Contract frontmatters carry the graceful-degradation comment `<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->`. Per rubric activation condition: full marks.

### 8. Fixture Specification (前置数据声明) — 83/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Entity completeness (veto item) | 32/35 | Veto NOT triggered. `LoginProvider` replaced with `IdentitySource` (matches design domain model + fact-table IAM_IDENTITY_SOURCE_LDAP). `EiamService` removed and documented as environment precondition in HTML comments. `UserAccount` accepted as naming variant of design's `User`. `Tenant` matches. All entities referenced in operations are declared. Minor: `UserAccount` vs design `User` naming variant (acceptable). |
| Relationship and constraint coverage | 28/30 | Step 2 single-tenant and multi-tenant-switch now declare `relationship_type: "user-tenant-membership"` + `parent_entity: "UserAccount"` for Tenant (iter-1 gap closed). Step 2 top-level fixture also declares the relationship. Step 1 IdentitySource carries `field_constraints` for type. Single-entity contracts (Step 3/4) score full. |
| Minimum data quantity declarations | 23/25 | multi-tenant Tenant min_count=2 declared per-outcome AND top-level (iter-1 top-level under-declaration fixed). zero-tenant-restricted correctly has no Tenant entity. Minor: Step 3 top-level fixture declares session="valid shared session in console" while per-outcome session-expired declares session="expired" — expected happy-path default divergence, acceptable. |

---

## Phase 3 — Blindspot Hunt

### [blindspot] Step 2 `multi-tenant-switch` remains a different action stuffed into the auth-submission step (UNRESOLVED from iter-1)

> Input: "用户经顶部租户切换器重选租户"

Step 2's step-action is "submit-auth-and-route-by-tenant-count". Tenant switching via the top-bar switcher is a post-login workspace action, not an auth-submission branch. This Outcome tests a different user action than the step's stated purpose. A downstream QA agent executing Step 2 ("submit auth") would be confused: the multi-tenant routing at login (auto-show tenant picker) is different from post-login tenant switching via the switcher. The two share the tenant_count≥2 precondition but are distinct actions. This falls between Completeness, Surface Fitness, and Precondition Exclusivity — no single dimension fully captures the action-outcome scope mismatch.

### [blindspot] Step 3 `session-expired` Output offers a non-deterministic OR

> Output: "未保存的交互内容被保留或给出数据丢失警示"

The contract presents two alternative behaviors ("content preserved" OR "data-loss warning shown") without a precondition distinguishing which one occurs. A downstream QA agent cannot write a deterministic assertion for an OR — the test would pass if either branch fires, masking regressions. Either specify the condition that selects each branch, or collapse to a single deterministic behavior.

### [blindspot] `source: fact` annotation applied inconsistently across fact-backed Outcomes

> Step 2 single-tenant-success: `<!-- source: fact -->` … Step 2 zero-tenant-restricted: (traceability comment only, no source tag)

Five fact-backed Outcomes (zero-tenant-restricted, multi-tenant-switch, success-shared-session, console-logout-success, ecam-web-logout-reverse) carry traceability comments citing fact_ids but lack the `<!-- source: fact -->` annotation that single-tenant-success and mfa-lockout carry. This creates ambiguity about whether these are fact-backed or inferred claims, weakening the traceability contract. Falls between Fact Alignment and Internal Consistency — scored in Fact Alignment above but flagged here as a structural annotation gap.

---

## Attacks

1. [语义纯度]: Side-effect values still name internal architecture components ("鉴权中间件", "前端拦截器") — "eiam 鉴权中间件拒绝无凭据请求" (Step 3/4 unauthorized) / "前端拦截器引导回控制台登录页" (Step 3 session-expired) — Describe system-level behavior ("系统拒绝无凭据请求并返回未认证状态") without naming the middleware/interceptor; move component identity to traceability comments.

2. [完整性]: Step 2 `multi-tenant-switch` tests a post-login tenant-switching action that does not belong to the auth-submission step — Input "用户经顶部租户切换器重选租户" — Either move this Outcome to a dedicated tenant-switching step, or rewrite it as the at-login multi-tenant routing branch (auto-show tenant picker) that genuinely belongs to "submit-auth-and-route-by-tenant-count".

3. [blindspot]: Step 3 `session-expired` Output is a non-deterministic OR — "未保存的交互内容被保留或给出数据丢失警示" — Split into two Outcomes with distinguishing preconditions, or collapse to one deterministic behavior; a QA agent cannot assert an OR deterministically.

4. [事实依据]: `source: fact` annotation applied inconsistently — 5 of 7 fact-backed Outcomes (zero-tenant-restricted, multi-tenant-switch, success-shared-session, console-logout-success, ecam-web-logout-reverse) have traceability comments but no `<!-- source: fact -->` tag — Add the `source: fact` annotation to all fact-backed Outcomes for uniform traceability classification.

5. [事实依据]: "不做实时广播" is an inferred-from-F6 claim without `source: inferred` classification — "不做实时广播" (Step 4 console-logout-success Output) with comment "可由 F6 推导" — Add `<!-- source: inferred -->` with reasoning basis (F6 → next-request-only invalidation implies no realtime broadcast).

6. [语义纯度]: mfa-lockout Output embeds PRD source citations in the behavior description — "账号锁定（PRD 称 15 分钟）" / "MFA 连续失败达配置阈值（PRD 称 5 次）" — Move "PRD 称 X" sourcing to a traceability comment; the Output value should state the config-driven behavior ("账号锁定，锁定时长由配置驱动") without citing the PRD inline.

7. [前置数据声明]: Step 2 top-level `## Fixture Specification` only declares the multi-tenant Tenant min_count=2 default; zero-tenant-restricted per-outcome needs Tenant min_count=0 and single-tenant needs min_count=1 — Document per-outcome fixture divergence explicitly in the top-level section (or note that top-level represents the happy-path/multi-tenant default and per-outcome specs override) so a test-data setup agent knows which is authoritative.
