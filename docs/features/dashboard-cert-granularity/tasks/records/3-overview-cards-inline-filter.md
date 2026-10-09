---
status: "completed"
started: "2026-09-22 13:06"
completed: "2026-09-22 13:22"
time_spent: "~16m"
---

# Task Record: 3 前端 OverviewCards 分级卡/特殊卡改页内筛选 + 豁免 + 移除跳台账

## Summary
到期看板 OverviewCards 分级卡/特殊卡改页内筛选 + 豁免 + 移除跳台账：分级卡点击改页内筛选（selected 高亮、再点取消、aria-pressed 两态），emit navigate-level→select-level，移除「点击跳转台账」静态文案与 router.push('/certs')（grep src/views/cert/dashboard/ 残留 0）；未激活副文案「可见 N · 隐藏 M」（countsByLevel 双口径 per-level hidden 分量），激活提示总计数；激活经任务 2 状态机（isRiskFilterActive + show-all-change）强制 includeHidden 重拉，主表行数 == 该档 countsByLevel 总计数；特殊卡 diff/exempt 副标题标注「域名计数」，filterDashboardItems 增 probeByDomain 可选参数实现逐 SAN 谓词（任一 SAN 为该状态即命中，exempt 被行聚合更差态掩盖仍可筛中；探测明细缺失降级行徽标相等），index.vue 挂载时拉取 GET /certs/probes 构建 SAN 映射；筛选状态 ⇄ URL query 深链同步（filterToQuery/filterFromQuery 纯函数 + index.vue watch router.replace 无历史噪音 + mount 恢复），filter 单一状态源机制沿用（与特殊卡 toggle 一致，Hard Rule 总规则豁免仅由风险维度触发未动）

## Changes

### Files Created
- e-cam-web/src/views/cert/dashboard/components/__tests__/OverviewCards.test.ts

### Files Modified
- e-cam-web/src/views/cert/dashboard/format.ts
- e-cam-web/src/views/cert/dashboard/format.test.ts
- e-cam-web/src/views/cert/dashboard/index.vue
- e-cam-web/src/views/cert/dashboard/components/OverviewCards.vue

### Key Decisions
- 逐 SAN 谓词数据源复用 GET /certs/probes（LatestPerDomain）按 SAN 域名映射（与抽屉多 SAN 视图同源），DashboardItem 载荷不含 per-SAN 状态故不改任务 1 契约；明细加载失败/未就绪降级行聚合徽标相等（diff 为最差态二者等价，exempt 仅明细缺失窗口内少筛不误筛）
- 分级卡激活豁免复用任务 2 既有状态机（resolveDashboardHiddenViewState + show-all-change 上抛 + includeHidden 重拉），本任务只补触发端 toggle，不另写豁免路径
- URL query 深链以纯函数 filterToQuery/filterFromQuery 落位 format.ts（可 node 单测、不引 vue-router 依赖；非法值丢弃），index.vue watch 中与当前 query 一致时跳过 replace 防导航重复
- aria-pressed 两态断言落位新组件测试 OverviewCards.test.ts（无 ElementPlus 组件依赖，纯 button 断言）
- commit 619726f 显式 pathspec 5 个文件，未卷入并发会话的 network/* 脏文件

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 分级卡点击页内筛选：选中高亮、再点取消恢复；grep src/views/cert/dashboard/ 无 router.push('/certs') 残留
- [x] 激活该档 → 豁免孤儿隐藏（总规则）→ 主表行数 == countsByLevel[tier].total（纯关系断言）；未激活显示「可见 N · 隐藏 M」
- [x] 特殊卡 diff/exempt：计数保持域名级（副标题「域名计数」），行集合 = 任一 SAN 为该状态（逐 SAN 谓词）；单测断言「一个 diff 域名挂 2 张证 → 卡 1 行 2」及 exempt 被更差态掩盖仍可筛中
- [x] 筛选状态同步 URL query 保留深链（弥合移除 router.push）
- [x] 分级卡选中/未选两态 aria-pressed 各有 vitest 断言；format.ts 测试随逐 SAN 谓词更新
- [x] vitest + vue-tsc 全绿

## Notes
门禁证据：vue-tsc -b 0 错误；prettier --check 通过；just lint 仅 src/api/types/*.ts 三个与本项目无关的既有 no-empty-object-type 错误（沿任务 2 记录，本任务改动文件 0 问题）；vitest 全仓 49 文件 902/902 通过（较任务 2 的 884 新增 18 用例：format.test.ts 逐 SAN 谓词 5 + levelCardSubtitle 3 + query round-trip 3 + OverviewCards.test.ts 7）。coverage 字段省略：@vitest/coverage-v8 未安装（npm 安装两度失败，沿 bdd9513/任务 2 先例）。台账页 /certs?daysLeft 深链接收逻辑保留未动（外部深链仍生效）。本任务不新增台账入口（eval 残留 #9：只读角色台账入口由菜单承担，联调验证归任务 4）
