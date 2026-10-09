---
created: "2026-09-24"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: 网络页对齐主机标准 — eip/lb/vpc/vswitch/eni/cdn/waf + image

## Problem

网络系页面（eip/lb/vpc/vswitch/eni/cdn/waf）与 `compute/image` 是另一代骨架：`<PageContainer>` + `ManagerHeader`（标题 + 同步动作 + 内联 stats-bar）+ 专用 `XxxFilters` 组件 + `<el-table>`。它们**没有**大统计卡（stats-bar 是内联紧凑条），但有两个与「主机页紧凑标准」不一致之处：① `PageContainer/ManagerHeader` 外层壳与主机页手写 `page-top` 不同源；② 表格带一个 `selectable 恒 false` 的 disabled 多选列（死列占 40px，主机页已无多选列）。为让全站云产品页观感与交互统一，本批按锁定的「主机页紧凑标准」重写这 8 页（用户已确认 eip 原型 `F:\front_design\prototype-host-eip.html`）。

### Evidence

- 8 页均 `PageContainer` 包裹，`ManagerHeader` 内嘴上 `stats-bar`（总数/某态数），无 `page-top` 手写骨架
- eip/lb/vpc/vswitch/eni/waf 均有 `type="selection"` 列且 `selectable 恒 false`（注释自述「未接线」）
- 设计标准已锁定 `ui-unify-rds-redis-pilot`（后续批沿用），本批唯一新载体是「去 PageContainer 壳 + 亮度内联状态标」，已由 eip 原型定稿

### Urgency

网络系是最后一批两代骨架收敛，闭合后全站云产品成本页（compute/databases/middleware/network/storage）观感统一。

## Proposed Solution

按锁定的「主机页紧凑标准」（eip 原型定稿）重写 8 页：
- 页头：去 PageContainer/ManagerHeader → `page-top`（标题 + tab + 5 内联状态标，竖分隔，无大卡片）
- 操作栏：按各产品领域给动作（eip 申请/绑定/解绑/带宽/释放；lb 创建/监听管理；vpc 创建/子网管理；镜像 创建/导入… 等），保留原「同步」动作，勿照搬主机电源操作
- 搜索：字段搜索栏（回车加条件 + chips）取代内联关键词搜索
- 表格：无默认多选列（删 disabled selection 列）；状态点 + 平台色标 + 等宽 ID；首/末列 20px 对齐
- 详情：行点击/详情开右侧 Drawer（保留现有 DetailDrawer）
- 主题：深/浅一致，Linear 靛紫令牌

## Scope

### In Scope

- network/eip、network/lb、network/vpc、network/vswitch、network/eni、network/cdn、network/waf、compute/image 8 页重写对齐标准

### Out of Scope

- 令牌/主题改动；`compute/provision`（开通向导）、`compute/template`（模板管理）非列表页；数据源/后端接口改动

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 各页领域动作/列梳理不准 | M | M | 以各页现有字段派生；每页执行时实测后端字段有效性 |
| Filters 组件 → 字段搜索栏的能力丢失（导出/自定义列） | M | M | 导出/自定义列收敛为 action-right 图标按钮（主机同款） |

## Success Criteria

- [ ] 8 页均为紧凑标准骨架（page-top + 5 内联状态标、无 PageContainer/ManagerHeader、无默认多选列）
- [ ] 操作栏按各自领域给动作；字段搜索栏 + chips 可用；导出/自定义列保留
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题与标准一致、无控制台错误

consistency_check_result:
  status: pass
  pairs_checked: 8
  conflicts_found: 0

## Next Steps

- Proceed to `/quick-tasks` to generate tasks