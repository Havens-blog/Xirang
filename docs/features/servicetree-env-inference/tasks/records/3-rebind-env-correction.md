---
status: "completed"
started: "2026-09-22 14:27"
completed: "2026-09-22 14:27"
time_spent: ""
---

# Task Record: 3 改绑重算纳入环境，纠正历史 691 条环境错标

## Summary
改绑重算纳入环境，纠正历史环境错标。PreviewRebind/ApplyRebind 的 rule 绑定目标 env 改用 resolveEnvID(inferInstanceEnvCode(inst), cur.EnvID, envIDByCode) 重算（推断优先、未命中沿用当前绑定 env 而非规则写死值），候选判定从「节点不同」扩为「节点或环境不同」。嘉立创租户活体验证：改绑预览 723 条全部「仅环境」候选（664 dev→prod、59 dev→uat），apply 后根子树 suspicious 723→0、环境分布 dev 1062→339/prod 1→665/uat 0→59，重算预览 0 候选收敛。

## Changes

### Files Created
无

### Files Modified
- internal/cam/servicetree/service/rebind_plan.go
- internal/cam/servicetree/service/rebind_plan_test.go

### Key Decisions
无

## Test Results
- **Tests Executed**: Yes
- **Passed**: 133
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] computeRebindCandidates 目标 env 用推断结果（非规则写死），候选判定含环境变化
- [x] 「仅环境变化、节点不变」的 rule 绑定也产生候选（当前被漏）
- [x] ApplyRebind 用推断环境调用 UpdateTarget，环境纠正落库
- [x] manual 锁定不变：手动绑定资产不入候选
- [x] 单测：仅环境变化 candidate 命中、确认后 env 改对、manual 跳过；go build + go test 全绿
- [x] 运行时佐证：改绑后根子树 suspicious 显著下降（无信号资产除外）

## Notes
AC6 由主会话完成（子执行器错判数据不可达）：本地 8001 以 X-Tenant-ID:3 + admin Bearer 直接 curl 即命中 1063 条绑定数据，子执行器的「全为 0」源于其租户切换未对齐 jlc。改绑为「预览→确认」两步（预览 723 全为正确方向的环境纠偏后才 apply）。resolveEnvID 兜底参数传 cur.EnvID 落实「未命中沿用当前绑定 env」；envIDByCode 加载与 ExecuteRules 同口径降级；未改 rule_engine.go/node_asset.go/infer_env.go/domain VO。
