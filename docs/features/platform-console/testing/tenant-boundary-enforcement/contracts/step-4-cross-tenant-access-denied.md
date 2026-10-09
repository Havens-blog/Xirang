---
journey: "tenant-boundary-enforcement"
step: "4"
step-action: "cross-tenant-governance-access-attempt"
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

# Contract: tenant-boundary-enforcement / Step 4: 租户管理员访问其他租户治理资源

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "forbidden-403"
- Preconditions: "目标资源归属其他租户；当前会话有效且归属本租户（区别于会话过期与表单校验失败）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "role"
            value: "tenant_admin"
          - field: "session"
            value: "valid, current tenant"
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "other tenant (access target)"
- Input: "租户管理员尝试访问不属于本租户的治理资源（其他租户用户/角色/授权）"
- Output: "eiam 返回 403，页面内提示无权限并保持当前页（不跳转登录页），数据不变；越权提示不暴露其他租户资源是否存在"
- State: "无状态变更；租户边界保持"
- Side-effect: "eiam 服务端拒绝越权请求，返回 403；403 原地收敛（跳转区分由前端拦截器实现，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）"
  <!-- 实现锚点: checker 返回 ErrForbidden; fact IAM_403_NO_EXISTENCE_LEAK -->
- Invariants: "403 不泄露存在性（fact: IAM_403_NO_EXISTENCE_LEAK）"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes session-expired 派生；fact IAM_SESSION_EXPIRED_VS_403_DISTINCT（PARTIAL）、IAM_UNAUTHORIZED_401 -->
- Preconditions: "操作过程中会话已过期（区别于会话有效的 403 越权与表单校验失败）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "role"
            value: "tenant_admin"
          - field: "session"
            value: "expired"
- Input: "租户管理员在会话过期后再次操作"
- Output: "页面跳转控制台登录页，提示会话已过期，未保存的表单数据得到保留或丢失告警；登录流程可从过期状态进入"
- State: "共享会话失效；需重新认证"
- Side-effect: "eiam 鉴权返回 401；前端拦截器跳转登录页（跳转属前端行为，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）"
  <!-- 实现锚点: session.Get 失败 -> 401 -->

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes validation-error 派生；PRD Story 2 AC3 表单校验口径（无对应 IAM 事实，PRD 级要求） -->
- Preconditions: "提交的表单含非法字段（必填项为空或格式不符），区别于越权与会话过期"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "role"
            value: "tenant_admin"
          - field: "session"
            value: "valid"
- Input: "租户管理员提交含非法字段的治理表单"
- Output: "对应字段标红并回显错误原因，表单数据保留，列表数据不变，可修正后重新提交；请求不送达服务端"
- State: "无状态变更；表单数据保留"
- Side-effect: "none"

## Outcome "policy-def-entry-unreachable"
- Preconditions: "操作者为租户管理员，目标为策略定义管理（区别于本租户内治理操作与其他越权场景）"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "role"
            value: "tenant_admin"
          - field: "session"
            value: "valid, current tenant"
      - entity_type: "Policy"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "scope"
            value: "system-scoped policy definition"
- Input: "租户管理员尝试访问策略定义增删改入口，或绕过前端直接以有效凭据请求"
- Output: "策略定义增删改入口不渲染；即使绕过前端以有效凭据直请策略定义写操作，eiam 是否返回 403 取决于操作者是否持有对应 grant（PARTIAL，fact IAM_POLICY_DEFINITION_WRITE_SCOPE）——若被拒则 403 原地收敛、页面内提示无权限并保持当前页、数据不变、不泄露资源存在性；若操作者恰持有 grant 则可能不被拒"
- State: "策略定义不变"
- Side-effect: "eiam 对策略定义写操作按 scope 与 grant 判定；拒绝非硬性 scope 规则（PARTIAL）"
  <!-- 实现锚点: checker 拒绝 ScopeSystem 动作, 但 policy handler 未声明 DefaultScope(ScopeSystem), 拒绝依赖 grant; fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL -->

## Journey Invariants

- 租户边界由 eiam 服务端强制——前端菜单裁剪仅为体验层；任何越界访问在服务端被拒（fact: IAM_TENANT_BOUNDARY）。
- 403 原地收敛——越权请求保持当前页、不跳转登录页、数据不变；与「会话过期跳登录」严格分流（403 原地 vs 过期跳转；跳转区分由前端拦截器实现，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）。
- 401 与 403 语义分离——无凭据/凭据失效返回 401（认证失败）；有效凭据但越权返回 403（授权失败），二者不混淆（fact: IAM_UNAUTHORIZED_401, IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 策略定义口径——租户管理员仅可绑定已有策略、不可增删改策略定义；该入口在前端不渲染、在服务端被拒（拒绝依赖 grant，fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）。
- 403 不泄露存在性——所有 403 越权响应统一不暴露其他租户资源是否存在（fact: IAM_403_NO_EXISTENCE_LEAK）。
