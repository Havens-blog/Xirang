---
status: "completed"
started: "2026-09-22 10:18"
completed: "2026-09-22 10:32"
time_spent: "~14m"
---

# Task Record: 1 迁移 7 个视图级 scss 的旧蓝字面量为靛紫

## Summary
按映射表把 7 个视图级 scss 的 26 行旧蓝字面量机械迁移为靛紫：#2563eb→#5e6ad2（主按钮 hover 暗化语义保留）、rgba(59,130,246,X)→rgba(113,112,255,X)（X 不变、间距风格保持）；#3b82f6→#7170ff 映射行在当前代码中零命中（no-op）。不改选择器结构、不新增 token、不改任何行为。构建 + 浏览器双主题实测验收：cmdb/models 旧蓝计算样式命中深/浅两主题均归零（深色基线 35），其余 6 页抽查无回归、无控制台错误。

## Changes

### Files Created
无

### Files Modified
- src/views/cmdb/models/styles/models.scss
- src/views/cmdb/instances/styles/instances.scss
- src/views/cmdb/relations/styles/relations.scss
- src/views/cmdb/topology/styles/topology.scss
- src/views/compute/ecs/styles/instances.scss
- src/views/storage/styles/export-dialog.scss
- src/views/storage/styles/column-settings.scss

### Key Decisions
- 逐行清单以任务文件 per-file 表为权威（合计 26 行；任务 prose 的「25 行」为计数笔误），grep 实扫 26 行与表完全一致后一次 sed 批量迁移
- #3b82f6 映射行经实扫确认当前 7 文件无命中，不虚构替换
- rgba 替换保留原有带/不带空格两种写法的风格（59, 130, 246 → 113, 112, 255 与 59,130,246 → 113,112,255 各自对应）
- git status 非净（src/views/network 3 文件为并发会话 WIP，与本任务 7 文件零交集），按项目惯例不阻塞、不触碰

## Test Results
- **Tests Executed**: Yes
- **Passed**: 93
- **Failed**: 0
- **Coverage**: 3.0%

## Acceptance Criteria
- [x] 按映射表迁移全部文件行：#3b82f6→#7170ff、rgba(59,130,246,X)→rgba(113,112,255,X)、#2563eb→#5e6ad2，不改选择器结构、不新增 token
- [x] grep -rn "3b82f6|2563eb|59,\s*130,\s*246" src/views --include="*.scss" 零命中（排除 __tests__）
- [x] pnpm build 与 pnpm typecheck 均通过
- [x] 浏览器 :8888/cam/ 深/浅两主题：cmdb/models 旧蓝计算样式命中 = 0（深色基线 35），抽查其余 6 页无渲染回归、无控制台错误

## Notes
coverage 2.96% 为 --coverage.include 覆盖 src/views/storage/**+compute/** 全目录的信息性数值（v8 对 .scss 天然不计）；本任务 7 个改动文件均为 scss，无 JS 覆盖率语义，相对基线零变化。lint 3 错/12 警，基线为 3 错/16 警，零新增。旧蓝断言相关 4 个测试文件（MainLayout/alert events/assets/dashboard）均为「不含旧蓝」负向断言，迁移后自然增强，全部 PRESERVE 未改动。浏览器实测经 nginx :8888/cam/ 真实登录态（chrome-devtools MCP getComputedStyle 全 DOM 扫描 rgb(59,130,246)/rgb(37,99,235)/rgba 两系），测毕主题已恢复 dark。
