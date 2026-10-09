---
journey: "tenant-boundary-enforcement"
step: "5"
step-action: "http-direct-governance-api-access"
generated: "2026-10-08"
sources:
  - docs/features/platform-console/testing/tenant-boundary-enforcement/journey.md

anchors:
  api: { endpoint: "", method: "", content_type: "", auth_required: }
  cli: { command: "", subcommand: "", flags: [], aliases: [] }
  tui: { command: "", interactive_prompt: "", keybindings: [] }
  web: { page: "", route: "", requires_auth: , layout: "" }
  mobile: { screen: "", navigation_path: [], deeplink: "", platform: "" }
last_anchor_sync: ""
---

# Contract: tenant-boundary-enforcement / Step 5: 以 HTTP 客户端直接访问治理接口

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "unauthorized-401"
<!-- source: inferred -->
<!-- reasoning: api surface required_outcomes unauthorized 派生；fact IAM_UNAUTHORIZED_401 -->
- Preconditions: "请求未携带有效凭据，或携带的凭据已失效（区别于有效凭据的各场景）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "credentials"
            value: "absent or invalid"
- Input: "调用方以 HTTP 请求访问 eiam 治理接口，未携带有效凭据"
- Output: "返回 401 Unauthorized，响应体含认证错误信息，不泄露任何租户数据；不与 403 越权混淆"
- State: "无状态变更；请求被拒"
- Side-effect: "eiam 鉴权中间件返回 401（fact IAM_UNAUTHORIZED_401）"
  <!-- 实现锚点: session.Get 失败或返回 nil -> AbortWithStatus(401) -->

## Outcome "cross-tenant-403"
- Preconditions: "凭据有效且归属本租户，但请求目标属于其他租户（区别于本租户内有效访问与策略定义写）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "credentials"
            value: "valid, current tenant"
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "other tenant (access target)"
- Input: "调用方以有效凭据访问归属其他租户的治理资源"
- Output: "返回 403 Forbidden，响应不泄露其他租户资源是否存在，数据不变"
- State: "无状态变更；租户边界保持"
- Side-effect: "eiam 服务端拒绝跨租户请求，返回 403"
  <!-- 实现锚点: checker 返回 ErrForbidden; fact IAM_403_NO_EXISTENCE_LEAK -->
- Invariants: "403 不泄露存在性（fact: IAM_403_NO_EXISTENCE_LEAK）"

## Outcome "policy-def-write-403"
- Preconditions: "凭据有效但操作者为租户管理员，目标为策略定义写操作（区别于跨租户访问与本租户内有效访问）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "credentials"
            value: "valid, tenant admin"
      - entity_type: "Policy"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "scope"
            value: "system-scoped policy definition"
- Input: "调用方以租户管理员凭据尝试增删改策略定义"
- Output: "eiam 是否返回 403 取决于操作者是否持有对应 grant（PARTIAL，fact IAM_POLICY_DEFINITION_WRITE_SCOPE）；若被拒则返回 403 Forbidden、不泄露资源存在性、数据不变；若操作者恰持有 grant 则可能不被拒"
- State: "策略定义不变"
- Side-effect: "eiam 对策略定义写操作按 scope 与 grant 判定；拒绝非硬性 scope 规则（PARTIAL）"
  <!-- 实现锚点: checker 拒绝 ScopeSystem 动作, 但 policy handler 未声明 DefaultScope(ScopeSystem), 拒绝依赖 grant; fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL -->

## Outcome "in-tenant-read-200"
- Preconditions: "凭据有效，请求目标属于本租户且为读操作、在租户管理员权限内（区别于越权与未认证）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "credentials"
            value: "valid, current tenant"
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "current tenant (read target)"
- Input: "调用方以有效凭据对本租户治理资源发起读操作（代表操作：查询本租户用户列表）"
- Output: "返回 200，响应体仅含本租户数据，不含其他租户数据"
- State: "无状态变更"
- Side-effect: "eiam 按租户上下文返回本租户数据"
  <!-- 实现锚点: ctxutil.GetTenantID 过滤; fact IAM_TENANT_BOUNDARY -->

## Outcome "in-tenant-write-201"
- Preconditions: "凭据有效，请求目标属于本租户且为写操作、在租户管理员权限内（区别于越权与未认证）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "credentials"
            value: "valid, current tenant"
      - entity_type: "Role"
        min_count: 0
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "lifecycle"
            value: "created during operation (current tenant)"
- Input: "调用方以有效凭据对本租户治理资源发起写操作（代表操作：在本租户内创建角色）"
- Output: "返回 201，新资源写入本租户，不跨租户"
- State: "新资源写入本租户"
- Side-effect: "eiam 在本租户内创建资源；策略定义写入口对租户管理员部分受限"
  <!-- 实现锚点: POST /api/role/create; fact IAM_ROLE_CREATE, IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL -->

## Journey Invariants

- 租户边界由 eiam 服务端强制——前端菜单裁剪仅为体验层；任何越界访问在服务端被拒（fact: IAM_TENANT_BOUNDARY）。
- 403 原地收敛——越权请求保持当前页、不跳转登录页、数据不变；与「会话过期跳登录」严格分流（403 原地 vs 过期跳转；跳转区分由前端拦截器实现，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）。
- 401 与 403 语义分离——无凭据/凭据失效返回 401（认证失败）；有效凭据但越权返回 403（授权失败），二者不混淆（fact: IAM_UNAUTHORIZED_401, IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 策略定义口径——租户管理员仅可绑定已有策略、不可增删改策略定义；该入口在前端不渲染、在服务端被拒（拒绝依赖 grant，fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）。
- 403 不泄露存在性——所有 403 越权响应统一不暴露其他租户资源是否存在（fact: IAM_403_NO_EXISTENCE_LEAK）。
