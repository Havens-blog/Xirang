---
feature: "platform-console"
generated: "2026-09-30"
status: draft
---

# Business Rules: 平台控制台（统一入口 + 身份治理 + 权限管理 + 工作台）

> 抽取自 `prd/prd-spec.md`、`prd/prd-user-stories.md`、`proposals/platform-console/proposal.md`。
> `[CROSS]` = 跨特性领域不变量；`[LOCAL]` = 仅本特性内有效。

## 授权与角色

### BIZ-001: 服务端强制授权，前端裁剪非安全边界

**Rule**: eiam 对所有 CRUD 端点做服务端权限校验；前端菜单裁剪 / 按钮隐藏仅为体验层，不构成安全边界。
**Context**: 平台治理直接操作租户/用户/角色/授权，任何前端兜底都可被绕过（直调接口 / URL 访问），授权必须由 eiam 服务端强制。
**Scope**: [CROSS]
**Source**: prd-spec.md §Security Requirements；proposal §NFR 安全；tech-design §Security

### BIZ-002: 四角色职责边界与权限分层

**Rule**: 平台管理员（全局治理，跨租户）、租户管理员（本租户内用户/角色/授权，可绑定已有策略但不可增删策略定义）、普通用户（消费）、只读/审计角色（只读查看，不含租户与策略管理）。策略定义（UF-10）仅平台管理员可增删改。
**Context**: 身份治理与权限管理的核心域不变量，决定菜单裁剪、越权判定与审计口径。
**Scope**: [CROSS]
**Source**: prd-spec.md §Who；prd-user-stories.md Story 3/4/5

### BIZ-003: 租户边界越权 403 原地提示不跳转登录页

**Rule**: 越权请求（前端已裁剪后直调接口或经 URL 访问）eiam 返回 403，页面原地提示无权限、保持当前页、数据不变，不跳转登录页。
**Context**: 403 与 401 语义分离——403 是授权拒绝（用户已登录但无权），跳登录页会造成「已登录却被踢出」的误解。
**Scope**: [CROSS]
**Source**: prd-spec.md §Business Flow #8；prd-user-stories.md Story 3/5

### BIZ-004: 账号锁定（密码 + MFA 同口径）

**Rule**: 连续 5 次密码错误或 MFA 失败锁定账号 15 分钟；锁定期间凭正确密码亦不可登录；登录页提示剩余解锁时间，可等待或改用其他已启用方式。
**Context**: 暴力破解防御；密码与 MFA 共用同一锁定口径避免绕过。
**Scope**: [CROSS]
**Source**: prd-spec.md E2/E9；prd-user-stories.md Story 1

### BIZ-005: 双向单点登出

**Rule**: 任一侧（控制台 / e-cam-web；过渡期含 ecmdb-web）登出即销毁共享 Redis session，其余应用 / 标签页在下一次请求时 401 收敛并跳回控制台登录页；多标签页不做实时广播（v1 不引入 WebSocket / BroadcastChannel）。
**Context**: 共享 session 拓扑下的退出语义；实时广播成本高且窗口狭窄，显式接受下次请求收敛。
**Scope**: [CROSS]
**Source**: prd-spec.md §Business Flow #7；prd-user-stories.md Story 6

### BIZ-006: 凭据错误防枚举统一文案

**Rule**: 凭据错误统一文案「用户名或密码错误」；锁定提示中性化（不区分账号级 / 方式级）；不泄露用户名是否存在。
**Context**: 登录页是凭据攻击面，统一文案阻断用户名枚举。
**Scope**: [CROSS]
**Source**: prd-spec.md §Security Requirements 防枚举；prd-user-stories.md Story 1

### BIZ-007: 密码重置唯一路径（管理员重置）

**Rule**: 改密唯一路径 = 管理员重置 → 生成一次性初始密码（展示一次）→ 该用户首次登录强制改密。v1 控制台不提供用户自助改密 / passkey 设备丢失自助重绑（passkey 重绑唯一路径 = 管理员撤销原绑定后重新发起注册引导）。
**Context**: 凭据生命周期治理；自助流程成本与风险不划算，v1 收敛到管理员路径。
**Scope**: [CROSS]
**Source**: prd-spec.md E1 / §Out of Scope；tech-design §Interfaces C 档 用户·密码重置

### BIZ-008: 会话过期跳转登录页不白屏

**Rule**: 会话过期 → 跳转控制台登录页（不白屏、不无限 loading）；重新登录后回到原操作路径。
**Context**: 会话失效是常态，必须优雅承接。
**Scope**: [CROSS]
**Source**: prd-spec.md §Business Flow #8；proposal §Key Scenarios 边界情况

## 身份数据完整性

### BIZ-009: 用户名格式约束

**Rule**: 用户名 3–32 字符，正则 `^[a-z][a-z0-9_-]{2,31}$`；前端即时拦截，唯一性由 eiam 校验。
**Context**: 用户名是跨系统稳定标识，格式约束保证可读性与一致引用。
**Scope**: [CROSS]
**Source**: prd-spec.md E10

### BIZ-010: 用户-租户为 1:N（禁止 M:N）

**Rule**: eiam 禁止用户-租户多对多关系；`tenant/batch_assign` 载荷为 `{user_ids, tenant_ids}` 但语义受 1:N 约束。
**Context**: 租户是数据隔离边界，M:N 会破坏隔离；与租户边界越权判定同源。
**Scope**: [CROSS]
**Source**: tech-design §Interfaces A 档 tenant/batch_assign

### BIZ-011: 依赖删除被拒

**Rule**: 删除被依赖对象（下属组织 / 关联用户 / 用户绑定）被拒：弹窗列出阻断计数，确认按钮置灰。
**Context**: 防止孤儿数据与悬空引用；删除前置依赖检查由 eiam 服务端强制。
**Scope**: [CROSS]
**Source**: prd-spec.md E11

### BIZ-012: 并发编辑 last-submitter-wins

**Rule**: 同一对象并发编辑：后提交者生效；若 eiam 返回 409 版本冲突则提示「数据已被他人修改，请刷新后重试」并自动重拉当前页（Phase 0 核查 eiam 是否支持 409，不支持则降级并入 validation）。
**Context**: 无乐观锁时的显式取舍；409 是 eiam 支持时的最优解。
**Scope**: [CROSS]
**Source**: prd-spec.md E6；tech-design §Error Handling conflict

### BIZ-013: 用户状态词表仅 active / disable

**Rule**: 用户状态原值为 `active` | `disable`；未知串归一 `unknown`（写错会静默降级、启停操作失效）。
**Context**: 状态词表是启停操作的契约边界，写错值不报错但功能失效是高危静默失败。
**Scope**: [CROSS]
**Source**: tech-design §Data Models User.status

## 降级与韧性

### BIZ-014: eiam 不可用降级

**Rule**: eiam 不可用时，管理页错误态可重试、工作台静态内容仍渲染、不白屏不无限 loading；已登录的 e-cam-web 会话不受影响。
**Context**: 控制台为纯静态站，资源加载与 eiam 可达性解耦；降级不能波及共享 session 的其他应用。
**Scope**: [CROSS]
**Source**: prd-spec.md E12 / §Business Flow #8；proposal §Key Scenarios

### BIZ-015: passkey 浏览器不支持隐藏入口

**Rule**: WebAuthn 不可用时隐藏 passkey 入口，提示改用密码 / LDAP，无 JS 报错。
**Context**: 浏览器能力探测失败不应崩溃登录页。
**Scope**: [CROSS]
**Source**: prd-spec.md E8

### BIZ-016: 零/单/多租户登录分流

**Rule**: 登录后查询所属租户数：0 个 → 直达工作台受限视图（仅欢迎 + 我的信息，访问管理页提示「无所属租户」）；1 个 → 自动选定唯一租户直达工作台；≥ 2 个 → 进入租户选择页。平台管理员豁免零租户限制（职责跨租户、不依赖所属租户，零所属租户仍可访问全部管理页）。
**Context**: 租户上下文是数据隔离前提；平台管理员职责跨租户需豁免。
**Scope**: [LOCAL]
**Source**: prd-spec.md E3/E4/E5 / §Business Flow #3

## 平台运营治理

### BIZ-017: eiam API 契约基线与变更通知

**Rule**: Phase 0 冻结控制台依赖的 eiam 端点快照（路径 / 方法 / 鉴权要求 / 关键字段）登记入 `parity-checklist.md` 作为契约基线；eiam 侧变更基线内任一端点（含字段增删、语义调整）须提前 ≥ 5 个工作日书面通知并附影响评估，未经评估的变更不得发布；发布后发现的未通知变更按缺口处置登记并触发受影响场景回归。
**Context**: 控制台零后端、强依赖 eiam API 稳定性；契约基线是「eiam 完全不动」假设的对冲。
**Scope**: [CROSS]
**Source**: prd-spec.md §Data Requirements；tech-design §Interfaces 契约基线

### BIZ-018: 平台管理操作审计由 eiam 承接

**Rule**: 平台管理操作审计由 eiam 承接（纳入 Phase 0 parity 清单逐项核对），控制台不自建审计。
**Context**: 审计是合规底线，复用 eiam 既有能力避免双份。
**Scope**: [CROSS]
**Source**: prd-spec.md §Data Requirements
