---
status: "completed"
started: "2026-09-23 14:52"
completed: "2026-09-23 15:09"
time_spent: "~17m"
---

# Task Record: 1 MongoDB 页重写对齐主机页紧凑标准

## Summary
MongoDB 页（252 行）重写为「主机页紧凑标准」：page-top（标题+tab+5 内联状态标，全局后端口径 3 连查）、MongoDB 领域动作（新建/重启/扩容/备份，禁用+tooltip「功能开发中」，主题 B 决策）、字段搜索栏（回车加条件+chips，仅映射 ListAssetsParams 的 name/status/provider/region）、去多选列/全选/选中N、名称列 .cell-sub 等宽 ID 副行、行操作「详情」、根类名 database-page→mongodb-page（margin:-24px 同 ecs .vm-page 法）、首/末列 20px 对齐。localStorage 键 mongodb-column-settings、H-03 query.search 预填、导出配置等价保留；AssetExportDialog :selected-ids="[]" 字面量。仅改 index.vue 一个文件，组件子目录与令牌层未动。

## Changes

### Files Created
无

### Files Modified
- src/views/databases/mongodb/index.vue

### Key Decisions
- 领域动作取 新建/重启/扩容/备份（备份有本页抽屉「备份」tab 佐证），沿用 F-C2/F-C3 主题 B 禁用+tooltip 先例，不照搬主机电源操作
- 列保持本页原 11 列派生（状态/版本/类型/规格/存储/副本集/平台/区域 + 隐藏 VPC/计费方式/创建时间），不加新列
- 状态标改后端全局口径（listMongoDBAssetsApi({limit:1,offset:0}) 三连查，fetchInstances 成功后同步刷），替换原当前页 computed
- 紧凑样式按 ui-unify-rds-redis-pilot 任务 1 同法整版内联（不抽共享 scss），localStorage 键保留 mongodb-column-settings

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] MongoDB 页为紧凑标准骨架：page-top（标题 + tab + 内联状态标，无大卡片）
- [x] 操作栏为 MongoDB 领域动作（从本页实际能力派生，勿照搬主机电源操作）
- [x] 表格无默认多选列；字段搜索栏（回车加条件 + chips）可用
- [x] pnpm build 与 pnpm typecheck 通过；pnpm lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
coverage 0 为 mongodb 模块实测值：该模块无测试文件（与 RDS/Redis 先例一致，重构任务 maintain 策略零下降）；全量 vitest 902/902 绿（49 文件，与基线持平）。lint 3 error（src/api/types/{cost,statistics,task}.ts 历史）+ 12 源码警告 = 基线（第 13 条警告在 coverage/ 运行产物，非源码）。浏览器实测经 nginx :8888（dist 构建）：深 rgb(8,9,10)/浅 rgb(247,248,248) 背景走令牌、双主题靛紫强调色 rgb(113,112,255)、状态点 rgb(16,185,129)、首/末列 20px 计算样式断言全过、搜索真实过滤（ali-jlc-prod-overseas→1/1）、行点击开抽屉且 .close-corner 关闭、console 零错误、主题经 pinia setTheme 实切并恢复 dark。截图归档 docs/features/ui-unify-mongo-es-kafka/screenshots/task1/。git 基线：src/views/network/** 3 文件为并发会话 WIP，未触碰。
