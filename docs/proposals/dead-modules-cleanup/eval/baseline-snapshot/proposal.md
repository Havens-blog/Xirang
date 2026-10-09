---
created: "2026-09-09"
author: "Haven"
status: Draft
intent: "cleanup"
---

# Proposal: 悬空模块清理（internal/{task,sync,servicetree} 删除 + account/asset 定位正名）

## Problem

2026-09-08 架构审查确认 e-cam-service 存在多组"迁移半成品"模块：既无外部消费者、又带着误导性的"待迁移"文档，持续抬高理解成本与误改风险。本轮一次清理，并让两个半活模块（account/asset）的文档说真话。

### Evidence

仅统计 import 语句（排除注释）的依赖实测（2026-09-09）：

- `internal/task`：外部 importer **0**。cam 侧两处仅为注释提及（"旧版/备用执行器，不参与运行时"，`internal/cam/service/asset_sync.go:13`、`internal/cam/task/executor/sync_assets.go:12`）。注意与活跃的 `internal/cam/task` 同名不同包。
- `internal/sync`：外部 importer **0**（与活跃的 `internal/cam/sync` 同名不同包）。
- `internal/servicetree`：外部 importer **0**；服务树实际接线的是 `internal/cam/servicetree`（`camModule.ServiceTreeModule`）。
- `internal/account`：半活。`repository(+dao)`、`service` 被 cert 域全家（module + 3 services + tests）、logquery、mcp、cam-cost、ioc/{cert,logquery}.go、cmd/tmp-k8s-cred 共约 20 处消费；死的只有 `doc.go` 的过时迁移计划（其所述"迁移到 account/"方向已被放弃，cam 域自有实现为权威写路径）与 `module.go` 外壳（0 importer）。
- `internal/asset`：半活。删掉 task/servicetree 后仅剩 cert 域（覆盖率分母只读聚合，`cert/service/coverage_meta.go` Hard Rule 注释自证）+ `ioc/cert.go` 消费；死的只有 `module.go` 外壳（0 importer）。
- `cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`：临时诊断工具，非部署单元。
- `ioc/`、`Makefile`、`main.go` 对上述三死模块**零引用**，删除无构建面影响。

### Urgency

- 新云接入、CMDB 收敛等主线重构都要求先看懂模块图；悬空模块让"哪份是实现"永远需要二次考证。
- `internal/servicetree` 与 `cam/servicetree`、`internal/task` 与 `cam/task`、`internal/sync` 与 `cam/sync` 三对同名包并存，是最容易改错文件的地方。
- CMDB 收敛（决策 0001）Phase 4 将删除 `e-cam-service/internal/cmdb`，届时若悬空模块仍在，删除面会重叠、返工成本更高。

## Proposed Solution

纯删除 + 文档正名，零行为变化：

1. **整目录删除**：`internal/task`、`internal/sync`、`internal/servicetree`、`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`。
2. **account 正名**：删除零引用的 `internal/account/module.go`；重写 `doc.go`——"跨域账号只读访问层（cert/logquery/mcp/cam-cost 共享）；权威写路径在 cam 域自有实现；归属收编迁移随 CMDB 决策 0001 一并决断"。
3. **asset 正名**：删除零引用的 `internal/asset/module.go`；重写 `doc.go`——"cert 域覆盖率分母的只读聚合数据源；直写 c_instance 的历史包袱与归属随 CMDB 决策 0001 Phase 4 一并决断"。
4. **决策文档联动**：`docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md` 的 Phase 4 补记 asset 归属决断项，防止遗忘。

用户体验：无（纯内部清理）；体验改善发生在开发者侧——模块图与文档第一次与运行时事实一致。

### Innovation Highlights

无创新，属常规死代码清理。唯一方法论点：以"仅 import 语句"为判死标准（排除注释引用），避免把"文档提到"误当"代码依赖"。

## Requirements Analysis

### Key Scenarios

清理类无用户场景。关键"安全场景"：

- 删除后 `go build ./...`、`go vet ./...`、全量 `go test ./...` 必须与删除前等价（基线：当前全部通过）。
- 同名包保护：删除路径必须精确匹配 `internal/{task,sync,servicetree}`，不得波及 `internal/cam/{task,sync,servicetree}`。

### Non-Functional Requirements

- 零行为变化：不改任何活跃代码路径、不改 Mongo 数据与索引。
- 构建产物不变（tmp 工具均非部署单元；`go build ./...` 通过即可）。

### Constraints & Dependencies

- 依赖决策 0001（CMDB 收敛）已存在，本清理不改变其任何结论，仅补记联动项。
- account/asset 的"收编迁移"明确不在本轮（受 CMDB 决策影响，避免二次返工）。

## Alternatives & Industry Benchmarking

### Industry Solutions

常规做法即"删除 + 编译验证"（Go 生态无死代码链接问题，删除即终点）；更进一步的做法是先收编再删（本仓曾有 internal/account/doc.go 的迁移计划，方向已被放弃）。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 保守清理（本轮） | 本提案 | 零行为风险；文档与事实一致；为 CMDB Phase 4 铺路 | account/asset 仍位于独立顶层目录 | **采纳** |
| 彻底收编（account→cam、asset→cert） | account/doc.go 原迁移计划 | 目录彻底干净 | 约 25 个 import 站点改动；asset 归属受 CMDB 决策影响可能二次返工 | Rejected: 收益不成比例 |
| Do nothing | — | 零成本 | 同名包误改风险持续；CMDB Phase 4 删除面重叠 | Rejected: 已确认三对同名包并存 |

## Scope

### In Scope

- 删除 `internal/task`、`internal/sync`、`internal/servicetree`（整目录）
- 删除 `cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`（整目录）
- 删除 `internal/account/module.go`、`internal/asset/module.go`（零引用外壳）
- 重写 `internal/account/doc.go`、`internal/asset/doc.go`（真实定位 + 指向决策 0001）
- `docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md` Phase 4 补记 asset 归属决断项

### Out of Scope

- account/asset 的收编迁移（随 CMDB 决策 0001 后续阶段）
- 同步逻辑三代收敛（另立提案）
- `internal/cam/{task,sync,servicetree}`（全部活跃，不动）
- Mongo 数据/索引清理（只删代码，不动数据）

## Risks & Mitigations

| # | Risk | Mitigation |
|---|------|-----------|
| 1 | 同名包误删（cam/task、cam/sync、cam/servicetree 均活跃） | 删除用精确路径；删除后 `go build ./...` 验证活跃模块完好 |
| 2 | 死模块存在未知隐式注册（init() 副作用） | 已实测零外部 importer（无 import 即无 init 触发）；删除后全量测试兜底 |
| 3 | account/asset 新 doc.go 定位随后续迁移漂移 | doc.go 显式链接决策 0001 的 Phase 4 联动项作为唯一事实源 |
| 4 | tmp 工具被某人本地工作流依赖 | 诊断逻辑属一次性（临时排查用）；提交信息中明示删除，需要时可从 git 历史找回 |

## Success Criteria

- [ ] `internal/task`、`internal/sync`、`internal/servicetree`、`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred` 目录不存在
- [ ] import 语句级 grep：`internal/task`、`internal/sync`、`internal/servicetree` 零命中（`internal/cam/*` 同名路径除外）
- [ ] `internal/account`、`internal/asset` 无 `module.go`，`doc.go` 描述与消费方事实一致并链接决策 0001
- [ ] `docs/decisions/0001` Phase 4 含 asset 归属决断记录
- [ ] `go build ./...`、`go vet ./...`、`go test ./...` 全部通过，与清理前基线一致

## SC Consistency Check

- SC-1（目录删除）与 SC-5（构建通过）：可满足性已由"零外部 importer + ioc/Makefile 零引用"证据链支持，无矛盾。
- SC-2 与 SC-1：SC-2 是 SC-1 的可验证表达（grep 判据），其中显式豁免 `internal/cam/*` 同名活跃路径，与 Out of Scope 一致。
- SC-3 与 In-Scope（account/asset 仅删外壳+正名）：一致；SC-3 不要求收编迁移，与 Out of Scope 不冲突。
- SC-4 仅涉及文档追加，与 SC-5 构建判据无耦合。
- 结论：无 SC↔SC、SC↔InScope 矛盾。
