# Journey Eval — iteration 3

**Journey**: readonly-audit-access
**Rubric**: journey.md (1000 pts, target 850)
**Scorer persona**: Senior QA Engineer (adversarial)
**Date**: 2026-10-08

## SCORE: 890/1000

**Pass condition**: PASS (total 890 ≥ 850; all dimensions above min threshold)

## DIMENSIONS

| Dimension | Score | Max | Min | Pass? |
|-----------|-------|-----|-----|-------|
| 完整性 (Completeness) | 162 | 175 | 105 | PASS |
| 语义纯度 (Semantic Purity) | 157 | 175 | 105 | PASS |
| 前置条件互斥性 (Precondition Exclusivity) | 114 | 130 | 80 | PASS |
| 事实依据 (Fact Alignment) | 107 | 130 | 80 | PASS |
| Surface 适配 (Surface Fitness) | 111 | 130 | 80 | PASS |
| 一致性 (Internal Consistency) | 118 | 130 | 80 | PASS |
| 工作流覆盖度 (Workflow Coverage) | 121 | 130 | 80 | PASS |
| **Total** | **890** | **1000** | **850** | **PASS** |

## Iteration-2 Attack Resolution Check

| iter-2 Attack | Status | Evidence |
|---------------|--------|----------|
| #1 Step 1 Precondition contradicts Outcome B (data exists vs data empty) | **RESOLVED** | Step 1 Precondition now: "用户/角色/授权页面可能含可读数据，亦可能为空" — Outcome B reachable |
| #2 Step 3 Precondition contradicts Outcome B (logged in vs token expired) | **RESOLVED** | Step 3 Precondition now: "只读角色持有令牌（令牌可能有效，亦可能缺失/过期）" — Outcome B reachable; explicit mutual-exclusivity note added |
| #3 Step 2 Outcome A tagged [F1] but claims beyond F1 | **RESOLVED** | Step 2 Outcome A now tagged [F5][F7]; F5/F7 added to Fact Table covering read-only viewing + search/pagination correctness |
| #4 Edge Case numbering stale (Step 1b search, Step 2b URL-access) | **RESOLVED** | Step 1b renamed to "Step 2b: 搜索输入非法字符" (parent Step 2 = search); old "Step 2b: 经 URL 直接访问策略管理" removed and merged into Step 4 Outcome B |
| #5 contracts/ empty, 50/50 unverifiable | **RESOLVED (acknowledged)** | New "Contract Coverage Gap" section explicitly registers the gap and downstream gen-contracts plan |
| #6 Edge Cases use single-outcome "Expected Result" vs Happy Path multi-outcome | **RESOLVED** | All Edge Cases now use multi-outcome "Outcomes" format (A/B) |
| #7 Step 3 action implementation-level ("绕过前端裁剪，直接调用任意写接口") | **RESOLVED** | Step 3 User Action now: "只读角色尝试修改用户/角色/授权数据（创建、编辑、删除）" — domain-level |
| #8 Generic inferred rule citations ("required_outcomes 派生") | **RESOLVED** | All inferred outcomes now cite specific rules (surface-web.md–network-error, surface-api.md–unauthorized, surface-web.md–validation-error, surface-web.md–session-expired) |
| #9 [blindspot] Dual-surface Golden Path execution model undefined | **RESOLVED** | New "跨面执行模型" paragraph describes browser→token-extract→HTTP→browser sequential execution within one test case |
| #10 [blindspot] Fact Table sources homogeneous (all Story 5 ACs) | **RESOLVED** | Fact Table expanded to F1–F8; F5/F6/F7/F8 sourced from prd-spec.md (Who 表, Business Flow #8, 只读查看, E7 分页边界) |

All 10 iteration-2 attacks addressed. Remaining deductions are **residual/new issues** not fully closed by the revision.

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Step 2b Outcome A fact-tag mismatch.** Outcome A is tagged `[F1]` but describes search-input validation interception ("前端即时拦截并标红提示格式要求"), while F1 covers only "增删改入口不渲染". Search validation is an inferred `validation-error` behavior (the step Precondition correctly cites `surface-web.md–validation-error`), but the Outcome itself is tagged with an irrelevant fact. The tag implies factual grounding that does not cover the claim.

2. **Step 4 Outcome B unclassified claim.** Outcome B is tagged `[F4][F6]` but asserts "响应体不含租户/策略定义详情（不泄露受限管理内容）". F4 covers "页面提示无权限且数据不变"; F6 covers "403 原地不跳转登录页". Neither fact covers response-body content non-leakage. This claim is neither factual-with-traceability nor inferred-with-annotation.

3. **F7 over-claims PRD source.** F7 states "浏览、搜索、分页、筛选以只读方式正常执行并返回正确结果" sourced from "prd-spec.md（只读查看页面）+ Story 5 AC1". PRD spec line 28 says only "只读查看用户/组织/身份源/角色/授权页面...不可增删改" — it does not explicitly guarantee that search/filter "返回正确结果". The "返回正确结果" portion is a reasonable inference from "只读查看" but is presented as a direct fact.

4. **Step 1c/1d numbering residual.** "Step 1c: 审计浏览中途会话过期" and "Step 1d: 审计角色查看时不触发任何状态变更" are numbered under Step 1 (login + navigate) but their content (session expiry during browsing, query/pagination/sorting side-effects) belongs to Step 2's domain or is cross-cutting. Step 1d's User Action ("只读角色在只读页面上进行查询/翻页/排序") duplicates Step 2's action ("执行搜索、分页、筛选").

## Phase 2 — Dimension Scoring

### 1. Completeness — 162/175

- **Metadata (44/45)**: name `readonly-audit-access` kebab-case ✓; `risk_level: Low` justified by read-only content ✓; `golden_path: true` ✓; `surface_types: ["web","api"]` ✓; `sources` lists both PRD files ✓; `surface_keys: ["haven-console","eiam"]` present ✓. Minor: `surface_keys` values not explained in-document.
- **Steps complete (67/70)**: 4 Happy Path steps + 3 Edge Cases; each has Precondition + User Action + multi-outcome Outcomes (A/B or A/B/C). All steps now use consistent multi-outcome format (iteration-2's single-outcome "Expected Result" inconsistency resolved). Coherent ordered sequence. Minor: 3 dedicated Edge Cases is fewer than iteration-2's 5, but consolidated boundary outcomes within Happy Path steps absorb the coverage.
- **Outcomes cover happy + derived (51/60)**: Happy path ✓. Mandatory Web derived: `validation-error` (Step 2b) ✓, `session-expired` (Step 1c) ✓. Mandatory API derived: `unauthorized` (Step 3 Outcome B) ✓. Additional Web: `network-error` (Step 2 Outcome C) ✓. Gap: no API-level additional outcomes (surface-api.md lists `validation-error` 400, `not-found` 404, `rate-limit` 429, `internal-error` 500, `pagination`) for the single API step — only 403 and 401 covered.

### 2. Semantic Purity — 157/175

- **Natural language (63/70)**: Outcomes in Chinese natural language; no regex/CSS/XPath/framework assertion calls. `403`/`401` are system-level status codes — acceptable. `@web-e2e`/`@api-functional` tags required by surface rules. Residual: the "跨面执行模型" guidance paragraph contains "提取会话令牌" and "HTTP 段发起越权写请求" — implementation-level mechanism terms in a journey-level note (not in outcomes/actions, but visible to downstream agents).
- **Preconditions declarative (50/55)**: Preconditions are now declarative state descriptions: "只读/审计角色已登录；用户/角色/授权页面可能含可读数据，亦可能为空" ✓; "只读角色持有令牌（令牌可能有效，亦可能缺失/过期）" ✓. Step 2b Precondition "搜索框存在格式校验规则；只读角色待提交非法搜索词" — "待提交非法搜索词" is slightly procedural (describes an imminent action rather than a state).
- **No implementation coupling (44/50)**: Step 3 action now domain-level ("尝试修改用户/角色/授权数据") ✓. `eiam` confined to Setup/Invariants. Residual: Invariant 2 retains "eiam 对所有写端点做服务端校验" — "写端点" is API-level vocabulary. The "跨面执行模型" note's "提取会话令牌" is an implementation mechanism. Step 3 Precondition "后端对所有用户/角色/授权的写操作做服务端权限校验" is system-level but acceptable.

### 3. Precondition Exclusivity — 114/130

- **Distinct across Outcomes (42/50)**: Outcomes within each step have distinguishable trigger conditions (data-present vs data-empty; search-match vs no-match vs network-error; 403-valid-token vs 401-expired-token; menu-hidden vs URL-blocked). Step 3 Outcome B explicitly states "语义上与 Outcome A 互斥——403 为已认证但越权，401 为未认证". However, trigger conditions remain embedded in outcome descriptions, not stated as separate per-outcome Preconditions — the step-level Precondition is shared.
- **Sufficient to uniquely select (40/45)**: Iteration-2's two precondition-outcome mismatches (Step 1 data-exists vs data-empty; Step 3 logged-in vs token-expired) are **resolved** — Preconditions now neutral ("可能含可读数据，亦可能为空"; "令牌可能有效，亦可能缺失/过期"). All Outcomes now reachable. Residual: trigger conditions live in outcome text, not dedicated per-outcome Precondition fields, so selection requires reading outcome descriptions.
- **Missing Preconditions for error/boundary (32/35)**: Error/boundary outcomes state triggers in descriptions ("搜索请求网络错误时"). Step 2b and 1c carry `source: inferred` Precondition annotations with specific rule citations. Minor: triggers in outcome text, not dedicated per-outcome Precondition fields.

### 4. Fact Alignment — 107/130

- **Factual claims traceable to fact_id or UNKNOWN (40/50)**: F1–F8 traceable to PRD Story 5 AC1/AC2 + prd-spec.md (verified against source files). Iteration-2's Step 2 Outcome A [F1] mis-tag **resolved** → now [F5][F7]. **New/residual mis-tags**: (a) Step 2b Outcome A tagged `[F1]` but describes search-validation interception — F1 covers "增删改入口不渲染", irrelevant to search validation; (b) Step 4 Outcome B tagged `[F4][F6]` but claims "响应体不含租户/策略定义详情" — neither F4 nor F6 covers response-body content; (c) Step 1d Outcome A tagged `[F1]` for "零写副作用" — F1 is entry-non-rendering, a semantic stretch.
- **Inferred claims have rule support + `source: inferred` (43/45)**: All inferred outcomes now cite **specific** rules (iteration-2's generic "required_outcomes 派生" resolved): `surface-web.md–network-error`, `surface-api.md–unauthorized`, `surface-web.md–validation-error`, `surface-web.md–session-expired`. Annotation quality uniform. Minor: F7's "返回正确结果" is presented as a fact but is partially inferred (spec says "只读查看", not "search returns correct results").
- **No hallucinated unclassified claims (24/35)**: Step 4 Outcome B's "响应体不含租户/策略定义详情（不泄露受限管理内容）" is an unclassified claim — not in F4/F6, not marked `source: inferred`. Step 2b Outcome A's search-validation behavior tagged `[F1]` is a mis-classification (should be inferred `validation-error`). These are partial hallucinations/mis-classifications.

### 5. Surface Fitness — 111/130

- **Mandatory derived Outcomes present (48/50)**: Web `validation-error` (Step 2b) ✓, Web `session-expired` (Step 1c) ✓, API `unauthorized` (Step 3 Outcome B) ✓. All mandatory derived outcomes present for both declared surfaces. Near-full marks.
- **Test strategy proportions (30/45)**: `contracts/` directory still empty, but now explicitly acknowledged in new "Contract Coverage Gap" section: "Contract 侧规范（contracts/ 目录）当前为空...下游 gen-contracts 产出后即满足比例要求." Acknowledgment mitigates the gap (iteration-2's attack #5 accepted explicit acknowledgment as resolution), but the 50/50 Contract/Journey split remains structurally unverifiable from artifacts. Journey side is deep (4 Happy Path + 3 Edge, multi-outcome branching); Contract side entirely absent.
- **Surface-specific environment realistic (33/35)**: Web steps use browser navigation, URL bar, search/pagination/filter ✓. API step uses direct calls, 403/401 ✓. Cross-surface execution model now described ✓. `@web-e2e`/`@api-functional` tags correctly applied ✓. Invariant 4 correctly distinguishes 403-stay vs session-expired-redirect, matching PRD spec line 90 ✓. Minor: execution model assumes session token extractable from browser context (see blindspot).

### 6. Internal Consistency — 118/130

- **Invariants hold in every Step (48/50)**: Five invariants verified across all steps. Inv 1 (no write entries rendered): Steps 1, 2, 4 ✓. Inv 2 (403 final boundary): Steps 3, 4 ✓. Inv 3 (zero side-effects): Steps 1, 2, 1d ✓. Inv 4 (403 stays, not login redirect): Steps 3, 4; Step 1c contrasts session-expired ✓. Inv 5 (audit goal achievable): Steps 1, 2 ✓ — now grounded in F5/F7 rather than the overstated "数据完整". Minor: Invariant 2 retains "写端点" API-level vocabulary.
- **Cross-Step references consistent (35/45)**: Iteration-2's two numbering errors (Step 1b search, Step 2b URL-access) **resolved** — Step 1b renamed to Step 2b (correct parent), old Step 2b URL-access merged into Step 4 Outcome B (redundancy eliminated). **Residual**: "Step 1c: 审计浏览中途会话过期" and "Step 1d: 审计角色查看时不触发任何状态变更" remain numbered under Step 1 (login+navigate) but their content (session expiry during browsing; query/pagination/sorting side-effects) belongs to Step 2's domain or is cross-cutting. Step 1d User Action "只读角色在只读页面上进行查询/翻页/排序" duplicates Step 2's "执行搜索、分页、筛选".
- **Risk level consistent (35/35)**: `Low` risk + read-only observational workflow — fully consistent.

### 7. Workflow Coverage — 121/130

- **Golden Path existence (47/50)**: `golden_path: true` ✓. 4 contiguous steps (≥3 required) ✓. Golden Path note: "登录进入只读视图 → 搜索分页验证审计完整性 → 直调写接口被拒 → 访问受限管理页被裁剪且 URL 直达被拒" ✓. Steps use domain-level operations ✓. Anchored to PRD Story 5 AC1+AC2 ✓. Cross-surface execution model now described (browser→token→HTTP→browser) ✓. Minor: Step 3 (API write-denial) is a negative/security test in the Golden Path — defensible since Story 5 AC1 explicitly includes "直接调用写接口时 eiam 返回 403" as acceptance criteria, but stretches the Golden Path toward boundary testing.
- **Multi-step coverage depth (42/45)**: State transitions: navigate → search → write-denied → URL-denied ✓. Entity coverage: user/role/auth + tenant/strategy management ✓. Error recovery: 403 (Steps 3, 4), 401 (Step 3B), validation-error (2b), session-expired (1c), network-error (2C) ✓. Cross-entity: parent (management) vs child (user/role/auth) access control ✓. Good depth.
- **Workflow completeness against PRD/Design (32/35)**: PRD Story 5 AC1 (write entries hidden + 403 on direct write API) → Steps 1, 3 ✓. AC2 (tenant/strategy menu hidden + 403 on URL access) → Step 4 ✓. Positive audit viewing → Steps 1–2 ✓. Minor gap: PRD spec line 28 says read-only role views "用户/组织/身份源/角色/授权页面" (5 types) but journey steps cover only "用户/角色/授权" (3 types, matching Story 5 AC1). F5 mentions all 5 types but no step exercises 组织/身份源 pages — auxiliary for Story 5 but present in the spec's role definition.

## Phase 3 — Blindspot Hunt

- **[blindspot]** Cross-surface execution model assumes session token is extractable from browser context — "复用浏览器上下文中已认证只读角色的会话令牌，对该令牌直接发起写请求以断言 403/401" — if authentication is httpOnly-cookie-based (common for eiam/console SSO), the token is not directly extractable via Playwright JS evaluation; the model would need to forward cookies via `context.cookies()` or use a shared HTTP session. The journey does not distinguish bearer-token auth from cookie-based auth, leaving the cross-surface handoff mechanism underspecified. No rubric dimension captures the testability of the cross-surface token-extraction mechanism.

- **[blindspot]** Step 1d is semantically a duplicate of Step 2's scope — Step 1d User Action "只读角色在只读页面上进行查询/翻页/排序" vs Step 2 User Action "只读角色在用户、角色、授权列表分别执行搜索、分页、筛选". Both cover query/pagination/sorting on read-only pages. Step 1d adds only the "zero side-effects" invariant angle, which is already captured by Invariant 3 ("只读操作零副作用"). The edge case is structurally redundant with Step 2 + Invariant 3 — no rubric dimension directly penalizes scenario redundancy between an Edge Case and a Happy Path step.

- **[blindspot]** Fact Table F7 over-claims the PRD source — F7 asserts "浏览、搜索、分页、筛选以只读方式正常执行并返回正确结果" sourced from "prd-spec.md（只读查看页面）+ Story 5 AC1（审计目的）", but the spec line 28 says only "只读查看...页面...不可增删改" and Story 5 AC1 says "打开用户/角色/权限页面" + "增删改入口不渲染". Neither source explicitly guarantees that search/filter "返回正确结果" — that is a reasonable inference from "只读查看" but is presented as a fact in the Fact Table. A fact that over-claims its source undermines the traceability contract the Fact Table is meant to provide. This overlaps Fact Alignment but the root cause is a Fact Table accuracy defect, not an outcome-tagging defect.

## ATTACKS

1. [事实依据]: Step 2b Outcome A mis-tagged [F1] — "**Outcome A（前端即时拦截·不提交）** [F1]：前端即时拦截并标红提示格式要求" — F1 covers "增删改入口不渲染", not search-input validation; must tag the outcome as `source: inferred` (validation-error, matching the Precondition annotation) or add a dedicated fact for search-validation interception behavior.

2. [事实依据]: Step 4 Outcome B unclassified claim — "**Outcome B（URL 直达被拒·403·不泄露定义）** [F4][F6]：...响应体不含租户/策略定义详情（不泄露受限管理内容）" — neither F4 ("页面提示无权限且数据不变") nor F6 ("403 原地不跳转登录页") covers response-body content non-leakage; must add a new fact (e.g., F9: 403 response body excludes restricted management content) or mark the claim `source: inferred` with a security-best-practice rule citation.

3. [一致性]: Step 1c/1d numbering stale relative to content — "### Step 1c: 审计浏览中途会话过期" and "### Step 1d: 审计角色查看时不触发任何状态变更" — session-expiry-during-browsing and query/pagination/sorting side-effects belong to Step 2's domain (search/pagination) or are cross-cutting, not Step 1 (login+navigate); Step 1d User Action "只读角色在只读页面上进行查询/翻页/排序" duplicates Step 2's action — must renumber to "Step 2c/2d" or relabel as cross-cutting edge cases.

4. [语义纯度]: 跨面执行模型 uses implementation-level terms — "提取会话令牌 → HTTP 段发起越权写请求 → 浏览器段继续 URL 直达验证" — token extraction and HTTP-segment invocation are test-implementation mechanisms, not journey-level behavior; should describe the cross-surface handoff at domain level or relocate to a separate test-execution appendix outside the journey body.

5. [Surface 适配]: contracts/ directory still empty despite acknowledgment — "Contract 侧规范（contracts/ 目录）当前为空...下游 gen-contracts 产出后即满足比例要求" — explicit acknowledgment mitigates the gap but the 50/50 Contract/Journey proportion remains unverifiable from current artifacts; must produce contract specs (or have downstream gen-contracts run) to close the gap before the proportion claim is satisfied.

6. [事实依据]: F7 over-claims PRD source — "F7 | 只读角色对用户/角色/授权页面的浏览、搜索、分页、筛选以只读方式正常执行并返回正确结果（审计可读性） | PRD prd-spec.md（只读查看页面）+ Story 5 AC1（审计目的）" — spec line 28 says only "只读查看...页面...不可增删改", not that search/filter "返回正确结果"; must either mark the "返回正确结果" portion as `source: inferred` within F7 or cite the specific spec line that guarantees search/filter correctness.

7. [完整性]: No API-level additional boundary outcomes — Step 3 (the only `@api-functional` step) covers only 403 (Outcome A) and 401 (Outcome B); surface-api.md additional outcomes (`validation-error` 400, `not-found` 404, `rate-limit` 429, `internal-error` 500, `pagination`) are not considered even though the API step involves write operations that could yield 400/404/409 — must add at least one API additional boundary outcome for write-endpoint completeness.

8. [blindspot]: Cross-surface token extraction mechanism underspecified — "复用浏览器上下文中已认证只读角色的会话令牌，对该令牌直接发起写请求" — the model does not distinguish bearer-token auth (extractable) from httpOnly-cookie auth (not directly extractable via JS); must specify the auth mechanism or describe a cookie-forwarding handoff so a downstream agent can implement the cross-surface step deterministically.
