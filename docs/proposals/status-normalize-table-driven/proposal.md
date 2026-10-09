---
created: "2026-09-17"
author: "Haven"
status: Approved
intent: "refactor"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: 状态归一化 switch 表驱动化（types/status.go + eni.go）

## Problem

`internal/shared/cloudx/types` 的状态归一化用逐字 switch 硬编码同构的「字符串 → 枚举」映射，26+ case 全部是 `case "foo": return X` 单行体，新增状态/厂商需逐一追加 case，且各 normalizer 间口径靠肉眼保持一致。

### Evidence

- `types/status.go`：8 个 `NormalizeXStatus`（RDS/Redis/MongoDB/SecurityGroup/Image/Disk/Snapshot + 通用），合计 26+ case，每 case 体均为单行 return。
- `types/eni.go`：`NormalizeENIStatus(provider, status)` 用 `switch provider` 派发到 5 个厂商函数，每个函数又是纯 `case → 枚举` switch（aliyun/aws/huawei/tencent/volcano）。
- 既有测试锁定行为：`status_test.go` + `eni_test.go` 存在，是安全的改造护栏。

### Urgency

- 无现行 bug，属可维护性收敛。不做的代价：case 列表继续增长、厂商新增时需再抄一份 ENI switch。
- 用户已在 /quick 范围确认中选择本候选。

## Proposed Solution

两个文件的状态归一化 switch 收敛为数据表（map），行为逐字节不变（零漂移）：

1. **status.go**：每个 `NormalizeXStatus` 的 `strings.ToLower` + switch → 一张 `map[string]string`（小写输入 → 规范状态）+ 保留 `default: return lower` 备选。
2. **eni.go**：`NormalizeENIStatus` 的 `switch provider` 派发 + 五个厂商函数 → `map[string]map[string]string`（provider → 原始状态串 → 规范状态），volcano/volcengine 别名保留；`default: return status` 保留。

### Innovation Highlights

直白的表驱动重构，无创新点；价值在消除逐字复制与口径漂移。

## Requirements Analysis

### Key Scenarios

- 8 个 status normalizer 对既有输入输出完全不变（含 `""`→unknown、未匹配→lower）。
- ENI 五厂商大小写敏感映射不变（Available/ACTIVE/in-use 等原样匹配），未匹配→原值。
- volcano 与 volcengine 两个 provider 名路由到同一映射。

### Non-Functional Requirements

- **行为等价**：零漂移是本特征的硬约束（用户已确认）；既有 `status_test.go`/`eni_test.go` 零改动作为锁定。

### Constraints & Dependencies

- 仅改 `internal/shared/cloudx/types/status.go` 与 `eni.go`。
- 谓词函数（IsDiskAvailable/IsRunningStatus 等）是 normalizer 薄封装，不动。
- 不做大小写归一（用户已否决——ENI 大小写敏感是有意行为）。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 表驱动 map + 原 fallback | 本提案 | 数据即代码，易扩展 | 无 | Adopt（用户已确认） |
| 顺带大小写归一（eni 也 ToLower） | 备选 | 统一口径 | 改变 ENI 行为，需改测试与可能影响调用方 | Rejected：用户已选零漂移 |
| Do nothing | — | 零改动 | case 列表持续膨胀 | Rejected：用户已确认收敛 |

## Scope

### In Scope

- `status.go` 8 个 normalizer 表驱动化（`map[string]string` + lower fallback，26+ case → 数据表）。
- `eni.go` `NormalizeENIStatus` 表驱动化（`map[provider]map[string]string` + 原值 fallback，4 switch + 5 厂商函数 → 数据表，volcano/volcengine 别名保留）。
- 既有测试断言不改；新增表结构一致性测试（可选）。

### Out of Scope

- 谓词函数（Is*）、其它 types 状态族、调用方迁移、大小写归一统一。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 表驱动改写漏掉某 case（映射残缺）→ 行为漂移 | Low | 逐 normalizer 对照原 switch 抄表；既有 26+ 断言测试回归 |
| 腾讯 ENI 怪例（`"BINDbindingd"` 带尾空格等）抄错 | Low | 逐字符对照，测试锁定 |
| fallback 语义丢失（`return lower` vs `return status` 混淆） | Medium | 表驱动包装保留原 fallback 行，逐函数对照 |

## Success Criteria

- [ ] `status.go` 无 `case "..."` 残留（8 个 normalizer 全部表驱动，谓词函数除外）。
- [ ] `eni.go` 无 `switch provider` 派发（map 查找替代），五厂商函数 case 清零。
- [ ] `go test ./internal/shared/cloudx/types/...` 全绿（既有 status_test + eni_test 零改动通过）。
- [ ] 新增结构一致性单测：每 normalizer 抽样输入输出与 map 表一致、ENI 别名 volcano/volcengine 路由一致。
- [ ] 全仓 `go build ./...` 通过；行为等价抽查 diff 仅 switch→map 结构变化。