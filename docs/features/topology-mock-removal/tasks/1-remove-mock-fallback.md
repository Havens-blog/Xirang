---
id: "1"
title: "拓扑页移除 mock 回退，改诚实错误态"
priority: "P0"
estimated_time: "1h"
complexity: "high"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.fix"
mainSession: false
---

# 1: 拓扑页移除 mock 回退，改诚实错误态

## Description
**代码仓：D:/Haven/e-cam-web**（Vue3+TS+pnpm）。`views/topology/` 后端 API 失败时静默 fallback 到
`MOCK_DOMAINS`/`MOCK_GRAPH`，且 `index.vue` 顶部无条件挂绿色「实时」徽章——用户把示例拓扑当真实业务链路。
改为：移除全部 mock 回退；API 失败 → store 记录错误态 + `el-alert` 提示 + 空画布；徽章随数据状态联动
（成功=绿「实时」、失败=灰/橙「离线/异常」、加载中不误显实时）；空数据正常空态。删除 `useMockData.ts`。

## Reference Files
- `docs/proposals/topology-mock-removal/proposal.md` — Proposed Solution, Scope > In Scope, Risks, Success Criteria
- `D:/Haven/e-cam-web/src/views/topology/composables/useTopologyData.ts` — 两个 catch 块 mock 回退（loadDomains:22-25 / loadTopology:39-41）
- `D:/Haven/e-cam-web/src/views/topology/composables/useMockData.ts` — 待删除（自带「后端可用后删除」注释）
- `D:/Haven/e-cam-web/src/stores/topology.ts` — 无错误态字段（新增 loadError 或状态枚举）
- `D:/Haven/e-cam-web/src/views/topology/index.vue:7-10` — 写死「实时」徽章（绿色 + 呼吸动画）
- `D:/Haven/e-cam-web/src/config/index.ts:30` — `useMock` 死开关（可选清理）

## Acceptance Criteria
- [ ] `useTopologyData.ts` 无 `MOCK_DOMAINS`/`MOCK_GRAPH` 引用；`useMockData.ts` 已删除（grep 断言全前端无残留）
- [ ] 后端 API 失败 → `el-alert` 错误提示 + 空画布；徽章不再显示绿色「实时」（灰/橙「离线/异常」）
- [ ] 后端成功 → 绿「实时」徽章 + 真实拓扑（行为与现状一致）；后端空数据 → 正常空态（不误报错误）
- [ ] 刷新/筛选重新加载时错误状态正确更新；loading 中徽章不误显「实时」
- [ ] `pnpm typecheck` + 相关 vitest 通过；`git diff` 仅拓扑 4 文件（+ 测试/`config/index.ts` 可选）

## Hard Rules
仅改：`src/views/topology/composables/useTopologyData.ts`、`src/stores/topology.ts`、`src/views/topology/index.vue`、删除
`src/views/topology/composables/useMockData.ts`（+ 必要时新增测试、`src/config/index.ts` 清死开关）。
**不触碰** `src/views/accounts/index.vue`（并发会话脏文件）、`src/views/logs/`、后端任何文件、其它 topology 组件。
提交用 `git -C D:/Haven/e-cam-web add <pathspec>` 显式路径，不 `git add -A`。

## Implementation Notes
- store 新增错误/数据状态（推荐 `loadError: string | null` 或 `dataStatus: 'idle'|'loading'|'live'|'error'`），
  `queryParams`/`filteredNodes` 等既有字段不动。
- 徽章联动：加载中→灰「加载中」；失败→橙「离线/异常」（附错误文案）；成功→绿「实时」（现状）。
- 删除 `useMockData.ts` 前再 grep 确认仅 `useTopologyData.ts` 引用（提案已全前端确认）。
- 保留 `refresh()` 重试路径；错误提示指引「后端未启动/接口失败，请刷新重试」。
- 空数据（后端正常返回空数组）与失败（异常）必须区分：空数据→空态文案，失败→错误态。
- 参考 `useTopologyData.ts:18-45` 现有 try/catch 结构；组件 `TopologyCanvas`/`FilterPanel`/`DetailPanel` 不改（对空 graphData 已有容忍）。
