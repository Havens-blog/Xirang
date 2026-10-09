---
id: "6"
title: "CDN 页重写对齐主机页紧凑标准"
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

# 6: CDN 页重写对齐主机页紧凑标准

## Description

CDN 加速页（538 行）当前为 `PageContainer + ManagerHeader + stats-bar` 骨架（「在线域名」统计）。重写为「主机页紧凑标准」。

## Reference Files

- `docs/proposals/ui-unify-network-pages/proposal.md` — Source proposal
- `F:\front_design\prototype-host-eip.html` — 视觉样板
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）
- `docs/features/ui-unify-security-group-snapshot/tasks/1-rewrite-security-group.md` — 同批先例
- `src/views/network/cdn/index.vue` — 待重写页面

## Acceptance Criteria

- [ ] CDN 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标：总数/在线/离线/HTTPS 开启/证书 等派生）
- [ ] 表格删除 disabled selection 列；操作栏为 CDN 领域动作（添加域名/刷新缓存 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [ ] 列按 CDN 领域：加速域名(等宽ID)、CNAME、业务类型、HTTPS、证书、服务区域、状态(状态点)、平台色标（现页 `business_type/cname/https_enabled/cert_name/service_area/status`）
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/network/cdn/**` 文件
- 不改令牌层；不引入大统计卡/PageContainer

## Implementation Notes

**视觉标准**：同 eip 原型。列/动作/状态标按 CDN 领域派生并实测后端字段。CNAME 列保留 mono 样式 + 复制提示（若现页有）。去壳手法同先例。