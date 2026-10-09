---
feature: "waf-flood-diagnosis"
created: "2026-09-19"
status: completed
mode: quick
---

# Feature (Quick): waf-flood-diagnosis

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/waf-flood-diagnosis/proposal.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | WAF 被刷判定规则引擎(风险分/等级/攻击类型/措施映射,纯函数+单测) | pending | 1-rule-engine.md |
| 2 | 后端诊断接口 POST /logs/diagnose(当前窗+前窗聚合,调规则引擎,预留 summary) | pending | 2-diagnose-api.md |
| 3 | 前端 WAF Tab 内嵌「流量诊断」卡(触发/展示/降级) | pending | 3-frontend-diagnose-card.md |
| T-quick-doc-drift | Detect Spec Drift | pending | quick-drift-detection.md |
