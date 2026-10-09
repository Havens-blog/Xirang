---
status: "completed"
started: "2026-09-19 17:06"
completed: "2026-09-19 17:34"
time_spent: "~28m"
---

# Task Record: 3 前端 WAF Tab 内嵌「流量诊断」卡(触发/展示/降级)

## Summary
WAF Tab 新增「流量诊断」折叠卡(纯新增区块):新增 LogDiagnoseCard 组件(折叠头+手动触发+结果/降级展示)与 diagnose.ts 展示辅助(风险等级/攻击类型标签、徽标色映射、占比/突增/千分位格式化),API 层新增 diagnoseLogsApi(POST /cam/logs/diagnose)与 LogDiagnoseRequest/Response 等类型;index.vue 仅当 activeType==='waf' 渲染,诊断上下文复用当前 timeRange/keyword/selectedClouds/selectedResources/buildFilters()。展示:风险等级徽标(高=danger/中=warning/低=primary/无=success)+风险分+疑似攻击类型、Top 攻击源 IP(请求数+占比,后端 Top5)、来源分布(按云账号)、趋势对比(当前 vs 前窗突增倍数)、措施清单(≥3 条可复制);降级不白屏(prev 缺省+prev_error、prev.total=0、degraded、dimension_notes 各有提示),加载态计时文字回馈,失败可重试;AI summary 占位空串不渲染,未引入任何模型调用。

## Changes

### Files Created
- e-cam-web/src/views/logs/diagnose.ts
- e-cam-web/src/views/logs/diagnose.test.ts
- e-cam-web/src/views/logs/components/LogDiagnoseCard.vue
- e-cam-web/src/views/logs/components/LogDiagnoseCard.test.ts

### Files Modified
- e-cam-web/src/api/logs.ts
- e-cam-web/src/api/types/logs.ts
- e-cam-web/src/views/logs/index.vue
- e-cam-web/src/views/logs/index.test.ts

### Key Decisions
- 响应类型以 T2 已合入后端真实 wire 契约为准(service.DiagnoseResponse):risk_level/attack_type/measures/top_sources/surge_multiplier 在 result.* 下,prev={total,top_ip_count}+prev_error,无 geo 字段 —— 任务 Implementation Notes 的字段草图(risk_level 平铺/geo/prev_window)与后端不符,按后端契约实现,否则卡片渲染为空
- proposal 的「来源分布(Top 地域)」后端 T2 未聚合 geo,前端如实降级为「来源分布(按云账号)」(sources[].total/错误标注),不伪造地域数据
- 诊断卡独立于统计/明细:v-if=activeType==='waf' 纯新增区块,index.vue 仅加 template 块+import+diagnoseContext computed,不触碰任何既有 handler;页级测试断言触发诊断后 searchApi/aggregateApi 零调用
- vitest 环境中 app.use(ElementPlus) 不解析 el-tag(渲染为未知原生元素)、ElTag 根为 transition 会被 test-utils 存根 —— 卡片测试显式注册真实 ElTag 并按 .risk-tag .el-tag 断言徽标色;运行时无影响
- coverage 字段不填:@vitest/coverage-v8 未安装(--no-save 安装两度失败,npm 'matches' 错误),无 runner 输出不臆造,沿 bdd9513 前端任务先例;新代码 37 用例覆盖全部状态机分支(仅剪贴板 reject 分支未覆盖)

## Test Results
- **Tests Executed**: Yes
- **Passed**: 485
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] WAF Tab 出现"流量诊断"入口(折叠卡),点击触发后调用 /logs/diagnose 并渲染
- [x] 展示:风险等级(无/低/中/高,徽标色区分)、疑似攻击类型、Top5 攻击源 IP(请求数+占比)、来源分布、趋势对比(当前 vs 前窗突增倍数)、措施清单(≥3 条,可复制)
- [x] 前窗无数据/接口降级时给出对应提示(不白屏);加载态/失败可重试
- [x] vue-tsc --noEmit 通过;既有 451 前端测试回归通过

## Notes
Hard Rules 合规:仅改 e-cam-web 仓;纯新增区块(页级测试证明诊断触发时 search/aggregate 零调用,既有行为不变);前端零模型调用(summary 恒空串不渲染)。AC4 的"451"为陈旧基线,任务开始时全量即 458/21 files,完成后 485/23 files(458+27 新增),零回归。vue-tsc --noEmit exit 0;eslint(8 个变更文件) 0 errors 且无 --fix 改写。并发会话脏文件(.env.development、element-theme.scss、components.d.ts、accounts/index.vue、.claude/)未触碰,提交须显式 pathspec。测试基建事实记录:happy-dom+ElementPlus 全插件环境下 el-tag 无法经全局插件解析(app 级),需局部显式注册;axios 拦截器对业务错误会先 ElMessage.error 再 reject,卡片错误态与全局消息并存(与既有 doSearch 行为一致)。
