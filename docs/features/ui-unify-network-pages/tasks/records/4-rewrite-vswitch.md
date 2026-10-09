---
status: "completed"
started: "2026-09-24 12:58"
completed: "2026-09-24 13:21"
time_spent: "~23m"
---

# Task Record: 4 VSwitch 页重写对齐主机页紧凑标准

## Summary
VSwitch 页重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader 壳，改为 page-top（标题「交换机/子网 (VSwitch)」+ tab 全部/默认子网/非默认子网 + 5 内联状态标）；删 disabled selection 列；操作栏为子网领域动作（创建子网 禁用+tooltip + 同步实例 primary）；新增字段搜索栏 + chips（搜索字段按实测筛选）；新增导出（AssetExportDialog）与自定义列（ColumnSettingsDialog，键 vswitch-column-settings）；保留 handleVpcClick VPC 跳转、VSwitchDetailDrawer、VpcDetailDrawer 与 query.vpc_id 跳入预填（改为 chip 呈现）；删除 VSwitchFilters.vue。

## Changes

### Files Created
- src/views/network/vswitch/components/ColumnSettingsDialog.vue

### Files Modified
- src/views/network/vswitch/index.vue

### Key Decisions
- 状态映射基于全量 860 条实测：值域仅 Available(613)/available(130)/ACTIVE(117) 三个大小写变体、全为成功族，无 pending/异常态 → 创建中(orange)/异常(red) 两标为防御性计数置 0；默认子网按 attributes.is_default 派生（实测 true 72 条）
- 5 状态标 = 总数 860 / 可用(blue) 860 / 默认子网 72 / 创建中 0 / 异常 0；实测后端 total 与前端渲染完全一致
- 搜索字段按实测排除三类判据筛选：名称/云上ID（name LIKE 同匹配 asset_name+asset_id）、所属VPC（vpc_id 精确，实测真值→18/伪值→0）、可用区（zone 精确有效，不在 TS ListAssetsParams 类型但实测为准）、平台（固定选项）、区域；status 参数有效但三变体同为「可用」语义无区分度，按排除判据③不作搜索字段
- 实例数列：实测 860 行均无 instance_count 属性 → 按 AC 保留领域列位，缺值显 -（不伪造 0，同 VPC 批 shared_machines 先例）；可用IP数/总IP数/网关IP/IPv6网段/默认子网/路由表/云账号/创建时间 保留为可勾选隐藏列
- 所属VPC 列展示 vpc_id 并可点击（实测 vpc_name 属性 0/860 全空，属死属性不采用）；点击经 handleVpcClick 打开 VpcDetailDrawer（原交互逐字保留）
- 本页原无导出/列设置能力，按 proposal 风险缓解条款收敛为 action-right 图标按钮新增（LB 任务 2 同款），存储键 vswitch-column-settings 为新建、无历史漂移
- 抽屉侧 statusLabels（VSwitchDetailDrawer）已覆盖实测三变体值域，零改动；列表映射键族与抽屉同源

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 58.5%

## Acceptance Criteria
- [x] VSwitch 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标，按子网状态域派生）
- [x] 表格删除 disabled selection 列；操作栏为子网领域动作（创建子网 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [x] 列按子网领域：名称/ID(等宽ID)、CIDR、所属 VPC、可用区、实例数、状态(状态点)、平台色标、地域
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
删除文件：src/views/network/vswitch/components/VSwitchFilters.vue（能力由字段搜索栏承接，无外部引用，grep 验证）。验证证据：just e-cam-web-compile 通过；pnpm build 通过；prettier --check 干净；全仓 eslint 3 error 12 warning 与历史基线一致（scoped vswitch 目录零告警）；全量 vitest 49 文件 902/902 全绿；coverage 全仓 58.5% 与基线持平（vswitch 模块无测试文件，maintain 零下降）。浏览器实测（nginx :8888，构建即生效，SPA 内路由进入）：标题/tabs/5 状态标（860/860/72/0/0 与后端探查一致）/动作（创建子网禁用+同步主色）/右栏 3 图标/20 行渲染/列序（云上ID名称、CIDR、所属VPC、可用区、实例数、状态、平台、地域、操作）/首末列 padding 20px/状态点全部「可用」/20 个 vpc-link；搜索「vsw」→ chip「名称: vsw」共 311 条（与后端实测 311 一致），清除还原 860；行点击开 VSwitchDetailDrawer，抽屉内 VPC 链接点击成功打开 VpcDetailDrawer（跨抽屉跳转保留）；console 零错误零警告；深/浅主题计算样式均命中原型令牌（dark bg rgb(8,9,10)/文字 rgb(247,248,248)，light bg rgb(247,248,248)/文字 rgb(17,17,17)，经 pinia setTheme 双向切换验证，主题状态已恢复 light）。提交范围仅 src/views/network/vswitch/**（重写吸收该文件原有 1 行并发会话 WIP：过时 eslint-disable 注释移除）；工作区其余脏区（logs 3 文件、SG/VPC drawer WIP、coverage/、.claude/）不属本任务，未提交。Write 后经 U+FFFD 检查（index.vue 与 ColumnSettingsDialog.vue 均 0 命中）。
