---
title: "前端状态管理约定"
domains: [store, pinia, persist, tenant, state, abortcontroller, list]
---

# 技术约定：前端状态管理

> 跨特性可复用约定。源：feature/platform-console。

## 持久化

### TECH-frontend-state-001: persistedstate 仅挂 theme

**Requirement**: `persistedstate` 插件仅挂 `theme` 一个 store（key `hc_theme`）；任何会话 / 列表 store 不得声明 `persist`——杜绝插件默认全量写 localStorage；凭据不落 localStorage。
**Source**: feature/platform-console tech-design §Architecture Store 清单

## list store 契约

### TECH-frontend-state-002: list store 统一契约

**Requirement**: list store 统一契约：`state: { items, total, page, pageSize(默认 20、上限 100), loading, error }` + `fetch / create / update / remove`；每个 store 持有 `AbortController`，fetch 前中断在途请求。
**Source**: feature/platform-console tech-design §Architecture Store 清单

## 租户上下文

### TECH-frontend-state-003: 租户上下文唯一归属

**Requirement**: 租户上下文唯一归属 `stores/tenant.ts`；api 层零租户状态（仅「参数进、类型出」纯调用）；stores 可持有分页 / 筛选 UI 态。
**Source**: feature/platform-console tech-design §Architecture Store 清单 职责边界

### TECH-frontend-state-004: 切租户必走 tenant/switch + 整页 reload

**Requirement**: 所有切租户路径成功后必走 `POST /api/iam/tenant/switch` → 整页 `location.replace()`，杜绝 SPA 内只换 store 不刷页（e-cam-web TenantSelector 同款约定）。tenant_id 唯一存在于服务端共享 session，不引入 per-tab 快照、不做 BroadcastChannel 广播。
**Source**: feature/platform-console tech-design §Integration Specs 6

### TECH-frontend-state-005: 响应回读租户比对自愈

**Requirement**: list store 每次 fetch 成功后回读响应所属租户标识与 `currentTenantId` 比对（响应无租户标识则跳过此层），不一致 → 丢弃渲染并 `location.replace()` 自愈。
**Source**: feature/platform-console tech-design §Integration Specs 6
