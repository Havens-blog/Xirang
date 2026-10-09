# Eval-proposal — Baseline Report (Iteration 0)

**Eval-Proposal Complete**
**Final Score**: 762/1000 (target: 900)
**Iterations Used**: 1/3 (ITERATION=1, baseline; no prior report)

### Score Progression
| Iteration | Score | Delta |
|-----------|-------|-------|
| 0 (baseline) | 762 | — |

### Verdict (CTO persona)

提案小而聚焦，问题真实、方案方向正确，风险表和口径闭合意识高于平均水平。但有三个层面的问题阻止其接近 900 分：**(1)** 核心过滤谓词 `refCount=0` 与本仓库自身已固化的引用三态语义（`has_refs / no_refs_scanned / blind_spot`，见 `e-cam-service/internal/cert/domain/reference_status.go`："**未发现引用 ≠ 无引用**……blind_spot 表示无成功快照或扫描范围未覆盖，禁止据此放行删除"）直接冲突——`refCount` 在 DTO 中是 `int`，blind_spot 与 no_refs_scanned 均显示 0，按现提案写法 blind_spot 过期证书会被默认隐藏，而这恰是信息最少、最可能藏着漂移的行；**(2)** 行业对标一节是全文档最薄弱处，无任何具体产品/项目/模式引用；**(3)** 风险表没有任何可能性/影响评级，SC 把会腐烂的快照数字（11、10、1）写成了永久断言。

---

## Phase 1 — Reasoning Audit (pre-score anchors)

1. **Problem → Solution**: 方案（默认隐藏 expired+refCount=0 + 可切换 + 主动筛选豁免）确实针对"默认视图噪音与风险混排"的问题，方向成立。但 Problem 断言噪音"不可操作、无风险信号"，而 Evidence 自己承认被隐藏对象"非保护期，可删除"——它是台账里少数**有**可执行动作的行，前提与证据自相矛盾（见 Attack）。
2. **Solution → Evidence**: Evidence 给出了具体数字与命名证书（`*.jlc-robot.com`，-11 天，fingerprint_only），可核查性强（仓库中 `ledger_handler.go` / `listCertsApi` / daysLeft expired 分档 / commit 7cdd488 均已独立验证存在）。但 93/11/1 为单点快照，且提案自己也说"换证操作正在持续进行"——证据半衰期短，却未体现在 SC 设计上。
3. **Evidence → Success Criteria**: SC4 "11 行 = 看板「已过期」卡计数 11" 把当期快照硬编码为断言；换证完成数量变化后该 SC 永假或误导实现写死数据。相等关系（行数=卡计数）才是本值，具体数值不是。
4. **Self-contradiction check**:
   - NFR "性能：无新增查询" vs In Scope "响应返回隐藏计数"：隐藏计数需对同一过滤条件做一次不排除口径的统计（或单查询双 total），与"无新增查询"字面冲突。非致命，但会诱导实现者让前端推算计数——恰好回到风险表 #1 要杜绝的前端口径漂移。
   - **SC/InScope 一致性深挖（聚类 + 双向可满足性）**：SC1（默认隐藏）、SC2（切换）、SC3（有引用过期可见）、SC4（expired 筛选全显）、SC5（正交）、SC6（测试）与 In Scope 四条聚类后逐对双向推导，**未发现 mutual-exclusion 型硬矛盾**。但存在 **ambiguous 对**：SC2 "点击显示全部、再点恢复隐藏" ↔ SC4/InScope-3 "daysLeft=expired 时不排除"——用户先点"查看全部"再选"已过期"再清除筛选后处于哪个状态、切换态与筛选态优先级、分页是否重置，文档均未定义；标记为 `ambiguous — requires author clarification`（type: direction-clash 候选，尚不构成确认矛盾）。
   - 最严重问题不是 SC 内部矛盾，而是 **Solution↔代码库语义矛盾**（blindspot #1）：方案声称"被线上引用的过期证书始终显示（风险信号）"，但 `refCount=0` 谓词无法区分"已扫描确认为 0"与"从未扫描成功（blind_spot）"，后者恰恰可能是线上在用。

---

## Phase 2 — Dimension Breakdown

### 1. Problem Definition: 88/110
| Criterion | Score | Notes |
|-----------|-------|-------|
| Problem stated clearly | 34/40 | 问题（默认视图噪音与风险混排）明确，两名读者不会歧义；扣分点：「不可操作」的判定与自身证据冲突（见 Attack 2） |
| Evidence provided | 32/40 | 有具体数字与命名实例（93/11/1、`*.jlc-robot.com`），远好于"我们觉得"；但全部为单点快照、无用户反馈/工单类证据，"近 2 天手动换证/导入活动频繁"是无佐证断言 |
| Urgency justified | 22/30 | 诚实承认"当前噪音仅 1 张"，紧迫性建立在"未来会增多"的推测上；delay 成本只有一句"可操作性被稀释"，无量化 |

### 2. Solution Clarity: 98/120
| Criterion | Score | Notes |
|-----------|-------|-------|
| Approach is concrete | 34/40 | 服务端过滤+工具栏切换可复述；扣：切换态×筛选态状态机、组合过滤下隐藏计数是否按当前过滤上下文重算均未定义 |
| User-facing behavior described | 36/45 | Key Scenarios 六条覆盖主路径与空态；扣：切换后分页行为（停在非首页时）、空态下提示条是否可见、跨导航/刷新的切换态持久性未定义 |
| Technical direction clear | 28/35 | 指名了 ledger_handler.go / ledger_service / 前端 ledger 目录与 7cdd488，方向可执行；扣：daysLeft=expired 豁免落在后端强制还是前端不传参被写成待定语义（"需与 daysLeft/expired 联动的语义"），这是关键实现分叉，提案阶段应定死 |

### 3. Industry Benchmarking: 84/120
| Criterion | Score | Notes |
|-----------|-------|-------|
| Industry solutions referenced | 18/40 | 全节引用仅有"证书管理工具的过期证书归档/折叠"一句泛称，**无任何产品名/开源项目/已发表模式**；rubric 明确要求 "Product names, open-source projects, or published patterns" |
| At least 3 meaningful alternatives | 24/30 | 四个备选含 do nothing，均为真差异而非稻草人；扣：「默认隐藏全部过期」与「折叠/灰化」同属展示层弱化族，差异化边际 |
| Honest trade-off comparison | 21/25 | Pros/Cons 与 verdict 齐备且不 cherry-pick（承认本方案"需后端过滤参数 + 提示切换"成本） |
| Chosen approach justified against benchmarks | 21/25 | "Rejected: 丢风险"的拒绝理由绑定本项目约束（风险信号可见性），成立；扣：未说明为何不采纳行业标准做法的最小变体（如归档桶但保留引用>0 过期行） |

### 4. Requirements Completeness: 86/110
| Criterion | Score | Notes |
|-----------|-------|-------|
| Scenario coverage | 30/40 | 六场景含组合过滤与空态，好于均值；缺：切换×筛选交叉状态机、切换后分页重置、空态下提示条可见性 |
| Non-functional requirements | 30/40 | 口径一致/可发现性是本需求正确的关键 NFR；扣：「性能：无新增查询」与"响应返回隐藏计数"自相矛盾（隐藏计数需要第二个统计口径）；可发现性 NFR 未覆盖空态路径 |
| Constraints & dependencies | 26/30 | 后端/前端文件、7cdd488、与 daysLeft 联动均已点名；扣：未提及 refCount 派生依赖"最近 done 快照"的新鲜度窗口（风险表里倒提了一句） |

### 5. Solution Creativity: 65/100
| Criterion | Score | Notes |
|-----------|-------|-------|
| Novelty over industry baseline | 30/40 | 与"一刀切隐藏过期"的差异化（按 refCount 归零判定噪音）表述清晰且有价值，但属小步改良非创新 |
| Cross-domain inspiration | 15/35 | 无任何跨域借鉴（如邮件客户端"归档而非删除"、SRE 控制台"降级信号折叠但保留计数徽标"等模式均可借） |
| Simplicity of insight | 20/25 | 改动小而准，无过度工程；扣分因洞察本身依赖 refCount 单一信号，简化得略粗糙（见 blindspot 1） |

### 6. Feasibility: 78/100
| Criterion | Score | Notes |
|-----------|-------|-------|
| Technical feasibility | 32/40 | 已验证 `ListCerts` 现有 hostingStatus/daysLeft/search 服务端过滤与分页，扩展一个布尔参数完全可行；扣：`refCount=0` 谓词在现有三态模型（`blind_spot` 也为 0）下不能直接实现为提案描述的语义，需要先补 `ReferenceStatus` 维度 |
| Resource & timeline feasibility | 20/30 | 范围小、团队已熟悉该代码面（7cdd488 同域）；但全文无任何工作量/时间预估 |
| Dependency readiness | 26/30 | 全部依赖已存在并经独立验证（ledger_handler.go、ParseDaysLeftTier("expired")、commit 7cdd488 "fix(cert): 到期看板分级卡改跳转台账按档过滤"）；扣：refCount 快照新鲜度窗口未定义 |

### 7. Scope Definition: 72/80
| Criterion | Score | Notes |
|-----------|-------|-------|
| In-scope items are concrete | 26/30 | 四条均为可交付物（后端过滤+返回计数/前端切换/豁免/测试），可执行 |
| Out-of-scope explicitly listed | 23/25 | 四条有名有因（批量删除/告警/archive/看板页），边界干净 |
| Scope is bounded | 23/25 | 可在短周期内完成；扣：豁免语义"待定"一项实质是未闭合的范围项 |

### 8. Risk Assessment: 60/90
| Criterion | Score | Notes |
|-----------|-------|-------|
| Risks identified | 24/30 | 五条均非琐碎且直击要害（口径漂移/误藏/11≠11/静默吞行/refCount 语义）；缺三条：blind_spot 误判、组合过滤下计数重算、切换态状态机边界 |
| Likelihood + impact rated | 12/30 | **风险表只有 Risk/Mitigation 两列，无任何可能性/影响评级**——rubric 要求 "Is the assessment honest?" 的前提是有评级 |
| Mitigations are actionable | 24/30 | #1/#2/#3/#4 可执行（服务端过滤、单测锁定、联调断言）；#5 "refCount 来自最近 done 快照，语义稳定"是断言不是缓解动作，且与代码库三态语义不符 |

### 9. Success Criteria: 62/80
| Criterion | Score | Notes |
|-----------|-------|-------|
| Criteria are measurable and testable | 22/30 | 六条均可写成测试/清单；扣：SC4 "11 行 = …计数 11" 与 SC1 "（当前 1 张）" 把会腐烂的快照数字硬编码为永久断言（提案自证"换证操作正在持续进行"） |
| Coverage is complete | 21/25 | 覆盖全部 In Scope 项；缺：空态下提示条可见性、切换后分页行为无对应 SC |
| SC internal consistency | 19/25 | 聚类+双向推导未发现 mutual-exclusion 矛盾；SC2↔SC4 的切换态与筛选态优先级标记为 `ambiguous — requires author clarification` |

### 10. Logical Consistency: 69/90
| Criterion | Score | Notes |
|-----------|-------|-------|
| Solution addresses the stated problem | 26/35 | 对"噪音累积"成立；但对"风险可见性"承诺打折扣——blind_spot 过期证书（refCount 显示 0 但实为未知）会被误藏，方案部分重新引入了它声称要避免的"藏掉风险"问题 |
| Scope ↔ Solution ↔ SC aligned | 23/30 | 主体对齐；NFR "无新增查询" vs "响应返回隐藏计数" 存在实现层矛盾；SC 快照数字与"换证持续进行"的前提冲突 |
| Requirements ↔ Solution coherent | 20/25 | 映射干净、无孤儿需求；扣：Problem 前提"不可操作、无风险信号"与 Evidence"可删除"内在张力贯穿全文未被承认 |

**Cross-dimension coherence check**: Scope/Solution/SC 主干一致；最突出的跨维裂隙是 Solution 的核心谓词（refCount=0）在 D10/D6 中与代码库三态语义冲突，而风险表（D8）与 SC（D9）均未捕获——单一根因贯穿四维，修订时应一次性修掉。

---

## Phase 3 — Blindspot Hunt

1. **[blindspot] 核心谓词与本仓库已固化的引用三态语义直接冲突，blind_spot 过期证书会被误藏**（Reasoning audit flagged this independently of dimension scoring.）— 方案写"过滤条件严格限定 `expired AND refCount=0`"，但 `e-cam-service/internal/cert/domain/reference_status.go` 明文规定："**'未发现引用' ≠ '无引用'**——no_refs_scanned 仅表示'已扫描但无匹配'，可能因云 API 权限不足、产品未覆盖而漏报；blind_spot 表示无成功快照或扫描范围未覆盖，禁止据此放行删除"，且 DTO 注释确认 "RefCount int——has_refs 时为最新成功快照引用计数，**否则 0**"。即 refCount=0 同时覆盖"已扫描确认为零"与"从未扫描成功"两种状态，后者（blind_spot）是信息最少的过期证书，恰恰最可能藏着"线上在用"的漂移。提案声称"被线上引用的过期证书始终显示（风险信号）"在该谓词下**无法保证**。必须改进：过滤条件改为 `expired AND ReferenceStatus=no_refs_scanned`（或等价的"扫描已完成且显式为 0"），blind_spot/未知一律保守显示；并把 refCount=null/blind_spot 用例写进后端单测与 Risks 表。
2. **[blindspot] 全文无回滚计划** — 方案以 `status: Approved` 落地，In Scope 与 Risks 中均无任何关闭/回退开关的设计；一旦"默认隐藏"造成一次漂移漏看（本方案存在的唯一理由就是防漂移），唯一的恢复路径是发版回滚。引用："后端 `GET /certs` 新增过滤：排除「已过期且 refCount=0」的行"。必须改进：补一行缓解（如过滤参数可由前端 query 强制关闭 / 隐藏计数超阈值时告警或自动降级为不隐藏），并在 Risks 表中登记。
3. **[blindspot] Evidence 内部矛盾：被定义为"不可操作"的行恰是唯一可操作的清理对象** — Problem 断言无引用过期证"不可操作、无风险信号"，同一段 Evidence 承认该证书"非保护期，**可删除**"。隐藏它等于把台账里唯一的清理机会藏进"点击查看"二级入口，噪音治理（提升可操作性）与手段（隐藏可操作对象）方向相反。必须改进：至少在 Key Scenarios 中承认该取舍，或在提示文案中带上"可删除"属性（"已隐藏 N 张无引用过期证书（可删除）"），让被隐藏行的可操作性对用户可见。
4. **[blindspot] 方案文案自身漂移：同一按钮三个名字** — Solution 写"点击查看"（"工具栏显示行内提示「已隐藏 N 张无引用过期证书 · 点击查看」"），In Scope 写"查看全部"（"工具栏行内提示「已隐藏 N 张无引用过期证书 · 查看全部」"），Key Scenarios 又写提示变"再次点击隐藏无引用过期证书"。小缺陷，但文案是本方案的核心用户界面承诺，三处不一致会直接带进实现。必须改进：统一为一个 canonical 文案（含默认态/展开态两个字符串）。

---

## Outcome

**Target NOT reached** — 762/1000 < 900（baseline iteration 0；按协议修订后进入 iteration 1）。

### Revision priorities (for reviser)
1. （最高）过滤谓词改用 `ReferenceStatus=no_refs_scanned`，blind_spot 保守显示；补单测与风险条目（blindspot 1）。
2. 修 NFR 矛盾：把"无新增查询"改写为诚实的"同条件双统计口径（visibleTotal + hiddenCount，同一过滤上下文）"；定义隐藏计数按当前其余过滤条件重算。
3. SC 去硬编码：SC1/SC4 改为动态相等断言，不写死 11/10/1。
4. 定死 daysLeft=expired 豁免收敛层（建议后端强制豁免）+ 切换态×筛选态状态机 + 分页重置 + 空态提示条可见性。
5. 风险表补可能性/影响两列；补回滚/降级方案。
6. Industry Benchmarking 补具体产品/项目/模式引用（如 cert-manager、Venafi/Keyforge 类产品的过期证书归档行为、各云控制台过期证书默认折叠策略），并说明本方案相对行业基线的取舍。
