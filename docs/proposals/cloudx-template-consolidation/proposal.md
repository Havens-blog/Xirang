---
created: "2026-09-16"
author: "Haven"
status: Approved
intent: "refactor"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: cloudx 五厂商模板收敛（RateLimiter / containsString / strVal）

## Problem

`internal/shared/cloudx` 下同一实现被逐字节复制到五个云厂商子包，且已开始漂移（复制处注释出现乱码 `实例�?`），新增厂商时需再抄第五份、最容易漏。

### Evidence

- **RateLimiter**：`common/{aliyun,aws,huawei,tencent,volcano}/rate_limiter.go`（aliyun/aws 为 `ratelimit.go`）五份各 ~25 行，`golang.org/x/time/rate` 包装，除 package 名与注释外逐字节相同（diff 为空）。引用面 ~17 文件（aliyun 2 / aws 7 / huawei 3 / tencent 4 / volcano 1）。
- **containsString**：`cloudx/{aliyun/nas.go:279, aws/efs.go:280, huawei/sfs.go:469, tencent/cfs.go:308, volcano/tos.go:256}` 五份相同签名 `func containsString(slice []string, s string) bool`（线性扫描），与 stdlib `slices.Contains` 语义一致（cloudx/logquery 已在用 stdlib）。
- **strVal**：`billing/{huawei:306, tencent:205, volcano:273}/adapter.go` 三份 `func strVal(s *string) string` 逐字节相同。

### Urgency

- 无现行功能 bug，属预防性收敛 + 维护负担消除。
- 不做的代价：复制漂移持续累积（已见乱码注释）；新增厂商（如 azure）需再抄三份模板。
- 用户已选择本候选（/quick 范围确认）。

## Proposed Solution

三处真·逐字节复制分别收敛为单一事实源，行为逐字节不变（内部实现重组，无外部可观察行为变化）：

1. **RateLimiter** → 新包 `cloudx/common/ratelimit`（`RateLimiter`/`NewRateLimiter(qps)`/`Wait(ctx)`/`Allow()`，保持原 API），删除五厂商文件，更新 ~17 个引用文件改导入。
2. **containsString** → 直接替换调用点为 stdlib `slices.Contains`，删除五个私有函数。
3. **strVal** → 共享 `billing.DerefStr(s *string) string`，三处调用点改用它。

### Innovation Highlights

直白的标准做法（stdlib + 父包共享），无创新点；价值在"同一知识一份拷贝"的长期可维护性。

## Requirements Analysis

### Key Scenarios

- 现有所有使用 `common/<provider>.NewRateLimiter(qps)` / `*.RateLimiter` 的适配器（iam/billing/asset 等）编译与行为不变。
- 五个资产文件（nas/efs/sfs/cfs/tos）的 containsString 调用（各 2 处）返回相同结果。
- 三处 billing strVal 解引用语义不变（nil → ""）。

### Non-Functional Requirements

- **行为等价**：所有改动后同一输入的输出与错误返回逐字节一致（本次为 refactor，零行为漂移是本特征的首要 NFR）。

### Constraints & Dependencies

- Go 1.21+（`slices` stdlib 可用；项目 Go 1.25）。
- `golang.org/x/time/rate` 已是既有依赖，共享包继续用它，不引入新依赖。
- 不动 `client.go`/`error.go`（各厂商 SDK 不同，非真复制）；不动 billing `asAuthError/isRetryable`（错误类型/码厂商相关，控制流统一风险高收益低）；不动 `aliyun/client.go` 手工令牌桶 `RateLimiter`（docs-only 引用，独立死代码候选，另行处理）。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 共享包 + 改导入（RateLimiter） | 本提案 | 真正的单一事实源，无残留 | 改 ~17 个导入 | Adopt（用户已确认） |
| 类型别名 `type RateLimiter = ratelimit.RateLimiter` | 备选 | 引用方零改动 | 保留 5 个残留 stub，收敛不彻底 | Rejected：用户已选共享包 |
| Do nothing | — | 零改动 | 复制漂移持续，新增厂商需抄第五份 | Rejected：用户已确认收敛 |
| containsString → slices.Contains | 本提案 | stdlib，cloudx 已在用 | 5 个调用点改动 | Adopt |
| strVal → billing.DerefStr | 本提案 | 一处实现 | 3 处改动 | Adopt |

## Scope

### In Scope

- `cloudx/common/ratelimit` 新包 + 删除 `common/{aliyun,aws,huawei,tencent,volcano}` 五份 rate_limiter/ratelimit 文件 + 更新全部引用文件导入。
- 五个资产文件 containsString 调用改 `slices.Contains` + 删除五个私有函数。
- billing 三处 strVal 改 `billing.DerefStr` + 新增 helper。
- 单测：`ratelimit` 包基础行为（NewRateLimiter/Wait/Allow 冒烟）+ `DerefStr`（nil/空/非空）。

### Out of Scope

- billing `asAuthError/isRetryable` 控制流统一。
- `common/<provider>/client.go`、`error.go`。
- `aliyun/client.go` 手工令牌桶 `RateLimiter` 的删除（docs-only 引用，需另行确认）。
- 其它同类复制（如 iam 五厂商 group.go，各厂商 API 不同）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 改导入漏改某引用文件导致编译失败 | Low（机械替换可 grep 全量） | `go build ./...` 全仓验证；改前 `grep -rl "common/<p>.*RateLimiter"` 穷举引用面 |
| 行为漂移（共享包与旧副本语义不一致） | Low（逐字节相同才收敛） | 共享包复制原实现；ratelimit 冒烟单测 |
| 与用户并行开发冲突 | Medium（用户在 CDN 成本/指标上活跃） | 只动 common/ 与 billing strVal 三处；改前 git status 确认工作树无冲突文件 |

## Success Criteria

- [ ] `cloudx/common/ratelimit` 存在且五厂商 rate_limiter/ratelimit 文件全部删除（无 alias 残留）。
- [ ] `grep -rn "containsString" internal/` 无残留；五个资产文件改用 `slices.Contains`。
- [ ] `grep -rn "func strVal" internal/` 无残留；billing 三处用 `billing.DerefStr`。
- [ ] 全仓 `go build ./...` 通过；`go test ./internal/shared/cloudx/...` 通过。
- [ ] ratelimit 与 DerefStr 单测存在且通过。
- [ ] 行为等价抽查：抽样一个适配器编译产物行为与收敛前一致（人工核对 diff 仅导入变化）。
