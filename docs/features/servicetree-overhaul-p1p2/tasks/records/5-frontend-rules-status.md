---
status: "completed"
started: "2026-09-21 19:00"
completed: "2026-09-21 19:00"
time_spent: ""
---

# Task Record: 5 前端：规则列表真实状态呈现（去「-」遮掩）

## Summary
规则列表真实状态呈现：匹配数列改绑后端任务 1 透出的 last_match_count（last_executed_at 为 0 显示「未执行」而非「-」）；新增最近执行时间列（formatDateTime 格式化毫秒时间戳，未执行显示「未执行」）；目标节点列直绑 node_name（后端 4e196b5 已回填）；行内启用切换沿用既有 updateRuleApi 并新增 loading 防重复提交，失败由单向绑定自然回滚。仅改 rules/index.vue。

## Changes

### Files Created
无

### Files Modified
- src/views/service-tree/rules/index.vue

### Key Decisions
- Hard Rules 限定仅改 rules/index.vue，故后端新透出的 last_executed_at/last_match_count 以本地 interface BindingRuleWithStats 扩展承载，不改 src/api/types/service-tree.ts
- 匹配数/执行时间的「从未执行」判定以后端 last_executed_at 毫秒时间戳 0 值为准（commit 4e196b5 契约），不依赖旧 match_count 字段
- 行内启用切换未改为乐观更新：现有单向 :model-value 绑定在 API 失败时自然保持原状态（等效回滚），成功后 loadRules 刷新，满足 AC「调既有更新 API 后刷新」
- 首次提交 24d38bd 误卷入并发会话 staged 的 assets 组件删除文件，git reset --soft 后以显式 pathspec 重提为 628d192，并发会话暂存区原样保留

## Test Results
- **Tests Executed**: Yes
- **Passed**: 824
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 目标节点列显示真实节点名；状态列真实反映 enabled（含行内切换，调既有更新 API 后刷新）
- [x] 匹配数列：有执行记录显示数字，从未执行显示「未执行」（非「-」）；最近执行时间列新增
- [x] pnpm typecheck + vitest 全绿（并发在途错误除外）；git diff 仅 rules/index.vue

## Notes
静态验证：npx vue-tsc -b 退出码 0（无任何类型错误，含并发在途文件）。vitest 证据：全量套件 830 用例中 6 失败均在 RuleFormDialog.test.ts——该组件正被并发任务 4 在途修改（工作区 M 状态），与本任务改动无关（本任务 diff 仅 rules/index.vue，AC 3 约定「并发在途错误除外」）；剔除该在途文件后 npx vitest run --exclude **/RuleFormDialog.test.ts → 43 文件 824/824 全绿。分支确认为 main。git show --stat HEAD 仅 src/views/service-tree/rules/index.vue（1 file changed, 24 insertions, 4 deletions）。工作区其他并发在途文件（src/views/assets/index.vue 修改、AssetFilters/AssetTable 删除）未触碰未提交。提交：628d192f555eaba3bb3bfa3fd81484a91e1229ff。
