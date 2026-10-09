---
feature: "dashboard-cert-granularity"
created: "2026-09-22"
status: completed
mode: quick
---

# Feature (Quick): dashboard-cert-granularity

> 到期看板分级卡弃跳台账：主表重铸为证书粒度 + 页内筛选 + 探测行级聚合 + 孤儿与台账同构默认隐藏（双口径）。

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/dashboard-cert-granularity/proposal.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | 后端 GET /certs/dashboard 证书行化 + 服务端孤儿隐藏 + 双口径 | pending | [tasks/1-dashboard-cert-rows-hide.md](tasks/1-dashboard-cert-rows-hide.md) |
| 2 | 前端 DashboardTable 证书行渲染 + 隐藏切换 + 云选项派生 + 抽屉多 SAN 视图 | pending | [tasks/2-dashboard-table-cert-rows.md](tasks/2-dashboard-table-cert-rows.md) |
| 3 | 前端 OverviewCards 分级卡/特殊卡改页内筛选 + 豁免 + 移除跳台账 | pending | [tasks/3-overview-cards-inline-filter.md](tasks/3-overview-cards-inline-filter.md) |
| 4 | 联调验证：证书行三态实测 + 看板/台账动态相等 | pending | [tasks/4-verify-three-state-dashboard.md](tasks/4-verify-three-state-dashboard.md) |
| T-quick-doc-drift | Detect Spec Drift | pending | [tasks/quick-drift-detection.md](tasks/quick-drift-detection.md) |