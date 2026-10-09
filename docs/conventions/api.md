---
title: "eiam API 契约约定"
domains: [api, envelope, contract, eiam, serialization, pagination, http-client]
---

# 技术约定：eiam API 契约

> 跨特性可复用约定。源：feature/platform-console。

## 信封与归一

### TECH-api-001: eiam 信封格式

**Requirement**: 所有 eiam 响应为 ginx 信封 `{ code: number; msg: string; data: T }`，成功 `code === 0`；字段是 `msg` 不是 `message`（实盘核验自 e-cam-web `src/api/eiam-mapper.ts`）。
**Source**: feature/platform-console tech-design §Interfaces 信封类型

### TECH-api-002: ListQuery / Page\<T> 归一类型

**Requirement**: 请求统一 `interface ListQuery { offset: number; limit: number; keyword: string }`；响应 api 层归一为 `interface Page<T> { total: number; items: T[] }`（把 eiam 的 `{total, users|roles|tenants}` 归一为 `items`）。每端点一个具名函数（如 `listUsers(q): Promise<Page<User>>`）。
**Source**: feature/platform-console tech-design §Interfaces 统一请求/归一类型

## HTTP 客户端

### TECH-api-003: eiamAxios 请求客户端模式

**Requirement**: axios 实例 `eiamAxios`：`withCredentials: true`、请求拦截从 cookie 读 token 注入 `Authorization: Bearer`、响应拦截 401 → `redirectToLogin()` 跳 `/console/login`（`isRedirectingToLogin` 防抖，不清共享 cookie）、15s 超时。
**Source**: feature/platform-console tech-design §Overview / §Architecture 组件关系图

## 分页

### TECH-api-004: 分页边界行为

**Requirement**: 超界页码回落最后一页；空页显示空状态；删除末页唯一记录后回退上一页。
**Source**: feature/platform-console prd-spec.md E7
