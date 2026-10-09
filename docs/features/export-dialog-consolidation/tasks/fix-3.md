---
id: "fix-3"
title: "fix lint: just lint failure in quality gate"
priority: "P0"
estimated_time: "15min"
dependencies: []
status: pending
breaking: false
type: "coding.cleanup"
---

# fix lint: just lint failure in quality gate

## Root Cause

Quality gate step `just lint` failed during quality-gate hook.

Error output saved to: `tests/results/unit-raw-output.txt`

Concise error:
```
cd e-cam-service && go vet ./internal/...
# github.com/Havens-blog/e-cam-service/internal/shared/cloudx/aliyun_test
# [github.com/Havens-blog/e-cam-service/internal/shared/cloudx/aliyun_test]
vet.exe: internal\shared\cloudx\aliyun\oss_probe_manual_test.go:155:24: cannot use fmt.Sprintf("%d", page) (value of type string) as requests.Integer value in assignment
error: recipe `lint` failed on line 13 with exit code 1

```

## Reference Files

- Source: oss_probe_manual_test.go
- Tool output: tests/results/unit-raw-output.txt

## Surface Inference

This cleanup-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `oss_probe_manual_test.go` to extract the first file path (comma-separated).
2. Run `forge surfaces --json <file-path>` to resolve surface-key/type.
3. Use the resolved surface-type to load the appropriate `rules/surfaces/<type>.md` for test orchestration guidance.

If `forge surfaces --json` fails (no surfaces configured, command not found), proceed without surface information — this does not block the cleanup.

## Cleanup Guidelines

Fix only the reported style/lint issues. Do not refactor adjacent code.

1. Read the tool output and identify each violation
2. Fix each violation with minimal changes
3. Re-run the failing tool to confirm the fix

## Verification

After fixing, verify the cleanup works:
1. Run targeted tests on changed packages: `go test -race ./affected/package/...`
2. Replace the path with the actual packages you modified

> **Note:** Full project-wide tests run at CLI submit (`forge task submit`) — agent runs targeted tests only.

E2e regression is verified by the dispatcher, not by this cleanup task.

When this task is recorded as completed via `task record`, the source task N/A (project-wide gate) is automatically restored to pending if all its dependencies are completed.
