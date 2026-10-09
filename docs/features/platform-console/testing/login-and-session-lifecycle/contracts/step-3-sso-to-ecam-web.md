---
journey: "login-and-session-lifecycle"
step: "3"
step-action: "sso-jump-to-ecam-web"
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

# Contract: login-and-session-lifecycle / Step 3: 单点跳转 e-cam-web

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->

## Outcome "success-shared-session"
- Preconditions: "用户已在控制台工作台登录态；点击「云管」应用入口发起跨应用跳转；e-cam-web 与控制台同登录域且共享会话可达"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid shared session in console"
- Input: "已登录用户在工作台点击「云管」应用入口"
- Output: "单点跳转到 e-cam-web 且不提示重复登录；共享会话生效；e-cam-web 不要求再次认证"
- State: "e-cam-web 复用控制台共享会话；无需重新认证"
- Side-effect: "none"
  <!-- traceability: fact F6, IAM_SESSION_SHARED -->

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes session-expired 派生；区别于主动登出 -->
- Preconditions: "用户在 e-cam-web 工作流中操作时间超过会话有效期，下一次交互时会话已过期（区别于主动登出与无凭据访问）"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "expired in e-cam-web workflow"
- Input: "用户在会话过期后再次交互（搜索/翻页/跳转）"
- Output: "用户被引导回控制台登录页并提示会话已过期；未保存的交互内容被保留或给出数据丢失警示；登录入口可从过期态直接可达"
- State: "共享会话失效；用户需重新认证"
- Side-effect: "eiam 鉴权拒绝过期会话；前端拦截器引导回控制台登录页"
  <!-- traceability: fact IAM_SESSION_EXPIRED_VS_403_DISTINCT, IAM_UNAUTHORIZED_401；401 路径 eiam/internal/pkg/middleware/permission.go:24-26 -->

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: api surface required_outcomes unauthorized 派生；fact IAM_UNAUTHORIZED_401 -->
- Preconditions: "请求方在无有效共享会话凭据的情况下访问 e-cam-web 鉴权端点（会话从未建立或已被另一侧登出销毁），区别于会话过期中途态"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "absent or destroyed"
- Input: "请求方在无有效共享会话凭据时访问 e-cam-web 鉴权端点"
- Output: "鉴权端点拒绝请求并返回未认证状态，引导回退到控制台登录页；响应不泄露任何受保护资源数据"
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
          value: "valid shared session in console"
```
