---
created: "2026-10-09"
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: WatchAlert 子项目并入（含 eiam 统一鉴权）

> 设计层：定义界面「长什么样、怎么交互」，与需求层（[prd-ui-functions.md](../prd/prd-ui-functions.md)）分离。
> 平台：**web**。三个 UI 功能均为**对既有 WatchAlert 前端页面的改造**（existing-page 模式），不新建完整页面。

## Design System

选用 **Shadcn**（zinc 中性、功能性极简，适用于运维告警管理后台）。核心令牌内联如下，全部通过 CSS 变量实现 light/dark 切换，组件禁止硬编码色值。

### Visual Theme & Atmosphere

干净、功利，无装饰——每个视觉元素都服务于功能。表面扁平、1px 细边框。美学目标是「隐形 UI」：用户聚焦内容而非 chrome。默认浅色，支持暗色切换。

### Color Palette（Light）

| Role | Variable | Value |
|------|----------|-------|
| Background | `--background` | #ffffff |
| Foreground | `--foreground` | #09090b (zinc-950) |
| Card | `--card` | #ffffff |
| Card Foreground | `--card-foreground` | #09090b |
| Primary | `--primary` | #18181b (zinc-900) |
| Primary Foreground | `--primary-foreground` | #fafafa |
| Secondary | `--secondary` | #f4f4f5 (zinc-100) |
| Muted | `--muted` | #f4f4f5 |
| Muted Foreground | `--muted-foreground` | #52525b (zinc-600) |
| Destructive | `--destructive` | #ef4444 |
| Border | `--border` | #e4e4e7 (zinc-200) |
| Input | `--input` | #e4e4e7 |
| Ring | `--ring` | #18181b |

> **对比度修正**：`--muted-foreground` 由 zinc-500 (#71717a，对 `--muted` 背景 4.4:1，不满足 WCAG AA 4.5:1) 上调至 zinc-600 (#52525b，对 `--muted` 背景 5.74:1，通过 AA)。该值适用于表头文字、Component 3 只读说明条、引导文案等所有 muted 文字场景。

> **语义色例外**：宽限横幅与错误态属「语义状态色」，允许在 zinc 中性基础上引入**受限**的告警色（amber 黄 / destructive 红），不进入通用色板。规范如下：宽限横幅用 `amber-50` 底 / `amber-800` 文字 / `amber-600` 1px 边（light）；会话过期/旧账号 404/写操作被拒提示用 `--destructive` 红。其余一律回归 zinc 中性，禁止混搭蓝/绿等冷暖非语义色。

### Color Palette（Dark）

暗色模式通过 `[data-theme="dark"]`（手动，源自 TECH-frontend-state-001 持久化键 `hc_theme`）**优先于** `@media (prefers-color-scheme: dark)`（系统）重写同名 CSS 变量，组件零改动。优先级规则：`hc_theme` 存在时以手动选择为准（覆盖系统偏好），`hc_theme` 缺省时回退系统 `prefers-color-scheme`——避免「系统暗色 + 手动切浅色」失效的 OR 语义陷阱。CSS 实现：`[data-theme="dark"]` / `[data-theme="light"]` 属性选择器写在 `@media (prefers-color-scheme)` 规则之后（或以更高特异性），确保手动覆盖在层叠中胜出。所有暗色值已验证 WCAG AA 对比度 ≥ 4.5:1。

| Role | Variable | Value |
|------|----------|-------|
| Background | `--background` | #09090b (zinc-950) |
| Foreground | `--foreground` | #fafafa (zinc-50) |
| Card | `--card` | #18181b (zinc-900) |
| Card Foreground | `--card-foreground` | #fafafa |
| Primary | `--primary` | #fafafa |
| Primary Foreground | `--primary-foreground` | #18181b |
| Secondary | `--secondary` | #27272a (zinc-800) |
| Muted | `--muted` | #27272a |
| Muted Foreground | `--muted-foreground` | #a1a1aa (zinc-400，对 #27272a 背景 5.89:1) |
| Destructive | `--destructive` | #f87171 (red-400) |
| Border | `--border` | #27272a (zinc-800) |
| Input | `--input` | #3f3f46 (zinc-700) |
| Ring | `--ring` | #d4d4d8 (zinc-300) |

> **暗色语义色例外**：宽限横幅暗色用 `amber-950` (#1c1917) 底 / `amber-200` (#fde68a) 文字 / `amber-500` (#f59e0b) 1px 边（amber-200 对 amber-950 对比度 10.3:1）；destructive 文字用 `--destructive` (#f87171，对 #09090b 背景 5.91:1)。

### Typography

| Role | Font | Weight | Size |
|------|------|--------|------|
| H1 | Inter / System | 700 | 36px |
| H3 | Inter / System | 600 | 24px |
| Body | Inter / System | 400 | 16px |
| Small | Inter / System | 500 | 14px |
| Muted | Inter / System | 400 | 14px |
| Mono | JetBrains Mono | 400 | 14px |

字体栈：`-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif`。

### Components

- **Button**：默认 primary（`--primary` 底 / `--primary-foreground` 字），rounded-md（6px）、h-10、px-4、text-sm、font-medium；hover 仅 opacity 变化、无阴影、150ms 过渡。变体：secondary / outline / ghost / destructive。
- **Card**：bg `--card`、rounded-lg（8px）、`1px solid var(--border)`、内边距 p-6、无阴影。
- **Card（compact 变体）**：同 Card，但内边距 p-4；用于信息密度高的概览/统计卡片（如 Component 3 告警概览），与通用 Card 形成可预期的「紧凑展示 vs 标准表单/详情」对照，非随意取值——这是对 p-6 基准的显式有依据偏离，不是遗漏。
- **Badge**：rounded-full、px-2.5、text-xs；变体 secondary / destructive / outline。
- **Table**：干净网格，表头 bg-muted，行 hover bg-muted/50；仅行间 `border-b`，无外框。
- **Banner（宽限横幅专用）**：全宽条带，1px 底边（amber-600），圆角 0（贴顶）。

### Do's and Don'ts

| Do | Don't |
|-----|------|
| zinc 中性为主 + 语义状态色例外 | 在通用 UI 混搭 amber/blue 冷暖非语义色 |
| 全部 1px 细边框 | 粗/双层边框 |
| 输入 rounded-md、卡片 rounded-lg | 随意混用圆角值 |
| CSS 变量实现 dark mode | 硬编码色值 |
| disabled 用 opacity-50 | 用 display:none 隐藏禁用态 |

### Icons

图标库统一使用 **Lucide React**（`lucide-react`，MIT，与 Shadcn 原生配套，tree-shakeable）。所有图标 `size={18}`、`strokeWidth={1.5}`、颜色继承 `currentColor`，禁止硬编码 stroke 色。导航与语义图标映射如下：

| 语义关键词 | Lucide 图标名 | 用途位置 |
|-----------|--------------|---------|
| 告警/alert | `bell` | 顶部导航「告警看板」入口 + 概览卡片标题 |
| 规则/rule | `shield` | 顶部导航「告警规则」入口 + 规则列表标题 |
| 故障/event | `activity` | 顶部导航「故障中心」入口 + 事件时间线标题 |
| 鉴权降级 | `alert-triangle` | Component 2 宽限横幅图标 |
| 只读 | `lock` | Component 3 只读说明条图标 |
| 无权限 | `ban` | Component 4 403 占位卡标题图标 |
| 登录跳转 | `log-in` | Component 1 主按钮（图标右置） |

### Motion & Transitions

所有过渡统一使用 `cubic-bezier(0.4, 0, 0.2, 1)`（标准 ease）缓动函数。具体时长：

| 场景 | 时长 | 属性 |
|------|------|------|
| Button hover | 150ms | opacity |
| Banner 展开/收起（height 0↔auto） | 200ms | height, opacity |
| 403 占位卡淡入 | 150ms | opacity, transform(translateY 8px→0) |
| 只读视图入场 | 200ms | opacity |
| 只读说明条展开 | 200ms | height, opacity |
| 状态切换（Loading→Populated 等） | 150ms | opacity |

> **Banner 高度动画实现**：`height: 0 ↔ auto` 非 CSS 可直接 transition 的值。采用「测量高度 + max-height」技术——展开前以 `scrollHeight` 测量目标高度并设为 `max-height`，过渡结束后清除 `max-height` 改回 `auto`（允许内容自适应后续变化）；收起反向操作。`opacity` 与 `max-height` 同步过渡 200ms。不使用 FLIP（无位移变换需求）。`prefers-reduced-motion: reduce` 时跳过测量、瞬切显隐。
>
> 所有动画在 `@media (prefers-reduced-motion: reduce)` 下降级为瞬时切换（duration: 0ms），仅保留透明度变化不保留位移。

### Accessibility

全局无障碍规范，所有组件必须遵守：

1. **跳转导航**：每个页面首个可聚焦元素前提供「跳到主内容」skip link，`href="#main-content"`，默认 `sr-only`，聚焦时可见。
2. **焦点管理**：
   - 状态切换时移动焦点至新内容容器（banner 展开后焦点移至 banner 容器；403 渲染后焦点移至 403 卡片标题）。
   - Component 1 Loading 态：焦点保持在主按钮，`aria-busy="true"`。
   - 自动跳转（Component 3 已恢复）前不抢焦点，跳转由浏览器接管。
3. **键盘导航**：所有交互元素可 Tab 聚焦，可见 focus ring（`outline: 2px solid var(--ring); outline-offset: 2px`）；禁用元素设 `tabindex="-1"` + `aria-disabled="true"`。ESC 无特殊行为（无 modal/drawer）。
4. **ARIA / Live Regions**：
   - Component 2 宽限横幅容器：`role="alert"` + `aria-live="assertive"`（降级状态需立即通告）。
   - Component 2 恢复收起：通过同一容器的 `aria-live` 区域通告「鉴权服务已恢复」后移除。
   - Component 1 Loading spinner：`aria-label="正在跳转安全登录"`（spinner 为装饰性图标时用 `aria-hidden="true"`，由按钮 `aria-busy` 表达状态）。
   - Component 3 只读说明条：`role="status"` + `aria-live="polite"`。
   - Component 4 403 占位卡：渲染时 `aria-live="polite"` 通告。
   - Component 3 数据加载/空态切换：`role="status"` + `aria-live="polite"`。**Error 态为该规则的特例覆盖**：升级为 `role="alert"` + `aria-live="assertive"`（请求失败需立即通告，不等待用户空闲）；Error 解除回到 Populated 时降级回 `polite`。Component 3 States 表 Error 行的 `assertive` 即此覆盖的具象化，二者不冲突——通用规则管 Loading/Empty/Populated，Error 行管请求失败态。
5. **语义化结构**：导航用 `<nav aria-label="主导航">`，主内容区 `<main id="main-content">`，横幅用 `<aside>`，表格保留 `<table>` + `<thead>`/`<tbody>` + `<th scope>`。

---

## Navigation & App Shell

共享应用壳（除登录页外的所有页面继承），来自 PRD `Navigation Architecture`：

- **顶部导航栏**（top nav bar）：品牌名「WatchAlert」+ 三个一级导航入口 `告警看板 / 告警规则 / 故障中心`，当前页高亮（下划线或背景强调）。
- **横幅插槽（banner slot）**：导航栏正下方、页面内容上方，全宽常驻插槽，供「鉴权不可用宽限横幅」（组件 2）按需占位或隐藏。
- **响应式**：断点 `≥1024px` 桌面全宽 / `768px–1023px` 平板紧凑 / `<768px` 移动端折叠。<768px 时一级导航折叠为汉堡菜单（`aria-expanded` 跟随切换，展开面板首个链接自动聚焦）。

### 全局响应式行为

| 组件 | ≥1024px | 768px–1023px | <768px |
|------|---------|-------------|--------|
| 顶部导航 | 全宽，三入口平铺 | 全宽，三入口平铺（间距收窄） | 汉堡菜单折叠 |
| 宽限横幅（Component 2） | 主文案 + 副文案同一行 | 主文案单行 + 副文案换行 | 主文案截断至 2 行（`-webkit-line-clamp:2`），副文案折叠为「详情」可展开；徽标右浮 |
| 只读说明条（Component 3） | 单行全宽 | 单行全宽 | 文案可换行，`text-align:left`，图标 + 文字一行 |
| 只读看板内容（Component 3） | 3 列网格（见 Component 3 布局） | 2 列网格 | 单列堆叠，表格水平滚动（`overflow-x:auto`） |
| 403 占位卡（Component 4） | max-w-[460px] 居中 | max-w-[460px] 居中 | full-width，`px-4`，卡片 `mx-4` |
| 登录页卡片（Component 1） | max-w-[400px] 居中 | max-w-[400px] 居中 | full-width，`mx-4`，`min-h-screen` flex 不变 |
- **二级页**：`eiam 登录跳转页` 由 UF-1 通过浏览器整页跳转实现（无站内二级路由），浏览器「返回」为回退机制。

> 原型中所有共享导航 HTML 在各页面必须字节一致（仅 active 态不同），标签与目标严格对齐 PRD Navigation Architecture 表。

---

## Component 1: eiam 统一登录入口（UF-1）

### Placement

- **Mode**: existing-page
- **Target**: 登录页（原「本地账密登录」页改造）
- **Position**: 居中单卡片，全屏 `min-height: 100vh` flex 垂直·水平居中。**移除**本地用户名输入框、密码输入框、「登录」提交按钮与「记住我/忘记密码」等本地认证相关元素；替换为品牌区 + 单一「使用平台账号登录」主按钮 + 引导文案。

### Layout Structure

```
<body class="login-body">                       // min-h-screen, flex center
  └─ <main class="login-card" id="main-content">  // Card, max-w-[400px], p-6
       ├─ CardHeader                             // 品牌区
       │    ├─ 品牌名「WatchAlert」              // H3, 居中
       │    └─ 副标题「运维告警平台」             // Muted, 居中
       ├─ CardContent                            // 登录动作区
       │    ├─ Button「使用平台账号登录」(primary, 全宽, Lucide log-in 图标右置)
       │    └─ 引导文案「请使用运维平台统一账号（eiam）登录」  // Muted、text-sm、居中
       └─ (按状态条件渲染的错误区 ─ 见 States)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default（未登录） | 品牌区 + 主按钮 + 引导文案 | 点击主按钮 → 浏览器整页跳转 eiam 登录（携带回跳地址参数） |
| Loading（跳转中） | 主按钮禁用（`aria-disabled="true"` + `aria-busy="true"`）并显示「正在跳转安全登录…」 + 内联 spinner（`aria-hidden="true"`，状态由按钮 `aria-busy` 表达） | 跳转发起后锁定按钮，防重复点击；不阻塞浏览器跳转 |
| Error（登录失败） | 卡片顶部插入 destructive 红字提示条「登录失败，请重试」+ 重试按钮与主按钮并存 | eiam 回跳错误时展示统一文案（不泄露用户名是否存在），可重试 |
| 403 无权限（认证成功但无内部角色） | **不渲染在登录页**：渲染在受保护页「原地」的全宽占位（见组件 4） | 保持当前页不跳登录页（BIZ-auth-003） |
| 配置缺失（`eiamAuthUrl` 为 null） | 卡片顶部插入 destructive 提示「登录服务未配置，请联系管理员」+ 主按钮禁用（`aria-disabled="true"` + `tabindex="-1"`，不可点击） | 前端启动时校验配置下发；缺省不渲染可点击的登录按钮，避免死链跳转。后台每 30s 重试拉取配置，命中后回 Default |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击「使用平台账号登录」 | 拼接 eiam 授权地址（携带 `redirect_uri` = 当前页 URL）并整页跳转 | 按钮进入 Loading 态 |
| eiam 认证成功回跳 | 校验回跳地址白名单 → 建立会话 → 进入告警看板 | 页面跳转至原访问页面 |
| eiam 返回失败 | 停留登录页 | 顶部红色错误提示，可重试 |
| eiam 认证成功但无 `watchalert_role_binding` | 渲染组件 4（403 占位） | 403 原地提示，不跳登录页 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 登录状态（是否跳 eiam） | `isAuthenticated: boolean` | 会话（前置检测） |
| 回跳地址 | `redirect_uri: string` | 当前页 URL（经白名单校验） |
| eiam 登录地址 | `eiamAuthUrl: string \| null` | 后端配置下发；缺省时为 `null`，触发「配置缺失」态（见 States） |

### Validation Rules（对应 PRD UF-1）

- 登录页**不得保留任何本地用户名/密码输入框**。
- 回跳地址白名单/清洗：仅 `host` 匹配告警平台域名、`path` 以 `/` 开头的相对路径；禁止绝对外链、`//` 开头协议相对 URL、`javascript:`/`data:` scheme；校验失败回退到告警看板默认路径。

---

## Component 2: 鉴权不可用宽限横幅（UF-2）

### Placement

- **Mode**: existing-page
- **Target**: 全站顶部（**所有页面共享**，登录页除外——登录页走 UF-1 的失败态）。
- **Position**: 顶部导航栏正下方、页面内容上方的**全宽常驻插槽**。无横幅时插槽不占高（`height:0`），出现时展开固定高度条带。

### Layout Structure

```
<nav class="nav-bar">…</nav>                    // 顶部导航（app shell）
<aside class="banner-slot" role="alert" aria-live="assertive">  // 全宽插槽，height 0 或 auto
  └─ <div class="degraded-banner">               // amber 语义色条带，h-14(56px)、px-6、flex items-center gap-3
       ├─ Lucide alert-triangle (size 18, amber-600) + 主文案「鉴权服务暂不可用，已切换宽限模式(只读)」  // 左对齐、font-medium、text-sm
       ├─ 副文案「剩余 X 小时 X 分；此期间告警规则/通知等写操作不可用」 // Small、muted
       └─ Badge「只读」(secondary variant，ml-auto)              // writeAllowed=false 时显示
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 正常（探针偶发 1-2 次失败） | 无横幅（插槽高 0） | 对用户不可见，页面正常使用 |
| 宽限中 | 黄色（amber）横幅展开 | 写操作按钮整体禁用（opacity-50 + cursor-not-allowed），只读可继续 |
| 已恢复 | 横幅收起（滞回：探针连 3 次成功移除） | 写按钮恢复可用；不弹「恢复」toast，静默回正 |
| 宽限过期 | 横幅消失，页面切至只读免登录视图（组件 3） | 2 小时窗口到期，整页降级 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 后端探针信号（SSE/轮询）判定 eiam 不可用 | 横幅展开、倒计时启动 | 黄色横幅 + 剩余时长展示 |
| 用户点击已禁用的写按钮（编辑规则/确认告警/改通知） | 前端拦截并展示写拒绝提示 | 内联提示「鉴权服务暂不可用，操作已被禁止」，不发起请求 |
| 后端写接口返回 503 | 前端 HTTP 拦截器（全局 axios/fetch wrapper）捕获 503 响应码 | toast「鉴权服务暂不可用，操作失败」 |
| 探针连 3 次成功判定恢复 | 横幅收起、写按钮恢复 | 静默回正，无额外提示 |

> **503 捕获机制说明**：eiam 不可用时，后端所有写接口（规则增删改、告警确认、通知配置）统一返回 HTTP 503。前端触发 503 捕获的场景：(a) 用户在横幅展示前（探针滞回窗口内）发起写请求——此时写按钮尚未被 `writeAllowed=false` 禁用，请求到达后端后被拒。该 503 捕获由全局 HTTP 拦截器统一处理，不依赖按钮禁用逻辑，与「前端拦截写按钮」是互补的双保险，非互斥。跨标签页登出（BIZ-auth-005）场景在其余标签页的下一请求返回 **401** 而非 503，归入下表 401 分派，不混入 503。

### SSE/Polling 传输层故障处理

探针信号依赖 SSE/轮询传输，传输层自身的连接失败与 eiam 健康判定是**正交**的两类故障：

| 传输故障 | 判定条件 | 前端行为 |
|---------|---------|---------|
| SSE 连接断开 / 轮询请求超时 | EventSource `onerror` 或轮询 fetch 连续 2 次超时（各 10s） | 前端**假定 eiam 健康不变**（维持上一个已知 `authHealth` 值），横幅不主动展开；同时启动指数退避重连（初始 2s，上限 30s，`jitter ±20%`） |
| 重连成功且探针信号正常 | 连续 2 次成功探针 | 恢复正常监听，清除重连计时器 |
| 传输中断 > 60s 且上一状态为 normal | 持续重连失败超过 60s | 横幅展开并文案改为「鉴权服务连接中断，写操作已暂停」，写按钮禁用（保守降级）；继续重连 |
| 传输中断 > 60s 且上一状态为 degraded | 持续重连失败超过 60s | 维持 degraded 横幅 + 副文案追加「（连接中断，恢复时间未知）」 |

> 传输层故障不触发 Component 3 只读降级（只读降级仅由「宽限 2h 窗口到期」驱动），避免因网络抖动导致整页降级。

### 多标签页状态陈旧窗口（BIZ-auth-005 / TECH-frontend-state-004）

跨标签页不做实时广播（无 WebSocket / BroadcastChannel）。Tab A 登出 / eiam 恢复后，Tab B 的 `authHealth`、`writeAllowed`、横幅显隐均以**本标签页下一次探针 tick** 为感知上限。陈旧边界与对账规则：

| 量 | 陈旧上限 | 对账方式 |
|----|---------|---------|
| `authHealth` / 横幅显隐 | ≤ 探针周期 10s + 滞回 3 次 ≈ 30s | 下一次探针收敛；陈旧期间 Tab B 可能仍展示已失效的横幅或仍允许点击写按钮（请求到达后端被 401/503 拦截，见全局 HTTP 错误分派表，不依赖按钮态） |
| `graceRemaining`（宽限剩余时长） | 不漂移 | 后端统一下发**绝对剩余时长**，前端仅展示不下发本地计秒，故跨标签页倒计时一致；仅「展示与否」可能滞后 ≤30s |
| 会话有效性 | 下一请求 | 跨标签页登出后 Tab B 下一请求 401，由全局拦截器分派跳登录页（BIZ-auth-005） |

> 不做跨标签页状态对账（无快照、无广播，TECH-frontend-state-004）；陈旧由下一次探针 / 请求自然收敛，可接受上限 30s。切租户场景走整页 `location.replace()`（TECH-frontend-state-004），不在此列。

### 会话过期处理（BIZ-auth-008）

已登录用户在正常浏览期间会话过期（Redis session TTL 到期）时：

1. 下一次 API 请求收到 HTTP 401。
2. 前端全局拦截器捕获 401 → 将当前页完整路径（含 query + hash）存入 `redirect_uri` 参数 → `location.replace()` 跳转 eiam 登录页。
3. 用户重新认证成功后回跳至原操作路径，恢复中断操作。
4. 不白屏、不无限 loading——跳转前显示瞬时 loading overlay（150ms 淡出），跳转后由 eiam 登录页接管渲染。

> 401（会话过期→跳登录页）与 403（越权→原地提示，BIZ-auth-003）语义严格分离：401 走跳转流程，403 走 Component 4 原地占位。

### 全局 HTTP 错误分派表（401 / 403 / 503）

全局 HTTP 拦截器（axios/fetch wrapper）按响应码统一分派，三码语义严格分离、不混用，避免散落三处拼装：

| 码 | 语义 | 触发场景 | 拦截器动作 | 目标组件 |
|----|------|---------|-----------|---------|
| 401 | 会话过期 / 未登录（BIZ-auth-005、BIZ-auth-008） | Redis session TTL 到期；跨标签页登出后本标签页下一请求（不做实时广播） | 存当前路径（含 query + hash）为 `redirect_uri` → 150ms loading overlay → `location.replace()` 跳 eiam 登录页 | Component 1（跳转接管） |
| 403 | 越权（已认证但无 `watchalert_role_binding`，BIZ-auth-003） | eiam 认证成功但角色绑定缺失 | 原地渲染 403 占位，不跳登录页，当前数据不变 | Component 4 |
| 503 | eiam 不可用、写操作被拒（BIZ-resilience-001） | 宽限窗口内写请求到达后端被拒（含横幅展示前滞回窗口内的写请求） | toast「鉴权服务暂不可用，操作失败」；不跳转、不清会话 | Component 2（横幅 + toast） |

> 跨标签页登出（BIZ-auth-005）在其余标签页的下一请求表现为 401，归入 401 分派，**不**归入 503——此前 503 note 的 case (b) 已据此移除。401 与 403 语义分离（401 跳登录页、403 原地提示）；503 仅与 eiam 写操作降级相关。

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 宽限剩余时长文案 | `graceRemaining: string` | 后端状态（宽限 2h 倒计时） |
| 写操作许可 | `writeAllowed: boolean` | 后端状态（决定是否禁用写按钮） |
| 探针状态 | `authHealth: AuthHealth` | 后端 SSE/轮询（10s 探针，滞回 3 次） |

> **共享类型契约**：`AuthHealth = 'normal' | 'degraded' | 'recovered'`，全站（Component 2、Component 3 及所有页面级 store）统一使用此枚举类型，来源为同一后端探针。语义：`normal` = eiam 健康；`degraded` = eiam 不可用、宽限窗口内或已过期进入只读；`recovered` = 探针连续 3 次成功判定恢复（瞬时态，前端收到后触发恢复流程，随后回落 `normal`）。禁止任何组件将此字段收窄为 `boolean`。

### Validation Rules（对应 PRD UF-2）

- 横幅必须明确告知「**哪些操作被禁止**」（写操作），而非只提示「服务不可用」。
- 1-2 次探针失败不得展示横幅（滞回防抖），仅连续 3 次失败才判定不可用；连续 3 次成功才判恢复。

---

## Component 3: 只读免登录视图（UF-3）

### Placement

- **Mode**: existing-page
- **Target**: 告警看板（降级态，同一页面在「只读免登录」态下的重渲染）。
- **Position**: 复用告警看板整体布局，但**移除所有写操作按钮与登录入口**；顶部展示只读说明条。

### Layout Structure

```
<nav class="nav-bar">…</nav>                     // 顶部导航（只读态下仍显示三个入口，但页面链接同样只读）
<aside class="readonly-notice">                  // 只读说明条：全宽、bg-muted、h-12、px-6、flex items-center、gap-2、1px solid var(--border)（底边，遵循「全部 1px 细边框」Do）、rounded-0
  └─「鉴权服务不可用，当前为只读浏览。请等待恢复或联系管理员」
<main id="main-content" class="max-w-[1200px] mx-auto px-6 py-8">
  ├─ 告警概览卡片 (grid)                          // grid grid-cols-3 gap-4（≥1024px）/ grid-cols-2（768-1023px）/ grid-cols-1（<768px）
  │    └─ 各卡片：Card（compact 变体：rounded-lg, p-4, 1px border），内容 = 规则名称/级别/状态/触发条件
  ├─ 规则列表 (Table)                             // mt-6；Table 全宽（max-w-[1200px]）；列：规则名称(min-w-[160px]) / 级别(w-[80px]) / 状态(w-[100px]) / 触发条件(min-w-[200px]) / 敏感字段(***, w-[120px])
  │    └─ 表格容器 overflow-x:auto（<768px 时启用水平滚动），行 hover bg-muted/50
  └─ 事件时间线                                   // mt-6；max-w-[800px]；垂直时间线，每条目 = 时间(stamp) + 级别Badge + 状态变更文案；间距 py-3
```

> **尺寸规范**：主内容区 `max-w-[1200px]` + `mx-auto` + `px-6`；概览卡片网格 `gap-4`；规则列表与时间线之间垂直间距 `mt-6`（24px）；时间线区域收窄至 `max-w-[800px]` 以保证可读行宽。表格单元格文字超长时 `truncate`（单行省略号），鼠标 hover 显示 `title` 全文；Badge 级别列固定 `w-[80px]` 防止列宽抖动。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading（数据拉取中） | 概览卡片区域显示 3 个骨架占位（skeleton：`bg-muted rounded-lg h-[120px]` 脉动动画 `animate-pulse`）；规则列表显示 5 行骨架行；时间线区域显示 3 条骨架条目 | 进入只读态后立即渲染骨架，`readonlyAlerts` fetch 返回前不显示空白；`role="status"` + `aria-label="加载中"` |
| Populated（有数据） | 概览卡片 + 规则列表 + 事件时间线正常渲染；敏感字段 `***` 脱敏 | 从 Loading 平滑切换（150ms opacity）；`aria-live="polite"` 通告内容就绪 |
| Empty（无数据） | 概览区域显示空态占位卡（居中、`py-12`、Lucide `inbox` 图标 48px muted + 文案「当前无告警数据」+ 副文案「鉴权服务恢复后可查看实时数据」）；规则列表与时间线同样显示各自空态（`「暂无规则」`/`「暂无事件」`） | 数据 fetch 成功但 `readonlyAlerts.length === 0` 时展示；不与 Loading 混淆 |
| Error（只读接口请求失败） | 概览区域显示错误态占位（Lucide `cloud-off` 图标 48px destructive + 文案「数据加载失败」+ 「重试」Button outline variant）；规则列表与时间线隐藏 | 只读接口返回 5xx 或网络超时（15s）时展示；点击「重试」重新发起 fetch；`role="alert"` + `aria-live="assertive"` |
| 只读免登录（Populated 的稳定态） | 只读看板 + 只读说明条 + 敏感字段 `***` 脱敏；无任何登录/写按钮 | 只读浏览；页面底部 Small 提示「等待鉴权服务恢复」 |
| 已恢复 | 后端探针连 3 次成功 → 前端自动整页跳转 eiam 登录页 | **无用户可点击的登录按钮**，跳转由后端探针信号驱动（防重新引入本地登录入口）；跳转前 150ms 显示瞬时过渡文案「鉴权服务已恢复，正在跳转登录…」`role="status"` |

> **状态转移**：进入只读态 → Loading → (Populated | Empty | Error)。Populated 稳定后进入「只读免登录」展示态。Error 态用户点击重试 → 回到 Loading。收到 `authHealth='recovered'` → 进入已恢复态。慢网络：fetch 超过 5s 未返回时骨架上方追加 Small 文案「加载较慢，请稍候…」（不阻塞、可并行显示最终结果）。

### Edge Cases

| 边界场景 | 处理规则 |
|---------|---------|
| 长文本溢出（规则名称/触发条件） | 单行 `truncate` + `title` 全文；表格列设 `min-w` 防压缩 |
| 规则列表行数 > 50 | 客户端分页（pageSize=20，符合 TECH-frontend-state-002），底部分页器；不分页时表格 `max-h-[600px] overflow-y:auto` |
| 事件时间线条目 > 100 | 同上分页，每页 20 条 |
| 敏感字段原文意外下发 | 前端二次过滤：已知敏感字段名（webhook_secret / api_key / password / datasource_pwd）一律前端覆写为 `***`，不依赖后端 |
| 只读接口 404（无匿名态数据源） | 归入 Error 态，文案改为「只读数据暂不可用」 |
| 用户刷新只读页 | 重新走 Loading → Populated/Empty/Error 流程 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 进入（宽限过期 / 未登录且 eiam 不可用） | 渲染只读看板 | 只读说明条 + 数据脱敏 |
| 前端经 SSE/轮询获知恢复信号（`authHealth='recovered'`） | 先渲染 150ms 瞬时过渡文案「鉴权服务已恢复，正在跳转登录…」（`role="status"`），随后 `location.replace()` 跳转 eiam 登录页 | 无用户可点击的登录按钮入口；过渡文案与 States 表「已恢复」行为一致——以 States 表为准，Interactions 不再表述「无提示停顿」 |
| 用户尝试任何写操作 | 无写操作按钮可点（已移除） | — |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 只读告警数据（看板/规则/事件） | `readonlyAlerts: list` | 后端只读接口（匿名态） |
| 鉴权服务状态（含恢复信号） | `authHealth: AuthHealth` | 后端探针（10s，滞回 3 次）；与 Component 2 共享类型契约。恢复即 `authHealth` 由 `degraded` 转为 `'recovered'`（探针连 3 次成功），前端据此触发已恢复跳转——**不引入独立 `recoverySignal` 字段**，恢复信号是 `authHealth` 枚举值之一，避免双通道冗余与数据契约歧义 |
| 敏感字段脱敏 | `masked: '***'` | 后端只读接口不下发敏感字段原文 |

### 脱敏清单（对应 PRD UF-3）

- **匿名可见字段**：告警规则名称/级别/状态/触发条件、告警事件时间线（时间/级别/状态变更）、事件当前状态。
- **敏感字段脱敏**：通知渠道凭据（webhook secret、API key）、数据源密码、webhook URL 查询参数等一律 `***` 不展示；后端只读接口在匿名态下不下发原文。

### Validation Rules（对应 PRD UF-3）

- 只读视图**不得出现任何登录入口**（避免重新引入被移除的本地登录路径）；恢复跳转由后端探针信号驱动。

---

## Component 4: 403 无权限占位（UF-1 的 in-place 状态）

### Placement

- **Mode**: existing-page
- **Target**: 任意受保护页面（告警看板/规则/故障中心）。
- **Position**: 取代页面主内容区的全宽居中占位（顶部导航仍保留）。

### Layout Structure

```
<main class="forbidden" id="main-content" aria-live="polite">  // 全宽居中，渲染时焦点移至标题
  └─ Card（max-w-[460px], p-6, 居中）
       ├─ Lucide ban (size 24, muted-foreground) + 标题「无访问权限」
       ├─ 说明「请使用平台账号登录后，联系管理员在告警平台分配角色」
       └─ 无跳转/无返回登录按钮（BIZ-auth-003：原地提示，不跳登录页）
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default（403） | 居中 403 占位卡 | 原地停留，不跳登录页，不提供「重新登录」按钮 |
| 角色已分配（恢复） | 占位卡追加 Small 提示「检测到权限已更新，正在刷新…」 + 内联 spinner | 每 60s 轮询 `watchalert_role_binding`（TECH-frontend-state-004 不做多标签页广播，故依赖本标签页轮询，不做 BroadcastChannel）；命中非 null 角色 → `location.replace()` 重载当前页进入正常态。用户手动刷新立即触发一次检测 |

> **状态转移**：Default →（轮询命中角色 | 用户手动刷新命中角色）→ 正常态。无角色期间持续 Default，不自动跳转登录页。轮询间隔 60s 为权限变更传播的可接受上限——管理员分配角色后本标签页最长 60s 内感知（跨标签页因 BIZ-auth-005 不广播而同样受此时延约束）。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| eiam 认证成功但无 `watchalert_role_binding` 匹配 | 渲染 403 占位 | 文案提示联系管理员配置角色 |
| 轮询/手动刷新检测到 `roleBinding` 非 null | `location.replace()` 重载当前页 | 文案提示「检测到权限已更新，正在刷新…」 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 角色匹配结果 | `roleBinding: null \| 'admin' \| 'user'` | 后端鉴权中间件（`watchalert_role_binding` 表查询） |

---

## Page Composition

| Page | Type | 涉及的 UI 组件 | 说明 |
|------|------|----------------|------|
| 登录页 | existing | 组件 1 | 移除本地账密，改 eiam 跳转 |
| 告警看板 | existing | 组件 2（横幅）+ 组件 3（只读）+ 组件 4（403） | 正常/宽限/只读/403 四态 |
| 告警规则 | existing | 组件 2 + 组件 3（只读态脱敏）+ 组件 4 | 共享横幅 + 只读规则列表 |
| 故障中心 | existing | 组件 2 + 组件 3（只读事件时间线）+ 组件 4 | 共享横幅 + 只读事件 |