---
id: "T-review-doc"
title: "Review Documentation Quality"
priority: "P1"
estimated_time: "30min"
dependencies: ["4"]
type: "doc.review"
surface-key: ""
surface-type: ""
---

Review documentation quality for the remove-ecmdb-stack feature (quick mode).

## Acceptance Criteria Summary

The following acceptance criteria are pre-extracted from doc tasks. Use these as the review baseline.

### 4-supersede-cmdb-decision
- [ ] 决策 0001 的状态由「已决策」改为「已废弃」，并引用新决策 0003
- [ ] 新决策 0003 记录：能力已废弃（用户 2026-10-08 确认）、eiam 走已发布模块 v1.11.0、e-cam-service 零依赖、nginx 根入口 301 /console/
- [ ] `docs/decisions/` 下无其他仍声明「保留 ecmdb」的落地决策


## Discovery Strategy

Scan ONLY the following allowlist of directories for target documents:
- docs/features/remove-ecmdb-stack/ (prd/, design/, testing/, and any subdirectories)
- docs/proposals/remove-ecmdb-stack/

EXCLUDE the following from scanning — do NOT read or process these:
- tasks/ directory (task definitions are not deliverables)
- tasks/records/ directory (execution records are not deliverables)
- manifest.md (build artifact)
- index.json (build artifact)

Only .md files under the allowlist directories are target deliverables.

## Acceptance Criteria

- [ ] All acceptance criteria met
