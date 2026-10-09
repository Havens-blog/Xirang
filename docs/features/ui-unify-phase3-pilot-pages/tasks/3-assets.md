---
id: "3"
title: "assets 资产列表等价重写（FilterBar/DataTable/抽屉 + 浏览器实测）"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 3: assets 资产列表等价重写（FilterBar/DataTable/抽屉 + 浏览器实测）

## Description
列表页标杆重写（617 行）：PageContainer + FilterBar（现有 5 筛选字段无损表达）+ DataTable（列等价 + 三态 + 批量操作条）+ 规格指定增强：**行点击开 DetailDrawer 2.0** 看资产详情。此页是 220+ 列表页迁移的主样板。

## Reference Files
- `docs/proposals/ui-unify-phase3-pilot-pages/proposal.md` — Proposed Solution / Key Scenarios / Success Criteria
- `src/views/assets/index.vue` — 重写对象（筛选字段/列结构/分页/操作盘点入 AC）
- `src/components/{PageContainer,FilterBar,DataTable,DetailDrawer}/index.vue` — Phase 2 组件
- `docs/features/ui-unify-phase2-components/tasks/8-ui-kit-demo.md` — demo 页列表页样板参照

## Acceptance Criteria
- [ ] 等价重写：旧页用户路径盘点（云厂商/资产类型/区域/状态/资产名称筛选、列结构、分页、查看/编辑/删除操作、导出）逐项保留
- [ ] 规格增强生效：行点击打开 DetailDrawer 2.0 展示资产详情；DataTable 密度切换生效
- [ ] 页面 scoped 样式 <100 行，零新增硬编码色字面量
- [ ] chrome-devtools-mcp 实测：导航 /cam/assets、截图归档、FilterBar 筛选→DataTable 三态→分页链路走通、计算样式断言
- [ ] chrome-devtools-mcp 实测：行点击抽屉打开且加载态/内容正确、批量选择出现操作条、双主题 + hover
- [ ] 相关测试更新后全量绿，`pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 功能等价边界：不顺手新增/优化规格外功能
- 零新增硬编码颜色字面量

## Implementation Notes
- 旧页的 查看/编辑/删除 操作列保留（DataTable columns actions 配置承载）；行点击与行内操作按钮并存时注意事件冒泡（操作列点击不触发行点击）
- Drawer 详情数据加载失败时由 DetailDrawer 内置 StateBlock 错误态承载
