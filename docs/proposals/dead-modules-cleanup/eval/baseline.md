# Proposal Eval — Baseline (Iteration 1)

- **Document**: `D:\Haven\docs\proposals\dead-modules-cleanup\proposal.md`
- **Scorer persona**: CTO（惯于批准/否决提案的技术负责人）
- **Rubric**: `proposal.md`（scale 1000，target 900）
- **Date**: 2026-09-09
- **SCORE: 684/1000**

---

## 0. 独立代码核验记录（不采信文档自述，逐条实证）

评分前对文档的承重断言做了独立对账（非转述既有评审）：

| # | 文档断言 | 核验结果 |
|---|---------|---------|
| V1 | `internal/{task,sync,servicetree}` 外部 importer 0 | **成立**。import 语句级全仓 grep，命中均为 0（排除各自目录与 `internal/cam/*`） |
| V2 | `internal/account/module.go`、`internal/asset/module.go` 0 importer | **成立**。对包根 `internal/account"` / `internal/asset"` 的 import 全仓 0 命中 |
| V3 | `ioc/`、`Makefile`、`main.go` 对三死模块零引用 | **成立**。grep 0 命中；`.github/workflows/{ci,deploy}.yml` 及全树亦无 tmp 工具引用，"非部署单元"假设成立 |
| V4 | `internal/account/doc.go` 现文为已被放弃的迁移计划 | **成立**（"当前状态：待迁移…迁移计划：1. 将 cam/domain/account.go 迁移到 account/domain/"） |
| V5 | 决策文档 0001 存在且 Phase 4 可补记 | **成立，但有坐标问题**。文件实际位于**工作区根** `D:/Haven/docs/decisions/`，`e-cam-service/docs/` 下**无** decisions/ 目录；Phase 4 现文"删除 `e-cam-service/internal/cmdb`、其路由注册与 collection 直写代码"与提案第 4 条联动前提一致 |
| V6 | `cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred` 存在且为临时诊断工具 | **成立**，但**同族第三只 `cmd/tmp-k8s-diag/` 存在而全文未提**（Evidence、In Scope、Out of Scope 三处皆无） |
| V7 | Risk 4 缓解"需要时可从 git 历史找回" | **被代码证伪**。`git status --short cmd/` → `?? cmd/tmp-cas-diag/`、`?? cmd/tmp-k8s-cred/`、`?? cmd/tmp-k8s-diag/`——三只工具均为未跟踪文件，从未进过 git 历史（`.claude/backups/40-backup-*.md` 显示它们是会话内现写的）。一旦删除，诊断逻辑永久丢失 |
| V8 | 基线"当前全部通过" | **被代码证伪**。实测 `go test ./internal/cert/domain/ -run TestBsonTagsMatchSchema` → **FAIL**（`TestBsonTagsMatchSchema/K8sCredential`：bson 字段 `displayName` 与 schema.sql 不一致）。基线今天就是红的，与死模块无关 |
| V9 | account 新 doc.go 文案"跨域账号只读访问层（cert/logquery/mcp/cam-cost 共享）" | **被代码证伪，两处**。(a) 非只读：`internal/mcp/tools_sync.go:85` 调用 `SyncAccount`，而 `internal/account/service/account.go:338` 有 `s.taskQueue.Submit(t)`——状态变更路径；(b) 消费方漏 `internal/cam/task`（`internal/cam/task/module.go:6`、`executor/sync_billing.go:11` 均消费 `accountservice.CloudAccountService`）。另 `internal/mcp/deps.go:59` `NewCloudAccountService(accountRepo, nil, logger)` 注入 **nil 队列**——该"只读"文档将紧挨着一条一调用即 panic 的活路径 |
| V10 | Evidence 对注释级引用的盘点（仅列 cam 侧 2 处） | **不完整**。实测 live 代码中 `internal/task` 注释级引用共 **10 处**：cam 2 处之外，cert 侧 8 处（`internal/cert/service/execute_service.go:25,26,114,962,968`、`internal/cert/module.go:231`、`internal/cert/scheduler/inspection_job.go:3`、`internal/cert/service/execute_service_test.go:1208`），且含指令性表述"参考 internal/task executor 既有用法"。删除后这些注释成为指向不存在路径的悬空指针 |

**核验小结**：判死证据链（V1/V2/V3/V4）成立，提案不是想象出来的；但**安全叙事层有 4 处被代码直接证伪**（V7/V8/V9/V10）+ 1 处盘点缺口（V6）+ 1 处坐标错位（V5）。

---

## 1. Phase 1 — Reasoning Audit（pre-score anchors）

论证链追踪：

1. **Problem → Solution**：问题="悬空模块抬高理解成本与误改风险 + 文档说谎"。解法=删死模块 + 文档正名。方向正确，但**解法只兑现了问题的一半**：① "让文档说真话"这一半，其交付物文本本身与代码事实不符（V9）；② "同名包最容易改错文件"这一半，cert 侧 8 处现役注释仍把活跃框架叫作 "internal/task 框架"（V10），删除后误导性不降反升（指针悬空）。
2. **Solution → Evidence**：判死证据扎实（import 级实测）。但 Evidence 被当作**清理边界**使用——"import 为 0"推不出"注释可以不管"。判死标准与清理边界被混为一谈。
3. **Evidence → Success Criteria**：SC-2 只测 import 语句级 grep，对注释残留零感知；SC-5 要求"全部通过 + 与基线一致"，而基线已红（V8）——两个子句在当前状态下恰好互斥（"全部通过"今天必失败，"与基线一致"今天必通过），执行时必然产生"清理引入了问题"的错误归因。
4. **Self-contradiction check（SC 聚类 + 双向可满足性）**：
   - SC-1↔SC-2：SC-2 是 SC-1 的可验证表达，`internal/cam/*` 豁免与 Out of Scope 一致。无矛盾。
   - SC-1↔SC-5：零外部 importer + ioc/Makefile 零引用（V1/V3），可同时满足。无矛盾。
   - SC-3↔InScope-3/4：一致（不要求收编迁移）。
   - **SC-3 ↔ Proposed Solution 第 2 条**：SC-3 要求 "`doc.go` 描述与消费方事实一致"，而 Solution 规定的文本（"只读访问层…cert/logquery/mcp/cam-cost 共享"）按 V9 不满足该判据——照抄 Solution 文本必然挂 SC-3。这是**跨节矛盾**（记入 D10），非 SC 集内矛盾。
   - **SC-5 内部**："全部通过" 与 "与清理前基线一致" 在红色基线下是两个不同判据，文档未择一或定义等价关系（记入 D9）。
   - 结论：SC 集内部无 mutual-exclusion；矛盾集中在 SC 与 Solution/现实之间。

Pre-score anchors（待 Phase 3 判定是否为 blindspot）：
- A1：文档自设 "SC Consistency Check" 节并自我判定"无矛盾"，但漏掉了唯一真实的矛盾（SC-3 ↔ Solution 文本）。**自我盖章 ≠ 验证**。
- A2：全文档的验证叙事（基线、等价、git 找回）默认"仓库干净、目标已入库"，而实际是 79 项未提交变更 + 删除目标为未跟踪文件的脏工作区。变更不可审计、删除不可回滚。
- A3：Urgency 引用"CMDB Phase 4 删除面重叠、返工成本更高"作为紧迫性，同时方案把 account/asset 的归属决断**主动推迟**到同一个 Phase 4——紧迫性逻辑对方案自己保留的部分同样成立。属可辩护的分期取舍，但文档未解释为何同一逻辑不一以贯之。

---

## 2. Phase 2 — Rubric Scoring

### D1. Problem Definition — 94/110

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 38/40 | "迁移半成品"问题定义清晰无歧义，且正确收窄为"一次清理 + 文档正名"。唯死模块与半活模块两类对象混在一个问题陈述里，边界靠 Evidence 补足。 |
| Evidence provided | 28/40 | 证据是本提案最强项：import 语句级实测、日期、file:line 引用（V1/V2/V3/V4 全部核验成立）。但两处**盘点不完整**："cam 侧两处仅为注释提及"把注释足迹呈现为 2 处，实测 10 处（V10）；"`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`：临时诊断工具"漏掉同族的 `cmd/tmp-k8s-diag`（V6）。证据链在承重处成立、在边界处失真。 |
| Urgency justified | 28/30 | 紧迫性具体可感：三对同名包并存、CMDB Phase 4 删除面重叠。成本 of delay 明确。 |

### D2. Solution Clarity — 111/120

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 36/40 | 五项动作全部是精确到文件/目录的交付物，可复述、可执行。扣分：第 4 条决策文档联动引用 `docs/decisions/0001-...md`，但该文件在工作区根而非仓库内（V5）——四项动作之一按文档执行会找不到目标。 |
| User-facing behavior described | 42/45 | "用户体验：无（纯内部清理）；体验改善发生在开发者侧"——对清理类提案，明确声明"无"并指明受益面，是诚实且正确的写法。 |
| Technical direction clear | 33/35 | 精确路径删除防同名误删、build/vet/test 三闸门、doc.go 指向决策 0001——技术方向足够。注：Solution 规定的 doc.go **文本**本身失实（V9），这是正确性问题非清晰性问题，在 D10 记账。 |

### D3. Industry Benchmarking — 84/120

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Industry solutions referenced | 16/40 | 全节只有"常规做法即'删除 + 编译验证'"与"先收编再删"两句泛述。**零具名引用**：Go 生态有现成工具与 published patterns（`golang.org/x/tools/cmd/deadcode` 的可达性分析、staticcheck U1000、ADR/MADR 决策记录实践）一个都没提。"仅 import 语句判死"是自造方法，未对照任何标准工具验证。 |
| At least 3 meaningful alternatives | 27/30 | 三方案（保守清理/彻底收编/do nothing）真实互异、非稻草人，均给出量化代价（"约 25 个 import 站点改动"）。扣分：备选空间恰好绕开了真正的难点（注释残留处置、tmp 工具入库归档），未提供"删除 + 注释正名"这类更贴近问题的中间选项。 |
| Honest trade-off comparison | 20/25 | 选中方案的 con（"account/asset 仍位于独立顶层目录"）如实列出，未粉饰。扣分：选中方案的 cons **恰好漏掉它自己制造的代价**——"10 处悬空注释留存"与"新 doc.go 与事实不符"，这是对所选方案的 cherry-picking。 |
| Chosen approach justified against benchmarks | 21/25 | "收益不成比例" + CMDB 依赖 + 二次返工风险，理由链完整成立。扣分：判死方法论本身未与行业标准工具对比（谁更可靠、各自漏什么）。 |

### D4. Requirements Completeness — 68/110

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Scenario coverage | 18/40 | 把"安全场景"作为清理类替代用户场景是合理重构，同名包保护是好的边界场景。但**唯一会真实发生的错误场景缺席**：基线已红（V8），`go test ./...` 今天必然失败，文档没有任何处置预案（既有红灯登记 / 失败集合比对）；**数据丢失场景缺席**：删除未跟踪文件不可找回（V7）。只覆盖了金色路径。 |
| Non-functional requirements | 28/40 | 零行为变化、不动 Mongo 数据与索引、构建产物不变——对清理提案是正确的 NFR 集。扣分：**可回滚性/可恢复性 NFR 缺位**（只在 Risk 4 以虚假缓解出现，V7）；文档准确性这一本轮核心 NFR 无对应验收手段。 |
| Constraints & dependencies | 22/30 | 决策 0001 依赖、收编迁移显式推迟、`internal/cam/*` 不动——清楚。扣分：执行环境约束缺席（79 项未提交变更的在途工作区、删除目标未入库），这两条直接决定本次清理能否被审计。 |

### D5. Solution Creativity — 46/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Novelty over industry baseline | 20/40 | 诚实自评"无创新，属常规死代码清理"。唯一方法论点"以'仅 import 语句'为判死标准"确有辨析价值（朴素 grep 会把注释当依赖），与基准的差异化说清了。但它恰是本轮盲区来源：判死标准被误用为清理边界（见 [blindspot] 3）。 |
| Cross-domain inspiration | 10/35 | 无跨域借鉴。"决策文档联动防遗忘"隐含借用了 ADR 实践但未自觉表述；未从任何其他领域（编译器 DCE、特性开关下线流程、数据库 schema 弃用）汲取方法。 |
| Simplicity of insight | 16/25 | 洞察简单直接、不造作。但"import 即终点"的简洁同时把"删除后留下的 10 处活注释"划出了视野——简洁以完整性为代价。 |

### D6. Feasibility — 61/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Technical feasibility | 26/40 | 删除本身零技术难度（V1/V2/V3 实证）。但提案把**验证闸门**建在"全量 `go test ./...` 通过"上，而该闸门今天就是断的（V8）——按文档执行，验收必然失败且原因与清理无关。技术可行性对"删除"成立，对"按写明的验收标准收尾"不成立。 |
| Resource & timeline feasibility | 15/30 | 全文档**无任何工作量估算与时间线**。范围虽小（半天量级），但"scope realistic for the proposed timeline"无从谈起——没有 proposed timeline。 |
| Dependency readiness | 20/30 | 决策 0001 存在且 Phase 4 文本可对接（V5）。扣分：真正的前置依赖——干净的基线 commit 与在途变更的处置——既未识别也未安排；依赖决策 0001 文件的路径坐标还写错了仓库（V5）。 |

### D7. Scope Definition — 62/80

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 28/30 | 五条 In Scope 全是精确交付物，无一条是模糊领域。扣分：第 5 条决策文档路径坐标错位（V5）。 |
| Out-of-scope explicitly listed | 14/25 | Out of Scope 四条写得清楚且理由充分。但存在**"第三态"缺口**：① `cmd/tmp-k8s-diag` 与 In Scope 第 2 条同族同类（同为未跟踪一次性诊断工具），既不在 In 也不在 Out，"本轮清掉 tmp 工具"的边界是任意画出来的（V6）；② 10 处注释级引用既不清理也不声明为 follow-up，实现了"既不声称、也不负责"。 |
| Scope is bounded | 20/25 | 范围有界、可在一个明确时段内完成。扣分：注释残留与 tmp-k8s-diag 的归属未决，使"有界"依赖实现者的自由裁量。 |

### D8. Risk Assessment — 46/90

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 26/30 | 四条风险均非平凡（同名误删、init() 副作用、doc 漂移、工具依赖），Risk 2 的"无 import 即无 init 触发"推理正确。扣分：**两条最可能兑现的风险缺席**——红色基线导致的错误归因风险、删除未跟踪文件的不可逆丢失风险。 |
| Likelihood + impact rated | 8/30 | **完全无评级**。表格只有 Risk | Mitigation 两列，没有任何 likelihood/impact 标注，无法据以排序或设门槛。该 criterion 的基本要求（有评级且诚实）未满足。 |
| Mitigations are actionable | 12/30 | Risk 1/2/3 的缓解可执行（精确路径、构建闸门、链接唯一事实源）。但 **Risk 4 的缓解是假命题**："需要时可从 git 历史找回"——三只 tmp 工具从未入库（V7），该缓解对它要保护的唯一场景恰好无效。一条被证伪的缓解比没有缓解更危险，因为它制造了虚假的安全感。 |

### D9. Success Criteria — 61/80

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable and testable | 21/30 | SC-1/2/4 可机械验证，SC-5 可执行。扣分：SC-3 的核心子句"doc.go 描述与消费方事实一致"**无机械判据**（不列消费方清单就无法逐行核对，"一致"沦为自由心证）；SC-5 如上所述今天不可达成。 |
| Coverage is complete | 21/25 | 五条 SC 与五条 In Scope 一一对应，覆盖完整。扣分：NFR 层（构建产物不变、零行为变化）仅由 SC-5 弱覆盖，而 SC-5 自身失效。 |
| SC internal consistency | 19/25 | SC↔SC、SC↔InScope 双向可满足性逐对核查通过（含 cam/* 豁免与 Out of Scope 一致）。扣分：**SC-5 内部二义**——"全部通过"与"与清理前基线一致"在红色基线（V8）下是两个不同且方向相反的判据，文档未定义二者关系。 |

### D10. Logical Consistency — 51/90

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Solution addresses the stated problem | 20/35 | 删除动作确凿解决问题的一半。但方案自己宣称的收益"模块图与文档**第一次与运行时事实一致**"按当前交付物**兑现不了**：新 doc.go 文本出生即失实（"只读"被 `SyncAccount`→`taskQueue.Submit` 写路径反驳、消费方漏 `internal/cam/task`，V9）；cert 侧 8 处现役注释继续把活跃框架叫 "internal/task 框架"（V10）。方案在治"文档说谎"时交付了新的谎言。 |
| Scope ↔ Solution ↔ Success Criteria aligned | 18/30 | In Scope 与 Solution 逐条镜像、SC 与 In Scope 逐条对应——骨架对齐。三处错位：① Solution 第 2 条规定的 doc 文本与 SC-3 的"与消费方事实一致"**互斥**（照抄文本必挂 SC-3）；② `cmd/tmp-k8s-diag` 三处皆无（V6）；③ 决策文档路径坐标不一致（V5）。 |
| Requirements ↔ Solution coherent | 13/25 | 零行为变化等需求与解法吻合、无孤儿特性。但关键需求"必须与删除前等价（基线：当前全部通过）"的**前提被实测证伪**（V8）——需求以假前提书写且解法未调和；"删除无构建面影响"成立，但唯一需要该需求兜底的错误场景（验收失败）无解法承接。 |

### Cross-Dimension Coherence Check

- Scope / Solution / SC 三表逐条镜像，骨架一致性良好（这在低分提案里少见）。
- ��三处跨节裂缝贯穿全文：**"基线全绿"假设**（D4/D6/D8/D9/D10 五处共用同一假前提）、**"import 判死 = 清理边界"混同**（D1 Evidence / D5 方法论 / D7 Scope / D10 收益主张）、**"文档说真话"目标与失实交付物文本**（D2 → D10）。任一裂缝修复需同步动四处，修订时应整体处理而非逐点补丁。

---

## 3. Phase 3 — Blindspot Hunt

1. **[blindspot] 自我盖章式一致性检查** — 文档自设 "SC Consistency Check" 节并下结论"无 SC↔SC、SC↔InScope 矛盾"，但该检查恰好漏掉唯一真实的矛盾（SC-3"描述与消费方事实一致" ↔ Solution 规定的失实 doc 文本）。一份把"自证清白"写成正文章节的提案，把复核者的注意力引向"已经检查过了"的方向。Reasoning audit flagged this independently of dimension scoring. **Must improve**: 删除自我结论性段落，或将该节改写为"已核查项 + 每项的机械判据与核验输出"，明确自查不能替代独立验证。

2. **[blindspot] 变更不可审计、删除不可回滚的执行环境盲区** — 全文档验证叙事（"基线：当前全部通过"、"需要时可从 git 历史找回"）默认仓库干净、目标已入库；实测 79 项未提交变更 + 删除目标全是未跟踪文件。删除 commit 将混在在途工作之上，"零行为变化"无从 diff 审计；三只工具删除即永久丢失。任何 rubric 维度都只覆盖了它的碎片（D6 依赖、D8 缓解），"本次变更本身能否被事后审计"无维度承接。Reasoning audit flagged this independently of dimension scoring. **Must improve**: 增加"先入库/先提交、再删除"的两步流程与基线 commit 钉扎，或显式声明放弃审计性与理由。

3. **[blindspot] 判死标准与清理边界的范畴混同** — "以'仅 import 语句'为判死标准（排除注释引用），避免把'文档提到'误当'代码依赖'"作为方法论被单列为 Innovation Highlights 的唯一论点，但该标准只回答"什么死了"，不回答"删除后留下什么"。cert 侧现役注释里的指令性表述"参考 internal/task executor 既有用法"正是被该标准排除在外的"文档提到"——而它会在删除后继续给开发者指路。提案把方法论上正确的东西当成了范围上完备的东西。**Must improve**: 显式区分"判死判据"与"清理边界"，为注释级残留（实测 10 处）做出 In/Out 或 follow-up 的命名决定。

---

## 4. Attack List（按修订优先级）

1. [D8/Risk 4] **虚假缓解必须先修**："需要时可从 git 历史找回" — 三只 tmp 工具未跟踪（`?? cmd/tmp-cas-diag/` 等），从未入库。改为"先提交入库存档、再删除"，并把 `cmd/tmp-k8s-diag` 一并纳入处置。
2. [D4/D6/D8/D9/D10] **基线前提被证伪**："（基线：当前全部通过）" — 实测 `internal/cert/domain` 红灯（displayName vs schema.sql）。必须：钉基线 commit、登记既有红灯、把 SC-5 改写为"失败集合与基线逐项一致、无新增失败"。
3. [D10] **交付物文本失实**："重写 doc.go——'跨域账号只读访问层（cert/logquery/mcp/cam-cost 共享）'" — SyncAccount 是写路径且 mcp 注入 nil 队列、cam/task 是漏掉的消费方。doc 文本按 import 清单重写并列全消费方，nil 队列坏路径显式记录。
4. [D1/D7] **盘点缺口**："cam 侧两处仅为注释提及"、"`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`" — 注释级引用实测 10 处（cert 8 处含指令性表述）、tmp 工具实为三只。补全 Evidence 并为每处命名处置方式。
5. [D8] **风险无评级**："Risk | Mitigation" 两列表格 — 补 likelihood/impact 两列并诚实评级。
6. [D6] **无时间线/工作量**："纯删除 + 文档正名，零行为变化" — 补工作量与阶段划分（含前置的基线处置步骤）。
7. [D2/V5] **路径坐标错位**："`docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md`" — 该文件在工作区根而非 e-cam-service 仓库内，改为绝对锚点或迁入仓库。
8. [D3] **零具名行业引用**："常规做法即'删除 + 编译验证'" — 引入 `golang.org/x/tools/cmd/deadcode`、staticcheck U1000、ADR 实践并对比取舍。
9. [D9] **SC-5 内部二义**："全部通过，与清理前基线一致" — 基线红时两子句方向相反，择一并定义等价关系。
10. [blindspot 1] **删除自我盖章段落**："结论：无 SC↔SC、SC↔InScope 矛盾。" — 以机械判据 + 核验输出替代自证结论。

---

## 5. Scores Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 94 | 110 |
| 2. Solution Clarity | 111 | 120 |
| 3. Industry Benchmarking | 84 | 120 |
| 4. Requirements Completeness | 68 | 110 |
| 5. Solution Creativity | 46 | 100 |
| 6. Feasibility | 61 | 100 |
| 7. Scope Definition | 62 | 80 |
| 8. Risk Assessment | 46 | 90 |
| 9. Success Criteria | 61 | 80 |
| 10. Logical Consistency | 51 | 90 |
| **Total** | **684** | **1000** |

**Rubric deduction rules audit**: 无占位符（TBD/TODO）；无明显未量化模糊语（"零行为变化""约 20 处"等均有量化或限定）；无稻草人备选。扣分全部来自准则级缺口与被证伪断言，非规则罚分。
