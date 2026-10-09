---
created: "2026-09-22"
author: "Haven"
status: Approved
intent: "refactor"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: UI Unify 残留旧蓝清扫（视图级 scss 7 文件）

## Problem

Phase 1 令牌层把 `src/assets/styles/` 与 Element Plus 桥接全部切到 Linear 靛紫后，视图级 scss 里还残留着旧蓝字面量（`#3b82f6` / `#2563eb` / `rgba(59,130,246,X)`），导致站内「蓝 / 靛紫」两套系统并存——同样的 `.blue` 徽章文字已用 `var(--accent-blue)` 但背景仍是硬编码旧蓝。

### Evidence

- 精确扫描（排除测试断言串）：7 个视图级 scss、共 25 行残留旧蓝：
  - `views/cmdb/models/styles/models.scss`（9）、`cmdb/instances/styles/instances.scss`（4）、`cmdb/relations/styles/relations.scss`（4）、`cmdb/topology/styles/topology.scss`（3）、`compute/ecs/styles/instances.scss`（4）、`storage/styles/export-dialog.scss`（1）、`storage/styles/column-settings.scss`（1）
- 浏览器实测 `cmdb/models` 页在深色主题下 **35 处计算样式命中**旧蓝（`group-item.active`、`group-count`、`.model-icon.blue`、`.badge.blue`、`.el-button--primary` 的 glow `rgba(59,130,246,.25)` 等）——即真实可见的双蓝并存
- 直接先例 `fc0b2e4`（"JS/TS 数据层旧蓝色字面量迁移至靛紫系"）已用同一约定清掉数据层旧蓝，本提案是其视图级 scss 的对应补完

### Urgency

这是 Phase 1 收尾的最后一块硬编码残留：不清理，Phase 3/4 之后 220+ 页迁移样板仍带双蓝反例，且每次页面改动都在延续错误色值。成本低（25 行机械迁移）、收益确定（全站单一靛紫体系），宜尽快闭环。

## Proposed Solution

按既定约定，把 7 个视图级 scss 的 25 行旧蓝字面量机械迁移为靛紫（不改选择器结构、不改任何行为、不新增变量）：

| 旧值 | 新值 | 语义 |
|------|------|------|
| `#3b82f6` | `#7170ff` | 实心 accent-blue（与 fc0b2e4 一致） |
| `rgba(59,130,246,X)` | `rgba(113,112,255,X)` | accent-blue 微透（与 fc0b2e4 一致） |
| `#2563eb`（主按钮 hover 暗化） | `#5e6ad2` | hover 保留「变暗」语义（用户确认，见 Assumptions Challenged） |

全部沿用 `fc0b2e4` 的数据层约定，唯一例外是 `#2563eb` 由用户拍板走「暗化 → `#5e6ad2`」而非 Phase 1 计划对 EP 桥接的 `#828fff`（因为这几处基色是 `--accent-blue`，非 `--accent-primary`，且原始意图是 hover 变暗）。

### Innovation Highlights

无创新——这是对既有约定（fc0b2e4）+ 既定设计系统（Linear 规格）的**忠实补完**，刻意不引入新抽象，避免 7 个文件里再冒出一套新 token/混合色约定。诚实说明：方案的价值不在创意，而在把遗留双蓝彻底抹平。

## Requirements Analysis

### Key Scenarios

- **正常路径**：7 页在深/浅两主题下渲染，旧蓝位置全部呈现对应靛紫，无报错、无结构位移
- **边界**：`.blue` 类同时含 `background: rgba(59,130,246,X)` 与 `color: var(--accent-blue)`——背景迁到 `rgba(113,112,255,X)` 后与文字 token 同色系，视觉一致
- **异常**：任意一行改错导致 scss 编译失败 → `pnpm build` 即暴露；改错色值但能编译 → 浏览器计算样式断言兜底

### Non-Functional Requirements

- **可访问性**：hover 暗化 `#5e6ad2` 与基色 `#7170ff` 的对比度需仍可辨识（浏览器 hover 态验收覆盖）
- **一致性**：迁移后所有 `.blue` 徽章/激活态/glow 同色系
- 其余 NFR（性能/安全/兼容/可扩展）不涉及，不展开

### Constraints & Dependencies

- 依赖：Phase 1 已定义的 `--accent-blue`/`--accent-primary` token（已合并 main，08904e9）
- 约束：不动 `src/views/**` 之外的文件；不新增 scss 变量；不改选择器结构（Phase 1 的「只换值」原则）
- 前置：`fc0b2e4` 数据层迁移已落地（上诉约定以此为准）

## Alternatives & Industry Benchmarking

### Industry Solutions

设计系统落地中，旧主题色残留的清扫通常是 grep 扫描 + 逐点迁移 + 构建/视觉回归，无专门工具。本方案即此标准做法。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零成本 | 双蓝并存长期可见，样板带反例，漂移累积 | Rejected：成本低收益确定，无理由拖延 |
| 全盘用 Phase 1 字面规则（`#2563eb→#828fff`） | Phase 1 计划 | 全库一条规则 | 这几处基色是 accent-blue，`#828fff` 比基色更亮、与原始「hover 暗化」意图相反 | Rejected：语义失真 |
| **按 fc0b2e4 约定 + `#2563eb→#5e6ad2` 暗化** | fc0b2e4 沿用 | 与数据层一致、保留 hover 暗化语义、复用既有 token 值 | 需用户对 `#2563eb` 单独拍板（已拍） | **Selected：语义最忠实** |

## Feasibility Assessment

### Technical Feasibility

纯 scss 字面量替换，无任何技术障碍；scss 经 vite `additionalData` 全局注入 variables，无需额外 `@use`。

### Resource & Timeline

7 文件 / 25 行，单人一轮即可完成（迁移 + 构建 + 浏览器验收），无外部依赖。

### Dependency Readiness

上游 token（`--accent-blue`/`--accent-primary`）已就绪；`fc0b2e4` 迁移约定已存在。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| `#2563eb` 应按 Phase 1 字面规则统一映射 `#828fff` | Assumption Flip + 语义核对 | Overturned：Phase 1 的 `#2563eb→#828fff` 只针对 `--accent-primary` 基色的 EP 按钮（hover 变亮）；这 7 处基色是 `--accent-blue` 且原始意图是 hover 变暗，照搬会反向。用户选定 `#5e6ad2`（暗化）。 |
| 该清扫可与 filters bug 打包成一个 feature | Occam's Razor + XY | Refined：两项 intent 不同（refactor vs fix），用户决定拆两轮，以 /quick 规则「一次一 feature」为准。 |
| 需为徽章背景新增「accent-blue tint」token 才能一致 | Occam's Razor | Overturned：fc0b2e4 已直接用字面量 `rgba(113,112,255,X)` 而非 token，沿用即可，不引入新抽象。 |

## Scope

### In Scope

- 迁移 7 个视图级 scss 的 25 行旧蓝字面量（映射见 Proposed Solution 表）
- `pnpm build` + `pnpm typecheck` 通过
- 浏览器实测 `cmdb/models`（旧蓝最集中页）深/浅两主题旧蓝计算样式归零，并抽查其余 6 页无回归

### Out of Scope

- 测试文件里的断言串（`expect(...).not.toContain('#3b82f6')`）——本就正确，不动
- `resource/tag/index.vue` 的 `const filters` 常量 bug —— 下一轮单独走 fix-bug
- Phase 4 的 stylelint 防回潮规则
- Element Plus 桥接与令牌层（Phase 1 已完成）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| hover 暗化 `#5e6ad2` 对比度不足 | M | L | 浏览器 hover 态实测，不足则回退 `#4f5ab8`（light 主题 dark-2） |
| 某文件存在 `rgba(59,130,246)` 无空格变体被 grep 漏掉 | L | L | 迁移后用 `59,\s*,\s*130` 正则复核清零 |
| 改错色值但能编译 | L | M | 浏览器计算样式断言（旧蓝 35→0）兜底 |

## Success Criteria

- [ ] SC-1：`grep -rn "3b82f6\|2563eb\|59,\s*130,\s*246" src/views --include="*.scss"` 零命中（排除测试）
- [ ] SC-2：`pnpm build` 通过
- [ ] SC-3：`pnpm typecheck` 通过
- [ ] SC-4：浏览器 `cmdb/models` 深/浅两主题旧蓝计算样式命中 = 0（当前深色 35）
- [ ] SC-5：`pnpm lint` 相对基线零新增（基线含 3 个历史 error，允许存在但不新增）

consistency_check_result:
  status: pass
  pairs_checked: 15
  conflicts_found: 0

## Next Steps

- Proceed to `/quick-tasks` to generate tasks