# Scorer Report — Iteration 1 (Annotated Blind Review)

**Document**: `D:\Haven\docs\proposals\cdn-cache-analysis\proposal.md`
**Rubric**: proposal.md (1000 pts, target 900)
**Scorer persona**: CTO
**Iteration**: 1 (no previous report)
**Date**: 2026-09-21

---

## Phase 1 — Reasoning Audit (pre-score anchors)

### Argument chain trace

1. **Problem → Solution**: 问题为"运维/优化无法快速回答 CDN 缓存健康度(命中率、失效域名/URI、回源流量浪费、优化建议)"。方案为 CDN Tab 内嵌「缓存分析」卡(手动触发,聚合 + 纯函数规则引擎 + AI 后置),直接产出双命中率、域名/URI 未命中排行、结构化优化项。**对齐良好,无"换题"**。

2. **Solution → Evidence**: 证据(四源 cache_hit 已归一、/aggregate 已支持 cache_hit 维度与 sum_bytes、1453万/1h 实测、tencent/eo.go 映射)支撑"数据面零新增采集"主张,相关性高、非 cherry-pick。

3. **Evidence → Success Criteria**: SC 直接测分析输出(入口、≤2 帧、内容项、可复现、单测≥80%),**基本测在正点上**;但 SC2 的"每域名前几"URI TOP 与 IS1 聚合粒度不符(见深潜)。

4. **Self-contradiction check**: 发现 3 处张力:
   - 场景"并监控漂移"↔ Out of Scope"定时看护(diagwatch)后续立项"(L40 vs L113);
   - "按各域名历史命中率分位数初始化档位边界"↔ Constraint"不新增采集/存储"(L40 vs L57,历史分位数数据来源未定义);
   - SC2"字节命中率(双口径)"↔ Innovation"字节命中率 = hit+partial 字节/总字节"单一公式(L133 vs L31)。

5. **核心逻辑风险(pre-score anchor,非维度覆盖)**: "可缓存两档"是头号创新点,但剔除"动态接口、POST、带认证、Range"所需的请求属性(method/auth/range)不在 IS1 聚合输入维度(cache_hit×host、×uri、×status、sum_bytes)中,规则引擎面对聚合行无法识别 POST/认证/动态接口。"可缓存口径"的**输入数据来源未定义**,核心指标可行性悬空。

### SC Consistency Deep-Dive

区域聚类与逐对双向可满足性检查:

- **集群 A — 聚合通道/后端**(IS1, SC1, SC3, SC4):
  - SC1(≤2 帧)↔ IS1(当前窗并发多维度聚合 + 前窗 1 帧):兼容(帧=窗口级概念),但"当前窗 N 次维度聚合"的 N 无上限,成本承诺弱。
  - SC3(可复现,同一时间边界 + 同一执行批次并发)↔ IS1 并发聚合:一致;文档已自限"可复现断言限定于此"。
  - SC4"明细/聚合既有行为回归通过" ↔ In Scope: **孤儿验收项**,无任何 In Scope 交付回归测试。标记 `ambiguous — requires author clarification`。
- **集群 B — 规则引擎/输出**(IS2, SC2, SC5):
  - SC2"字节命中率(双口径)" ↔ IS2/Innovation 单一字节命中率公式(L31): **方向不一致**。第二口径(全请求/可缓存分档?partial 口径?)全文未定义。标记 `ambiguous — requires author clarification`。
  - SC2"未命中 URI TOP(查询串归一后,每域名前几)" ↔ IS1"cache_hit×host(域名)、×uri(...)"为**分离的单维度聚合,无 host×uri 交叉**:产出"每域名前几"需二维聚合粒度,IS1 未承诺。标记 `ambiguous — requires author clarification`。
  - IS2 健康档位(优/中/差)↔ SC2:一致;SC5 单测项 ↔ IS2:一致。
- **集群 C — 缺失源降级/前端**(IS4, SC4, IS3):一致。EdgeOne 已映射 → "缺失源"实为未来新源,SC4 与 IS4 措辞互洽。
- **跨区**: 场景(缓存健康)"并监控漂移"↔ Out of Scope"定时看护…后续立项" + Constraint"不新增采集/存储":
  ```
  CONTRADICTION: 场景「缓存健康」"并监控漂移" ↔ Out of Scope "定时看护(CDN 版 diagwatch)——结构可复用,后续立项" / Constraint "不新增采集/存储"
  | Type: direction-clash / resource-contention
  | Evidence: 满足"监控漂移"需周期性(定时)巡检或长期历史数据沉淀 → 与"定时看护后续立项"、"不新增采集/存储"冲突;满足"不新增存储 + 只对比前一等长窗口"则无法形成历史分位数基线。
  | Resolution required: 声明漂移检测随分析触发、把"初始化档位边界"的数据来源(一次性历史查询 or 新存储)写清,或把漂移监控显式移交 diagwatch 立项。
  ```
  > 注:该短语位于 `<!-- pre-revised: medium -->` 区域,系预修订新增内容 → 修订引入的新张力。

---

## Phase 2 — Rubric Scoring (verification stance)

### 1. Problem Definition (110)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Problem stated clearly | 38/40 | "当前命中率多高(请求/字节)、哪些域名/URI 缓存失效最重、回源流量浪费多少、该怎么做缓存优化"无歧义。 |
| Evidence provided | 34/40 | 数据可用性证据强(四源归一、/aggregate 支持、实测请求量)。但**只给原始请求量、未给当前命中率/未命中流量基线**:L17"线上实测(2026-09-20):CDN aliyun 1h 全窗 1453 万请求,域名 datasink.jlc.com 单域名 286 万次" — 既然 L16 称 /aggregate 已支持 cache_hit 维度,证据本可先算当前命中率以量化"失效面大/浪费",却止步于请求量。 |
| Urgency justified | 24/30 | "命中率每低 1 个点,热点域名的回源带宽/流量费都会放大"(L22)"放大"未量化;但"数据面已齐 + 即日可用 + 手动触发成本可控"给出即时理由。 |

**Subtotal: 96/110**

### 2. Solution Clarity (120)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Approach concrete | 36/40 | 端点名 `POST /logs/cache-analyze`、规则引擎 cdn-cache、聚合维度、前窗 1 帧、SWR、per-source 降级、AI env 门控均具体。扣分:核心剔除判据未定义(见下)。 |
| User-facing behavior | 43/45 | 内嵌卡、手动触发、前窗对比并标注"对比基期/基于 X/Y 源"、双命中率总览、域名排行、未命中 URI、状态码、可复制优化项、AI 区,可复述。 |
| Technical direction clear | 32/35 | 复用 diagnose 编排骨架、llm 包、/aggregate 联邦通道、查询串归一、golden 回放、单一归一映射源,方向明确。扣分:可缓存口径识别需的输入字段未指明。 |

**Subtotal: 111/120**

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
| Scenario coverage | 34/40 | 覆盖:健康档、未命中集中、大/小文件未命中、4xx/5xx、前窗对比、错误场景,面广。扣分:(a)**可缓存剔除判据无操作定义** — L31"剔除动态接口、POST、带认证、4xx/5xx、Range 等天然不可缓存","动态接口/带认证"从聚合行如何识别未定义,且 IS1 聚合维度不含 method/auth/range;(b)状态码命中率主口径 L44"单维度 /aggregate 无法做'状态 × cache_hit'二维交叉时,退化为状态码分布单独展示"——当前能力下恒退化,发布时交付的是已有"状态码分布",头版承诺价值近零。 |
| Non-functional requirements | 32/40 | 成本/性能(≤2 帧+预估扫描)、覆盖率≥80%、可复现、per-source 状态均有。扣分:(a)L50"超限拦截或提示确认"行为二义(拦截 or 确认由谁决);(b)"≤ 2 帧"未约束"当前窗 N 次维度聚合"的 N 次 SLS 扫描,成本口径低估;(c)**无安全/隐私 NFR** — AI 后置会把含 query 串(可能带 PII)的 URI 级聚合外发 LLM,无脱敏表述。 |
| Constraints & dependencies | 26/30 | /aggregate 依赖、四源 cache_hit 语义差异→单一映射源、SLS 计费手动触发,均点名。扣分:场景"历史命中率分位数初始化档位边界"的历史数据来源与"不新增采集/存储"(L57)冲突,约束未自洽。 |

**Subtotal: 92/110**

### 5. Solution Creativity (100)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Novelty over baseline | 35/40 | 全请求/可缓存双档 + 请求/字节双口径 + 结构化可执行优化项(域名/URI 前缀+动作+miss 流量占比),超出云控制台与 Grafana。 |
| Cross-domain inspiration | 30/35 | 复用内部 WAF 诊断卡交互/编排(跨产品借鉴)。 |
| Simplicity of insight | 21/25 | "请求 vs 字节命中率 gap = 回源带宽浪费"直观优雅;但可缓存档引入的分类复杂度(动态/认证识别)削弱"Why didn't I think of that"的简洁性。 |

**Subtotal: 86/100**

### 6. Feasibility (100)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Technical feasibility | 35/40 | 数据面已实测(DCDN hit_info 聚合、eo.go 映射、/aggregate 生产可用),无 showstopper。扣分:状态×cache_hit 二维交叉不可用 → 主口径指标不可达(恒退化);可缓存分类输入缺失。 |
| Resource & timeline | 20/30 | L86"同构复用估 1-1.5 天"。交付面(新端点+≥80% 覆盖纯函数引擎+未定义的分类启发式+前端卡+降级+前窗+回归)显著大于"同构",**1-1.5 天明显乐观**。 |
| Dependency readiness | 27/30 | 四源已映射、llm 包已落地、WAF 骨架已落地、数据流正常。 |

**Subtotal: 82/100**

### 7. Scope Definition (80)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| In-scope concrete | 28/30 | 5 条均交付物级(端点、引擎、卡、降级标注、summary 字段)。 |
| Out-of-scope explicit | 23/25 | diagwatch、一键下发、独立页/历史趋势、UA/边缘粒度、WAF/SLB,具名。 |
| Scope bounded | 23/25 | 固定窗口对、固定维度集、Top N 封顶,可执行。扣分:SC2 每域名 URI TOP 未体现在 IS 聚合粒度(越界需求)。 |

**Subtotal: 74/80**

### 8. Risk Assessment (90)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Risks identified | 27/30 | 5 条(缺源、partial 上偏、扫描成本、口径漂移、bytes_sent 归一批量),有意义。 |
| Likelihood + impact | 27/30 | M/M、M/M、M/M、M/M、L/M,分布诚实,非全"低风险高影响"。 |
| Mitigations actionable | 26/30 | 标"partial 上偏"、预估扫描+上限、跨源口径测试、降级为仅请求命中率,均可执行。扣分:(a)"待源日志可折算分片命中字节再折算"无排期、被动;(b)**无"建议错误→错误运维动作"风险**(见 blindspot);(c)无回滚/开关计划。 |

**Subtotal: 80/90**

### 9. Success Criteria (80)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Measurable/testable | 27/30 | 入口、≤2 帧、窗口上限、内容项存在性、固定边界可复现、单测≥80%,均可写测试。 |
| Coverage complete | 21/25 | SC 覆盖 IS1/IS2/IS4/IS5;缺口:IS3 前端卡的"AI 解读区/优化项可复制"无 SC 对应;SC4"回归通过"无 In Scope 交付支撑。 |
| SC internal consistency | 19/25 | 见深潜:SC2"双口径"未定义、SC2"每域名前几"超 IS1 粒度、SC4 回归孤儿、SC1"≤2 帧"对 N 无约束。 |

**Subtotal: 67/80**

### 10. Logical Consistency (90)

| Criterion | Score | 依据/扣分 |
|---|---|---|
| Solution addresses problem | 33/35 | 分析卡直接回答四问。 |
| Scope↔Solution↔SC aligned | 22/30 | 漂移监控 vs 定时看护冲突;历史分位数 vs 零存储冲突;每域名 URI vs 单维聚合粒度差。 |
| Requirements↔Solution coherent | 21/25 | 场景"漂移/分位数初始化"为孤儿需求(无 In Scope/SC 承接);可缓存分类输入需求未被聚合设计承接。 |

**Subtotal: 76/90**

### Cross-dimension coherence check

D4 的"可缓存分类输入缺失"与 D2 的"技术方向明确度"、D6 的"技术可行性"同源共振 — 主扣分集中在 D4/D9/D10。

**Total: 96+111+109+92+86+82+74+80+67+76 = 873/1000** (target 900 — 未达标,需修订)

---

## Bias Detection Report

标注区(带 `<!-- pre-revised -->`)= 23 段;未标注区 = 22 段。
- 标注区攻击点:7 段内攻击 = 密度 **0.30**(攻击 1、2、3、4、5、6、9)
- 未标注区攻击点:5 段内攻击 = 密度 **0.23**(攻击 7、8 及 blindspot 10、11、12)
- 比值(标注/未标注)= **1.30**

结论:密度接近平衡,标注区略高(符合预修订集中在高风险段、修订引入的新问题自然聚集在标注区的预期)。**未发现对已修订区域的宽纵偏置**(若对标注区过于宽松,比值应显著 <1)。标注区攻击均针对"修订新引入或未解决"的问题(如"监控漂移"短语、SC2 双口径、可缓存剔除枚举),非重审已修复问题。

---

## Phase 3 — Blindspot Hunt

以下问题不落入任何 rubric 维度,但属 CTO 识别的高风险失败模式:

- **[blindspot] 决策建议无护栏,错误建议可导向错误生产动作**: 规则引擎输出"可执行"优化项,操作员可能照做改 CDN 缓存规则;而核心分类判据("动态接口")未定义,误判 → 对动态接口加缓存 → 用户可见回退。风险表(L121-127)无任何一条覆盖"建议错误→错误运维动作"。quote: L41"优化项:该前缀加/调缓存规则、忽略查询参数"。
- **[blindspot] AI 后置外发数据的隐私边界未声明**: URI query 串可携带 PII(令牌/用户标识),"AI 解读"将 URI 级聚合喂给外部 LLM 无脱敏/范围表述。quote: L27"AI 解读后置(复用已落地的 `llm` 包,env 开启才调,失败降级)"。
- **[blindspot] 无回滚/开关计划**(CTO 人设: Missing rollback plans): 新端点复用共享聚合通道(诊断编排骨架),共享代码回归会波及既有日志查询;仅有"回归通过"测试,无 feature flag/回滚策略。quote: L82"新增 `POST /logs/cache-analyze`(复用 diagnose 编排骨架)"。

---

## Attack List (all with quotes)

1. **[9-SC 内部一致性]**: SC2"字节命中率(双口径)"但全文仅定义单一字节命中率公式 — "请求命中率(全请求/可缓存两档)+ 字节命中率(双口径)" vs "字节命中率 = hit+partial 字节 / 总字节,partial 按全命中计入" — 需定义第二口径或删"双口径"。(标注区;同缺陷亦见于未标注区 L27)
2. **[4-Requirements]**: 可缓存口径剔除判据无操作定义,且聚合输入不含所需字段 — "剔除动态接口、POST、带认证、4xx/5xx、Range 等天然不可缓存" — 需明确判定字段/启发式或扩展聚合维度(method/auth)。(标注区)
3. **[4-Requirements/2]**: 状态码命中率主口径当前不可实现,发布时恒退化为已有"状态码分布" — "单维度 /aggregate 无法做'状态 × cache_hit'二维交叉时,退化为状态码分布单独展示" — 需提升聚合能力或调整对外文案。(标注区)
4. **[4-NFR]**: "≤ 2 帧"成本承诺未约束"当前窗 N 次维度聚合"的 N 次 SLS 扫描;超限行为二义 — "一次分析 ≤ 2 帧窗口聚合(当前窗 N 次维度聚合 + 前窗 1 帧)…超限拦截或提示确认" — 需明确 N 上限与拦截/确认决策。(标注区)
5. **[10-Logical Consistency / 9-SC]** (`conflict-with-pre-revision`): 场景新增"并监控漂移"与 Out of Scope"定时看护后续立项"、Constraint"不新增采集/存储"冲突;历史分位数数据来源未定义 — "上线初期按各域名历史命中率分位数初始化档位边界,并监控漂移" vs "定时看护(CDN 版 diagwatch)——结构可复用,后续立项" / "不新增采集/存储" — 需声明漂移检测归属与分位数数据来源。(标注区,短语系预修订新增)
6. **[7-Scope↔SC]**: SC2"每域名前几"要求 host×uri 交叉粒度,IS1 仅单维度 ×uri 聚合 — "未命中 URI TOP(查询串归一后,每域名前几)" vs "cache_hit×host(域名)、×uri(查询串归一后的 URI 未命中)" — 需补 host×uri 二维聚合或改写 SC2。(标注区)
7. **[1-Problem/Evidence]**: 证据止步于请求量,未量化当前命中率/回源浪费;且 /aggregate 已支持 cache_hit,基线本可给出 — "线上实测(2026-09-20):CDN aliyun 1h 全窗 1453 万请求,域名 datasink.jlc.com 单域名 286 万次" — 需补当前命中率/未命中流量基线支撑"失效面大"。(未标注区)
8. **[6-Feasibility]**: 1-1.5 天工期与真实交付面不符 — "同构复用估 1-1.5 天" — 需按新增面(分类启发式、字节口径、降级、回归)复核工期。(未标注区)
9. **[9-SC coverage]**: SC4"回归通过"为孤儿验收项,In Scope 无对应交付 — "明细/聚合既有行为回归通过" — 需在 In Scope 明确回归测试交付。(标注区)
10. **[blindspot]**: 错误优化建议无护栏,操作员照做可造成用户可见回退;风险表无此项 — "优化项:该前缀加/调缓存规则、忽略查询参数" — 需加建议可信度分级/改动前验证。(未标注区)
11. **[blindspot]**: AI 后置外发含 query 串(可能带 PII)的 URI 级聚合至 LLM,无隐私/脱敏声明 — "AI 解读后置(复用已落地的 `llm` 包,env 开启才调,失败降级)" — 需声明外发数据范围与脱敏。(未标注区)
12. **[blindspot]**: 无回滚/开关计划;新端点复用共享聚合通道,共享代码回归波及既有查询 — "新增 `POST /logs/cache-analyze`(复用 diagnose 编排骨架)" — 需补 feature flag/回滚/灰度计划。(未标注区)

---

## Outcome

**Final Score: 873/1000** (target 900) — 未达标。优先修订:攻击 2(可缓存分类输入)、5(漂移/分位数来源)、6(每域名 URI 粒度)、1(双口径定义)。
