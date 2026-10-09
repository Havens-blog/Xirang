# Freeform Expert Review

**评审人视角**：长期维护大型 Go 服务的重构架构师（多租户/多云服务、行为保持型结构清理）。
**评审方式**：对提案中每一条可证伪的断言逐条与 D:\Haven\e-cam-service 实际代码对账（import 语句级，区分注释与真实引用），证据不足不下结论，代码反驳的断言直接点名。

## Background Assessment

这份提案要解决的是"迁移半成品"问题：`internal/task`、`internal/sync`、`internal/servicetree` 三个整目录死模块、两只一次性诊断工具，外加 `internal/account`、`internal/asset` 两个半活模块的"文档说谎"（`doc.go` 仍写着已被放弃的迁移计划），通过纯删除 + 文档正名让模块图与运行时事实一致，并为 CMDB 决策 0001 Phase 4 铺路。核心方法论是以"仅 import 语句"为判死标准，排除注释提及。

我把提案 Evidence 节的每条断言都对了一遍代码，先说结论：**判死的核心证据链是站得住的**，这不是一份靠想象写出来的提案——

- `internal/task`、`internal/sync`、`internal/servicetree` 的确零外部 importer：全仓 import 语句级检索，命中全部落在各自目录内部（如 `internal/task/web/handler.go` 引 `internal/task` 根包属死簇自引用，整目录删除即消解）。cam 侧两处引用确为注释，且行号精确（`internal/cam/service/asset_sync.go:13`、`internal/cam/task/executor/sync_assets.go:12`）。
- `internal/servicetree` 的服务树接线确为 `internal/cam/servicetree`（`internal/cam/init.go:86` `module.ServiceTreeModule = stModule`，`ioc/gin.go:144-151` 路由注册）。
- `internal/account/module.go`、`internal/asset/module.go` 两个外壳 0 importer：全仓对这两个包根的引用不存在，所有消费都走 `account/repository|service`、`asset/domain|repository` 子包。account 外部消费我数出 23 条 import 语句（其中 2 条在死模块 internal/task 内，清理后剩 21）——提案"约 20 处"与被拒方案"约 25 个 import 站点改动"均与实测吻合，没有夸大。
- asset 清理后仅剩 cert 域 + `ioc/cert.go`：与提案一致；`cert/service/coverage_meta.go` 的 Hard Rule 注释（"coverageMeta.total 必须来自 internal/asset 聚合……本端口是分母唯一入口"）和 `NewAssetRepositoryCounts(instances assetrepo.InstanceRepository)` 证实"只读聚合分母"定性准确；cam 注释"两条同步路径写入同一个 c_instance 集合"证实"直写 c_instance 的历史包袱"说法有据。
- `ioc/`、`Makefile`、`main.go` 对三个死模块零引用：grep 证实。`docs/decisions/0001` 存在且 Phase 4 现文为"删除 `e-cam-service/internal/cmdb`、其路由注册与 collection 直写代码"——补记 asset 决断项的前提成立。
- 两个删除目标工具的包注释原话就是"一次性只读诊断（用后即删）"，Risk 4 的定性有代码自证。
- 我实测 `go build ./...` 与 `go vet ./...` 全绿。另外三个死模块内 **0 个 `_test.go` 文件**，删除不损失任何测试覆盖——这让"删除不影响测试"比提案自己声称的还要容易达成。
- 但**全量测试基线并不绿**（详见风险二）：`go test -p 1 ./...` 实测至少 `internal/cert/domain` 与 `tests/no-snapshot-guidance` 两个包失败，且失败与三个死模块毫无关系。

背景判断到此。以下是我对账后认为提案没有说对或没有说全的部分。

## Key Risks

风险：Risk 4 的缓解措施"需要时可从 git 历史找回"对它要保护的对象是假的——两只 tmp 工具根本没进过 git 历史。
> "tmp 工具被某人本地工作流依赖 | 诊断逻辑属一次性（临时排查用）；提交信息中明示删除，需要时可从 git 历史找回"

我查了：`git status --short cmd/` 显示 `?? cmd/tmp-cas-diag/`、`?? cmd/tmp-k8s-cred/`（连同第三只 `?? cmd/tmp-k8s-diag/`），`git ls-files cmd/tmp-cas-diag cmd/tmp-k8s-cred cmd/tmp-k8s-diag` 返回为空，也未命中任何 .gitignore 规则。也就是说这两只工具从未被提交过——"从 git 历史找回"这个缓解措施对唯一需要它的场景（未跟踪文件被删）恰好不成立。一旦按提案删除，121 行 + 75 行诊断逻辑永久丢失，Risk 4 等于没有缓解。这是全提案第一处被代码直接证伪的断言，必须先修。

风险：基线"当前全部通过"不成立——实测当前测试套件就是红的，SC-5 按原文写下"全部通过"在纯删除的前提下无法达成；基线同时未被钉住，测量发生在 79 项未提交变更的脏工作区上。
> "删除后 `go build ./...`、`go vet ./...`、全量 `go test ./...` 必须与删除前等价（基线：当前全部通过）。"

我实测（2026-09-09，与提案同日）：`go build ./...`、`go vet ./...` 通过；但 `go test -p 1 ./...` 至少两个包红——(1) `internal/cert/domain` 的 `TestBsonTagsMatchSchema/K8sCredential`："bson 字段与 schema.sql 不一致，got: [_id apiEndpoint clusterName createdAt displayName kubeconfig]，want: [无 displayName]"——`internal/cert/domain/k8s_credential.go:17` 的 `bson:"displayName,omitempty"` 与 `schema.sql` 脱节，时间上正对 HEAD 提交 `743a9a8`（"凭证回填扩为 fetchMeta（可读名+endpoint）"）；(2) `tests/no-snapshot-guidance` 的两个 smoke 测试期望 200 实得 202（响应体 `status: running`）——触发扫描接口已异步化，smoke 期望值未跟上，且这类测试依赖运行中的服务。两者都与三个死模块零关联，纯删除不可能把它们变绿。也就是说：SC-5 要求"全部通过，与清理前基线一致"，可清理前基线不是全绿——这个判据今天执行必然失败，失败原因还与本次清理无关，将白白产生一次"清理引入了问题"的错误归因。此外基线无 commit 可回（`git status` 79 个未提交条目，含已删的 gRPC proto、改动的 `go.mod`/`go.sum`、`internal/cam/init.go` 等），删除 commit 混在在途工作之上，"零行为变化"的审计无从做起。另注：并行 `go test ./...` 在本机直接把链接器跑崩（cmd/link 内存崩溃），本机要靠 `-p 1` 才能完成一次可信的全量跑——"全量测试兜底"这条缓解在本环境的执行成本比提案设想的高。

风险：account 的"正名"文案与代码事实不符，新 `doc.go` 将出生即不准确——与本次清理要治的病是同一个病。
> "重写 doc.go——"跨域账号只读访问层（cert/logquery/mcp/cam-cost 共享）；权威写路径在 cam 域自有实现；归属收编迁移随 CMDB 决策 0001 一并决断""

三处对不上账。其一，"只读"被 mcp 反驳：`internal/mcp/tools_sync.go:85` 调用 `s.deps.AccountSvc.SyncAccount(ctx, accountID, req)`，而 internal/account 的 `SyncAccount`（`internal/account/service/account.go:296` 起）会构造 `taskx.Task` 并 `s.taskQueue.Submit(t)`（:338）——这是一条触发资产同步的状态变更路径，不是只读。其二，这条路径在 mcp 的接线里还是坏的：`internal/mcp/deps.go:59` 用 `NewCloudAccountService(accountRepo, nil, logger)` 注入了 **nil** 队列，而 `pkg/taskx/queue.go:226` 的 `Submit` 第一行就是 `q.mu.RLock()`，nil 接收者直接 panic——MCP 的 sync_account 工具一调用就是运行时崩溃。doc.go 若按提案文本宣称"只读访问层"，等于把一颗已经埋好的雷描述成了和平地带。其三，消费方清单漏了 `internal/cam/task`（`internal/cam/task/module.go:6` 与 `executor/sync_billing.go:11` 消费 `accountservice.CloudAccountService`），新 doc.go 罗列 "cert/logquery/mcp/cam-cost" 会导致文档漏掉一个真实消费方。附带一提：`"cam:sync_assets"` 这个任务类型字符串在 `internal/account/service/account.go:19` 与 `internal/cam/task/executor/sync_assets.go:40` 双重定义，是一条纯靠字符串对齐的跨包契约——这正是"正名"文档最该记录的耦合，提案文本只字未提。

风险：Evidence 节对注释级引用的盘点不完整，删除后将在 live 代码里留下 10 处指向已删包的悬空注释，提案自己定义的收益（文档与事实一致）只兑现了一半。
> "`internal/task`：外部 importer **0**。cam 侧两处仅为注释提及（"旧版/备用执行器，不参与运行时"，`internal/cam/service/asset_sync.go:13`、`internal/cam/task/executor/sync_assets.go:12`）"

cam 侧两处确凿且行号精确，但 Evidence 以此给出的注释全景是不完整的。全仓检索 live 代码中 `internal/task` 的注释级引用共 10 处：cam 2 处之外，cert 侧 8 处——`internal/cert/service/execute_service.go` 的 25、26、114、962、968 五处，`internal/cert/scheduler/inspection_job.go:3`、`internal/cert/module.go:231`、`internal/cert/service/execute_service_test.go:1208` 各一处，另有 `docs/features/ssl-cert-management/tasks/index.json:788` 在任务标题里写"挂 internal/task 框架"。其中 cert 的现役 taskx 执行器框架在注释里通篇被称为"internal/task 框架"，还带指令性表述"子任务执行器挂 internal/task 框架（pkg/taskx.Queue + TaskExecutor，**参考 internal/task executor 既有用法**）"（execute_service.go:25-26）。删除之后，这 8 处注释变成指向不存在路径的悬空指针，而 cert 自己的框架仍顶着死模块的名字——提案 Urgency 里最刺痛的那条"同名包并存，是最容易改错文件的地方"，会在 cert 的注释里原样存活。SC-2 只测 import 语句级，对这一切零感知；"模块图与文档第一次与运行时事实一致"这句话按当前 scope 是兑现不了的。判死标准用 import 级是对的（方法论点我认可，且实测确实有注释引用会被朴素 grep 误判），但判死标准不等于清理边界——提案把前者当成了后者。

问题：盘点漏了第三只同类工具 `cmd/tmp-k8s-diag`，而"非部署单元"的判据对它同样成立。
> "`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`：临时诊断工具，非部署单元。"

`cmd/tmp-k8s-diag/main.go`（140 行，包注释"诊断：集群凭证状态 + 最新扫描快照的 K8s 通道结果 + crd 引用落库情况 + APIServer 可达性（TCP 拨号）"）与提案要删的两只属同一族、同一形态（`package main` 一次性只读诊断）。它既不在 Evidence、不在 In Scope，也不在 Out of Scope——三种都不沾，只能理解为遗漏而非取舍。如果它因为 K8s 凭证链路排查仍在进行而被有意保留，这个理由必须写出来；否则本轮"清掉 tmp 工具"的边界就是任意画出来的。顺带指出：它同样是未跟踪文件，Risk 4 的找回问题对它同样存在。

问题：决策文档的路径锚点是工作区根，不是仓库根，与其余所有路径的坐标系不一致。
> "4. **决策文档联动**：`docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md` 的 Phase 4 补记 asset 归属决断项，防止遗忘。"

实测 `docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md` 位于 `D:/Haven/docs/decisions/`（工作区根），e-cam-service 仓库自己的 `docs/` 目录下**没有** decisions/ 目录。提案其余所有路径（`internal/...`、`cmd/...`、`ioc/...`）都是仓库相对路径，唯独这条是工作区相对路径。按提案执行的实现者会在 `e-cam-service/docs/decisions/` 下找不到文件。这属于小概率但零成本可免的执行摩擦。

## Improvement Suggestions

建议：把 Risk 4 的缓解措施改为"先提交、再删除"，并覆盖三只工具。
Addresses: Risk 4"git 历史找回不成立"与"tmp-k8s-diag 盘点缺口"两个发现。
> What changes: 在删除 commit 之前增加一个独立 commit，把 `cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`（以及处置后的 tmp-k8s-diag）先 `git add` 进历史并提交，提交信息注明"临时诊断工具入库存档，随后删除"；若决定不进主历史，则把关键诊断逻辑（如 DescribeClusterUserKubeconfig 双形态计时、cas 引用计数样例）摘录进提案或 docs 附录再删。删除 commit 的信息里写明这两步。执行后 Risk 4 的"可找回"才从假命题变成真命题。

建议：把基线从"断言"升级为"证据"，先处置既有红灯，再把 SC-5 的"等价"定义为通过/失败集合一致而非"全部通过"。
Addresses: "基线不绿 + 未钉住 + 脏工作区"风险。
> What changes: 执行顺序固定为——(1) 先将在途 79 项变更提交或 stash，得到干净基线 commit；(2) 在该 commit 上运行 `go build ./...`、`go vet ./...`、`go test -p 1 ./...` 并将输出（含 commit hash）留档到本提案或 PR 描述；当前实测已红的 `internal/cert/domain`（displayName 与 schema.sql 脱节，疑似 HEAD `743a9a8` 遗留）与 `tests/no-snapshot-guidance`（202 异步化后 smoke 期望未更新）必须二选一：要么在清理前修掉，要么在提案里显式登记为"既有红灯、与本次清理无关"，并把 SC-5 改写为"失败集合与基线逐项一致、无新增失败"；(3) 删除 + 文档正名单独成 commit，commit 内不含任何在途变更文件；(4) 删除后重跑同样三命令逐项比对。我在评审中实测过 `go build`/`go vet` 全绿、三个死模块含 0 个测试文件，这两点可写进提案作为既有证据。

建议：account 的 `doc.go` 文本按代码事实重写，消费方列全、把坏路径写实。
Addresses: account 正名文案与代码不符的风险。
> What changes: 文本至少改为——"跨域账号**读取**访问层：cert（module + 3 services + tests）、logquery、mcp、cam-cost、**cam/task** 共享；资产同步触发路径（`SyncAccount`，taskType `cam:sync_assets` 字符串与 `internal/cam/task/executor` 双重定义）在本仓存在两份实现，权威实现在 cam 域（`internal/cam/service/account.go`，经 wire 注入真实队列）；mcp 侧注入了 nil 队列，`sync_account` 工具当前调用即 panic，属已知坏路径，处置随 CMDB 决策 0001 一并决断"。若团队愿意扩一点 scope，更彻底的做法是本轮顺手删除 mcp 的 `sync_account` 工具或补上队列注入——但这违反"零行为变化"，可以作为二选一明确记录，不能默认沉默。

建议：把注释级悬空引用纳入本轮 scope（或显式声明为 follow-up 并写明理由）。
Addresses: "10 处悬空注释 + cert 注释沿用死模块命名"风险。
> What changes: 在 In Scope 增加一条"注释正名"：将 cert 侧 8 处注释中的"internal/task 框架"改称"pkg/taskx 框架"，把"参考 internal/task executor 既有用法"改指 `internal/cam/task/executor/sync_assets.go`（`pkg/taskx/README.md:347` 的示例引用本来就指向后者，改后全仓口径统一）；cam 侧两处"旧版/备用执行器"注释删除该行或改为"已随 internal/task 删除（2026-09）"。这些是注释改动，零行为风险，与本轮"文档说真话"的目标完全同质，没有理由留给下一轮。若坚持不动注释，则必须在 Out of Scope 写明"注释级悬空引用（10 处，清单见 Evidence 补充）"并立 follow-up 项，同时把 SC-2 扩一条可执行判据：`grep -rn "internal/task" --include="*.go"` 的残留命中数与处置记录一一对应。

建议：显式处置 `cmd/tmp-k8s-diag`，并在提案中写明取舍。
Addresses: tmp 工具盘点缺口问题。
> What changes: 二选一并落进文档——(a) 若其诊断目标（K8s 凭证链路）排查已完成，随本轮一并删除并纳入"先提交再删除"流程；(b) 若排查仍在进行，在 Out of Scope 写明"cmd/tmp-k8s-diag 保留至 XX 验证完成，随下次清理删除"。三条 tmp 工具要么都在盘子里，要么每条都有名字和理由，不允许第三种状态。

建议：给决策文档引用补上明确的路径锚点。
Addresses: 决策文档路径坐标系不一致问题。
> What changes: 将 In Scope 第 5 条改写为"在 `D:/Haven/docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md`（工作区根，非 e-cam-service 仓库内）的 Phase 4 补记 asset 归属决断项"，或者把决策文档迁入仓库 `docs/decisions/` 使全部引用同坐标系。补记内容建议直接对齐 Phase 4 现文（"删除 `e-cam-service/internal/cmdb`、其路由注册与 collection 直写代码"），追加一句 asset 归属与 `internal/asset` 直写 `c_instance` 包袱在该阶段的决断点。

建议：给 SC-3 补一个可机械执行的验收判据，避免"描述与消费方事实一致"沦为自由心证。
Addresses: account/asset 正名的可验证性（间接覆盖 account 文案失真风险）。
> What changes: SC-3 追加机械判据——`grep` 出 `internal/account`、`internal/asset` 的全部外部 import 站点清单（我实测 account 清理后 21 条、asset 清理后 9 条，按 file:line 列进提案），新 `doc.go` 中的消费方列表必须与该清单一一对应；account 的 doc 还需显式记录 `SyncAccount`/nil 队列现状。这样"正名"就从修辞变成可以逐行核对的事实陈述，与提案自己"仅 import 语句"的方法论一脉相承。

## 结语

这份提案的判死结论我全部复核过并认可：三个死模块零外部 importer、两只工具确属"用后即删"、同名包风险真实存在、精确路径删除 + 构建闸门是对的缓解、决策 0001 联动有据。它的证据链在**判死**这个动作上是扎实的，问题集中在**安全叙事与清理边界**：git 找回措施是假的（未跟踪文件）、基线"当前全部通过"被实测反驳（至少两个包红，且红在与清理无关的地方）、正名文案有一处被代码反驳、注释级悬空引用被系统性低估、第三只 tmp 工具无理由缺席。以上几处修完，这就是一次可以放心执行的清理；修不完，至少要把每一处的"不做"写成显式决定，而不是留成默认。
