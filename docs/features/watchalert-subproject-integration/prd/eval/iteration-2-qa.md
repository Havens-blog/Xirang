# Iteration 2 — QA Scorer Report (Adversarial)

**Scorer**: Senior QA Engineer (对抗视角)
**Mode**: A (Feature WITH UI)
**Iteration**: 2 (revision against iteration-1-merged 14 attack points)

---

## Verification: Iteration-1 Issues (14 条) — Resolution Status

| # | Issue | Status | Evidence |
|---|-------|--------|----------|
| 1 | 宽限期时长未定义 | **FIXED** | "宽限期（2 小时）" — prd-spec L54, L95; UF-2 L89 |
| 2 | Story 1 缺错误 AC | **FIXED** | Story 1 now has 5 ACs covering auth fail, eiam unavailable, 403, session expiry — prd-user-stories L14-28 |
| 3 | 会话过期场景缺失 | **FIXED** | Flow + Mermaid ExpMid node — prd-spec L74-75, L116-118 |
| 4 | 单点登出未覆盖 | **FIXED** | Flow + Story 5 + In-Scope — prd-spec L57, L77-78; Story 5 L75-87 |
| 5 | 迁移切流缺 Mermaid 图 | **FIXED** | 迁移切流流程图 with 阶段门控 + 回滚分支 — prd-spec L135-152 |
| 6 | 403 缺失 + 401/404 语义不一致 | **FIXED** | Role diamond → Forbidden 403; 404 for removed endpoint — prd-spec L104-107, L197 |
| 7 | 迁移切流无对应故事 | **FIXED** | Story 6 — prd-user-stories L90-107 |
| 8 | 回滚步骤与验证门控缺失 | **FIXED** | Per-stage pass/fail + rollback trigger — prd-spec L84-92 |
| 9 | 角色映射机制空白 | **FIXED** | watchalert_role_binding — prd-spec L31, L53 |
| 10 | 回跳地址开放重定向风险 | **FIXED** | White-list validation — prd-spec L196; UF-1 L66 |
| 11 | token 多进程存储/验签矛盾 | **PARTIALLY FIXED — NEW CONTRADICTION INTRODUCED** | HMAC-SHA256 ticket described, but contradicts "复用平台共享会话" and BIZ-auth-005's "共享 Redis session" (see Attack 1) |
| 12 | eiam 抖动无防抖 | **FIXED** | 滞回机制: 3 consecutive failures / 3 successes — prd-spec L95 |
| 13 | 降级探针机制未定义 | **FIXED** | GET /eiam/health, 10s, 3s timeout — prd-spec L191 |
| 14 | 目标未量化 + 清单自证矛盾 | **FIXED** | Goals quantified; checklist updated — prd-spec L38-45, L203-213 |

**Summary**: 13/14 fully resolved. Issue #11 (session mechanism) was addressed with an HMAC ticket design that introduces a **new internal contradiction** and a **business-rule violation** — the fix created a worse problem than it solved.

---

## Phase 1: Reasoning Audit (Pre-Score Anchors)

### Argument Chain Trace

1. **Problem → Solution**: Problem = WatchAlert fragmented (two accounts, code outside workspace, no unified gate). Solution = bring into workspace as subproject + unify login to eiam + remove local login. **Sound match.**

2. **Solution → Evidence**: Goals have metrics (100% eiam, ≤compile+30s, 404 for removed endpoints, quarterly PR, ≥99.9% auth success, P95 ≤300ms). Evidence supports solution. **Sound.**

3. **Evidence → Success Criteria**: Success criteria test what matters (account unification, startup, backdoor removal, upstream contribution, migration safety). **Mostly sound**, but "质量门禁覆盖" goal metric is "可一键触发" — a capability flag, not a measurable quality target (e.g., coverage %). Mild proxy.

4. **Self-contradiction check — CRITICAL FINDING**: The PRD contains a direct internal contradiction on session architecture:
   - prd-spec L78: "WatchAlert 不自建独立会话存储，**复用平台共享会话**"
   - prd-spec L195: "**不依赖共享存储读会话**。会话票**仅存内存**（不落盘、不明文缓存），登出时加入**内存撤销列表**"
   - BIZ-auth-005: "销毁**共享 Redis session**"

   These three statements describe two mutually exclusive architectures:
   - (A) Shared Redis session store (BIZ-auth-005, L78) — centralized, cross-process logout works
   - (B) Self-signed in-memory HMAC ticket with per-process revoke list (L195) — decentralized, cross-process logout **cannot work**

   The fix for iteration-1 issue #11 adopted architecture (B) while the flow description and In-Scope items still claim architecture (A). This is a **mutual-exclusion contradiction** that breaks the single sign-out goal.

---

## Phase 2: Rubric Scoring

### Dimension 1: Background & Goals — 95/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Background 三要素 (Reason/Target/Users) | 30/30 | Reason (L15), Target (L19), Users table (L23-27) all present and specific. Three roles with clear descriptions. |
| Goals 量化 | 28/30 | Multiple numeric targets: 100% eiam, ≤compile+30s, 404, 季度≥1PR, ≥99.9%, P95≤300ms. **Docked 2**: "质量门禁覆盖" metric is "可一键触发" — a capability flag, not a measurable quality target (e.g., coverage %, pass rate). |
| Background 与 Goals 逻辑一致 | 37/40 | Goals follow from stated problem. **Docked 3**: Performance Requirements L181 says "并发：满足现有告警平台用户规模，不因鉴权委托引入明显回退" — "满足现有规模" and "明显回退" are unquantified, inconsistent with the checklist claim "无含糊措辞". |

### Dimension 2: Flow Diagrams — 140/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Mermaid 图存在 | 50/50 | Three diagrams: login flow (L99-119), developer startup (L123-131), migration cutover (L135-152). |
| 主路径完整 | 48/50 | Login happy path start→end complete. **Docked 2**: Recovery transition (eiam 恢复 → 自动回正) mentioned in text (L95, L117) but not shown as an arrow in the login flow diagram — ReadOnly/Anon nodes have no edge back to Dash. |
| 决策点 + 错误分支 | 42/50 | Decision diamonds and error branches present (EiamOK→Err, Role→Forbidden, Deg→Banner, G1-G5→rollback). **Docked 8**: Single sign-out flow (BIZ-auth-005) described in text (L77-78) but **not diagrammed** — logout → destroy session → other tabs 401 is a distinct flow with its own decision logic that downstream agents need to visualize. |

### Dimension 3: Functional Specs — 185/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Placement & Interaction 完整性 | 68/70 | All 3 UF have Placement + Interaction Flow. **Docked 2**: UF-2 interaction flow (L81-83) does not include a step for "eiam 恢复 → 横幅移除 → 恢复写操作" — the recovery-during-grace interaction is absent. |
| Data Requirements & States 清晰度 | 60/70 | Field tables and state tables filled for all UF. **Docked 10**: (a) UF-2 state table (L96-97) only has "宽限中" and "宽限过期" — **missing "已恢复" state** for when eiam recovers within grace window; UF-3 has "已恢复" (L131) but UF-2 does not, creating an asymmetry. (b) UF-3 Data Requirements (L123) says "只读告警数据 \| 列表 \| 后端只读接口 \| 看板/规则/事件" — does not specify which data fields are safe for **unauthenticated** access; alarm rules may contain sensitive notification channel configs or data source credentials. |
| Validation Rules 可执行性 | 57/60 | UF-1 redirect white-list (L66) is detailed and actionable. UF-3 (L134) is actionable. **Docked 3**: UF-2 validation rule (L100) "横幅须明确告知「哪些操作禁止」" is directive but does not enumerate the specific operations in the rule itself (the interaction flow L83 lists "编辑规则/确认告警/改通知" but the validation rule should consolidate); also no validation rule specifying consistent error code (503 per L83) for write rejection. |

### Dimension 4: User Stories — 185/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| 覆盖: 每个目标用户一个故事 | 50/50 | 运维/值班工程师 (Story 1, 2, 5), 告警平台管理员 (Story 3, 6), 平台开发者 (Story 4). All three roles covered. |
| 格式正确 (As a / I want / So that) | 50/50 | All 6 stories follow format. Actions concrete ("用平台统一账号登录", "确认旧本地登录入口已彻底移除", "执行迁移切流"). |
| AC per story (Given/When/Then) | 50/50 | Every story has ≥1 AC in Given/When/Then format. Story 1 has 5 ACs. Story 6 has multi-scenario ACs. |
| AC 可验证性 & 边界覆盖 | 35/50 | ACs are generally testable. **Docked 15**: (a) **Story 5 AC unverifiable under stated session architecture** — "其余标签页下次请求时后端返回 401" (L85) cannot be guaranteed if the other tab's request hits a **different process** that doesn't have the in-memory revoke list entry (see session contradiction, Attack 1). (b) Story 4 AC only covers backend compile failure (L70) — **frontend startup failure** not covered (what if :3001 fails to bind?). (c) Story 6 AC covers config diff=0 (L99) but not **partial config import failure** (some rules imported, some failed — cleanup? retry?). (d) Story 2 AC (L41) covers grace period but not the **exact boundary** (at 2h:00:00, what state transition?) or the **race** between grace expiry and eiam recovery. |

### Dimension 5: Scenario Completeness — 115/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| 端到端场景覆盖 | 53/60 | Login SSO, session expiry, single sign-out, developer startup, migration cutover, eiam degradation — all covered trigger→final state. **Docked 7**: (a) "只读免登录视图" scenario (UF-3) lacks specification of **which backend endpoints serve unauthenticated read-only data** — an unstated prerequisite for the scenario to function. (b) Migration "配置迁移" scenario (L87) lacks specification of the **tooling/process** — "导出...导入" is described at outcome level, not execution level. |
| 隐式假设显式化 | 30/40 | Role binding table, shared session assumption, probe endpoint, shared key distribution — all stated. **Docked 10**: (a) Read-only unauthenticated endpoints assumed to exist but never specified — what endpoints, what data filtering? (b) `watchalert_role_binding` pre-population method unspecified — SQL? Admin UI? CLI? "由告警平台管理员在配置迁移时预置" (L31) is an outcome, not an interface. (c) "平台密钥管理分发至各进程" (L195) — key distribution mechanism assumed but not specified (how are keys rotated? what happens on key compromise?). |
| 业务规则一致性 | 32/50 | BIZ-auth-003 (403 原地提示) ✓, BIZ-auth-005 (单点登出) **contradicted**, BIZ-auth-006 (统一文案) ✓, BIZ-auth-008 (会话过期) ✓, BIZ-resilience-001 (降级) ✓. **Docked 18**: (a) **BIZ-auth-005 contradiction** — business rule says "销毁共享 Redis session" (shared store), PRD says "会话票仅存内存...不依赖共享存储" (no shared store) — these are mutually exclusive architectures. The PRD's in-memory per-process revoke list cannot deliver BIZ-auth-005's "其余应用/标签页在下一次请求时 401 收敛" across processes. (b) BIZ-auth-009 (eiam API 契约基线) not referenced — PRD depends on eiam endpoints (login, /eiam/health) but doesn't freeze a contract baseline or change-notification process. (c) BIZ-auth-010 (审计由 eiam 承接) — PRD says "审计归平台" (L40) but doesn't specify that WatchAlert does not self-build audit logging. |

### Dimension 6: Edge Case Coverage — 75/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| 错误路径文档化 | 33/40 | eiam auth fail, eiam unavailable, 403 no-role, session expiry, compile failure, migration stage fail, old credentials 404, redirect URL attack — all covered. **Docked 7**: Cross-process logout failure path undocumented — if a tab's request hits a different process after logout, the in-memory revoke list won't have the entry, so 401 is not guaranteed. This is an **untested error path** that directly undermines a stated goal. |
| 边界条件覆盖 | 25/35 | Grace window 2h ✓, probe 3 consecutive ✓, migration 10%→100% ✓, config diff=0 ✓, rollback ≤2min ✓. **Docked 10**: (a) **Partial config migration failure** — "重导或回滚至并行阶段" (L88) doesn't specify cleanup of partially imported config in the new instance. (b) **Grace-expiry / eiam-recovery race** — if eiam recovers at exactly 2h:00:00, does the user stay in read-only or get redirected to login? (c) **Concurrent logout from multiple tabs** — what if two tabs logout simultaneously? (d) **Frontend startup failure** — only backend compile failure is covered (L81, L128). |
| 失败恢复描述 | 17/25 | Migration rollback detailed ✓, eiam recovery ✓, compile failure → old process continues ✓. **Docked 8**: (a) **Rollback-failure recovery** — "反代回切旧实例（≤ 2 分钟完成）" (L92) doesn't describe what happens if the rollback itself fails (反代不可达? 切流脚本超时?). (b) **Key compromise recovery** — shared HMAC key distributed to all processes; if compromised, rotation process not described. (c) **Revoke list memory growth** — "直至过期" (L195) addresses cleanup, but no bound on list size during high-volume logout bursts. |

### Dimension 7: Scope Clarity — 94/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope 可交付 | 33/35 | Most items are concrete deliverables. **Docked 2**: "作废旧接入决策，新增子项目接入决策记录" (L59) — "决策记录" is vague: where stored? what format? what content? A downstream agent cannot implement "记录一个决策" without knowing the artifact specification. |
| Out-of-scope 显式列出 | 30/30 | Four deferred items named explicitly: historical data migration, portal embedding, eiam enhancement, role sync. Not implied by absence. |
| Scope 与功能规格/故事一致 | 31/35 | In-scope items map to UF and Stories. Out-of-scope "前端嵌入门户" → UI Scope "独立前端" ✓. **Docked 4**: Out-of-scope "权限/角色随 eiam 同步" (L66) says authorization stays in WatchAlert, but the session mechanism contradiction (L78 vs L195) means it's unclear whether WatchAlert's session is part of the platform shared session (which carries platform-level auth context) or independent — scope boundary between platform session and WatchAlert session is blurred. |

---

## Phase 3: Blindspot Hunt

### [blindspot] 1: Read-only unauthenticated endpoint data exposure

**Quote**: "展示只读免登录视图：告警看板 + 规则只读 + 事件历史，无任何写操作与登录入口" (UF-3, L112)

**Issue**: The degradation scenario exposes alarm dashboard, rules, and event history to **unauthenticated** users. Alarm rules may contain sensitive configuration (notification channel credentials, data source connection strings, webhook URLs with secrets). The PRD never specifies what data is safe for anonymous access or what filtering/redaction applies. A downstream agent implementing "后端只读接口" (L123) has no guidance on data classification for anonymous exposure. This is a security blindspot that no rubric dimension fully captures — it spans Functional Specs (data requirements), Edge Case (security boundary), and Scenario Completeness (implicit assumption), but the core issue is an **unstated data classification prerequisite**.

**Resolution required**: Specify which alarm data fields are safe for unauthenticated read-only access; define a redaction/filtering rule for sensitive fields (notification credentials, data source passwords, webhook secrets) in the read-only degradation endpoints.

### [blindspot] 2: Session architecture self-contradiction (Reasoning Audit flagged independently)

**Quote**: "复用平台共享会话" (L78) ↔ "不依赖共享存储读会话。会话票仅存内存" (L195)

**Issue**: The PRD describes two mutually exclusive session architectures in two different sections. The flow description (L78) and In-Scope (L57) claim WatchAlert reuses the platform's shared session (Redis, per BIZ-auth-005). The Security Requirements (L195) describe a self-signed in-memory HMAC ticket with no shared storage and a per-process revoke list. These cannot both be true. If architecture (B) is adopted, single sign-out across processes is broken — a logout from one process adds to that process's in-memory revoke list only, and another process will still validate the ticket via HMAC (which hasn't changed) and return 200 instead of 401. This breaks Story 5's AC and the In-Scope "单点登出" item. The fix for iteration-1 issue #11 introduced this contradiction.

**Resolution required**: Pick one architecture. If shared Redis session (BIZ-auth-005 compliant), remove the in-memory HMAC ticket design and describe how WatchAlert integrates with the platform Redis session. If in-memory HMAC ticket, explain how the revoke list propagates cross-process (e.g., pub/sub, shared cache, sticky routing) and reconcile with BIZ-auth-005.

### [blindspot] 3: Migration config export/import tooling undefined

**Quote**: "导出旧实例告警规则/数据源/通知配置，导入新实例；pass 标准 = 配置条数对齐（旧=N，新=N，diff=0）" (L88)

**Issue**: The migration flow defines a pass criterion (diff=0) but never specifies the export/import mechanism — is it a CLI tool? A SQL dump? An API-based import? A manual admin UI operation? What format is the export? How are data source credentials and notification channel secrets handled during export (are they re-encrypted? re-entered?). A downstream agent cannot implement "导出...导入" without knowing the tooling, format, and secret-handling process. This is an executability gap that falls between Scenario Completeness (implicit assumption) and Scope Clarity (concrete deliverable), but is fundamentally a **missing implementation interface** that no dimension fully scores.

**Resolution required**: Specify the migration tool/interface (CLI/API/SQL), export format, and secret-handling process (re-encryption, manual re-entry, key rotation) for config migration.

### [blindspot] 4: `watchalert_role_binding` population interface undefined

**Quote**: "由告警平台管理员在配置迁移时预置" (L31)

**Issue**: The role binding table is described at the data-model level (`eiam_uid → admin/user`) but the population interface is unspecified. How does the administrator "预置" bindings? Direct SQL insert? An admin API endpoint? A CLI command? A bulk import from a CSV? Without an interface specification, a downstream agent cannot implement the binding management feature, and the administrator cannot perform the pre-population step that the entire login flow depends on (no binding → 403).

**Resolution required**: Specify the role binding management interface (API endpoint, CLI command, or admin UI) and the bulk pre-population method available during config migration.

---

## Cross-Dimension Coherence Check

| # | Finding | Dimensions Affected |
|---|---------|---------------------|
| 1 | Session architecture contradiction (L78 vs L195 vs BIZ-auth-005) | Scenario Completeness, User Stories, Edge Case, Scope Clarity |
| 2 | UF-2 missing recovery state during grace window | Functional Specs, Flow Diagrams |
| 3 | Read-only unauthenticated endpoint data scope | Functional Specs, Scenario Completeness, [blindspot] |
| 4 | "质量门禁覆盖" goal is a capability flag, not a measurable metric | Background & Goals |
| 5 | Performance Requirements "满足现有规模" unquantified despite checklist claim | Background & Goals |

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Background & Goals | 95 | 100 |
| 2. Flow Diagrams | 140 | 150 |
| 3. Functional Specs | 185 | 200 |
| 4. User Stories | 185 | 200 |
| 5. Scenario Completeness | 115 | 150 |
| 6. Edge Case Coverage | 75 | 100 |
| 7. Scope Clarity | 94 | 100 |
| **Total** | **889** | **1000** |

**Target**: 900 | **Gap**: -11

---

## Attack Points (Priority-Ordered)

1. **[Scenario Completeness]** Session architecture contradicts BIZ-auth-005 — "会话票仅存内存（不落盘、不明文缓存）...不依赖共享存储读会话" (L195) vs "复用平台共享会话" (L78) vs BIZ-auth-005 "销毁共享 Redis session" — Must pick one architecture: either use shared Redis session (matching BIZ-auth-005) and remove in-memory HMAC ticket design, or explain how per-process revoke list propagates cross-process for single sign-out to work.

2. **[blindspot]** Read-only unauthenticated endpoint data exposure — "只读看板 + 规则只读 + 事件历史" (UF-3 L112) exposed to unauthenticated users without data classification — Must specify which fields are safe for anonymous access and define redaction rules for sensitive config (notification credentials, data source passwords, webhook secrets).

3. **[User Stories]** Story 5 AC unverifiable under stated session architecture — "其余标签页下次请求时后端返回 401" (L85) cannot be guaranteed if the request hits a different process without the revoke list entry — Must either use shared session store or specify cross-process revoke propagation mechanism.

4. **[Edge Case Coverage]** Cross-process logout failure path undocumented — "登出时加入内存撤销列表直至过期" (L195) is per-process, so a different process will still validate the HMAC ticket — Must document the failure path and recovery when cross-process logout doesn't propagate.

5. **[Functional Specs]** UF-2 missing recovery state during grace window — state table only has "宽限中" and "宽限过期" (L96-97), no "已恢复" state — Must add a recovery state with trigger "探针连续 3 次成功" to remove banner and re-enable write operations during grace window.

6. **[Flow Diagrams]** Single sign-out flow not diagrammed — only session expiry (ExpMid) is in the Mermaid diagram, logout is text-only (L77-78) — Must add a logout flow diagram showing logout → destroy session → other tabs 401 convergence.

7. **[blindspot]** Migration config export/import tooling undefined — "导出旧实例告警规则/数据源/通知配置，导入新实例" (L88) has no tool, format, or secret-handling specification — Must specify migration interface, export format, and credential re-encryption process.

8. **[blindspot]** `watchalert_role_binding` population interface undefined — "由告警平台管理员在配置迁移时预置" (L31) has no interface (SQL/API/CLI/UI) — Must specify the binding management interface and bulk pre-population method.

9. **[Scenario Completeness]** BIZ-auth-009 (eiam API 契约基线) not referenced — PRD depends on eiam endpoints (login, GET /eiam/health) but doesn't freeze a contract baseline — Must reference or establish an eiam endpoint contract baseline with change-notification process per BIZ-auth-009.

10. **[Edge Case Coverage]** Rollback-failure recovery undocumented — "反代回切旧实例（≤ 2 分钟完成）" (L92) doesn't describe what happens if rollback itself fails — Must specify recovery steps when the rollback mechanism fails (反代不可达, script timeout).

11. **[User Stories]** Partial config migration failure not covered in Story 6 — "配置条数对齐（旧=N，新=N，diff=0）" (L99) implies all-or-nothing but doesn't address partially imported config cleanup — Must add AC for partial import failure and cleanup.

12. **[Background & Goals]** "质量门禁覆盖" goal metric is a capability flag — "可一键触发" (L42) doesn't measure quality (coverage %, pass rate) — Must add a measurable quality metric (e.g., unit test coverage ≥ X%, gate pass rate 100%).
