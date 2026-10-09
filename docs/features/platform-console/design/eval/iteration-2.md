# 技术设计评分报告 — 平台控制台（迭代 2）

- **评审人**：Staff Architect（正式评分，迭代 2）
- **评分对象**：`docs/features/platform-console/design/tech-design.md`
- **对照输入**：`prd/prd-spec.md`、`prd/prd-user-stories.md`、`prd/prd-ui-functions.md`、`ui/ui-design.md`（文案契约 L543 实存）
- **注入事实核查（本迭代实盘复核）**：`e-cam-web/package.json`、`src/api/request/eiam.ts`、`src/api/request/index.ts`、`src/api/eiam-users.ts`、`src/api/eiam-tenant.ts`、`src/api/eiam-mapper.ts`、`src/api/types/eiam.ts`、`src/stores/user.ts`、`src/stores/user-mapper.ts`、`src/layouts/MainLayout.vue`、`src/components/PlatformNav/index.vue`、`src/router/routes.ts`、`src/assets/styles/theme-variables.scss` / `element-theme.scss`
- **总分**：**875 / 1000**（目标 900，未达标；迭代 1 为 721）
- **关键门槛**：Breakdown-Readiness **159/180 < 160** → **仍阻断进入 `/breakdown-tasks`**（差距 1 分，缺口清单见维度 5，修复面极小且明确）

> 评分纪律：只按当前页面内容评分，不给「努力/改进」加分。迭代 1 缺陷在当前页面确已解决的，按当前内容回分；修订**新引入**的错误按当前内容扣分（本迭代实核出 3 处，见维度 7 与 Phase 1）。

---

## Phase 1 — Reasoning Audit（独立推理审计）

### 论证链追踪

1. **Problem → Solution**：成立。零后端 SPA 直调 eiam + 共享 Redis session，确实解决「宿主消失」且不自造鉴权。本迭代把两条 In Scope 跨仓改动（RC#1/RC#2）纳入设计后，方案覆盖面与 PRD 承诺对齐。
2. **Solution → Evidence**：本迭代最强的部分。A 档 13 个端点的请求/响应形状**逐条与仓库实况吻合**（本人逐一实核：`eiam-users.ts` 的 `{total, users|roles|tenants}`、`batch_assign {usernames, role_codes}`、`tenant/batch_assign {user_ids, tenant_ids}` 与 M:N 禁止注释、`eiam-tenant.ts` 的 `X-Active-Tenant-ID` 头 + 「销毁旧 session 并重签 JWT、调用方须 reload」、`stores/user.ts` 的 ginx 信封 `{code, msg, data}`、`eiam-mapper.ts` 的角色 `desc` 非 `description`、`parseEiamStatus` 的 `'active'|'disable'` 归一 `'unknown'`、`user-mapper.ts` 的 profile 六字段）。B/C 档诚实标注「推定 + Phase 0 核查」。证据姿态合格。
3. **Evidence → Success Criteria**：测试目标可量化（≥80%、E1–E12 + UF-1~11 逐条映射、负向+边界 ≥1/3、验收宿主锁 nginx 同域、禁止独立端口 dev server 判定）；性能预算有分解、手段、测量三列且算术自洽（60+140+50+50=300）。
4. **Self-contradiction check（SC/In-Scope 聚类 + 两两可满足性）**：
   - **持久化簇**：Overview 引入 persistedstate ↔ Architecture「persist 仅 `theme`（key `hc_theme`）」↔「无本地数据持久化（除 cookie 会话与 theme localStorage）」——已通过显式口径闭合，可同时满足。
   - **租户簇**：A 档 tenant/switch「销毁旧 session 并重签 JWT」↔ Integration 6「整页 `location.replace()`」↔ 401 kind「中断在途 + store 复位」——可满足（新 JWT 经共享 cookie 被其他标签逐请求读取，读旧渲染接受 staleness、写有守卫）。cookie 重签如何到达客户端（Set-Cookie vs 前端写 cookie）标 Phase 0 矩阵，页内已声明，不算矛盾。
   - **跨标签一致性**：Integration 6「不引入 per-tab 快照」↔ 写前守卫「进入页面时的租户快照」——措辞有张力但非互斥（后者是页进入时刻的 UI 态观测值，非租户真相第二份存储），记轻度，不扣分。
   - **PRD In Scope 8 项聚类**：1–7 项均有设计落点；第 8 项「产出**下线就绪清单**」只承接了机制（「移除该 location 即完成『下线就绪』部署项」），**清单工件本身（下线判据文档）无产出承诺**——部分承接。
   - **未发现逻辑互斥型矛盾**（双向推导均通过）。
5. **修订新引入的事实错误（本迭代实核）**：
   - **错误 A（RC#1 引用盘点）**：设计断言「`eiam-mapper.ts` 保留——`stores/user.ts` 的 `mapEiamProfile` 在用」。实况：`stores/user.ts` 第 4 行 `import { mapEiamProfile } from './user-mapper'`——该函数在 **`stores/user-mapper.ts`**，`api/eiam-mapper.ts` **根本不导出 `mapEiamProfile`**（仅导出 `eiam/parseEiamStatus/mapEiamUser/mapEiamRole/diffSets`）。RC#1 删除 `eiam-users.ts` 与 `UserForm.vue` 后，`api/eiam-mapper.ts` 的唯一剩余引用方是它自己的测试文件——按设计自己的三选一分类应判「迁移/删除」，却以假依据判了「保留」。这是把 Phase 0 产出预写成断言、且未经验证的直接后果。
   - **错误 B（RC#2 改动文件清单）**：设计列 `src/api/request/eiam.ts` 为「`redirectToLogin()` 目标值」改动文件——实况 `eiam.ts` 只 import 并调用 `redirectToLogin()`，**目标值只存在于 `index.ts:24`**（`VITE_ECMDB_LOGIN_URL || '/login'`），eiam.ts 无需任何改动。清单噪音会变成无意义任务。
   - **错误 C（RC#2 隐式耦合未识别）**：`VITE_ECMDB_LOGIN_URL` 在 e-cam-web 有**三处**用途：① 401 目标（`request/index.ts:24`）、② 登出跳转（`stores/user.ts:91`）、③ **「个人设置」跳转基址**（`MainLayout.vue:58`，拼 `${base}/profile/index` 打开 ecmdb-web 个人中心）。把该变量默认值改为 `/console/login` 会把「个人设置」跳转打成 `/console/login/profile/index` 死链；`MainLayout.vue` 不在 RC#2 清单中，设计对这一耦合零字提及——正是「模块间隐式耦合未在接口层声明」的教科书案例。
   - 发现 A/B/C 归入维度 2/5/7 评分。

### Pre-score anchors（转入攻击）

- 设计把 Phase 0 从 PRD 的「端点差集**盘点**」收窄为「差异核查（**不再从零盘点**）」——收窄本身使设计自己 C 档枚举的遗漏（密码重置、角色-策略绑定、角色已绑清单读取、passkey 注册引导发起/撤销）**失去任何兜底网**，直到 M3/M4 才暴露。
- UF-4「锁定状态」与 UF-8「策略绑定」在 eiam 侧是否经 Authorization 三元组实现，设计未调和（Policy 扁平形状 vs UF-10 Statement 数组）。

---

## Phase 2 — Rubric Scoring（验证姿态）

> 姿态：每条断言视为未证实，直至页内证据或实盘证据支撑。本迭代所有可实核断言均已实核。

### 1. Architecture Clarity — 153/170

| Criterion | Score | 依据 |
|-----------|-------|------|
| Layer placement explicit | 56/60 | 「单 SPA 分层，镜像 e-cam-web 目录约定」+ 单向数据流；本迭代补齐职责边界：「api 层零租户状态」「租户上下文唯一归属 `stores/tenant.ts`」「stores 可持有分页/筛选 UI 态，api 层仅『参数进、类型出』」「persistedstate 仅挂 theme」。扣 4：目录树职责仍是一行注释粒度；共享 components 层粒度粗（仅 3 个具名）。 |
| Component diagram present | 52/60 | **多方会话拓扑图已补**：浏览器三标签 / nginx 四路由 + 301 / eiam ↔ 共享 Redis ↔ e-cam-service，SSO、双向单点登出、401 收敛、cookie 域、过渡期三方全部入图——系统最高风险面已有组件关系图。扣 8：SPA **内部**组件关系（哪些 view 消费哪个 store、路由守卫流、共享组件被哪些页复用）仍只有一行数据流文字 + 目录树，图只覆盖部署/会话拓扑。 |
| Dependencies listed | 45/50 | 全栈具名且版本经实核（vue 3.5.22 / vite 7.1.7 / ts 5.9.3 / element-plus 2.10.5 / pinia 3.0.1 / persistedstate 4.2.0 / sass / axios / vitest 4.1.5 / vue-tsc / unplugin 全在 manifest）；`packageManager` 缺位以「控制台新仓为首个显式锚点」自证；Playwright ^1.55 + 仅 chromium 矩阵已锁。扣 5：单元测试 DOM 环境未具名（theme store 持久化断言必需 localStorage 环境，e-cam-web devDeps 里现成的 happy-dom ^20.11.6 未采用）；@vue/test-utils 在现仓存在而设计不用亦未声明取舍。 |

### 2. Interface & Model Definitions — 128/170（db-schema=no 路径）

| Criterion | Score | 依据 |
|-----------|-------|------|
| Interface signatures typed | 46/60 | 迭代 1 的「显式弃权」已消除：A 档 13 端点逐条请求→响应类型（实核逐字吻合），`Envelope<T>` / `ListQuery` / `Page<T>` 类型化，api 客户端具名签名（`listUsers(q: ListQuery): Promise<Page<User>>`、`switchTenant(id: number): Promise<void>`），`unwrapEnvelope<T>()` 单入口。B 档 4 端点类型化（`LoginResult`）。扣 14——接口空洞（均为控制台新依赖、按设计自己的分类应入 C 档却缺席）：① **密码重置端点**（E1/UF-4「管理员重置→一次性初始密码展示一次」；实核 `UpdateEiamUserRequest` 无 password 字段，A 档无此能力）；② **角色-策略绑定端点**（UF-8「绑定策略」核心流；C 档 policy 仅 CRUD，`Role.policies: Ref[]` 无任何端点来源）；③ 角色已绑用户/策略**读取**端点（角色详情页数据源；A 档 `role/list/attached/user` 是 user→roles 方向）；④ passkey 注册引导**发起/撤销**端点（模型 `PasskeyRegistrationSession` 在、端点不在，仅「机制本身列 Phase 0 核查」）；⑤ C 档授权词表枚举端点无响应类型。 |
| Models concrete | 46/60 | 12 个模型全部字段具名带类型与约束（`ConnParams` 含端口 1–65535/超时 1–30 默认 5、`SessionClaims`、`User` 含 `'active'|'disable'|'unknown'` 归一语义与注释、`Tenant` 对齐实盘 EiamTenant 四字段）。扣 14：① `Policy` 扁平单 effect 结构（`effect: 'allow'|'deny'` + 顶层 actions/resources）与 UF-10「**Statement 数组**：Effect ∈ {Allow, Deny}、Action…、Resource…、Condition 可选」**形状分歧未声明、未标 Phase 0**（Statement 数组无法无损坍缩为单 effect 扁平结构）；`PolicyCondition` 的 operator 词表标了 Phase 0 但整个 Statement-vs-flat 问题没标；② `User` 缺所属组织 ref（UF-4 表单「所属租户/组织」）与「锁定」状态口径（UF-4 状态含锁定，eiam 实况无此值，未标注差异处置）；③ `Role.policies/users` 两个字段无端点来源（同上接口空洞）；④ E2/E9「提示剩余解锁时间」无数据来源字段（`LoginResult` 无 lockUntil/剩余时间，锁定文案要倒计时但没有时间源）。 |
| Directly implementable | 36/50 | A/B 档全链路（api 客户端、store、users/tenants/roles 列表页、登录/登出/租户切换、401/403/降级、跨标签守卫）可不猜类型直接编码——这已覆盖大部分工作量。扣 14：UF-8 策略绑定流、E1 重置流、UF-4 passkey 引导流、授权词表四处在「不发明端点就无法实现」状态；且 Phase 0 已被设计自己收窄为「差异核查、不再从零盘点」，这些空洞不在任何核查清单上。 |

### 3. Error Handling — 124/130

| Criterion | Score | 依据 |
|-----------|-------|------|
| Error types defined | 42/45 | `ErrorKind` 六值联合、`ApiError extends Error { kind, code, httpStatus }`、`FieldErrors`、唯一解包入口 `unwrapEnvelope<T>()`（所有 api 函数必经）。「前端不假设 eiam 业务码区段；先 HTTP 状态、后信封 code」是清醒立场。扣 3：业务码本体仍未枚举（契约 key 集合悬于 ui-design 与 Phase 0 矩阵，设计侧无最小集占位）。 |
| Propagation strategy clear | 44/45 | 两种错误形态（2xx + code≠0 / 非 2xx 带信封 / 非 2xx 不可解析）统一解包三态完备；401「中断当前路由全部在途请求、list store 复位 loading/error」补上了迭代 1 缺口；409「eiam 不支持则降级并入 validation」预留分支；unavailable「重试 ≥2 次仍失败 → 顶部降级 + 写操作禁用」承接 E12；字段级 aria-describedby 不弹全局 toast；文案契约零二次映射。扣 1：401 防重复跳转（现仓 `isRedirectingToLogin` 防抖）在控制台侧未继承声明。 |
| HTTP status codes mapped | 38/40 | 401/403/400/422/**409（含降级）**/5xx/15s 超时/网络错/unknown 全映射 + Phase 0「状态 × code → kind」对照表产出承诺。扣 2：429/限流类未入表（登录锁定走业务码还是状态码未定，E2/E9 的 kind 归属悬空）。 |

### 4. Testing Strategy — 121/130

| Criterion | Score | 依据 |
|-----------|-------|------|
| Per-layer test plan | 40/45 | 单元（stores/api/校验 utils/文案契约）+ 真实浏览器 E2E（E1–E12 与 UF-1~11 每条至少 1 映射、负向+边界 ≥1/3、逐条附 Playwright/MCP 证据）+ typecheck/build 门禁 + **验收宿主锁 nginx 同域形态、禁止独立端口 dev server 判定**（直击本项目真实历史坑，仍是全文最清醒约束之一）。扣 5：组件级测试层仍缺席（登录表单、文案契约渲染组件无验证路径；@vue/test-utils 在现仓 devDeps 现成而设计不用）。 |
| Coverage target numeric | 45/45 | 「覆盖 ≥80%」，数值明确。 |
| Test tooling named | 36/40 | Vitest ^4.1.5、Playwright ^1.55、vue-tsc、pnpm、浏览器矩阵仅 chromium（迭代 1 缺口已闭合）。扣 4：DOM 测试环境仍未指定（happy-dom/jsdom 二选一未落）；Playwright 在现仓无先例，版本仅 `^1.55` 无 lockfile 锚定机制说明。 |

### 5. Breakdown-Readiness ★ — 159/180（**低于 160，仍阻断**）

| Criterion | Score | 依据 |
|-----------|-------|------|
| Components enumerable | 62/65 | 可数性已达到设计文档的高水位：10 个 store（3 全局具名 + 7 页面级具名，含统一 state 契约）、11 views（与 ui-design 11 页一一对应）、9 个 api 模块具名、共享组件 3 个、utils 3 项、e2e 目录、nginx 配置、`shared-hashes.json` 3 处拷贝面、RC#1 六处改动文件 + 三项引用盘点、RC#2 三处、回切三件套。扣 3：共享组件粒度粗（11 页的表单弹窗/树/JSON 编辑器/详情抽屉未列，属可接受拆解产物）。 |
| Tasks derivable | 54/65 | 每接口→任务、每模型→任务的主干成立：脚手架+`packageManager` 锚定、主题拷贝×2+hash 登记+CI 漂移 job+主题冒烟 E2E、请求层+解包+错误分类学、A/B 档逐端点客户端、10 store、路由守卫、11 页、跨标签三机制、RC#1 分 commit 删除+回归 smoke+revert 预案、RC#2 env 注入+双产物+软链+Playwright 断言、nginx 三 location、性能预算构建任务、E2E 矩阵、staging 演练。扣 11：① 四处接口空洞使 UF-8 策略绑定、E1 重置、UF-4 passkey 引导、角色详情读取**写不出可验收的 API 任务**（-8）；② RC#1 引用盘点一条假依据（mapper）产生一个错误「保留」任务且漏一个清理任务（-2）；③ RC#2 清单含一个无需改动的文件、漏一个受耦合影响的文件（-1）。 |
| PRD AC coverage | 43/50 | 六个 Story 的全部 AC 均有设计落点（登录矩阵/SSO/双向登出/403 原地/逐请求判定/只读裁剪/字段标红回显/凭据不泄露）；性能 Goal 有预算表+测量；监控（探活+告警、回切演练计时入部署清单）已承接；认证矩阵会签已恢复。扣 7：「下线就绪**清单**」工件无产出承诺，只承接了移除 location 的部署动作（-4）；Story 4 / UF-8 策略绑定流无接口锚（-2）；Story 4 AC3 授权缓存 ≤30s 语义未入 Phase 0 清单（-1）。 |

**门槛判定：159 < 160 → 仍阻断进入 `/breakdown-tasks`。** 判定理由：缺口集中且明确——(a) 四个 PRD 必需流的接口空洞（在 Phase 0 被收窄为「差异核查」后无兜底网）；(b) 两条跨仓引用盘点各含一处实盘错误。修复面极小：C 档补 4 行端点 + 恢复「差集盘点」口径或逐项标注 + 修正两处盘点。修复后本维度可达 ~175。

### 6. Security Considerations — 69/80

| Criterion | Score | 依据 |
|-----------|-------|------|
| Threat model present | 33/40 | 具名威胁：前端裁剪非安全边界（+Phase 0 逐端点核查矩阵）、凭据枚举（统一文案 + 锁定提示中性化，账号级/方式级口径已引 PRD）、cookie name/domain/path/SameSite/Secure 错配（兼容矩阵 + **eiam owner 书面会签已恢复**）、点击劫持/XSS（nginx 注入 CSP/XFO/HSTS）、第三方脚本供应链（零第三方）、跨域直连（SameSite/CORS 约束）。扣 7：token 窃取/会话固定/重放仍未具名——且设计自己的机制（「cookie 注入 Bearer」）**意味着 session cookie 必须可被 JS 读取**（非 httpOnly），这一 XSS 立面在威胁清单中零字对峙；CSP 仍无指令基线。 |
| Mitigations concrete | 36/40 | 威胁→缓解配对完整且大多具体：掩码不回显 + 初始密码展示一次 + 首登强制改密、服务端强制授权 + 核查矩阵、同登录域约束、写前守卫阻断错租户写、双人复核回切。扣 4：CSP 无 default-src 基线与起草人；「CSRF 归属 eiam」未写前提（SameSite 属性为何足以免除前端配合）；写前守卫依赖的「进入页面时租户快照」取值机制未定义（守卫本身是缓解，其取值不可实现则缓解落空）。 |

### 7. Implementation Feasibility — 121/140

| Criterion | Score | 依据 |
|-----------|-------|------|
| Dependencies available | 45/50 | 逐项实核通过（见维度 1）。扣 5：pnpm 与 Playwright 不在任何现有 manifest——新仓首锚的自证只覆盖 pnpm（`packageManager` 字段），Playwright 的版本锚定机制未写明（仅 `^1.55`）。 |
| Architecture fits project structure | 48/50 | 实核：`eiam.ts` 行为与设计逐字吻合（15s 超时/withCredentials/cookie 注入 Bearer/401 跳转/nginx `/api/iam/* → eiam :9000`）；主题两文件实存；RC#1 六处改动文件**6/6 与仓库实况一致**（`views/platform/users/index.vue`、`UserForm.vue`、`eiam-users.ts` 仅被两文件 import、`routes.ts:111`、`MainLayout.vue:331` 菜单项、`PlatformNav/index.vue` 恰好两处引���）；`eiam-tenant.ts` 保留理由属实（`TenantSelector.vue:24` 在用）。扣 2：RC#2 把 `eiam.ts` 列为改动文件与现仓结构不符（该文件无目标值）。 |
| Technical claims grounded | 28/40 | 可验证断言绝大多数实核通过（A 档 13 端点、`desc` 非 `description`、`active|disable` 与 ParseStatus 静默降级、switch 销毁重签 + reload 约定、profile 六字段、M:N 禁止、ginx 信封 `msg`、`:8888/cam` 前例、vite 直连缺 `/api/iam` 代理）。扣 12——修订新引入的两处与仓库事实相悖：① 「`eiam-mapper.ts` 保留——`stores/user.ts` 的 `mapEiamProfile` 在用」为**假**（实况见 Phase 1 错误 A；RC#1 后该文件仅剩自测引用，成死代码）（-6）；② `VITE_ECMDB_LOGIN_URL` 三处用途耦合未识别，改默认值将使「个人设置」跳转（`MainLayout.vue:58`）落到 `/console/login/profile/index` 死链，`MainLayout.vue` 不在 RC#2 清单，与回切三件套「不散落硬编码」的声明相冲突（-6）。 |

### Cross-dimension coherence check

- 「零后端、零数据库」↔ RC#1/RC#2 均为纯前端改动 + nginx 配置：一致。
- 持久化口径（persist 仅 theme）↔「无本地数据持久化」：已闭合，一致。
- 性能预算算术（60+140+50+50=300）↔ CI 门禁：一致。
- 验收宿主约束 ↔ Integration Spec 5 部署形态：一致。
- RC#2「机制 = build 期 env 注入，不散落硬编码」↔ 同一 env 变量第三处消费方（MainLayout 个人设置基址）的事实：**冲突**（已记维度 7）。
- RC#1 引用盘点「Phase 0 产出」的自我定位 ↔ 盘点内容已按断言预写且一条为假：**内部张力**（已记维度 5/7）。
- 未发现维度间其他硬冲突。

---

## Phase 3 — Blindspot Hunt

### [blindspot] B1 — 「写方向由写前守卫阻断」的保证范围被夸大：守卫只存在于控制台，e-cam-web 与过渡期 ecmdb-web 的写路径无任何守卫

引用：「显式接受边界：切租户后其他标签页已渲染数据保持 staleness……**读方向 staleness 接受，写方向由写前守卫阻断**」（Integration Spec 6）；拓扑图：「e-cam-web …… **读同一份**」共享 Redis session；PRD Data Requirements：「核查前按『错租户数据』为最高优先风险对待」。

推导：租户切换经共享 session 跨标签即时生效（tab A 切 X→Y 后 tab B 的下一个请求即按 Y 判定）。设计的写前守卫实现在**控制台的 7 个 list store**；e-cam-web 的写操作（资产/证书/日志页的编辑提交）与过渡期 ecmdb-web 的写操作同样面临「呈现 X 租户数据、发出 Y 租户请求」的写错租户形态，而 e-cam-web 明确 Out of Scope（「仅边界清理与 401 切换改动」）——即设计中无任何机制、也无显式接受声明覆盖这半边风险。当前的写法使读者误以为「错租户写」已被阻断，实则只覆盖控制台自身。此问题不属于任何 rubric 维度（非错误处理、非接口、非安全清单条目，是缓解措施的**覆盖边界**与真实保证之间的错位）。**必须改进**：二选一并写明——(a) 把 e-cam-web 侧残余风险列入显式接受边界（含理由与 UI 表现），或 (b) 在 RC#2 顺带为 e-cam-web 请求层加最小租户回读/写前校验（若判定超范围则回归 (a)）。

### [blindspot] B2 — 「Phase 0 差异核查」与「Phase 0 产出引用盘点」两个自我定位互相拆台，使接口空洞既不在核查清单上、盘点又可预写失实

引用：「Phase 0 只做**差异核查**（路径/字段/鉴权逐项对表登记 `parity-checklist.md`），不再从零盘点」（Interfaces）；「**引用盘点**（Phase 0 产出、并入 parity-checklist）」（RC#1，随后预写了含假依据的盘点结论）。

推导：前一句把 Phase 0 的安全网收窄为「核对已枚举三档」，后一句又让 Phase 0「产出」盘点——但盘点已在页内预写为断言且至少一条与仓库事实相反（mapper）。两个定位合成的净效果：**未枚举的端点永远进不了 parity-checklist**（密码重置、角色-策略绑定、passkey 引导、授权词表形状），而预写的盘点又会以「已核验」面目绕过 Phase 0 的核查动作。Reasoning audit 独立标记；其数值影响已计入维度 2/5/7，此处作为结构性风险单列。**必须改进**：统一口径——Phase 0 恢复 PRD 的「差集盘点」（以 UF 反推所需端点全集再对 eiam 求差），或至少把「引用盘点/契约断言」降格为「待 Phase 0 核验的假设清单」并逐条标注证据等级。

---

## 结论

迭代 2 的设计在「可拆解性」上完成了实质收敛：接口层从显式弃权变为实核吻合的 A 档契约、store/页面/跨仓改动全部可数、错误分类学与租户一致性模型补齐、性能预算落地。两处 blindspot（迭代 1 的 B1/B2）在页面上已有真实承接。

剩余 125 分缺口集中在四类，全部可精确修复：

1. **四个接口空洞**（E1 密码重置、UF-8 角色-策略绑定 + 角色已绑读取、UF-4 passkey 注册引导发起/撤销、授权词表响应类型）——C 档补 4 行即可解锁大部分分数。
2. **两处跨仓盘点实盘错误**（RC#1 mapper 假依据、RC#2 eiam.ts 误列 + MainLayout.vue 耦合遗漏）——按本报告引用的行号修正。
3. **Policy 形状分歧**（扁平 vs UF-10 Statement 数组）未声明处置。
4. **blindspot B1/B2**：写前守卫的保证边界、Phase 0 两个自我定位的调和。

**迭代 3 修复优先级**：
1. C 档补端点行（重置/绑定/绑定读取/passkey 引导/词表形状）并把 Phase 0 口径恢复为「差集盘点」或逐项降格为「待核验假设」——解锁 Breakdown-Readiness 门槛（159→~175）。
2. 修正 RC#1 mapper 三选一落点（实况：RC#1 后仅剩自测引用，应判迁移/删除并给出处置）；修正 RC#2 文件清单（删 eiam.ts、补 MainLayout.vue 或为个人设置跳转单独引入变量并写明切换行为）。
3. Policy 模型按 UF-10 Statement 数组重构或显式声明归一规则并标 Phase 0。
4. 写前守卫的取值机制 + 「写方向已阻断」声明范围收敛到控制台（或扩至 e-cam-web 最小守卫），e-cam-web/ecmdb-web 残余风险显式登记。
5. 补单元测试 DOM 环境与 @vue/test-utils 取舍声明、CSP 指令基线、`LoginResult` 锁定剩余时间数据源。
