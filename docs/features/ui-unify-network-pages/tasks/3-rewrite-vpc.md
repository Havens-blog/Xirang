---
id: "3"
title: "VPC 页重写对齐主机页紧凑标准"
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

# 3: VPC 页重写对齐主机页紧凑标准

## Description

VPC 专有网络页（606 行）当前为 `PageContainer + ManagerHeader + stats-bar` 骨架（带 disabled selection 列）。重写为「主机页紧凑标准」。

## Reference Files

- `docs/proposals/ui-unify-network-pages/proposal.md` — Source proposal
- `F:\front_design\prototype-host-eip.html` — 视觉样板
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）
- `docs/features/ui-unify-security-group-snapshot/tasks/1-rewrite-security-group.md` — 同批先例
- `src/views/network/vpc/index.vue` — 待重写页面

## Acceptance Criteria

- [ ] VPC 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标，按 VPC 状态/默认域派生）
- [ ] 表格删除 disabled selection 列；操作栏为 VPC 领域动作（创建 VPC/子网管理 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [ ] 列按 VPC 领域：名称/ID(等宽ID)、CIDR、状态(状态点)、子网数、共享服务器、默认/网络域、平台色标、地域（现页 `cidr_block/vswitch_count/enable_internet_access/is_default/network_domain/shared_machines/status`）
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/network/vpc/**` 文件
- 不改令牌层；不引入大统计卡/PageContainer

## Implementation Notes

**视觉标准**：同 eip 原型。字段搜索栏 chips 按现页 `col.key` 派生并实测后端有效参数。保留现有 DetailDrawer（含安全组/子网子查询能力）。去壳手法同先例。