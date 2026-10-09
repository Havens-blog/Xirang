---
status: "completed"
started: "2026-09-24 14:16"
completed: "2026-09-24 14:40"
time_spent: "~24m"
---

# Task Record: 7 WAF 页重写对齐主机页紧凑标准

## Summary
WAF 防火墙页重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader/StatCard 壳，改为 page-top（标题 + 4 tab + 5 内联状态标）+ 操作栏（创建实例/防护配置 禁用+tooltip + 同步实例 primary）+ 字段搜索栏（chips）+ 无多选列表格 + 分页栏；新增 ColumnSettingsDialog（键 waf-column-settings）；列按实测值域派生；WafDetailDrawer 未动

## Changes

### Files Created
- src/views/network/waf/components/ColumnSettingsDialog.vue

### Files Modified
- src/views/network/waf/index.vue

### Key Decisions
- 值域全量实测（425 条翻页聚合）：status 仅 active(160)/suspended(252)/pending(13) 全小写；edition 实测 EdgeOne(13)/sparta-waf(100)/空(312)，原 basic/pro 筛选选项与 editionMap 为虚构值，已按实测重写并保留旧键兼容
- 5 状态标 总数425/正常 blue/已暂停/待接入 orange/异常 red(未知状态族防御性计数=0)；原「正常运行」StatCard（当前页口径）并入「正常」标并升为全局口径（分页全量拉取本地计数，eip/cdn 同款）
- 后端参数实测（传参看 total）：name（LIKE 匹配 asset_name+asset_id）/status/edition/provider/region 有效；protection_mode/domain_count/qps/waf_enabled/cloud_account_name 后端忽略，不作搜索字段；搜索字段 5 个
- region 选项从统计全量拉取动态派生：实测区域值域被采集任务实时改写（10 分钟内 195 条区域 cn-shanghai→cn-hongkong→ap-northeast-1 三变），静态枚举必漂移失效
- 原 4 tab 全保留（攻击日志 tab 为 2e65651 接线入口，/logs?t=waf 原样；防护域名/安全报表占位 tab 保留开发中文案）；QPS（425 行全 0）与到期时间（全空）降为可勾选隐藏列不进默认视图
- 多选列随紧凑标准删除（AC 口径），AssetExportDialog 传 selected-ids=[] 保持「已选中」禁用零漂移；导出字段/文件名前缀原样保留，status/edition 导出文案与列表同源

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 58.5%

## Acceptance Criteria
- [x] WAF 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标），无大统计卡
- [x] 表格删除 selection 列；操作栏 创建实例/防护配置（禁用+tooltip）+ 保留同步 primary；字段搜索栏 + chips；导出/自定义列保留（新增 ColumnSettingsDialog）
- [x] 列按 WAF 领域：实例（等宽 ID 副行）、状态（状态点）、版本、防护域名、防护模式、规则数、平台 IconFont、地域
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增（3 errors/12 warnings 与基线一致）
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误（badges 425/160/252/13/0 与实测一致；搜索/列设置/导出/抽屉/4 tab 交互全通过）

## Notes
lint 全量含未跟踪 coverage/ 生成目录（并发会话产物）会多 1 warning，排除后恰为基线 3E/12W；vitest 902/902、coverage 58.5% 与基线持平；浏览器验证用 evaluate_script + take_snapshot（未用 take_screenshot），主题切换走 pinia setTheme 并已恢复 light
