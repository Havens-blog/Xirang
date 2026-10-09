---
feature: "平台控制台（统一入口 + 身份治理 + 权限管理 + 工作台）"
---

# 平台控制台 — UI Functions

> 需求层：定义 UI「做什么」，不定义「长什么样」（那是 ui-design.md）。

## UI Scope

全新纯前端控制台，全部为新页面：登录页、主布局（左侧边栏 + 顶部 header + 应用切换器）、工作台、租户选择、身份治理四页（用户/组织/租户/身份源）、权限管理三页（角色/授权/策略）。

## Navigation Architecture

- **Platform**: web

### Primary Navigation (shared across pages — 左侧边栏分组菜单)

| 分组 | 菜单项 | Target Page | Icon Keyword |
|------|--------|-------------|-------------|
| 概览 | 工作台 | /workbench | home |
| 身份治理 | 用户管理 | /users | user |
| 身份治理 | 组织管理 | /organizations | org |
| 身份治理 | 租户管理 | /tenants | tenant |
| 身份治理 | 身份源管理 | /identity-sources | ldap |
| 权限管理 | 角色管理 | /roles | role |
| 权限管理 | 授权管理 | /authorizations | grant |
| 权限管理 | 策略管理 | /policies | policy |

> 侧边栏 220px（可折叠 64px）；顶部 header 承载 面包屑 / 租户选择器 / 应用切换器（新标签跳 e-cam-web）/ 主题切换 / 用户头像。可见范围按角色裁剪（唯一口径，与各 UF Validation Rules 一致）：普通用户仅见工作台；只读/审计可见 工作台/用户/组织/身份源/角色/授权（只读，不渲染增删改入口）；租户管理员见 工作台/用户/角色/授权（租户内，角色页可绑定已有策略）；租户管理 与 策略管理 仅平台管理员可见（只读/审计亦不渲染）。

### Secondary Pages (navigated from a parent page)

| Page | Entry Point (UF# or action) | Return Target |
|------|-----------------------------|---------------|
| 登录页 /login | 未登录访问任意页 | /workbench |
| 用户详情/编辑 | UF-4 列表「编辑」 | /users |
| 角色详情/绑定 | UF-8 列表「编辑」 | /roles |

### Navigation Rules

- Primary navigation is shared across pages
- Every secondary page must have back navigation targeting its entry point page
- Every navigation target must correspond to a page defined in this document

## UI Function 1: 登录页

### Placement
- **Mode**: new-page
- **Target Page**: /login（无会话访问入口）
- **Position**: 独立整页

### Description
按 eiam 支持的登录方式完成认证：密码（system）、LDAP、passkey，必要时 MFA 二次验证（连续 5 次失败锁定 15 分钟）；认证成功后按 UF-11 分流进入租户选择/工作台。

### User Interaction Flow
用户访问 → 无会话 → 跳 /login → 选登录方式 → 提交凭据 → 成功则写 session 并进入租户选择 → 失败提示错误可重试。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 登录方式 | enum | 本地 | system/ldap/passkey |
| username/password | string | 用户输入 | 密码登录 |
| MFA 令牌 | string | 用户输入 | 需要时 |
| 租户 | 选择 | eiam | 认证后按 UF-11 处理 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 未登录 | 登录表单 | 无有效会话 |
| 认证中 | loading | 提交凭据 |
| 失败 | 错误提示 | 凭据错误/MFA 失败 |
| MFA 锁定 | 锁定提示+剩余解锁时间 | MFA 连续失败 5 次 |
| passkey 不可用 | 隐藏 passkey 入口，提示改用密码/LDAP | 浏览器不支持 WebAuthn |
| 成功 | 跳转 UF-11 租户选择/工作台 | 认证通过 |

### Validation Rules
- 凭据为空不可提交；错误提示不含敏感信息
- 「不泄露用户名是否存在」仅限凭据错误提示：用户不存在与密码错误返回同一文案（如「用户名或密码错误」）
- 账号锁定提示（E2/E9）为中性文案：仅含「尝试次数过多，账号已锁定，剩余 X 分钟」与可改用的其他登录方式，不提示账号是否存在、不回显账号信息

## UI Function 2: 顶部导航 + 应用切换器

### Placement
- **Mode**: new-page
- **Target Page**: 主布局（所有登录后页共享）
- **Position**: 顶部通栏

### Description
全局导航展示可见菜单 + 当前租户（切换见 UF-11）+ 应用切换入口（跳 e-cam-web）+ 登出。

### User Interaction Flow
点击菜单 → 路由跳转；点应用切换器 → 单点跳 e-cam-web；点登出 → 双侧会话失效。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 当前用户 | object | eiam session | 我的信息 |
| 当前租户 | tenant_id | eiam session | |
| 可见菜单 | 按角色 | eiam 权限 | 角色裁剪 |
| 应用清单 | 枚举 | 本地 | 穷举：v1 仅「云管（e-cam-web）」一项，无其他应用；新增应用须改代码发布，不做运行时注册 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 已登录 | 菜单+头像+应用切换 | 有效会话 |
| 过期 | 跳登录页 | 401 |
| 无权限 | 原地错误提示，不跳转登录页 | 接口返回 403 |

### Validation Rules
- 越权菜单不渲染；无权限操作明确提示

## UI Function 3: 工作台

### Placement
- **Mode**: new-page
- **Target Page**: /workbench
- **Position**: 登录后首页

### Description
欢迎语 + 我的信息（来自 eiam）+ 应用快捷导航卡片。

### User Interaction Flow
登录后默认进入 → 展示欢迎/我的信息/应用卡片 → 点击卡片跳转对应应用或页面。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 用户名/昵称 | string | eiam session | 我的信息 |
| 应用卡片 | 枚举 | 本地 | 与 UF-2 应用清单同源：仅云管（e-cam-web） |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载完成 | 欢迎+卡片 | 进入 |
| eiam 不可用 | 降级提示 | 接口失败 |

### Validation Rules
- eiam 不可用时静态内容仍渲染、明确提示、不白屏

## UI Function 4: 用户管理

### Placement
- **Mode**: new-page
- **Target Page**: /users
- **Position**: 独立列表页

### Description
平台用户的列表/搜索/新增/编辑/删除（直调 eiam）。

### User Interaction Flow
列表加载 → 搜索过滤 → 新增/编辑弹窗 → 提交 CRUD → 列表刷新 → 删除需确认。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 用户列表 | array | eiam | 分页/搜索 |
| 用户名 | string | 表单 | 必填；3–32 字符，`^[a-z][a-z0-9_-]{2,31}$`，唯一性由 eiam 校验 |
| 初始密码 | string | 表单/系统生成 | 新增密码账号时二选一：管理员设置或系统生成（展示一次）；首次登录强制改密 |
| 登录方式 | enum | eiam | 密码/passkey（passkey 经注册引导绑定，WebAuthn 可用时） |
| 状态 | enum | eiam | 启用/禁用/锁定 |
| 所属租户/组织 | ref | eiam | 关联选择 |
| 角色绑定 | ref[] | eiam | 多选 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | skeleton | 进入 |
| 空 | 空状态 | 无数据 |
| 列表 | 表格 | 有数据 |
| 错误 | 错误提示 | 接口失败 |

### Validation Rules
- 用户名格式/长度前端即时拦截，唯一性反馈来自 eiam；租户边界内过滤
- 密码账���必须发放初始密码（设置或生成），passkey 账号完成注册引导才可提交
- 分页与并发边界见 prd-spec.md E6/E7

### passkey 注册引导（定义）

- **主体与执行方式**：平台管理员在用户新增/编辑页对目标用户发起「引导 passkey 注册」→ eiam 生成一次性注册会话（有效期 15 分钟、单次使用；该机制是否由 eiam 现成提供列入 Phase 0 parity 核查，若无进入缺口处置）→ 用户本人在其常用设备浏览器于有效期内完成 WebAuthn 注册。私钥仅存于用户设备，控制台与 eiam 不代持、不导出
- **完成态**：注册成功后该用户「登录方式」标记为已绑定 passkey；注册会话过期未完成即作废，须重新引导
- **设备丢失重绑**：唯一路径 = 管理员撤销原 passkey 绑定 → 重新发起同一注册引导；重绑完成前用户凭密码/LDAP 登录不受影响。若 passkey 为该账号唯一登录方式且设备丢失，账号恢复走管理员重置密码路径（E1）

## UI Function 5: 组织管理

### Placement
- **Mode**: new-page
- **Target Page**: /organizations
- **Position**: 独立列表页

### Description
组织的树形/列表管理（新增/编辑/删除/架构层级）。

### User Interaction Flow
加载组织树 → 新增子组织/编辑/删除 → CRUD → 刷新。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 组织树 | tree | eiam | 层级 |
| 组织名 | string | 表单 | 必填 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | skeleton | 进入 |
| 空 | 空状态 | 无数据 |
| 树/列表 | 组织树 | 有数据 |
| 错误 | 提示 | 失败 |

### Validation Rules
- 组织名非空，1–64 字符；删除需校验无下属组织且无关联用户
- 删除被拒 UI 态：弹窗列出下属组织数/关联用户数，确认按钮置灰（E11）
- 组织树 ≤ 1000 节点、深度 ≤ 8，超限禁止再建子级

## UI Function 6: 租户管理

### Placement
- **Mode**: new-page
- **Target Page**: /tenants
- **Position**: 独立列表页

### Description
租户的列表/新增/编辑/禁用（平台管理员可见）。

### User Interaction Flow
列表 → 新增/编辑租户 → 提交 → 刷新；禁用需确认。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 租户列表 | array | eiam | 分页 |
| 租户名 | string | 表单 | 必填 |
| 状态 | enum | eiam | 启用/禁用 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | skeleton | 进入 |
| 空 | 空状态 | 无数据 |
| 列表 | 表格 | 有数据 |
| 错误 | 错误提示+重试 | 接口失败 |

### Validation Rules
- 仅平台管理员可见（只读/审计角色亦不渲染，口径与导航注记一致）；租户名唯一，1–64 字符
- 禁用前置（可测）：该租户下未过期会话数 = 0（活跃用户 = 存在未过期共享 session 的用户）；不满足则拒绝并显示活跃会话数

## UI Function 7: 身份源管理

### Placement
- **Mode**: new-page
- **Target Page**: /identity-sources
- **Position**: 独立列表页

### Description
登录身份源（LDAP 等）的配置与启停。

### User Interaction Flow
列表 → 新增/编辑身份源（连接参数）→ 保存/测试连接 → 启停。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 身份源列表 | array | eiam | |
| 连接地址 | string | 表单 | `ldap://` 或 `ldaps://host:port`，端口 1–65535 |
| Bind DN / Bind 密码 | string | 表单 | 必填；密码仅写入，掩码展示不回显明文 |
| Base DN | string | 表单 | 必填 |
| 用户属性映射 | object | 表单 | 用户名/邮箱/DN 属性名；用户名属性必填 |
| 超时 | number | 表单 | 1–30 秒，默认 5 |
| 启用状态 | enum | eiam | 启用/停用 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | skeleton | 进入 |
| 空 | 空状态 | 无配置 |
| 列表 | 表格 | 有数据 |
| 连接失败 | 错误提示 | 测试连接失败 |

### Validation Rules
- 地址/Base DN/Bind DN/密码/用户名属性非空，地址/端口/超时格式与范围校验；保存前可测试连接；凭据掩码不回显明文

## UI Function 8: 角色管理

### Placement
- **Mode**: new-page
- **Target Page**: /roles
- **Position**: 独立列表页

### Description
角色的列表/新增/编辑/删除（含绑定策略/用户）。

### User Interaction Flow
列表 → 新增/编辑角色 → 绑定策略与用户 → 提交 → 刷新。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 角色列表 | array | eiam | |
| 角色名 | string | 表单 | 必填 |
| 策略绑定 | ref[] | eiam | 多选 |
| 用户绑定 | ref[] | eiam | 多选 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | skeleton | 进入 |
| 空 | 空状态 | 无角色 |
| 列表 | 表格 | 有数据 |
| 错误 | 错误提示+重试 | 接口失败 |

### Validation Rules
- 角色名唯一，1–64 字符；删除前校验无用户绑定，被拒时提示绑定用户数并阻止（E11）
- 租户管理员仅限本租户；策略绑定下拉仅列 eiam 返回的可绑定策略，无新建/编辑策略入口（见 Story 4 口径）

## UI Function 9: 授权管理

### Placement
- **Mode**: new-page
- **Target Page**: /authorizations
- **Position**: 独立列表页

### Description
授权关系的列表/新增/撤销。

### User Interaction Flow
列表 → 新增授权（主体+资源+动作）→ 提交 → 撤销需确认。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 授权列表 | array | eiam | |
| 主体 | ref | eiam | 用户/角色，选项来自 eiam 返回的主体清单 |
| 资源 | ref | eiam | 选项来自 eiam 资源词表，不支持自由文本 |
| 动作 | ref | eiam | 选项来自所选资源关联的动作词表，不支持自由文本 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | skeleton | 进入 |
| 空 | 空状态 | 无授权 |
| 列表 | 表格 | 有数据 |
| 错误 | 错误提示+重试 | 接口失败 |

### Validation Rules
- 三元组均必填，且资源/动作只能从 eiam 返回的词表中选取；词表来源 = eiam 枚举端点（Phase 0 核查是否现成提供，若无进入缺口处置，UI 临时降级为 parity-checklist 登记的静态词表并标注来源版本）
- 唯一性：同一（主体, 资源, 动作）组合唯一；提交前前端按已加载列表预检提示「授权已存在」，eiam 服务端兜底拒绝并在页面提示，重复数据不入库
- 撤销需确认

## UI Function 10: 策略管理

### Placement
- **Mode**: new-page
- **Target Page**: /policies
- **Position**: 独立列表页

### Description
策略的列表/新增/编辑/删除。

### User Interaction Flow
列表 → 新增/编辑策略 → 提交 → 刷新。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 策略列表 | array | eiam | |
| 策略名称 | string | 表单 | 1–64 字符，唯一（eiam 校验） |
| 策略内容 | JSON | 表单 | Statement 数组：Effect ∈ {Allow, Deny}、Action 非空字符串数组、Resource 非空字符串数组、Condition 可选 |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | skeleton | 进入 |
| 空 | 空状态 | 无策略 |
| 列表 | 表格 | 有数据 |
| 错误 | 错误提示+重试 | 接口失败 |

### Validation Rules
- 仅平台管理员可见/可编辑（租户管理员与只读/审计角色入口均不渲染，口径与导航注记一致）
- 策略内容合法性判定：先前端 JSON 语法校验，再 eiam 服务端 schema 校验（字段名/枚举/结构）；任一失败标红对应字段并阻止保存
- 已绑定角色的策略删除被拒：提示绑定角色数并阻止

## UI Function 11: 租户选择

### Placement
- **Mode**: new-page
- **Target Page**: /tenant-select（认证通过后、进入工作台前）
- **Position**: 独立整页

### Description
按所属租户数量分流：0 个不进入本页，直达工作台受限视图（prd-spec.md E3；平台管理员豁免零租户限制）；1 个自动选定直达工作台；≥ 2 个展示所属租户列表供选择。tenant_id 写入共享 session（位置与上下文以 Phase 0 事实核查为准）。

### User Interaction Flow
认证通过 → 查询所属租户 → 0/1 个自动分流；多个：点选 → 写 session → 进工作台。登录后可经 UF-2 顶部导航重新切换，切换后刷新页面数据。

### Data Requirements
| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 所属租户列表 | array | eiam | 仅当前用户所属租户 |
| 选中租户 | tenant_id | 用户选择 | 写入共享 session |

### States
| State | Display | Trigger |
|-------|---------|---------|
| 多租户 | 租户选择列表 | 所属租户 ≥ 2 |
| 加载中 | skeleton | 查询所属租户 |
| 错误 | 错误提示+重试 | 接口失败 |

### Validation Rules
- 列表仅含 eiam 返回的所属租户，不可输入任意 tenant_id；未选定前不进入管理页

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| /login | new | UF-1 | 独立登录页 |
| 主布局 | new | UF-2 | 顶部导航通栏，所有登录后页共享 |
| /workbench | new | UF-3 | 登录后首页 |
| /users | new | UF-4 | 用户列表页 |
| /organizations | new | UF-5 | 组织列表页 |
| /tenants | new | UF-6 | 租户列表页 |
| /identity-sources | new | UF-7 | 身份源列表页 |
| /roles | new | UF-8 | 角色列表页 |
| /authorizations | new | UF-9 | 授权列表页 |
| /policies | new | UF-10 | 策略列表页（仅平台管理员） |
| /tenant-select | new | UF-11 | 认证后、工作台前（零/单租户自动跳过） |