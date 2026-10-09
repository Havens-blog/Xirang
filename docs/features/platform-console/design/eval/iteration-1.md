# 技术设计评分报告 — 平台控制台（迭代 1）

- **评审人**：Staff Architect（正式评分，迭代 1）
- **评分对象**：`docs/features/platform-console/design/tech-design.md`
- **对照输入**：`prd/prd-spec.md`、`prd/prd-user-stories.md`、`prd/prd-ui-functions.md`、`ui/ui-design.md`
- **注入事实核查**：`e-cam-web/package.json`、`e-cam-web/src/api/request/eiam.ts`、`e-cam-web/src/assets/styles/theme-variables.scss`、`element-theme.scss`
- **总分**：**721 / 1000**（目标 900，未达标）
- **关键门槛**：Breakdown-Readiness **120/180 < 160** → **阻断进入 `/breakdown-tasks`**

---

## Phase 1 — Reasoning Audit（独立推理审计）

### 论证链追踪

1. **Problem → Solution**：问题 = ecmdb-web 下线后平台身份/权限/工作台 UI 断供、eiam 无自有前端。方案 = 新建纯前端 SPA 直调 eiam 既有 API + 复用共享 Redis session。链条成立：零后端方案确实解决「宿主消失」且不重蹈「依赖他人前端」。`EcmdbAuthMiddleware` 同源 session 复用策略避免了自造鉴权，方向正确。
2. **Solution → Evidence**：证据经实盘核验成立——`e-cam-web/src/api/request/eiam.ts` 确实存在且与设计声称一致（`withCredentials: true`、`timeout: 15000`、401 → `redirectToLogin`、经 nginx `/api/iam/* → eiam :9000`）；`theme-variables.scss` / `element-theme.scss` 实存；package.json 版本与设计宣称逐项吻合（vue 3.5 / vite 7 / ts 5.9 / element-plus 2.10 / pinia 3 / vitest 4 / vue-tsc / unplugin）。这是本设计最扎实的部分。
3. **Evidence → Success Criteria**：测试目标可量化（≥80% 覆盖、E1–E12 + UF-1~11 逐条映射、负向+边界 ≥1/3、验收宿主锁定 nginx 同域形态），质量高。
4. **Self-contradiction check**：
   - **发现 A（断言 vs 缺席）**：PRD In Scope 8 项中，设计只承载了 6 项。Related Change #1（`platform/users` 迁出，含引用盘点/回归/回退/双签前置门）与 Related Change #2（e-cam-web 401 跳转目标切换）在设计全文中**零字提及**——Architecture 目录树、Integration Specs、Testing 均不含这两项跨仓改动。设计的组件清单因此不完整。
   - **发现 B（性能目标蒸发）**：PRD Goal「首屏静态资源 gzip ≤ 300 KB；同域内网 LCP ≤ 2.5s（二次 ≤ 1s）」在设计中无任何承接（无预算、无测量计划、无懒加载/分包策略）。
   - **发现 C（持久化口径自相矛盾，轻度）**：Architecture 声明「无本地数据持久化（除 cookie 会话与 theme localStorage）」，但 Overview 引入 Pinia 3（persistedstate）——persistedstate 默认写 localStorage，哪些 store 持久化、与「无本地持久化」口径如何相容，未说明。
   - **发现 D（Phase 0 依赖倒挂）**：Interfaces 一节自认「端点（示例，以 Phase 0 盘点为准）」。PRD 顺序上 Phase 0 先于 M1 成立，但技术设计作为拆任务输入，其接口层实质为空壳——本可当场枚举的 eiam 既有端点（e-cam-web 已在调用部分端点）被整体推迟。发现的深度问题归入维度 2/5 评分。
   - **SC/In-Scope 两两可满足性**：未发现互斥型矛盾（发现 A/B/C 属于「承诺 X 实现遗漏 Y」，非逻辑互斥）。

### Pre-score anchors（转入 blindspot/攻击）

- 跨标签页租户一致性未建模（共享 session tenant_id + 多标签无广播 → 混合租户渲染态）；PRD 明言「错租户数据」为最高优先风险，设计未承接。
- 「拷贝 + 锁同版本」共享机制在消除 `platform/users` 双处维护的同时**新造了**主题/请求层文件的第二处维护面，且无漂移检测机制。

---

## Phase 2 — Rubric Scoring（验证姿态）

> 姿态：文档中每条断言视为未证实，直至有页内证据或实盘证据支撑。

### 1. Architecture Clarity — 117/170

| Criterion | Score | 依据 |
|-----------|-------|------|
| Layer placement explicit | 50/60 | 「单 SPA 分层，镜像 e-cam-web 目录约定」+ 单向数据流 `views → stores → api/eiamAxios →(HTTP)→ eiam` 明确；但各目录职责仅一行注释，store 与 api 层职责边界（如租户上下文归属谁）、组件层粒度未定。 |
| Component diagram present | 25/60 | **只有目录树，没有组件关系图**。唯一的「关系」表达是一行文字数据流。本设计最复杂、风险最高的部分——console / e-cam-web / 过渡期 ecmdb-web / eiam / 共享 Redis / nginx 同域反代的多方会话拓扑（SSO、双向单点登出、401 收敛、cookie 域）——**纯散文描述，无图**。目录树是文件清单，不满足「components and relationships」。Prose-only where diagram expected：本应 -50，考虑目录树 + 数据流行有部分图示价值，记 25/60。 |
| Dependencies listed | 42/50 | 依赖清单完整且经实盘核验（版本全对）。扣分项：pnpm 未在任何现有 manifest 出现（e-cam-web 未声明 packageManager）；Playwright 不在 e-cam-web 依赖中、版本未锁；「Vitest 4」仅给主版本。 |

### 2. Interface & Model Definitions — 85/170（db-schema=no 路径）

| Criterion | Score | 依据 |
|-----------|-------|------|
| Interface signatures typed | 25/60 | Interfaces 一节是「域 → 端点示例 → 用途」三列表 + 明确弃权声明：「端点（示例，以 Phase 0 盘点为准）」。没有任何一个端点的请求/响应 payload 类型；没有任何 api 客户端函数签名（`auth.ts` 等仅是文件名）；统一信封 `{code, message, data}` 是散文而非 `Envelope<T>` 类型。「端点由 Phase 0 parity-checklist 冻结」——eiam 端点是**已存在**的，不是待建的：设计本可现在枚举，却全部推迟。 |
| Models concrete | 35/60 | 8 个 interface 有字段名 + 联合类型，方向正确。但存在未定义类型与逃逸口：`User.roles: RoleRef[]` 与 `Role.policies: Ref[]`——**`RoleRef` 和 `Ref` 全文未定义**；`IdentitySource.conn: ConnParams`——UF-7 已列出连接参数字段（地址/Bind DN/密码/Base DN/属性映射/超时），模型却用一个未定义的 `ConnParams` 吞掉；`Policy.condition?: object` 无形状；`// + passkey_session 注册态` 是注释不是字段；无分页模型（page/pageSize/total）、无信封类型。 |
| Directly implementable | 25/50 | 可以搭出骨架（views/stores/router），但**无法在不猜类型的情况下写 api 层**：每个端点的请求/响应形状未知；表单层无法从 `ConnParams` 落地；文案契约 keys 在另一文档（可接受，引用可解析——ui-design.md L543 起确有契约表）。 |

### 3. Error Handling — 86/130

| Criterion | Score | 依据 |
|-----------|-------|------|
| Error types defined | 20/45 | **零自定义错误类型/错误码分类学**。「文案契约：错误码 → 逐字文案映射」引用了错误码但未枚举（在 ui-design 有表，设计侧无 `ApiError {code, message}` 之类的 TS 定义）；拦截器错误对象形状未定义。 |
| Propagation strategy clear | 38/45 | 传播策略是全文最强段落之一：axios 响应拦截器集中处理 401（跳 `/login` 不清共享 cookie）/403（原地 toast）/网络错+15s 超时（页面错误态+重试）/eiam 不可用（静态内容仍渲染+降级+写操作禁用）；信封解包 `code != 0` → 契约文案；字段级 → `aria-describedby`。扣分：未说明「eiam 返回非 2xx 但带信封」与「200 + code != 0」两种形态如何统一解包；401 重定向后各 store 在途状态如何清理未提。 |
| HTTP status codes mapped | 28/40 | 401/403/超时/不可用已映射。缺失：400/422 字段校验错误的状态映射（字段级处理提了但状态码未挂钩）；**409 未出现**（PRD E6 并发冲突依赖它，Phase 0 才核查——设计错误处理章至少应预留分支）；5xx 与降级的边界未写。 |

### 4. Testing Strategy — 119/130

| Criterion | Score | 依据 |
|-----------|-------|------|
| Per-layer test plan | 40/45 | 单元（Vitest：stores/api/utils/文案契约）+ 真实浏览器 E2E（Playwright，映射 E1–E12 与 UF-1~11，逐条证据）+ typecheck/build 门禁 + **验收宿主锁定 nginx 同域形态**（「不得在独立端口 dev server 上判定通过」）——最后这条是少见的清醒约束，直击本项目真实历史坑。扣分：无组件级测试层（@vue/test-utils 未提）；单元测试 DOM 环境未指定。 |
| Coverage target numeric | 45/45 | 「覆盖 ≥80%」，数值明确。 |
| Test tooling named | 34/40 | Vitest 4、Playwright、vue-tsc、pnpm、浏览器 MCP 证据形态均已点名。扣分：Playwright 版本/浏览器矩阵未锁；happy-dom/jsdom 环境未指定。 |

### 5. Breakdown-Readiness ★ — 120/180（**低于 160，阻断**）

| Criterion | Score | 依据 |
|-----------|-------|------|
| Components enumerable | 50/65 | 新仓内部组件可枚举（8 个 api 文件、11 页面、stores、共享组件、e2e、nginx 配置）。扣分：(a) 「stores/ # Pinia：user（会话）、tenant、theme、**各页 list store**」——「各页」含糊，8 个 CRUD 页是每页一个还是共享泛型 store，数量不可数；(b) **跨仓工作项从组件版图中消失**：e-cam-web 401 跳转目标切换、`platform/users` 迁出（含引用盘点/回归/回退）、ecmdb-web 过渡期重定向改动——三项 In Scope/Related Change 在设计架构中无落点。 |
| Tasks derivable | 50/65 | 骨架任务可派生（脚手架、主题拷贝、请求层、auth 流、租户流、8 个 CRUD 页、布局、E# 映射 E2E、nginx 配置），M1–M4 里程碑对齐良好。扣分：api 客户端任务在 Phase 0 产出前无法写出可验收的任务描述（无请求/响应形状）；「各页 list store」无法直接成任务。 |
| PRD AC coverage | 20/50 | Stories 1–6 的 AC 均可在设计中找到承接（登录矩阵/SSO/双向登出/403 原地/只读降级/服务端强制授权）。但：**Related Change #1、#2 零承接**（PRD 明文：「移出 `platform/users`」「会话过期/401 跳转目标……改为控制台登录页」，设计全文未出现）；**性能 Goal（gzip ≤300KB / LCP ≤2.5s / 二次 ≤1s）零承接**；「下线就绪清单」产出物未提。按每条未承接 -30 计已远超本 criterion 分值，封底 20/50。 |

**门槛判定：120 < 160 → 阻断进入 `/breakdown-tasks`。** 判定理由：拆任务的第一输入是完整组件清单，当前设计对其自身 In Scope 的 1/4（跨仓改动）不建模，且接口层为显式弃权状态。

### 6. Security Considerations — 67/80

| Criterion | Score | 依据 |
|-----------|-------|------|
| Threat model present | 32/40 | 具名威胁：前端裁剪非安全边界（直调 API 越权）、凭据枚举（统一文案）、cookie 域/路径错配（跨应用会话面）、点击劫持/XSS（CSP/X-Frame-Options/HSTS）、第三方脚本供应链（零第三方脚本）、跨域直连（SameSite/CORS）。扣分：无系统性威胁清单（token 窃取/会话固定/重放未具名）；暴力破解速率限制仅存在于 PRD 的锁定机制（5 次/15 分钟），设计 Security 节未承接；PRD 要求的「认证链路兼容矩阵经 eiam owner 书面会签」在设计侧只写了「Phase 0 出兼容矩阵」，会签动作丢失。 |
| Mitigations concrete | 35/40 | 威胁→缓解配对具体：裁剪→eiam 服务端强制 + Phase 0 逐端点核查矩阵；枚举→统一文案 + 中性锁定提示；凭据→掩码不回显 + 初始密码展示一次 + 首登强制改密；header→nginx 注入。扣分：CSP 只写「nginx 注入 CSP」无指令基线（谁起草、default-src 什么）；「CSRF 归属 eiam」未说明控制台侧前提（cookie SameSite 属性是否足以免除前端配合）；「回切开关」作为安全兜底手段无机制设计。 |

### 7. Implementation Feasibility — 127/140

| Criterion | Score | 依据 |
|-----------|-------|------|
| Dependencies available | 45/50 | 逐项实盘核验通过：vue ^3.5.22、vite ^7.1.7、typescript ~5.9.3、vue-router ^4.5.0、pinia ^3.0.1、pinia-plugin-persistedstate ^4.2.0、element-plus ^2.10.5、@element-plus/icons-vue、unplugin-vue-components/unplugin-auto-import、sass、axios、vitest ^4.1.5、vue-tsc——**全部存在于 e-cam-web/package.json**。扣分：pnpm 与 Playwright 在现有 manifest 中均不存在（-5）。 |
| Architecture fits project structure | 50/50 | 镜像 e-cam-web 目录约定属实：`api/request/eiam.ts` 实存且行为与设计描述逐字吻合（15s 超时、withCredentials、Bearer 注入、401 跳转、nginx `/api/iam/* → eiam :9000`）；theme 两文件实存；stores/router/views 分层与现仓一致。新增的 403/降级处理是对现有单 401 处理的扩展，无模式冲突。「拷贝 + 锁同版本」替代抽象组件库，与 PRD Out of Scope（共享组件库另立提案）自洽。 |
| Technical claims grounded | 32/40 | 可验证断言均有实据（eiamAxios 模式、:8888/cam 反代前例、共享 session 同源、CDN 同域约束）。扣分：「上线前回切开关演练（10 分钟内）」——**回切开关是什么（nginx upstream 切换？蓝绿目录？DNS？）零设计**，10 分钟承诺无机制支撑；端到端登录链路（eiam 写 session → 中间件读）标了「以 Phase 0 认证链路事实核查为准」是诚实标注，但意味着核心链路当前未被证实。 |

### Cross-dimension coherence check

- Overview 承诺「零后端、零数据库」↔ 全文无 API 之外的组件：一致。
- 「无本地数据持久化」↔ Pinia persistedstate 引入：轻度矛盾（发现 C），记在 Architecture。
- Interfaces 「示例」弃权 ↔ Breakdown-Readiness 门槛：设计自己的契约基线依赖 Phase 0，而设计的价值承诺是可直接拆任务——结构性错位，主扣分已在维度 2/5。
- Testing 验收宿主约束 ↔ Integration Spec 5 部署形态：一致。
- 未发现维度间硬冲突。

---

## Phase 3 — Blindspot Hunt

### [blindspot] B1 — 跨标签页租户一致性未建模，而 PRD 把「错租户数据」列为最高优先风险

引用：「tenant_id 存共享 session，跳转后由 eiam/e-cam-service 中间件读同一份」（Integration Specs #4）与「多标签页不做实时广播」（Integration Specs #3）。

双向推导：tenant_id 在服务端 session（跨标签共享）+ 登出不做实时广播是显式接受的决策，但**租户切换**同理跨标签生效——标签 A 切租户 X→Y 后，标签 B 的已渲染数据仍是 X 租户的，而 B 的后续读写请求按 session 中的 Y 租户被中间件判定。结果：B 页面呈现 X 的数据、发出 Y 的请求——正是 PRD Data Requirements 声明的「按『错租户数据』为最高优先风险对待」的形态，设计中无任何缓解（无 per-tab tenant 快照、无切租户强制全页刷新约定、无响应租户校验）。此问题不属于任何 rubric 维度（非错误处理、非接口、非安全清单条目），故以 blindspot 记。**必须改进**：明确跨标签租户一致性模型（如：租户切换写 session 后强制 `location.reload` 所有导航路径 + 关键响应回读 tenant_id 与 UI 态比对，或明示接受 staleness 的边界与 UI 表现）。

### [blindspot] B2 — 「拷贝 + 锁同版本」重新引入了本提案宣称消除的双处维护

引用：「**从 e-cam-web 拷贝**、引用同一套 `--accent-primary #5e6ad2` / ... 令牌」（Overview）；PRD In Scope：「从 e-cam-web 迁出平台身份 UI（`platform/users`），**消除双处维护**」。

域专家失败模式「Solutions that reintroduce patterns they claim to eliminate」直接命中：迁移消除了一处双维护（platform/users），拷贝机制又新造一处（theme-variables.scss / element-theme.scss / eiamAxios 模式三份拷贝），且**无漂移检测机制**（无 CI 比对、无版本锚点文件、无变更通知约定——PRD 的 ≥5 工作日通知只覆盖 eiam 端点，不覆盖 e-cam-web 主题/请求层文件）。Reasoning audit 已独立标记。**必须改进**：为拷贝面建立漂移防线（如 CI hash 比对 e-cam-web 源文件、或提取为带版本号的共享包/子模块并在设计中写明升级流程）。

---

## 结论

设计的工程品味良好（真实栈核验全过、验收宿主约束清醒、安全姿态正确），但作为**拆任务输入**不合格：接口层显式弃权给 Phase 0、两条 In Scope 跨仓改动从架构中消失、性能目标蒸发、多方会话拓扑无图。迭代 2 修复优先级：

1. 补 Related Change #1/#2 的设计落点（e-cam-web 改动文件清单 + platform/users 迁出方案 + 回归/回退）——解锁 Breakdown-Readiness 门槛。
2. 枚举 eiam 实际端点 + 每端点请求/响应 TS 类型（eiam 已存在，无需等 Phase 0；Phase 0 只做差异核查）。
3. 补多方会话拓扑组件图 + 跨标签租户一致性模型。
4. 补性能预算承接（gzip ≤300KB 分包/测量计划）与补齐未定义类型（Ref/RoleRef/ConnParams/Envelope<T>）。
5. 回切开关机制化 + 拷贝面漂移检测。
