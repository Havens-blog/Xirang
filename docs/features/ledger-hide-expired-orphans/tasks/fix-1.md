---
id: "fix-1"
title: "Fix: e-cam-service just compile 失败（ledger_service.go:360 snap 未使用）"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
type: "coding.fix"
---

# Fix: e-cam-service just compile 失败（ledger_service.go:360 snap 未使用）

## Root Cause

Quality gate 分类：复杂/跨仓阻塞。forge task submit 1 的 quality gate 在 just compile 步骤失败：e-cam-service internal/cert/service/ledger_service.go:360:2 declared and not used: snap，go build exit 1。该文件当前为未提交 WIP（e-cam-service 大量 M 状态文件，另一会话进行中），不属任务 1（e-cam-web StateBlock 前端组件）范围，任务 1 的前端检查全部通过（vitest 561 全绿、coverage 100%、pnpm build exit 0、typecheck 相对基线零新增）。修复须在 e-cam-service 仓库内解决 compile 错误后，任务 1 才能通过 gate 完成提交。

## Reference Files

- Source: e-cam-service/internal/cert/service/ledger_service.go
- Test script: just compile (workspace root D:/Haven)
- Test results: internal/cert/service/ledger_service.go:360:2: declared and not used: snap; error: recipe compile failed on line 9 with exit code 1

## Surface Inference

This fix-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `e-cam-service/internal/cert/service/ledger_service.go` to extract the first file path (comma-separated).
2. Run `forge surfaces --json <file-path>` to resolve surface-key/type.
3. Use the resolved surface-type to load the appropriate `rules/surfaces/<type>.md` for test orchestration guidance.

If `forge surfaces --json` fails (no surfaces configured, command not found), proceed without surface information — this does not block the fix.

## Fix Boundaries

When fixing test failures, observe these boundaries:

**Forbidden:**
- Starting dev server (`npx expo start`, `npm run dev`, etc.)
- Running `npm install` more than 3 times — mark task as blocked if dependency installation fails 3 times
- Running full test suite — regression is verified by the dispatcher after fix completes
- Manually opening browser to verify rendering

**Correct workflow:**
1. Read failing test + corresponding component source
2. Compare test's expected testID/selectors vs actual DOM structure
3. Modify component (add testID) or test (adjust selectors/assertions)
4. Run targeted tests on affected packages — unit tests must pass
5. Record completion

## Verification

After fixing, verify the fix works:
1. Run targeted tests on changed packages: `go test -race ./affected/package/...`
2. Replace the path with the actual packages you modified

> **Note:** Full project-wide tests run at CLI submit (`forge task submit`) — agent runs targeted tests only.

Full regression is verified by the dispatcher, not by this fix task.

When this task is recorded as completed via `task record`, the source task 1 is automatically restored to pending if all its dependencies are completed.
