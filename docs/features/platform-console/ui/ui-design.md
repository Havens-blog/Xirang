---
created: "2026-09-29"
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: 平台控制台（统一入口 + 身份治理 + 权限管理 + 工作台）

## Design System

**Linear 靛紫** — 与 e-cam-web 现网同源的设计系统：颜色令牌取自 `e-cam-web/src/assets/styles/theme-variables.scss`，布局骨架取自 `e-cam-web/src/layouts/MainLayout.vue`（左侧边栏 + 顶部 header）。**深色为默认主题**。用途：企业后台/管理控制台。

> 组件库选型（Element Plus 系，与 e-cam-web 对齐）由 /tech-design 决策。本文规格中的语义 Token 一律组件库无关：若最终组件库不同，仅替换映射右列，语义值不变。

### 色彩（CSS 变量，深色默认 + 浅色双主题）

强调色与语义色两主题一致：

| Token | 值 | 用途 |
|-------|-----|------|
| --accent-primary | #5e6ad2 | 主按钮实底、active 菜单项背景（15% 着色） |
| --accent-blue | #7170ff | hover/active 文字与图标、focus ring、链接 |
| --accent-cyan / --accent-green / --accent-purple | #06b6d4 / #10b981 / #8b5cf6 | 图表与辅助强调（状态徽章复用 accent-green） |
| --color-success | #27a644 | 启用状态、成功反馈 |
| --color-warning | #f59e0b | 锁定状态、警告反馈 |
| --color-danger | #ef4444 | 禁止/删除等危险操作 |
| --color-info | #8a8f98 | 中性信息 |

深色（默认主题）表面 / 文本 / 边框：

| Token | 值 | Token | 值 |
|-------|-----|-------|-----|
| --bg-base = --bg-page | #08090a | --text-primary | #f7f8f8 |
| --bg-elevated | #0f1011 | --text-regular | #d0d6e0 |
| --bg-surface = --bg-overlay | #191a1b | --text-secondary | #8a8f98 |
| --bg-hover | #28282c | --text-tertiary | #62666d |
| --border-subtle / base / strong | rgba(255,255,255,0.05 / 0.08 / 0.12) | --text-muted | #52525b |

浅色（`:root.light`）：

| Token | 值 | Token | 值 |
|-------|-----|-------|-----|
| --bg-base = --bg-page | #f7f8f8 | --text-primary | #111111 |
| --bg-elevated = --bg-surface | #ffffff | --text-regular | #333333 |
| --bg-hover | #f3f4f5 | --text-secondary | #666666 |
| --border-base | #e6e6e6 | --text-tertiary | #999999 |
| --border-strong | #d0d6e0 | --text-muted | #bbbbbb |

派生令牌两主题各备浅深取值：玻璃拟态 --glass-bg / --glass-bg-hover / --glass-border；阴影三档 --shadow-sm / --shadow-base / --shadow-lg（深色自 0 2px 8px rgba(0,0,0,.3) 起，浅色自 0 1px 3px rgba(0,0,0,.08) 起）；另有 scrollbar / input / table / dialog / dropdown / tag 组件级令牌（见 theme-variables.scss）。

主题策略：**深色为默认主题**；主布局 header 提供月/日切换按钮（html 根节点切换 `light` class），选择持久化 localStorage；不跟随系统 `prefers-color-scheme` 自动切换。

### 字体
- 正文：`'Inter Variable', 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif`（正文 16px / 1.5；控件文本 14px；辅助 12–13px）
- 标题：Inter，H1 36px/700，H2 30px/600，H3 24px/600，H4 20px/600
- 等宽（ID/代码）：JetBrains Mono 14px

### 组件规范（统一 8px 圆角）
- **按钮**：圆角 8px、h-40；primary = accent-primary 实底白字，secondary = bg-surface + 1px border，另有 outline / ghost / destructive（danger 90% 加深实底白字）四变体；hover 为背景/边框色变化，150ms
- **卡片**：bg-surface + 1px border-subtle、圆角 8px、p-6、shadow-sm；CardHeader/CardContent/CardFooter 分层
- **输入**：input-bg 半透明底 + 1px input-border、圆角 8px、h-40；focus 边框 accent-blue + 35% 光圈（ring 2px）；disabled opacity-50
- **徽章**：圆角 8px、px-2.5、text-xs；软着色（语义色 16% 透明底 + 语义色文字）；状态列着色一律按「状态徽章映射」，不硬编码色值
- **表格**：容器 bg-surface + border-subtle、圆角 8px + shadow-sm；header 用 table-header-bg、行 hover 用 table-row-hover、行间 table-border；长文本列（DN/邮箱/授权主体/URL 等单行串）一律单行截断省略号 + hover title 显示全文，不换行撑高行
- **对话框**：遮罩 rgba(0,0,0,.7) + blur、内容 dialog-bg 圆角 8px + shadow-lg、fade-in + scale(0.95→1)

### 状态徽章映射（跨页统一）
「状态」列三态全站同一视觉；success/warning 语义色仅用于状态与结果反馈、不用于装饰，danger 保留给删除/危险操作。徽章 = 语义色 16% 透明底 + 语义色文字（浅色主题文字经 color-mix 加深以保 AA），不硬编码色值。

| 状态 | 底色 token | 文字 token | 实测对比度（Dark / Light） |
|------|-----------|-----------|---------------------------|
| 启用 | accent-green 16% 底 | accent-green（浅色 color-mix 72% 黑） | 5.2:1 / 4.7:1 |
| 禁用 | tag-bg | text-regular | 9.0:1 / 11.0:1 |
| 锁定 | color-warning 16% 底 | color-warning（浅色 color-mix 60% 黑） | 6.1:1 / 4.9:1 |

UF-6 两态复用「启用/禁用」两行；UF-7 启用维持开关形态不套徽章。

### 语义 Token 映射（组件库无关）
| Token | 值 | 说明 |
|-------|-----|------|
| control-height | 40px | 输入/按钮/菜单项 |
| header-height | 56px | 顶部 header 与侧边栏 logo 区 |
| sidebar-width / sidebar-collapsed | 220px / 64px | 左侧边栏（折叠态仅图标） |
| card-padding / section-gap | 24px | spacing-base 4px 的倍数（gap-4=16px） |
| radius | 8px | 按钮/卡片/输入/徽章/表格/对话框统一圆角 |
| ring-width / ring-offset | 2px / 2px | focus ring（accent-blue） |

### 布局骨架与响应式（侧边栏 + 顶部 header，与 e-cam-web MainLayout 同构）
- 左侧边栏：宽 220px（可折叠 64px）；顶部 logo 区 56px（靛紫 logo 图标 + 产品名「Platform」）；下方**分组菜单**，组标题 uppercase 11px 灰色；菜单项 40px（图标 18px + 标题 14px）。active 态 = background color-mix(accent-primary 15%, transparent) + 文字/图标 accent-blue；hover 态 = bg-hover；折叠态仅图标（logo 文本与组标题隐藏，tooltip 提示）。
- 顶部 header：高 56px。左 = 折叠按钮 + 面包屑（首页 / 当前页）；右 = 租户选择器 + 应用切换器 + 主题切换（月/日）+ 用户头像（accent-blue→accent-purple 渐变 32px）+ 用户名。
- 主内容区：padding 24px，背景 bg-base，容器最大 1440px 居中。
- 最小支持分辨率 1280×720；v1 仅桌面后台、不做移动端适配。内容区宽 < 1280 时不折叠布局，表格容器横向滚动（表格最小可用宽度 960px），不产生页面级横滚；< 1024px 侧边栏转为抽屉（点折叠按钮开合）。
- UF-5 双栏定比：组织树 : 右侧（详情 + 成员）= 2:3，树最小宽 280px；其余管理页为单列表格。
- 工作台应用卡片列数定死：视口 1280–1599px 为 2 列、≥1600px 为 3 列（仅此两个决策断点，v1 不另引入断点）。

### 深度与间距
- 阴影三档系统（--shadow-sm / base / lg）：配合背景表面阶（base → elevated → surface → hover）+ 边框三档（subtle/base/strong）表达层级；对话框/下拉用 shadow-base/lg
- 玻璃拟态令牌（--glass-bg / --glass-bg-hover / --glass-border）供悬浮层背景使用
- 间距 4px 基准：gap-4(16px)、p-6(24px)、space-y-6(24px)

### Do / Don't
- Do：靛紫强调 + Linear 灰阶、统一 8px 圆角、阴影三档、深色默认双主题（html.light 切换）、font-medium 强调、侧边栏分组菜单
- Don't：引入靛紫以外的品牌主色、粗/双边框、乱用圆角、硬编码色值、自创间距、回到顶部导航条布局

### Accessibility
- **键盘导航**：全部交互可 Tab 到达并显示 focus ring（accent-blue）。组织树 ↑↓ 移动、←→ 折叠/展开、Enter 选中；表格行操作按 Tab 顺序可达；Dialog 打开即 focus trap（Tab 循环于弹窗内）、Esc 关闭、关闭后焦点归还触发元素；登录页 Tab 顺序 = 登录方式 Tab → 表单字段 → 提交按钮。
- **表单标签**：每个输入配显式 label（不依赖 placeholder）；必填以「*」标注；字段校验错误经 aria-describedby 关联到错误文案。
- **状态播报**：全局 toast 与表单错误条用 role="alert"；skeleton 区域标 aria-busy；锁定倒计时、「撤销生效（≤30s）」等时变信息用 aria-live="polite" 播报——loading/error/锁定状态均不以纯视觉传达。
- **图标文本替代**：侧边栏 8 项一律「图标 + 文字」同渲染，折叠态隐藏文字时以 tooltip + aria-label 兜底；icon-only 按钮（如主题切换、表格行内操作）必须带 aria-label。
- **对比度**：文本一律取色彩 token 组合，明暗两主题均满足 WCAG AA（正文/控件文本 ≥4.5:1），关键组合逐组实测如下；强调色作文字使用时（active 菜单项/链接/图标）满足图形 ≥3:1，正文不使用强调色文字。

| 组合（文字 / 底） | Dark | Light |
|------------------|------|-------|
| text-primary / bg-base | 18.5:1 | 17.7:1 |
| text-regular / bg-base | 13.6:1 | 11.8:1 |
| text-secondary / bg-base | 6.1:1 | 5.4:1 |
| 白字 / accent-primary（主按钮实底） | 4.7:1 | 4.7:1 |
| 白字 / danger 90% 加深实底（危险按钮） | 4.6:1 | 4.6:1 |
| 状态徽章文字 / 自身 16% 着色底 | 5.2–9.0:1 | 4.7–11.0:1 |

### 通用状态迁移（各列表页共用）
Loading（skeleton）→ 接口返回非空数组 → Populated；返回空数组 → Empty；请求失败 → Error（展示重试按钮，重试后回到 Loading）。各列表页组件共用上述迁移，其 States 表不再重复描述迁移条件。

### 边界行为约定（E6/E7，各列表页共用）
- **并发编辑（E6）**：提交保存时后端返回 409（数据已被他人修改）→ 保留表单当前输入，顶部 toast（逐字见契约 save.conflict）并提供「重新加载」按钮；确认后以服务端最新数据覆盖表单，本地未提交修改不自动保留。
- **分页边界（E7）**：page 超出总页数时自动回落至最后一页；删除使当前页变空时回退至前一页重查（整体为空则转 Empty 态）；切换过滤/搜索条件时 page 重置为 1。UF-4/UF-6 分页表格共用。

---

## Component: 登录页（UF-1）

### Placement
- **Mode**: new-page
- **Target**: /login
- **Position**: 独立整页，居中卡片

### Layout Structure
外层居中（flex min-h-screen items-center justify-center，bg-base + 頂部靛紫径向微光）→ 卡片（w-full max-w-sm，bg-elevated + shadow-base）→ CardHeader（产品名 + Logo 位）→ CardContent（登录方式 Tab + 表单）→ CardFooter（登录按钮/帮助）。下方无其他区块。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Default | 登录方式 Tab + 账号密码表单 | 可提交 |
| MFA challenge | 追加 6 位验证码输入 | 阻断进入工作台 |
| 强制改密 | 改密表单（新密码 + 确认） | 初始密码首次登录触发（E1），改密成功后方可进入租户选择 |
| Loading | 按钮转 loading、禁用 | 提交中 |
| Error | 表单顶部统一中性错误条：凭据错误逐字见契约 auth.invalid_credentials，MFA 码错误见 auth.mfa_invalid（均不提示账号是否存在） | 保留输入可重试 |
| Locked | 中性锁定提示，逐字见文案契约 auth.account_locked | 倒计时后恢复 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 提交登录 | 调 eiam 登录接口 → 写共享 session | 成功跳 /tenant-select（0/1 租户自动分流、≥2 租户列表选择，见 UF-11）；失败错误条 |
| 选登录方式 | 切 Tab（密码/LDAP/passkey） | 表单随之切换；WebAuthn 不可用隐藏 passkey 入口 |
| MFA 校验 | 连续 5 次失败 | 锁定 15 分钟，中性文案提示 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 用户名 | username | 用户输入 |
| 密码 | password | 用户输入 |
| MFA 码 | mfa_token | 用户输入 |
| 剩余锁定时间 | lock_remain | eiam 响应 |
| 登录方式 | login_method | 本地枚举 |
| 帮助入口 | help_url | 本地静态配置（v1 指向运维支持页） |

---

## Component: 主布局（侧边栏 + 顶部 header）（UF-2）

### Placement
- **Mode**: new-page
- **Target**: 主布局（所有登录后页共享）
- **Position**: 左侧固定侧边栏 + 内容区顶部通栏 header（.layout 100vh 两栏，侧边栏与内容区各自滚动）

### Layout Structure
.layout 横向两栏：左侧边栏 + 右侧 .main-area（header + 内容 outlet）。
- **侧边栏（220px / 折叠 64px）**：logo 区 56px（靛紫 logo 图标 + 产品名「Platform」，点击回工作台）→ 分组菜单（组标题 uppercase 11px 灰色）：「概览」工作台；「身份治理」用户管理 / 组织管理 / 租户管理 / 身份源管理；「权限管理」角色管理 / 授权管理 / 策略管理。菜单项 40px（图标 18px + 标题 14px），按角色裁剪：普通用户仅工作台；只读/审计可见 2/3/5/6/7 仅查看、不渲染增删改入口；租户管理员见用户/授权/角色；租户与策略管理仅平台管理员。active = color-mix(accent-primary 15%) 底 + accent-blue 文字/图标；hover = bg-hover。
- **顶部 header（56px）**：左 = 侧边栏折叠按钮 + 面包屑（首页 / 当前页，当前页取自侧边栏 active 项文案）；右 = 租户选择器 + 应用切换器 + 主题切换（月/日）+ 用户头像下拉（登出）+ 用户名。
- **内容区**：padding 24px、bg-base，路由 outlet。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Default | 分组菜单 + header 全量 | 可操作 |
| 侧边栏折叠 | 宽 220px→64px，仅图标；<1024px 转抽屉 | 再次点击恢复；悬浮显示 tooltip |
| 403 | 原地 toast（逐字见文案契约 forbidden） | 不跳转、保持当前页 |
| 受限（零租户） | 租户选择器禁用占位并显示契约 tenant.none 文案；应用切换器隐藏（无可用应用）；菜单仅「工作台」 | 不报错不白屏；平台管理员豁免口径见 UF-3 受限视图 |
| 过期（401） | 跳登录页 | 全局 401 拦截（含 Dialog 打开态）：中断请求、丢弃未提交草稿、跳 /login；v1 不保留编辑上下文 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 点侧边栏菜单 | 路由跳转对应页 | active 态高亮，面包屑同步 |
| 点折叠按钮 | 侧边栏 220px ↔ 64px（<1024px 为抽屉开合） | 300ms 过渡 |
| 点应用切换器 | 新标签打开 e-cam-web（共享会话） | 控制台标签页保留，工作台上下文不丢失 |
| 点主题切换 | html 根节点切 light class（月/日图标随主题互换） | 全站即时换肤，选择持久化 localStorage |
| 切换租户 | header 租户选择器选择目标 → 写 session tenant_id | 丢弃未提交草稿并关闭 Dialog（与 401 策略一致，v1 不保留编辑上下文）；SPA 内按新租户重取数据并回工作台（数据重取，非整页 reload） |
| 点登出 | 销毁共享会话 | 双侧下次请求 401 → 回登录页 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 当前用户 | user {name, avatar} | eiam session |
| 当前租户 | tenant_id | eiam session |
| 租户选择器选项 | tenants[]（当前用户所属） | eiam（与 UF-11 同源） |
| 可见菜单 | menus[] | eiam 权限（按角色） |
| 应用清单 | apps[] | 本地枚举（v1 仅「云管 e-cam-web」一项） |
| 主题偏好 | theme（dark / light，深色默认） | 本地 localStorage |

---

## Component: 工作台（UF-3）

### Placement
- **Mode**: new-page
- **Target**: /workbench
- **Position**: 登录后首页

### Layout Structure
内容区 space-y-6：顶部欢迎横幅（H2 欢迎 + 我的信息：用户名、昵称、所属租户、角色四字段只读展示）→ 应用快捷导航卡片组（列数定死见「布局骨架与响应式」：1280–1599px 2 列、≥1600px 3 列）。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Default | 欢迎 + 我的信息 + 应用卡片 | 可跳转 |
| 受限视图（零租户） | 仅欢迎 + 我的信息；应用卡片禁用；不渲染管理菜单 | 所属租户=0（E3，平台管理员豁免）；URL 直入管理页提示「无所属租户」，不报错不白屏 |
| eiam 不可用 | 静态内容仍渲染 + 顶部降级提示 | 卡片跳转禁用 |

迁移：进入 /workbench 时按 eiam 所属租户查询结果二选一 Default / 受限视图（0 个 → 受限视图）。

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 点应用卡片 | 跳转对应应用/页面 | 高亮 hover |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 用户名/昵称 | user.name | eiam session |
| 我的信息 | profile {username, display_name, tenant_name, roles[]} | eiam session（与 UF-2 当前用户同源） |
| 应用卡片 | apps[] | 本地枚举（v1 仅云管） |

---

## Component: 用户管理（UF-4）

### Placement
- **Mode**: new-page
- **Target**: /users

### Layout Structure
页面头（标题 + 「新增用户」按钮）→ 搜索栏（关键词输入 + 角色过滤）→ 表格（列：用户名/所属租户/组织/登录方式/角色/状态/操作）→ 分页。新增用 Dialog；行内「编辑」跳用户详情/编辑子页（Secondary Page）。

#### Secondary Page: 用户详情/编辑（/users/:id）
- **Entry**：列表「编辑」；**Return**：页头「← 返回用户列表」→ /users（满足 PRD 二级页返回导航规则）
- **Layout**：页头（返回 + 用户名 + 保存）→ 表单（用户名只读、登录方式、角色绑定多选、所属租户/组织关联选择、状态）→ passkey 注册引导区块
- **Binding**：复用本组件 Data Binding；单条用户由 eiam 按 id 查询

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 表格 skeleton | 加载中 |
| Empty | 空状态插画 + 文案 | 无数据 |
| Populated | 表格数据 | 可操作 |
| Error | 行内错误 + 重试 | 接口失败可重试 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 新增 | Dialog（用户名/初始密码/登录方式/角色绑定/所属租户组织） | 保存后列表 ≤2s 刷新 |
| 编辑 | 跳 /users/:id 详情/编辑页 | 保存后返回 /users，列表 ≤2s 刷新 |
| 保存遇并发冲突 | 提交返回 409 | 见「边界行为约定（E6）」：toast save.conflict + 重新加载 |
| 删除 | 删除确认 | 删除被拒显示原因 |
| 搜索/过滤 | 关键词/角色过滤 | 列表即时过滤 |
| 引导 passkey 注册 | 对目标用户发起 → eiam 生成一次性注册会话（15 分钟、单次有效） | 表单内提示有效期与状态；过期未完成即作废，须重新引导 |
| passkey 完成态 | 用户在其设备完成 WebAuthn 注册 | 该用户登录方式标记「已绑定 passkey」 |
| 撤销重绑 | 设备丢失：撤销原 passkey 绑定 → 重新发起同一引导 | 重绑前用户凭密码/LDAP 登录不受影响；passkey 为唯一登录方式时走管理员重置密码路径（E1） |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 用户名 | username（3–32 字符，正则 ^[a-z][a-z0-9_-]{2,31}$） | 表单 |
| 初始密码 | initial_password（管理员设置或系统生成，展示一次；首次登录强制改密，承接 E1） | 表单/生成 |
| 登录方式 | login_method（密码/passkey 注册引导） | 表单 |
| passkey 注册会话 | passkey_session {id, expires_at=+15min} | eiam 生成 |
| 角色绑定 | roles[] | eiam 多选 |
| 所属租户/组织 | tenant_id / org_id | eiam 关联选择 |
| 用户列表（表格） | users[]（username/tenant_id/org_id/login_method/roles/status） | eiam 分页查询 |
| 分页器 | page / page_size / total | eiam |
| 状态 | status（启用/禁用/锁定） | eiam |

---

## Component: 组织管理（UF-5）

### Placement
- **Mode**: new-page
- **Target**: /organizations

### Layout Structure
左：组织树（Card）；右：选中组织详情（名称/下属组织数/关联用户数）+ 成员列表（Card）。树深度 ≤8、节点 ≤1000；树节点名超宽单行截断（省略号 + title 全文）。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 树 skeleton | 加载中 |
| Empty | 空状态 | 无组织 |
| Populated | 树 + 详情 | 可操作 |
| Error | 错误 + 重试 | 加载失败 |
| 删除被拒 | Dialog 列出阻断计数（下属/关联用户），确认置灰 | 不可确认 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 新增子组织/编辑/删除 | 弹窗操作 | 树刷新 |
| 树超限禁建 | 深度已达 8 或节点数已达 1000 时「新增子组织」按钮禁用 | tooltip 提示已达上限，不可再建子级（PRD 上限保护） |
| 选中节点 | 加载详情 | 右侧更新 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 组织树 | org_tree（树/层级） | eiam |
| 组织详情（右侧） | org {org_name, child_count, user_count} | eiam 按选中节点查询 |
| 成员列表（右侧） | members[]（username, roles[]） | eiam |
| 组织名 | org_name（1–64 字符，超长前端拦截） | 表单 |

---

## Component: 租户管理（UF-6）

### Placement
- **Mode**: new-page
- **Target**: /tenants

### Layout Structure
页面头 + 表格（列：租户名/状态/活跃会话数/操作）→ 分页。仅平台管理员可见。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | skeleton | 加载中 |
| Empty | 空状态 | 无数据 |
| Populated | 表格 | 可操作 |
| Error | 错误 + 重试 | 加载失败 |
| 禁用被拒 | 提示文案逐字见契约 tenant.disable_blocked（{n} = active_session_count，与表格列同源） | 阻断 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 新增/编辑 | 弹窗 | 刷新 |
| 禁用 | 校验未过期会话数=0 → 确认 | 被拒显示活跃会话数 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 租户列表 | tenants[] | eiam |
| 分页器 | page / page_size / total | eiam（与 UF-4 同源口径，边界行为见 E7） |
| 租户名 | tenant_name（1–64 字符，全局唯一，冲突提示契约 validation.name_exists） | 表单 |
| 状态 | status（启用/禁用） | eiam |
| 活跃会话数列 | active_session_count（未过期共享 session 计数） | eiam（与禁用校验同一字段） |

---

## Component: 身份源管理（UF-7）

### Placement
- **Mode**: new-page
- **Target**: /identity-sources

### Layout Structure
表格（列：名称/类型/启用/操作）→ 新增/编辑 Dialog（连接参数表单）。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | skeleton | 加载中 |
| Empty | 空状态 | 无配置 |
| Populated | 表格 | 可操作 |
| Error | 错误 + 重试 | 加载失败 |
| 连接失败 | 测试连接报错 | 保存前可测试 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 新增/编辑 | 6 字段表单：连接地址（ldap(s)://host:port，端口 1–65535）、Bind DN、Bind 密码（仅写入：掩码显示、不回显已存明文，留空 = 不修改）、Base DN、用户属性映射（用户名属性必填）、超时 1–30s | 保存 |
| 测试连接 | 调 eiam 测试 | 成功/失败提示 |
| 启停 | 开关切换 | 状态更新 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 名称/类型列 | name / type（ldap） | eiam |
| 连接参数 | ldap_url/port/bind_dn/base_dn/attr_map/timeout | 表单 |
| Bind 密码 | password（仅写入，掩码展示、不回显明文；留空表示沿用原值） | 表单（type=password） |
| 启用状态 | enabled | eiam |

---

## Component: 角色管理（UF-8）

### Placement
- **Mode**: new-page
- **Target**: /roles

### Layout Structure
页面头 + 表格（列：角色名/所属租户/绑定策略数/绑定用户数/操作）→ 行内「编辑」跳角色详情/绑定子页（Secondary Page）。

#### Secondary Page: 角色详情/绑定（/roles/:id）
- **Entry**：列表「编辑」；**Return**：页头「← 返回角色列表」→ /roles
- **Layout**：页头（返回 + 角色名 + 保存）→ 角色名/所属租户 → 策略绑定多选（租户管理员仅列 eiam 返回的可绑定策略，无策略定义入口）+ 用户绑定多选（支持搜索）——双多选在整页承载，避免 Dialog 内可用性问题
- **Binding**：复用本组件 Data Binding；单条角色由 eiam 按 id 查询

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | skeleton | 加载中 |
| Empty | 空状态 | 无角色 |
| Populated | 表格 | 可操作 |
| Error | 错误 + 重试 | 加载失败 |
| 删除被拒 | 提示绑定用户数 | 阻断 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 新增 | Dialog（角色名/所属租户） | 保存后列表 ≤2s 刷新 |
| 编辑 | 跳 /roles/:id 详情/绑定页 | 保存后返回 /roles，列表 ≤2s 刷新 |
| 删除 | 校验无用户绑定 | 被拒显示绑定数 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 角色列表（表格） | roles[]（role_name/tenant_id/policy_count/user_count） | eiam |
| 绑定策略数/绑定用户数列 | policy_count / user_count | eiam（user_count 亦为删除校验来源） |
| 角色名 | role_name（1–64 字符，租户内唯一，冲突提示契约 validation.name_exists） | 表单 |
| 所属租户 | tenant_id | eiam |
| 策略绑定 | policies[] | eiam 多选 |
| 用户绑定 | users[] | eiam 多选 |

---

## Component: 授权管理（UF-9）

### Placement
- **Mode**: new-page
- **Target**: /authorizations

### Layout Structure
页面头 + 表格（列：主体/资源/动作/操作）→ 新增/撤销 Dialog。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | skeleton | 加载中 |
| Empty | 空状态 | 无授权 |
| Populated | 表格 | 可操作 |
| Error | 错误 + 重试 | 加载失败 |
| 撤销 pending | 行内徽标「生效中」、操作按钮禁用 | 撤销提交成功至生效（≤30s）期间；生效后刷新列表移除该行，aria-live 播报 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 新增授权 | Dialog（主体 + 资源 + 动作，均下拉枚举） | 前端预检唯一性 → 提交 |
| 撤销 | 确认弹窗 | 行进入 pending 态（见 States），生效延迟 ≤30s，aria-live 播报 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 主体 | subject | eiam 词表（下拉，非自由文本） |
| 资源 | resource | eiam 词表（下拉） |
| 动作 | action | eiam 词表（下拉） |

---

## Component: 策略管理（UF-10）

### Placement
- **Mode**: new-page
- **Target**: /policies

### Layout Structure
页面头 + 表格（列：策略名/Effect/绑定对象数/操作）→ 编辑 Dialog（策略名 + JSON Statement 表单）。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | skeleton | 加载中 |
| Empty | 空状态 | 无策略 |
| Populated | 表格 | 可操作 |
| Error | 错误 + 重试 | 加载失败 |
| 校验失败 | 错误字段标红 + 阻止保存 | 前端 JSON 语法失败 |

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 新增/编辑 | JSON Statement（Effect 枚举/Action/Resource 非空数组/Condition） | 前端语法 → eiam schema 校验，失败标红 |
| 删除 | 校验无绑定 | 被拒提示 |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 绑定对象数列 | bound_role_count | eiam（删除被拒提示同源：绑定角色数） |
| 策略名 | policy_name（1–64 字符，唯一，冲突提示契约 validation.name_exists） | 表单 |
| Effect | effect（Allow/Deny 枚举） | 表单 |
| Action/Resource | actions[]/resources[] 非空数组 | 表单 |
| Condition | condition | 表单 |

---

## Component: 租户选择（UF-11）

### Placement
- **Mode**: new-page
- **Target**: /tenant-select

### Layout Structure
居中卡片：标题「选择租户」+ 租户列表（单选 radio）或自动提示。「零租户」直达受限工作台（仅欢迎 + 我的信息，禁用管理入口）。

### States
| State | Visual | Behavior |
|-------|--------|----------|
| Loading | skeleton | 查询所属租户中（登录后必经异步请求） |
| 多租户 | 租户单选列表 | 选择后进入工作台 |
| 单租户 | 自动选定，toast 短暂提示后直接进工作台 | 无停留交互 |
| 零租户 | 受限工作台（即 UF-3 受限视图态）+ 提示「无所属租户」 | 平台管理员豁免可进管理页 |
| Error | 错误提示 + 重试按钮 | 查询失败；重试回到 Loading |

迁移：Loading → eiam 返回 0 个直达受限工作台、1 个自动选定进工作台、≥2 个渲染多租户列表。

### Interactions
| Trigger | Action | Feedback |
|---------|--------|----------|
| 选租户 | 写入共享 session tenant_id | 跳工作台 |
| 登录后切换租户 | 经 UF-2 顶部 header 租户选择器（触发器与草稿处置定义见 UF-2「切换租户」行） | 丢弃未提交草稿、关闭 Dialog（与 401 策略一致）；数据重取非整页 reload |

### Data Binding
| UI Element | Data Field | Source |
|------------|-----------|--------|
| 租户列表 | tenants[] | eiam 查询所属 |
| 选中租户 | tenant_id | 写入 session（口径待 Phase 0 核查） |

---

## 文案契约（错误码 → 逐字文案）

系统级错误、确认与关键交互文案唯一口径，实现不得改写措辞；字段级校验文案（必填/格式/长度提示）不在本契约内，由各表单 aria-describedby 错误文案就近维护（唯一性冲突除外，见 validation.name_exists）。安全级文案不提示账号是否存在、不回显账号信息。v1 仅中文，按 key 维护以备 i18n。

| Key / 错误码 | 逐字文案 | 场景 |
|--------------|----------|------|
| auth.invalid_credentials | 用户名或密码错误 | 用户不存在与密码错误返回同一文案（防账号枚举） |
| auth.mfa_invalid | 验证码错误，请重新输入 | MFA 码错误（UF-1 MFA challenge 态，与凭据错误文案区分；同样不提示账号是否存在） |
| auth.account_locked | 尝试次数过多，账号已锁定，剩余 {X} 分钟 | E2/E9；X 逐分钟刷新，副行「可改用其他已启用的登录方式」 |
| auth.empty_credentials | 请输入用户名和密码 | 凭据为空不可提交（前端拦截） |
| auth.passkey_unsupported | 当前浏览器不支持 Passkey，请改用密码或 LDAP 登录 | E8 |
| auth.first_login_change_password | 首次登录请修改密码 | E1 强制改密页标题 |
| session.expired | 登录已过期，请重新登录 | 401 全局拦截提示（UF-2） |
| forbidden | 无权限执行该操作 | 403 toast（UF-2） |
| tenant.none | 无所属租户 | E3 受限视图 / URL 直入管理页 |
| tenant.disable_blocked | 无法禁用：该租户仍有 {n} 个活跃会话 | UF-6 禁用被拒（{n} = active_session_count） |
| eiam.unavailable | 身份服务暂不可用，请稍后重试 | E12 各管理页 Error、工作台降级提示 |
| grant.duplicate | 授权已存在 | UF-9 前端预检与 eiam 兜底 |
| delete.blocked_org | 无法删除：存在 {n} 个下属组织、{m} 个关联用户 | E11（UF-5 弹窗，确认置灰） |
| delete.blocked_role | 无法删除：该角色已绑定 {n} 个用户 | UF-8 |
| delete.blocked_policy | 无法删除：该策略已被 {n} 个角色绑定 | UF-10 |
| validation.name_exists | 名称已存在：{name} | 租户/角色/策略名称唯一性冲突（1–64 字符约束，保存时阻断） |
| save.conflict | 数据已被他人修改，请重新加载后再试 | E6 并发编辑 409（各管理页保存提交） |
