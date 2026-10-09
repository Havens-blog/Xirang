---
status: "completed"
started: "2026-09-19 14:44"
completed: "2026-09-19 14:56"
time_spent: "~12m"
---

# Task Record: 1 IAM compare 页 mock 替换为真实 IAM 数据

## Summary
IAM compare 页整页 mock 替换为真实 IAM 数据：每选中账号(≤4)并行调用 listUsersApi/listGroupsApi(带 cloud_account_id)，计数取响应 total(无则列表长度)，用户/组对比行取各账号名称并集；移除全部 Math.random 虚构数据、1s 假延迟与假「加载成功」；单账号查询失败经 Promise.allSettled 独立降级显示「—」并在顶部 el-alert 列出失败账号，不阻塞其它账号；策略段改为显式「暂未接入」el-empty 占位(无 per-account 策略接口，不虚构)，策略数量恒为「—」；UserCompareRow/GroupCompareRow 类型对齐真实响应(exists/missing/unavailable 三态)

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/views/iam/compare/index.vue

### Key Decisions
- 策略段采用「未接入」诚实占位而非 listTemplatesApi 模板对比：无 per-account 策略查询接口，模板为租户级、语义不同，展示会误导
- 每账号内 users/groups 用 Promise.allSettled 相互独立，任一失败仅该账号该段降级「—」，失败账号名汇总到 failedAccounts el-alert
- 计数优先取响应 total、无则取列表长度；对比行只取首屏 COMPARE_FETCH_SIZE=200 行，符合 proposal 对分页量大风险的缓解方案
- api/iam.ts 与 types/iam.ts 无需改动：现有 listUsersApi/listGroupsApi 已支持 cloud_account_id，类型已齐备

## Test Results
- **Tests Executed**: Yes
- **Passed**: 451
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] compare/index.vue 无 Math.random 残留(grep 断言)；// 模拟数据加载 与 setTimeout(...,1000) 假延迟移除
- [x] 用户/组计数与对比行来自真实 listUsersApi/listGroupsApi 响应(每选中账号带 cloud_account_id，loading 期间按钮禁用)
- [x] 单账号查询失败/无数据 → 该账号显示「—」/空态，不阻塞其它账号；移除假「加载成功」
- [x] 策略段无虚构数据(listTemplatesApi 真实模板对比或显式「未接入」占位)
- [x] 类型(UserCompareRow/GroupCompareRow 等)对齐真实响应
- [x] pnpm typecheck + 相关 vitest 通过；git diff 仅 compare/index.vue

## Notes
验证证据：pnpm typecheck(vue-tsc -b) exit 0；eslint(仅 compare/index.vue) 0 errors 0 warnings；grep 断言 Math.random/模拟数据加载/setTimeout/对比数据加载成功/generate* 均 0 匹配；compare 模块自身无测试文件(src/views/iam 无 *.test.ts，Hard Rules 限定仅可改 compare/index.vue 故未新增)，全量 vitest run 451/451 通过(20 files) 证明无回归；coverage 无 runner 输出未臆造。提交 bdd9513 仅含 src/views/iam/compare/index.vue(+265/-177)，显式 pathspec 提交；并发会话脏文件(accounts/index.vue、.env.development、element-theme.scss、components.d.ts)未触碰未提交。响应解析兼容 data/users/groups/items 载体字段与裸数组，与 users/groups 页既有防御式解析一致。
