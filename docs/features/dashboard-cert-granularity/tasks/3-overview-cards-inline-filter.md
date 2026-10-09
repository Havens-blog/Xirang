---
id: "3"
title: "前端 OverviewCards 分级卡/特殊卡改页内筛选 + 豁免 + 移除跳台账"
priority: "P1"
estimated_time: "2h"
complexity: "medium"
dependencies: ["1"]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 3: 前端 OverviewCards 分级卡/特殊卡改页内筛选 + 豁免 + 移除跳台账

## Description

分级卡点击由「跳转台账按档过滤」改为页内筛选（选中高亮、再点取消、aria-pressed 两态）；激活即按总规则豁免孤儿隐藏（前端以 includeHidden=true 重拉，见任务 2），行数 == 该档 countsByLevel[tier].total；未激活时该档显示"可见 N · 隐藏 M"。移除 router.push('/certs') 跳转与静态文案，筛选状态同步 URL query 保留深链。特殊卡 diff/exempt 计数保持域名级（副标题标注「域名计数」），行集合 = 存在任一 SAN 为该状态的证书行（逐 SAN 谓词，非行徽标相等）。

## Reference Files
- `docs/proposals/dashboard-cert-granularity/proposal.md` — In Scope / Key Scenarios / Success Criteria
- `e-cam-web/src/views/cert/dashboard/components/OverviewCards.vue` — 分级卡交互改页内 filter (ref: ### In Scope)
- `e-cam-web/src/views/cert/dashboard/index.vue` — onLevelCard router.push 移除 + 页内筛选接线 (ref: ### In Scope)
- `e-cam-web/src/views/cert/dashboard/format.ts` — filterDashboardItems 逐 SAN 谓词 + aria-pressed (ref: ### In Scope)

## Acceptance Criteria
- [ ] 分级卡点击页内筛选：选中高亮、再点取消恢复；grep `src/views/cert/dashboard/` 无 `router.push('/certs'` 残留
- [ ] 激活该档 → 豁免孤儿隐藏（总规则）→ 主表行数 == countsByLevel[tier].total（纯关系断言，不写死数值）；未激活显示"可见 N · 隐藏 M"
- [ ] 特殊卡 diff/exempt：计数保持域名级（副标题「域名计数」），行集合 = 任一 SAN 为该状态（逐 SAN 谓词）；单测断言"一个 diff 域名挂 2 张证 → 卡 1 行 2"映射及 exempt 行被更差态掩盖仍可筛中
- [ ] 筛选状态同步 URL query 保留深链（弥合移除 router.push）
- [ ] 分级卡选中/未选两态 aria-pressed 各有 vitest 断言；format.ts 测试随逐 SAN 谓词更新
- [ ] vitest + vue-tsc 全绿

## Hard Rules
豁免仅由风险维度搜索触发（到期分级卡 + diff/exempt 探测卡），云/托管/issuer/关键字等属性维度筛选不豁免（孤儿保持隐藏，仅当前可见行集再过滤）；filter 单一状态源机制沿用（与特殊卡交互一致）。

## Implementation Notes
- eval 残留 #9：移除跳台账后，有权限角色的台账入口由菜单「证书管理 > 证书台账」承担；需验证只读角色无台账入口时看板是否仍满足只读浏览需求（本任务不新增入口）。
- 特殊卡副标题「域名计数」是用户理解"卡值小于筛出行数（1→2）"的关键文案，实现时保留。
- 分级卡激活与"查看全部"开关的优先级：激活时开关强制开启并禁用，清除筛选恢复原开关态（见任务 2 状态机）。