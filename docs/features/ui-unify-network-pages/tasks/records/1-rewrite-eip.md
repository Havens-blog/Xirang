---
status: "completed"
started: "2026-09-24 11:30"
completed: "2026-09-24 11:55"
time_spent: "~25m"
---

# Task Record: 1 EIP 页重写对齐主机页紧凑标准

## Summary
EIP 页重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader 壳，改为 page-top（标题 + tab 全部/已绑定/未绑定 + 5 内联状态标）+ 领域操作栏 + 字段搜索栏（回车加条件 + chips）+ 手写紧凑表格（无多选列、名称列等宽 ID 副行、状态点、首/末列 20px）。删除未接线的 EipFilters.vue，导出/自定义列收敛为 action-right 图标按钮，保留同步对话框、DetailDrawer、导出与列设置能力。

## Changes

### Files Created
无

### Files Modified
- src/views/network/eip/index.vue
- src/views/network/eip/components/EipDetailDrawer.vue

### Key Decisions
- 骨架整版移植 SG 批（506c545）同 AC 措辞的已验收实现；prototype-host-eip.html 定义内容层（标题文案、tab、5 状态标、列集、动作集），与已上线 8 页令牌一致
- 全量探查 311 条实测：状态值域仅 InUse(239)/Attached(39)/Available(33)，Attached（华为云绑定态）补入 EIP_STATUS_LABELS 归已绑定族（列表/抽屉/导出三处同源），否则 39 条显示裸英文且状态标口径缺 39
- 搜索字段按实测筛除：charge_type/account_name 后端忽略（传参 total 不变）不作字段；ip_address/instance_id 后端有效但精确匹配；name 同时匹配 asset_name/asset_id
- 带宽告警/异常状态标：带宽告警无数据源置 0（主机页同款），异常按未知状态族防御性计数（当前值域下为 0）
- 领域动作 申请 EIP/绑定/解绑/修改带宽/释放/标签 均无后端能力，禁用 + tooltip（佐证=抽屉 F-EIP-02 同口径）；同步为唯一已接线动作故持有 primary 主色
- col key 与历史 eip-column-settings 存储零漂移；实例类型/云账号/创建时间改为默认隐藏但保留可勾选，已保存设置无需迁移
- H-03 query.search 跳入预填改为搜索条件 chip（可见可移除），无 query 直达行为不变

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] EIP 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab 全部/已绑定/未绑定 + 5 内联状态标：总数/已绑定/未绑定/带宽告警/异常）
- [x] 表格删除 disabled selection 列；操作栏为 EIP 领域动作（申请 EIP/绑定/解绑/修改带宽/释放/标签 + 保留同步）；字段搜索栏 + chips 可用；导出/自定义列保留为图标
- [x] 列按 EIP 领域：名称/ID(等宽ID副行)、公网 IP、状态(状态点)、带宽、计费方式、关联实例、平台色标、地域
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与原型一致、无控制台错误

## Notes
实测证据：列表参数 total 311，已绑定 278=InUse 239+Attached 39，未绑定 33；状态筛选传参 InUse=239/Available=33/Attached=39 精确生效；搜索 名称:AI代理→2 条、状态:未绑定→33 条 chips 增删均实测。浏览器验收（:8888，SPA 内路由进入）：深色 page-top 背景 rgb(15,16,17)、浅色 rgb(255,255,255)、强调色 rgb(113,112,255) 与原型 token 一致；th/td 首末列 padding 20px 实测；控制台零 error（仅与 SG 页同款的 input 无 name 属性 a11y 提示）。pnpm build 通过（chunk 体积警告为既有）；vue-tsc -b 零错误；eslint 3 错 12 警与基线持平；全量 vitest 902/902 绿；eip 模块无测试文件故 coverage 0%（runner 实报）。EipFilters.vue 已删除（仅 index.vue 引用过）。带宽列对 bandwidth=0（103 条，全为 AWS）显示 '-'，与原页口径一致。
