---
id: "2"
title: "IAM 五个 adapter.go 改导入 cloudx/common/ratelimit"
priority: "P0"
estimated_time: "1h"
complexity: "low"
dependencies: [1]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 2: IAM 五个 adapter.go 改导入 cloudx/common/ratelimit

## Description
Task 1 完成后，剩余 5 个 RateLimiter 引用文件（`internal/shared/cloudx/iam/{aliyun,aws,huawei,tencent,volcano}/adapter.go`）
的 `common/<p>.RateLimiter`/`NewRateLimiter` 引用改为导入 `common/ratelimit`。

## Reference Files
- `docs/proposals/cloudx-template-consolidation/proposal.md` — Proposed Solution, In Scope
- `internal/shared/cloudx/iam/aliyun/adapter.go` — 引用文件（rateLimiter 字段 + NewRateLimiter 调用）
- `internal/shared/cloudx/iam/huawei/adapter.go` — 引用文件

## Acceptance Criteria
- [ ] 五个 iam adapter.go 的 RateLimiter/NewRateLimiter 引用全部改导入 `common/ratelimit`；若仍用 `common/<p>` 其它符号则保留原导入
- [ ] 结构体字段类型同步更新（如 `rateLimiter *huaweicommon.RateLimiter` → `*ratelimit.RateLimiter`），构造调用 `huaweicommon.NewRateLimiter(15)` → `ratelimit.NewRateLimiter(15)`
- [ ] `grep -rn "common/(aliyun|aws|huawei|tencent|volcano).*RateLimiter\|common/(aliyun|aws|huawei|tencent|volcano).*NewRateLimiter" internal/` 无残留
- [ ] 全仓 `go build ./...` 通过
- [ ] 行为等价抽查：任一 iam 适配器 diff 中除导入/类型前缀外无其它改动

## Hard Rules
仅修改以下 5 个文件：`internal/shared/cloudx/iam/aliyun/adapter.go`、`iam/aws/adapter.go`、`iam/huawei/adapter.go`、`iam/tencent/adapter.go`、`iam/volcano/adapter.go`。

## Implementation Notes
- 依赖 Task 1（共享包须先存在）。
- 机械替换，不触碰任何方法体逻辑。
- 五个文件均同时 import `cloudxiam` 与 `domain`，注意新导入分组与 gofmt。
