---
id: "1"
title: "抽取公共环境推断 + 修正 tag 键 + 环境码对齐 + envRepo 注入"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: 抽取公共环境推断 + 修正 tag 键 + 环境码对齐 + envRepo 注入

## Description
**代码仓：D:/Haven/e-cam-service**。把 `node_asset.go` 里散落的 `matchNameEnvCode`/`normalizeEnvCode`/`isStandardEnvCode`/`extractTagEnv`/`loadEnvCodeMap` 收敛成单份公共环境推断能力，供绑定（任务 2）、改绑（任务 3）、概览三方复用（SC：单份实现的根）。同时修一个实测确认的 bug：`extractTagEnv` 读 `tags["env"]`，但数据标签键是 `tags["environment"]`（192 条）——重写为 `environment` 优先、`env` 兜底。环境码对齐：canonical `staging`（`-uat-` 归一化结果）与租户环境表 code `uat` 匹配；`fat` 等未知码视为无信号。`ruleEngineService` 注入 `envRepo` 以便 code→env_id 映射（当前仅 nodeAssetService 有）。

## Reference Files
- `docs/proposals/servicetree-env-inference/proposal.md` — Proposed Solution / Scope
- `internal/cam/servicetree/service/node_asset.go` — detectEnvMismatch/matchNameEnvCode/normalizeEnvCode/isStandardEnvCode/extractTagEnv/loadEnvCodeMap（509-560 行附近）
- `internal/cam/servicetree/service/rule_engine.go` — ruleEngineService 结构体（需加 envRepo）
- `internal/cam/servicetree/ioc/wire_gen.go` — NewRuleEngineService 注入点

## Acceptance Criteria
- [ ] 新增公共 `inferEnvCode(instance)`：命名模式 `-prod-/-uat-/-test-/-dev-` 优先，其次 `tag.environment`（兼容 `tag.env`），返回 canonical 码 `dev/test/staging/prod` 或空串
- [ ] tag 键修正：读 `tags["environment"]` 优先、`tags["env"]` 兜底（实测 environment 192 条而 env 0 条）
- [ ] 环境码对齐：infer 的 `staging` 与租户 code `uat` 经 `normalizeEnvCode` 归一后等价；`fat` 等非标准码归类为无信号
- [ ] `node_asset.go` 的 `detectEnvMismatch`/`extractTagEnv` 改用公共单份实现，删除/收敛重复函数
- [ ] `ruleEngineService` 注入 `envRepo`（`NewRuleEngineService` 签名 + `wire_gen.go` 同步），提供 code→env_id 映射
- [ ] 单测：命名命中/未命中、tag.environment 命中、uat↔staging 对齐、fat 无信号；`go build ./...` + `go test ./internal/cam/servicetree/...` 全绿

## Hard Rules
仅改 `internal/cam/servicetree/` 内文件（含 ioc）。不碰 task/executor、前端。提交 `git -C D:/Haven/e-cam-service add <显式 pathspec>`，绝不用 `git add -A`。

## Implementation Notes
- 公共方法建议放 `service/infer_env.go`（新建）或 rule_engine.go 顶部，命名避开 node_asset.go 既有符号名冲突。
- code→env_id 映射：`envRepo.List` 一次加载，按 `normalizeEnvCode(code)` 建反向 map，`inferEnvCode` 结果直接查。
- `NewRuleEngineService` 变更会牵动 wire 与既有测试的 `newTestRuleEngine` 桩——同步更新测试桩注入 envRepo。