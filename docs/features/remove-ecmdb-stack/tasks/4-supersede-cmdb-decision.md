---
id: "4"
title: "决策 0001 标废弃并登记移除 ecmdb 的新决策"
priority: "P1"
estimated_time: "0.5h"
dependencies: []
type: "doc"
mainSession: false
---

# 4: 决策 0001 标废弃并登记移除 ecmdb 的新决策

## Description
`docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md`（保留 ecmdb 为唯一 CMDB）已被本 feature 推翻，将其标为废弃/被取代，并新增「移除 ecmdb + ecmdb-web」的替代决策，记录能力废弃事实与依赖安全判定。

## Reference Files
- `docs/proposals/remove-ecmdb-stack/proposal.md` — 决策反转依据与依赖安全判定（ref: Problem / Assumptions Challenged）
- `docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md` — 标废弃，附新决策引用（ref: Scope）

## Affected Files

### Create
| File | Description |
|------|-------------|
| `docs/decisions/0003-remove-ecmdb.md` | 新决策：移除 ecmdb + ecmdb-web，记录能力已废弃与依赖安全判定 |

### Modify
| File | Changes |
|------|---------|
| `docs/decisions/0001-cmdb-consolidation-keep-ecmdb.md` | 状态字段改为「已废弃（被 0003 取代）」，附指向说明 |

### Delete
| File | Reason |
|------|--------|
| （无） | — |

## Acceptance Criteria
- [ ] 决策 0001 的状态由「已决策」改为「已废弃」，并引用新决策 0003
- [ ] 新决策 0003 记录：能力已废弃（用户 2026-10-08 确认）、eiam 走已发布模块 v1.11.0、e-cam-service 零依赖、nginx 根入口 301 /console/
- [ ] `docs/decisions/` 下无其他仍声明「保留 ecmdb」的落地决策

## Implementation Notes
- 决策编号按现有 `docs/decisions/` 顺序登记（当前已有 0001、0002，新决策用 0003）。
- 仅标废弃、不删除 0001 原文（保留历史审计线索）；在 0001 顶部加「已废弃」横幅即可。