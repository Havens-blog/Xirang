---
status: "completed"
started: "2026-09-19 16:02"
completed: "2026-09-19 16:16"
time_spent: "~14m"
---

# Task Record: 1 拓扑页移除 mock 回退，改诚实错误态

## Summary
拓扑页移除全部 mock 回退，改为诚实错误态：useTopologyData.ts 两个 catch 块不再回退 MOCK_DOMAINS/MOCK_GRAPH，改为记录错误状态并置空数据；stores/topology.ts 新增 dataStatus(idle/loading/live/error) 与 loadError 字段；views/topology/index.vue 徽章随数据状态联动（成功=绿实时、失败=橙异常、加载中/未加载=灰）并新增 el-alert 错误提示指引刷新重试；删除 useMockData.ts；顺带清理 config/index.ts 无消费者的 useMock 死开关。新增 useTopologyData.test.ts 7 例覆盖错误态/空态/加载中/重试恢复。

## Changes

### Files Created
- e-cam-web/src/views/topology/composables/useTopologyData.test.ts

### Files Modified
- e-cam-web/src/views/topology/composables/useTopologyData.ts
- e-cam-web/src/stores/topology.ts
- e-cam-web/src/views/topology/index.vue
- e-cam-web/src/config/index.ts
- e-cam-web/src/views/topology/composables/useMockData.ts

### Key Decisions
- store 同时新增 dataStatus 枚举与 loadError 文案：徽章用枚举、el-alert 用文案，二者互补（任务 Implementation Notes 推荐二选一，取并集以同时满足徽章联动与错误提示）
- dataStatus 仅由 loadTopology 驱动（loading 起点→live/error 结果）；loadDomains 失败只写 loadError 并置空 domainList，不改徽章状态——初始仅加载域名列表时徽章保持灰「未加载」，不误显实时
- loadTopology 失败时 graphData 置为 null：TopologyCanvas 对 null/空 nodes 已有空态文案（选择域名前/后两套），组件零改动即满足「空画布」
- 后端成功返回空数组 → dataStatus='live'：空数据与失败严格区分，不误报错误
- loadTopology 起点清空 loadError：刷新/筛选重载时错误状态随最新一次结果更新
- config/index.ts 的 useMock 死开关已清理（grep 确认 src 内无消费者）；.env.development/.env.production/env.d.ts/README 中的 VITE_USE_MOCK 残留声明按 Hard Rules 文件边界不触碰（.env.development 为并发会话脏文件）
- 提交用显式 pathspec（6 个路径，含删除与新增测试），未触碰并发会话脏文件 .env.development/element-theme.scss/components.d.ts/accounts/index.vue

## Test Results
- **Tests Executed**: Yes
- **Passed**: 458
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] useTopologyData.ts 无 MOCK_DOMAINS/MOCK_GRAPH 引用；useMockData.ts 已删除（grep 断言全前端无残留）
- [x] 后端 API 失败 → el-alert 错误提示 + 空画布；徽章不再显示绿色「实时」（灰/橙「离线/异常」）
- [x] 后端成功 → 绿「实时」徽章 + 真实拓扑（行为与现状一致）；后端空数据 → 正常空态（不误报错误）
- [x] 刷新/筛选重新加载时错误状态正确更新；loading 中徽章不误显「实时」
- [x] pnpm typecheck + 相关 vitest 通过；git diff 仅拓扑 4 文件（+ 测试/config/index.ts 可选）

## Notes
验证证据：pnpm typecheck（vue-tsc -b）0 错误；pnpm vitest run 全量 21 文件 458/458 通过（既有 451 + 新增 7）；grep 断言 src 内 MOCK_DOMAINS/MOCK_GRAPH/useMockData/useMock 零残留。提交 8cedfbd（e-cam-web main，6 files changed, +199/-106），显式 pathspec 提交，并发会话脏文件均未暂存。徽章文案：未加载/加载中=灰、实时=绿(呼吸动画)、异常=橙(附 el-alert type=error 不可关闭，随下次加载清除)。
