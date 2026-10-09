# Journey Eval — iteration 2

**Journey**: readonly-audit-access
**Rubric**: journey.md (1000 pts, target 850)
**Scorer persona**: Senior QA Engineer (adversarial)
**Date**: 2026-10-08

## SCORE: 839/1000

**Pass condition**: FAIL (total 839 < 850; all dimensions above min threshold)

## DIMENSIONS

| Dimension | Score | Max | Min | Pass? |
|-----------|-------|-----|-----|-------|
| 完整性 (Completeness) | 155 | 175 | 105 | PASS |
| 语义纯度 (Semantic Purity) | 155 | 175 | 105 | PASS |
| 前置条件互斥性 (Precondition Exclusivity) | 100 | 130 | 80 | PASS |
| 事实依据 (Fact Alignment) | 95 | 130 | 80 | PASS |
| Surface 适配 (Surface Fitness) | 105 | 130 | 80 | PASS |
| 一致性 (Internal Consistency) | 108 | 130 | 80 | PASS |
| 工作流覆盖度 (Workflow Coverage) | 121 | 130 | 80 | PASS |
| **Total** | **839** | **1000** | **850** | **FAIL** |

## Iteration-1 Attack Resolution Check

| iter-1 Attack | Status | Evidence |
|---------------|--------|----------|
| #1 Golden Path veto (`golden_path: false`, <3 steps) | **RESOLVED** | `golden_path: true` + 4-step contiguous path anchored to Story 5 |
| #2 No Fact Table, no `fact_id`/`UNKNOWN`/`source: inferred` | **RESOLVED** | Fact Table F1–F4 with PRD AC traceability; inferred outcomes annotated |
| #3 Web mandatory outcomes (`validation-error`, `session-expired`) absent | **RESOLVED** | Step 1b (validation-error), Step 1c (session-expired) added |
| #4 Happy Path steps lack Precondition field | **RESOLVED** | Steps 1–4 each carry explicit Precondition |
| #5 Single-outcome steps, no branching | **PARTIALLY RESOLVED** | Happy Path steps now multi-outcome (A/B/C); Edge Cases remain single-outcome "Expected Result" |
| #6 `golden_path: false` contradicts Overview | **RESOLVED** | `golden_path: true` |
| #7 `eiam` in step action/expected-result | **RESOLVED** | `eiam` moved to Setup/Invariants; steps use "服务端权限校验" |
| #8 Audit-purpose untested (blindspot) | **RESOLVED** | Step 2 Outcome A verifies positive audit capability |
| #9 No Fact Table artifact (blindspot) | **RESOLVED** | Fact Table section present with F1–F4 |
| #10 Over-solves write-denial, under-solves audit viewing (blindspot) | **RESOLVED** | Steps 1–2 cover positive read/search/paginate workflow |

All 10 iteration-1 attacks addressed. Remaining deductions are **new/residual issues** introduced or not fully resolved by the revision.

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Step-level Precondition vs outcome trigger contradiction.** Step 1 Precondition states "用户/角色/授权页面存在可读数据" (data exists) but Outcome B requires "对应页数据为空" (data empty). Step 3 Precondition states "只读角色已登录" (logged in) but Outcome B requires "令牌缺失/过期" (token expired). The step-level Preconditions are too narrow to encompass all declared Outcomes — some Outcomes are unreachable under their step's stated Precondition.

2. **[F1] tag scope mismatch.** Step 2 Outcome A is tagged `[F1]` but contains claims — "搜索、分页、筛选返回正确结果", "数据完整" — that go well beyond F1's scope ("增删改入口不渲染"). These extra claims are neither traceable to F1 nor marked `source: inferred`/`UNKNOWN`. The tag implies factual grounding that does not cover the full claim set.

3. **Edge Case numbering stale after Happy Path expansion.** Iteration 1 had 2 Happy Path steps; iteration 2 added Steps 3–4 but did not renumber Edge Cases. "Step 1b: 搜索输入非法字符" relates to Step 2's search topic; "Step 2b: 经 URL 直接访问策略管理" relates to Step 4's URL-access topic. Both are cross-reference errors.

4. **Dual-surface Golden Path execution gap.** The Golden Path spans Steps 1–4 mixing `@web-e2e` (browser automation) and `@api-functional` (HTTP client) tags. No execution model is described for how a single contiguous Golden Path test switches between browser automation and HTTP client mid-flow.

## Phase 2 — Dimension Scoring

### 1. Completeness — 155/175

- **Metadata (43/45)**: name `readonly-audit-access` kebab-case ✓; `risk_level: Low` justified by read-only content ✓; `golden_path: true` ✓; `surface_types: ["web","api"]` ✓; `sources` lists both PRD files ✓; `surface_keys` present ✓. Minor: `surface_keys` values not explained.
- **Steps complete (65/70)**: 4 Happy Path steps + 5 Edge Cases; each has Precondition + User Action + Outcomes/Expected Result. Steps form coherent ordered sequence. Deduction: Edge Cases use "Expected Result" (single-outcome) format while Happy Path uses "Outcomes" (multi-outcome A/B/C) — structural format inconsistency across steps.
- **Outcomes cover happy + derived (47/60)**: Happy path covered ✓. Mandatory Web derived outcomes present: `validation-error` (Step 1b) ✓, `session-expired` (Step 1c) ✓. Mandatory API derived outcome present: `unauthorized` (Step 3 Outcome B, 401) ✓. Additional `network-error` present (Step 2 Outcome C) ✓. Minor gap: no API-level additional outcomes (e.g., `not-found`, API `validation-error`) for the single API step.

### 2. Semantic Purity — 155/175

- **Natural language (63/70)**: Outcomes in Chinese natural language; no regex/CSS/XPath/framework assertion calls. `403`/`401` are status codes — acceptable as system-level observable behavior. `@web-e2e`/`@api-functional` tags are required by surface rules, not implementation coupling. Minor deduction for tags in step headers.
- **Preconditions declarative (48/55)**: Preconditions are mostly declarative state descriptions ("只读/审计角色已登录；用户/角色/授权页面存在可读数据"). Step 1b Precondition "只读角色在搜索框输入非法字符（如超长串、禁用符号）" is slightly procedural (describes an action in the precondition field).
- **No implementation coupling (44/50)**: `eiam` successfully moved to Setup/Invariants; steps use "服务端权限校验" instead. Residual: Step 3 action "只读角色绕过前端裁剪，直接调用任意写接口" — "绕过前端裁剪" and "写接口" are implementation-level terms; a domain-level description would be "只读角色尝试修改用户/角色/授权数据". Step 3 Precondition "后端对所有写端点" retains API-level vocabulary.

### 3. Precondition Exclusivity — 100/130

- **Distinct across Outcomes (38/50)**: Outcomes within each step have distinguishable trigger conditions (data-present vs data-empty; search-match vs no-match vs network-error; 403 vs 401; menu-hidden vs URL-blocked). However, trigger conditions are embedded in outcome descriptions, not stated as separate per-outcome Preconditions — the step-level Precondition is shared.
- **Sufficient to uniquely select (32/45)**: **Step 1 contradiction**: Precondition states "用户/角色/授权页面存在可读数据" (data exists) but Outcome B requires "对应页数据为空" (data empty) — Outcome B is unreachable under the stated Precondition. **Step 3 contradiction**: Precondition states "只读角色已登录" (logged in) but Outcome B requires "令牌缺失/过期" (token missing/expired) — if logged in, token should be valid, making Outcome B unreachable. Both are precondition-outcome mismatches that prevent unambiguous outcome selection.
- **Missing Preconditions for error/boundary (30/35)**: Error/boundary outcomes state their triggers in descriptions (e.g., "搜索请求网络错误时"). Step 1b/1c Edge Cases carry Precondition fields with `source: inferred` annotations. Minor: triggers are in outcome text, not in dedicated Precondition fields per outcome.

### 4. Fact Alignment — 95/130

- **Factual claims traceable to fact_id or UNKNOWN (40/50)**: F1–F4 traceable to PRD Story 5 AC1/AC2 (verified against `prd-user-stories.md`). Outcomes correctly tagged: Step 1 Outcome A [F1], Step 3 Outcome A [F2], Step 4 Outcome A [F3], Step 4 Outcome B [F4], Invariants [F1][F2], Step 2b [F4]. **Key issue**: Step 2 Outcome A tagged `[F1]` but contains claims beyond F1 — "跨用户、角色、授权三个域页面均可读且数据完整；搜索、分页、筛选返回正确结果" — F1 only covers "增删改入口不渲染". "数据完整" and "搜索、分页、筛选返回正确结果" are not in F1, not in any other fact, and not marked UNKNOWN/inferred. Misleading fact tagging.
- **Inferred claims have rule support + `source: inferred` (35/45)**: Most inferred outcomes properly annotated: Step 1 Outcome B, Step 2 Outcomes B/C, Step 3 Outcome B, Step 1b, Step 1c, Step 2b all carry `source: inferred` with rule citations. **Issue**: Some citations are generic — Step 1 Outcome B and Step 2 Outcome B cite "rule: surface-web.md/required_outcomes 派生" without specifying which required_outcomes rule (validation-error? session-expired? network-error?). `required_outcomes` is a section name, not a specific rule. Step 2 Outcome C correctly cites "surface-web.md–network-error" as a specific rule — this specificity should be applied uniformly.
- **No hallucinated unclassified claims (20/35)**: Claims are PRD-anchored, not pure hallucinations. But Step 2 Outcome A's "搜索、分页、筛选返回正确结果" and "审计角色可据此完成合规审计" are claims tagged `[F1]` that are not actually covered by F1 — misaligned classification. These are neither factual-with-traceability (F1 doesn't cover them) nor inferred-with-annotation (no `source: inferred`). Partial credit only.

### 5. Surface Fitness — 105/130

- **Mandatory derived Outcomes present (45/50)**: Web `validation-error` (Step 1b) ✓, Web `session-expired` (Step 1c) ✓, API `unauthorized` (Step 3 Outcome B) ✓. All mandatory derived outcomes present for both declared surfaces. Near-full marks.
- **Test strategy proportions (27/45)**: surface-web.md and surface-api.md both mandate balanced 50/50 Contract/Journey. The `contracts/` directory is **empty** — no contract specifications exist. The 50/50 split is struct unverifiable from artifacts. Additionally, the journey is web-heavy: 7 web-tagged steps vs 1 API-tagged step, which doesn't reflect a balanced dual-surface distribution. Journey-side depth is good (4 Happy Path + 5 Edge Cases with multi-outcome branching), but the Contract side is entirely absent.
- **Surface-specific environment realistic (33/35)**: Web steps use browser navigation, URL bar access, search/pagination/filter ✓. API step uses direct API calls, 403/401 responses ✓. `@web-e2e`/`@api-functional` tags correctly applied per surface rules ✓. Invariant 4 correctly distinguishes 403-stay from session-expired-redirect, matching PRD spec line 90 ("无权限（403）→ 原地提示、不跳转登录页" vs "会话过期 → 跳转控制台登录页"). Minor: no explicit browser-state cleanup or async-wait assumptions mentioned.

### 6. Internal Consistency — 108/130

- **Invariants hold in every Step (47/50)**: Five invariants verified across all steps:
  - Inv 1 (no write entries rendered): Steps 1, 2, 4 ✓ (Step 3 is API, no rendering) ✓
  - Inv 2 (403 is final boundary): Steps 3, 4 ✓
  - Inv 3 (zero side-effects): Steps 1, 2, 1d ✓
  - Inv 4 (403 stays, not login redirect): Steps 3, 4 ✓; Step 1c contrasts session-expired ✓
  - Inv 5 (audit goal achievable): Steps 1, 2 ✓
  - Minor: Invariant 5 claims "数据完整" which is not a verified fact (see Fact Alignment issue).
- **Cross-Step references consistent (26/45)**: **Two numbering errors** introduced by adding Steps 3–4 without updating Edge Case numbering:
  - "Step 1b: 搜索输入非法字符" — search validation is Step 2's topic (Step 2 action: "执行搜索、分页、筛选"), not Step 1's (navigate to pages). Should be "Step 2b" or "Step 2-edge-1".
  - "Step 2b: 经 URL 直接访问策略管理" — URL access to restricted management is Step 4's topic (Step 4: "访问受限管理页被菜单裁剪且 URL 直达被拒"), not Step 2's. Should be "Step 4b".
  - Additionally, Step 2b **duplicates** Step 4 Outcome B's scenario: both cover "经 URL 直接访问 → 403" for restricted management pages [F4]. The only addition in 2b is "不泄露策略定义内容（响应体不含策略详情）" — this should be an additional outcome on Step 4, not a separate edge case with misleading numbering.
- **Risk level consistent (35/35)**: `Low` risk + read-only observational workflow — fully consistent.

### 7. Workflow Coverage — 121/130

- **Golden Path existence (47/50)**: `golden_path: true` ✓. 4 contiguous steps (≥3 required) ✓. Golden Path note explicitly describes: "登录进入只读视图 → 搜索分页验证审计完整性 → 直调写接口被拒 → 访问受限管理页被裁剪且 URL 直达被拒" ✓. Steps use domain-level operations ("导航至用户管理、角色管理、授权管理页面", "执行搜索、分页、筛选", "进入租户管理或策略管理") ✓. Semantically tied to PRD Story 5 AC1+AC2 ✓. Minor: Step 3 (direct API call) is a negative/security test included in the Golden Path — defensible since Story 5 AC1 explicitly includes "直接调用写接口时 eiam 返回 403" as acceptance criteria, but it stretches the "primary user story" toward a boundary test.
- **Multi-step coverage depth (41/45)**: State transitions: navigate → search → write-denied → URL-denied ✓. Entity coverage: user/role/auth domains + tenant/strategy management ✓. Error recovery: 403 (Steps 3, 4), 401 (Step 3B), validation-error (1b), session-expired (1c), network-error (2C) ✓. Cross-entity interactions: parent (management) vs child (user/role/auth) access control ✓. Good depth.
- **Workflow completeness against PRD/Design (33/35)**: PRD Story 5 AC1 (write entries hidden + 403 on direct write API) → Steps 1, 3 ✓. AC2 (tenant/strategy menu hidden + 403 on URL access) → Step 4 ✓. Positive audit viewing → Steps 1–2 ✓. Minor gap: PRD spec line 28 says read-only role views "用户/组织/身份源/角色/授权页面" (5 types) but journey covers only "用户/角色/授权" (3 types, matching Story 5 AC1). The 2 uncovered page types (组织/身份源) are auxiliary for Story 5 but present in the spec's role definition.

## Phase 3 — Blindspot Hunt

- **[blindspot]** Dual-surface Golden Path execution model undefined — "### Step 1... @web-e2e ... ### Step 3... @api-functional" — the Golden Path is a single contiguous sequence mixing browser automation (`@web-e2e`) and HTTP client (`@api-functional`) tags. No execution model describes how a single test switches from Playwright/browser to HTTP client mid-flow, or whether the Golden Path should be split into per-surface paths. The rubric's Golden Path criterion evaluates semantic contiguity, but the practical testability of a cross-surface path is a structural concern no dimension captures.

- **[blindspot]** Reasoning audit flagged independently: Fact Table sources are homogeneous — all 4 facts (F1–F4) are sourced from "PRD Story 5 AC1" or "AC2" only, despite `sources` listing both `prd-user-stories.md` and `prd-spec.md`. PRD spec line 28 ("只读查看用户/组织/身份源/角色/授权页面") and line 90 ("无权限（403）→ 原地提示、不跳转登录页") contain verifiable facts relevant to this journey that are not captured in the Fact Table. The Fact Table under-utilizes available source material, limiting traceability depth.

- **[blindspot]** Step 2b and Step 4 Outcome B test the same scenario (URL → 403 for restricted management pages, [F4]) with only a marginal addition ("不泄露策略定义内容"). This redundancy is not a contradiction but a structural inefficiency — the boundary test ("response body doesn't leak strategy definitions") should be an outcome on Step 4, not a separate edge case with stale numbering. No rubric dimension directly penalizes scenario redundancy across steps.

## ATTACKS (summary)

1. [前置条件互斥性]: Step 1 Precondition contradicts Outcome B — "用户/角色/授权页面存在可读数据" vs Outcome B "对应页数据为空时显示空状态占位" — must broaden the step Precondition (e.g., "页面可能含数据或为空") or add per-outcome Preconditions so all Outcomes are reachable.

2. [前置条件互斥性]: Step 3 Precondition contradicts Outcome B — "只读角色已登录" vs Outcome B "令牌缺失/过期·401 未认证" — if logged in, token should be valid, making the 401 Outcome unreachable — must state Precondition neutrally or add per-outcome Preconditions distinguishing authenticated vs token-expired states.

3. [事实依据]: Step 2 Outcome A tagged `[F1]` but contains claims beyond F1 — "[F1]：跨用户、角色、授权三个域页面均可读且数据完整；搜索、分页、筛选返回正确结果" — F1 only covers "增删改入口不渲染"; "数据完整" and "搜索、分页、筛选返回正确结果" are unverified — must add new facts (e.g., F5: read-only search/pagination works) or mark these claims `source: inferred`/`UNKNOWN`.

4. [一致性]: Edge Case numbering stale after Happy Path expansion — "### Step 1b: 搜索输入非法字符" (Step 2's search topic) and "### Step 2b: 经 URL 直接访问策略管理" (Step 4's URL-access topic) — must renumber to match parent steps (1b→2-edge, 2b→4b).

5. [Surface 适配]: `contracts/` directory is empty despite surface rules mandating 50/50 Contract/Journey — no contract specifications exist to verify the balanced split — must add contract specs or acknowledge the gap explicitly.

6. [完整性]: Edge Cases use "Expected Result" (single-outcome) while Happy Path uses "Outcomes" (multi-outcome A/B/C) — "**Expected Result**: 前端即时拦截并标红提示格式要求" — must use consistent multi-outcome format across all steps.

7. [语义纯度]: Step 3 action uses implementation-level language — "只读角色绕过前端裁剪，直接调用任意写接口" — "绕过前端裁剪" and "写接口" are implementation-coupled; should describe domain-level action (e.g., "只读角色尝试修改用户/角色/授权数据").

8. [事实依据]: Generic inferred rule citations — "source: inferred, rule: surface-web.md/required_outcomes 派生" (Step 1 Outcome B, Step 2 Outcome B) — `required_outcomes` is a section name, not a specific rule; must cite the specific outcome type (e.g., `surface-web.md/required_outcomes–validation-error`) as done correctly in Step 2 Outcome C and Step 1b.

9. [blindspot]: Dual-surface Golden Path execution model undefined — "### Step 1... @web-e2e ... ### Step 3... @api-functional" — a single contiguous Golden Path mixing browser automation and HTTP client has no described execution model — must address how the cross-surface path is executed as one test or split into separate per-surface Golden Paths.

10. [blindspot]: Fact Table sources are homogeneous (all from Story 5 ACs only) despite `sources` listing `prd-spec.md` — PRD spec lines 28 and 90 contain verifiable facts ("只读查看用户/组织/身份源/角色/授权页面", "无权限（403）→ 原地提示、不跳转登录页") not captured as facts — must enrich Fact Table with spec-sourced facts to deepen traceability.
