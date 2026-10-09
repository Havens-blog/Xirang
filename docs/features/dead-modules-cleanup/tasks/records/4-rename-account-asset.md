---
status: "completed"
started: "2026-09-09 13:08"
completed: "2026-09-09 13:12"
time_spent: "~4m"
---

# Task Record: 4 account/asset 正名：删死外壳 + 按代码事实重写 doc.go

## Summary
RESUME verification of already-committed work (0a7bcbc): deleted zero-importer shells internal/account/module.go and internal/asset/module.go, rewrote both doc.go with code-fact consumer lists. Verified: module.go files absent in working tree and recorded as deletions in commit 0a7bcbc (69 + 95 lines removed); fresh post-deletion import-level grep confirms one-to-one file-level match with doc.go lists (account: 14 consumer files; asset: 5 consumer files, zero extra/missing). account doc.go registers write path SyncAccount via taskQueue.Submit (internal/account/service/account.go:338, anchor re-verified) and bad path mcp nil-queue injection (internal/mcp/deps.go:59, anchor re-verified). Both doc.go link decision 0001 as sole source of truth. go build ./... exit 0; go test -count=1 ./internal/account/... ./internal/asset/... ./internal/cert/domain/ all green (cert/domain 32 pass, 0 fail — previously red TestBsonTagsMatchSchema/K8sCredential fixed by a15ebc0). No new commit created.

## Changes

### Files Created
无

### Files Modified
- internal/account/doc.go
- internal/asset/doc.go
- internal/account/module.go
- internal/asset/module.go

### Key Decisions
- No new commit: deliverable already present in commit 0a7bcbc; verification-only resume
- Consumer lists in doc.go re-validated against fresh import-level grep (file-level one-to-one)

## Test Results
- **Tests Executed**: Yes
- **Passed**: 32
- **Failed**: 0
- **Coverage**: 95.2%

## Acceptance Criteria
- [x] internal/account/module.go、internal/asset/module.go 不存在
- [x] 删除后 import 级 grep 与 doc.go 消费方清单文件级一一对应（account 14 文件、asset 5 文件，无多列无漏列）
- [x] account doc.go：跨域账号访问层（以读为主）定位 + 写路径 SyncAccount 经 taskQueue.Submit（account.go:338）+ 坏路径 mcp nil 队列（deps.go:59）
- [x] asset doc.go：cert 域覆盖率分母只读聚合数据源定位 + 直写 c_instance 历史包袱登记随决策 0001 Phase 4 决断
- [x] 两份 doc.go 显式链接决策 0001（0001-cmdb-consolidation-keep-ecmdb.md）为唯一事实源

## Notes
Gate previously blocked only by internal/cert/domain TestBsonTagsMatchSchema/K8sCredential — pre-existing baseline red since fixed by commit a15ebc0 (schema.sql + test fixture aligned). Working tree contains unrelated in-flight changes (grpc removal etc.) not belonging to this task; all verification ran against current tree with those changes present and passed.
