# Eval Report — Journey: tenant-boundary-enforcement (Iteration 1)

**Scorer persona**: Senior QA Engineer
**Rubric**: journey.md (7 dimensions, 1000 pts, target 850, pass = total ≥ 850 AND every dimension ≥ threshold)
**Surface types declared**: `["web", "api"]`
**Iteration**: 1

---

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution**: The Overview promises to verify that any out-of-tenant-boundary request is rejected (403) by eiam. The Happy Path covers within-boundary governance (Steps 1–2) + boundary-violation rejection (Step 3). Edge cases cover multiple violation vectors (direct API, URL, policy-definition CRUD, concurrent editing). The solution addresses the stated problem, but only tests the *happy case of boundary enforcement* — there is no outcome where the boundary fails (e.g., eiam misconfiguration allows cross-tenant access). An adversarial test plan for a security boundary should include a negative-of-negative case.

2. **Solution → Evidence**: Claims ("eiam 返回 403", "eiam 对所有 CRUD 端点做服务端权限校验") are grounded in PRD Story 3 AC2 and PRD Spec Security Requirements, but there is **no Fact Table**, no `fact_id` traceability, and no `UNKNOWN` markings. Inferred boundary outcomes (403 non-leakage of existence) carry no `source: inferred` annotation and no `required_outcomes` rule citation.

3. **Evidence → Success Criteria**: The success criteria (Expected Results) assert 403 behavior but do not test any of the surface-mandated derived outcomes (web: validation-error, session-expired; api: unauthorized/401). The 403 (forbidden) outcomes present are authorization-rejection scenarios, not the authentication-failure (401) scenarios the API surface requires.

4. **Self-contradiction check**:
   - **CONTRADICTION**: frontmatter `golden_path: false` ↔ Happy Path section contains 3 contiguous steps mapping to PRD Story 3. Type: direction-clash. The author declares no golden path while the content provides a 3-step sequence semantically tied to a user story.
   - **CONTRADICTION**: Invariant "403 不泄露存在性" is stated as a journey-wide invariant, but only Step 1b's Expected Result explicitly reflects it ("不泄露其他租户用户存在性"). Steps 3, 2b, and 3b produce 403 outcomes that do not mention existence non-leakage. If the invariant is journey-wide, all 403 outcomes must reflect it; otherwise the invariant scope is over-claimed.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 115/175 (threshold 105) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Journey metadata complete | 40/45 | Name `tenant-boundary-enforcement` is kebab-case ✓; risk_level `Medium` is a valid value and borderline-justified (multi-step interaction; role creation is reversible state mutation; 403 keeps data unchanged). Could argue High for security-boundary enforcement. |
| Steps complete with required fields | 55/70 | Every Step has a step name, User Action, and Expected Result. Steps form a coherent ordered sequence. Deductions: happy-path Steps 1–3 lack explicit Precondition (rely on Setup); each step has exactly ONE outcome — edge cases are split into separate steps (1b/2b/3b/1c) rather than alternative outcomes within the same step, yielding a shallow single-outcome structure. |
| Outcomes cover happy path + required derived scenarios | 20/60 | Boundary/edge outcomes present (403 from multiple vectors, concurrent editing). But **mandatory derived outcomes from both declared surfaces are absent**: Web requires `validation-error` + `session-expired` — neither present; API requires `unauthorized` (401) — absent (only 403/forbidden outcomes exist). The invariant even distinguishes 403 from session-expiry but never tests session-expiry. |

### 2. Semantic Purity (语义纯度) — 115/175 (threshold 105) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Outcome descriptions use natural language | 55/70 | Outcomes are mostly natural-language behavioral descriptions. "eiam 返回 403" is an HTTP status-code reference — acceptable for the api surface but slightly implementation-coupled for web outcomes. No regex, CSS selectors, or framework assertion calls. |
| Preconditions are declarative, not procedural | 35/55 | Some preconditions describe actions rather than states. Step 1b Precondition: "租户管理员绕过前端裁剪，直接以其他租户 ID 调用户列表接口" describes an action (bypassing, calling), not a state. Step 3b Precondition: "租户管理员经浏览器地址栏直接访问其他租户管理页 URL" is procedural. A declarative form would state the condition that holds (e.g., "租户管理员持有其他租户的管理页 URL 且前端菜单已裁剪"). |
| No implementation coupling in Step descriptions | 25/50 | Multiple steps couple to implementation: Step 3 "经直接调接口或经 URL 访问其他租户资源"; Step 1b "直接调用其他租户的列表接口"; Step 2b "经 URL 或直接调接口"; Step 3b "粘贴其他租户管理页 URL 并访问". These reference API calls and URL-access methods rather than domain-level user operations. |

### 3. Precondition Exclusivity (前置条件互斥性) — 90/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes within each Step | 35/50 | Each step has exactly one outcome — there is no overlap (trivially satisfied), but also no demonstration of exclusivity through multiple precondition-gated outcomes. The step structure never leverages precondition-based outcome selection. |
| Preconditions sufficient to uniquely select an Outcome | 30/45 | With single outcomes per step there is no ambiguity, but the structure provides no selection mechanism to verify. Edge-case preconditions (1b/2b/3b/1c) are distinct from each other. |
| No missing Preconditions for error/boundary Outcomes | 25/35 | Edge cases 1b/2b/3b/1c all have explicit Preconditions ✓. But happy-path Steps 1–3 have no explicit Precondition (rely on Setup). Step 3 ("越权请求被服务端拒绝") is a boundary scenario in the Happy Path section with no precondition. |

### 4. Fact Alignment (事实依据) — 65/130 (threshold 80) ✗ BELOW THRESHOLD

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 25/50 | No Fact Table exists in the document. Claims like "eiam 返回 403", "eiam 对所有 CRUD 端点做服务端权限校验", "分配后该用户下一次请求即具备对应权限" are factual assertions about eiam behavior. They are grounded in PRD sources but lack `fact_id` traceability and are not marked `UNKNOWN`. The rubric requires formal traceability, not just PRD alignment. |
| Inferred claims have required_outcomes rule support and source: inferred | 15/45 | The 403 boundary outcomes are derived from the journey's purpose but carry no `source: inferred` annotation and cite no `required_outcomes` rule from the surface configuration. No inferred claim in the document follows the required annotation format. |
| No hallucinated claims without classification | 25/35 | Most claims trace to PRD Story 3 / Spec. But "不泄露其他租户用户存在性" (Step 1b and Invariant 4) is a behavioral assertion **not present in the PRD** — the PRD never states that 403 responses must not leak resource existence. This is an unclassified inferred security requirement (hallucination without classification). |

### 5. Surface Fitness (Surface 适配) — 55/130 (threshold 80) ✗ BELOW THRESHOLD

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes from surface required_outcomes present | 10/50 | Web mandatory: `validation-error` — ABSENT; `session-expired` — ABSENT (invariant mentions it but no outcome tests it). API mandatory: `unauthorized` (401) — ABSENT. Only 403 (forbidden) outcomes exist, which are authorization-rejection, not the mandated derived outcomes. Per rubric: "Score 0 if mandatory Outcomes are completely absent" — partial credit only because 403 outcomes are thematically related to api authorization testing. |
| Test strategy proportions match surface guidance | 25/45 | Web/API guidance: balanced 50/50 (Contract 50% / Journey 50%). The journey has 7 steps mixing web interactions (page state, URL access) and api interactions (direct calls, 403). Rough balance exists but the document does not distinguish Contract-level vs Journey-level test coverage, and does not declare which surface each step targets. |
| Surface-specific environment and execution assumptions realistic | 20/35 | Web assumptions (page state, no redirect, URL access) are realistic. API assumptions (403 status codes) are realistic. But "直接调接口" (directly call API) is presented as a user action in multiple steps — this is an API testing technique, not a web user interaction. The journey conflates web user actions with API testing techniques without separating surfaces. |

### 6. Internal Consistency (一致性) — 90/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step | 35/50 | Invariant 1 (eiam server-side enforcement) holds in all steps ✓. Invariant 2 (403 原地收敛) holds in Steps 3/1b/2b/3b ✓. Invariant 3 (策略定义口径) holds in Steps 2/2b ✓. Invariant 4 (403 不泄露存在性) is **only explicitly reflected in Step 1b** — Steps 3, 2b, 3b produce 403 outcomes that do not mention existence non-leakage. A journey-wide invariant must hold in all applicable 403 outcomes. |
| Cross-Step references consistent | 30/45 | No dangling data references. But Step 1c introduces "两个不同租户的管理员" — a second tenant admin not mentioned in Setup (which says "租户管理员账号已存在", singular). Step 2 creates a role and assigns it to a user, but no later step references the created role/user — no lifecycle verification. |
| Risk level consistent with content | 25/35 | Medium = "multi-step interaction without irreversible side effects". The content involves role creation (state mutation) and tenant-boundary security enforcement. Role creation is reversible; 403 outcomes keep data unchanged. Medium is borderline-defensible, but tenant-boundary is a security concern and role creation is state mutation — could warrant High. |

### 7. Workflow Coverage (工作流覆盖度) — 83/130 (threshold 80) ✓

| Criterion | Score | Notes |
|-----------|-------|-------|
| Golden Path existence (veto item) | 30/50 | 3 contiguous happy-path steps exist and map to PRD Story 3 (view users → create role/bind policy → boundary rejection). Not a veto (steps are semantically related to a user story). Deductions: frontmatter declares `golden_path: false` contradicting the 3-step sequence; Step 3 ("越权请求被服务端拒绝") is a negative/boundary scenario, not a domain-level happy-path success operation; Step 3 uses API-level description ("直接调接口"). |
| Multi-step coverage depth | 28/45 | Covers boundary enforcement from multiple vectors (direct API, URL, policy-definition CRUD, concurrent editing). But no entity lifecycle (create → update → delete) — Step 2 creates a role but never updates or deletes it. No state-machine coverage. Moderate depth. |
| Workflow completeness against PRD/Design scope | 25/35 | Story 3 (tenant-admin boundary governance) well covered. Story 4 tenant-admin AC covered (Step 2). E6 (concurrent editing) covered (Step 1c). Gaps: E7 (pagination boundary) relevant to Step 1's user-list viewing but uncovered; the 403-vs-session-expiry distinction (PRD Flow #8) is mentioned in invariants but never tested as an outcome. |

---

## Phase 3 — Blindspot Hunt

### [blindspot] 1: Aspirational security invariant not grounded in PRD

> "403 不泄露存在性——越权提示不暴露其他租户资源是否存在。"

The invariant "403 不泄露存在性" is presented as a journey-wide invariant, but the PRD (prd-spec.md, prd-user-stories.md) never states that 403 responses must not leak resource existence. The PRD's Edge & Negative Scenarios (E1–E12) and Story 3 ACs mention 403 behavior and "不跳转登录页" but say nothing about existence non-leakage. This is an inferred security requirement presented as fact without classification. It should either cite a security-requirement source or be marked `source: inferred` with the reasoning basis.

### [blindspot] 2: No separation of web-surface vs api-surface testing despite declaring both

> `surface_types: ["web", "api"]`

The journey declares both web and api surfaces but never clarifies which surface each step targets. Steps mix browser interactions (URL access, page state, "保持当前页") with API testing techniques ("直接调接口", "调用户列表接口") within the same step. For a dual-surface journey, each step should indicate whether it tests the web surface (browser interaction) or the api surface (HTTP client), and the outcomes should match the surface's verification model. This falls between Surface Fitness and Completeness — no single dimension fully captures the surface-conflation problem.

### [blindspot] 3: Boundary enforcement tested only on the happy side — no negative-of-negative case

> "eiam 返回 403，页面内提示无权限并保持当前页（不跳转登录页），数据不变。"

Every boundary-violation outcome assumes eiam correctly returns 403. A Senior QA Engineer would include at least one outcome where the boundary enforcement itself fails (e.g., eiam misconfiguration, a newly added endpoint not yet covered by the authorization matrix, or a caching lag allowing stale permissions through). The journey tests "does 403 happen when it should" but never tests "does 403 fail to happen when eiam is misconfigured." The PRD's Phase 0 emphasizes "服务端授权核查（CRUD 端点逐项确认）" — the journey should reflect the risk that a missed endpoint bypasses the boundary.

### [blindspot] 4: Step 1 only checks UI-level data scoping, not API-response-level scoping

> "仅显示本租户用户，无法看到其他租户数据；分页/筛选在本租户范围内生效。"

Setup states "其他租户数据存在且租户管理员不应可见." Step 1's Expected Result says "仅显示本租户用户" — this only verifies the *displayed* (UI) data, not that other-tenant data is absent from the API *response*. A tenant admin could still receive other-tenant data in the API response if the frontend fails to filter. The boundary should be enforced server-side (per the invariant), but Step 1's outcome only checks the web-surface display, not the api-surface response payload. This is a reasoning-audit flag: the invariant says "前端菜单裁剪仅为体验层" but Step 1's verification is purely UI-level.

---

## Dimension Summary

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| Completeness (完整性) | 115/175 | 105 | ✓ |
| Semantic Purity (语义纯度) | 115/175 | 105 | ✓ |
| Precondition Exclusivity (前置条件互斥性) | 90/130 | 80 | ✓ |
| Fact Alignment (事实依据) | 65/130 | 80 | ✗ BELOW |
| Surface Fitness (Surface 适配) | 55/130 | 80 | ✗ BELOW |
| Internal Consistency (一致性) | 90/130 | 80 | ✓ |
| Workflow Coverage (工作流覆盖度) | 83/130 | 80 | ✓ |
| **Total** | **613/1000** | 850 | ✗ FAIL |

**Pass condition**: FAILED. Total 613 < 850, and two dimensions (Fact Alignment, Surface Fitness) are below their minimum thresholds.

---

## Priority Fixes for Iteration 2

1. **[CRITICAL] Add Fact Table with fact_ids** — every factual claim ("eiam 返回 403", "eiam 对所有 CRUD 端点做服务端权限校验", "逐请求判定") must trace to a `fact_id` or be marked `UNKNOWN`. Add `source: inferred` annotations to derived boundary outcomes with the `required_outcomes` rule that triggered them.
2. **[CRITICAL] Add mandatory surface derived outcomes** — Web: `validation-error` (form submission with invalid data) + `session-expired` (session timeout during workflow); API: `unauthorized` (401, missing/expired credentials, distinct from 403 forbidden).
3. **[HIGH] Resolve `golden_path: false` contradiction** — either declare `golden_path: true` (the 3 happy-path steps map to Story 3) or restructure so the golden path is the primary success workflow (within-boundary governance) and move boundary-violation to edge cases.
4. **[HIGH] Restructure steps to have multiple outcomes with precondition-based selection** — merge happy-path and edge-case steps so each step has alternative outcomes distinguished by Preconditions.
5. **[HIGH] Remove implementation coupling from step descriptions** — replace "直接调接口"/"经 URL 访问" with domain-level user operations; move access-method details to test-execution notes.
6. **[MEDIUM] Apply "不泄露存在性" invariant to ALL 403 outcomes** or narrow the invariant scope.
7. **[MEDIUM] Clarify surface per step** — label each step as web-surface or api-surface and align outcomes to the surface's verification model.
8. **[MEDIUM] Add declarative preconditions to happy-path steps** and convert procedural preconditions to state-describing form.
