# PRD Evaluation Report — Iteration 1 (PM / Adversary)

**Evaluator Persona**: Senior Product Manager (adversarial)
**Rubric**: Mode A (Feature WITH UI), 7 dimensions, 1000 points
**Date**: 2026-10-09
**Documents evaluated**:
- `prd-spec.md`
- `prd-user-stories.md`
- `prd-ui-functions.md`

**Business rules loaded for consistency check**:
- `auth.md` (BIZ-auth-001 ~ 010)
- `identity.md` (BIZ-identity-001 ~ 005)
- `resilience.md` (BIZ-resilience-001 ~ 002)

---

## Phase 1: Reasoning Audit

### Problem → Solution trace

The stated problem is threefold: (1) WatchAlert source lives outside the platform workspace and is not governed by unified orchestration/quality gates; (2) startup requires manual two-step process; (3) login is fragmented — WatchAlert maintains its own local accounts, decoupled from the platform's eiam unified identity.

The proposed solution addresses all three: source into workspace, one-command startup, unify login to eiam and remove local login. The solution does address the stated problem. No "easier problem" substitution detected.

### Solution → Evidence trace

Evidence is largely self-referential. The PRD does not cite baseline metrics from the current state (e.g., current startup time measured, current auth failure rate, number of users affected by dual-account overhead). Goals like "启动耗时 ≤ 编译 + 数十秒" reference the current state implicitly ("手动两步") but never quantify the baseline. This weakens the evidentiary chain.

### Evidence → Success Criteria trace

| Goal | Success criterion | Verdict |
|------|------------------|---------|
| 账号统一 | "100% 走 eiam" | Testable (grep for local login handler, verify no token issuance) |
| 一键启动 | "≤ 编译 + 数十秒" | "数十秒" is a range, not a threshold — partially testable |
| 质量门禁覆盖 | "可一键触发" | Tests capability, not outcome — "can trigger" ≠ "passes consistently" |
| 本地登录后门清零 | "返回 401/404" | Ambiguous: two codes listed, which one is the contract? |
| 上游贡献不中断 | "仍能向开源上游提交通用修复" | No metric — untestable as stated |

Three of five goals have verifiability gaps.

### Self-contradiction check

1. **"eiam 只负责认证" vs. BIZ-auth-001**: BIZ-auth-001 states "eiam 对所有 CRUD 端点做服务端权限校验". The PRD explicitly scopes authorization OUT of eiam ("授权留在 WatchAlert 内部"). While one can argue BIZ-auth-001 applies to platform-console endpoints specifically, the PRD does not clarify whether WatchAlert's CRUD endpoints are exempt from eiam's server-side authorization enforcement or whether WatchAlert itself enforces server-side auth. This ambiguity is a latent contradiction.

2. **"无含糊措辞" self-certification vs. actual content**: The Quality Checklist item "无含糊措辞（阈值/清单/口径均量化）" is checked `[x]`, yet the Goals table contains "数十秒" (a range, not a number) and the grace period ("宽限窗口") is never quantified anywhere in the document. The document self-certifies a quality it does not meet.

3. **401/404 for removed login endpoint vs. BIZ-auth-003**: BIZ-auth-003 defines 401 = not authenticated, 403 = authenticated but no permission. A removed endpoint should return 404 (not found), not 401 (not authenticated). Offering "401/404" as either-or creates semantic ambiguity and risks conflating "endpoint removed" with "not authenticated".

### Pre-score anchors

- A1: Grace period duration is never defined — appears in flow description, UI function states, and user story ACs, but no numeric value exists anywhere.
- A2: eiam-to-WatchAlert role mapping mechanism is completely absent — the PRD says WatchAlert internal roles govern post-login access but never explains how an eiam-authenticated user acquires a WatchAlert role.
- A3: Single sign-out (BIZ-auth-005) is never mentioned — logout/session destruction behavior is absent.
- A4: Session expiry mid-use (BIZ-auth-008) is never addressed — the flow diagram only checks auth at initial access.

---

## Phase 2: Rubric Scoring

### Dimension 1: Background & Goals (100 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Background has three elements (Reason/Target/Users) | 28/30 | All three present and specific. Users table defines 3 roles with clear responsibilities. Minor: "平台开发者" role description ("通过平台统一命令操作") is thin — doesn't describe what they need from the feature beyond startup. |
| Goals are quantified | 15/30 | "100% 走 eiam" is quantified. "≤ 编译 + 数十秒" — "数十秒" is a vague range, not a threshold. "可一键触发" tests capability not outcome. "401/404" is specific but ambiguous (two codes). "仍能向开源上游提交通用修复" has zero quantification. Two of five goals are adequately quantified. |
| Background and goals are logically consistent | 32/40 | Goals follow from the stated problem (dual accounts → eiam unification, manual startup → one-command, no quality gate → unified gate). But "上游贡献不中断" is a constraint masquerading as a goal — it doesn't measure feature success, it measures ongoing fork hygiene. This inflates the perceived goal coverage. |

**Subtotal: 75/100**

---

### Dimension 2: Flow Diagrams (150 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Mermaid diagram exists | 42/50 | Two Mermaid diagrams present (login/degradation flow + developer startup flow). However, the first diagram contains a stray character: `Deg{C 认证服务可用?}` — the "C" appears to be a typo or artifact, reducing diagram clarity. |
| Main path complete (start → end) | 32/50 | Login flow covers Start → CheckAuth → EiamLogin → Session → Dash (complete). Developer flow covers D → Build → Run → Access (complete). But 主流程 3 (迁移切流) — described in text as a 4-step process — has **no Mermaid diagram at all**. This is a significant coverage gap for a core business flow. |
| Decision points + error branches covered | 28/50 | Login flow has diamond nodes (CheckAuth, EiamOK, Deg, Expire) and one error branch (Err for auth failure). Developer flow has build-failure branch. But missing: session expiry branch (BIZ-auth-008 — session can expire mid-use), eiam callback error/forgery branch, migration rollback branch. The error node `Err[返回登录错误提示]` is generic — no differentiation between network error, callback invalid, or user cancelled. |

**Subtotal: 102/150**

---

### Dimension 3: Functional Specs (200 pts) — evaluates prd-ui-functions.md

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Placement & Interaction completeness | 48/70 | All 3 UI Functions have Placement (Mode/Target Page/Position). Interaction flows cover the happy path for each. But: UF-1 interaction flow step 4 covers auth failure but not network timeout during eiam redirect, not eiam callback tampering. UF-2 step 3 says "后端返回 503" for writes during degradation but doesn't cover what the user sees if the backend itself is degraded (not just eiam). UF-3 doesn't cover the transition from read-only back to authenticated — "自动回正" is stated but no interaction steps describe how the user experiences this (auto-refresh? redirect? prompt?). |
| Data Requirements & States clarity | 42/70 | Field tables and state tables exist for all 3 UFs. However: (1) **Grace period duration is never defined** — UF-2 shows "宽限剩余时长 | 数值 | 后端状态 | 横幅展示「剩余 X 小时」" but X is never specified anywhere in any document. This is a critical unimplemented parameter. (2) UF-1 "回跳地址 | 字符串 | 当前页 URL" — no mention of open-redirect validation on this field. (3) UF-3 "鉴权服务状态 | 布尔 | 后端探针" — probe mechanism and polling frequency undefined. (4) State tables have only 2-3 states each — no transitional/error states between them. |
| Validation Rules explicit | 32/60 | One validation rule per UF. UF-1: "不得保留任何本地用户名/密码输入框" — actionable. UF-2: "横幅须明确告知「哪些操作禁止」" — "明确告知" is subjective; what constitutes "明确"? Should list specific operations. UF-3: "只读视图不得出现任何登录入口" — actionable. Missing: no validation rule for redirect URL safety (open redirect), no validation for grace period display edge cases (negative values, expiry during display), no validation for concurrent state transitions (eiam recovers while user is mid-write). |

**Subtotal: 122/200**

---

### Dimension 4: User Stories (200 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Coverage: one story per target user | 45/50 | All 3 background roles covered: 运维/值班工程师 (Stories 1, 2), 告警平台管理员 (Story 3), 平台开发者 (Story 4). Minor: no story for the "平台管理员" role defined in BIZ-auth-002 (platform admin who governs cross-tenant) — though this may be intentionally out of scope. |
| Format correct (As a / I want / So that) | 43/50 | All 4 stories follow As a / I want / So that format. Actions are mostly concrete. Minor: Story 3's "I want to 确认...已彻底移除" is a verification action, not a feature desire — it reads more like a test case than a user need. |
| AC per story (Given/When/Then) | 45/50 | All stories have AC in Given/When/Then format. Story 2 and Story 4 have additional "And" clauses extending coverage. Format is correct throughout. |
| AC verifiability & boundary coverage | 22/50 | Multiple verifiability problems: (1) Story 1: "能按 WatchAlert 内部角色使用相应功能" — "相应功能" is undefined, not objectively testable. (2) Story 2: "宽限期过后降级" — grace period duration undefined, making the AC unverifiable without a concrete threshold. (3) Story 3: "返回 401/404" — two options; which is the contract? (4) Story 4: "状态给出显著标记" — "显著" is subjective. "统一质量门禁可触发...且通过" — "通过" threshold undefined (100%? 80%?). Missing ACs: no error-case AC for eiam callback failure, no AC for session expiry, no AC for concurrent access, no AC for migration rollback. |

**Subtotal: 155/200**

---

### Dimension 5: Scenario Completeness (150 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| End-to-end scenario coverage | 30/60 | Three main flows described: login (full lifecycle), startup (full lifecycle), migration (bullet list only). The migration flow ("并行运行、配置迁移、切流、旧实例退役（含回滚方式）") is a compressed outline, not a scenario walkthrough — no trigger, no step-by-step, no final state per step. Missing scenarios: session expiry during active use (BIZ-auth-008), single sign-out propagation (BIZ-auth-005), concurrent user during cutover, rollback execution. The degradation scenario is the best-described but still lacks the exact grace-period boundary behavior. |
| Implicit assumptions surfaced | 12/40 | Major unstated assumptions: (1) **Grace period duration** — used in flow, UI, and ACs but never defined. (2) **eiam-to-WatchAlert role mapping** — the PRD says "eiam 只负责认证" and "WatchAlert 内部 admin/user 角色决定" access, but never explains HOW an eiam-authenticated user gets assigned a WatchAlert internal role. Is it pre-provisioned? First-login admin assignment? Automatic? (3) **Session sharing model** — BIZ-auth-005 references shared Redis session; the PRD doesn't state whether WatchAlert joins this shared session or maintains its own. (4) **eiam session lifetime vs. WatchAlert session lifetime** — are they aligned? What if they diverge? |
| Business-rules consistency | 15/50 | Multiple business-rule violations/gaps: (1) **BIZ-auth-005 (single sign-out)**: Not addressed anywhere — when a user logs out from the platform console, the WatchAlert session should also be destroyed via shared Redis session. PRD is silent. (2) **BIZ-auth-008 (session expiry)**: "会话过期 → 跳转登录页（不白屏、不无限 loading）" — the PRD flow diagram only checks auth at initial access (`CheckAuth{是否已登录?}`), not mid-session. No scenario covers session expiry during active WatchAlert use. (3) **BIZ-auth-003 (403 vs 401)**: PRD says old login returns "401/404" — per BIZ-auth-003, 401 = not authenticated (semantic mismatch for a removed endpoint); should be 404 only. (4) **BIZ-resilience-001**: PRD degradation is richer than the rule but doesn't mention "已登录的 e-cam-web 会话不受影响" — degradation of WatchAlert's auth should not cascade to other apps sharing the session. (5) **BIZ-auth-001 (server-side authorization)**: PRD says eiam only does authentication, authorization stays in WatchAlert — but doesn't confirm WatchAlert enforces server-side authorization for its own CRUD endpoints, potentially leaving a gap where eiam's server-side enforcement is bypassed. |

**Subtotal: 57/150**

---

### Dimension 6: Edge Case Coverage (100 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Error paths documented | 18/40 | Covered: eiam auth failure (UF-1 "登录失败" state), eiam unavailable (UF-2/UF-3 degradation), build failure (developer flow "旧进程继续跑"). Missing: (1) eiam callback tampering/forgery — no validation of callback integrity. (2) Network timeout during eiam redirect. (3) Token validation failure mid-session. (4) Migration cutover failure — "含回滚方式" is mentioned but no rollback steps described. (5) Race condition: eiam recovers while user is mid-write in degraded mode. |
| Boundary conditions covered | 8/35 | Not covered: (1) Grace period boundary — what happens at the exact moment of expiry? Does an in-flight write succeed or fail? (2) Concurrent logins from multiple tabs — BIZ-auth-005 says no real-time broadcast, so multi-tab behavior needs definition. (3) Empty alarm data in read-only view. (4) Maximum session duration / token refresh boundary. (5) Config migration with large rule sets — performance during import. (6) Concurrent cutover with active alarm ingestion. |
| Failure recovery described | 8/25 | Mentioned but not detailed: (1) "eiam 恢复后自动回正" (UF-3) — what does "自动回正" mean? Auto-refresh? Redirect? Does the user need to re-login? (2) Migration rollback — "含回滚方式" with zero steps described. (3) Build failure — "旧进程继续跑" is stated but recovery (rebuild, deploy) is not. No recovery step includes user communication or expected downtime. |

**Subtotal: 34/100**

---

### Dimension 7: Scope Clarity (100 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete deliverables | 23/35 | Most items are concrete: "源码纳入 workspace", "一键启动后端 + 前端", "移除本地登录入口（后端 handler + 签名密钥 + 前端入口）". But: (1) "迁移切流：并行运行、配置迁移、切流、旧实例退役（含回滚方式）" — "含回滚方式" is a procedure, not a deliverable. (2) "作废旧接入决策，新增子项目接入决策记录" — this is a documentation/administrative task, not a feature deliverable; it's unclear what artifact is produced. (3) "eiam 不可用时的降级：已登录用户在宽限窗口内继续访问，过期后只读免登录" — "宽限窗口" is undefined, making this deliverable's acceptance criteria unclear. |
| Out-of-scope explicitly lists deferred items | 28/30 | Four items listed and named: "历史数据迁移", "前端嵌入门户控制台布局", "eiam 功能性/能力增强", "权限/角色随 eiam 同步". All are specific and clear. Minor: "仅允许为下游鉴权所需的最小改动" for eiam is vague — what constitutes "最小"? |
| Scope consistent with functional specs and user stories | 22/35 | In-scope items largely match UI functions and user stories. But: (1) "生产反代入口调整，开发态直接由编排拉起" — no user story or UI function covers the reverse proxy change. This is a backend/ops deliverable with no user-facing story. (2) "作废旧接入决策，新增子项目接入决策记录" — no user story or UI function covers this. (3) The scope includes "纳入平台统一质量门禁" but no UI function or user story covers the quality gate interaction (only Story 4 mentions it tangentially in an AC). (4) Scope says "迁移切流" but no user story covers the migration/cutover experience for any user role. |

**Subtotal: 73/100**

---

## Cross-Dimension Coherence Check

1. **Grace period appears in 4 places, defined in 0**: Flow Description ("宽限期内"), UI Function 2 ("宽限剩余时长"), User Story 2 AC ("宽限期过后"), Goals (implied by degradation). No document specifies the actual duration. This is a cross-dimension gap that manifests most severely in Functional Specs and Scenario Completeness.

2. **Role mapping gap spans Background → User Stories → Functional Specs**: Background says "WatchAlert 内部 admin/user 角色决定" access. User Story 1 AC says "按 WatchAlert 内部角色使用相应功能". But no document explains the mapping mechanism. This makes Story 1's AC unverifiable and creates an implicit assumption gap in Scenario Completeness.

3. **Scope promises migration, stories don't deliver**: Scope In-Scope includes "迁移切流" with parallel run, config migration, cutover, rollback. But no user story covers migration — neither the admin performing cutover nor the user experiencing it. The migration flow has no diagram. This is a scope-vs-stories inconsistency.

---

## Phase 3: Blindspot Hunt

### [blindspot] B1: eiam-to-WatchAlert role mapping mechanism is entirely absent

**Quote**: "授权边界（已确认）：eiam 只负责认证（你是谁）；登录后能干什么（看告警/配置规则/管通知）仍由 WatchAlert 内部 admin/user 角色决定，不随 eiam 角色/组织映射。"

**Issue**: The PRD establishes that eiam authenticates identity and WatchAlert's internal roles govern authorization. But it never explains HOW an eiam-authenticated user acquires a WatchAlert internal role. Is it:
- Pre-provisioned (admin manually assigns WatchAlert roles to eiam users)?
- First-login assignment (first eiam login gets default "user" role)?
- Identity-claim based (eiam passes a claim that WatchAlert interprets)?

Without this mechanism, the entire authorization model is a black box. User Story 1's AC ("按 WatchAlert 内部角色使用相应功能") is unverifiable because the role assignment path is undefined. This is a fundamental gap that no single rubric dimension fully captures — it spans Background (who are the users), User Stories (AC verifiability), and Scenario Completeness (implicit assumptions).

**Required**: Add a section describing the role provisioning model — how eiam users are mapped to WatchAlert internal roles, including first-login behavior, role assignment authority, and role persistence.

---

### [blindspot] B2: Quality Checklist self-certification contradicts document content

**Quote**: Quality Checklist: "[x] 无含糊措辞（阈值/清单/口径均量化）" vs Goals table: "启动耗时 ≤ 编译 + 数十秒"

**Issue**: The document self-certifies that all thresholds are quantified, but "数十秒" (tens of seconds) is a range, not a number. Additionally, the grace period ("宽限窗口") is used in multiple sections but never quantified. The self-certification creates false confidence and would cause a reviewer to skip verification of quantification.

**Required**: Either quantify all thresholds (replace "数十秒" with a concrete number like "≤ 编译 + 30s"; define grace period as a concrete duration like "2 hours") or uncheck the Quality Checklist item and list the unquantified items.

---

### [blindspot] B3: Single sign-out (BIZ-auth-005) is completely absent from all documents

**Quote**: PRD spec In-Scope: "登录统一到 eiam（认证委托），移除本地登录入口（后端 handler + 签名密钥 + 前端入口）"

**Issue**: The PRD focuses exclusively on login (authentication delegation) and login entry removal. It never mentions logout/session destruction. BIZ-auth-005 requires bidirectional single sign-out via shared Redis session — when a user logs out from the platform console, all other apps (including WatchAlert) must destroy the session on next request (401 convergence). The PRD doesn't state whether WatchAlert joins the shared Redis session, whether it implements logout, or how it handles 401 convergence from console sign-out. This is a cross-cutting concern between auth business rules and the PRD's scope that falls between the cracks of individual dimensions.

**Required**: Add single sign-out behavior to the scope, flow description, and user stories. Specify whether WatchAlert joins the shared Redis session and how it handles session destruction propagated from the console.

---

### [blindspot] B4: "上游贡献不中断" conflates a constraint with a goal

**Quote**: Goals table: "上游贡献不中断 | 迁入后仍能向开源上游提交通用修复 | 防止「死 fork」"

**Issue**: This is listed as a Goal with a Metric column, but the "metric" ("仍能向开源上游提交通用修复") is not a measurable target — it's an ongoing capability. Goals should define a state to achieve; this defines a constraint to maintain. Listing it as a goal inflates perceived goal coverage and implies it has a success criterion, which it doesn't. This is a fundamental reasoning flaw: the document treats a non-functional constraint as a measurable goal, which would cause a team to either skip measuring it or invent an arbitrary metric.

**Required**: Move "上游贡献不中断" to a Constraints section (not Goals), or define a concrete metric (e.g., "迁入后 90 天内至少 1 个 PR 合入上游").

---

### [blindspot] B5: Migration/cutover has no user story and no flow diagram despite being In-Scope

**Quote**: In-Scope: "迁移切流：并行运行、配置迁移、切流、旧实例退役（含回滚方式）"

**Issue**: Migration/cutover is listed as a core in-scope deliverable with 4 sub-steps, yet:
- No user story covers it (no story for the admin performing cutover, no story for the user experiencing the transition)
- No Mermaid diagram exists for 主流程 3 (unlike 主流程 1 and 2 which both have diagrams)
- The "回滚方式" is mentioned but never described

This is a structural misalignment: scope promises a deliverable, but neither user stories nor flow diagrams deliver it. The migration is arguably the highest-risk part of the feature (production cutover with rollback), yet it has the least specification depth.

**Required**: Add at least one user story for migration/cutover (e.g., "As a 告警平台管理员, I want to execute a controlled cutover with rollback capability..."), add a Mermaid flow diagram for the migration process, and describe rollback steps.

---

## Score Summary

| Dimension | Max | Score |
|-----------|-----|-------|
| 1. Background & Goals | 100 | 75 |
| 2. Flow Diagrams | 150 | 102 |
| 3. Functional Specs | 200 | 122 |
| 4. User Stories | 200 | 155 |
| 5. Scenario Completeness | 150 | 57 |
| 6. Edge Case Coverage | 100 | 34 |
| 7. Scope Clarity | 100 | 73 |
| **Total** | **1000** | **618** |

**Target: 900/1000 — NOT MET (gap: 282 points)**
