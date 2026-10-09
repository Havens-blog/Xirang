# Contract Evaluation Report — Iteration 1

**Journey**: tenant-boundary-enforcement
**Surface types**: web, api
**Documents evaluated**: 6 Contract files (step-1 through step-6)
**Rubric**: contract.md (1000 pts, target 850)
**Expert persona**: Senior QA Engineer
**Date**: 2026-10-08

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

### Problem → Solution
The journey addresses tenant boundary enforcement in a multi-tenant IAM system. The 6 contracts cover: viewing users within tenant (Step 1), creating roles + binding policy (Step 2), per-request role assignment (Step 3), cross-tenant access denial (Step 4), HTTP direct API governance (Step 5), concurrent cross-tenant edit isolation (Step 6). Coverage is comprehensive — the solution set addresses the stated problem of verifying server-side tenant boundary enforcement.

### Solution → Evidence
Contracts cite both journey-local fact IDs (F1–F10, I1–I4) and global Fact Table IDs (IAM_*). However, two critical facts are UNKNOWN or PARTIAL yet presented as definitive in behavioral Output fields:
- **IAM_VERSION_CONFLICT_409** = UNKNOWN (no 409 code path in eiam). Step 6 Output claims "若同租户内同一对象并发触发 409 版本冲突，则提示刷新后重试" as expected behavior. The UNKNOWN marking appears only in Side-effect, not in Output.
- **IAM_POLICY_DEFINITION_WRITE_SCOPE** = PARTIAL (denial depends on grants, not a hard scope rule). Steps 4 and 5 Output present "eiam 返回 403" as guaranteed behavior.

### Evidence → Success Criteria
Most Output descriptions are testable ("返回 403 Forbidden", "列表仅显示本租户用户"). However, Step 5's in-tenant-200 Outcome is overly broad — "按各 CRUD 语义返回 200/201/204" covers all CRUD operations in a single Outcome, making it untestable as a single scenario. Step 3's "下一次请求即具备对应权限" doesn't account for the 60s casbin convergence ceiling (IAM_AUTH_CACHE), risking flaky test execution.

### Self-Contradiction Check
No intra-Step or cross-Step contradictions found. The 403-in-place vs session-expired-redirect distinction is consistently maintained across all contracts. Step 3's precondition ("Step 2 创建的角色已存在") is achievable from Step 2's State ("新角色写入本租户").

### SC Consistency Deep-Dive
Clustered SC/InScope entries by affected area (eiam governance endpoints). Bidirectional derivation performed on all Outcome pairs within Steps 4 and 5 (the multi-Outcome contracts). No mutual-exclusion, direction-clash, or resource-contention contradictions detected. Preconditions within each Step are distinguishable by session validity, credential validity, and target ownership.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 123/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Every Outcome has 4 mandatory dimensions | 48/50 | All Outcomes have non-empty Preconditions, Input, Output, State. Minor: some Outcomes lack Invariants (optional). |
| Journey Invariants section ≥1 entry | 50/50 | All 6 contracts have 5 invariant entries each. |
| Happy path + required derived scenarios | 25/40 | Web Steps 1–3 have only happy-path Outcomes — no validation-error or session-expired derived Outcomes. Step 4 (web) has 4 Outcomes including derived. Step 5 (API) has unauthorized. Derived scenarios are consolidated into Steps 4/5 rather than present per-Contract. |

### 2. Semantic Purity (语义纯度) — 107/170

| Criterion | Score | Notes |
|-----------|-------|-------|
| Natural language, not code/regex | 45/70 | No regex/CSS selectors. Preconditions, Output mostly NL. But Side-effect fields pervasively contain code references (file paths, function names). Input fields in Steps 2, 3 contain API endpoint paths. |
| Preconditions declarative | 42/50 | Preconditions are declarative state descriptions throughout. Minor: some contain distinguishing-clause parentheticals ("区别于...") which are procedural hints. |
| No implementation coupling | 20/50 | Pervasive coupling: `ctxutil.GetTenantID (checker.go:29)`, `errs.ErrForbidden (checker.go:45-46)`, `POST /api/role/create (role/handler.go:41)`, `middleware.CheckPermission (permission.go:24-26)`, `casbin.go:77,85`, `IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL`. ~13 instances across Side-effect, Input, and State fields. |

### 3. Precondition Exclusivity (前置条件互斥性) — 123/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes | 46/50 | Steps 4, 5 (multi-Outcome) have distinct Preconditions. Each Outcome distinguishes by session validity, credential validity, or target ownership. |
| Preconditions sufficient to uniquely select | 42/50 | Minor ambiguity: Step 5 cross-tenant-403 ("凭据有效且归属本租户，但请求目标属于其他租户") vs policy-def-write-403 ("凭据有效但操作者为租户管理员，目标为策略定义写操作") — a tenant admin requesting a policy-definition write that happens to be in another tenant's scope could match both. Distinguishable by target type (tenant resource vs system-scoped policy definition), but not explicitly exclusive. |
| Error/boundary Outcomes state triggering conditions | 35/40 | All error Outcomes explicitly state triggering conditions. session-expired, validation-error, unauthorized-403, cross-tenant-403 all have explicit Preconditions. |

### 4. Fact Alignment (事实依据) — 96/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or UNKNOWN | 38/50 | Contracts cite F1–F10 (journey fact table) and IAM_* (global fact table). Traceable. However, Step 6 Output claims 409 behavior without UNKNOWN marking in the Output field — UNKNOWN is only in Side-effect. Step 4/5 policy-def Outcomes present 403 as definitive despite PARTIAL fact. |
| Inferred claims have required_outcomes support + source: inferred | 30/50 | session-expired (Step 4), validation-error (Step 4), unauthorized-401 (Step 5) have proper `<!-- source: inferred -->` + `<!-- reasoning: ... required_outcomes ... -->` annotations. However, Outcomes citing I1 (forbidden-403, cross-tenant-403, policy-def-entry-unreachable, policy-def-write-403) lack their own source annotation — they rely on I1 being annotated in the journey, not at the Outcome level. |
| No hallucinated claims without classification | 28/40 | Step 6 Output's 409 claim is effectively unclassified in the Output field (UNKNOWN marking is in Side-effect only). Step 4/5 policy-def 403 claims are presented as definitive when the backing fact is PARTIAL. These are borderline hallucinations — claims presented with more certainty than the evidence supports. |

### 5. Surface Fitness (Surface 适配) — 67/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes present | 22/35 | Web: validation-error + session-expired present only in Step 4, absent in Steps 1–3. API: unauthorized present in Step 5. Partial coverage — derived Outcomes are consolidated rather than per-Contract. |
| Surface-appropriate language | 20/30 | Web steps (1–4) use web-appropriate language in Input/Output ("用户管理页", "页面内提示无权限并保持当前页"). API steps (5–6) use HTTP status codes appropriately. But Side-effect fields across all contracts use backend implementation language (function names, file paths) rather than surface-appropriate descriptions. |
| TUI async timeout Outcomes | 25/25 | Non-TUI surface. Full marks. |

### 6. Internal Consistency (一致性) — 128/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Contract | 48/50 | All 5 invariants (tenant boundary, 403 convergence, 401/403 separation, policy definition scope, 403 no-existence-leak) are respected across all 6 contracts. No violations. |
| Cross-Contract state references consistent | 45/50 | Step 3 references "Step 2 创建的角色已存在" — Step 2's State confirms "新角色写入本租户". Consistent. Minor: no explicit cross-reference to Step 2's role identifier. |
| Outcome Preconditions consistent with preceding State changes | 35/40 | Step 3's precondition (role exists) is achievable from Step 2's State (role written). Consistent. All other steps are self-contained or reference preceding steps correctly. |

### 7. Anchor Integrity (锚点完整性) — 90/90

No handbook file exists. Per backward-compatible rule, full marks.

> Handbook not found — anchor integrity check skipped (backward-compatible).

### 8. Fixture Specification (前置数据声明) — 58/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Entity completeness (veto item) | 20/35 | Steps 5, 6 use abstract entity types (`Caller`, `TargetResource`, `EditableObject`) that do not match domain model entities (User, Role, Policy, Tenant). Step 4 forbidden-403 bottom fixture_spec lists only PlatformUser but Input references users/roles/authorizations as possible targets. Dual fixture_spec declarations (inline per-Outcome vs bottom section) are inconsistent across contracts. Veto NOT triggered: entities are non-empty and pre-existing referenced entities are mostly present; created entities (Role in Step 2, authorization in Step 3) are exempt as operation outputs. |
| Relationship and constraint coverage | 18/30 | Step 1 declares `relationship_type: belongs_to` for PlatformUser→Tenant. Step 3 lacks relationship between Role and PlatformUser (the assignment target). Step 6 lacks relationship between TenantAdmin and EditableObject (which admin owns which object). Step 2 lacks Role→PolicyDefinition binding relationship. |
| Minimum data quantity | 20/25 | Step 1 bottom fixture_spec declares PlatformUser min_count 1 but the Contract requires ≥2 users (one per tenant). Step 6 correctly declares min_count 2 for both admins and objects. Others are adequate. |

---

## Phase 3 — Blindspot Hunt

### [blindspot] 1: Step 5 in-tenant-200 Outcome is untestably broad
Step 5's `in-tenant-200` Outcome covers all CRUD operations (user/role/authorization/policy) in a single Outcome: "调用方以有效凭据访问本租户治理资源（用户/角色/授权 CRUD）" with Output "按各 CRUD 语义返回 200/201/204". A downstream test agent cannot write a single test that validates "all CRUD semantics" — each operation (GET list, POST create, PUT update, DELETE) has distinct request/response shapes. This Outcome should be split per CRUD operation or scoped to a representative operation. No rubric dimension captures "testability of a single Outcome covering multiple operations."

### [blindspot] 2: Step 3 ignores casbin convergence ceiling (60s)
Step 3's Output claims "该用户下一次请求即具备对应权限，逐请求判定、无需重新登录". The backing fact IAM_AUTH_CACHE documents a convergence ceiling of 60s (`enforcer.StartAutoLoadPolicy(time.Minute)`) with near-real-time via Redis watcher. A test that immediately sends the "next request" after role assignment may fail if the casbin policy hasn't reloaded yet. The Contract doesn't specify a wait/retry strategy, risking flaky test execution. This timing/reliability gap is not covered by any rubric dimension.

### [blindspot] 3: Web Steps 1–3 lack loading-state Outcomes for async operations
Steps 1–3 are web-e2e steps involving page loads and form submissions that trigger async API calls. The web surface rule lists "loading-state" as a common boundary Outcome ("Async operation in progress. Assert: loading indicator visible"). No web Contract includes a loading-state Outcome. While loading-state is not in the mandatory `required_outcomes` list, its absence across all web Contracts means the test plan won't verify that users see appropriate feedback during async operations — a common source of user-reported bugs. This gap falls between the Completeness and Surface Fitness dimensions.

---

## Score Summary

| Dimension | Score | Max | Threshold | Pass |
|-----------|-------|-----|-----------|------|
| 完整性 | 123 | 140 | 85 | ✓ |
| 语义纯度 | 107 | 170 | 100 | ✓ |
| 前置条件互斥性 | 123 | 140 | 85 | ✓ |
| 事实依据 | 96 | 140 | 85 | ✓ |
| Surface 适配 | 67 | 90 | 55 | ✓ |
| 一致性 | 128 | 140 | 85 | ✓ |
| 锚点完整性 | 90 | 90 | 55 | ✓ |
| 前置数据声明 | 58 | 90 | 55 | ✓ |
| **Total** | **792** | **1000** | **850** | **✗** |

**Result**: FAIL — Total 792 < 850 target. All dimensions pass their individual thresholds, but the aggregate score falls short.

---

## Top Issues for Revision

1. **[语义纯度] Remove implementation coupling from dimension values** — Move code references (file paths, function names, API endpoint paths) from Side-effect/Input/State fields into annotations or comments. Dimension values should describe system-level behavior only.

2. **[完整性] Add derived Outcomes to web Steps 1–3** — Each web Contract should consider validation-error and session-expired Outcomes, not just the consolidated Step 4.

3. **[事实依据] Mark UNKNOWN/PARTIAL facts in Output fields** — Step 6's 409 claim and Steps 4/5's policy-def 403 claim must reflect the uncertainty of their backing facts in the Output, not just in Side-effect.

4. **[前置数据声明] Replace abstract entity types with domain model entities** — Steps 5, 6 should use concrete entities (User, Role, Policy, Tenant) instead of Caller/TargetResource/EditableObject.

5. **[前置数据声明] Reconcile dual fixture_spec declarations** — Inline per-Outcome and bottom section fixture_specs conflict; consolidate to a single authoritative declaration per Contract.

6. **[blindspot] Split Step 5 in-tenant-200 into per-CRUD Outcomes** — A single Outcome covering all CRUD operations is not testable by a downstream agent.

7. **[blindspot] Address casbin convergence timing in Step 3** — Specify a wait/retry strategy to handle the 60s policy reload window.
