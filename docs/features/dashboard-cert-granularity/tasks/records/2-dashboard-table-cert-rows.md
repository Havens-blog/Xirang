---
status: "completed"
started: "2026-09-22 12:43"
completed: "2026-09-22 13:04"
time_spent: "~21m"
---

# Task Record: 2 前端 DashboardTable 证书行渲染 + 隐藏切换 + 云选项派生 + 抽屉多 SAN 视图

## Summary
前端 DashboardTable 证书粒度化：cert.ts 类型对齐任务 1 载荷（证书行 commonName/sans/issuer/referenceStatus/lastScanAt/hidden、countsByLevel 升级 {total,visible,hidden} + hiddenCount、getCertDashboardApi 增 includeHidden 参数）；DashboardTable row-key 改 certId（同域两证不串行）、列改证书行渲染（CN+SAN 折叠计数/引用云/剩余天数/托管/签发者/探测徽标/豁免）、工具栏「查看全部」横幅开关（localStorage cert.dashboard.hiddenExpanded 持久化、风险维度卡激活强制开启并禁用、清除恢复原态）、云筛选选项由可见行派生（hidden 行不入选项）、空态保留「查看全部」入口、show-all-change 上抛父级以 includeHidden 重拉（豁免接线，前端不派生三态）；ProbeDetailDrawer 扩展证书多 SAN 视图（头部证书摘录 + 全部 SAN 逐条探测，最差优先 stable 排序，SAN>20 折叠最差优先前 20 可展开）；OverviewCards countsByLevel 双口径卡值消费修正（取 total，页内筛选交互为任务 3 范围）

## Changes

### Files Created
- e-cam-web/src/views/cert/dashboard/components/__tests__/DashboardTable.test.ts

### Files Modified
- e-cam-web/src/api/cert.ts
- e-cam-web/src/views/cert/dashboard/format.ts
- e-cam-web/src/views/cert/dashboard/format.test.ts
- e-cam-web/src/views/cert/dashboard/index.vue
- e-cam-web/src/views/cert/dashboard/components/DashboardTable.vue
- e-cam-web/src/views/cert/dashboard/components/OverviewCards.vue
- e-cam-web/src/views/cert/dashboard/components/ProbeDetailDrawer.vue

### Key Decisions
- 隐藏三态完全消费服务端结果（row.hidden + summary.hiddenCount），前端零派生（Hard Rule）
- 云选项派生保留在 format.ts cloudFilterOptions（跳过 hidden 行）并由 DashboardTable 从服务端行全集计算（不过滤后行集，避免筛选态选项收缩），组件经 defineExpose 暴露供锁定
- 开关状态机落位 dashboard/format.ts（resolveDashboardHiddenViewState + isRiskFilterActive），与台账 resolveHiddenViewState 同构但豁免条件不同（看板=level/special 任一，台账=daysLeft=expired 挂起）
- 抽屉多 SAN 数据源复用 GET /certs/probes（LatestPerDomain）按 SAN 域名映射，打开时拉取、失败降级「未探测」并提示，不塌陷聚合徽标
- index.vue 仅做豁免接线（includeHidden 初始取持久化态 + show-all-change 变化重拉），OverviewCards 页内筛选/router.push 移除留给任务 3
- commit 421bd45：显式 pathspec 提交本任务 8 个文件，未卷入并发会话的 network/* 脏文件

## Test Results
- **Tests Executed**: Yes
- **Passed**: 884
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 证书行渲染：CN+SAN 折叠计数+issuer+daysLeft+托管+引用云+探测徽标；row-key 改证书 ID（组件测试断言同域两证不串行）
- [x] 孤儿默认隐藏 + 工具栏横幅开关（localStorage 持久化、刷新保留）+ 空态「查看全部」入口保留
- [x] 云筛选选项由当前可见行派生，隐藏态下不含孤儿所属云（组件测试锁定）
- [x] 豁免接线：风险维度卡激活或开关开启 → includeHidden=true 重新拉取全量；前端不派生三态
- [x] 行点击 → 证书多 SAN 探测视图（头部证书摘录 + 全部 SAN 探测列表，SAN>20 折叠为最差优先前 20）
- [x] 组件测试：开关状态机（分级卡激活强制开启并禁用、清除恢复原态）、localStorage 持久化、隐藏态云选项；vitest + vue-tsc 全绿

## Notes
门禁证据：just compile（vue-tsc -b）0 错误；just fmt（prettier --check）通过；just lint 仅 src/api/types/*.ts 三个与本任务无关的既有 no-empty-object-type 错误（本任务改动文件 0 问题）；vitest 全仓 48 文件 884/884 通过（含新增 DashboardTable.test.ts 11 用例 + format.test.ts 新增状态机/持久化/最差优先/云选项用例）。coverage 字段省略：@vitest/coverage-v8 未安装（npm 安装两度失败，沿 bdd9513 先例）。el-table 在 happy-dom 下 MutationObserver 跨 realm 崩溃，测试沿 assets 先例 stub NoopMutationObserver。OverviewCards 卡值改为 countsByLevel[idx].total 属任务 1 载荷的类型消费必需（否则渲染 [object Object]），卡交互改造（页内筛选/aria-pressed/移除跳台账）未动，归任务 3。
