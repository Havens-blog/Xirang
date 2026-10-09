---
feature: "WatchAlert 作为子项目并入运维平台（含 eiam 统一鉴权）"
---

# WatchAlert 子项目并入（含 eiam 统一鉴权） — UI Functions

> 需求层：定义 UI 要「做什么」。不定义「长什么样」（那是 ui-design）。

## UI Scope

本 feature 的 UI 面集中在告警平台前端（独立前端，不嵌入门户控制台）：登录入口改造 + 鉴权降级两态展示。共 3 个 UI 功能。

## Navigation Architecture

- **Platform**: web

### Primary Navigation (shared across pages)

| # | Label | Target Page | Icon Keyword |
|---|-------|-------------|-------------|
| 1 | 告警看板 | 告警概览 | 告警/alert |
| 2 | 告警规则 | 规则管理 | 规则/rule |
| 3 | 故障中心 | 当前事件 | 故障/event |

### Secondary Pages (navigated from a parent page)

| Page | Entry Point (UF# or action) | Return Target |
|------|-----------------------------|---------------|
| eiam 登录跳转页 | UF-1（未登录访问任意页） | 原访问页面 |

## UI Function 1: eiam 统一登录入口

### Placement
- **Mode**: existing-page
- **Target Page**: 登录页（原本地账密登录页改造）
- **Position**: 移除本地账密表单，改为「使用平台账号登录」跳转入口

### Description
取消本地账号密码登录，改为跳转 eiam 统一登录。

### User Interaction Flow
1. 用户访问未登录 → 系统检测未登录 → 展示「使用平台账号登录」按钮
2. 用户点击 → 跳转 eiam 登录
3. eiam 认证成功 → 回跳（回跳地址经白名单校验）→ 建立会话 → 进入告警看板
4. eiam 认证失败 → 回跳登录页展示错误信息，可重试
5. eiam 认证成功但无 WatchAlert 内部角色 → 返回 403，原地提示「无访问权限，请联系管理员配置角色」，不跳登录页（BIZ-auth-003）

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 登录状态 | 布尔 | 会话 | 决定是否跳转 eiam |
| 回跳地址 | 字符串 | 当前页 URL | 登录后回到原页面；须通过白名单校验 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 未登录 | 显示「使用平台账号登录」按钮 | 访问受保护页面 |
| 登录中 | 跳转中/加载态 | 点击登录后 |
| 登录失败 | 错误提示 + 重试按钮 | eiam 返回失败 |
| 无权限 | 403 原地提示「无访问权限，请联系管理员配置角色」 | 认证通过但无 `watchalert_role_binding` 匹配 |

### Validation Rules
- 不得保留任何本地用户名/密码输入框
- 回跳地址白名单/清洗：仅允许 host 匹配告警平台域名、path 以 `/` 开头的相对路径；禁止绝对外链、`//` 开头协议相对 URL、`javascript:`/`data:` scheme；校验失败回退到告警看板默认路径（防开放重定向）

---

## UI Function 2: 鉴权不可用宽限横幅

### Placement
- **Mode**: existing-page
- **Target Page**: 全站顶部（所有页面共享）
- **Position**: 页面顶部常驻横幅

### Description
eiam 不可用时在页面顶部展示宽限横幅，提示当前处于降级模式，写操作被禁止。

### User Interaction Flow
1. 系统检测 eiam 不可用且用户仍持有效会话 → 顶部展示横幅「鉴权服务暂不可用，已切换宽限模式」
2. 用户可继续只读查看告警列表/规则/历史事件
3. 用户尝试写操作（编辑规则/确认告警/改通知）→ 后端返回 503，前端提示不可写

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 宽限剩余时长 | 数值 | 后端状态 | 宽限期 2 小时；横幅展示「剩余 X 小时 X 分」 |
| 写操作许可 | 布尔 | 后端状态 | 决定是否禁用写按钮 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常（探针偶发失败） | 无横幅，正常使用 | 1-2 次探针失败（未达 3 次阈值），对用户不可见 |
| 宽限中 | 黄色横幅 + 写按钮禁用 | 探针连续 3 次失败（30 秒内）判定 eiam 不可用 + 会话有效 |
| 已恢复 | 横幅消失，恢复正常可用 | 宽限期内探针连续 3 次成功判定 eiam 恢复（滞回防抖） |
| 宽限过期 | 切至只读免登录视图（UF-3） | 宽限窗口（2 小时）到期 |

### Validation Rules
- 横幅须明确告知「哪些操作禁止」，而非仅提示「服务不可用」
- 1-2 次探针失败不得展示横幅（滞回防抖，避免抖动闪烁），仅连续 3 次失败才判定不可用

---

## UI Function 3: 只读免登录视图

### Placement
- **Mode**: existing-page
- **Target Page**: 告警看板（降级态）
- **Position**: 无操作按钮的只读页面

### Description
宽限过期后，或未登录用户在 eiam 不可用期间访问时，展示只读免登录视图：告警看板 + 规则只读 + 事件历史，无任何写操作与登录入口。恢复时由后端探针信号驱动前端自动跳转 eiam 登录页（非用户点击按钮入口），不引入被移除的本地登录路径。

### User Interaction Flow
1. 用户进入（宽限过期 或 未登录且 eiam 不可用）→ 看到只读看板
2. 页面提示「请联系管理员或等待鉴权服务恢复」
3. 后端探针连续 3 次成功判定 eiam 恢复 → 前端经 SSE/轮询获知探针信号后自动跳转 eiam 登录页（无用户操作按钮，非登录入口残留）

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 只读告警数据 | 列表 | 后端只读接口 | 看板/规则/事件 |
| 鉴权服务状态 | 布尔 | 后端探针 | 后端每 10s 探针 `GET /eiam/health`（超时 3s）；连续 3 次成功判定恢复，前端经 SSE/轮询获知 |

**匿名可见字段清单**：告警规则名称/级别/状态/触发条件、告警事件时间线（时间/级别/状态变更）、事件当前状态。

**敏感字段脱敏规则**：通知渠道凭据（webhook secret、API key）、数据源密码、webhook URL 查询参数等敏感字段一律以 `***` 脱敏不展示；后端只读接口在匿名态下不下发敏感字段原文。

### States

| State | Display | Trigger |
|-------|---------|---------|
| 只读免登录 | 只读看板（敏感字段脱敏）+ 恢复提示 | 宽限过期 / 未登录且 eiam 不可用 |
| 已恢复 | 自动跳转 eiam 登录页 | 探针连续 3 次成功判定恢复（滞回防抖，与进入不可用对称） |

### Validation Rules
- 只读视图不得出现任何登录入口（避免重新引入被移除的本地登录路径）；恢复跳转由后端探针信号驱动，非用户可点击的登录按钮

---

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| 登录页 | existing | UF-1 | 移除本地账密，改 eiam 跳转 |
| 全站顶部 | existing | UF-2 | 鉴权不可用时的宽限横幅 |
| 告警看板 | existing | UF-3 | 宽限过期/未登录的只读降级态 |