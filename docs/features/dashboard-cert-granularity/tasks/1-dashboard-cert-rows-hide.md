---
id: "1"
title: "后端 GET /certs/dashboard 证书行化 + 服务端孤儿隐藏 + 双口径"
priority: "P0"
estimated_time: "3h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: true
type: "coding.enhancement"
mainSession: false
---

# 1: 后端 GET /certs/dashboard 证书行化 + 服务端孤儿隐藏 + 双口径

## Description

将看板聚合从"域名粒度（同域多证仅取 notAfter 最新者独占一行）"重铸为"证书粒度（每证一行）"：items 由证书循环直出，同域新旧两证同时成行、独立分档；行级下发 referenceStatus；孤儿（过期且 no_refs_scanned 且扫描快照在新鲜度阈值内）由服务端隐藏，has_refs/blind_spot 恒可见；countsByLevel 升级为每档 {total, visible, hidden} + 全局 hiddenCount，与 items 同一快照计算；探测徽标按最差优先序行内聚合；wildcardSkippedCount 按 domain 去重；items 按 notAfter asc（并列指纹字典序）排序。

## Reference Files
- `docs/proposals/dashboard-cert-granularity/proposal.md` — In Scope / Constraints & Dependencies / Non-Functional Requirements / Success Criteria
- `e-cam-service/internal/cert/service/dashboard_service.go` — owner 域名收敛改证书直出、CountsByLevel 升级 (ref: ### In Scope)
- `e-cam-service/internal/cert/web/dashboard_handler.go` — items VO 与 includeHidden 参数 (ref: ### In Scope)
- `e-cam-service/internal/cert/service/ledger_service.go` — 三态判定 + 扫描新鲜度阈值（复用不重写） (ref: ### Constraints & Dependencies)

## Acceptance Criteria
- [ ] items 每证书一行：includeHidden=true 时长度 == 全部证书数；单测断言同域新旧两证同时成行、独立分档，每行携带 referenceStatus
- [ ] 隐藏谓词 = expired AND referenceStatus=no_refs_scanned AND 扫描快照在新鲜度阈值内；has_refs 与 blind_spot（引用状态未知）默认视图恒可见（单测以裸 refCount=0 反例锁定不得用 refCount 判定）
- [ ] 快照超新鲜度阈值时降级为保守显示（不隐藏），并下发展示最近扫描时间
- [ ] countsByLevel 每档 {total, visible, hidden}（total=visible+hidden）+ 全局 hiddenCount，与 items 同一快照计算（单测断言 includeHidden=true 下各档 total 之和 == items.length）
- [ ] 探测徽标按 diff > change_linked_diff > unreachable > wildcard_skipped > exempt > 健康（未探测排序尾）聚合到行；wildcardSkippedCount 按 domain 去重（共享通配符 SAN 两证计数 1）
- [ ] items 按 notAfter asc 排序、并列按证书指纹字典序稳定取序；go test ./... 全绿

## Hard Rules
复用台账 ledger 的三态判定与扫描新鲜度阈值（提取共享 helper，编译/单测锁定无 import cycle），不得在看板侧另写判定；隐藏不得用裸 refCount 作谓词；仅修改 `internal/cert` 的 dashboard 路径文件与必要的共享 helper 提取。

## Implementation Notes
- eval 残留 #10（扫描新鲜度契约）：隐藏谓词必须复配台账的 scanFreshness 阈值，超阈值保守显示、不静默隐藏，否则最近获得引用的过期证会被误隐藏。
- 三态派生批量内存归约（一次拉 refs + 最近 done 快照），仅对「过期且 refCount=0」候选派生，逐证 N+1 为反例（NFR 性能）。
- countsByLevel 与 items 必须同一快照内收敛，避免两遍读取漂移（SC 同快照断言）。
- 破坏性变更：items 载荷结构改变，与前端同 release 协同上线（见风险表）；加载荷契约回归测试持续断言 items 形状。