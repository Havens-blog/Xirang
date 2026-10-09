---
status: "completed"
started: "2026-09-20 17:31"
completed: "2026-09-20 17:34"
time_spent: "~3m"
---

# Task Record: fix-1 fix unit-test: just unit-test failure in quality gate

## Summary
修复质量门 just unit-test 失败：cert/ledger/index.vue setup 现调用 useRoute()（cert 侧并发新增），而 ledger 测试的 vue-router mock 仅提供 useRouter，致 3 个挂载类用例报 No "useRoute" export is defined 全挂并阻塞共享门。测试 mock 补 useRoute: () => ({ query: {}, params: {} })（测试基建一行修复，不动页面逻辑）。vitest 551/551 全绿。

## Changes

### Files Created
无

### Files Modified
- src/views/cert/ledger/index.test.ts

### Key Decisions
- 仅补测试 mock（+2 行注释），不改 cert 页面逻辑——页面 useRoute 使用属并发会话 cert 工作的正确改动
- mock 返回空 query/params，与既有用例断言无耦合

## Test Results
- **Tests Executed**: Yes
- **Passed**: 551
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 3 个 cert/ledger 挂载类用例恢复通过
- [x] vitest 全量 551/551 绿，质量门解锁

## Notes
提交 a459473（显式 pathspec 单文件）。失败源为并发会话 cert 在途改动（页面加 useRoute 但未同步测试 mock），本修复为测试基建最小补齐。
