---
status: "completed"
started: "2026-09-21 18:42"
completed: "2026-09-21 19:02"
time_spent: "~20m"
---

# Task Record: 3 后端：同步后自动执行规则（异步挂点 + 幂等）

## Summary
资产同步完成后自动执行服务树规则引擎：SyncAssetsExecutor 加完成后挂点（triggerRulesAfterSync，仅加回调不改同步既有语义），按本次落库租户触发 servicetree RuleEngineService.ExecuteRulesAsync（goroutine + panic recover + 2min 超时 + 失败仅日志，不阻塞同步主链路，同租户 in-flight 去重）。ExecuteRules 内存去重由 envID-resourceID 对齐为资源级（与 dao/init.go 既有唯一索引 tenant_id+resource_type+resource_id 一致）：重复执行零重复绑定（幂等），已绑定（含 manual 任意环境）资产整体跳过（手动优先 locked）。注入循 SetChangeTracker 既有 wire 先例：executor 本地接口 RuleAutoExecutor + SetRuleExecutor + task/module.go 透传 + ioc WireRuleExecutor 显式接线（wire_gen.go），无全局单例、无依赖环。

## Changes

### Files Created
- internal/cam/servicetree/service/rule_engine_async.go
- internal/cam/servicetree/service/rule_engine_async_test.go
- internal/cam/task/executor/sync_assets_rule_hook_test.go
- ioc/servicetree.go

### Files Modified
- internal/cam/servicetree/service/rule_engine.go
- internal/cam/task/executor/sync_assets.go
- internal/cam/task/module.go
- ioc/wire_gen.go

### Key Decisions
- 幂等唯一键确认：dao/init.go initBindingIndexes 已有资源级唯一索引 (tenant_id, resource_type, resource_id)，强于提案所述 instance+node+env，无需补齐；内存去重从 envID-resourceID 改为 resourceID 级对齐该键，顺带消除跨环境绑定触发整批 InsertMany 唯一键冲突失败的隐患
- 手动优先 locked 语义：资源已存在任意绑定（manual/rule、任意环境）即跳过规则匹配，单测锁定
- 异步职责放 servicetree 侧（ExecuteRulesAsync 内部 goroutine），executor 挂点仅同步调用立即返回；失败仅 elog，不向上传播（CDN 指标采集教训）
- 接线遵循仓库既有 SetChangeTracker/WireChangeTracker 先例（wire 不支持无返回值副作用调用，wire_gen.go 显式接线），task/module.go 透传 setter 为该模式必要组成，超出 servicetree+executor 严格目录但为 AC4 注入要求的最小必要文件
- ruleAsyncTimeout 为包级 var（默认 2min）便于测试覆写；同租户并发触发经 sync.Map in-flight 标记跳过而非排队

## Test Results
- **Tests Executed**: Yes
- **Passed**: 12
- **Failed**: 0
- **Coverage**: 29.1%

## Acceptance Criteria
- [x] 资产同步完成后异步触发 ExecuteRules（不阻塞同步返回；panic recover + 超时 + 失败仅日志）
- [x] 幂等单测：同一资产重复执行零重复绑定
- [x] 手动绑定优先单测：已手动绑定资产规则执行跳过（locked 语义）
- [x] 触发点注入方式遵循仓库既有 wire 模式（不引入全局单例）
- [x] go build ./... + go test ./internal/cam/servicetree/... ./internal/cam/task/... 全绿
- [x] git diff 仅 servicetree + task/executor 挂点最小文件（另含注入模式必要的 task/module.go 透传与 ioc 接线 2 文件）

## Notes
testsPassed=12 为本任务新增测试（TestExecuteRulesIdempotentRepeat 1 + TestExecuteRulesManualBindingPriority 4 子测试 + TestExecuteRulesAsync 4 子测试 + TestTriggerRulesAfterSync 3 子测试），既有套件（TestExecuteRulesStatsPersistence 等 servicetree/task 全量）回归全绿。coverage 29.1 为 servicetree/service 包 go test -cover 实测（executor 包 69.7）。证据：go build ./... 通过；go test ./internal/cam/servicetree/... ./internal/cam/task/... 全 ok；gofmt 变更文件干净。并发脏文件 docs/features/rds-ops-insight/tasks/index.json 未触碰、未入库；提交用显式 pathspec。触发点在 SyncAssetsExecutor.Execute 账号循环完成后、进度 95 更新前，只加完成回调不改既有逻辑语义（Hard Rule 满足）。手动同步入口（handler 直连同步）不在本挂点范围，与提案『生产主入口=异步任务队列』一致。
