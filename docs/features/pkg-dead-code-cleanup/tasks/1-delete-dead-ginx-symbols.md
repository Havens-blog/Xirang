---
id: "1"
title: "删除 pkg/ginx 四个确认死的导出符号"
priority: "P0"
estimated_time: "30m"
complexity: "low"
dependencies: []
surface-key: ""
surface-type: ""
breaking: true
type: "coding.cleanup"
mainSession: false
---

# 1: 删除 pkg/ginx 四个确认死的导出符号

## Description
删除 `pkg/ginx` 的 4 个确认死导出符号：`wrap.go` 的 `WrapFinder`/`WrapFinderBody`/`Ws`（保留在用 `Wrap`/`WrapBody`）+ `test/recorder.go` 的 `MustScan`（保留 `JSONResponseRecorder` 其它方法）。核验已完成：本仓（含测试/ioc/cmd）零引用、仓外（ecmdb/eiam）零 import、私有仓。

## Reference Files
- `docs/proposals/pkg-dead-code-cleanup/proposal.md` — Proposed Solution, In Scope, Success Criteria
- `internal/../pkg/ginx/wrap.go` → `e-cam-service/pkg/ginx/wrap.go` — WrapFinder(41)/WrapFinderBody(57)/Ws(71) 定义
- `e-cam-service/pkg/ginx/test/recorder.go` — MustScan(24) 定义

## Acceptance Criteria
- [ ] `pkg/ginx/wrap.go` 删除 `WrapFinder`/`WrapFinderBody`/`Ws` 三个函数，`Wrap`/`WrapBody` 保留
- [ ] `pkg/ginx/test/recorder.go` 删除 `MustScan` 方法，`JSONResponseRecorder` 其它方法保留
- [ ] `grep -rn "WrapFinder\|WrapFinderBody\|\.Ws(\|MustScan" pkg/ internal/ ioc/ cmd/ --include="*.go"` 无残留（含定义处）
- [ ] `go test ./pkg/...` 全绿；`go build ./...` 通过

## Hard Rules
仅修改 `e-cam-service/pkg/ginx/wrap.go` 与 `e-cam-service/pkg/ginx/test/recorder.go`。不触碰 mongox、ecmdb fork、internal 其它死符号。任务文件路径（若含 pkg/ 前缀）以 `e-cam-service` 仓库根为基准。

## Implementation Notes
- 删除后检查 wrap.go / recorder.go 是否有随之未用的 import，清理之。
- 提交信息注明 API 移除（breaking，但零消费者已核验）。
- 工作树并发会话（cert/logquery）脏文件与本作用域无关，仅显式提交两个目标文件。
- 调用点核验已由 proposal 完成，本任务无需重复全仓 grep（但执行前后各跑一次确认）。