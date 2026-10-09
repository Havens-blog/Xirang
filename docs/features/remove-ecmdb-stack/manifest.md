---
feature: "remove-ecmdb-stack"
created: "2026-10-08"
status: completed
mode: quick
---

# Feature (Quick): remove-ecmdb-stack

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/remove-ecmdb-stack/proposal.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | dev.ps1 移除 ecmdb 与 ecmdb-web 服务条目 | pending | 1-remove-devps1-ecmdb-entries.md |
| 2 | nginx 摘除 /api/cmdb 路由并把默认路由 / 改为 301 /console/ | pending | 2-rewire-nginx-default-route.md |
| 3 | 删除 ecmdb 与 ecmdb-web 两个 checkout 目录 | pending | 3-delete-ecmdb-directories.md |
| 4 | 决策 0001 标废弃并登记移除 ecmdb 的新决策 | pending | 4-supersede-cmdb-decision.md |
| T-quick-doc-drift | Detect Spec Drift | pending | — |
| T-review-doc | Review Documentation Quality | pending | — |