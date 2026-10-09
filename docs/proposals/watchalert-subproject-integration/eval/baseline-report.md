# Baseline Scoring Report — WatchAlert 子项目并入提案

**Scorer**: CTO Adversary
**Mode**: Baseline (原始未修订提案)
**Date**: 2026-10-09
**Rubric**: proposal.md (1000 pts, 10 dimensions)

---

## Phase 1 — Reasoning Audit (Pre-score Anchors)

### Argument Chain Trace

1. **Problem → Solution**: Problem = WatchAlert runs outside Haven workspace orchestration + auth fragmented from eiam. Solution = clone forks as sibling repos + four-tool-suite registration + delegate auth to eiam. Mapping is sound; both stated problems are addressed.

2. **Solution → Evidence**: Evidence is mostly concrete infrastructure references (decision 0002, dev.ps1 service registry, Justfile, forge config). However, the claim "eiam 已具备登录认证能力" is treated as fact, while the document later admits "对接协议需 tech-design 明确，可能需要 eiam 侧最小改动" — an unstated assumption treated as established fact.

3. **Evidence → Success Criteria**: Most SC are testable (git log, git push mine, dev.ps1 start, just recipes, forge quality-gate). But "用 eiam 账号登录 WatchAlert 成功（SSO）" does not define what "成功" means — no token validation, session persistence, or fallback verification specified.

4. **Self-contradiction check**:
   - **Anchor A — eiam auth mechanism logically broken**: Assumptions Challenged states "不要求 eiam 成为完整 OIDC IdP", but WatchAlert's OIDC login mode requires an OIDC IdP endpoint to point to. If eiam is not an OIDC IdP, the mechanism is undefined. The "Refined" finding creates a gap it claims to resolve.
   - **Anchor B — eiam modification scope boundary undefined**: Constraints say "可能需要 eiam 侧最小改动"; Out of Scope says "eiam 本体功能增强...另立 feature". The line between "最小必要改动" (in scope) and "功能增强" (out of scope) is never drawn.
   - **Anchor C — No timeline**: "单人可完成，分阶段（见 Next Steps）" but Next Steps contains only "Proceed to /write-prd". No phases, no estimates, no milestones.
   - **Anchor D — eiam-unavailable degradation has no solution**: Edge case "鉴权服务(eiam)不可用时的降级" is listed under Key Scenarios, but no solution detail, no risk entry, and no SC address it. SC6 removes local JWT login — meaning if eiam is down, WatchAlert has no login path at all.

---

## Phase 2 — Rubric Scoring

### Dimension 1: Problem Definition — 90/110

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 32/40 | Problem is stated but conflates two issues (orchestration integration + auth unification) without declaring which is primary. Quote: "WatchAlert 后端与前端以独立目录/独立进程运行...且登录仍是 WatchAlert 自身的本地 JWT，与运维平台的身份体系（eiam）割裂。" — two problems joined by "且" without priority. |
| Evidence provided | 34/40 | Concrete file references (0002 decision, dev.ps1 registry, Justfile, forge config, PR numbers #257/#263/#264). Solid but no quantitative data (no metrics on how often manual startup occurs, no incident counts). |
| Urgency justified | 24/30 | Cost of delay explained (fork divergence, manual startup, audit risk). Quote: "愈晚迁移，与上游 fork 的历史分叉愈深、迁移成本愈高。" — reasonable but no quantification of actual cost. |

### Dimension 2: Solution Clarity — 88/120

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 34/40 | Reader can explain what will be built: clone forks, register in four-tool-suite, delegate auth. Concrete enough. |
| User-facing behavior described | 30/45 | SSO login and dev.ps1 start described. But the actual SSO user flow is thin — no description of redirect flow, token exchange, session lifecycle. Quote: "凭 eiam 账号...登录，无需 WatchAlert 独立账号" — states outcome, not observable behavior. |
| Technical direction clear | 24/35 | Vague language on the critical auth integration. Quote: "对接协议需 tech-design 明确，可能需要 eiam 侧最小改动" — the most technically risky part is explicitly deferred. Vague language penalty: -20 for "可能需要" (uncertainty on core mechanism). |

### Dimension 3: Industry Benchmarking — 92/120

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Industry solutions referenced | 30/40 | Patterns cited: git submodule, git subtree, fork+clone, OIDC/LDAP IdP delegation. But no specific product names or open-source project names referenced — only generic pattern categories. |
| At least 3 meaningful alternatives | 24/30 | 5 alternatives listed including "do nothing". However, git submodule and git subtree are effectively straw men — rejected on a premise the document itself establishes ("D:\Haven 是 workspace 无此前提"). Straw-man penalty: -20 for subtree alternative rejected on self-defined premise. |
| Honest trade-off comparison | 18/25 | Selected option's cons are minimal: "需遵守向 fork 同步纪律" — a trivial con compared to substantive cons of alternatives. Cherry-picked. |
| Chosen approach justified against benchmarks | 20/25 | Justified as "与现有工作区形态一致且最简". Reasonable rationale. |

### Dimension 4: Requirements Completeness — 82/110

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Scenario coverage | 32/40 | Happy path + some edge cases (fork divergence, eiam unavailable, cmd/ compile). But eiam-unavailable degradation has no solution detail. Quote: "鉴权服务(eiam)不可用时的降级" — listed but unresolved. |
| Non-functional requirements | 24/40 | Security and compatibility mentioned. No performance NFRs (e.g., SSO login latency target, concurrent user capacity). No accessibility. Missing relevant NFRs. |
| Constraints & dependencies | 26/30 | eiam dependency, Windows env, fork remotes named. Good but eiam API is unconfirmed ("需确认"). |

### Dimension 5: Solution Creativity — 65/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Novelty over industry baseline | 30/40 | Genuine reframing: "D:\Haven 本来就是多仓库 workspace，无需 subtree 迁入 monorepo". But "just clone the fork" is essentially the industry default for multi-repo workspaces — limited novelty. |
| Cross-domain inspiration | 15/35 | No cross-domain borrowing evident. Solution stays within git/workspace domain. |
| Simplicity of insight | 20/25 | "fork 本身就是 git 仓库，直接 clone 下沉为同级子项目" is an elegant simplification. |

### Dimension 6: Feasibility — 72/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Technical feasibility | 34/40 | git clone + dev.ps1 patterns proven (eiam/opsagent already using). But eiam auth integration feasibility is asserted not demonstrated. Quote: "eiam 已具备登录认证能力，WatchAlert 已内置 OIDC/LDAP 登录 —— 两端能力就绪" — asserted without evidence of compatibility. |
| Resource & timeline feasibility | 18/30 | "单人可完成" but no timeline. Quote: "单人可完成，分阶段（见 Next Steps）" — Next Steps has no phases. |
| Dependency readiness | 20/30 | Fork remotes ready, but key dependency (eiam auth API) is unconfirmed. Quote: "eiam 认证 API 需确认（是否有可复用的登录校验端点...）" |

### Dimension 7: Scope Definition — 64/80

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 26/30 | Most items are concrete deliverables. But "鉴权统一到 eiam：WatchAlert 登录委托 eiam 认证" is vague — no definition of what "委托" entails technically. |
| Out-of-scope explicitly listed | 22/25 | Four items listed clearly. |
| Scope is bounded | 16/25 | No timeframe. "分阶段" without phases. Quote: "单人可完成，分阶段（见 Next Steps）" — but no phases defined. |

### Dimension 8: Risk Assessment — 72/90

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 26/30 | 5 meaningful risks (fork sync, eiam protocol, cmd/ compile, auth code changes, frontend labels). |
| Likelihood + impact rated | 26/30 | Honest mix (M/H, M/H, H/M, M/M, M/L). Not all low-likelihood-high-impact. |
| Mitigations are actionable | 20/30 | Some actionable (service table custom build command). Some vague: "文档化...纪律" and "tech-design 阶段先摸清" — process deferrals, not concrete mitigations. |

### Dimension 9: Success Criteria — 62/80

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable and testable | 24/30 | Most testable. But SC6 "用 eiam 账号登录 WatchAlert 成功（SSO）" — "成功" is not precisely defined (no token validation, no session persistence check). Vague language penalty: -20 for "成功" (unmeasurable qualifier). |
| Coverage is complete | 20/25 | In Scope item "nginx dev 入口调整（dev.ps1 直接起，消除手动启动流程）" has no corresponding SC. |
| SC internal consistency | 18/25 | SC6 combines "eiam SSO login" + "local JWT removed". The edge case scenario "鉴权服务(eiam)不可用时的降级" requires a fallback login, but SC6 removes the only fallback (local JWT). SC6 is satisfiable alone but contradicts the degradation scenario requirement — ambiguous pair flagged. |

### Dimension 10: Logical Consistency — 70/90

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Solution addresses the stated problem | 30/35 | Both problems (orchestration + auth) addressed. |
| Scope ↔ Solution ↔ Success Criteria aligned | 22/30 | InScope "nginx dev 入口调整" has no SC. Auth mechanism vague. Quote: "鉴权统一到 eiam：WatchAlert 登录委托 eiam 认证" (scope) vs "对接协议需 tech-design 明确" (solution) — scope promises a deliverable, solution defers its definition. |
| Requirements ↔ Solution coherent | 18/25 | "鉴权服务(eiam)不可用时的降级" scenario has no solution detail and no SC. Orphan requirement. |

---

## Cross-Dimension Coherence Check

- **Scope ↔ SC gap**: InScope "nginx dev 入口调整" has no SC — scored in D9 (coverage) and D10 (alignment).
- **Solution ↔ Feasibility gap**: Auth integration is the riskiest part but is deferred to tech-design — scored in D2 (technical direction) and D6 (technical feasibility).
- **Requirements ↔ Solution gap**: eiam-unavailable degradation scenario is orphaned — scored in D4 (scenario coverage) and D10 (requirements coherence).

---

## Phase 3 — Blindspot Hunt

### [blindspot] 1 — eiam auth mechanism is logically incoherent

**Quote**: "eiam 作统一认证源，WatchAlert 复用已有 OIDC/LDAP 登录对接之，不要求 eiam 成为完整 OIDC IdP"

**Issue**: WatchAlert's built-in OIDC login mode requires an OIDC IdP endpoint (authorization server, token endpoint, userinfo endpoint). If eiam is "不要求...成为完整 OIDC IdP", then what does WatchAlert's OIDC configuration point to? The LDAP path is plausible (WatchAlert queries eiam via LDAP), but the document simultaneously claims OIDC reuse. The mechanism is logically broken — the "Refined" finding creates a gap it claims to bridge. This is a fundamental reasoning flaw independent of any dimension.

**Improve**: Define the concrete auth protocol — is it LDAP bind against eiam's directory? Is it a custom token-validation endpoint? Is eiam acting as an OIDC IdP or not? Pick one and specify the integration surface.

### [blindspot] 2 — No timeline or phase breakdown exists

**Quote**: "单人可完成，分阶段（见 Next Steps）。"

**Issue**: "分阶段" implies phases, but Next Steps contains only "Proceed to /write-prd to formalize requirements（全流程，非 quick）". No phases, no estimates, no milestones. A CTO reviewing this proposal cannot assess whether the scope is realistic without any timeline signal. This falls between D6 (resource feasibility) and D7 (scope bounded) but is a structural gap neither fully captures.

**Improve**: Provide at minimum a phased breakdown (e.g., Phase 1: clone + orchestration registration; Phase 2: auth integration; Phase 3: decision record update) with rough effort estimates.

### [blindspot] 3 — eiam-unavailable degradation is an orphaned requirement

**Quote**: "鉴权服务(eiam)不可用时的降级" (Key Scenarios) vs SC6 "本地 JWT 登录入口移除/隐藏"

**Issue**: The edge case scenario requires a degradation path when eiam is unavailable. But SC6 removes the only fallback (local JWT login). No risk entry covers "eiam outage = WatchAlert login fully down". No SC tests degradation. This is a cross-section contradiction between a stated requirement and the success criteria — it falls between D4, D8, D9, and D10 but is a structural misalignment no single dimension fully scores.

**Improve**: Either (a) define a degradation path (e.g., cached token, emergency admin login) and add an SC for it, or (b) explicitly accept "eiam down = no login" as a documented trade-off with risk assessment.

### [blindspot] 4 — "最小通用改动" boundary with WatchAlert code is undefined

**Quote**: Out of Scope: "WatchAlert 深度内核改造（除鉴权衔接所需的最小通用改动外）" vs Assumptions: "可能需要 WatchAlert 少量通用改动（调用 eiam 认证）"

**Issue**: The proposal both claims WatchAlert auth changes are "最小通用改动" (in scope) and excludes "深度内核改造" (out of scope), but never defines the boundary. If the eiam integration requires touching WatchAlert's session middleware, token validation, or login handler — is that "最小" or "深度"? This boundary ambiguity is a hidden scope-creep risk that no rubric dimension captures directly.

**Improve**: Enumerate the specific WatchAlert code areas that may be modified (e.g., auth middleware, login handler, token store) and declare which are in-scope "minimal" vs out-of-scope "core".

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 90 | 110 |
| 2. Solution Clarity | 88 | 120 |
| 3. Industry Benchmarking | 92 | 120 |
| 4. Requirements Completeness | 82 | 110 |
| 5. Solution Creativity | 65 | 100 |
| 6. Feasibility | 72 | 100 |
| 7. Scope Definition | 64 | 80 |
| 8. Risk Assessment | 72 | 90 |
| 9. Success Criteria | 62 | 80 |
| 10. Logical Consistency | 70 | 90 |
| **TOTAL** | **757** | **1000** |

**Target**: 900. Gap: 143 points.
