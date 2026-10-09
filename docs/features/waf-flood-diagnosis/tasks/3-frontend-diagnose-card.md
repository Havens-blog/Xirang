---
id: "3"
title: "前端 WAF Tab 内嵌「流量诊断」卡(触发/展示/降级)"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: [2]
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 3: 前端 WAF Tab 内嵌「流量诊断」卡(触发/展示/降级)

## Description
在日志查询 WAF Tab 新增"流量诊断"折叠卡:手动触发按钮(取当前查询的窗口/域名/检索式/筛选),调用 `/logs/diagnose`,渲染诊断结果——风险等级徽标 + 疑似攻击类型、Top 攻击源 IP(请求数/占比)、来源分布、趋势对比(当前 vs 前窗突增)、措施清单(可复制)。含加载态/错误态/前窗缺失降级提示。诊断卡独立于既有统计视图,不影响明细/聚合行为。

## Reference Files
- `docs/proposals/waf-flood-diagnosis/proposal.md` — Source proposal;«Proposed Solution» «In Scope»
- `e-cam-web/src/views/logs/index.vue`: WAF Tab 现有结构/聚合条/统计卡(诊断卡同区块插入)
- `e-cam-web/src/api/logs.ts`: API 封装(diagnose 接口新增)
- `e-cam-web/src/api/types/logs.ts`: 响应类型(DiagnoseResponse)

## Acceptance Criteria
- [ ] WAF Tab 出现"流量诊断"入口(折叠卡),点击触发后调用 `/logs/diagnose` 并渲染
- [ ] 展示:风险等级(无/低/中/高,徽标色区分)、疑似攻击类型、Top5 攻击源 IP(请求数+占比)、来源分布、趋势对比(当前 vs 前窗突增倍数)、措施清单(≥3 条,可复制)
- [ ] 前窗无数据/接口降级时给出对应提示(不白屏);加载态/失败可重试
- [ ] `vue-tsc --noEmit` 通过;既有 451 前端测试回归通过

## Hard Rules
仅修改 e-cam-web 仓;诊断卡不得改变既有明细/统计/聚合行为(纯新增区块);不得在前端引入模型调用(AI 解读后置,后端 summary 占位即可)。

## Implementation Notes
- 触发按钮放置:WAF Tab 统计卡/聚合条附近(如"流量诊断"按钮 + 折叠面板),复用当前 `timeRange`/`keyword`/`selectedClouds`/`selectedResources`/`buildFilters()`。
- 诊断仅在 logType==='waf' 时展示入口(CDN/SLB 不显示)。
- 复用现有风格:风险等级徽标用 el-tag type(danger/warning/primary/success 映射 高/中/低/无);措施清单用纯文本列表 + 复制按钮。
- `diagnoseApi` 请求体 = 当前聚合请求字段(不含 dimension/metric);响应类型 `DiagnoseResponse`(risk_level/attack_type/sources/geo/summary/measures/prev_window)。
