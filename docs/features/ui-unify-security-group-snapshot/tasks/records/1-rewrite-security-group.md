---
status: "completed"
started: "2026-09-24 10:16"
completed: "2026-09-24 10:41"
time_spent: "~25m"
---

# Task Record: 1 SecurityGroup 页重写对齐主机页紧凑标准

## Summary
SecurityGroup 页（240 行薄骨架）重写为「主机页紧凑标准」（约 860 行，仅改 src/views/compute/security-group/index.vue）：page-top 标题 + tab + 5 内联状态标（总数/企业级/普通/未关联VPC/无规则）；安全组领域动作（新建/规则管理/关联实例/标签，禁用+tooltip，佐证=详情抽屉 tab）；字段搜索栏（名称/云上ID/平台/区域/VPC ID）+ 条件 chips；去常驻多选列与全选/已选 UI；名称列等宽 ID 副行、首/末列 20px 对齐；分页右对齐带 jumper；行点击开详情抽屉；导出弹窗 selected-ids 传 [] 零漂移。浏览器实测（nginx :8888，910 条真实数据）深/浅两主题渲染一致、0 控制台错误。

## Changes

### Files Created
无

### Files Modified
- e-cam-web/src/views/compute/security-group/index.vue

### Key Decisions
- 状态标按领域维度派生（OSS 无 status 先例同型）：实测 910 条安全组顶层 status 全空、status/security_group_type 参数后端忽略（传参返回全量），故 5 标取 类型分布（企业级 73/普通 837）+ VPC 覆盖（未关联VPC 127）+ 规则覆盖（无规则 32）；空类型与列展示同口径归普通（原页 enterprise?企业级:普通 零漂移）；统计用 limit=1000 全量拉取本地计数（实测一次取全，带 10 页 cap）
- 搜索字段仅映射 ListSecurityGroupParams 后端支持参数（name/provider/region/vpc_id）：security_group_type 无后端参数、status 无值域，均不作搜索字段；name 参数实测同时匹配 asset_id（云上ID 字段映射 name）
- 列 key/label/localStorage 键 security-group-column-settings 与原页完全一致 → 历史保存设置零迁移，components/ 下 ColumnSettingsDialog 与 SecurityGroupDetailDrawer 两组件未改动
- 整版移植最近已提交标准页 NAS（525217e）样式骨架，去 ops-strip/status-dot（本页无状态列），保留原页 env-tag 类型样式；filters.provider 用 CloudProvider 类型约束通过 vue-tsc（原页 as any 改类型化）
- 基线 3 个 network/ 脏文件（并发会话 WIP）与本任务无关未触碰未提交；components.d.ts/auto-imports.d.ts 构建再生成，按惯例不入提交

## Test Results
- **Tests Executed**: Yes
- **Passed**: 902
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] SecurityGroup 页为紧凑标准骨架：page-top（标题 + tab + 5 内联状态标，无大卡片）
- [x] 操作栏为安全组领域动作（勿照搬主机电源操作）；表格无默认多选列；字段搜索栏 + chips 可用
- [x] 列按安全组领域派生：安全组 ID/名称、类型（企业级/普通）、VPC、规则数、关联实例、平台色标、地域等
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
验证链：npx vue-tsc -b 全绿；pnpm build 通过（chunk>1500kB 警告为既有 advisory）；eslint 3 错 12 警均为历史基线（src/api/types/{cost,statistics,task}.ts 等，零 security-group 提及）；vitest 全量 902/902 绿（49 文件）；coverage-v8 实测 security-group 模块 0%（无测试文件，maintain 策略零下降）。浏览器实测：evaluate_script 断言 5 状态标真实值（910/73/837/127/32）、chips 增删（平台=火山引擎→212 条、VPC ID→19 条、清除→910，与 API 探查一致）、行点击开抽屉（5 tab）、.close-corner 关闭（overlay display:none）、深浅主题计算样式（dark: surface #191a1b/text #f7f8f8；light: #ffffff/#111111，取值时避开 200ms 主题过渡竞态）、首末列 20px、控制台 0 错误 0 警告；测试后已恢复该 tab 原 URL 与 dark 主题。
