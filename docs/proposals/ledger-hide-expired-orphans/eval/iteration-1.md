# Iteration 1 Report — Rubric Scoring (CTO Persona)

- Document: `D:\Haven\docs\proposals\ledger-hide-expired-orphans\proposal.md`（已修订版）
- Rubric: `proposal.md`（target 900/1000）
- Iteration: 1（无先前 rubric 报告；iteration-0 为 freeform 发现清单，仅用于核对「是否仍存在已登记缺陷」，评分只按当前页面内容）

## Eval-Proposal Complete
**Final Score**: 902/1000 (target: 900)
**Iterations Used**: 1/3

### Score Progression
| Iteration | Score | Delta |
|-----------|-------|-------|
| 1 | 902 | — |

### Dimension Breakdown (final)

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 96 | 110 |
| 2. Solution Clarity | 108 | 120 |
| 3. Industry Benchmarking | 113 | 120 |
| 4. Requirements Completeness | 98 | 110 |
| 5. Solution Creativity | 89 | 100 |
| 6. Feasibility | 84 | 100 |
| 7. Scope Definition | 75 | 80 |
| 8. Risk Assessment | 83 | 90 |
| 9. Success Criteria | 72 | 80 |
| 10. Logical Consistency | 84 | 90 |
| **Total** | **902** | **1000** |

### Outcome
Target reached（902 ≥ 900，余量极小——下列攻击点仍值得下一轮修订吸收）

---

## Phase 1 — Reasoning Audit（pre-score anchors）

链条审计：Problem（过期+零引用=噪音）→ Solution（按三态 no_refs_scanned 服务端过滤 + 可见提示 + expired 豁免）→ Evidence（93/11/1 数字）→ SC（6 条可测断言）。主链成立，方案确实解决的是它声称的问题（降噪不吞风险）。

自相矛盾检查（SC × In Scope 聚类 + 双向可满足性）：

- 聚类 A（后端 `GET /certs` / ledger_service）：InScope-1（过滤+双口径）、SC-1、SC-3、SC-4。SC-1（默认 total 与行一致）与 SC-4（expired 豁免显示全部）通过「exclude 不改变 daysLeft=expired 语义」兼容，双向可满足。无矛盾。
- 聚类 B（前端 ledger）：InScope-2/3、SC-2、SC-5、SC-6。切换态×筛选态状态机自洽（挂起/恢复/重置第 1 页均已定义）。无矛盾。
- 发现 1 处 **ambiguous**：SC-4 的「daysLeft=expired 筛选行数 = 看板『已过期』卡计数」在叠加 search/hostingStatus 时必然不等（卡计数是全局口径），SC 未声明断言作用域（裸跳转）——非矛盾，标记 `ambiguous — requires author clarification`（计入 D9）。
- 发现 1 处载荷性前提缺口（Reasoning audit 独立标记）：Evidence 用 refCount=0 计数噪音（1 张），而判定谓词是 referenceStatus=no_refs_scanned；文档自己警告「refCount=0 可能同时来自 no_refs_scanned 与 blind_spot」，却从未断言 `*.jlc-robot.com` 的三态实为 no_refs_scanned。若该证实为 blind_spot，当前可隐藏数为 0，Urgency 前提坍塌。

代码库核验（verification stance）：
- 三态语义存在：`e-cam-service/internal/cert/domain/reference_status.go` — 且明文「ReferenceStatus 引用三态**派生量（DTO 派生，非存储字段）**」。
- `ledger_service.go ListCerts` 当前行级数据只有 `counts[c.Fingerprint]`（「refCount 取最新成功快照计数（无快照=0）」），**未派生三态**；区分 no_refs_scanned 与 blind_spot 需要快照 CoverageMeta（`reference_query_service.go deriveRefStatusFor`）。
- `7cdd488`、`listCertsApi`、`e-cam-web/src/views/cert/ledger/`（index.vue / components/CertTable.vue / format.ts）均核验存在。
- 结论：方案技术方向可行，但「服务端过滤派生字段 × SQL offset/limit 分页」这一最难的实现落点文档未触及（见 D2/D6 扣分）。

---

## Phase 2 — Rubric 评分明细（含扣分 quote）

### 1. Problem Definition — 96/110
- 问题清晰 38/40：谓词式问题定义，无歧义。
- 证据 32/40：数字具体（93/11/1、指纹、refCount 21/11），但核心证据与判定谓词错位——quote: 「其中 **1 张 refCount=0**（`*.jlc-robot.com`，-11 天，fingerprint_only，非保护期，可删除）」对照其自身警告「裸 refCount 数字无法区分『明确零引用』与『扫描未覆盖』」——文档从未给出该证的三态为 no_refs_scanned 的证据。
- 紧迫性 26/30：「近 2 天手动换证/导入活动频繁」无任何量化（几次换证？导入多少张？），是未验证的活动性断言。

### 2. Solution Clarity — 108/120
- 方案具体 38/40：谓词、双口径、状态机、豁免落点全部定死。
- 用户行为 44/45：Key Scenarios 覆盖默认/切换/豁免/组合/空态，文案三处一致（先前漂移已修）。
- 技术方向 26/35：quote: 「（服务端过滤，保证分页/total 一致）」——referenceStatus 是「DTO 派生，非存储字段」（reference_status.go 原文），在 offset/limit 分页查询上按派生字段过滤需要明确落点（扩展 CertListFilter 子查询 vs 全量取回内存过滤分页），文档对此零字提及；这是本方案唯一有实质实现分歧的决策点。

### 3. Industry Benchmarking — 113/120
- 行业方案引用 37/40：cert-manager、Venafi/Keyfactor 具名；但「商业证书管理产品……普遍采用『归档/折叠出默认列表』」是无出处的类目级断言。
- ≥3 替代方案 28/30：4 个方案含 do nothing，均非稻草人（各有真实 pros）。
- 诚实取舍 24/25：Accepted 方案也列了代价（「需后端过滤参数 + 提示切换」），Key Scenarios 还登记了「清理入口变深」取舍。
- 对标论证 24/25：对齐「降噪不吞风险」共识，论证清楚。

### 4. Requirements Completeness — 98/110
- 场景覆盖 38/40：含空态、正交性、blind_spot、状态机，覆盖完整。
- NFR 34/40：口径一致/可发现性/性能三条到位，但性能声明低估实现成本——quote: 「性能：复用 listCertsApi 主查询，仅新增一次同条件统计口径（或单查询双 total 返回 `visibleTotal + hiddenCount`），不新增独立接口」——当前 ListCerts 行级仅有 refCount（「无快照=0」），派生三态需引入快照 CoverageMeta，这不是「一次统计口径」能覆盖的新增派生成本。
- 约束依赖 26/30：文件、语义、7cdd488 均已具名；漏掉「referenceStatus 非存储字段」这一直接影响过滤落点的结构性约束。

### 5. Solution Creativity — 89/100
- 相对基线的新颖性 36/40：三态保守显示 vs 行业「默认隐藏已过期」，差异化表述清晰。
- 跨域借鉴 30/35：cert-manager「是否被 workload 引用」判定思路的移植明确。
- 洞察简洁性 23/25：单谓词 + 提示开关，无过度工程。

### 6. Feasibility — 84/100
- 技术可行 30/40：可行但最难关卡未讨论——派生字段过滤 × 分页（同 D2）；且 Evidence 中「无快照=0」意味着 refCount=0 实际覆盖三种来源（no_refs_scanned / 无快照 blind_spot / 覆盖外 blind_spot），三态判定在列表路径的实现量被低估。
- 资源/时间 26/30：范围小、测试明确，现实；未给时间框（提案类文档可接受）。
- 依赖就绪 28/30：reference_status.go / listCertsApi / 7cdd488 核验存在；但「当前 1 张可隐藏」的数据前提未经三态核验。

### 7. Scope Definition — 75/80
- In-scope 具体 28/30：四条全是可交付物。
- Out-of-scope 明确 24/25：四条具名（删除/告警/归档/看板）。
- 有界 23/25：可一次执行；但 Risk 4 缓解引入的配置开关不在 In Scope（见 D8），范围边界与风险缓解存在 1 处脱节。

### 8. Risk Assessment — 83/90
- 风险识别 30/30：7 条均为实质风险（#3 blind_spot 误藏、#4 状态漂移漏看尤其好）。
- 评级诚实 28/30：非全「低/高」，#3/#6 给了「中」，诚实。
- 缓解可执行 25/30：quote: 「过滤行为加配置开关可一键关闭隐藏（回归显示全部），后续可按 blind_spot 占比等阈值自动降级」——该 kill switch 既不在 In Scope 也不在 SC，是缓解措施承诺了范围外交付物（隐藏成本模式）；实现者按 In Scope 执行时该缓解落空。

### 9. Success Criteria — 72/80
- 可测 28/30：六条均可写成测试/检查单。
- 覆盖完整 23/25：覆盖全部 In Scope 条目。
- SC 内部一致性 21/25：集合整体可满足；SC-4 存在歧义——quote: 「daysLeft=expired 筛选行数 = 看板『已过期』卡计数（动态相等断言，不写死数值）」未声明断言作用域，叠加 search/hostingStatus 时必然失败（卡计数为全局口径），`ambiguous — requires author clarification`。

### 10. Logical Consistency — 84/90
- 方案解决问题 33/35：主链成立；唯 Evidence→谓词错位（refCount=0 ≠ no_refs_scanned）使「当前 1 张噪音」的论证有 1 步未接通。
- Scope↔Solution↔SC 对齐 27/30：kill switch 缓解不在 scope（已在 D8 扣），其余对齐。
- 需求↔方案 24/25：无孤儿需求、无无需求的功能。

**跨维度一致性检查**：Solution/In Scope/SC/Key Scenarios 四处文案与谓词一致（先前文案漂移已收敛）；「切换重置第 1 页」在 Solution、Key Scenarios、In Scope 三处一致。唯一跨节脱节为 Risk-4 缓解 vs Scope（D8 已计）。

---

## Phase 3 — Blindspot Hunt

1. **[blindspot] 谓词无「过期新近度」宽限窗口**：quote: 「台账默认列表**不显示『已过期且 referenceStatus=no_refs_scanned（已成功扫描且明确零引用）』的证书**」——一张昨天刚过期的零引用证书立即从默认视图消失，而「刚过期」恰是漂移确认价值最高的观察窗口（行业做法如 cert-manager 用 renewal/过期时间阈值分档）。文档未讨论是否需要 daysSinceExpiry 阈值（如仅隐藏过期 >7 天者），也未把「新过期即隐藏」登记为有意取舍。改进：补时间窗决策或登记取舍。

2. **[blindspot] 载荷前提未核验，可能使问题为空**：quote: 「当前噪音仅 1 张，但换证操作正在持续进行」——若 `*.jlc-robot.com` 的三态实为 blind_spot（扫描未覆盖其云/产品，恰与 Evidence 只给出 refCount=0、fingerprint_only 相容），则今日可隐藏数为 0，Urgency 整段建立在未核验假设上。Reasoning audit flagged this independently of dimension scoring. 改进：在 Evidence 中给出该证 referenceStatus=no_refs_scanned 的系统内实测值（或改述为「至多 1 张，以三态实测为准」）。

3. **[blindspot] 切换态的持久化未定义**：quote: 「清除筛选后回到切换态」——状态机定义了筛选联动却未定义刷新/再次进入页面后切换态是否保留（sessionStorage/URL 参数/每次重置为隐藏）。若刷新即重置为隐藏，用户展开态在一次刷新后丢失，与「不得静默吞行」精神存在张力。改进：明确切换态生命周期（至少声明「会话内不持久，属有意取舍」）。

---

## 偏差检测报告（annotated-blind-review）

- 标注区域（11 个标注块：Problem/Evidence/Solution/Innovation/Key Scenarios/NFR/Constraints/Industry/In Scope/Risks/SC）：7 攻击点 / 11 段落 = 密度 0.64
- 未标注区域（Urgency/Comparison Table/Out of Scope）：1 攻击点 / 3 段落 = 密度 0.33
- 比率（标注/未标注）：1.9 — 预修订标记成功吸引了注意力分配，且标注区域攻击点均为残留/新增问题而非已纠正问题的重复评估；未发现与预修订方向冲突的判断（无 conflict-with-pre-revision 项）。

---

## 攻击点汇总（供下一轮修订）

1. [Problem Definition/Evidence]: 噪音计数用 refCount=0 佐证，谓词却是三态——quote: 「其中 **1 张 refCount=0**（`*.jlc-robot.com`…）」 + 「裸 refCount 数字无法区分…」 — 必须给出该证 referenceStatus=no_refs_scanned 的实测依据，或按三态重新计数。
2. [Solution Clarity/Feasibility]: 派生字段（非存储）在分页查询上的服务端过滤落点未定义 — quote: 「（服务端过滤，保证分页/total 一致）」 — 必须写明实现落点（扩展 CertListFilter 子查询 vs 内存过滤后分页）。
3. [Requirements/NFR]: 性能声明低估三态派生成本 — quote: 「仅新增一次同条件统计口径（或单查询双 total…）」 — 必须承认列表路径需引入快照 CoverageMeta 派生三态。
4. [Risk Assessment]: kill switch 缓解承诺了范围外交付物 — quote: 「过滤行为加配置开关可一键关闭隐藏」 — 必须纳入 In Scope + SC，或显式标注为后续项并降级该缓解表述。
5. [Success Criteria]: SC-4 相等断言作用域歧义 — quote: 「daysLeft=expired 筛选行数 = 看板『已过期』卡计数」 — 必须声明为裸跳转（无其他过滤）下的断言，或定义按上下文的卡计数口径。
6. [Problem Definition/Urgency]: 活动性断言无量化 — quote: 「近 2 天手动换证/导入活动频繁」 — 给出数据或弱化表述。
7. [blindspot]: 谓词无过期新近度宽限窗 — quote: 「不显示『已过期且 referenceStatus=no_refs_scanned…』的证书」 — 补 daysSinceExpiry 阈值决策或登记取舍。
8. [blindspot]: 「当前噪音仅 1 张」前提未按三态核验，可能为 0 — quote: 「当前噪音仅 1 张」 — 以系统实测三态确认。
9. [blindspot]: 切换态刷新后是否保留未定义 — quote: 「清除筛选后回到切换态」 — 明确切换态生命周期。
