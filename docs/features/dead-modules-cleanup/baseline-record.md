# 基线留档：dead-modules-cleanup 清理前 build/vet/test 失败集合

- **执行日期**：2026-09-09
- **任务**：dead-modules-cleanup 任务 1（记录清理前基线）
- **Go 工作区**：`D:\Haven\e-cam-service`（go1.24.1 windows/amd64）
- **用途**：提案 SC-5 的验收比对基准——删除后 `go test ./...` 失败集合必须是本记录失败集合的**子集**（只允许减少、不允许新增）。

## 工作区状态

**脏工作区，含 79 项在途变更**（`git status --short` 实测）：

- 63 项已跟踪文件修改/删除（`M`/`D`），包括 grpc/protobuf 移除相关（`api/proto/**`、`pkg/grpcx/**`、`ioc/grpc.go`、`ioc/ecmdb.go`、`internal/shared/middleware/check_policy*.go` 等）、cloudx IAM 多云适配改动、`cmd/start/start.go`、`go.mod`/`go.sum` 等
- 12 项未跟踪新增（`??`）：`cmd/tmp-cas-diag/`、`cmd/tmp-k8s-cred/`、`cmd/tmp-k8s-diag/`、`internal/shared/cloudx/iam/registry*.go`、`internal/shared/cloudx/providers/`、`ioc/policy.go`、`ioc/gin_test.go`、`go.work.sum` 等

结论：基线反映的是**当前在途变更叠加下的状态**，非干净 main。后续 SC-5 比对时若在途变更已合入/回退，需知悉本基线的这一前提。

## 构建与静态检查

| 命令 | 结果 |
|------|------|
| `go build ./...` | **通过**（exit 0，无输出） |
| `go vet ./...` | **通过**（exit 0，无输出） |

与提案 Feasibility 评估一致（2026-09-09 实测 vet 通过）。

## 全量测试（`go test -count=1 ./...`）

### 运行记录说明（重要）

按 SC-5 口径执行单次连续运行，实际发生两次运行，原因与定性如下：

1. **第 1 次运行**（默认并行度，`go test -count=1 ./...`）：**因环境资源耗尽作废，不作基线依据**。Go 工具链在并行编译测试二进制时 OOM（`runtime: VirtualAlloc of ... bytes failed with errno=1455` / `fatal error: out of memory`），导致 8 个包报 `[build failed]`（cam/cost/budget、cam/dictionary、cam/tag、cam/task/executor、cam/template、cert、cert/deployer、cert/scheduler、tests/no-snapshot-guidance、tests/permission-boundary、tests/placeholder-fingerprint-backfill、tests/unsupported-entries-skip 中的 8 项）。这些包的测试**从未执行**，属瞬态环境失败而非代码/测试失败，故登记为运行作废而非基线失败成员。该次运行中已实际执行的失败项为 TestGetCostSummary_DateRanges 与 TestBsonTagsMatchSchema/K8sCredential。
2. **第 2 次运行**（`go test -count=1 -p 2 ./...`，**权威基线**）：降低包级并行度以规避 OOM，范围仍为全量 `./...`、无 `-run` 过滤、单次连续运行。汇总：**62 个包 ok，4 个包 FAIL，6 个失败测试**（另 114 个包无测试文件）。

两次运行对共同执行的失败项结果一致（TestGetCostSummary_DateRanges、TestBsonTagsMatchSchema/K8sCredential 均失败），**无 flaky 判定**；第 1 次作废运行中 OOM 的包在第 2 次运行均实际执行完毕，无同测试两次结果不一致情形。

### 基线失败集合（权威，以第 2 次运行为准）

| # | 包 | 失败测试 | 失败位置 | 现象摘要 |
|---|----|----------|----------|----------|
| 1 | `internal/cam/cost/analysis` | `TestGetCostSummary_DateRanges` | `service_test.go:422` | 断言日期区间列表应含 3 项，实得 2 项（`[{1 0 2026-09-01 2026-09-09 0 0} {1 0 2026-08-01 2026-08-31 0 0}]`） |
| 2 | `internal/cert/domain` | `TestBsonTagsMatchSchema/K8sCredential` | `models_test.go:110` | **提案已登记的存量红灯**：bson 字段与 `schema.sql` 不一致——got 含 `displayName`，want 不含（bson 端多出 `displayName`，schema.sql 脱节） |
| 3 | `tests/no-snapshot-guidance` | `TestNoSnapshotGuidance_Smoke` | `no_snapshot_guidance_smoke_test.go:51` | HTTP 状态码断言 expected 200 / actual 202（响应体 `success:true, status:"running"`，扫描已改异步语义） |
| 4 | `tests/no-snapshot-guidance` | `TestStep2_TriggerScan_Success` | `step2_trigger_scan_test.go:57` | 同上，expected 200 / actual 202 |
| 5 | `tests/placeholder-fingerprint-backfill` | `TestStep4_BackfillReferences_WrongBackfillRescanRecoverable` | `step4_backfill_references_test.go:116` | 同上，expected 200 / actual 202 |
| 6 | `tests/placeholder-fingerprint-backfill` | `TestStep5_VerifyReferenceList_PartialFailureRescanRefresh` | `step5_verify_reference_list_test.go:78` | 同上，expected 200 / actual 202 |

**失败集合汇总（4 包 6 测试）**：

- `github.com/Havens-blog/e-cam-service/internal/cam/cost/analysis` — 1 项
- `github.com/Havens-blog/e-cam-service/internal/cert/domain` — 1 项（`TestBsonTagsMatchSchema/K8sCredential`）
- `github.com/Havens-blog/e-cam-service/tests/no-snapshot-guidance` — 2 项
- `github.com/Havens-blog/e-cam-service/tests/placeholder-fingerprint-backfill` — 2 项

### 判读说明

- #2（cert/domain）为提案明文登记的存量红灯，与本次删除无关，删除后**预期仍在失败集合中**。
- #1、#3~#6 均疑似与在途变更相关（cost 汇总日期区间、扫描触发从同步 200 改为异步 202），属脏工作区前提下的存量失败；本次纯删除任务不触碰这些路径，删除后同样**预期原样保留或减少、不得新增**。
- SC-5 比对时以"包 + 测试名"为集合成员；flaky 处置按提案口径（同一测试两次结果不一致按基线失败集合成员处理，不计新增）。
