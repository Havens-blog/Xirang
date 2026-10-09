---
id: "1"
title: "EIP 页重写对齐主机页紧凑标准"
priority: "P0"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: EIP 页重写对齐主机页紧凑标准

## Description

弹性公网 IP 页（597 行）当前为 `PageContainer + ManagerHeader + stats-bar + EipFilters + el-table`（带 disabled selection 列）。重写为「主机页紧凑标准」（用户已确认原型 `F:\front_design\prototype-host-eip.html`）。

## Reference Files

- `docs/proposals/ui-unify-network-pages/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope）
- `F:\front_design\prototype-host-eip.html` — 本页定稿原型（视觉标准，锁定）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）
- `docs/features/ui-unify-security-group-snapshot/tasks/1-rewrite-security-group.md` — 同批先例（去壳逻辑 + 验收方法）
- `src/views/network/eip/index.vue` — 待重写页面

## Acceptance Criteria

- [ ] EIP 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab 全部/已绑定/未绑定 + 5 内联状态标：总数/已绑定/未绑定/带宽告警/异常）
- [ ] 表格删除 disabled selection 列；操作栏为 EIP 领域动作（申请 EIP/绑定/解绑/修改带宽/释放/标签 + 保留同步）；字段搜索栏 + chips 可用；导出/自定义列保留为图标
- [ ] 列按 EIP 领域：名称/ID(等宽ID副行)、公网 IP、状态(状态点)、带宽、计费方式、关联实例、平台色标、地域
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与原型一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/network/eip/**` 文件
- 不改令牌层；不引入大统计卡/PageContainer；改动不得影响后端接口契约

## Implementation Notes

**视觉标准**：以 `F:\front_design\prototype-host-eip.html` 为准；去壳手法同 `1-rewrite-security-group.md`。字段搜索栏 chips 按 EIP 现有 `col.key` 派生（ip_address/status/charge_type/instance_id/platform/region/account_name），实测后端有效参数再映射。保留现有 `EipFilters` 的导出/自定义列能力（收敛为 action-right 图标）与 DetailDrawer。