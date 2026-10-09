# Eval Report — Journey: tenant-boundary-enforcement (Iteration 2)

**Scorer persona**: Senior QA Engineer
**Rubric**: journey.md (7 dimensions, 1000 pts, target 850, pass = total ≥ 850 AND every dimension ≥ threshold)
**Surface types declared**: `["web", "api"]`
**Iteration**: 2

---

## Iteration-1 Attack Resolution Verification

| Iteration-1 Attack | Severity | Status | Verification |
|--------------------|----------|--------|--------------|
| Add Fact Table with fact_ids | CRITICAL | RESOLVED | Fact Table F1–F10 added with source column; every outcome carries `[fact_id: ...]` or `[inferred: ...]`. |
| Add mandatory surface derived outcomes | CRITICAL | RESOLVED | Web `validation-error` (Step 4 C), `session-expired` (Step 4 B); API `unauthorized`/401 (Step 5 A). All present. |
| Resolve `golden_path: false` contradiction | HIGH | RESOLVED | frontmatter now `golden_path: true`; explicit Golden Path block in Happy Path maps to Story 3/4. |
| Restructure steps to multi-outcome precondition selection | HIGH | RESOLVED | Steps 4 & 5 now have 4 outcomes each with distinct Preconditions. |
| Remove implementation coupling from step descriptions | HIGH | PARTIAL | Happy-path steps are domain-level; Step 4 Outcome D still embeds "绕过前端直接以有效凭据请求" (API bypass technique in a @web-e2e step). |
| Apply "不泄露存在性" to ALL 403 outcomes | MEDIUM | RESOLVED | Step 4 A/D, Step 5 B/C all state non-leakage; Invariant 5 marked `inferred: I1`. |
| Clarify surface per step | MEDIUM | RESOLVED | Steps tagged `@web-e2e` (1–4) / `@api-functional` (5–6). |
| Add declarative preconditions to happy-path steps | MEDIUM | RESOLVED | Steps 1–3 now have declarative state Preconditions. |

No new contradictions were introduced by the revision. The `golden_path: true` ↔ 3-step sequence contradiction from iteration-1 is eliminated. Invariant 4 (existence non-leakage) is now consistently reflected in all 403 outcomes and correctly classified as inferred.

---

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution**: The Overview promises within-boundary governance + cross-boundary rejection enforced by eiam. Happy Path (Steps 1–3) covers within-boundary governance; Edge Cases (Steps 4–6) cover boundary violations across web + api surfaces + concurrency. The solution addresses the stated problem. The negative-of-negative case (eiam misconfiguration allowing cross-tenant access) remains absent — noted as a blindspot, though arguably out of scope for a user-workflow journey.

2. **Solution → Evidence**: Fact Table F1–F10 provides traceability to PRD sources. Inferred Claims I1–I4 are annotated `source: inferred` with reasoning bases. I2/I3/I4 cite `required_outcomes` rules; I1 cites a security principle (Story 1 AC) rather than a required_outcomes rule — a format gap (see Fact Alignment).

3. **Evidence → Success Criteria**: Expected Results now assert 403, 401, validation-error, session-expired, and concurrent-edit behavior. The mandated surface derived outcomes are testable. However, Step 4 Outcomes B/C (session-expired, validation-error) are mandatory outcomes that are semantically detached from the step's action ("访问其他租户治理资源") — they satisfy the checklist but don't organically arise from tenant-boundary enforcement.

4. **Self-contradiction check**: No confirmed SC↔SC or SC↔InScope contradictions. Invariants 1–5 hold across all applicable steps. Cross-step reference (Step 3 → Step 2 role) is consistent. Setup singular "租户管理员账号已存在" vs Step 6 "两个不同租户的管理员" is a minor unreconciled reference.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 158/175 (threshold 105) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Journey metadata complete | 43/45 | Name kebab-case ✓; risk_level Medium valid ✓; golden_path true ✓; surface_types declared ✓; sources + surface_keys present ✓. Small deduction: Medium is borderline for a security-boundary feature (could warrant High). |
| Steps complete with required fields | 63/70 | Every Step has name, Precondition, User Action, Expected Result. Steps 4–5 have multi-outcome structure with per-outcome Preconditions. Coherent ordered sequence. Deductions: Happy-path Steps 1–3 retain single-outcome structure (acceptable for golden path but shallow). |
| Outcomes cover happy path + required derived scenarios | 52/60 | Required derived outcomes present: web validation-error + session-expired (Step 4 B/C); api unauthorized/401 (Step 5 A). Boundary outcomes across multiple vectors ✓. Concurrent edit ✓. Deductions: Step 4 B/C are mandatory outcomes artificially attached to a cross-tenant-access step — their Preconditions don't relate to the step's boundary action. Common web boundary outcomes (network-error, navigation-guard) absent but not mandatory. |

### 2. Semantic Purity (语义纯度) — 152/175 (threshold 105) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Outcome descriptions use natural language | 60/70 | Outcomes are natural-language behavioral descriptions. API status codes (401/403/200/201/204) are appropriate for the api surface per surface-api.md. Web outcomes reference "eiam 返回 403" — the backend service name is mildly implementation-coupled for @web-e2e (the user observes "页面提示无权限", not "eiam 返回 403"). No regex/CSS/XPath/assertion calls. |
| Preconditions are declarative, not procedural | 50/55 | Steps 1–3 Preconditions are declarative states ✓. Step 4 outcome Preconditions ("目标资源归属其他租户，当前会话有效" / "操作过程中会话已过期" / "提交的表单含非法字段") are declarative ✓. Step 5 outcome Preconditions declarative ✓. Minor: Step 4 Outcome D "操作者为租户管理员，目标为策略定义管理" is declarative but could be more precise about the write-operation target. |
| No implementation coupling in Step descriptions | 42/50 | Happy-path User Actions are domain-level ✓. Step 4 User Action "尝试访问不属于本租户的治理资源" is domain-level ✓. Step 5 User Action "调用方以 HTTP 请求访问 eiam 治理接口（用户/角色/授权/策略 CRUD）" is appropriate for @api-functional. Deduction: Step 4 Outcome D "即使绕过前端直接以有效凭据请求" embeds an API bypass testing technique within a @web-e2e outcome — implementation coupling + surface conflation. |

### 3. Precondition Exclusivity (前置条件互斥性) — 118/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes within each Step | 45/50 | Step 4: A (cross-tenant, session valid) / B (session expired) / C (invalid form fields) / D (policy-definition target) — all distinct. Step 5: A (no creds) / B (valid creds, cross-tenant) / C (valid creds, policy-def write) / D (valid creds, own tenant) — all distinct. No overlap. |
| Preconditions sufficient to uniquely select an Outcome | 40/45 | Step 5 outcomes are mutually exclusive and exhaustive (no-creds vs valid-cross vs valid-policy-def vs valid-own). Step 4 Outcome A explicitly states "当前会话有效" which excludes B (session expired) — good exclusivity engineering. Deduction: Step 4 Outcome D's Precondition ("操作者为租户管理员，目标为策略定义管理") could overlap with A if the policy-definition target also belongs to another tenant — ambiguous edge. |
| No missing Preconditions for error/boundary Outcomes | 33/35 | All edge/boundary outcomes have explicit Preconditions ✓. Happy-path Steps 1–3 have Preconditions ✓. Minor: Step 6 Precondition "两个不同租户的管理员同时编辑" doesn't specify session validity or credential state. |

### 4. Fact Alignment (事实依据) — 113/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 45/50 | F1–F10 traceable to specific PRD sources with citations ✓. Every outcome carries `[fact_id: ...]` or `[inferred: ...]`. Strong traceability. Minor: F9 "若 eiam 返回 409 版本冲突" assumes a specific HTTP status for version conflict — the PRD E6 citation should be verified to confirm 409 is specified vs inferred. |
| Inferred claims have required_outcomes rule support and source: inferred | 35/45 | I2 (`web/validation-error`), I3 (`web/session-expired`), I4 (`api/unauthorized`) all cite required_outcomes rules ✓ and `source: inferred` ✓. **I1 lacks required_outcomes rule support** — its reasoning basis is "Story 1 AC 存在性不泄露原则推广" (a security principle from another story), not a surface required_outcomes rule. I1 is the most heavily-referenced inferred claim (4 outcomes + Invariant 5) yet doesn't meet the inferred-claim format requirement. |
| No hallucinated claims without classification | 33/35 | All claims are classified as either factual (F1–F10) or inferred (I1–I4). No unclassified hallucinations. I1 is borderline (inferred without rule support) but is honestly marked `source: inferred` with reasoning — not a hallucination. Minor deduction for I1's classification ambiguity. |

### 5. Surface Fitness (Surface 适配) — 111/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes from surface required_outcomes present | 45/50 | Web: validation-error (Step 4 C) ✓, session-expired (Step 4 B) ✓. API: unauthorized/401 (Step 5 A) ✓. All mandatory derived outcomes present. Near-full. |
| Test strategy proportions match surface guidance | 38/45 | Web/API balanced 50/50 guidance. Steps 1–4 @web-e2e, Steps 5–6 @api-functional — roughly balanced. Tags applied correctly. Deduction: no explicit Contract-level vs Journey-level split declared; the document doesn't indicate which outcomes are Contract-testable vs Journey-smoke-testable. |
| Surface-specific environment and execution assumptions realistic | 28/35 | Web steps use browser interaction language (page state, URL, form) ✓. API steps use HTTP status codes, credential/auth semantics ✓. Step 5 properly separates 401/403/200 by scenario. Deduction: Step 4 Outcome D "即使绕过前端直接以有效凭据请求，eiam 返回 403" embeds an API-client bypass technique inside a @web-e2e tagged step — a web-e2e test cannot "绕过前端直接以有效凭据请求"; this is an API testing technique conflated into the web surface. |

### 6. Internal Consistency (一致性) — 114/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step | 46/50 | Inv1 (eiam server-side enforcement): all steps ✓. Inv2 (403 原地收敛): Step 4 A/D (web, "保持当前页") ✓; Step 5 B/C (api, "数据不变" without "保持当前页") ✓ — correctly surface-adapted. Inv3 (策略定义口径): Steps 2, 4 D, 5 C ✓. Inv4 (403 不泄露存在性): Step 4 A/D, Step 5 B/C all reflect it ✓. Inv5 (401 vs 403 separation): Step 5 A vs B/C ✓. Minor: Inv2's "不跳转登录页" is only explicitly stated in Step 4 A, not Step 4 D or Step 5 B/C. |
| Cross-Step references consistent | 38/45 | Step 3 Precondition "Step 2 创建的角色已存在于本租户" — Step 2 creates a role ✓. No dangling references. Deduction: Step 6 introduces "两个不同租户的管理员" but Setup states "租户管理员账号已存在" (singular) — a second tenant admin is not established in Setup. |
| Risk level consistent with content | 30/35 | Medium = "multi-step interaction without irreversible side effects." Role creation is reversible state mutation; 403 outcomes keep data unchanged. Defensible. But tenant-boundary is a security concern — a Senior QA Engineer would argue High is more appropriate for a security-boundary enforcement journey. |

### 7. Workflow Coverage (工作流覆盖度) — 106/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Golden Path existence (veto item) | 45/50 | 3 contiguous happy-path steps (查看用户 → 创建角色绑定策略 → 分配角色) semantically map to PRD Story 3/4 tenant-admin AC. Domain-level operations. `golden_path: true`. Not vetoed. Deduction: the Golden Path covers only within-boundary governance; the feature's namesake purpose (boundary enforcement / 403 rejection) is relegated to Edge Cases. The Overview promises "验证任何越出租户边界的访问均被 eiam 服务端拒绝" but the golden path never asserts a boundary rejection. |
| Multi-step coverage depth | 33/45 | Covers within-boundary governance (create role → bind policy → assign), boundary enforcement across web + api + policy-def + concurrent. Good breadth. Deductions: no entity lifecycle (create → update → delete) — Step 2 creates a role but never updates/deletes it. No state-machine coverage. Concurrent edit (Step 6) is api-only though surface-web.md lists concurrent-edit as a web boundary outcome. |
| Workflow completeness against PRD/Design scope | 28/35 | Story 3 ✓, Story 4 ✓, Story 6 (logout/session) partially via F7/I4 ✓, E6 concurrent ✓, Flow #8 (403 vs session-expiry) ✓. Gaps: E7 pagination boundary — Step 1 asserts "分页与筛选在本租户范围内生效" but never tests pagination-based cross-tenant data leakage (e.g., manipulating page/offset to reach other-tenant records). Update/delete workflows absent. |

---

## Phase 3 — Blindspot Hunt

### [blindspot] 1: No negative-of-negative case — boundary enforcement assumed infallible

> "eiam 对所有 CRUD 端点做服务端权限校验（租户边界为安全边界）"

Every boundary-violation outcome assumes eiam correctly returns 403. A Senior QA Engineer testing a security boundary would include at least one outcome where enforcement fails (e.g., a newly added endpoint not yet covered by the authorization matrix, a caching lag allowing stale permissions, or a misconfigured policy). The journey tests "does 403 happen when it should" but never tests "does 403 fail to happen when eiam has a coverage gap." The PRD Phase 0 emphasizes "服务端授权核查（CRUD 端点逐项确认）" — the journey should reflect the risk that a missed endpoint bypasses the boundary. This reasoning-audit flag is independent of dimension scoring.

### [blindspot] 2: Step 1 verifies only UI-level data scoping, not API-response-level scoping

> "列表仅显示本租户用户，不显示其他租户数据；分页与筛选在本租户范围内生效。"

Step 1 is @web-e2e and its Expected Result verifies the *displayed* (UI) data. But the invariant "前端菜单裁剪仅为体验层、非安全边界" (F4) implies the real boundary is server-side. Step 1's outcome only checks the web-surface display, not that the API response payload excludes other-tenant data. A tenant admin could still receive other-tenant data in the API response if the frontend filters but the backend doesn't. The journey should add an outcome verifying the API response (not just the rendered list) is tenant-scoped. This falls between Surface Fitness and Completeness.

### [blindspot] 3: I1 (existence non-leakage) is a security inference presented without required_outcomes rule grounding

> "I1 | 403 越权响应不泄露其他租户资源是否存在 | source: inferred | Story 1 AC 明确「凭据错误文案不泄露用户名是否存在」的存在性不泄露原则；将该原则推广至 403 越权响应（安全最小信息披露）"

I1 is the most heavily-referenced inferred claim (Step 4 A, Step 4 D, Step 5 B, Step 5 C, Invariant 5) yet its reasoning basis is a cross-story security principle, not a surface `required_outcomes` rule. The rubric's inferred-claim format requires "which required_outcomes rule triggered the derivation." I1 is honestly marked `source: inferred` (so not a hallucination), but it doesn't fit the rubric's inferred category — it's a security-domain inference with no surface-rule anchor. This is a classification gap that no single rubric dimension fully captures (it manifests partially in Fact Alignment).

---

## Dimension Summary

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| Completeness (完整性) | 158/175 | 105 | ✓ |
| Semantic Purity (语义纯度) | 152/175 | 105 | ✓ |
| Precondition Exclusivity (前置条件互斥性) | 118/130 | 80 | ✓ |
| Fact Alignment (事实依据) | 113/130 | 80 | ✓ |
| Surface Fitness (Surface 适配) | 111/130 | 80 | ✓ |
| Internal Consistency (一致性) | 114/130 | 80 | ✓ |
| Workflow Coverage (工作流覆盖度) | 106/130 | 80 | ✓ |
| **Total** | **872/1000** | 850 | ✓ PASS |

**Pass condition**: PASSED. Total 872 ≥ 850, and every dimension ≥ its minimum threshold. Two dimensions (Fact Alignment, Surface Fitness) that were below threshold in iteration-1 now pass (65→113, 55→111). All iteration-1 CRITICAL/HIGH attacks are resolved or partially resolved.

---

## Priority Fixes for Iteration 3 (optional — document already passes)

1. **[MEDIUM] Ground I1 properly** — either cite a `required_outcomes` rule or explicitly reclassify I1 as a security-domain inference with a non-rule reasoning basis acknowledged in the document format.
2. **[MEDIUM] De-conflate Step 4 Outcome D** — move the "绕过前端直接以有效凭据请求" API-bypass assertion into Step 5 (api-surface) or split into a separate api outcome; keep Step 4 D web-surface-only ("入口不渲染").
3. **[MEDIUM] Add pagination boundary test (E7)** — Step 1 or a new edge outcome should test that pagination/page-size manipulation cannot reach other-tenant records, not just assert within-tenant pagination works.
4. **[LOW] Reconcile Setup with Step 6** — add a second tenant admin to Setup or declare Step 6's Precondition as self-establishing.
5. **[LOW] Add API-response-level scoping verification** to Step 1 (verify the response payload, not just the rendered UI, excludes other-tenant data).
