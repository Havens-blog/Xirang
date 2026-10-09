# Contract Evaluation Report — Iteration 2

- **Journey**: identity-permission-governance-golden-path
- **Surfaces**: web (haven-console) + api (eiam)
- **Rubric**: contract.md (1000 pts, target 850, per-dimension thresholds)
- **Scorer persona**: Senior QA Engineer (adversarial)
- **Iteration**: 2
- **Fact Table**: D:\Haven\.forge\fact-table.json (23 facts, IAM_* IDs authoritative; F-* = journey legacy, remapped)
- **Handbook**: absent → Anchor Integrity scored full marks (backward-compatible)

---

## Iteration-1 Attack Re-Verification

| iteration-1 attack | status | evidence |
|---|---|---|
| F-* IDs cited in Invariants untraceable to fact-table.json (used IAM_* only) | RESOLVED | Step 1–5 Invariants now cite `IAM_CONSOLE_DIRECT_EIAM_NO_BFF`, `IAM_IDENTITY_SOURCE_LDAP`, `IAM_POLICY_DEFINITION_WRITE_SCOPE`, `IAM_AUTH_CACHE`, `IAM_UNAUTHORIZED_401`, `IAM_ROLE_REVOKE`, etc. — all exist in fact-table.json. No F-* IDs remain in dimension prose. |
| `dependency-delete-blocked` / `cross-tenant-policy-denied` mis-tagged `source: inferred` while citing facts | RESOLVED | Step 3 dependency-delete-blocked re-tagged `<!-- fact: IAM_DEPENDENCY_DELETE_BLOCK (UNKNOWN ...) -->`; Step 4 cross-tenant-policy-denied and Step 5 cache-delay-boundary carry `<!-- 本 Outcome 由 fact 派生（非 required_outcomes 规则），故标记为 factual 而非 inferred -->`. |
| Happy-path Step 1 bundles unimplemented password-reset side-effect | RESOLVED (new blindspot introduced) | Step 1 success Side-effect now "密码重置行为不在此 success 契约中断言（见独立 Outcome initial-password-reset-unknown）"; split into a dedicated `initial-password-reset-unknown` Outcome with UNKNOWN Output/State. Honest split, but the new Outcome is not executable as a test (see blindspot 1). |
| Step 3 delete-blocked tests a different action than step-action | PARTIALLY addressed | Action discriminator added: "操作动作为删除角色或策略（区别于本步骤的创建动作）". Selection now logically possible, but the Outcome still sits structurally inside a create-step Contract (action mismatch persists). |
| Forward cross-Contract reference Step 3 → Step 4 | PARTIALLY addressed | Precondition now explicitly scoped: "本 Outcome 为 post-journey 边缘场景，须在授权关系建立后执行". Ambiguity resolved, but linear Step 3 executability still broken for this Outcome. |
| Action-mixing (delete in create-step, policy-CRUD in assign-step) | PARTIALLY addressed | Action discriminators added on both Outcomes; structural relocation not performed. |
| Cache 30s (journey) vs 60s (Contract) contradiction | RESOLVED at Contract scope | All Contract Invariants consistently state 60s with "PRD ≤30 秒未获代码支持，UNKNOWN — see IAM_AUTH_CACHE". Contracts follow the authoritative IAM_AUTH_CACHE fact; journey's 30s flagged UNKNOWN. |
| API `unauthorized` required per authenticated endpoint — missing on Steps 1,2,3,5 | NOT addressed | Steps 1, 2, 3, 5 still carry no unauthorized Outcome for their authenticated eiam endpoints. Only Step 4 has one. |
| Implementation coupling (file:line, function names, "casbin") in dimension values | RESOLVED | All file:line refs, `middleware.CheckPermission`, `errs.ErrForbidden`, `casbin`, `handler.go:NN` removed from dimension values; moved into HTML comments. Dimension prose is now natural language. |
| Missing `relationship_type`/`parent_entity` for multi-entity Contracts | RESOLVED | Relationships now declared: PlatformUser `created_by`/`assigned_to`/`edited_by`/`same_username`; LdapServer `managed_by`; Role `binds_policy`; PolicyDefinition `bound_to_role`; AuthorizationRecord `links`. |
| Step 1 bottom PlatformUser min_count 0 | acceptable | Happy path legitimately needs 0 pre-existing users; conflict/concurrent inline specs declare min 1. No change needed. |

---

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem→Solution**: 5 Contracts map 1:1 to the journey's domain-action sequence (create user → configure source → create role+bind policy → assign → revoke+converge). Solution addresses the stated problem. ✓
2. **Solution→Evidence**: F-* remapping to IAM_* is now consistent with the authoritative Fact Table. Traceability restored. ✓
3. **Evidence→Success Criteria**: The new `initial-password-reset-unknown` Outcome has Output/State entirely "UNKNOWN — 不可断言". This is honest (admits the unimplemented endpoint) but yields a test contract no QA agent can assert against — a reasoning gap between "documented gap" and "executable test". (→ blindspot 1)
4. **Self-contradiction check**:
   - Cache ceiling: Contract-level consistent (60s, 30s flagged UNKNOWN). Cross-artifact discrepancy with journey.md remains but is explicitly flagged — not a Contract-internal contradiction.
   - Forward reference: Step 3 dependency-delete-blocked still depends on Step 4's authorization state; now scoped "post-journey" but not relocated. Linear executability broken for that Outcome within Step 3.
   - Action-mixing: Step 3 (create) still hosts a delete-blocked Outcome; Step 4 (assign) still hosts a policy-CRUD-denied Outcome. Action discriminators mask the mismatch but do not resolve the structural scope/Outcome misalignment.
   - SC consistency deep-dive: Contracts use Outcomes, not SC/InScope pairs — the structured SC clustering check is N/A. The bidirectional satisfiability check on Outcomes within each Step finds no intra-Step mutual-exclusion (action discriminators separate them), but finds cross-action placement friction (Outcomes whose action ≠ step-action).

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 130/140

| Criterion | Score | Notes |
|---|---|---|
| Every Outcome has all four mandatory dimensions non-empty (Preconditions, Input, Output, State) | 50/50 | All 5 Contracts × all Outcomes carry non-empty Preconditions/Input/Output/State. Step 1 `initial-password-reset-unknown` Output/State are "UNKNOWN — …" which is a non-empty UNKNOWN-marked value (acceptable per Fact Alignment UNKNOWN convention). Side-effect present or "none". |
| Journey Invariants section present with ≥1 entry in every Contract | 50/50 | All 5 Contracts have `## Journey Invariants` with multiple entries. |
| Outcomes cover happy path + required derived scenarios per surface type | 30/40 | Happy path present in every step. Web `validation-error` (Step 1) + `session-expired` (Step 4) present. API `unauthorized` present only in Step 4 — missing on Steps 1, 2, 3, 5 (each calls an authenticated eiam endpoint: POST /api/user/create, /api/identity_source/save, /api/role/create, /api/role/unassign). The per-authenticated-endpoint mandate is unmet on 4/5 endpoints. −10. |

### 2. Semantic Purity (语义纯度) — 154/170

| Criterion | Score | Notes |
|---|---|---|
| Dimension values use natural language, not code/regex | 64/70 | No regex, CSS/XPath, framework assertion calls, file:line refs, or internal function names remain in dimension values (all moved to HTML comments). Residual: inline fact-ID tags and "代码实测" annotations mixed into prose — e.g. Step 5 Invariants "撤销生效延迟上限 60 秒（代码实测；PRD ≤30 秒未获代码支持，UNKNOWN — see IAM_AUTH_CACHE）" and Step 3 Side-effect "（fact: IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）". These are traceability metadata leaking into behavioral description; a purer Contract keeps citations in HTML comments only. HTTP status codes (401/403) in Output are surface-appropriate for api surface, not a purity violation. |
| Preconditions are declarative state descriptions, not procedural steps | 48/50 | Preconditions are uniformly declarative state ("管理员已登录且会话有效；…"). Action discriminators ("操作动作为新建用户" / "操作动作为删除角色或策略（区别于本步骤的创建动作）") are declarative action-context statements. Mild intra-Contract dependency: Step 1 `initial-password-reset-unknown` Preconditions reference "success 已创建新用户" (state from a prior Outcome) — reads as state, acceptable. |
| No implementation coupling in dimension values | 42/50 | API endpoint paths, file:line, and function names removed. Residual: State/Side-effect name the internal cache mechanism — Step 4 State "授权策略缓存重载使新权限对后续请求即时生效", Step 5 State "授权策略缓存重载使权限对后续请求收敛为拒绝". "缓存重载" describes the implementation mechanism (casbin policy cache autoload); a system-level description would be "权限变更在 60 秒内对后续请求生效/收敛" without naming the cache-reload path. Minor. |

### 3. Precondition Exclusivity (前置条件互斥性) — 128/140

| Criterion | Score | Notes |
|---|---|---|
| Preconditions are distinct across Outcomes within each Step | 46/50 | Within each Contract, Outcome Preconditions are semantically distinct. Step 4 `session-expired` ("会话已超时过期", web mid-workflow) vs `unauthorized-401` ("令牌已过期/无效", api direct call) overlap on "no valid session" but are distinguishable by surface and workflow context. Residual ambiguity: an expired-token request could match both without the surface/context discriminator. |
| Preconditions sufficient to uniquely select an Outcome | 44/50 | Action discriminators ("操作动作为删除…（区别于本步骤的创建动作）"; "操作动作为策略定义增删改（区别于本步骤的角色分配动作）") now make Outcomes logically selectable. But the discriminators reveal that several Outcomes' action ≠ step-action — a downstream agent driving the step's nominal action cannot reach these Outcomes without action-switching. Selection is achievable but placement creates friction. |
| Error/boundary Outcomes state triggering conditions explicitly | 38/40 | All boundary Outcomes state triggers (format violation, duplicate, concurrent stale version, backend unavailable, cache window ≤60s, last-page sole record, mid-workflow expiry, missing creds, cross-tenant, dependency referenced). Strong. Minor: Step 1 `initial-password-reset-unknown` trigger "该重置端点实现状态为 UNKNOWN" is a status condition, not a behavioral trigger. |

### 4. Fact Alignment (事实依据) — 135/140

| Criterion | Score | Notes |
|---|---|---|
| Factual claims traceable to fact_id or marked UNKNOWN | 49/50 | All Invariants and Side-effects cite canonical IAM_* IDs verifiable in fact-table.json (IAM_CONSOLE_DIRECT_EIAM_NO_BFF, IAM_USER_PASSWORD_RESET, IAM_USER_USERNAME_FORMAT, IAM_IDENTITY_SOURCE_LDAP, IAM_ROLE_CREATE, IAM_POLICY_DEFINITION_WRITE_SCOPE, IAM_ROLE_ASSIGN, IAM_ROLE_REVOKE, IAM_AUTH_CACHE, IAM_DEPENDENCY_DELETE_BLOCK, IAM_TENANT_BOUNDARY, IAM_403_NO_EXISTENCE_LEAK, IAM_UNAUTHORIZED_401, IAM_VERSION_CONFLICT_409, IAM_PAGINATION_LAST_PAGE). UNKNOWN marking consistent (U1, U4, IAM_* UNKNOWN facts). PARTIAL status propagated where applicable. Minor: inline "fact: IAM_X PARTIAL" tags duplicate the HTML-comment traceability in prose. |
| Inferred claims have required_outcomes rule support and source: inferred | 48/50 | Inferred Outcomes (validation-error-format, conflict-duplicate, concurrent-version-conflict, save-failure, session-expired, unauthorized-401, pagination-last-page-fallback) all carry `<!-- source: inferred -->` with explicit required_outcomes rule citations (web validation-error, api conflict, web concurrent-edit, web network-error, api internal-error, web session-expired, api unauthorized, api pagination). Fact-derived Outcomes (dependency-delete-blocked, cross-tenant-policy-denied, cache-delay-boundary) correctly reclassified as factual, not inferred. Minor: factual-derived Outcomes cite facts that are themselves UNKNOWN/PARTIAL — deriving a behavioral Outcome from an unverified fact is a slightly weak basis, though UNKNOWN propagation is honest. |
| No hallucinated claims without classification | 38/40 | No unclassified hallucinations. The "代码实测" (code-measured) 60s claim is backed by IAM_AUTH_CACHE ("Convergence ceiling 60s — casbin.go:85"). Minor: the "casbin Redis watcher 近实时生效" claim implied in Step 4 Side-effect "近实时生效" is backed by IAM_ROLE_ASSIGN (which mentions Redis watcher) but not explicitly cited there. |

### 5. Surface Fitness (Surface 适配) — 79/90

| Criterion | Score | Notes |
|---|---|---|
| Mandatory derived Outcomes from surface required_outcomes present | 25/35 | Web `validation-error` (Step 1) + `session-expired` (Step 4) present. API `unauthorized` present only in Step 4 — absent on Steps 1, 2, 3, 5, each of which calls an authenticated eiam endpoint. The api surface rule mandates `unauthorized` per authenticated endpoint; 4/5 endpoints uncovered. −10. |
| Dimension descriptions use surface-appropriate language | 29/30 | Web language (用户管理页, 表单, 标红提示, 弹窗, 列表) and api language (401/403 status codes, 请求/响应) used appropriately for the dual surface. file:line refs (surface-inappropriate for both web and api) removed from dimension values. Minor: inline fact-ID tags in prose are not surface-specific language but serve traceability. |
| TUI async timeout Outcomes | 25/25 | N/A (non-TUI surface) — full marks. |

### 6. Internal Consistency (一致性) — 122/140

| Criterion | Score | Notes |
|---|---|---|
| Invariants hold in every Step Contract | 46/50 | Contract-level invariants consistently state 60s convergence with 30s flagged UNKNOWN — internally consistent across all 5 Contracts and aligned with the authoritative IAM_AUTH_CACHE fact. The journey.md invariant still says "≤ 30 秒" but Contracts explicitly flag this as PRD-only/UNKNOWN; within Contract scope no violation. Residual: the cross-artifact value discrepancy (Contract 60s vs journey 30s) is flagged but not reconciled into a single canonical statement across artifacts. |
| Cross-Contract state references are consistent | 40/50 | Step 4→Step 5 references consistent (授权关系 from Step 4 used in Step 5). Step 1 intra-Contract `initial-password-reset-unknown` → `success` state reference consistent. But Step 3 `dependency-delete-blocked` precondition "此引用状态在本 journey 中于授权分配后建立；本 Outcome 为 post-journey 边缘场景，须在授权关系建立后执行" still depends on Step 4's state — linear Step 3 execution cannot satisfy it. The explicit post-journey scoping resolves ambiguity but does not restore linear executability. |
| Outcome Preconditions consistent with preceding Steps' State changes | 36/40 | Step 5 success precondition "该用户当前具备该角色对应的权限" achievable from Step 4's State ✓. Step 5 `cache-delay-boundary` now has a measurable threshold ("删除授权后处于缓存收敛窗口内（≤ 60 秒）") distinguishing it from success. Minor: the boundary between `cache-delay-boundary` (window ≤60s, may pass) and `success` (post-window, denied) is timing-based and relies on a non-deterministic cache window, making the two Outcomes' Preconditions race-condition-adjacent in practice. |

### 7. Anchor Integrity (锚点完整性) — 90/90

Handbook absent (no api-handbook.md / page-map.md in the design directory). Per rubric activation condition, all criteria score full marks (backward-compatible). Anchor frontmatter fields are present but empty per graceful-degradation rule — no penalty.

### 8. Fixture Specification (前置数据声明) — 80/90

| Criterion | Score | Notes |
|---|---|---|
| Entity completeness (veto item) | 30/35 | Inline per-Outcome fixture_specs declare all operated entities (PlatformAdmin, PlatformUser, LdapServer, PolicyDefinition, Role, TenantAdmin, AuthorizationRecord). Veto NOT triggered (inline coverage exists). BUT Step 4 bottom `## Fixture Specification` lists only PlatformAdmin/PlatformUser/Role — omitting TenantAdmin and PolicyDefinition, which are operated in the `cross-tenant-policy-denied` Outcome. A strict reviewer reading only the canonical bottom spec would trigger the entity-completeness veto. −5. |
| Relationship and constraint coverage | 27/30 | Relationships now declared across all multi-entity Outcomes (PlatformUser `created_by`/`assigned_to`/`edited_by`/`same_username`; LdapServer `managed_by`; Role `binds_policy`; PolicyDefinition `bound_to_role`; AuthorizationRecord `links`). field_constraints capture necessary constraints (session valid, username same, version stale, bindable, bound_policy, assignments ≥1, status active, page_position last-page-sole-record). Minor: Step 4 `cross-tenant-policy-denied` TenantAdmin + PolicyDefinition carry no relationship_type between them (acceptable — they are not parent-child in that scenario). |
| Minimum data quantity declarations | 23/25 | min_count values adequate (Step 1 concurrent PlatformAdmin min 2 ✓; Step 5 pagination AuthorizationRecord min 1 ✓; Step 3 dependency Role/PolicyDefinition min 1 ✓). Step 1 bottom PlatformUser min_count 0 is correct for the happy path (no pre-existing user) but a downstream agent reading only the bottom spec would under-provision for conflict/concurrent scenarios — inline specs compensate. Minor under-declaration at bottom-summary level. |

---

## Phase 3 — Blindspot Hunt

- **[blindspot] `initial-password-reset-unknown` Outcome has entirely UNKNOWN Output/State, yielding no executable test contract.** Step 1: `Output: "UNKNOWN — 一次性初始密码展示、首登强制改密行为未获代码支持（fact IAM_USER_PASSWORD_RESET）；不可断言重置是否生效"` and `State: "UNKNOWN — 无法断言密码重置是否写入或生效"`. A downstream QA agent cannot write an assertion for "不可断言". The split (iteration-1 fix) correctly avoids bundling an unimplemented operation into the happy path, but the replacement Outcome documents a gap without providing any testable behavior — it is documentation, not a contract. Either mark the Outcome explicitly as `skip — endpoint unimplemented` with a fallback assertion (e.g., "the admin UI exposes no reset control for the new user" or "a reset request returns not-implemented/404"), or relocate it out of the executable Contract suite. Reasoning audit flagged this independently of dimension scoring.

- **[blindspot] Action-mixing structural misalignment persists under an action-discriminator mask.** Step 3 (step-action `create-role-and-bind-policy`) contains `dependency-delete-blocked` whose Preconditions open with "操作动作为删除角色或策略（区别于本步骤的创建动作）"; Step 4 (step-action `assign-role-to-new-user`) contains `cross-tenant-policy-denied` whose Preconditions open with "操作动作为策略定义增删改（区别于本步骤的角色分配动作）". The action discriminators make the Outcomes logically selectable, but a QA engineer driving the step's nominal action (create / assign) cannot reach these Outcomes without action-switching — the Outcome tests a different action than the step scope promises. This is a structural scope/Outcome misalignment (scope promises X, success criteria test Y) that no single rubric dimension fully captures; Precondition Exclusivity scores the selection-ambiguity aspect, but the scope/Outcome-action mismatch itself is a reasoning flaw. Relocate delete-blocked to a delete-focused step and policy-CRUD-denied to a policy-CRUD-focused step (or a dedicated cross-cutting authorization Contract).

- **[blindspot] API `unauthorized` treated as journey-level rather than per-endpoint.** Only Step 4 carries an `unauthorized-401` Outcome. Steps 1, 2, 3, 5 each invoke an authenticated eiam endpoint (POST /api/user/create, /api/identity_source/save, /api/role/create, /api/role/unassign) with no unauthorized Outcome. The api surface rule mandates unauthorized "for every API Journey involving authenticated endpoints" — the author satisfied the rule once (journey-level) rather than per-endpoint, leaving 4 endpoints' auth-failure behavior unspecified. This is a reasoning flaw (pattern-once instead of per-endpoint) manifesting in Surface Fitness and Completeness, but the underlying misinterpretation of "per authenticated endpoint" is a blindspot.

---

## Summary

| Dimension | Score | Max | Threshold | Pass |
|---|---|---|---|---|
| 完整性 | 130 | 140 | 85 | ✓ |
| 语义纯度 | 154 | 170 | 100 | ✓ |
| 前置条件互斥性 | 128 | 140 | 85 | ✓ |
| 事实依据 | 135 | 140 | 85 | ✓ |
| Surface 适配 | 79 | 90 | 55 | ✓ |
| 一致性 | 122 | 140 | 85 | ✓ |
| 锚点完整性 | 90 | 90 | 55 | ✓ |
| 前置数据声明 | 80 | 90 | 55 | ✓ |
| **Total** | **918** | **1000** | **850** | **✓** |

All per-dimension thresholds met; total 918 ≥ 850 target → **PASS (iteration 2)**.

**Remaining gaps for future hardening** (not blocking pass):
1. API `unauthorized` Outcome on Steps 1, 2, 3, 5 (per-authenticated-endpoint mandate).
2. Step 4 bottom `## Fixture Specification` missing TenantAdmin + PolicyDefinition (entity-completeness veto risk under strict reading).
3. Inline fact-ID / "代码实测" tags in dimension prose — move to HTML comments only.
4. "授权策略缓存重载" mechanism naming in State/Side-effect — raise abstraction to system-level behavior.
5. Step 3 `dependency-delete-blocked` forward reference + action-mismatch — relocate to a delete-step or post-journey Contract.
6. Step 1 `initial-password-reset-unknown` — provide a fallback testable assertion or mark as skip.
