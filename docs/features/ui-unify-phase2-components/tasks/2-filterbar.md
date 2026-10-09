---
id: "2"
title: "FilterBar 配置化筛选栏组件"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 2: FilterBar 配置化筛选栏组件

## Description
Phase 2 布局层组件。以声明式字段数组自动渲染筛选栏（input/select/date 等 el-form-item），消灭各页面手写的 `xxx-filters` 结构（现状参考 src/views/assets 里的 asset-filters）。字段过多时响应式折叠进"展开更多"。消费 Linear 令牌。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Proposed Solution / Key Scenarios
- `src/views/assets/index.vue` — 现有手写筛选栏的真实用法（反推 API 的锚点）

## Acceptance Criteria
- [ ] `src/components/FilterBar/index.vue` 落地：fields 数组配置（type: input/select/date，含 label/placeholder/options/clearable）自动渲染，v-model 双向绑定筛选值对象
- [ ] 支持 select 远程/本地 options、日期范围、字段 change 事件透传
- [ ] 字段数超过阈值（默认一行放不下）时自动折叠 + "展开/收起"切换
- [ ] TSDoc 覆盖 100% 对外 props/slots/events；单测覆盖字段渲染/双向绑定/折叠逻辑，行覆盖率 ≥80%
- [ ] 组件源码零硬编码颜色字面量（仅 var(--*)）
- [ ] `pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 不得引入新运行时依赖
- 零硬编码颜色字面量（只用 var(--*) 令牌）

## Implementation Notes
- API 反推锚点：assets 页现有筛选字段（云厂商 select/资产类型 select/区域 input/状态 select/资产名称 input）必须能用 FilterBar 无损表达
- 折叠阈值用容器宽度测量（ResizeObserver）而非固定条数
