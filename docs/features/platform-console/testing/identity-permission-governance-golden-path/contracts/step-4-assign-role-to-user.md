---
journey: "identity-permission-governance-golden-path"
step: "4"
step-action: "assign-role-to-new-user"
generated: "2026-10-08"
sources:
  - docs/features/platform-console/testing/identity-permission-governance-golden-path/journey.md

anchors:
  api: { endpoint: "", method: "", content_type: "", auth_required: }
  cli: { command: "", subcommand: "", flags: [], aliases: [] }
  tui: { command: "", interactive_prompt: "", keybindings: [] }
  web: { page: "", route: "", requires_auth: , layout: "" }
  mobile: { screen: "", navigation_path: [], deeplink: "", platform: "" }
last_anchor_sync: ""
---

# Contract: identity-permission-governance-golden-path / Step 4: 将角色分配给新用户

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->
<!-- Implementation anchors (endpoint POST /api/role/batch_assign -> handler BatchAssignRole; single /api/role/assign 已注释; fact IAM_ROLE_ASSIGN) moved out of dimension values per semantic-purity rule. -->

## Outcome "success"
- Preconditions: "管理员已登录且会话有效；先前步骤创建的用户存在；先前步骤创建的角色存在且带已绑定策略；操作动作为将角色分配给用户"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "PlatformUser"
        min_count: 1
        relationship_type: "assigned_to"
        parent_entity: "Role"
      - entity_type: "Role"
        min_count: 1
        relationship_type: "binds_policy"
        parent_entity: "PolicyDefinition"
        field_constraints:
          - field: "bound_policy"
            value: "present"
- Input: "管理员将已创建的新角色分配给已创建的新用户，建立授权关系"
- Output: "授权创建成功；该用户下一次请求即具备对应权限（逐请求判定，无需重新登录）；列表在短时间内反映授权变更"
- State: "授权关系写入后端；授权策略缓存重载使新权限对后续请求即时生效"
- Side-effect: "后端写入授权关系；权限经缓存重载近实时生效，上限 60 秒（代码实测；PRD ≤30 秒未获代码支持，UNKNOWN — see IAM_AUTH_CACHE）"
- Invariants: "列表刷新时序 ≤ 2 秒（UNKNOWN — see U1）"

## Outcome "cross-tenant-policy-denied"
<!-- fact: IAM_POLICY_DEFINITION_WRITE_SCOPE (PARTIAL) + IAM_TENANT_BOUNDARY + IAM_403_NO_EXISTENCE_LEAK -->
<!-- 本 Outcome 由 fact 派生（非 required_outcomes 规则），故标记为 factual 而非 inferred -->
- Preconditions: "操作动作为策略定义增删改（区别于本步骤的角色分配动作）；当前操作者为租户管理员（非平台管理员）；策略定义增删改仅平台管理员可用"
  fixture_spec:
    entities:
      - entity_type: "TenantAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "PolicyDefinition"
        min_count: 1
- Input: "租户管理员尝试新建、编辑或删除策略定义"
- Output: "入口不渲染；直接调用接口时后端拒绝请求（403），页面内提示无权限并保持当前页（不跳转登录页），数据不变"
- State: "策略定义不变；租户管理员会话保持有效"
- Side-effect: "后端拒绝越权请求并返回禁止访问错误；不泄露资源存在性（fact: IAM_403_NO_EXISTENCE_LEAK）"
- Invariants: "403 原地收敛不跳转登录页；不泄露资源存在性"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes session-expired 派生；fact IAM_UNAUTHORIZED_401 -->
- Preconditions: "操作动作为提交授权分配（会话已过期）；管理员在授权分配流程中途，会话已超时过期——表单已填写完毕，但提交时会话不再有效"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "expired mid-workflow"
      - entity_type: "PlatformUser"
        min_count: 1
        relationship_type: "assigned_to"
        parent_entity: "Role"
      - entity_type: "Role"
        min_count: 1
- Input: "管理员在角色分配表单填写完毕后提交，此时会话已超时过期"
- Output: "提交后系统检测到会话过期，跳转至登录页面或显示会话过期提示；未保存的授权数据不丢失或明确提示数据丢失风险；登录后可重新进入授权分配流程"
- State: "授权关系未创建；表单数据保留或明确告警数据丢失"
- Side-effect: "后端返回未认证错误（401）；前端引导跳转登录页（fact: IAM_UNAUTHORIZED_401）"

## Outcome "unauthorized-401"
<!-- source: inferred -->
<!-- reasoning: api surface required_outcomes unauthorized 派生；fact IAM_UNAUTHORIZED_401 -->
- Preconditions: "操作动作为调用授权分配接口（无有效凭证）；请求未携带有效认证凭证，或携带的令牌已过期/无效"
  fixture_spec:
    entities:
      - entity_type: "PlatformUser"
        min_count: 1
      - entity_type: "Role"
        min_count: 1
- Input: "直接调用授权分配接口时未携带有效认证凭证（令牌缺失或已过期）"
- Output: "后端拒绝请求并返回未认证错误（401）；响应不泄露敏感数据"
- State: "授权关系未创建；无状态变更"
- Side-effect: "后端拒绝并返回未认证错误；响应不泄露敏感数据（fact: IAM_UNAUTHORIZED_401）"

## Journey Invariants

- eiam 是唯一后端——所有 CRUD 与授权判定均直调 eiam，控制台零新增后端（fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF）。
- 权限判定逐请求生效——分配/撤销授权无需用户重新登录；撤销生效延迟上限 60 秒（代码实测；PRD ≤30 秒未获代码支持，UNKNOWN — see IAM_AUTH_CACHE）。
- 403 原地收敛不跳转登录页——与「会话过期跳登录」严格分流。
- 错误态不白屏——校验失败/后端错误均明确提示、保留表单数据、列表数据不变。

## Fixture Specification

This Contract requires the following pre-existing data state. See `rules/fixture-spec.md` for schema details.

```yaml
fixture_spec:
  entities:
    - entity_type: "PlatformAdmin"
      min_count: 1
    - entity_type: "PlatformUser"
      min_count: 1
      relationship_type: "assigned_to"
      parent_entity: "Role"
    - entity_type: "Role"
      min_count: 1
      relationship_type: "binds_policy"
      parent_entity: "PolicyDefinition"
      field_constraints:
        - field: "bound_policy"
          value: "present"
```
