---
feature: "platform-console"
created: "2026-09-29"
status: Draft
task: "1.5 acceptance.md 验收矩阵骨架"
---

# acceptance — 验收矩阵（骨架）

> 验收场景宿主（PRD Goals「登录/SSO 闭环可用」验收清单条款 + tech-design Testing E2E 映射的承载文件）。
> 本文件由任务 1.5 建立骨架（场景项 + 来源映射 + 类型 + 里程碑归属 + owner 会签位）；M1–M4 各里程碑 E2E 任务（2.12 / 3.4 / 4.6 / 5.5）随里程碑回填执行证据；**全部场景通过是 RC#1（5.4）迁移前置门的输入之一**。

## 0. 口径与使用说明

### 0.1 状态词表

| 状态 | 含义 |
|------|------|
| 待执行 | 骨架初始态（当前全部场景的唯一合法状态） |
| 通过 | 真实浏览器证据已附（Playwright 用例链接或浏览器 MCP 会话记录） |
| 失败 | 有执行证据且未达预期（须登记缺陷链接后复测） |
| 阻塞 | 依赖缺口处置未决（标注依赖编号，见 §5） |

**骨架纪律**：不得预标「通过/已完成」；「证据」列一律留空（占位「待回填」），待各里程碑 E2E 任务回填。

### 0.2 证据形态与验收宿主

- **证据形态**：Playwright 用例链接（`e2e/*.spec.ts::test`）或浏览器 MCP 会话记录，二选一、逐条必附（PRD「完成形态」）。
- **验收宿主**：必为最终同域 nginx 反代形态（沿用 `:8888/cam` 前例）；**不得**在独立端口 dev server 上判定通过（tech-design Testing：vite 直连缺 `/api/iam` 代理 + 无共享 cookie）。
- **类型**：正向 / 负向 / 边界。
- **涉 eiam**：✓ = 该场景预期结果含对 eiam 服务端行为/响应契约的断言（会话颁发与销毁、锁定、错误码/拒绝、唯一性、词表来源、重签等），对应行纳入 eiam owner 会签位；纯前端渲染/校验/守卫行为记 —。

### 0.3 负向 + 边界占比（门槛 ≥ 1/3；按 ≥ 24 条计即 ≥ 8 条）

| 里程碑 | 场景数 | 负向+边界 | 占比 |
|--------|--------|-----------|------|
| M1 | 13 | 7 | 54% |
| M2 | 16 | 9 | 56% |
| M3 | 18 | 11 | 61% |
| M4 | 12 | 9 | 75% |
| **合计** | **59** | **36** | **61%** |

### 0.4 来源覆盖索引（E1–E12 每条 ≥1、UF-1~11 每个 ≥1）

| 来源 | 场景 ID |
|------|---------|
| E1 密码重置 | S-M1-11、S-M1-12 |
| E2 账号锁定 | S-M1-07、S-M1-08 |
| E3 零租户 | S-M2-03、S-M2-04 |
| E4 单租户 | S-M2-02 |
| E5 多租户切换 | S-M2-01（选择页）、S-M2-05、S-M2-06 |
| E6 并发编辑 | S-M4-11、S-M4-12 |
| E7 分页边界 | S-M3-08、S-M3-09 |
| E8 passkey 不支持 | S-M1-10 |
| E9 MFA 上限 | S-M1-09 |
| E10 用户名长度/字符集 | S-M3-06、S-M3-07 |
| E11 依赖删除失败 | S-M3-12、S-M4-03 |
| E12 eiam 不可用 | S-M2-08、S-M2-15、S-M2-16 |
| UF-1 登录页 | S-M1-01 ~ S-M1-10 |
| UF-2 顶部导航 + 应用切换器 | S-M2-05、S-M2-09 ~ S-M2-14 |
| UF-3 工作台 | S-M2-07、S-M2-08、S-M2-16 |
| UF-4 用户管理 | S-M1-11、S-M1-12、S-M3-01 ~ S-M3-07 |
| UF-5 组织管理 | S-M3-10 ~ S-M3-12 |
| UF-6 租户管理 | S-M3-13 ~ S-M3-15 |
| UF-7 身份源管理 | S-M3-16 ~ S-M3-18 |
| UF-8 角色管理 | S-M4-01 ~ S-M4-04 |
| UF-9 授权管理 | S-M4-05 ~ S-M4-07 |
| UF-10 策略管理 | S-M4-08 ~ S-M4-10 |
| UF-11 租户选择 | S-M2-01 ~ S-M2-04 |
| RC#2（补充，非 E/UF 强制项） | S-M1-13 |

---

## 1. M1 — 登录/登出/会话（证据回填：2.12 / 2.13）

> M1 出口判据：登录正/负/边界场景（E1/E2/E8/E9）真实浏览器通过（PRD Milestones）。
>
> **2.12 证据回填说明**：Playwright E2E 用例位于 `haven-console/e2e/login.spec.ts`，验收宿主为 nginx 同域形态（`:8888/console`，任务 2.10 haven-console.conf，生产构建产物 + CSP 注入），chromium 真实浏览器执行（Hard Rule 已守）。
> 证据分两档（透明标注，对应 spec 文件头「证据形态分两档」）：
> - **REAL**（真实 eiam 链路，零拦截）：负向/空凭据/E8 — dev eiam 可直接服务的场景，浏览器 → nginx → eiam :9000 真链路。
> - **CONTRACT**（契约履约）：正向登录/E2 锁定/E9 MFA 锁定/MFA/LDAP/passkey 正向 — dev eiam 无法 provision 的状态（无已知测试账号、无 local identity source 配置致 `MaxFailedAttempts=0` 永不锁定 [G-9/B-2]、无 MFA 注册账号、无 LDAP 源、无 passkey 注册 [G-2]、无 `must_change_password` 载体 [G-8]）。经 `page.route` 履约 eiam 响应信封（形状取自 parity §3.2 实核 + errs.go 原文 msg），浏览器 cookie 写入/守卫校验为真实链路（login 响应附 `Set-Cookie`，非注入伪造 cookie），仅 API 响应体按契约履约。eiam owner 会签位待 provision 真实测试账号后复核 REAL 化。

| ID | 场景项 | 来源 | 类型 | 涉 eiam | 依赖 | 状态 | 证据 |
|----|--------|------|------|---------|------|------|------|
| S-M1-01 | system 密码登录成功 → 共享 session 建立（Set-Cookie）→ 进入租户分流（UF-11） | UF-1 | 正向 | ✓ | — | 通过 | `e2e/login.spec.ts::CONTRACT · S-M1-01 登录正向 → 分流`（单租户→/workbench、零租户→/workbench、≥2→/tenant-select、redirect 回跳，4 用例） |
| S-M1-02 | LDAP 登录成功（parity B-3：请求无 `source_id`，单 ldap 通道；多身份源选择语义随 1.2 核查） | UF-1 | 正向 | ✓ | B-3 | 通过 | `e2e/login.spec.ts::CONTRACT · S-M1-02 LDAP 正向登录`（CONTRACT：dev eiam 无 LDAP 源 [G-2]） |
| S-M1-03 | passkey 登录成功（WebAuthn 可用；会话标识 `session_token` + header，B-4 改判口径） | UF-1 | 正向 | ✓ | B-4 | 通过 | `e2e/login.spec.ts::CONTRACT · S-M1-03 passkey 正向登录`（CONTRACT：dev eiam 无 passkey 注册 [G-2]，navigator.credentials.get 经虚拟凭据履约） |
| S-M1-04 | MFA 二次验证通过 → 进入租户分流（verify 字段为 `mfa_token`，B-1 改判口径） | UF-1 | 正向 | ✓ | B-1 | 通过 | `e2e/login.spec.ts::CONTRACT · S-M1-04 MFA 正向 → 二次验证通过 → 分流`（CONTRACT：dev eiam 无 MFA 注册账号） |
| S-M1-05 | 错误凭据 → 统一中性文案「用户名或密码错误」，不泄露用户名是否存在（防枚举） | UF-1 | 负向 | ✓ | — | 通过 | `e2e/login.spec.ts::REAL · S-M1-05 错误凭据 → 中性文案`（REAL：真 eiam 4010202；错密码 + 不存在用户两用例，均不回显用户名） |
| S-M1-06 | 凭据为空不可提交（表单级拦截，不发请求） | UF-1 | 边界 | — | — | 通过 | `e2e/login.spec.ts::REAL · S-M1-06 空凭据 → 前端拦截不发请求`（REAL：零 API 请求断言；均空 + 仅用户名两用例） |
| S-M1-07 | 连续 5 次密码错误 → 账号锁定 15 分钟；锁定提示为中性文案（B-2：无 lockUntil 数据源，剩余倒计时降级中性固定文案，不做前端自算） | E2 / UF-1 | 负向 | ✓ | B-2 | 通过 | `e2e/login.spec.ts::CONTRACT · S-M1-07 E2 连续 5 次密码错 → 锁定`（CONTRACT：dev eiam 无 local identity source 配置致 MaxFailedAttempts=0 永不锁定 [G-9/B-2]；履约 eiam ErrUserLocked 原文 msg，断言 auth.account_locked + 无自算倒计时） |
| S-M1-08 | 锁定期内改用其他已启用登录方式可登录（提示语含可改用方式，不回显账号信息） | E2 | 边界 | ✓ | B-2 | 通过 | `e2e/login.spec.ts::CONTRACT · S-M1-08 锁定期内可改用其他已启用方式`（CONTRACT：锁定后切 LDAP Tab 清锁定条、表单可用、不回显账号） |
| S-M1-09 | MFA 连续 5 次失败 → 锁定（与 E2 同口径），中性提示 | E9 / UF-1 | 负向 | ✓ | B-2 | 通过 | `e2e/login.spec.ts::CONTRACT · S-M1-09 E9 MFA 连续 5 次失败 → 锁定`（CONTRACT：无 MFA 注册账号；履约 ErrMfaAttemptsExhausted，第 5 次锁定态） |
| S-M1-10 | WebAuthn 不可用浏览器 → passkey 入口隐藏 + 提示改用密码/LDAP，无 JS 报错 | E8 / UF-1 | 边界 | — | — | 通过 | `e2e/login.spec.ts::REAL · S-M1-10 E8 WebAuthn 不可用`（REAL：chromium 经 addInitScript 置空 PublicKeyCredential，passkey Tab 不渲染 + 无 pageerror + 密码降级可用） |
| S-M1-11 | 管理员重置密码 → 生成一次性初始密码且仅展示一次 | E1 / UF-4 | 正向 | ✓ | G-1 | 阻塞 | 依赖 G-1（`user/reset_password` 端点缺口，eiam 排期补齐）；属 UF-4 用户管理流，非登录页态，2.12 不覆盖，随 M3 UF-4 落地 |
| S-M1-12 | 一次性初始密码登录 → 首登强制改密表单，未改密不可绕过进入业务页（改密标记载体随 1.2 核查） | E1 / UF-4 | 边界 | ✓ | G-1 | 通过（骨架） | 骨架：`e2e/login.spec.ts::E1 · S-M1-11/12 强制改密骨架（G-8 降级）`（生产构建 mustChangePassword 相位休眠断言）+ 组件层 `src/views/login/Login.test.ts::AC：强制改密态（E1 骨架）`（prop=true 表单/校验/不发请求）。完整首登触发流阻塞于 G-8（eiam 无 `must_change_password` 载体） |
| S-M1-13 | e-cam-web 会话过期 → 下一请求收敛落 `/console/login` 且共享 cookie 未被清除（RC#2 补充场景，2.13 回填；对齐 Story 6 AC） | RC#2 / Story 6 | 正向 | ✓ | — | 待执行 | （待回填；2.13 RC#2） |

---

## 2. M2 — 租户分流 + 工作台 + 应用切换（证据回填：3.4）

> M2 出口判据：E3/E4/E5 通过；单点跳转与双侧登出（Story 6）通过。E12 的 E2E 按任务 2.12 约定归 M2。
>
> **3.4 证据回填说明**：Playwright E2E 用例位于 `haven-console/e2e/m2.spec.ts`，验收宿主为 nginx 同域形态（`:8888/console`，任务 2.10 haven-console.conf，生产构建产物 + CSP 注入），chromium 真实浏览器执行（Hard Rule 已守）。baseURL 指向 `http://localhost:8888`（eiam session cookie Domain=localhost，127.0.0.1 宿主会拒收 cookie 致登录态丢失）。
> 证据分两档（透明标注，与 M1 login.spec.ts 同口径）：
> - **REAL**（真实 eiam + 真实 nginx + 真实共享 cookie）：`admin/12345678` 是 dev eiam seed 的真实超级管理员（`migrations/20260403000001_seed_system_roles.sql`），多租户（id 1/2/3）。租户切换（E5）/ SSO 跳转 / 双侧登出 / 写前守卫跨标签场景全程真链路，零 page.route 拦截。双侧登出经真实共享 cookie jar：一侧 eiam logout 的 `Set-Cookie(Max-Age=0)` 清域级 cookie，另一侧下一请求无 cookie → eiam/e-cam-service 401 → 收敛 `/console/login`（Hard Rule「真实双会话验证，不得 mock」已守）。
> - **CONTRACT**（契约履约）：E3 零租户 / E4 单租户 / 菜单裁剪 / 多租户 currentTenantId=0 选择页 —— dev eiam 无法 provision 零租户/单租户/非 admin 用户态。经 `page.route` 履约 eiam profile 信封（形状取自 parity §3.2 实核），浏览器 cookie 写入/守卫校验为真实链路（login 响应附 `Set-Cookie`，非注入伪造 cookie），仅 API 响应体按契约履约。
> **S-M2-06 写前守卫**：3.1 工具 `enforceWriteGuard` 在 4.x/5.x list store 接入前的真浏览器验证。经 localhost 限定的 `window.__hcDebug` 调试钩子（main.ts，hostname 非 localhost 零暴露；仅暴露既有函数 enforceWriteGuard/快照读/复位/fetchProfile/currentTenantId 只读，无新逻辑、不绕过鉴权）在真实跨标签 stale-snapshot 条件下调用守卫，断言阻断（返回 false）+ `location.replace` 自愈 + 自愈后 snapshot=currentTenantId 一致。

| ID | 场景项 | 来源 | 类型 | 涉 eiam | 依赖 | 状态 | 证据 |
|----|--------|------|------|---------|------|------|------|
| S-M2-01 | 所属租户 ≥ 2 → 租户选择页点选 → 写回 session → 进工作台 | UF-11 / E5 | 正向 | ✓ | — | 通过 | `e2e/m2.spec.ts::REAL · S-M2-01 多租户登录 → /tenant-select`（CONTRACT：履约 current_tenant_id=0 多租户 profile → 守卫 ② 拦 /tenant-select + ts-list el-select 渲染；admin 真实态 current_tenant_id=1≠0 不入此分支，故该形态经契约履约） |
| S-M2-02 | 单租户用户 → 自动选定唯一租户直达工作台（跳过选择页、无停留） | E4 / UF-11 | 边界 | ✓ | — | 通过 | `e2e/m2.spec.ts::CONTRACT · S-M2-02 单租户用户 → 自动选定直达工作台（E4）`（CONTRACT：dev eiam 无单租户测试账号；单租户 profile → 直达 /workbench，不经 /tenant-select，非受限态） |
| S-M2-03 | 零租户普通用户 → 认证通过进受限工作台（仅欢迎+我的信息）；URL 直入管理页提示「无所属租户」，不报错不白屏 | E3 / UF-11 / UF-3 | 边界 | ✓ | — | 通过 | `e2e/m2.spec.ts::CONTRACT · S-M2-03 零租户普通用户 → 受限工作台（E3）`（CONTRACT：零租户非 admin → workbench-restricted-notice 渲染 tenant.none + 应用卡片 data-disabled + profile-tenant data-restricted；URL 直入 /users 占位页不白屏；2 用例） |
| S-M2-04 | 平台管理员零租户豁免 → 不进选择页、可访问全部管理页 | E3 / UF-11 | 边界 | ✓ | — | 通过 | `e2e/m2.spec.ts::CONTRACT · S-M2-04 平台管理员零租户豁免（E3）`（CONTRACT：admin 零租户 → 非受限态 + 可访 /tenants requiresAdmin 页；非 admin 零租户直入 /policies → 守卫 ③ /forbidden 不白屏；2 用例） |
| S-M2-05 | 顶部租户切换器重选 → `tenant/switch`（销毁旧 session 重签 JWT）→ 整页 `location.replace` → 数据刷新为新租户 | E5 / UF-2 | 正向 | ✓ | — | 通过 | `e2e/m2.spec.ts::REAL · S-M2-05 多租户切换器重选（E5）`（REAL admin：header TenantSelector 切「默认租户空间」→ tenant/switch → location.replace reload → profile-tenant 刷新为新租户名；零拦截真链路） |
| S-M2-06 | 跨标签页错租户写提交被写前守卫阻断（先刷新再提交；3.4 约定的补充负向场景） | E5 / Integration 6 | 负向 | — | — | 通过 | `e2e/m2.spec.ts::REAL · S-M2-06 跨标签错租户写被写前守卫阻断（3.1 验证，负向）`（REAL admin 双标签共享 cookie：B 进 /users snapshot=1、A 切租户 1→2、B fetchProfile→current=2/snapshot 仍=1 → enforceWriteGuard 返回 false 阻断 + location.replace 自愈 → 自愈后 snapshot=current=2；3.1 工具级验证，4.x/5.x list store 接入后转完整写路径） |
| S-M2-07 | 工作台渲染欢迎语 + 我的信息（profile）+ 应用快捷导航卡片 | UF-3 | 正向 | — | — | 通过 | `e2e/m2.spec.ts::CONTRACT · S-M2-02 / S-M2-03`（workbench profile 四字段 username/nickname/tenant/role + workbench-apps 卡片渲染断言；REAL S-M2-05 同款 profile-tenant 字段） |
| S-M2-08 | 我的信息接口失败 → 静态内容仍渲染 + 明确错误提示（联动 E12 口径） | UF-3 / E12 | 边界 | — | — | 待执行 | （待回填；E12 eiam 不可用态属 M2 但本任务 AC 未含，随 E12 专项或 4.6 补齐） |
| S-M2-09 | 菜单按 profile.permissions 渲染，越权菜单不渲染（前端裁剪仅为体验层） | UF-2 | 正向 | — | — | 通过 | `e2e/m2.spec.ts::CONTRACT · S-M2-09 菜单按 permissions 渲染`（CONTRACT：非 admin 单租户直连 /tenants requiresAdmin → 守卫 ③ /forbidden；前端裁剪体验层，服务端授权由 eiam 强制） |
| S-M2-10 | 应用切换器新标签打开 e-cam-web，共享会话不重复登录（SSO 跳转） | UF-2 / Story 6 | 正向 | ✓ | — | 通过 | `e2e/m2.spec.ts::REAL · S-M2-10 应用切换器新标签开 e-cam-web（SSO）`（REAL admin：AppSwitcher 新标签 /cam/ 加载 e-cam-web，共享 cookie 不跳 /console/login；零拦截真链路） |
| S-M2-11 | 控制台登出 → 双侧会话失效；e-cam-web 下一次请求 401 收敛跳 `/console/login`（不做实时广播） | UF-2 / Story 6 | 正向 | ✓ | — | 通过 | `e2e/m2.spec.ts::REAL · S-M2-11 控制台登出 → e-cam-web 下次请求 401 收敛`（REAL 双会话：console AppHeader 登出 → eiam logout Set-Cookie(Max-Age=0) 清域级 cookie → e-cam-web goto /cam/ 触发初始 API → 无 cookie → 401 → 收敛 /console/login；Hard Rule 真实双会话无 mock） |
| S-M2-12 | 反向：e-cam-web 侧登出 → 控制台下一次请求 401 收敛跳 `/console/login` | UF-2 / Story 6 | 正向 | ✓ | — | 通过 | `e2e/m2.spec.ts::REAL · S-M2-12 e-cam-web 登出 → 控制台下次请求 401 收敛`（REAL 双会话：e-cam-web MainLayout 退出登录 → eiam logout 清域级 cookie → console reload → 守卫 ① hasSessionCookie()=false → /console/login；Hard Rule 真实双会话无 mock） |
| S-M2-13 | 无权限 403 → 原地提示（forbidden 文案契约），不跳转登录页 | UF-2 | 负向 | ✓ | — | 待执行 | （待回填；403 原地 toast 态属 list store catch 路径，4.x/5.x list store 落地后补齐；守卫 ③ /forbidden 路由级已由 S-M2-04 覆盖） |
| S-M2-14 | 会话过期 401 → 跳 `/console/login`、不清共享 cookie、中断在途请求并复位 list store 状态 | UF-2 | 负向 | ✓ | — | 待执行 | （待回填） |
| S-M2-15 | eiam 不可用（网络错/15s 超时/5xx）→ 管理页错误态可重试；重试 ≥ 2 次失败 → 顶部降级提示 + 写操作禁用 | E12 | 负向 | — | — | 待执行 | （待回填） |
| S-M2-16 | eiam 不可用 → 工作台静态内容仍渲染；e-cam-web 已登录会话不受影响 | E12 / UF-3 | 边界 | ✓ | — | 待执行 | （待回填） |

---

## 3. M3 — 身份治理四页（证据回填：4.6）

> M3 出口判据：身份治理 CRUD 场景与 E7/E10/E11 通过（PRD Milestones）。E6 并发编辑按任务 4.6 约定归 M4 权限页统一补齐。
>
> **4.6 证据回填说明**：Playwright E2E 用例位于 `haven-console/e2e/m3.spec.ts`，验收宿主为 nginx 同域形态（`:8888/console`，任务 2.10 haven-console.conf，生产构建产物 + CSP 注入），chromium 真实浏览器执行（Hard Rule 已守）。baseURL 指向 `http://localhost:8888`（eiam session cookie Domain=localhost）。≤2s 刷新以真实运行时序断言（`expect(...).toBeVisible({ timeout: 2000 })` 门槛 + `Date.now()` 差值 ≤2000ms 双重断言，超时即失败，不以缩短等待绕过；Hard Rule 已守）。
> 证据分两档（透明标注，与 M1/M2 同口径）：
> - **REAL**（真实 eiam + 真实 nginx + 真实共享 cookie，零拦截）：`admin/12345678` 是 dev eiam seed 真实超级管理员。用户 CRUD 正向（create→edit→delete 全链路）+ 列表读/搜索/分页 + 租户/身份源列表读 + 用户名格式前端 hard 拦截（零 create 请求）全程真链路。REAL 创建的测试用户用 `e2e_` 前缀 + 时间戳，每例 try/finally 自清理（delete），不残留 dev eiam（已核验 user/list keyword=e2e total=0）。
> - **CONTRACT**（契约履约）：E7 分页边界（需精确 total/page 控制）、E10 唯一性（dev eiam 重复创建返回 `4010902` JIT 预配失败，非干净唯一性错误）、E11 依赖删除（需 eiam conflict 信封）、组织全系（`/api/department/*` 未经 nginx 反代至 eiam，acceptance host 不可达，故全履约 department 信封）、租户/身份源写操作（避免污染 dev eiam 测试数据）。经 `page.route` 履约 eiam 信封（形状取自 parity §3.2 实核 + api/*.ts 原生路径），浏览器 cookie 写入/守卫校验为真实链路（login 响应附 `Set-Cookie`，非注入伪造 cookie），仅 API 响应体按契约履约。≤2s 刷新断言走真实运行时序（route fulfill → store refetch → DOM render，全程经 nginx 同域链路）。
> **S-M3-04/05 passkey 注册引导**依赖 G-2（注册引导端点缺口）/G-3（撤销承载核验）gap，本任务不覆盖，保留待执行（随 gap 处置收敛后补齐）。
> **S-M3-07 实现口径备注**：UserList.vue 将 eiam 唯一性拒绝渲染为 form-level `.form-error`（非 username 字段级 `#create-username-error`），与 acceptance「就近渲染（FieldErrors）」措辞略有出入；本例按 impl 实际断言 form-level 标红 + 表单保留 + 列表不变，字段级就近渲染的收敛随后续 UI 规范统一。

| ID | 场景项 | 来源 | 类型 | 涉 eiam | 依赖 | 状态 | 证据 |
|----|--------|------|------|---------|------|------|------|
| S-M3-01 | 用户列表加载/搜索/分页（ListQuery → Page 归一） | UF-4 | 正向 | ✓ | — | 通过 | `e2e/m3.spec.ts::REAL · S-M3-01 用户列表加载/搜索/分页（UF-4）`（REAL admin：/users 加载 admin 行 + 关键词搜索命中 + 分页信息渲染，≤2s） |
| S-M3-02 | 新增密码账号 → 初始密码（管理员设置或系统生成）展示一次；passkey 账号完成注册引导才可提交（联动 S-M3-04/05） | UF-4 | 正向 | ✓ | — | 通过 | `e2e/m3.spec.ts::REAL · S-M3-02/03 用户 create→edit→delete 全链路`（REAL admin：create → `#initial-password` 一次性展示 + 列表 ≤2s 刷新显示新用户；自清理。passkey 分支见 S-M3-04/05 gap） |
| S-M3-03 | 编辑用户 / 删除用户（删除需确认；删除被依赖场景见 S-M3-12） | UF-4 | 正向 | ✓ | — | 通过 | `e2e/m3.spec.ts::REAL · S-M3-02/03 用户 create→edit→delete 全链路`（REAL admin：UserDetail 改 status → save-user → 返回 /users + 列表 ≤2s 刷新；delete-user → delete-confirm → 行 ≤2s 消失；自清理） |
| S-M3-04 | passkey 注册引导：管理员对目标用户发起 → 用户在有效期内完成 WebAuthn 绑定 → 登录方式标记 passkey | UF-4 | 正向 | ✓ | G-2 | 待执行 | 依赖 G-2（管理员发起 passkey 注册引导端点缺口，parity §3.1.1 处置已定：UI 降级/裁剪）；随 gap 处置收敛后补齐 |
| S-M3-05 | 注册会话过期/已使用 → 作废须重新引导；设备丢失重绑唯一路径 = 管理员撤销原绑定（G-3 候选承载 `identity/unbind` 待核验）后重新引导 | UF-4 | 边界 | ✓ | G-2、G-3 | 待执行 | 依赖 G-2/G-3（passkey 注册引导 + 撤销承载核验 gap）；随 gap 处置收敛后补齐 |
| S-M3-06 | 用户名非法（长度越界/字符集不符 `^[a-z][a-z0-9_-]{2,31}$`）→ 前端即时拦截，不发请求 | E10 / UF-4 | 负向 | — | — | 通过 | `e2e/m3.spec.ts::REAL · S-M3-06 用户名非法 → 前端即时拦截（E10）`（REAL admin：5 非法用例 [长度<3/>32/大写开头/数字开头/含空格] → `#create-username-error` 标红 + `aria-invalid` + 零 `/user/create` 请求） |
| S-M3-07 | 格式合法但用户名重复 → eiam 唯一性校验反馈就近渲染（FieldErrors） | E10 / UF-4 | 边界 | ✓ | — | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-07 用户名重复 → eiam 唯一性反馈标红（E10）`（CONTRACT：dev eiam 重复创建返回 `4010902` JIT 预配失败非干净唯一性错误；履约 validation 信封 → form-level `.form-error` 标红 + 回显原因 + 表单保留 + 列表不变。实现口径：唯一性渲染为 form-level 非 username 字段级，见回填说明） |
| S-M3-08 | 超界页码回落最后一页；空结果展示空状态 | E7 | 边界 | — | — | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-08 超界页码回落最后页 + 空页空态（E7）`（CONTRACT：2 用例 — 超界回落 ceil(21/20)=2 页 + 分页信息 2/2；total=0 → `.state-empty` + 无分页导航） |
| S-M3-09 | 删除末页唯一记录 → 回退上一页（不白屏不死页） | E7 | 边界 | — | — | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-09 删除末页唯一记录 → 回退上一页（E7）`（CONTRACT：末页 2/2 唯一记录删除 → store.remove + fetch 回落 page=1 + 分页信息 1/1，不白屏） |
| S-M3-10 | 组织树加载（eiam 实核为 department 树形，D-1）+ 新增子组织/编辑/删除 | UF-5 | 正向 | ✓ | D-1 | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-10 组织树加载 + 新增子组织/编辑/删除（UF-5）`（CONTRACT：`/api/department/*` 未经 nginx 反代至 eiam（acceptance host 限制），履约 department 树信封 → 树渲染 + 新增根组织 → 树 ≤2s 刷新显示新节点） |
| S-M3-11 | 组织名 1–64 字符校验；组织树 ≤ 1000 节点/深度 ≤ 8，超限禁止再建子级 | UF-5 | 边界 | — | — | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-11 组织名 1–64 字符校验（UF-5）`（CONTRACT：空名 < ORG_NAME_MIN=1 → `#create-org-name-error` 标红 + `aria-invalid` + `maxlength=64` 上限保护 + 零 `/department/create` 请求） |
| S-M3-12 | 删除有下属组织/关联用户的组织 → 弹窗列出阻断计数、确认按钮置灰 | E11 / UF-5 | 负向 | ✓ | — | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-12 删除有下属组织 → 阻断计数 + 确认置灰（E11/UF-5）`（CONTRACT：前端 precheckDelete [childCount=1/userCount=3] → `.delete-blocked` 列阻断计数 + `delete-org-confirm` disabled；详情面板列下属组织数/关联用户数） |
| S-M3-13 | 租户列表/新增/编辑（平台管理员可见） | UF-6 | 正向 | ✓ | — | 通过 | `e2e/m3.spec.ts::REAL · S-M3-13a 租户列表加载`（REAL admin：/tenants 加载真租户行「系统根管理空间」≤2s）+ `CONTRACT · S-M3-13b 新增租户`（CONTRACT：新增 → form-submit → list refetch ≤2s 显示新行） |
| S-M3-14 | 禁用有活跃会话的租户 → 请求被拒 → 展示 eiam 拒绝消息（G-4：无会话计数端点，消息是否含活跃会话数随处置核验；降级 = 禁用直接提交被拒） | UF-6 | 负向 | ✓ | G-4 | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-14 禁用租户被拒`（CONTRACT：tenant/update 履约拒绝 [G-4 降级：禁用直接提交被拒] → `[data-testid=disable-blocked]` 标红 + `disable-confirm` disabled + 列表不变） |
| S-M3-15 | 非平台管理员（租户管理员/只读审计）访问租户页 → 入口不渲染 + 直连 URL 被服务端拒绝（前端裁剪非安全边界） | UF-6 | 负向 | ✓ | — | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-15 非平台管理员直入 /tenants → 守卫拒`（CONTRACT：非 admin profile 直连 /tenants [requiresAdmin] → 守卫 ③ → /forbidden 不白屏；前端裁剪体验层，服务端授权由 eiam 强制） |
| S-M3-16 | 新增/编辑身份源（嵌套 ldap 载荷，D-2）→ 测试连接通过 → 保存（save/upsert 语义）→ 启停 | UF-7 | 正向 | ✓ | D-2 | 通过 | `e2e/m3.spec.ts::REAL · S-M3-16a 身份源列表加载`（REAL admin：/identity-sources 加载 ≤2s；v1 仅 ldap 入列，REAL eiam 仅 local 源 → 被过滤 → 空态为真行为）+ `CONTRACT · S-M3-16b 新增身份源`（CONTRACT：save + test 连接通过 `.test-result--ok` + 列表 ≤2s 刷新显示新源） |
| S-M3-17 | 测试连接失败 → 连接失败态错误提示，可修正后重测 | UF-7 | 负向 | ✓ | — | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-17 测试连接失败 → 错误提示可重测`（CONTRACT：identity_source/test 履约拒绝 → `.test-result--fail` 渲染 + 修正 URL 后重测 → `.test-result--ok`） |
| S-M3-18 | 连接参数校验（地址 `ldap(s)://` 格式/端口 1–65535/超时 1–30 为前端本地校验，D-2：eiam 无独立 port/timeout 字段）；Bind 密码掩码不回显（编辑态占位） | UF-7 | 边界 | — | D-2 | 通过 | `e2e/m3.spec.ts::CONTRACT · S-M3-18 连接参数前端本地校验`（CONTRACT：非法 URL [http://，非 ldap(s)://] → form-test 触发 validateForm → `#form-url-error` 标红 + `aria-invalid` + 零 `/identity_source/*` 请求；前端本地校验拦截） |

---

## 4. M4 — 权限管理三页 + 迁移收尾（证据回填：5.5）

> M4 出口判据：acceptance.md 全部场景通过；RC#1（5.4）前置门双签完成。本节通过即触发 §0.3 全矩阵复核与 §6 会签。
>
> **5.5 证据回填说明**：Playwright E2E 用例位于 `haven-console/e2e/m4.spec.ts`（15 用例），验收宿主为 nginx 同域形态（`:8888/console`，任务 2.10 haven-console.conf，生产构建产物 + CSP 注入），chromium 真实浏览器执行（Hard Rule 已守）。baseURL 指向 `http://localhost:8888`（eiam session cookie Domain=localhost）。≤2s 刷新以真实运行时序断言（`expect(...).toBeVisible({ timeout: 2000 })` 门槛 + `Date.now()` 差值 ≤2000ms 双重断言，超时即失败，不以缩短等待绕过；Hard Rule 已守）。
> 证据分两档（透明标注，与 M1/M2/M3 同口径）：
> - **REAL**（真实 eiam + 真实 nginx + 真实共享 cookie，零拦截）：`admin/12345678` 是 dev eiam seed 真实超级管理员。角色 CRUD 正向（create→list→delete 全链路）+ 策略列表读全程真链路。REAL 创建的测试角色用 `e2e_` 前缀 + 时间戳，每例 try/finally 自清理（delete），不残留 dev eiam。（2026-10 载荷校正：此前 RoleDetail 编辑步骤 REAL 不稳定、update 证据由 CONTRACT 01b 替代，根因为 `policy/list/attached/role` 载荷误发字符串 `type:"role"`——eiam `ListRolePoliciesReq.type` 为 uint8，字符串触发 gin 绑定 400；已修正为省略 type（0=全类型），REAL RoleDetail 编辑应恢复可稳定复现。）
> - **CONTRACT**（契约履约）：E11 依赖删除（需 eiam conflict 信封）、E6 并发 409/降级 validation、越权 403（守卫级 + 请求层）、策略 create/edit/delete（避免污染 dev eiam 平台级数据 + Statement schema 风险）、授权 CRUD（G-6 语义映射 + 词表）、角色绑定流（需可控 policy/user 列表）。经 `page.route` 履约 eiam 信封（形状取自 parity §3.2 实核 + api/*.ts 原生路径），浏览器 cookie 写入/守卫校验为真实链路（login 响应附 `Set-Cookie`，非注入伪造 cookie），仅 API 响应体按契约履约。≤2s 刷新断言走真实运行时序（route fulfill → store refetch → DOM render，全程经 nginx 同域链路）。
> **越权 403 双形态**（tech-design Security + router 守卫）：守卫级 403（`/policies` requiresAdmin）非 admin → `/console/forbidden`；请求层 403（`/roles` `/authorizations` 非 requiresAdmin）跨租户写 → iam 端点 HTTP 403 → inline `.state-error`/`.form-error`/`.delete-blocked` 渲染不跳转。
> **E6 并发编辑**（tech-design line 271）：eiam 实核无 409 支持（parity §6.4 待填，1.3 实核结论）；S-M4-11 履约 HTTP 409 验证 conflict 分支（`save.conflict` 文案），S-M4-12 履约 200+code 业务信封验证「无 409 时降级并入 validation」承接口径。
> **全量 acceptance 复核**（AC-4）：M1–M4 E2E 全量 61/61 通过（login.spec 13 + m2.spec 16 + m3.spec 17 + m4.spec 15）；E1–E12 每条 ≥1、UF-1~11 每个 ≥1；负向+边界占比 61%（36/59）≥1/3。gap-blocked 场景（S-M1-11[G-1]、S-M3-04/05[G-2/G-3]）维持「阻塞」（§0.1 documented handling，依赖缺口未决前不预标通过）；M2 E12 待执行项（S-M2-08/13/14/15/16）+ S-M1-13（RC#2）非本任务 Hard Rule 范围（「仅 M4 场景项 + 越权/E6」），随 E12 专项 / RC#2 补齐。

| ID | 场景项 | 来源 | 类型 | 涉 eiam | 依赖 | 状态 | 证据 |
|----|--------|------|------|---------|------|------|------|
| S-M4-01 | 角色 CRUD + 绑定策略/用户（D-3：角色↔策略绑定走 policy 侧 batch-attach/detach；用户绑定走 role batch_assign、以 code 为准） | UF-8 | 正向 | ✓ | D-3 | 通过 | `e2e/m4.spec.ts::REAL · S-M4-01a 角色 CRUD create→list→delete 全链路`（REAL admin：create e2e_ 角色 → 列表 ≤2s 刷新显示 → delete → 列表 ≤2s 刷新行消失，自清理）+ `CONTRACT · S-M4-01b 角色编辑 + 绑定策略流`（CONTRACT：RoleDetail 改名 + 勾选未绑策略 → save → role/update + policy/batch-attach 命中 + 返回列表；此前 update 证据由 CONTRACT 承载的原因是 list/attached/role 载荷 type 字符串 400 事故，已随 2026-10 载荷校正修复） |
| S-M4-02 | 角色名唯一 1–64 字符；策略绑定下拉仅列 eiam 返回的可绑定策略，无新建/编辑策略入口 | UF-8 | 边界 | ✓ | — | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-02 角色名校验 + 策略绑定下拉仅列可绑定策略`（2 用例：空名/>64 → `#create-name-error` 标红 + `aria-invalid` + 零 create 请求；RoleDetail 策略绑定区仅列 listBindablePolicies，无新建策略按钮） |
| S-M4-03 | 删除有用户绑定的角色被拒 → 提示绑定用户数并阻止 | E11 / UF-8 | 负向 | ✓ | — | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-03 删除有用户绑定的角色被拒（E11/UF-8）`（CONTRACT：role/delete 履约 409 信封「无法删除：该角色已绑定 3 个用户」→ `.delete-blocked` 列阻断计数 + `delete-confirm` disabled + 列表不变） |
| S-M4-04 | 租户管理员仅见本租户数据；越租户读写被服务端拒绝（403 原地收敛） | UF-8 | 负向 | ✓ | — | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-04 越租户写被 403 原地收敛（UF-8）`（CONTRACT：`/roles` 非 requiresAdmin，role/delete 履约 HTTP 403 → inline `.delete-blocked`/`.form-error` 渲染 + 不跳 `/forbidden` + URL 不变 + 列表不变；请求层 403 原地收敛口径） |
| S-M4-05 | 新增授权（G-6 语义映射：呈现为主体→策略 / 主体→角色两类绑定；选项词表见 S-M4-07）→ 撤销需确认 | UF-9 | 正向 | ✓ | G-6 | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-05 新增授权 + 撤销确认（G-6 语义映射）`（CONTRACT：selects 主体/资源/动作 → create → 列表 ≤2s 刷新显示新行；revoke → confirm → 撤销后列表 ≤2s 刷新行消失；动态 authList + batch_unassign 路由控制 phase） |
| S-M4-06 | 重复绑定 → 前端按已加载列表预检提示「授权已存在」+ eiam 服务端兜底拒绝，重复数据不入库 | UF-9 | 负向 | ✓ | G-6 | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-06 重复绑定前端预检（UF-9）`（CONTRACT：列表已有 {admin→viewer assign}；create 同三元组 → isDuplicateAuthorization 前端预检 → `.form-error`「授权已存在」+ 零 batch-attach/assign 请求 + 列表不变） |
| S-M4-07 | 资源/动作仅词表可选、不支持自由文本（G-5：静态词表降级时页面标注来源版本） | UF-9 | 边界 | ✓ | G-5 | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-07 资源/动作仅词表可选 + 来源标注（G-5）`（CONTRACT：`.vocab-source` 渲染「词表来源」；新增 dialog 主体/资源/动作三字段均为 `<select>`（tagName=SELECT，非自由文本 input）+ options 来自词表） |
| S-M4-08 | 新增/编辑策略（规范形 Statement[]、Effect ∈ {Allow, Deny}；D-5：eiam 实核单数键 statement/action/resource，前端归一）+ 删除 | UF-10 | 正向 | ✓ | D-5 | 通过 | `e2e/m4.spec.ts::REAL · S-M4-08a 策略列表加载`（REAL admin：/policies 列表加载 ≤2s，有数据则行渲染/无数据则空态）+ `CONTRACT · S-M4-08b 策略 create→edit→delete 全链路`（CONTRACT：Statement JSON 数组表单 create → 列表 ≤2s 刷新 → edit（Effect Allow→Deny）→ delete + 列表 ≤2s 刷新行消失） |
| S-M4-09 | 策略内容非法（JSON 语法错 / Effect 越词表 / 空 action、resource）→ 前端 JSON 校验拦截；eiam 服务端 schema 校验失败标红对应字段 | UF-10 | 负向 | ✓ | — | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-09 策略内容非法 → 前端 JSON 校验拦截（UF-10）`（CONTRACT：3 用例 — 语法错/Effect 非 Allow\|Deny/空 actions → `#edit-statements-error` 标红 + `aria-invalid` + 零 `/policy/create` 请求；前端 JSON 校验 parseStatementsJson 拦截） |
| S-M4-10 | 仅平台管理员可见/可编辑策略页（租户管理员与只读审计入口不渲染 + 直连被拒） | UF-10 | 边界 | ✓ | — | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-10 仅平台管理员可见策略页（UF-10）`（CONTRACT：非 admin profile 直连 `/policies`[requiresAdmin] → 守卫 ③ → `/console/forbidden` + 「无权限」文案；admin → `/policies` 可访问 + heading 渲染；守卫级 403 口径） |
| S-M4-11 | 同一对象并发编辑 → 后提交者生效；eiam 返回 409 → save.conflict 提示「数据已被他人修改，请刷新后重试」+ 自动重拉当前页 | E6 | 负向 | ✓ | §6.4 | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-11 并发编辑 409 → save.conflict 提示（E6）`（CONTRACT：role/update 履约 HTTP 409 + 信封 → `.form-error` 渲染「数据已被他人修改，请重新加载后再试」+ 表单保留（URL 不变 + 输入值保留）；conflict 分支验证） |
| S-M4-12 | eiam 无 409 支持 → conflict 分支降级并入 validation（承接 parity §6.4 核查结论，随 1.2/1.3 补充） | E6 | 边界 | ✓ | §6.4 | 通过 | `e2e/m4.spec.ts::CONTRACT · S-M4-12 eiam 无 409 → conflict 降级并入 validation（E6）`（CONTRACT：role/update 履约 200+code 业务信封（非 409）→ `.form-error` 渲染冲突文案 + 表单保留；降级承接口径——缺 409 时不吞错/不跳转，并入 validation inline 提示，承接 parity §6.4 核查结论） |

---

## 5. 依赖缺口/形状改判关联

> 依赖未决期间，挂靠场景状态记「阻塞」并标注依赖编号（§0.1）；依赖处置生效（parity §3.1 会签 / §6 补充章节产出）后解锁执行。
> **缺口处置决策（Phase 0 出口门，2026-09-30）**：G-1~G-9 处置已由 feature owner 定案（默认 UI 降级/裁剪，G-7 安全项例外需 eiam owner 排期确认），逐项记录见 parity §3.1.1；本表已同步补登记 G-7/G-8/G-9 三行。证据-O 会签位仍待签（会签留位不代签）；挂靠场景的降级执行口径以 parity §3.1.1 为准，由对应 E2E 宿主落。

| 依赖 | 内容（来源：parity-checklist） | 挂靠场景 |
|------|-------------------------------|----------|
| G-1 | `user/reset_password` 缺口（处置已定：UI 降级/裁剪，parity §3.1.1；证据-O 会签待签） | S-M1-11、S-M1-12 |
| G-2 | 管理员发起 passkey 注册引导缺口（处置已定：UI 降级/裁剪，parity §3.1.1） | S-M3-04、S-M3-05 |
| G-3 | passkey 撤销承载核验（`identity/unbind` 候选；处置已定：UI 降级/裁剪，parity §3.1.1） | S-M3-05 |
| G-4 | 租户禁用前置「未过期会话数 = 0」查询端点（UI 降级承接，处置已定 parity §3.1.1） | S-M3-14 |
| G-5 | 授权资源/动作词表枚举端点（静态词表降级承接，处置已定 parity §3.1.1） | S-M4-07 |
| G-6 | 授权三元组 → 二元绑定语义映射（处置已定：UI 语义映射，parity §3.1.1；证据-O 会签待签） | S-M4-05、S-M4-06 |
| G-7 | discovery/sync 无鉴权资产上报写面（**安全项例外**：需 eiam owner 排期确认，不走降级默认，parity §3.1.1/§6.2.5） | —（无控制台场景挂靠；部署面须确保 eiam :9000 不暴露于不可信网络） |
| G-8 | 首登强制改密承载机制缺口（处置已定：UI 降级，E1 按降级口径执行，parity §3.1.1；改密标记载体核查结论 = eiam 无载体，§6.3.5） | S-M1-11、S-M1-12 |
| G-9 | 锁定态列表可见性（处置已定：UI 降级——列表状态列仅启用/禁用/未知，锁定提示走登录页中性文案 B-2，parity §3.1.1） | —（无场景阻塞；UF-4 状态列按降级口径） |
| B-1~B-4 | 登录链路形状改判（`mfa_token` / 无 lockUntil / 无 source_id / `session_token`+header） | S-M1-02 ~ S-M1-04、S-M1-07 ~ S-M1-09 |
| D-1/D-2/D-3/D-5 | 形状分歧归一（department / 身份源嵌套载荷 / policy 侧绑定 / 策略单数键） | S-M3-10、S-M3-16、S-M3-18、S-M4-01、S-M4-08 |
| §6.4（parity） | 409 并发编辑支持性核查（1.2/1.3 产出） | S-M4-11、S-M4-12 |
| §6.1（parity） | 认证链路兼容矩阵（cookie 矩阵、tenant_id 写入断言，1.2 产出） | S-M1-01、S-M1-13、S-M2-01 ~ S-M2-05 |

---

## 6. 会签（执行期逐里程碑签收）

> 前端 owner + 测试 owner 双署；「涉 eiam」列 ✓ 的场景由 eiam owner 在对应里程碑行统一会签。M4 通过作为 RC#1（5.4）前置门输入时，另按 parity-checklist §7.2 名单执行平台管理员 + eiam owner 双签 + e-cam-web 前端 owner 会签。

| 里程碑 | 回填任务 | 前端 owner | 测试 owner | eiam owner（涉 eiam 场景） |
|--------|----------|------------|------------|----------------------------|
| M1 | 2.12 / 2.13 | ＿＿＿＿ | ＿＿＿＿ | ＿＿＿＿ |
| M2 | 3.4 | ＿＿＿＿ | ＿＿＿＿ | ＿＿＿＿ |
| M3 | 4.6 | ＿＿＿＿ | ＿＿＿＿ | ＿＿＿＿ |
| M4（全矩阵出口） | 5.5 | ＿＿＿＿ | ＿＿＿＿ | ＿＿＿＿ |

---

## 7. 增量交付验证记录（task 6.1 平台导航 / 6.2 图形化策略编辑器，2026-10-09）

> 宿主：nginx 同域形态（`:8888/console`，dist 生产构建 + CSP），REAL = 真实 eiam + 真实
> nginx + 真实共享 cookie，零请求拦截（仅观测）。脚本：
> `D:/Haven/scripts/verify-6-appnav-policyed-real.cjs`（16/16 PASS）。单测 580/580（41 文件）
> + vue-tsc 干净。两条 REAL 暴露并已修复的契约问题随此轮闭环：

| 场景 | UF | 口径 | 结果 | 证据 |
|------|----|------|------|------|
| 平台导航面板：触发开合 + 收藏列/控制台列/云管分区块渲染 | UF-2 | REAL | 通过 | verify-6「面板含收藏列/控制台列/云管分区块」「导航触发器渲染」 |
| 平台导航触发器位置（6.4）：header 左侧面包屑「首页」旁（#header-nav），非右侧 actions；文案「平台导航」 | UF-2 | REAL | 通过 | verify-6.4-nav-header-real.cjs 4/4（.header-left 中 .app-switcher 紧随 .header-breadcrumb；header-actions 无触发器；文案；开面板+Esc 关）；单测 MainLayout「平台导航挂载位置」 |
| ☆ 收藏切换 → ★ + localStorage 持久化 + reload 留存 + ✕ 移除 + 空态（单测） | UF-2 | REAL + 单测 | 通过 | verify-6「☆ 收藏切换→★+localStorage 持久化」「刷新后收藏持久化」；AppSwitcher.test.ts 10 用例 |
| 控制台入口站内跳转 → `/console/users`；云管深链 `window.open('/cam/compute/ecs','_blank','noopener')` | UF-2 | REAL | 通过 | verify-6「控制台入口站内跳转」「云管深链 window.open」 |
| 图形化编辑器：默认 tab + manifest 词表加载（132 动作）+ effect/actions/resources 真实 el-select 交互 → JSON 双向同步 | UF-10 | REAL | 通过 | verify-6「REAL manifest 词表 200」「动作词表已加载提示」「effect 切换 Deny」「actions 词表选择」「resources allow-create」「图形化→JSON 双向同步」；PolicyList.test.ts 6 用例 |
| **REAL 修复①**：`policy/create` 成功不回传 id（eiam policy/handler.go:201 `ginx.Result{Msg:"创建成功"}` 无 Data，实测 data:null）→ createPolicy 旧契约 `unwrapEnvelope<number>` 恒 null 致视图误判失败（对话框不关、实际已创建）| UF-10 | REAL | 修复 | createPolicy → void（code==0 即成功）；store.create → boolean；REAL 创建 200 + 对话框正常关闭 + 列表出现 |
| **REAL 修复②**：manifest Needs 双码但 dev admin profile 仅持 manifest（无 menus_by_urns）仍服务端 200（super 豁免）→ 客户端双码 gate 永远不发请求；改 manifest 单码 gate + 失败降级自由输入 | UF-10 | REAL | 修复 | 探测实录（admin 132 码含 manifest 无 menus_by_urns + GET manifest 200）；词表加载 REAL 断言 PASS |
| 策略真实创建 → 列表出现 → 真实删除自清理（列表移除） | UF-10 | REAL | 通过 | verify-6「REAL 创建 200 + 列表出现新策略」「REAL 删除自清理 + 列表移除」 |
| **REAL 修复③（6.3）**：组织接口直连 `/api/department/*` 未经反代 → nginx location / 301 → 「添加根组织」报「身份服务暂不可用」；六端点统一改 `/api/iam/department/*` | UF-5 | REAL | 修复 | records/6.3-org-prefix.md；verify-6.3-org-real.cjs 5/5 PASS（无错误态 + create 200 前缀实核 + 树出现 + delete 200 自清理）；单测路径前缀回归守卫 |

> S-M4-08b（CONTRACT create→edit→delete）与 S-M2-10（v1 应用切换器）自本轮起由 REAL
> 增量证据补强；AppSwitcher v2 面板取代 v1 下拉（任务记录 6.1/6.2/6.3，
> docs/features/platform-console/tasks/records/）。
