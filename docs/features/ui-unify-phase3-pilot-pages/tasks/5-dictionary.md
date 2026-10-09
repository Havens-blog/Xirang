---
id: "5"
title: "system/dictionary 字典管理等价重写（表单页样板 + 浏览器实测）"
priority: "P1"
estimated_time: "1.5h"
complexity: "high"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 5: system/dictionary 字典管理等价重写（表单页样板 + 浏览器实测）

## Description
表单页标杆重写（330 行）：PageContainer + DataTable 列表 + 新增/编辑弹窗表单统一布局与校验反馈。作为 220+ 页中"列表 + 弹窗表单"组合类型的迁移样板。

## Reference Files
- `docs/proposals/ui-unify-phase3-pilot-pages/proposal.md` — Proposed Solution / Key Scenarios
- `src/views/system/dictionary/index.vue` — 重写对象（CRUD 链路盘点入 AC）
- `src/views/system/dictionary/components/` — DictTypeForm/DictItemForm/DictItemTable/DictTypeList 子组件（可沿用，布局统一）
- `src/components/{PageContainer,DataTable}/index.vue` — Phase 2 组件

## Acceptance Criteria
- [ ] 等价重写：旧页 CRUD 用户路径盘点（字典类型列表/新增/编辑/删除、字典项管理）逐项保留
- [ ] 弹窗表单布局与校验反馈样式统一（错误提示/label 对齐/按钮区），页面 scoped 样式 <100 行、零新增硬编码色
- [ ] chrome-devtools-mcp 实测：导航本页、截图归档、CRUD 链路走通（新增→编辑→删除全链路，测试数据自建自清）、计算样式断言
- [ ] chrome-devtools-mcp 实测：表单校验错误态视觉正确（触发一次校验失败）、双主题正常
- [ ] 相关测试更新后全量绿，`pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 功能等价边界：不顺手新增/优化规格外功能
- 零新增硬编码颜色字面量

## Implementation Notes
- CRUD 实测用自建测试数据（名称带 e2e 前缀），结束清理，不污染真实字典
- 子组件（DictTypeForm 等）如无需改动则不动，重写聚焦 index.vue 与布局
