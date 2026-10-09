---
id: "1"
title: "dev.ps1 移除 ecmdb 与 ecmdb-web 服务条目"
priority: "P1"
estimated_time: "0.5h"
complexity: "low"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 1: dev.ps1 移除 ecmdb 与 ecmdb-web 服务条目

## Description
从 `dev.ps1` 的 `$Services` 注册数组移除 ecmdb（Ports 8000/8097）与 ecmdb-web（Port 3333）两个废弃服务，使本地一键脚本不再拉起/管理它们。这是「移除 ecmdb 栈」的摘引用第一步。

## Reference Files
- `docs/proposals/remove-ecmdb-stack/proposal.md` — 完整需求与成功标准（ref: Scope / Success Criteria）
- `dev.ps1` — 删除 `$Services` 中 `ecmdb`（8000/8097）与 `ecmdb-web`（3333）两条注册项（ref: Proposed Solution）

## Acceptance Criteria
- [ ] `dev.ps1` 的 `$Services` 数组不再含 ecmdb、ecmdb-web 条目
- [ ] `dev.ps1 status` 输出仅列出 e-cam-service / eiam / e-cam-web / nginx 四个服务
- [ ] 其余命令（start/stop/restart/build-console/console-dev）行为不变

## Implementation Notes
- 目标条目：`dev.ps1` 第 25 行 `@{ Name = 'ecmdb'; ... }` 与第 28 行 `@{ Name = 'ecmdb-web'; ... }`。
- 删除条目后 `$Services` 数组长度从 6 缩到 4，`Get-Target` 的「未知服务」报错列表随之自动更新。
- 验证：`.\dev.ps1 status` 确认只列 4 服务；`.\dev.ps1 start` 不尝试拉起 ecmdb/ecmdb-web。