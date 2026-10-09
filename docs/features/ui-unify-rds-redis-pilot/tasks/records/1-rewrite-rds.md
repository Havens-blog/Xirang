---
status: "completed"
started: "2026-09-22 13:41"
completed: "2026-09-22 14:08"
time_spent: "~27m"
---

# Task Record: 1 RDS 页重写对齐主机页紧凑标准

## Summary
RDS 页重写对齐主机页紧凑标准：page-top（标题+tab+5 内联状态标：总数/运行中/已停止/监控异常/异常，无大统计卡）；操作栏换为 RDS 领域动作（新建/重启/暂停/变更配置/扩容，未实现按主题 B 决策禁用+tooltip，不照搬主机电源操作）；新增主机页同款字段搜索栏（回车加条件+条件 chips，搜索字段仅覆盖 ListAssetsParams 支持的 名称/云上ID/状态/平台/区域）；表格去除 col-checkbox 多选列与「全选/选中 N」分页左侧区块，名称列下增等宽 ID 副行，规格/存储/连接/端口/VPC 列等宽字体，行操作改「详情」链接；根类名 database-page→rds-page，样式改为自包含（对齐 ecs index.vue 内联样式，弃用共享 storage.scss 导入）。RdsDetailDrawer/ColumnSettingsDialog/AssetExportDialog 契约与 H-03 query.search 预填、列设置 localStorage 键均保留。浏览器实测（nginx :8888 真实登录态）：深/浅双主题渲染与原型一致，状态点/accent-green/等宽 ID/无多选列/搜索 chips 增删全部实测通过，全程控制台零错误；截图归档 docs/features/ui-unify-rds-redis-pilot/screenshots/task1/。

## Changes

### Files Created
无

### Files Modified
- src/views/databases/rds/index.vue

### Key Decisions
- 操作栏 5 个领域动作全部禁用+tooltip「功能开发中」：沿用主机页 F-C2/F-C3 主题 B 决策先例，本页无对应实现，避免点击零反馈
- 搜索字段仅派生自 ListAssetsParams 实际支持的筛选维度（name/status/provider/region），不提供后端不支持的 engine/instance_type 等字段以免假筛选
- 状态标计数改为后端全局口径（listRDSAssetsApi limit:1 三连查，主机页 fetchStatusCounts 同款），替换原当前页 computed 口径
- 样式自包含（内联 scoped，移植 ecs 紧凑样式子集），Hard Rule 限定仅改 rds/** 故不抽共享 scss、不动 ecs/storage 现有文件
- 监控异常/异常两标按主机页先例固定为 0（后端无对应状态字段），浏览器验证通过

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] RDS 页为紧凑标准骨架：page-top（标题 + tab + 内联状态标，竖分隔，无大统计卡）
- [x] 操作栏为 RDS 领域动作（新建/重启/暂停/变更配置/扩容，不照搬主机电源操作）
- [x] 表格无默认多选列；字段搜索栏（回车加条件 + 条件 chips）可用
- [x] pnpm build 与 pnpm typecheck 通过；pnpm lint 相对基线零新增
- [x] 浏览器深/浅两主题：RDS 页渲染与原型一致（状态点/平台色标/等宽 ID），无控制台错误

## Notes
coverage 0% 为 vitest v8 runner 对 src/views/databases/rds/** 实测报告值——该模块重构前后均无测试文件（本任务为 coding.refactor，覆盖率策略 maintain，基线即 0%，零下降）。全量 902 测试通过验证零行为回归。浏览器证据：nginx :8888/cam/databases/rds 真实登录态，深/浅主题计算样式断言（status-dot.running=rgb(16,185,129)、stat-num blue=rgb(113,112,255)、cell-sub 为 JetBrains Mono、col-checkbox=0、5 个禁用领域按钮），搜索条件 chip 增/删实测，抽屉打开实测，控制台零 error/warn。
