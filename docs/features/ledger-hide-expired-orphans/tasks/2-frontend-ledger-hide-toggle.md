---
id: "2"
title: "前端台账默认隐藏 + 行内提示切换 + 状态机"
priority: "P1"
estimated_time: "2h"
complexity: "medium"
dependencies: [1]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 2: 前端台账默认隐藏 + 行内提示切换 + 状态机

## Description

台账默认应用服务端过滤（任务 1 交付），前端消费响应的 visibleTotal / hiddenCount 双口径：工具栏显示「已隐藏 N 张无引用过期证书 · 查看全部」，点击切换为显示全部（提示变「再次点击隐藏无引用过期证书」），再点恢复隐藏。定义切换态×筛选态状态机（切换态仅作用默认视图；daysLeft=expired 时挂起显示全部；清除筛选后恢复；切换导致 total 变化重置到第 1 页）；空态视图保留提示条与「查看全部」入口。

## Reference Files
- `docs/proposals/ledger-hide-expired-orphans/proposal.md` — In Scope / Key Scenarios / Success Criteria
- `e-cam-web/src/views/cert/ledger/index.vue` — 列表筛选状态源与总数消费 (ref: ### Key Scenarios)
- `e-cam-web/src/views/cert/ledger/components/CertTable.vue` — 工具栏筛选与表格渲染 (ref: ### In Scope)
- `e-cam-web/src/views/cert/ledger/format.ts` — 展示纯逻辑（filter/文案）与 vitest (ref: ### Success Criteria)

## Acceptance Criteria
- [ ] 默认视图使用服务端 visibleTotal/hiddenCount：不显示被排除的无引用过期证书，总数与行一致
- [ ] 工具栏显示「已隐藏 N 张无引用过期证书 · 查看全部」，点击显示全部（提示变「再次点击隐藏无引用过期证书」），再点恢复隐藏，计数与行实时一致
- [ ] 切换导致 total 变化时重置到第 1 页
- [ ] daysLeft=expired 筛选或 /certs?daysLeft=expired 跳转时切换态挂起（显示全部、不显示隐藏提示），清除筛选后恢复切换态
- [ ] 空态视图仍显示「已隐藏 N 张…」提示条与「查看全部」入口（提示条不随表格主体消失）
- [ ] 前端 vitest 覆盖（format/filter + 切换交互 + 状态机）+ vue-tsc 通过

## Hard Rules
canonical 文案全文统一两字符串：默认态「已隐藏 N 张无引用过期证书 · 查看全部」/ 展开态「再次点击隐藏无引用过期证书」；不得静默隐藏（无提示不得吞行）。

## Implementation Notes
- 隐藏逻辑与其他筛选（search/hostingStatus）正交：先按其余条件过滤，再应用隐藏；hiddenCount 随该上下文联动（服务端已算）。
- 状态机：切换态（展开/隐藏）仅作用于默认视图；daysLeft=expired 时挂起；清筛恢复；total 变化重置第 1 页。
- 被隐藏行含「可删除」清理对象（如 *.jlc-robot.com），查看入口是其唯一发现路径——提示文案可保留「（可删除）」属性（proposal Key Scenarios 已记录该取舍）。
- 组合过滤场景为空/边界：空态下提示条保留。