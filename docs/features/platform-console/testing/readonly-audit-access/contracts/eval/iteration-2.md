# Contract Eval — iteration 2

**Journey**: readonly-audit-access
**Rubric**: contract.md (1000 pts, target 850)
**Scorer persona**: Senior QA Engineer (adversarial)
**Date**: 2026-10-08
**ITERATION**: 2

## SCORE: 930/1000

**Pass condition**: PASS (total 930 >= 850; all dimensions above min threshold)

## DIMENSIONS

| Dimension | Score | Max | Min | Pass? |
|-----------|-------|-----|-----|-------|
| 完整性 (Completeness) | 134 | 140 | 85 | PASS |
| 语义纯度 (Semantic Purity) | 159 | 170 | 100 | PASS |
| 前置条件互斥性 (Precondition Exclusivity) | 132 | 140 | 85 | PASS |
| 事实依据 (Fact Alignment) | 132 | 140 | 85 | PASS |
| Surface 适配 (Surface Fitness) | 82 | 90 | 55 | PASS |
| 一致性 (Internal Consistency) | 132 | 140 | 85 | PASS |
| 锚点完整性 (Anchor Integrity) | 90 | 90 | 55 | PASS (no handbook) |
| 前置数据声明 (Fixture Specification) | 69 | 90 | 55 | PASS |
| **Total** | **930** | **1000** | **850** | **PASS** |

## Iteration-1 Attack Resolution Audit

All 10 iteration-1 attacks verified as resolved on the page:

1. **Side-effect implementation coupling** — RESOLVED. No file paths/function names remain. Side-effects now use system-level language ("写操作被服务端权限校验拒绝，返回 403 越权错误").
2. **Step 1 empty-state vs zero-side-effect-error overlap** — RESOLVED. zero-side-effect-error now restricts to "后端身份服务不可达（网络错误）"; "或空状态" removed.
3. **corrected-retry unclassified claim** — RESOLVED. Now carries Invariants "修正后重试返回正确结果，不丢失已加载列表（fact: F7）".
4. **Generic "required_outcomes" citation** — RESOLVED. All 4 inferred outcomes cite specific rule paths (e.g., "surface-web.md–session-expired 派生").
5. **Missing relationship declarations** — RESOLVED. Steps 1/2 declare relationship_type/target_entity (assigned_to, binds, has_role).
6. **Step 2 min_count: 2 insufficient** — RESOLVED. Now min_count: 25 for PlatformUser/Role/Authorization.
7. **Step 3 Output surface mixing** — RESOLVED. write-denied-403 Output now pure API language ("写接口返回 403 越权错误；响应体不泄露敏感数据").
8. **Invalid YAML frontmatter** — RESOLVED. auth_required: null, requires_auth: null.
9. **EiamService missing from bottom-level aggregate** — RESOLVED. Steps 1/2 bottom-level now include EiamService.
10. **Step 1 non-exclusive Preconditions (success-data-present vs zero-side-effect-normal)** — RESOLVED. success-data-present = "角色首次进入页面"; zero-side-effect-normal = "已在只读页面执行搜索/分页/筛选等交互操作".

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution alignment**: The 4 contracts still cover PRD Story 5's two halves (positive audit capability + write denial). Solution addresses the stated problem.

2. **Surface-language regression in Step 4**: While Step 3's surface mixing was fixed, Step 4 url-direct-403 retains the same pattern — a web-e2e outcome describing HTTP "响应体" (response body) alongside page behavior. The reviser fixed the named attack but did not sweep the same class of issue in the sibling web step.

3. **Precondition completeness inconsistency in Step 2**: 2 of 5 Step 2 outcomes omit the "会话有效" precondition that all their siblings declare. This makes those outcomes less self-contained for unique selection and breaks the per-step precondition uniformity.

4. **Fact-claim semantic mismatch**: zero-side-effect-normal cites IAM_READONLY_MENU_PRUNE (a fact about menu rendering/pruning) for a zero-side-effect (no write data mutation) claim. The cited fact's subject is menu pruning, not data-mutation absence.

5. **Entity modeling concerns**: ReadOnlyRole is used as a distinct entity_type but is a constrained Role variant (no seeded readonly role per fact-table). EiamService is infrastructure, not a domain entity. Step 3 Input references PlatformUser/Role/Authorization but Step 3 fixture_spec declares only ReadOnlyRole.

6. **Data type inconsistency**: Step 1 bottom-level EiamService field_constraint value is the string "may be unreachable" while the inline zero-side-effect-error spec uses boolean false — downstream fixture tooling may expect a consistent type.

## Phase 2 — Dimension Scoring

### 1. Completeness — 134/140

- **Every Outcome has four mandatory dimensions (0-50)**: All 14 outcomes across 4 contracts have non-empty Preconditions, Input, Output, State. Side-effect present (defaults to "none" where applicable). **50/50**

- **Journey Invariants section present (0-50)**: All 4 contracts have `## Journey Invariants` with 5 entries each, citing facts (F1-F8, IAM_*). **50/50**

- **Outcomes cover happy path + derived scenarios (0-40)**: Web `validation-error` (Step 2) ✓, `session-expired` (Step 1) ✓. API `unauthorized` (Step 3) ✓. Step 4 (@web-e2e) has no `session-expired` outcome despite surface-web requiring it per web step — session could expire during restricted-URL access. Journey does not define it for Step 4, but the surface rule mandates consideration. **34/40**

### 2. Semantic Purity — 159/170

- **Natural language, not code/regex (0-70)**: No regex, CSS selectors, XPath, or framework assertions anywhere. All Side-effects are now clean system-level descriptions. **70/70**

- **Preconditions declarative, not procedural (0-50)**: Most Preconditions are declarative state. Borderline procedural: Step 2 corrected-retry "只读角色修正非法搜索词为合法值后重新搜索" (action sequence); Step 4 url-direct-403 "只读角色经浏览器地址栏直接粘贴 URL 访问受限管理页" (action modality). Both include (区别于...) annotations but describe actions rather than holding state. **44/50**

- **No implementation coupling in dimension values (0-50)**: Largely clean. Residual coupling: Step 1 session-expired Side-effect "服务端返回 401 未认证" injects an HTTP status code into a web-e2e step; Step 4 url-direct-403 Output/Side-effect uses "响应体不含租户/策略定义详情" — "响应体" (response body) is API-surface language in a web-e2e outcome. **45/50**

### 3. Precondition Exclusivity — 132/140

- **Preconditions distinct across Outcomes (0-50)**: Step 1 now distinct (first-entry vs already-interacting vs empty vs network-error vs session-expired). Step 2 distinct (audit-integrity vs search-empty vs network-error vs validation-error vs corrected-retry). Steps 3, 4 distinct. **48/50**

- **Sufficient to uniquely select an Outcome (0-50)**: Step 2 search-validation-error ("搜索框存在格式校验规则...只读角色待提交非法搜索词") and corrected-retry ("只读角色修正非法搜索词为合法值后重新搜索") omit the "会话有效" precondition that all 3 sibling outcomes declare, making their selection boundary less self-contained — without session-validity stated, an expired-session state could also match. **44/50**

- **Error/boundary Outcomes state triggering conditions (0-40)**: All error/boundary outcomes explicitly state triggers: session-expired ("会话超时过期"), network-error ("搜索请求网络错误"), search-validation-error ("非法搜索词"), unauthorized-401 ("请求未携带有效令牌或令牌已过期"), empty-state ("对应页数据为空"), zero-side-effect-error ("后端身份服务不可达"). **40/40**

### 4. Fact Alignment — 132/140

- **Factual claims traceable to fact_id or UNKNOWN (0-50)**: F1-F8 traceable to journey Fact Table; IAM_* traceable to fact-table.json. Gap: Step 1 zero-side-effect-normal Invariants cites "fact: IAM_READONLY_MENU_PRUNE" for a zero-side-effect (no data mutation) claim, but IAM_READONLY_MENU_PRUNE's subject is "Readonly role menu pruning (write entries not rendered)" — the fact is about menu rendering, not about absence of write side-effects. The journey's own Invariant 3 cites [F1] for the same claim; the contract swapped to a fact whose subject does not match the claim. **44/50**

- **Inferred claims have rule support + source: inferred (0-50)**: All 4 inferred outcomes (session-expired, network-error, search-validation-error, unauthorized-401) carry `source: inferred` and cite specific rule paths (surface-web.md–session-expired, surface-web.md–network-error, surface-web.md–validation-error, surface-api.md–unauthorized). **50/50**

- **No hallucinated claims without classification (0-40)**: All claims are either fact-referenced or inferred-annotated. corrected-retry now cites F7. Minor residual: fact-claim semantic match is slightly off in one instance (above). **38/40**

### 5. Surface Fitness — 82/90

- **Mandatory derived Outcomes present (0-35)**: Web `validation-error` (Step 2) ✓, `session-expired` (Step 1) ✓. API `unauthorized` (Step 3) ✓. **35/35**

- **Surface-appropriate language (0-30)**: Web steps use user-interaction language (navigation, search, URL access). API step uses HTTP status codes (403, 401). However: Step 4 url-direct-403 Output "被服务端权限校验拒绝（403），页面原地提示无权限且数据不变；不跳转登录页；响应体不含租户/策略定义详情" mixes web page language ("页面原地提示") with API language ("响应体不含") in the same @web-e2e outcome. Step 1 session-expired Side-effect "服务端返回 401 未认证" injects API status into web. The reviser fixed Step 3's mixing but left the same pattern in Step 4. **22/30**

- **TUI timeout Outcomes (0-25)**: Not TUI surface. Full marks. **25/25**

### 6. Internal Consistency — 132/140

- **Invariants hold in every Step Contract (0-50)**: No invariant is violated. session-expired redirects to login (session expiry, not 403) so Invariant 4 ("403 原地收敛不跳转登录页") holds. However, Step 2 search-validation-error and corrected-retry omit "会话有效" while their siblings declare it — an internal precondition-completeness inconsistency within the same step (the zero-side-effect invariant implicitly assumes a valid session). **46/50**

- **Cross-Contract state references consistent (0-50)**: Step 2 "页面已加载" → Step 1 "只读视图就绪" ✓. Step 3 "持有有效令牌" → Step 1 valid session ✓. Step 4 "已登录" consistent ✓. No dangling references. **48/50**

- **Outcome Preconditions consistent with preceding State changes (0-40)**: Step 2 achievable from Step 1 ✓. Step 3 achievable from Step 1 ✓. Step 4 consistent ✓. Minor: Step 3 unauthorized-401 "令牌已过期" semantically overlaps with Step 1 session-expired "会话超时过期" — both describe expired sessions, distinguished only by surface (API vs Web). The cross-face model justifies this but the overlap could confuse cross-step state flow. **38/40**

### 7. Anchor Integrity — 90/90

No handbook exists for any surface type. Per rubric: "When no handbook exists for the surface type: score full marks (90/90)." All contracts note "Handbook missing. Anchor fields empty per graceful-degradation rule." **90/90**

### 8. Fixture Specification — 69/90

- **Entity completeness (veto item, 0-35)**: Steps 1, 2, 4 declare entities they reference. However:
  - Step 3 write-denied-403 Input "只读角色尝试修改用户/角色/授权数据（创建、编辑、删除），直调写接口" references PlatformUser/Role/Authorization, but Step 3 fixture_spec.entities declares only ReadOnlyRole. Per rubric, entity types referenced in Input must be in fixture_spec.entities. Denial semantics (403 before data access) partially justify omission, but the textual reference triggers the entity-completeness rule.
  - `ReadOnlyRole` is used as a distinct entity_type but is a constrained Role variant (fact-table notes no seeded readonly role — only super_admin + admin). Semantic verification against the domain model cannot confirm ReadOnlyRole as a domain entity; it should be `Role` with field_constraints (e.g., role_type: readonly).
  - `EiamService` is infrastructure (backend service availability), not a domain model entity.
  **18/35**

- **Relationship and constraint coverage (0-30)**: Steps 1, 2 declare relationships (assigned_to, binds, has_role). Step 2 declares PlatformUser→Role (has_role) AND Role→PlatformUser (assigned_to) bidirectionally — redundant but not wrong. Steps 3, 4 single-entity (ReadOnlyRole only) — full marks for single-entity. **26/30**

- **Minimum data quantity declarations (0-25)**: Step 2 min_count: 25 for PlatformUser/Role/Authorization — sufficient for multi-page pagination. Step 1 success-data-present min_count: 1 reasonable; empty-state min_count: 0 correct. Steps 3, 4 min_count: 1 reasonable. **25/25**

## Phase 3 — Blindspot Hunt

- **[blindspot] Step 4 url-direct-403 retains the surface-mixing pattern that Step 3 was fixed for** — "被服务端权限校验拒绝（403），页面原地提示无权限且数据不变；不跳转登录页；响应体不含租户/策略定义详情（不泄露受限管理内容）" — a @web-e2e outcome should describe page observables ("页面不显示受限管理内容"), not HTTP "响应体". The reviser fixed the named attack (Step 3) but did not sweep the same class of issue in the sibling web step. Reasoning audit flagged this independently of dimension scoring.

- **[blindspot] Step 1 bottom-level EiamService field_constraint value type inconsistent with inline spec** — bottom-level "value: \"may be unreachable\"" (string) vs inline zero-side-effect-error "value: false" (boolean) — downstream fixture tooling may expect a consistent type for the `reachable` field across the aggregate and inline subsets.

- **[blindspot] Step 3 fixture_spec entity coverage gap masked by denial semantics** — Input "只读角色尝试修改用户/角色/授权数据" references PlatformUser/Role/Authorization but fixture_spec declares only ReadOnlyRole. The 403-denial-before-data-access semantics may justify not seeding target data, but the contract must either declare the referenced entities or rewrite the Input to avoid entity-type references (e.g., "只读角色尝试调用写接口"). Leaving the reference unbacked is an entity-completeness gap that the rubric treats strictly.

## ATTACKS (summary)

1. [Surface 适配]: Step 4 url-direct-403 mixes web and API surface language — "被服务端权限校验拒绝（403），页面原地提示无权限且数据不变；不跳转登录页；响应体不含租户/策略定义详情（不泄露受限管理内容）" — web-e2e outcome must describe page observables, not HTTP "响应体"; replace with "页面不显示受限管理内容".

2. [前置数据声明]: Step 3 fixture_spec omits entities referenced in Input — "只读角色尝试修改用户/角色/授权数据（创建、编辑、删除），直调写接口" — fixture_spec only declares ReadOnlyRole; must declare PlatformUser/Role/Authorization or rewrite Input to avoid entity-type references.

3. [一致性]: Step 2 search-validation-error and corrected-retry omit "会话有效" while 3 siblings declare it — "搜索框存在格式校验规则（长度上限、禁用符号）；只读角色待提交非法搜索词" / "只读角色修正非法搜索词为合法值后重新搜索" — must add "会话有效" for self-contained precondition completeness.

4. [事实依据]: zero-side-effect-normal cites IAM_READONLY_MENU_PRUNE (menu pruning fact) for a zero-side-effect claim — "只读操作零副作用——查看/搜索/分页不触发任何数据变更（fact: IAM_READONLY_MENU_PRUNE）" — cite a fact whose subject matches zero-write-side-effect (journey uses [F1] for the same invariant); IAM_READONLY_MENU_PRUNE is about menu rendering.

5. [前置数据声明]: ReadOnlyRole is a constrained Role variant, not a separate domain entity; EiamService is infrastructure — "entity_type: \"ReadOnlyRole\"" / "entity_type: \"EiamService\"" — use Role with field_constraints (role_type: readonly); clarify EiamService as test-infra fixture, not a domain entity.

6. [完整性]: Step 4 (@web-e2e) lacks session-expired outcome — surface-web mandates session-expired per web step — must add a session-expired outcome or explicitly justify its omission (e.g., session expiry already covered journey-wide in Step 1).

7. [语义纯度]: Preconditions are borderline procedural action sequences — "只读角色修正非法搜索词为合法值后重新搜索" / "只读角色经浏览器地址栏直接粘贴 URL 访问受限管理页" — describe the holding state (post-correction / URL-paste attempt) rather than the action sequence.

8. [blindspot]: Step 1 bottom-level EiamService field_constraint value "may be unreachable" (string) inconsistent with inline zero-side-effect-error "value: false" (boolean) — must align the `reachable` field type across aggregate and inline subsets to prevent downstream fixture-parser ambiguity.
