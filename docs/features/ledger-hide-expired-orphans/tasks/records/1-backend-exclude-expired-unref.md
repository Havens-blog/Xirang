---
status: "completed"
started: "2026-09-21 11:04"
completed: "2026-09-21 11:04"
time_spent: ""
---

# Task Record: 1 后端 GET /certs 排除无引用过期证书 + 双口径返回

## Summary
GET /certs 默认排除「已过期且 referenceStatus=no_refs_scanned」行并返回 visibleTotal+hiddenCount 双口径：service 层新增隐藏路径（listCertsHideExpiredNoRefs，三态判定复用 deriveRefStatusFor，禁止裸 refCount 判定；daysLeft=expired 后端强制豁免），handler 新增 includeExpiredNoRefs 布尔参数与 hiddenCount 字段；新增 3 个测试覆盖四态过期证/豁免/上下文重算/分页闭合/参数校验

## Changes

### Files Created
无

### Files Modified
- e-cam-service/internal/cert/service/ledger_service.go
- e-cam-service/internal/cert/service/ledger_service_test.go
- e-cam-service/internal/cert/web/ledger_handler.go
- e-cam-service/internal/cert/web/ledger_handler_test.go

### Key Decisions
- 服务端过滤落点：referenceStatus 为 DTO 派生量（需最新成功快照+跨快照历史引用+CoverageMeta 跨集合三态推导），无法下推单集合子查询——隐藏生效时 ListPage(limit<=0) 取当前筛选全量命中行，内存完成「过期 && no_refs_scanned」排除后重算 total 并按仓储同序（notAfter asc, _id asc）切片分页
- 隐藏判定收敛在 ledger_service 一处（风险表 #7 缓解）：候选仅限「已过期且快照计数 0」行，复用 deriveRefStatusFor 区分 no_refs_scanned/blind_spot；无成功快照视为引用状态未知保守显示
- 豁免落点 service 层：q.DaysLeft==expired 恒走直查路径（HiddenCount=0），includeExpiredNoRefs 参数不改变 expired 语义，杜绝 URL 直达绕过
- 仅对 no_refs_scanned 候选行惰性拉取历史引用（fetchHistory），has_refs/未过期行零额外查询，控制三态派生成本（eval 残留 #3）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 376
- **Failed**: 0
- **Coverage**: 77.1%

## Acceptance Criteria
- [x] GET /certs 新增过滤参数，默认视图不返回「已过期且 no_refs_scanned」行
- [x] has_refs / blind_spot / 引用状态未知的过期证书默认视图始终返回（保守显示）
- [x] daysLeft=expired 后端强制豁免：include 参数不改变该筛选语义，返回全部过期
- [x] 同一响应返回 visibleTotal+hiddenCount，hiddenCount 按当前 search/hostingStatus 上下文统计
- [x] 分页 total、行数、hiddenCount 同上下文闭合（前端不推算）
- [x] 后端单测覆盖四态过期证、daysLeft=expired 豁免、hiddenCount 上下文重算

## Notes
coverage 为 service 包 77.1%（web 包 79.9%），达 60% 目标即停。just lint 的 3 个 error 均为 e-cam-web 既有文件（cost/statistics/task.ts 空接口），与本次后端改动无关；just fmt 对本任务文件无变更（仓库既有 662 个脏文件非本次产生）。
