---
status: "completed"
started: "2026-09-21 19:15"
completed: "2026-09-21 19:15"
time_spent: ""
---

# Task Record: 6 后端：子树资产聚合 API + 环境错绑检测

## Summary
新增 GET /service-tree/nodes/:id/asset-summary 子树资产聚合 API：返回该节点子树（含自身）聚合的 total/by_environment/by_provider/by_type/by_bind_type 分布与环境疑异清单 suspicious[]（含 asset_id/asset_name/bound_env_code/reason）。聚合先取子树节点 ID 集合（ListByPath 路径前缀一次查），再对绑定表 node_id IN 一次取全量绑定（非逐节点递归 N+1），二段批量查 CMDB 实例明细后内存归并分布。环境错绑检测双信号：资产命名模式（-prod-/-uat-/-test-/-dev-，uat 归一化为 staging）+ tag.env，与绑定环境矛盾即入 suspicious，只提示不改绑；绑定环境未知或信号非标准环境代码时不误报。

## Changes

### Files Created
- internal/cam/servicetree/domain/asset_summary.go
- internal/cam/servicetree/service/node_asset_summary_test.go

### Files Modified
- internal/cam/servicetree/service/node_asset.go
- internal/cam/servicetree/web/handler.go
- internal/cam/servicetree/wire_gen.go

### Key Decisions
- 聚合实现按任务 Implementation Notes 取「二段查」：绑定表 node_id IN 一次查全量绑定（不加 limit），再 cmdbRepo.ListByIDs 批量取实例，内存归并分布——满足 AC 的非 N+1 递归要求，未引入新 DAO 聚合查询
- byEnvironment 按环境代码（dev/test/staging/prod）为 key，通过注入 EnvironmentRepository 做 env_id→code 映射，未知环境回退 env_<id>；为此 NewNodeAssetService 增加 envRepo 参数并同步更新 wire_gen.go（注入顺序前置 environmentRepository）
- 错绑判定归一化：命名/tag.env 值小写化，uat→staging（环境代码体系无 uat）；绑定环境代码未知（自定义环境）或 tag.env 非标准环境值时不判定，避免误报
- Suspicious/分布 map 空态返回非 nil 空集合，前端可直接遍历
- handler 直接返回 domain.AssetSummary（沿用 DryRunRules 返回 domain 类型的先例），未新增重复 VO

## Test Results
- **Tests Executed**: Yes
- **Passed**: 21
- **Failed**: 0
- **Coverage**: 36.8%

## Acceptance Criteria
- [x] 新路由 GET /nodes/:id/asset-summary 返回 {total, byEnvironment{}, byProvider{}, byType{}, byBindType{}, suspicious[]}（suspicious 含 asset_id/asset_name/绑定环境/疑异原因）
- [x] 聚合覆盖子树全部节点（含自身）；聚合查询走 group by（非 N+1 递归）
- [x] 环境错绑检测：命名模式（-prod-/-uat-/-test-/-dev-）与 tag.env 双信号，矛盾即入 suspicious
- [x] 单测：多级子树聚合正确/空子树/错绑检测命中与不命中；go build + go test ./internal/cam/servicetree/... 全绿
- [x] git diff 仅 servicetree 模块内文件

## Notes
Hard Rules 遵守：仅改 internal/cam/servicetree/ 内 5 个文件，git add 显式 pathspec，未卷入并发在途的 docs/features/rds-ops-insight/。静态检查适配：golangci-lint 本机缺失用 go vet 替代；-race 需 CGO，本机 CGO_ENABLED=0（无 gcc）跳过；go test -cover 对 repository/dao 包报 go: no such tool "covdata"（Go 工具链缺 covdata），service 包正常，coverage 36.8% 为 service 包语句覆盖率（新增代码及检测辅助函数全覆盖，任务起点 29.1%）。gofmt 对 node_asset.go 报整文件 diff 系 CRLF 行尾（仓库既有状态），未整文件重排以免污染 diff；新增文件 gofmt 干净。
