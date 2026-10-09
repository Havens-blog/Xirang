---
id: "3"
title: "联调验证：三态实测 + 看板/台账动态相等"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies: [2]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 3: 联调验证：三态实测 + 看板/台账动态相等

## Description

任务 1/2 上线后做端到端联调验证（复用 Playwright 自动登录浏览器验证流程）：实测被隐藏行的引用三态真实性（Evidence 判定依据），确认默认视图计数口径闭合、daysLeft=expired 时看板卡跳转与台账行数动态相等。

## Reference Files
- `docs/proposals/ledger-hide-expired-orphans/proposal.md` — Success Criteria / Evidence (ref: ### Success Criteria)
- `e-cam-web/src/views/cert/ledger/index.vue` — 台账默认视图 (ref: ### Success Criteria)
- `e-cam-web/src/views/cert/dashboard/index.vue` — 看板「已过期」卡 (ref: ### Success Criteria)

## Acceptance Criteria
- [ ] 实测 `*.jlc-robot.com` 的 referenceStatus=no_refs_scanned（非 blind_spot），确认其为当前唯一可隐藏行；若实测为 blind_spot 则回改任务 1 判定与 proposal Evidence
- [ ] 台账默认视图不含无引用过期行，分页 total 与可见行数一致；展开后行数与 hiddenCount 一致
- [ ] 任一 `blind_spot` / 引用状态未知的过期证书在默认视图可见（若有）
- [ ] /certs?daysLeft=expired 筛选行数 = 看板「已过期」卡计数（动态相等，不写死数值）

## Hard Rules
无（验证任务）。

## Implementation Notes
- 复用 browser-verify 自动登录（admin/12345678 → 租户 jlc → 台账）与 diag 探针模式。
- eval 残留 #6「过期新近度宽限窗口」决策记录于此（本次不做，观察噪音累积速率后另行评估）。