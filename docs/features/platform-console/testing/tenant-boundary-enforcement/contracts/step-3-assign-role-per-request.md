---
journey: "tenant-boundary-enforcement"
step: "3"
step-action: "assign-role-and-verify-per-request"
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

# Contract: tenant-boundary-enforcement / Step 3: 将角色分配给本租户用户并验证逐请求生效

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "success"
- Preconditions: "Step 2 创建的角色已存在于本租户；待分配用户属于本租户；租户管理员会话有效"
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
      - entity_type: "Role"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "current tenant"
      - entity_type: "User"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "current tenant (assignable target)"
- Input: "租户管理员将该角色分配给本租户用户，随后该用户发起一次需该权限的请求"
- Output: "分配成功；由于授权策略存在收敛窗口（上限 60s），该用户后续请求在策略重载后生效；逐请求判定、无需重新登录。测试须采用等待或重试策略（轮询直至权限生效或等待超过收敛上限）以避免时序 flaky"
- State: "授权关系写入本租户；授权策略重载使权限对后续请求生效"
- Side-effect: "eiam 在本租户内写入授权关系；策略重载近实时生效，存在 60s 收敛上限"
  <!-- 实现锚点: POST /api/role/batch_assign; casbin StartAutoLoadPolicy(time.Minute)+Redis watcher; fact IAM_ROLE_ASSIGN, IAM_AUTH_CACHE -->
- Invariants: "逐请求判定无需重新登录（fact: IAM_ROLE_ASSIGN, IAM_AUTH_CACHE）"

## Outcome "assign-denied-403"
<!-- source: inferred -->
<!-- reasoning: web+api 双 surface 越权分配派生；fact IAM_403_NO_EXISTENCE_LEAK、IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL -->
- Preconditions: "租户管理员尝试将角色分配给不属于本租户的用户，或分配自身无权授予的角色（区别于成功分配与会话过期）"
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
            value: "other tenant (assignment target)"
- Input: "租户管理员尝试越权分配角色（目标用户归属其他租户）"
- Output: "区分两面——API 面：eiam 返回 403 Forbidden，响应体为通用越权提示、不泄露目标资源是否存在；Web 面：页面内提示无权限并保持当前页（不跳转登录页），表单数据不变"
- State: "无状态变更；授权关系未写入"
- Side-effect: "eiam 服务端拒绝越权分配，返回 403 原地收敛；跳转区分由前端拦截器实现（fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）"
  <!-- 实现锚点: checker 返回 ErrForbidden; fact IAM_403_NO_EXISTENCE_LEAK -->

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes session-expired 派生；fact IAM_SESSION_EXPIRED_VS_403_DISTINCT（PARTIAL）、IAM_UNAUTHORIZED_401 -->
- Preconditions: "操作过程中会话已过期（区别于会话有效的成功分配与越权分配）"
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
- Input: "租户管理员在会话过期后提交角色分配"
- Output: "页面跳转控制台登录页，提示会话已过期，未保存的表单数据得到保留或丢失告警；登录流程可从过期状态进入"
- State: "共享会话失效；需重新认证；授权关系未写入"
- Side-effect: "eiam 鉴权返回 401；前端拦截器跳转登录页（跳转属前端行为，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）"
  <!-- 实现锚点: session.Get 失败 -> 401 -->

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes validation-error 派生；PRD Story 2 AC3 表单校验口径（无对应 IAM 事实，PRD 级要求） -->
- Preconditions: "提交的角色分配表单含非法字段（未选择角色或目标用户），区别于越权与会话过期"
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
- Input: "租户管理员提交含非法字段的角色分配表单"
- Output: "对应字段标红并回显错误原因，表单数据保留，列表数据不变，可修正后重新提交；请求不送达服务端"
- State: "无状态变更；表单数据保留"
- Side-effect: "none"

## Journey Invariants

- 租户边界由 eiam 服务端强制——前端菜单裁剪仅为体验层；任何越界访问在服务端被拒（fact: IAM_TENANT_BOUNDARY）。
- 403 原地收敛——越权请求保持当前页、不跳转登录页、数据不变；与「会话过期跳登录」严格分流（403 原地 vs 过期跳转；跳转区分由前端拦截器实现，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）。
- 401 与 403 语义分离——无凭据/凭据失效返回 401（认证失败）；有效凭据但越权返回 403（授权失败），二者不混淆（fact: IAM_UNAUTHORIZED_401, IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 策略定义口径——租户管理员仅可绑定已有策略、不可增删改策略定义；该入口在前端不渲染、在服务端被拒（拒绝依赖 grant，fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）。
- 403 不泄露存在性——所有 403 越权响应统一不暴露其他租户资源是否存在（fact: IAM_403_NO_EXISTENCE_LEAK）。
