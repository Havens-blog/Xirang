---
journey: "identity-permission-governance-golden-path"
step: "1"
step-action: "create-platform-user-and-reset-initial-password"
generated: "2026-10-08"
sources:
  - docs/features/platform-console/testing/identity-permission-governance-golden-path/journey.md

anchors:
  api:
    endpoint: ""
    method: ""
    content_type: ""
    auth_required:
  cli:
    command: ""
    subcommand: ""
    flags: []
    aliases: []
  tui:
    command: ""
    interactive_prompt: ""
    keybindings: []
  web:
    page: ""
    route: ""
    requires_auth:
    layout: ""
  mobile:
    screen: ""
    navigation_path: []
    deeplink: ""
    platform: ""
last_anchor_sync: ""
---

# Contract: identity-permission-governance-golden-path / Step 1: 创建平台用户并重置初始密码

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->
<!-- Handbook missing (no api-handbook.md / page-map.md). Anchor fields left empty per graceful-degradation rule. -->
<!-- Implementation anchors (endpoint POST /api/user/create -> handler Create; fact IAM_USER_CREATE) moved out of dimension values per semantic-purity rule. -->

## Outcome "success"
- Preconditions: "管理员已登录且会话有效；输入的用户名符合格式要求（小写字母开头、仅含小写字母/数字/下划线/连字符、长度 3–32 字符）且系统中不存在同名用户；操作动作为新建用户"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "PlatformUser"
        min_count: 0
- Input: "管理员在用户管理页填写合规用户名并提交新增用户请求"
- Output: "操作完成后用户列表在短时间内刷新并显示该新用户"
- State: "新增用户记录写入后端用户存储；用户列表反映新用户"
- Side-effect: "后端写入新用户记录并触发列表刷新；密码重置行为不在此 success 契约中断言（见独立 Outcome initial-password-reset-unknown）"
- Invariants: "列表刷新时序 ≤ 2 秒（UNKNOWN — see U1）；控制台零新增后端"

## Outcome "initial-password-reset-unknown"
<!-- fact: IAM_USER_PASSWORD_RESET (UNKNOWN — 无管理员密码重置端点，代码未实现；PRD Story 2 only) -->
<!-- 本 Outcome 由 blindspot 拆分而来：密码重置端点未实现，不得并入 success 契约 -->
- Preconditions: "success 已创建新用户；管理员对该用户发起密码重置以生成一次性初始密码；该重置端点实现状态为 UNKNOWN"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "PlatformUser"
        min_count: 1
        relationship_type: "created_by"
        parent_entity: "PlatformAdmin"
- Input: "管理员对新建用户触发密码重置，期望获得一次性初始密码"
- Output: "UNKNOWN — 一次性初始密码展示、首登强制改密行为未获代码支持（fact IAM_USER_PASSWORD_RESET）；不可断言重置是否生效"
- State: "UNKNOWN — 无法断言密码重置是否写入或生效"
- Side-effect: "none（行为未实现，不断言）"
- Invariants: "密码重置行为整体记为 UNKNOWN — see IAM_USER_PASSWORD_RESET"

## Outcome "validation-error-format"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes validation-error 派生；用户名格式事实 IAM_USER_USERNAME_FORMAT 为 PRD-only（代码无独立校验器），前端做即时格式拦截 -->
- Preconditions: "输入的用户名不符合格式要求——首字符非小写字母，或长度不在 3 至 32 字符范围内，或含非法字符；操作动作为提交新增用户"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
- Input: "管理员在用户名输入框输入不符合格式要求的值（如首字符非小写字母、长度越界、含非法字符）并尝试提交"
- Output: "前端即时拦截并标红提示格式要求，表单不提交至后端；列表数据不变；用户名唯一性最终由后端校验"
- State: "无后端状态变更；表单数据保留待修正"
- Side-effect: "none"
- Invariants: "校验失败不白屏；表单数据保留"

## Outcome "conflict-duplicate"
<!-- source: inferred -->
<!-- reasoning: api surface required_outcomes conflict + web required_outcomes validation-error 派生；用户名唯一性由后端校验，重复提交返回错误（fact IAM_USER_CREATE）。409 版本冲突响应 UNKNOWN — see IAM_VERSION_CONFLICT_409 -->
- Preconditions: "提交的用户名已存在，或后端返回校验/写入错误（区别于格式错误与并发版本冲突）；操作动作为提交新增用户"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 1
        field_constraints:
          - field: "session"
            value: "valid active session"
      - entity_type: "PlatformUser"
        min_count: 1
        relationship_type: "same_username"
        parent_entity: "PlatformAdmin"
        field_constraints:
          - field: "username"
            value: "same as submitted value"
- Input: "管理员提交新增用户请求，用户名与已存在用户重复"
- Output: "对应字段标红并回显错误原因，表单数据保留，列表数据不变；修正后可重新提交；页面不白屏"
- State: "无新用户写入；用户列表不变"
- Side-effect: "none"

## Outcome "concurrent-version-conflict"
<!-- source: inferred -->
<!-- reasoning: web surface required_outcomes concurrent-edit + api required_outcomes conflict 派生；并发版本冲突事实 IAM_VERSION_CONFLICT_409 为 UNKNOWN（代码无乐观锁），按 PRD E6 描述 -->
- Preconditions: "两个管理员会话同时编辑同一用户对象；先提交者的改动已生效，后提交者基于过期版本提交；操作动作为提交编辑"
  fixture_spec:
    entities:
      - entity_type: "PlatformAdmin"
        min_count: 2
        field_constraints:
          - field: "session"
            value: "two concurrent valid sessions"
      - entity_type: "PlatformUser"
        min_count: 1
        relationship_type: "edited_by"
        parent_entity: "PlatformAdmin"
        field_constraints:
          - field: "version"
            value: "stale in the later submitter's view"
- Input: "后提交者基于过期版本提交同一用户对象的编辑"
- Output: "后提交者收到版本冲突提示，要求刷新后重试，不覆盖他人改动"
- State: "对象保持先提交者的改动；后提交者编辑不生效"
- Side-effect: "none"

## Journey Invariants

- eiam 是唯一后端——所有 CRUD 与授权判定均直调 eiam，控制台零新增后端（fact: IAM_CONSOLE_DIRECT_EIAM_NO_BFF）。
- 权限判定逐请求生效——分配/撤销授权无需用户重新登录；撤销生效延迟上限 60 秒（代码实测；PRD 称 ≤30 秒未获代码支持，记为 UNKNOWN — see IAM_AUTH_CACHE）。
- 前端菜单裁剪仅为体验层、非安全边界——eiam 对所有 CRUD 端点做服务端权限校验；直接调用写接口必被拒绝（403）。
- 列表刷新时序——CRUD 操作成功返回后列表 ≤ 2 秒内反映变更（UNKNOWN — see U1, Phase 0 核查）。
- 错误态不白屏——校验失败/后端错误均明确提示、保留表单数据、列表数据不变。

## Fixture Specification

This Contract requires the following pre-existing data state. See `rules/fixture-spec.md` for schema details.

```yaml
fixture_spec:
  entities:
    - entity_type: "PlatformAdmin"
      min_count: 1
      field_constraints:
        - field: "session"
          value: "valid active session"
    - entity_type: "PlatformUser"
      min_count: 0
  # Outcome-specific fixture deltas declared inline per Outcome above.
```
