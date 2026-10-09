# UI Design 评分报告 — Iteration 1

- **文档**: `docs/features/platform-console/ui/ui-design.md`
- **需求源**: `docs/features/platform-console/prd/prd-ui-functions.md`（辅以 `prd-spec.md` E1–E12 交叉核对）
- **Rubric**: `ui-web`（1000 分制；4 视角 × 250）
- **评分者**: Senior UX Engineer（scorer-protocol 三阶段）
- **日期**: 2026-09-29

---

## 总分

| 视角 | 得分 | 满分 |
|------|------|------|
| 1. Requirement Coverage（PM） | **110** | 250 |
| 2. User Experience（终端用户） | **155** | 250 |
| 3. Design Integrity（设计师） | **200** | 250 |
| 4. Implementability（开发者） | **160** | 250 |
| **合计** | **625** | **1000** |

**一句话结论**: 设计系统基座（Shadcn token 表 + 组件规范）是全文最强的部分；但 11 个 UF 的页面规格停留在骨架级——导航架构与 PRD 直接冲突、4 个 PRD 状态缺失、约 10 个孤儿元素、辅助功能（a11y）零覆盖、E1/E6/E7 等 PRD 边界场景未落 UI。距离可开工的实现规格还有明显差距。

---

## Phase 1 — Reasoning Audit（独立推理审计）

**问题 → 方案**: PRD 要求 11 个 UI Function 的全新纯前端控制台。设计逐一给出 Placement/Layout/States/Interactions/Data Binding，UF-1～UF-11 全部有对应 Component，问题对位正确。

**方案 → 证据**: 关键断言大多可回溯到 PRD，但发现一条**流程级自相矛盾**：

- UF-1 Interactions 写「提交登录 | 调 eiam 登录接口 → 写共享 session | **成功跳工作台**」；
- 而 UF-11 整个组件的存在前提是「认证通过后、进入工作台前」的租户分流（PRD UF-1 flow 明确「成功则写 session 并进入租户选择」，PRD States「成功 | 跳转 UF-11 租户选择/工作台」）。

按设计文档字面执行，登录将**绕过租户选择页**，租户上下文（tenant_id）永远不会经 UF-11 写入 session——这会使 UF-6 租户边界、UF-8 租户管理员范围全部失去前提。这是最高优先级修复项。

**证据 → 成功标准**: 设计无独立成功标准（依赖 PRD acceptance.md），无冲突。

**自相矛盾追加**: UF-11 零租户态承诺「受限工作台（仅欢迎 + 我的信息，禁用管理入口）」，但 UF-3 工作台组件自己的 States 表只有 Default 与「eiam 不可用」两态，**不存在受限视图状态**——UF-11 引用了一个 UF-3 未定义的形态。

**Phase 1 前置锚点（转入后续评分/blindspot）**:
1. 登录→工作台 vs UF-11 分流矛盾（计入 Requirement Coverage，-30）。
2. UF-3 缺受限视图态（计入 State requirement coverage）。
3. 设计系统声明双主题 + 「实际组件库映射在 /tech-design 落地」——组件库未定却已写出 h-10/p-6/space-y-6 等 Tailwind 风格原子类（转入 blindspot）。

---

## Phase 2 — Rubric Scoring

### 1. Requirement Coverage — 110/250（PM 视角）

**UI function coverage: 60/80**

- UF-1～UF-11 全部有对应 Component，目标路由与 PRD Page Composition 一一相符（/login、主布局、/workbench、/users、/organizations、/tenants、/identity-sources、/roles、/authorizations、/policies、/tenant-select）。基础分成立。
- **扣分 1（passkey 注册引导 UI 缺失，-8）**: PRD UF-4 有整节「passkey 注册引导（定义）」——管理员发起引导、一次性注册会话 15 分钟有效期、完成态标记、设备丢失时「管理员撤销原 passkey 绑定 → 重新发起」的唯一重绑路径。设计仅在 Data Binding 括号里写「login_method（密码/passkey 注册引导）」，**无任何发起/有效期/完成/撤销绑定的交互与状态**。
- **扣分 2（E1 首次登录强制改密缺失，-4）**: PRD E1「管理员重置 → 一次性初始密码（展示一次）；该用户**首次登录强制改密**」。设计的 initial_password 有「展示一次」，但登录页 States 无强制改密态，全文档「强制改密」零命中。
- **扣分 3（UF-4 所属租户/组织缺失，-4）**: PRD UF-4 Data Requirements 明确「所属租户/组织 | ref | eiam | 关联选择」；设计的 Dialog 字段列表（用户名/初始密码/登录方式/角色绑定）与表格列（用户名/登录方式/角色/状态/操作）均无此项，Data Binding 表也无此字段。
- **扣分 4（只读/审计角色视图态缺失，-4）**: PRD 导航注记「只读/审计可见 2/3/5/6/7，只读查看、**不渲染增删改入口**」。设计的 UF-6 写了「仅平台管理员可见」、UF-8 写了租户管理员策略绑定口径，但 UF-4/5/7/9 的 States/Interactions 均未建模「只读模式下隐藏新增/编辑/删除入口」的条件渲染。

**Navigation Architecture coverage: 0/40**

- Primary Navigation 8 个入口对应页面全部存在 ✓（但 UF-2 布局仅写「主菜单（按角色裁剪）」，未落 PRD 的 8 项菜单与 Icon Keyword，靠读者回查 PRD）。
- **Secondary Pages 3 项中 2 项被降级为 Dialog，直接违反 PRD 导航架构（-20 × 2 = -40）**: PRD Secondary Pages 表定义「用户详情/编辑 | UF-4 列表『编辑』 | Return Target /users」与「角色详情/绑定 | UF-8 列表『编辑』 | Return Target /roles」，且 Navigation Rules 要求「Every secondary page must have back navigation targeting its entry point page」。设计写「新增/编辑用 Dialog」「编辑弹窗」，详情/编辑页不存在、返回导航无从谈起。UF-8 的角色绑定（策略多选 + 用户多选）塞进 Dialog 的可用性也存疑——这正是 PRD 把它定为页面的动机。
- 登录页 /login（第 3 项 Secondary Page）覆盖 ✓。

**State requirement coverage: 55/80**

PRD 各 UF States 共约 41 条，设计缺失 4 条（每条约 -6）：
- **UF-2 缺「过期」态**: PRD「过期 | 跳登录页 | 401」。设计 UF-2 States 仅 Default 与 403 两行，全文档 401 只出现在登出交互（「双侧下次请求 401 → 回登录页」）中，**被动会话过期→跳登录页的全局行为无任何归属**。
- **UF-11 缺「加载中」与「错误」态**: PRD 明确「加载中 | skeleton | 查询所属租户」「错误 | 错误提示+重试 | 接口失败」；设计只有 多租户/单租户/零租户 三行。查询所属租户是登录后的必经异步请求，缺 Loading/Error 意味着失败时用户卡死在空白分流点。
- **UF-3 缺受限视图态**: UF-11 零租户态引用的「受限工作台」在 UF-3 组件内无定义（Phase 1 锚点 2）；PRD E3 的「访问管理页提示『无所属租户』，不报错不白屏」同样无组件承接。

其余 UF-4～UF-10 的 Loading/Empty/Populated/Error 四态齐全，且 UF-5/6/8 额外补齐了删除/禁用被拒态，UF-10 补了校验失败态——这部分做得好。

**Edge case handling: 25/50**

已覆盖（加分面）：E2/E9 锁定（Locked 态 + 倒计时 + 中性文案）、E3 零租户（UF-11，含平台管理员豁免）、E4 单租户自动选定、E8 WebAuthn 不可用隐藏 passkey、E11 删除被拒（UF-5 弹窗列阻断计数确认置灰，UF-8/UF-10 被拒提示）、E10 用户名正则 `^[a-z][a-z0-9_-]{2,31}$`、E12 eiam 不可用（UF-3 降级 + 各管理页 Error+重试）。

未覆盖（扣分面，-25）：
- **E6 并发编辑**: 全文档「并发/409」零命中；PRD UF-4 明确「分页与并发边界见 prd-spec.md E6/E7」，设计未承接。
- **E7 分页边界**: 超界页码回落、删除末页唯一记录回退——设计各页只写「分页」二字，无边界行为。
- **UF-7 凭据掩码**: PRD「密码仅写入，掩码展示不回显明文」；设计的密码字段无掩码约束（仅 UF-4 编辑有无关的「变更回显」）。
- **长文本溢出**: 用户名/邮箱/DN 等长字符串在表格与树节点中的截断/省略规则全文档未提。
- **树超限**: UF-5 写了「树深度 ≤8、节点 ≤1000」但缺 PRD「超限禁止再建子级」的 UI 行为（新增按钮禁用/提示）。
- **UF-1 空凭据拦截**: PRD「凭据为空不可提交」未进入设计的 Validation 呈现。

**Cross-section inconsistency（-30）**: UF-1「成功跳工作台」↔ UF-11 组件存在性与 PRD 分流流程（Phase 1 锚点 1），按 rubric 规则每冲突 -30。

**小计**: 60 + 0 + 55 + 25 = 140；-30 = **110/250**。

### 2. User Experience — 155/250（终端用户视角）

**Information hierarchy: 65/80**

- 页面骨架符合后台扫描习惯：页面头（标题 + 主操作按钮）→ 过滤区 → 主表格/主内容，主操作（新增用户/新增授权）位置一致；工作台「欢迎横幅 → 应用卡片」层次清晰；登录页 max-w-sm 居中卡片是成熟范式。
- 扣分：UF-3「我的信息」只有一句话（「H2 欢迎 + 我的信息」），信息内容、字段数、呈现密度全部未定义，无法评估该区块的信息层级；UF-4 搜索栏「关键词输入 + 过滤」的过滤维度（角色下拉？状态？）未展开，扫描动线不完整；表格页普遍未说明列优先级/省略策略，长 DN/邮箱类内容会破坏扫描节奏。

**Interaction intuitiveness: 65/80**

- 惯例模式选型正确：Tab 切登录方式、Dialog 做 CRUD、左树右详情 master-detail、radio 选租户、确认弹窗做破坏性操作、skeleton 表加载——用户零学习成本。MFA 6 位码、锁定倒计时、单租户「短暂提示」自动分流都是清晰反馈。
- 扣分（歧义伤直觉）：
  - UF-2「点应用切换器 | 单点跳 e-cam-web | **新标签/同标签跳转**」——二选一未决，用户心智模型（新标签保工作台上下文 vs 同标签导航）是交互设计的基本决策，不是实现细节。
  - UF-11 单租户「自动选定（**短暂提示**）」——提示形态与时长未定。
  - UF-4「列表**即时过滤**」——分页场景下前端过滤当前页还是触发服务端查询，体验完全不同，未决。
  - UF-5「弹窗操作」、UF-6/UF-8「刷新」——反馈时机缺失（对比 UF-4 有「≤2s 刷新」，同类页面口径不一）。

**Accessibility: 25/90**

全文 accessibility 覆盖接近于零，这是本文档最大的单点失分：

- **键盘导航**: 零提及。登录 Tab、组织树（键盘树导航是硬需求）、表格行操作、Dialog 焦点圈闭（focus trap）、Esc 关闭——均无。组织树这类组件没有键盘方案等于对键盘用户不可用。
- **表单标签**: 零提及。UF-11 radio 组、UF-1 Tab、各 Dialog 字段的 label/必填标识/错误关联（aria-describedby）未定义。
- **状态的可感知性**: rubric 明确问「loading/error 是否以非纯视觉方式传达」——skeleton、错误条、降级提示、锁定倒计时全部只有视觉形态，无 aria-live/role=alert 约定。
- **对比度**: 唯一可算作加分的是完整 token 表（zinc 系 + foreground/background 组合天然满足 AA）与输入 focus ring（「focus ring 2px + ring-offset 2px」是全文唯一焦点相关内容）；但 muted-foreground #71717a 仅在白底达标，用于 Dark 主题 #a1a1aa 的场景未做声明性承诺。
- **图标**: PRD 导航表给了 8 个 Icon Keyword，设计只字未提图标及其文字替代（icon-only 按钮的 aria-label）。
- 综合给 25/90：token 体系与 focus ring 提供了间接基础，但无任何显式 a11y 决策。

**小计**: 65 + 65 + 25 = **155/250**。

### 3. Design Integrity — 200/250（设计师视角）

**Design system adherence: 70/80**

- 强项：Shadcn 定位准确；色彩 token 表 Light/Dark 双主题完整且取值与 shadcn zinc 默认一致；6 类组件规格给了可执行数值（按钮 h-10/圆角 6/150ms、卡片 p-6/无阴影、输入 focus ring、徽章四变体、表格 header bg-muted、对话框遮罩 blur + scale 动效）；「深度与间距」（4px 基准）与 Do/Don't 清单（禁硬编码色值、禁混圆角）是真实的防走样约束。
- 扣分 1（-6）：**语义色缺口**——调色板只有 destructive，无 success/warning 角色；而「启用/禁用/锁定」状态列（UF-4/6/7）、eiam 降级提示（UF-3）、连接失败提示（UF-7）都需要非 destructive 的语义色，徽章四变体（default/secondary/outline/destructive）里没有对应变体。状态着色将被迫自创，恰好违反自家 Don't「硬编码色值」。
- 扣分 2（-4）：字体声明自相矛盾于单一家族惯例——正文系统栈 vs 标题 Inter，且 16px 正文对高密度管理表格偏大但未说明表格字号。

**Visual coherence: 75/90**

- 强项：全局共享单一主布局组件（UF-2），跨页导航元素天然一致（本次评分未发现共享导航不一致，未触发 -30 规则）；「无阴影系统 + 1px border + 两级圆角」贯穿登录卡、管理页卡片、对话框；各管理页统一「页面头 + 表格 + Dialog」模式，跨页一致性好。
- 扣分 1（-6）：状态列视觉无规范——同一「状态」概念在 UF-4 是三态（启用/禁用/锁定）、UF-6 两态、UF-7 开关，徽章色映射未定义，跨页状态语言不统一。
- 扣分 2（-5）：反馈时限口径不一——UF-4「保存后列表 ≤2s 刷新」vs UF-6/UF-8 泛泛「刷新」，同类 CRUD 页的反馈承诺不同。
- 扣分 3（-4）：UF-1 CardFooter 的「帮助」入口无任何去向定义，登录页出现一个无归宿元素破坏完成度。

**State completeness: 55/80**

- 多数交互组件四态齐全（UF-4～UF-10 全部有 Loading/Empty/Populated/Error），UF-5/6/8 还加了被拒态，状态表习惯本身已养成。
- 扣分：UF-2 缺 Loading（菜单/会话异步加载期壳层显示什么）与 过期（401）态（-10）；UF-11 缺 Loading/Error（-10）；UF-3 缺受限视图态（-5，与 Requirement Coverage 的状态缺口同源但此处按组件完整性计）。
- **状态迁移未描述**: rubric 要求「组件如何从 Loading 走到 Empty」——全部组件的迁移条件（接口返回空数组 vs 报错）均靠读者默认推断，无一处显式迁移说明（-累计已含在上述扣分内，不再重复计）。

**小计**: 70 + 75 + 55 = **200/250**。「Happy-path only（-50）」不触发——状态表习惯良好。

### 4. Implementability — 160/250（开发者视角）

**Layout specificity: 55/80**

- 已给出：header h-14 + sticky + border-b、登录卡 max-w-sm、内容区 space-y-6、卡片 p-6、间距基准 4px、对话框动效参数（fade + scale 0.95→1）。
- 扣分（-25）：
  - UF-5「左：组织树；右：详情/成员」**无任何宽度比例/最小宽度**——树容器与详情面板的分配全凭开发猜。
  - UF-3「grid 2–3 列」——2 还是 3、断点如何切换，未决。
  - **全文档零响应式声明**（响应式/断点/最小视口 grep 零命中）。内网后台可以桌面优先，但连「最小支持分辨率 1280px」这类底线都没写，树+详情双栏布局在窄屏下的行为完全未定义。
  - Dialog 尺寸档位（策略 JSON 编辑与 6 字段表单显然不同宽）未定义。
  - 表格列宽/溢出策略未定义（与长文本问题叠加）。

**Data binding explicit: 45/80**

孤儿/未绑定元素清单（每处均无 Data Field/Source）：

| # | 组件 | 孤儿元素 | 说明 |
|---|------|----------|------|
| 1 | UF-4 | 用户列表（表格本体） | Data Binding 表只有表单字段 + status，列表数据源缺失 |
| 2 | UF-4 | 分页器 | page/total 无绑定 |
| 3 | UF-4 | 所属租户/组织 | PRD 有此数据需求，设计连元素带绑定一起消失 |
| 4 | UF-5 | 右侧「详情/成员」面板 | 详情字段、成员列表均无字段定义、无绑定 |
| 5 | UF-6 | 「用户数」列 | 设计自创列，无来源（且与禁用校验所需的「活跃会话数」是两个概念，混淆风险高） |
| 6 | UF-7 | 「名称」「类型」两列 | Data Binding 表无 name/type 字段 |
| 7 | UF-8 | 「绑定策略数」「绑定用户数」两列 | 无来源 |
| 8 | UF-10 | 「绑定对象数」列 | 无来源（而删除被拒恰需绑定角色数） |
| 9 | UF-2 | 租户选择器的选项列表 | 只绑了当前 tenant_id，可选项来源未定义 |
| 10 | UF-1 | 「帮助」入口 | 无绑定、无交互 |

约 10 处孤儿元素，按 rubric 规则属系统性缺口而非零星遗漏。绑定质量好的一面：字段命名与 PRD 高度对齐（username 正则、attr_map、policies[]/users[]、effect 枚举），UF-9 三元组「词表下拉非自由文本」的来源声明清晰。

**Interaction unambiguity: 60/90**

- 合格面：多数链条 Trigger→Action→Feedback 完整且带参数——UF-1「调 eiam 登录接口 → 写共享 session」，UF-9「前端预检唯一性 → 提交」「撤销生效（延迟 ≤30s 提示）」，UF-10「前端语法 → eiam schema 校验，失败标红」，UF-6「校验未过期会话数=0 → 确认」，UF-7「端口 1–65535 / 超时 1–30s」。
- 扣分（-30）：
  - 「新标签/同标签跳转」未决（交互基本问题）。
  - UF-5「弹窗操作」——弹窗内容只有 org_name 一个绑定，字段与校验链条不完整；UF-6「弹窗」同理。
  - UF-4「即时过滤」前端/后端未决；UF-6/UF-8「刷新」无时限（与 UF-4 ≤2s 口径分裂）。
  - UF-7「**6 字段表单**」实列 7 项（ldap(s)地址/端口、Bind DN、密码、Base DN、属性映射、超时），且 PRD 的「用户名属性必填」等字段级校验未进入设计。
  - UF-11「短暂提示」「（口径待 Phase 0 核查）」——后者忠实转抄 PRD，可接受，但作为设计交付物核心绑定仍悬置。
  - 各删除/禁用确认弹窗未说明需输入确认（如输入对象名）还是单击确认——破坏性操作强度未分级。

**小计**: 55 + 45 + 60 = **160/250**。

---

### 跨维度一致性核查

- UF-1「成功跳工作台」↔ UF-11 分流（RC -30，已计）。
- UF-11 零租户态引用的「受限工作台」↔ UF-3 无此状态（RC 状态、DI 状态完整性双处计）。
- UF-6 禁用校验需要「活跃会话数」↔ 表格列却是「用户数」（Impl 绑定缺口 #5，两概念并存易实现错）。
- 设计系统「支持明暗双主题」↔ UF-2 壳层无主题切换/默认主题决策（转入 blindspot）。
- UF-8 Dialog 内做策略+用户双多选 ↔ PRD 将角色详情/绑定定为 Secondary Page（导航架构扣分的体验佐证）。

---

## Phase 3 — Blindspot Hunt

1. **[blindspot] 组件库映射悬置，动摇全文规格的可执行前提** — 原文：「Shadcn — ……（与 e-cam-web 的 Element Plus 企业后台审美对齐，实际组件库映射在 /tech-design 落地）」。文档一边把 h-10/p-6/space-y-6/ring-offset 等 Tailwind 原子类写成硬规格，一边声明真实组件库待 tech-design 才定；若最终映射到 Element Plus 系，全部组件规范数值需要重写。Reasoning audit flagged this independently of dimension scoring. 必须改进：明确「Shadcn token 体系 + 某一确定组件库」的映射关系，或把原子类规格降级为 token 语义规格（如 control-height: 40px）。

2. **[blindspot] 双主题有声明无决策** — 原文：「### 色彩（CSS 变量，支持明暗双主题）」；而 UF-2 壳层「右：租户选择器 + 应用切换器 + 用户头像下拉（登出）」无主题切换入口，全文档无默认主题、无 prefers-color-scheme 跟随策略。Token 表是死资产，除非有消费决策。必须改进：声明默认主题与跟随策略，或在壳层补切换器及其状态。

3. **[blindspot] 文案契约缺失，中性化安全要求被转述稀释** — 原文：「Locked | 中性锁定时提示（仅剩余时间 + 其他可用方式）」「表单顶部统一中性错误条（不泄露用户名）」。PRD 给出了精确文案契约（「用户名或密码错误」统一文案；「尝试次数过多，账号已锁定，剩余 X 分钟」），这是防账号枚举的安全级要求；设计只做了转述且措辞漂移（「不泄露用户名」≠「不泄露账号是否存在」），最终文案、复数/倒计时刷新、i18n 均未定。必须改进：设文案表（含错误码→文案映射），锁定/失败文案逐字落入设计。

4. **[blindspot] 会话过期打断编辑场景无承接** — 原文：UF-2 States 仅「Default | 菜单 + 头像 + 应用切换」「403 | 原地 toast……」。用户在 Dialog 填写一半（策略 JSON、6 字段连接参数）时 401 到达的行为——草稿是否保留、重登后是否回跳原页——全文档无归属。该问题横跨 UX/RC/DI 三维边界，无单一维度对「会话中断的上下文保持」负责。必须改进：定义全局 401 拦截交互（含 Dialog 打开态的行为）。

---

## 修复优先级（给 reviser）

1. **P0** 修正登录成功去向为 /tenant-select 分流（-30 矛盾源）。
2. **P0** 补 UF-2 过期(401) 全局行为、UF-11 Loading/Error、UF-3 受限视图态（4 条缺失状态 + blindspot 4）。
3. **P0** 解决 Secondary Pages：恢复详情/编辑页面 + 返回导航，或在设计中显式提出 Dialog 方案并声明与 PRD 的差异（不可默默降级）。
4. **P1** 消灭 ~10 个孤儿元素（逐一补 Data Binding；UF-6「用户数」改为「活跃会话数」或补来源）。
5. **P1** 补 passkey 注册引导完整 UI 链（发起/有效期/完成/撤销重绑）与 E1 强制改密态；补 UF-4 所属租户/组织。
6. **P1** 新增 Accessibility 小节：键盘导航（树/表格/Dialog focus trap）、label 约定、aria-live 状态播报、图标替代文本、对比度承诺。
7. **P2** 语义色补 success/warning 角色；统一 CRUD 反馈时限口径；补 E6/E7 边界、掩码不回显、树超限、长文本省略、响应式底线（最小分辨率）；落实组件库映射与默认主题决策；落文案契约表。
