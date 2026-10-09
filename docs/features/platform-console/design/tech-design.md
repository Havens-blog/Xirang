---
feature: "平台控制台（统一入口 + 身份治理 + 权限管理 + 工作台）"
status: Draft
created: "2026-09-29"
---

# 平台控制台 — 技术设计

## Overview

新建**独立纯前端 SPA 仓库**（与 e-cam-web 平级，建议名 `haven-console`），直调 eiam 既有 HTTP API，**零后端、零数据库**。技术栈与 e-cam-web 完全同源，v1 采用「拷贝 + 锁同版本」共享机制：

- **框架**：Vue 3.5 + Vite 7 + TypeScript 5.9 + Vue Router 4.5 + Pinia 3（persistedstate）
- **UI**：Element Plus 2.10（锁与 e-cam-web 同版本）+ @element-plus/icons-vue，unplugin-vue-components / auto-import 按需自动引入
- **样式**：Sass + `theme-variables.scss`（Linear 靛紫令牌，深色默认）+ `element-theme.scss`（Element Plus 令牌覆盖），**从 e-cam-web 拷贝**、引用同一套 `--accent-primary #5e6ad2` / `--accent-blue #7170ff` / `--bg-*` / `--text-*` 令牌
- **HTTP**：axios（`eiamAxios` 模式：`withCredentials: true`、cookie 注入 Bearer、401 跳登录、15s 超时）
- **测试**：Vitest ^4.1.5（单元）+ Playwright ^1.55（真实浏览器 E2E，PRD 硬性要求；浏览器矩阵仅 chromium，与内网统一浏览器一致）
- **包管理**：pnpm ≥ 9，新仓 package.json 写入 `packageManager` 字段锚定精确版本（e-cam-web 现有 manifest 未声明 packageManager，控制台新仓为首个显式锚点；建仓时以 lockfile 锁死）

核心约束：登录态复用 eiam 写入的共享 Redis session（与 e-cam-service `EcmdbAuthMiddleware` 同源），**不自造鉴权**；服务端授权由 eiam 强制。

## Architecture

单 SPA 分层，镜像 e-cam-web 目录约定：

```
haven-console/
├── package.json / vite.config.ts / tsconfig.json
├── src/
│   ├── assets/styles/        # theme-variables.scss + element-theme.scss（拷贝自 e-cam-web）
│   ├── api/                  # eiam 客户端 + 类型
│   │   ├── request/eiam.ts   # eiamAxios 实例（拦截器）
│   │   ├── auth.ts           # 登录/登出/MFA/passkey
│   │   ├── users.ts / organizations.ts / tenants.ts / identity-sources.ts
│   │   └── roles.ts / authorizations.ts / policies.ts
│   ├── stores/               # Pinia，共 10 个：3 全局 + 7 页面级（清单见下）
│   ├── router/               # 路由 + 全局守卫（401/403/租户）
│   ├── layouts/              # 主布局（侧边栏 + header，见 ui-design UF-2）
│   ├── components/           # 共享 UI（租户选择器、状态徽章、文案契约渲染）
│   ├── views/                # 11 页面（对应 ui-design 组件）
│   └── utils/                # cookie、文案契约 keys、校验
├── e2e/                      # Playwright 测试（acceptance.md 映射）
└── nginx/                    # same-domain 反代配置（/api/iam -> eiam，静态托管）
```

数据流（单一方向）：`views → stores → api/eiamAxios →(HTTP)→ eiam`。

**Store 清单（10 个，可直接成任务）**：全局 3 个——`user.ts`（会话/profile/permissions）、`tenant.ts`（currentTenantId + 切换动作）、`theme.ts`；页面级 7 个——`userList.ts / orgList.ts / tenantList.ts / idpList.ts`（身份治理四页）+ `roleList.ts / authorizationList.ts / policyList.ts`（权限管理三页）。7 个 list store 统一契约：`state: { items, total, page, pageSize(默认 20、上限 100), loading, error }` + `fetch / create / update / remove`，每个 store 持有 `AbortController`，fetch 前中断在途请求。**持久化口径**：`persistedstate` 插件仅挂 `theme` 一个 store（key `hc_theme`），任何会话/列表 store 不得声明 `persist`——与「无本地数据持久化（除 cookie 会话与 theme localStorage）」一致，杜绝插件默认全量写 localStorage。**职责边界**：租户上下文唯一归属 `stores/tenant.ts`，api 层零租户状态；stores 可持有分页/筛选 UI 态，api 层仅「参数进、类型出」纯调用。

### SPA 内部组件关系（视图 ↔ 状态 ↔ 请求层）

```
┌──────────────── router 全局守卫链（beforeEach）────────────────┐
│ ① cookie 缺失 / profile 拉取失败 → /console/login              │
│ ② mustSelectTenant → /tenant-select                            │
│ ③ permissions 裁剪不通过 → 403 页（与菜单裁剪同口径）            │
│ ④ 放行 → 记录「页面进入租户快照」（写前守卫取值，Integration 6）  │
└──────────────────────────┬─────────────────────────────────────┘
                           ▼ 渲染
 ┌─── views（11，与 ui-design 一一对应）───┐ 消费  ┌──── stores（10）────┐
 │ 登录 / 工作台 / 用户 / 组织 / 租户 /      │ ────► │ user·tenant·theme   │
 │ 身份源 / 角色 / 授权 / 策略 / 租户选择    │       │ userList…policyList │
 │ + 主布局（UF-2）                        │       └─────────┬───────────┘
 └──────┬──────────────────────────────────┘                 │ 调用
        │ 复用                                               ▼
        ▼                                   ┌── api/*（9 模块）──┐
 ┌─── components（3 共享）───────────┐       │ 唯一出口 unwrapEnvelope │
 │ TenantSelector（主布局 + 租户选择）│       │ → ApiError 单一错误对象 │
 │ StatusBadge（7 个列表页共用）      │       └─────────┬──────────┘
 │ 文案契约渲染（全部 11 页文案出口）  │                 │ 唯一实例
 └───────────────────────────────────┘                 ▼
                                   request/eiam.ts（eiamAxios）
                                   · 请求拦截：cookie → Bearer 注入
                                   · 响应拦截：401 → redirectToLogin
                                     （isRedirectingToLogin 防抖，继承现仓同款）
                                   · store catch → 置 error / FieldErrors → 视图渲染
```

视图 → store 映射（一对一，可直接成任务）：用户/组织/租户/身份源/角色/授权/策略 7 页各消费同名 list store；登录页与工作台消费 user store；租户选择页消费 tenant store；主布局消费 user + tenant（TenantSelector + 应用切换器）。数据流单向闭环：view 只经 store 触发请求，api 错误经 `ApiError` 上抛由 store 归位为 `error` / `FieldErrors`，视图零直连 axios。

### 多方会话拓扑（组件关系图）

```
┌────────────────────── 浏览器（同一登录域，共享 cookie）──────────────────────┐
│  标签A  haven-console /console/*     标签B  e-cam-web /cam/*    过渡期  ecmdb-web │
└────┬──────────────────────────────────┬───────────────────────────┬────────┘
     │  静态资源 + API（withCredentials，Bearer 注入自共享 cookie）              │
┌────▼──────────────────────────────────▼───────────────────────────▼────────┐
│ nginx 同域入口                                                              │
│  /console/* → 控制台静态（可 CDN，须同登录域）   /cam/* → e-cam-web 静态        │
│  /api/iam/* → eiam :9000 /api/*     /api/* → e-cam-service（读同一 session）   │
│  过渡期：旧登录入口 301 → /console/login（ecmdb-web 重定向同规则）               │
└────┬──────────────────────────────────────────────────────────┬────────────┘
     │ 写/读/销毁 JWT session                                     │ 读同一份
┌────▼───────────────┐                             ┌─────────────▼──────────────┐
│ eiam :9000          │◄────── 共享 Redis ──────►│ e-cam-service               │
│ 认证/租户/权限/CRUD   │  （uid/username/          │ EcmdbAuthMiddleware         │
└────────────────────┘   tenant_id）             └────────────────────────────┘
```

- **SSO**：认证只发生在控制台 `/console/login` → eiam 写 Redis session；cookie 共享使 e-cam-web 免重复登录（应用切换器新标签直开）。
- **双向单点登出**：任一侧 `POST /api/iam/user/logout` 销毁 Redis session，另一侧下一个请求 401 收敛跳 `/console/login`（多标签页无实时广播，显式接受）。
- **401 收敛**：控制台 eiamAxios 与 e-cam-web 请求层统一 `redirectToLogin('/console/login')`，且不清共享 cookie。
- **cookie 域**：cookie name/domain/path/SameSite/Secure 以 Phase 0 认证链路兼容矩阵为准（eiam owner 书面会签）。
- **过渡期三方**：ecmdb-web 在线期间旧登录入口 301 至 `/console/login`，ecmdb-web 自身读同一 Redis session；ecmdb-web 下线时移除该 location。

### 拷贝面漂移防线（对「拷贝 + 锁同版本」自造第二维护面的对冲）

- **登记**：新仓 `shared-hashes.json` 逐条登记拷贝文件 × sha256 × 源仓 commit，覆盖 `theme-variables.scss`、`element-theme.scss`、`api/request/eiam.ts`（eiamAxios 模式文件）三处拷贝面。
- **CI 防线**：CI job 从 e-cam-web checkout 对应源文件重算 sha256 比对，不一致 → 红；恢复绿必须同步拷贝 + 更新 hash + 跑一条主题冒烟 E2E，漂移来源 commit 记入变更记录。
- **升级流程**：e-cam-web 侧改动主题/请求层文件 → 控制台 CI 当日变红 → 人工评估差异 → 同步拷贝 + hash 更新 + 主题冒烟回归。此防线覆盖 PRD ≥5 工作日通知条款未触及的 e-cam-web 侧文件（该条款只覆盖 eiam 端点）。

## Interfaces

### eiam API 契约（消费方）

**信封类型**：所有 eiam 响应为 ginx 信封（实盘核验自 e-cam-web `src/api/eiam-mapper.ts`——字段是 `msg` 不是 `message`）：

```ts
interface Envelope<T> { code: number; msg: string; data: T }   // 成功 code === 0
```

端点按证据分三档枚举。Phase 0 执行 PRD 的**端点差集盘点**：以 UF-1~11 逐条反推控制台所需端点全集，与 eiam 实有端点求差集，缺口端点与形状分歧逐项登记 `parity-checklist.md` 并给出处置（eiam 排期补齐 / 降级路径 / 裁剪该流）——A/B/C 三档是盘点的预填输入而非盘点边界，三档之外由 UF 反推出的端点一律进盘点。证据等级：**A 档 = 已实核**（e-cam-web 生产代码逐字吻合，证据 = 源文件）；**B/C 档 = 待核验假设**（无现仓调用先例，Phase 0 逐项核验后方为契约）。

**A 档 — e-cam-web 生产在用（实盘核验）**

| 端点 | 请求 → 响应 data |
|------|------------------|
| `POST /api/iam/user/list` | `ListQuery` → `{total, users}`，api 层归一为 `Page<User>` |
| `POST /api/iam/user/create` | `CreateUserRequest` → `number`（新 id） |
| `POST /api/iam/user/update` | `UpdateUserRequest` → `null` |
| `DELETE /api/iam/user/delete/:id` | — → `null` |
| `POST /api/iam/role/list` | `ListQuery` → `{total, roles}`（eiam 角色字段为 `id/code/name/desc`，注意是 `desc` 非 `description`） |
| `POST /api/iam/role/list/attached/user` | `{user_id, offset, limit}` → `{total, roles}` |
| `POST /api/iam/role/batch_assign` / `batch_unassign` | `{usernames, role_codes}` → `null`（角色分配以 code 为准） |
| `POST /api/iam/tenant/list` | `ListQuery` → `{total, tenants}` |
| `POST /api/iam/tenant/list/attached/user` | `{user_id, offset, limit}` → `{total, tenants}` |
| `POST /api/iam/tenant/batch_assign` / `batch_unassign` | `{user_ids, tenant_ids}` → `null`（eiam 禁止用户-租户 M:N） |
| `POST /api/iam/tenant/switch` | header `X-Active-Tenant-ID: <id>` → `null`；eiam **销毁旧 session 并重签 JWT**，调用方必须随后整页 reload |
| `GET /api/iam/user/profile` | — → `ProfileData`（`user` + `permissions: string[]` + `tenants` + `current_tenant_id` + `is_admin` + `must_select_tenant`） |
| `POST /api/iam/user/logout` | — → `null` |

**B 档 — eiam 提供、e-cam-web 未直调（登录链路；e-cam-web 现登录页在 ecmdb-web 侧。形状为 PRD UF-1 推定，Phase 0 逐项核查）**

| 端点 | 请求 → 响应 data（推定） |
|------|--------------------------|
| `POST /api/iam/user/system/login` | `{username, password}` → `LoginResult` |
| `POST /api/iam/user/ldap/login` | `{username, password, source_id?}` → `LoginResult` |
| `POST /api/iam/user/passkey/login/start` + `/finish` | start → `{options, ticket}`；finish → `{ticket, assertion}` → `LoginResult` |
| `POST /api/iam/user/login/mfa/verify` | `{mfa_ticket, code}` → `LoginResult`；连续 5 次失败锁 15 分钟（E2/E9） |

```ts
interface LoginResult {
  mfaRequired?: boolean
  mfaTicket?: string
  token?: string
  lockUntil?: number   // 账号级/方式级锁定截止时间（Unix 秒）——E2/E9「剩余解锁倒计时」唯一数据源；
                       // 随登录失败响应返回（在 data 内独立字段还是结构化 msg，Phase 0 核查）。
                       // 若 eiam 不携带：降级为中性化固定文案（不含倒计时，见 Security 防枚举），不做前端自算。
}
```

**C 档 — 控制台新依赖、无 e-cam-web 前例（Phase 0 parity 重点核查）**

| 域 | 推定端点与形状 | 用途 |
|----|----------------|------|
| 组织 | `POST /api/iam/organization/list|create|update|delete`，树形 → `Page<Organization>`（≤1000 节点/深度 ≤8） | UF-5 |
| 身份源 | `POST /api/iam/identity_source/list|create|update|delete`，载荷 `ConnParams`；`POST .../test` → 通断结果（保存前测试连接） | UF-7 |
| 策略 | `POST /api/iam/policy/list|create|update|delete`，内容为 **Statement 数组**（Effect ∈ {Allow, Deny}，见 Data Models `Policy`）；eiam 原生是 Statement[] 还是扁平单 effect 结构 → Phase 0 重点核查（归一规则见模型注释） | UF-10 |
| 授权 | 三元组 `subject/resource/action` 词表枚举端点 → `{subjects: string[], resources: string[], actions: string[]}`（响应类型推定） | UF-9 |
| 用户·密码重置 | `POST /api/iam/user/reset_password` `{user_id}` → `{initial_password}`（一次性展示 + 首登强制改密）。实核：`UpdateEiamUserRequest` 无 password 字段、`user/create` 的 password 仅建号载荷——重置无法走 `user/update` 承载 | E1 / UF-4（passkey 唯一登录方式且回收时的账号恢复路径） |
| 角色·策略绑定 | `POST /api/iam/role/policy/batch_assign|batch_unassign` `{role_id, policy_ids}` → `null`（载荷对齐 A 档 `batch_assign` 惯例，真实路径 Phase 0 核查） | UF-8 绑定策略 |
| 角色·已绑清单读取 | `POST /api/iam/role/attached/users` `{role_id, offset, limit}` → `{total, users}`；`POST /api/iam/role/attached/policies` → `{total, policies}`（方向与 A 档 `role/list/attached/user` 的 user→roles 相反，路径推定） | UF-8 角色详情页数据源（`Role.policies/users` 两字段的端点来源） |
| passkey·注册引导 | `POST /api/iam/user/passkey/register/start` → `{options, ticket}`（形状对齐 B 档 `passkey/login/start`）；`/finish` `{ticket, assertion}` → `null`；`DELETE /api/iam/user/passkey/:id`（撤销）→ `null` | UF-4 管理员发起引导 / 设备丢失重绑的撤销前置（`PasskeyRegistrationSession` 的端点来源） |

**统一请求/归一类型**：

```ts
interface ListQuery { offset: number; limit: number; keyword: string }
interface Page<T> { total: number; items: T[] }   // api 层把 eiam 的 {total, users|roles|tenants} 归一为 items
```

api 客户端每端点一个具名函数（如 `listUsers(q: ListQuery): Promise<Page<User>>`、`switchTenant(id: number): Promise<void>`），全部经 `unwrapEnvelope<T>()` 解包（见 Error Handling）。**契约基线**：Phase 0 差集盘点将 A/B/C 三档与 UF 反推全集逐项登记 `parity-checklist.md`（A 档核对差异；B/C 档按「待核验假设」核对形状与存在性，核验通过才升格为契约），eiam 变更基线内任一端点须提前 ≥5 工作日书面通知 + 影响评估（PRD Data Requirements）。

## Data Models

纯前端 TypeScript 类型（无 DB schema，db-schema=no）：

```ts
// ---- 共享引用类型 ----
interface Ref { id: number; name: string }
/** 角色引用；eiam 角色分配以 code 为准（实盘：role/batch_assign 载荷为 role_codes） */
interface RoleRef { id: number; code: string; name: string }

/** UF-7 六字段连接参数；password 仅写入、响应掩码不回显 */
interface ConnParams {
  url: string          // ldap:// 或 ldaps://host:port
  port: number         // 1–65535
  bindDn: string       // 必填
  password: string     // 写后不回显，编辑态占位 '******'
  baseDn: string       // 必填
  attrMap: { username: string; email?: string; dn?: string }   // 用户名属性必填
  timeoutSec: number   // 1–30，默认 5
}

/** passkey 注册引导会话：15 分钟有效、单次使用（机制本身列 Phase 0 parity 核查） */
interface PasskeyRegistrationSession { ticket: string; expiresAt: number; used: boolean }

// ---- 领域模型 ----
// 会话 claims（来自共享 Redis JWT）
interface SessionClaims { uid: number; username: string; tenant_id: number }

interface User {
  id: number; username: string; displayName: string
  tenantId: number; roles: RoleRef[]
  /** eiam 原值为 'active' | 'disable'（ParseStatus 未知串归一 'unknown'）；写错会静默降级、启停操作失效 */
  status: 'active' | 'disable' | 'unknown'
  loginMethod: 'password' | 'ldap' | 'passkey'
  passkeyRegistered: boolean          // 注册态显式字段（eiam 暴露方式 Phase 0 核查）
  mustChangePassword?: boolean        // 首登强制改密标记（E1；profile 是否携带 Phase 0 核查）
}

/** eiam 实盘形状（user-mapper EiamTenant：id/name/code/domain） */
interface Tenant { id: number; name: string; code: string; domain: string }
// UF-6 禁用前置需「未过期会话数 = 0」：tenant/list 现不返回会话计数（Phase 0 核查）；
// 若无端点，禁用前置改由「禁用请求被 eiam 拒绝 + 拒绝消息中的活跃会话数展示」承接
interface Organization { id: number; name: string; parentId?: number }
interface IdentitySource { id: number; name: string; type: 'ldap'; conn: ConnParams; enabled: boolean }
interface Role { id: number; code: string; name: string; desc: string; tenantId: number; policies: Ref[]; users: Ref[] }
interface Authorization { subject: string; resource: string; action: string }   // 三元组，词表自 eiam
/** 策略内容为 JSON schema（Effect/Action/Resource/Condition）：前端 JSON 语法校验，eiam 服务端 schema 校验并标红对应字段 */
type PolicyCondition = Array<{ key: string; operator: string; values: string[] }>   // operator 词表 Phase 0 核查
/** 单条声明；Effect 首字母大写对齐 UF-10 词表（eiam 原值大小写随形状核查一并登记 parity-checklist） */
interface PolicyStatement {
  effect: 'Allow' | 'Deny'
  actions: string[]
  resources: string[]
  condition?: PolicyCondition
}
/** UF-10 形状：策略 = Statement 数组，一条策略可同时含 Allow 与 Deny 声明。
 *  归一规则（Phase 0 核查项）：若实核 eiam 为「扁平单 effect」存储（顶层 effect + actions/resources），
 *  前端规范形仍为 Statement[]，扁平结构 ≡ 恰含一条 Statement 的数组；扁平结构无法表达多 Statement 策略时，
 *  该形态上报 validation（不静默截断），结论登记 parity-checklist。 */
interface Policy { id: number; name: string; statements: PolicyStatement[] }
```

## Error Handling

- **错误类型**：视图层只 catch 唯一错误对象：

```ts
type ErrorKind = 'unauthorized' | 'forbidden' | 'validation' | 'conflict' | 'unavailable' | 'unknown'
class ApiError extends Error {
  kind: ErrorKind
  code: number | null          // 信封业务码原值（透传文案契约 key）
  httpStatus: number | null
}
type FieldErrors = Record<string, string>   // 400/422 字段级错误
```

- **两种错误形态统一解包**（单入口 `unwrapEnvelope<T>()`，位于 `api/request/eiam.ts`，所有 api 客户端函数必经）：
  1. HTTP 2xx + `code != 0` → 抛 `ApiError`（kind 按分类学映射，携带业务 code）；
  2. HTTP 非 2xx：响应体可解析为 `Envelope` → 取 `{code, msg}` 抛同一 `ApiError`（业务错误与 HTTP 错误走同一文案契约）；不可解析 → 按 HTTP 状态映射 kind，message 取状态短语。
- **错误码分类学**（前端不假设 eiam 业务码区段；先 HTTP 状态、后信封 code，Phase 0 产出「状态 × code → kind」对照表入 parity-checklist）：

  | kind | 触发 | 处理 |
  |------|------|------|
  | `unauthorized` | 401 | `redirectToLogin()`（跳 `/console/login`，不清共享 cookie）+ 中断当前路由全部在途请求、list store 复位 loading/error（401 后在途状态清理） |
  | `forbidden` | 403 | 原地 toast（文案契约 `forbidden`），不跳转 |
  | `validation` | 400 / 422 | `FieldErrors` 就近渲染（aria-describedby），不弹全局 toast |
  | `conflict` | **409** | E6 并发编辑：提示「数据已被他人修改，请刷新后重试」+ 自动重拉当前页；E11 依赖删除被拒：弹窗列阻断计数、确认置灰。Phase 0 核查 eiam 是否返回 409，不支持则该分支降级并入 `validation` |
  | `unavailable` | 网络错 / 15s 超时 / 5xx | 页面错误态 + 重试；重试 ≥2 次仍失败 → 顶部降级提示 + 写操作禁用（E12：静态内容仍渲染，e-cam-web 会话不受影响） |
  | `unknown` | 其余 | 兜底文案，不白屏 |

- **文案契约**：`code != 0` 的业务码原值透传给文案契约 key（`code → 逐字文案`映射表见 ui-design），`ApiError.message` 即契约文案，视图层零二次映射；字段级校验就近 aria-describedby。

## Integration Specs

1. **登录**：控制台 `/console/login` → `POST /api/iam/user/system/login`（LDAP/passkey/MFA 同）→ eiam 写共享 Redis session → 控制台存 cookie → 租户选择（0/1/≥2 分流）→ 工作台。
2. **SSO / 应用切换**：控制台与 e-cam-web 同登录域共享 cookie；应用切换器新标签打开 e-cam-web（共享会话，不重复登录）。
3. **单点登出（双向）**：任一侧登出销毁共享 session；另一侧下次请求 401 收敛跳 `/console/login`（多标签页不做实时广播）。过渡期 ecmdb-web 在线时，权威登录入口 = 控制台 `/console/login`，ecmdb-web 重定向至此。
4. **租户上下文**：tenant_id 存共享 session，跳转后由 eiam/e-cam-service 中间件读同一份；口径以 Phase 0 认证链路事实核查为准。
5. **部署与回切开关**：nginx 同域反代（`/api/iam/*` → eiam :9000），静态资源可 CDN 加速但不得跨域直连 eiam（SameSite/CORS）；控制台与 e-cam-web 同运维面探活 + 告警。回切开关为三件套，全部无 rebuild 可执行：① nginx `/console` location 注释行切换（或 upstream 双池）+ `nginx -s reload`（秒级）；② e-cam-web 401 目标以 build 期 env 注入、新旧双产物并存，切换 = 换静态目录软链；③ 过渡期旧登录入口 301 同挂 nginx location。执行人：平台管理员 + eiam owner 双人复核（与 Related Change #1 双签同名单）；上线前在 staging 完整演练一次回切并计时（PRD 承诺 10 分钟内），演练记录（步骤/耗时/回退验证）入部署清单。
6. **跨标签页租户一致性**（PRD「错租户数据」最高优先风险的承接）：
   - **真相源**：tenant_id 唯一存在于服务端共享 session，多标签共享同一份；不引入 per-tab 快照、不做 BroadcastChannel 广播。
   - **强制刷新约定**：所有切租户路径（UF-11 首选、UF-2 切换器）成功后必走 `POST /api/iam/tenant/switch` → 整页 `location.replace()`（e-cam-web TenantSelector 同款约定），杜绝 SPA 内只换 store 不刷页。
   - **响应回读比对**：list store 每次 fetch 成功后回读响应所属租户标识与 `currentTenantId` 比对（回读字段 Phase 0 核查；响应无租户标识则跳过此层），不一致 → 丢弃渲染并 `location.replace()` 自愈。
   - **写前守卫（硬门，保证范围 = 控制台自身）**：create/update/remove 提交前比对「进入页面时的租户快照」与 `currentTenantId`，不一致先刷新再提交——阻断「标签 B 呈现 X 租户数据、发出 Y 租户请求」的写错租户形态。快照取值机制：路由守卫在每次路由进入/切换成功后把 user store 的 `currentTenantId` 记入模块级单值快照（非持久化、非第二份租户真相存储，仅页进入时刻的观测值）。
   - **显式接受边界（含跨仓残余风险登记）**：切租户后其他标签页已渲染数据保持 staleness，直至该页发起下一次请求或用户交互触发刷新；读方向 staleness 接受，写方向由写前守卫阻断——**该写保证仅覆盖控制台的 7 个 list store**。e-cam-web（资产/证书/日志等页的写提交）与过渡期 ecmdb-web 的写路径**无任何守卫**，同样存在「呈现 X 租户数据、发出 Y 租户请求」形态：服务端按 session tenant_id 强制授权，越权不发生，但数据可能写入切换后的新租户。**登记为显式接受**，理由：① 请求层租户校验侵入 e-cam-web 全部业务写路径，超出 PRD In Scope 对该仓的授权（仅边界清理与 401 切换改动）；② 触发窗口狭窄——需用户在 A 标签切租户后、B 标签未刷新的前提下完成一次写提交；③ ecmdb-web 下线即消除第三方暴露面。UI 表现：e-cam-web / ecmdb-web 页面继续呈现旧租户快照直至下一次请求。若 Phase 0 核查发现响应普遍携带租户标识、且 RC#2 范围获扩张批准，可在 e-cam-web 请求层加最小租户回读（不一致丢弃渲染 + 提示刷新），届时本边界收窄并同步修订本文档。

## 跨仓改动设计（Related Change #1/#2 承接）

### RC#1 — e-cam-web 移出 `platform/users`（M4 收尾）

- **改动文件**（路径实盘核验）：删除 `src/views/platform/users/index.vue` + `views/platform/users/components/UserForm.vue` + `src/api/eiam-users.ts` 及仅被其引用的类型；`src/router/routes.ts` 删 `/platform/users` 路由；`src/layouts/MainLayout.vue` 删菜单项 `platform-users`；`src/components/PlatformNav/index.vue` 删两处引用（图标映射 + 导航项 `go('/platform/users')`）。
- **引用盘点**（修订时实核结论，证据 = 文件:行号；Phase 0 执行时重跑全仓 grep 复核、连同 evidence 并入 parity-checklist，复核后判定生效）：
  - `eiam-users.ts`：仅被 `views/platform/users/index.vue` 与 `components/UserForm.vue` 两处 import（实核），随页面整体迁出；
  - `eiam-tenant.ts`：**保留**——`components/TenantSelector.vue` 在用（实核，cam 侧租户切换）；
  - `eiam-mapper.ts`：**判「迁移/删除」**。实况：该文件导出 `eiam / parseEiamStatus / mapEiamUser / mapEiamRole / diffSets`，**不导出 `mapEiamProfile`**——`stores/user.ts:4` 引用的是 `stores/user-mapper.ts` 的同名函数，与本文件无关（修订稿曾误判此依据，已更正）。其全部引用方 = `eiam-users.ts` + `UserForm.vue`（均随 RC#1 删除）+ 自测 `eiam-mapper.test.ts`，RC#1 后仅剩自测引用、成死代码。处置：e-cam-web 侧连测试文件一并删除；控制台侧若需 `parseEiamStatus` / `diffSets` 同款归一，在新仓重建并携带同款单测（不跨仓 import）。

  盘点逐项标注「迁移 / 保留 / 有意不迁移」三选一落点（PRD 能力映射门）。
- **回归**：e-cam-web 全量路由 smoke（菜单无死链、`/platform/users` 返回 404 页而��白屏）+ TenantSelector 租户切换回归（与 E5 同口径）；控制台侧 acceptance.md E1–E12 全过为迁移输入（PRD 前置门）。
- **回退**：迁移独立成单 commit（页面/API 删除与路由/菜单删除分 commit），回退 = `git revert` + 重发版，无数据迁移、无状态残留。
- **双签前置门**：parity-checklist 逐项 evidence（截图/接口响应/用例链接）+ 能力映射表，平台管理员 + eiam owner 双签，另须 e-cam-web 前端 owner 会签（跨仓改动归属方）。

### RC#2 — e-cam-web 会话过期/401 跳转目标切换（M1 内）

- **改动文件**（目标值实核到行）：
  - `src/api/request/index.ts:24` —— `redirectToLogin()` 登录目标值**唯一所在地**（`VITE_ECMDB_LOGIN_URL || '/login'`，并拼 `?redirect=` 回跳）；`src/api/request/eiam.ts` 仅 import 并调用该函数、**无目标值，不列改动清单**（修订稿曾误列，已更正）；
  - `src/stores/user.ts:91` —— 登出跳转目标（同一变量、同款 fallback）；
  - `src/layouts/MainLayout.vue:58` —— **「个人设置」跳转基址**（隐式耦合第三处，修订稿曾遗漏）：拼 `${base}/profile/index` 新标签打开 ecmdb-web 个人中心（fallback `window.location.origin`）。若仅把变量默认值改为 `/console/login`，该跳转会落到 `/console/login/profile/index` 死链。
- **变量用途拆分**：`VITE_ECMDB_LOGIN_URL` 现承担两个互斥语义——① 登录页目标（401 / 登出，应切 `/console/login`）；② ecmdb-web 站点基址（个人设置跳转，过渡期必须继续指向 ecmdb-web 根）。拆为两个变量：`VITE_LOGIN_TARGET`（登录目标，默认 `/console/login`；index.ts 与 user.ts 改读此变量）与 `VITE_ECMDB_WEB_BASE`（个人中心基址，值维持现 env 不变，MainLayout.vue 一行改读新变量——语义正名、零行为变化）。过渡期该基址不受旧登录入口 301 波及（301 仅覆盖登录 location）；ecmdb-web 下线时个人设置跳转随「下线就绪清单」一并处置。机制 = build 期 env 注入，不散落硬编码。
- **验证**：Playwright 断言——e-cam-web 会话过期后下一请求落 `/console/login` 且共享 cookie 未被清除（对齐 S6 AC「跳回控制台登录页」）；并入 M1 出口判据（E1/E2/E8/E9）。
- **回退**：新旧目标双产物并存（各一份 dist），切换 = nginx 静态目录软链，秒级可回切（回切三件套之一）。

### 过渡期 ecmdb-web 登录重定向

nginx 层实现（旧登录入口 301 → `/console/login`），ecmdb-web 仓库零改动；ecmdb-web 下线时移除该 location 即完成「下线就绪」部署项。会话语义见拓扑图「过渡期三方」。

## 性能预算（PRD Goal 承接）

| 指标 | 预算分解 | 手段 | 测量 |
|------|----------|------|------|
| 全量静态资源 gzip ≤ 300 KB | entry 自研 ≤ 60 + element-plus ≤ 140 + vue/router/pinia ≤ 50 + routes 懒加载视图 ≤ 90 + 其余（vendor/css/字体/图标）≤ 50（单位 KB；桶额=分类护栏，总计 300 为硬顶，2026-10-09 口径修正：懒加载视图 chunk 独立 routes 桶，不再混入 remainder，records/6.5） | 11 页路由级全 `import()` 懒加载；Element Plus unplugin 按需 + 图标按需；manualChunks 三分包；构建期预生成 .gz + nginx `gzip_static on` | 构建脚本汇总 `dist/**/*.gz` 体积，CI 超 300 KB 即 fail |
| LCP ≤ 2.5s（同域内网） | 首屏仅 entry + workbench chunk + element-plus 核心 css | 上述分包 + 静态资源 `Cache-Control: max-age=31536000, immutable`（content-hash 文件名），index.html no-cache | Playwright trace + Lighthouse，登录页与工作台两个 LCP 采样点，同域内网基线，实测报告入验收材料（PRD 验收形态） |
| 二次访问 ≤ 1s | 命中 immutable 缓存，仅 revalidate html | 同上 | 同上二次采样 |

## Testing

- **单元（Vitest，DOM 环境具名 happy-dom ^20.11.6，与 e-cam-web devDeps 同版本）**：stores / api 客户端 / 校验 utils / 文案契约，覆盖 ≥80%，`pnpm typecheck`（vue-tsc）+ `pnpm build` 通过。选 happy-dom 而非 jsdom：theme store 的 `localStorage` 持久化断言与 cookie utils 需真实 DOM 存储实现，且与现仓测试环境一致、拷贝来的 utils 测试可直接复跑。
- **组件层（@vue/test-utils，取舍声明）**：现仓 devDeps 已含 @vue/test-utils，控制台沿用，但范围收敛到三个契约敏感组件——登录表单（校验规则 / 提交 / 锁定倒计时态）、文案契约渲染组件（code → 文案逐字断言，防二次映射回归）、StatusBadge（状态归一词表渲染）。页面级交互不写组件测试、统一由 E2E 承接，避免同一路径双重维护。
- **浏览器 E2E（Playwright，硬性）**：登录/登出、SSO 跳转、双向单点登出、身份/权限 CRUD 的正向+负向+边界。验收场景映射 `acceptance.md` 的 E1–E12 与 UF-1~11（每条至少 1 个映射，负向+边界 ≥1/3），逐条附 Playwright/MCP 执行证据。
- **验收宿主**：必为最终同域 nginx 反代形态（沿用 `:8888/cam` 前例），**不得**在独立端口 dev server 上判定通过（vite 直连缺 `/api/iam` 代理 + 无共享 cookie）。

## Security

<!-- Override: Security review enabled by signal "认证/授权/权限/token" -->

- **服务端授权由 eiam 强制**：所有 CRUD 端点服务端校验，前端菜单裁剪/按钮隐藏**非安全边界**（纳入 Phase 0 核查矩阵）。
- **会话**：复用 eiam 共享 Redis session，不自造鉴权；cookie `domain/路径` 严格与 e-cam-web 一致（Phase 0 认证链路事实核查出兼容矩阵）。
- **登录页安全面**：nginx 注入 CSP / X-Frame-Options / HSTS；CSRF 归属 eiam——前提是 session cookie 按兼容矩阵携带 `SameSite=Lax/Strict`，跨站请求不携带凭据，前端因此无需 CSRF token 配合（该前提列入 Phase 0 cookie 矩阵逐项确认，前提不成立则重议）；零第三方脚本。
- **token 窃取 / XSS（具名威胁）**：控制台继承 eiamAxios「从 cookie 读 token 注入 Bearer」机制 ⇒ session cookie 必须可被 JS 读取（非 httpOnly），任何 XSS 可直接窃取平台凭证并冒用会话。缓解配对：① CSP 基线（下表）收紧脚本与连接源；② 零第三方脚本 + 构建产物无内联脚本；③ Vue 模板默认转义，**禁用 `v-html` 渲染任何 eiam 返回内容**（文案契约、字段错误一律纯文本/组件化渲染）；④ cookie `Secure/SameSite/Path/Domain` 以 Phase 0 矩阵为准并经 eiam owner 会签；⑤ 会话固定/重放：凭据由 eiam 服务端签发与销毁（租户切换即销毁重签），前端不生成、不续期、凭据不落 localStorage（persist 仅 theme，见 Architecture）。
- **CSP 指令基线**（nginx 注入、覆盖全部 `/console` 路由；起草人 = 平台管理员，复核 = eiam owner，随认证链路兼容矩阵双签；Phase 0 按真实构建产物实测收敛）：

  | 指令 | 基线 | 说明 |
  |------|------|------|
  | default-src | `'self'` | 未列指令的兜底收紧项 |
  | script-src | `'self'` | 禁 inline / eval，vite 产物均外链 content-hash 文件 |
  | style-src | `'self'` `'unsafe-inline'` | Element Plus 弹层运行时内联定位样式所需；实测可去除则收紧并记录 |
  | connect-src | `'self'` | API 仅同源 `/api/iam`，禁止跨域直连 eiam |
  | img-src | `'self'` data: | 头像 / 图标 |
  | frame-ancestors / object-src / base-uri | `'none'` / `'none'` / `'self'` | XFO 的现代等价 + 防基址劫持 |
- **防枚举**：凭据错误统一文案「用户名或密码错误」；锁定提示中性化（账号级/方式级已在 PRD 厘清）。
- **敏感面**：初始密码展示一次 + 首次登录强制改密；身份源凭据掩码不回显。