# Iteration 3 PM Scorer Report (Final, Adversarial)

**Scorer**: Senior PM (adversarial)  
**Iteration**: 3 (final)  
**Score**: 997/1000  
**Target**: 900 — **PASS**

---

## Phase 1: Reasoning Audit

### Problem → Solution trace
- **Problem**: WatchAlert 独立目录 + 独立进程 + 反代游离于平台 workspace 之外；本地账号与 eiam 割裂；启动手动两步。
- **Solution**: 子项目纳入 workspace + 统一启动编排 + 统一质量门禁 + 登录统一 eiam + 移除本地登录入口。
- **Verdict**: Solution directly addresses every facet of the stated problem. No scope drift.

### Solution → Evidence trace
- 登录统一 → Main flow 1 (SSO) + Story 1 + UF-1 + 回跳白名单 + 403 无角色分支 + 后门清零 404。
- 降级韧性 → 异常流程 + Story 2 + UF-2/UF-3 + 探针滞回 + 脱敏清单。
- 迁移切流 → 主流程 3 + Story 6 + 阶段门控 + 回滚 + 回滚失败兜底 + 部分导入失败 + 抽样口径。
- 启动编排 → 主流程 2 + Story 4 + 编译失败标记。
- **Verdict**: Evidence supports the solution at every point.

### Evidence → Success Criteria trace
- Goals 表量化：100% eiam / ≤ 编译+30s / ≥ 80% 覆盖率 / 404 后门清零 / 季度 ≥ 1 PR 上游。全部可测。
- 切流 pass 标准：鉴权成功率 ≥ 99.9% / P95 ≤ 300ms / diff=0 / 告警周期无遗漏。可测。
- **Verdict**: Success criteria test what matters, not proxy metrics.

### Self-contradiction check (SC clustering)
- 会话架构：spec L209、Story 5 AC L84、In Scope L61、Dependency Baseline L221 全部一致表述「复用平台共享会话存储 / 共享 session 存平台 Redis / 不自建独立会话存储 / 不签发进程内自签名会话票」。**Iteration 2 的 CRITICAL 矛盾（进程内 HMAC 内存票 + 内存撤销列表 vs 共享会话）已彻底消除**——"进程内自签名会话票" 仅以否定形式出现（"不签发…"），无任何正向残留。
- 写操作状态码：spec L109「写操作返回 503」、UI UF-2 L83「后端返回 503」、Story 2 AC L41「写操作返回 503」三文档一致。
- 探针参数：spec L205「每 10 秒 / 超时 3 秒 / 连续 3 次」、UF-2 L97「连续 3 次失败（30 秒内）」、UF-3 L127「每 10s 探针 / 超时 3s / 连续 3 次成功」一致。3×10s=30s 逻辑自洽。
- 角色匹配时序：spec L31「eiam 用户登录后…匹配预置角色表」→ 隐含角色变更在下次登录生效，与 403 无匹配分支一致。
- 降级恢复路径：UF-2「已恢复」状态（探针 3 次成功 → 横幅消失）与 UF-3「已恢复」状态（探针 3 次成功 → 自动跳 eiam 登录页）分别面向已登录/匿名用户，无矛盾。
- **No new contradictions introduced by the revision.**

---

## Phase 2: Rubric Scoring

### 1. Background & Goals — 100/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Three elements (Reason/Target/Users) | 30/30 | Why L15（游离+割裂+手动两步）、What L19（子项目+统一 eiam+移除本地入口）、Who L23-28（三角色表含备注）全部具体。 |
| Goals quantified | 30/30 | 100% eiam / ≤ 编译+30s / ≥ 80% 覆盖率 / 404 / 季度 ≥ 1 PR。至少 5 个数值目标。 |
| Logical consistency | 40/40 | 双账号问题→100% eiam 目标；手动两步→一键启动目标；缺失门禁→≥ 80% 覆盖率。因果链完整。 |

### 2. Flow Diagrams — 150/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Mermaid exists | 50/50 | 4 张图：主流程（L113）、单点登出（L84）、开发者启动（L137）、迁移切流门控（L149）。 |
| Main path complete | 50/50 | 登录→认证→角色→会话→看板→降级→会话过期全链路覆盖。 |
| Decision + error branches | 50/50 | 菱形节点：是否已登录、认证成功、有角色、认证服务可用、宽限期过、会话过期、编译成功、各阶段 pass/fail、回滚分支。含回滚失败兜底路径。 |

### 3. Functional Specs (prd-ui-functions.md) — 200/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Placement & Interaction | 70/70 | 3 个 UF 均有 Placement（mode/target/position）+ User Interaction Flow 全路径。Page Composition 表完整。 |
| Data Requirements & States | 70/70 | 字段表含 Source/Trigger；状态表 4 态（正常/宽限中/已恢复/宽限过期）完整。匿名可见字段清单 + 脱敏规则明确。 |
| Validation Rules explicit | 60/60 | UF-1 回跳白名单（host/path/scheme 禁止清单）可执行；UF-2 探针防抖阈值（3 次）可执行；UF-3 无登录入口 + 探针驱动恢复可执行。 |

### 4. User Stories — 200/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Coverage per target user | 50/50 | 运维工程师（Story 1/2/5）、告警平台管理员（Story 3/6）、平台开发者（Story 4）三角色全覆盖。 |
| Format correct | 50/50 | 全部 As a / I want / So that；动作具体（登录/登出/切流/启动），无 "manage/handle" 泛词。 |
| AC per story (G/W/T) | 50/50 | 6 个 Story 均含 Given/When/Then，部分含多个 AC 覆盖正常+异常分支。 |
| AC verifiability & boundary | 50/50 | "不白屏、不无限 loading" 源自 BIZ-auth-008 业务规则原文，可二元验证（页面有内容/loading 态收敛）。"回原操作路径" 可验证。403/404/503/401 状态码均可断言。抽样口径（样本量/方法/一致定义）完全量化。 |

### 5. Scenario Completeness — 148/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| End-to-end coverage | 58/60 | 登录/会话过期/单点登出/降级/启动/切流全生命周期覆盖。-2：10% 灰度→100% 切流的门控用词「无 5xx/鉴权失败**激增**」（L102）未量化阈值，回滚触发线（< 99% / > 1%）提供了量化锚点但灰度阶段门控本身仍含主观判断空间。 |
| Implicit assumptions surfaced | 40/40 | Dependency Baseline 逐项标注现有/新建；eiam 契约基线（BIZ-auth-009）含 5 工作日通知流程；回跳白名单前提已声明。 |
| Business-rules consistency | 50/50 | BIZ-auth-003（403 不跳登录）✓ L76/L121；BIZ-auth-005（共享会话销毁+401 收敛+不实时广播）✓ L82/L209；BIZ-auth-006（统一文案防枚举）✓ L19/L49；BIZ-auth-008（过期不白屏+回原路径）✓ L79/L131；BIZ-auth-009（契约基线+变更通知）✓ L226；BIZ-resilience-001（eiam 不可用降级）✓ L108-109。 |

### 6. Edge Case Coverage — 99/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Error paths documented | 40/40 | 部分导入失败（L100-101）、403 无角色（L76）、404 后门清零（L211）、探针超时 3s（L205）、eiam 不可用降级（L108）。 |
| Boundary conditions | 34/35 | 并发：回滚双实例只读 503 防双写（L106）；空输入：diff=0 隐式覆盖零规则；非法角色值 null/空串/非标准→403（L33）；schema_version 不兼容中止（L199）。-1：会话在宽限期（2h）内过期的过渡场景未显式描述（若 session TTL > 2h 则不触发，但 spec 未声明 session TTL，隐式靠"未登录用户→只读视图"兜底）。 |
| Failure recovery described | 25/25 | 回滚步骤 + 回滚失败兜底（DNS 切换/双实例只读/人工确认）+ 部分导入增量重导 + eiam 恢复滞回机制。 |

### 7. Scope Clarity — 100/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope concrete deliverables | 35/35 | 每项是具体功能/接口/流程（源码纳入/一键启动/门禁/eiam 委托/角色绑定/降级/反代/切流/单点登出/会话过期/接入决策记录）。 |
| Out-of-scope named | 30/30 | 4 项显式列出：历史数据迁移、前端嵌入门户、eiam 增强、角色同步。 |
| Scope consistent with specs/stories | 35/35 | In Scope 项与 UF（3 个）、Story（6 个）、Related Changes 表逐一对应，无悬空项。 |

---

## Phase 3: Blindspot Hunt

经全面审查，未发现落在 rubric 维度之外的盲点。Phase 1 推理审计未标记任何维度评分无法吸收的根本性缺陷。

---

## Iteration 2 — 12 条修复验证

| # | 问题 | 状态 | 验证 |
|---|------|------|------|
| 1 | 会话架构矛盾（CRITICAL） | **FIXED** | spec L209「复用平台共享会话存储…不签发进程内自签名会话票」；Story 5 L84 同口径；Dependency Baseline L221「平台共享 Redis session=现有」。"进程内 HMAC 内存票 + 内存撤销列表" 设计已彻底移除，仅余否定表述。 |
| 2 | 隐性依赖未标现有/新建 + 缺 eiam 契约基线 | **FIXED** | Dependency Baseline L217-224 逐项标注（eiam/health=现有、role_binding=新建、Redis session=现有、密钥=现有、迁移接口=新建、绑定接口=新建）；eiam 契约基线 L226 含 5 工作日通知 + 影响评估。 |
| 3 | 回滚失败未覆盖 | **FIXED** | spec L106「回滚失败兜底：手动 DNS 切换…双实例只读 503…人工确认」；Story 6 L109 同口径。 |
| 4 | 单点登出缺 Mermaid | **FIXED** | spec L84-92 补图：登出→销毁共享会话→其余标签页 401→收敛跳 eiam 登录。 |
| 5 | UF-2/UF-3 状态缺口与矛盾 | **FIXED** | UF-2 补「已恢复」状态（L98）+「正常（探针偶发失败）」状态（L96）；UF-3 L115「恢复时由后端探针信号驱动前端自动跳转 eiam 登录页（非用户点击按钮入口）」。矛盾消除。 |
| 6 | Story 6 抽样口径未定义 | **FIXED** | Story 6 L98「样本量=max(≥10% 总数, ≥50 条)；抽样方法=随机覆盖 P0/P1/P2+各数据源类型；一致=同输入触发，时间差≤5s，内容字段完全相同」。 |
| 7 | 写操作状态码三文档不一致 | **FIXED** | spec L109「503」、UI L83「503」、Story 2 L41「503」三文档统一。 |
| 8 | 性能含糊 + 质量门禁不可度量 | **FIXED** | Goals L46「≥ 80% 覆盖率」；Performance L195「≥ 200 并发用户」「P95 回退 ≤ 50ms」。 |
| 9 | role_binding 非法值 + 预置接口未定义 | **FIXED** | spec L33「null/空串/非 admin/user→403」；L35 绑定管理接口 POST/DELETE/GET + 批量导入 JSON 文件。 |
| 10 | 只读视图脱敏缺失 | **FIXED** | spec L212 匿名可见字段清单 + 脱敏规则；UF-3 L129-131 同口径。 |
| 11 | 迁移导出/导入工具未定义 | **FIXED** | spec L199 接口/格式/schema_version/凭据重新加密流程完整。 |
| 12 | Story 6 部分导入失败未覆盖 | **FIXED** | Story 6 L100-101 部分失败 AC：中止批次+待确认标记+错误清单+增量重导/回滚全量+不进切流直至 diff=0。 |

**12/12 全部修复，无新矛盾引入。**
