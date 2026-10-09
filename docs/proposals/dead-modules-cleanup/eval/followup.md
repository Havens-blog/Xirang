# Follow-up: internal/task 注释级悬空引用清单（10 处）

- **来源**: 提案 `docs/proposals/dead-modules-cleanup/proposal.md`（Out of Scope 登记项的跟踪锚点）
- **触发条件**: `internal/task` 删除后，以下注释成为悬空引用
- **清除载体**: 随 CMDB 决策 0001 的文档一致性 pass 一并处理
- **判定方法**: import 语句级 grep 对这些行不命中（均为注释级）；逐行人工改写或删除

## cam 侧（2 处）

| # | 位置 | 内容摘要 |
|---|------|---------|
| 1 | `internal/cam/service/asset_sync.go:13` | "旧版/备用执行器，不参与运行时" |
| 2 | `internal/cam/task/executor/sync_assets.go:12` | "旧版/备用执行器，不参与运行时" |

## cert 侧（8 处）

| # | 位置 | 内容摘要 |
|---|------|---------|
| 3 | `internal/cert/service/execute_service.go:25` | 子任务执行器挂 internal/task 框架 |
| 4 | `internal/cert/service/execute_service.go:26` | 指导性注释："参考 internal/task executor 既有用法" |
| 5 | `internal/cert/service/execute_service.go:114` | SubtaskDispatcher 派发端口 → internal/task 框架 |
| 6 | `internal/cert/service/execute_service.go:962` | internal/task 框架挂载（pkg/taskx） |
| 7 | `internal/cert/service/execute_service.go:968` | ChangeItemExecutor 与 internal/task 既有用法对齐 |
| 8 | `internal/cert/service/execute_service_test.go:1208` | internal/task 框架挂载注释 |
| 9 | `internal/cert/module.go:231` | 变更项子任务执行器注册到 internal/task 队列 |
| 10 | `internal/cert/scheduler/inspection_job.go:3` | 注册挂入 internal/task 框架由 7.1 完成 |

（清单 2026-09-09 与 proposal.md Evidence 节核对一致。）
