---
feature: "platform-console"
journey: "tenant-boundary-enforcement"
risk_level: "Medium"
golden_path: true
surface_types: ["web", "api"]
surface_keys: ["haven-console", "eiam"]
sources:
  - docs/features/platform-console/prd/prd-user-stories.md
  - docs/features/platform-console/prd/prd-spec.md
generated: "2026-10-08"
---

# Journey: tenant-boundary-enforcement

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
  Medium = Workflow involves multi-step interaction without irreversible side effects
  Low    = Workflow is read-only or purely observational
-->

## Overview

租户管理员在租户边界内治理本租户用户/角色/授权（可绑定已有策略、不可增删策略定义），并验证任何越出租户边界的访问均被 eiam 服务端拒绝、原地收敛且数据不变。本 Journey 覆盖 Web（浏览器交互）与 API（HTTP 客户端）两个 surface：正常治理流程以 Web E2E 验证，越界与无凭据访问以 API Functional 验证服务端边界。

## Setup

- 租户管理员账号已存在且已登录，已选定本租户（单租户自动选定或多租户经切换器选定）
- 本租户内存在用户数据；其他租户数据存在且租户管理员不应可见
- eiam 服务可用；eiam 对所有 CRUD 端点做服务端权限校验（租户边界为安全边界）
- 控制台前端菜单已按租户管理员角色裁剪（策略定义增删改入口不渲染）

## Fact Table

| fact_id | Claim | Source |
|---------|-------|--------|
| F1 | 租户管理员查看用户列表时仅显示本租户用户，无法看到其他租户数据 | prd-user-stories.md Story 3 AC1 |
| F2 | 租户管理员越权请求时 eiam 返回 403，页面内提示无权限并保持当前页（不跳转登录页），数据不变 | prd-user-stories.md Story 3 AC2; prd-spec.md Flow #8 |
| F3 | 策略定义增删改仅平台管理员可用；租户管理员仅可在本租户内创建角色、绑定已有策略并分配给用户 | prd-user-stories.md Story 4 口径; prd-spec.md Who 表 |
| F4 | eiam 对所有 CRUD 端点做服务端权限校验，前端菜单裁剪仅为体验层、非安全边界 | prd-spec.md Security Requirements |
| F5 | 会话过期 → 跳转控制台登录页；无权限（403）→ 原地提示、不跳转登录页（两者分流） | prd-spec.md Flow #8 |
| F6 | 表单提交字段校验失败时对应字段标红并回显错误原因、表单数据保留、列表数据不变；修正后可重新提交 | prd-user-stories.md Story 2 AC3 |
| F7 | 任一侧登出即销毁共享 session，其余应用/标签页在下一次请求时 401 收敛并跳回控制台登录页 | prd-user-stories.md Story 6; prd-spec.md Flow #7 |
| F8 | 角色分配后该用户下一次请求即具备对应权限（逐请求判定，无需重新登录） | prd-user-stories.md Story 4 AC1 |
| F9 | 同一对象并发编辑：后提交者生效；若 eiam 返回 409 版本冲突则提示刷新后重试 | prd-spec.md E6 |
| F10 | 可绑定策略仅含 eiam 返回的列表，无新建/编辑/删除策略定义入口 | prd-user-stories.md Story 4 AC2 |

### Inferred Claims

| claim_id | Claim | Source | Reasoning Basis |
|----------|-------|--------|-----------------|
| I1 | 403 越权响应不泄露其他租户资源是否存在 | source: inferred | Story 1 AC 明确「凭据错误文案不泄露用户名是否存在」的存在性不泄露原则；将该原则推广至 403 越权响应（安全最小信息披露） |
| I2 | 租户管理员提交含非法字段的治理表单时显示字段级校验错误 | source: inferred, required_outcomes: `web/validation-error` | 由 web surface `required_outcomes` 规则派生；事实依据 F6（Story 2 AC3 校验口径） |
| I3 | 租户管理员在治理流程中会话过期时跳转登录页且未保存数据得到提示 | source: inferred, required_outcomes: `web/session-expired` | 由 web surface `required_outcomes` 规则派生；事实依据 F5（Flow #8 会话过期分流） |
| I4 | 无有效凭据调用治理接口时返回 401 Unauthorized | source: inferred, required_outcomes: `api/unauthorized` | 由 api surface `required_outcomes` 规则派生；事实依据 F7（Flow #7 登出后 401 收敛） |

## Happy Path

> Golden Path：租户管理员在本租户边界内完成「查看用户 → 创建角色并绑定策略 → 分配角色」的治理闭环，对应 PRD Story 3 / Story 4 租户管理员 AC。三条 Step 均为边界内的正向治理操作。

### Step 1: 查看本租户用户列表 @web-e2e

**Precondition**: 租户管理员已登录并已选定本租户；本租户与其他租户均存在用户数据。

**User Action**: 租户管理员在用户管理页查看用户列表。

**Expected Result**: 列表仅显示本租户用户，不显示其他租户数据；分页与筛选在本租户范围内生效。 [fact_id: F1]

### Step 2: 在本租户创建角色并绑定已有策略 @web-e2e

**Precondition**: 租户管理员在本租户内；可绑定策略列表由 eiam 返回。

**User Action**: 租户管理员在本租户内创建一个新角色，从可绑定策略列表中选择一个策略绑定。

**Expected Result**: 绑定成功；可绑定策略仅含 eiam 返回的列表，无新建/编辑/删除策略定义入口。 [fact_id: F10, F3]

### Step 3: 将角色分配给本租户用户并验证逐请求生效 @web-e2e

**Precondition**: Step 2 创建的角色已存在于本租户；待分配用户属于本租户。

**User Action**: 租户管理员将该角色分配给本租户用户，随后该用户发起一次需该权限的请求。

**Expected Result**: 分配成功；该用户下一次请求即具备对应权限，逐请求判定、无需重新登录。 [fact_id: F8]

## Edge Cases

> 以下 Step 通过 Precondition 区分多 Outcome。边界越权与无凭据访问场景移出 Happy Path，归入此处。

### Step 4: 租户管理员访问其他租户治理资源 @web-e2e

**User Action**: 租户管理员尝试访问不属于本租户的治理资源（其他租户用户/角色/授权，或策略定义管理）。

**Outcome A — 越权访问被服务端拒绝（Precondition: 目标资源归属其他租户，当前会话有效）**:
eiam 返回 403，页面内提示无权限并保持当前页（不跳转登录页），数据不变；越权提示不暴露其他租户资源是否存在。 [fact_id: F2, F4; inferred: I1]

**Outcome B — 会话过期（Precondition: 操作过程中会话已过期）**:
页面跳转控制台登录页，提示会话已过期，未保存的表单数据得到保留或丢失告警；登录流程可从过期状态进入。 [fact_id: F5; inferred: I3]

**Outcome C — 表单校验失败（Precondition: 提交的表单含非法字段，如必填项为空或格式不符）**:
对应字段标红并回显错误原因，表单数据保留，列表数据不变，可修正后重新提交；请求不送达服务端。 [inferred: I2]

**Outcome D — 策略定义增删改入口不可达（Precondition: 操作者为租户管理员，目标为策略定义管理）**:
策略定义增删改入口不渲染；即使绕过前端直接以有效凭据请求，eiam 返回 403，页面内提示无权限并保持当前页，数据不变，不泄露资源存在性。 [fact_id: F3, F2; inferred: I1]

### Step 5: 以 HTTP 客户端直接访问治理接口 @api-functional

**User Action**: 调用方以 HTTP 请求访问 eiam 治理接口（用户/角色/授权/策略 CRUD）。

**Outcome A — 无有效凭据（Precondition: 请求未携带有效凭据，或携带的凭据已失效）**:
返回 401 Unauthorized，响应体含认证错误信息，不泄露任何租户数据；不与 403 越权混淆。 [fact_id: F7; inferred: I4]

**Outcome B — 有效凭据但跨租户访问（Precondition: 凭据有效且归属本租户，但请求目标属于其他租户）**:
返回 403 Forbidden，响应不泄露其他租户资源是否存在，数据不变。 [fact_id: F2, F4; inferred: I1]

**Outcome C — 有效凭据但尝试增删改策略定义（Precondition: 凭据有效但操作者为租户管理员，目标为策略定义写操作）**:
返回 403 Forbidden，不泄露资源存在性，数据不变。 [fact_id: F3, F2; inferred: I1]

**Outcome D — 有效凭据访问本租户资源（Precondition: 凭据有效，请求目标属于本租户且操作在租户管理员权限内）**:
按各 CRUD 语义返回 200/201/204，响应仅含本租户数据。 [fact_id: F1, F4]

### Step 6: 并发跨租户编辑互不干扰 @api-functional

**Precondition**: 两个不同租户的管理员同时编辑各自租户内的同一类型对象。

**User Action**: 双方各自提交编辑。

**Expected Result**: 各租户内后提交者各自生效；不发生跨租户数据串改；若同租户内同一对象并发触发 409 版本冲突，则提示刷新后重试。租户边界在并发下仍成立。 [fact_id: F9, F4]

## Journey Invariants

- **租户边界由 eiam 服务端强制** [fact_id: F4]——前端菜单裁剪仅为体验层；任何越界访问在服务端被拒。
- **403 原地收敛** [fact_id: F2, F5]——越权请求保持当前页、不跳转登录页、数据不变；与「会话过期跳登录」严格分流（403 原地 vs 过期跳转）。
- **401 与 403 语义分离** [fact_id: F7, F2]——无凭据/凭据失效返回 401（认证失败）；有效凭据但越权返回 403（授权失败），二者不混淆。
- **策略定义口径** [fact_id: F3, F10]——租户管理员仅可绑定已有策略、不可增删改策略定义；该入口在前端不渲染、在服务端被拒。
- **403 不泄露存在性** [inferred: I1]——所有 403 越权响应统一不暴露其他租户资源是否存在；此为推断性安全要求（source: inferred），依据 Story 1 AC 存在性不泄露原则推广，非 PRD 直接条款。
