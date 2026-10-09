---
status: "completed"
started: "2026-09-19 16:28"
completed: "2026-09-19 17:03"
time_spent: "~35m"
---

# Task Record: 2 后端诊断接口 POST /logs/diagnose(当前窗+前窗聚合,调规则引擎,预留 summary)

## Summary
新增后端诊断接口 POST /api/v1/cam/logs/diagnose(WAF 专属):服务层 Diagnose 复用联邦聚合通道,当前窗并发 4 维聚合(client_ip/user_agent/status/action),前一等长窗口仅 1 帧 client_ip 聚合(一次产出精确 Total + Top IP,不放大扫描),拼装 DiagnoseInput 调任务 1 规则引擎;响应含风险分/等级/疑似攻击类型/Top5 攻击源(请求数+占比)/措施(≥3)/突增倍数/per-source 状态/降级标注/summary 空串占位(不引入 LLM);诊断级 SWR 缓存与 Cached/CacheStale 标注同 search/aggregate;SLB/CDN 明确 400 拒绝;诊断结果(duration_ms/风险等级)落日志

## Changes

### Files Created
- e-cam-service/internal/logquery/service/diagnose.go
- e-cam-service/internal/logquery/service/diagnose_test.go
- e-cam-service/internal/logquery/web/diagnose_test.go

### Files Modified
- e-cam-service/internal/logquery/web/handler.go
- e-cam-service/internal/logquery/diagnose/diagnose.go
- e-cam-service/internal/logquery/diagnose/diagnose_test.go

### Key Decisions
- 前窗仅 1 帧:client_ip 维度一次聚合同时产出精确 Total(分桶求和)与 Top client_ip,满足『≤2 窗口帧、不放大 SLS 扫描量』硬规则;前窗等长窗口由服务层推导(start-span, start),不可指定
- 聚合层源级失败被隔离(Sources[].Error)而非调用级错误 —— 前窗全源失败按『前窗不可用』降级(Prev=nil+PrevError=去重源错误),源正常但窗口为空保留空前窗(Prev={0,0}),两者均降级为当前窗绝对量判定不报错
- 突增倍数落引擎 DiagnoseResult.SurgeMultiplier(与风险分突增分量同源,防响应与评分口径漂移);Total 倍数与 Top IP 倍数取大
- 非主维度(UA/status/action)失败仅记 DimensionNotes(该维度判据退化),主帧 client_ip 失败才整体报错,防误导性『无风险』结论
- 路由挂既有 /api/v1/cam/logs 组(RequireTenant 同边界);诊断级 cachedCall(同参 60s 命中,测试验证命中不重放聚合帧)

## Test Results
- **Tests Executed**: Yes
- **Passed**: 33
- **Failed**: 0
- **Coverage**: 75.8%

## Acceptance Criteria
- [x] POST /logs/diagnose(WAF)返回风险等级、疑似攻击类型、Top5 攻击源 IP(含请求数与占比)、来源分布、≥3 条措施、summary 空字段、per-source 状态
- [x] 前窗对比:前一等长窗口执行一次聚合(Total+Top client_ip),输出突增倍数;前窗无数据降级为当前窗绝对量判定,不报错
- [x] 聚合失败源 per-source 状态标注,已聚合源照常判定
- [x] 单次诊断 ≤2 窗口帧(当前窗+前窗各一),前窗仅 1 帧聚合;鉴权/租户边界与现有 /logs/* 一致(同组 RequireTenant,403 验证)
- [x] go build ./internal/logquery/... 与相关单测全绿;明细/聚合既有行为回归通过

## Notes
覆盖率(runner 实测,logquery 域合并):diagnose 引擎 91.8%、service 69.9%、web 72.2%,总计 75.8%;新增 service/diagnose.go 主路径 Diagnose 94.1% / diagnoseUncached 97.5%(未覆盖为防御性 nil 分支)。测试含 403 租户边界、SLB 明确 400、前窗恰 1 帧(5 次聚合帧断言)、同参缓存不重放。go build ./... 全仓通过;仓库存在大量并行任务脏文件,提交仅限本任务 6 个文件(显式 pathspec)。
