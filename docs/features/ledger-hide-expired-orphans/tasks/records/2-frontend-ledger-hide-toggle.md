---
status: "completed"
started: "2026-09-21 11:09"
completed: "2026-09-21 12:06"
time_spent: "~57m"
---

# Task Record: 2 前端台账默认隐藏 + 行内提示切换 + 状态机

## Summary
前端台账默认隐藏无引用过期证书：消费任务1服务端 visibleTotal/hiddenCount 双口径（api/cert.ts 新增 includeExpiredNoRefs 参数与 hiddenCount 字段），ledger/index.vue 实现默认隐藏切换状态机（默认态/展开态切换、daysLeft=expired 筛选时挂起显示全部、清除筛选恢复、total 变化重置第 1 页），CertTable.vue 工具栏渲染 canonical 文案横幅（默认「已隐藏 N 张无引用过期证书 · 查看全部」/展开「再次点击隐藏无引用过期证书」），空态视图保留提示条与查看全部入口，format.ts 展示纯逻辑（filter/文案）+ vitest 覆盖。代码已提交 e-cam-web：6393454（主体）、fbf31aa、57e918a。本记录为补录（原执行 DONE 后执行记录缺失）。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/api/cert.ts
- e-cam-web/src/views/cert/ledger/index.vue
- e-cam-web/src/views/cert/ledger/components/CertTable.vue
- e-cam-web/src/views/cert/ledger/format.ts
- e-cam-web/src/views/cert/ledger/format.test.ts
- e-cam-web/src/views/cert/ledger/index.test.ts

### Key Decisions
- 隐藏逻辑采用服务端过滤 + 双总数口径（visibleTotal/hiddenCount），前端不自行过滤行数据，与 search/hostingStatus 正交
- 切换态仅作用默认视图：daysLeft=expired 时挂起显示全部且不显示隐藏提示，清除筛选后恢复
- 切换导致 total 变化时重置到第 1 页；空态下提示条不随表格主体消失
- 被隐藏行含可删除清理对象（如 *.jlc-robot.com），查看全部入口是其唯一发现路径，保留提示属性

## Test Results
- **Tests Executed**: Yes
- **Passed**: 54
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 默认视图使用服务端 visibleTotal/hiddenCount：不显示被排除的无引用过期证书，总数与行一致
- [x] 工具栏显示「已隐藏 N 张无引用过期证书 · 查看全部」/展开态「再次点击隐藏无引用过期证书」双向切换，计数与行实时一致
- [x] 切换导致 total 变化时重置到第 1 页
- [x] daysLeft=expired 筛选或 /certs?daysLeft=expired 跳转时切换态挂起，清除筛选后恢复
- [x] 空态视图仍显示隐藏提示条与「查看全部」入口
- [x] 前端 vitest 覆盖（format/filter + 切换交互 + 状态机）+ vue-tsc 通过

## Notes
补录记录：原执行会话报告 DONE 并提交代码（e-cam-web commits 6393454/fbf31aa/57e918a，已核实存在于 git log，cert.ts 含 includeExpiredNoRefs/hiddenCount），但执行记录缺失。本次为 verify-only 复核：任务作用域测试套件 vitest run src/views/cert/ledger/format.test.ts index.test.ts → 54/54 通过；原执行时全量 vitest 572/572 + vue-tsc 通过。复核时全仓 vue-tsc -b 报 3 个错误，全部位于 src/components/FilterBar/index.vue —— 该文件为另一并发会话的未跟踪脏文件（git status untracked），不属于本任务任何提交，本记录未触碰。
