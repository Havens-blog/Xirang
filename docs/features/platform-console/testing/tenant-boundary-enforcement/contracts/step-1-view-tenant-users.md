---
journey: "tenant-boundary-enforcement"
step: "1"
step-action: "view-in-tenant-user-list"
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

# Contract: tenant-boundary-enforcement / Step 1: 查看本租户用户列表

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "success-scoped"
- Preconditions: "租户管理员已登录并已选定本租户；本租户与其他租户均存在用户数据；会话有效"
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
            value: "valid, tenant selected"
      - entity_type: "Tenant"
        min_count: 2
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "current tenant (visible)"
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "other tenant (must be invisible)"
- Input: "租户管理员在用户管理页查看用户列表"
- Output: "列表仅显示本租户用户，不显示其他租户数据；分页与筛选在本租户范围内生效"
- State: "无状态变更；租户边界由 eiam 服务端强制"
- Side-effect: "eiam 服务端按调用方租户上下文过滤用户列表，跨租户数据不可见"
  <!-- 实现锚点: ctxutil.GetTenantID 过滤用户列表; fact IAM_TENANT_BOUNDARY -->
- Invariants: "租户边界由 eiam 服务端强制（fact: IAM_TENANT_BOUNDARY）"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes session-expired 派生；fact IAM_SESSION_EXPIRED_VS_403_DISTINCT（PARTIAL，跳转区分由前端拦截器）、IAM_UNAUTHORIZED_401 -->
- Preconditions: "操作过程中会话已过期（区别于会话有效的 403 越权与筛选条件校验失败）"
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
- Input: "租户管理员在会话过期后查看用户列表"
- Output: "页面跳转控制台登录页，提示会话已过期；登录流程可从过期状态进入"
- State: "共享会话失效；需重新认证"
- Side-effect: "eiam 鉴权返回 401；前端 HTTP 拦截器跳转登录页（跳转属前端行为，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）"
  <!-- 实现锚点: session.Get 失败 -> 401; 跳转由前端拦截器 -->

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes validation-error 派生；PRD Story 2 AC3 表单/筛选校验口径（无对应 IAM 事实，PRD 级要求） -->
- Preconditions: "用户列表筛选条件含非法值（如页码越界或关键字格式不符），区别于越权与会话过期"
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
- Input: "租户管理员提交含非法筛选条件的列表查询"
- Output: "对应筛选条件标红并回显错误原因，列表数据不变，可修正后重新查询；请求不送达服务端"
- State: "无状态变更；筛选条件保留"
- Side-effect: "none"

## Journey Invariants

- 租户边界由 eiam 服务端强制——前端菜单裁剪仅为体验层；任何越界访问在服务端被拒（fact: IAM_TENANT_BOUNDARY）。
- 403 原地收敛——越权请求保持当前页、不跳转登录页、数据不变；与「会话过期跳登录」严格分流（403 原地 vs 过期跳转；跳转区分由前端拦截器实现，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）。
- 401 与 403 语义分离——无凭据/凭据失效返回 401（认证失败）；有效凭据但越权返回 403（授权失败），二者不混淆（fact: IAM_UNAUTHORIZED_401, IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 策略定义口径——租户管理员仅可绑定已有策略、不可增删改策略定义；该入口在前端不渲染、在服务端被拒（拒绝依赖 grant，fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）。
- 403 不泄露存在性——所有 403 越权响应统一不暴露其他租户资源是否存在（fact: IAM_403_NO_EXISTENCE_LEAK）。
