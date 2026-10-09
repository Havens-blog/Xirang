---
status: "completed"
started: "2026-09-22 14:09"
completed: "2026-09-22 14:24"
time_spent: "~15m"
---

# Task Record: 2 Redis 页重写对齐主机页紧凑标准

## Summary
Redis 云数据库页重写对齐主机页紧凑标准：page-top 紧凑骨架（标题 + tab + 5 内联状态标，后端全局口径）、Redis 领域动作（新建/重启/密码重置/参数配置，禁用 + tooltip）、字段搜索栏（回车加条件 + 条件 chips，仅映射 ListAssetsParams 支持的 name/status/provider/region）、去多选列、名称列等宽 ID 副行、行操作「详情」、根类名 redis-page（margin:-24px 抵消外层 padding）。紧凑样式从任务 1 RDS 页（源自 ecs index.vue 内联标准）移植。

## Changes

### Files Created
无

### Files Modified
- src/views/databases/redis/index.vue

### Key Decisions
- 紧凑样式整体内联进 redis/index.vue（沿用任务 1 RDS 先例，不抽共享 scss、不动 ecs/storage 文件）
- 状态标口径：监控异常/异常固定 0（后端无对应字段，主机页同款）；总数/运行中/已停止走 listRedisAssetsApi({limit:1,offset:0}) 三连查全局口径
- 搜索字段仅映射 ListAssetsParams 支持的维度（asset_name/asset_id 归 filters.keyword→name 参数）
- 保留 Redis 特有行为：getStatusClass/getStatusText 的 normal→运行中 映射（阿里云 Redis 状态为 Normal）、内存列 MB 单位、redis-column-settings localStorage 键不变
- 动作按 Redis 领域派生（密码重置/参数配置 替代 RDS 的 暂停/变更配置/扩容），均禁用 + tooltip「功能开发中」（F-C2/F-C3 主题 B 决策）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] Redis 页为紧凑标准骨架：page-top（标题 + tab + 内联状态标，竖分隔，无大统计卡）
- [x] 操作栏为 Redis 领域动作（新建/重启/密码重置/参数配置，不照搬主机电源操作）
- [x] 表格无默认多选列；字段搜索栏（回车加条件 + 条件 chips）可用
- [x] pnpm build 与 pnpm typecheck 通过；pnpm lint 相对基线零新增
- [x] 浏览器深/浅两主题：Redis 页渲染与标准一致，无控制台错误

## Notes
验证链：just e-cam-web-compile（vue-tsc -b）绿；vite build 绿；lint 3 错 12 警全为历史基线（src/api/types/{cost,statistics,task}.ts 等），redis 文件零新增；全量 vitest 902/902 绿（49 文件）。浏览器实测经 nginx :8888/cam/databases/redis：深/浅双主题计算样式断言（page-top 背景/边框、accent rgb(113,112,255)、无 col-checkbox、4 领域动作 disabled）+ 搜索 chip 增/删/过滤（tabtin 199→1）实测通过，控制台零错误。截图归档 screenshots/task2/（redis-dark.png / redis-light.png）。coverage 0% 为 redis 模块实测（无测试文件，重构任务 maintain 策略零下降；全量基线不受影响）。
