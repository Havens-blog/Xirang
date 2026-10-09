---
journey: "login-and-session-lifecycle"
step: "1"
step-action: "enter-login-page-and-select-method"
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

# Contract: login-and-session-lifecycle / Step 1: 进入登录页并选择登录方式

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "success-render"
- Preconditions: "用户浏览器同域访问控制台，无有效会话；eiam 支持密码/LDAP/passkey 登录方式；浏览器支持 WebAuthn（区别于 passkey-unavailable 的不支持路径）"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
      - entity_type: "IdentitySource"
        min_count: 1
        field_constraints:
          - field: "type"
            value: "password, ldap, or passkey"
- Input: "用户访问控制台，进入登录页"
- Output: "登录页正常渲染，无 JS 报错；可用登录方式入口可见且可键盘导航；页面安全响应头存在，页面不可被 iframe 嵌套"
  <!-- security-headers / iframe-nested claim: UNKNOWN — 事实表中无对应 fact_id，未在代码中核查；待 Phase 0 核查后补 fact_id -->
- State: "无状态变更；登录页就绪等待用户输入"
- Side-effect: "none"

## Outcome "passkey-unavailable"
- Preconditions: "未登录且浏览器不支持 WebAuthn（区别于支持 WebAuthn 的成功路径）；eiam 仍支持密码/LDAP 登录方式"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
- Input: "不支持 WebAuthn 的浏览器访问登录页"
- Output: "passkey 入口不渲染，显示「改用密码/LDAP 登录」提示；无 JS 报错；其余登录方式正常可用"
- State: "无状态变更；passkey 入口隐藏"
- Side-effect: "none"

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
    - entity_type: "IdentitySource"
      min_count: 1
      field_constraints:
        - field: "type"
          value: "password, ldap, or passkey"
```
