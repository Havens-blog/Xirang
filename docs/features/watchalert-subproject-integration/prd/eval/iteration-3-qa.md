# Iteration 3 QA Scorer Report (Final Round)

**Scorer**: Senior QA Engineer (Adversarial)
**Mode**: A (Feature WITH UI)
**Target**: 900/1000
**Score**: 979/1000

---

## Phase 0: Previous Iteration Issue Verification

All 12 issues from iteration-2-merged.md were checked against the current revision.

| # | Issue | Status | Evidence |
|---|-------|--------|----------|
| 1 (CRITICAL) | 会话架构自相矛盾（进程内 HMAC 票 vs 共享会话） | **FIXED** | Security Requirements: "复用平台共享会话存储（与 BIZ-auth-005 同口径，共享 session 存平台 Redis），不自建独立会话存储、不签发进程内自签名会话票". No mention of "进程内 HMAC 内存票" or "内存撤销列表" anywhere. Consistent across spec, user stories (Story 5 AC), and Dependency Baseline. |
| 2 | 隐性依赖未标现有/新建 + 缺 eiam 契约基线 | **FIXED** | Dependency Baseline table labels all 6 items as 现有/新建. eiam 契约基线 (BIZ-auth-009) documented: "本 feature 依赖的 eiam 端点快照登记入契约基线文件（`parity-checklist.md`）...eiam 侧变更基线内任一端点须提前 ≥ 5 个工作日书面通知并附影响评估". |
| 3 | 回滚失败未覆盖 | **FIXED** | "**回滚失败兜底**：若反代回切本身失败（反代控制面不可达或回切超 > 2 分钟），执行手动 DNS 切换将流量指回旧实例 VIP；切换期间双实例并存且均设只读模式（写操作统一 503）". Story 6 AC also covers this. |
| 4 | 单点登出流程缺 Mermaid 图 | **FIXED** | Mermaid diagram added: `LO([用户在告警平台登出]) --> Dest[销毁平台共享会话 Redis session] --> Other{其余标签页发请求?} --> ...` |
| 5 | UF-2/UF-3 状态缺口与自相矛盾 | **FIXED** | UF-2 added "已恢复" and "正常（探针偶发失败）" states. UF-3 clarified: "恢复时由后端探针信号驱动前端自动跳转 eiam 登录页（非用户点击按钮入口），不引入被移除的本地登录路径". Validation rule: "只读视图不得出现任何登录入口...恢复跳转由后端探针信号驱动，非用户可点击的登录按钮". |
| 6 | Story 6 抽样验证口径未定义 | **FIXED** | "抽样口径：样本量 = max（≥ 10% 规则总数，≥ 50 条），抽样方法 = 随机抽取覆盖各规则级别（P0/P1/P2）与各数据源类型的规则；「一致」定义 = 同一规则用相同输入触发，新旧实例告警产出时间差 ≤ 5 秒且告警内容（级别/标题/接收人）字段完全相同". |
| 7 | 写操作状态码三文档不一致 | **FIXED** | Spec: "写操作返回 503". UI UF-2: "后端返回 503，前端提示不可写". Story 2 AC: "写操作返回 503 且前端提示不可写". All three consistently use 503. |
| 8 | 性能含糊 + 质量门禁不可度量 | **FIXED** | Goals: "启动耗时 ≤ 编译 + 30 秒", "单元测试覆盖率 ≥ 80%". Performance: "P95 ≤ 300ms", "支撑 ≥ 200 并发用户", "P95 回退 ≤ 50ms". All quantified. |
| 9 | role_binding 非法值 + 预置接口未定义 | **FIXED** | "非法角色值处理：role 字段为 null、空串或非 admin/user 值时，一律视为无有效角色，走 403 分支". Binding management API: POST/DELETE/GET endpoints with batch import support. |
| 10 | 只读免登录视图数据分级/脱敏缺失 | **FIXED** | Security Requirements + UF-3 Data Requirements: "匿名可见字段清单：告警规则名称/级别/状态/触发条件、告警事件时间线、事件当前状态" and "敏感字段脱敏规则：通知渠道凭据...一律以 `***` 脱敏不展示". |
| 11 | 迁移导出/导入工具未定义 | **FIXED** | "迁移接口：旧实例 `GET /api/watchalert/export/config`...新实例 `POST /api/watchalert/import/config`". Export format, schema_version, credential re-encryption all documented. |
| 12 | Story 6 部分导入失败未覆盖 | **FIXED** | Story 6 AC: "Given 配置导入过程中部分条目失败...Then 导入中止当前批次，已成功导入的条目保留并标记为「待确认」，失败条目输出错误清单...管理员可选择修复后重导失败条目（增量导入）或回滚至并行阶段". |

**Result**: 12/12 issues fixed. No regressions detected.

---

## Phase 1: Reasoning Audit

### Problem → Solution
Stated problem: WatchAlert isolated (separate source, process, reverse proxy, local auth separate from eiam). Proposed solution: bring source into workspace as subproject, unified startup, unified quality gates, unify auth to eiam, remove local login. The solution directly addresses every facet of the stated problem. No scope substitution detected.

### Solution → Evidence
Evidence is in the form of specific, measurable metrics (P95 ≤ 300ms, ≥ 200 concurrent users, ≥ 80% coverage, 404 for removed endpoints, quarterly upstream PR). These metrics are relevant to the solution and not cherry-picked.

### Evidence → Success Criteria
Goals are quantified and testable: 100% eiam auth, ≤ compile+30s startup, ≥ 80% coverage, 404 + no token for removed endpoints, ≥ 1 PR/quarter + CI green. Each tests what matters, not easy proxies.

### Self-contradiction Check
- Session architecture: Consistently "复用平台共享会话存储" across spec, stories, UI functions, and dependency baseline. No contradiction.
- 403/401 usage: 403 for authenticated-but-no-role (BIZ-auth-003 compliant), 401 for unauthenticated/session-expired, 404 for removed endpoint, 503 for write-during-degradation. Consistent.
- Auth boundary: "eiam 只负责认证" consistent with Out of Scope "权限/角色随 eiam 同步（仅统一认证，授权留在 WatchAlert 内部）". No contradiction.
- Scope ↔ Functional Specs ↔ User Stories: All in-scope items have corresponding UI functions and/or user stories. No orphaned or missing items.

### SC Consistency Deep-Dive (Clustering + Intra-group Satisfiability)

**Cluster: Auth Module / Login Handler**
- SC: 账号统一 (100% eiam), 本地登录后门清零 (404, no token)
- InScope: 登录统一到 eiam + 移除本地登录入口, 角色预置绑定, 降级处理, 单点登出, 会话过期承接
- Bidirectional check: Achieving 100% eiam requires removing local login (InScope item). Removing local login endpoint returns 404 (SC). Role binding enables eiam-based auth. Degradation handles eiam-unavailable edge. All mutually satisfiable. No contradiction.

**Cluster: Platform Orchestration / Quality Gate**
- SC: 一键启动 (≤ compile+30s), 质量门禁覆盖 (≥ 80%)
- InScope: 源码纳入 workspace, 统一启动编排, 统一质量门禁
- Bidirectional check: Unified startup achieves ≤ compile+30s. Quality gate includes unit tests with ≥ 80% coverage. All compatible. No contradiction.

**Cluster: Migration**
- SC: (no direct SC, but Story 6 has migration ACs)
- InScope: 迁移切流 (并行/配置迁移/切流/退役/回滚)
- Bidirectional check: Migration phases are sequential with gates. Rollback branches exist. No internal contradiction. No contradiction.

**Cluster: Session Management**
- InScope: 单点登出, 会话过期承接
- Bidirectional check: Single logout destroys shared session; session expiry redirects to login. Both use shared session store. No conflict. No contradiction.

**New contradictions from revision?** The revision added role binding API, rollback failure handling, probe hysteresis, desensitization rules, and dependency baseline. All additions are internally consistent with existing SC and InScope entries. No new contradictions introduced.

---

## Phase 2: Rubric Scoring

### 1. Background & Goals (100/100)

**Background three elements (30/30)**: Why (WatchAlert isolated, dual accounts, no unified audit), What (subproject integration + eiam auth unification + local login removal), Who (运维/值班工程师, 告警平台管理员, 平台开发者 — all with descriptions and notes). All present and specific.

**Goals quantified (30/30)**: Five goals, all with numeric metrics: 100% eiam, ≤ compile+30s, ≥ 80% coverage, 404+no token, ≥ 1 PR/quarter + CI green.

**Background-Goals consistency (40/40)**: Each goal directly derives from a stated problem: dual accounts → 100% eiam; manual two-step startup → one command ≤ compile+30s; not in quality gates → ≥ 80% coverage; local login backdoor → 404+no token; dead fork risk → ≥ 1 PR/quarter. Fully consistent.

### 2. Flow Diagrams (142/150)

**Mermaid diagram exists (50/50)**: Four Mermaid diagrams: main business flow, single logout, developer startup, migration cutover.

**Main path complete (50/50)**: Main flow covers: user访问 → check auth → eiam login → role check → session → dashboard → degradation → session expiry → re-login. Full happy path from start to dashboard with all key states.

**Decision points + error branches (42/50)**: Multiple diamond nodes (CheckAuth, EiamOK, Role, Deg, Expire, ExpMid) and error branches (auth failure, 403 no-role, degradation banner, session expiry). Migration diagram has 5 stage gates with 3 rollback branches.

**Deduction (-8)**: The main flow diagram is missing the "eiam unavailable for unauthenticated user → read-only view" branch. The diagram shows `CheckAuth -->|否| EiamLogin[eiam 统一登录]`, assuming eiam is always reachable for unauthenticated users. The text explicitly covers this case: "未登录用户直接进入只读视图，提示等待恢复", but the diagram does not show this branch. The degradation check (`Deg{认证服务可用?}`) only triggers from `Dash` (after login), not from the unauthenticated path. This is a gap between the text and the diagram for an important error branch.

### 3. Functional Specs (200/200) — evaluates prd-ui-functions.md

**Placement & Interaction completeness (70/70)**: All 3 UI functions have Mode, Target Page, and Position. User Interaction Flows cover full paths (UF-1: 5 steps from access to dashboard/403; UF-2: 3 steps from detection to write-blocked; UF-3: 3 steps from read-only to recovery-redirect). Page Composition table maps pages to UFs.

**Data Requirements & States clarity (70/70)**: Field tables and state tables filled for all 3 UFs with explicit sources and triggers. UF-1: 4 states (未登录/登录中/登录失败/无权限). UF-2: 4 states including "已恢复" and "正常（探针偶发失败）". UF-3: 2 states plus 匿名可见字段清单 and 敏感字段脱敏规则. All sources and triggers explicit.

**Validation Rules explicit (60/60)**: UF-1: redirect whitelist with specific rules (host match, path prefix `/`, no absolute URLs, no `//`, no `javascript:`/`data:`, fallback to default). UF-2: banner must specify prohibited operations; 1-2 probe failures must not show banner (hysteresis threshold = 3). UF-3: no login entry in read-only view; recovery redirect by backend probe signal, not user button. All actionable, not vague.

### 4. User Stories (196/200)

**Coverage: one story per target user (50/50)**: 运维/值班工程师 → Stories 1, 2, 5. 告警平台管理员 → Stories 3, 6. 平台开发者 → Story 4. All three user types from Background covered.

**Format correct (50/50)**: All stories follow "As a / I want to / So that". Actions are concrete: "用平台统一账号登录", "确认旧本地登录入口已彻底移除", "用平台统一命令一键启动", "将生产流量切流到新实例". No vague "manage" or "handle".

**AC per story (50/50)**: Every story has at least one AC in Given/When/Then format. Story 1: 5 ACs. Story 2: 2 ACs. Story 3: 1 AC with And. Story 4: 1 AC with Ands. Story 5: 1 AC with Ands. Story 6: multiple ACs covering all migration phases.

**AC verifiability & boundary coverage (46/50)**: ACs are generally objectively testable with specific status codes (404, 401, 403, 503), numeric thresholds (≥ 99.9%, P95 ≤ 300ms, ≤ 2 minutes), and verifiable behaviors.

**Deduction (-4)**: Story 4's failure-case AC — "后端编译失败时状态给出显著标记（区别于普通业务服务静默续跑）" — does not specify the form of "显著标记" (exit code? log line? UI indicator?) and does not verify that the old backend process continues running and frontend remains accessible after compile failure. The spec text states "旧进程继续跑" and "拉起后端与前端", but the AC only verifies the status marker, not the continuation behavior. A QA engineer cannot verify "旧进程继续跑" from this AC alone.

### 5. Scenario Completeness (146/150)

**End-to-end scenario coverage (60/60)**: All user-facing scenarios described from trigger to final state: login SSO (all branches), session expiry, single logout, developer startup, migration cutover (5 phases with gates), degradation (grace window → read-only → recovery). Each covers full lifecycle.

**Implicit assumptions surfaced (36/40)**: Dependency Baseline table makes all dependencies explicit (eiam health endpoint, role binding table, shared Redis session, eiam key distribution, migration/export-import endpoints, role binding management API — all labeled 现有/新建). eiam contract baseline (BIZ-auth-009) established.

**Deduction (-4)**: Session continuity during migration cutover is an implicit assumption not explicitly surfaced. During cutover (10% → 100% traffic shift), users shifted from old instance (local WatchAlert auth) to new instance (eiam auth) would need to re-authenticate via eiam — their old local sessions do not carry over to the new eiam-based auth. The document states "反代灰度切流，先 10% 流量观察 15 分钟无 5xx/鉴权失败激增" and "鉴权成功率 ≥ 99.9%", but does not explicitly state that users caught in the cutover may experience re-authentication prompts, nor whether re-auth via eiam redirect counts as "鉴权成功" or "鉴权失败" in the success metric. This implicit assumption affects both the user experience and the metric interpretation.

**Business-rules consistency (50/50)**: Checked against all loaded business rules:
- BIZ-auth-003 (403 for authz denial, stay on page): "返回 403 原地提示...保持当前页不跳登录页". ✓
- BIZ-auth-005 (shared Redis session, 401 convergence, no broadcast): "登出即销毁共享会话，其余进程/标签页下次验会话返回 401 收敛跳 eiam 登录页（不引入实时广播）". ✓
- BIZ-auth-006 (credential error unified message): "统一文案，不泄露用户名是否存在". ✓
- BIZ-auth-008 (session expiry, no white screen, return to path): Covered. ✓
- BIZ-auth-009 (eiam API contract baseline): "端点快照登记入契约基线文件...变更须提前 ≥ 5 个工作日书面通知". ✓
- BIZ-resilience-001 (eiam unavailable, graceful degradation, no white screen): Covered with grace window + read-only mode. ✓
No contradictions with any loaded business rule.

### 6. Edge Case Coverage (95/100)

**Error paths documented (40/40)**: Comprehensive: invalid input (redirect whitelist failure → fallback), permission denied (no role → 403), resource not found (removed endpoint → 404), timeout (eiam probe 3s timeout), eiam unavailable (degradation flow), compilation failure (old process continues), partial import failure (batch abort + error report), rollback failure (manual DNS switch + dual read-only).

**Boundary conditions covered (30/35)**: Illegal role values (null/empty/non-admin/user → 403). Export/import schema_version mismatch → abort. Concurrent dual instances during rollback failure (both read-only, 503 for writes). eiam probe hysteresis (1-2 failures don't trigger, 3 consecutive required).

**Deduction (-5)**: Missing boundary conditions for: (a) `eiam_uid` max-length in role binding — no length constraint or overflow behavior documented; (b) concurrent role binding management — two admins simultaneously setting the same `eiam_uid` to different roles, no conflict resolution specified (unlike BIZ-identity-004 which addresses concurrent edits); (c) very large config migration — no explicit bound on config size or performance expectation for export/import of thousands of rules.

**Failure recovery described (25/25)**: eiam recovery (probe 3x success → remove banner / auto-redirect to login). Migration rollback (reverse proxy rollback ≤ 2min). Rollback failure recovery (manual DNS switch). Partial import failure (fix + incremental re-import or full rollback). Compilation failure (old process continues, developer retries).

### 7. Scope Clarity (100/100)

**In-scope items are concrete deliverables (35/35)**: All 11 in-scope items are specific: source into workspace, unified startup orchestration, unified quality gates, auth to eiam + remove local login, role binding table, degradation handling, reverse proxy adjustment, migration cutover, single logout, session expiry handling, decision record. No vague areas.

**Out-of-scope explicitly lists deferred items (30/30)**: Four named deferred items: historical alert data migration (only config), frontend portal integration (alert stays separate), eiam functional enhancement (minimal changes only), role/permission sync with eiam (auth only, authorization stays internal). All explicitly named, not implied by absence.

**Scope consistent with functional specs and user stories (35/35)**: Every in-scope item maps to corresponding UI functions and/or user stories. No orphaned scope items, no unscoped features described in specs or stories. Fully consistent.

---

## Phase 3: Blindspot Hunt

### [blindspot] Bootstrap of first admin role binding — chicken-and-egg gap

**Quote**: "由告警平台管理员在配置迁移时预置" and "管理员通过 `POST /api/watchalert/role-bindings`（单条/批量 JSON Body：`[{eiam_uid, role}]`）预置绑定"

**Issue**: The role binding management API (`POST /api/watchalert/role-bindings`) requires authentication — but authentication depends on the role binding table being populated. This creates a chicken-and-egg problem: the first admin cannot authenticate to the API to create the first admin's role binding. The document mentions "建表脚本随迁移交付" and "配置迁移时支持批量导入（JSON 文件，格式同批量接口 Body）" which implies a migration-tool-based bootstrap (direct database write, bypassing API auth), but this is never explicitly documented as the bootstrap mechanism. The document does not state whether the batch import during migration bypasses API authentication, nor whether it's a direct database operation or an internal endpoint.

**Improvement required**: Explicitly document the bootstrap process — e.g., "迁移工具在导入配置时同时导入初始角色绑定，通过迁移脚本直接写入数据库（不经过 API 鉴权，因 API 鉴权依赖角色绑定表，存在循环依赖）。初始绑定导入完成后，后续角色管理通过已鉴权的 API 进行。"

---

## Cross-Dimension Coherence Check

- Scope ↔ Solution: In-scope items match the Related Changes table and UI functions. ✓
- Solution ↔ Success Criteria: Each goal has corresponding in-scope items. ✓
- Success Criteria ↔ User Stories: User story ACs verify the goals (Story 3 → backdoor zeroing, Story 4 → startup, Story 6 → migration metrics). ✓
- Status codes: 404/401/403/503 consistent across spec, UI functions, and user stories. ✓
- Session architecture: "复用平台共享会话存储" consistent across all documents. ✓
- eiam contract: Dependency Baseline + BIZ-auth-009 baseline consistent. ✓

No cross-dimension inconsistencies found.

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 100 | 100 |
| Flow Diagrams | 142 | 150 |
| Functional Specs | 200 | 200 |
| User Stories | 196 | 200 |
| Scenario Completeness | 146 | 150 |
| Edge Case Coverage | 95 | 100 |
| Scope Clarity | 100 | 100 |
| **Total** | **979** | **1000** |

## Attack Points (for reviser)

1. [Flow Diagrams]: 主流程图缺"未登录用户 + eiam 不可用 → 只读视图"分支 — "CheckAuth -->|否| EiamLogin[eiam 统一登录]" 假定 eiam 总可达，文本有"未登录用户直接进入只读视图"但图中无此分支 — 主流程图 CheckAuth→否 路径需增加 `{eiam 可用?}` 决策节点，不可用分支指向只读免登录视图
2. [User Stories]: Story 4 编译失败 AC 未验证旧进程续跑与前端可达 — "后端编译失败时状态给出显著标记（区别于普通业务服务静默续跑）" — 补 AC：Given 编译失败 When 启动完成 Then 前端(:3001)仍可访问且旧后端进程继续运行（或明确"显著标记"的具体形式：退出码/日志级别/UI 状态）
3. [Scenario Completeness]: 迁移切流期间会话连续性未显式说明 — "反代灰度切流，先 10% 流量观察 15 分钟无 5xx/鉴权失败激增" — 补说明：切流期间从旧实例（本地认证）转到新实例（eiam 认证）的用户需重新经 eiam 认证；明确"鉴权成功率"是否将 eiam 重定向+成功登录计入"成功"
4. [Edge Case Coverage]: 角色绑定边界条件缺失 — "`watchalert_role_binding`：eiam_uid → admin/user" — 补：eiam_uid 长度约束、并发修改同一 eiam_uid 的冲突处理（last-submitter-wins 或 409）、大批量配置导出导入的性能预期
5. [blindspot]: 首个管理员角色绑定的 bootstrap 流程未显式文档化（鸡生蛋问题） — "由告警平台管理员在配置迁移时预置" + "管理员通过 `POST /api/watchalert/role-bindings`...预置绑定" — 显式说明迁移工具直接写入初始绑定（绕过 API 鉴权，因 API 鉴权依赖绑定表），初始导入完成后后续管理走已鉴权 API
