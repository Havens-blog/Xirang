---
journey: "readonly-audit-access"
step: "1"
step-action: "enter-readonly-identity-permission-pages"
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

# Contract: readonly-audit-access / Step 1: 只读角色登录并进入身份与权限页面（只读视图）

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "success-data-present"
- Preconditions: "只读/审计角色已登录且会话有效；用户/角色/授权页面含可读数据；角色首次进入页面（区别于后续交互操作与数据为空的空状态）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
      - entity_type: "PlatformUser"
        min_count: 1
      - entity_type: "Role"
        min_count: 1
        relationships:
          - relationship_type: "assigned_to"
            target_entity: "PlatformUser"
      - entity_type: "Authorization"
        min_count: 1
        relationships:
          - relationship_type: "binds"
            parent_entity: "Role"
            target_entity: "PlatformUser"
- Input: "只读角色在控制台导航至用户管理、角色管理、授权管理页面"
- Output: "增删改入口不渲染（不可用）；页面以只读形式展示数据；可正常浏览。审计目标可达成——身份与权限现状对只读角色可读且完整"
- State: "无状态变更；只读视图就绪"
- Side-effect: "none"
- Invariants: "只读角色永不渲染写入口（fact: IAM_READONLY_MENU_PRUNE）"

## Outcome "empty-state"
- Preconditions: "只读角色已登录且会话有效；对应页数据为空（区别于含可读数据与网络错误路径）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
      - entity_type: "PlatformUser"
        min_count: 0
- Input: "只读角色进入数据为空的页面"
- Output: "对应页数据为空时显示空状态占位，仍不渲染任何写入口；空状态不报错，登录态保留"
- State: "无状态变更；空状态占位显示"
- Side-effect: "none"
- Invariants: "空页显示空状态不报错（fact: F8）"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: surface-web.md–session-expired 派生；journey Step 1c 的两个子结果共享同一前置（会话过期），按互斥规则合并，403 区分作为不变量 -->
- Preconditions: "只读角色在审计浏览过程中会话超时过期（区别于会话有效的数据/空状态路径）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "expired mid-audit"
- Input: "只读角色在会话过期后再次操作（搜索/翻页/跳转）"
- Output: "重定向至控制台登录页或提示会话已过期；登录入口可达；重新登录后可恢复审计上下文"
- State: "共享会话失效；需重新认证"
- Side-effect: "服务端返回 401 未认证；前端重定向至登录页"
- Invariants: "过期态不与 403 越权混淆——403 越权保持登录态原地收敛，会话过期则需重新登录（fact: IAM_SESSION_EXPIRED_VS_403_DISTINCT, IAM_UNAUTHORIZED_401）"

## Outcome "zero-side-effect-normal"
- Preconditions: "只读角色已在只读页面执行搜索/分页/筛选等交互操作（会话有效，正常路径；区别于首次进入页面的数据加载）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
- Input: "只读角色在只读页面上进行查询/翻页/排序"
- Output: "全程无任何写操作副作用；查询/翻页/排序不触发除审计日志外的状态变更；只读语义在所有交互路径上成立"
- State: "无状态变更"
- Side-effect: "none"
- Invariants: "只读操作零副作用——查看/搜索/分页不触发任何数据变更（fact: IAM_READONLY_MENU_PRUNE）"

## Outcome "zero-side-effect-error"
- Preconditions: "只读角色查看时后端身份服务不可达（网络错误，区别于数据为空的空状态与会话过期路径），会话有效"
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
- Input: "只读角色在错误态下进行查询/翻页/排序"
- Output: "即使发生网络错误，亦不产生任何写副作用；错误态下只读语义仍成立"
- State: "无状态变更"
- Side-effect: "none"
- Invariants: "错误态仍只读无副作用（fact: F7）"

## Journey Invariants

- 只读角色永不渲染写入口——增删改入口在所有身份/权限页面均不渲染（fact: IAM_READONLY_MENU_PRUNE）。
- 服务端 403 是最终安全边界——任何写接口直调均被后端权限校验拒绝；前端裁剪非安全边界（fact: IAM_READONLY_ROLE_WRITE_DENY）。
- 只读操作零副作用——查看/搜索/分页不触发任何数据变更（fact: IAM_READONLY_MENU_PRUNE）。
- 403 原地收敛不跳转登录页——与「会话过期跳登录」语义区分（fact: IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 审计目标可达成——只读角色对用户/角色/授权三域页面均可读且数据完整（fact: F5, F7）。

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
      min_count: 1
    - entity_type: "Role"
      min_count: 1
      relationships:
        - relationship_type: "assigned_to"
          target_entity: "PlatformUser"
    - entity_type: "Authorization"
      min_count: 1
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
