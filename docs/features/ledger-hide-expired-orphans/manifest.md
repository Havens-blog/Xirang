---
feature: "ledger-hide-expired-orphans"
created: "2026-09-21"
status: completed
mode: quick
---

# Feature (Quick): ledger-hide-expired-orphans

> 台账默认隐藏「已过期且 referenceStatus=no_refs_scanned」的证书（可切换查看），有引用过期证/盲点保守显示；看板跳转口径一致。

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/ledger-hide-expired-orphans/proposal.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | 后端 GET /certs 排除无引用过期证书 + 双口径返回 | pending | [tasks/1-backend-exclude-expired-unref.md](tasks/1-backend-exclude-expired-unref.md) |
| 2 | 前端台账默认隐藏 + 行内提示切换 + 状态机 | pending | [tasks/2-frontend-ledger-hide-toggle.md](tasks/2-frontend-ledger-hide-toggle.md) |
| 3 | 联调验证：三态实测 + 看板/台账动态相等 | pending | [tasks/3-verify-three-state-and-dynamic-equality.md](tasks/3-verify-three-state-and-dynamic-equality.md) |