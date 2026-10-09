---
journey: "readonly-audit-access"
step: "3"
step-action: "readonly-write-attempt-denied"
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

# Contract: readonly-audit-access / Step 3: 只读角色尝试修改用户/角色/授权数据被服务端拒绝

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "write-denied-403"
- Preconditions: "只读角色持有有效令牌；后端对所有用户/角色/授权的写操作做服务端权限校验（区别于令牌缺失/过期的 401）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "token"
            value: "valid"
- Input: "只读角色尝试修改用户/角色/授权数据（创建、编辑、删除），直调写接口"
- Output: "写接口返回 403 越权错误；响应体不泄露敏感数据、不含资源详情；写操作不生效；响应不触发重定向、不破坏会话令牌（403 原地收敛）"
- State: "无状态变更；写操作未生效"
- Side-effect: "写操作被服务端权限校验拒绝，返回 403 越权错误；响应体不泄露资源详情"
- Invariants: "403 原地收敛不跳转登录页（fact: IAM_READONLY_ROLE_WRITE_DENY, IAM_SESSION_EXPIRED_VS_403_DISTINCT）"

## Outcome "unauthorized-401"
<!-- source: inferred -->
<!-- reasoning: surface-api.md–unauthorized 派生；与 403 互斥——403 为已认证越权，401 为未认证 -->
- Preconditions: "请求未携带有效令牌或令牌已过期（区别于持有有效令牌的 403）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "token"
            value: "absent or expired"
- Input: "请求方未携带有效令牌或令牌已过期时尝试写操作"
- Output: "返回 401 未认证；响应体不泄露敏感数据；可重新认证后重试。语义上与 write-denied-403 互斥——403 为「已认证但越权」，401 为「未认证」"
- State: "无状态变更"
- Side-effect: "请求未通过认证，返回 401 未认证；响应体不泄露敏感数据"

## Journey Invariants

- 只读角色永不渲染写入口——增删改入口在所有身份/权限页面均不渲染（fact: IAM_READONLY_MENU_PRUNE）。
- 服务端 403 是最终安全边界——任何写接口直调均被后端权限校验拒绝；前端裁剪非安全边界（fact: IAM_READONLY_ROLE_WRITE_DENY）。
- 只读操作零副作用——查看/搜索/分页不触发任何数据变更（fact: IAM_READONLY_MENU_PRUNE）。
- 403 原地收敛不跳转登录页——与「会话过期跳登录」语义区分；只读越权不丢失登录态（fact: IAM_SESSION_EXPIRED_VS_403_DISTINCT）。
- 审计目标可达成——只读角色对三域页面均可读且数据完整（fact: F5, F7）。

## Fixture Specification

This Contract requires the following pre-existing data state. See `rules/fixture-spec.md` for schema details. The aggregate below is the authoritative entity set for this Step; inline `fixture_spec` blocks on individual Outcomes are subsets of this set.

```yaml
fixture_spec:
  entities:
    - entity_type: "ReadOnlyRole"
      min_count: 1
      field_constraints:
        - field: "token"
          value: "valid"
```
