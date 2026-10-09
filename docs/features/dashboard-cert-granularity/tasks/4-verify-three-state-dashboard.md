---
id: "4"
title: "联调验证：证书行三态实测 + 看板/台账动态相等"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies: ["2", "3"]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 4: 联调验证：证书行三态实测 + 看板/台账动态相等

## Description

任务 1/2/3 上线后做端到端联调验证（复用浏览器自动登录 Playwright 流程：admin/12345678 → 租户 jlc → 走 8888/cam/）：实测证书行三态正确性（no_refs_scanned 被隐藏、has_refs/blind_spot 恒可见）、分级卡激活行数与卡片计数动态相等、与台账 daysLeft=expired 行数口径一致。

## Reference Files
- `docs/proposals/dashboard-cert-granularity/proposal.md` — Success Criteria
- `e-cam-web/src/views/cert/dashboard/index.vue` — 看板页 (ref: ### Success Criteria)
- `D:\Haven\browser-verify\` — 既有自动登录与 diag 探针脚本 (ref: ### Success Criteria)

## Acceptance Criteria
- [ ] 实测同名多证成行：过期旧证与有效新证同时在看板主表成两行（不再域名掩盖）
- [ ] 实测 `*.jlc-robot.com` referenceStatus=no_refs_scanned 默认隐藏；任一 has_refs/blind_spot 过期证默认可见
- [ ] 分级卡激活 → 主表行数 == 该档卡片计数（动态相等）；探针验证豁免接线（includeHidden=true 重拉）
- [ ] 与台账 `daysLeft=expired` 经 includeExpiredNoRefs 的行数口径一致（不写死数值）
- [ ] 浏览器走通：行渲染、隐藏切换、分级卡页内筛选、行点击多 SAN 抽屉

## Hard Rules
无（验证任务；只读，不改生产数据）。

## Implementation Notes
- 复用 browser-verify 自动登录与 diag 探针；浏览器必须走 8888/cam/（nginx），5173 直连 auth 必挂。
- 若实测 blind_spot 而非 no_refs_scanned（扫描范围变化），记录证据回改任务 1 判定与提案 Evidence，不硬凑。