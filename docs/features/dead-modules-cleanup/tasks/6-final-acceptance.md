---
id: "6"
title: "全量验收：失败集合子集比对"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies: [4]
surface-key: ""
surface-type: "api"
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 6: 全量验收：失败集合子集比对

## Description
对照任务 1 留档的基线，执行清理后的全量验收（提案 SC-5）：`go build ./...`、`go vet ./...` 通过；`go test -count=1 ./...` 单次连续运行的失败集合必须是基线失败集合的子集（只允许减少、不允许新增；通过集合作废不比对）。比对结果留档。

## Reference Files
- `docs/proposals/dead-modules-cleanup/proposal.md` — Success Criteria (SC-5); Key Scenarios

## Acceptance Criteria
- [ ] `go build ./...` 与 `go vet ./...` 通过
- [ ] `go test -count=1 ./...` 单次连续运行的失败集合 ⊆ 基线失败集合（对照 `baseline-record.md`）；出现新增失败时先重跑 1 次确认可复现，仍新增即验收失败；同一测试两次结果不一致（flaky）按基线失败集合成员处理，不计新增
- [ ] 比对结果（两个失败集合与判定结论）留档于 `docs/features/dead-modules-cleanup/acceptance-record.md`

## Hard Rules
- 禁止反复重跑取绿（last-run-wins）

## Implementation Notes
- 存量红灯预期仍在（如 internal/cert/domain 的 TestBsonTagsMatchSchema/K8sCredential），属"减少或持平"，不是新增
- 若验收失败：revert 删除 commit 恢复原状（tmp 工具已留档无丢失），排查后重新发起；已通过的验收项不受影响
