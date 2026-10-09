---
status: "completed"
started: "2026-09-20 17:15"
completed: "2026-09-20 17:21"
time_spent: "~6m"
---

# Task Record: 4 H-03 发送方：handleViewDetail 改发 asset_name（接收方就绪后切换）

## Summary
H-03 发送方切换：cmdb/instances/index.vue handleViewDetail 的 query.search 从 row.asset_id 改为 row.asset_name || row.asset_id。后端各列表 name 模糊搜索匹配展示名 asset_name，云 ID 匹配不上；接收方 10 页（任务 2/3，commits f4fd016/807fae6）已就绪后切换，无发 name 而目标页不读的空窗。InstanceVO.asset_name 可空，空时兜底 asset_id。assetTypeRouteMap 10 个路由目标与未映射类型的 ElMessage.warning 分支行为不变。单行改动（+1/-1），commit 11d61f3。

## Changes

### Files Created
无

### Files Modified
- src/views/cmdb/instances/index.vue

### Key Decisions
- Hard Rules 仅允许改 src/views/cmdb/instances/index.vue，覆盖了 prompt 的 60% coverage 策略——该页无既有测试文件，新建测试文件会越出 Hard Rules 文件边界，故未新增测试；验证依赖全量 vitest + typecheck
- coverage 字段省略：仓库未装 @vitest/coverage-v8，vitest 不输出覆盖率指标，按『指标必须来自真实 runner 输出』规则不虚构数值
- 全量 vitest（未过滤）548 passed / 3 failed，3 个失败全部位于 src/views/cert/ledger/index.test.ts（该测试自己的 vue-router mock 缺 useRoute 导出），属并发会话在途工作，与本任务文件无关；记录用剔除该文件的 scoped run 真实计数 548/0
- pnpm typecheck 仅 2 个错误且全部位于 src/views/logs/ 测试文件（top_uris 类型/缺字段），属并发会话在途工作；cmdb/instances/index.vue 零错误
- prettier --check 对该文件报警告为存量问题（HEAD 版本同样不通过，含行尾/尾随空格），非本次单行改动引入，未做整文件重排以保持 surgical diff

## Test Results
- **Tests Executed**: Yes
- **Passed**: 548
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] query.search 值改为 row.asset_name || row.asset_id
- [x] 10 个路由目标与未映射类型的警告分支行为不变
- [x] pnpm typecheck + vitest 通过（并发在途错误除外）；git diff 仅 cmdb/instances/index.vue

## Notes
Scoped vitest run（排除并发在途失败的 cert/ledger 测试文件）: Test Files 27 passed (27), Tests 548 passed (548), 0 failed。全量 run 为 548/3，3 个失败均在 cert/ledger（并发会话）。typecheck (vue-tsc -b): 本文件 0 错误，仅 src/views/logs/ 两个存量错误。ESLint 对本文件 exit 0。提交: e-cam-web main 11d61f3，仅暂存 src/views/cmdb/instances/index.vue（未用 git add -A；.env.development/element-theme.scss/components.d.ts/accounts/index.vue 为并发会话脏文件，未触碰未提交）。
