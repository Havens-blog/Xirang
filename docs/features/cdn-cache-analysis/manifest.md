---
feature: "cdn-cache-analysis"
created: "2026-09-21"
status: completed
mode: quick
---

# Feature (Quick): cdn-cache-analysis

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/cdn-cache-analysis/proposal.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | cdn-cache 缓存命中率规则引擎(双命中率/健康档位/gap 归因/结构化优化项,纯函数+单测) | pending | tasks/1-cdn-cache-rule-engine.md |
| 2 | 后端接口 POST /logs/cache-analyze(当前窗并发聚合+前窗 1 帧,调规则引擎,降级标注,预留 summary) | pending | tasks/2-cache-analyze-api.md |
| 3 | 前端 CDN Tab 内嵌「缓存分析」卡(手动触发/双命中率总览/域名排行/未命中 URI/状态码分布/优化项复制/AI 解读区/前窗对比降级) | pending | tasks/3-frontend-cache-card.md |
| T-quick-doc-drift | Detect Spec Drift | pending | tasks/quick-drift-detection.md |
