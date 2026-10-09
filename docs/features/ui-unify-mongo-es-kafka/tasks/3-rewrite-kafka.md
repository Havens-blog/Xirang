---
id: "3"
title: "Kafka 页重写对齐主机页紧凑标准"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 3: Kafka 页重写对齐主机页紧凑标准

## Description

Kafka 中间件页（224 行）当前为早年薄骨架。重写为「主机页紧凑标准」，动作/列/状态标按 Kafka 领域派生（Topic/分区/副本维度）。

## Reference Files

- `docs/proposals/ui-unify-mongo-es-kafka/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）(ref: Proposed Solution)
- `docs/features/ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md` — 同批先例（设计标准表 + 验收方法）
- `src/views/middleware/kafka/index.vue` — 待重写页面

## Acceptance Criteria

- [ ] Kafka 页为紧凑标准骨架：page-top（标题 + tab + 内联状态标，无大卡片）
- [ ] 操作栏为 Kafka 领域动作（从本页实际能力派生，勿照搬主机电源操作）
- [ ] 表格无默认多选列；字段搜索栏（回车加条件 + chips）可用
- [ ] `pnpm build` 与 `pnpm typecheck` 通过；`pnpm lint` 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/middleware/kafka/**` 文件
- 不改令牌层；不引入大统计卡/PageContainer

## Implementation Notes

**设计标准**：与 `ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md` 设计标准表一致（锁死）；Kafka 领域动作/列以本页现有字段派生（Topic/分区/副本/消费者组等）。

**首/末列对齐**：`th:first-child/td:first-child { padding-left: 20px }`、`last-child { padding-right: 20px }`。

**验收注意**：`pnpm lint` 历史 3 error 判据「相对基线零新增」；浏览器实测经 nginx `http://localhost:8888/cam/middleware/kafka`，用 evaluate_script 断言深/浅主题 + 首列对齐 + 无 console error。