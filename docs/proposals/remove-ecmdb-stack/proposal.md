---
created: "2026-10-08"
author: "Haven"
status: Approved
intent: "cleanup"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: 移除 Duke1616 的 ecmdb 与 ecmdb-web

## Problem

Haven 平台「去他人依赖 / 100% 可控」的目标尚未闭环：Duke1616 的两个独立 checkout 仍占位在 `D:\Haven` 下——`ecmdb`（CMDB 后端）与 `ecmdb-web`（平台前端）。它们承载的能力已全部废弃，纯遗留。

### Evidence

- 统一入口（登录 / 身份治理 / 权限管理 / 工作台）已由 `haven-console` 接管（对 eiam 直连，2026-10 交付）。
- `e-cam-service` 运行时**零**对 ecmdb 服务的调用：全仓仅剩 `ecmdb-token-key`（共享 cookie 名）、Mongo 用户名 `ecmdb` 等命名残留，`internal/shared/middleware/ecmdb_auth.go` 是 eiam 共享 session 鉴权，不依赖 ecmdb 服务本身。
- `eiam` 依赖的是**已发布模块** `github.com/Duke1616/ecmdb v1.11.0`（仅 `pkg/cryptox`，经 go.sum/module 缓存解析），**非本地 `ecmdb` 目录**——删本地目录不打断 eiam 编译。
- `ecmdb-web` 剩余页面 `alert / change / cmdb / task / ticket` 已无人使用（用户 2026-10-08 确认「都废弃，直接全删」）。
- 两个 checkout 的 git remote 均指向 Duke1616 上游（`Duke1616/ecmdb`、`Duke1616/ecmdb-web`），是外来依赖源。

### Urgency

- `dev.ps1 start/status` 每轮都要拉起两个无用服务（ecmdb 走 CompileDaemon 编译、ecmdb-web 走 npm dev），拖慢本地开发。
- nginx 默认路由 `/`（`:8888` 根入口）仍被 ecmdb-web 占着，访问根路径落到废弃前端而非新入口。

## Proposed Solution

纯清理：物理删除两个 checkout，摘除 `dev.ps1` 的服务条目与 `nginx.dev.conf` 的路由，把 nginx 默认路由 `/` 改为 301 重定向到 `/console/`，并将「保留 ecmdb」的决策 0001 标为废弃。

### Innovation Highlights

无创新点——这是直接删除死代码/死依赖的标准清理。值得说明的是**依赖安全判定**：通过 go.mod 判定 eiam 依赖已发布模块而非本地目录，从而确认删除本地 checkout 零构建影响，这是本次删除的可行性前提。

## Requirements Analysis

### Key Scenarios

- 删除后 `dev.ps1 status` 只列 4 个在役服务（e-cam-service / eiam / e-cam-web / nginx），不再出现 ecmdb、ecmdb-web。
- 访问 `:8888/` → 301 → `/console/`；`/console/`、`/cam/`、`/api/iam/`、`/api/v1/*` 全部不受影响。
- e-cam-service / eiam / e-cam-web / haven-console 的启动与编译均不受删除影响。

### Non-Functional Requirements

- 不引入新的明文密钥或网络依赖。
- 删除为不可逆操作：执行前确认两目录无未提交且需保留的本地改动（ecmdb 曾有 Phase 0/1 本地 commit，均属被删对象的自变更）。

### Constraints & Dependencies

- `eiam` 保留，其 go.mod 对 `ecmdb v1.11.0/cryptox` 的**库**依赖不在本范围（属「eiam 去依赖」的独立议题）。
- `ecmdb-token-key` cookie 名、Mongo 用户 `ecmdb` 等**命名残留**不改动（eiam/e-cam-service 共享约定，重命名风险大于收益）。

## Alternatives & Industry Benchmarking

### Industry Solutions

去外来依赖、下线遗留系统是标准的「退役（decommissioning）」工程：先确认依赖面，再摘引用，再物理删除。本提案即按此顺序。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零工作量 | 偏离「100% 可控」目标，持续维护两个无用服务 + 根路由被占用 | Rejected：目标明确，遗留无价值 |
| 先迁剩余能力再删 | 迁移方案 | 保住 CMDB/工单等能力 | 能力已废弃，迁空壳纯属浪费 | Rejected：用户确认无使用方 |
| 只删前端留后端 | 分期 | 风险最小 | 留下半个死依赖，二次返工 | Rejected：后端同样无使用方 |
| **全删 + 摘引用 + 根路由重定向** | 退役工程 | 一次闭环，根入口理顺 | 删除不可逆（有本地 commit） | **Selected：能力已废弃，一次性清理最干净** |

## Feasibility Assessment

### Technical Feasibility

完全可行。已查证：`e-cam-service` 零 ecmdb 依赖；`eiam` 依赖已发布模块而非本地目录；`ecmdb-web`/`ecmdb` 独立 git 仓，无交叉引用会因删除而破坏。

### Resource & Timeline

小工程：删除 2 目录 + 改 `dev.ps1` 2 行 + 改 `nginx.dev.conf` 2 处路由 + 更新 1 个决策文档。单次提交可完成。

### Dependency Readiness

无外部 API/服务前置条件。前置是用户对「能力已废弃」的确认（已获），与「根入口指向 /console/」的确认（已获）。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 「ecmdb 被 eiam/e-cam-service 依赖，删不得」 | 代码核查（go.mod / import） | **推翻**：eiam 走已发布 v1.11.0 模块，e-cam-service 零 import，仅命名残留 |
| 「CMDB/工单能力可能还在用」 | 用户确认 | **推翻**：2026-10-08 确认「都废弃，直接全删」 |
| 「根路径还得留个页面」 | 用户确认 | **推翻**：改 301 到 /console/ 即可 |

## Scope

### In Scope

- 物理删除 `D:\Haven\ecmdb\` 目录（Duke1616 CMDB 后端 checkout）。
- 物理删除 `D:\Haven\ecmdb-web\` 目录（Duke1616 平台前端 checkout）。
- `dev.ps1`：移除 ecmdb（Ports 8000/8097）与 ecmdb-web（Port 3333）两个服务注册条目。
- `nginx-1.25.4/conf/vhost/nginx.dev.conf`：删除 `/api/cmdb` 路由；默认路由 `/` 改为 `301 /console/`。
- 决策文档：`docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md` 标废弃/新增替代决策。

### Out of Scope

- `eiam` 及其 go.mod 对 `ecmdb v1.11.0/cryptox` 的库依赖（「eiam 去依赖」独立议题）。
- 重命名 `ecmdb-token-key` cookie / Mongo 用户名 `ecmdb` / 共享 session 命名。
- `e-cam-web`、`e-cam-service`、`haven-console` 的任何改动。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 删除后某服务编译/启动失败 | Low | High | 已核查 eiam 走模块缓存、e-cam-service 零依赖；删后跑 `dev.ps1 status` + 编译验证 |
| 根路由重定向影响既有书签/入口 | Low | Medium | 301 到 `/console/`（新统一入口），语义无损 |
| 两个 checkout 有需保留的本地改动 | Low | Medium | 删除前检查 git status；ecmdb 的 Phase 0/1 本地 commit 属被删对象自身，无需保留 |
| 遗漏的启动/编排引用（compose/CI） | Medium | Low | 执行时全盘 grep `ecmdb-web`/`ecmdb`，逐个摘除 |

## Success Criteria

- [ ] `D:\Haven` 下不存在 `ecmdb\` 与 `ecmdb-web\` 目录。
- [ ] `dev.ps1 status` 不再列出 ecmdb、ecmdb-web，仅剩 4 个在役服务可正常 start/stop。
- [ ] `:8888/` 返回 301 到 `/console/`；`:8888/api/cmdb` 不再被路由（无 502/残留 upstream）。
- [ ] e-cam-service / eiam / e-cam-web / haven-console 四服务仍可正常编译、启动、访问。
- [ ] 决策 0001 标废弃，或新增「移除 ecmdb」替代决策并登记。

## Next Steps

- Proceed to `/write-prd` to formalize requirements