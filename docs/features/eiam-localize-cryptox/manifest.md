---
feature: "eiam-localize-cryptox"
created: "2026-10-09"
status: completed
mode: quick
---

# Feature (Quick): eiam-localize-cryptox

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/eiam-localize-cryptox/proposal.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | 内联移植 cryptox 到 eiam/internal/cryptox（含 round-trip 单测） | pending | 1-port-cryptox.md |
| 2 | 用原库生成 KAT 已知密文向量并固化字节兼容测试 | pending | 2-kat-compat.md |
| 3 | 切换 5 件 import、删除 go.mod ecmdb 依赖并全量验证 | pending | 3-swap-drop-verify.md |