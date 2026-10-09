---
feature: "eiam-localize-etask"
created: "2026-10-09"
status: completed
mode: quick
---

# Feature (Quick): eiam-localize-etask

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/eiam-localize-etask/proposal.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | 新建 internal/grpcx（Server + JWT 拦截器 + 单测） | pending | 1-grpcx-jwt-package.md |
| 2 | 接线替换（ioc/grpc.go + app.go + 删 registry.go）并重跑 wire | pending | 2-rewire-ioc-drop-registry.md |
| 3 | 删 etask 依赖、简化 config 并全量验证 | pending | 3-drop-etask-verify.md |