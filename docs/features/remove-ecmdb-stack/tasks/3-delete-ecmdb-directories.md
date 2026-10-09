---
id: "3"
title: "删除 ecmdb 与 ecmdb-web 两个 checkout 目录"
priority: "P0"
estimated_time: "0.5h"
complexity: "low"
dependencies: ["1", "2"]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 3: 删除 ecmdb 与 ecmdb-web 两个 checkout 目录

## Description
物理删除 `D:\Haven\ecmdb\`（Duke1616 CMDB 后端）与 `D:\Haven\ecmdb-web\`（Duke1616 平台前端）两个独立 git checkout。前置任务已摘除 dev.ps1/nginx 引用，本次为不可逆删除，须确认依赖安全。

## Reference Files
- `docs/proposals/remove-ecmdb-stack/proposal.md` — 依赖安全判定与成功标准（ref: Feasibility Assessment / Success Criteria）
- `ecmdb/`、`ecmdb-web/` — 整目录删除（独立 git 仓，origin 指向 Duke1616 上游）（ref: Scope）

## Acceptance Criteria
- [ ] `D:\Haven` 下不存在 `ecmdb\` 与 `ecmdb-web\` 目录
- [ ] eiam / e-cam-service 删除后仍可编译启动（eiam 依赖已发布模块 `ecmdb v1.11.0`，非本地目录；e-cam-service 零依赖）
- [ ] 全盘 grep 无指向 `ecmdb/`、`ecmdb-web/` 目录的启动/路由/构建引用残留

## Hard Rules
- 仅删除两个目录本身，不得改动 eiam、e-cam-service、e-cam-web、haven-console 的代码或配置

## Implementation Notes
- 删除前检查两个目录 `git status --short`：ecmdb 的 Phase 0/1 本地 commit 属被删对象自身的自变更，无需保留。
- 安全判定来源（提案 §Feasibility Assessment）：`eiam/go.mod` 依赖 `github.com/Duke1616/ecmdb v1.11.0`（发布模块，经 go module 缓存解析）；`e-cam-service` 无任何 `Duke1616/ecmdb` import（仅 `ecmdb-token-key` cookie 名、Mongo 用户名 `ecmdb` 等命名残留，不需删）。
- 验证：`go build ./...`（eiam、e-cam-service 各跑一次）确认编译不依赖本地 ecmdb 目录。