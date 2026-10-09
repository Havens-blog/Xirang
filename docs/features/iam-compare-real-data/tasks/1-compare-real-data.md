---
id: "1"
title: "IAM compare 页 mock 替换为真实 IAM 数据"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.fix"
mainSession: false
---

# 1: IAM compare 页 mock 替换为真实 IAM 数据

## Description
**代码仓：D:/Haven/e-cam-web**（非 e-cam-service）。`src/views/iam/compare/index.vue` 整页 mock
（Math.random 计数、虚构用户/组/策略、1s 假延迟 + 假「加载成功」）替换为真实 IAM 数据：
每选中账号（≤4）调 `listUsersApi`/`listGroupsApi`（带 `cloud_account_id`）拉真实计数与对比行；
移除全部 Math.random 与假延迟；真实 loading/错误/空态（单账号失败降级「—」不阻塞）；策略段用
`listTemplatesApi` 真实模板或「未接入」占位（绝不虚构）。

## Reference Files
- `docs/proposals/iam-compare-real-data/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `D:/Haven/e-cam-web/src/views/iam/compare/index.vue` — 全页 mock（loadCompareData 340/生成函数 369+）
- `D:/Haven/e-cam-web/src/api/iam.ts` — listUsersApi(52)/listGroupsApi(127)/listTemplatesApi(210)，均支持 cloud_account_id
- `D:/Haven/e-cam-web/src/api/types/iam.ts` — ListUsersParams/ListGroupsParams/响应类型

## Acceptance Criteria
- [ ] `compare/index.vue` 无 `Math.random` 残留（grep 断言）；`// 模拟数据加载` 与 `setTimeout(..., 1000)` 假延迟移除
- [ ] 用户/组计数与对比行来自真实 `listUsersApi`/`listGroupsApi` 响应（每选中账号，带 `cloud_account_id`，loading 期间按钮禁用）
- [ ] 单账号查询失败/无数据 → 该账号显示「—」/空态，不阻塞其它账号；移除假「加载成功」
- [ ] 策略段无虚构数据（`listTemplatesApi` 真实模板对比，或显式「未接入」占位）
- [ ] 类型（UserCompareRow/GroupCompareRow 等）对齐真实响应
- [ ] `pnpm typecheck` + 相关 vitest 通过；`git diff` 仅 compare/index.vue（+ 必要时 api/iam.ts 补客户端）

## Hard Rules
仅修改 `D:/Haven/e-cam-web/src/views/iam/compare/index.vue`（+ 必要时 `src/api/iam.ts`/`types/iam.ts`）。
**不触碰** `src/views/accounts/index.vue`（并发会话脏文件）、`src/views/logs/`、后端任何文件。提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- 并行拉取：选中账号 × 用户/组 可用 `Promise.all`（≤4 账号 × 2 = ≤8 请求）；每账号独立 try/catch 降级。
- 计数从响应取：优先 total 字段，无则 len(list)。
- 策略段倾向「未接入」占位（无 per-account 策略接口，模板语义不同，需明确标注才可展示）。
- 类型核对 `src/api/types/iam.ts` 现有 ListUsersResponse 形状，必要时扩展。
