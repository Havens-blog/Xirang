# Eval Report — iteration 3 (final)

- Proposal: `D:\Haven\docs\proposals\dashboard-cert-granularity\proposal.md`
- Rubric: `proposal.md` (1000 pts, target 900)
- Scorer: adversarial CTO persona. Verification stance: unverified assertions are gaps. Only the two input files were read.

---

## Phase 1 — Reasoning Audit

Trace: Problem (粒度矛盾：卡=证书粒度、主表=域名粒度 → 旧证被同名新证掩盖，只能跳台账) → Solution (主表重铸证书粒度 + 卡点击页内筛选 + 行级探测徽标 + 孤儿服务端同构隐藏) → Evidence (file:line、commit、用户反馈、94证/93域) → SC (9 条，均挂单测/组件测试)。Trace 整体闭合。

### Self-contradiction / SC Consistency Deep-Dive

Clusters by affected area: API 层（items/includeHidden/countsByLevel）、默认视图隐藏（总规则/孤儿）、卡筛选（分级卡/特殊卡）、云筛选、行渲染（key/排序/徽标）、开关状态机。

**CONTRADICTION-1（总规则两条子句互斥，且与风险表冲突）**

- 子句 A：“任一筛选激活（分级卡或特殊卡）即豁免孤儿隐藏”
- 子句 B：“孤儿仅在无筛选默认视图隐藏”（字面：任何筛选——含云筛选——激活即豁免）
- 风险表：“云筛选选项若含被隐藏孤儿的云则选中即空态”——该空态仅在**云筛选不豁免**时可能发生；若子句 B 按字面成立，选中含孤儿云的筛选会立即豁免隐藏、出现行，空态不可能发生。
- 两种读法必有一假：要么括号穷尽豁免范围（则 B 对云筛选为假），要么 B 字面成立（则风险场景不可能 + mitigation “云选项来源改为「当前可见行」派生”掩盖了一个不存在的风险）。SC3 “默认视图（无筛选）……总规则唯一生效点” 同样未裁决云筛选归属。**ambiguous — requires author clarification**（倾向读法：仅分级/特殊卡豁免，但必须明示）。

**AMBIGUOUS-1：`countsByLevel[tier].total` 字段未定义。** SC2 断言 “rows.length === countsByLevel[tier].total”，而 API 契约只声明 “countsByLevel 升级为每档 {visible, hidden}”——`total` 键从未在契约中出现（应 = visible+hidden，但契约未写）。

**AMBIGUOUS-2：豁免的服务端/客户端接线缺失。** “孤儿隐藏由**服务端**执行” + “前端只消费 visible 行 + hiddenCount，不派生三态” + “includeHidden 查询参数：默认 false=服务端隐藏生效” + 分级卡点击为**页内筛选**（“复用现有 filter 状态源”）。默认 fetch（includeHidden=false）不含孤儿 → 客户端页内筛选永远筛不出孤儿 → 行数 ≠ 该档 total，除非 开关强制开启→includeHidden=true→重新拉取。开关→includeHidden 的映射与“切换是否重新请求”全文未写。

**AMBIGUOUS-3：云选项“当前可见行”基准。** 开关被强制开启时，“云筛选选项来源改为「当前可见行」派生”的“可见”是默认视图可见（不含孤儿）还是开关开启后可见？未裁决。

其余配对（SC2↔分级卡 In Scope、SC3↔孤儿 In Scope、SC5↔特殊卡 In Scope、SC6↔测试 In Scope、SC7↔排序 NFR）双向可满足，未发现硬矛盾。`conflict-with-pre-revision`：未发现——所有 `pre-revised` 区域与上下文自洽。

---

## Phase 2 — Rubric Scoring

### 1. Problem Definition — 104/110

- 清晰 40/40：粒度矛盾表述无歧义，两读者不会读偏。
- Evidence 36/40：file:line、commit 自陈注释、用户实测反馈、94证/93域 实测均扎实。扣 4：“用户点击卡片的第一反应是**就地看到该档具体证书**（依据见 Assumptions Challenged 表）”——该“依据”是 5 Whys 推演而非用户数据；把推演结论当作用户意图陈述，属 overstated evidence。
- Urgency 28/30：只读角色默认页 + 口径分裂窗口成立；扣 2：“用户信任损耗越大” 无量化或先例支撑。

### 2. Solution Clarity — 111/120

- Approach concrete 36/40：主链路可复述；扣 4 = AMBIGUOUS-2（豁免机制在服务端隐藏与页内筛选之间的接线未写，读者无法复述开关如何驱动 includeHidden）。
- User-facing behavior 42/45：快乐路径、两态文案「可见 N · 隐藏 M」、副标签「生效归属另一证书」均可见可测；扣 3：开关切换是否伴随重新加载（用户可感知的 loading/闪烁）未描述。
- Technical direction 33/35：三处后端+四处前端点名；扣 2：“>5000 行即启用分页/虚拟化” 的实现落点未出现在 In Scope 的 DashboardTable 条目。

### 3. Industry Benchmarking — 100/120

- Industry solutions referenced 28/40：仅 Innovation Highlights 里两句“cert-manager 逐证呈现到期，AWS ACM 控制台同理”，无版本/链接/出处；“Industry Solutions” 正文一段 (“证书到期仪表盘主流形态：证书行主表 + 分级计数卡 + 到期筛选”) 通篇无名有姓的参照物；“业界证书列表无此合并” 为未验证的全称断言。
- ≥3 meaningful alternatives 28/30：do-nothing/方案B/方案A/方案A变体，四个且各有真实 cons，无 straw-man；扣 2：方案 B 标注来源“画布备选”，属于自家备选而非外部方案池。
- Honest trade-offs 22/25：cons 如实（“后端+前端中等改造”）；扣 3：方案 A 变体（全量显示）只给了口径分裂一个 con，未评估“台账全显、看板同构全显”这一第三选项（改台账口径）是否更省。
- Justified vs benchmarks 22/25：拒绝理由链完整；扣 3：未对比 cert-manager/AWS ACM 如何处理“孤儿/无引用证”呈现，恰是本方案差异化所在。

### 4. Requirements Completeness — 102/110

- Scenarios 38/40：快乐+3边界+错误路径；扣 2：错误路径仅覆盖“探测数据缺失”，API 失败/加载中/部分快照过期未列。
- NFR 36/40：性能预算、a11y、兼容、排序齐备且“预算预决”；扣 4：分页/虚拟化触发条件写了（>5000），但交付形态（本期实现 vs 上线后追加）在 NFR/风险/In Scope 三处口径不一。
- Constraints & dependencies 28/30：三态判定单点、import cycle、复用清单齐备；扣 2：`probe_service.go` 域级归属基准仅点名文件，未说明版本/符号锚点。

### 5. Solution Creativity — 86/100

- Novelty 34/40：差异化 articulated（“探测态与到期态并置于一张证书行——业界证书列表无此合并”），但该“业界无”断言本身未经基准核查。
- Cross-domain 30/35：CI/监控“最差优先”reducer 迁移到探测聚合，真实跨域。
- Simplicity 22/25：核心洞察（粒度对齐）优雅；开关状态机（强制开启/禁用/恢复原态/持久化）是方案中复杂度最高的部件，其必要性依赖 CONTRADICTION-1 的裁决。

### 6. Feasibility — 91/100

- Technical 38/40：文件级落点+复用清单，“无新依赖”可信；扣 2：抽屉“扩展为证书多 SAN 视图”的改造量被一笔带过（见 blindspot 2）。
- Resource & timeline 25/30：“约 1 个 feature 周期（2~3 任务）” vs SC 列出的 8 组测试门禁（徽标序、状态机、云选项、排序稳定性、去重、Out-of-Scope 不变量……）——测试矩阵明显超出 2~3 任务的常规容量，估算偏乐观且无带宽佐证。
- Dependency readiness 28/30：三个 commit 号可查，判定单点已合入。

### 7. Scope Definition — 52/80

- In-scope concrete 10/30：条目多为可交付物，但含一处未决二选一（vague instance，Deduction Rules -20）：“开关状态按用户持久化于 localStorage/sessionStorage”——两者语义不同（跨 tab/跨会话），且开关与状态机（强制开启、恢复原态）交互，持久化语义是承载细节，不能留“或”；其余条目具体度尚可（基础 30，-20 → 10）。
- Out-of-scope explicitly listed 25/25：五条具名，含“不变量”式排除。
- Scope bounded 17/25：2~3 任务 + 三处后端改动有界；扣 8：NFR “>5000 行即启用分页/虚拟化（预算预决，不待上线后再定）” 若为本期承诺则 In Scope 漏列一项实质交付物，若非本期则“预决”用词误导——范围边界因此不闭合。

### 8. Risk Assessment — 84/90

- Risks identified 30/30：10 条风险，含两条自曝型（row-key 失效、wildcardSkippedCount 重复计数），非 trivial。
- Likelihood + impact 28/30：H/M 组合诚实（聚合歧义 H/M、row-key H/M）；扣 2：“非 1:1 租户 L/M” 的 L 评级在无跨租户数据下给出，mitigation 自己承认“上线前核查各租户比例”——评级先于核查。
- Mitigations actionable 26/30：多数带单测锚点；扣 4：云选项 mitigation 建立在 CONTRADICTION-1 未裁决的语义上；“上线后核对”无 owner/阈值/告警条件。

### 9. Success Criteria — 69/80

- Measurable/testable 28/30：全部挂单测/组件/grep 断言；扣 2 = AMBIGUOUS-1（`.total` 字段契约缺失，断言对象未定义）。
- Coverage 21/25：In Scope 各项基本有 SC 对应；扣 4：tooltip “注明冲突 SAN 明细” 与 a11y “通告 total 基准明确为**可见行数**” 两条交付承诺无 SC 锁定。
- Internal consistency 20/25：SC 集内部整体可满足；扣 5 = CONTRADICTION-1 波及 SC3（“总规则唯一生效点”未裁决云筛选归属，使 SC3 与风险表/SC6 云选项断言之间存在 ambiguous pair）。

### 10. Logical Consistency — 79/90

- Solution addresses problem 35/35：证书粒度化直接消除“同名新证掩盖”根因，卡计数⇔行数闭合，无 gap。
- Scope↔Solution↔SC aligned 24/30：扣 6 = AMBIGUOUS-2（服务端隐藏 vs 页内筛选 vs 默认 includeHidden=false 三者需开关→参数→重拉的接线才能同时成立，该接线在 Scope/Solution/SC 均未落字）+ 分页/虚拟化在 NFR 与风险中有、In Scope 中无。
- Requirements↔Solution coherent 20/25：扣 5 = CONTRADICTION-1/AMBIGUOUS-3（云筛选豁免语义不确定使“云选项派生”需求与总规则需求互斥或冗余，二选一未定）。

**Cross-dimension coherence check**：扣分点互相印证（CONTRADICTION-1 同时压 D9/D10；豁免接线同时压 D2/D10），无互斥扣分。

**Deduction tally**：vague instance（localStorage/sessionStorage）-20 ×1；无 placeholder、无 straw-man。

---

## Phase 3 — Blindspot Hunt

1. **[blindspot] 新字段暴露给最广读者群的安全等价性是断言不是论证** — “只读角色可见字段即 In Scope 所列行内新字段；证书指纹为公开哈希且台账已展示，看板行沿用同一可见性”。`referenceStatus∈{no_refs_scanned, blind_spot}` 暴露的是内部引用巡检的覆盖缺口，看板是“只读角色的默认页与唯一菜单入口”——受众比台账更宽；“台账已展示指纹”不等于“巡检盲区标记可在默认页全员可见”已过安全审视。必须补一句安全裁决（或明确该字段已在与台账同权限位下展示）。
2. **[blindspot] 抽屉扩展的规模边界未定义** — “探测详情抽屉组件（扩展为证书多 SAN 视图）”与“行点击进入证书多 SAN 探测视图（全部 SAN，最差优先排前）”。通配符/多域名证书可有数十 SAN，抽屉是否分页/折叠/截断未写；这是唯一被复用改造却无任何风险条目覆盖的组件。
3. **[blindspot] 回滚声明依赖未验证的部署原子性** — “破坏性变更以前后端同一 release/PR 协同上线……回滚=同 release 整体 revert，无数据迁移”。“同一 release”不保证前端静态资源（缓存/CDN）与后端同时回退；revert 窗口内旧前端打新 API 或新前端打旧 API 的兼容矩阵未检查，“无数据迁移”≠“无兼容窗口”。

---

## Bias Detection Report

- `pre-revised` 标注区域约 15 处，覆盖 Solution 主句、总规则四段、SC 高门槛五条、破坏性变更/回滚。
- 攻击落点分布：标注区域内 6 处（CONTRADICTION-1、AMBIGUOUS-1/2/3、localStorage vague、回滚 blindspot），非标注区域 6 处（证据越权、基准稀薄、估算乐观、分页口径、安全 blindspot、抽屉 blindspot）。
- 密度比 ≈ 1:1（标注区域面积远小于全文，标注区攻击密度约为非标注区的 2~3 倍）。判读：修订把复杂度（总规则/状态机/双口径）集中引入了标注区域，attacks 聚焦于真实新增复杂度而非标注本身——未发现“用标注诱导注意力转移、掩盖未标注缺陷”的迹象；两处 blindspot 恰在非标注区，佐证非标注区仍被完整扫过。
- `conflict-with-pre-revision`：0。

---

## Verdict

方案根因诊断正确、可测性是同类提案中的上乘水平；但“豁免总规则”在云筛选上的语义未裁决、豁免机制的 fetch 接线未落字、以及一处 -20 vague instance（localStorage/sessionStorage），使总分停在 878（<900 目标）。三处修复成本低（裁决云筛选归属、写明开关→includeHidden→重拉映射、契约补 total 键+选定存储介质），修复后可达标。
