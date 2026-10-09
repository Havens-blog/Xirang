---
created: "2026-09-17"
author: "Haven"
status: Approved
intent: "enhancement"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: 规则引擎双实现收敛（共享 MatchValue + 大小写口径统一）

## Problem

`equals/contains/regex` 等规则操作符的求值语义在 `internal/cam/tag` 与 `internal/cam/servicetree` 各实现一份，且**大小写口径不一致**：tag 侧 contains/prefix/suffix 大小写不敏感，servicetree 侧 Contains 大小写敏感。同类操作符跨域语义漂移，维护与行为理解成本高。

### Evidence

- `internal/cam/tag/tag_rules.go:358` `matchOperator`：`contains`/`prefix`/`suffix` 用 `strings.ToLower` 双侧后比较（不敏感）；`equals`/`regex` 敏感。操作符：equals/contains/prefix/suffix/regex。（注：重构后该函数更名为 `matchCondition`（tag_rules.go:322），行号/函数名已漂移，详见文末 Drift Verification。）
- `internal/cam/servicetree/service/rule_engine.go:268` `compareValue`：`Eq/Ne/Contains/Regex/In/NotIn/Exists`，`Contains` 用 `strings.Contains`（**大小写敏感**），其余精确/正则。操作符集与 tag 部分重叠（Eq/Contains/Regex）。
- 同一语义（contains 子串匹配）两处口径不同，任一修改需两侧同步，易漂移。

### Urgency

- 无现行 bug 报告，属语义归一 + 收敛。不做的代价：双实现持续漂移。
- 用户已确认口径（统一不敏感）与形态（抽共享 MatchValue）。

## Proposed Solution

- 新增共享求值函数 `MatchValue(operator, actual, expected string) bool`，覆盖 **equals/contains/prefix/suffix/regex**；其中 contains/prefix/suffix **大小写不敏感**（对齐 tag 侧现口径，用 ToLower 双侧），equals/regex 保持敏感。
- `tag_rules.matchOperator` 改为委托 `MatchValue`（现口径已一致，纯重构）。
- `rule_engine.compareValue` 的 Eq/Contains/Regex 改委托 `MatchValue`（**Contains 语义由敏感放宽为不敏感**——唯一行为变化）；Ne/In/NotIn/Exists 留在原处（tag 侧无对应，不收敛）。
- 共享函数放置：`internal/cam/domain`（tag 与 servicetree 均能导入且语义属领域层）。

### Innovation Highlights

直白的语义归一，无创新点；价值在消除双实现漂移 + 统一大小写体验（规则配置宽容匹配）。

## Requirements Analysis

### Key Scenarios

- tag 规则 contains 匹配：大小写不敏感（现状不变，回归确认）。
- servicetree 分类规则 Contains：**从敏感变不敏感**——之前 "Prod" 模式不匹配 "prod-01" 的现在匹配（放宽，需测试确认分类结果无意外）。
- regex 匹配：行为不变（用户用模式控制大小写）。
- equals/Ne/In/NotIn：精确匹配不变。

### Non-Functional Requirements

- 共享 `MatchValue` 覆盖 5 个重叠操作符，语义单点定义。
- 行为变化仅限 servicetree Contains 放宽（用户已确认）。

### Constraints & Dependencies

- 操作符命名：`MatchValue` 用 tag 侧字符串名（equals/contains/prefix/suffix/regex）作为规范；servicetree 常量映射到这些名。
- 仅改 `internal/cam/tag/tag_rules.go`、`internal/cam/servicetree/service/rule_engine.go` + 新增 domain helper。
- field mapping（tag_rules.go:259 / rule_engine.go:236）不在范围（字段语义两侧差异大）。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 共享 MatchValue + 不敏感统一 | 本提案 | 单点语义、宽容匹配 | servicetree Contains 放宽（已确认） | Adopt（用户已确认） |
| 仅对齐大小写（不抽 helper） | 备选 | 改动最小 | 双实现仍存，漂移风险依旧 | Rejected：用户已选共享 |
| 含 field mapping 收敛 | 备选 | 收敛更彻底 | 字段语义差异大、风险高 | Rejected：用户已排除 |
| Do nothing | — | 零改动 | 双实现漂移持续 | Rejected：用户已确认 |

## Scope

### In Scope

- `internal/cam/domain` 新增 `MatchValue(operator, actual, expected string) bool`（equals/contains/prefix/suffix/regex；后三者 ToLower 双侧不敏感）+ 单测。
- `tag_rules.matchOperator` 委托 `MatchValue`（现口径一致，纯重构）。
- `rule_engine.compareValue` 的 Eq/Contains/Regex 委托 `MatchValue`（Contains 放宽不敏感）；Ne/In/NotIn/Exists 保留。
- 单测：`MatchValue`（各 op 大小写用例、regex 转义）+ tag/servicetree 既有测试回归（servicetree Contains 用例补不敏感断言）。

### Out of Scope

- field mapping（attribute→字段名）收敛。
- servicetree Ne/In/NotIn/Exists 迁移。
- 规则引擎的 Mongo filter 构建（conditionToFilter 等）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| servicetree Contains 放宽导致拓扑分类结果意外变化 | Medium | 既有 rule_engine 测试回归 + 新增不敏感断言；放宽只增匹配不漏 |
| 操作符名映射错位（tag 名 vs servicetree 常量值） | Low | 单测锁定各 op 输入输出；servicetree 常量→字符串映射显式 |
| ToLower 的 Unicode 语义（非 ASCII 大小写） | Low | 沿用 tag 侧现有 ToLower 实现（行为一致），不引入 EqualFold |

## Success Criteria

- [ ] `internal/cam/domain` 存在 `MatchValue`，覆盖 5 个操作符，contains/prefix/suffix 大小写不敏感（单测断言 "Prod" 模式匹配 "prod-01"）。
- [ ] `tag_rules.matchOperator` 委托 `MatchValue`（tag 行为零变化，既有 tag 测试零改动通过）。
- [ ] `rule_engine.compareValue` Eq/Contains/Regex 委托 `MatchValue`，Contains 不敏感（新增断言）；Ne/In/NotIn/Exists 原样保留。
- [ ] `go test ./internal/cam/tag/... ./internal/cam/servicetree/...` 全绿。
- [ ] 全仓 `go build ./...` 通过；grep 无残留的重复 contains 求值实现（`matchOperator`/`compareValue` 内不再内联 contains/regex 逻辑）。

## Drift Verification (2026-09-17)

任务 T-quick-doc-drift 对本提案逐条 Success Criteria 与实际代码核查，结论：**5/5 全部达成，无行为级漂移**；仅发现 2 处文本级漂移并已修正。

| SC | 结论 | 证据 |
|----|------|------|
| MatchValue 覆盖 5 操作符、contains/prefix/suffix 不敏感 | 达成 | `internal/cam/domain/rule_match.go`（ToLower 双侧）；单测 `rule_match_test.go` 含 "Prod"~"prod-01" 双向不敏感断言 |
| tag 侧委托、行为零变化 | 达成 | `tag_rules.go:324` `matchCondition` 委托 `camdomain.MatchValue` |
| servicetree Eq/Contains/Regex 委托；Ne/In/NotIn/Exists 保留；Contains 放宽有断言 | 达成 | `rule_engine.go:269` `compareValue` 三分支委托 + 其余原样；`rule_engine_test.go:24-25` 显式不敏感断言 |
| `go test ./internal/cam/tag/... ./internal/cam/servicetree/...` 全绿 | 达成 | tag/service/domain 三包 ok |
| `go build ./...` 通过；无残留内联 contains/regex 求值 | 达成 | build OK；tag_rules.go 残留 regexp 仅为 Mongo filter 构建（提案明确 Out of Scope） |

**文本级漂移（已修正）**：

1. Evidence 中 `tag_rules.go:358` `matchOperator` 行号/函数名漂移——重构后函数更名为 `matchCondition`（:322），已在 Evidence 行内加注。
2. 本文档 Evidence 第二条 `rule_engine.go:268` 行号漂移——`compareValue` 现位于 :269（委托实现，注释标明放宽语义），属重构后自然行号偏移，语义无变化。
3. feature manifest.md 任务表状态漂移（3 任务显示 pending，实际 index.json 均 completed），已同步为 completed。

**项目级 spec 说明**：`docs/business-rules/`、`docs/conventions/` 项目级 spec 目录在本 quick-mode 工作区不存在，无项目级 spec 可漂移核查（与 log-query-optimization / cert-multicloud-deployers 漂移核查结论一致）。