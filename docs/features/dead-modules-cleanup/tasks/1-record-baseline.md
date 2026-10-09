---
id: "1"
title: "记录清理前基线（build/vet/test 失败集合留档）"
priority: "P0"
estimated_time: "0.5h"
dependencies: []
mainSession: false
type: "doc"
---

# 1: 记录清理前基线（build/vet/test 失败集合留档）

## Description
清理执行前必须先钉住验收基线：e-cam-service 当前工作区（含 79 项在途变更）的 build/vet/test 状态。提案实测基线并非全绿（internal/cert/domain 的 TestBsonTagsMatchSchema/K8sCredential 存量红灯），SC-5 的验收判据是"删除后失败集合 ⊆ 基线失败集合"，因此基线失败集合必须先留档。

## Reference Files
- `docs/proposals/dead-modules-cleanup/proposal.md` — Constraints & Dependencies; Key Scenarios; Success Criteria (SC-5)

## Acceptance Criteria
- [ ] 在 D:\Haven\e-cam-service 运行 `go build ./...` 与 `go vet ./...`，结果留档（预期均通过）
- [ ] 运行 `go test -count=1 ./...`（单次连续运行，禁止重跑取绿），失败测试清单逐条留档（含 internal/cert/domain 的 TestBsonTagsMatchSchema/K8sCredential）
- [ ] 留档文件记录执行时间与工作区状态说明（在途变更的脏工作区）

## Affected Files

### Create
| File | Description |
|------|-------------|
| `docs/features/dead-modules-cleanup/baseline-record.md` | 基线留档：build/vet/test 失败集合与运行环境说明 |

### Modify
| File | Description |
|------|-------------|
| （无） |  |

## Implementation Notes
- 输出基于 Go 工具链，在 D:\Haven\e-cam-service 下执行
- 若测试出现同一测试两次结果不一致（flaky），按基线失败集合成员登记并注明
- 留档是任务 6（全量验收）的比对依据，务必在删除类任务开始前完成
