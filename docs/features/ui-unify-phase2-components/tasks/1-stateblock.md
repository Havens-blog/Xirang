---
id: "1"
title: "StateBlock 统一三态组件（skeleton/empty/error）"
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

# 1: StateBlock 统一三态组件（skeleton/empty/error）

## Description
Phase 2 布局层第一个组件。提供 `status` 驱动的统一三态块：loading（skeleton 骨架屏）/ empty（空状态）/ error（错误态含重试按钮），默认插槽承载成功态内容。吸收现有 EmptyState（自定义文案/图标）与 ErrorDisplay（错误信息+重试）的能力，但**两个旧组件文件保留不动**（迁移期由 Phase 3/4 处理）。消费 Phase 1 Linear 令牌（仅 var(--*)），供 DataTable 与业务页面复用。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Proposed Solution / Key Scenarios / Non-Functional Requirements
- `src/components/EmptyState.vue` — 被吸收能力的参考（文案/图标自定义）
- `src/components/ErrorDisplay.vue` — 被吸收能力的参考（错误信息+重试）

## Acceptance Criteria
- [ ] `src/components/StateBlock/index.vue` 落地：status=loading/empty/error/success 四态，loading 渲染 el-skeleton，error 渲染错误信息 + 重试按钮并 emit `retry`，默认插槽为成功态内容
- [ ] 支持自定义文案、图标、错误详情透传（覆盖 EmptyState/ErrorDisplay 既有能力）
- [ ] TSDoc 覆盖 100% 对外 props/slots/events
- [ ] 单测覆盖四态切换、retry 事件、插槽透传，行覆盖率 ≥80%，测试通过且输出无噪音
- [ ] 组件源码零硬编码颜色字面量（仅 var(--*)；TSDoc 示例文本除外）
- [ ] `pnpm build` + `pnpm typecheck` 通过

## Hard Rules
- 不得修改或删除 src/components/EmptyState.vue 与 src/components/ErrorDisplay.vue
- 零硬编码颜色字面量（只用 var(--*) 令牌）

## Implementation Notes
- 参考 Phase 1 令牌：--text-secondary/--text-tertiary 用于文案层级，--accent-* 用于强调
- 现有 28 个测试文件的组织模式（src/components/__tests__/）沿用
