---
id: "2"
title: "Redis 页重写对齐主机页紧凑标准"
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

# 2: Redis 页重写对齐主机页紧凑标准

## Description

Redis 云数据库页与 RDS 同病——共享早年 `storage.scss`，观感薄于主机页。本任务把它同样重写为「主机页紧凑标准」，动作/列/状态标按 Redis 领域定制。

## Reference Files

- `docs/proposals/ui-unify-rds-redis-pilot/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope / Key Risks）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）(ref: Proposed Solution)
- `src/views/databases/redis/index.vue` — 待重写页面（253 行）

## Acceptance Criteria

- [ ] Redis 页为紧凑标准骨架：`page-top`（标题 + tab + 内联状态标，竖分隔，无大统计卡）
- [ ] 操作栏为 Redis 领域动作（从本页实际能力派生：如新建/重启/密码重置/参数配置，不照搬主机电源操作）
- [ ] 表格无默认多选列；字段搜索栏（回车加条件 + 条件 chips）可用
- [ ] `pnpm build` 与 `pnpm typecheck` 通过；`pnpm lint` 相对基线零新增
- [ ] 浏览器深/浅两主题：Redis 页渲染与标准一致，无控制台错误

## Hard Rules

- 仅修改 `src/views/databases/redis/**` 文件
- 不改令牌层；不引入大统计卡；不引入 PageContainer 大改

## Implementation Notes

**设计标准**：与任务 1（RDS）同一套，见 `docs/features/ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md` 的 Implementation Notes 设计标准表。区别仅在 Redis 的领域动作与列：以 `src/views/databases/redis/index.vue` 现有字段/操作派生（Redis 特有：内存容量/版本/密码重置等），勿与 RDS/主机混用。

**参考**：原型 `F:\front_design\prototype-host-disk.html`（RDS 示例，Redis 用同一骨架）；主机实现 `src/views/compute/ecs/index.vue`。

**验收注意**：
- `pnpm lint` 历史 3 error 判据「相对基线零新增」。
- 浏览器实测经 nginx `http://localhost:8888/cam/databases/redis`，用 chrome-devtools-mcp 断言深/浅主题渲染 + 无控制台错误。