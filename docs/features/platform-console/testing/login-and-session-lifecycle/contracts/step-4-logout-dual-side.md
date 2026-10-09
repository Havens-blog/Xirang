---
journey: "login-and-session-lifecycle"
step: "4"
step-action: "logout-triggers-dual-side-session-invalidation"
generated: "2026-10-08"
sources:
  - docs/features/platform-console/testing/login-and-session-lifecycle/journey.md

anchors:
  api: { endpoint: "", method: "", content_type: "", auth_required: }
  cli: { command: "", subcommand: "", flags: [], aliases: [] }
  tui: { command: "", interactive_prompt: "", keybindings: [] }
  web: { page: "", route: "", requires_auth: , layout: "" }
  mobile: { screen: "", navigation_path: [], deeplink: "", platform: "" }
last_anchor_sync: ""
---

# Contract: login-and-session-lifecycle / Step 4: 登出触发双侧会话失效

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->
<!-- Environment preconditions (not fixture entities):
     - eiam service reachable on same login domain; shared session store supports destroy (fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF). -->

## Outcome "console-logout-success"
- Preconditions: "用户双侧会话均处于登录态；登出发起侧为控制台（主路径），区别于 e-cam-web 侧反向登出与会话已过期登出"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "dual-side logged in"
- Input: "用户在控制台发起登出并确认"
- Output: "控制台登出生效后共享会话立即销毁；已打开的 e-cam-web 会话在下一次请求时被拒绝并跳回控制台登录页；不做实时广播"
  <!-- traceability: fact F6, IAM_SESSION_SHARED；不做实时广播可由 F6 推导（仅下一次请求时失效） -->
- State: "共享会话销毁；e-cam-web 侧会话在下一次请求时失效"
- Side-effect: "eiam 销毁共享会话；e-cam-web 下一次请求被拒绝"
  <!-- traceability: fact IAM_SESSION_SHARED, IAM_UNAUTHORIZED_401；POST /api/user/logout (eiam/internal/web/user/user.go:89)；401 eiam/internal/pkg/middleware/permission.go:24-26 -->
- Invariants: "双侧登出语义对称（fact: F6, IAM_SESSION_SHARED）"

## Outcome "ecam-web-logout-reverse"
- Preconditions: "用户双侧会话均处于登录态；登出发起侧为 e-cam-web（反向场景），区别于控制台主路径"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "dual-side logged in"
- Input: "用户在 e-cam-web 发起登出"
- Output: "控制台会话失效并回到登录页；双侧登出语义对称"
- State: "共享会话销毁；控制台侧会话在下一次请求时失效"
- Side-effect: "eiam 共享会话被销毁；控制台下一次请求被拒绝"
  <!-- traceability: fact IAM_SESSION_SHARED, IAM_UNAUTHORIZED_401；401 eiam/internal/pkg/middleware/permission.go:24-26 -->
- Invariants: "双侧登出语义对称（fact: F6, IAM_SESSION_SHARED）"

## Outcome "session-already-expired-logout"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes session-expired 派生（会话过期场景覆盖登出时已过期）；区别于主路径有效会话主动登出 -->
- Preconditions: "用户在控制台发起登出时共享会话已过期（区别于双侧有效会话的主路径登出与无有效会话的未授权调用）"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "expired before logout"
- Input: "用户在控制台发起登出，此时共享会话已过期"
- Output: "登出视为 no-op（会话已失效），系统直接重定向回控制台登录页；不报错不白屏"
- State: "共享会话保持失效态；无额外销毁动作"
- Side-effect: "eiam 鉴权拒绝过期会话；前端引导回控制台登录页"
  <!-- traceability: fact IAM_SESSION_EXPIRED_VS_403_DISTINCT, IAM_UNAUTHORIZED_401 -->

## Outcome "network-error"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes network-error 派生；登出提交时网络错误或 eiam 不可用 -->
- Preconditions: "用户发起登出并确认时网络错误或 eiam 不可用（区别于登出成功与会话已过期登出）"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "dual-side logged in"
- Input: "用户在控制台发起登出并确认，此时网络错误或 eiam 不可用"
- Output: "显示明确错误与重试入口，不白屏；本地会话状态待确认，提示用户重试登出"
- State: "共享会话状态不确定；登出未完成"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: api surface required_outcomes unauthorized 派生；无有效会话凭据调用登出端点 -->
- Preconditions: "请求方在无有效共享会话凭据的情况下调用登出端点（会话从未建立或已被另一侧登出销毁），区别于有效会话主动登出与会话已过期登出"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "absent or destroyed"
- Input: "请求方在无有效会话凭据时调用登出端点"
- Output: "登出端点拒绝请求并返回未认证状态，引导回退到控制台登录页；响应不泄露任何受保护资源数据"
- State: "无状态变更；请求被拒"
- Side-effect: "eiam 鉴权中间件拒绝无凭据请求"
  <!-- traceability: fact IAM_UNAUTHORIZED_401；eiam middleware.CheckPermission eiam/internal/pkg/middleware/permission.go:24-26 -->

## Journey Invariants

- 共享会话是双侧会话的唯一真相源——任一侧登出即销毁共享会话，其余应用/标签页在下一次请求时被拒绝并收敛回控制台登录页（fact: IAM_SESSION_SHARED）。
- 登录页安全面不泄露用户名是否存在——凭据错误文案统一（fact: IAM_LOGIN_ERROR_NO_LEAK）。
- 错误态不白屏——网络错误/eiam 不可用均保留输入、明确提示、可重试；静态内容仍可加载。
- 同登录域约束——控制台须与 e-cam-web/eiam 同域，跨域 CDN 直调 eiam 不支持（fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF）。
- 平台管理员豁免零租户限制——职责跨租户，零所属租户仍可访问全部管理页（fact: IAM_TENANT_COUNT_ROUTING）。

## Fixture Specification

This Contract requires the following pre-existing data state. See `rules/fixture-spec.md` for schema details.

```yaml
fixture_spec:
  entities:
    - entity_type: "UserAccount"
      min_count: 1
      field_constraints:
        - field: "session"
          value: "dual-side logged in"
  # eiam service reachability is an environment precondition, not a fixture entity (fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF).
```
