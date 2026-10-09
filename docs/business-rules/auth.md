---
title: "授权、会话与凭据业务规则"
domains: [auth, authorization, session, sso, tenant, role, credential, audit]
---

# 业务规则：授权、会话与凭据

> 跨特性领域不变量。源：feature/platform-console。

## 服务端授权

### BIZ-auth-001: 服务端强制授权，前端裁剪非安全边界

**Rule**: eiam 对所有 CRUD 端点做服务端权限校验；前端菜单裁剪 / 按钮隐藏仅为体验层，不构成安全边界。
**Context**: 平台治理直接操作租户 / 用户 / 角色 / 授权，任何前端兜底都可被绕过（直调接口 / URL 访问），授权必须由 eiam 服务端强制。
**Source**: feature/platform-console prd-spec.md §Security Requirements；tech-design §Security

## 角色与权限分层

### BIZ-auth-002: 四角色职责边界与权限分层

**Rule**: 平台管理员（全局治理，跨租户）、租户管理员（本租户内用户 / 角色 / 授权，可绑定已有策略但不可增删策略定义）、普通用户（消费）、只读 / 审计角色（只读查看，不含租户与策略管理）。策略定义仅平台管理员可增删改。
**Context**: 身份治理与权限管理的核心域不变量，决定菜单裁剪、越权判定与审计口径。
**Source**: feature/platform-console prd-spec.md §Who；prd-user-stories.md Story 3/4/5

### BIZ-auth-003: 租户边界越权 403 原地提示不跳转登录页

**Rule**: 越权请求（前端已裁剪后直调接口或经 URL 访问）eiam 返回 403，页面原地提示无权限、保持当前页、数据不变，不跳转登录页。
**Context**: 403 与 401 语义分离——403 是授权拒绝（用户已登录但无权），跳登录页会造成「已登录却被踢出」的误解。
**Source**: feature/platform-console prd-spec.md §Business Flow #8；prd-user-stories.md Story 3/5

## 凭据与锁定

### BIZ-auth-004: 账号锁定（密码 + MFA 同口径）

**Rule**: 连续 5 次密码错误或 MFA 失败锁定账号 15 分钟；锁定期间凭正确密码亦不可登录；登录页提示剩余解锁时间，可等待或改用其他已启用方式。
**Context**: 暴力破解防御；密码与 MFA 共用同一锁定口径避免绕过。
**Source**: feature/platform-console prd-spec.md E2/E9；prd-user-stories.md Story 1

### BIZ-auth-006: 凭据错误防枚举统一文案

**Rule**: 凭据错误统一文案「用户名或密码错误」；锁定提示中性化（不区分账号级 / 方式级）；不泄露用户名是否存在。
**Context**: 登录页是凭据攻击面，统一文案阻断用户名枚举。
**Source**: feature/platform-console prd-spec.md §Security Requirements 防枚举

### BIZ-auth-007: 密码重置唯一路径（管理员重置）

**Rule**: 改密唯一路径 = 管理员重置 → 生成一次性初始密码（展示一次）→ 该用户首次登录强制改密。passkey 设备丢失重绑唯一路径 = 管理员撤销原绑定后重新发起注册引导。控制台不提供用户自助改密 / 自助重绑。
**Context**: 凭据生命周期治理；自助流程成本与风险不划算，收敛到管理员路径。
**Source**: feature/platform-console prd-spec.md E1 / §Out of Scope；tech-design §Interfaces C 档

## 会话与单点登出

### BIZ-auth-005: 双向单点登出

**Rule**: 任一侧（控制台 / e-cam-web）登出即销毁共享 Redis session，其余应用 / 标签页在下一次请求时 401 收敛并跳回控制台登录页；多标签页不做实时广播（不引入 WebSocket / BroadcastChannel）。
**Context**: 共享 session 拓扑下的退出语义；实时广播成本高且窗口狭窄，显式接受下次请求收敛。
**Source**: feature/platform-console prd-spec.md §Business Flow #7；prd-user-stories.md Story 6

### BIZ-auth-008: 会话过期跳转登录页不白屏

**Rule**: 会话过期 → 跳转控制台登录页（不白屏、不无限 loading）；重新登录后回到原操作路径。
**Context**: 会话失效是常态，必须优雅承接。
**Source**: feature/platform-console prd-spec.md §Business Flow #8；proposal §Key Scenarios

## 平台运营治理

### BIZ-auth-009: eiam API 契约基线与变更通知

**Rule**: Phase 0 冻结控制台依赖的 eiam 端点快照（路径 / 方法 / 鉴权要求 / 关键字段）登记入 `parity-checklist.md` 作为契约基线；eiam 侧变更基线内任一端点（含字段增删、语义调整）须提前 ≥ 5 个工作日书面通知并附影响评估，未经评估的变更不得发布；发布后发现的未通知变更按缺口处置登记并触发受影响场景回归。
**Context**: 控制台零后端、强依赖 eiam API 稳定性；契约基线是「eiam 完全不动」假设的对冲。
**Source**: feature/platform-console prd-spec.md §Data Requirements；tech-design §Interfaces 契约基线

### BIZ-auth-010: 平台管理操作审计由 eiam 承接

**Rule**: 平台管理操作审计由 eiam 承接（纳入 Phase 0 parity 清单逐项核对），控制台不自建审计。
**Context**: 审计是合规底线，复用 eiam 既有能力避免双份。
**Source**: feature/platform-console prd-spec.md §Data Requirements
