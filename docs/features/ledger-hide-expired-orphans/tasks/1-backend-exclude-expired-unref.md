---
id: "1"
title: "后端 GET /certs 排除无引用过期证书 + 双口径返回"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.enhancement"
mainSession: false
---

# 1: 后端 GET /certs 排除无引用过期证书 + 双口径返回

## Description

台账默认视图不显示「已过期且 referenceStatus=no_refs_scanned（已成功扫描且明确零引用）」的证书——服务端过滤保证分页/total 一致。判定一律以引用三态（has_refs / no_refs_scanned / blind_spot）为准：仅 no_refs_scanned 可隐藏，blind_spot（无成功快照或扫描范围未覆盖）与引用状态未知保守显示。daysLeft=expired 时豁免由后端强制（exclude 参数不改变该语义）。同一响应返回 visibleTotal + hiddenCount 双口径，hiddenCount 按当前 search/hostingStatus 过滤上下文统计。

## Reference Files
- `docs/proposals/ledger-hide-expired-orphans/proposal.md` — In Scope / Constraints & Dependencies / Non-Functional Requirements
- `e-cam-service/internal/cert/web/ledger_handler.go` — GET /certs 端点与过滤参数解析 (ref: ### In Scope)
- `e-cam-service/internal/cert/service/ledger_service.go` — 列表查询与三态判定收敛 (ref: ### Constraints & Dependencies)

## Acceptance Criteria
- [ ] GET /certs 新增过滤参数：排除「已过期且 referenceStatus=no_refs_scanned」的行，默认视图不返回该类行
- [ ] referenceStatus=has_refs 与 blind_spot / 引用状态未知的过期证书在默认视图始终返回（保守显示，绝不隐藏）
- [ ] daysLeft=expired 时后端强制豁免：exclude 参数不改变该筛选语义，返回全部过期证书
- [ ] 同一响应返回 visibleTotal + hiddenCount：hiddenCount 为满足当前 search/hostingStatus 其余过滤条件但被排除的行数（与 visibleTotal 同一过滤上下文）
- [ ] 分页 total、行数、hiddenCount 三者在该上下文内闭合（前端不推算任何计数）
- [ ] 后端单测覆盖：has_refs / no_refs_scanned / blind_spot / 引用状态未知 四种过期证、daysLeft=expired 豁免、hiddenCount 按上下文重算

## Hard Rules
仅修改 e-cam-service `internal/cert` 列表查询路径文件（web/ledger_handler.go、service/ledger_service.go 及必要 repo/DTO）；引用三态语义已固化（reference_status.go），不得用裸 refCount 作隐藏判定（refCount=0 混淆 no_refs_scanned 与 blind_spot）。

## Implementation Notes
- referenceStatus 为 DTO 派生字段：需定死 offset/limit 分页下的服务端过滤落点（子查询下推 or 内存过滤后重算 total），实现中注明；参考 eval 残留攻击 #2/#3。
- 三态派生需引入最近 done 快照 CoverageMeta/引用状态判定，承认该成本（eval 残留 #3）。
- 判定逻辑收敛在 ledger_service 一处（风险表 #7 缓解），便于未来语义调整。
- daysLeft=expired 豁免为后端强制，前端无豁免逻辑（杜绝 URL 直达绕过）。