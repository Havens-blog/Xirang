# Iteration 2 — PM Adversarial Scoring Report

**Scorer**: Senior Product Manager (adversarial)  
**Iteration**: 2 (verifies 14 iteration-1 fixes, hunts new contradictions)  
**Date**: 2026-10-09  

---

## Iteration-1 Fix Verification (14 points)

| # | Original Attack | Status | Evidence |
|---|-----------------|--------|----------|
| 1 | 宽限期时长未定义 | **FIXED** | "宽限期（2 小时）" appears consistently in spec Goals, UF-2 Data Requirements, Story 2 AC |
| 2 | Story 1 缺错误 AC | **FIXED** | Story 1 now has 5 ACs: happy path, eiam failure, eiam unavailable, 403 no-role, session expiry |
| 3 | 会话过期场景缺失 | **FIXED** | "会话过期流程（BIZ-auth-008）" in Flow Description + Mermaid node `ExpMid → ExpLogin → ReAuth` |
| 4 | 单点登出未覆盖 | **FIXED** | "单点登出流程（BIZ-auth-005）" in Flow Description, In Scope item, Story 5 |
| 5 | 迁移切流缺 Mermaid 图 | **FIXED** | Dedicated migration cutover diagram with 5 phase gates + 3 rollback branches |
| 6 | 403 场景缺失 + 401/404 语义不一致 | **FIXED** | 403 branch in diagram (`Role -->|否| Forbidden`), 404 for removed endpoints, 401 for unauthenticated |
| 7 | 迁移切流无对应故事 | **FIXED** | Story 6 added with 7 Given/When/Then blocks |
| 8 | 迁移回滚步骤与验证门控缺失 | **FIXED** | Pass/fail criteria per phase, rollback triggers ("鉴权成功率 < 99%、5xx 率 > 1%"), timing (≤2 min), data consistency guarantee |
| 9 | 角色映射机制空白 | **FIXED** | `watchalert_role_binding` table documented, eiam_uid → admin/user, pre-provisioned by admin, 403 for no match |
| 10 | 回跳地址开放重定向风险 | **FIXED** | Whitelist validation in UF-1 Validation Rules + spec Security Requirements |
| 11 | token 多进程存储/验签机制矛盾 | **PARTIALLY FIXED — NEW CONTRADICTION INTRODUCED** | HMAC-SHA256 signing resolved verification, but in-memory revocation list in multi-process creates a new contradiction (see Attack 2) |
| 12 | eiam 抖动无防抖/滞回 | **FIXED** | Hysteresis: 3 consecutive failures (30s) to declare unavailable, 3 consecutive successes to recover |
| 13 | 降级探针机制未定义 | **FIXED** | `GET /eiam/health`, 10s interval, 3s timeout, SSE/polling to frontend |
| 14 | 目标未量化 + 自证矛盾 | **FIXED** | All Goals quantified; checklist updated to match. But Performance Requirements still contain vague language (see Attack 6) |

**Summary**: 13/14 fully resolved. Point 11 partially resolved — the signing/verification gap closed, but the fix introduced a multi-process revocation contradiction.

---

## Phase 1: Reasoning Audit (Pre-Score Anchors)

### Anchor 1: Session revocation contradicts multi-process verification model

The spec states simultaneously:
- "多进程下任一进程可凭共享密钥独立验签，不依赖共享存储读会话" (each process verifies independently, no shared storage)
- "会话票仅存内存（不落盘、不明文缓存），登出时加入内存撤销列表直至过期" (in-memory revocation list only)
- "登出销毁共享会话，其余标签页下次请求 401 收敛" (BIZ-auth-005: all tabs must get 401)

In a multi-process deployment, per-process in-memory revocation lists are NOT shared. Process A handles logout and adds the ticket to ITS revocation list. Process B has no knowledge of this revocation. A subsequent request hitting process B will still accept the revoked ticket (HMAC signature is still valid, ticket hasn't expired). This directly violates BIZ-auth-005's requirement that "其余应用 / 标签页在下一次请求时 401 收敛".

Furthermore, "不依赖共享存储" contradicts "复用平台共享会话" — if WatchAlert reuses the platform's shared session store (e.g., Redis), it DOES depend on shared storage. If it uses stateless HMAC tickets, it is NOT reusing a shared session store.

### Anchor 2: Implicit prerequisites on eiam endpoints and tables

The spec assumes but does not confirm existence of:
- `GET /eiam/health` endpoint (probe target) — is this an existing eiam endpoint or does it need to be built? If new, it's an eiam change, which is partially Out of Scope
- `watchalert_role_binding` table — where does it live (WatchAlert DB? platform DB? eiam?)? Who creates the schema? Is there a management UI or is it manual SQL?
- Platform "密钥管理" system that distributes HMAC keys to all processes — is this an existing system?
- eiam returning a "用户唯一标识" suitable for matching — does eiam currently expose such an identifier?

These are unstated environmental dependencies that could block implementation.

### Anchor 3: UF-3 self-contradiction on login entry in read-only view

UF-3 Validation Rules: "只读视图不得出现任何登录入口"  
UF-3 User Interaction Flow step 3: "eiam 恢复 → 自动回正到正常登录流程"

If no login entry point can exist in the read-only view, how does the user transition to login when eiam recovers? "自动回正" implies auto-redirect, but this is not explicitly stated — it could also mean showing a login button, which would violate the validation rule.

---

## Phase 2: Rubric Scoring

### Dimension 1: Background & Goals (100 pts)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Three elements (Reason/Target/Users) | 30/30 | All present, specific. User table with 3 roles + authorization boundary note. |
| Goals quantified | 28/30 | 5 goals all have metrics (100% eiam, ≤编译+30s, 404/no-token, quarterly ≥1 PR + CI green, 2h grace). "启动耗时 ≤ 编译 + 30 秒" is relative but testable. "上游贡献" is a process metric, not a product outcome metric — minor. |
| Logical consistency | 38/40 | Goals follow from problems, except "上游贡献不中断" which has no direct counterpart in the stated problems (problems are about integration, auth, quality gates — not upstream divergence). The upstream goal is a legitimate guardrail but is not derived from the problem statement. |

**Subtotal: 96/100**

### Dimension 2: Flow Diagrams (150 pts)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Mermaid diagram exists | 50/50 | Three Mermaid flowcharts: main login flow, developer startup, migration cutover. |
| Main path complete (start → end) | 47/50 | Login → dashboard complete. Developer startup → access complete. Migration → retirement complete. But the **logout flow is not in any Mermaid diagram** — logout is described in text ("单点登出流程") but has no visual representation showing the shared-session destruction → other-tab-401-convergence path. |
| Decision points + error branches | 45/50 | Rich diamond coverage: CheckAuth, EiamOK, Role, Deg, ExpMid, compile success/fail, 5 migration gate diamonds with 3 rollback branches. Error branches: auth failure → error prompt, no role → 403, degradation → banner/read-only, compile fail → old process continues. Missing: logout → 401 convergence branch is not visualized. |

**Subtotal: 142/150**

### Dimension 3: Functional Specs — prd-ui-functions.md (200 pts)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Placement & Interaction completeness | 63/70 | All 3 UF have Placement (mode/target page/position). Interaction flows cover full paths. Gaps: UF-3 flow step 3 "eiam 恢复 → 自动回正到正常登录流程" is ambiguous (auto-redirect? show button? for unauthenticated users?). Secondary Pages table lists "eiam 登录跳转页" as a secondary page of the alert platform, but this is eiam's external page — misleading classification. |
| Data Requirements & States clarity | 62/70 | Field tables and state tables filled for all 3 UF. Sources and triggers explicit. Gaps: UF-2 has no state for "probe failures in progress, not yet 3 consecutive" (pre-hysteresis detection state) — user sees normal UI while probes are failing but not yet declared unavailable. UF-3 "只读告警数据 \| 列表 \| 后端只读接口" — which specific interfaces? Not named. "鉴权服务状态" field mentions SSE/轮询 but no endpoint specified. |
| Validation Rules explicit | 52/60 | UF-1 has detailed whitelist rules (host match, path constraints, scheme blacklist) — excellent. UF-2 "横幅须明确告知「哪些操作禁止」" is a design guideline, not an actionable validation rule — what specific text must appear? What operations must be enumerated? UF-3 "只读视图不得出现任何登录入口" is testable but **contradicts** the flow step 3 "自动回正到正常登录流程" (if auto-redirect, it's fine; if a button is shown, it violates the rule — the spec doesn't disambiguate). |

**Subtotal: 177/200**

### Dimension 4: User Stories (200 pts)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Coverage: one story per target user | 48/50 | 运维/值班工程师 (Stories 1, 2, 5), 告警平台管理员 (Stories 3, 6), 平台开发者 (Story 4). All 3 target users covered. Minor: Story 4 doesn't cover the quality gate failure scenario (what if gate fails — is it a blocker or warning?). |
| Format correct (As a / I want / So that) | 48/50 | All stories follow the format. Actions are concrete ("用平台统一账号登录", "执行统一启动命令", "执行迁移切流"). Minor: Story 1 AC "并能按 WatchAlert 内部角色使用相应功能" — "相应功能" is slightly vague. |
| AC per story (Given/When/Then) | 48/50 | All 6 stories have Given/When/Then ACs. Story 6 has multiple G/W/T blocks. Minor: Story 4 uses "And" clauses after "Then" which is acceptable Gherkin but mixes structural levels. |
| AC verifiability & boundary coverage | 40/50 | Most ACs are objectively testable. Key gaps: **Story 6 AC "抽样规则触发验证一致"** — sample size and methodology undefined, cannot objectively verify what "一致" means. **Story 2 AC** — doesn't specify HTTP status for rejected writes (503 is in UI doc only, not in story AC; boundary of grace period expiry mid-operation not covered). **Story 4 AC "状态/日志可查"** — vague: what format? what interface? **Story 2** — no AC for write operation attempted at exact moment of grace period expiry. |

**Subtotal: 184/200**

### Dimension 5: Scenario Completeness (150 pts)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| End-to-end scenario coverage | 52/60 | All major scenarios covered from trigger to end state: login SSO, session expiry, logout, developer startup, migration cutover (5 phases), degradation. Gap: **`watchalert_role_binding` pre-provisioning is not a described scenario** — the spec says "管理员配置迁移时预置" but doesn't describe how (manual SQL? admin UI? batch import from old accounts?). This is a missing end-to-end scenario for a critical onboarding step. |
| Implicit assumptions surfaced | 28/40 | Multiple unstated prerequisites: (1) `GET /eiam/health` endpoint existence — not confirmed as existing vs. new build; if new, it's an eiam change partially Out of Scope. (2) `watchalert_role_binding` table location/schema — not specified. (3) Platform "密钥管理" system for HMAC key distribution — assumed but not confirmed. (4) eiam returning a suitable unique identifier — assumed. (5) WatchAlert joining the platform shared session mechanism — listed in In Scope as "纳入平台统一启动编排" but shared session joining is not an explicit step. |
| Business-rules consistency | 42/50 | BIZ-auth-003 (403 原地提示) ✓. BIZ-auth-005 (单点登出) — **partially violated**: the in-memory per-process revocation model cannot guarantee "其余应用/标签页在下一次请求时 401 收敛" across multiple processes without shared storage. BIZ-auth-008 (会话过期) ✓. BIZ-auth-006 (防枚举统一文案) ✓. BIZ-resilience-001 (eiam 不可用降级) ✓. BIZ-auth-009 (eiam API 契约基线) — not addressed: the feature depends on eiam endpoints but doesn't mention registering them in the parity checklist. |

**Subtotal: 122/150**

### Dimension 6: Edge Case Coverage (100 pts)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Error paths documented | 35/40 | eiam auth failure, eiam unavailable, 403 no-role, session expiry, compile failure, migration rollback, 404 for removed endpoints. Missing: what if `watchalert_role_binding` contains an invalid role value (not admin/user)? The spec says "admin/user" but doesn't address invalid values — does the user get 403? 500? Undefined behavior? |
| Boundary conditions covered | 23/35 | Grace period boundary (2h), probe timeout (3s), migration gate thresholds (99.9%, P95 ≤ 300ms), 10% gray rollout covered. Missing: (1) **Multi-process revocation boundary** — in-memory revocation list not shared across processes, ticket revoked on one process still valid on another. (2) In-flight write at grace period expiry — user submits a write at 1h59m59s, what happens? (3) Large config migration performance — "配置条数对齐" but no threshold for large datasets. (4) Concurrent `watchalert_role_binding` modification during migration. (5) Probe slow-but-not-timing-out (e.g., consistent 2.9s response). |
| Failure recovery described | 18/25 | eiam recovery (auto via probe), migration rollback (≤2min, old instance continues), compile failure (old process continues). Missing: (1) **Rollback failure recovery** — "反代回切旧实例（≤ 2 分钟）" assumes rollback always succeeds; what if reverse proxy can't switch? (2) Partial config migration failure — "重导或回滚至并行阶段" but what about partially imported configs? (3) Revocation list overflow — if many users logout simultaneously, in-memory revocation list grows unbounded until expiry. |

**Subtotal: 76/100**

### Dimension 7: Scope Clarity (100 pts)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| In-scope items are concrete deliverables | 32/35 | Most items are specific (source into workspace, unified startup, quality gates, eiam auth delegation, role binding, degradation, migration phases, logout, session expiry). Gap: "作废旧接入决策，新增子项目接入决策记录" — what format? Where stored? Who reviews? This is a vague process item, not a concrete deliverable. |
| Out-of-scope explicitly lists deferred items | 28/30 | 4 deferred items named (historical data migration, portal embedding, eiam enhancement, role sync). Gap: "eiam 功能性/能力增强（仅允许为下游鉴权所需的最小改动）" — "最小改动" is vague. The spec itself requires `GET /eiam/health` (if new) and `watchalert_role_binding` (if in eiam DB) — are these "最小改动" or are they Out of Scope? The boundary is unclear. |
| Scope consistent with functional specs and user stories | 33/35 | In Scope items map to user stories and UI functions consistently. "作废旧接入决策" has no corresponding user story or functional spec — minor inconsistency. All other items have clear traceability. |

**Subtotal: 93/100**

---

## Phase 3: Blindspot Hunt

### [blindspot] B1: Cross-document status code gap for write rejection during degradation

The spec's Flow Description says "写操作被拒" (no HTTP status). The UI Functions doc says "后端返回 503" (UF-2, step 3). Story 2 AC says "写操作被拒" (no status). The three documents are not aligned on the rejection mechanism. A developer reading the spec would not know to return 503; a developer reading the UI doc would. This is a cross-document inconsistency that falls between dimensions.

> Quote (spec): "宽限期（2 小时）内可继续只读查看、写操作被拒"  
> Quote (UI): "用户尝试写操作（编辑规则/确认告警/改通知）→ 后端返回 503，前端提示不可写"  
> Quote (Story 2): "宽限期（2 小时）内我可只读查看、写操作被拒"

### [blindspot] B2: watchalert_role_binding invalid role value — undefined behavior

The spec states roles are "admin/user" and unmatched users get 403. But what happens if a binding exists with an invalid role value (e.g., `"superadmin"`, `null`, empty string)? The spec doesn't address this. BIZ-identity-005 establishes that unknown status values cause silent degradation — the same pattern applies here. An invalid role binding could either grant unintended access (if treated as admin) or silently deny access (if treated as 403).

> Quote: "eiam 用户登录后，WatchAlert 后端以 eiam 返回的用户唯一标识匹配预置角色表（`watchalert_role_binding`：eiam_uid → admin/user）"

### [blindspot] B3: Session ticket "仅存内存" vs "复用平台共享会话" — self-contradictory session model

The spec uses two incompatible session models simultaneously:
1. Stateless HMAC ticket: "不依赖共享存储读会话" — no shared storage needed for verification
2. Shared session: "复用平台共享会话" — implies shared storage (Redis)

These are architecturally incompatible. Either WatchAlert uses stateless HMAC tickets (no shared session store) OR it joins the platform's shared session store (depends on Redis). The spec cannot have both. The logout revocation mechanism depends on which model is used:
- If stateless: revocation list must be shared (contradicts "仅存内存" per-process)
- If shared session: logout destroys the Redis key (consistent with BIZ-auth-005), but then "不依赖共享存储" is false

> Quote: "多进程下任一进程可凭共享密钥独立验签，不依赖共享存储读会话。会话票仅存内存（不落盘、不明文缓存）"  
> Quote: "WatchAlert 不自建独立会话存储，复用平台共享会话。"

---

## Cross-Dimension Coherence Check

1. **Scope vs Functional Specs**: In Scope "eiam 不可用时的降级" maps to UF-2 + UF-3 + Story 2 ✓. In Scope "迁移切流" maps to Story 6 ✓. Consistent.
2. **Goals vs User Stories**: Goal "本地登录后门清零" maps to Story 3 ✓. Goal "一键启动" maps to Story 4 ✓. Consistent.
3. **Security Requirements vs Flow Description**: Session ticket mechanism (Security) contradicts shared session destruction (Flow) — see blindspot B3.
4. **Performance Requirements vs Quality Checklist**: Checklist self-certifies "无含糊措辞" but Performance Requirements contain "明显回退" and "现有告警平台用户规模" — cross-section inconsistency.

---

## Final Score

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 96 | 100 |
| Flow Diagrams | 142 | 150 |
| Functional Specs | 177 | 200 |
| User Stories | 184 | 200 |
| Scenario Completeness | 122 | 150 |
| Edge Case Coverage | 76 | 100 |
| Scope Clarity | 93 | 100 |
| **Total** | **890** | **1000** |

**Gap to target (900): 10 points**

---

## Attack Points Summary

1. [Scenario Completeness] Implicit prerequisites not surfaced — `GET /eiam/health`, `watchalert_role_binding` schema/location, HMAC key distribution, shared session joining — all assumed but not confirmed as existing or new build. — **Must explicitly state whether each dependency is existing or requires new build, and if new, reconcile with Out of Scope "eiam 功能性/能力增强".**

2. [Edge Case Coverage] Multi-process session revocation contradiction — "会话票仅存内存（不落盘），登出时加入内存撤销列表" but "多进程下任一进程可凭共享密钥独立验签" — per-process in-memory revocation lists cannot guarantee BIZ-auth-005's "其余标签页下次请求 401 收敛" across processes. — **Must specify either: (a) shared revocation store (contradicts "不依赖共享存储"), or (b) accept that revocation is per-process and cannot guarantee cross-process logout convergence.**

3. [Functional Specs] UF-3 self-contradiction — "只读视图不得出现任何登录入口" vs "eiam 恢复 → 自动回正到正常登录流程" — if recovery shows a login button, it violates the validation rule; if it auto-redirects, this must be stated explicitly. — **Must clarify: auto-redirect to eiam on recovery (not a login button), and state that the redirect is triggered by the backend probe signal, not a user-facing entry point.**

4. [User Stories] Story 6 AC "抽样规则触发验证一致" is not objectively verifiable — sample size, sampling methodology, and "一致" criteria are undefined. — **Must specify: sample size (e.g., ≥10% of rules or ≥50 rules, whichever is greater), sampling methodology (random vs. stratified), and "一致" definition (same alert output for same input within X seconds).**

5. [Flow Diagrams] Logout flow missing from all Mermaid diagrams — "单点登出流程（BIZ-auth-005）" is described in text but has no visual representation. — **Add a logout branch to the main flow diagram: User Logout → Destroy Shared Session → Other Tab Next Request → 401 → Converge to eiam Login.**

6. [Background & Goals] Performance Requirements contain vague language despite Quality Checklist self-certifying "无含糊措辞" — "不因鉴权委托引入明显回退" (what threshold is "明显"?) and "满足现有告警平台用户规模" (what number?). — **Quantify: e.g., "鉴权延迟 P95 ≤ 300ms, P99 ≤ 500ms" (already stated elsewhere) and "支持 ≥ N 并发用户" with a specific N. Remove the checklist self-certification or make it accurate.**

7. [blindspot] Cross-document status code gap — spec says "写操作被拒" (no code), UI doc says "后端返回 503", Story 2 AC says neither. — **Align all three documents: spec Flow Description should state 503, Story 2 AC should include "Then 写操作返回 503 且前端提示不可写".**

8. [Edge Case Coverage] Rollback failure not addressed — "反代回切旧实例（≤ 2 分钟）" assumes rollback always succeeds; no fallback for reverse proxy switch failure. — **Must specify: what happens if reverse proxy cannot switch back (e.g., config error, network partition)? Fallback: manual DNS switch, or keep both instances in read-only.**

9. [Functional Specs] UF-2 missing pre-hysteresis detection state — probe failures in progress (1-2 failures, not yet 3) have no corresponding UI state. — **Add a state: "探针进行中" (probe in progress) or explicitly state that 1-2 probe failures are invisible to the user (acceptable design decision, but must be documented).**

10. [blindspot] `watchalert_role_binding` invalid role value — undefined behavior for bindings with role values other than "admin"/"user" (e.g., null, empty, "superadmin"). — **Must specify: invalid role values are treated as 403 (no access), matching the "no match" behavior, and a validation rule should be added to the binding management process.**
