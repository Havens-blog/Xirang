---
status: "completed"
started: "2026-09-09 13:04"
completed: "2026-09-09 13:08"
time_spent: "~4m"
---

# Task Record: fix-1 Fix: internal/cert/domain TestBsonTagsMatchSchema/K8sCredential 存量红灯

## Summary
对齐 K8sCredential bson 标签与 schema.sql：displayName 由 commit 36157af（feat cert: 集群凭证区分集群名/集群ID）作为持久化字段有意引入（云端拉取回填，幂等仅空缺回填），故判定 schema.sql 为过期侧。cert_k8s_credentials 的 $jsonSchema properties 补 displayName（非 required，omitempty 语义一致），models_test.go 的 K8sCredential 预期清单（转录自 schema.sql）同步补 displayName。

## Changes

### Files Created
无

### Files Modified
- e-cam-service/docs/features/ssl-cert-management/design/schema.sql
- e-cam-service/internal/cert/domain/models_test.go

### Key Decisions
- schema.sql 判定为过期侧而非删除 bson 标签：git log -S 证实 displayName 是 36157af 的功能性新增（登记键 ClusterName 不变，仅展示区分），服务层有回填逻辑
- 测试预期清单是 schema.sql 的转录，故两处同步修改以维持 TestBsonTagsMatchSchema 的契约
- displayName 不加入 required：与 bson omitempty 一致（手动登记时为空）

## Test Results
- **Tests Executed**: Yes
- **Passed**: 6
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
无

## Notes
go test -count=1 ./internal/cert/domain/ 全绿；go build ./... 首次因链接器 OOM(errno=1455) 环境性失败，go build -p 2 ./... 重试通过（BUILD_OK），非代码问题。提交 a15ebc0 仅含 2 个修复文件。
