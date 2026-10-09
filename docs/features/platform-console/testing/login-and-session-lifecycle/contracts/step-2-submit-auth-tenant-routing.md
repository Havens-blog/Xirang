---
journey: "login-and-session-lifecycle"
step: "2"
step-action: "submit-auth-and-route-by-tenant-count"
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

# Contract: login-and-session-lifecycle / Step 2: 提交认证并完成租户分流

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->
<!-- Environment preconditions (not fixture entities):
     - eiam service reachable on same login domain; shared session store read/write available (fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF).
     - network-error Outcome: eiam unreachable / network failure during submit. -->

## Outcome "single-tenant-success"
- Preconditions: "用户在登录页选择密码方式，已输入有效账号密码；账号所属租户数 = 1（单租户，区别于零租户与多租户分支）；账号未启用 MFA 或 MFA 已通过"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "credentials"
            value: "valid password"
          - field: "tenant_count"
            value: 1
      - entity_type: "Tenant"
        min_count: 1
        relationship_type: "user-tenant-membership"
        parent_entity: "UserAccount"
- Input: "用户输入有效账号密码并提交登录"
- Output: "认证成功后系统建立共享会话；单租户自动选定唯一租户跳过选择页，直达工作台展示我的信息与应用快捷导航"
  <!-- traceability: fact F1, F7；eiam 登录端点 POST /api/user/system/login (eiam/internal/web/user/user.go:67) -->
- State: "共享会话建立；会话凭据中写入 tenant_id 驱动租户上下文"
  <!-- tenant_id 写入语义 UNKNOWN — fact F7，待 Phase 0 核查（见 fact-table IAM_TENANT_COUNT_ROUTING） -->
- Side-effect: "eiam 校验凭据通过后建立共享会话并写入会话凭据"
  <!-- traceability: fact IAM_LOGIN_ENDPOINT, IAM_SESSION_SHARED -->

## Outcome "mfa-lockout"
<!-- source: fact -->
<!-- reasoning: fact F2（MFA 连续失败达阈值锁定账号）；fact IAM_LOGIN_ENDPOINT 显示锁定阈值为配置驱动而非硬编码字面量 -->
- Preconditions: "账号已启用 MFA，MFA 连续失败达配置阈值（PRD 称 5 次）；区别于凭据正确或网络错误"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "mfa_enabled"
            value: true
          - field: "mfa_failed_attempts"
            value: "at configured lockout threshold"
- Input: "用户提交 MFA 验证码，连续失败达锁定阈值"
- Output: "账号锁定（PRD 称 15 分钟），登录页提示剩余解锁时间；锁定期间凭正确密码亦不可登录；可等待或改用其他已启用方式"
- State: "账号进入锁定态；锁定期间登录被拒"
- Side-effect: "锁定阈值与时长由 eiam 配置驱动（非硬编码字面量）"
  <!-- traceability: fact F2, IAM_LOGIN_ENDPOINT；eiam LocalConfig.MaxFailedAttempts/LockoutDuration (eiam/internal/domain/identity_source.go:105-106) -->

## Outcome "network-error"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes network-error 派生；fact F3（网络错误/eiam 不可用保留输入） -->
- Preconditions: "登录提交时网络错误或 eiam 不可用，区别于凭据错误与 MFA 锁定"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
- Input: "用户提交登录，此时网络错误或 eiam 不可用"
- Output: "已输入内容保留、显示明确错误与重试入口，不白屏；凭据错误文案不泄露用户名是否存在；e-cam-web 已登录会话不受影响"
- State: "无新会话建立；e-cam-web 既有会话不受影响"
- Side-effect: "none"

## Outcome "zero-tenant-restricted"
- Preconditions: "认证通过但用户所属租户数 = 0，且用户为普通用户/租户管理员（非平台管理员）；区别于单租户与多租户分支"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "credentials"
            value: "valid"
          - field: "tenant_count"
            value: 0
          - field: "role"
            value: "not platform admin"
- Input: "用户认证通过后系统按所属租户数分流，发现租户数 = 0"
- Output: "不进租户选择页，直达工作台受限视图（仅欢迎 + 我的信息）；访问管理页提示「无所属租户」，不报错不白屏；平台管理员豁免本规则"
- State: "建立受限会话；无租户上下文"
  <!-- traceability: fact F4；零租户受限分支在 eiam 代码中 UNKNOWN（见 fact-table IAM_TENANT_COUNT_ROUTING）；平台管理员豁免为 PRD 规则 -->
- Side-effect: "系统进入零租户受限工作台分支"

## Outcome "multi-tenant-switch"
- Preconditions: "用户所属租户数 ≥ 2，已进入租户选择页或工作台；区别于单租户与零租户分支"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
        field_constraints:
          - field: "credentials"
            value: "valid"
          - field: "tenant_count"
            value: ">=2"
      - entity_type: "Tenant"
        min_count: 2
        relationship_type: "user-tenant-membership"
        parent_entity: "UserAccount"
- Input: "用户经顶部租户切换器重选租户"
- Output: "当前租户上下文切换并刷新数据；切换后列表仅显示该租户数据"
- State: "会话租户上下文切换为目标租户"
  <!-- traceability: fact F5；租户切换器为前端行为（见 fact-table IAM_TENANT_COUNT_ROUTING） -->
- Side-effect: "前端切换租户上下文并刷新当前租户数据"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes validation-error 派生；form-submission 校验属本 step（提交认证）而非 step 1（进入页面） -->
- Preconditions: "用户提交登录表单时必填字段为空或字段格式非法（如用户名含非法字符），区别于凭据错误、MFA 锁定与网络错误"
  fixture_spec:
    entities:
      - entity_type: "UserAccount"
        min_count: 1
- Input: "用户提交登录表单，必填字段为空或字段格式非法"
- Output: "登录页在相关字段就近显示错误提示，表单不提交，用户可修正后重试；不泄露用户名是否存在"
- State: "无状态变更；表单数据保留"
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
    - entity_type: "Tenant"
      min_count: 2
      relationship_type: "user-tenant-membership"
      parent_entity: "UserAccount"
  # eiam service reachability is an environment precondition, not a fixture entity (fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF).
```
