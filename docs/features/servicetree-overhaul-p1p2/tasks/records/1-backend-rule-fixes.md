---
status: "completed"
started: "2026-09-21 17:53"
completed: "2026-09-21 18:12"
time_spent: "~19m"
---

# Task Record: 1 后端：规则基础修复（node_name 回填/region 字段/执行结果落库）

## Summary
规则引擎三处基础修复（一期方案 1/4、SC1）：① ListRules 回填 node_name——NodeRepository/NodeDAO 新增 GetByIDs（$in 批量查询+节点 ID 去重，N+1 避免），服务层建映射回填，RuleVO 透出 node_name；② getFieldValue 补 region case——从 Instance attributes 取 region（采集侧 sync_* 均写 attributes["region"]，与 node_asset.go 取值口径一致），缺失返回空串不匹配；③ 执行结果落库——BindingRule/dao.Rule 增加 last_executed_at/last_match_count（Mongo 零值容忍，仅 $set 不触碰业务字段），ExecuteRules 每次执行后按规则落库（未命中记 0，统计写失败仅告警不阻塞绑定主流程），RuleVO 透出执行统计。

## Changes

### Files Created
无

### Files Modified
- e-cam-service/internal/cam/servicetree/domain/rule.go
- e-cam-service/internal/cam/servicetree/repository/dao/rule.go
- e-cam-service/internal/cam/servicetree/repository/dao/node.go
- e-cam-service/internal/cam/servicetree/repository/rule.go
- e-cam-service/internal/cam/servicetree/repository/node.go
- e-cam-service/internal/cam/servicetree/service/rule_engine.go
- e-cam-service/internal/cam/servicetree/service/rule_engine_stats_test.go
- e-cam-service/internal/cam/servicetree/web/handler.go
- e-cam-service/internal/cam/servicetree/web/vo.go

### Key Decisions
- region 键名实证：采集侧 sync_ecs.go:65 等全部资产均写 attributes["region"]，node_asset.go:116 已用 GetStringAttribute("region")，故 region case 走同口径取值
- node_name 回填放在服务层（ruleEngineService.ListRules/fillNodeNames）而非 handler，handler 无节点仓储依赖，且服务层可用接口桩单测验证批量去重
- 执行统计在 ExecuteRules 匹配循环后、绑定创建前统一落库：无实例数据/零命中也记录执行时间（规则确实执行过）；统计写失败仅 Warn 不阻塞绑定主流程（对齐提案失败容忍要求）
- last_match_count 语义为「本次执行新增匹配绑定数」——已被绑定的实例跳过不计入，与前端匹配数展示直觉一致
- dao.Rule 统计字段不加入 Update() 的 $set，规则编辑不会清掉执行统计

## Test Results
- **Tests Executed**: Yes
- **Passed**: 31
- **Failed**: 0
- **Coverage**: 19.1%

## Acceptance Criteria
- [x] ListRules 的 RuleVO 回填 node_name（批量查询节点名映射，N+1 避免）；前端 row.node_name 可直接用
- [x] getFieldValue 支持 region（从 Instance attributes 取 region，缺失返回空串不匹配）
- [x] ExecuteRules 执行后规则级统计落库（last_executed_at/last_match_count），ListRules 透出
- [x] go build ./... + go test ./internal/cam/servicetree/... 全绿（新增 region/回填单测）
- [x] git diff 仅 servicetree 模块内文件

## Notes
提交 4e196b5（e-cam-service main）。测试计数来自 go test -v 实跑：31 个 PASS（含子测试）/0 FAIL；coverage 19.1% 为 servicetree/service 包 go test -cover 输出（该包 dao/web 层无既有单测基建，本次新增单测聚焦规则引擎逻辑层）。静态检查：gofmt 已应用、go vet 通过、go build ./... OK（GOFLAGS=-p=2 防 Windows OOM）；无 justfile，项目用 Makefile。新增 region/回填/落库单测共 10 个用例。git add 全程显式 pathspec，未触碰 task/executor、前端、cert/logs/apm-push 等并发在途文件。
