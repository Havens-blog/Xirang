# Eval Report — login-and-session-lifecycle (Iteration 2)

**Scorer**: Senior QA Engineer (adversarial)
**Rubric**: journey.md (7 dims / 1000, target 850, pass = total ≥ 850 AND every dim ≥ threshold)
**Surface types declared**: `["web", "api"]`
**Date**: 2026-10-08

---

## Iteration-1 Attack Resolution Check

| # | Iteration-1 Attack | Resolved? | Evidence |
|---|---------------------|-----------|----------|
| 1 | tenant_id-in-JWT asserted as fact, not UNKNOWN | YES | F7 Fact Table row: `UNKNOWN — 待 Phase 0 核查写入语义`; Outcome 2.1 hedges: `F7（会话凭据写入语义待 Phase 0 核查，见 F7 UNKNOWN）` |
| 2 | Web/API mandatory derived Outcomes missing | YES | Outcome 1.3 `validation-error`, Outcome 3.2 `session-expired`, Outcome 3.3 `unauthorized` all present |
| 3 | `golden_path: false` vs body Happy Path | YES | Frontmatter now `golden_path: true`; Overview declares golden path explicitly |
| 4 | Golden Path denied by metadata | YES | Same as #3 |
| 5 | Expected Result couples to Redis/JWT/nginx | YES | No Redis/JWT/nginx tokens in step descriptions; technical detail moved to Fact Table |
| 6 | `source: inferred` framework unengaged | YES | Three inferred Outcomes carry `source: inferred — required_outcomes rule: web#.../api#...` |
| 7 | Single-Outcome-per-Step, no branch modeling | YES | Each Step now has 2–5 Outcomes with distinct Preconditions |
| 8 | Test type tags missing | YES | Each Step carries `**Test type**: @web-e2e` and/or `@api-functional` |
| 9 | No Fact Table / fact_id registry | YES | Fact Table with F1–F7 present, with fact_id / Source / Status columns |
| 10 | golden_path blindspot | YES | Same as #3 |

**All 10 iteration-1 attacks addressed.** Revision introduced new issues (see Phase 1/3 below).

---

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem→Solution fit**: Journey covers login→workbench→SSO→dual-logout lifecycle (Story 1 + Story 6). Step sequence directly addresses the dual-side session problem. Fit sound.

2. **Solution→Evidence**: Fact Table F1–F6 verified against PRD E#/Story references. F7 correctly marked UNKNOWN (DF002 tenant_id write pending Phase 0). **BUT** two outcomes cite fact_ids that don't support their claims:
   - Outcome 2.1 cites `F1, F7` — F1 is about login page rendering (E8), not about post-auth session establishment or single-tenant routing. The correct source E4 (single tenant skip selection page) is absent from the Fact Table entirely.
   - Outcome 3.1 cites `F6` — F6 is about logout destroying shared session (Story 6), not about SSO shared-session working. The correct sources (Story 1 AC, BF#6) have no Fact Table entry.
   These are broken traceability chains — the fact_id exists but the cited fact doesn't support the outcome's behavioral claim.

3. **Evidence→Success Criteria**: Outcomes now describe user-observable behavior for the most part. Residual technical tokens (`无 JS 报错`, `页面安全响应头存在`, `鉴权端点`) leak implementation-level concepts but don't cross into regex/CSS/assertion-call territory.

4. **Self-contradiction check**: `golden_path: true` now matches the body's Happy Path — resolved. Invariants hold across Steps except Invariant 2 (`凭据错误文案统一`) which declares credential-error non-leakage but no Outcome tests the wrong-credential (valid username + wrong password) scenario. No SC/InScope entries to cluster.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 163/175 (threshold 105) ✓

| Sub | Score | Notes |
|-----|-------|-------|
| Metadata (kebab name, risk_level justified) | 43/45 | Name kebab ✓; `golden_path: true` ✓; risk Medium justified via comment. `surface_keys` include `haven-console` (non-standard but acceptable). |
| Steps complete (name/action/outcome) | 65/70 | Every Step has User Action + multiple Outcomes with Preconditions and Results. Coherent ordered sequence. Minor: Outcome 2.1 (main path) lacks explicit precondition — relies on vague step-level `账号所属租户数决定后续分流分支`. |
| Happy + required derived Outcomes | 55/60 | All three mandatory derived Outcomes present (validation-error, session-expired, unauthorized). 7 additional boundary/error Outcomes (passkey unavailable, MFA lockout, network error, zero-tenant, multi-tenant, reverse logout, session-expired). Gaps: no wrong-credential Outcome (valid username + wrong password); no already-has-session redirect (PRD BF#1). |

### 2. Semantic Purity (语义纯度) — 147/175 (threshold 105) ✓

| Sub | Score | Notes |
|-----|-------|-------|
| Outcomes natural language, no code/regex | 55/70 | No regex/CSS/XPath/assertion calls. Residual technical tokens: `无 JS 报错` (browser console), `页面安全响应头存在` (HTTP headers), `鉴权端点拒绝请求并返回未认证状态` (API endpoint). Not in prohibited category but leak implementation-level concepts. |
| Preconditions declarative, not procedural | 52/55 | All Preconditions are declarative state descriptions (`未登录且浏览器不支持 WebAuthn`, `账号已启用 MFA，MFA 连续失败达 5 次`, `用户所属租户数 ≥ 2`). Clean. |
| No implementation coupling in Steps | 40/50 | Step 1 User Action: `登录页按 eiam 支持方式渲染` references backend service name `eiam`. Outcome 3.3: `鉴权端点` is API-level. Significantly improved from iteration-1 (Redis/JWT/nginx removed from step descriptions). |

### 3. Precondition Exclusivity (前置条件互斥性) — 95/130 (threshold 80) ✓

| Sub | Score | Notes |
|-----|-------|-------|
| Preconditions distinct across Outcomes within Step | 35/50 | Step 2: 2.2/2.3/2.4/2.5 have distinct explicit Preconditions ✓; but 2.1 (main path) has no explicit precondition — inherits step-level `账号所属租户数决定后续分流分支` without stating `租户数 = 1`. Step 1: 1.1 (main path) also lacks explicit precondition. |
| Preconditions sufficient to uniquely select Outcome | 30/45 | Without explicit `租户数 = 1` on Outcome 2.1, the main-path branch is only identifiable by reading the Result text (`单租户自动选定唯一租户`), not from Preconditions. Step 1: 1.2 and 1.3 Preconditions can co-occur (user without WebAuthn who submits invalid form) — not strictly competing but not cleanly exclusive. |
| No missing Preconditions for error/boundary Outcomes | 30/35 | All error/boundary Outcomes (1.2, 1.3, 2.2, 2.3, 2.4, 2.5, 3.2, 3.3, 4.2) state Preconditions ✓. Main paths (1.1, 2.1, 4.1) lack explicit Preconditions but are not error/boundary. |

### 4. Fact Alignment (事实依据) — 105/130 (threshold 80) ✓

| Sub | Score | Notes |
|-----|-------|-------|
| Factual claims traceable to fact_id or UNKNOWN | 33/50 | Most outcomes cite correct fact_ids (1.1→F1, 1.2→F1, 2.2→F2, 2.3→F3, 2.4→F4, 2.5→F5, 4.1→F6, 4.2→F6). F7 correctly UNKNOWN. **BUT**: Outcome 2.1 cites `F1` (login page rendering) for session establishment / single-tenant routing — F1 doesn't support this; E4 (single tenant skip) missing from Fact Table entirely. Outcome 3.1 cites `F6` (logout destroys session) for SSO shared-session working — F6 doesn't support this; Story 1/BF#6 have no Fact Table entry. Two broken traceability chains. |
| Inferred claims have required_outcomes rule + source: inferred | 42/45 | Three inferred Outcomes properly annotated: 1.3 `source: inferred — required_outcomes rule: web#validation-error`; 3.2 `web#session-expired`; 3.3 `api#unauthorized`. All carry `source: inferred` and cite the triggering rule. |
| No hallucinated unclassified claims | 30/35 | All claims classified (fact-backed or inferred). No unclassified hallucinations. Misattributions (F1 on 2.1, F6 on 3.1) are traceability defects, not unclassified claims. |

### 5. Surface Fitness (Surface 适配) — 115/130 (threshold 80) ✓

| Sub | Score | Notes |
|-----|-------|-------|
| Mandatory derived Outcomes present | 47/50 | Web `validation-error` (Outcome 1.3) ✓; Web `session-expired` (Outcome 3.2, timeout during active workflow not logout) ✓; API `unauthorized` (Outcome 3.3, no valid credentials on authenticated endpoint) ✓. All well-modeled and properly annotated. |
| Test strategy proportions (Web/API balanced 50/50) | 38/45 | Dual-surface tags present on every Step (`@web-e2e`, `@api-functional`). Step 1 web-only (correct — pure rendering). Steps 2–4 dual-tagged. Tags are at Step level, not Outcome level — a test generator can't tell whether Outcome 2.3 (network error) should be browser-driven or API-driven. |
| Surface-specific environment realistic | 30/35 | Web: browser, WebAuthn, keyboard navigation ✓. API: session validation, unauthorized access ✓. Setup mentions same-domain constraint ✓. No explicit headless/browser-automation framework mention (convention-level). |

### 6. Internal Consistency (一致性) — 116/130 (threshold 80) ✓

| Sub | Score | Notes |
|-----|-------|-------|
| Invariants hold in every Step | 42/50 | Invariants 1, 3, 4, 5 hold across all Steps ✓. **Invariant 2** (`登录页安全面不泄露用户名是否存在——凭据错误文案统一`) is declared but no Outcome tests the wrong-credential scenario (valid username + wrong password). Outcome 1.3 tests form validation (empty/invalid format), not credential failure. Outcome 2.3 mentions credential-error text under a network-error precondition — semantically mismatched. The most common login failure mode has no dedicated Outcome. |
| Cross-Step references consistent | 42/45 | Step 3 references shared session from Step 2 ✓. Step 4 references session from Steps 2/3 ✓. `golden_path: true` matches body Happy Path ✓ (iteration-1 contradiction resolved). Overview explicitly declares golden path sequence ✓. |
| Risk level consistent with content | 32/35 | Medium defensible — login/logout reversible, no irreversible data loss. Authentication/security angle could argue High, but Medium is within tolerance given session destruction is reversible (re-login). |

### 7. Workflow Coverage (工作流覆盖度) — 113/130 (threshold 80) ✓

| Sub | Score | Notes |
|-----|-------|-------|
| Golden Path existence (veto item) | 45/50 | Veto NOT triggered: Step 1 → Outcome 2.1 → Outcome 3.1 → Outcome 4.1 is a contiguous 4-step domain-level golden path mapping to Story 1 (login→workbench→SSO) + Story 6 (dual logout). `golden_path: true` ✓. Steps reference domain-level user operations (`进入登录页`, `提交认证并完成租户分流`, `单点跳转 e-cam-web`, `登出触发双侧会话失效`), not bare HTTP methods. |
| Multi-step coverage depth | 38/45 | State transitions: tenant 0/1/many ✓, session active/expired/destroyed ✓, MFA lockout ✓. Cross-entity: console↔e-cam-web session interaction ✓. Error recovery: network error retry ✓, MFA lockout recovery ✓. Reverse scenario: e-cam-web side logout (4.2) ✓. No entity-lifecycle CRUD (scope-appropriate). |
| Workflow completeness vs PRD scope | 30/35 | Stories 1 + 6 covered ✓. PRD BF#8 session expiry → Outcome 3.2 ✓. PRD BF#1 `若有有效会话直接进工作台` (returning user with active session) has no Outcome — Step 1 Precondition only covers `无有效会话`. PRD E2 (password-error lockout) not covered — only E9 (MFA lockout) via Outcome 2.2. 403 scenario defensibly out of scope (authorization, not authentication). |

---

## Phase 3 — Blindspot Hunt

- **[blindspot]**: F2 cites `E2/E9` but only describes MFA lockout — `"MFA 连续失败 5 次锁定账号 15 分钟 | prd-spec E2/E9"` — E2 covers password-error lockout (`连续 5 次密码错误锁定 15 分钟`) but F2's claim text only describes MFA failure. The password-failure lockout path (5 wrong passwords → lockout) has no dedicated Fact or Outcome. A tester following F2 would test MFA lockout but miss password lockout despite E2 being cited. Must either split F2 into separate password/MFA facts or add a password-lockout Outcome.

- **[blindspot]**: Outcome 2.3 conflates network error with credential-error non-leakage — Precondition `"登录提交时网络错误或 eiam 不可用"` but Result includes `"凭据错误文案不泄露用户名是否存在"` — credential-error text behavior is semantically irrelevant to the network-error precondition. While traceable to PRD Story 1 AC (which also conflates them), at the Journey level this makes the Outcome's verification scope ambiguous: a tester verifying network-error behavior would also need to verify credential-error non-leakage, which requires a different input scenario. Must separate credential-error non-leakage into its own Outcome under the correct wrong-credential precondition.

- **[blindspot]**: Test-type tags are at Step level, not Outcome level — `**Test type**: @web-e2e, @api-functional` on Step 2 — a dual-surface Step with 5 Outcomes gives a test generator no signal whether Outcome 2.3 (network error) should be a browser-driven `@web-e2e` test or an API-client `@api-functional` test. For a dual-surface Journey, per-Outcome surface tagging (or at minimum per-Outcome surface indication) would let the generator route correctly.

---

## Summary

**SCORE: 854/1000**
**PASS: YES** — total 854 ≥ 850; all dimensions above per-dimension thresholds.

```
SCORE: 854/1000
DIMENSIONS:
  完整性: 163/175
  语义纯度: 147/175
  前置条件互斥性: 95/130
  事实依据: 105/130
  Surface 适配: 115/130
  一致性: 116/130
  工作流覆盖度: 113/130
ATTACKS:
1. [事实依据]: F1 misattributed on Outcome 2.1 — "source: fact F1, F7" — F1 covers login page rendering (E8), not session establishment/single-tenant routing; E4 (single tenant skip selection page) missing from Fact Table entirely; must add Fact for E4 and correct the citation on Outcome 2.1
2. [事实依据]: F6 misattributed on Outcome 3.1 — "source: fact F6" — F6 covers logout destroying shared session (Story 6), not SSO shared-session working; must cite Story 1 AC / BF#6 and add corresponding Fact Table entry
3. [前置条件互斥性]: Outcome 2.1 main path lacks explicit precondition — Step 2 Precondition "账号所属租户数决定后续分流分支" without stating tenant=1 for Outcome 2.1 — must add "用户所属租户数 = 1" as explicit Precondition on Outcome 2.1 to distinguish from 2.4 (tenant=0) and 2.5 (tenant≥2)
4. [语义纯度]: Implementation-level tokens in outcome descriptions — "无 JS 报错" / "页面安全响应头存在" / "鉴权端点拒绝请求并返回未认证状态" — user-observable results embed browser-console/HTTP-header/API-endpoint concepts; must rephrase as pure user-observable behavior and move technical verification details to Fact Table or implementation notes
5. [一致性]: Invariant 2 untested by any Outcome — "登录页安全面不泄露用户名是否存在——凭据错误文案统一" — no Outcome covers wrong-credential (valid username + wrong password); validation-error (1.3) covers form validation only, MFA lockout (2.2) covers MFA failures; must add a wrong-credential Outcome verifying non-leakage
6. [完整性]: No Outcome for returning user with active session — PRD BF#1 "若有有效会话直接进工作台" uncovered; Step 1 Precondition "用户浏览器同域访问控制台，无有效会话" only covers no-session case; must add Outcome for already-authenticated redirect to workbench
7. [blindspot]: F2 cites E2/E9 but only describes MFA lockout — "MFA 连续失败 5 次锁定账号 15 分钟 | prd-spec E2/E9" — E2 covers password-error lockout but F2 omits it; password-failure lockout path has no dedicated Fact or Outcome; must split F2 or add password-lockout Outcome
8. [blindspot]: Outcome 2.3 conflates network error with credential-error non-leakage — Precondition "登录提交时网络错误或 eiam 不可用" but Result "凭据错误文案不泄露用户名是否存在" — credential-error text behavior is irrelevant to network-error precondition; must separate into its own wrong-credential Outcome
```
