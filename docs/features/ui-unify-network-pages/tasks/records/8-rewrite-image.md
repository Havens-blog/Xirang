---
status: "completed"
started: "2026-09-24 14:43"
completed: "2026-09-24 15:08"
time_spent: "~25m"
---

# Task Record: 8 Image 页重写对齐主机页紧凑标准

## Summary
镜像管理页（compute/image）整版重写对齐主机页紧凑标准：去 PageContainer/ManagerHeader/StatCard/AccountSidebar 壳 → page-top（标题「镜像管理」+ 5 tab 全部/公共镜像/自定义镜像/共享镜像/市场镜像沿原侧栏镜像类型域 + 5 内联状态标 总数/公共(blue)/自定义/共享/导入中(orange)）；操作栏 领域动作 创建镜像/导入镜像（禁用+tooltip）+ 同步镜像 primary（原 ImageSyncDialog 保留）；7 字段搜索栏+chips（名称ID/云账号/架构/操作系统/镜像平台/云平台/地域，全部后端实测有效）；表格去 disabled 多选列，首列 云上ID/名称+等宽副行，列 类型(tag)/操作系统/架构/大小/状态点/平台IconFont/地域 + 隐藏列 云账号/创建时间；行操作 创建实例（原已接线跳转逐字保留）+详情；新增导出（AssetExportDialog selected-ids=[] 零漂移）与自定义列（新键 image-column-settings 无历史漂移）；首/末列 20px 对齐；ImageDetailDrawer 未动

## Changes

### Files Created
- src/views/compute/image/components/ColumnSettingsDialog.vue

### Files Modified
- src/views/compute/image/index.vue
- src/views/compute/image/components/AccountSidebar.vue (deleted)
- src/views/compute/image/components/ImageFilters.vue (deleted)
- src/views/compute/image/components/OsIcon.vue (deleted)

### Key Decisions
- 写码前全量探值域（2026-09-24 全量 669 条翻页聚合）：status 仅 Available(141)/available(506)/active(16)/NORMAL(6) 四变体全成功族、无 creating/failed —— 原页 statusLabels 缺 active/NORMAL 会裸显英文，按成功/导入中/不可用三族映射 + lenient 兜底；「导入中」状态标为创建族派生计数（实测 0，真实数据态）
- image_owner_alias 实测 669 条全部 self（自定义），stats 接口口径 total 669/system 0/custom 669/shared 0 与扫描一致 → 公共/共享/市场 tab 与状态标显示真实 0（CDN 证书 0 同款诚实态）
- 搜索字段按实测排除两类：status（参数有效但四变体全成功族无区分度，排除判据③）、cloud_account_name（后端忽略）；架构/镜像平台/地域选项从全量扫描动态派生防值域漂移（WAF region 漂移坑），os_type 为稳定域枚举静态给出（大写 Linux 实测查 0）；云账号搜索选项来自 cloud-accounts 接口，承接原侧栏账号筛选能力（account_id 实测有效 141）
- 操作系统列 platform 优先、os_name 兜底：实测 AWS 系 360/669 条 os_name 为「Created for policy: …」采集描述垃圾值，platform 才是干净系统族；size=0 显示 '-'（EIP bandwidth=0 同款）
- 原侧栏（AccountSidebar）删除：镜像类型域 → 页头 tab、账号筛选 → 云账号搜索字段、同步快捷操作 → 同步镜像按钮；行内已接线的 创建实例 跳转（/compute/template/create?image_id=）逐字保留为行操作链接
- tab 切换与搜索 chips 各自管理独立参数（image_owner_alias 归 tab，applySearchConditions 不清该键），互不冲突

## Test Results
- **Tests Executed**: Yes
- **Passed**: 27
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] Image 页为紧凑标准骨架：去 PageContainer/ManagerHeader → page-top（标题 + tab + 5 内联状态标：总数/公共/自定义/共享/导入中）
- [x] 表格删除 disabled selection 列；操作栏为镜像领域动作（创建镜像/导入）+ 保留同步；字段搜索栏 + chips；导出/自定义列保留
- [x] 列按镜像领域：名称/ID(等宽ID)、类型(公共/自定义/共享)、OS、架构、大小、状态(状态点)、平台色标、地域
- [x] 首/末列 20px 对齐；pnpm build/typecheck 通过、lint 相对基线零新增
- [x] 浏览器深/浅两主题渲染与标准一致、无控制台错误

## Notes
仅改动 src/views/compute/image/** 五文件：index.vue 整版重写（421→约1240 行）、新增 ColumnSettingsDialog.vue、删除 AccountSidebar/ImageFilters/OsIcon（均仅原 index.vue 引用）；ImageDetailDrawer/ImageSyncDialog 未动。验证：just e-cam-web-compile 绿、vite build 绿、eslint image 模块 0 错 0 警（全仓 3 错 13 警，第 13 警为未跟踪 coverage/ 运行产物非源码，源码基线 3 错 12 警零新增）、定向 vitest views/compute 27/27 绿、coverage-v8 实测 image 模块 0%（无测试文件，与 SG/Snapshot 同状）。浏览器 :8888 构建实测：snapshot 结构/状态标数值（669/0/669/0/0）/搜索 7 字段与动态选项（地域 10 项、平台 15 项、账号 5 项）/chips 过滤（eu-central-1→430）/tab 切换（公共镜像→0 空态）/详情抽屉（3 tab）/同步对话框/自定义列增删+localStorage 持久化+恢复默认/导出对话框/创建实例导航往返 全通过；深浅两主题计算样式均为令牌值（靛紫 #7170ff 常量、明暗表面层次正确）、console 零错误零警告
