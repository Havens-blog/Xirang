---
id: "2"
title: "LB 页重写对齐主机页紧凑标准"
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

# 2: LB 页重写对齐主机页紧凑标准

## Description

负载均衡页（434 行）当前为 `PageContainer + ManagerHeader + stats-bar` 骨架（带 disabled selection 列）。重写为「主机页紧凑标准」（eip 原型为视觉样板）。

## Reference Files

- `docs/proposals/ui-unify-network-pages/proposal.md` — Source proposal
- `F:\front_design\prototype-host-eip.html` — 视觉样板（列/动作按 LB 领域改）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）
- `docs/features/ui-unify-security-group-snapshot/tasks/1-rewrite-security-group.md` — 同批先例
- `src/views/network/lb/index.vue` — 待重写页面

## Acceptance Criteria

- [ ] LB 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标，按 LB 状态域派生）
- [ ] 表格删除 disabled selection 列；操作栏为 LB 领域动作（创建 LB/监听管理/后端服务器 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [ ] 列按 LB 领域：名称/ID(等宽ID)、类型(SLB/ALB/NLB)、状态(状态点)、VIP、监听协议/端口、后端实例数、平台色标、地域
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/network/lb/**` 文件
- 不改令牌层；不引入大统计卡/PageContainer

## Implementation Notes

**视觉标准**：同 eip 原型，列/动作/状态标按 LB 领域派生（现页 `col.key`/字段实测）。去壳手法同先例。状态标 5 个按 LB 状态/类型域派生。字段搜索栏 chips 实测后端有效参数。