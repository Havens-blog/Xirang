---
id: "2"
title: "前端 DashboardTable 证书行渲染 + 隐藏切换 + 云选项派生 + 抽屉多 SAN 视图"
priority: "P1"
estimated_time: "3h"
complexity: "high"
dependencies: ["1"]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 2: 前端 DashboardTable 证书行渲染 + 隐藏切换 + 云选项派生 + 抽屉多 SAN 视图

## Description

看板表格按证书行渲染（CN、SANS 折叠计数、issuer、daysLeft、托管、引用云、探测徽标），行 key 改证书 ID（现用 row-key=domain，证书粒度化后域名不再唯一会导致 Vue 行复用错乱）；孤儿默认隐藏 + 工具栏横幅开关（状态按用户持久化 localStorage，对账工作流刷新不丢）；云筛选选项由当前可见行派生（隐藏态不含孤儿所属云）；豁免（风险维度卡激活或开关开启）时以 includeHidden=true 重新拉取；行点击进入由现有探测抽屉组件扩展的"证书多 SAN 视图"（SAN 超 20 折叠，最差优先前 20，其余可展开）。

## Reference Files
- `docs/proposals/dashboard-cert-granularity/proposal.md` — In Scope / Non-Functional Requirements / Success Criteria
- `e-cam-web/src/views/cert/dashboard/components/DashboardTable.vue` — 行渲染与 row-key (ref: ### In Scope)
- `e-cam-web/src/views/cert/dashboard/components/ProbeDetailDrawer.vue` — 抽屉扩展证书多 SAN 视图 (ref: ### In Scope)
- `e-cam-web/src/api/cert.ts` — 载荷类型与 includeHidden 参数 (ref: ### Non-Functional Requirements)

## Acceptance Criteria
- [ ] 证书行渲染：CN + SANS 折叠计数 + issuer + daysLeft + 托管 + 引用云 + 探测徽标；row-key 改为证书 ID（组件测试断言同域两证不串行）
- [ ] 孤儿默认隐藏 + 工具栏横幅开关（localStorage 持久化、刷新保留）+ 空态提示"查看全部"入口保留（不静默吞行）
- [ ] 云筛选选项由当前可见行派生，隐藏态下不含孤儿所属云（组件测试锁定）
- [ ] 豁免接线：风险维度卡激活或开关开启 → includeHidden=true 重新拉取全量；前端不派生三态（只消费 visible 行 + hiddenCount）
- [ ] 行点击 → 证书多 SAN 探测视图（头部证书摘录 + 该证全部 SAN 探测列表，SAN>20 折叠为最差优先前 20）
- [ ] 组件测试：开关状态机（分级卡激活强制开启并禁用、清除恢复原态）、localStorage 持久化、隐藏态云选项；vitest + vue-tsc 全绿

## Hard Rules
隐藏判定不得在前端重写三态（复用服务端结果）；cert.ts 类型与后端 VO json tag 对齐；行 key 必须证书 ID。

## Implementation Notes
- eval 残留 #12：localStorage 键名需带命名空间（如 cert.dashboard.hiddenExpanded），前端回退旧版本时无异常行为。
- 豁免 fetch 接线是总规则闭环的关键：分级/特殊卡激活或开关开启 → includeHidden=true；清除 → 回到 false。缺此接线则客户端页内筛选永远凑不齐该档 total。
- SAN>20 折叠阈值已在 NFR 锚定，实现时沿用。