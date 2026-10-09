---
status: "completed"
started: "2026-09-23 16:58"
completed: "2026-09-23 17:31"
time_spent: "~33m"
---

# Task Record: 2 NAS 页重写对齐主机页紧凑标准

## Summary
NAS 页重写对齐主机页紧凑标准：page-top 紧凑骨架（标题 + tab + 5 内联状态标，全局口径三连查 95/95/0/0/0），运营指标卡保留为紧凑指标条（ops-strip 包裹、三态 warn/empty/data 原样保留、纵向 padding 14px），操作栏改 NAS 领域动作 新建/扩容/挂载点管理/标签（全禁用 + tooltip），新增字段搜索栏（7 字段：名称/云上ID/状态/平台/文件系统类型/协议类型/区域，回车加条件 + chips），去多选列与全选/已选，名称列加等宽 ID 副行，行操作改「详情」，首/末列 20px 对齐。状态选项改用后端精确匹配的真实小写值域（实测 95 条全为 running，旧下拉大写 Running 恒 0）。仅改 src/views/storage/nas/index.vue（544 行 → 约 1000 行）。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/views/storage/nas/index.vue

### Key Decisions
- 整版移植自 disk/index.vue（49e870d）紧凑标准，列保持本页原 11 列派生不造新列（指标表无吞吐字段，proposal 提示的吞吐不成列）
- 状态计数探测定案：/assets/nas 后端 status 精确匹配且真实值域为小写 running（实测 95 条全 running、大写 Running=0），状态选项与 fetchStatusCounts 均用小写值
- 运营指标条保留 nasMetrics.ts 判定逻辑与 NAS 自有四指标（总容量/已用/平均使用率/文件系统），warn/empty/data 三态模板原样保留
- 导出配置等价保留（字段/文件名 NAS文件系统/默认勾选），:selected-ids 传 [] 字面量（多选列移除后「已选中」保持禁用）；H-03 route.query.search 预填按 kafka 批同款保留
- nas-column-settings localStorage 键保留（写入方在 components/ColumnSettingsDialog.vue:74 未动），initColumnSettings 用 disk 批稳健版（剔除下线列/回填标签/补缺失列）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 7.4%

## Acceptance Criteria
- [x] NAS 页为紧凑标准骨架：page-top（标题 + tab + 5 内联状态标，无大卡片）
- [x] 运营指标卡保留为紧凑指标条（warn/empty 态不丢）
- [x] 操作栏为 NAS 领域动作；表格无默认多选列；字段搜索栏 + chips 可用
- [x] 首/末列 20px 对齐；build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
验证全部用 evaluate_script + take_snapshot（take_screenshot 本环境禁用）。实测证据：状态标 95/95/0/0/0（全局口径）；首列文本 left 240 == 页头标题 left 240、th/td 首末 20px；指标条 data 态 7.2TB/1.6TB/14.9%/5·截至2026-09-22，warn 态（琥珀图标/警示边框/重试按钮）与 empty 态用临时源码开关实测后精确还原；chip「状态:运行中」→95、「名称:外贸」→4、组合→4、清除还原；H-03 ?search=外贸 → Total 4；双主题 dark 8,9,10/15,16,17/25,26,27+glass0.03、light 247,248,248/白系+glass0.8；console 零错误。全量 vitest 902/902（49 文件）与基线一致；nasMetrics.test.ts 18/18；lint 恰为历史基线 3 error+12 warning 零新增。coverage 7.39% 为 nas 模块行覆盖（nasMetrics.ts 18 测覆盖、index.vue 无页面测试，与 disk 批同形态，maintain 零下降）。范围外发现（未修，供后续参考）：后端任务列表实际序列化的 failures 键为压平形式 accountid/errorcount/lasterror，而 nasMetrics.ts 的 NASCollectFailure 期望 account_id/error_count/last_error（Go executor 结构体 json tag 是对的，压平发生在任务持久化层），导致 warn 派生在本环境不可达——本任务前后行为一致非回归；diskMetrics.ts 疑似同模式。git 脏区 src/views/network/** 3 文件属并发会话 WIP，未触碰未提交。
