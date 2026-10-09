---
id: "2"
title: "后端：规则 dry-run 预览 API（不落库）"
priority: "P0"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.feature"
mainSession: false
---

# 2: 后端：规则 dry-run 预览 API（不落库）

## Description
**代码仓：D:/Haven/e-cam-service**。新 API：按规则条件（可不落库的临时规则体）预览命中资产清单——
`POST /service-tree/rules/dry-run`，请求体 = BindingRule 条件（conditions + 可选 node/env），响应 =
命中资产列表（asset_id/asset_name/provider/region/当前绑定状态）。复用 `MatchInstance` 逐资产匹配，
**只读不落库**。规则保存前验证条件写对没有。

## Reference Files
- `docs/proposals/servicetree-overhaul-p1p2/proposal.md` — 一期方案 3、SC3
- `D:/Haven/e-cam-service/internal/cam/servicetree/service/rule_engine.go:100` — MatchInstance（复用）
- `D:/Haven/e-cam-service/internal/cam/servicetree/web/handler.go` — 路由注册模式（:70 execute 先例）

## Acceptance Criteria
- [ ] 新路由 `POST /service-tree/rules/dry-run`：接受 conditions（与 RuleCondition 同构），返回命中资产清单（含当前绑定状态：未绑/手动绑节点/规则绑节点）
- [ ] 只读：不产生任何 binding 落库、不改规则统计
- [ ] 复用 MatchInstance/matchCondition 逻辑（不复制匹配实现）；资产遍历复用 ExecuteRules 的取数路径
- [ ] 单测：命中/不命中/正则条件/条件为空 400；`go build` + `go test ./internal/cam/servicetree/...` 全绿
- [ ] `git diff` 仅 servicetree 模块内文件

## Hard Rules
仅改 `internal/cam/servicetree/` 内文件。**不碰** task/executor、前端、并发在途文件。提交显式 pathspec。

## Implementation Notes
- 资产取数：ExecuteRules 已有遍历资产的路径（rule_engine.go:126 附近），抽公共函数或直接复用其查询。
- 命中资产带「当前绑定状态」便于用户判断规则会新增/跳过哪些（手动绑定跳过语义在任务 3 落地，dry-run 如实标注）。
- 数量上限保护：单次 dry-run 返回上限（如 500 条）+ total 计数，防全租户大列表。
