---
status: "completed"
started: "2026-09-20 09:41"
completed: "2026-09-20 09:56"
time_spent: "~15m"
---

# Task Record: 1 新建共享 AssetExportDialog 配置驱动组件 + 单测

## Summary
新建共享配置驱动导出弹窗组件 src/components/AssetExportDialog.vue（props 与现中档 ExportDialog 完全一致：visible/instances/selectedIds/total/fetchAllRows?，新增 config prop 导出 ExportFieldConfig：fields/getValue/filename/defaultFields），UI 采用 ecs 全功能卡样式（范围按钮+格式卡+字段勾选+全量拉取行内进度，680px 自带内联 scss 不跨目录 import），导出语义逐字保留自 ecs：范围解析 current→selected→all、csv|xlsx（BOM/引号转义/制表符或逗号分隔）、文件名 {config.filename}_{YYYYMMDD}.{ext}、fetchAllRows 进度回调与失败降级（ElMessage.error 透出 message 且不关闭弹窗，非 Error 用兜底文案）、未提供闭包时 all 退回当前页。新增单测 9 例覆盖全部 AC。仅新增组件+测试两个文件，未改任何现有 ExportDialog/index.vue/accounts/logs/后端/components.d.ts。

## Changes

### Files Created
- e-cam-web/src/components/AssetExportDialog.vue
- e-cam-web/src/components/__tests__/AssetExportDialog.test.ts

### Files Modified
无

### Key Decisions
- ExportFieldConfig 接口放独立 <script lang="ts"> 块导出（script setup 不允许 export），供迁移任务 2/3/4 import 类型
- 组件显式 import ElMessage/icons，不依赖 unplugin 自动导入登记（规避并发会话 components.d.ts 脏文件冲突）
- resolveExportRows 复制 ecs 的 current→selected→all 顺序（Implementation Notes 指定），文件名/格式化逻辑逐字复制 ecs（含 formatDate 本地时区 YYYYMMDD）
- 范围按钮「全部数据」文案采用 oss/任务 AC 措辞（ecs 为「当前过滤条件全部数据」，属 UI 文案非导出语义）
- getValue 为纯函数 prop，页面特例（nas metric-map 等）经闭包捕获，不做特殊 prop

## Test Results
- **Tests Executed**: Yes
- **Passed**: 512
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 新建 AssetExportDialog.vue，props 与现中档一致 + config prop（ExportFieldConfig）+ emit update:visible
- [x] UI 含范围按钮（已选中无数据禁用）、格式卡（xlsx/csv）、字段勾选（全选/取消全选）、全量拉取行内进度（fetchingAll 禁用按钮）
- [x] 导出语义逐字保留：范围解析三态/csv|xlsx（BOM、引号转义、分隔符）/文件名/失败 ElMessage.error 返回 null 不关闭/成功关闭弹窗
- [x] 单测覆盖：范围解析三态、csv/xlsx 内容（BOM+转义）、fetchAllRows 进度、失败降级、无 fetchAllRows 时 all 退 current
- [x] pnpm typecheck 通过；vitest 全量 512/512 通过；git diff 仅新增共享组件 + 测试文件

## Notes
提交 732534bd97e129283e04b8b6e11822266d604e50（git -C D:/Haven/e-cam-web，显式 pathspec add 两文件）。coverage 未上报：本机 coverage-v8 不可用（历史既知），指标均来自真实测试输出。并发会话脏文件（.env.development/element-theme.scss/components.d.ts/accounts/index.vue/.claude/）未触碰未提交。单测中 csv 表头不加引号、xlsx 表头加引号——与 ecs 原实现逐字一致（csv headers.join 裸拼接，xlsx 表头进 quote map）。
