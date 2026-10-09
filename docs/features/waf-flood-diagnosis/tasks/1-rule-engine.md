---
id: "1"
title: "WAF 被刷判定规则引擎(风险分/等级/攻击类型/措施映射,纯函数+单测)"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: []
surface-key: "e-cam-service"
surface-type: "api"
breaking: false
type: "coding.feature"
mainSession: false
---

# 1: WAF 被刷判定规则引擎(风险分/等级/攻击类型/措施映射,纯函数+单测)

## Description
实现"网站有没有被刷"的确定性判定引擎(纯函数、零 I/O、可单测)。输入:当前窗聚合行(Total、时间分桶、Top client_ip、Top UA、状态码分布、动作分布)+ 前窗对比(Total 突增倍数、Top IP 突增)。输出:风险分 / 风险等级(无/低/中/高)/ 疑似攻击类型(CC 刷量 / 恶意爬虫 / 接口爆破 / 正常高流量)/ 措施文案列表。判据来自 proposal:单 IP 最高频率、UA 异常、请求突增、5xx/攻击动作占比;权重与阈值集中配置(可校准)。

## Reference Files
- `docs/proposals/waf-flood-diagnosis/proposal.md` — Source proposal;«Proposed Solution» «Key Risks» «Success Criteria»
- `e-cam-service/internal/logquery/service/federation.go`: 聚合编排/AggregateParams/结果结构
- `e-cam-service/internal/shared/cloudx/logquery/aggregate.go`: AggregateResult/TopNItem/AggregateBucket(诊断输入结构同源)

## Acceptance Criteria
- [ ] 风险分 = 单 IP 最高频率 + 突增倍数 + UA 异常 + 5xx/攻击动作占比加权;权重与阈值集中配置(常量或配置表,可校准)
- [ ] 风险等级映射(无/低/中/高)边界有单测覆盖(临界分)
- [ ] 攻击类型判定 ≥3 类:CC 刷量 / 恶意爬虫 / 接口爆破 / 正常高流量;大促高流量(突增但 UA 多样、来源分散、无攻击动作)判低风险不误报
- [ ] 措施映射:按判定输出 ≥3 条具体文案(封禁 Top 攻击源 IP+请求数、开启 CC 防护/限流、异常 UA 黑名单)
- [ ] 纯函数(无网络/DB),输入为聚合行结构与前窗对比值,单测覆盖率 ≥80%(风险分/等级边界/类型判定/措施映射)

## Hard Rules
仅修改 e-cam-service 仓;不得新增数据面(不新增大数据组件/采集),输入仅来自现有聚合通道可产出的结构;不得引入 LLM/外部调用(本任务纯规则)。

## Implementation Notes
- 结构建议:`internal/logquery/service/diagnose.go`(或 `internal/logquery/diagnose/`):定义 `DiagnoseInput`(当前窗聚合 + 前窗对比)、`DiagnoseResult`(RiskScore/RiskLevel/AttackType/Measures/TopSources)、纯函数 `Evaluate(input) Result`。
- 判据锚定(实测校准):阿里 WAF 近 6h 全窗 8000 万行;单 IP 请求数量级需从真实 Top client_ip 聚合观察后定阈值(首版给保守默认,标注可调)。
- 攻击类型判据草案:CC=单 IP 极高频率且请求突增;爬虫=UA 异常/单一 + 4xx 高;爆破=单 IP 高频 + 4xx 高 + URI 集中;正常高流量=突增但 UA 多样、地理分散、无攻击动作。
- 措施文案需带具体值(IP、请求数),由判定数据插值生成,非纯静态。
