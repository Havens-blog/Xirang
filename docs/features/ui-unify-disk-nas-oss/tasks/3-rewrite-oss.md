---
id: "3"
title: "OSS 页重写对齐主机页紧凑标准"
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

# 3: OSS 页重写对齐主机页紧凑标准

## Description

对象存储 OSS 页（532 行）当前为早年薄骨架 + 运营指标卡。重写为「主机页紧凑标准」，运营指标卡保留为紧凑指标条，动作/列/状态标按 OSS 领域派生（桶/存储类型/容量/用量）。

## Reference Files

- `docs/proposals/ui-unify-disk-nas-oss/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）(ref: Proposed Solution)
- `docs/features/ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md` — 同批先例（设计标准表 + 验收方法）
- `src/views/storage/oss/index.vue` — 待重写页面

## Acceptance Criteria

- [ ] OSS 页为紧凑标准骨架：page-top（标题 + tab + 5 内联状态标，无大卡片）
- [ ] 运营指标卡保留为紧凑指标条（warn/empty 态不丢）
- [ ] 操作栏为 OSS 领域动作；表格无默认多选列；字段搜索栏 + chips 可用
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/storage/oss/**` 文件
- 不改令牌层；不引入大统计卡/PageContainer

## Implementation Notes

**设计标准**：同 `ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md`（锁死）。OSS 保留紧凑指标条（与 Disk 任务 1 同法）；动作/列以 `src/views/storage/oss/index.vue` 现有字段派生（桶名/存储类型/容量/用量等）。

**首/末列对齐**：`th:first-child/td:first-child { padding-left: 20px }`、`last-child { padding-right: 20px }`。

**验收注意**：`pnpm lint` 历史基线零新增判据；浏览器实测经 nginx `http://localhost:8888/cam/storage/oss`，用 evaluate_script 断言深/浅主题 + 首列对齐 + 指标条 + 无 console error。