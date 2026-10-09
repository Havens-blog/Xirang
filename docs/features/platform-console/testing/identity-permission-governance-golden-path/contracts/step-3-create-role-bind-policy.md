---
journey: "identity-permission-governance-golden-path"
step: "3"
step-action: "create-role-and-bind-policy"
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

# Contract: identity-permission-governance-golden-path / Step 3: 创建角色并绑定已有策略

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->
<!-- Implementation anchors (endpoint POST /api/role/create -> handler Create; fact IAM_ROLE_CREATE) moved out of dimension values per semantic-purity rule. -->

## Outcome "success"
- Preconditions: "管理员已登录且会话有效；至少一个可绑定策略（policy definition）已存在且对管理员可见；角色名称未重复；操作动作为创建角色并绑定策略"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "PolicyDefinition"
        min_count: 1
        field_constraints:
          - field: "bindable"
            value: "returned by system list"
      - entity_type: "Role"
        min_count: 1
        relationship_type: "binds_policy"
        parent_entity: "PolicyDefinition"
        field_constraints:
          - field: "bound_policy"
            value: "present"
- Input: "管理员在角色管理页创建新角色，从系统提供的可绑定策略列表中选择一个策略绑定到该角色"
- Output: "角色创建成功并带已绑定策略；可绑定策略列表仅含系统返回项，无新建/编辑/删除策略定义的越权入口"
- State: "新角色记录写入后端并关联已绑定策略；策略定义增删改仅平台管理员可用"
- Side-effect: "后端写入新角色并关联已绑定策略；策略定义写入口对租户管理员部分受限（fact: IAM_POLICY_DEFINITION_WRITE_SCOPE PARTIAL）"
- Invariants: "可绑定策略仅含系统返回项，无越权入口（fact: IAM_POLICY_DEFINITION_WRITE_SCOPE）"

## Outcome "dependency-delete-blocked"
<!-- fact: IAM_DEPENDENCY_DELETE_BLOCK (UNKNOWN — 代码无显式阻断计数处理器，按 PRD E8 描述) -->
<!-- post-journey 边缘：本 Outcome 须在授权关系建立后执行，非本步骤线性流内；引用状态在本 journey 中于授权分配后建立 -->
- Preconditions: "操作动作为删除角色或策略（区别于本步骤的创建动作）；待删除对象仍被引用——目标角色仍已分配给至少一个用户，或绑定到该角色的策略仍被该角色引用（此引用状态在本 journey 中于授权分配后建立；本 Outcome 为 post-journey 边缘场景，须在授权关系建立后执行）"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "Role"
        min_count: 1
        relationship_type: "binds_policy"
        parent_entity: "PolicyDefinition"
        field_constraints:
          - field: "assignments"
            value: ">=1 referencing user"
      - entity_type: "PolicyDefinition"
        min_count: 1
        relationship_type: "bound_to_role"
        parent_entity: "Role"
        field_constraints:
          - field: "bound_to_role"
            value: true
- Input: "管理员尝试删除已被引用的角色，或尝试删除绑定到角色的策略"
- Output: "弹窗列出阻断计数（引用该对象的用户数或角色数），确认按钮置灰；对象不被删除；关系解除后方可重试"
- State: "对象未被删除；引用关系不变"
- Side-effect: "none"

## Journey Invariants

- eiam 是唯一后端——所有 CRUD 与授权判定均直调 eiam，控制台零新增后端（fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF）。
- 策略定义增删改仅平台管理员可用；可绑定策略列表仅含系统返回项，无越权入口（fact: IAM_POLICY_DEFINITION_WRITE_SCOPE）。
- 错误态不白屏——校验失败/后端错误均明确提示、保留表单数据、列表数据不变。

## Fixture Specification

This Contract requires the following pre-existing data state. See `rules/fixture-spec.md` for schema details.

```yaml
fixture_spec:
  entities:
    - entity_type: "PlatformAdmin"
      min_count: 1
    - entity_type: "PolicyDefinition"
      min_count: 1
      field_constraints:
        - field: "bindable"
          value: "returned by system list"
    - entity_type: "Role"
      min_count: 1
      relationship_type: "binds_policy"
      parent_entity: "PolicyDefinition"
      field_constraints:
        - field: "bound_policy"
          value: "present"
```
