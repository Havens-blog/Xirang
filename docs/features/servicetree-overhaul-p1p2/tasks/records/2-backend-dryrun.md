---
status: "completed"
started: "2026-09-21 18:38"
completed: "2026-09-21 18:38"
time_spent: ""
---

# Task Record: 2 后端：规则 dry-run 预览 API（不落库）

## Summary
新增 POST /service-tree/rules/dry-run 规则试运行预览 API：按临时规则条件（RuleCondition 同构 + 可选 node/env）返回命中资产清单（asset_id/asset_name/provider/region/当前绑定状态 unbound/manual/rule + 已绑节点名回填），只读不落库。复用 matchRule/matchCondition 匹配实现，资产遍历抽 listTenantInstances 与 ExecuteRules 共用取数路径；结果上限 500 条 + total 真实计数（capped 标记）。条件为空返回 400。

## Changes

### Files Created
- internal/cam/servicetree/domain/dryrun.go
- internal/cam/servicetree/service/rule_engine_dryrun_test.go
- internal/cam/servicetree/web/handler_dryrun_test.go

### Files Modified
- internal/cam/servicetree/service/rule_engine.go
- internal/cam/servicetree/web/handler.go
- internal/cam/servicetree/web/vo.go

### Key Decisions
- 绑定状态取值复用 BindType（manual/rule），新增 BindStatusUnbound=unbound；提供 env_id 时按该环境口径判定绑定状态
- DryRunRules 与 ExecuteRules 共用抽出的 listTenantInstances 取数路径，匹配走同一 matchRule 实现（不复制逻辑）
- dry-run 不读取已保存规则列表，仅按请求体条件匹配，天然无规则统计写入路径

## Test Results
- **Tests Executed**: Yes
- **Passed**: 11
- **Failed**: 0
- **Coverage**: 26.8%

## Acceptance Criteria
- [x] 新路由 POST /service-tree/rules/dry-run 返回命中资产清单（含当前绑定状态）
- [x] 只读：不产生 binding 落库、不改规则统计
- [x] 复用 MatchInstance/matchCondition 与 ExecuteRules 资产取数路径
- [x] 单测：命中/不命中/正则/条件为空 400；go build + go test ./internal/cam/servicetree/... 全绿
- [x] git diff 仅 servicetree 模块内文件（显式 pathspec 提交）

## Notes
commit 2c28d8c on main (e-cam-service)。全包覆盖率 26.8%（service）/5.5%（web），为既有未覆盖代码摊薄；新增 dry-run 路径全覆盖。-race 需 CGO，本机 CGO_ENABLED=0 未跑。工作区另有并发在途文件 docs/features/rds-ops-insight/tasks/index.json 的改动，未纳入本次提交。
