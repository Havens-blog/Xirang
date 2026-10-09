---
status: "completed"
started: "2026-09-19 15:53"
completed: "2026-09-19 16:25"
time_spent: "~32m"
---

# Task Record: 1 WAF 被刷判定规则引擎(风险分/等级/攻击类型/措施映射,纯函数+单测)

## Summary
实现 WAF 被刷判定规则引擎(纯函数包 internal/logquery/diagnose):DiagnoseInput(当前窗聚合行 Top client_ip/UA/URI/状态码/动作 + 前窗对比)→ Evaluate 输出 RiskScore/RiskLevel/AttackType/Measures/TopSources。风险分为单 IP 频率+突增(Total 与 Top IP 取大)+UA 异常+5xx/攻击动作占比四分量加权,权重/阈值/分段评分表/特征字典全部集中在 config.go DefaultConfig 可校准;攻击类型判 CC 刷量(突增路径+绝对量降级路径)/恶意爬虫/接口爆破(URI 维度缺失时退化)/正常高流量(大促不误报,等级封顶低)/正常;措施映射按判定输出 ≥3 条带具体值插值的文案;前窗无数据/窗口缺失输出降级标注。

## Changes

### Files Created
- e-cam-service/internal/logquery/diagnose/diagnose.go
- e-cam-service/internal/logquery/diagnose/config.go
- e-cam-service/internal/logquery/diagnose/diagnose_test.go

### Files Modified
无

### Key Decisions
- 新建独立纯函数包 internal/logquery/diagnose(零 I/O,仅依赖 cloudx/logquery 类型),任务 2 诊断接口可直接调用 Evaluate
- 风险分四分量权重 40/25/15/20,分段线性评分表集中配置;Top IP 突增与 Total 突增取大作为突增分量
- CC 判定加单 IP 占比下限(CCMinIPShare=0.2)防大促高流量误报;正常高流量判定后等级显式封顶低风险
- 输入结构严格同源现有聚合通道(TopNItem/AggregateBucket);TopN 截断使占比为下界,判据按保守设计并在注释说明
- 静态检查回退:任务模板的 just e-cam-service-* 配方在本机无 justfile,改用 go build/gofmt/go vet(golangci-lint 本机不可用)

## Test Results
- **Tests Executed**: Yes
- **Passed**: 22
- **Failed**: 0
- **Coverage**: 91.8%

## Acceptance Criteria
- [x] 风险分 = 单 IP 最高频率 + 突增倍数 + UA 异常 + 5xx/攻击动作占比加权;权重与阈值集中配置(可校准)
- [x] 风险等级映射(无/低/中/高)边界有单测覆盖(临界分 19/20、39/40、69/70)
- [x] 攻击类型判定 ≥3 类(CC 刷量/恶意爬虫/接口爆破/正常高流量);大促高流量判低风险不误报
- [x] 措施映射:按判定输出 ≥3 条具体文案(封禁 Top 攻击源 IP+请求数、开启 CC 防护/限流、异常 UA 黑名单)
- [x] 纯函数(无网络/DB),输入为聚合行结构与前窗对比值,单测覆盖率 ≥80%

## Notes
测试 22 项全过(10 个测试函数含 12 个子测试),go test -cover 报告 91.8%。回归:logquery service/web 包测试通过,go build ./... 通过。编译首次因环境内存不足链接器 OOM(errno=1455),-p=1 限流后通过,非代码问题。Hard Rules 遵守:仅改 e-cam-service 仓、零新增数据面、无 LLM/外部调用。
