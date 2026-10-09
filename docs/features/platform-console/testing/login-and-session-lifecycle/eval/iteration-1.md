# Eval Report — login-and-session-lifecycle (Iteration 1)

**Scorer**: Senior QA Engineer (adversarial)
**Rubric**: journey.md (7 dims / 1000, target 850, pass = total ≥ 850 AND every dim ≥ threshold)
**Surface types declared**: `["web", "api"]`
**Date**: 2026-10-08

---

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem→Solution fit**: Journey scopes itself to login/SSO/logout lifecycle (Story 1 + Story 6). Solution (shared Redis session, single login domain, 401-on-next-request convergence) directly addresses the stated dual-side session problem. Fit is sound.
2. **Solution→Evidence**: Most behavioral claims (MFA 5-strike lockout, zero-tenant restricted workbench, dual-side logout) are backed by E#/Story references to PRD. BUT one load-bearing claim — `JWT claims 含 uid/username/tenant_id` in Step 2 — is asserted as fact while PRD DF002 explicitly flags tenant_id write semantics as `待 Phase 0 核查` (unverified). This is an unclassified factual assertion.
3. **Evidence→Success Criteria**: The "Expected Result" blobs conflate user-observable behavior with implementation internals (Redis, JWT claims, nginx headers), so success criteria test a mix of behavior and internals — weakening what they actually verify.
4. **Self-contradiction**: Frontmatter declares `golden_path: false`, yet the document body contains a `## Happy Path` with 4 contiguous steps (Step 1–4) that map directly onto Story 1 (login → workbench → SSO jump) and Story 6 (dual logout). The body IS a golden path; the metadata denies it. Downstream tooling reading `golden_path: false` would treat this journey as having no golden path — fatal ambiguity.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 117/175 (threshold 105) ✓ threshold

| Sub | Score | Notes |
|-----|-------|-------|
| Metadata (kebab name, risk_level justified) | 42/45 | Name kebab ✓; risk Medium justified via comment. But `golden_path: false` is a questionable metadata declaration contradicted by body. |
| Steps complete (name/action/outcome) | 50/70 | Every Step has User Action + Expected Result. BUT no Step models multiple alternative Outcomes with Preconditions — alternatives are split into separate Step* letters instead of Outcome variants within a Step. |
| Happy + required derived Outcomes | 25/60 | Web mandatory `validation-error` (invalid form input) absent; `session-expired` (timeout during active workflow) absent — only logout-triggered invalidation covered. API mandatory `unauthorized` (authenticated endpoint w/o credentials) absent. `network-error` (Step 2c) present as additional. |

### 2. Semantic Purity (语义纯度) — 130/175 (threshold 105) ✓ threshold

| Sub | Score | Notes |
|-----|-------|-------|
| Outcomes natural language, no code/regex | 50/70 | No regex/CSS/assertion calls. But outcomes embed technical tokens: `JWT claims 含 uid/username/tenant_id`, `CSP/X-Frame-Options/HSTS`, `401`, `403`. Implementation-detail leakage in user-observable result descriptions. |
| Preconditions declarative, not procedural | 50/55 | Edge-case preconditions are declarative state (`未登录且浏览器不支持 WebAuthn`, `租户数 = 0`). Clean. |
| No implementation coupling in Steps | 30/50 | Step 1 Expected Result: `CSP/X-Frame-Options/HSTS 由 nginx 注入`. Step 2: `写回 eiam 共享 Redis session（JWT claims 含 uid/username/tenant_id）`. Direct coupling to Redis/JWT/nginx internals in step-level descriptions. |

### 3. Precondition Exclusivity (前置条件互斥性) — 90/130 (threshold 80) ✓ threshold

| Sub | Score | Notes |
|-----|-------|-------|
| Preconditions distinct across Outcomes within Step | 35/50 | Each Step carries a single Outcome; alternatives are modeled as separate Steps (2b/2c/2d/2e), not as co-located Outcomes. Vacuously non-overlapping but the dimension's intent (distinguishability within a step) is unexercised. |
| Preconditions sufficient to uniquely select Outcome | 30/45 | Single-outcome-per-step trivially selects, but the structure provides no evidence the author reasoned about mutual exclusivity between competing branches. |
| No missing Preconditions for error/boundary Outcomes | 25/35 | Edge cases DO state preconditions (good). But happy-path Steps 1–4 carry NO preconditions at all, so the boundary between "when does happy path apply vs edge" is implicit. |

### 4. Fact Alignment (事实依据) — 78/130 (threshold 80) ✗ BELOW THRESHOLD

| Sub | Score | Notes |
|-----|-------|-------|
| Factual claims traceable to fact_id or UNKNOWN | 38/50 | E#/Story references (E8, E2/E9, E12, E3, E5, Story 6) act as informal traceability to PRD edge table. But no formal `fact_id` mapping and the `tenant_id in JWT claims` claim is NOT marked UNKNOWN despite PRD DF002 flagging it pending Phase 0. |
| Inferred claims have required_outcomes rule + `source: inferred` | 12/45 | Zero `source: inferred` annotations anywhere. No citation of surface `required_outcomes` rules (web: validation-error/session-expired; api: unauthorized). The inferred-outcome framework is entirely unengaged. |
| No hallucinated unclassified claims | 28/35 | Most claims grounded in PRD. One genuine unclassified assertion: `JWT claims 含 uid/username/tenant_id` stated as fact while PRD marks tenant_id write as unverified. |

### 5. Surface Fitness (Surface 适配) — 55/130 (threshold 80) ✗ BELOW THRESHOLD

| Sub | Score | Notes |
|-----|-------|-------|
| Mandatory derived Outcomes present | 15/50 | Web `validation-error`: absent. Web `session-expired`: only logout-triggered invalidation, NOT timeout-during-workflow. API `unauthorized`: absent (no authenticated-endpoint-without-credentials outcome). Mandatory outcomes largely missing — not zeroed because partial session-invalidation coverage exists. |
| Test strategy proportions (Web/API balanced 50/50) | 15/45 | No Contract/Journey split indicated. No `@web-e2e` / `@api-functional` test tags. Mixed web+api journey gives no signal on test-level proportions. |
| Surface-specific environment realistic | 25/35 | Web assumptions (browser, WebAuthn, CSP) realistic ✓. API assumptions (401, shared session) realistic ✓. Blended surface handling is coherent but no explicit per-surface execution model. |

### 6. Internal Consistency (一致性) — 100/130 (threshold 80) ✓ threshold

| Sub | Score | Notes |
|-----|-------|-------|
| Invariants hold in every Step | 45/50 | All 5 declared invariants hold across Steps (shared-session truth source, no username leak, no white-screen, same-domain, admin exempt). |
| Cross-Step references consistent | 25/45 | Step 4 references Step 2's session; Step 4b references Step 3's e-cam-web login — consistent. BUT frontmatter `golden_path: false` contradicts the body's Happy Path (a declared golden path that the metadata denies) — a structural contradiction that breaks downstream consistency. |
| Risk level consistent with content | 30/35 | Medium defensible (logout reversible, no irreversible data loss). Auth/security angle could argue High, but Medium is within tolerance. |

### 7. Workflow Coverage (工作流覆盖度) — 80/130 (threshold 80) ✓ threshold (barely)

| Sub | Score | Notes |
|-----|-------|-------|
| Golden Path existence (veto item) | 25/50 | Veto NOT triggered: body contains a semantically valid 4-step Happy Path mapping to Story 1 + Story 6. BUT `golden_path: false` frontmatter explicitly denies golden-path designation — heavy deduction for fatal ambiguity (downstream consumers cannot rely on this journey as the golden path). |
| Multi-step coverage depth | 30/45 | Edge cases cover tenant 0/1/many state transitions, MFA lockout, reverse logout, WebAuthn fallback. Cross-entity console↔e-cam-web session interaction covered. No entity-lifecycle CRUD (scope-appropriate for a login journey). |
| Workflow completeness vs PRD scope | 25/35 | Story 1 + Story 6 covered. PRD Business Flow #8 `会话过期 → 跳转控制台登录页` (session timeout, distinct from logout) has no dedicated step. Stories 2–5 correctly out of scope for this journey. |

---

## Phase 3 — Blindspot Hunt

- **[blindspot]**: `golden_path: false` in frontmatter vs `## Happy Path` body — "golden_path: false" / "### Step 1 ... ### Step 4" — Reasoning audit flagged this independently of dimension scoring; the declaration makes the journey unusable as a canonical golden-path source for the feature while its content is exactly that. Must set `golden_path: true` or rename Happy Path and drop the primary-user-story framing.
- **[blindspot]**: No `## Fact Table` / fact_id registry anywhere in the journey — "sources: docs/features/platform-console/prd/prd-spec.md" — the Fact Alignment framework presupposes a Fact Table with traceable fact_ids; its absence makes every E# reference informal and prevents a downstream agent from verifying any specific behavioral assertion back to a canonical fact.
- **[blindspot]**: Test-type tag discipline absent — surface rules mandate `@web-e2e` for web and `@api-functional` (NOT `@e2e`) for API; this dual-surface journey declares neither, so a test generator cannot disambiguate which steps are browser-driven vs HTTP-client-driven.

---

## Summary

**SCORE: 550/1000**
**PASS: NO** — total 550 < 850; Fact Alignment (78 < 80) and Surface Fitness (55 < 80) below per-dimension thresholds.

```
SCORE: 550/1000
DIMENSIONS:
  完整性: 117/175
  语义纯度: 130/175
  前置条件互斥性: 90/130
  事实依据: 78/130
  Surface 适配: 55/130
  一致性: 100/130
  工作流覆盖度: 80/130
ATTACKS:
1. [事实依据]: tenant_id-in-JWT 断言为事实未标 UNKNOWN — "JWT claims 含 uid/username/tenant_id" — PRD DF002 标注 tenant_id 写入待 Phase 0 核查；须标 UNKNOWN 或移除该断言
2. [Surface 适配]: Web/API 强制派生 Outcome 缺失 — "validation-error" / "session-expired" / "unauthorized" 均无对应 Outcome — 须补表单校验失败、会话超时(非登出)、无凭据访问鉴权端点三类 Outcome
3. [一致性]: frontmatter golden_path: false 与 body Happy Path 矛盾 — "golden_path: false" vs "## Happy Path ### Step 1...Step 4" — 须将 golden_path 改 true 或重命名 Happy Path 并放弃主用户故事框架
4. [工作流覆盖度]: Golden Path 被元数据否认 — "golden_path: false" — 下游无法据此判定黄金路径；须修正声明使内容与元数据一致
5. [语义纯度]: Expected Result 耦合实现细节 — "CSP/X-Frame-Options/HSTS 由 nginx 注入" / "写回 eiam 共享 Redis session（JWT claims 含 uid/username/tenant_id）" — 须改为用户可观测行为，将 Redis/JWT/nginx 细节移至 Fact Table 或实现注解
6. [事实依据]: 推理声明框架完全未启用 — 无任何 `source: inferred` 标注或 required_outcomes 规则引用 — 须为派生边界 Outcome 标注来源与触发规则
7. [完整性]: 单 Outcome/Step 结构未建模分支 — 各 Step 仅单一 Expected Result，分支拆为 Step 2b/2c/2d/2e — 须在同一 Step 内以 Preconditions 区分多 Outcome
8. [Surface 适配]: 测试类型标签缺失 — 无 @web-e2e / @api-functional 标注 — 双面 Journey 须为每 Step 标注对应测试类型以供生成器分流
9. [blindspot]: 无 Fact Table / fact_id 登记 — "sources: ...prd-spec.md" 仅列文件级来源 — 须建 Fact Table 逐条登记 fact_id 供断言追溯
10. [blindspot]: golden_path: false 使本 Journey 既含主用户故事黄金路径又否认之 — "golden_path: false" + "Step 1...Step 4 覆盖 Story 1/6" — Reasoning audit 独立标记；须决断本 Journey 是否为黄金路径并统一声明
```
