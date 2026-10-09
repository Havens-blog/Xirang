# PRD Evaluation Report — Iteration 1

**Scorer**: Senior QA Engineer (adversarial, verification stance)
**Rubric**: Mode A (prd-ui-functions.md present), 7 dimensions, 1000 pts
**Date**: 2026-10-09

---

## Phase 1: Reasoning Audit

### Problem → Solution
The stated problem — WatchAlert fragmented as independent directory/process/proxy with its own local accounts, causing dual-account maintenance and unified-audit inability — is addressed by the proposed solution: integrate WatchAlert as a subproject into the platform workspace and delegate authentication to eiam. The alignment is sound at the surface level.

### Solution → Evidence
The PRD provides no evidence that eiam actually supports delegated authentication from WatchAlert. No mention of eiam's existing API capabilities, token issuance mechanism, or session-sharing infrastructure (e.g., shared Redis session per BIZ-auth-005). The bidirectional claim "eiam 只负责认证（你是谁）" assumes eiam can issue tokens that WatchAlert backend can validate — but the validation mechanism (synchronous call? shared secret? JWT?) is never specified.

### Evidence → Success Criteria
- "100% 走 eiam" and "401/404 且不签发 token" are objectively testable.
- "启动耗时 ≤ 编译 + 数十秒" — "数十秒" (tens of seconds) is not a metric; no QA can write a test for "数十秒".
- "上游贡献不中断" — no metric, no test. How would a test verify "仍能向开源上游提交通用修复"?
- "质量门禁覆盖...可一键触发" — "可触发" is a binary check, but "覆盖" is vague: covered to what degree?

### Self-Contradiction Check
1. **"审计归平台" vs "授权留在 WatchAlert 内部"**: Goal says "消除双账号，审计归平台" (audit unified to platform). But BIZ-auth-010 says "平台管理操作审计由 eiam 承接" — WatchAlert's internal authorization actions (configure rules, manage notification channels) are NOT audited by eiam since authorization stays in WatchAlert. The goal of unified audit is only partially achieved (authentication audit unified, authorization audit still in WatchAlert). This is an internal contradiction.

2. **"eiam 下发的 token 仅存内存" vs multi-process architecture**: WatchAlert backend is a separate process. If the token is "仅存内存" — whose memory? The eiam side? The frontend? If the WatchAlert backend needs to validate the token on each request, it must either call eiam synchronously (latency concern, contradicting P95 ≤ 300ms) or hold the token in its own memory (but then how does it get it? and how does it participate in BIZ-auth-005's shared Redis session for single sign-out?).

3. **SC consistency — degradation scope vs resilience rule**: PRD In Scope says "已登录用户在宽限窗口内继续访问，过期后只读免登录". BIZ-resilience-001 says "已登录的 e-cam-web 会话不受影响" — meaning other apps sharing the session should be unaffected. The PRD's degradation affects only WatchAlert, which is consistent, but the PRD never mentions the shared Redis session infrastructure that BIZ-auth-005 and BIZ-resilience-001 both depend on. The implicit assumption that WatchAlert has or will have access to this shared session store is unstated.

---

## Phase 2: Rubric Scoring

### Dimension 1: Background & Goals (82/100)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Three elements present & specific | 28/30 | Reason (Why), Target (What), Users (Who) all present. "平台开发者" role is somewhat tangentially defined — "通过平台统一命令操作" describes how they work, not what they need. Minor. |
| Goals quantified | 21/30 | "100% 走 eiam" ✓ numeric. "启动耗时 ≤ 编译 + 数十秒" — "数十秒" is not a real metric, a QA cannot write a test asserting "tens of seconds". "可一键触发" is binary but "覆盖当前缺失" is vague. "上游贡献不中断" has zero quantification. |
| Background & goals consistent | 33/40 | Goals follow logically from the problem. But "审计归平台" (audit unified) is only partially achieved by "eiam只负责认证" — authorization audit remains in WatchAlert. The goal overpromises what the solution delivers. |

**Deduction**: "启动耗时 ≤ 编译 + 数十秒" — vague language without quantification, -5 pts (applied within quantified criterion). "上游贡献不中断" — no metric at all, -3 pts.

---

### Dimension 2: Flow Diagrams (110/150)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Mermaid diagram exists | 50/50 | Two Mermaid flowcharts present (login + startup). |
| Main path complete | 30/50 | Login flow and startup flow have diagrams. **But 主流程 3 (迁移切流) — the most complex and highest-risk flow — has NO diagram.** Quote: "迁移切流：并行运行、配置迁移、切流、旧实例退役（含回滚方式）" — described only in prose, no visual flow. A QA cannot trace cutover decision points (when to proceed, when to rollback) from the text alone. |
| Decision points + error branches | 30/50 | Login diagram has diamonds (CheckAuth, EiamOK, Deg, Expire) and error branches. But: (a) migration has no decision diamonds at all (no rollback trigger, no verification gate); (b) startup diagram's "编译失败" branch says "旧进程继续跑" but doesn't show what happens to the developer's workflow after — dead-end branch. |

---

### Dimension 3: Functional Specs (120/200) — evaluates prd-ui-functions.md

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Placement & Interaction completeness | 50/70 | All 3 UF have Placement (mode, target page, position). But: UF-2 step 3: "后端返回 503，前端提示不可写" — what does the UI actually show? Toast? Inline error? Disabled button feedback? A downstream agent cannot implement this. UF-3 step 3: "eiam 恢复 → 自动回正" — how does the frontend detect recovery? Polling interval? WebSocket? SSE? Not specified. UF-1 step 4: "回跳登录页展示错误信息" — what error info? Per BIZ-auth-006, error message must be unified anti-enumeration text, but this is not specified. |
| Data Requirements & States clarity | 40/70 | Field tables exist but are thin. UF-2: "宽限剩余时长" source "后端状态" — which endpoint? What response shape? "写操作许可" source "后端状态" — same vagueness. UF-3: "鉴权服务状态" source "后端探针" — what probe? Health check endpoint? Polling frequency? State tables have triggers but triggers reference unspecified mechanisms ("后端探针", "宽限窗口到期" — how is expiry detected?). |
| Validation Rules explicit | 30/60 | UF-1: "不得保留任何本地用户名/密码输入框" — actionable ✓. UF-2: "横幅须明确告知哪些操作禁止" — but the specific prohibited operations are only listed in the interaction flow ("编辑规则/确认告警/改通知"), not in the validation rule itself; the rule says "明确告知" but doesn't enumerate. UF-3: "只读视图不得出现任何登录入口" — actionable ✓. **Missing**: no validation for eiam redirect URL (open redirect vulnerability — the "回跳地址" field in UF-1 Data Requirements takes "当前页 URL" without sanitization). No validation for token format or tampering. |

---

### Dimension 4: User Stories (163/200)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Coverage: one story per target user | 48/50 | 3 user types in Background → 4 stories. 运维工程师 (S1, S2), 告警平台管理员 (S3), 平台开发者 (S4). All covered. |
| Format correct (As a / I want / So that) | 45/50 | All stories follow format. Actions are concrete ("用平台统一账号登录", "一键启动"). S2 "仍能看到告警现状" is slightly abstract but acceptable. |
| AC per story (Given/When/Then) | 42/50 | All stories have AC in Given/When/Then. S4 uses "And" clauses for additional conditions — acceptable. |
| AC verifiability & boundary coverage | 28/50 | **S1 AC covers only happy path** — no error AC (eiam authentication failure, eiam unreachable, session expired). **S2 AC**: "宽限期内" — duration unspecified; a QA cannot write a boundary test (e.g., "at 1 second before expiry, write is still blocked; at expiry, view degrades") without knowing the window length. **S3 AC** is testable ✓. **S4 AC**: "状态/日志可查" — how verified? What constitutes "可查"? No negative test: what happens when quality gate (lint/static-check/unit-test) fails? The AC says "且通过" but doesn't describe failure behavior. **No story covers migration/cutover from operator perspective** — a critical lifecycle event has no user story. |

**Deduction**: Cross-section inconsistency — migration cutover is In Scope but has no corresponding user story, -5 pts (within verifiability criterion).

---

### Dimension 5: Scenario Completeness (67/150)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| End-to-end scenario coverage | 35/60 | Login, startup, migration, degradation flows described. But migration cutover is high-level only: "同步告警规则/数据源配置到新实例" — how? Export/import? API sync? Manual config copy? "观察一个告警周期无遗漏" — **"告警周期" is undefined** (1 hour? 24 hours? one full incident lifecycle?). No scenario for concurrent user access during cutover (can users log in to both instances?). No scenario for cutover verification failure. |
| Implicit assumptions surfaced | 12/40 | **Unstated prerequisites**: (a) eiam can accept delegated auth from WatchAlert — not confirmed as a capability; (b) WatchAlert internal role system can accept eiam-provided identity (username/unique ID) as input — mechanism unstated; (c) shared Redis session infrastructure (BIZ-auth-005, BIZ-resilience-001 both depend on it) — PRD never mentions Redis or shared session store; (d) token validation mechanism between eiam and WatchAlert backend — unstated; (e) network path between WatchAlert and eiam (PRD says "平台内网 + 双向认证" but doesn't confirm mTLS infrastructure exists). These are critical environmental dependencies that should be surfaced as pre-conditions. |
| Business-rules consistency | 20/50 | **BIZ-auth-005 (双向单点登出)**: PRD addresses SSO login but completely ignores single sign-out. If a user logs out from the platform console, does the WatchAlert session also get destroyed via shared Redis session? Quote: "eiam 只负责认证（你是谁）" — but session lifecycle (creation, validation, **destruction**) is not addressed. This contradicts BIZ-auth-005's mandate. **BIZ-auth-008 (会话过期跳转登录页不白屏)**: Not addressed — what happens when a WatchAlert session expires? Does it redirect to eiam login or white-screen? **BIZ-auth-003 (403原地提示不跳转登录页)**: WatchAlert endpoints may return 403 (user authenticated by eiam but lacks WatchAlert internal role) — behavior undefined. PRD's login flow diagram shows only 401 (unauthenticated) path, not 403 (unauthorized) path. **BIZ-resilience-001**: Aligned ✓. |

---

### Dimension 6: Edge Case Coverage (34/100)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| Error paths documented | 18/40 | eiam不可用 degradation is well-covered. But: **eiam authentication failure** (wrong credentials) — only mentioned in UF-1 step 4, no AC covers it. **Network timeout during eiam redirect** — not covered. **Token expired mid-session** — not covered (no refresh mechanism described). **Concurrent eiam login from multiple tabs/devices** — not covered. **Migration config sync failure** — "同步告警规则/数据源配置" but what if sync fails partway? Not covered. **eiam returns malformed response** — not covered. |
| Boundary conditions covered | 8/35 | **宽限期 duration**: unspecified numerically — cannot test boundary (at T-1s before expiry vs at T+1s after). **eiam flapping** (rapidly available/unavailable): not covered — would cause rapid state switching between banner/normal. **Max concurrent users during degradation**: not covered. **Empty alarm dataset in read-only view**: not covered. **Large alarm dataset in degraded read-only mode**: performance impact not addressed. **Max-length redirect URL** in UF-1 "回跳地址": not covered (potential open redirect / buffer issue). **Zero users logged in when eiam goes down**: not explicitly distinguished from "already logged in" case. |
| Failure recovery described | 8/25 | "eiam 恢复后自动回正" — described but **no probe mechanism**: how does the system detect eiam recovery? Polling interval? Health check endpoint? What if eiam appears recovered but is partially functional? **Migration rollback**: "含回滚方式" mentioned in scope but **rollback steps never described** — what triggers rollback? What are the steps? How is data consistency ensured after rollback? |

---

### Dimension 7: Scope Clarity (76/100)

| Criterion | Score | Rationale |
|-----------|-------|-----------|
| In-scope items concrete deliverables | 26/35 | Most items are specific. But: "迁移切流：并行运行、配置迁移、切流、旧实例退役（含回滚方式）" — "回滚方式" is mentioned but not defined; this is a scope item that promises something it doesn't specify. "作废旧接入决策，新增子项目接入决策记录" — vague: what decision records? Where? Format? |
| Out-of-scope explicitly lists deferred items | 28/30 | 4 items clearly listed and named: history data migration, frontend embedding, eiam enhancement, role sync. Good. |
| Scope consistent with functional specs and user stories | 22/35 | **Inconsistency**: Scope includes "迁移切流" (migration cutover) but **no user story covers migration** — the operator/admin perspective on cutover is absent from user stories. Scope mentions "eiam 不可用时的降级" but the **宽限期 duration is undefined** — scope item doesn't match functional spec detail (functional spec references "宽限窗口" but never defines its length). Scope mentions "生产反代入口调整" but no UI function or user story covers it (acceptable as infra, but creates a gap in traceability). |

---

## Phase 3: Blindspot Hunt

### [blindspot] B1: Single sign-out completely unaddressed — security hole
**Quote**: "eiam 只负责认证（你是谁）；登录后能干什么（看告警/配置规则/管通知）仍由 WatchAlert 内部 admin/user 角色决定"
**Issue**: The PRD addresses SSO login but completely ignores single sign-out. BIZ-auth-005 mandates "任一侧登出即销毁共享 Redis session，其余应用在下一次请求时 401 收敛". The PRD never mentions Redis session, shared session store, or logout behavior. If a user logs out from the platform console, the WatchAlert session behavior is undefined — it could persist indefinitely, creating a security hole where a "logged out" user retains access to the alerting platform. This is a fundamental session-lifecycle gap that the rubric's "Scenario Completeness" dimension only partially captures.
**Improvement required**: Add a logout/single sign-out flow (both directions: console→WatchAlert and WatchAlert→console), specify the shared session mechanism (Redis), and add user story + AC for single sign-out.

### [blindspot] B2: Token storage contradiction — "仅存内存" is architecturally ambiguous
**Quote**: "eiam 下发的 token 仅存内存，不落盘、不明文缓存"
**Issue**: WatchAlert backend is a separate process/service from eiam. The PRD says token "仅存内存" but never specifies whose memory. If the token lives only in eiam's memory, how does WatchAlert backend validate it on each request (it would need to call eiam synchronously, contradicting P95 ≤ 300ms under load)? If the token lives in the browser, how does the backend validate it without a shared secret or public key? If it lives in WatchAlert backend memory, how is it distributed across multiple backend instances? And how does this interact with BIZ-auth-005's shared Redis session? The statement is security-positive in intent but architecturally incoherent as written.
**Improvement required**: Specify the exact token flow: who issues, who stores, who validates, what validation mechanism (JWT signature verification? introspection endpoint? shared Redis session lookup?), and how multi-instance WatchAlert backend handles it.

### [blindspot] B3: Reasoning audit — "审计归平台" goal not fully achievable by the stated solution
**Quote**: "消除双账号，审计归平台" (Goal) vs "授权留在 WatchAlert 内部" (Scope, Out of Scope: "权限/角色随 eiam 同步（仅统一认证，授权留在 WatchAlert 内部）")
**Issue**: The goal promises unified audit ("审计归平台"), but the solution explicitly keeps authorization in WatchAlert. BIZ-auth-010 says "平台管理操作审计由 eiam 承接" — but WatchAlert's internal authorization actions (rule configuration, notification channel management, alert acknowledgement/silencing) are performed by WatchAlert's own role system, NOT by eiam. These actions would be audited by WatchAlert, not by eiam, meaning audit is NOT unified to the platform — only authentication audit is. The goal overpromises what the architecture delivers.
**Improvement required**: Either (a) revise the goal to "认证审计归平台，授权审计留在告警子系统" (accurate but weaker), or (b) add a mechanism for WatchAlert to forward authorization audit events to the platform's audit system.

### [blindspot] B4: Migration cutover has no rollback criteria or verification gates
**Quote**: "迁移切流：并行运行、配置迁移、切流、旧实例退役（含回滚方式）" and "观察一个告警周期无遗漏 → 旧实例保留只读两周后退役"
**Issue**: "含回滚方式" is promised but never delivered — no rollback trigger conditions, no rollback steps, no data-consistency guarantees after rollback. "观察一个告警周期" is undefined (what constitutes a "周期"?). There are no verification gates between cutover phases (when is it safe to proceed from "parallel run" to "config sync" to "traffic switch"?). A QA cannot write a test for "safe to retire old instance" because the success criteria for each phase are undefined. This falls between Scope Clarity and Scenario Completeness dimensions but is fundamentally a reasoning gap.
**Improvement required**: Define explicit cutover phase gates with pass/fail criteria, define "告警周期" numerically, describe rollback trigger conditions and steps, and specify verification criteria for retiring the old instance.

### [blindspot] B5: 403 (unauthorized) path absent from login flow diagram
**Quote**: Login flow diagram: "EiamOK -->|是| Session[建立会话]" → "Dash[告警看板（按内部角色使用）]"
**Issue**: The flow assumes that once authenticated by eiam, the user enters the dashboard. But BIZ-auth-003 requires 403 (越权) to be handled with "原地提示不跳转登录页". A user authenticated by eiam but lacking WatchAlert internal role (e.g., a platform admin with no WatchAlert role) would hit a 403 — but the diagram has no 403 branch. This means the diagram conflates authentication success with authorization success, hiding the 403 edge case. This is not captured by any rubric dimension (Flow Diagrams checks "decision points + error branches" but the 403 is an authorization error, not an authentication error).
**Improvement required**: Add a 403 branch to the login flow diagram after session establishment: "Session → RoleCheck{has WatchAlert role?} →|no| 403原地提示 →|yes| Dash".

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Background & Goals | 82 | 100 |
| 2. Flow Diagrams | 110 | 150 |
| 3. Functional Specs | 120 | 200 |
| 4. User Stories | 163 | 200 |
| 5. Scenario Completeness | 67 | 150 |
| 6. Edge Case Coverage | 34 | 100 |
| 7. Scope Clarity | 76 | 100 |
| **Total** | **652** | **1000** |

Target: 900/1000. **NOT MET** (652 < 900).
