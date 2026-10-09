---
status: "completed"
started: "2026-09-24 13:23"
completed: "2026-09-24 13:48"
time_spent: "~25m"
---

# Task Record: 5 ENI 页重写对齐主机页紧凑标准

## Summary
ENI 页重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader 壳 → page-top（标题「弹性网卡 (ENI)」+ tab 全部/使用中/可用 + 5 内联状态标 总数/使用中(blue)/可用/绑定率/异常(red)，原「使用中/绑定率」StatCard 统计并入状态标）；删 disabled selection 列；操作栏 创建网卡/绑定实例/解绑（禁用+tooltip）+ 同步实例（primary，唯一已接线动作）；内联筛选器 → 字段搜索栏 + chips（7 个后端实测有效字段）；表格改手写 data-table（首列 网卡ID/名称 等宽副行 + 私有IP/状态点/绑定实例/MAC地址/类型/平台色标/地域，VPC/云账号/安全组数/创建时间 为隐藏列）；新增 ColumnSettingsDialog（键 eni-column-settings，本页无历史键零漂移）；导出/EniDetailDrawer 逐字保留。先全量探查后端（2026-09-24，5814 条）：状态值域仅 in_use(4799)/available(1014)/deleting(1) 三值、类型值域 11 个（huawei=Secondary、aliyun=Primary/Secondary/Bond/Trunk/Member、volcano=primary/secondary/branch/nlb/vpclink/transit_router）、provider=huawei/aliyun/volcano；参数有效性实测：name（LIKE asset_name+asset_id）/status/provider/type（均精确区分大小写）/vpc_id/instance_id 有效，ip_address/mac_address/zone 后端忽略已排除。类型映射由原二元判断补齐 11 值域（导出同源对齐）。

## Changes

### Files Created
- src/views/network/eni/components/ColumnSettingsDialog.vue

### Files Modified
- src/views/network/eni/index.vue

### Key Decisions
- 状态标全局口径（全量分页拉取计数，不受筛选影响），deleting 为过渡态不计入任一状态标、列表状态点如实呈现；异常标=未知状态族防御性计数（当前值域恒 0）
- 类型列/导出取值对齐全量实测 11 值域（原 Primary→主网卡否则辅助网卡 的二元判断会误标 Bond/Trunk/Member），未知值 labelOfLenient 原样透出
- 搜索字段按实测排除 ip_address/mac_address/zone（传参 total 不变）；status/type/provider 选项按大小写变体逐一列出（后端精确匹配）
- 主色给唯一已接线的「同步实例」而非禁用的领域动作（诚实 UI，同 EIP/LB 先例）
- 原多选列未接线（selectedIds 恒空）随列删除，导出 selected-ids 传 [] 保持「已选中」禁用（零漂移）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 58.5%

## Acceptance Criteria
- [x] ENI 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标 总数/使用中/可用/绑定率/异常）
- [x] 表格删除 disabled selection 列；操作栏为 ENI 领域动作（创建网卡/绑定实例/解绑 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [x] 列按 ENI 领域：名称/ID(等宽ID)、私有 IP、状态(状态点)、绑定实例、MAC/类型、平台色标、地域
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
验证证据：vue-tsc -b 零错误；pnpm build 通过；eslint 全量 15 problems（3 错 12 警）= 基线零新增；vitest 全量 902/902；coverage 58.5%（与基线持平，eni 无测试文件）。浏览器 :8888 实测：状态标 5814/4799/1014/83%/0 与后端探查值一致；搜索 chips（状态→使用中）total 4799；自定义列勾选 VPC 保存后表头即时生效且恢复默认正常；详情抽屉四个 tab 正常；深/浅主题切换（pinia setTheme）后 computed style 与 a11y 快照均符合标准、console 零消息。prettier --check 对 3 个 eni 文件有 warn，但已提交的 eip/vswitch 同款文件同样被标（基线如此，未执行 --write 以免偏离先例骨架风格）。仅修改 src/views/network/eni/** 两个文件，EniDetailDrawer.vue 未动；未触碰工作区其他并发会话 WIP 脏区。Write 产出的 1 处 U+FFFD 已字节级修复并回读校验（grep 归零）。
