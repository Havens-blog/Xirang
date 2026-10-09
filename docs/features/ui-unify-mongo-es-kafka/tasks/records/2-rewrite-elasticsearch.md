---
status: "completed"
started: "2026-09-23 15:11"
completed: "2026-09-23 15:27"
time_spent: "~16m"
---

# Task Record: 2 Elasticsearch 页重写对齐主机页紧凑标准

## Summary
Elasticsearch 页重写对齐主机页紧凑标准：单文件重写 src/views/middleware/elasticsearch/index.vue（221 行 → 约 1160 行），替换 storage.scss 薄骨架。页头 page-top（标题 + tab + 5 内联状态标，总数/运行中/已停止走 listElasticsearchAssetsApi({limit:1,offset:0}) 三连查全局口径，监控异常/异常固定 0）；操作栏 ES 领域动作 新建/索引管理/重启/扩容（索引管理证据 = EsDetailDrawer 索引 tab，扩容证据 = node_amount/disk_size 字段；无实现统一禁用 + tooltip「功能开发中」）；新增字段搜索栏（回车加条件 + chips，字段仅映射 ListAssetsParams 支持的 asset_name/asset_id→name、status、provider、region）；去多选列/全选/选中 N 块；名称列下 .cell-sub 等宽 ID 副行；行操作「详情」；首/末列 20px 对齐；样式按 ffe1b50 MongoDB 版同标准整版内联（根类 elasticsearch-page，margin:-24px）。等价保留：localStorage 键 elasticsearch-column-settings、H-03 route.query.search 预填、导出字段配置与文件名 Elasticsearch实例、AssetExportDialog :selected-ids="[]" 字面量、ES 自有状态映射（running/active、stopped/inactive）。组件 EsDetailDrawer/ColumnSettingsDialog 未改动。

## Changes

### Files Created
无

### Files Modified
- src/views/middleware/elasticsearch/index.vue

### Key Decisions
- 模板直接复用同 feature 任务 1 已验证的 MongoDB 版（ffe1b50）整版移植，仅换领域动作/列/API/键名，比回 ecs 取样式更贴近已验证基线
- 领域动作取 新建/索引管理/重启/扩容：索引管理由 EsDetailDrawer 的「索引」tab 佐证，扩容由 node_amount/disk_size 字段佐证；全部禁用 + tooltip（F-C2/F-C3 主题 B 决策延续）
- 列保持本页原 10 列派生（状态/版本/规格/节点数/存储/访问地址/平台/区域 + 隐藏 VPC/创建时间），未按 proposal 提示新造分片列——本页无分片字段，「从本页字段派生」为准
- 状态映射保留 ES 自有 running/active、stopped/inactive（等价保留，Redis normal→运行中 同理），未换成 MongoDB 版映射

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] Elasticsearch 页为紧凑标准骨架：page-top（标题 + tab + 内联状态标，无大卡片）
- [x] 操作栏为 ES 领域动作（从本页实际能力派生，勿照搬主机电源操作）
- [x] 表格无默认多选列；字段搜索栏（回车加条件 + chips）可用
- [x] pnpm build 与 pnpm typecheck 通过；pnpm lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
验证证据：vue-tsc -b 零错误；vite build 通过（仅历史 chunk-size 警告）；eslint 3 error + 12 warning = 历史源码基线零新增。vitest 全量 902 测试 49 文件全绿；elasticsearch 模块 runner 实测覆盖率 0%（无测试文件，与 RDS/Redis/MongoDB 先例一致，maintain 策略零下降）。浏览器实测 nginx :8888 真实数据：5 状态标���局口径（总数 35/运行中 34/已停止 0，本页 20 行）；状态搜索 chip「状态: 创建中」加条件后过滤为 1 条真实命中（jlc-prod-search-platform-es），删 chip 恢复 20/35；行「详情」开 EsDetailDrawer（.close-corner 关闭正常）；深/浅双主题计算样式断言过（首/末列 20px、状态点绿 6px、badge 靛紫 rgb(113,112,255)、等宽 ID 副行 JetBrains Mono）；console 零错误。截图归档 docs/features/ui-unify-mongo-es-kafka/screenshots/task2/（es-task2-light.png / es-task2-dark.png）。git 脏区 src/views/network/** 3 文件为并发会话 WIP，未动未提交。
