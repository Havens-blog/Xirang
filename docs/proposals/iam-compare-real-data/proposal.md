---
created: "2026-09-19"
author: "Haven"
status: Approved
intent: "fix"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: IAM compare 页接真实数据（移除整页 mock）

## Problem

`e-cam-web/src/views/iam/compare/index.vue` 整页是**伪造数据**：用户/组/策略计数用 `Math.random()` 生成、账号存在性与成员数随机、对比行（admin/developer/operator…）虚构、还有 1 秒假延迟后弹「加载成功」。用户在对比视图看到随机数误以为是真实跨账号 IAM 数据——**CRITICAL 数据完整性缺陷**。

### Evidence

- `compare/index.vue:375-377`：`userCount/groupCount/policyCount` 全 `Math.random()`；`:402/416/438` 账号存在性/成员数随机。
- `:340` `// 模拟数据加载` + `setTimeout(resolve, 1000)` 假延迟，随后 `ElMessage.success('对比数据加载成功')`。
- `generateUserCompareData`（369）虚构 `['admin','developer','operator','viewer','auditor']` 用户名与随机存在性。
- 真实数据源已存在：`e-cam-web/src/api/iam.ts:52` `listUsersApi`（`/iam/users`，支持 `cloud_account_id` 按账号）、`:127` `listGroupsApi`（同）；后端 `internal/cam/iam/web/{user,group}_handler.go` 均按 `cloud_account_id` 查询。
- 策略：无 per-account 策略 list 客户端（仅 `listTemplatesApi` 模板），策略段需诚实降级或模板对比。

### Urgency

- **CRITICAL**：向用户展示虚构数据，可能导致基于错误计数的决策。用户已确认修复范围（接真实数据，前端）。

## Proposed Solution

`compare/index.vue` 的 mock 生成全部替换为真实 IAM 数据：
- 对每个选中账号（≤4）调 `listUsersApi`/`listGroupsApi`（带 `cloud_account_id`），从真实返回计数与对比行（用户/组名、各账号存在性）。
- 移除全部 `Math.random()` 与假延迟；真实 loading/错误/空态（无数据账号显示「—」/空，失败显示错误提示而非假成功）。
- 策略段：用真实可用数据（`listTemplatesApi` 模板对比）或显式「策略对比未接入」空态，绝不虚构。
- 仅前端改动，不新增后端接口。

### Innovation Highlights

诚实数据 > 优雅占位。价值在把误导性 mock 视图变为可信工具；无创新实现。

## Requirements Analysis

### Key Scenarios

- 选中 2-4 个账号：并行拉各账号 users/groups，计数与行均来自真实响应。
- 某账号 IAM 未配置/查询失败：该账号对应段显示「—」/错误提示，不阻塞其它账号。
- 账号无用户/组：空列表 → 计数 0，行表空态。
- 策略段无接口：显示「未接入」占位（诚实），不虚构。

### Non-Functional Requirements

- 无任何 `Math.random()`/假数据残留（grep 断言）。
- 页面加载即真实查询（移除假延迟）。

### Constraints & Dependencies

- 仅改 `e-cam-web/src/views/iam/compare/index.vue`（+ 必要时 `api/iam.ts` 补 per-account 策略查询客户端，若后端支持）。
- 不动 `accounts/index.vue`（并发会话脏文件）、logs 相关文件。
- 后端 IAM 接口不变。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 接真实数据（users/groups）+ 策略诚实占位 | 本提案 | 可信对比、修 CRITICAL、仅前端 | 策略段暂无真实对比 | Adopt（用户已确认） |
| 仅删随机数 + 占位 | 备选 | 最小改动 | 页面失去对比价值 | Rejected：用户已选真数据 |
| 后端新增每账号统计接口 | 备选 | 一次查询高效 | 后端+前端、需评估 | Deferred：当前 list 接口够用 |

## Scope

### In Scope

- `compare/index.vue`：mock 生成函数（loadCompareData/generate*）全部替换为真实 `listUsersApi`/`listGroupsApi` 调用（每账号，带 `cloud_account_id`）。
- 真实 loading/错误/空态；移除假延迟与假成功消息。
- 策略段：`listTemplatesApi` 真实模板对比或「未接入」占位（依实现可行性与产品倾向，倾向占位）。
- 类型调整（`UserCompareRow` 等结构对齐真实响应）。

### Out of Scope

- 后端新增统计接口、IAM 数据同步逻辑、其它 mock 页面（拓扑 mock 回退另行）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| per-account list 接口分页/量大（每账号拉全量用户） | Medium | 用 count/分页参数控制（ListUsersParams 已有）；只取计数与首屏行 |
| 某账号 IAM 未接入导致查询报错 | Medium | 逐账号 try/catch，单账号失败降级「—」不阻塞 |
| 策略段误用模板数据当 per-account 策略 | Low | 模板对比需明确标注「策略模板」，或直接空态 |

## Success Criteria

- [ ] `compare/index.vue` 无 `Math.random` 残留（grep 断言）；`// 模拟数据加载` 与假延迟移除
- [ ] 用户/组计数与对比行来自真实 `listUsersApi`/`listGroupsApi` 响应（每选中账号）
- [ ] 单账号查询失败/无数据 → 该账号显示「—」/空态，不阻塞其它；移除假「加载成功」
- [ ] 策略段无虚构数据（真实模板对比或「未接入」占位）
- [ ] `pnpm typecheck` + 相关 vitest 通过；`git diff` 仅 compare/index.vue（+必要时 api/iam.ts）