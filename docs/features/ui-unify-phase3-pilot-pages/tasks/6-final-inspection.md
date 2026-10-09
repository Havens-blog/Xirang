---
id: "6"
title: "最终全站双主题浏览器巡检（回归确认）"
priority: "P1"
estimated_time: "1h"
complexity: "medium"
dependencies: [1, 2, 3, 4, 5]
surface-key: "e-cam-web"
surface-type: "web"
breaking: false
type: "coding.feature"
mainSession: false
---

# 6: 最终全站双主题浏览器巡检（回归确认）

## Description
四页重写 + MainLayout 适配全部落地后的收尾巡检：四标杆页 + /dev/ui-kit demo 页 + 主题切换，深浅双主题全量过一遍，确认无回归、无破相。发现问题清单化，阻断性问题当场修复。

## Reference Files
- `docs/proposals/ui-unify-phase3-pilot-pages/proposal.md` — Success Criteria（最终巡检条款）
- `src/views/dev/ui-kit/index.vue` — demo 页

## Acceptance Criteria
- [ ] chrome-devtools-mcp 巡检：四标杆页 + demo 页深/浅双主题逐页截图归档，无布局破损/无不可读文本/无旧蓝残留
- [ ] 全量测试无回归（≥775 基线全绿），`pnpm build` + `pnpm typecheck` 零新增错误
- [ ] 巡检发现的问题清单化写入执行报告；阻断性问题当场修复并复验，非阻断项记录缓期

## Hard Rules
- 修复仅限阻断性视觉/功能缺陷；非阻断项一律记录不修（防 scope creep）

## Implementation Notes
- 巡检路径：dashboard → assets → alert/events → dictionary → /dev/ui-kit → ⌘K 抽查 → 双主题切换
- 巡检也算 Phase 3 的最终用户验收素材：截图清单随执行记录归档
