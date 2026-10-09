---
journey: "tenant-boundary-enforcement"
step: "2"
step-action: "create-in-tenant-role-and-bind-policy"
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

# Contract: tenant-boundary-enforcement / Step 2: 在本租户创建角色并绑定已有策略

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "success"
- Preconditions: "租户管理员在本租户内；可绑定策略列表由 eiam 返回；会话有效"
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
      - entity_type: "Policy"
        min_count: 1
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "bindable"
            value: "returned by eiam list"
- Input: "租户管理员在本租户内创建一个新角色，从可绑定策略列表中选择一个策略绑定"
- Output: "绑定成功；可绑定策略仅含 eiam 返回的列表，无新建/编辑/删除策略定义入口"
- State: "新角色写入本租户；角色关联已绑定策略"
- Side-effect: "eiam 在本租户内创建角色并绑定策略；策略定义写入口对租户管理员部分受限"
  <!-- 实现锚点: POST /api/role/create; fact IAM_ROLE_CREATE, IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL -->
- Invariants: "可绑定策略仅含 eiam 返回列表（fact: IAM_ROLE_CREATE, IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes session-expired 派生；fact IAM_SESSION_EXPIRED_VS_403_DISTINCT（PARTIAL）、IAM_UNAUTHORIZED_401 -->
- Preconditions: "操作过程中会话已过期（区别于会话有效的创建成功与表单校验失败）"
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
- Input: "租户管理员在会话过期后提交角色创建表单"
- Output: "页面跳转控制台登录页，提示会话已过期，未保存的表单数据得到保留或丢失告警；登录流程可从过期状态进入"
- State: "共享会话失效；需重新认证；角色未创建"
- Side-effect: "eiam 鉴权返回 401；前端拦截器跳转登录页（跳转属前端行为，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）"
  <!-- 实现锚点: session.Get 失败 -> 401 -->

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes validation-error 派生；PRD Story 2 AC3 表单校验口径（无对应 IAM 事实，PRD 级要求） -->
- Preconditions: "提交的角色创建表单含非法字段（必填项为空或角色名格式不符），区别于越权与会话过期"
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
- Input: "租户管理员提交含非法字段的角色创建表单"
- Output: "对应字段标红并回显错误原因，表单数据保留，列表数据不变，可修正后重新提交；请求不送达服务端"
- State: "无状态变更；表单数据保留"
- Side-effect: "none"

## Journey Invariants

- 租户边界由 eiam 服务端强制——前端菜单裁剪仅为体验层；任何越界访问在服务端被拒（fact: IAM_TENANT_BOUNDARY）。
- 403 原地收敛——越权请求保持当前页、不跳转登录页、数据不变；与「会话过期跳登录」严格分流（403 原地 vs 过期跳转；跳转区分由前端拦截器实现，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）。
- 401 与 403 语义分离——无凭据/凭据失效返回 401（认证失败）；有效凭据但越权返回 403（授权失败），二者不混淆（fact: IAM_UNAUTHORIZED_401, IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 策略定义口径——租户管理员仅可绑定已有策略、不可增删改策略定义；该入口在前端不渲染、在服务端被拒（拒绝依赖 grant，fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）。
- 403 不泄露存在性——所有 403 越权响应统一不暴露其他租户资源是否存在（fact: IAM_403_NO_EXISTENCE_LEAK）。
