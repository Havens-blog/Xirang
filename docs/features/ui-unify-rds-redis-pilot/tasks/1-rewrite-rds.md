---
id: "1"
title: "RDS 页重写对齐主机页紧凑标准"
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

# 1: RDS 页重写对齐主机页紧凑标准

## Description

RDS 云数据库页当前共享早年 `storage.scss`，观感「薄」于主机页（仅 3 状态标、关键词搜索塞操作栏、无字段搜索、无标签筛选、常驻多选列）。本任务把它重写为「主机页紧凑标准」，视觉与交互与主机页同一水准，动作/列/状态标按 RDS 领域定制。

## Reference Files

- `docs/proposals/ui-unify-rds-redis-pilot/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope / Key Risks）
- `src/views/compute/ecs/index.vue` — 参考标准（主机页骨架：page-top + 内联状态标 + 字段搜索栏 + 密集表格）(ref: Proposed Solution)
- `src/views/databases/rds/index.vue` — 待重写页面（356 行）

## Acceptance Criteria

- [ ] RDS 页为紧凑标准骨架：`page-top`（标题 + tab + 内联状态标，竖分隔，无大统计卡）
- [ ] 操作栏为 RDS 领域动作（新建/重启/暂停/变更配置/扩容 等，从本页实际能力派生，不照搬主机「开机/关机」电源操作）
- [ ] 表格无默认多选列（去掉 `col-checkbox` 列与「选中 N 条」字样）；字段搜索栏（回车加条件 + 条件 chips）可用
- [ ] `pnpm build` 与 `pnpm typecheck` 通过；`pnpm lint` 相对基线零新增
- [ ] 浏览器深/浅两主题：RDS 页渲染与原型一致（状态点/平台色标/等宽 ID），无控制台错误

## Hard Rules

- 仅修改 `src/views/databases/rds/**` 文件
- 不改令牌层；不引入大统计卡；不引入 PageContainer 大改

## Implementation Notes

**设计标准（用户三轮定稿，锁死）**：
- 页头：标题 + tab + 内联状态标（`label + 数字` 竖分隔，5 个：总数/运行中/已停止/监控异常/异常）
- 操作栏：**按产品领域**给动作（RDS：新建/重启/暂停/变更配置/扩容/标签），勿照搬主机电源操作
- 搜索：字段搜索栏（回车加条件）＋条件 chips
- 表格：无默认多选列；状态点（绿=运行/灰=停止/黄=异常）＋平台色标＋等宽 ID/IP；列可自定义
- 详情：行点击/「详情」开右侧 Drawer

**参考**：原型 `F:\front_design\prototype-host-disk.html`（RDS 示例页）；主机实现 `src/views/compute/ecs/index.vue`。RDS 现有实现约 356 行，参照主机页裁剪即可。

**验收注意**：
- `pnpm lint` 有 3 个历史 error（`src/api/types/{cost,statistics,task}.ts`），判据「相对基线零新增」，勿让历史 error 阻塞。
- 浏览器实测经 nginx `http://localhost:8888/cam/databases/rds`（勿直连 :5173，无共享 cookie），用 chrome-devtools-mcp 断言深/浅主题渲染 + 无控制台错误。