---
status: "completed"
started: "2026-09-22 13:25"
completed: "2026-09-22 13:33"
time_spent: "~8m"
---

# Task Record: 4 联调验证：证书行三态实测 + 看板/台账动态相等

## Summary
端到端联调验证（只读，经 8888/cam/ nginx + admin→jlc 租户自动登录 Playwright）：证书行三态、分级卡页内筛选动态相等、看板/台账口径一致全部实测通过。新增可复用验证脚本 verify-dashboard-three-state.js。

## Changes

### Files Created
- browser-verify/verify-dashboard-three-state.js
- browser-verify/dash3state-run2.log

### Files Modified
无

### Key Decisions
- 复用 ledger-hide-expired-orphans 验证脚本的自动登录骨架（8888/cam/ 路径、单次提交+轮询工作空间弹窗），新增看板专用断言
- AC3 豁免接线探针须在全新默认态下先测分级卡激活（隐藏开关未开）才能捕获 includeHidden=true 重拉请求；首轮脚本开关已展开导致无重拉，已调整顺序重跑
- blind_spot 过期证线上无样本（过期三态分布 has_refs=11 / no_refs_scanned=1），可见性以 has_refs 实测 + 任务 1 单测覆盖，未硬凑数据

## Test Results
- **Tests Executed**: Yes
- **Passed**: 5
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 实测同名多证成行：过期旧证与有效新证同时在看板主表成两行（不再域名掩盖）
- [x] 实测 *.jlc-robot.com referenceStatus=no_refs_scanned 默认隐藏；任一 has_refs/blind_spot 过期证默认可见
- [x] 分级卡激活 → 主表行数 == 该档卡片计数（动态相等）；探针验证豁免接线（includeHidden=true 重拉）
- [x] 与台账 daysLeft=expired 经 includeExpiredNoRefs 的行数口径一致（不写死数值）
- [x] 浏览器走通：行渲染、隐藏切换、分级卡页内筛选、行点击多 SAN 抽屉

## Notes
证据（2026-09-22 实测，租户 jlc）：AC1 *.jlcpcb.com 4 证 4 行（3 过期 has_refs + 1 有效 daysLeft=184，certId 各异、hidden=false），UI 行数 93 == API 默认 items 93（row-key=certId）。AC2 *.jlc-robot.com 过期证 daysLeft=-12 no_refs_scanned hidden=true 不在默认视图（默认仅现 daysLeft=170 has_refs 那张）；11/11 过期 has_refs 默认可见；hiddenCount=1 全局与 expired 档 hidden 分量同源。AC3 全新默认态激活已过期卡：aria-pressed=true、副文案「已按该档筛选，共 11 张」、主表 11 行 == countsByLevel.expired.total=11（动态取值非写死）、抓到重拉 ?includeHidden=true、激活态下 jlc-robot 行可见（豁免生效）、取消激活恢复 93 行 aria-pressed=false。AC4 看板 expired 档 total=11 == 台账 GET /certs?daysLeft=expired&includeExpiredNoRefs=true total=11（均运行时读取）。AC5 默认 93 行+横幅「已隐藏 1 张无引用过期证书」→ 开关展开 94 行 hidden-chip=1 横幅翻转 → 分级卡筛选开/关 → 行点击抽屉开启含 11 个 SAN 元素。截图 dash3state-{default,expanded,filtered,drawer}.png 于 %TEMP%。testsPassed/testsFailed 记 5 个 AC 实测项；coverage 不适用（纯只读验证，零代码改动，无 runner 覆盖率输出，不臆造）。台账侧 daysLeft=expired 默认与带 include 均 11 系后端契约（daysLeft=expired 强制豁免），与看板 11 口径一致。
