---
id: "3"
title: "删除三死模块与三只 tmp 工具目录"
priority: "P0"
estimated_time: "1h"
complexity: "high"
dependencies: [2]
surface-key: ""
surface-type: "api"
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 3: 删除三死模块与三只 tmp 工具目录

## Description
删除零外部 importer 的三个悬空模块与已留档的三只 tmp 工具：`internal/task`、`internal/sync`、`internal/servicetree`、`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`、`cmd/tmp-k8s-diag`（整目录）。同名活跃包 `internal/cam/{task,sync,servicetree}` 不得波及。删除单独成 commit。

## Reference Files
- `docs/proposals/dead-modules-cleanup/proposal.md` — In Scope; Key Scenarios; Success Criteria (SC-1/SC-2); Risks & Mitigations (Risk 1/2/6/7)

## Acceptance Criteria
- [ ] `internal/task`、`internal/sync`、`internal/servicetree`、`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`、`cmd/tmp-k8s-diag` 六个目录均不存在
- [ ] import 语句级 grep：`internal/task`、`internal/sync`、`internal/servicetree` 零命中（`internal/cam/*` 同名活跃路径除外）
- [ ] `go build ./...` 与 `go vet ./...` 通过（活跃模块完好）
- [ ] 删除单独成 commit（不与在途变更混提交），便于整体 revert
- [ ] 删除 commit 的描述列明 10 处悬空注释清单（`eval/followup.md`）并注明"待 follow-up 清除"

## Hard Rules
- 仅删除以下目录：`internal/task`、`internal/sync`、`internal/servicetree`、`cmd/tmp-cas-diag`、`cmd/tmp-k8s-cred`、`cmd/tmp-k8s-diag`；不得改动 `internal/cam/{task,sync,servicetree}` 与任何 live 代码
- 不改 Mongo 数据与索引

## Implementation Notes
- 判死标准是 import 语句级（排除注释）；若删除中途 build 红灯暴露隐藏消费者，revert 恢复后停下重查 Evidence 结论（Key Scenarios 处置）
- 死模块已实测无 init() 副作用面（零外部 importer），全量测试由任务 6 兜底
- 10 处注释级悬空引用本轮不动（Out of Scope，清单见 `eval/followup.md`）
