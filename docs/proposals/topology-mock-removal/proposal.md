---
created: "2026-09-19"
author: "Haven"
status: Approved
intent: "fix"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: 拓扑页移除 mock 回退（诚实错误态）

## Problem

`e-cam-web/src/views/topology/` 在后端 API 失败时**静默回退到 mock 假数据**，页面顶部同时无条件挂着绿色「实时」徽章——用户把示例拓扑当真实业务链路。与已修复的 IAM compare 页同类数据完整性缺陷。

### Evidence

- `composables/useTopologyData.ts:22-25`（domains）与 `:39-41`（graph）`catch` 块仅 `console.warn` 后赋值 `MOCK_DOMAINS`/`MOCK_GRAPH`。
- `views/topology/index.vue:7-10` 徽章写死 `实时`（绿色 + 呼吸动画），无论数据真假。
- 后端 `e-cam-service/internal/topology/` 已完整实现（web handler、`live_builder.go`、`dag_builder.go`、collectors、DAO）；前端 `api/topology.ts` 全部端点就绪。
- `useMockData.ts:5` 自带注释「后端可用后删除此文件，在 useTopologyData.ts 中去掉 fallback 即可」——后端已可用。
- 全前端 grep 确认：生产 mock 回退仅此一处；其余 `vi.mock` 均为 vitest 测试基建；`config/index.ts:30` 的 `useMock` 开关是死配置（无人消费）。

### Urgency

- **HIGH**：故障时展示假拓扑并标「实时」，可能误导故障研判决策。范围小、后端已就绪、修复路径清晰。

## Proposed Solution

拓扑页移除全部 mock 回退与假数据，恢复诚实显示：
- `useTopologyData.ts`：去掉 `MOCK_DOMAINS`/`MOCK_GRAPH` 引用；API 失败 → 记录错误状态，`domainList`/`graphData` 保持空。
- `stores/topology.ts`：新增错误/数据状态字段（如 `loadError: string | null` 或数据状态枚举），供 UI 区分 实时 / 空数据 / 失败。
- `views/topology/index.vue`：徽章与真实状态联动（成功→绿色「实时」；失败→灰/橙「离线」或「异常」）；失败时显示可读错误提示（`el-alert`），画布空态诚实展示（无数据/失败均有空态文案）。
- 删除 `composables/useMockData.ts`。

### Innovation Highlights

诚实数据 > 优雅占位，与 IAM compare 修复同口径。无创新实现，价值在消除最后一处生产 mock。

## Requirements Analysis

### Key Scenarios

- 后端正常返回：与现状一致，绿「实时」徽章 + 真实拓扑。
- 后端返回空（无 DNS 记录/无拓扑）：正常空态文案，不视为错误。
- 后端 API 失败（网络/5xx）：`el-alert` 错误提示 + 画布空态；徽章灰/橙「离线」或「异常」；刷新按钮可重试。
- 刷新/筛选变化触发重新加载：错误状态随之更新。

### Non-Functional Requirements

- 生产代码无 `MOCK_DOMAINS`/`MOCK_GRAPH` 残留（grep 断言）；`useMockData.ts` 删除。
- `config/index.ts` 的 `useMock` 死开关同步清理（可选，若无人引用）。

### Constraints & Dependencies

- 仅改拓扑页 4 个文件：`useTopologyData.ts`、`stores/topology.ts`、`views/topology/index.vue`、删除 `useMockData.ts`（+ 必要测试）。
- 不触碰 `accounts/index.vue`（并发会话脏文件）、logs 文件、后端。
- 组件 `TopologyCanvas`/`FilterPanel`/`DetailPanel` 不改（对空 `graphData` 已有或可容忍空态）。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 移除 mock + 诚实错误态 | 本提案 | 可信数据、消除「实时」误导、符合文件自带删除注释 | 失败时无图可看（诚实代价） | Adopt（用户已确认） |
| 保留 mock 但标注「示例数据」 | 备选 | 保留开发脚手架 | 仍展示假数据，运维可能误用 | Rejected：用户已选移除 |
| 仅修「实时」徽章联动 | 备选 | 最小改动 | 假数据仍显示，误导仍在 | Rejected：用户已选移除 |
| Do nothing | — | 零改动 | 假数据+「实时」徽章误导持续 | Rejected |

## Scope

### In Scope

- `composables/useTopologyData.ts`：移除 mock 回退；API 失败 → 设置错误状态，数据保持空。
- `stores/topology.ts`：新增错误/数据状态字段。
- `views/topology/index.vue`：徽章与数据状态联动；失败错误提示（`el-alert`）；空态文案。
- 删除 `composables/useMockData.ts`。
- （可选）`config/index.ts` 清理 `useMock` 死开关。

### Out of Scope

- `TopologyCanvas`/`FilterPanel`/`DetailPanel` 组件改造。
- 后端拓扑模块改动。
- 其它页面的 mock 排查（grep 已确认仅此一处）。
- 拓扑节点详情、声明确认等其它既有行为。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 移除 fallback 后开发环境无后端时画布空白，排障体验下降 | Medium | 错误提示 `el-alert` 明确指引（后端未启动/接口失败）；保留刷新重试；文档/README 注明后端依赖 |
| store 新增状态字段影响既有引用 | Low | 只加不改；`filteredNodes`/`queryParams` 等现有字段不动 |
| 徽章状态与 loading 竞态（刷新中短暂灰/绿跳变） | Low | 用 loading 态占位（刷新中灰「加载中」），成功后再绿 |
| 删除 useMockData.ts 后其它文件引用报编译错 | Low | 全局 grep 已确认仅 useTopologyData.ts 引用；删除前再 grep 断言 |

## Success Criteria

- [ ] `useTopologyData.ts` 无 `MOCK_DOMAINS`/`MOCK_GRAPH` 引用；`useMockData.ts` 已删除（grep 断言全前端无残留）
- [ ] 后端 API 失败 → `el-alert` 错误提示 + 空画布；徽章灰/橙「离线」或「异常」（不再显示绿色「实时」）
- [ ] 后端返回空数据 → 正常空态（不误报错误）；后端成功 → 绿「实时」徽章 + 真实拓扑（行为与现状一致）
- [ ] 刷新/筛选重新加载时错误状态正确更新；加载中徽章不误显「实时」
- [ ] `pnpm typecheck` + 相关 vitest 通过；`git diff` 仅拓扑 4 文件（+ 测试/config 可选）
