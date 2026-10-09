---
id: "8"
title: "/dev/ui-kit 组件 demo 页（迁移样板 + 验收靶场）"
priority: "P1"
estimated_time: "1.5h"
complexity: "medium"
dependencies: [1, 2, 3, 4, 5, 6]
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 8: /dev/ui-kit 组件 demo 页（迁移样板 + 验收靶场）

## Description
内部 demo 路由页（hideInMenu）：集中展示 6 个组件的典型用法与状态组合，每个示例代码可直接复制进业务页。作为 Phase 3 迁移 220+ 页的"照抄样板"与双主题可视化验收靶场。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Proposed Solution / Success Criteria（demo 页条款）
- `src/router/routes.ts` — 路由注册模式（hideInMenu 用法）
- `src/components/` — Task 1-6 产出

## Acceptance Criteria
- [ ] `/dev/ui-kit` 路由可达且 hideInMenu: true（不进侧栏菜单）
- [ ] 6 组件典型用法全部展示：StateBlock 四态、FilterBar 折叠态、DataTable 三态+批量操作、PageContainer 完整骨架、DetailDrawer 分区块、CommandPalette 唤起入口
- [ ] 每个示例的源码片段在页面上可见或可一键复制（照抄样板定位）
- [ ] 内置深/浅主题切换预览，双主题下浏览器实测视觉正确（含 hover 态）
- [ ] `pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- demo 页自身也零硬编码颜色字面量（它本身是样板）
- 不修改 Task 1-6 的组件源码；发现组件缺陷回到对应任务修复

## Implementation Notes
- 页面结构按组件分区块，锚点导航；示例源码用字符串常量内嵌展示，避免运行时读文件
