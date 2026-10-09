---
title: "错误处理约定"
domains: [error, status, response, forbidden, validation, conflict, unavailable]
---

# 技术约定：错误处理

> 跨特性可复用约定。源：feature/platform-console。

## 统一错误对象

### TECH-error-001: unwrapEnvelope 单入口 + ApiError 单一错误对象

**Requirement**: 所有 api 客户端函数必经 `unwrapEnvelope<T>()` 解包；视图层只 catch 唯一错误对象 `ApiError`（`kind: ErrorKind`, `code: number|null`, `httpStatus: number|null`）。两种错误形态统一解包：① HTTP 2xx + `code != 0` → 抛 ApiError；② HTTP 非 2xx 可解析为信封 → 取 `{code,msg}` 抛同一 ApiError，不可解析 → 按 HTTP 状态映射 kind。`ErrorKind = 'unauthorized' | 'forbidden' | 'validation' | 'conflict' | 'unavailable' | 'unknown'`。`FieldErrors = Record<string,string>`。前端不假设 eiam 业务码区段；先 HTTP 状态、后信封 code，产出「状态 × code → kind」对照表。
**Source**: feature/platform-console tech-design §Error Handling

## 状态分类学

### TECH-error-002: 401 处理

**Requirement**: `unauthorized`(401) → `redirectToLogin('/console/login')`（不清共享 cookie）+ 中断当前路由全部在途请求 + list store 复位 loading/error。
**Source**: feature/platform-console tech-design §Error Handling 分类学表

### TECH-error-003: 403 原地 toast 不跳转

**Requirement**: `forbidden`(403) → 原地 toast（文案契约 `forbidden`），不跳转登录页。
**Source**: feature/platform-console tech-design §Error Handling 分类学表

### TECH-error-004: 400/422 字段级错误就近渲染

**Requirement**: `validation`(400/422) → `FieldErrors` 就近渲染（`aria-describedby` 关联），不弹全局 toast。
**Source**: feature/platform-console tech-design §Error Handling 分类学表

### TECH-error-005: 409 并发与依赖删除

**Requirement**: `conflict`(409) → 并发编辑提示「数据已被他人修改，请刷新后重试」+ 自动重拉当前页；依赖删除被拒弹窗列阻断计数、确认置灰。Phase 0 核查 eiam 是否返回 409，不支持则该分支降级并入 `validation`。
**Source**: feature/platform-console tech-design §Error Handling 分类学表

### TECH-error-006: unavailable 降级

**Requirement**: `unavailable`(网络错 / 15s 超时 / 5xx) → 页面错误态 + 重试；重试 ≥ 2 次仍失败 → 顶部降级提示 + 写操作禁用（静态内容仍渲染，e-cam-web 会话不受影响）。
**Source**: feature/platform-console tech-design §Error Handling 分类学表

## 文案契约

### TECH-error-007: 文案契约零二次映射

**Requirement**: `code != 0` 的业务码原值透传给文案契约 key（`code → 逐字文案`映射表见 ui-design）；`ApiError.message` 即契约文案，视图层零二次映射；字段级校验就近 `aria-describedby`。
**Source**: feature/platform-console tech-design §Error Handling 文案契约
