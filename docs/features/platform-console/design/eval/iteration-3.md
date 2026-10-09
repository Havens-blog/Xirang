# 技术设计评分报告 — 平台控制台（迭代 3，最终轮）

- **评审人**：Staff Architect（正式评分，迭代 3，最终轮）
- **评分对象**：`docs/features/platform-console/design/tech-design.md`（迭代 2 修订后）
- **对照输入**：`prd/prd-spec.md`、`prd/prd-user-stories.md`、`prd/prd-ui-functions.md`、`ui/ui-design.md`（文案契约 L543 实存）
- **注入事实核查（本迭代实盘复核）**：`e-cam-web`：`package.json`、`src/api/request/index.ts`、`src/api/request/eiam.ts`、`src/api/eiam-mapper.ts`（导出与全部引用方）、`src/api/eiam-users.ts`、`src/api/eiam-tenant.ts`、`src/api/types/eiam.ts`、`src/stores/user.ts`、`src/layouts/MainLayout.vue`、`src/components/PlatformNav/index.vue`、`src/router/routes.ts`、`src/assets/styles/*`、`nginx.dev.conf`、`vite.config.ts`；`eiam`：`internal/web/user/user.go`、`internal/web/policy/handler.go`、`internal/web/identity_source/handler.go`、`internal/web/department/handler.go`、`internal/domain/policy.go`、`internal/service/user/login.go`
- **总分**：**948 / 1000**（目标 900，**达标**）
- **关键门槛**：Breakdown-Readiness **172/180 ≥ 160** → **放行进入 `/breakdown-tasks`**

> 评分纪律：只按当前页面内容评分，不给「努力/改进」加分。迭代 2 指出缺陷在当前页面已存在的按当前内容回分；本迭代对修订新写入的全部实核断言逐条重验（**未发现新引入的错误**，见 Phase 1）。

---

## Phase 1 — Reasoning Audit（独立推理审计）

### 论证链追踪

1. **Problem → Solution**：成立。零后端 SPA 直调 eiam + 共享 Redis session 解决「宿主消失」且不自造鉴权；RC#1/RC#2 两条跨仓改动承接 In Scope，方案覆盖面与 PRD 承诺对齐。
2. **Solution → Evidence**：本迭代实核结论——**页面上所有可实核断言与两个仓库实况一致**：
   - A 档 13 端点与 `e-cam-web/src/api/eiam-users.ts`、`eiam-tenant.ts` 逐字吻合（`{total, users|roles|tenants}`、`batch_assign {usernames, role_codes}`、`tenant/batch_assign {user_ids, tenant_ids}` + M:N 禁止注释、`X-Active-Tenant-ID` 头 + 销毁重签 + reload 约定、角色 `desc` 非 `description`）。
   - **B 档 4 个登录端点实核全部真实存在**（`eiam/internal/web/user/user.go:66-80`：`system/login`、`ldap/login`、`passkey/login/start`、`passkey/login/finish`、`login/mfa/verify`）；锁定机制真实（`service/user/login.go` 的 `IsLocked`/`MaxFailedAttempts`/`LockoutDuration`）。
   - **RC#1 修正后 100% 准确**：`eiam-mapper.ts` 实况恰为「导出 `eiam/parseEiamStatus/mapEiamUser/mapEiamRole/diffSets`、不导出 `mapEiamProfile`」，全部引用方恰为 `eiam-users.ts` + `UserForm.vue` + 自测 `eiam-mapper.test.ts` 三处（grep 全仓验证）——判「迁移/删除 + 连测试文件一并删除 + 控制台侧重建」与实况及三选一分类自洽。
   - **RC#2 修正后 100% 准确**：`VITE_ECMDB_LOGIN_URL` 消费方全仓 grep **恰好三处**（`request/index.ts:24`、`stores/user.ts:91`、`MainLayout.vue:58`），与设计声明完全一致，无第四处遗漏；变量拆分（`VITE_LOGIN_TARGET` / `VITE_ECMDB_WEB_BASE`）使「登录目标」与「站点基址」两个互斥语义各归其主，回退（旧产物仍读旧变量）自洽。
   - 测试栈实核：`happy-dom ^20.11.6`、`@vue/test-utils ^2.4.6`、`vitest ^4.1.5` 均在 e-cam-web devDeps；`packageManager` 在现仓 manifest 确实缺位（grep 无命中）；`request/eiam.ts` 行为（timeout 15000 / withCredentials / cookie 注入 Bearer / 401 → redirectToLogin 且不清 cookie）与设计逐字一致；`isRedirectingToLogin` 防抖实存（`index.ts:16-17`）。
   - nginx 映射实核：`e-cam-web/nginx.dev.conf:27-28` `/api/iam/* → http://127.0.0.1:9000/api/`，与设计「/api/iam/* → eiam :9000 /api/*」一致。
   - C 档两个「推定」实核为真：`passkey/register/start` + `/finish` 真实存在（`user.go:92-93`）；`identity_source` 域真实存在（save/list/delete/detail/test/toggle）。`reset_password` 实核 **不存在** 于 eiam（全仓无 reset 路由）——设计「无法走 user/update 承载」（`UpdateEiamUserRequest` 无 password 字段，实核属实）+ C 档登记 + Phase 0 处置的写法与事实相符。
3. **Evidence → Success Criteria**：测试目标可量化（≥80%、E1–E12 + UF-1~11 逐条映射、负向+边界 ≥1/3、验收宿主锁 nginx 同域）；性能预算分解算术自洽（60+140+50+50=300）且三列（预算/手段/测量）齐备；CI 门禁挂钩。
4. **Self-contradiction check（SC/In-Scope 聚类 + 两两可满足性）**：
   - **持久化簇**：persistedstate 引入 ↔ persist 仅 theme（key `hc_theme`）↔「无本地数据持久化」——闭合，可同时满足。
   - **租户簇**：tenant/switch 销毁重签 ↔ 整页 reload ↔ 401 在途清理 ↔ 写前守卫快照（「非持久化、非第二份租户真相存储，仅页进入时刻的观测值」）——双向推导均通过；同标签页切租户必 reload，快照陈旧只能来自他标签切换，恰是守卫针对的形态，逻辑自洽。
   - **守卫范围簇**：Integration 6「写保证仅覆盖控制台的 7 个 list store」↔ 拓扑图 e-cam-web「读同一份」↔ 显式接受边界（三条理由 + UI 表现 + Phase 0 收窄条款）——迭代 2 的 B1 夸大已消除，页内声明自洽。
   - **RC#2 变量簇**：拆分后 `VITE_LOGIN_TARGET`（index.ts + user.ts）与 `VITE_ECMDB_WEB_BASE`（MainLayout）无交叉引用；「过渡期该基址不受旧登录入口 301 波及（301 仅覆盖登录 location）」与 nginx location 粒度一致。
   - **PRD In Scope 8 项**：1–7 项均有设计落点；第 8 项「产出**下线就绪清单**」仍只承接机制（移除 location / 个人设置跳转处置两处提到清单），**清单工件与下线判据本身无产出承诺**——部分承接（唯一遗留，见维度 5）。
   - **未发现逻辑互斥型矛盾**；未发现修订新引入的自相矛盾。
5. **本迭代新发现的事实缺口（页内当前状态，非修订错误）**：
   - eiam 实际域模型为 **department**（`internal/web/department/handler.go`：`GET /list` 部门树 + `assign/remove/members` 成员分配），设计 C 档推定的 `/api/iam/organization/*` 命名与实况不符（该信息在同工作区 eiam 仓当日可查，设计的实核范围止于 e-cam-web，未越界断言，故记为缺口而非错误）。
   - eiam 身份源域有 `toggle/:id` 启停端点（`identity_source/handler.go:47`），设计 C 档行未列启停端点，而 UF-7 明确要求「配置与启停」——UF-7 启停操作无接口锚。
   - eiam policy 域实况为 **Statement 结构**（`domain/policy.go`：`Effect` 首字母大写 `Allow/Deny`、`Action []string`、`Condition`），设计的 UF-10 Statement 数组模型与实况吻合，归一规则前瞻正确。
   - 发现归入维度 2/5/7 评分。

### Pre-score anchors（转入攻击）

- Phase 0 的证据**来源与执行方法**未定义（A 档有证据定义「源文件」，B/C 档只说「逐项核验后方为契约」）——结构性风险，转 blindspot B1。
- `LoginResult.lockUntil` 的降级路径（「若 eiam 不携带：降级为中性化固定文案」）单方面放弃了 E2/E9 的倒计时 AC，未走 PRD 缺口处置的 owner 书面确认门——记维度 5 AC 覆盖。

---

## Phase 2 — Rubric Scoring（验证姿态）

> 姿态：每条断言视为未证实，直至页内证据或实盘证据支撑。本迭代所有可实核断言均已实核（含 eiam 侧路由与域模型）。

### 1. Architecture Clarity — 166/170

| Criterion | Score | 依据 |
|-----------|-------|------|
| Layer placement explicit | 58/60 | 「单 SPA 分层，镜像 e-cam-web 目录约定」+ 单向数据流 `views → stores → api/eiamAxios → eiam`；职责边界显式：「api 层零租户状态」「租户上下文唯一归属 `stores/tenant.ts`」「stores 可持有分页/筛选 UI 态，api 层仅『参数进、类型出』纯调用」「persistedstate 仅挂 theme」；跨仓改动逐处标注归属层（请求层/router/layout/nginx）。扣 2：目录树职责仍是一行注释粒度；共享 components 层仅 3 个具名（11 页的表单弹窗/树/JSON 编辑器归入页面内部组成，可接受）。 |
| Component diagram present | 58/60 | 两张图均已实存：① SPA 内部组件关系图——路由守卫链 ①cookie/profile ②mustSelectTenant ③permissions ④租户快照、views(11)↔stores(10)↔api(9) 消费关系、共享组件复用点（TenantSelector 主布局+租户选择、StatusBadge 7 页共用、文案契约全部 11 页出口）；② 多方会话拓扑图——浏览器三标签 / nginx 四路由 + 301 / eiam ↔ 共享 Redis ↔ e-cam-service，SSO、双向登出、401 收敛、过渡期三方入图。扣 2：回切三件套的切换机制（env 双产物 + 软链状态机）仅文字描述，无图。 |
| Dependencies listed | 50/50 | 全栈具名且版本经实核（vue 3.5 / vite 7 / ts 5.9 / element-plus 2.10 / pinia 3 / persistedstate / sass / axios / vitest ^4.1.5 / happy-dom ^20.11.6 / @vue/test-utils ^2.4.6 均在现仓 manifest）；`packageManager` 锚定 + 「建仓时以 lockfile 锁死」覆盖 Playwright ^1.55；chromium 单矩阵已锁。无缺项。 |

### 2. Interface & Model Definitions — 150/170（db-schema=no 路径）

| Criterion | Score | 依据 |
|-----------|-------|------|
| Interface signatures typed | 54/60 | 迭代 2 的四处接口空洞**已在 C 档补齐且类型化**：密码重置 `{user_id}` → `{initial_password}`（附实核理由）、角色-策略绑定 `{role_id, policy_ids}` → `null`、角色已绑读取 `{role_id, offset, limit}` → `{total, users\|policies}`、passkey 注册引导 start/finish/撤销；授权词表端点已给响应类型 `{subjects, resources, actions}`；`Envelope<T>` / `ListQuery` / `Page<T>` / `LoginResult`（含 `lockUntil` 数据源与降级分支）全类型化；api 客户端具名签名 + `unwrapEnvelope<T>()` 单入口。扣 6：① 身份源「测试连接」响应仅「通断结果」散文描述，无类型（-2）；② **UF-7 启停端点缺席**——C 档行仅 `list\|create\|update\|delete` + test，UF-7 明确「配置与启停」，模型 `IdentitySource.enabled: boolean` 无写路径（eiam 实有 `toggle/:id`，当日可查）（-2）；③ **user↔organization 关联端点缺席**——UF-4「所属租户/组织」的组织半边无选择/分配锚（eiam department 有 assign/remove）（-2）。 |
| Models concrete | 54/60 | 全部模型字段具名带类型与约束：`ConnParams`（端口 1–65535、超时 1–30 默认 5、attrMap 用户名属性必填、password 写后不回显）、`PasskeyRegistrationSession`（15 分钟、单次使用）、`SessionClaims`、`User`（`'active'\|'disable'\|'unknown'` 归一 + `passkeyRegistered` + `mustChangePassword` 均标 Phase 0）、`Tenant`（实盘四字段）、`Role`（`desc` + policies/users refs 有端点来源）、`PolicyStatement[]`（`'Allow'\|'Deny'` 对齐 UF-10 与 eiam 域模型实况；扁平归一规则：扁平 ≡ 单 Statement 数组、无法表达多 Statement 时上报 validation 不静默截断）。扣 6：① `User` 无所属组织 ref（UF-4 表单「所属租户/组织」，组织半边悬空）（-3）；② `User.status` 词表缺「锁定」且未标 Phase 0——UF-4 状态枚举为「启用/禁用/**锁定**」，eiam 登录锁定（IsLocked）是否体现在 user/list 未处置（-3）。 |
| Directly implementable | 42/50 | A/B 档全链路（api 客户端、10 store、7 列表页、登录矩阵、租户分流/切换、401/403/降级、跨标签守卫）+ C 档补齐后的 UF-8 绑定流 / E1 重置流 / UF-4 passkey 引导流均可不猜类型编码——覆盖绝大部分工作量。扣 8：UF-4 组织关联选择、UF-7 启停操作、测试连接响应形状、UF-4 锁定状态展示四处仍需猜测或 Phase 0 兜底（UF-4/UF-7 均为 PRD 必需流）。 |

### 3. Error Handling — 125/130

| Criterion | Score | 依据 |
|-----------|-------|------|
| Error types defined | 42/45 | `ErrorKind` 六值联合、`ApiError extends Error { kind, code, httpStatus }`、`FieldErrors`、唯一解包入口 `unwrapEnvelope<T>()`（所有 api 函数必经）；「前端不假设 eiam 业务码区段；先 HTTP 状态、后信封 code」立场清醒。扣 3：业务码本体未在本文档枚举（契约 key 集合悬于 ui-design L543 与 Phase 0 对照表，无最小占位集）。 |
| Propagation strategy clear | 45/45 | 三态解包完备（2xx + code≠0 / 非 2xx 带信封 / 非 2xx 不可解析）；401「中断当前路由全部在途请求 + list store 复位 loading/error + `isRedirectingToLogin` 防抖继承现仓（实核 index.ts:16-17 存在）+ 不清共享 cookie」；403 原地 toast 不跳转；validation 字段级 aria-describedby 不弹全局；conflict 409 含「eiam 不支持则降级并入 validation」预留；unavailable「重试 ≥2 次 → 顶部降级提示 + 写操作禁用」承接 E12；unknown 兜底不白屏。无缺项。 |
| HTTP status codes mapped | 38/40 | 401/403/400/422/409（含降级）/5xx/15s 超时/网络错/unknown 全映射 + Phase 0「状态 × code → kind」对照表产出承诺。扣 2：429/限流类未入表；登录锁定（E2/E9，`lockUntil` 随失败响应返回）应归哪个 kind 未指明。 |

### 4. Testing Strategy — 128/130

| Criterion | Score | 依据 |
|-----------|-------|------|
| Per-layer test plan | 43/45 | 单元（stores/api 客户端/校验 utils/文案契约）+ 组件层（@vue/test-utils，范围收敛三组件并声明取舍理由：登录表单、文案契约渲染 code→文案逐字断言、StatusBadge；「页面级交互统一由 E2E 承接，避免双重维护」）+ E2E（E1–E12 与 UF-1~11 每条至少 1 映射、负向+边界 ≥1/3、逐条附证据）+ **验收宿主锁 nginx 同域形态、禁止独立端口 dev server 判定**（直击本项目真实历史坑）。扣 2：TenantSelector（租户切换 + reload，契约敏感）不在组件层测试范围，仅靠 E2E 间接覆盖；拷贝面 CI 漂移 job 自身的失败路径无测试声明。 |
| Coverage target numeric | 45/45 | 「覆盖 ≥80%」+ `pnpm typecheck`（vue-tsc）+ `pnpm build` 门禁。数值明确。 |
| Test tooling named | 40/40 | Vitest ^4.1.5（具名 DOM 环境 happy-dom ^20.11.6，理由充分：localStorage 持久化断言 + 与现仓测试环境一致、拷贝 utils 测试可复跑——版本实核在现仓 devDeps）、@vue/test-utils ^2.4.6（实核在现仓）、Playwright ^1.55 + chromium 单矩阵 + lockfile 锁死、vue-tsc、pnpm。无缺项。 |

### 5. Breakdown-Readiness ★ — 172/180（**≥160，门槛通过**）

| Criterion | Score | 依据 |
|-----------|-------|------|
| Components enumerable | 64/65 | 可数性达到全文最高水位：10 个 store（3 全局 + 7 页面级，统一 state 契约 `items/total/page/pageSize/loading/error` + `fetch/create/update/remove` + AbortController）、11 views、9 api 模块、共享组件 3、utils 3、e2e 目录、nginx 配置、`shared-hashes.json` 3 处拷贝面 × sha256 × 源 commit、RC#1 六处改动文件 + 引用盘点三选一落点（实核全对，含测试文件处置）、RC#2 三处文件 + 两个环境变量、回切三件套、CSP 表 6 指令、性能预算表。扣 1：页面内部组成（弹窗/树/JSON 编辑器）未枚举（可接受拆解产物）。 |
| Tasks derivable | 63/65 | 接口→任务、模型→任务主干完整：C 档补齐后 UF-8 策略绑定、E1 重置、UF-4 passkey 引导、角色详情读取**均可写出可验收的 API 任务**（迭代 2 的 -8 解除）；RC#1 清理任务含测试文件删除（迭代 2 漏项已闭合）；RC#2 三文件 + 变量拆分任务无噪音（eiam.ts 已从清单移除）。扣 2：UF-4 组织关联、UF-7 启停两处缺接口锚，对应 API 任务写不成可验收形态（-2）。 |
| PRD AC coverage | 45/50 | 六个 Story 全部 AC 有落点：登录矩阵/SSO/双向登出/403 原地/逐请求判定/只读裁剪/字段标红回显/凭据不泄露；性能 Goal 预算表 + 测量；监控（探活 + 回切演练计时入部署清单）；认证矩阵双签；E2/E9 倒计时数据源 `lockUntil` 已定义。扣 5：① 「下线就绪**清单**」工件无产出承诺——仅两处提到清单（「移除该 location 即完成『下线就绪』部署项」「个人设置跳转随『下线就绪清单』一并处置」），下线判据清单本体与判据项未列，PRD In Scope 第 8 项部分承接（-3）；② Story 1 AC4 / E2/E9：`lockUntil` 降级路径「降级为中性化固定文案（不含倒计时）」是单方面放弃 AC 必需的倒计时展示，未按 PRD Goals「每项『放弃』须业务 owner 书面确认并登记理由与 UI 降级」走缺口处置门（-1）；③ Story 4 AC3 授权缓存 ≤30s 语义仍未入 Phase 0 清单（-1）。 |

**门槛判定：172 ≥ 160 → 放行进入 `/breakdown-tasks`。** 稳健性：即使按最严口径（枚举 62 / 任务 60 / AC 44 = 166）仍高于门槛，判定不依赖临界分。

### 6. Security Considerations — 75/80

| Criterion | Score | 依据 |
|-----------|-------|------|
| Threat model present | 37/40 | 具名威胁清单完整：**token 窃取/XSS（具名机制链）**——「cookie 注入 Bearer ⇒ session cookie 必须可被 JS 读取（非 httpOnly），任何 XSS 可直接窃取平台凭证并冒用会话」，缓解五条配对（CSP / 零第三方 / 模板转义 + 禁 v-html / cookie 属性矩阵会签 / 会话固定重放由 eiam 签发销毁 + 租户切换即重签 + 不落 localStorage）；凭据枚举（统一文案 + 锁定提示中性化）；点击劫持（frame-ancestors none）；供应链（零第三方脚本）；跨域直连（connect-src 'self'）；cookie 属性错配（矩阵 + eiam owner 书面会签）；服务端强制授权为唯一安全边界（前端裁剪非边界，纳入核查矩阵）；CSRF 前提显式化（SameSite=Lax/Strict 前提列入 Phase 0 逐项确认，前提不成立则重议）。扣 3：会话 TTL/续期策略（Redis TTL、滑动 vs 绝对）未列入 Phase 0 矩阵——「前端不生成、不续期」意味着续期语义完全在 eiam 侧，矩阵却未核对；同账号多端并发会话策略未具名。 |
| Mitigations concrete | 38/40 | 威胁→缓解配对具体：CSP 指令基线表 6 行（default-src 'self' / script-src 禁 inline+eval / style-src 'unsafe-inline' 附 Element Plus 弹层理由与「实测可去除则收紧」/ connect-src 'self' / img-src / frame-ancestors+object-src 'none' + base-uri 'self'），起草人/复核人/双签齐全，Phase 0 按真实构建产物实测收敛；初始密码展示一次 + 首登强制改密；身份源凭据掩码不回显；写前守卫（快照取值机制已定义：路由守卫写模块级单值，非持久化非第二真相源）；双人复核回切 + staging 演练计时。扣 2：`lockUntil` 交付机制未定（data 独立字段 vs 结构化 msg，Phase 0 核查——降级已有但锁定 kind 归属随之悬空）；CSP style-src 'unsafe-inline' 是常设放宽项，仅承诺实测收敛无时间点。 |

### 7. Implementation Feasibility — 132/140

| Criterion | Score | 依据 |
|-----------|-------|------|
| Dependencies available | 48/50 | 逐项实核通过（见维度 1/4：vitest/happy-dom/@vue/test-utils/主题文件/axios 行为全在现仓）。扣 2：pnpm 与 Playwright 不在任何现有 manifest——新仓锚点自证覆盖 pnpm（packageManager），Playwright 依赖 lockfile 锁死声明成立但属工作区新工具链。 |
| Architecture fits project structure | 50/50 | 实核全部吻合：RC#1 六处改动文件 **6/6 与仓库一致**（`views/platform/users/index.vue`、`UserForm.vue`、`eiam-users.ts` 仅两处 import、`routes.ts:111`、`MainLayout.vue:331` 菜单项、`PlatformNav/index.vue` 恰好两处——L51 图标映射 + L141 导航项）；RC#2 三处 **3/3 行号精确**（`index.ts:24` 目标值唯一所在地、`user.ts:91` 登出目标、`MainLayout.vue:58` 个人设置基址，且全仓恰此三处消费）；`eiam-tenant.ts` 保留理由属实（TenantSelector 在用）；`/api/iam/* → :9000/api/*` 与 `nginx.dev.conf:27-28` 一致；env 拆分符合 build 期注入模式；「不跨仓 import、新仓重建归一函数」符合仓边界。无扣分项。 |
| Technical claims grounded | 34/40 | 实核通过项：A 档 13 端点、B 档 4 登录端点**在 eiam 实存**、passkey/register 实存、policy Statement 形状与 eiam 域模型一致（`Effect Allow/Deny` 首字母大写，设计「原值大小写随形状核查登记」覆盖大小写差异）、`reset_password` 实存性判断准确（eiam 无此路由 → C 档 + Phase 0 处置是正确姿态）、MFA 锁定机制实存（IsLocked/MaxFailedAttempts/LockoutDuration）、`:8888/cam` 前例、vite 直连缺代理坑。扣 6：① C 档组织端点推定 `/api/iam/organization/*` 与同工作区 eiam 实况不符——实况为 **department** 域（`GET /list` 部门树 + `assign/remove/members` 成员分配），且树读为 GET 非推定的 POST；该事实当日可查而实核范围止于 e-cam-web，导致 UF-5 锚点错名而非未验证（-3）；② UF-7 启停锚缺失而 eiam 实有 `toggle/:id`（-3）。 |

### Cross-dimension coherence check

- 「零后端、零数据库」↔ RC#1/RC#2 纯前端 + nginx 改动：一致。
- 持久化口径（persist 仅 theme）↔「无本地数据持久化」：闭合。
- 性能预算算术（60+140+50+50=300）↔ CI 门禁：一致。
- 验收宿主约束（nginx 同域、禁 dev server 判定）↔ Integration 5 部署形态：一致。
- RC#2 变量拆分 ↔ 回切三件套「不散落硬编码」：拆分后一致（迭代 2 的冲突已消除）。
- Integration 6 守卫范围声明 ↔ 拓扑图 e-cam-web「读同一份」：一致（显式接受边界覆盖残余）。
- Policy 模型（Statement[]）↔ UF-10 ↔ eiam 域模型实况：三者一致。
- `lockUntil` 降级路径 ↔ PRD E2/E9 倒计时 AC ↔ PRD Goals「放弃须 owner 书面确认」：**张力**（降级未走缺口处置门，已记维度 5）。
- 未发现其他维度间硬冲突。

---

## Phase 3 — Blindspot Hunt

### [blindspot] B1 — Phase 0 核验的证据来源与执行方法未定义，使「核验通过才升格为契约」不可执行

引用：「Phase 0 执行 PRD 的**端点差集盘点**：以 UF-1~11 逐条反推控制台所需端点全集，**与 eiam 实有端点求差集**……A 档 = 已实核（e-cam-web 生产代码逐字吻合，**证据 = 源文件**）；B/C 档 = 待核验假设（无现仓调用先例，Phase 0 逐项核验后方为契约）」。

推导：A 档的证据定义精确到「源文件」，但 B/C 档的「逐项核验」没有任何证据来源定义——eiam 的「实有端点」从何而来（读 eiam 路由注册代码？运行时探测？eiam owner 书面确认？）、由谁执行、evidence 何种形态可签收，页内零字。这不是假设性问题：本迭代实核表明 eiam 路由在同工作区当日可查（`internal/web/*/handler.go` 的 `g.POST/GET/DELETE` 注册 + `Capability` 授权注记连鉴权要求都可一并核出），而两轮修订下来 C 档仍保留了 organization（实为 department）、create|update（实为 save）、缺 toggle（实有 `toggle/:id`）三处当日可纠正的偏差——根因正是核验方法未定义，导致「Phase 0 再查」成为无限期推迟。Reasoning audit 独立标记；其数值影响已计入维度 2/7，此处作为结构性风险单列。**必须改进**：为 Phase 0 每类核查项定义证据来源与形态（端点存在性 = eiam 路由注册 文件：行号；形状 = 响应样本/域模型结构体；鉴权 = Capability 注记；行为语义 = owner 书面确认），指明执行人，并把「eiam 路由注册全量导出」列为 Phase 0 第一产出——C 档三处偏差当日即可被该方法捕获。

---

## 结论

迭代 3 修订在页面当前状态下达成双目标：**总分 948/1000 ≥ 900**，**Breakdown-Readiness 172/180 ≥ 160 放行**。迭代 2 报告的全部 4 类缺口（四处接口空洞、两处盘点实盘错误、Policy 形状分歧、B1/B2 blindspot）在当前页面均已真实闭合，且经本迭代对两个仓库的逐条实核**未发现修订新引入的错误**——RC#1 引用盘点、RC#2 三处文件行号、`VITE_ECMDB_LOGIN_URL` 三消费方、mapper 导出/引用方、B 档端点存在性、测试栈版本、nginx 映射全部与仓库实况一致。

剩余 52 分缺口（均已定位，不阻断拆解）：

1. **UF-7 启停端点缺席**（eiam 实有 `toggle/:id`）+ **UF-4 组织关联与锁定状态**（User 模型无 org ref、status ��表无锁定）——维度 2/5/7 的主要扣分源。
2. **下线就绪清单工件**无产出承诺（PRD In Scope 第 8 项部分承接）。
3. **Phase 0 证据来源/执行方法未定义**（blindspot B1）——`/breakdown-tasks` 执行前建议补一节。
4. 次要项：测试连接响应类型、429/锁定 kind 归属、授权缓存 ≤30s 入 Phase 0 清单、`lockUntil` 降级走 owner 确认门、会话 TTL 入矩阵。

**判定：可通过 `/breakdown-tasks` 进入任务拆解。**
