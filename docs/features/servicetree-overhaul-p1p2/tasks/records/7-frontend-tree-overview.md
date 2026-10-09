---
status: "completed"
started: "2026-09-21 19:31"
completed: "2026-09-21 19:31"
time_spent: ""
---

# Task Record: 7 前端：树徽标 + 关联资产聚合语义 + 环境疑异标红

## Summary
服务树总览二期可见性：①树节点子树资产数徽标（消费任务6 GET /nodes/:id/asset-summary，前端5并发批量拉取，非0高亮/0灰显）；②节点详情卡新增「子树资产」数与徽标同源一致，关联资产tab统计卡同源（total/by_type）；③关联资产tab聚合语义（提案#8）：include_children默认开启，根/中间节点展示子树聚合资产，新增「来源节点」列（node_id→节点名前端映射，NodeAssetVO已含node_id无需后端改动）；④环境疑异标红（提案#9）：选中节点summary.suspicious按asset_id匹配绑定/关联资产两表，行背景标红+名称tooltip疑异原因+醒目提示条，只提示不改绑。新增单测7例覆盖AC1-AC3。

## Changes

### Files Created
- src/views/service-tree/__tests__/index.test.ts

### Files Modified
- src/views/service-tree/index.vue
- src/api/service-tree.ts
- src/api/types/service-tree.ts

### Key Decisions
- 徽标数据源决策：任务6为单节点接口（GET /nodes/:id/asset-summary），未要求后端加批量参数。树使用 default-expand-all，展开事件懒加载不会触发，故选择前端小并发（5）批量拉取：服务树常态几十~百级节点，单点为轻量聚合查询（路径前缀+绑定表IN一次归并），5并发可控且单点失败容忍不阻塞。如未来节点规模上千可再评估后端批量接口。
- 「来源节点」列零后端改动：后端 NodeAssetVO 已返回 node_id（资产实际绑定节点），前端用 flat 节点列表映射 node_id→name 展示，无需扩展接口。
- 关联资产统计卡数据源从 getNodeAssetStatsApi（by_asset_type）切换为任务6的 getNodeAssetSummaryApi（by_type），与树徽标/详情卡同源口径，保证「点击节点详情与徽标一致」。
- 关联资产tab include_children 默认 true：根/中间节点展示子树聚合；叶子节点含子树查询等价于直属，满足聚合语义且无需按节点层级分支。原有手动开关保留可切回直属视图。
- 疑异匹配按云资产 asset_id：summary.suspicious 为选中节点子树超集，绑定tab（直属）与关联资产tab均用其匹配标红，无漏报可能有余报（子树内其他节点的疑异资产在父节点视图也标红，属预期提示行为）。CMDB instance 类绑定无云 asset_id 不参与匹配（疑异检测本身基于资产）。

## Test Results
- **Tests Executed**: Yes
- **Passed**: 837
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 树节点显示子树资产数徽标（非0数字标签，0灰显或隐藏）；点击节点详情与徽标一致
- [x] 关联资产tab：根/中间节点显示子树聚合资产清单（含来源节点列），叶子节点显示直属
- [x] suspicious资产行醒目标红（行背景+名称）+ tooltip展示疑异原因
- [x] pnpm typecheck + vitest全绿（并发在途错误除外）；git diff仅service-tree总览相关前端文件

## Notes
分支确认 main 后提交 d453dec，显式 pathspec（src/api/service-tree.ts、src/api/types/service-tree.ts、src/views/service-tree/index.vue、src/views/service-tree/__tests__/index.test.ts），未卷入并发会话文件。Hard Rules 遵守：未触碰 rules 页/accounts/index.vue/logs/components.d.ts。测试基建沿用仓库既有模式（happy-dom + NoopMutationObserver/NoopResizeObserver 桩 + ElementPlus 全插件 + 子组件 vi.mock 替身）。
