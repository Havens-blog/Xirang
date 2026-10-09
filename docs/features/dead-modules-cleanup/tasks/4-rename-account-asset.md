---
id: "4"
title: "account/asset 正名：删死外壳 + 按代码事实重写 doc.go"
priority: "P1"
estimated_time: "1.5h"
complexity: "high"
dependencies: [3]
surface-key: ""
surface-type: "api"
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 4: account/asset 正名：删死外壳 + 按代码事实重写 doc.go

## Description
`internal/account` 与 `internal/asset` 是半活模块：活的 repo/service 层被多域消费，死的只有 module.go 外壳（0 importer）与过时的"待迁移"doc.go。删除外壳，重写 doc.go——按代码事实写，消费方清单以任务 3 删除后的 import 级 grep 结果为准（文件级一一对应）。

## Reference Files
- `docs/proposals/dead-modules-cleanup/proposal.md` — Proposed Solution (第 2/3 条); Evidence; Success Criteria (SC-3)

## Acceptance Criteria
- [ ] `internal/account/module.go`、`internal/asset/module.go` 不存在
- [ ] 删除完成后对两包做 import 级 grep 得外部消费方清单，`doc.go` 所列与之文件级一一对应（grep 命中的每个消费方文件须在 doc.go 清单中，多列、漏列均判失败）；注意 grep 必须在删除后执行（删除前会计入 `cmd/tmp-k8s-cred`、`internal/task/module.go:8`、`internal/servicetree/service/rule_engine.go:9-10` 三个待删对象自身）
- [ ] account doc.go：定位为"跨域账号访问层（以读为主）"，如实登记写路径——`SyncAccount` 经 `taskQueue.Submit` 触发同步（`internal/account/service/account.go:338`）；坏路径——mcp 注入 nil 队列（`internal/mcp/deps.go:59`），`sync_account` 调用即 panic，处置随 CMDB 决策 0001 一并决断
- [ ] asset doc.go：定位为"cert 域覆盖率分母的只读聚合数据源"，登记直写 c_instance 的历史包袱与归属随决策 0001 Phase 4 一并决断
- [ ] 两份 doc.go 均显式链接决策 0001（`D:\Haven\docs\decisions\0001-cmdb-consolidation-keep-ecmdb.md`）作为唯一事实源

## Hard Rules
- 仅修改以下文件：`internal/account/module.go`（删）、`internal/account/doc.go`（重写）、`internal/asset/module.go`（删）、`internal/asset/doc.go`（重写）
- 不迁移、不收编（Out of Scope：随 CMDB 决策 0001 后续阶段）

## Implementation Notes
- doc.go 是 Go 源文件（package 注释），改完需 `go build ./internal/account/... ./internal/asset/...` 验证
- 消费方实测锚点：cert 域全家（module + 3 services + tests）、logquery、mcp、cam-cost、cam-task（`internal/cam/task/module.go:6`、`executor/sync_billing.go:11`）、ioc/{cert,logquery}.go——以删除后 grep 为准，此清单仅作对照
