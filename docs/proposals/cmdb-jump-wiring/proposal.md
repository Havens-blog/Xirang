---
created: "2026-09-20"
author: "Haven"
status: Approved
intent: "fix"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: CMDB 跳转接线修复（H-02 模型→实例 + H-03 实例→资产）

## Problem

CMDB 两条核心下钻链路全部断线，用户点击跳转后落到**未过滤**的列表页：
1. **H-02**：模型列表「查看实例」跳 `/cmdb/instances?model_uid=<uid>`，但 instances 页**从不读取** `model_uid`（未 import `useRoute`）——落全量实例列表。
2. **H-03**：实例详情「查看资产」跳 10 个资产类型列表页并带 `query.search`，但 10 个目标页**全部不读取**该参数（仅 ecs 读 `query.uid`，键名还不匹配）——落未过滤列表。

### Evidence

- `cmdb/models/index.vue:428`：`router.push({ path: '/cmdb/instances', query: { model_uid: model.uid } })`。
- `cmdb/instances/index.vue:236` 仅 `import { useRouter }`；`filters = { keyword, asset_type, provider, region }` 无 model_uid；后端 `cmdb/web/instance_handler.go:213` **已支持** `model_uid` 查询参数——纯前端接线。
- `cmdb/instances/index.vue:578-580`：`router.push({ path: routeInfo.route, query: { search: row.asset_id } })`，`assetTypeRouteMap` 覆盖 10 类型（ecs/rds/redis/mongodb/vpc/eip/nas/oss/kafka/elasticsearch）。
- 10 目标页搜索语义全部为后端 `name` 模糊，前端三种形态：`filters.keyword`（rds/redis/mongodb/disk/kafka/elasticsearch）、`filters.name`（vpc/eip）、`searchKeyword` ref（nas/oss）；ecs 有 `filters.keyword`（映射后端 `name`）+ 既有 `query.uid` 读取。
- **语义错配**：发送方传 `row.asset_id`（云 ID），后端 name 搜索匹配的是 `asset_name`（展示名）——即使目标页读了 `search` 也匹配不上；应发 `asset_name`（`InstanceVO.asset_name?` 可空，`asset_id` 兜底）。

### Urgency

- **HIGH**：两条下钻是 CMDB 的核心导航动线；点击「查看实例/查看资产」后落到未过滤列表，用户需手动重搜，功能形同虚设。修复纯前端、路径清晰。

## Proposed Solution

统一接线，四段式：
1. **H-02**：instances 页 onMounted 读 `route.query.model_uid` → 加入列表请求参数（后端已支持）+ 可清除的过滤标签（明示当前按模型过滤）。
2. **H-03 接收方**：10 页统一模式读 `route.query.search` → 各自搜索字段并触发列表加载：keyword 系 5 页（rds/redis/mongodb/kafka/elasticsearch → `filters.keyword`）、name 系 2 页（vpc/eip → `filters.name`）、searchKeyword 系 2 页（nas/oss → `searchKeyword`）；ecs 把 `search` 读入 `filters.keyword`（与既有 `query.uid` 读取并存）。
3. **H-03 发送方**：`handleViewDetail` 改发 `search: row.asset_name || row.asset_id`（匹配后端 name 搜索语义），在接收方接线完成后切换。

### Innovation Highlights

直白接线修复，无创新实现；价值在打通 CMDB 两条核心下钻动线。

## Requirements Analysis

### Key Scenarios

- 模型列表点「查看实例」→ 实例页仅显示该模型实例，清除标签可回到全量。
- 实例详情点「查看资产」（10 类型）→ 目标列表页按 asset_name 预填搜索并加载过滤结果。
- 直达目标页（无 query）→ 行为与现状完全一致。

### Non-Functional Requirements

- 零冗余请求：读取 query 在首次 fetch 前完成（不发生先全量后过滤的二次请求）。
- 各页既有筛选行为不变（无 query 时零改动）。

### Constraints & Dependencies

- 仅前端：`cmdb/instances/index.vue`（H-02 + 发送方）+ 10 个目标页 index.vue。
- 不触碰 `accounts/index.vue`（并发脏）、logs、后端（参数已支持）、`components.d.ts`。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 统一接线（本提案） | 审计建议 | 两条动线全通；后端零改动 | 12 文件触达 | Adopt（用户已确认） |
| 只做 H-02 | 备选 | 1 文件 | H-03 十页断线仍在 | Rejected：同主题拆两次 |
| 目标页统一加 keyword 过滤器 | 备选 | 键名统一 | vpc/eip/nas/oss 需新增 UI 字段，超出接线范围 | Rejected：用各页既有搜索字段 |

## Scope

### In Scope

- `cmdb/instances/index.vue`：读 `model_uid` 入请求参 + 过滤标签；`handleViewDetail` 改发 `asset_name || asset_id`。
- 10 目标页接线（6 keyword + 2 name + 2 searchKeyword + ecs keyword 并存）。
- 各页读取时机在首次 fetch 前。

### Out of Scope

- 后端参数新增（均已支持或无需）。
- vpc/eip/nas/oss 新增可见搜索框 UI（用既有字段）。
- 其它模块跳转（accounts A-C1 等待并发）、`components.d.ts`。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| asset_name 为空或重名导致定位不准 | Medium | 发送方 `asset_name || asset_id` 兜底；name 模糊搜索重名时列表已收敛到候选集（优于现状全量） |
| 读取时机晚于首次 fetch（二次请求） | Medium | 统一在各页 onMounted 的首次 fetch 之前赋值；代码评审点 |
| ecs 双参数并存冲突（uid + search 同到） | Low | 现实中单跳只带其一；两者作用于不同过滤维度（uid 精确 / keyword 模糊），共存无害 |
| 10 页模式复制走样 | Medium | 统一三行模式（读 query → 赋字段 → 依赖既有首载）；逐页 typecheck |

## Success Criteria

- [ ] H-02：模型「查看实例」→ 实例页仅显示该模型实例，请求携带 `model_uid`；过滤标签可清除恢复全量
- [ ] H-03：10 类型「查看资产」→ 目标页按 `asset_name` 预填搜索字段且首屏即为过滤结果（5 keyword + 2 name + 2 searchKeyword + ecs）
- [ ] 发送方改发 `asset_name || asset_id`
- [ ] 无 query 直达各页行为与现状一致（零改动验证）
- [ ] `pnpm typecheck` + vitest 全绿（并发在途错误除外）；`git diff` 仅 12 个 index.vue
