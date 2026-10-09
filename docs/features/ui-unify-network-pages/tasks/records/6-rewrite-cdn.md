---
status: "completed"
started: "2026-09-24 13:51"
completed: "2026-09-24 14:14"
time_spent: "~23m"
---

# Task Record: 6 CDN 页重写对齐主机页紧凑标准

## Summary
CDN 加速页重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader 壳改手写 page-top（标题「CDN 加速域名」+ tab 全部/在线/离线 + 5 内联状态标 总数/在线(blue)/离线/HTTPS 开启/证书）；删 disabled selection 死列与 el-table，改主机页手写密集表格（首列加速域名 + 条件等宽 ID 副行、状态点、平台 IconFont、首/末列 20px 对齐）；操作栏 CDN 领域动作（添加域名/刷新缓存禁用+tooltip、同步实例 primary、成本明细按钮保留原大卡入口）；内联筛选改字段搜索栏 + chips（6 字段按后端实测有效性派生）；导出/自定义列/详情抽屉/成本面板/同步对话框全保留，col key 与 cdn-column-settings 零漂移。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/views/network/cdn/index.vue
- e-cam-web/src/views/network/cdn/components/ColumnSettingsDialog.vue

### Key Decisions
- 值域探查（2026-09-24 全量 598 条实测）：status 仅 online(566)/offline(31)/InProgress(1) 三统一枚举值（历史原始值 Started/Deployed 等已归一），CDN_STATUS_LABELS 全覆盖无裸显英文；https_enabled true=513/false=85；cert_name 598 行全空 → 证书标显示 0 为真实数据状态
- 参数有效性实测：name 按域名 LIKE 匹配（不匹配 CNAME）、status/business_type/service_area 统一枚举精确匹配（历史原始值如 Started/page/mainland 传参 total 不变）、https_enabled 'true'/'false' 有效；cname/cert_name/http2_enabled 后端忽略不作搜索字段 → 综合有效搜索字段 6 个
- 状态标改 eip 同款全量分页本地计数（单次 1000/页拉取，在线/离线/HTTPS/证书同口径派生），弃后端 status=online 参数法（对历史原始值行有漏计风险）
- 首列副行条件渲染：asset_id 实测 598 行恒等于域名，仅在两者不同时显示，避免整列重复值
- 原「本月成本」大卡（CdnCostPanel 唯一入口）收敛为操作栏按钮保留能力；「近2日流量」纯展示卡无其余承接面随大卡移除（getCdnTopDomainsApi 调用删除）
- CNAME 列保留 mono 样式；原页无复制功能（grep 无 clipboard），以 title 原生 tooltip 等价保留 show-overflow-tooltip 能力

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 58.5%

## Acceptance Criteria
- [x] CDN 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标：总数/在线/离线/HTTPS 开启/证书）
- [x] 表格删除 disabled selection 列；操作栏为 CDN 领域动作（添加域名/刷新缓存 + 保留同步）；字段搜索栏 + chips；导出/自定义列保留
- [x] 列按 CDN 领域：加速域名(等宽ID)、CNAME、业务类型、HTTPS、证书、服务区域、状态(状态点)、平台色标
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
验证证据：vue-tsc -b 通过；prettier --check 通过；eslint 3 错 12 警=基线（均不在 cdn 文件）；vitest 全量 902/902、coverage 58.5% 与基线持平；浏览器 :8888 构建实测——状态标计数与全量探查逐一吻合、搜索 chip「域名: easyeda」过滤 9 行后清除恢复 20 行、详情抽屉/导出/自定义列/同步对话框均开合正常、深浅主题令牌与原型一致（light #f7f8f8/#fff/#111、dark #08090a/#0f1011/#f7f8f8）、控制台无错误。提交范围仅 cdn 两文件，未触碰并发会话脏区（logs/、security-group drawer、vpc drawer）。
