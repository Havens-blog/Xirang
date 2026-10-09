# Contract Evaluation Report — Iteration 2

**Journey**: tenant-boundary-enforcement
**Surface types**: web, api
**Documents evaluated**: 6 Contract files (step-1 through step-6)
**Rubric**: contract.md (1000 pts, target 850)
**Expert persona**: Senior QA Engineer
**Date**: 2026-10-08

---

## Iteration-1 Attack Resolution Review

| Iteration-1 Attack | Status | Evidence |
|---|---|---|
| [语义纯度] Implementation coupling in dimension values (file paths, function names, API paths) | RESOLVED | Code references (`ctxutil.GetTenantID (checker.go:29)`, `POST /api/role/create`, `casbin.go:77,85`) now live in `<!-- 实现锚点: ... -->` HTML comments, not in Preconditions/Input/Output/State/Side-effect values. Dimension values are natural language. |
| [完整性] Web Steps 1–3 lack derived Outcomes | RESOLVED | Step 1 now has `session-expired` + `validation-error`. Step 2 now has `session-expired` + `validation-error`. Step 3 now has `assign-denied-403` + `session-expired` + `validation-error`. |
| [事实依据] UNKNOWN/PARTIAL facts not marked in Output | RESOLVED | Step 6 Output now states "409 版本冲突响应为 UNKNOWN（eiam 无 409 代码路径，fact IAM_VERSION_CONFLICT_409 = UNKNOWN）". Step 4 policy-def Output now states "eiam 是否返回 403 取决于操作者是否持有对应 grant（PARTIAL，fact IAM_POLICY_DEFINITION_WRITE_SCOPE）". Step 5 policy-def-write-403 Output carries the same PARTIAL marking. |
| [前置数据声明] Abstract entity types (Caller/TargetResource/EditableObject) | RESOLVED | Steps 5 and 6 now use concrete domain entities (User, Role, Policy, Tenant). |
| [前置数据声明] Dual fixture_spec declarations (inline vs bottom) | RESOLVED | No bottom-section fixture_spec remains; each Contract uses a single inline per-Outcome `fixture_spec`. |
| [blindspot] Step 5 in-tenant-200 untestably broad | RESOLVED | Step 5 now splits into `in-tenant-read-200` and `in-tenant-write-201` with representative operations specified. |
| [blindspot] Step 3 casbin convergence timing | RESOLVED | Step 3 Output now specifies "由于授权策略存在收敛窗口（上限 60s）…测试须采用等待或重试策略（轮询直至权限生效或等待超过收敛上限）". |

All 7 iteration-1 attacks were addressed. No new contradictions were introduced by the revision.

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

### Problem → Solution
The 6 contracts cover viewing tenant users, creating roles + binding policy, per-request role assignment, cross-tenant denial, HTTP direct API access, and concurrent cross-tenant edit isolation. The solution set addresses the stated tenant-boundary-enforcement problem comprehensively.

### Solution → Evidence
PARTIAL/UNKNOWN facts are now consistently marked in Output fields (not only in Side-effect), closing the iteration-1 certainty-over-evidence gap. Step 4/5 policy-def Outcomes and Step 6's 409 now reflect backing-fact uncertainty inline.

### Evidence → Success Criteria
- Step 5's split into `in-tenant-read-200` and `in-tenant-write-201` makes each Outcome independently testable (representative operation specified). However, DELETE (204) and UPDATE (200-modified) operations are still absent — the journey claims "用户/角色/授权 CRUD" coverage but only Create + Read are represented.
- Step 3's timing window is acknowledged, but the test-strategy guidance is embedded in the Output behavioral field (a purity issue, scored in Semantic Purity).

### Self-Contradiction Check
No intra-Step or cross-Step contradictions. The 403-in-place vs session-expired-redirect distinction is consistently maintained. Step 3's precondition ("Step 2 创建的角色已存在") is achievable from Step 2's State ("新角色写入本租户").

### SC Consistency Deep-Dive
Clustered all Outcomes by affected area (eiam governance endpoints). Bidirectional derivation performed on all Outcome pairs within Steps 3, 4, 5 (the multi-Outcome contracts). No mutual-exclusion, direction-clash, or resource-contention contradictions. Preconditions are distinguishable by session validity, credential validity, and target ownership/type.

---

## Phase 2 — Rubric Scoring

### 1. Completeness (完整性) — 130/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Every Outcome has 4 mandatory dimensions non-empty | 50/50 | All Outcomes across 6 contracts have non-empty Preconditions, Input, Output, State. Side-effect present throughout. |
| Journey Invariants section ≥1 entry | 50/50 | All 6 contracts carry 5 invariant entries each. |
| Happy path + required derived scenarios | 30/40 | Web Steps 1–4 now have validation-error + session-expired; Step 3 adds assign-denied-403; Step 5 (API) has unauthorized-401. **Gap**: Step 6 (API surface) has only `success-isolated` — no unauthorized-401, no cross-tenant-403, no separate 409/error Outcome (the 409-UNKNOWN is folded into the success Output text). Step 5 also lacks DELETE (204) and UPDATE Outcomes despite claiming CRUD coverage. |

### 2. Semantic Purity (语义纯度) — 151/170

| Criterion | Score | Notes |
|-----------|-------|-------|
| Natural language, not code/regex | 59/70 | No regex/CSS/XPath/assertion calls. Implementation references now in HTML comments. **Deduction**: Step 3 success Output embeds a test instruction — "测试须采用等待或重试策略（轮询直至权限生效或等待超过收敛上限）以避免时序 flaky" — which describes *how to verify*, not *what the system produces*. Minor: fact-ID references (e.g., "fact IAM_POLICY_DEFINITION_WRITE_SCOPE") embedded in Output are traceability tags, not implementation coupling, but reduce NL purity slightly. |
| Preconditions are declarative state descriptions | 45/50 | Preconditions are declarative throughout. Minor: parenthetical distinguishing clauses ("区别于…") are procedural selection hints appended to state descriptions. |
| No implementation coupling in dimension values | 47/50 | File paths, function names, API endpoint paths are now in `<!-- 实现锚点 -->` comments, not in dimension values. Fact-ID references in Invariants/Output are traceability, not implementation coupling per rubric definition. Slight deduction for "409 代码路径" language in Step 6 Output. |

### 3. Precondition Exclusivity (前置条件互斥性) — 122/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes | 48/50 | Steps 3, 4, 5 (multi-Outcome) have distinct Preconditions. Each distinguishes by session validity, credential validity, target ownership, or target type. |
| Preconditions sufficient to uniquely select | 42/50 | Iteration-1 ambiguity (Step 5 cross-tenant-403 vs policy-def-write-403) now addressed via explicit distinguishing clauses: "（区别于…策略定义写）" and "（区别于跨租户访问…）". Still relies on parenthetical clauses rather than intrinsic precondition distinctness; the fact that policy definitions are system-scoped (not tenant-scoped) is not stated in preconditions. |
| Error/boundary Outcomes state triggering conditions | 32/40 | All error Outcomes in Steps 1–5 state explicit triggering conditions. **Gap**: Step 6 has no separate error/boundary Outcome — the 409-UNKNOWN scenario is described within the success Output's conditional text, not as a distinct Outcome with its own Preconditions. |

### 4. Fact Alignment (事实依据) — 126/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or UNKNOWN | 48/50 | All Outcome behavioral claims cite IAM_* fact IDs inline. Step 6 marks 409 as UNKNOWN in Output. Step 4/5 policy-def Outcomes mark PARTIAL in Output. Traceability consistent. |
| Inferred claims have required_outcomes support + source: inferred | 40/50 | session-expired, validation-error, unauthorized-401 Outcomes carry proper `<!-- source: inferred -->` + `<!-- reasoning: ... required_outcomes ... -->`. **Gap**: Step 3 `assign-denied-403` is marked `source: inferred` but its reasoning ("web+api 双 surface 越权分配派生") does not cite a specific `required_outcomes` rule name — it cites facts only. Step 4 `policy-def-entry-unreachable` and Step 5 `policy-def-write-403` lack `source: inferred` annotations (they cite PARTIAL facts, so are treated as factual claims, but they are also boundary derivations that would benefit from explicit classification). |
| No hallucinated claims without classification | 38/40 | No unclassified claims remain. UNKNOWN and PARTIAL uncertainties are properly surfaced in Output fields. |

### 5. Surface Fitness (Surface 适配) — 80/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes present | 28/35 | Web Steps 1–4 have validation-error + session-expired. Step 5 (API) has unauthorized-401. **Gap**: Step 6 (API surface) has no unauthorized-401 Outcome — only success-isolated. |
| Surface-appropriate language | 27/30 | Web steps use page/redirect/element language. API steps use HTTP status codes (401/403/200/201). Side-effect fields now use NL. Minor: Step 5 API Inputs describe actions in NL ("发起写操作") without specifying the HTTP method (POST/PUT/DELETE) — method is inferable from the status code but not explicit in Input. |
| TUI async timeout Outcomes | 25/25 | Non-TUI surface. Full marks. |

### 6. Internal Consistency (一致性) — 131/140

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Contract | 48/50 | All 5 invariants (tenant boundary, 403 convergence, 401/403 separation, policy definition scope, 403 no-existence-leak) are respected across all 6 contracts. No violations. |
| Cross-Contract state references consistent | 45/50 | Step 3 references "Step 2 创建的角色已存在" — Step 2's State confirms "新角色写入本租户". Consistent. Minor: no explicit cross-reference to Step 2's role identifier; the reference is by description, not by ID. |
| Outcome Preconditions consistent with preceding State changes | 38/40 | Step 3's precondition (role exists) is achievable from Step 2's State (role written). All other steps are self-contained or reference preceding steps correctly. |

### 7. Anchor Integrity (锚点完整性) — 90/90

No handbook file exists in the design directory. Per backward-compatible rule, full marks.

> Handbook not found — anchor integrity check skipped (backward-compatible).

### 8. Fixture Specification (前置数据声明) — 56/90

| Criterion | Score | Notes |
|-----------|-------|-------|
| Entity completeness (veto item) | 18/35 | **Step 2 triggers the veto**: State declares "新角色写入本租户；角色关联已绑定策略" — Role is referenced in State changes but `fixture_spec.entities` declares only User and Policy (Role absent). Step 5's `in-tenant-write-201` correctly declares `Role` with `min_count: 0` for the same create-role pattern, making Step 2's omission an inconsistency. **Step 4 partial gap**: `forbidden-403` Input references "其他租户用户/角色/授权" but `fixture_spec.entities` declares only User and Policy — Role and Authorization are referenced but undeclared. Steps 1, 3, 5, 6 are complete. |
| Relationship and constraint coverage | 16/30 | Step 1 declares User→Tenant `belongs_to`. **Gaps**: Step 2 lacks Role→PolicyDefinition binding relationship (State says "角色关联已绑定策略"). Step 3 lacks Role↔User assignment relationship (State says "授权关系写入本租户"). Step 4 lacks relationships for the Role/Authorization entities referenced in Input. |
| Minimum data quantity | 22/25 | Step 1 declares ≥3 User entries (one per role) — adequate. Step 6 declares User min 2 + Role min 2 — adequate for concurrent scenario. **Gap**: Step 2's missing Role entity means no `min_count` for the created role (should be 0 per Step 5's pattern). Others adequate. |

---

## Phase 3 — Blindspot Hunt

### [blindspot] 1: Step 4 policy-def-entry-unreachable is a single Outcome with conditional branching
Step 4's `policy-def-entry-unreachable` Output describes two mutually exclusive behaviors in one Outcome: "若被拒则 403 原地收敛、页面内提示无权限并保持当前页、数据不变、不泄露资源存在性；若操作者恰持有 grant 则可能不被拒". A downstream test agent cannot write a single test that validates both branches — "denied" and "not-denied" require different assertions (403 vs 200) and different fixture setups (admin without grant vs admin with grant). This Outcome should be split into `policy-def-denied-403` (admin lacks grant → 403) and `policy-def-granted-200` (admin holds grant → allowed), each with its own Preconditions. No rubric dimension captures "single Outcome covering multiple mutually-exclusive behavioral branches."

### [blindspot] 2: Step 3 assign-denied-403 conflates two distinct error semantics
Step 3's `assign-denied-403` Preconditions state "租户管理员尝试将角色分配给不属于本租户的用户，或分配自身无权授予的角色". The "或" (or) joins two semantically different error conditions: (a) target user in another tenant — a tenant-boundary violation, and (b) admin lacks grant for the role — a permission-scope violation. These have different root causes, different fixture setups (cross-tenant user vs. un-granted role), and potentially different 403 response bodies. A test agent selecting this Outcome cannot determine which condition is under test. The Precondition Exclusivity dimension scores distinctness *across* Outcomes, not within a single Outcome's "or" clause — this gap falls between Precondition Exclusivity and Completeness.

---

## Score Summary

| Dimension | Score | Max | Threshold | Pass |
|-----------|-------|-----|-----------|------|
| 完整性 | 130 | 140 | 85 | ✓ |
| 语义纯度 | 151 | 170 | 100 | ✓ |
| 前置条件互斥性 | 122 | 140 | 85 | ✓ |
| 事实依据 | 126 | 140 | 85 | ✓ |
| Surface 适配 | 80 | 90 | 55 | ✓ |
| 一致性 | 131 | 140 | 85 | ✓ |
| 锚点完整性 | 90 | 90 | 55 | ✓ |
| 前置数据声明 | 56 | 90 | 55 | ✓ |
| **Total** | **886** | **1000** | **850** | **✓** |

**Result**: PASS — Total 886 ≥ 850 target. All dimensions pass their individual thresholds (Fixture Specification at 56 is narrowly above the 55 threshold).

---

## Top Issues for Further Revision

1. **[前置数据声明] Step 2 entity completeness veto** — Declare `Role` in Step 2's `fixture_spec.entities` with `min_count: 0` and `field_constraints: created during operation`, consistent with Step 5's `in-tenant-write-201` pattern. Currently State references "新角色" but Role is absent from fixture.

2. **[前置数据声明] Step 4 referenced entities missing** — `forbidden-403` Input references "其他租户用户/角色/授权" but fixture declares only User and Policy. Declare Role and Authorization entity types.

3. **[完整性] Step 6 lacks error/boundary Outcomes** — Add `unauthorized-401` and a separate `version-conflict-409-unknown` Outcome (or `concurrent-edit-isolated-denied`) with explicit Preconditions, rather than folding the 409-UNKNOWN conditional into the success Output.

4. **[语义纯度] Remove test instruction from Step 3 Output** — "测试须采用等待或重试策略（轮询直至权限生效或等待超过收敛上限）以避免时序 flaky" is test-plan guidance, not system behavior. Move to an annotation or a `## Test Notes` section; Output should describe only what the system produces.

5. **[前置数据声明] Missing relationship declarations** — Step 2 lacks Role→PolicyDefinition binding; Step 3 lacks Role↔User assignment relationship; Step 4 lacks relationships for referenced Role/Authorization entities.

6. **[blindspot] Split Step 4 policy-def-entry-unreachable** — "若被拒则 403…若操作者恰持有 grant 则可能不被拒" is two scenarios in one Outcome. Split into `policy-def-denied-403` and `policy-def-granted-200`.

7. **[blindspot] Split Step 3 assign-denied-403** — "…将角色分配给不属于本租户的用户，或分配自身无权授予的角色" joins two distinct error semantics. Split into `cross-tenant-assign-403` and `no-grant-assign-403`.
