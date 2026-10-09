---
id: "3"
title: "改绑重算纳入环境，纠正历史 691 条环境错标"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies:
  - "1"
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 3: 改绑重算纳入环境，纠正历史 691 条环境错标

## Description
**代码仓：D:/Haven/e-cam-service**。`PreviewRebind`/`ApplyRebind`（`rebind_plan.go`，三期 282e256）当前重算口径只有节点——目标 env 用规则写死的 env。改为：rule 绑定资产的目标环境也用 `resolveEnvID(inferEnvCode(instance), …)` 计算，候选判定从「节点不同」扩为「**节点或环境不同**」。这样历史 691 条已错绑（prod 命名绑 dev）的资产在改绑预览里以「环境变化」候选出现，经既有「规则改绑」预览→确认入口一次性纠正。`UpdateTarget(id, nodeID, envID, ruleID)` 原语已带 env 参数，无需改仓储层。

## Reference Files
- `docs/proposals/servicetree-env-inference/proposal.md` — Proposed Solution / Success Criteria / Key Risks
- `internal/cam/servicetree/service/rebind_plan.go` — computeRebindCandidates/ApplyRebind
- 任务 1 产物：公共 `inferEnvCode` / `resolveEnvID`

## Acceptance Criteria
- [ ] `computeRebindCandidates` 目标 env 用推断结果（非规则写死），候选判定含环境变化
- [ ] 「仅环境变化、节点不变」的 rule 绑定也产生候选（当前被漏）
- [ ] `ApplyRebind` 用推断环境调用 `UpdateTarget`，环境纠正落库
- [ ] manual 锁定不变：手动绑定资产不入候选
- [ ] 单测：仅环境变化的 candidate 命中、确认后 env 改对、manual 跳过；`go build ./...` + `go test ./internal/cam/servicetree/...` 全绿
- [ ] 运行时佐证：改绑执行后根子树 `GET /nodes/4/asset-summary` 的 `suspicious` 数量较改绑前显著下降（无信号资产除外）

## Hard Rules
仅改 `internal/cam/servicetree/service/rebind_plan.go`（+ 测试文件）。不碰 rule_engine.go（任务 2）、node_asset.go（任务 1）。提交显式 pathspec。

## Implementation Notes
- `RebindCandidate` 已有 `FromEnvID/ToEnvID` 字段，直接填推断结果即可，无需改 VO。
- 目标 env 推断链：`inferEnvCode(instance)` → `normalizeEnvCode` → env_id；未命中沿用当前绑定 env（而非规则 env，避免误改无信号资产）。
- 前端预览弹窗已渲染 `from→to` 节点；如要更直观可补 env 显示，但本任务不强制（Out of Scope 前端）。