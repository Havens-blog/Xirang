---
id: "5"
title: "决策文档 0001 Phase 4 补记 asset 归属决断项"
priority: "P1"
estimated_time: "0.5h"
dependencies: []
mainSession: false
type: "doc"
---

# 5: 决策文档 0001 Phase 4 补记 asset 归属决断项

## Description
在 CMDB 收敛决策文档（工作区根 `D:\Haven\docs\decisions\0001-cmdb-consolidation-keep-ecmdb.md`，非 e-cam-service 仓库内）的 Phase 4 节补记 `internal/asset` 归属决断项：该目录直写 c_instance 的历史包袱与收编/删除归属随 CMDB 收敛 Phase 4 一并决断，防止遗忘（与 account doc.go 的链接形成双向锚点）。

## Reference Files
- `docs/proposals/dead-modules-cleanup/proposal.md` — Proposed Solution (第 4 条); In Scope; Success Criteria (SC-4)

## Acceptance Criteria
- [ ] 决策文档 Phase 4 节含 `internal/asset` 归属决断记录（直写 c_instance 包袱 + 收编/删除决断随 Phase 4）
- [ ] 记录注明坐标系：该文档位于工作区根 `D:\Haven\docs\decisions\`，不在 e-cam-service 仓库内

## Affected Files

### Create
| File | Description |
|------|-------------|
| （无） |  |

### Modify
| File | Description |
|------|-------------|
| `D:\Haven\docs\decisions\0001-cmdb-consolidation-keep-ecmdb.md` | Phase 4 节追加 asset 归属决断项 |

## Implementation Notes
- 仅追加，不改写决策 0001 既有结论（Constraints：本清理不改变其任何结论）
- 本任务与任务 3/4 无依赖，可并行；但需在任务 6（全量验收）前完成
