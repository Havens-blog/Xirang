---
created: "2026-09-19"
author: "Haven"
status: Approved
intent: "cleanup"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: pkg/ginx 死代码清理（WrapFinder/WrapFinderBody/Ws/MustScan）

## Problem

`pkg/ginx` 的 4 个导出符号零引用（本仓含测试、仓外 Go 仓），属确认死代码；删除导出 API 有理论外部消费者风险，需先核验再清理。

### Evidence

- `pkg/ginx/wrap.go:41` `WrapFinder`、`:57` `WrapFinderBody`、`:71` `Ws`：本仓全量 grep（含 `internal/`、`ioc/`、`cmd/`、测试）零引用。
- `pkg/ginx/test/recorder.go:24` `MustScan()`：仅定义处一处（方法接收者写法），无调用。
- 仓外核验：`D:/Haven/ecmdb`、`D:/Haven/eiam` 无 `Havens-blog/e-cam-service/pkg` 的 Go import（ecmdb 有**自己的** `pkg/ginx` fork，非本仓消费者）；私有仓（Havens-blog）。
- `pkg/mongox/mongo.go:33` `Collections` 曾疑似死代码——核验有 8 处引用，**排除**（agent 误报已甄别）。

### Urgency

- 无功能影响，纯技术债清理。不做的代价：4 个误导性导出符号留在 API 面，新代码可能误用。
- 用户已确认删除（核验零消费者）。

## Proposed Solution

删除 4 个确认死的导出符号：
- `pkg/ginx/wrap.go`：删 `WrapFinder`、`WrapFinderBody`、`Ws`（保留仍在使用的 `Wrap`/`WrapBody`）。
- `pkg/ginx/test/recorder.go`：删 `MustScan`（保留 `JSONResponseRecorder` 其它方法）。

### Innovation Highlights

直白清理，无创新点；价值在 API 面瘦身与消除误用风险。

## Requirements Analysis

### Key Scenarios

- pkg/ginx 既有消费者仅用 `Wrap`/`WrapBody`——删除后不受影响。
- 理论外部消费者：私有仓 + 仓外零 import 核验 → 风险可接受。

### Non-Functional Requirements

- 行为等价：删除后其余符号功能不变，`go build ./...` 通过。

### Constraints & Dependencies

- 仅改 `pkg/ginx/wrap.go`、`pkg/ginx/test/recorder.go`。
- 不动 `pkg/mongox`（Collections 在用）、不动 ecmdb 的 ginx fork、不动 `internal/` 的其它死符号（ProcessFailedTasks 等，另行处理）。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 删除 4 个死符号 | 本提案 | API 面瘦身、消除误用 | 导出 API 移除（零消费者已核验） | Adopt（用户已确认） |
| 保留导出符号 | 备选 | 理论外部兼容 | 死代码滞留、误用风险 | Rejected：用户已确认删除 |
| 删 MustScan 仅 | 备选 | 最低风险 | 清理不彻底 | Rejected |

## Scope

### In Scope

- `pkg/ginx/wrap.go` 删除 `WrapFinder`/`WrapFinderBody`/`Ws`。
- `pkg/ginx/test/recorder.go` 删除 `MustScan`。
- 构建与测试回归。

### Out of Scope

- `pkg/mongox` Collections（在用）。
- ecmdb/eiam 的 ginx fork。
- `internal/cam/iam/service` 的 `ProcessFailedTasks`/`ProcessPendingTasks`/`InitBuiltInTemplates`、`iam/factory.GetCachedAdapter`（internal 死符号，另行专项）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 理论外部消费者被破坏 | Low | 仓外零 import 已核验；私有仓；提交信息注明 API 移除 |
| 误删在用符号 | Low | 全仓 grep 核验（含测试、ioc、cmd）+ `go build ./...` 回归 |

## Success Criteria

- [ ] `grep -rn "WrapFinder\|WrapFinderBody\|\.Ws(\|MustScan" pkg/ internal/ ioc/ cmd/ --include="*.go"` 无残留（定义处也删净）。
- [ ] `pkg/ginx` 其余符号（Wrap/WrapBody/JSONResponseRecorder 其他方法）不受影响，`go test ./pkg/...` 全绿。
- [ ] 全仓 `go build ./...` 通过。