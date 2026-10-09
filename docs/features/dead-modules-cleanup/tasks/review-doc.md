---
id: "T-review-doc"
title: "Review Documentation Quality"
priority: "P1"
estimated_time: "30min"
dependencies: ["5", "1"]
type: "doc.review"
surface-key: ""
surface-type: ""
---

Review documentation quality for the dead-modules-cleanup feature (quick mode).

## Acceptance Criteria Summary

The following acceptance criteria are pre-extracted from doc tasks. Use these as the review baseline.

### 1-record-baseline
- [ ] 在 D:\Haven\e-cam-service 运行 `go build ./...` 与 `go vet ./...`，结果留档（预期均通过）
- [ ] 运行 `go test -count=1 ./...`（单次连续运行，禁止重跑取绿），失败测试清单逐条留档（含 internal/cert/domain 的 TestBsonTagsMatchSchema/K8sCredential）
- [ ] 留档文件记录执行时间与工作区状态说明（在途变更的脏工作区）


### 5-decision-doc-linkage
- [ ] 决策文档 Phase 4 节含 `internal/asset` 归属决断记录（直写 c_instance 包袱 + 收编/删除决断随 Phase 4）
- [ ] 记录注明坐标系：该文档位于工作区根 `D:\Haven\docs\decisions\`，不在 e-cam-service 仓库内


## Discovery Strategy

Scan ONLY the following allowlist of directories for target documents:
- docs/features/dead-modules-cleanup/ (prd/, design/, testing/, and any subdirectories)
- docs/proposals/dead-modules-cleanup/

EXCLUDE the following from scanning — do NOT read or process these:
- tasks/ directory (task definitions are not deliverables)
- tasks/records/ directory (execution records are not deliverables)
- manifest.md (build artifact)
- index.json (build artifact)

Only .md files under the allowlist directories are target deliverables.

## Acceptance Criteria

- [ ] All acceptance criteria met
