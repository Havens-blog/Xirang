# Contract Eval — iteration 1

**Journey**: readonly-audit-access
**Rubric**: contract.md (1000 pts, target 850)
**Scorer persona**: Senior QA Engineer (adversarial)
**Date**: 2026-10-08
**ITERATION**: 1

## SCORE: 819/1000

**Pass condition**: FAIL (total 819 < 850; all dimensions pass min threshold but total below target)

## DIMENSIONS

| Dimension | Score | Max | Min | Pass? |
|-----------|-------|-----|-----|-------|
| 完整性 (Completeness) | 136 | 140 | 85 | PASS |
| 语义纯度 (Semantic Purity) | 122 | 170 | 100 | PASS |
| 前置条件互斥性 (Precondition Exclusivity) | 108 | 140 | 85 | PASS |
| 事实依据 (Fact Alignment) | 85 | 140 | 85 | PASS (barely) |
| Surface 适配 (Surface Fitness) | 82 | 90 | 55 | PASS |
| 一致性 (Internal Consistency) | 130 | 140 | 85 | PASS |
| 锚点完整性 (Anchor Integrity) | 90 | 90 | 55 | PASS (no handbook) |
| 前置数据声明 (Fixture Specification) | 66 | 90 | 55 | PASS |
| **Total** | **819** | **1000** | **850** | **FAIL** |

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution alignment**: The 4 contracts cover both halves of PRD Story 5 — positive audit capability (Steps 1-2: read/search/paginate) and write denial (Steps 3-4: API 403, URL 403). Solution addresses the stated problem.

2. **Implementation coupling in Side-effect dimension**: 5 outcomes across Steps 1, 3, 4 use Side-effect values containing file paths (`permission.go:20`, `checker.go:45-46`) and function names (`middleware.CheckPermission`, `errs.ErrForbidden`). These are implementation details, not system-level behavior descriptions. The rubric's Semantic Purity dimension explicitly prohibits "internal function calls, database queries, API endpoint paths, or file system paths" in dimension values.

3. **Precondition overlap in Step 1**: Five outcomes in Step 1 have overlapping Preconditions:
   - `empty-state` ("数据为空") and `zero-side-effect-error` ("网络错误或空状态") both match when data is empty.
   - `success-data-present` ("含可读数据") and `zero-side-effect-normal` ("进行搜索、分页、筛选...会话有效") both match when data is present and user interacts.
   The `(区别于...)` annotations help readability but do not add structural exclusivity — both Preconditions can be true simultaneously.

4. **Unclassified claim in corrected-retry**: Step 2's `corrected-retry` outcome asserts "可重新搜索，返回正确结果" with no fact_id reference and no `source: inferred` annotation. The journey's Step 2b Outcome B cites [F7], but the contract drops the traceability. Per rubric, this is an unclassified claim.

5. **Inferred outcomes cite `required_outcomes` generically**: All 4 inferred outcomes reference "required_outcomes" as the rule source (e.g., "web surface required_outcomes session-expired 派生") instead of the specific rule path (e.g., "surface-web.md–session-expired") as explicitly instructed by the journey's Fact Table note.

6. **Step 3 write-denied-403 Output mixes surfaces**: The outcome is in an @api-functional step but its Output says "前端提示无权限且数据不变" — frontend UI behavior in an API-surface contract.

7. **Bottom-level Fixture Specification inconsistent with inline specs**: Steps 1 and 2 have inline fixture_specs including `EiamService`, but the bottom-level aggregate `## Fixture Specification` section omits it, creating ambiguity about authoritative entity set.

## Phase 2 — Dimension Scoring

### 1. Completeness — 136/140

- **Every Outcome has four mandatory dimensions (0-50)**: All 14 outcomes across 4 contracts have non-empty Preconditions, Input, Output, and State. Side-effect defaults to "none" where applicable. **50/50**

- **Journey Invariants section present (0-50)**: All 4 contracts have `## Journey Invariants` with 5 entries each, all citing facts (F1-F8, IAM_*). **50/50**

- **Outcomes cover happy path + derived scenarios (0-40)**: Web mandatory `validation-error` (Step 2 search-validation-error) and `session-expired` (Step 1) present. API mandatory `unauthorized` (Step 3 unauthorized-401) present. `network-error` also present (Step 2). Happy path outcomes in all steps. Minor: Step 4 has no session-expired outcome, but it is less applicable (menu pruning scenario, no form interaction). **36/40**

### 2. Semantic Purity — 122/170

- **Natural language, not code/regex (0-70)**: Preconditions, Input, Output, State are clean natural language throughout — no regex, CSS selectors, or XPath. However, Side-effect values contain code-level references: `permission.go:24-26`, `permission.go:20`, `checker.go:45-46`, `middleware.CheckPermission`, `errs.ErrForbidden` across 5 outcomes (Step 1 session-expired, Step 3 both outcomes, Step 4 url-direct-403, Step 4 menu-not-rendered). **55/70**

- **Preconditions declarative (0-50)**: Most Preconditions are declarative state descriptions. Borderline cases: `zero-side-effect-normal` ("只读角色在查看过程中进行搜索、分页、筛选等操作") describes actions not state; `corrected-retry` ("只读角色修正非法搜索词为合法值后重新搜索") is a procedural sequence; `url-direct-403` ("只读角色经浏览器地址栏直接粘贴 URL 访问") is procedural. All include `(区别于...)` exclusivity annotations. **42/50**

- **No implementation coupling (0-50)**: Side-effect values in Steps 1, 3, 4 contain implementation coupling — file paths and function calls instead of system-level behavior:
  - Step 1 session-expired: "eiam 鉴权返回 401（permission.go:24-26）；前端拦截器跳登录"
  - Step 3 write-denied-403: "eiam middleware.CheckPermission 拒绝（permission.go:20）；403 经 errs.ErrForbidden（checker.go:45-46）"
  - Step 3 unauthorized-401: "eiam middleware.CheckPermission 返回 401（permission.go:24-26）"
  - Step 4 url-direct-403: same as Step 3 write-denied-403
  - Step 4 menu-not-rendered: "前端菜单裁剪为体验层（IAM_READONLY_MENU_PRUNE 前端）"
  **25/50**

### 3. Precondition Exclusivity — 108/140

- **Preconditions distinct across Outcomes (0-50)**: Step 1 has two overlap pairs:
  - `empty-state` ("对应页数据为空") vs `zero-side-effect-error` ("网络错误或空状态") — the "空状态" condition in the latter subsumes the former.
  - `success-data-present` ("含可读数据") vs `zero-side-effect-normal` ("进行搜索、分页、筛选...会话有效") — both match when data is present and user searches.
  Steps 2, 3, 4 are distinct. **35/50**

- **Sufficient to uniquely select an Outcome (0-50)**: The Step 1 overlaps mean that given the system state (data empty, user searching), two outcomes are simultaneously selectable. The distinguishing factor is in Input/Output focus, not in Preconditions. **35/50**

- **Error/boundary Outcomes state triggering conditions (0-40)**: All error/boundary outcomes explicitly state triggers: session-expired ("会话超时过期"), network-error ("搜索请求网络错误"), search-validation-error ("搜索框存在格式校验规则...非法搜索词"), unauthorized-401 ("请求未携带有效令牌或令牌已过期"), empty-state ("对应页数据为空"). **38/40**

### 4. Fact Alignment — 85/140

- **Factual claims traceable to fact_id or UNKNOWN (0-50)**: Most outcomes cite facts in Invariants fields (F1-F8, IAM_READONLY_ROLE_WRITE_DENY, IAM_READONLY_MENU_PRUNE, IAM_403_NO_EXISTENCE_LEAK). Two gaps:
  - Step 1 `zero-side-effect-normal`: no fact reference (journey Step 1d Outcome A cites [F1]).
  - Step 2 `corrected-retry`: no fact reference (journey Step 2b Outcome B cites [F7]).
  Step 1 session-expired Side-effect cites code path but not fact_id IAM_UNAUTHORIZED_401. **35/50**

- **Inferred claims have rule support + source: inferred (0-50)**: 4 inferred outcomes all carry `source: inferred` and reasoning annotations. However, all cite "required_outcomes" generically (e.g., "web surface required_outcomes session-expired 派生") instead of the specific rule path (e.g., "surface-web.md–session-expired") as explicitly instructed by the journey's Fact Table note: "不引用段落名 `required_outcomes` 本身". **35/50**

- **No hallucinated claims without classification (0-40)**: Step 2 `corrected-retry` asserts "可重新搜索，返回正确结果；不报错、不丢失已加载列表" with no fact_id and no `source: inferred` — unclassified claim per rubric definition. Step 1 `zero-side-effect-normal` also lacks annotation but is indirectly covered by contract-level Journey Invariants citing F1. Per rubric: -30 pts per hallucinated unclassified claim. **15/40** (40 - 30 + 5 partial)

### 5. Surface Fitness — 82/90

- **Mandatory derived Outcomes present (0-35)**: Web `validation-error` (Step 2) ✓, `session-expired` (Step 1) ✓. API `unauthorized` (Step 3) ✓. **35/35**

- **Surface-appropriate language (0-30)**: Web steps use user-interaction language (navigation, search, URL access) ✓. API step uses HTTP status codes (403, 401) ✓ but:
  - Side-effect values use implementation details (function names, file paths) instead of API-appropriate language.
  - Step 3 write-denied-403 Output mixes surfaces: "服务端权限校验拒绝（403），前端提示无权限且数据不变" — API surface should describe HTTP response, not frontend UI behavior.
  **22/30**

- **TUI timeout Outcomes (0-25)**: Not TUI surface. Full marks. **25/25**

### 6. Internal Consistency — 130/140

- **Invariants hold in every Step Contract (0-50)**: All 5 invariants verified across all 4 contracts — no violations. Session-expired outcome redirects to login, but this is session expiry (not 403), so invariant 4 ("403 原地收敛不跳转登录页") is not violated. **50/50**

- **Cross-Contract state references consistent (0-50)**: Step 2 "页面已加载" → Step 1 "只读视图就绪" ✓. Step 3 "持有有效令牌" → Step 1 valid session ✓. Step 4 "已登录" consistent ✓. No dangling references. **45/50**

- **Outcome Preconditions consistent with preceding State changes (0-40)**: Step 2 Preconditions achievable from Step 1 ✓. Step 3 achievable from Step 1 ✓. Step 4 consistent ✓. Minor: Step 3 unauthorized-401 "令牌已过期" semantically overlaps with Step 1 session-expired "会话超时过期" — both describe expired sessions, distinguished only by surface (API vs Web). The overlap could cause confusion in cross-step state flow. **35/40**

### 7. Anchor Integrity — 90/90

No handbook exists for any surface type. Per rubric: "When no handbook exists for the surface type: score full marks (90/90)." All contracts note "Handbook missing. Anchor fields empty per graceful-degradation rule." **90/90**

### 8. Fixture Specification — 66/90

- **Entity completeness (veto item, 0-35)**: All outcomes' inline fixture_specs include entities they reference. Veto not triggered. However:
  - Bottom-level aggregate `## Fixture Specification` omits `EiamService` (present in inline specs for Steps 1, 2).
  - Entity type `ReadOnlyRole` is conceptual — fact-table.json notes no seeded readonly role exists (only super_admin + admin). Semantic verification against Design document cannot confirm match.
  **30/35**

- **Relationship and constraint coverage (0-30)**: Multi-entity contracts (Steps 1, 2) involve PlatformUser, Role, Authorization but declare no relationships (relationship_type, parent_entity). These entities are related in the domain (users have roles, roles have authorizations) but the fixture_specs list them as independent. **18/30**

- **Minimum data quantity declarations (0-25)**: Step 2 audit-integrity declares min_count: 2 for PlatformUser, Role, Authorization while Preconditions state "含多页以支撑分页审计" (multi-page). With only 2 items, pagination may not be meaningfully testable. Step 1 success-data-present min_count: 1 is reasonable. **18/25**

## Phase 3 — Blindspot Hunt

- **[blindspot] Step 3 write-denied-403 Output mixes API and Web surface language** — "服务端权限校验拒绝（403），前端提示无权限且数据不变" — the outcome is in an @api-functional step but describes frontend UI behavior ("前端提示无权限"). API surface contracts should describe HTTP response body and status code, not frontend rendering. Reasoning audit flagged this independently of dimension scoring.

- **[blindspot] Invalid YAML in frontmatter** — `api: { endpoint: "", method: "", content_type: "", auth_required: }` and `web: { page: "", route: "", requires_auth: , layout: "" }` — `auth_required:` and `requires_auth:` have no values (bare keys), which is invalid YAML. Downstream parsers may fail. Must use `null` or a boolean.

- **[blindspot] Bottom-level Fixture Specification inconsistent with inline fixture_specs** — Steps 1 and 2 inline fixture_specs include `EiamService` (for network-error and zero-side-effect-error outcomes) but the bottom-level `## Fixture Specification` section omits it. This creates ambiguity about which entity set is authoritative for test data seeding.

- **[blindspot] Reasoning audit: Step 1 has 5 outcomes but 2 pairs are not truly exclusive** — the contract merges journey Steps 1, 1c, 1d into a single Contract with 5 outcomes, but the merging introduced Preconditions that overlap (empty-state vs zero-side-effect-error, success-data-present vs zero-side-effect-normal). The journey had these as separate edge cases with distinct Precondition fields; merging them into one Contract without refining Preconditions to be structurally exclusive is a design flaw not captured by any single rubric dimension. Reasoning audit flagged this independently of dimension scoring.

## ATTACKS (summary)

1. [语义纯度]: Side-effect dimension values contain implementation coupling (file paths, function names) — "eiam middleware.CheckPermission 拒绝（permission.go:20）；403 经 errs.ErrForbidden（checker.go:45-46）" — must replace with system-level behavior descriptions (e.g., "写操作被服务端权限校验拒绝，返回 403").

2. [前置条件互斥性]: Step 1 empty-state and zero-side-effect-error Preconditions overlap on "空状态" — "只读角色查看时发生网络错误或空状态" vs "对应页数据为空" — must add distinguishing condition (e.g., restrict zero-side-effect-error to "网络错误" only, removing "或空状态") to make outcomes mutually exclusive.

3. [事实依据]: corrected-retry outcome is unclassified (no fact reference, no source: inferred) — "可重新搜索，返回正确结果；不报错、不丢失已加载列表" — must add fact reference (F7 per journey Step 2b Outcome B) or source: inferred annotation with rule citation.

4. [事实依据]: Inferred outcomes cite "required_outcomes" generically instead of specific rule paths — "web surface required_outcomes session-expired 派生" — must cite specific rule path like "surface-web.md–session-expired" per journey's Fact Table explicit instruction.

5. [前置数据声明]: Multi-entity contracts lack relationship declarations — entities listed as independent without relationship_type/parent_entity — must declare relationships (e.g., Role assigned to PlatformUser, Authorization bound to Role).

6. [前置数据声明]: Step 2 min_count: 2 insufficient for pagination testing — Preconditions state "含多页以支撑分页审计" but min_count: 2 may not fill multiple pages — must increase min_count to exceed typical page size.

7. [Surface 适配]: Step 3 write-denied-403 Output mixes API and Web surface language — "服务端权限校验拒绝（403），前端提示无权限且数据不变" — API surface outcome must describe HTTP response (status code, response body), not frontend UI behavior.

8. [blindspot]: Invalid YAML in frontmatter — `auth_required:` and `requires_auth:` have no values — must provide valid YAML values (null or boolean) to prevent downstream parser failures.

9. [blindspot]: Bottom-level Fixture Specification inconsistent with inline fixture_specs — EiamService present in inline specs but missing from bottom-level aggregate in Steps 1 and 2 — must reconcile to a single authoritative entity set.

10. [blindspot]: Step 1 merging of journey Steps 1/1c/1d introduced non-exclusive Preconditions — success-data-present vs zero-side-effect-normal and empty-state vs zero-side-effect-error are not structurally distinguishable by Preconditions alone — must refine Preconditions so each outcome is uniquely selectable.
