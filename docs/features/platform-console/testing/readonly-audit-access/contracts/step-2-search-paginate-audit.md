---
journey: "readonly-audit-access"
step: "2"
step-action: "search-and-paginate-audit-data"
generated: "2026-10-08"
sources:
  - docs/features/platform-console/testing/readonly-audit-access/journey.md

anchors:
  api: { endpoint: "", method: "", content_type: "", auth_required: null }
  cli: { command: "", subcommand: "", flags: [], aliases: [] }
  tui: { command: "", interactive_prompt: "", keybindings: [] }
  web: { page: "", route: "", requires_auth: null, layout: "" }
  mobile: { screen: "", navigation_path: [], deeplink: "", platform: "" }
last_anchor_sync: ""
---

# Contract: readonly-audit-access / Step 2: 只读角色搜索与分页浏览以验证审计数据完整

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "audit-integrity"
- Preconditions: "身份与权限页面已加载；用户/角色/授权列表含可读数据（含多页以支撑分页审计）；会话有效"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
      - entity_type: "PlatformUser"
        min_count: 25
        relationships:
          - relationship_type: "has_role"
            target_entity: "Role"
      - entity_type: "Role"
        min_count: 25
        relationships:
          - relationship_type: "assigned_to"
            target_entity: "PlatformUser"
      - entity_type: "Authorization"
        min_count: 25
        relationships:
          - relationship_type: "binds"
            parent_entity: "Role"
            target_entity: "PlatformUser"
- Input: "只读角色在用户、角色、授权列表分别执行搜索、分页、筛选"
- Output: "跨用户、角色、授权三个域页面均可读且数据完整；搜索、分页、筛选以只读方式正常执行并返回正确结果；审计角色可据此完成合规审计"
- State: "无状态变更"
- Side-effect: "none"
- Invariants: "正向审计能力得到验证（fact: F5, F7）"

## Outcome "search-empty-state"
- Preconditions: "搜索无匹配结果（区别于含匹配数据的正向路径），会话有效"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
- Input: "只读角色输入搜索词，无匹配结果"
- Output: "搜索无匹配时显示空状态，不报错；已加载列表不丢失"
- State: "无状态变更；空状态占位显示"
- Side-effect: "none"
- Invariants: "空状态不报错（fact: F8）"

## Outcome "network-error"
<!-- source: inferred -->
<!-- reasoning: surface-web.md–network-error 派生；journey Step 2 Outcome C -->
- Preconditions: "搜索请求网络错误（区别于无匹配与正常路径），会话有效"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
      - entity_type: "EiamService"
        min_count: 1
        field_constraints:
          - field: "reachable"
            value: false
- Input: "只读角色发起搜索请求，此时网络错误"
- Output: "搜索请求网络错误时显示错误提示与重试入口，不丢失已加载数据；只读语义在错误态下仍成立（无写副作用）"
- State: "无状态变更；已加载数据保留"
- Side-effect: "none"

## Outcome "search-validation-error"
<!-- source: inferred -->
<!-- reasoning: surface-web.md–validation-error 派生；journey Step 2b Outcome A -->
- Preconditions: "搜索框存在格式校验规则（长度上限、禁用符号）；只读角色待提交非法搜索词（区别于合法搜索词路径）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
- Input: "只读角色输入非法搜索词并尝试提交搜索"
- Output: "前端即时拦截并标红提示格式要求，搜索请求不提交至后端；列表数据不变；只读语义保持（无写副作用）"
- State: "无状态变更；搜索词保留待修正"
- Side-effect: "none"

## Outcome "corrected-retry"
- Preconditions: "只读角色修正非法搜索词为合法值后重新搜索（区别于提交非法词被拦截）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
- Input: "只读角色修正为合法搜索词后重新搜索"
- Output: "可重新搜索，返回正确结果；不报错、不丢失已加载列表"
- State: "无状态变更；搜索结果刷新"
- Side-effect: "none"
- Invariants: "修正后重试返回正确结果，不丢失已加载列表（fact: F7）"

## Journey Invariants

- 只读角色永不渲染写入口——增删改入口在所有身份/权限页面均不渲染（fact: IAM_READONLY_MENU_PRUNE）。
- 服务端 403 是最终安全边界——任何写接口直调均被后端权限校验拒绝；前端裁剪非安全边界（fact: IAM_READONLY_ROLE_WRITE_DENY）。
- 只读操作零副作用——查看/搜索/分页不触发任何数据变更（fact: IAM_READONLY_MENU_PRUNE）。
- 403 原地收敛不跳转登录页——与「会话过期跳登录」语义区分（fact: IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 审计目标可达成——只读角色对三域页面均可读且数据完整（fact: F5, F7）。

## Fixture Specification

This Contract requires the following pre-existing data state. See `rules/fixture-spec.md` for schema details. The aggregate below is the authoritative entity set for this Step; inline `fixture_spec` blocks on individual Outcomes are subsets of this set.

```yaml
fixture_spec:
  entities:
    - entity_type: "ReadOnlyRole"
      min_count: 1
      field_constraints:
        - field: "session"
          value: "valid"
    - entity_type: "PlatformUser"
      min_count: 25
      relationships:
        - relationship_type: "has_role"
          target_entity: "Role"
    - entity_type: "Role"
      min_count: 25
      relationships:
        - relationship_type: "assigned_to"
          target_entity: "PlatformUser"
    - entity_type: "Authorization"
      min_count: 25
      relationships:
        - relationship_type: "binds"
          parent_entity: "Role"
          target_entity: "PlatformUser"
    - entity_type: "EiamService"
      min_count: 1
      field_constraints:
        - field: "reachable"
          value: "may be unreachable"
```
