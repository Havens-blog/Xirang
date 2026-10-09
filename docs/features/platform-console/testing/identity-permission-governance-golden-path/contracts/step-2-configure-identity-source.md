---
journey: "identity-permission-governance-golden-path"
step: "2"
step-action: "configure-ldap-identity-source"
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

# Contract: identity-permission-governance-golden-path / Step 2: 配置身份源（LDAP）

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing. Anchor fields empty per graceful-degradation rule. -->
<!-- Implementation anchors (endpoint POST /api/identity_source/save -> handler Save; fact IAM_IDENTITY_SOURCE_LDAP) moved out of dimension values per semantic-purity rule. -->

## Outcome "success"
- Preconditions: "管理员已登录且会话有效；LDAP 服务参数完整且格式正确（绑定地址、搜索基准、凭据齐全）；操作动作为保存身份源配置"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "LdapServer"
        min_count: 1
        relationship_type: "managed_by"
        parent_entity: "PlatformAdmin"
        field_constraints:
          - field: "reachable"
            value: true
- Input: "管理员在身份源管理页填写 LDAP 绑定地址、搜索基准、凭据并保存配置"
- Output: "配置保存成功；该身份源立即可用于登录；不重启控制台即生效"
- State: "新增 LDAP 身份源记录写入后端；身份源列表反映新项"
- Side-effect: "后端写入身份源配置并即时生效（无需重启）；登录页是否渲染对应入口 UNKNOWN — see U4"
- Invariants: "身份源配置保存失败不影响已登录会话（fact: IAM_IDENTITY_SOURCE_LDAP；会话独立性由直调架构推断 — see IAM_CONSOLE_DIRECT_EIAM_NO_BFF）"

## Outcome "save-failure"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes network-error + api required_outcomes internal-error 派生；事实 IAM_IDENTITY_SOURCE_LDAP（保存端点）；会话独立性由直调架构推断（fact IAM_CONSOLE_DIRECT_EIAM_NO_BFF） -->
- Preconditions: "后端不可用，或身份源配置字段校验失败（区别于参数完整且服务可达的成功路径）；操作动作为保存身份源配置"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "LdapServer"
        min_count: 1
        relationship_type: "managed_by"
        parent_entity: "PlatformAdmin"
        field_constraints:
          - field: "reachable"
            value: false
- Input: "管理员保存 LDAP 配置，此时后端不可用或字段校验失败"
- Output: "明确错误提示与可重试入口，静态内容仍渲染；已有登录方式不受影响；已登录会话不受影响"
- State: "身份源配置未写入；已登录会话保持有效；已有身份源不变"
- Side-effect: "none"

## Journey Invariants

- eiam 是唯一后端——所有 CRUD 与授权判定均直调 eiam，控制台零新增后端（fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF）。
- 错误态不白屏——校验失败/后端错误均明确提示、保留表单数据、列表数据不变。
- 身份源配置保存失败不影响已登录会话（fact: IAM_IDENTITY_SOURCE_LDAP；会话独立性由直调架构推断 — see IAM_CONSOLE_DIRECT_EIAM_NO_BFF）。

## Fixture Specification

This Contract requires the following pre-existing data state. See `rules/fixture-spec.md` for schema details.

```yaml
fixture_spec:
  entities:
    - entity_type: "PlatformAdmin"
      min_count: 1
    - entity_type: "LdapServer"
      min_count: 1
      relationship_type: "managed_by"
      parent_entity: "PlatformAdmin"
```
