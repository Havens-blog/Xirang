---
id: "5"
title: "DetailDrawer 2.0 统一详情抽屉（向后兼容 v1）"
priority: "P1"
estimated_time: "1.5h"
complexity: "medium"
dependencies: [1]
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 5: DetailDrawer 2.0 统一详情抽屉（向后兼容 v1）

## Description
Phase 2 交互层组件。增强现有 DetailDrawer（src/components/DetailDrawer/index.vue）：描述列表（label-value 网格）、分区块（sections）、内置加载态（复用 Task 1 StateBlock）。v1 API 向后兼容——现有调用方（accounts/cmdb/iam/network/storage 等多处 DetailDrawer 使用）零改动。

## Reference Files
- `docs/proposals/ui-unify-phase2-components/proposal.md` — Proposed Solution / Key Scenarios
- `src/components/DetailDrawer/index.vue` — v1 现状与现有调用契约
- `src/views/accounts/components/AccountDetailDrawer.vue` — 真实使用形态参考

## Acceptance Criteria
- [ ] v1 对外 API（props/slots/events）逐项保留，现有调用方零改动通过 `pnpm build` + `pnpm typecheck`
- [ ] 新增描述列表模式：items 配置（label/value/formatter/span）渲染 label-value 网格，支持 sections 分区块
- [ ] 内置加载态（复用 StateBlock loading）
- [ ] TSDoc 覆盖 100% 对外 API；单测覆盖兼容契约 + 新增能力，行覆盖率 ≥80%
- [ ] 组件源码零硬编码颜色字面量（仅 var(--*)）
- [ ] 深浅主题下视觉正确（demo 页联验前先自测）

## Hard Rules
- 不得修改任何现有 DetailDrawer 调用方文件
- 零硬编码颜色字面量（只用 var(--*) 令牌）

## Implementation Notes
- 兼容优先：v1 的 slot 用法（若有）必须原样工作；新增能力走新 props/slots，不改旧签名
