---
id: "fix-1"
title: "Fix: e-cam-web CacheAnalyzeCard.test.ts 缺 byte_hit_rate 字段阻塞 submit 门禁"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
type: "coding.fix"
---

# Fix: e-cam-web CacheAnalyzeCard.test.ts 缺 byte_hit_rate 字段阻塞 submit 门禁

## Root Cause

环境类阻塞（非本任务代码缺陷）：任务 1 已实现并通过本地质量验证（go build/vet/gofmt clean、internal/cert 全部包 go test 全绿、dashboard_service.go 覆盖率 91.6%），但 forge task submit 的共享质量门禁 just compile 在 e-cam-web vue-tsc 步骤失败——cd n-cache-analysis feature 提交 d4ca4e5（2026-09-22 12:07）给 CacheDomainHitStat 类型新增必填字段 byte_hit_rate/byte_hit_available 后，CacheAnalyzeCard.test.ts:56-57 的 fixture 未同步更新。该阻塞为已提交的确定性断点（30+ 分钟未自愈），与本任务（仅改 e-cam-service internal/cert dashboard 路径）无关，且任务 Hard Rule 限定修改面禁止本任务越界修复。修复方式：按 d4ca4e5 的新契约补齐 CacheAnalyzeCard.test.ts 两处 fixture 的 byte_hit_rate/byte_hit_available 字段（语义应与该 feature 域名排行字节命中率列一致），门禁通过后 forge task submit 1 --data docs/features/dashboard-cert-granularity/tasks/process/record.json 即可封档（record.json 已就绪，testsPassed=25/testsFailed=0/coverage=91.6）。

## Reference Files

- Source: e-cam-web/src/views/logs/components/CacheAnalyzeCard.test.ts
- Test script: cd /d/Haven/e-cam-web && npx vue-tsc -b
- Test results: src/views/logs/components/CacheAnalyzeCard.test.ts(56,17): error TS2739: Type '{ host: string; requests: number; hit_rate: number; cacheable_hit_rate: number; miss_traffic_ratio: number; grade: string; }' is missing the following properties from type 'CacheDomainHitStat': byte_hit_rate, byte_hit_available (lines 56,57)

## Surface Inference

This fix-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `e-cam-web/src/views/logs/components/CacheAnalyzeCard.test.ts` to extract the first file path (comma-separated).
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
