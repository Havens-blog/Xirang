---
id: "fix-2"
title: "fix unit-test: just unit-test failure in quality gate"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
type: "coding.fix"
---

# fix unit-test: just unit-test failure in quality gate

## Root Cause

Quality gate step `just unit-test` failed during quality-gate hook.

Error output saved to: `tests/results/unit-raw-output.txt`

Concise error:
```
...
355|         expect(styleBody).not.toMatch(/#[0-9a-fA-F]{3,8}\b/)
|                               ^
356|         expect(styleBody).not.toMatch(/rgba?\(/)
357|         expect(styleBody).toContain('var(--')
⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[7/7]⎯
Test Files  1 failed | 45 passed (46)
Tests  7 failed | 849 passed (856)
Start at  19:38:05
Duration  15.19s (transform 14.79s, setup 0ms, import 80.70s, tests 45.42s, environment 23.71s)
error: recipe `e-cam-web-unit-test` failed on line 34 with exit code 1
```

## Reference Files

- Source: runtime-core.cjs.js, reactivity.cjs.js
- Test script: just unit-test
- Test results: tests/results/unit-raw-output.txt

## Surface Inference

This fix-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `runtime-core.cjs.js, reactivity.cjs.js` to extract the first file path (comma-separated).
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

When this task is recorded as completed via `task record`, the source task N/A (project-wide gate) is automatically restored to pending if all its dependencies are completed.
