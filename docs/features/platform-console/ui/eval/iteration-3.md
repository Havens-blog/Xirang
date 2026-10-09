# UI Design 评分报告 — Iteration 3（最终轮）

- **文档**: `docs/features/platform-console/ui/ui-design.md`
- **需求源**: `docs/features/platform-console/prd/prd-ui-functions.md`（辅以 `prd-spec.md` E1–E12 交叉核对）
- **Rubric**: `ui-web`（1000 分制；4 视角 × 250）
- **评分者**: Senior UX Engineer（scorer-protocol 三阶段；迭代 3 正式评分，最终轮）
- **日期**: 2026-09-29
- **HARD-RULE 声明**: 不给「努力/改进」加分，只按当前页面内容评分。

---

## 总分

| 视角 | 得分 | 满分 | 迭代 2 |
|------|------|------|--------|
| 1. Requirement Coverage（PM） | **234** | 250 | 205 |
| 2. User Experience（终端用户） | **232** | 250 | 206 |
| 3. Design Integrity（设计师） | **229** | 250 | 215 |
| 4. Implementability（开发者） | **230** | 250 | 193 |
| **合计** | **925** | **1000** | 819 |

**一句话结论**: 迭代 2 的 7 组修复优先级中，P0 三组（username 三方矛盾、destructive 对比度、边界场景整层）全部实质落地——其中对比度表经逐组复算全部成立，E6/E7/掩码/溢出/树超限/长度唯一性/零租户壳层从零命中到成建制落进「边界行为约定」「布局与响应式」「状态徽章映射」三个共享小节；悬置交互决策落了三分之二（应用切换器、切换租户草稿策略、布局量化）。剩余差距收窄为两类：**未决项**（即时过滤机制、删除确认强度、UF-2 壳层 Loading、Dialog 提交中、UF-5/6 刷新时限、图标清单）与**微结构性缺口**（badge 变体清单与状态映射不一致、passkey_unsupported 无渲染面、锁定文案语义张力）。

---

## 迭代 2 修复优先级解决核查（7 组）

| # | 迭代 2 优先级 | 现状 | 证据 |
|---|--------------|------|------|
| 1 | P0 username 长度（1–64 vs 正则 3–32 三方矛盾） | **已解决** | UF-4 Data Binding:「username（3–32 字符，正则 ^[a-z][a-z0-9_-]{2,31}$）」——与自身正则、与 PRD 一致 |
| 2 | P0 destructive 对比度断言被证伪 | **已解决（复算通过）** | destructive 改深档 Light #dc2626；「关键组合逐组实测」表 10+ 组对比度逐组复算全部成立（如 #fafafa/#15803d ≈ 4.80:1、#fafafa/#dc2626 ≈ 4.62:1、#f87171/#09090b ≈ 7.19:1），AA 断言不再与 token 表冲突 |
| 3 | P0 边界场景整层（E6/E7/掩码/溢出/树超限/长度唯一性/零租户壳层） | **已解决（全项落地）** | 「边界行为约定（E6/E7）」409 逐字 toast + 重新加载、分页回落/空页回退/过滤重置；UF-7「Bind 密码（仅写入：掩码显示、不回显已存明文，留空 = 不修改）」；表格组件「长文本列…单行截断省略号 + hover title」+ UF-5 树节点同规则；UF-5「树超限禁建…按钮禁用 + tooltip」；org_name/tenant_name/role_name/policy_name 全部落 1–64 + 唯一性 + validation.name_exists；UF-2 新增「受限（零租户）」态（选择器禁用占位、切换器隐藏、菜单仅工作台） |
| 4 | P1 三项交互悬置（应用切换器/即时过滤/确认强度） | **1/3 解决** | 应用切换器已决「新标签打开 e-cam-web（共享会话）」✓；即时过滤仍为「列表即时过滤」未明前端/后端 ✗；删除/禁用确认强度仍未分级 ✗ |
| 5 | P1 状态补齐（UF-2 Loading/切换租户行/UF-5 右面板/UF-6 分页/UF-9 pending/Dialog 提交中） | **3/6 解决** | 切换租户行已补进 UF-2 且草稿策略与 401 对齐 ✓；UF-6 分页器绑定「page / page_size / total（与 UF-4 同源口径，边界行为见 E7）」✓；UF-9「撤销 pending」态（徽标「生效中」+ 按钮禁用 + aria-live）✓；UF-2 壳层 Loading ✗、UF-5 右面板加载态 ✗、Dialog 提交中 ✗ |
| 6 | P1 设计系统（success/warning/阴影例外/刷新时限统一） | **2/3 解决** | success/warning token + 「状态徽章映射（跨页统一）」三态实测表 ✓；「对话框阴影为唯一显式例外」✓；UF-5/6 反馈仍为「树刷新」「刷新」无时限 ✗ |
| 7 | P2（契约扩容/租户草稿/我的信息/图标清单/布局量化） | **4/5 解决** | 契约 16 key + 口径收窄声明（「字段级校验文案…不在本契约内，由各表单 aria-describedby 就近维护」）+ auth.mfa_invalid 补入 ✓；租户切换「丢弃未提交草稿并关闭 Dialog（与 401 策略一致…数据重取，非整页 reload）」✓；「我的信息 \| profile {username, display_name, tenant_name, roles[]}」✓；布局量化（UF-5 双栏 2:3 + 树最小宽 280px、工作台列数两断点、最小分辨率 1280×720、表格最小可用宽 960px）✓；图标清单/Icon Keyword 落表 ✗ |

**结论**: 7 组中 4 组全解、3 组部分解决；本轮未发现修订新引入的事实性错误（对比度与唯一键约束经复算均成立）。

---

## Phase 1 — Reasoning Audit（独立推理审计）

**问题 → 方案**: 11 个 UF ↔ 11 个 Component 一一对应，Placement 与 PRD Page Composition 相符；UF-1 →（强制改密）→ UF-11 分流 → UF-3 / UF-2 的主链自洽；401 / 切换租户两个上下文销毁场景现已同策（「与 401 策略一致，v1 不保留编辑上下文」）。

**方案 → 证据**: 全文档交叉复算未发现新引入的事实错误——

- 状态徽章映射表 6 组对比度逐组重算（success/warning/destructive × 明暗 + disabled 徽章）全部与声明吻合（误差 ≤0.1）。
- username/tenant_name/role_name/policy_name 四处长度标注与 PRD 逐字一致，唯一性契约 key 全部有引用方。
- 契约 16 key 与引用组件逐条核对：forbidden、save.conflict、tenant.disable_blocked、delete.blocked_org/role/policy、grant.duplicate 均逐字匹配，迭代 2 的「UF-2 toast『无权限』≠ 契约」措辞冲突已消除（「原地 toast（逐字见文案契约 forbidden）」）。

**证据 → 成功标准**: 设计无独立成功标准，无冲突。

**SC/状态簇一致性核查**（对 States/Interactions/Data Binding 做簇内双向推导）:

- UF-11 零租户「受限工作台（即 UF-3 受限视图态）」↔ UF-3 已定义 → 满足。
- UF-6「活跃会话数列 active_session_count…与禁用校验同一字段」↔ 禁用被拒「{n} = active_session_count，与表格列同源」→ 同源满足。
- 「状态徽章映射（跨页统一）」↔ 组件规范「徽章：default/secondary/outline/destructive 四变体」→ **冲突**：映射表要求 success/warning 实底徽章，而变体清单不含 success/warning 变体——按字面实现，「启用/锁定」徽章无合法变体可用，正撞自家 Don't「硬编码色值」（转入 DI）。
- E7「UF-4/UF-6 分页表格共用」↔ PRD 仅 UF-4/UF-6 声明分页 → 范围一致，满足。
- UF-9「撤销 pending…生效后刷新列表移除该行」→ pending 为客户端瞬态，≤30s 窗口内刷新/离开后该行的呈现与重复撤销反馈未定义（转入 RC 边界）。

**Phase 1 前置锚点**: 无新矛盾锚点；badge 变体冲突与 pending 刷新缺口分别计入 DI / RC。

---

## Phase 2 — Rubric Scoring

### 1. Requirement Coverage — 234/250（PM 视角）

**UI function coverage: 76/80**

- UF-1～UF-11 全覆盖；路由与 Page Composition 一一相符；passkey 链（发起/15 分钟单次会话/完成态/撤销重绑 + 唯一登录方式走重置密码）、E1 强制改密态、Secondary Pages 均在位。
- 扣分（-4）：UF-7「用户属性映射（用户名属性必填）」仍未枚举 PRD 的「用户名/邮箱/DN 属性名」子字段集，binding 仅 attr_map 一笔带过（-2）；只读/审计「不渲染增删改入口」仍仅在 UF-2 导航层声明一次，UF-4/5/7/9 页内 Interactions 全为无条件行且无交叉引用（-2）。

**Navigation Architecture coverage: 36/40**

- Primary 8 项全列于 UF-2 且四档角色裁剪口径完整转录；3 个 Secondary Pages 带 Entry/Return ✓。
- 扣分（-4）：PRD 8 个 Icon Keyword（home/user/org/tenant/ldap/role/grant/policy）未逐一落表，设计仅写「图标+文字」（-2）；UF-1 帮助入口「help_url（v1 指向运维支持页）」指向一个**未在 PRD Page Composition 或本设计中定义的页面**，违反 PRD Navigation Rules「Every navigation target must correspond to a page defined in this document」——运维支持页是站内路由还是外链未决（-2）。

**State requirement coverage: 78/80**

- 逐 UF 核对 PRD States 表全部着落；UF-2 受限态、UF-9 pending 态为本轮新增覆盖。
- 扣分（-2）：UF-1 的「passkey 不可用」仍未作为 States 行存在（PRD 有此状态行），仅散在 Interactions（「WebAuthn 不可用隐藏 passkey 入口」）与契约 key；行为已定义、状态表缺行。

**Edge case handling: 44/50**

已覆盖（grep 复核全命中）：E1/E2/E9/E3/E4/E8/E11×3/E12、E6 409、E7 分页回落、掩码不回显、长文本溢出（表格+树）、树超限禁建、四处长度/唯一性约束、零租户壳层、空凭据拦截、401 丢草稿、授权三元组预检、策略 JSON 双层校验、租户禁用会话数校验。

未覆盖（-6）：
- **重复提交防护**（-4）：各管理页 Dialog 保存无提交中禁用/防双击约定（UF-1 按钮有 loading，其余弹窗无）——「并发动作」在 UI 层未建模，仅 E6 覆盖了数据层冲突。
- **UF-9 pending 态的刷新复活缺口**（-2）：「行内徽标『生效中』、操作按钮禁用」是客户端瞬态；≤30s 窗口内用户刷新或离开再回，行恢复 Populated 常态且「撤销」重新可点，对已提交撤销的授权再次点撤销的反馈未定义。

**小计**: 76 + 36 + 78 + 44 = **234/250**。

### 2. User Experience — 232/250（终端用户视角）

**Information hierarchy: 74/80**

- 页面骨架扫描动线一致（页面头+主操作 → 过滤区 → 主表格）；「我的信息」字段集已落地（profile 四字段）；表格长文本截断规则落地，扫描节奏保护成立。
- 扣分（-6）：表格列优先级与默认列宽未定义——截断规则有了，但哪些列优先、宽列如何分配仍靠实现者判断（-3）；表格字号未声明，16px 正文直接用于多列高密度表格偏大且与「紧凑后台」定位不符（-3）。

**Interaction intuitiveness: 72/80**

- 惯例模式选型正确且悬置决策大量收敛：应用切换器新标签保留控制台上下文、租户切换草稿丢弃与 401 同策、409「重新加载」按钮流、E7 分页回落、树键盘三键位、单租户 toast 自动分流、Dialog focus trap/Esc/焦点归还。
- 扣分（-8）：删除/禁用确认强度仍未分级——「删除确认」是单击还是输入对象名，破坏性操作的摩擦度未设计（-3）；「列表即时过滤」前端当前页过滤还是服务端查询仍未明示，分页场景下二者体验完全不同（E7「切换过滤/搜索条件时 page 重置为 1」暗示服务端但未写明）（-3）；UF-5「树刷新」/UF-6「刷新」与 UF-4/UF-8「≤2s 刷新」同类操作不同承诺（-2）。

**Accessibility: 86/90**

- 对比度表逐组复算全部成立（本轮核心修复）；状态徽章映射自带实测列；aria-live 覆盖锁定倒计时与「撤销生效（≤30s）」；focus trap/焦点归还/登录 Tab 顺序/icon-only aria-label 均在位。
- 扣分（-4）：表格语义化（th scope/caption）与分页器键盘可达性未提（-2）；toast 出现时焦点策略未提（-1）；长文本全文揭示「hover title 显示全文」是纯悬停机制，键盘用户与读屏用户的等价揭示手段未定义，与「全部交互可 Tab 到达」的承诺存在张力（-1）。

**小计**: 74 + 72 + 86 = **232/250**。

### 3. Design Integrity — 229/250（设计师视角）

**Design system adherence: 74/80**

- success/warning 语义色补齐且「仅用于状态与结果反馈、不用于装饰」的用途纪律明确；状态徽章三态映射带实测对比度；对话框阴影已声明为「无阴影系统」的显式例外；语义 Token 映射与主题策略维持。
- 扣分（-6）：**badge 变体清单与状态映射冲突**——组件规范「default/secondary/outline/destructive 四变体」vs 映射表要求「success/warning 实底」徽章，变体族缺 success/warning 两员，「状态列着色一律按映射、不硬编码」在现有变体集内不可满足（-3）；字体仍双家族并行（正文系统栈 vs 标题 Inter）无 rationale 声明（-3）。

**Visual coherence: 84/90**

- 「状态」列跨页同一视觉达成（映射表 + UF-6/UF-7 例外显式声明）；全站共享单一主布局；两级圆角（6/8px）、4px 间距基准、无阴影体系统一。
- 扣分（-6）：八个菜单图标与页内 icon-only 按钮的图标集来源未声明（-2）；反馈时限口径仍分裂（UF-5/6 无时限 vs UF-4/8「≤2s」）（-2）；UF-9「行内徽标『生效中』」、UF-4「已绑定 passkey」标记等同类 inline 徽标未纳入「跨页统一」映射，实现者将自行发明样式（-2）。

**State completeness: 71/80**

- 通用状态迁移 + UF-3/UF-11 专属迁移 + UF-9 pending → 生效 → 移除行的转换链完整；E6 409 的恢复路径（保留输入 → toast → 重新加载 → 服务端数据覆盖）是全文档最完整的一次状态转换描述。
- 扣分（-9）：UF-2 壳层仍无 Loading 态——菜单按角色异步加载期间头部显示什么未定义（-4）；UF-5 右侧详情/成员面板按选中节点加载，其加载/错误形态未定义（树有 skeleton、右面板没有）（-2）；各管理页 Dialog 提交中状态未定义（-3）。

**小计**: 74 + 84 + 71 = **229/250**。Happy-path only（-50）不触发；跨页共享导航不一致（-30）不触发。

### 4. Implementability — 230/250（开发者视角）

**Layout specificity: 74/80**

- 布局量化到位：UF-5 双栏定比 2:3 + 树最小宽 280px、工作台列数仅两个决策断点、最小分辨率 1280×720、表格容器横滚 + 最小可用宽 960px、header h-14、登录 max-w-sm、对话框动效参数。
- 扣分（-6）：Dialog 尺寸档位未定义——策略 JSON 编辑器与 6 字段 LDAP 表单显然不同宽，唯一宽度约定缺位（-3）；表格默认列宽/列宽分配未定义，截断规则无法反推列布局（-3）。

**Data binding explicit: 79/80**

- 绑定质量高：passkey_session 带结构、org 详情带字段集、「与 UF-4 同源口径」「{n} 与表格列同源」等来源声明消除歧义；UF-6 分页器缺口已补。
- 扣分（-1）：UF-7 attr_map 为单行绑定，PRD 的「用户名/邮箱/DN 属性名」三个子键无对应绑定行，6 字段表单中该字段的内部结构需实现者猜。

**Interaction unambiguity: 75/90**

- 大面积收敛：应用切换器新标签 ✓；UF-2 补「切换租户」行且草稿策略与 401 对齐 ✓；E6 409 链（逐字 toast + 重新加载 + 覆盖规则）✓；E7 回落规则可编码 ✓；掩码「留空 = 不修改」语义明确 ✓；passkey 三链完整 ✓；UF-9 pending→生效→移除 + aria-live ✓；UF-6 禁用前置「校验未过期会话数=0 → 确认」+ 逐字文案 ✓。
- 扣分（-15）：「列表即时过滤」机制未明（前端/后端）（-5）；删除/禁用确认强度未分级，「删除确认」一词不可编码（-3）；UF-5「新增子组织/编辑/删除 | 弹窗操作 | 树刷新」弹窗字段集仍未明示（仅名称一字段还是含更多，未写）（-2）；UF-5/UF-6 反馈「刷新」无时限，完成判定不可验证（-2）；UF-1「剩余锁定时间 lock_remain | eiam 响应」与契约「X 逐分钟刷新」之间缺刷新机制（轮询 or 由 deadline 客户端计时），绑定值语义不足（-1）；UF-7「测试连接 | 调 eiam 测试 | 成功/失败提示」超时与进行中形态未定义（-2）。

**小计**: 74 + 79 + 75 = **230/250**。

---

### 跨维度一致性核查

- 「明暗两主题均满足 WCAG AA」↔ token 实测表 → 一致 ✓（本轮复算通过）。
- 「状态徽章映射（跨页统一）」↔ 组件规范 badge 四变体 → 冲突（DI -3）。
- E7 共用范围 ↔ PRD 分页声明范围 → 一致 ✓。
- 401 与切换租户草稿策略 → 一致 ✓。
- 帮助入口 target「运维支持页」↔ PRD「Every navigation target must correspond to a page defined in this document」→ 冲突（RC -2）。
- 「hover title 显示全文」↔「全部交互可 Tab 到达」→ 张力（UX -1）。

---

## Phase 3 — Blindspot Hunt

1. **[blindspot] auth.passkey_unsupported 有契约文案、无渲染面** — 原文：UF-1 Interactions「WebAuthn 不可用**隐藏** passkey 入口」+ 契约「auth.passkey_unsupported | 当前浏览器不支持 Passkey，请改用密码或 LDAP 登录 | E8」。PRD 要求「隐藏 passkey 入口，**提示改用密码/LDAP**」——入口既被隐藏，这条逐字文案在页面上没有任何声明过的渲染位置（登录卡片 notice？toast？）。契约自称「实现不得改写措辞」，但第一个问题是无处可写。Reasoning audit flagged this independently of dimension scoring. 必须改进：为 E8 定义提示条的渲染位置与形态（如登录卡片顶部中性 notice）。

2. **[blindspot] 锁定文案的语义张力：账号级 vs 方式级锁定未决** — 原文：契约「auth.account_locked | 尝试次数过多，**账号已锁定**，剩余 {X} 分钟」+「副行『可改用其他已启用的登录方式』」。文案断言「账号已锁定」（账号级），副行却承诺其他登录方式可用（方式级）——若锁是账号级，副行误导用户；若锁是方式级（MFA 失败锁 MFA），「账号已锁定」措辞失实。设计逐字转录了 PRD 的上游张力，但 UI 规格是文案的最后守门人，实现时必然要二选一。Reasoning audit flagged this independently of dimension scoring. 必须改进：与 eiam 确认锁定作用域，改写为「登录方式已锁定」或删去副行。

---

## 修复优先级（最终轮遗留清单，供排期参考）

1. **P1** 补齐三处未决交互：即时过滤机制（前端/后端）、删除/禁用确认强度分级、UF-5/6 刷新时限统一为「≤2s」。
2. **P1** 补齐状态缺口：UF-2 壳层 Loading、UF-5 右面板加载/错误态、各 Dialog 提交中 + 重复提交防护、UF-1 passkey 不可用 States 行、UF-9 pending 刷新复活行为。
3. **P2** badge 变体清单补 success/warning 两变体（或声明映射表变体独立于四变体）；UF-9「生效中」/UF-4「已绑定 passkey」徽标纳入映射。
4. **P2** 两个 blindspot：passkey_unsupported 渲染面、锁定作用域与文案对齐。
5. **P2** 图标集来源 + 8 个 Icon Keyword 落表；表格字号/列宽/Dialog 尺寸档位量化；帮助入口 target 定义（站内页或外链二选一）；attr_map 拆三子键绑定。
6. **P3** a11y 尾巴：表格 th scope/caption、分页器键盘可达、toast 焦点策略、截断全文的键盘等价揭示。
