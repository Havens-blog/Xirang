# Journey Eval Report — iteration 2

**Journey**: identity-permission-governance-golden-path
**Rubric**: journey.md (1000 pts, target 850, pass = total ≥ 850 AND every dimension ≥ threshold)
**Surface types declared**: web, api (surface_keys: haven-console, eiam)
**Scorer persona**: Senior QA Engineer
**Iteration**: 2

## Iteration-1 Attack Point Resolution Check

| Iteration-1 attack | Status | Evidence |
|---|---|---|
| Missing mandatory web `session-expired` | **RESOLVED** | Step 4c added with `source: inferred; rule: web/required_outcomes/session-expired` |
| Missing mandatory API `unauthorized` (401) | **RESOLVED** | Step 4d added with `source: inferred; rule: api/required_outcomes/unauthorized; http_status: 401` |
| Raw regex `^[a-z][a-z0-9_-]{2,31}$` in Step 1 | **RESOLVED** | Replaced with natural-language format description |
| HTTP status codes in Expected prose | **RESOLVED** | Status codes moved to Annotations metadata; Expected descriptions are natural language |
| Implementation coupling in Step 3 ("eiam 返回的可绑定策略列表") | **PARTIALLY RESOLVED** | Step 3 now says "系统提供的"; but Invariants still reference "eiam" and "CRUD 端点" |
| No Fact Table | **RESOLVED** | Fact Table added with F1–F12 (verified) + U1–U4 (UNKNOWN) |
| Missing `source: inferred` annotations | **MOSTLY RESOLVED** | 7/9 derived outcomes annotated; 5b-a and 3b-a still missing |
| Happy-path Steps lack Preconditions | **RESOLVED** | All happy-path Outcomes now carry explicit Preconditions |
| E-tag references (E6/E11/E7) instead of in-prose conditions | **RESOLVED** | All edge Preconditions now state full trigger conditions in prose |
| Step 3b widening beyond Step 3 scope | **RESOLVED** | Step 3b now scoped to "Step 3 创建的角色" and its bound policy |
| [blindspot] session-expired mid-workflow gap | **RESOLVED** | Step 4c exercises mid-workflow session timeout on assign-role submit |

**New issues introduced by revision**: None structural. Residual annotation gaps on 5b-a and 3b-a; new executability gap (tenant admin setup missing for Step 4b) surfaced by the added Step 4b context.

## Phase 1 — Reasoning Audit (pre-score anchors)

- **Problem → Solution**: Stated problem = "创建平台用户、配置身份源、创建角色并绑定已有策略、将角色分配给用户、最后撤销授权并验证权限即时收敛". 5 happy-path steps map 1:1. No easier-problem substitution.
- **Solution → Evidence**: Fact Table now grounds verified claims (F1–F12) and marks unknowns (U1–U4). "Phase 0 核查" markers consistently applied to U1–U4. Acceptable.
- **Evidence → Success Criteria**: Expected Results test user-observable behavior (列表刷新, 弹窗, 跳转登录页, 拒绝请求). No internal-state coupling. OK.
- **Self-contradiction check**: Invariant "权限判定逐请求生效" with "≤ 30 秒" caveat vs Step 5b cache-window — consistent (both carry the same UNKNOWN caveat). Step 4c (session-expired, web) and Step 4d (unauthorized, api) test related but distinct surface manifestations. No contradiction.
- **SC cluster check**: N/A (journey doc, no SC/InScope entries).

Pre-score anchors channeled into Fact Alignment + Surface Fitness below.

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 160/175

- Metadata (42/45): name `identity-permission-governance-golden-path` kebab-case ✓; `risk_level: High` justified by content (authorization mutations, irreversible revoke/delete). Frontmatter complete with surface_types, surface_keys, sources, generated date. Risk justification in HTML comment rather than structured field — minor.
- Steps complete (65/70): 5 happy + 10 edge steps (1b, 1c, 1d, 2b, 3b, 4b, 4c, 4d, 5b, 5c), each with User Action + Outcome (Precondition + Expected + Annotations). Coherent ordered sequence. Single-outcome-per-step structure satisfies "at least one Outcome" but yields no multi-outcome branching within Steps.
- Outcomes cover happy + required derived (53/60): All mandatory derived Outcomes present: validation-error (1b), session-expired (4c), unauthorized 401 (4d). Additional boundary Outcomes: conflict/concurrent-edit (1c/1d), pagination (5c), network-error (2b). Several optional boundary Outcomes absent (navigation-guard, responsive-layout for web; not-found, rate-limit for api) — acceptable as non-mandatory.

### 2. Semantic Purity (语义纯度) — 144/175

- Outcome descriptions natural language (58/70): Regex removed from Step 1 ✓. HTTP status codes now in Annotations metadata, not Expected prose ✓. Expected descriptions are user-observable. Residual: "后端拒绝请求" and "直接调用接口" appear in Expected/Action prose (Step 4b, 4d, 5-a) — implementation-level terms.
- Preconditions declarative (48/55): All Preconditions are state descriptions. Minor: Step 1c "后端返回校验/写入错误" describes backend action rather than pure state; Step 2b "后端不可用" is borderline procedural.
- No implementation coupling in Steps (38/50): Step descriptions mostly user-level. Step 4d User Action "直接调用授权分配接口" references API endpoint detail. Setup and Invariants reference "eiam", "直调 eiam", "CRUD 端点" — metadata sections, not Step descriptions, but still implementation-coupled.

### 3. Precondition Exclusivity (前置条件互斥性) — 115/130

- Preconditions distinct across Outcomes (45/50): Each Step has exactly one Outcome — trivially distinct, no intra-step overlap. Structure avoids ambiguity by construction.
- Sufficient to uniquely select (38/45): Single-outcome-per-step → trivially unique. Deduction: Step 1c and Step 2b OR multiple distinct trigger conditions in one Precondition ("提交的用户名已存在，或后端返回校验/写入错误"), reducing selection precision.
- No missing Preconditions for error/boundary (32/35): All edge Outcomes carry full in-prose Preconditions. No E-tag shorthand remnants. Minor: Step 1c and 2b OR-ed conditions conflate distinct triggers into one Precondition.

### 4. Fact Alignment (事实依据) — 111/130

- Factual claims traceable to fact_id or UNKNOWN (43/50): Fact Table with F1–F12 (verified, sourced to PRD) + U1–U4 (UNKNOWN, with reasoning). Outcomes reference facts via `fact: F1` etc. UNKNOWN claims marked with "UNKNOWN — see U1". Minor: Setup claim "平台管理员豁免零租户限制" not in Fact Table — a system-behavior assertion without traceability.
- Inferred claims have required_outcomes rule support + source: inferred (38/45): 7/9 derived Outcomes properly annotated:
  - 1b-a ✓, 1c-a ✓, 1d-a ✓, 2b-a ✓, 4c-a ✓, 4d-a ✓, 5c-a ✓
  - **5b-a ✗**: labeled "(derived)" but no `source: inferred`, no rule citation. Cache-window behavior ("窗口内可能仍放行") goes beyond F7's scope.
  - **3b-a ✗**: labeled "(derived)" but content is factual (restates F8 exactly). No `source: inferred` or rule. Classification label inconsistent with factual nature.
- No hallucinated unclassified claims (30/35): Most claims trace to F-facts or U-unknowns. 5b-a's cache-window behavior is grounded in U2+F7 but improperly annotated (not a true hallucination). Step 2b Precondition mentions "字段校验失败" (validation error) but cited rules are network-error/internal-error — rule-to-trigger mismatch leaves the validation-error aspect unclassified.

### 5. Surface Fitness (Surface 适配) — 105/130

- Mandatory derived Outcomes present (45/50): web `validation-error` (1b) ✓ + `session-expired` (4c) ✓; api `unauthorized` (4d, 401) ✓. All mandatory outcomes present. session-expired tested on one mutating step only (Step 4) — satisfies mandatory presence.
- Test strategy proportions (30/45): Target 50/50 Contract/Journey for web/api. Actual: 5 journey-level / 9 contract-level Outcomes ≈ 36/64. Skewed contract-heavy. Happy-path steps are journey-level; all edge cases are contract-level — no journey-level edge exploration.
- Surface-specific environment realistic (30/35): @web-e2e and @api-functional tags correctly applied. Step 4c web-only (session-expired is web concern) ✓; Step 4d api-only (unauthorized 401 is api concern) ✓. Browser interaction (前端拦截, 弹窗, 列表刷新) + HTTP status codes in annotations. Realistic assumptions.

### 6. Internal Consistency (一致性) — 120/130

- Invariants hold in every Step (44/50): eiam唯一后端 ✓; 逐请求判定 with ≤30s caveat ✓; ≤2s refresh ✓; 错误不白屏 ✓; 前端菜单裁剪非安全边界 ✓ (tested by Step 4b). All hold. Minor: "eiam 是唯一后端" not directly testable within Journey Steps.
- Cross-Step references consistent (43/45): Step 4 → "Step 1 创建的新用户" ✓; Step 4 → "Step 3 创建的新角色" ✓; Step 5 → "Step 4 中建立的该用户授权关系" ✓; Step 3b → "Step 3 创建的角色" + "Step 4 已建立授权关系" ✓. All unambiguous. Minor: Step 3b references "绑定到该角色的策略" — policy existed before Step 3 (selected from list), reference is traceable but slightly indirect.
- Risk level consistent (33/35): High — authorization mutations, irreversible revoke, tenant boundary enforcement, dependency deletion blocks. Fully consistent with content.

### 7. Workflow Coverage (工作流覆盖度) — 107/130

- Golden Path existence (47/50): Steps 1→2→3→4→5 = 5 contiguous steps (create user → configure LDAP → create role+bind policy → assign role → revoke+verify convergence). Domain-level operations, not bare HTTP calls. Maps to PRD Story 2 + Story 4. Veto not triggered. Minor: Step 2 (LDAP config) is tangential to the core user→role→permission chain but is in the stated overview.
- Multi-step coverage depth (35/45): Covers create→assign→revoke lifecycle, concurrent edit, dependency deletion, session expiry, unauthorized, pagination, validation error, network error, tenant boundary. Multi-entity (user, role, policy, auth, identity source). State transitions (assign→revoke). Good depth. Missing: edit-user lifecycle, delete-user happy path, role-unbind workflow, strategy-definition lifecycle.
- Workflow completeness against PRD scope (25/35): Covers Stories 2/3/4. Story 3 tenant-boundary via 4b. Gap: Step 5 requires the created user to make an authenticated request ("该用户以原权限发起下一次请求") but no preceding Step establishes the user's login/session — the user is created in Step 1 with initial password + forced change, but never logs in. Also missing: edit-user, delete-user happy steps, role-unbind workflow.

## Phase 3 — Blindspot Hunt

- **[blindspot] Setup does not establish tenant admin session for Step 4b**: The Setup section establishes only a platform admin session ("平台管理员账号已存在且已启用 MFA，已登录控制台"), but Step 4b's Precondition requires "当前操作者为租户管理员（非平台管理员）". No Step or Setup instruction describes how/when the tenant admin session is established. A downstream agent executing this Journey cannot set up Step 4b without this context. This spans Internal Consistency (Setup-to-Step) and Workflow Coverage (role-switch workflow), but the Setup-to-Step executability gap is not fully captured by either dimension's criteria (Internal Consistency checks Step→Step references; Workflow Coverage checks PRD scope coverage). Quote: "平台管理员账号已存在且已启用 MFA，已登录控制台" vs Step 4b "当前操作者为租户管理员（非平台管理员）". Must improve: add Setup instruction or a Step that establishes the tenant admin session before Step 4b.

## Result

SCORE: 862/1000 — **PASS** (target 850; all dimensions above threshold).

Improvements from iteration-1: +127 pts (735 → 862). All below-threshold dimensions recovered: Fact Alignment 72→111, Surface Fitness 75→105. All iteration-1 mandatory-outcome gaps closed (session-expired, unauthorized 401). Fact Table and source: inferred annotations added. Regex and HTTP status codes purged from Expected prose.
