---
created: "2026-09-09"
author: "Haven"
status: Approved
intent: "cleanup"
---

# Proposal: 悬空模块清理（internal/{task,sync,servicetree} 删除 + account/asset 定位正名）

## Problem

2026-09-08 架构审查确认 e-cam-service 存在多组"迁移半成品"模块：既无外部消费者、又带着误导性的"待迁移"文档，持续抬高理解成本与误改风险。本轮一次清理，并让两个半活模块（account/asset）的文档说真话。

### Evidence

仅统计 import 语句（排除注释）的依赖实测（2026-09-09）：

- `internal/task`：外部 importer **0**。注释级提及共 **10 处**：cam 侧 2 处（"旧版/备用执行器，不参与运行时"，`internal/cam/service/asset_sync.go:13`、`internal/cam/task/executor/sync_assets.go:12`）；cert 侧 8 处（internal/cert/ 下）：execute_service.go 25/26/114/962/968、execute_service_test.go:1208、module.go:231、scheduler/inspection_job.go:3，含指导性注释"参考 internal/task executor 既有用法"。注意与活跃的 `internal/cam/task` 同名不同包。
- `internal/sync`：外部 importer **0**（与活跃的 `internal/cam/sync` 同名不同包）。
- `internal/servicetree`：外部 importer **0**；服务树实际接线的是 `internal/cam/servicetree`（`camModule.ServiceTreeModule`）。
- `internal/account`：半活。`repository(+dao)`、`service` 被 cert 域全家（module + 3 services + tests）、logquery、mcp、cam-cost、cam-task（`internal/cam/task/module.go:6`、`executor/sync_billing.go:11`）、ioc/{cert,logquery}.go、cmd/tmp-k8s-cred 共约 20 处消费；死的只有 `doc.go` 的过时迁移计划（其所述"迁移到 account/"方向已被放弃，cam 域自有实现为权威写路径）与 `module.go` 外壳（0 importer）。
- `internal/asset`：半活。删掉 task/servicetree 后仅剩 cert 域（覆盖率分母只读聚合，`cert/service/coverage_meta.go` Hard Rule 注释自证）+ `ioc/cert.go` 消费；死的只有 `module.go` 外壳（0 importer）。
- `cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`、`cmd/tmp-k8s-diag`：三只同族临时诊断工具（`package main` 一次性只读诊断），非部署单元；均未入 git（2026-09-09 实测 `git ls-files` 为空），删除即无法从历史找回。
- `ioc/`、`Makefile`、`main.go` 对上述三死模块**零引用**，删除无构建面影响。

### Urgency

- 新云接入、CMDB 收敛等主线重构都要求先看懂模块图；悬空模块让"哪份是实现"永远需要二次考证。
- `internal/servicetree` 与 `cam/servicetree`、`internal/task` 与 `cam/task`、`internal/sync` 与 `cam/sync` 三对同名包并存，是最容易改错文件的地方。
- CMDB 收敛（决策 0001）Phase 4 将删除 `e-cam-service/internal/cmdb`，届时若悬空模块仍在，删除面会重叠、返工成本更高。

## Proposed Solution

纯删除 + 文档正名，零行为变化：

1. **整目录删除**：`internal/task`、`internal/sync`、`internal/servicetree`、`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`、`cmd/tmp-k8s-diag`（tmp 三只删除前留档，处置方式按 Risk 4/Risk 5 区分）。
2. **account 正名**：删除零引用的 `internal/account/module.go`；重写 `doc.go`，**按代码事实写**——消费方列全至文件级（cert 域全家/logquery/mcp/cam-cost/cam-task/ioc，以 SC-3 删除后 grep 清单为准）；"只读"降级为"以读为主"（`SyncAccount` 经 `taskQueue.Submit` 触发同步，`internal/account/service/account.go:338`）；如实登记坏路径——mcp 注入 nil 队列（`internal/mcp/deps.go:59`），`sync_account` 调用即 panic，处置随 CMDB 决策 0001 一并决断。
3. **asset 正名**：删除零引用的 `internal/asset/module.go`；重写 `doc.go`——"cert 域覆盖率分母的只读聚合数据源；直写 c_instance 的历史包袱与归属随 CMDB 决策 0001 Phase 4 一并决断"。
4. **决策文档联动**：`D:\Haven\docs\decisions\0001-cmdb-consolidation-keep-ecmdb.md` 的 Phase 4 补记 asset 归属决断项，防止遗忘。注意坐标系：该文件在**工作区根** `D:\Haven\docs\decisions\`，不在 e-cam-service 仓库内。

用户体验：无（纯内部清理）；体验改善发生在开发者侧——模块图在 **import 级**与运行时事实一致（三死模块零 importer、doc.go 与实测消费方一一对应）。口径说明：live 代码中 10 处注释级悬空引用不在本轮（已锚定 follow-up，见 Out of Scope），注释级的"文档与事实一致"待该轮完成后成立。

### Innovation Highlights

无创新，属常规死代码清理。唯一方法论点：以"仅 import 语句"为判死标准（排除注释引用），避免把"文档提到"误当"代码依赖"。

## Requirements Analysis

### Key Scenarios

清理类无用户场景。关键"安全场景"：

- 删除后 `go build ./...`、`go vet ./...`、全量 `go test ./...` 的**失败集合必须是基线失败集合的子集**（统一判据：只允许减少、不允许新增；不要求失败集合相同，通过集合随测试宇宙变化不作判据。基线本非全绿，存量红灯见 SC-5；基线失败集合删除前留档）。
- 同名包保护：删除路径必须精确匹配 `internal/{task,sync,servicetree}`，不得波及 `internal/cam/{task,sync,servicetree}`。
- 在途分支冲突：某在途分支若正在改动引用被删路径的代码，合并删除 commit 时会构建失败。处理：删除单独成 commit，回滚即 revert 该 commit；在途分支 rebase 后按 followup.md 清单移除失效 import。死模块零 importer，正常不应出现此情形——出现即说明有未登记的消费者，须停下重查 Evidence 结论。
- 部分删除失败：删除中途 build 红灯暴露某目录仍有隐藏消费者。处理：revert 删除 commit 即恢复原状（tmp 工具已按 SC-1 在删除前留档，无内容丢失）；排查后重新发起，已通过的验收项不受影响。

### Non-Functional Requirements

- 零行为变化：不改任何活跃代码路径、不改 Mongo 数据与索引。
- 构建产物不变（tmp 工具均非部署单元；`go build ./...` 通过即可）。

### Constraints & Dependencies

- 依赖决策 0001（CMDB 收敛）已存在，本清理不改变其任何结论，仅补记联动项。
- 先记录基线失败集合（build/vet/test 输出留档）作为验收证据；删除单独成 commit，不与在途变更混提交。
- account/asset 的"收编迁移"明确不在本轮（受 CMDB 决策影响，避免二次返工）。

## Feasibility Assessment

- 工作量：**半天内单人完成**。内容为纯目录删除 + 两个 `doc.go` 重写 + 一处决策文档补记，无代码编写；唯一耗时项是 SC-3 要求的删除后消费方 grep 与 doc.go 清单逐文件比对，以及 SC-1 的留档步骤。
- 验收可行性：验收全部由既有工具链承载（`go build`/`go vet`/`go test` + git + grep），无新工具引入；`go vet ./...` 在当前树已实测通过（2026-09-09），SC-5 的 build/vet 闸门在动手前即已满足。
- 带宽与技能：操作者为项目维护者（提案作者本人），不依赖任何新技能或外部资源；无运行时依赖、无停机窗口要求，可插入任意空档执行，无排期约束。

## Alternatives & Industry Benchmarking

### Industry Solutions

Go 生态的行业参照：`staticcheck` 的 **U1000**（unused）检查、`golang.org/x/tools/cmd/deadcode`（基于调用图报告不可达函数）、golangci-lint 的 `unused` linter 是三类主流死代码探测工具；对仍有消费者的代码，通行模式是 deprecate-then-delete（先标记弃用、留观察窗口，再删）。本提案不引入探测工具的原因：三死模块外部 importer 实测为 0，工具能给出的信息已被 import 级实测覆盖且更精确（工具不处理 `internal/task` vs `cam/task` 同名不同包的歧义）；deprecate-then-delete 仅对仍有消费者的代码有意义，此处消费面为零，"删除 + 编译验证"即终点（Go 无死代码链接问题）。"先收编再删"的取舍量化见下表。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 保守清理（本轮） | 本提案 | 零行为风险；文档与事实一致；为 CMDB Phase 4 铺路 | account/asset 仍位于独立顶层目录 | **采纳** |
| 彻底收编（account→cam、asset→cert） | account/doc.go 原迁移计划 | 目录彻底干净 | 约 32 个 import 站点改动（2026-09-09 实测，剔除待删 task 模块后；全仓现状约 41）；涉及 cert 域全家 + 3 服务 + 测试文件的批量改动，与 CMDB 决策 0001 后续阶段重叠，asset 归属存在二次返工风险 | Rejected: 收益为单一目录归属（asset 直属 cert 域、account 不再游离顶层）并消除 doc.go 与目录事实再漂移的维护面，但模块图问题已由 doc.go 正名解决，收益不成比例于约 32 个 import 站点改动；且 asset 归属受 CMDB 决策 0001 影响可能二次返工 |
| Do nothing | — | 零成本 | 同名包误改风险持续；CMDB Phase 4 删除面重叠 | Rejected: 已确认三对同名包并存 |

## Scope

### In Scope

- 删除 `internal/task`、`internal/sync`、`internal/servicetree`（整目录）
- 删除 `cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`、`cmd/tmp-k8s-diag`（整目录；删除前留档：tmp-cas-diag/tmp-k8s-diag 先提交入库再删，tmp-k8s-cred 摘录诊断逻辑留档后删——见 Risk 5）
- 删除 `internal/account/module.go`、`internal/asset/module.go`（零引用外壳）
- 重写 `internal/account/doc.go`、`internal/asset/doc.go`（真实定位 + 指向决策 0001）
- `D:\Haven\docs\decisions\0001-cmdb-consolidation-keep-ecmdb.md`（工作区根 `D:\Haven\docs\decisions\`，非 e-cam-service 仓库内）Phase 4 补记 asset 归属决断项

### Out of Scope

- account/asset 的收编迁移（随 CMDB 决策 0001 后续阶段）
- 同步逻辑三代收敛（另立提案）
- `internal/cam/{task,sync,servicetree}`（全部活跃，不动）
- Mongo 数据/索引清理（只删代码，不动数据）
- live 代码中指向 `internal/task` 的注释级引用清理（10 处，清单见 Evidence）：删除后成悬空注释；登记为 follow-up，**跟踪锚点**：本提案 eval 目录 `eval/followup.md`（本轮已建，逐条 file:line，与 Evidence 清单一致），随 CMDB 决策 0001 的文档一致性 pass 一并清除——无锚点即视为未登记，与"Phase 4 补记防遗忘"同一方法。不纳入理由：守"纯删除、不动 live 代码"边界，不与删除混 commit。

## Risks & Mitigations

评级口径：L = 可能性，I = 影响（低/中/高）。

| # | Risk | L | I | Mitigation |
|---|------|---|---|-----------|
| 1 | 同名包误删（cam/task、cam/sync、cam/servicetree 均活跃） | 低 | 高 | 删除用精确路径；删除后 `go build ./...` 验证活跃模块完好 | |
| 2 | 死模块存在未知隐式注册（init() 副作用） | 低 | 中 | 已实测零外部 importer（无 import 即无 init 触发）；删除后全量测试兜底 | |
| 3 | account/asset 新 doc.go 定位随后续迁移漂移 | 中 | 低 | doc.go 显式链接决策 0001 的 Phase 4 联动项作为唯一事实源 | |
| 4 | tmp 工具被某人本地工作流依赖 | 低 | 中 | 三只均未入 git（见 Evidence），"从历史找回"不成立；缓解为删除前留档，处置方式按 Risk 5 区分：tmp-cas-diag/tmp-k8s-diag 先提交入库再删（提交信息注明存档用途），tmp-k8s-cred 走摘录留档 | |
| 5 | 存档入史扩大涉敏面：`cmd/tmp-k8s-cred` 是凭证处理工具（从 Mongo 读账户记录、解密 AccessKey 并调云 API，`main.go:55`），按"先提交再删"整目录入库会把该涉敏调用链永久写进 git 历史——不可逆，且本提案无历史重写计划 | 低 | 中 | 区分处置：`tmp-k8s-cred` 不整目录入库，仅摘录关键诊断逻辑（Mongo 查询结构与调用链说明，不含凭证读取代码）留档至本提案目录后再删；`tmp-cas-diag`/`tmp-k8s-diag` 已核实无硬编码凭证（均经 `v.GetString` 读配置），确认后整目录入库 | |
| 6 | 悬空注释主动误导：删除后至 CMDB 文档一致性 pass 前，live 代码中 10 处 `internal/task` 注释仍指向不存在的路径，其中 `internal/cert/service/execute_service.go:26`"参考 internal/task executor 既有用法"是指导性注释，读者会按图索骥扑空 | 中 | 低 | 清单已逐条登记 `eval/followup.md`（file:line，与 Evidence 一致）；删除 commit 的 PR 描述/提交公告中列明悬空注释清单并注明"待 follow-up 清除" | |
| 7 | 并发在途分支引用被删路径，合并删除 commit 时构建失败 | 低 | 中 | 删除单独成 commit（可整体 revert，见 Key Scenarios 处置）；删除 commit 公告列明被删路径，在途分支 rebase 后按 followup.md 移除失效 import；出现此类引用即说明存在未登记消费者，须停下重查 Evidence 结论 | |

## Success Criteria

- [ ] 存档先行：`cmd/tmp-cas-diag`、`cmd/tmp-k8s-diag` 的存档 commit 存在（`git log --oneline -- cmd/<dir>` 非空），`cmd/tmp-k8s-cred` 的诊断逻辑摘录文件存在于本提案目录；随后 `internal/task`、`internal/sync`、`internal/servicetree`、三只 `cmd/tmp-*` 目录均不存在
- [ ] import 语句级 grep：`internal/task`、`internal/sync`、`internal/servicetree` 零命中（`internal/cam/*` 同名路径除外）
- [ ] `internal/account`、`internal/asset` 无 `module.go`，`doc.go` 描述与消费方事实一致并链接决策 0001。机械判据：**删除完成后**对两包做 import 级 grep 得外部消费方清单（必须在删除后执行——删除前该 grep 会把三个待删对象自身计入：`cmd/tmp-k8s-cred`，以及 `internal/task/module.go:8` import account、`internal/servicetree/service/rule_engine.go:9-10` import asset），`doc.go` 所列与之一一对应（粒度为**文件级**：grep 命中的每个消费方文件须在 doc.go 清单中，多列、漏列均判失败；不要求到行级）
- [ ] `D:\Haven\docs\decisions\0001`（工作区根）Phase 4 含 asset 归属决断记录
- [ ] `go build ./...`、`go vet ./...` 通过；`go test ./...` 失败集合是清理前基线失败集合的**子集**（统一判据：只允许减少、不允许新增，不要求两集合相同——删除会改变测试宇宙，通过集合作废）（存量红灯例：`internal/cert/domain` 的 `TestBsonTagsMatchSchema/K8sCredential`，bson `displayName` 与 `schema.sql` 脱节，与本次删除无关）。flaky 处置：基线与删除后失败集合各以**单次连续运行**（`go test -count=1 ./...`）为准，禁止反复重跑取绿；出现新增失败时先重跑 1 次确认可复现，仍新增即判失败；同一测试两次结果不一致（flaky）按基线失败集合成员处理，不计为新增

## SC Consistency Check

- SC-1（目录删除）与 SC-5（构建通过）：可满足性已由"零外部 importer + ioc/Makefile 零引用"证据链支持，无矛盾。
- SC-2 与 SC-1：SC-2 是 SC-1 的可验证表达（grep 判据），其中显式豁免 `internal/cam/*` 同名活跃路径，与 Out of Scope 一致。
- SC-3 与 In-Scope（account/asset 仅删外壳+正名）：一致；SC-3 不要求收编迁移，与 Out of Scope 不冲突。
- SC-4 仅涉及文档追加，与 SC-5 构建判据无耦合。
- 结论：无 SC↔SC、SC↔InScope 矛盾。以上为作者自检；执行时以机械验收复核（SC-3 判据 + SC-2 的 import 级 grep），不依赖作者自证。
