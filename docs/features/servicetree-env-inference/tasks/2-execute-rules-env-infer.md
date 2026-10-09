---
id: "2"
title: "ExecuteRules 绑定环境按推断优先、规则 env_id 兜底"
priority: "P0"
estimated_time: "1.5h"
complexity: "medium"
dependencies:
  - "1"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 2: ExecuteRules 绑定环境按推断优先、规则 env_id 兜底

## Description
**代码仓：D:/Haven/e-cam-service**。`ExecuteRules` 当前创建 `ResourceBinding` 时 `EnvID` 直接用 `rule.EnvID`。改为：`EnvID = resolveEnvID(inferEnvCode(instance), rule.EnvID)`——推断命中时用资产真实环境映射到的 env_id，未命中沿用规则 env_id。这是纠正 691 条错标的核心（新绑/未来同步的资产直接落对环境）。**不改变** create-only 幂等、manual 锁定、优先级首个命中语义。

## Reference Files
- `docs/proposals/servicetree-env-inference/proposal.md` — Proposed Solution / Success Criteria
- `internal/cam/servicetree/service/rule_engine.go` — ExecuteRules（事件 4 落库循环，~207-230 行）
- 任务 1 产物：公共 `inferEnvCode` / `resolveEnvID`

## Acceptance Criteria
- [ ] 绑定落库 `EnvID`：推断命中 → 资产环境码对应 env_id；未命中 → `rule.EnvID` 兜底
- [ ] create-only 幂等不变：已绑定资源（manual/rule）仍整体跳过，重复执行零重复绑定
- [ ] manual 锁定不变：手动绑定资产不被规则覆盖
- [ ] 优先级语义不变：仍取首个命中的规则（节点按规则，仅环境被资产纠正）
- [ ] 单测：命名 `-prod-` 命中落 prod env、未命中落规则 env、manual 跳过；`go build ./...` + `go test ./internal/cam/servicetree/...` 全绿

## Hard Rules
仅改 `internal/cam/servicetree/service/rule_engine.go`（+ 测试文件）。不碰 node_asset.go（任务 1）、rebind_plan.go（任务 3）。提交显式 pathspec。

## Implementation Notes
- 复用任务 1 的 `resolveEnvID` 与 `ruleEngineService.envRepo`；映射在 `ExecuteRules` 开篇一次加载，循环内查表。
- 注意与任务 3 的文件边界：本任务只动 `rule_engine.go` 的 ExecuteRules，任务 3 动 `rebind_plan.go`。