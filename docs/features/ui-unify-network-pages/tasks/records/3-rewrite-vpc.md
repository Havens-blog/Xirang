---
status: "completed"
started: "2026-09-24 12:29"
completed: "2026-09-24 12:55"
time_spent: "~26m"
---

# Task Record: 3 VPC 页重写对齐主机页紧凑标准

## Summary
VPC 专有网络页（606 行）重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader 壳 → page-top（标题「专有网络 (VPC)」+ 全部/本地IDC/公有云 tab + 5 内联状态标 总数/可用(blue)/默认VPC/创建中(orange)/异常(red)），删 disabled selection 死列，操作栏为 VPC 领域动作 创建 VPC/子网管理（禁用+tooltip）+ 同步实例（唯一已接线动作落主色），字段搜索栏（名称/云上ID/状态/平台/区域 5 字段）+ chips 取代 VpcFilters，导出/自定义列收进 action-right 圆钮，VpcDetailDrawer 原样保留（安全组/IP子网子查询能力不受影响），VpcFilters.vue 删除。整版移植 eip(67ac656)/lb(5ecfd54) 骨架。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/views/network/vpc/index.vue
- e-cam-web/src/views/network/vpc/components/ColumnSettingsDialog.vue
- e-cam-web/src/views/network/vpc/components/VpcFilters.vue (deleted)

### Key Decisions
- 全量探查 96 条实测：status 仅 Available(47)/available(27)/OK(22) 三个大小写变体、全为成功族 → 状态映射补 OK（原缺键会裸显英文），创建中/异常两标为防御性 0 计数；默认域标按 attributes.is_default 派生（实测 true 26 条）
- 后端参数实测（fetch 传参看 total）：name 有效（LIKE 同时匹配 asset_name/asset_id，中段子串均命中）、status/provider 精确且大小写敏感（Available→47/available→27/OK→22/ok→0，选项须逐变体列出）、region 有效；cidr_block/is_default 后端忽略（传参 total 不变）不作搜索字段
- 列 key 与历史 vpc-column-settings 零漂移（12 键全保留，历史保存设置零迁移）；标签按 AC 调整（IP数→子网数、区域→地域、共享机器→共享服务器）；默认可见列 8 = AC 领域口径（CIDR/状态/子网数/共享服务器/默认VPC/网段域/平台/地域），IPv6/允许外网访问/云账号/创建时间保留为可勾选隐藏列
- 实测 enable_internet_access/network_domain/shared_machines 属性 96 行全缺失（列位按 AC 保留、缺值显 -，allow_internet 不再把未知兜底成「禁用」）；account_name 列取值修正为 cloud_account_name（原 account_name 键全量缺失恒显 -）
- VpcDetailDrawer 接口未动（ecs 抽屉与 vswitch 页也在引用该组件）；ColumnSettingsDialog 的 defaultColumns 与 index.vue 同步对齐（恢复默认口径同源）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 58.5%

## Acceptance Criteria
- [x] VPC 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标，按 VPC 状态/默认域派生）
- [x] 表格删除 disabled selection 列；操作栏为 VPC 领域动作（创建 VPC/子网管理 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [x] 列按 VPC 领域：名称/ID(等宽ID)、CIDR、状态(状态点)、子网数、共享服务器、默认/网络域、平台色标、地域
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
验证：just e-cam-web-compile（vue-tsc -b）零错误；pnpm build 通过（仅既有 chunk 体积警告）；just e-cam-web-fmt 与 eip/lb 标杆一致（vpc/index.vue 与标杆文件同为 prettier 不洁存量口径，未对自身文件跑 --write）；eslint 3 错 12 警与基线持平且 vpc 文件零命中；全量 vitest 902/902 绿，coverage 58.5%（runner 实报，vpc 模块无测试文件）。浏览器验收（:8888，SPA 内路由进入 /cam/network/vpc）：状态标 96/96/26/0/0 与探值域一致；搜索 chip 状态=正常(OK)→22 条、名称:法兰克福→4 条、云上ID 中段子串→1 条，增删清均实测；行点击开抽屉（详情/安全组/IP子网/路由表 tabs 在）、导出/列设置/同步对话框均开合正常；深浅主题 token round-trip 一致（page-top 白↔rgb(15,16,17)、强调色 rgb(113,112,255) 不变）；控制台零 error/warn。工作区遗留：VpcDetailDrawer.vue 存在并发会话未提交 WIP（eslint 注释两行），本次未触碰未纳入提交。
