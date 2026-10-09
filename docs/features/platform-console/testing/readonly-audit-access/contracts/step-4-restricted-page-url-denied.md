---
journey: "readonly-audit-access"
step: "4"
step-action: "restricted-management-page-url-direct-denied"
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

# Contract: readonly-audit-access / Step 4: 只读角色访问受限管理页被菜单裁剪且 URL 直达被拒

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "menu-not-rendered"
- Preconditions: "只读角色已登录；租户管理与策略管理对只读角色受限（会话有效）"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
- Input: "只读角色查看菜单入口（先看租户管理与策略管理入口是否渲染）"
- Output: "租户管理与策略管理的菜单入口不渲染；仅可读域入口可见"
- State: "无状态变更；菜单按只读角色裁剪"
- Side-effect: "菜单裁剪为前端体验层，非服务端安全边界"
- Invariants: "前端菜单裁剪非安全边界（fact: IAM_READONLY_MENU_PRUNE）"

## Outcome "url-direct-403"
- Preconditions: "只读角色经浏览器地址栏直接粘贴 URL 访问受限管理页（区别于菜单查看），会话有效"
  fixture_spec:
    entities:
      - entity_type: "ReadOnlyRole"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid"
- Input: "只读角色经 URL 直接访问租户管理或策略管理"
- Output: "被服务端权限校验拒绝（403），页面原地提示无权限且数据不变；不跳转登录页；响应体不含租户/策略定义详情（不泄露受限管理内容）"
- State: "无状态变更；受限内容不泄露"
- Side-effect: "受限页访问被服务端权限校验拒绝，返回 403 越权错误；响应体不含租户/策略定义详情"
- Invariants: "不泄露受限管理内容（fact: IAM_403_NO_EXISTENCE_LEAK, IAM_SESSION_EXPIRED_VS_403_DISTINCT）"

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
```
