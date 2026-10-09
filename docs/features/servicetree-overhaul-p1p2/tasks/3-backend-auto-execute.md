---
id: "3"
title: "后端：同步后自动执行规则（异步挂点 + 幂等）"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.feature"
mainSession: false
---

# 3: 后端：同步后自动执行规则（异步挂点 + 幂等）

## Description
**代码仓：D:/Haven/e-cam-service**。资产同步完成后自动执行规则引擎（事件驱动，规则自动为主的核心）：
在资产同步链路（`internal/cam/task/executor`）完成后异步触发 `ExecuteRules`（goroutine + 超时 + 失败容忍，
**不阻塞同步主链路**——对齐 CDN 指标采集教训）。幂等：重复执行不产生重复绑定（instance+node+env 去重，
确认既有唯一键或补齐）；手动绑定优先（规则跳过已手动绑定资产——确认既有行为并补单测锁定）。

## Reference Files
- `docs/proposals/servicetree-overhaul-p1p2/proposal.md` — 一期方案 4/5、SC4、NFR 幂等
- `D:/Haven/e-cam-service/internal/cam/task/executor/` — 同步链路（挂点位置：sync_assets 完成后）
- `D:/Haven/e-cam-service/internal/cam/servicetree/service/rule_engine.go:126` — ExecuteRules
- `D:/Haven/e-cam-service/internal/cam/servicetree/repository/dao/binding.go` — 绑定唯一键确认

## Acceptance Criteria
- [ ] 资产同步完成后异步触发 ExecuteRules（不阻塞同步返回；panic recover + 超时 + 失败仅日志）
- [ ] 幂等单测：同一资产重复执行零重复绑定
- [ ] 手动绑定优先单测：已手动绑定资产规则执行跳过（locked 语义）
- [ ] 触发点注入方式遵循仓库既有 wire 模式（不引入全局单例）
- [ ] `go build ./...` + `go test ./internal/cam/servicetree/... ./internal/cam/task/...` 全绿
- [ ] `git diff` 仅 servicetree + task/executor 挂点最小文件

## Hard Rules
改 `internal/cam/servicetree/` + `internal/cam/task/executor/` 挂点最小文件。
**不碰** sync_registry/sync_assets 的既有逻辑语义（只加完成后回调）、并发在途文件、前端。
提交 `git -C D:/Haven/e-cam-service add <pathspec>` 显式路径。

## Implementation Notes
- 挂点形态参考：executor 完成后调 `servicetree.RuleService.ExecuteRulesAsync(tenantID)`——注意循环依赖，
  经 wire 注入接口或事件回调注册（module.go 组装），勿直接 import 造成环。
- CDN 指标教训：后台任务失败只 log，不向上传播；任务名/耗时打日志便于观测。
- 执行范围：仅 Enabled 规则（确认 ExecuteRules 现有过滤）。
