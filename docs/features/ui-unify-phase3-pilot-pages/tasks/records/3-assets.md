---
status: "completed"
started: "2026-09-21 18:01"
completed: "2026-09-21 19:19"
time_spent: "~1h 18m"
---

# Task Record: 3 assets 资产列表等价重写（FilterBar/DataTable/抽屉 + 浏览器实测）

## Summary
assets 资产列表等价重写：PageContainer(#header 标题副标题+#actions 发现/同步+#filters+#footer 分页) + FilterBar(云厂商/资产类型/区域/状态/资产名称 5 字段 + 重置 + 折叠展开) + DataTable(fetch 模式内置三态、批量操作条、密度切换、stripe/max-height 冻结表头) + 规格增强行点击开 DetailDrawer 2.0；旧页 13 条用户路径逐项保留（发现/同步/编辑/删除对话框与抽屉三标签页原样迁移），删除孤儿子组件 AssetFilters.vue/AssetTable.vue；新增 24 项单测，全量 830 绿，vue-tsc -b 与 pnpm build 通过。

## Changes

### Files Created
- src/views/assets/__tests__/index.test.ts

### Files Modified
- src/views/assets/index.vue
- src/views/assets/components/AssetFilters.vue
- src/views/assets/components/AssetTable.vue

### Key Decisions
- DataTable 采用 fetch 模式（组件自管 loading 骨架/empty/error+retry 三态），fetch 闭包在调用时读取最新 filters/pagination，total 由闭包侧写回（组件契约 fetcher 仅返回行数组）；筛选变更/分页变化显式 refresh()，与旧页显式 fetchAssets 控制流等价
- 旧页 ManagerHeader 的 @refresh 为死代码（#actions 插槽覆盖了默认刷新圆钮），新页按等价边界不设刷新按钮
- AC 文字中提到的「导出」在旧页不存在（AssetExportDialog 未被 assets 页引用），按 Hard Rule 功能等价边界不新增
- 批量操作条为规格指定增强，动作由既有删除能力组成：批量删除 → ElMessageBox 确认 → Promise.allSettled 逐行 deleteAssetApi → 部分失败报告 → clearSelection + refresh
- FilterBar 输入字段检索触发时机为 change（回车/失焦），替代旧页 500ms 防抖输入即搜（组件内建语义，筛选路径等价保留）
- ProviderIcon/AssetStatusBadge 槽位绑定补 `|| ''` 守卫（对齐 accounts/compute-image 页既有惯例），消除每次 refetch 触发的 2 条 Vue prop 校验警告
- 删除孤儿子组件 AssetFilters.vue / AssetTable.vue（重写后无任何引用）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 830
- **Failed**: 0
- **Coverage**: 88.4%

## Acceptance Criteria
- [x] 等价重写：旧页用户路径盘点（5 筛选字段、9 列结构、分页、查看/编辑/删除、发现/同步、详情抽屉）逐项保留
- [x] 规格增强生效：行点击打开 DetailDrawer 2.0；DataTable 密度切换生效
- [x] 页面 scoped 样式 <100 行（实测 90 行），零���增硬编码色字面量（单测守卫：色字面量 ⊆ 旧页基线=空集）
- [x] chrome-devtools-mcp 实测：导航 /cam/assets、截图归档 5 张、FilterBar 筛选→DataTable 三态→分页链路走通（网络参数 provider=aliyun&offset=0 / offset=20&limit=20 / name=openvpn 证据）、9 项计算样式令牌断言通过
- [x] chrome-devtools-mcp 实测：行点击抽屉打开且内容正确（标题=资产名/三标签页/描述列表/头操作）、批量选择出现操作条（已选 2 项）、双主题 + hover（light hover 行底 rgb(243,244,245)）
- [x] 相关测试更新后全量绿（830 passed），pnpm build 与 vue-tsc -b 通过

## Notes
coverage 88.42 为 src/views/assets/index.vue 语句覆盖率（分支 83.92%/行 89.14%，v8 报告）。浏览器证据归档 D:\Haven\docs\features\ui-unify-phase3-pilot-pages\screenshots\3-assets\（dark 全页/行点击抽屉/批量操作条/light hover/筛选展开）。真实登录态验收未执行删除/发现/同步提交（防脏数据），提交链路由单测覆盖；后端对 name 筛选参数的返回行为与旧页一致（前端传参链路正确）。AssetFilters.vue/AssetTable.vue 为删除（git rm）。修正过程中的 vite 转换缓存竞态导致的一次测试闪Fail（3 次复跑全绿）。
