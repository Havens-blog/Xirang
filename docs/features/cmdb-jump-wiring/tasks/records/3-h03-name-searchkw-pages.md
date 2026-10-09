---
status: "completed"
started: "2026-09-20 17:06"
completed: "2026-09-20 17:13"
time_spent: "~7m"
---

# Task Record: 3 H-03 接收方：name 系 2 页 + searchKeyword 系 2 页 + ecs

## Summary
H-03 接收方剩余 5 页接通 query.search：network/vpc、network/eip 首载前读 query.search 赋 filters.name；storage/nas、storage/oss 赋既有搜索框绑定 ref searchKeyword；compute/ecs 在既有 onMounted 的 query.uid 读取之后追加赋 filters.keyword（uid 逻辑未动，uid 精确/keyword 模糊不同维度共存）。统一三行模式同任务 2：import useRoute（缺失则补）→ onMounted 首次 fetch 前 const s = route.query.search + typeof s === 'string' && s 守卫赋值 → 依赖既有首载发出过滤请求。零二次全量请求；无 query 直达行为零改动。提交 807fae6，仅 5 个 index.vue，显式 pathspec。

## Changes

### Files Created
无

### Files Modified
- src/views/network/vpc/index.vue
- src/views/network/eip/index.vue
- src/views/storage/nas/index.vue
- src/views/storage/oss/index.vue
- src/views/compute/ecs/index.vue

### Key Decisions
- vpc/eip 已有 vue-router 导入，合并为 import { useRoute, useRouter }，不新增重复导入行
- nas/oss 原本无 vue-router 依赖，按任务 2 模式在 vue 导入后新增 import { useRoute } from 'vue-router'，并在 onMounted 前定义 route（注释同任务 2 措辞）
- nas/oss 单行 onMounted 展开为多行，searchKeyword.value 赋值置于 fetchData() 之前，保证首载即过滤（请求参 name: searchKeyword.value 已存在）
- ecs 的 search 读取插在现有 query.uid if/else 之后、fetchModels()/fetchInstances() 之前，uid 逻辑逐字未动
- 测试上报口径：任务相关单测（nasMetrics/ossMetrics）36 passed / 0 failed；全量 vitest 548 passed / 3 failed，3 个失败全部位于 src/views/cert/ledger/index.test.ts，为并发 cert 会话工作（commit 7cdd488 引入的 vue-router mock 缺 useRoute），本任务未触碰 cert，非本任务新增失败

## Test Results
- **Tests Executed**: Yes
- **Passed**: 36
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] vpc/eip：首载前读 query.search 赋 filters.name；nas/oss：赋 searchKeyword
- [x] ecs：query.search 赋 filters.keyword（保留既有 uid 读取逻辑不动）
- [x] 各页带 search 进入首屏即过滤结果；无 query 直达零改动
- [x] pnpm typecheck + vitest 通过；git diff 仅这 5 个 index.vue

## Notes
pnpm typecheck 仅剩 src/views/logs/ 2 个错误（并发 logs 会话预存，任务允许忽略）；本任务 5 文件 0 类型错误。vitest 全量 3 failed 均为 src/views/cert/ledger 预存并发失败（本任务文件与其无依赖关系，未新增���败）。git 提交 807fae6 经显式 pathspec 暂存，仅含 5 个 index.vue；并发脏文件（accounts/index.vue、components.d.ts、.env.development、element-theme.scss、.claude/）未触碰、未提交。发送方切换属任务 4，未动。
