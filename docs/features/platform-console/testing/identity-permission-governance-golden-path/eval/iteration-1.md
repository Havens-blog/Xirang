# Journey Eval Report — iteration 1

**Journey**: identity-permission-governance-golden-path
**Rubric**: journey.md (1000 pts, target 850, pass = total ≥ 850 AND every dimension ≥ threshold)
**Surface types declared**: web, api (surface_keys: haven-console, eiam)
**Scorer persona**: Senior QA Engineer
**Iteration**: 1

## Phase 1 — Reasoning Audit (pre-score anchors)

- **Problem → Solution**: Stated problem = "端到端治理：创建平台用户、配置身份源、创建角色并绑定已有策略、将角色分配给用户、最后撤销授权并验证权限即时收敛". The 5 happy-path steps map 1:1 to this problem statement. Solution addresses the stated problem — no easier-problem substitution.
- **Solution → Evidence**: Claims like "≤ 2 秒内刷新", "≤ 30 秒" cache delay, "409 版本冲突" are consistently marked "Phase 0 核查" — i.e. explicitly unverified. Acceptable as unknowns, but no Fact Table / fact_id traceability exists for the verified claims.
- **Evidence → Success Criteria**: Expected Results test user-observable behavior (列表刷新、提示、403 拒绝) — reasonable proxies, not internal-state coupling. OK.
- **Self-contradiction check**: Invariants ("权限判定逐请求生效") vs Step 5 / 5b ("如 eiam 存在授权缓存，生效延迟 ≤ 30 秒") — not contradictory because invariant itself carries the same "≤ 30 秒（Phase 0 核查）" caveat. Internally consistent.
- **SC cluster check**: No SC/InScope entries (journey doc, not contract). N/A.

Pre-score anchors channeled into Fact Alignment + Surface Fitness below.

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 138/175

- Metadata (42/45): name `identity-permission-governance-golden-path` kebab-case ✓; `risk_level: High` justified by content (authorization mutations, irreversible revoke/delete). Frontmatter well-populated.
- Steps complete (58/70): 5 happy + 8 edge steps, each with User Action + Expected Result, coherent ordered sequence. Deduction: happy-path Steps lack explicit Outcome-with-Precondition objects; Expected Result is a prose blob, not enumerated Outcomes.
- Outcomes cover happy + required derived (38/60): validation-error present (Step 1b); already-exists/conflict present (1c/1d); pagination present (5c); unauthorized-ish present (4b, but 403 not 401). **Missing mandatory web `session-expired`**; **missing mandatory API `unauthorized` (401)** for authenticated endpoints.

### 2. Semantic Purity (语义纯度) — 133/175

- Outcome descriptions natural language (50/70): Mostly natural, but raw regex `^[a-z][a-z0-9_-]{2,31}$` appears in Step 1 User Action and Step 1b Precondition — regex in what should be a user-observable description. HTTP status codes (403/409/401) leak API detail into Outcome prose.
- Preconditions declarative (48/55): Preconditions are state descriptions, not setup procedures. ✓.
- No implementation coupling in Steps (35/50): Steps reference "eiam 返回的可绑定策略列表", "eiam 侧数据同步变更", "直调 eiam" — backend/API internals leak into Step descriptions.

### 3. Precondition Exclusivity (前置条件互斥性) — 108/130

- Preconditions distinct across Outcomes (42/50): Edge cases are separate steps (1b/1c/1d/2b/3b/4b/5b/5c), each one Outcome — no intra-step overlap.
- Sufficient to uniquely select (38/45): Single-Outcome steps trivially unique; happy-path Steps have no Preconditions at all (rely on Setup section), so selection logic is underspecified for the happy path.
- No missing Preconditions for error/boundary (28/35): Edge cases carry Preconditions. But Step 1d's "（E6）", Step 3b's "（E11）", Step 5c's "（E7）" tag references instead of stating the trigger condition fully in-prose.

### 4. Fact Alignment (事实依据) — 72/130 (BELOW THRESHOLD 80)

- Factual claims traceable to fact_id or UNKNOWN (32/50): No Fact Table exists. Claims reference PRD edge IDs (E6/E7/E11/E12/E10/E1) — a form of traceability but not `fact_id`. "Phase 0 核查" markers correctly flag unknowns but the rubric expects `UNKNOWN` marking.
- Inferred claims have required_outcomes rule support + `source: inferred` (15/45): Derived boundary Outcomes (1b validation-error, 4b unauthorized, 5c pagination, 1c conflict) are clearly surface-rule-derived but carry **no `source: inferred` annotation** and cite **no `required_outcomes` rule**.
- No hallucinated unclassified claims (25/35): Most claims traceable to PRD E-scenarios or marked Phase 0 核查. Minor unclassified: Step 2 "登录页相应渲染 LDAP 入口" not in PRD; Step 5 "≤ 30 秒" cache window origin unstated beyond Phase 0 核查.

### 5. Surface Fitness (Surface 适配) — 75/130 (BELOW THRESHOLD 80)

- Mandatory derived Outcomes present (22/50): web requires `validation-error` (present, 1b) + `session-expired` (**absent**); api requires `unauthorized` (401) per authenticated endpoint (**absent** — 4b is 403 forbidden, not 401 unauthenticated). Two mandatory outcomes missing.
- Test strategy proportions (25/45): web/api both 50/50 Contract/Journey. Journey has 13 Outcomes all at Journey level; no Contract-level split or indication.
- Surface-specific environment realistic (28/35): web browser interaction (前端即时拦截, 弹窗, 列表刷新) + api status codes (403/409) mixed acceptably.

### 6. Internal Consistency (一致性) — 114/130

- Invariants hold in every Step (42/50): eiam唯一后端, 逐请求判定, ≤30s revoke delay, ≤2s list refresh, 错误不白屏 — all hold across Steps. Step 5b's cache-window-then-403 is consistent with the "≤ 30 秒" invariant caveat.
- Cross-Step references consistent (40/45): Step 4 → "Step 1 创建的新用户" ✓; Step 5 → "该用户" ✓; Step 3 → Step 3b role/strategy consistent. Minor: Step 3b "仍被引用的角色或策略" widens beyond Step 3's created role.
- Risk level consistent (32/35): High — authorization revoke + irreversible deletes + tenant boundary enforcement. Consistent.

### 7. Workflow Coverage (工作流覆盖度) — 95/130

- Golden Path existence (45/50): 5 contiguous steps (create user → configure LDAP → create role+bind policy → assign role → revoke+verify convergence) semantically maps to PRD Story 2 + Story 4. Domain-level operations, not bare HTTP calls. Veto not triggered.
- Multi-step coverage depth (28/45): Covers create→assign→revoke lifecycle + edge cases. Single entity family (user/role/auth); no org/tenant CRUD, no strategy-definition lifecycle (explicitly excluded for tenant admin). Moderate depth.
- Workflow completeness against PRD scope (22/35): Journey scoped to identity+permission governance (Stories 2/3/4). Story 3 tenant-boundary only partially (4b). Stories 1/5/6 out of scope (separate journeys) — acceptable — but within-scope gaps: no edit-user lifecycle step, no delete-user happy step, no role-unbind workflow.

## Phase 3 — Blindspot Hunt

- **[blindspot] Reasoning audit flagged independently**: The Journey's "Setup" preconditions ("平台管理员账号已存在且已启用 MFA，已登录控制台") silently assume login + MFA + tenant selection already completed — but the Journey's own surface mix (web) mandates a `session-expired` derived Outcome, and the PRD Business Flow #8 explicitly handles "会话过期 → 跳转控制台登录页". A 5-step mutating workflow that assumes a live session yet never tests session expiry mid-workflow is a real test-plan gap no single dimension fully captures (it surfaces partially in Completeness and Surface Fitness but the reasoning flaw — "assumes the precondition it never stress-tests" — is structural). Quote: "平台管理员账号已存在且已启用 MFA，已登录控制台". Must improve: add a `session-expired` edge Step exercising mid-workflow session timeout on at least one mutating action (e.g., Step 4 assign-role submit after session expiry).

## Result

SCORE: 735/1000 — **FAIL** (target 850; dimensions 事实依据 72<80 and Surface 适配 75<80 below threshold).
