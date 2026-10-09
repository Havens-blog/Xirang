---
feature: "platform-console"
journey: "identity-permission-governance-golden-path"
risk_level: "High"
golden_path: true
surface_types: ["web", "api"]
surface_keys: ["haven-console", "eiam"]
sources:
  - docs/features/platform-console/prd/prd-user-stories.md
  - docs/features/platform-console/prd/prd-spec.md
generated: "2026-10-08"
---

# Journey: identity-permission-governance-golden-path

**Risk Level**: High

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
  Medium = Workflow involves multi-step interaction without irreversible side effects
  Low    = Workflow is read-only or purely observational
-->

## Overview

平台管理员对平台身份与权限进行端到端治理：创建平台用户、配置身份源、创建角色并绑定已有策略、将角色分配给用户、最后撤销授权并验证权限即时收敛——覆盖「身份治理 + 权限管理」的核心域动作序列。

## Fact Table

| fact_id | Claim | Source | Status |
|---------|-------|--------|--------|
| F1 | 平台管理员可创建平台用户并重置初始密码，初始密码仅展示一次，首次登录强制改密 | PRD Story 2 | verified |
| F2 | 用户名格式要求：小写字母开头、仅含小写字母/数字/下划线/连字符、长度 3–32 字符 | PRD spec | verified |
| F3 | LDAP 身份源配置保存后立即可用于登录，无需重启控制台 | PRD E3 | verified |
| F4 | 平台管理员可创建角色并绑定已有策略 | PRD Story 4 | verified |
| F5 | 策略定义增删改仅平台管理员可用；可绑定策略列表仅含系统返回项，无越权入口 | PRD spec | verified |
| F6 | 角色分配后逐请求生效，无需重新登录 | PRD spec | verified |
| F7 | 撤销授权后权限收敛，后续请求被拒 | PRD spec | verified |
| F8 | 依赖对象删除被拒，弹窗列出阻断计数，确认按钮置灰 | PRD E11 | verified |
| F9 | 并发编辑同一用户对象触发版本冲突，后提交者须刷新后重试 | PRD E6 | verified |
| F10 | 分页末页唯一记录删除后列表回退至上一页，不停留空页 | PRD E7 | verified |
| F11 | 身份源配置保存失败不影响已登录会话 | PRD E12 | verified |
| F12 | 租户管理员尝试增删改策略定义被拒——入口不渲染且后端拒绝 | PRD spec | verified |
| U1 | 列表刷新时序 ≤ 2 秒 | — | UNKNOWN (Phase 0 核查) |
| U2 | 授权缓存生效延迟 ≤ 30 秒 | — | UNKNOWN (Phase 0 核查) |
| U3 | 后端是否支持 409 版本冲突响应 | — | UNKNOWN (Phase 0 核查) |
| U4 | 配置 LDAP 后登录页渲染对应身份源入口 | — | UNKNOWN (未在 PRD 明确，由 LDAP 配置流程合理推断) |

## Surface Configuration

| Surface | Key | Mandatory Required Outcomes (per surface rules) |
|---------|-----|-------------------------------------------------|
| web | haven-console | `validation-error`, `session-expired` |
| api | eiam | `unauthorized`（未认证 401） |

**Additional boundary Outcomes referenced**: web — `network-error`, `concurrent-edit`; api — `conflict`, `pagination`, `internal-error`, `validation-error`。

## Setup

- 平台管理员账号已存在且已启用 MFA，已登录控制台（平台管理员豁免零租户限制，可访问全部管理页）。**会话有效性是本 Journey 全程隐含前置条件——Step 4c 将对中途会话过期进行压力测试。**
- eiam 服务可用，eiam API 契约基线已冻结（parity-checklist.md 逐项签收）
- 至少一个可绑定策略（policy definition）已存在且对租户管理员可见
- 控制台直调 eiam，无自建后端；前端菜单裁剪仅为体验层、非安全边界

## Happy Path

### Step 1: 创建平台用户并重置初始密码

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员在用户管理页新增一个平台用户（用户名须以小写字母开头、仅含小写字母、数字、下划线与连字符、长度 3 至 32 字符），提交后对该用户执行密码重置以生成一次性初始密码。

**Outcomes**:

#### Outcome 1-a: 用户创建成功并刷新列表（happy path）

- **Precondition**: 管理员已登录且会话有效；输入的用户名符合格式要求且系统中不存在同名用户。
- **Expected**: 操作完成后用户列表在短时间内刷新并显示该新用户；初始密码仅展示一次；该用户首次登录将被强制修改密码。
- **Annotations**: fact: F1, F2; test_level: journey; refresh_target: ≤ 2 秒 (UNKNOWN — see U1)

## Edge Cases

### Step 1b: 用户名格式不合规被前端即时拦截

**Test types**: `@web-e2e`

**User Action**: 平台管理员在用户名输入框输入不符合格式要求的值（如首字符非小写字母、长度越界、含非法字符）并尝试提交。

**Outcomes**:

#### Outcome 1b-a: 前端即时拦截格式错误（derived）

- **Precondition**: 输入的用户名不符合格式要求——首字符非小写字母，或长度不在 3 至 32 字符范围内，或含非法字符。
- **Expected**: 前端即时拦截并标红提示格式要求，表单不提交至后端；列表数据不变；用户名唯一性最终由后端校验。
- **Annotations**: `source: inferred`; rule: `web/required_outcomes/validation-error`; fact: F2; test_level: contract

### Step 1c: 用户名重复或后端返回错误

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员提交新增用户请求。

**Outcomes**:

#### Outcome 1c-a: 重复或错误被拒绝并保留表单（derived）

- **Precondition**: 提交的用户名已存在，或后端返回校验/写入错误。
- **Expected**: 对应字段标红并回显错误原因，表单数据保留，列表数据不变；修正后可重新提交；页面不白屏。
- **Annotations**: `source: inferred`; rule: `api/required_outcomes/conflict`, `web/required_outcomes/validation-error`; fact: F2; test_level: contract; http_status: 409 (UNKNOWN — see U3)

### Step 1d: 并发编辑同一用户对象（版本冲突）

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 两个管理员会话同时编辑同一用户对象，后提交者基于过期版本提交编辑。

**Outcomes**:

#### Outcome 1d-a: 版本冲突被检测，要求刷新后重试（derived）

- **Precondition**: 两个管理员会话同时编辑同一用户对象；先提交者的改动已生效，后提交者基于过期版本提交。
- **Expected**: 后提交者收到版本冲突提示，要求刷新后重试，不覆盖他人改动。
- **Annotations**: `source: inferred`; rule: `web/required_outcomes/concurrent-edit`, `api/required_outcomes/conflict`; fact: F9; test_level: contract; http_status: 409 (UNKNOWN — see U3)

### Step 2: 配置身份源（LDAP）

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员在身份源管理页配置一个 LDAP 身份源（填写绑定地址、搜索基准、凭据），保存配置。

**Outcomes**:

#### Outcome 2-a: LDAP 配置保存成功并即时生效（happy path）

- **Precondition**: 管理员已登录且会话有效；LDAP 服务参数完整且格式正确。
- **Expected**: 配置保存后，该身份源立即可用于登录；不重启控制台即生效。
- **Annotations**: fact: F3; test_level: journey; note: 登录页是否渲染对应入口 (UNKNOWN — see U4)

### Step 2b: 身份源配置保存失败

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员保存 LDAP 配置。

**Outcomes**:

#### Outcome 2b-a: 保存失败提示且不影响已登录会话（derived）

- **Precondition**: 后端不可用，或身份源配置字段校验失败。
- **Expected**: 明确错误提示与可重试入口，静态内容仍渲染；已有登录方式不受影响；已登录会话不受影响。
- **Annotations**: `source: inferred`; rule: `web/required_outcomes/network-error`, `api/required_outcomes/internal-error`; fact: F11; test_level: contract

### Step 3: 创建角色并绑定已有策略

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员在角色管理页创建一个新角色，从系统提供的可绑定策略列表中选择一个策略绑定到该角色。

**Outcomes**:

#### Outcome 3-a: 角色创建成功并带已绑定策略（happy path）

- **Precondition**: 管理员已登录且会话有效；至少一个可绑定策略已存在且对管理员可见；角色名称未重复。
- **Expected**: 角色创建成功并带已绑定策略；可绑定策略列表仅含系统返回项，无新建/编辑/删除策略定义的越权入口（策略定义增删改仅平台管理员可用）。
- **Annotations**: fact: F4, F5; test_level: journey

### Step 3b: 依赖对象删除被拒（阻断计数提示）

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员尝试删除 Step 3 创建的角色（该角色已在 Step 4 中分配给用户），或尝试删除绑定到该角色的策略。

**Outcomes**:

#### Outcome 3b-a: 删除被阻断并提示引用计数（derived）

- **Precondition**: 待删除对象仍被引用——Step 3 创建的角色仍已分配给至少一个用户（Step 4 已建立授权关系），或绑定到该角色的策略仍被该角色引用。
- **Expected**: 弹窗列出阻断计数（引用该对象的用户数或角色数），确认按钮置灰；对象不被删除；关系解除后方可重试。
- **Annotations**: fact: F8; test_level: contract

### Step 4: 将角色分配给新用户

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员将 Step 3 创建的新角色分配给 Step 1 创建的新用户，建立授权关系。

**Outcomes**:

#### Outcome 4-a: 授权创建成功并即时生效（happy path）

- **Precondition**: 管理员已登录且会话有效；Step 1 创建的用户存在；Step 3 创建的角色存在且带已绑定策略。
- **Expected**: 授权创建成功；该用户下一次请求即具备对应权限（逐请求判定，无需重新登录）；列表在短时间内反映授权变更。
- **Annotations**: fact: F6; test_level: journey; refresh_target: ≤ 2 秒 (UNKNOWN — see U1)

### Step 4b: 租户管理员尝试创建策略定义被拒

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 租户管理员尝试新建、编辑或删除策略定义。

**Outcomes**:

#### Outcome 4b-a: 越权操作被拒且不跳转登录页（derived）

- **Precondition**: 当前操作者为租户管理员（非平台管理员）；策略定义增删改仅平台管理员可用。
- **Expected**: 入口不渲染；直接调用接口时后端拒绝请求，页面内提示无权限并保持当前页（不跳转登录页），数据不变。
- **Annotations**: fact: F12; test_level: contract; http_status: 403

### Step 4c: 提交授权时会话过期（mid-workflow session-expired）

**Test types**: `@web-e2e`

**User Action**: 平台管理员在角色分配表单填写完毕后提交，此时会话已超时过期。

**Outcomes**:

#### Outcome 4c-a: 会话过期被检测，引导重新登录（derived）

- **Precondition**: 管理员在 Step 4 的授权分配流程中途，会话已超时过期——表单已填写完毕，但提交时会话不再有效。
- **Expected**: 提交后系统检测到会话过期，跳转至登录页面或显示会话过期提示；未保存的授权数据不丢失或明确提示数据丢失风险；登录后可重新进入授权分配流程。
- **Annotations**: `source: inferred`; rule: `web/required_outcomes/session-expired`; test_level: contract

### Step 4d: API 请求未携带有效凭证被拒（unauthorized 401）

**Test types**: `@api-functional`

**User Action**: 直接调用授权分配接口时未携带有效认证凭证（如令牌缺失或已过期）。

**Outcomes**:

#### Outcome 4d-a: 未认证请求被拒且不泄露敏感数据（derived）

- **Precondition**: 请求未携带有效认证凭证，或携带的令牌已过期/无效。
- **Expected**: 后端拒绝请求并返回未认证错误；响应不泄露敏感数据。
- **Annotations**: `source: inferred`; rule: `api/required_outcomes/unauthorized`; test_level: contract; http_status: 401

### Step 5: 撤销授权并验证权限收敛

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员删除 Step 4 中建立的该用户授权关系，随后该用户以原权限发起下一次请求。

**Outcomes**:

#### Outcome 5-a: 授权删除成功且权限即时收敛（happy path）

- **Precondition**: 管理员已登录且会话有效；Step 4 创建的授权关系存在且有效；该用户当前具备该角色对应的权限。
- **Expected**: 授权删除成功；该用户下一次携带原权限的请求被后端拒绝；最终状态收敛为无对应权限。
- **Annotations**: fact: F7; test_level: journey; http_status: 403; cache_delay: ≤ 30 秒 (UNKNOWN — see U2)

### Step 5b: 授权缓存延迟边界

**Test types**: `@api-functional`

**User Action**: 该用户在缓存窗口内发起携带原权限的请求。

**Outcomes**:

#### Outcome 5b-a: 缓存窗口内可能放行，最终收敛为无权限（derived）

- **Precondition**: 后端存在授权缓存，删除授权后缓存尚未失效（短时间窗口内）。
- **Expected**: 窗口内可能仍放行（缓存命中）；超过短时间窗口后请求被拒；最终状态一致收敛为无权限。
- **Annotations**: fact: F7; test_level: contract; http_status: 403; cache_delay: ≤ 30 秒 (UNKNOWN — see U2)

### Step 5c: 分页边界——删除末页唯一记录后回退

**Test types**: `@web-e2e` `@api-functional`

**User Action**: 平台管理员删除授权列表末页的唯一一条授权记录。

**Outcomes**:

#### Outcome 5c-a: 删除后列表回退至上一页（derived）

- **Precondition**: 当前处于授权列表末页且该页仅有一条记录。
- **Expected**: 删除成功后列表回退至上一页，不停留在空页；空页显示空状态；超界页码回落至最后一页。
- **Annotations**: `source: inferred`; rule: `api/required_outcomes/pagination`; fact: F10; test_level: contract

## Journey Invariants

- **eiam 是唯一后端**——所有 CRUD 与授权判定均直调 eiam，控制台零新增后端。
- **权限判定逐请求生效**——分配/撤销授权无需用户重新登录；撤销生效延迟 ≤ 30 秒（UNKNOWN — see U2, Phase 0 核查）。
- **前端菜单裁剪仅为体验层、非安全边界**——eiam 对所有 CRUD 端点做服务端权限校验；直接调用写接口必被拒绝（http_status: 403）。
- **列表刷新时序**——CRUD 操作成功返回后列表 ≤ 2 秒内反映变更（UNKNOWN — see U1, Phase 0 核查）。
- **错误态不白屏**——校验失败/后端错误均明确提示、保留表单数据、列表数据不变。
