---
id: "fix-1"
title: "Fix: e-cam-web assets 测试文件 vue-tsc 类型错误阻塞 forge quality gate"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
type: "coding.fix"
---

# Fix: e-cam-web assets 测试文件 vue-tsc 类型错误阻塞 forge quality gate

## Root Cause

复杂/在途错误：forge task submit 2 的 quality gate 执行 just compile → e-cam-web-compile (vue-tsc -b) 失败。错误位于并发会话在途新增（未跟踪）的 src/views/assets/__tests__/index.test.ts，属前端文件，任务 2 Hard Rules 明确禁止触碰（仅限 internal/cam/servicetree/）。失败非本任务代码引入，重试无意义；待前端会话完成或由后续修复任务修正类型标注与 unused imports 后，forge task submit 2 --data docs/features/servicetree-overhaul-p1p2/records/2.record.json 即可完成落档。分类：跨包 gate 失败 + 并发在途文件冲突。

## Reference Files

- Source: D:/Haven/e-cam-web/src/views/assets/__tests__/index.test.ts
- Test script: cd e-cam-web && npx vue-tsc -b
- Test results: TS6133 x2 (ElInput/ElSelect unused imports); TS2322 x4 ("cloud_vm"/"cloud_rds"/"cloud_redis" not assignable to AssetType; "alicloud" not assignable to CloudProvider)

## Surface Inference

This fix-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `D:/Haven/e-cam-web/src/views/assets/__tests__/index.test.ts` to extract the first file path (comma-separated).
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

When this task is recorded as completed via `task record`, the source task 2 is automatically restored to pending if all its dependencies are completed.
