---
id: "2"
title: "Snapshot 页重写对齐主机页紧凑标准"
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

# 2: Snapshot 页重写对齐主机页紧凑标准

## Description

快照页（256 行）当前为早年薄骨架（2 状态标、关键词搜索塞操作栏、手写 table、常驻多选列）。重写为「主机页紧凑标准」，动作/列/状态标按快照领域派生（状态、容量、源磁盘/实例、进度）。

## Reference Files

- `docs/proposals/ui-unify-security-group-snapshot/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）(ref: Proposed Solution)
- `docs/features/ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md` — 设计标准表 + 验收方法（锁死）
- `src/views/compute/snapshot/index.vue` — 待重写页面

## Acceptance Criteria

- [ ] Snapshot 页为紧凑标准骨架：page-top（标题 + tab + 5 内联状态标，无大卡片）
- [ ] 操作栏为快照领域动作（勿照搬主机电源操作）；表格无默认多选列；字段搜索栏 + chips 可用
- [ ] 列按快照领域派生：快照 ID/名称、状态、容量、源磁盘/实例、平台色标、地域、创建时间等
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/compute/snapshot/**` 文件
- 不改令牌层；不引入大统计卡/PageContainer

## Implementation Notes

**设计标准**：同 `ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md`（锁死）。动作/列以 `src/views/compute/snapshot/index.vue` 现有字段派生（`status` 正常/创建中/失败、容量、源资源、`create_time` 等）。本页无运营指标卡，只把现有 2 个状态标扩为 5 个内联状态标。

**动作**（快照领域）：详情 + 领域相关动作（如删除快照/回滚），从现有能力派生，勿照搬主机电源操作。

**首/末列对齐**：`th:first-child/td:first-child { padding-left: 20px }`、`last-child { padding-right: 20px }`（去多选列后首列 12px 会错位 8px）。

**验收注意**：`pnpm lint` 历史基线零新增判据；浏览器实测经 nginx `http://localhost:8888/cam/compute/snapshot`，用 evaluate_script 断言深/浅主题 + 首列对齐 + 无 console error。