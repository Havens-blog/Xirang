# Scorer Report — Iteration 2 (Annotated Blind Review)

**Document**: `D:\Haven\docs\proposals\cdn-cache-analysis\proposal.md`
**Rubric**: proposal.md (1000 pts, target 900)
**Scorer persona**: CTO
**Iteration**: 2 (previous report: `eval/iteration-1.md`)
**Date**: 2026-09-21

---

## Phase 1 — Reasoning Audit (pre-score anchors)

### Argument chain trace

1. **Problem → Solution**: 分析卡(手动触发聚合 + 纯函数规则引擎 + AI 后置)直接回答"命中率 / 失效域名 / 回源浪费 / 优化建议"四问,对齐良好,无"换题"。

2. **Solution → Evidence**: 证据(四源 cache_hit 已归一、/aggregate 已支持、1453万/h 实测)支撑"数据面零新增采集"。修订后 Evidence 诚实标注"当前命中率基线可得、可先行用其出按域名基线"、"金额以未命中流量估算为主、不夸大",不再承诺交付后才有的口径。

3. **Evidence → Success Criteria**: SC 与 In Scope 粒度现已对齐(URI TOP 改全局 + 标注归属域名、状态码口径改"分布")。SC 测在正点上。

4. **Self-contradiction check — 修订引入 2 处新矛盾**:
   - **新矛盾 A(核心创新口径)**: Innovation(L31)"按 `cache_hit×status` 剔除 4xx/5xx"要求 **cache_hit×status 二维交叉**,而 L44 / IS1(L107)/ SC2(L137)三处均声明该交叉"单维度 /aggregate 做不到 / 本期不承诺 / 列为后续增强"。上一轮攻击 2(可缓存剔除无操作定义)与攻击 3(状态码口径不可达)的修复方案**互相打架**:为修攻击 2 引入的剔除机制,恰依赖被攻击 3 的修复方案(L44 改交付"状态码分布"、声明交叉不承诺)所否定的交叉。
   - **新矛盾 B(前端卡残留)**: IS3(L109,未标注区)仍承诺卡面"状态码命中率",而修订后后端/SC 交付的明确是"状态码分布(4xx/5xx 占比)",状态×cache_hit 交叉已移出本期。修订改了后端/SC 未同步改前端卡文案。

5. **核心逻辑风险(pre-score anchor,转入 blindspot)**: "未命中 URI TOP"依赖对 1453万请求/h 窗口做"查询串归一后单维度 ×uri 聚合",URI 高基数下 SLS GROUP BY 分组上限 / 性能未在文档中论证,全局 Top N(默认 5)在大基数下可能被截断/采样而不自知。

### SC Consistency Deep-Dive(复查修订后,含新矛盾)

区域聚类与逐对双向可满足性检查:

- **集群 A — 后端聚合/回归**(IS1, IS5, SC1, SC3, SC4): 通过。
  - SC1(≤2 帧 + N 上限 + 扫描量预估/拦截)↔ IS1 并发多维度聚合 ↔ L50(N≤5、24h、超限拦截):一致。
  - SC3(可复现,同边界 + 同批次并发)↔ IS1:一致;文档自限"可复现断言限定于此"。
  - SC4"回归通过" ↔ IS5(回归测试交付,含跨源/跨层对照 + golden 回放):**上轮孤儿已消除**。
- **集群 B — 规则引擎/输出**(IS1, IS2, SC2, SC5):
  - SC2 字节双口径 ↔ IS2/Innovation(L31 现定义两口径:全请求 = hit+partial 字节/总字节;可缓存 = 可缓存请求 hit+partial 字节/可缓存请求总字节):**上轮"双口径未定义"已解决**。
  - SC2 URI TOP 全局 + 标注归属域名 ↔ IS1 单维度 ×uri(不做 host×uri 交叉):**上轮粒度矛盾已解决**。
  - **新矛盾(SC2↔IS1)**:
    ```
    CONTRADICTION: SC2 "可缓存两档…双口径" ↔ IS1 "状态×cache_hit 二维交叉为后续增强" / L44 "单维度 /aggregate 做不到"
    | Type: direction-clash(机制依赖被同一文档声明的不可用能力)
    | Evidence: 满足 SC2 可缓存口径需按 Innovation(L31)"按 cache_hit×status 剔除 4xx/5xx"执行 → 需要 cache_hit×status 二维交叉;而 L44/L107/L137 均声明该交叉本期不承诺/做不到。满足 IS1(无交叉)则 L31 剔除机制无输入;满足 L31 机制则需在 IS1 例外加交叉查询。双向互斥。
    | Resolution required: 明确"可缓存剔除"的实际数据来源 —— 要么该特定聚合查询例外支持 cache_hit×status(需改写"二维交叉为后续增强"),要么改用不依赖交叉的近似剔除(如仅按 cache_hit=error 剔除 + 状态分布旁注)并重写 L31 表述
    ```
  - IS2 健康档位(优/中/差)↔ SC2 ↔ SC5 单测:一致。
- **集群 C — 前端卡**(IS3, SC1, SC2):
  - IS3 卡面"状态码命中率" ↔ SC2"状态码分布" / L44"默认交付为当前窗状态码分布单独展示":**失配(新)**。卡面描述的能力后端明确不交付。
  - IS3 其余项(折叠/手动触发/前窗对比/双命中率/域名排行/未命中 URI/优化项可复制/AI 解读区)↔ SC1/SC2:一致。
- **集群 D — 隐私/回滚 NFR**(L54, L55): 无对应 SC 验收项(见 Phase 3 / 攻击 5)。

---

## Phase 2 — Rubric Scoring (verification stance)

### 1. Problem Definition (110)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Problem stated clearly | 40/40 | "当前命中率多高(请求/字节)、哪些域名/URI 缓存失效最重、回源流量浪费多少、该怎么做缓存优化"无歧义。 |
| Evidence provided | 36/40 | 数据可用性证据强(四源归一、/aggregate 支持、实测请求量),修订后诚实标注量化边界("当前命中率基线可得…可先行用其出按域名的请求命中率基线…金额以未命中流量估算为主、不夸大")。扣分(残余):既然 /aggregate 已支持 cache_hit,文档仍只给请求量(1453万/h)而未给出任一当前命中率/未命中流量实测数,"基线可得/可先行"仍是承诺而非数据。 |
| Urgency justified | 25/30 | "命中率每低 1 个点,热点域名的回源带宽/流量费都会放大"(L22)"放大"仍未量化;但"数据面已齐 + 即日可用 + 手动触发成本可控"给出即时理由。 |

**Subtotal: 101/110**

### 2. Solution Clarity (120)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Approach concrete | 37/40 | 端点 `POST /logs/cache-analyze`、规则引擎 cdn-cache、N≤5 维度组、前窗 1 帧、SWR、per-source 降级、AI env 门控、可信度分级均具体。扣分:可缓存口径剔除机制(L31 用 cache_hit×status)与聚合能力声明矛盾,机制仍不落地。 |
| User-facing behavior | 44/45 | 内嵌卡、手动触发、前窗对比(标注"对比基期/基于 X/Y 源")、双命中率、域名排行、未命中 URI、状态码分布、可复制优化项、AI 区,可完整复述。 |
| Technical direction clear | 32/35 | 复用 diagnose 编排骨架、llm 包、/aggregate 联邦通道、查询串归一、golden 回放、单一归一映射源,方向明确。扣分:可缓存剔除的聚合输入路径未指明(见矛盾 A)。 |

**Subtotal: 113/120**

### 3. Industry Benchmarking (120)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Industry solutions referenced | 38/40 | 具名:阿里云 CDN/Akamai 控制台报表、Grafana+云指标、纯 LLM 分析。 |
| ≥3 meaningful alternatives | 28/30 | Do nothing + 云控制台 + 时序监控 + 纯 LLM,4 条,含行业验证方案,非稻草人。 |
| Honest trade-off | 22/25 | Pros/Cons 基于真实约束(跨云割裂、不下沉 URI、不可审计)。 |
| Chosen justified | 21/25 | "判据确定可靠,AI 解读后置"对比纯 LLM 的可审计性,理由成立。 |

**Subtotal: 109/120**

### 4. Requirements Completeness (110)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Scenario coverage | 36/40 | 覆盖:健康档、未命中集中、大/小文件未命中、4xx/5xx、前窗对比、错误场景,面广。修订后:可缓存剔除有近似判据且诚实标注口径边界(含 POST/认证/Range 无法剔除的声明),状态码口径改为"分布"并显式不承诺二维交叉。扣分(残余):可缓存剔除的"按 cache_hit×status"与"二维交叉本期不承诺"自相矛盾,头号场景的可缓存口径数据来源仍未闭合。 |
| Non-functional requirements | 37/40 | 成本(N≤5 + 扫描量预估 + 超限默认拦截确认后放行)、覆盖率≥80%、可复现、per-source 状态、AI 隐私(外发剥离 query/PII)、feature flag + 灰度 + 回滚(新增)均到位,上轮四项 NFR 缺口全部补齐。 |
| Constraints & dependencies | 25/30 | /aggregate 依赖、单一归一映射源、SLS 计费手动触发均点名。扣分:约束声明"单维度 /aggregate 做不到状态×cache_hit"(L44)与 Innovation 剔除机制(L31)依赖该交叉相冲突,约束不自洽。 |

**Subtotal: 98/110**

### 5. Solution Creativity (100)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Novelty over baseline | 36/40 | 全请求/可缓存双档 + 请求/字节双口径 + 结构化优化项(域名/URI 前缀+动作+miss 流量占比+可信度分级),超出云控制台与 Grafana。 |
| Cross-domain inspiration | 30/35 | 复用内部 WAF 诊断卡交互/编排(跨产品借鉴)。 |
| Simplicity of insight | 23/25 | "请求 vs 字节命中率 gap = 回源带宽浪费"直观;可缓存口径改为"聚合层近似 + 显式标注口径边界"比上轮"枚举剔除"更诚实、更简洁。 |

**Subtotal: 89/100**

### 6. Feasibility (100)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Technical feasibility | 35/40 | 数据面已实测(DCDN hit_info 聚合、eo.go 映射、/aggregate 生产可用),无 showstopper。扣分:(a) 可缓存口径机制依赖被声明"做不到"的二维交叉;(b) URI 高基数聚合在 1453万/h 下的 GROUP BY 上限/性能未论证(见 blindspot)。 |
| Resource & timeline | 26/30 | 修订后 2-3 天并给出分项(1 天端点+引擎+单测、0.5-1 天前端卡、0.5-1 天跨源对照与回归),与交付面(新端点+双口径+可信度分级+前端新卡+跨源回归)基本匹配,仍偏紧但可执行。 |
| Dependency readiness | 29/30 | 四源已映射、llm 包已落地、WAF 骨架已落地、数据流正常,无上游阻塞。 |

**Subtotal: 90/100**

### 7. Scope Definition (80)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| In-scope concrete | 29/30 | 6 条均交付物级(端点、引擎、卡、降级标注、回归测试、summary 字段)。 |
| Out-of-scope explicit | 24/25 | diagwatch、一键下发、独立页/历史趋势、UA/边缘粒度、WAF/SLB、host×uri 交叉、状态×cache_hit 交叉,均具名。 |
| Scope bounded | 23/25 | N≤5、Top N 封顶、窗口上限、feature flag 默认关,可执行。扣分:IS3 卡面"状态码命中率"越出后端交付范围(后端只出"状态码分布"),范围项内部失配。 |

**Subtotal: 76/80**

### 8. Risk Assessment (90)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Risks identified | 29/30 | 6 条(缺源、partial 上偏、扫描成本、口径漂移、bytes_sent 归一、优化建议误判),全部有意义;上轮"建议误判无护栏"已入表。 |
| Likelihood + impact | 28/30 | M/M、M/M、M/M、M/M、L/M、M/H,分布诚实,含高风险高影响项。 |
| Mitigations actionable | 27/30 | 可信度分级 + "建议,执行前请验证" + 低可信度折叠 + 无一键下发 + 扫描预估/拦截 + 跨源测试,均可执行。扣分:无针对"可缓存剔除机制不可达"与"URI 高基数聚合"的风险条目。 |

**Subtotal: 84/90**

### 9. Success Criteria (80)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Measurable/testable | 28/30 | 入口、≤2 帧、窗口上限、内容项存在性、固定边界可复现、单测≥80%,均可写测试。 |
| Coverage complete | 23/25 | SC1-5 覆盖 IS1/2/4/5/6。缺口:IS3 卡面"优化项可复制/AI 解读区"无 SC 对应;L54 AI 隐私、L55 开关/回滚等安全相关 NFR 无 SC 验收项。 |
| SC internal consistency | 21/25 | 见深潜:SC2 可缓存口径 ↔ IS1 无二维交叉(direction-clash);IS3"状态码命中率" ↔ SC2"状态码分布"失配。上轮三处矛盾(双口径、每域名 URI、回归孤儿)均已解决。 |

**Subtotal: 72/80**

### 10. Logical Consistency (90)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Solution addresses problem | 34/35 | 分析卡直接回答四问。 |
| Scope↔Solution↔SC aligned | 21/30 | 新矛盾 A(可缓存剔除依赖被"不承诺"的二维交叉,Innovation ↔ L44/IS1/SC2);新矛盾 B(IS3"状态码命中率" ↔ L44/SC2"状态码分布")。上轮漂移/分位数/每域名 URI 三处已解决。 |
| Requirements↔Solution coherent | 23/25 | 需求基本可追溯;残余:可缓存双口径需求与聚合输入(无交叉)未闭合,字节口径的可缓存分母同样继承该缺口。 |

**Subtotal: 78/90**

### Cross-dimension coherence check

新矛盾 A 在 D4(约束/场景)、D6(技术可行性)、D9(SC 内部一致)、D10(范围↔方案↔SC)四维同源共振 — 是本期主扣分。新矛盾 B 在 D7/D10 共振。上轮攻击 1-6、8-12 对应缺陷均已实质解决。

**Total: 101+113+109+98+89+90+76+84+72+78 = 910/1000** (target 900 — **达标**)

---

## Bias Detection Report

标注区(`<!-- pre-revised -->`)= 29 段;未标注区 ≈ 18 段。
- 标注区攻击点:2 段内攻击(矛盾 A、URI 基数 blindspot)= 密度 **0.07**
- 未标注区攻击点:2 段内攻击(IS3"状态码命中率"失配、SC 覆盖缺口)= 密度 **0.11**
- 比值(标注/未标注)= **0.64**

结论:标注区密度低于未标注区,但**不构成对修订区的宽纵**——原因是(1) 上轮 12 条攻击中绝大多数已实质解决,修订区自然少攻击;(2) 标注区仅剩的 2 条攻击均为**修订本身新引入**的问题(攻击 2/3 的修复方案互斥产生矛盾 A;对未标注的 IS3 未同步更新产生矛盾 B),而非重提已修复项,符合"修订引入新问题才提"的规则;(3) 未标注区 2 条攻击中一条(IS3 失配)系修订在别处改口径而漏改 IS3 所致,属修订盲区而非对修订区的纵容。未发现对已修订区域的刻意放水。

---

## Phase 3 — Blindspot Hunt

以下问题不落入任何 rubric 维度的显式判据,但属 CTO 识别的高风险失败模式:

- **[blindspot] URI 高基数聚合的 SLS GROUP BY 上限/成本未论证**: 本期旗舰输出"未命中 URI TOP"依赖对全窗口做"查询串归一后单维度 ×uri 聚合"。在证据自报的规模下(CDN aliyun 1h 全窗 1453 万请求),URI 基数可达百万级;SLS GROUP BY 有分组数上限(默认返回组数有限)且高基数 GROUP BY + ORDER BY 会显著推高扫描/计算成本,"全局 Top N(默认 5)"在大基数下可能被截断/采样,TOP 5 未必真实代表最重 miss,且文档无任何 URI 基数上限 / 截断检测 / 降级策略。quote: IS1"×uri(查询串归一后的 URI 未命中,单维度聚合)"(L107)与 SC2"未命中 URI TOP(查询串归一后,全局 Top N(默认 5))"(L137),规模引自 L17"CDN aliyun 1h 全窗 **1453 万请求**"。文档只给"扫描量预估/拦截",未给 URI 基数/分组上限的处理路径。需定义:高基数聚合的 GROUP BY 上限设定、miss 预筛(如先按 host 过滤再聚合 URI)、超限降级或采样标注。
- **[blindspot] 可缓存双口径的输入不可达风险被低估(Reasoning audit 独立于维度评分发现)**: 矛盾 A 的实质后果是"可缓存请求/可缓存字节"两个口径的计算输入在当前聚合能力下悬空 —— 这不是文案瑕疵而是核心指标可行性问题。Phase 2 的 D4/D6/D10 已按维度扣分,此处作为推理审计独立标记,防止被分散扣分掩盖其严重性。quote: L31"按 `cache_hit×status` 剔除 4xx/5xx 与 `cache_hit=error`" vs L44"单维度 /aggregate 做不到,列为**后续增强、本期不承诺**"。

---

## Attack List (all with quotes)

1. **[10-Logical Consistency / 4-Requirements / 9-SC]** (`conflict-with-pre-revision`): 可缓存口径剔除机制自相矛盾 — "按 `cache_hit×status` 剔除 4xx/5xx 与 `cache_hit=error`"(L31) vs "需"状态 × cache_hit"二维交叉,单维度 /aggregate 做不到,列为**后续增强、本期不承诺**"(L44)/ "状态×cache_hit 二维交叉为后续增强"(IS1, L107) — 上轮攻击 2/3 的修复方案互斥:为修攻击 2 引入的剔除机制恰依赖被攻击 3 的修复(声明交叉不承诺)所否定的能力。需统一:该聚合查询例外支持 cache_hit×status(改写"后续增强"),或改用不依赖交叉的近似剔除并重写 L31。(标注区;修订新引入)

2. **[10-Logical Consistency / 7-Scope]**: 前端卡仍承诺"状态码命中率",后端/SC 只交付"状态码分布" — "状态码命中率"(IS3, L109) vs "状态码口径**默认交付为当前窗状态码分布单独展示**"(L44)/ "状态码分布(4xx/5xx 占比;状态×cache_hit 二维交叉列为后续增强)"(SC2, L137) — 修订改了交付物未同步改卡文案。需将 IS3 改为"状态码分布"或注明仅分布展示。(未标注区;修订引入的失配)

3. **[blindspot]**: URI 高基数聚合可行性未论证 — "×uri(查询串归一后的 URI 未命中,单维度聚合)"(L107) + "CDN aliyun 1h 全窗 **1453 万请求**"(L17) + "未命中 URI TOP(查询串归一后,全局 Top N(默认 5))"(L137) — 百万级 URI 基数下 SLS GROUP BY 上限/截断/成本未评估,Top 5 miss 可能失真。需定义高基数处理(分组上限、miss 预筛、超限降级)。(标注区)

4. **[1-Problem/Evidence]**(低): 当前命中率/未命中流量基线仍为承诺非实测数据 — "当前命中率基线可得:`/aggregate` 已支持 `cache_hit` 维度,可先行用其出按域名的请求命中率基线"(L17) — 既然能力已具备,建议在 Evidence 直接给一个实测基线数值(如当前按域名请求命中率或未命中流量占比)支撑"失效面大"。(标注区;部分解决,残余缺口)

5. **[9-SC coverage]**(低): 安全相关 NFR 无 SC 验收项 — "AI 外发隐私:…外发前剥离 query 参数与可能的 PII"(L54)/ "开关与回滚(灰度)…异常时关闭开关即回滚"(L55) vs SC1-5 均无对应验收 — 隐私/回滚是本期关键安全属性,SC 未承接,存在"交付即无法客观验收"风险;IS3 卡面"优化项可复制/AI 解读区"亦无 SC。(未标注区)

---

## Outcome

**Final Score: 910/1000** (target 900) — **达标**。上轮 12 条攻击中 10 条实质解决(双口径定义、可缓存近似、状态码口径诚实化、N 上限与拦截、漂移/分位数来源、每域名 URI 粒度、工期复核、回归交付、建议护栏、AI 隐私、回滚开关)。本轮新引入 2 处内部矛盾(可缓存剔除依赖被"不承诺"的二维交叉;IS3"状态码命中率"残留)+ 1 个盲区(URI 高基数聚合)。优先修订:攻击 1(可缓存口径数据来源闭合,是最实质问题)、攻击 2(IS3 文案同步)、攻击 3(URI 基数处理)。
