---
id: "1"
title: "迁移 7 个视图级 scss 的旧蓝字面量为靛紫"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: []
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: 迁移 7 个视图级 scss 的旧蓝字面量为靛紫

## Description

Phase 1 令牌层切 Linear 靛紫后，视图级 scss 仍残留旧蓝字面量（`#3b82f6` / `#2563eb` / `rgba(59,130,246,X)`），导致站内蓝/靛紫两套并存。本任务按既定约定（沿用 `fc0b2e4` 数据层迁移）把 7 个视图级 scss 的 25 行旧蓝机械迁移为靛紫：不改选择器结构、不新增变量、不改任何行为，之后以构建 + 浏览器双主题实测验收。

## Reference Files

- `docs/proposals/ui-unify-old-blue-sweep/proposal.md` — Source proposal（Proposed Solution / Scope > In Scope / Key Risks）
- `src/views/cmdb/models/styles/models.scss:63-66` — 主按钮 glow + hover `#2563eb`（映射表核心样例）(ref: Proposed Solution)
- `src/views/cmdb/instances/styles/instances.scss:66-67` — hover `#2563eb` + 行 115 `.blue` (ref: Proposed Solution)

## Acceptance Criteria

- [ ] 按映射表迁移全部 7 文件 25 行：`#3b82f6`→`#7170ff`、`rgba(59,130,246,X)`→`rgba(113,112,255,X)`、`#2563eb`→`#5e6ad2`，不改选择器结构、不新增 token
- [ ] `grep -rn "3b82f6\|2563eb\|59, *130, *246" src/views --include="*.scss"` 零命中（排除 `__tests__`）
- [ ] `pnpm build` 与 `pnpm typecheck` 均通过
- [ ] 浏览器经 nginx `:8888/cam/` 深/浅两主题：`cmdb/models` 页旧蓝计算样式命中 = 0（当前深色 35），并抽查其余 6 页无渲染回归、无控制台错误

## Hard Rules

（无）

## Implementation Notes

**精确文件清单（仅这 7 个文件，共 25 行）**：

| 文件 | 行 | 数量 |
|------|-----|------|
| `src/views/cmdb/models/styles/models.scss` | 63,65,66,140,143,273,308,347,438 | 9 |
| `src/views/cmdb/instances/styles/instances.scss` | 66,67,115,302 | 4 |
| `src/views/cmdb/relations/styles/relations.scss` | 57,98,256,306 | 4 |
| `src/views/cmdb/topology/styles/topology.scss` | 150,211,255 | 3 |
| `src/views/compute/ecs/styles/instances.scss` | 66,67,115,302 | 4 |
| `src/views/storage/styles/export-dialog.scss` | 122 | 1 |
| `src/views/storage/styles/column-settings.scss` | 13 | 1 |

**映射表（唯一权威，勿自创色值）**：
- `#3b82f6` → `#7170ff`（实心 accent-blue）
- `rgba(59, 130, 246, X)` → `rgba(113, 112, 255, X)`（accent-blue 微透，X 不变）
- `#2563eb` → `#5e6ad2`（主按钮 hover 暗化，用户已拍板；勿用 `#828fff`）

**验收注意**：
- `pnpm lint` 有 3 个历史 error（`src/api/types/{cost,statistics,task}.ts`，均自 `ab809ae first commit` 起），判据为「相对基线零新增」，勿让历史 error 阻塞本任务。
- 浏览器验收沿用 Phase 1 经验：真实登录态经 `http://localhost:8888/cam/`（勿直连 :5173，会无共享 cookie），用 chrome-devtools-mcp 的 evaluate_script 对 `getComputedStyle` 扫 `rgb(59,130,246)`/`rgb(37,99,235)`/`rgba(59,130,246` 三类旧蓝，断言深/浅两主题命中归零。
- `rgba(59,130,246)` 存在带/不带空格两种写法，迁移后用 `59,\s*130,\s*246` 正则复核清零。