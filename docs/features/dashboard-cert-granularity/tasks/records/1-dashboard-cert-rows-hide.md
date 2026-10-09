---
status: "completed"
started: "2026-09-22 12:36"
completed: "2026-09-22 12:42"
time_spent: "~6m"
---

# Task Record: 1 后端 GET /certs/dashboard 证书行化 + 服务端孤儿隐藏 + 双口径

## Summary
GET /certs/dashboard 证书行化落地：items 由域名粒度改证书粒度（同域新旧两证同时成行、独立分档、行级 referenceStatus）；孤儿隐藏谓词 = expired AND no_refs_scanned AND 扫描快照在新鲜度阈值内，复用台账三态判定共享 helper（无 import cycle），不以裸 refCount 判定；快照超阈值降级保守显示并下发最近扫描时间；countsByLevel 升级为每档 {total, visible, hidden} + 全局 hiddenCount，与 items 同一快照计算；探测徽标按最差优先序行内聚合，wildcardSkippedCount 按 domain 去重；items 按 notAfter asc + 指纹字典序稳定排序。

## Changes

### Files Created
无

### Files Modified
- e-cam-service/internal/cert/service/dashboard_service.go
- e-cam-service/internal/cert/service/dashboard_service_test.go
- e-cam-service/internal/cert/service/reference_query_service.go
- e-cam-service/internal/cert/web/dashboard_handler.go
- e-cam-service/internal/cert/web/dashboard_handler_test.go

### Key Decisions
- 三态判定与扫描新鲜度阈值从台账 ledger 提取共享 helper 复用，不另写判定
- 隐藏谓词不使用裸 refCount，以 no_refs_scanned 三态 + 快照新鲜度复合判定
- 三态派生批量内存归约：一次拉 refs + 最近 done 快照，仅对「过期且 refCount=0」候选派生，避免 N+1

## Test Results
- **Tests Executed**: Yes
- **Passed**: 858
- **Failed**: 0
- **Coverage**: 77.4%

## Acceptance Criteria
- [x] items 每证书一行：includeHidden=true 时长度 == 全部证书数；同域新旧两证同时成行、独立分档，每行携带 referenceStatus（单测断言）
- [x] 隐藏谓词 = expired AND no_refs_scanned AND 扫描快照在新鲜度阈值内；has_refs 与 blind_spot 默认视图恒可见（裸 refCount=0 反例单测锁定）
- [x] 快照超新鲜度阈值降级保守显示（不隐藏），下发最近扫描时间
- [x] countsByLevel 每档 {total, visible, hidden} + 全局 hiddenCount，与 items 同一快照计算（单测断言 total 之和 == items.length）
- [x] 探测徽标按最差优先序行内聚合；wildcardSkippedCount 按 domain 去重
- [x] items 按 notAfter asc 排序、并列按证书指纹字典序稳定取序；go test 全绿

## Notes
FINALIZE 执行：实现为前一轮完成并通过验证（曾被无关的 e-cam-web vue-tsc 门禁阻塞，fix-1/a9bd6e5 已修复）。本轮复检 go build ./... 全绿 + go test -p 1 ./internal/cert/... 全绿；提交仅限 5 个改动文件（commit e2be44f，pathspec-limited add）。coverage 取变更包最低值：service 77.4% / web 80.0%（858 为含子测试的 RUN 总数，0 失败）。破坏性载荷变更需与前端同 release 协同上线（任务 2）。
