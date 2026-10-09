---
id: "1"
title: "新建 cloudx/common/ratelimit 共享包并删除五厂商 RateLimiter 副本"
priority: "P0"
estimated_time: "2h"
complexity: "medium"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: 新建 cloudx/common/ratelimit 共享包并删除五厂商 RateLimiter 副本

## Description
`internal/shared/cloudx/common/{aliyun,aws,huawei,tencent,volcano}` 五份逐字节相同的
RateLimiter（golang.org/x/time/rate 包装，除 package 名与注释外 diff 为空）收敛为单一
共享包 `common/ratelimit`，API 与旧实现一致（`RateLimiter`/`NewRateLimiter(qps)`/`Wait(ctx)`/`Allow()`）。
删除五厂商 rate_limiter/ratelimit 文件，并将首批 5 个引用文件改为导入共享包。

## Reference Files
- `docs/proposals/cloudx-template-consolidation/proposal.md` — Proposed Solution, Scope, Success Criteria
- `internal/shared/cloudx/common/aliyun/ratelimit.go` — 被收敛的原始实现（基准）
- `internal/shared/cloudx/common/huawei/rate_limiter.go` — 同构副本
- `internal/shared/cloudx/aliyun/cert.go` — 引用文件（改导入）
- `internal/shared/cloudx/iam/aliyun/adapter.go` — 引用文件（改导入，Task 2 覆盖 iam 其余）

## Acceptance Criteria
- [ ] `internal/shared/cloudx/common/ratelimit/ratelimit.go` 存在，导出 `RateLimiter`、`NewRateLimiter(qps int)`、`Wait(ctx)`、`Allow()`，实现与旧五份逐字节一致（复用 x/time/rate，burst=qps）
- [ ] `common/{aliyun/ratelimit.go, aws/ratelimit.go, huawei/rate_limiter.go, tencent/rate_limiter.go, volcano/rate_limiter.go}` 五个文件已删除，无 type alias 残留
- [ ] 首批 5 个引用文件（`aliyun/cert.go`, `aliyun/iam.go`, `aws/cert_discovery.go`, `huawei/cert_discovery.go`, `tencent/cert.go`）的 RateLimiter/NewRateLimiter 引用改导入 `common/ratelimit`；这些文件若还用到 `common/<p>` 其它符号（client/error），保留原导入
- [ ] `common/ratelimit` 包新增冒烟单测（NewRateLimiter 构造、Wait 无阻塞返回、Allow 限流语义）
- [ ] `go build ./internal/shared/cloudx/...` 通过

## Hard Rules
仅修改以下文件（Task 1 范围）：新增 `internal/shared/cloudx/common/ratelimit/ratelimit.go` + `ratelimit_test.go`；删除五个 common 厂商 rate_limiter 文件；修改 `aliyun/cert.go`、`aliyun/iam.go`、`aws/cert_discovery.go`、`huawei/cert_discovery.go`、`tencent/cert.go` 的导入。

## Implementation Notes
- 共享包路径 `internal/shared/cloudx/common/ratelimit`，package `ratelimit`。
- 不改变任何公开 API 签名；引用文件仅改导入路径与类型前缀（`aliyuncommon.RateLimiter` → `ratelimit.RateLimiter`）。
- iam 五个 adapter.go 的导入由 Task 2 处理，本任务不要触碰，避免并发冲突。
- 行为等价：共享包内容从 aliyun 版逐字复制（仅改注释去厂商名）。
