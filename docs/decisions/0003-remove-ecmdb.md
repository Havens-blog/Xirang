# 决策记录：移除 ecmdb + ecmdb-web —— 推翻 0001「保留 ecmdb」决策

| 项目 | 内容 |
|---|---|
| 状态 | 已决策（2026-10-08），执行见 feature `remove-ecmdb-stack` |
| 决策人 | 项目负责人 |
| 影响范围 | ecmdb、ecmdb-web、eiam、e-cam-service、e-cam-web、haven-console、nginx 根入口、`dev.ps1` 编排 |
| 取代 | [`0001-cmdb-consolidation-keep-ecmdb.md`](./0001-cmdb-consolidation-keep-ecmdb.md)（已标废弃） |

## 1. 背景

0001 决策（2026-09-08）认定平台存在两套 CMDB 实现且双写共库，结论为「保留 ecmdb 为唯一 CMDB，e-cam-service CMDB 退位」，并规划了 Phase 0~4 的收敛迁移。2026-10-08 的代码核查与用户确认推翻了该前提：ecmdb 自身承载的能力已全部废弃，无需收敛，应直接退役。

依据见 [`docs/proposals/remove-ecmdb-stack/proposal.md`](../proposals/remove-ecmdb-stack/proposal.md) 的 Problem / Assumptions Challenged / Feasibility Assessment 节。

## 2. 决策

**移除 `ecmdb` 与 `ecmdb-web` 两个 Duke1616 上游 checkout，0001「保留 ecmdb」决策废弃。** 这是纯退役清理，不迁移任何能力（能力已废弃），不删除 0001 原文（保留审计线索）。

## 3. 推翻依据（依赖安全判定）

| 被推翻的假设 | 核查方法 | 结论 |
|---|---|---|
| 「ecmdb 被 eiam/e-cam-service 依赖，删不得」 | 代码核查（go.mod / import） | **推翻**：eiam 依赖的是已发布模块 `github.com/Duke1616/ecmdb v1.11.0`（仅 `pkg/cryptox`，经 go.sum/module 缓存解析），**非本地 `ecmdb` 目录**；e-cam-service 运行时对 ecmdb 服务零调用，仅 `ecmdb-token-key` cookie 名、Mongo 用户名 `ecmdb` 等命名残留（共享约定，重命名风险大于收益，不在本范围）。删本地目录不打断 eiam/e-cam-service 编译。 |
| 「CMDB/工单等能力可能还在用」 | 用户确认 | **推翻**：2026-10-08 用户确认 ecmdb-web 剩余页面（alert / change / cmdb / task / ticket）「都废弃，直接全删」，无任何使用方。 |
| 「根路径还得留个页面」 | 用户确认 | **推翻**：nginx 根入口 `/` 改 301 重定向到 `/console/` 即可，语义无损。 |

> **关键可行性前提**：通过 go.mod 判定 eiam 走已发布模块而非本地目录，从而确认删除本地 checkout 对构建零影响——这是本次删除的可行性前提。`eiam` 及其 go.mod 对 `ecmdb v1.11.0/cryptox` 的**库**依赖属「eiam 去依赖」独立议题，不在本决策范围。

## 4. 执行范围

- 物理删除 `D:\Haven\ecmdb\`（Duke1616 CMDB 后端 checkout）。
- 物理删除 `D:\Haven\ecmdb-web\`（Duke1616 平台前端 checkout）。
- `dev.ps1`：移除 ecmdb（Ports 8000/8097）与 ecmdb-web（Port 3333）两个服务条目；删后仅剩 4 个在役服务。
- `nginx-1.25.4/conf/vhost/nginx.dev.conf`：删除 `/api/cmdb` 路由；默认路由 `/` 改为 `301 /console/`。
- 决策文档：0001 标废弃（顶部横幅 + 状态字段指向本决策），本决策登记替代。

## 5. 不在范围

- `eiam` go.mod 对 `ecmdb v1.11.0/cryptox` 的库依赖（「eiam 去依赖」独立议题）。
- 重命名 `ecmdb-token-key` cookie / Mongo 用户 `ecmdb` / 共享 session 命名（eiam/e-cam-service 共享约定，重命名风险大于收益）。
- `e-cam-web`、`e-cam-service`、`haven-console` 的任何改动。

## 6. 风险

| 风险 | 缓解 |
|---|---|
| 删除后某服务编译/启动失败 | 已核查 eiam 走模块缓存、e-cam-service 零依赖；删后跑 `dev.ps1 status` + 编译验证 |
| 根路由重定向影响既有书签/入口 | 301 到 `/console/`（新统一入口），语义无损 |
| 两 checkout 有需保留的本地改动 | 删除前检查 git status；ecmdb 的 Phase 0/1 本地 commit 属被删对象自身，无需保留 |
| 遗漏的启动/编排引用（compose/CI） | 执行时全盘 grep `ecmdb-web`/`ecmdb`，逐个摘除 |

## 7. 关联

- 取代：[`0001-cmdb-consolidation-keep-ecmdb.md`](./0001-cmdb-consolidation-keep-ecmdb.md)（已废弃）。
- 提案：[`docs/proposals/remove-ecmdb-stack/proposal.md`](../proposals/remove-ecmdb-stack/proposal.md)。
- 0001 的 Phase 0 盘点结论、gRPC 死链路拆除记录等历史内容保留于 0001 原文，供审计。
