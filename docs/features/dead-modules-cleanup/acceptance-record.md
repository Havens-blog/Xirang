# 验收留档：dead-modules-cleanup 清理后全量验收（SC-5）

- **执行日期**：2026-09-09
- **任务**：dead-modules-cleanup 任务 6（全量验收：失败集合子集比对）
- **Go 工作区**：`D:\Haven\e-cam-service`（go1.24.1 windows/amd64）
- **基线记录**：`docs/features/dead-modules-cleanup/baseline-record.md`（4 包 6 测试）
- **清理提交**：`dc3eee0`（删除三死模块与三只 tmp 工具目录）、`0a7bcbc`（account/asset 正名）、`a15ebc0`（对齐 K8sCredential bson 标签与 schema.sql）

## 构建与静态检查

| 命令 | 结果 |
|------|------|
| `go build ./...` | **通过**（exit 0，无输出） |
| `go vet ./...` | **通过**（exit 0，无输出） |

## 全量测试（`go test -count=1 -p 2 ./...`，单次连续运行）

- **运行口径**：范围全量 `./...`、无 `-run` 过滤、`-count=1` 禁缓存、单次连续运行。`-p 2` 仅为规避工具链并行编译测试二进制时的 OOM（与基线第 2 次权威运行口径一致，属并行度调整而非选择性重跑）。
- **结果汇总**：63 个包 ok，3 个包 FAIL，5 个失败测试（另 97 个包无测试文件）。无 `[build failed]` 项，未发生 OOM 作废，无需重跑。
- **未出现 flaky**：无任何测试被重跑，单次运行结果即最终结果。

## 失败集合比对

### 清理后失败集合（3 包 5 测试）

| # | 包 | 失败测试 | 与基线成员对应 |
|---|----|----------|----------------|
| 1 | `internal/cam/cost/analysis` | `TestGetCostSummary_DateRanges` | 基线 #1 |
| 2 | `tests/no-snapshot-guidance` | `TestNoSnapshotGuidance_Smoke` | 基线 #3 |
| 3 | `tests/no-snapshot-guidance` | `TestStep2_TriggerScan_Success` | 基线 #4 |
| 4 | `tests/placeholder-fingerprint-backfill` | `TestStep4_BackfillReferences_WrongBackfillRescanRecoverable` | 基线 #5 |
| 5 | `tests/placeholder-fingerprint-backfill` | `TestStep5_VerifyReferenceList_PartialFailureRescanRefresh` | 基线 #6 |

### 基线失败集合（4 包 6 测试，摘自 baseline-record.md）

1. `internal/cam/cost/analysis` — `TestGetCostSummary_DateRanges`
2. `internal/cert/domain` — `TestBsonTagsMatchSchema/K8sCredential`
3. `tests/no-snapshot-guidance` — `TestNoSnapshotGuidance_Smoke`、`TestStep2_TriggerScan_Success`
4. `tests/placeholder-fingerprint-backfill` — `TestStep4_BackfillReferences_WrongBackfillRescanRecoverable`、`TestStep5_VerifyReferenceList_PartialFailureRescanRefresh`

### 子集判定

以"包 + 测试名"为集合成员比对：清理后失败集合的每一项均为基线失败集合成员，**无任何新增失败**；基线 #2（`internal/cert/domain` 的 `TestBsonTagsMatchSchema/K8sCredential`）已由提交 `a15ebc0` 修复，本次运行该包结果为 `ok`，属预期内的"减少"。

**结论：清理后失败集合 ⊆ 基线失败集合（5 ⊆ 6，减少 1 项），判定通过。**

### 值得留意的通过项

- `internal/cert/domain` 由基线 FAIL 转为 **ok**：存量红灯 `TestBsonTagsMatchSchema/K8sCredential`（bson 端多出 `displayName`）已通过 `a15ebc0` 对齐 schema.sql 修复，是本次验收唯一较基线减少的失败项。
- 其余 5 项失败与基线现象一致（cost 日期区间断言 3 项实得 2 项；扫描触发 expected 200 / actual 202 的异步语义偏差），本次纯删除任务不触碰这些路径，按预期原样保留。

## 最终验收结论

| 验收项 | 结果 |
|--------|------|
| `go build ./...` 通过 | ✅ |
| `go vet ./...` 通过 | ✅ |
| 失败集合 ⊆ 基线失败集合 | ✅（5 ⊆ 6，无新增） |
| 比对结果留档 | ✅（本文件） |

**dead-modules-cleanup 提案 SC-5 验收通过。**
