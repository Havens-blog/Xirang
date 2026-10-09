# 评分报告 — 平台控制台子项目提案（Iteration 1）

- 评分人：CTO 评分子代理（正式评分）
- 日期：2026-09-29
- 评分模式：**注释盲审**（annotated blind review，pre-revision 版本）
- 总分：**811 / 1000**（target = 900，未达标）
- 依据：scorer-protocol.md 三阶段 + cto.md 专家视角 + rubrics/proposal.md

---

## Phase 1 — 推理审计（Pre-Score Anchors）

对论证链的独立判断（先于 rubric）：

1. **问题 → 解决方案**：成立。四类失去归属的能力（登录/导航、身份治理、权限管理、工作台）逐一被方案承接，边界清理与 401 跳转切换属必要联动。ticket/change 的处理是「显式退役」而非承接——显式是优点，但退役决策无使用证据支撑（见 D10/A27）。
2. **解决方案 → 证据**：基本成立。证据均为项目内可核验事实（文件路径、端点名、提案计数），但「eiam 决定保留」这一全案地基性决策无来源（见 D1/A1）。
3. **证据 → 成功标准**：大部分 SC 可测（真实浏览器、双标签闭环、部署清单），但 SC-2 的 parity 基线与缺口处置门存在条件性矛盾（见 D9/CONTRADICTION-1）。
4. **自我矛盾检查**：方案未重新引入其声称消除的问题（无 BFF、无微前端、无双处维护）；「零后端」与「扩 eiam 端点」字面兼容且文档已自我标注。发现两处结构性缺口进入盲点清单（B1 下线交接缺失、B2 Phase 0 自身无计划）。

---

## Phase 2 — 各维度评分与扣分理由

### 1. Problem Definition — 99/110

| Criterion | Score | 理由 |
|---|---|---|
| Problem stated clearly | 38/40 | 问题表述无歧义：「去掉 Duke1616 的 ecmdb-web 之后，平台级的『统一入口 + 权限管理 + 身份治理 + 工作台』将失去归属」。两名读者不会读出不同含义。扣 2：四能力的边界（尤其「工作台」的内容边界）在问题层未定义，直到方案层才收敛为「纯 eiam 侧」。 |
| Evidence provided | 36/40 | 证据具体可核验：`views/iam/` 五个子域、`/user/{system,ldap,passkey}/login`、`views/platform/users` 孤立页面、「29 个提案中没有任何一个涉及」。扣 4：「eiam 决定**保留**」是全案地基性决策，但无决策来源、时间、授权方——verification stance 下视为未核验断言；除提案计数外无任何使用数据（多少平台用户依赖这些 UI）。 |
| Urgency justified | 25/30 | 成本延迟已说明：「越晚建控制台，越长时间依赖 Duke1616 前端」「成本是『去他依赖 / 100% 可控』目标延期」。扣 5：无 ecmdb-web 下线的目标日期或触发条件，「中」级紧迫度缺乏可核验的时间锚点。 |

### 2. Solution Clarity — 105/120

| Criterion | Score | 理由 |
|---|---|---|
| Approach is concrete | 38/40 | 方案可复述：独立纯前端子项目、直调 eiam、零新增后端、共享 Redis session、同域 nginx 反代路径宿主、五项能力清单。扣 2：迁出 `platform/users` 的具体动作（先迁后删还是直接删、路由如何处理）未述。 |
| User-facing behavior described | 37/45 | 关键场景覆盖登录、增删改查、单点跳转、会话过期、双向登出。扣 8：LDAP/passkey/MFA 的用户可感流程仅一句「沿用 eiam 已有能力并补 UI 入口」——用户在 MFA 挑战页看到什么、passkey 注册入口在哪，完全未描述；CRUD 过程中 API 失败的用户侧表现未定义。 |
| Technical direction clear | 30/35 | 技术方向足够清楚（直调、同域、共享 session、组件库锁版本）。扣 5：前端框架/构建/测试栈从头到尾未点名（仅「组件库随 e-cam-web 现状确定」——版本锁了，身份没锁）；直调的 API 面仅举 `/api/iam/user/*` 一例。 |

### 3. Industry Benchmarking — 107/120

| Criterion | Score | 理由 |
|---|---|---|
| Industry solutions referenced | 36/40 | 引用了真实方案：阿里云/腾讯云控制台、Keycloak Admin Console、Casdoor、Authing、qiankun/wujie。扣 4：全部为点名式引用，无任何机制级借鉴（云厂商控制台的 SSO 会话机制、Keycloak 控制台与 IdP 的部署关系均未展开）。 |
| At least 3 meaningful alternatives | 28/30 | 四个备选真实互异且非稻草人（每个都列了真实 pros：fork 方案「eiam API 集成现成」、微前端「单入口体验统一」）。扣 2：「并入 e-cam-web」这一明显备选只在 Assumptions Challenged 表中出现，缺席对比表，读者无法在同一框架下比较它。 |
| Honest trade-off comparison | 22/25 | pros/cons 基于项目真实约束（血统、解耦成本、页面量），选中方案自曝缺点「页面需逐个写」。扣 3：「血统与解耦成本不划算」的拒绝理由是价值判断而非成本量化——fork 后清理耦合的工作量与逐页重写的工作量从未对比。 |
| Chosen approach justified against benchmarks | 21/25 | 对微前端的拒绝（YAGNI）与对「100% 可控」的匹配论证清楚。扣 4：声称 Keycloak/Casdoor「管理模式与本提案一致」，但两者的管理控制台由 IdP 自身托管，本提案是独立静态站 + 独立仓——与所引基准的分歧（谁托管、谁部署、谁升级）未解释。 |

### 4. Requirements Completeness — 84/110

| Criterion | Score | 理由 |
|---|---|---|
| Scenario coverage | 27/40 | happy path、会话过期、无权限提示、eiam 端点缺失均已覆盖。扣 13：① eiam 不可用/降级场景完全缺失——控制台全部能力直调 eiam，eiam 宕机即整站不可用，但无任何降级或提示要求；② 多标签/并发会话边界（A 标签登出后 B 标签的下一跳行为）未识别；③ ticket/change 的现有使用方场景零分析即宣告退役（见 A10）。 |
| Non-functional requirements | 30/40 | 安全（不自造鉴权）、兼容（cookie domain/path）、可验证性（真实浏览器 + 同域宿主，硬性）均已点名；无障碍/多语言显式延后有理由。扣 10：① 性能 NFR 不可测：「首屏资源可缓存」无预算、无指标；② 登录页作为凭据入口的安全 NFR 缺失（CSP/安全响应头、CSRF 防护归属 eiam 还是控制台）；③ 浏览器兼容矩阵缺失。 |
| Constraints & dependencies | 27/30 | eiam API 与 session 依赖、中间件读取逻辑不动、宿主钉死、租户语义以核查结论为准、组件库锁版本——约束具体。扣 3：「eiam 必须保留且**稳定**」中「稳定」无定义（何版本、何 SLA）；nginx 反代配置的归属方（谁改、谁审）未点名。 |

### 5. Solution Creativity — 64/100

| Criterion | Score | 理由 |
|---|---|---|
| Novelty over industry baseline | 26/40 | 文档诚实自评：「模式本身无创新」。差异化已清晰表述（两处克制），但差异本质是纪律而非新意——「纯前端直调、零 BFF」是基准方案的自然执行而非超越。 |
| Cross-domain inspiration | 18/35 | Assumptions Challenged 表（Assumption Flip / Occam's Razor / 5 Whys）是方法论层面的真实思考痕迹；「缺口三类处置决策门」（UI 降级 / 扩端点 / 放弃）是可复用的治理模式。但无任何跨领域产品借鉴。 |
| Simplicity of insight | 20/25 | 「外壳只装治理能力，不装业务」的洞察简洁优雅，直接推导出 SSO 平级架构，无过度设计。扣 5：该洞察的正确性依赖「工作台聚合无价值」这一未验证判断（风险评估中仅给 H/L，未做任何用户侧询问）。 |

### 6. Feasibility — 68/100

| Criterion | Score | 理由 |
|---|---|---|
| Technical feasibility | 30/40 | 认识论卫生良好：「初步可行，**待 Phase 0 盘点确认**——两项前置核查未完成前不下『无 showstopper』结论」。扣 10：① 认证链路核查只有通过分支，无失败分支——若兼容矩阵发现不兼容（如中间件要求控制台无法提供的字段），文档无任何应对路径，而「后端改造」在 Out of Scope 中被整体排除；② 「eiam API 面覆盖度」若核查结果为中低覆盖，v1 范围如何收缩未预设。 |
| Resource & timeline feasibility | 18/30 | 「团队已熟悉 e-cam-web 同生态（现有 26 个 UI 重写提案均在此栈落地）」是合理的技能论证。扣 12：全文无任何时间估算——无人数、无人周、无里程碑、无 Phase 0 时长；rubric 要求「scope realistic for the proposed timeline」，而本提案没有 timeline 可言。 |
| Dependency readiness | 20/30 | eiam 在生产运行是真实就绪项；Phase 0 盘点内容具体。扣 10：① Phase 0 兼容矩阵只读「eiam session 写入逻辑与 EcmdbAuthMiddleware 读取逻辑」两处源码，而 SC-4 后半段「e-cam-web 登出后控制台会话同样失效」依赖 e-cam-web 自身登出链路（它登出时是否删除共享 Redis session）——这第三处源码不在盘点范围内；② 「扩 eiam 端点」分支「须单独确认排期」但无 owner、无确认机制。 |

### 7. Scope Definition — 71/80

| Criterion | Score | 理由 |
|---|---|---|
| In-scope items are concrete | 27/30 | 8 项均为交付物形态（可建任务）。扣 3：「身份源管理」的边界依赖 Phase 0 结果，交付形态未定（是 CRUD 页面还是只读列表）。 |
| Out-of-scope explicitly listed | 22/25 | 7 项显式且论证清楚，ticket/change 退役「不为其归宿兜底」是罕见的诚实边界。扣 3：两处措辞与 In Scope 存在张力——「云管业务演进（e-cam-web 保持现职）」与 In Scope 中对 e-cam-web 的两处修改（迁出 platform/users、401 跳转切换）字面冲突，需靠读者自行理解「保持现职 ≠ 不动 e-cam-web」。 |
| Scope is bounded | 22/25 | v1 页面集枚举、迁出设前置门、退役显式。扣 3：「逐项覆盖完成」的前置门中「覆盖」一词双义（实现覆盖 or 处置覆盖）——若指实现覆盖，则任一「放弃能力」处置将永久阻塞迁出（见 CONTRADICTION-1 推导链）。 |

### 8. Risk Assessment — 75/90

| Criterion | Score | 理由 |
|---|---|---|
| Risks identified | 24/30 | 5 个风险均实质（SSO 不兼容、API 面不足、迁出牵连、工作台空、视觉不统一）。扣 6：缺失至少四个实质风险——① eiam 团队带宽（「扩 eiam 端点」需另一团队排期承诺，无对应风险项）；② ticket/change 使用方反弹；③ 控制台成为 e-cam-web 登录链路的强依赖（401 切换后，控制台静态站不可用 = 全员无法登录），可用性耦合无风险项；④ 新登录页面的钓鱼/凭据安全面。 |
| Likelihood + impact rated | 27/30 | 评级分布诚实（M/H、M/H、M/M、H/L、M/M），「工作台早期内容空、价值弱 H/L」尤其坦率。扣 3：风险 1（SSO 不兼容）标 M 似偏低——「后端零改动」尚非核验结论，且一旦不兼容即为 showstopper 级，H 更合理。 |
| Mitigations are actionable | 24/30 | 风险 1 的缓解链具体可执行（兼容矩阵 → 照 ecmdb-web 现状口径 → 同域实测闭环）；风险 3 有前置门 + 灰度 + 回退。扣 6：风险 5 缓解「v1 先视觉对齐」无操作定义（对齐什么、验收标准是什么）；风险 2 缓解「缺口显式列单」止步于列单，处置决策门的执行者与决策时点未指明。 |

### 9. Success Criteria — 59/80

| Criterion | Score | 理由 |
|---|---|---|
| Criteria are measurable and testable | 25/30 | 多数 SC 附带验证手段（真实浏览器、双标签闭环、部署清单可核验）。扣 5：SC-7「云管功能无回退（走查通过）」的「走查」无清单定义，主观；SC-1 的验证分档清楚但待验项机制有漏洞（见下）。 |
| Coverage is complete | 17/25 | 扣 8：① In Scope 的「e-cam-web 会话过期/401 跳转目标切换」无独立 SC（仅被 SC-4 的登出分支部分覆盖——401 过期路径与主动登出路径是两条链路）；② Key Scenario「无权限操作 → 明确提示而非白屏」无任何 SC；③ 缺口三类处置的执行结果无 SC 验证（处置完的登记物在哪、谁验收）。 |
| SC internal consistency | 17/25 | 执行了 SC×SC 与 SC×InScope 全对推导（见下方一致性深查）。扣 8：CONTRADICTION-1（SC-2 ↔ 缺口处置门）成立；SC-1 待验项机制（「登记为待验项，不计入通过」）无退出条件/负责人/期限，形成可永久存续的未验证面，且与 In Scope「登录/登出（eiam system/LDAP/passkey/MFA 流程）」的交付承诺之间存在口径落差。 |

#### SC 一致性深查（聚类 + 组内双向推导）

**Cluster A（登录/会话域）**：SC-1, SC-3, SC-4, SC-5 × InScope(登录登出/导航切换/401 切换)
- SC-1 ↔ SC-5：兼容（SC-5 是 SC-1 的宿主口径加严）。
- SC-3 ↔ SC-4：兼容（单向跳入与双向登出互不排斥）。
- SC-4 ↔ InScope(401 切换)：兼容，SC-4 依赖该 InScope 项，推导成立。
- 结论：无矛盾。

**Cluster B（iam CRUD 域）**：SC-2, SC-7 × InScope(身份治理/权限管理/迁出前置门) × Requirements(错误情况处置)

```
CONTRADICTION: SC-2 "平台用户/角色/租户/授权可增删改查，能力不低于原 ecmdb-web 的 iam 管理页（按 Phase 0 盘点清单逐项覆盖）"
               ↔ 缺口处置决策门 "放弃能力（该操作不迁移，登记为功能损失）"（Requirements 错误情况同款）
| Type: conditional mutual-exclusion
| Evidence: 正向——若对任一清单项作出「放弃能力」或「UI 降级」处置，则该能力低于原 ecmdb-web 页面，SC-2 按字面不成立；
|           反向——若坚持 SC-2 成立，则「放弃能力 / UI 降级」两个处置分支对全部清单项被禁用，处置门的三选一承诺落空。
| 延伸链：放弃能力 → InScope 迁出前置门「逐项覆盖完成」若按「实现覆盖」解读则永久不满足 → 迁出被阻塞 → SC-7 不可达
|         → ecmdb-web 下线（本提案的存在动机）被阻塞。
| Resolution required: 二选一——(a) SC-2 基线改写为「经缺口处置后保留的能力集，逐项覆盖且不低于 ecmdb-web 对应能力」；
|                      (b) 显式承诺 parity 项一律走「扩 eiam 端点」处置，放弃/降级仅适用于清单外能力。
| Status: conditional contradiction — requires author clarification
```

**Cluster C（工作台域）**：SC-6 × InScope(工作台)：兼容。

**Cluster D（边界清理域）**：SC-7 × InScope(迁出 + 401 切换)：兼容（同为 e-cam-web 但行为不重叠；SC-7 走查可覆盖 401 切换的回归）。

**Cluster E（部署形态域）**：SC-8 × Out of Scope(BFF) × 处置门「扩 eiam 端点」：字面兼容（eiam 非新增服务）；与方案层「零后端纯前端」表述存在口径张力，文档已自我标注，低危。

### 10. Logical Consistency — 79/90

| Criterion | Score | 理由 |
|---|---|---|
| Solution addresses the stated problem | 31/35 | 四类失去归属能力逐一承接 + 边界清理。扣 4：Problem/Evidence 明确列出 ecmdb-web 承担 `ticket`/`change`，方案对这两个能力的回答是「有意退役」，但退役决策无使用数据、无使用方确认——从「问题识别」到「能力死刑」之间缺一环论证（不要求重建，要求证据或签字）。 |
| Scope ↔ Solution ↔ Success Criteria aligned | 26/30 | 主链对齐：In Scope 各项均有 SC 或被 SC 覆盖，迁出前置门与 SC-2 引用同一 Phase 0 清单。扣 4：CONTRADICTION-1 的延伸链（放弃能力 → 迁出阻塞 → SC-7 不可达）暴露 Scope 与 SC 在处置分支上未对齐；「e-cam-web 保持现职」措辞张力（同 D7）。 |
| Requirements ↔ Solution coherent | 22/25 | 需求到方案映射基本干净，401 切换作为联动改动显式入范围是亮点。扣 3：「无权限操作 → 明确提示」为孤儿需求（方案无承载点、SC 无验证项）；「LDAP/passkey/MFA 补 UI 入口」在 In Scope 交付，但无对应 SC 验证入口可用性（仅冒烟到「流程存在」）。 |

---

## Phase 3 — Blindspot Hunt

### B1. [blindspot] 「去 ecmdb-web」交接条件缺失 — 下线就绪无定义

Reasoning audit flagged this independently of dimension scoring.

本提案自证紧迫性的唯一理由是它是下线的**前置依赖**：「一旦下线 ecmdb-web，平台身份/权限管理 UI 直接断供」「越晚建控制台，越长时间依赖 Duke1616 前端」。但全文既未说明 ecmdb-web 前端下线属不属于本提案范围（In Scope 无、Out of Scope 亦无对应条目），也没有任何 SC 定义「控制台做到什么程度，ecmdb-web 才可安全下线」。方案解决了能力断供，却对它存在的目的（解锁下线）不设完成判据——提案完成后，「能不能下线 ecmdb-web」仍然是另一个没人回答的问题。必须改进：增加一条下线就绪 SC（如「控制台对 ecmdb-web 平台级入口全量覆盖并经真实浏览器验收后，ecmdb-web 前端可退役」），并在 In/Out of Scope 中显式交代下线动作归属。

### B2. [blindspot] Phase 0 成为无计划的承重墙 — 全文 11 处悬置

Reasoning audit flagged this independently of dimension scoring.

逐条扣分已在 D6/D8/D9 体现，但存在一个跨维度、单维度评分不可见的结构性模式：Phase 0 被引用 11 处，是「后端零改动」承诺、同域 cookie 前提、租户语义、端点差集、迁出前置门、SC-2 验收基线、风险 1/2 缓解方案的共同前置——而 Phase 0 自身无 owner、无时长、无产出模板、无退出判据、无停滞处置。「以 Phase 0 事实核查……结论为准」「待 Phase 0 盘点确认」「按 Phase 0 盘点清单逐项覆盖」——文档的认识论诚实（值得肯定）同时成为承诺的替代品：多个小节以「Phase 0 会决定」收尾而非给出决策框架。单个维度里每一处都看似合理，合起来意味着提案的核心承诺集体处于待定态而没有任何力量推动其收敛。必须改进：为 Phase 0 定义产出物清单、负责人、时间盒与「核查失败 → 提案如何重定范围」的退出路径。

---

## Bias Detection Report

- Annotated regions: 10 attack points / 20 paragraphs = density 0.50
- Unannotated regions: 16 attack points / 56 paragraphs = density 0.286
- Ratio (annotated/unannotated): 1.75

解读：注释区攻击密度约为非注释区的 1.75 倍。方向符合预期（标注区即预修订着力区，是承重段落），倍数温和，未见对注释区的过度猎杀——10 个注释区攻击中 6 个为「修订引入新门/新分支但未向下 reconcile」类型（A15 无失败分支、A17 盘点漏第三处源码、A20 与 Out of Scope 冲突、A23 处置门与 SC-2 未同步、A24 修订新增项未配 SC、A25 待验项无退出），正是盲审规则要求聚焦的「修订是否引入新问题」。全部 26 个攻击点中 **0 个**标记 `conflict-with-pre-revision`：所有攻击均为修订方向的延伸要求（要退出条件、要失败分支、要 reconcile），无一处主张回退修订内容。

---

## 攻击点清单（完整）

| # | 维度 | 区域 | 攻击点 | 引文 | 改进要求 |
|---|---|---|---|---|---|
| A1 | D1 | 非注释 | 「eiam 决定保留」无决策来源 | 「eiam 决定**保留**，但它没有自己的前端」 | 补决策出处/时间/授权方，或降级为待确认假设 |
| A2 | D1 | 非注释 | 紧迫度无时间锚点 | 「**中**：不是数据正确性缺陷，但这是『去 ecmdb-web』重构的前置依赖」 | 给出下线目标窗口或触发条件 |
| A3 | D2 | 非注释 | LDAP/passkey/MFA 用户可感流程未描述 | 「沿用 eiam 已有能力并补 UI 入口」 | 补三种流程的用户侧交互描述 |
| A4 | D2 | 注释 | 前端栈未点名（修订锁了版本没锁身份） | 「组件库随 e-cam-web 现状确定」 | 点名框架/构建/测试栈 |
| A5 | D3 | 非注释 | 基准点名为零机制借鉴 | 「对标云厂商控制台、Keycloak/Casdoor 等身份服务自带管理控制台」 | 至少对 1 个基准做机制级展开 |
| A6 | D3 | 非注释 | 与所引基准的部署分歧未解释 | 「管理模式与本提案一致」 | 说明独立静态站与 IdP 托管控制台的差异及取舍 |
| A7 | D3 | 非注释 | 「并入 e-cam-web」缺席对比表 | Assumptions Challenged：「直接并入 e-cam-web 最省事 … Overturned」 | 纳入对比表统一比较 |
| A8 | D4 | 非注释 | eiam 不可用场景缺失 | （无对应条目——全案直调 eiam） | 补降级/提示要求 |
| A9 | D4 | 非注释 | 多标签/并发会话边界缺失 | （Key Scenarios 无对应条目） | 补跨标签会话行为定义 |
| A10 | D4 | 注释 | ticket/change 退役无使用证据 | 「随 ecmdb-web 下线一并消失，控制台不重建、本提案不为其归宿兜底」 | 补使用数据或使用方确认 |
| A11 | D4 | 非注释 | 登录页安全 NFR 与浏览器兼容缺失 | （NFR 节无对应条目） | 补 CSP/CSRF 归属与兼容矩阵 |
| A12 | D4 | 注释 | 性能 NFR 不可测 | 「首屏资源可缓存」 | 给出可测预算/指标 |
| A13 | D5 | 非注释 | 新颖度自评即结论 | 「模式本身无创新」 | 差异化需超出「执行纪律」层面才算创造 |
| A14 | D5 | 非注释 | 无跨领域借鉴 | （Assumptions 表仅方法论） | — |
| A15 | D6 | 注释 | Phase 0 无失败分支 | 「核查通过前不将该承诺视为成立」 | 定义核查失败后的重定范围路径 |
| A16 | D6 | 非注释 | 全文无时间估算 | 「团队已熟悉 e-cam-web 同生态 … 无新技能门槛」 | 补人周/里程碑/Phase 0 时长 |
| A17 | D6 | 注释 | Phase 0 盘点漏 e-cam-web 登出链路 | 「读 eiam session 写入逻辑与 `EcmdbAuthMiddleware` 读取逻辑源码」 | 纳入 e-cam-web 登出行为核查（SC-4 后半段的机制前提） |
| A18 | D6 | 注释 | 「扩 eiam 端点」无 owner/确认机制 | 「须单独确认排期」 | 指明确认人与时点 |
| A19 | D7 | 非注释 | 「保持现职」与两处 e-cam-web 改动字面冲突 | 「云管业务演进（e-cam-web 保持现职）」 | 改写为「不承载平台治理职责，边界清理类改动除外」 |
| A20 | D7 | 注释 | 共享 token 与「仅视觉对齐」张力（修订引入） | 「主题 token 以共享文件形式引用（非复制粘贴）」 × 「共享组件库 / 设计系统沉淀——另立提案（v1 仅视觉对齐）」 | reconcile 两处口径 |
| A21 | D8 | 非注释 | 缺 4 项实质风险（eiam 带宽/退役反弹/可用性耦合/登录页安全面） | （风险表无对应行） | 补风险项 |
| A22 | D8 | 非注释 | 缓解措施不可操作 | 「v1 先视觉对齐」 | 给出对齐的操作定义与验收 |
| A23 | D9 | 注释 | SC-2 ↔ 处置门条件性矛盾（含迁出阻塞延伸链） | 「能力不低于原 ecmdb-web 的 iam 管理页」 × 「放弃能力（该操作不迁移，登记为功能损失）」 | 按 CONTRADICTION-1 二选一改写 |
| A24 | D9 | 注释 | 401 切换与「无权限提示」无 SC | 「e-cam-web 会话过期/401 跳转目标切换为控制台登录页」 | 补两条独立 SC |
| A25 | D9 | 注释 | 待验项无退出机制 | 「登记为待验项，不计入通过」 | 补期限/owner/清零机制 |
| A26 | D10 | 非注释 | ticket/change 为孤儿能力（问题识别→能力退役之间无论证环） | 「ecmdb-web 承担 `login` … + `navigation` + `ticket`/`change`」 | 同 A10，补证据链 |
| A27 | D10 | 非注释 | 孤儿需求：无权限提示无方案承载点 | 「无权限操作 → 明确提示而非白屏」 | 指明承载组件并配 SC |
| B1 | blindspot | 非注释 | 下线交接无定义、下线就绪无 SC | 「一旦下线 ecmdb-web，平台身份/权限管理 UI 直接断供」 | 补下线就绪判据与范围归属 |
| B2 | blindspot | 混合 | Phase 0 无计划承重墙（11 处悬置） | 「以 Phase 0 事实核查……结论为准」 | 补 owner/时间盒/退出路径 |

---

## 结论

**811/1000，低于 target 900。**

本提案的认识论诚实（假设显式化、Phase 0 门控、能力自评无创新、退役显式化）是同类文档中少见的优点，问题定义与方案骨架扎实。距 900 的差距集中在五类可修复问题：

1. SC-2 与缺口处置门的条件性矛盾（含迁出阻塞延伸链）——最高优先；
2. Phase 0 无 owner/时长/失败分支，11 处承诺集体悬置；
3. 全文无时间估算；
4. ticket/change 退役与「eiam 决定保留」缺证据链；
5. 下线就绪判据缺失（提案为下线而生，却不定义下线何时可发生）。

修订后预计主要失分项均可收敛至 900 以上。
