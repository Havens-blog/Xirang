---
feature: "platform-console"
journey: "readonly-audit-access"
risk_level: "Low"
golden_path: true
surface_types: ["web", "api"]
surface_keys: ["haven-console", "eiam"]
sources:
  - docs/features/platform-console/prd/prd-user-stories.md
  - docs/features/platform-console/prd/prd-spec.md
generated: "2026-10-08"
---

# Journey: readonly-audit-access

**Risk Level**: Low

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
  Medium = Workflow involves multi-step interaction without irreversible side effects
  Low    = Workflow is read-only or purely observational
-->

## Overview

只读/审计角色登录控制台后，以只读形式查看用户、角色、授权等身份与权限页面以完成合规审计：浏览、搜索、分页均可正常使用且数据完整可读（正向审计能力可达成）；同时任何写入口（增删改、租户管理、策略管理）均不渲染，直接调用写接口或经 URL 访问受限管理页均被服务端权限校验拒绝（403），原地提示、数据不变、不丢失登录态。本旅程覆盖 PRD Story 5 的完整审计工作流（正向可读 + 写操作阻断）。

## Setup

- 只读/审计角色账号已存在且已登录控制台
- 用户/角色/授权页面通常存在可读数据（含多页数据以支撑分页审计），亦可能为空
- 身份与权限后端服务（eiam）可用；后端对所有用户/角色/授权的写操作做服务端权限校验，前端菜单裁剪仅为体验层、非安全边界
- 控制台前端菜单已按只读角色裁剪（增删改入口、租户管理与策略管理入口不渲染）

## Fact Table

| fact_id | Claim | Source |
|---------|-------|--------|
| F1 | 只读角色打开用户/角色/授权页面时增删改入口不渲染（不可用） | PRD Story 5 AC1 |
| F2 | 只读角色直接调用写接口被服务端拒绝（403），前端提示无权限且数据不变 | PRD Story 5 AC1 |
| F3 | 只读角色访问租户管理或策略管理时菜单入口不渲染 | PRD Story 5 AC2 |
| F4 | 只读角色经 URL 直接访问租户管理或策略管理被服务端拒绝（403），页面提示无权限且数据不变 | PRD Story 5 AC2 |
| F5 | 只读/审计角色可只读查看用户/组织/身份源/角色/授权页面（不含租户与策略管理），不可增删改 | PRD prd-spec.md（Who 表·只读/审计角色） |
| F6 | 无权限（403）→ 原地提示、不跳转登录页；会话过期 → 跳转控制台登录页（两者语义区分） | PRD prd-spec.md（Business Flow #8 异常分流） |
| F7 | 只读角色对用户/角色/授权页面的浏览、搜索、分页、筛选以只读方式正常执行并返回正确结果（审计可读性） | PRD prd-spec.md（只读查看页面）+ Story 5 AC1（审计目的） |
| F8 | 空页/空列表显示空状态占位（不报错、不白屏） | PRD prd-spec.md（E7 分页边界：空页显示空状态） |

> 事实类声明以 `[Fx]` 引用 Fact Table；推理类 Outcome（非 Fact Table 中的事实，由 surface 规则 `required_outcomes` 派生的边界场景）以 `source: inferred` 标注并引用**具体**触发规则（如 `surface-web.md–network-error`、`surface-web.md–validation-error`），不引用段落名 `required_outcomes` 本身。

## Contract Coverage Gap

surface-web.md 与 surface-api.md 均要求 Web/API 双面 50/50 的 Contract/Journey 比例。本文件为 Journey 侧产物；Contract 侧规范（`contracts/` 目录）当前为空，将下游由 `gen-contracts` 基于本 journey 自动生成。在该缺口闭合前，50/50 比例不可由工件直接验证；本条登记为已知缺口，下游 `gen-contracts` 产出后即满足比例要求。

## Happy Path (Golden Path)

> 以下 4 步构成一条锚定 PRD Story 5 的连续 Golden Path：登录进入只读视图 → 搜索分页验证审计完整性 → 直调写接口被拒 → 访问受限管理页被裁剪且 URL 直达被拒。
>
> **跨面执行模型**：Steps 1/2/4 为 `@web-e2e`（浏览器自动化，Playwright 驱动，覆盖渲染、菜单裁剪、URL 直达、空状态等浏览器可观测行为）；Step 3 为 `@api-functional`（HTTP 客户端），复用浏览器上下文中已认证只读角色的会话令牌，对该令牌直接发起写请求以断言 403/401。两段在同一测试用例内顺序执行：浏览器段完成登录与上下文建立 → 提取会话令牌 → HTTP 段发起越权写请求 → 浏览器段继续 URL 直达验证。如此 Golden Path 语义连续且跨面可单测执行，无需在运行期切换测试框架。

### Step 1: 只读角色登录并进入身份与权限页面（只读视图） @web-e2e

**Precondition**: 只读/审计角色已登录；用户/角色/授权页面可能含可读数据，亦可能为空。

**User Action**: 只读角色在控制台导航至用户管理、角色管理、授权管理页面。

**Outcomes**:
- **Outcome A（数据存在·只读可读）** [F1][F5]：增删改入口不渲染（不可用）；页面以只读形式展示数据；可正常浏览。审计目标可达成——身份与权限现状对只读角色可读且完整，不仅断言写被阻断。
- **Outcome B（数据为空·空状态）** [F8]：对应页数据为空时显示空状态占位，仍不渲染任何写入口；空状态不报错，登录态保留。

### Step 2: 只读角色搜索与分页浏览以验证审计数据完整 @web-e2e

**Precondition**: 身份与权限页面已加载；用户/角色/授权列表可能含多页数据，亦可能无匹配结果。

**User Action**: 只读角色在用户、角色、授权列表分别执行搜索、分页、筛选。

**Outcomes**:
- **Outcome A（审计完整性·正向可读）** [F5][F7]：跨用户、角色、授权三个域页面均可读且数据完整；搜索、分页、筛选以只读方式正常执行并返回正确结果；审计角色可据此完成合规审计——正向审计能力得到验证，而非仅断言写被阻断。
- **Outcome B（搜索无匹配·空状态）** [F8]：搜索无匹配时显示空状态，不报错；已加载列表不丢失。
- **Outcome C（搜索网络错误·可重试）**（source: inferred, rule: surface-web.md–network-error）：搜索请求网络错误时显示错误提示与重试入口，不丢失已加载数据；只读语义在错误态下仍成立（无写副作用）。

### Step 3: 只读角色尝试修改用户/角色/授权数据被服务端拒绝 @api-functional

**Precondition**: 只读角色持有令牌（令牌可能有效，亦可能缺失/过期）；后端对所有用户/角色/授权的写操作做服务端权限校验。

**User Action**: 只读角色尝试修改用户/角色/授权数据（创建、编辑、删除）。

**Outcomes**:
- **Outcome A（写操作被拒·403）** [F2][F6]：服务端权限校验拒绝（403），前端提示无权限且数据不变；写操作不生效；403 原地收敛不跳转登录页、不丢失登录态。
- **Outcome B（令牌缺失/过期·401 未认证）**（source: inferred, rule: surface-api.md–unauthorized）：请求未携带有效令牌或令牌已过期时返回 401 未认证；响应体不泄露敏感数据；可重新认证后重试。语义上与 Outcome A 互斥——403 为「已认证但越权」，401 为「未认证」。

### Step 4: 只读角色访问受限管理页被菜单裁剪且 URL 直达被拒 @web-e2e

**Precondition**: 只读角色已登录；租户管理与策略管理对只读角色受限。

**User Action**: 只读角色尝试进入租户管理或策略管理（先看菜单入口；再经浏览器地址栏直接粘贴 URL 访问）。

**Outcomes**:
- **Outcome A（菜单入口不渲染）** [F3]：租户管理与策略管理的菜单入口不渲染；仅可读域入口可见。
- **Outcome B（URL 直达被拒·403·不泄露定义）** [F4][F6]：直接经 URL 访问被服务端权限校验拒绝（403），页面原地提示无权限且数据不变；不跳转登录页；响应体不含租户/策略定义详情（不泄露受限管理内容）。

## Edge Cases

### Step 2b: 搜索输入非法字符被前端即时拦截 @web-e2e

**Precondition**: 搜索框存在格式校验规则（如长度上限、禁用符号）；只读角色待提交非法搜索词。（source: inferred, rule: surface-web.md–validation-error）

**User Action**: 只读角色输入非法搜索词并尝试提交搜索。

**Outcomes**:
- **Outcome A（前端即时拦截·不提交）** [F1]：前端即时拦截并标红提示格式要求，搜索请求不提交至后端；列表数据不变；只读语义保持（无写副作用）。
- **Outcome B（修正后重试·返回结果）** [F7]：用户修正为合法搜索词后可重新搜索，返回正确结果；不报错、不丢失已加载列表。

### Step 1c: 审计浏览中途会话过期 @web-e2e

**Precondition**: 只读角色在审计浏览过程中会话超时过期。（source: inferred, rule: surface-web.md–session-expired）

**User Action**: 只读角色在会话过期后再次操作（搜索/翻页/跳转）。

**Outcomes**:
- **Outcome A（重定向登录·可恢复）** [F6]：重定向至控制台登录页或提示会话过期；登录入口可达；重新登录后可恢复审计上下文。
- **Outcome B（与 403 越权区分·不混淆）** [F6]：过期态不与 403 越权混淆——403 越权保持登录态原地收敛，会话过期则需重新登录。

### Step 1d: 审计角色查看时不触发任何状态变更 @web-e2e

**Precondition**: 只读角色在查看过程中进行搜索、分页、筛选等操作。

**User Action**: 只读角色在只读页面上进行查询/翻页/排序。

**Outcomes**:
- **Outcome A（零写副作用）** [F1]：全程无任何写操作副作用；查询/翻页/排序不触发除审计日志外的状态变更；只读语义在所有交互路径上成立。
- **Outcome B（错误态仍只读·无副作用）** [F7]：即使发生网络错误或空状态，亦不产生任何写副作用；错误态下只读语义仍成立。

## Journey Invariants

- 只读角色永不渲染写入口——增删改入口在所有身份/权限页面均不渲染。[F1]
- 服务端 403 是最终安全边界——任何写接口直调均被后端权限校验拒绝（eiam 对所有写端点做服务端校验），前端裁剪非安全边界。[F2]
- 只读操作零副作用——查看/搜索/分页不触发任何数据变更。[F1]
- 403 原地收敛不跳转登录页——与「会话过期跳登录」语义区分；只读越权不丢失登录态，会话过期才需重新登录。[F6]
- 审计目标可达成——只读角色对用户/角色/授权三域页面均可读且数据完整，不仅断言写被阻断，亦验证正向审计能力。[F5][F7]
