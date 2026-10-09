# Proposal Evaluation — Iteration 2 (Adversarial Scorer: CTO Persona)

**Document**: `D:\Haven\docs\proposals\watchalert-subproject-integration\proposal.md`
**Rubric**: `proposal.md` (1000 pts, 10 dimensions)
**Scorer**: CTO expert persona — adversarial verification stance
**Date**: 2026-10-09
**Mode**: Annotated blind review (Pre-Revision markers present) — Round 2

---

## Iteration 1 Attack Point Resolution Audit

All 16 attack points from iteration-1.md were checked against the revised proposal. Summary:

| # | Iteration 1 Attack | Status | Evidence |
|---|-------------------|--------|----------|
| 1 | D1: Problem statement slightly abstract ("编排纳入") | Partially resolved | Evidence section still clarifies; minor residual |
| 2 | D1: Evidence self-referential, no metrics | Partially resolved | Urgency now quantified (15-20 commits, 2-4/month, 4-8 hr/month); still no user feedback or alert volume |
| 3 | D1: Urgency unquantified | **Resolved** | "fork 与上游已分叉约 15-20 个 commit…每延迟一月预计新增 2-4 个分叉 commit…迁移成本增加约 4-8 人时" |
| 4 | D2: Auth approach deferred to tech-design | Improved | Decision matrix (OIDC→LDAP→桥接) with sequential arbitration; still pending spike |
| 5 | D2: Degradation UX vague | **Resolved** | Detailed: grace window 4hr, read-only + write 503, new users → read-only view, banner with countdown |
| 6 | D2: No migration transition description | **Resolved** | "迁移过渡（切流）" scenario with progressive cutover, decommission criteria |
| 7 | D2: Auth direction unresolved (3 unordered paths) | Improved | Sequential decision matrix with fallback ordering and "同一时刻只采用一条" |
| 8 | D3: No industry products named | **Resolved** | Grafana (auth.generic_oauth) and Alertmanager (反代+IAM网关) specifically referenced |
| 9 | D3: Trade-off cherry-picked | **Resolved** | Comparison table now lists auth complexity, dev.ps1 gap, dead fork risk as cons |
| 10 | D4: Missing migration/rollback scenario | **Resolved** | Migration transition scenario + Rollback & Upstream Cadence section |
| 11 | D4: Missing performance/observability/maintainability NFRs | **Resolved** | Performance (≤300ms P95), observability (metrics + self-alert avoidance), security (memory-only token, mTLS) |
| 12 | D6: No timeline | **Resolved** | 4-phase timeline with person-day estimates (P1: 2-3d, P2: 3-5d, P3: 4-6d, P4: 2-3d + 2wk) |
| 13 | D7: nginx item vague | **Resolved** | Specific proxy_pass changes, :3001 entry, dev.ps1 status proxy health |
| 14 | D9: SC coverage gap (nginx, compile-fail) | **Resolved** | SC for nginx dev entry (:9001/:3001 reachable, status shows proxy health); SC for compile-fail marking |
| 15 | D10: NFR security vs solution gap | **Resolved** | Token in memory only, no disk, mTLS, pluggable interface as validation callback |
| 16 | BS-1: No rollback plan | **Resolved** | Rollback & Upstream Cadence section with runbook (trigger conditions, 4 steps, data preservation) |
| BS-2 | Migration transition undefined | **Resolved** | Progressive cutover: stop old write → sync config → nginx switch → observe 1 cycle → old read-only 2wk → decommission |
| BS-3 | Degradation only covers existing token holders | **Resolved** | "新/未登录用户…直接落到该只读免登录视图" |
| BS-4 | No upstream release cadence consideration | **Resolved** | Upstream cadence analysis, pinning strategy (track tag not HEAD), rebasing plan |
| BS-5 | Vague language without quantification | **Resolved** | Quantified throughout: 15-20 commits, 2-4/month, 4-8 hr, 300ms P95, 4hr grace, 5% failure, 2 weeks |

**Verdict**: All 16 attack points addressed. 12 fully resolved, 4 partially/improved. No revision-introduced contradictions found. New issues are minor and detailed below.

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

### Argument Chain Trace

1. **Problem → Solution**: Problem is dual — (a) WatchAlert outside workspace orchestration, (b) auth fragmented (local JWT vs eiam). Solution addresses both: clone forks as subprojects + four-piece set, and delegate auth to eiam with JWT removal. **No gap.**

2. **Solution → Evidence**: Evidence for the problem is strong (file references, decision 0002, PR numbers, quantified fork divergence). Solution feasibility for workspace parts is well-supported by existing patterns. Auth integration now has a decision matrix (OIDC→LDAP→桥接) with sequential fallback — better than iteration 1's 3 unordered paths. The 桥接 fallback (eiam adds token validation endpoint) is the most certain to be feasible since it's a new endpoint eiam can build. **Pre-score anchor: the auth integration is still an unverified hypothesis, but the fallback structure provides resilience — the worst case (all 3 channels fail) triggers rollback, which is now documented.**

3. **Evidence → Success Criteria**: SC now covers nginx (SC: :9001/:3001 via nginx, status shows proxy health) and compile-fail marking (SC: dev.ps1 status shows COMPILE-FAILED). SC-6↔SC-7 dependency now explicitly gated ("须在 SC-6 验证通过后执行"). **Pre-score anchor: migration/cutover success has no formalized SC — decommission criteria are in the scenario description but not in the SC list.**

4. **Self-contradiction check:
   - The Scope 判定线 (pre-revised) still uses "可能归入本提案" for eiam changes — scope uncertainty persists but is bounded by the 判定线.
   - The degradation path ("不回退到本地 JWT") and the rollback runbook ("重新启用 WatchAlert 本地 JWT") are NOT contradictory — degradation handles temporary eiam unavailability (token cache), rollback handles fundamental migration failure (JWT re-enabled via git revert). The distinction is operationally sound.
   - No new contradictions introduced by the revision.

### SC Consistency Deep-Dive

**Cluster: Auth/鉴权 area** (SC-6 eiam SSO, SC-7 JWT removal, InScope "鉴权统一到 eiam")

- SC-6 ↔ SC-7: Now explicitly gated — SC-7 is labeled "(须在 SC-6 验证通过后执行)" and SC-6 is "(门控前置)". Bidirectional derivation: assuming SC-6 satisfied (eiam login works), SC-7 can proceed. Assuming SC-7 executed after SC-6 verification, no contradiction. **The iteration-1 ambiguity is resolved. No contradiction.**
- SC-6 ↔ InScope "鉴权统一到 eiam: 移除本地 JWT 登录入口": Consistent — scope item enables SC.

**Cluster: dev.ps1 area** (SC-3 dev.ps1 start, SC compile-fail marking, InScope dev.ps1 items)

- SC-3 ↔ InScope "dev.ps1 服务表登记…含对 cmd/ 结构的自定义 build 命令支持": Consistent — custom build command is prerequisite.
- SC compile-fail marking ↔ InScope "dev.ps1 status 对 watch-alert 编译失败须给出显眼标记": **Consistent — SC now verifies the InScope deliverable. Iteration-1 coverage gap resolved.**

**Cluster: nginx area** (SC nginx dev entry, InScope "nginx dev 入口调整")

- SC nginx ↔ InScope nginx: **Consistent — SC now verifies :9001/:3001 via nginx + status proxy health. Iteration-1 coverage gap resolved.**

**Cluster: migration area** (Key Scenario "迁移过渡", P4 timeline, rollback runbook)

- No corresponding SC for migration/cutover success. **Coverage gap — decommission criteria exist in scenario description but are not formalized as SC.**

No mutual-exclusion contradictions found. No new contradictions introduced by revision.

---

## Phase 2 — Rubric Scoring (Verification Stance)

### Dimension 1: Problem Definition — 100/110

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Problem stated clearly | 38 | 40 | Problem is unambiguous: WatchAlert runs as standalone reverse-proxy service outside workspace orchestration, auth fragmented between local JWT and eiam. Two readers would interpret the core problem consistently. Minor deduction: "编排纳入" is slightly abstract — the Evidence section clarifies it means dev.ps1 + Justfile + forge, but the Problem statement itself could name these explicitly. |
| Evidence provided | 34 | 40 | Strong file-level evidence: references decision 0002, dev.ps1 service registry, Justfile recipe gaps, .forge/config.yaml surfaces, upstream PRs (#257/#263/#264). Urgency now quantified ("fork 与上游已分叉约 15-20 个 commit"). But evidence remains self-referential (internal docs and file inspections). "WatchAlert 已承载真实告警（CDN/ECS 指标、告警规则、故障中心）" — no data on alert volume, user count, or operational pain incidents. No user feedback cited. |
| Urgency justified | 28 | 30 | Now well-justified: quantified fork divergence ("15-20 个 commit", "每延迟一月预计新增 2-4 个分叉 commit", "迁移成本增加约 4-8 人时"), security/audit concerns from account fragmentation, quality gate gap. Minor deduction: no external deadline or business pressure cited — urgency is entirely internal technical debt. |

### Dimension 2: Solution Clarity — 104/120

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Approach is concrete | 36 | 40 | Very concrete for workspace integration: specific paths (D:\WatchAlert → Haven/watch-alert), ports (:9001/:3001), deliverables (dev.ps1 service table, Justfile recipes, forge surfaces). Auth approach now has a decision matrix (OIDC→LDAP→桥接) with sequential arbitration and "同一时刻只采用一条" — a concrete framework even though final selection is deferred to tech-design spike. Minor deduction: the actual auth channel remains unselected pending spike results. |
| User-facing behavior described | 40 | 45 | Happy paths described (SSO login "凭 eiam 账号登录", dev startup "一键拉起", quality gate). Degradation UX now detailed: grace window with countdown banner ("鉴权服务暂不可用，已切换宽限模式（token 剩余 X 小时）"), read-only allowed ops, write ops return 503, expired → read-only no-login view, new users → read-only view with "请联系管理员". Migration transition UX described (progressive cutover, old instance read-only 2 weeks). Minor deduction: the degradation exit condition (when does the system return to normal mode after eiam recovers?) is not described. |
| Technical direction clear | 28 | 35 | Clear for workspace parts: git clone, dev.ps1 Daemon, cmd/ build command, forge surface. Auth direction now has a sequential decision matrix (better than iteration 1's 3 unordered paths), but the actual implementation direction is still pending tech-design spike results. The decision matrix is a selection framework, not a direction — the proposal cannot tell the reader which auth protocol will be used. |

### Dimension 3: Industry Benchmarking — 103/120

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Industry solutions referenced | 32 | 40 | Now references specific products: Grafana (内置 generic OAuth2/OIDC provider, auth.generic_oauth 配置直连外部 IdP) and Alertmanager (无内置鉴权, 反代层 nginx/oauth2-proxy + 平台 IAM 网关收口身份). Git patterns (submodule, subtree, fork+clone) cited. These are specific and relevant to the problem domain. Minor deduction: no mention of how other multi-repo workspaces or monorepo-management tools handle fork integration at scale (e.g., how organizations manage upstream sync for forked components). |
| At least 3 meaningful alternatives | 28 | 30 | Five alternatives including "do nothing": (1) do nothing, (2) reverse-proxy + orchestration only, (3) git submodule, (4) git subtree, (5) clone fork. Each genuinely different. Minor deduction: submodule and subtree are somewhat pre-disqualified by the workspace nature established in the Problem section, reducing their value as genuine comparisons. |
| Honest trade-off comparison | 22 | 25 | Comparison table now has honest cons for the selected approach: "鉴权对接复杂度需 tech-design 收敛（见决策矩阵）；dev.ps1 自定义 build 命令字段当前缺失需新增；需持续向上游提通用 PR 以规避死 fork". Much improved from iteration 1's "需遵守向 fork 同步纪律". Minor deduction: the migration risk (moving a live alert system) is not listed as a con of the selected approach, though it is covered in the Risk table. |
| Chosen approach justified | 21 | 25 | "与现有工作区形态一致且最简" is valid rationale — consistency with existing pattern (eiam/opsagent already work this way). Auth approach justified against Grafana/Alertmanager patterns. Minor deduction: doesn't discuss whether the existing workspace pattern has known limitations that this proposal would inherit (e.g., no automated cross-repo dependency checking, manual sync discipline). |

### Dimension 4: Requirements Completeness — 96/110

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Scenario coverage | 35 | 40 | Happy path (dev startup, SSO, quality gate, upstream contribution), edge cases (fork divergence, eiam unavailable, Windows cmd/ compile), migration transition (progressive cutover with decommission criteria), degradation (token cache, read-only, new users). Good coverage. Missing: (1) what happens to currently logged-in WatchAlert users during auth cutover — the migration scenario says "配置为文件/DB 导出导入" for rules/config but doesn't mention user session/token migration; (2) degradation exit condition — when does the system return to normal after eiam recovers? |
| Non-functional requirements | 33 | 40 | Four NFRs now present: security (token in memory only, no disk, mTLS, pluggable interface as validation callback, no credentials in config/logs), performance (≤300ms P95, degradation trigger), observability (auth success rate, token validation latency P95/P99, degradation trigger count/duration, self-alert avoidance), compatibility (go build, frontend build, no regression). Good coverage. Deduction: the performance NFR's degradation trigger conflates "eiam slow" (P95 > 300ms) with "eiam unavailable" — the same heavy degradation (write 503, 4-hour read-only mode) applies to both, but they have different severity profiles. A 350ms latency doesn't warrant blocking all writes for 4 hours. No degradation exit condition specified. |
| Constraints & dependencies | 28 | 30 | Well documented: eiam auth API dependency (with readiness caveat and Scope 判定线), Windows dev environment constraint (cmd/ structure vs dev.ps1 build command), fork remotes, dev.ps1 custom build command gap explicitly flagged. The Scope 判定线 helps bound eiam changes. Minor deduction: the "可能归入本提案" language for eiam changes keeps the dependency scope uncertain — is the eiam token validation endpoint a dependency to be built (in scope) or an external dependency (not ready)? |

### Dimension 5: Solution Creativity — 60/100

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Novelty over industry baseline | 22 | 40 | The workspace insight (fork = git repo, clone as subproject) is practical but standard multi-repo workspace practice — the proposal acknowledges this ("与 eiam/e-cam-service 同构"). Auth approach (委托统一 IdP) is industry-standard SSO matching Grafana's pattern. The degradation path (token cache + read-only + write 503 + new users → read-only view instead of lockout) is a somewhat novel middle ground — most systems either lock out new users or provide emergency access; providing read-only access to unauthenticated users during outage is a thoughtful choice. Minor differentiation from industry baseline beyond "applying standard patterns correctly." |
| Cross-domain inspiration | 15 | 35 | Limited but present. The "避免 WatchAlert 自告自警" (avoid self-alerting — alert on WatchAlert auth degradation is carried by opsagent, not WatchAlert itself) shows awareness of monitoring anti-patterns crossing from observability best practices into alert system design. The degradation path's token cache + grace window echoes circuit-breaker/cache-stale patterns, though these connections are not explicitly named. No broader cross-domain borrowing (e.g., from OAuth offline-access patterns, CDN stale-while-revalidate, or distributed systems fail-safe patterns). |
| Simplicity of insight | 23 | 25 | The core insight remains elegant: "fork 本身就是 git 仓库，直接 clone 下沉为同级子项目" — why overcomplicate with subtree when the workspace is already multi-repo? The "why didn't I think of that" quality is present. Minor deduction: the auth degradation + rollback + upstream pinning + migration transition add significant operational complexity that somewhat undermines the "simple" narrative — the solution is more complex than the core insight suggests. |

### Dimension 6: Feasibility — 80/100

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Technical feasibility | 34 | 40 | Workspace integration clearly feasible — proven patterns (eiam/opsagent already work this way). eiam has auth capabilities, WatchAlert has OIDC/LDAP login. The decision matrix (OIDC→LDAP→桥接) provides fallback resilience: even if OIDC and LDAP fail, the 桥接 option (eiam adds a token validation endpoint) is a new endpoint eiam can build — guaranteed feasible with eiam-side development. The proposal cannot confirm which auth channel will work until P2 spike, but the fallback structure ensures at least one path is buildable. |
| Resource & timeline feasibility | 22 | 30 | Now has 4-phase timeline with person-day estimates: P1 (2-3d), P2 (3-5d), P3 (4-6d), P4 (2-3d + 2wk observation). Total: 11-17 person-days + 2 weeks observation. Reasonable for a single developer. Deduction: the P2 spike (3-5d) could easily expand if eiam needs changes — the timeline doesn't account for the eiam dependency readiness gap. If eiam's token validation endpoint (桥接 path) is needed, eiam-side development adds to the timeline but isn't sized. No calendar dates provided. |
| Dependency readiness | 24 | 30 | Fork remotes ready (Havens-blog/e-watch-alert, Havens-blog/WatchAlert-web). Two critical dependencies NOT ready: (1) eiam auth API — "需确认（是否有可复用的登录校验端点、是否需扩一个给下游应用用的验证接口）"; (2) dev.ps1 custom build command field — "当前版本不存在…需作为显式前置开发任务新增". Both have clear remediation paths (P1 prerequisite task, Scope 判定线 for eiam). Better than iteration 1 because remediation is concrete. Deduction: two of three dependency categories are unready. |

### Dimension 7: Scope Definition — 72/80

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| In-scope items are concrete | 28 | 30 | Most items are concrete deliverables: clone to specific paths, register specific ports, add specific recipes, remove JWT handler/keys/frontend entry. nginx entry now specific (proxy_pass change, :3001 entry, dev.ps1 status proxy health). Minor deduction: "更新/作废决策 0002，新增子项目接入决策记录（含向 fork 同步/提 PR 的约定）" is slightly vague — what specific content goes in the new decision record beyond "fork 同步/PR 约定"? |
| Out-of-scope explicitly listed | 24 | 25 | Clearly listed: data migration, deep kernel changes, frontend embedding, eiam enhancement. The Scope 判定线 helps bound eiam changes. Minor deduction: "WatchAlert 深度内核改造（除鉴权衔接所需的最小通用改动外）" — "最小通用改动" remains subjective and could be a scope creep vector. |
| Scope is bounded | 20 | 25 | Timeline now provides bounding (4 phases, person-day estimates, 2-week observation period as natural endpoint). But: (1) the eiam dependency uncertainty ("可能归入本提案") keeps scope somewhat fluid — if the 桥接 path is needed, eiam-side development expands scope; (2) P4 phase (切流与退役) is in the timeline but its deliverables (cutover plan, rollback runbook drill) are not in the In Scope list — see blindspot BS-1. No calendar dates. |

### Dimension 8: Risk Assessment — 77/90

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Risks identified | 27 | 30 | 9 risks identified, all meaningful: fork sync, eiam auth protocol, eiam unavailability, dev.ps1 cmd/ compile, auth code fork risk, frontend inconsistency, compile failure false positive, upstream contract change, migration rollback. Good coverage. Deduction: no risk identified for the scenario where all 3 auth channels fail (OIDC, LDAP, and 桥接 all prove infeasible). The rollback runbook would handle the outcome, but the proposal doesn't acknowledge "auth unification fundamentally fails" as a risk outcome — the decision matrix is presented as if one channel will work. |
| Likelihood + impact rated | 24 | 30 | Ratings: M/H, M/H, M/H, H/M, M/M, M/L, M/M, L/M, L/H. Now has two L risks (upstream contract: L/M, migration rollback: L/H) — better discrimination than iteration 1 which had only one L. The migration rollback at L/H is honest (low likelihood, high impact). Deduction: three M/H risks still cluster; the eiam auth protocol risk (M/H) seems like it should be H/H given it blocks P2/P3 and the proposal cannot confirm feasibility until spike. |
| Mitigations are actionable | 26 | 30 | Most mitigations are now actionable. The rollback runbook provides concrete steps (trigger conditions, 4-step revert, data preservation). The upstream pinning strategy is concrete (track tag, rebase plan with contract regression). The dev.ps1 custom build command is a concrete P1 task. The degradation path is a concrete design (token cache + read-only + write 503). Deduction: "文档化「改动前先 pull 上游、通用修复直推 fork」纪律" is still documentation-only without automated guards (e.g., pre-push hook checking upstream sync status). "tech-design 阶段先摸清 eiam 认证端点" is investigation, not risk mitigation. |

### Dimension 9: Success Criteria — 71/80

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Criteria are measurable and testable | 28 | 30 | Most are objectively testable: git log/remote check, git push mine, dev.ps1 start, just recipes, forge quality-gate, nginx reachable + status proxy health, compile-fail marking, JWT endpoint returns 401/404, eiam SSO login + P95 ≤ 300ms. Good measurability. Minor deduction: SC-6 ("用 eiam 账号登录 WatchAlert 成功（SSO）") depends on the unverified eiam auth assumption — if the assumption fails, this SC cannot be satisfied, but the SC doesn't flag this dependency explicitly (the gating label on SC-7 helps but SC-6 itself doesn't note its own dependency on P2 spike outcome). |
| Coverage is complete | 18 | 25 | Previous coverage gaps (nginx, compile-fail) now have SCs — resolved. But: (1) migration/cutover success has no formalized SC — the decommission criteria ("连续 2 周零告警丢失/误报、eiam 登录稳定、回滚 runbook 已演练") are in the Key Scenarios description, not in the SC list; (2) the rollback runbook's trigger condition ("新实例登录链路连续失败率 > 5%") introduces a metric with no corresponding SC or monitoring definition — how is "连续失败率" measured and over what window? |
| SC internal consistency | 25 | 25 | SC-6↔SC-7 now explicitly gated: SC-7 labeled "(须在 SC-6 验证通过后执行)", SC-6 labeled "(门控前置)". Bidirectional derivation passes: assuming SC-6 satisfied (eiam login works), SC-7 can proceed; SC-7 executed after SC-6 verification creates no contradiction. The iteration-1 ambiguity is fully resolved. No mutual-exclusion contradictions found across all SC clusters. |

### Dimension 10: Logical Consistency — 77/90

| Criterion | Score | Max | Justification |
|-----------|-------|-----|---------------|
| Solution addresses the stated problem | 32 | 35 | Solution directly addresses both problem dimensions: workspace integration (clone + four-piece set) and auth fragmentation (eiam delegation + JWT removal). Auth solution has decision matrix + degradation + rollback — well-rounded. Minor deduction: the auth solution's feasibility is still unverified (deferred to P2 tech-design spike), so there's a gap between "solution addresses problem" and "solution will definitely work as described." |
| Scope ↔ Solution ↔ Success Criteria aligned | 24 | 30 | Previous gaps (nginx SC, compile-fail SC) now resolved. SC-6↔SC-7 gating aligns scope and SC. Deduction: P4 phase (切流与退役) is in the Feasibility timeline but its deliverables (cutover plan, rollback runbook drill, decommission) are not in the In Scope list — scope and timeline are misaligned for the migration phase. The Scope 判定线's "可能归入本提案" for eiam changes keeps scope↔solution alignment uncertain. |
| Requirements ↔ Solution coherent | 21 | 25 | Requirements map mostly to solution. NFRs now have corresponding solution elements: security (token in memory, mTLS, pluggable interface), performance (≤300ms threshold), observability (metrics integration). The security NFR is well-addressed by the solution's credential handling design. Deduction: the observability NFR says "鉴权成功率、token 校验延迟 P95/P99…接入 eiam/opsagent 现有指标体系" but the solution doesn't describe how WatchAlert exports these metrics (metric format, endpoint, integration mechanism) — this creates a minor requirements↔solution gap. The performance NFR's degradation trigger (P95 > 300ms → full degradation) doesn't cleanly map to the degradation behavior described in the solution (which is designed for "eiam 不可用" / unavailable, not "eiam slow"). |

---

## Cross-Dimension Coherence Check

- **Scope vs Timeline gap**: P4 phase (切流与退役) is in the Feasibility timeline but its deliverables are not in the In Scope list — manifests in D7 (scope bounded) and D10 (scope↔SC alignment).
- **NFR vs Solution gap**: Performance NFR's degradation trigger (latency > 300ms) conflates "eiam slow" with "eiam unavailable"; the degradation behavior (4-hour read-only mode) is designed for unavailability, not transient latency — manifests in D4 (NFR) and D10 (requirements↔solution).
- **SC vs Scenario gap**: Migration scenario has decommission criteria but no formalized SC — manifests in D9 (coverage) and D10 (scope↔SC alignment).
- **Risk vs Decision Matrix gap**: Decision matrix has 3 fallback channels but no risk for "all 3 fail" — manifests in D8 (risks identified).

---

## Phase 3 — Blindspot Hunt

### [blindspot] BS-1: P4 phase deliverables not in In Scope

**Quote**: "P4 切流与退役（约 2-3 人日 + 2 周观察期）：并行运行、配置迁移、nginx 切流、旧实例退役"

**Issue**: The Feasibility section describes P4 as a timeline phase with concrete deliverables (parallel run, config migration, nginx cutover, decommission). The Rollback & Upstream section describes a rollback runbook and rebasing plan. But none of these migration/cutover deliverables appear in the In Scope list. The In Scope list focuses on code/config changes (clone, dev.ps1, Justfile, forge, auth, decision records, nginx entry, compile-fail marking) but omits the migration deliverables that P4 promises. This is a structural misalignment: the timeline commits to a phase whose deliverables aren't scoped, so a team cannot determine what "done" means for P4 from the In Scope list alone. This falls between Scope Definition (what's in scope) and Feasibility (timeline phases) — neither dimension alone captures the misalignment.

**Required improvement**: Add migration/cutover deliverables to In Scope: (1) migration/cutover runbook, (2) rollback runbook (already described, should be formally scoped), (3) decommission checklist with the 2-week observation criteria.

### [blindspot] BS-2: Rollback runbook doesn't explicitly cover frontend revert

**Quote**: "本地 JWT handler 与密钥配置在移除时打 revert tag 保留，回滚即 git revert 该 tag 部署"

**Issue**: The rollback step 1 says "恢复前端登录入口" but the mechanism (git revert tag) is described only for "本地 JWT handler 与密钥配置" — backend components in the Haven/watch-alert repo. The frontend login entry removal is in a separate repo (Haven/watch-alert-web, per the In Scope item "删除前端登录入口"). The runbook doesn't describe how the frontend revert is executed: is there a separate revert tag in the frontend repo? Does the frontend need to be rebuilt and redeployed? The runbook's revert mechanism appears to cover only the backend, leaving the frontend rollback implicit. This is a cross-section gap between the Rollback runbook (which describes a single-repo revert) and the Scope (which modifies two separate repos).

**Required improvement**: Specify the frontend rollback mechanism explicitly — e.g., "前端登录入口移除时同步在 Haven/watch-alert-web 打 revert tag，回滚时 git revert 该 tag 并重新 build 部署前端."

### [blindspot] BS-3: No SC for migration/cutover success

**Quote**: "退役标准：新实例连续 2 周零告警丢失/误报、eiam 登录稳定、回滚 runbook 已演练"

**Issue**: The migration transition scenario defines decommission criteria (2 weeks zero alert loss/misfire, eiam login stable, rollback runbook drilled), but these criteria appear only in the Key Scenarios description — they are not formalized as Success Criteria. SC-6 tests eiam login works, but no SC tests that the new instance correctly serves the full alert pipeline (rules → data source → notifications) after cutover, or that the cutover itself completed successfully. The decommission criteria are measurable and testable ("连续 2 周零告警丢失/误报") but exist outside the SC framework. This is a structural misalignment between what the scenario promises to verify and what the SC list actually measures — the most operationally critical phase (live system migration) has no formal success gate.

**Required improvement**: Add an SC for migration success — e.g., "切流后新实例连续 1 个告警周期内规则触发→数据源查询→通知投递全链路无丢失/误报；退役前连续 2 周零告警丢失/误报."

---

## Bias Detection Report

**Annotated regions (5 paragraphs with `<!-- pre-revised -->` markers)**:
- Line 33: Auth decision matrix (OIDC→LDAP→桥接)
- Line 46: eiam 不可用降级方向 (token cache + read-only)
- Line 116: 鉴权统一到 eiam scope item (JWT removal)
- Line 128: Scope 判定线 (eiam in/out-of-scope boundary)
- Line 172: SC-7 JWT removal (gated on SC-6)

Attack points in annotated regions: 2
1. Degradation trigger conflates "eiam slow" with "eiam unavailable" (scored in D4/D10)
2. "可能归入本提案" scope uncertainty persists (scored in D7)

**Unannotated regions** (new content from iteration revision): ~65 content blocks including migration transition scenario, NFR section, Rollback & Upstream Cadence section, risk table new rows, comparison table updated cons, timeline 4-phase breakdown, Grafana/Alertmanager references, SC additions (nginx, compile-fail).

Attack points in unannotated regions: 9
1. P4 deliverables not in In Scope (BS-1)
2. Rollback runbook frontend coverage gap (BS-2)
3. No SC for migration/cutover success (BS-3)
4. Degradation exit condition unspecified (scored in D4)
5. Timeline doesn't account for eiam dependency readiness expansion (scored in D6)
6. No risk for all-3-auth-channels-fail scenario (scored in D8)
7. Observability metrics export mechanism not described (scored in D10)
8. No user session/token migration during auth cutover (scored in D4)
9. Upstream cadence claim ("发版不规律") without evidence (scored in D8)

**Density comparison**:
- Annotated regions: 2 attack points / 5 paragraphs = density 0.40
- Unannotated regions: 9 attack points / ~65 paragraphs = density 0.14
- Ratio (annotated/unannotated): 2.9

**Assessment**: The density ratio dropped from 4.5 (iteration 1) to 2.9 (iteration 2). This reflects two factors: (1) the pre-revised regions were significantly improved — the auth decision matrix, degradation path, and SC gating now address most iteration-1 concerns, leaving only residual issues; (2) the new unannotated content (migration, rollback, NFRs) added substantial surface area with genuine minor issues. The lower ratio indicates reduced scorer bias toward annotated regions — the unannotated content received proportionally more scrutiny this round. No `conflict-with-pre-revision` tags applied: all pre-revision directions (auth decision matrix, degradation path, scope boundary, SC gating) are consistent with the revision's stated intent and improved the document.

---

## Score Summary

| # | Dimension | Score | Max | Δ from Iter 1 |
|---|-----------|-------|-----|---------------|
| 1 | Problem Definition | 100 | 110 | +12 |
| 2 | Solution Clarity | 104 | 120 | +16 |
| 3 | Industry Benchmarking | 103 | 120 | +20 |
| 4 | Requirements Completeness | 96 | 110 | +19 |
| 5 | Solution Creativity | 60 | 100 | +12 |
| 6 | Feasibility | 80 | 100 | +17 |
| 7 | Scope Definition | 72 | 80 | +7 |
| 8 | Risk Assessment | 77 | 90 | +10 |
| 9 | Success Criteria | 71 | 80 | +11 |
| 10 | Logical Consistency | 77 | 90 | +10 |
| **Total** | | **840** | **1000** | **+134** |

**Target**: 900/1000 — **NOT MET** (gap: 60 pts)

**Gap Analysis**: The remaining 60-point gap is distributed across all dimensions, with the largest absolute gaps in D5 Solution Creativity (40 pts below max) and D6 Feasibility (20 pts below max). D5 is inherently low for a practical integration proposal that applies standard patterns correctly rather than innovating. The remaining issues are minor and concentrated in: (1) the auth integration's unverified feasibility (inherent to proposal-stage deferment), (2) migration/cutover deliverables not formally scoped, (3) degradation design gaps (trigger/exit condition), and (4) SC coverage for migration success.
