---
journey: "identity-permission-governance-golden-path"
step: "5"
step-action: "revoke-authorization-and-verify-convergence"
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

# Contract: identity-permission-governance-golden-path / Step 5: 撤销授权并验证权限收敛

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->
<!-- Implementation anchors (endpoint POST /api/role/unassign -> handler UnassignRole; fact IAM_ROLE_REVOKE) moved out of dimension values per semantic-purity rule. -->

## Outcome "success"
- Preconditions: "管理员已登录且会话有效；先前步骤创建的授权关系存在且有效；该用户当前具备该角色对应的权限；操作动作为撤销授权"
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
        field_constraints:
          - field: "assigned_role"
            value: "present and effective"
      - entity_type: "Role"
        min_count: 1
        relationship_type: "binds_policy"
        parent_entity: "PolicyDefinition"
      - entity_type: "AuthorizationRecord"
        min_count: 1
        relationship_type: "links"
        parent_entity: "PlatformUser"
        field_constraints:
          - field: "status"
            value: "active"
- Input: "管理员删除已建立的该用户授权关系，随后该用户以原权限发起下一次请求"
- Output: "授权删除成功；该用户下一次携带原权限的请求被后端拒绝（403）；最终状态收敛为无对应权限"
- State: "授权关系从后端移除；授权策略缓存重载使权限对后续请求收敛为拒绝"
- Side-effect: "后端移除授权关系；权限经缓存重载收敛为拒绝，上限 60 秒（代码实测；PRD ≤30 秒未获代码支持，UNKNOWN — see IAM_AUTH_CACHE）"
- Invariants: "撤销生效延迟上限 60 秒（代码实测；PRD ≤30 秒未获代码支持，UNKNOWN — see IAM_AUTH_CACHE）"

## Outcome "cache-delay-boundary"
<!-- fact: IAM_AUTH_CACHE + IAM_ROLE_REVOKE -->
<!-- 本 Outcome 由 fact 派生（非 required_outcomes 规则），故标记为 factual 而非 inferred -->
- Preconditions: "操作动作为撤销授权后于缓存收敛窗口内发起请求（区别于已越过收敛窗口的成功路径）；删除授权后处于缓存收敛窗口内（≤ 60 秒），缓存尚未失效"
  fixture_spec:
    entities:
      - entity_type: "PlatformUser"
        min_count: 1
        relationship_type: "assigned_to"
        parent_entity: "Role"
        field_constraints:
          - field: "assigned_role"
            value: "just revoked, cache window active"
      - entity_type: "Role"
        min_count: 1
        relationship_type: "binds_policy"
        parent_entity: "PolicyDefinition"
      - entity_type: "AuthorizationRecord"
        min_count: 1
        relationship_type: "links"
        parent_entity: "PlatformUser"
        field_constraints:
          - field: "status"
            value: "revoked, cache window active"
- Input: "该用户在缓存收敛窗口内（≤ 60 秒）发起携带原权限的请求"
- Output: "窗口内可能仍放行（缓存命中）；超过收敛窗口后请求被拒（403）；最终状态一致收敛为无权限"
- State: "授权关系已删除；缓存窗口内权限判定可能暂时放行；窗口外收敛为拒绝"
- Side-effect: "none"
- Invariants: "最终状态一致收敛为无权限（fact: IAM_ROLE_REVOKE）"

## Outcome "pagination-last-page-fallback"
<!-- source: inferred -->
<!-- reasoning: api surface required_outcomes pagination 派生；fact IAM_PAGINATION_LAST_PAGE（UNKNOWN — 后端无回退逻辑，可能前端行为） -->
- Preconditions: "操作动作为删除授权列表末页唯一记录；当前处于授权列表末页且该页仅有一条记录"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "AuthorizationRecord"
        min_count: 1
        relationship_type: "links"
        parent_entity: "PlatformUser"
        field_constraints:
          - field: "page_position"
            value: "last page, sole record"
- Input: "管理员删除授权列表末页的唯一一条授权记录"
- Output: "删除成功后列表回退至上一页，不停留在空页；空页显示空状态；超界页码回落至最后一页"
- State: "该授权记录被删除；当前页回退至上一页或最后一页"
- Side-effect: "none"

## Journey Invariants

- eiam 是唯一后端——所有 CRUD 与授权判定均直调 eiam，控制台零新增后端（fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF）。
- 权限判定逐请求生效——撤销授权无需用户重新登录；撤销生效延迟上限 60 秒（代码实测；PRD ≤30 秒未获代码支持，UNKNOWN — see IAM_AUTH_CACHE）。
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
      field_constraints:
        - field: "assigned_role"
          value: "present and effective"
    - entity_type: "Role"
      min_count: 1
      relationship_type: "binds_policy"
      parent_entity: "PolicyDefinition"
    - entity_type: "AuthorizationRecord"
      min_count: 1
      relationship_type: "links"
      parent_entity: "PlatformUser"
      field_constraints:
        - field: "status"
          value: "active"
```
