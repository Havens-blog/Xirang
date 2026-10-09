---
feature: "platform-console"
generated: "2026-09-30"
status: draft
---

# Technical Specifications: 平台控制台（统一入口 + 身份治理 + 权限管理 + 工作台）

> 抽取自 `design/tech-design.md`、`prd/prd-spec.md` §Other Notes。
> `[CROSS]` = 跨特性可复用约定；`[LOCAL]` = 仅本特性内有效。

## API 契约

### TECH-001: eiam 信封格式

**Requirement**: 所有 eiam 响应为 ginx 信封 `{ code: number; msg: string; data: T }`，成功 `code === 0`；字段是 `msg` 不是 `message`（实盘核验自 e-cam-web `src/api/eiam-mapper.ts`）。
**Scope**: [CROSS]
**Source**: tech-design §Interfaces 信封类型

### TECH-002: unwrapEnvelope 单入口 + ApiError 单一错误对象

**Requirement**: 所有 api 客户端函数必经 `unwrapEnvelope<T>()` 解包；视图层只 catch 唯一错误对象 `ApiError`（`kind: ErrorKind`, `code: number|null`, `httpStatus: number|null`）。两种错误形态统一解包：① HTTP 2xx + `code != 0` → 抛 ApiError；② HTTP 非 2xx 可解析为信封 → 取 `{code,msg}` 抛同一 ApiError，不可解析 → 按 HTTP 状态映射 kind。`ErrorKind = 'unauthorized' | 'forbidden' | 'validation' | 'conflict' | 'unavailable' | 'unknown'`。`FieldErrors = Record<string,string>`。前端不假设 eiam 业务码区段；先 HTTP 状态、后信封 code，产出「状态 × code → kind」对照表。
**Scope**: [CROSS]
**Source**: tech-design §Error Handling

### TECH-003: ListQuery / Page\<T> 归一类型

**Requirement**: 请求统一 `interface ListQuery { offset: number; limit: number; keyword: string }`；响应 api 层归一为 `interface Page<T> { total: number; items: T[] }`（把 eiam 的 `{total, users|roles|tenants}` 归一为 `items`）。每端点一个具名函数（如 `listUsers(q): Promise<Page<User>>`）。
**Scope**: [CROSS]
**Source**: tech-design §Interfaces 统一请求/归一类型

### TECH-004: eiamAxios 请求客户端模式

**Requirement**: axios 实例 `eiamAxios`：`withCredentials: true`、请求拦截从 cookie 读 token 注入 `Authorization: Bearer`、响应拦截 401 → `redirectToLogin()` 跳 `/console/login`（`isRedirectingToLogin` 防抖，不清共享 cookie）、15s 超时。
**Scope**: [CROSS]
**Source**: tech-design §Overview / §Architecture 组件关系图

### TECH-005: eiam API 契约基线登记

**Requirement**: Phase 0 差集盘点将 A/B/C 三档端点与 UF 反推全集逐项登记 `parity-checklist.md`（A 档核对差异；B/C 档按「待核验假设」核对形状与存在性，核验通过才升格为契约）。
**Scope**: [LOCAL]
**Source**: tech-design §Interfaces 契约基线

### TECH-006: 分页边界行为

**Requirement**: 超界页码回落最后一页；空页显示空状态；删除末页唯一记录后回退上一页。
**Scope**: [CROSS]
**Source**: prd-spec.md E7

## 错误处理

### TECH-007: 401 处理

**Requirement**: `unauthorized`(401) → `redirectToLogin('/console/login')`（不清共享 cookie）+ 中断当前路由全部在途请求 + list store 复位 loading/error。
**Scope**: [CROSS]
**Source**: tech-design §Error Handling 分类学表

### TECH-008: 403 原地 toast 不跳转

**Requirement**: `forbidden`(403) → 原地 toast（文案契约 `forbidden`），不跳转登录页。
**Scope**: [CROSS]
**Source**: tech-design §Error Handling 分类学表

### TECH-009: 400/422 字段级错误就近渲染

**Requirement**: `validation`(400/422) → `FieldErrors` 就近渲染（`aria-describedby` 关联），不弹全局 toast。
**Scope**: [CROSS]
**Source**: tech-design §Error Handling 分类学表

### TECH-010: 409 并发与依赖删除

**Requirement**: `conflict`(409) → E6 并发编辑提示「数据已被他人修改，请刷新后重试」+ 自动重拉当前页；E11 依赖删除被拒弹窗列阻断计数、确认置灰。Phase 0 核查 eiam 是否返回 409，不支持则该分支降级并入 `validation`。
**Scope**: [CROSS]
**Source**: tech-design §Error Handling 分类学表

### TECH-011: unavailable 降级

**Requirement**: `unavailable`(网络错 / 15s 超时 / 5xx) → 页面错误态 + 重试；重试 ≥ 2 次仍失败 → 顶部降级提示 + 写操作禁用（静态内容仍渲染，e-cam-web 会话不受影响）。
**Scope**: [CROSS]
**Source**: tech-design §Error Handling 分类学表

### TECH-012: 文案契约零二次映射

**Requirement**: `code != 0` 的业务码原值透传给文案契约 key（`code → 逐字文案`映射表见 ui-design）；`ApiError.message` 即契约文案，视图层零二次映射；字段级校验就近 `aria-describedby`。
**Scope**: [CROSS]
**Source**: tech-design §Error Handling 文案契约

## 前端状态

### TECH-013: persistedstate 仅挂 theme

**Requirement**: `persistedstate` 插件仅挂 `theme` 一个 store（key `hc_theme`）；任何会话 / 列表 store 不得声明 `persist`——杜绝插件默认全量写 localStorage；凭据不落 localStorage。
**Scope**: [CROSS]
**Source**: tech-design §Architecture Store 清单

### TECH-014: list store 统一契约

**Requirement**: 7 个 list store 统一契约：`state: { items, total, page, pageSize(默认 20、上限 100), loading, error }` + `fetch / create / update / remove`；每个 store 持有 `AbortController`，fetch 前中断在途请求。
**Scope**: [CROSS]
**Source**: tech-design §Architecture Store 清单

### TECH-015: 租户上下文唯一归属

**Requirement**: 租户上下文唯一归属 `stores/tenant.ts`；api 层零租户状态（仅「参数进、类型出」纯调用）；stores 可持有分页 / 筛选 UI 态。
**Scope**: [CROSS]
**Source**: tech-design §Architecture Store 清单 职责边界

### TECH-016: 切租户必走 tenant/switch + 整页 reload

**Requirement**: 所有切租户路径（UF-11 首选、UF-2 切换器）成功后必走 `POST /api/iam/tenant/switch` → 整页 `location.replace()`，杜绝 SPA 内只换 store 不刷页（e-cam-web TenantSelector 同款约定）。tenant_id 唯一存在于服务端共享 session，不引入 per-tab 快照、不做 BroadcastChannel 广播。
**Scope**: [CROSS]
**Source**: tech-design §Integration Specs 6

### TECH-017: 响应回读租户比对自愈

**Requirement**: list store 每次 fetch 成功后回读响应所属租户标识与 `currentTenantId` 比对（回读字段 Phase 0 核查；响应无租户标识则跳过此层），不一致 → 丢弃渲染并 `location.replace()` 自愈。
**Scope**: [CROSS]
**Source**: tech-design §Integration Specs 6

### TECH-018: 写前守卫（租户快照比对）

**Requirement**: create / update / remove 提交前比对「进入页面时的租户快照」与 `currentTenantId`，不一致先刷新再提交——阻断「标签 B 呈现 X 租户数据、发出 Y 租户请求」的写错租户形态。快照取值：路由守卫在每次路由进入 / 切换成功后把 user store 的 `currentTenantId` 记入模块级单值快照（非持久化、非第二份租户真相存储，仅页进入时刻观测值）。保证范围 = 控制台自身 7 个 list store。
**Scope**: [LOCAL]
**Source**: tech-design §Integration Specs 6 写前守卫

## 数据模型

### TECH-019: 用户状态归一词表

**Requirement**: eiam 原值 `'active' | 'disable'`，`ParseStatus` 未知串归一 `'unknown'`；写错会静默降级、启停操作失效。
**Scope**: [CROSS]
**Source**: tech-design §Data Models User.status

### TECH-020: 角色分配以 code 为准

**Requirement**: eiam 角色分配以 `code` 为准（实盘：`role/batch_assign` 载荷为 `{usernames, role_codes}`）。
**Scope**: [CROSS]
**Source**: tech-design §Interfaces A 档 role/batch_assign

### TECH-021: tenant/switch 销毁旧 session 重签 JWT

**Requirement**: `POST /api/iam/tenant/switch`（header `X-Active-Tenant-ID`）→ eiam 销毁旧 session 并重签 JWT，调用方必须随后整页 reload。
**Scope**: [CROSS]
**Source**: tech-design §Interfaces A 档 tenant/switch

### TECH-022: Policy Statement[] 归一规则

**Requirement**: 前端规范形策略 = `PolicyStatement[]`（一条策略可同时含 Allow 与 Deny）。归一规则：若实核 eiam 为「扁平单 effect」存储（顶层 effect + actions/resources），扁平结构 ≡ 恰含一条 Statement 的数组；扁平结构无法表达多 Statement 策略时，该形态上报 validation（不静默截断）。
**Scope**: [CROSS]
**Source**: tech-design §Data Models Policy / PolicyStatement

### TECH-023: 身份源 ConnParams 密码写后不回显

**Requirement**: 身份源连接参数 `ConnParams.password` 仅写入、响应掩码不回显；编辑态占位 `'******'`。`ConnParams` 六字段：`url`（ldap:// 或 ldaps://host:port）、`port`(1–65535)、`bindDn`(必填)、`password`、`baseDn`(必填)、`attrMap.username`(必填)、`timeoutSec`(1–30，默认 5)。
**Scope**: [CROSS]
**Source**: tech-design §Data Models ConnParams

### TECH-024: PasskeyRegistrationSession

**Requirement**: passkey 注册引导会话 `{ ticket, expiresAt, used }`：15 分钟有效、单次使用。
**Scope**: [CROSS]
**Source**: tech-design §Data Models PasskeyRegistrationSession

## 安全

### TECH-025: CSP 指令基线（nginx 注入）

**Requirement**: nginx 注入 CSP 覆盖全部 `/console` 路由：`default-src 'self'`；`script-src 'self'`（禁 inline/eval）；`style-src 'self' 'unsafe-inline'`（Element Plus 弹层运行时内联定位样式所需，实测可去除则收紧并记录）；`connect-src 'self'`；`img-src 'self' data:`；`frame-ancestors 'none'`；`object-src 'none'`；`base-uri 'self'`。起草人 = 平台管理员，复核 = eiam owner，随认证链路兼容矩阵双签；Phase 0 按真实构建产物实测收敛。
**Scope**: [CROSS]
**Source**: tech-design §Security CSP 指令基线

### TECH-026: token 窃取 / XSS 缓解配对

**Requirement**: 控制台继承 eiamAxios「从 cookie 读 token 注入 Bearer」⇒ session cookie 必须可被 JS 读取（非 httpOnly），任何 XSS 可直接窃取平台凭证。缓解配对：① CSP 基线收紧脚本与连接源；② 零第三方脚本 + 构建产物无内联脚本；③ Vue 模板默认转义，**禁用 `v-html` 渲染任何 eiam 返回内容**（文案契约、字段错误一律纯文本 / 组件化渲染）；④ cookie `Secure/SameSite/Path/Domain` 以 Phase 0 矩阵为准并经 eiam owner 会签；⑤ 会话固定 / 重放：凭据由 eiam 服务端签发与销毁（租户切换即销毁重签），前端不生成、不续期、凭据不落 localStorage。
**Scope**: [CROSS]
**Source**: tech-design §Security token 窃取 / XSS

### TECH-027: cookie 域严格一致

**Requirement**: cookie `domain / path` 严格与 eiam、e-cam-web 一致（SSO 及共享 session 前提）；`SameSite / Secure` 以 Phase 0 认证链路兼容矩阵为准（eiam owner 书面会签）。
**Scope**: [CROSS]
**Source**: tech-design §Security 会话 / §Architecture 多方会话拓扑

### TECH-028: 登录页安全面

**Requirement**: CSP / `X-Frame-Options` / HSTS 由 nginx 注入；CSRF 归属 eiam（前提：session cookie 按兼容矩阵携带 `SameSite=Lax/Strict`，跨站请求不携带凭据，前端无需 CSRF token 配合；前提不成立则重议）；零第三方脚本。登录页仅作 eiam 流程 UI 壳，凭据校验全在 eiam 侧。
**Scope**: [CROSS]
**Source**: prd-spec.md §Security Requirements；tech-design §Security 登录页安全面

## 性能

### TECH-029: 首屏静态资源 gzip ≤ 300 KB 预算

**Requirement**: 首屏静态资源 gzip ≤ 300 KB，预算分解：entry 自研 ≤ 60 + element-plus ≤ 140 + vue/router/pinia ≤ 50 + 其余（css / 字体 / 图标）≤ 50（单位 KB）。手段：11 页路由级全 `import()` 懒加载；Element Plus unplugin 按需 + 图标按需；`manualChunks` 三分包；构建期预生成 `.gz` + nginx `gzip_static on`。测量：构建脚本汇总 `dist/**/*.gz` 体积，CI 超 300 KB 即 fail。
**Scope**: [CROSS]
**Source**: tech-design §性能预算

### TECH-030: LCP ≤ 2.5s / 二次 ≤ 1s

**Requirement**: 同域内网 LCP ≤ 2.5s（二次访问 ≤ 1s）。手段：首屏仅 entry + workbench chunk + element-plus 核心 css；静态资源 `Cache-Control: max-age=31536000, immutable`（content-hash 文件名），`index.html` no-cache。测量：Playwright trace + Lighthouse，登录页与工作台两个 LCP 采样点，同域内网基线。
**Scope**: [CROSS]
**Source**: tech-design §性能预算

### TECH-031: 静态站可 CDN 但须同登录域

**Requirement**: 纯静态前端站可 CDN 加速，但须与 e-cam-web / eiam 同登录域；跨域 CDN 直调 eiam 不支持（SameSite / CORS 限制）。仅当未来变更部署形态并专项核对 CORS / SameSite 后才可重议。
**Scope**: [CROSS]
**Source**: prd-spec.md §Performance Requirements

## 测试

### TECH-032: 单元测试 Vitest + happy-dom

**Requirement**: Vitest ^4.1.5（DOM 环境具名 happy-dom ^20.11.6，与 e-cam-web devDeps 同版本）；覆盖 ≥ 80%；`pnpm typecheck`（vue-tsc）+ `pnpm build` 通过。选 happy-dom 而非 jsdom：theme store 的 `localStorage` 持久化断言与 cookie utils 需真实 DOM 存储实现，且与现仓测试环境一致、拷贝来的 utils 测试可直接复跑。
**Scope**: [CROSS]
**Source**: tech-design §Testing

### TECH-033: 组件测试范围收敛

**Requirement**: @vue/test-utils 范围收敛到三个契约敏感组件——登录表单（校验规则 / 提交 / 锁定倒计时态）、文案契约渲染组件（code → 文案逐字断言，防二次映射回归）、StatusBadge（状态归一词表渲染）。页面级交互不写组件测试、统一由 E2E 承接，避免同一路径双重维护。
**Scope**: [CROSS]
**Source**: tech-design §Testing 组件层

### TECH-034: Playwright E2E 硬性

**Requirement**: Playwright ^1.55（真实浏览器 E2E，PRD 硬性要求；浏览器矩阵仅 chromium，与内网统一浏览器一致）。覆盖登录 / 登出、SSO 跳转、双向单点登出、身份 / 权限 CRUD 的正向 + 负向 + 边界。验收场景映射 `acceptance.md` 的 E1–E12 与 UF-1~11（每条至少 1 个映射，负向 + 边界 ≥ 1/3），逐条附 Playwright / MCP 执行证据。
**Scope**: [CROSS]
**Source**: tech-design §Testing 浏览器 E2E

### TECH-035: 验收宿主必同域 nginx 反代

**Requirement**: 验收宿主必为最终同域 nginx 反代形态（沿用 `:8888/cam` 前例），**不得**在独立端口 dev server 上判定通过（vite 直连缺 `/api/iam` 代理 + 无共享 cookie）。
**Scope**: [CROSS]
**Source**: tech-design §Testing 验收宿主

## 部署与运维

### TECH-036: 回切开关三件套

**Requirement**: 回切开关为三件套，全部无 rebuild 可执行：① nginx `/console` location 注释行切换（或 upstream 双池）+ `nginx -s reload`（秒级）；② e-cam-web 401 目标以 build 期 env 注入、新旧双产物并存，切换 = 换静态目录软链；③ 过渡期旧登录入口 301 同挂 nginx location。上线前在 staging 完整演练一次回切并计时（PRD 承诺 10 分钟内），演练记录（步骤 / 耗时 / 回退验证）入部署清单。执行人：平台管理员 + eiam owner 双人复核。
**Scope**: [CROSS]
**Source**: tech-design §Integration Specs 5

### TECH-037: 探活健康检查与告警接入

**Requirement**: 控制台 `/console` 静态资源探活（HTTP 200 健康检查）接入告警，与 e-cam-web 同一 nginx 运维面、同一告警渠道，不新增独立 SLA。
**Scope**: [CROSS]
**Source**: prd-spec.md §Monitoring Requirements；proposal §NFR 可用性

## 跨仓拷贝漂移防线

### TECH-038: shared-hashes.json 登记拷贝面

**Requirement**: 新仓 `shared-hashes.json` 逐条登记拷贝文件 × sha256 × 源仓 commit，覆盖 `theme-variables.scss`、`element-theme.scss`、`api/request/eiam.ts`（eiamAxios 模式文件）三处拷贝面。
**Scope**: [CROSS]
**Source**: tech-design §拷贝面漂移防线

### TECH-039: CI sha256 比对防漂移

**Requirement**: CI job 从 e-cam-web checkout 对应源文件重算 sha256 比对，不一致 → 红；恢复绿必须同步拷贝 + 更新 hash + 跑一条主题冒烟 E2E，漂移来源 commit 记入变更记录。
**Scope**: [CROSS]
**Source**: tech-design §拷贝面漂移防线 CI 防线

### TECH-040: 拷贝面升级流程

**Requirement**: e-cam-web 侧改动主题 / 请求层文件 → 控制台 CI 当日变红 → 人工评估差异 → 同步拷贝 + hash 更新 + 主题冒烟回归。此防线覆盖 PRD ≥ 5 工作日通知条款未触及的 e-cam-web 侧文件（该条款只覆盖 eiam 端点）。
**Scope**: [CROSS]
**Source**: tech-design §拷贝面漂移防线 升级流程

## 技术栈

### TECH-041: 技术栈锁同版本与 e-cam-web 同源

**Requirement**: Vue 3.5 + Vite 7 + TypeScript 5.9 + Vue Router 4.5 + Pinia 3（persistedstate）+ Element Plus 2.10（锁与 e-cam-web 同版本）+ @element-plus/icons-vue + unplugin-vue-components / auto-import 按需自动引入 + Sass。v1 采用「拷贝 + 锁同版本」共享机制。
**Scope**: [CROSS]
**Source**: tech-design §Overview

### TECH-042: packageManager 字段锚定

**Requirement**: pnpm ≥ 9，新仓 `package.json` 写入 `packageManager` 字段锚定精确版本（e-cam-web 现有 manifest 未声明 packageManager，控制台新仓为首个显式锚点；建仓时以 lockfile 锁死）。
**Scope**: [CROSS]
**Source**: tech-design §Overview 包管理
