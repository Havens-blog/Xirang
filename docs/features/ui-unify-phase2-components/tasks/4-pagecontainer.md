---
id: "4"
title: "PageContainer 2.0 页面骨架（向后兼容 v1）"
priority: "P0"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 4: PageContainer 2.0 页面骨架（向后兼容 v1）

## Description
Phase 2 布局层容器。将 v1（空壳 div，现状 22 行）升级为标准页面骨架：页头（title + actions 操作区）、筛选区、内容区（默认 slot）、底部分页区，统一间距与滚动行为。**v1 的默认 slot 用法天然兼容**——现有 29 个使用方零改动。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Proposed Solution / Assumptions Challenged（v1 空壳结论）
- `src/components/PageContainer/index.vue` — v1 现状（仅默认 slot + 样式壳）
- `src/views/assets/index.vue` — 页头（标题+描述+操作按钮）真实形态参考

## Acceptance Criteria
- [ ] 2.0 骨架：具名插槽 header（含 title prop 快捷方式 + actions slot）/ filters / 默认内容 / footer（分页位），间距与滚动统一由组件承载
- [ ] 向后兼容：不传任何新 prop/slot 时行为与 v1 完全一致（默认 slot 直出）
- [ ] 现有 29 个使用方零改动通过 `pnpm build` + `pnpm typecheck`（验证方式：全量构建 + grep 使用方清单抽样核对）
- [ ] TSDoc 覆盖 100% 对外 props/slots
- [ ] 单测覆盖骨架渲染/向后兼容场景/插槽分发，行覆盖率 ≥80%
- [ ] 组件源码零硬编码颜色字面量（仅 var(--*)）

## Hard Rules
- 不得修改任何 src/views/** 使用方文件
- 零硬编码颜色字面量（只用 var(--*) 令牌）

## Implementation Notes
- v1 使用方里 `title=` 是原生 HTML 属性透传（非组件 prop），2.0 定义 title prop 后该透传语义转为组件 prop，属兼容性增强而非破坏
- 29 个使用方清单以 `grep -rl "PageContainer" src/views/` 为准
