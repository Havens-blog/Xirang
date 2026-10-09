---
feature: "platform-console"
journey: "login-and-session-lifecycle"
risk_level: "Medium"
golden_path: true
surface_types: ["web", "api"]
surface_keys: ["haven-console", "eiam", "e-cam-web"]
sources:
  - docs/features/platform-console/prd/prd-user-stories.md
  - docs/features/platform-console/prd/prd-spec.md
generated: "2026-10-08"
---

# Journey: login-and-session-lifecycle

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
  Medium = Workflow involves multi-step interaction without irreversible side effects
  Low    = Workflow is read-only or purely observational
-->

## Overview

普通用户从登录控制台、按租户数分流到工作台、单点跳转 e-cam-web，到任一侧登出即双侧会话同步失效——覆盖登录/SSO/登出的完整会话生命周期。本 Journey 的 Happy Path（Step 1 → Step 2 主 Outcome → Step 3 主 Outcome → Step 4 主 Outcome）构成一条连续 4 步的领域级黄金路径，锚定 PRD Story 1（登录进入工作台）与 Story 6（双侧登出对称失效）。

## Setup

- 用户账号已存在且已分配至少一个租户（本 Journey happy path 为单租户场景）
- eiam 服务可用，共享会话存储可读写；e-cam-web 与控制台同登录域
- 浏览器支持 WebAuthn（happy path）；MFA 视账号配置启用
- 控制台为纯前端站，直调 eiam；e-cam-web 会话失效跳转目标已切换为控制台登录页

## Fact Table

| fact_id | Claim | Source | Status |
|---------|-------|--------|--------|
| F1 | 登录页按 eiam 支持方式渲染密码/LDAP/passkey 入口；浏览器不支持 WebAuthn 时 passkey 入口不渲染 | prd-spec E8 | verified |
| F2 | MFA 连续失败 5 次锁定账号 15 分钟，锁定期间凭正确密码亦不可登录 | prd-spec E2/E9 | verified |
| F3 | 登录提交遇到网络错误或 eiam 不可用时保留已输入内容、显示明确错误与重试入口、不白屏 | prd-spec E12 | verified |
| F4 | 认证通过但用户所属租户数 = 0 时进入受限工作台，访问管理页提示「无所属租户」；平台管理员豁免 | prd-spec E3 | verified |
| F5 | 多租户用户经顶部租户切换器重选租户，切换后列表仅显示该租户数据 | prd-spec E5 | verified |
| F6 | 任一侧登出即销毁共享会话，其余应用/标签页在下一次请求时失效并跳回控制台登录页；双侧登出语义对称 | prd-user-stories Story 6 | verified |
| F7 | 认证成功后会话凭据中写入 tenant_id 以驱动租户上下文 | prd-spec DF002 | UNKNOWN — 待 Phase 0 核查写入语义 |

## Happy Path

### Step 1: 进入登录页并选择登录方式

**Test type**: @web-e2e

**Precondition**: 用户浏览器同域访问控制台，无有效会话。

**User Action**: 用户访问控制台，进入登录页；登录页按 eiam 支持方式渲染密码/LDAP/passkey 入口。

**Outcomes**:

- **Outcome 1.1（主路径）**: 登录页正常渲染，无 JS 报错；可用登录方式入口可见且可键盘导航；页面安全响应头存在，页面不可被 iframe 嵌套。*source: fact F1*
- **Outcome 1.2（passkey 不可用）**
  - **Precondition**: 未登录且浏览器不支持 WebAuthn（F1/E8）。
  - **Result**: passkey 入口不渲染，显示「改用密码/LDAP 登录」提示；无 JS 报错；其余登录方式正常可用。*source: fact F1*
- **Outcome 1.3（表单校验失败 — validation-error）**
  - **Precondition**: 用户提交登录表单时必填字段为空或字段格式非法（如用户名含非法字符）。
  - **Result**: 登录页在相关字段就近显示错误提示，表单不提交，用户可修正后重试；不泄露用户名是否存在。**source: inferred** — *required_outcomes rule: web#validation-error*

### Step 2: 提交认证并完成租户分流

**Test type**: @web-e2e, @api-functional

**Precondition**: 用户在登录页选择密码方式，已输入账号凭据；账号所属租户数决定后续分流分支。

**User Action**: 用户输入有效账号密码提交登录；若账号启用 MFA 则完成 MFA 验证；认证成功后系统按所属租户数分流。

**Outcomes**:

- **Outcome 2.1（主路径 — 单租户）**: 认证成功后系统建立共享会话；单租户自动选定唯一租户跳过选择页，直达工作台展示我的信息与应用快捷导航。*source: fact F1, F7（会话凭据写入语义待 Phase 0 核查，见 F7 UNKNOWN）*
- **Outcome 2.2（MFA 锁定）**
  - **Precondition**: 账号已启用 MFA，MFA 连续失败达 5 次（F2/E2/E9）。
  - **Result**: 账号锁定 15 分钟，登录页提示剩余解锁时间；锁定期间凭正确密码亦不可登录；可等待或改用其他已启用方式。*source: fact F2*
- **Outcome 2.3（网络错误 / eiam 不可用）**
  - **Precondition**: 登录提交时网络错误或 eiam 不可用（F3/E12）。
  - **Result**: 已输入内容保留、显示明确错误与重试入口，不白屏；凭据错误文案不泄露用户名是否存在；e-cam-web 已登录会话不受影响。*source: fact F3*
- **Outcome 2.4（零租户受限工作台）**
  - **Precondition**: 认证通过但用户所属租户数 = 0，且用户为普通用户/租户管理员（非平台管理员）（F4/E3）。
  - **Result**: 不进租户选择页，直达工作台受限视图（仅欢迎 + 我的信息）；访问管理页提示「无所属租户」，不报错不白屏；平台管理员豁免本规则。*source: fact F4*
- **Outcome 2.5（多租户切换器重选）**
  - **Precondition**: 用户所属租户数 ≥ 2（F5/E5），已进入租户选择页或工作台。
  - **Result**: 用户经顶部租户切换器重选租户后，当前租户上下文切换并刷新数据；切换后列表仅显示该租户数据。*source: fact F5*

### Step 3: 单点跳转 e-cam-web

**Test type**: @web-e2e, @api-functional

**Precondition**: 用户已在控制台工作台登录态；点击「云管」应用入口发起跨应用跳转。

**User Action**: 已登录用户在工作台点击「云管」应用入口。

**Outcomes**:

- **Outcome 3.1（主路径 — 共享会话生效）**: 单点跳转到 e-cam-web 且不提示重复登录；共享会话生效；e-cam-web 不要求再次认证。*source: fact F6*
- **Outcome 3.2（会话超时 — session-expired）**
  - **Precondition**: 用户在 e-cam-web 工作流中操作时间超过会话有效期，下一次交互时会话已过期（区别于主动登出）。
  - **Result**: 用户被引导回控制台登录页并提示会话已过期；未保存的交互内容被保留或给出数据丢失警示；登录入口可从过期态直接可达。**source: inferred** — *required_outcomes rule: web#session-expired*
- **Outcome 3.3（无凭据访问鉴权端点 — unauthorized）**
  - **Precondition**: 请求方在无有效共享会话凭据的情况下访问 e-cam-web 鉴权端点（会话从未建立或已被另一侧登出销毁）。
  - **Result**: 鉴权端点拒绝请求并返回未认证状态，引导回退到控制台登录页；响应不泄露任何受保护资源数据。**source: inferred** — *required_outcomes rule: api#unauthorized*

### Step 4: 登出触发双侧会话失效

**Test type**: @web-e2e, @api-functional

**Precondition**: 用户双侧会话均处于登录态；登出发起侧可为控制台或 e-cam-web。

**User Action**: 用户在控制台发起登出并确认（主路径），或在 e-cam-web 侧发起登出（反向场景）。

**Outcomes**:

- **Outcome 4.1（主路径 — 控制台登出）**: 控制台登出生效后共享会话立即销毁；已打开的 e-cam-web 会话在下一次请求时被拒绝并跳回控制台登录页；不做实时广播。*source: fact F6*
- **Outcome 4.2（反向 — e-cam-web 侧登出）**
  - **Precondition**: 用户在 e-cam-web 发起登出（F6/Story 6 反向场景）。
  - **Result**: 控制台会话失效并回到登录页；双侧登出语义对称。*source: fact F6*

## Journey Invariants

- 共享会话是双侧会话的唯一真相源——任一侧登出即销毁共享会话，其余应用/标签页在下一次请求时被拒绝并收敛回控制台登录页。
- 登录页安全面不泄露用户名是否存在——凭据错误文案统一。
- 错误态不白屏——网络错误/eiam 不可用均保留输入、明确提示、可重试；静态内容仍可加载。
- 同登录域约束——控制台须与 e-cam-web/eiam 同域，跨域 CDN 直调 eiam 不支持。
- 平台管理员豁免零租户限制——职责跨租户，零所属租户仍可访问全部管理页。
