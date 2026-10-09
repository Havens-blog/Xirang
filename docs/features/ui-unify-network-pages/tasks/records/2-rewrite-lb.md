---
status: "completed"
started: "2026-09-24 11:59"
completed: "2026-09-24 12:27"
time_spent: "~28m"
---

# Task Record: 2 LB 页重写对齐主机页紧凑标准

## Summary
LB 页重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader → page-top（标题「负载均衡 (LB)」+ tab 全部/运行中/已停止 + 5 内联状态标 总数320/运行中314 blue/已停止6/公网101 orange/异常0 red，全量分页拉取本地计数）；删 disabled selection 列与 LbFilters.vue，操作栏 LB 领域动作（创建 LB/监听管理/后端服务器 禁用+tooltip + 同步实例 primary）；字段搜索栏 + chips（名称/云上ID/状态/类型/网络类型/平台/区域/VPC ID，实测有效参数）；列 名称/等宽ID副行、类型、状态点、VIP、监听协议/端口、后端实例数、平台色标、地域 + 6 可勾选隐藏列；新增 lb 本地 ColumnSettingsDialog（lb-column-settings）与 AssetExportDialog 导出；首/末列 20px；LbDetailDrawer 同源补 ACTIVE/INACTIVE 与 charge_type 实测值域。浏览器深/浅主题 + 全交互链路实测零控制台错误

## Changes

### Files Created
- src/views/network/lb/components/ColumnSettingsDialog.vue

### Files Modified
- src/views/network/lb/index.vue
- src/views/network/lb/components/LbDetailDrawer.vue

### Key Decisions
- 后端探值域（2026-09-24 全量 320 条实测）：status 仅 Active(144)/active(156)/ACTIVE(14)/inactive(6) 四个大小写变体且过滤精确区分大小写，LB_STATUS_LABELS 补 ACTIVE/INACTIVE 键（列表/抽屉/导出同源）
- load_balancer_type 实测 clb(155 众数)/alb(127)/nlb(21)/slb(17)，原映射缺 clb 会裸显英文 → 类型列补 CLB 标签
- address_type 实测 internet/Internet/intranet/Intranet 四变体，原页 ==='internet' 精确判断错标 69 条 → 大小写不敏感归一；status/address_type 搜索选项按变体逐一列出（诚实 UI）
- 监听协议/端口列：listeners 数组去重取前 2 组协议:端口 +N 余量，数组缺失（6 条）回退 listener_count；后端实例数取权威字段 backend_server_count（12 条与数组长度不符）
- charge_type/sort 实测后端忽略（total 不变）不作搜索字段；charge_type 值域 PayOnDemand/PrePay/POSTPAID_BY_HOUR 补入抽屉映射
- LB 原页无导出/自定义列能力，按 proposal 风险缓解「导出/自定义列收敛为 action-right 图标按钮（主机同款）」新增，列设置键 lb-column-settings 无历史漂移
- 领域动作全禁用时 primary 主色给唯一已接线的同步实例（诚实 UI 优先于原型像素，EIP 先例同款）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 58.5%

## Acceptance Criteria
- [x] LB 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标，按 LB 状态域派生）
- [x] 表格删除 disabled selection 列；操作栏为 LB 领域动作（创建 LB/监听管理/后端服务器 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [x] 列按 LB 领域：名称/ID(等宽ID)、类型(SLB/ALB/NLB)、状态(状态点)、VIP、监听协议/端口、后端实例数、平台色标、地域
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
改动仅 src/views/network/lb/** 三文件：index.vue 整版移植 eip(67ac656) 骨架、LbFilters.vue 删除、LbDetailDrawer.vue 最小同源补齐（+ACTIVE/INACTIVE/+charge_type 三值）、ColumnSettingsDialog.vue 新增。验证：vue-tsc -b 零错；pnpm build 通过（仅既有 chunk 警告）；lint 3错12警=基线持平且 lb 文件零命中；vitest 902/902 绿（全仓 coverage 58.5%）；浏览器 :8888 实测状态标计数 314/6/101 与探值域一致、搜索 chip 状态=inactive→6 行、抽屉/列设置/导出对话框、深浅主题切换均零 console error
