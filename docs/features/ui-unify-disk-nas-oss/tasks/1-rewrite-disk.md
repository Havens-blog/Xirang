---
id: "1"
title: "Disk 页重写对齐主机页紧凑标准"
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

# 1: Disk 页重写对齐主机页紧凑标准

## Description

云盘页（598 行）当前为早年薄骨架 + 运营指标卡。重写为「主机页紧凑标准」，**运营指标卡保留为紧凑指标条**（磁盘数/平均使用率/IO，保留 warn/empty 态），动作/列/状态标按 Disk 领域派生（磁盘类型/使用率/IOPS/吞吐量/挂载实例）。

## Reference Files

- `docs/proposals/ui-unify-disk-nas-oss/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架）(ref: Proposed Solution)
- `docs/features/ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md` — 同批先例（设计标准表 + 验收方法）
- `src/views/compute/disk/index.vue` — 待重写页面（含 diskMetrics.ts 指标数据源）

## Acceptance Criteria

- [ ] Disk 页为紧凑标准骨架：page-top（标题 + tab + 5 内联状态标，无大卡片）
- [ ] 运营指标卡保留为紧凑指标条（磁盘数/平均使用率/IO 繁忙，warn/empty 态不丢）
- [ ] 操作栏为 Disk 领域动作；表格无默认多选列；字段搜索栏 + chips 可用
- [ ] 首/末列 20px 对齐；`pnpm build`/`typecheck` 通过、lint 相对基线零新增
- [ ] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Hard Rules

- 仅修改 `src/views/compute/disk/**` 文件（`diskMetrics.ts` 可保留不改）
- 不改令牌层；不引入大统计卡/PageContainer

## Implementation Notes

**设计标准**：与 `ui-unify-rds-redis-pilot/tasks/1-rewrite-rds.md` 设计标准表一致（锁死）。**区别**：Disk 保留运营指标条——三卡并排、纵向 padding ≤16px、横向与页头 20px 同节奏，保留原 warn（采集异常）/empty（无指标数据）态；使用率列用进度条（对齐 RDS 先例）。

**动作**（Disk 领域）：新建/扩容/快照 等，从 `src/views/compute/disk/index.vue` 现有能力派生，勿照搬主机电源操作。

**首/末列对齐**：`th:first-child/td:first-child { padding-left: 20px }`、`last-child { padding-right: 20px }`。

**验收注意**：`pnpm lint` 历史 3 error 判据「相对基线零新增」；浏览器实测经 nginx `http://localhost:8888/cam/compute/disk`，用 evaluate_script 断言深/浅主题 + 首列对齐 + 指标条存在 + 无 console error。