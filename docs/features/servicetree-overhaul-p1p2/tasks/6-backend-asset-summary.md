---
id: "6"
title: "后端：子树资产聚合 API + 环境错绑检测"
priority: "P0"
estimated_time: "2.5h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.feature"
mainSession: false
---

# 6: 后端：子树资产聚合 API + 环境错绑检测

## Description
**代码仓：D:/Haven/e-cam-service**。二期可见性核心：`GET /service-tree/nodes/:id/asset-summary`——
返回该节点**子树**（含自身）聚合的资产统计：总数、按环境分布、按云平台分布、按类型（model_uid）分布、
按绑定来源（manual/rule）分布、环境疑异清单。聚合走绑定表 group by（非逐节点递归查资产）。环境错绑检测：
绑定环境与资产命名/tag.env 矛盾检测（如资产名含 `-prod-` 或 tag.env=prod 却绑 dev/test/uat → 疑异清单）。

## Reference Files
- `docs/proposals/servicetree-overhaul-p1p2/proposal.md` — 二期方案 6/9、SC5/SC6
- `D:/Haven/e-cam-service/internal/cam/servicetree/service/node_asset.go` — 既有节点资产查询（374 行）
- `D:/Haven/e-cam-service/internal/cam/servicetree/service/tree.go` — 子树遍历
- `D:/Haven/e-cam-service/internal/cam/servicetree/repository/dao/binding.go` — 绑定表结构

## Acceptance Criteria
- [ ] 新路由 `GET /nodes/:id/asset-summary`：返回 {total, byEnvironment{}, byProvider{}, byType{}, byBindType{}, suspicious[]}（suspicious 含 asset_id/asset_name/绑定环境/疑异原因）
- [ ] 聚合覆盖子树全部节点（含自身）；聚合查询走 group by（非 N+1 递归）
- [ ] 环境错绑检测：命名模式（`-prod-`/`-uat-`/`-test-`/`-dev-`）与 tag.env 双信号，矛盾即入 suspicious
- [ ] 单测：多级子树聚合正确/空子树/错绑检测命中与不命中；`go build` + `go test ./internal/cam/servicetree/...` 全绿
- [ ] `git diff` 仅 servicetree 模块内文件

## Hard Rules
仅改 `internal/cam/servicetree/` 内文件。**不碰** task/executor（任务 3）、前端、并发在途文件。
提交 `git -C D:/Haven/e-cam-service add <pathspec>` 显式路径。

## Implementation Notes
- 子树节点 ID 集合先查（tree.go 已有遍历），再对绑定表 `node_id IN (...)` 一次聚合。
- 资产明细字段（命名/tag.env）需关联实例集合——评估在聚合查询里 join 还是二段查（绑定→实例批量取），取后者简单。
- 疑异判定只提示不改绑（Out of Scope 明确）。
