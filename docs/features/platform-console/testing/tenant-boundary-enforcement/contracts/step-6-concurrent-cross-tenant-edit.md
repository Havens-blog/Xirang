---
journey: "tenant-boundary-enforcement"
step: "6"
step-action: "concurrent-cross-tenant-edit-isolated"
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

# Contract: tenant-boundary-enforcement / Step 6: 并发跨租户编辑互不干扰

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "success-isolated"
- Preconditions: "两个不同租户的管理员同时编辑各自租户内的同一类型对象；双方会话均有效且归属各自租户"
  fixture_spec:
    entities:
      - entity_type: "User"
        min_count: 2
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "role"
            value: "tenant_admin"
          - field: "tenant"
            value: "two distinct tenants"
          - field: "session"
            value: "valid, each in own tenant"
      - entity_type: "Role"
        min_count: 2
        relationship_type: "belongs_to"
        parent_entity: "Tenant"
        field_constraints:
          - field: "tenant"
            value: "one editable object per admin's tenant"
- Input: "双方各自提交编辑（以有效凭据直调 eiam 治理端点编辑本租户内角色）"
- Output: "各租户内后提交者各自生效；不发生跨租户数据串改；同租户内同一对象并发编辑的 409 版本冲突响应为 UNKNOWN（eiam 无 409 代码路径，fact IAM_VERSION_CONFLICT_409 = UNKNOWN）——若实现存在 409 则提示刷新后重试，否则后提交者覆盖先生效。租户边界在并发下仍成立"
- State: "各租户对象按各自提交生效；跨租户隔离保持"
- Side-effect: "eiam 按各自租户上下文隔离并发编辑；409 版本冲突响应 UNKNOWN（fact IAM_VERSION_CONFLICT_409）"
  <!-- 实现锚点: ctxutil.GetTenantID 隔离租户上下文; 无 409/乐观锁代码路径; fact IAM_TENANT_BOUNDARY, IAM_VERSION_CONFLICT_409 UNKNOWN -->
- Invariants: "租户边界在并发下仍成立（fact: IAM_TENANT_BOUNDARY）"

## Journey Invariants

- 租户边界由 eiam 服务端强制——前端菜单裁剪仅为体验层；任何越界访问在服务端被拒（fact: IAM_TENANT_BOUNDARY）。
- 403 原地收敛——越权请求保持当前页、不跳转登录页、数据不变；与「会话过期跳登录」严格分流（403 原地 vs 过期跳转；跳转区分由前端拦截器实现，fact IAM_SESSION_EXPIRED_VS_403_DISTINCT PARTIAL）。
- 401 与 403 语义分离——无凭据/凭据失效返回 401（认证失败）；有效凭据但越权返回 403（授权失败），二者不混淆（fact: IAM_UNAUTHORIZED_401, IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 策略定义口径——租户管理员仅可绑定已有策略、不可增删改策略定义；该入口在前端不渲染、在服务端被拒（拒绝依赖 grant，fact IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）。
- 403 不泄露存在性——所有 403 越权响应统一不暴露其他租户资源是否存在（fact: IAM_403_NO_EXISTENCE_LEAK）。
