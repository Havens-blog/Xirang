---
id: "fix-2"
title: "Fix: submit 质量门 just lint 失败（e-cam-service go vet NewDiscoveryImportService 参数不符 + gofmt 4 文件）"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
type: "coding.fix"
---

# Fix: submit 质量门 just lint 失败（e-cam-service go vet NewDiscoveryImportService 参数不符 + gofmt 4 文件）

## Root Cause

Complex/recurring blocker: forge task submit 3 的质量门（just lint）在 e-cam-service 失败——并发会话正在做 cert EnvelopeCrypto 重构（NewDiscoveryImportService 已加参，module.go 与两个测试调用方未更新），另有 4 个文件 gofmt 未格式化。任务 3（前端零漂移迁移）Hard Rules 禁碰后端，无法自行修复；前端侧 typecheck/vitest 512/512 已全绿，提交 bc28285 已落。等待后端重构收敛后修复 vet/gofmt，再重跑 forge task submit 3 即可（record.json 已在 docs/features/export-dialog-consolidation/tasks/process/record.json）。

## Reference Files

- Source: e-cam-service/internal/cert/module.go, e-cam-service/internal/cert/service/cert_sync_service_test.go, e-cam-service/internal/cert/web/discovery_handler_test.go, e-cam-service/internal/shared/cloudx/aliyun/cert_lb.go, e-cam-service/internal/shared/cloudx/logquery/aws/mapper.go, e-cam-service/internal/shared/cloudx/logquery/huawei/mapper.go, e-cam-service/internal/shared/cloudx/logquery/tencent/waf_test.go
- Test script: just lint (from D:/Haven)
- Test results: go vet: internal/cert/module.go:250, internal/cert/service/cert_sync_service_test.go:160, internal/cert/web/discovery_handler_test.go:310 — not enough arguments in call to service.NewDiscoveryImportService (missing *domain.EnvelopeCrypto); gofmt -l: aliyun/cert_lb.go, logquery/aws/mapper.go, logquery/huawei/mapper.go, logquery/tencent/waf_test.go

## Surface Inference

This fix-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `e-cam-service/internal/cert/module.go, e-cam-service/internal/cert/service/cert_sync_service_test.go, e-cam-service/internal/cert/web/discovery_handler_test.go, e-cam-service/internal/shared/cloudx/aliyun/cert_lb.go, e-cam-service/internal/shared/cloudx/logquery/aws/mapper.go, e-cam-service/internal/shared/cloudx/logquery/huawei/mapper.go, e-cam-service/internal/shared/cloudx/logquery/tencent/waf_test.go` to extract the first file path (comma-separated).
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

When this task is recorded as completed via `task record`, the source task 3 is automatically restored to pending if all its dependencies are completed.
