---
id: "1"
title: "后端：规则基础修复（node_name 回填/region 字段/执行结果落库）"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.feature"
mainSession: false
---

# 1: 后端：规则基础修复（node_name 回填/region 字段/执行结果落库）

## Description
**代码仓：D:/Haven/e-cam-service**（Go）。修规则引擎三个基础缺陷，为 dry-run/调度/前端呈现打底：
① `ListRules` VO 回填 `node_name`（按 NodeID 批量查节点名，当前恒空致前端显示「-」）；
② `rule_engine.go getFieldValue` 补 `region` case（UI 已有该条件选项但后端无此 case 永不匹配；region 在
Instance attributes 中，对齐既有 attributes 取值）；③ 执行结果持久化：BindingRule 增加/复用执行统计字段
（last_executed_at、last_match_count），ExecuteRules 执行后落库，列表接口透出。

## Reference Files
- `docs/proposals/servicetree-overhaul-p1p2/proposal.md` — 一期方案 1/4、SC1
- `D:/Haven/e-cam-service/internal/cam/servicetree/web/handler.go:628` — ListRules（VO 组装处补回填）
- `D:/Haven/e-cam-service/internal/cam/servicetree/service/rule_engine.go:234` — getFieldValue（补 region）
- `D:/Haven/e-cam-service/internal/cam/servicetree/domain/rule.go` — BindingRule（执行统计字段）
- `D:/Haven/e-cam-service/internal/cam/servicetree/repository/dao/rule.go` — DAO 字段/迁移

## Acceptance Criteria
- [ ] ListRules 的 RuleVO 回填 `node_name`（批量查询节点名映射，N+1 避免）；前端 `row.node_name` 可直接用
- [ ] `getFieldValue` 支持 `region`（从 Instance attributes 取 region，缺失返回空串不匹配）
- [ ] ExecuteRules 执行后规则级统计落库（last_executed_at/last_match_count），ListRules 透出
- [ ] `go build ./...` + `go test ./internal/cam/servicetree/...` 全绿（新增 region/回填单测）
- [ ] `git diff` 仅 servicetree 模块内文件

## Hard Rules
仅改 `internal/cam/servicetree/` 内文件。**不碰** task/executor（任务 3 挂点）、前端、
并发会话在途文件（cmd/apm-push 等）、cert/logs。提交 `git -C D:/Haven/e-cam-service add <pathspec>` 显式路径。
DAO 字段新增注意 MongoDB 集合兼容（缺字段零值容忍，勿 require）。

## Implementation Notes
- 现有 2 条规则 NodeID 均有值（域校验强制），node_name 回填即可让前端「-」消失。
- region 取值对齐 Instance attributes 的实际键名（先查采集侧 region 落在 attributes 的哪个键，如 `region`）。
- 统计字段若 DAO 无对应列，新增 bson 字段即可（Mongo 无需迁移）。
