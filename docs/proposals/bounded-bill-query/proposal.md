---
created: "2026-09-17"
author: "Haven"
status: Approved
intent: "enhancement"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: 成本账单查询分页化（消除无界全量加载）

## Problem

`internal/cam/cost` 的 5 个生产调用点以 `Limit=0` 调用 `ListUnifiedBills`，MongoDB `Find` 无分页上限 + `cursor.All` 把整月账单全量载入内存，配合服务端逐账单×规则嵌套循环，数据量增长时内存尖峰与任务耗时不可控。

### Evidence

- `allocation/service.go:135` `AllocateCosts`：`ListUnifiedBills{TenantID, StartDate, EndDate}` 无 Limit，随后 `for bill` × `for rule` 嵌套处理全部账单。
- `optimizer/service.go:150, 220, 231, 304`：4 处 `ListUnifiedBills` 均无 Limit（grep 确认无 `Limit:`）。
- DAO `bill.go:148`：`Limit > 0` 才 `SetLimit`，`Limit=0` 即无界；`ListUnifiedBills` 本身已支持 Offset/Limit（修复无需改 DAO 签名）。
- `cursor.All` 一次性解码全量到切片。

### Urgency

- 无 OOM 故障记录，属规模放大前的预防性加固。
- 不做的代价：账期账单越多（多账号 × 整月 × 明细行），AllocateCosts/优化任务的常驻内存与执行时长线性上涨。
- 用户已确认（/quick 范围选择「5 处分页」）。

## Proposed Solution

5 处 `ListUnifiedBills` 全量调用改为**分页遍历**：复用 DAO 既有 Offset/Limit 按页（每页 500）取数并处理，直至取回页数少于页大小为止。语义不变（每张账单仍被处理），内存有界。

抽共享分页 helper（`cost/repository` 层）承载「页循环 + 回调」，5 处调用点各传处理回调，避免 5 份重复页循环。

### Innovation Highlights

直白的服务端分页，无创新点；价值在不动 DAO/接口、不改语义的前提下把无界加载变为有界流式。

## Requirements Analysis

### Key Scenarios

- 整月账单 > 500：循环取第 2、3… 页直至处理完，账单无遗漏、无重复。
- 整月账单 ≤ 500：单页取完即止（与现状结果一致）。
- 页间有新账单写入（并发采集）：以"取回的页少于页大小即停"为准，最终一致性，不保证严格快照（与现状全量一次性读取的差异可接受，写入通常发生在采集任务完成后）。

### Non-Functional Requirements

- **内存有界**：任意时刻仅一页账单在内存（加输出积累）。
- **行为等价**：处理结果与全量加载一致（每个 bill 恰好处理一次）。

### Constraints & Dependencies

- 仅改 `internal/cam/cost/{allocation,optimizer,repository}`，不动 DAO 方法签名（复用既有 Offset/Limit）。
- 不触碰 `ListRawBills`（无生产调用方）与 `GetAllocationByDimension/Node`（结果集按维度/期绑定，中等量级，用户确认不在范围）。
- 页大小 500 为常量，可调。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 分页遍历（复用 Offset/Limit） | 本提案 | 不丢账单、内存有界、不动 DAO | 页循环样板需共享 helper 消除 | Adopt（用户已确认） |
| 简单 `SetLimit(N)` 截断 | 备选 | 改动最小 | **丢账单 → 分摊/优化结果错误** | Rejected：语义破坏 |
| Mongo 聚合下推（$lookup 分摊） | 备选 | 彻底流式 | 大改、规则引擎重写、风险高 | Rejected：超出 quick 范围 |
| Do nothing | — | 零改动 | 内存/耗时随数据线性涨 | Rejected：用户已确认加固 |

## Scope

### In Scope

- `allocation/service.go` AllocateCosts 账单获取改分页遍历。
- `optimizer/service.go` 4 处账单获取改分页遍历。
- `cost/repository` 新增共享分页 helper（页循环 + 回调），5 处共用。
- 单测：分页 helper（页大小边界、尾页早停、空结果、恰好整除）+ 一处调用点行为等价回归。

### Out of Scope

- `ListRawBills`（无生产调用方，疑似死代码，另行清理）。
- `GetAllocationByDimension/Node` 上限。
- collector 逐账号并发化（独立优化项）。
- 其它模块（logquery/cert）无界查询。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 分页循环漏页/重复（边界 bug）→ 账单漏处理 | Low | 页循环以"取回数 < 页大小"早停；单测覆盖整除/尾页/空；页间用 Offset 稳定排序（billing_date+ctime 已有排序） |
| 分页中途有并发写入 → 偏移漂移 | Low | 排序稳定 + 以尾页早停为准；写入发生在采集完成后的窗口，实际影响可忽略 |
| 共享 helper 过度抽象 | Low | helper 仅封装页循环，回调保持纯函数；5 处调用点传处理闭包 |

## Success Criteria

- [ ] 5 处 `ListUnifiedBills` 调用全部经分页 helper，grep 无残留的直接全量调用（审计：`Limit=0` 的调用点清零）。
- [ ] 分页 helper 单测覆盖：单页/多页/恰好整除/尾页早停/空结果，行为等价（每 bill 恰处理一次）。
- [ ] 一处调用点（AllocateCosts）既有测试零改动通过；`go test ./internal/cam/cost/...` 全绿。
- [ ] 全仓 `go build ./...` 通过。
- [ ] 人工抽查：页大小 500、早停条件为取回 < 500。