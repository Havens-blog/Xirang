---
id: "3"
title: "五份 containsString 改用 stdlib slices.Contains"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 3: 五份 containsString 改用 stdlib slices.Contains

## Description
`internal/shared/cloudx/{aliyun/nas.go, aws/efs.go, huawei/sfs.go, tencent/cfs.go, volcano/tos.go}`
五份相同的私有 `containsString(slice []string, s string) bool`（线性扫描，与 `slices.Contains` 语义一致）
替换调用点为 stdlib `slices.Contains`，删除五个私有函数。

## Reference Files
- `docs/proposals/cloudx-template-consolidation/proposal.md` — Proposed Solution, Success Criteria
- `internal/shared/cloudx/aws/efs.go` — containsString 定义 + 调用点
- `internal/shared/cloudx/volcano/tos.go` — containsString 定义 + 调用点

## Acceptance Criteria
- [ ] 五个文件（nas/efs/sfs/cfs/tos）的 containsString 调用点全部改为 `slices.Contains`
- [ ] 五个私有 `func containsString` 定义已删除
- [ ] `grep -rn "containsString" internal/` 无残留
- [ ] 相关包（cloudx/aliyun, aws, huawei, tencent, volcano）`go test` 通过
- [ ] 全仓 `go build ./...` 通过

## Hard Rules
仅修改以下 5 个文件：`internal/shared/cloudx/aliyun/nas.go`、`cloudx/aws/efs.go`、`cloudx/huawei/sfs.go`、`cloudx/tencent/cfs.go`、`cloudx/volcano/tos.go`。

## Implementation Notes
- 每个文件 2 个调用点，均为 `containsString(slice, x)` 形式，等价于 `slices.Contains(slice, x)`。
- 需补 `"slices"` 导入（stdlib，Go 1.21+，项目 Go 1.25）。
- 纯行为等价替换，不改变判断语义。
