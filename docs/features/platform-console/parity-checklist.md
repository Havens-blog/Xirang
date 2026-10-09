---
feature: "platform-console"
created: "2026-09-29"
status: Draft
task: "1.1 端点差集盘点与 eiam 契约基线冻结"
---

# parity-checklist — 端点差集盘点与 eiam 契约基线

> Phase 0 清单宿主（PRD Goals「Phase 0 parity 覆盖率 100%」与 Data Requirements 契约基线条款的承载文件）。
> 本文件由任务 1.1 建立骨架与盘点主体；1.2（认证链路）、1.3（服务端授权）、1.4（能力映射）各自在 §6 占位章节补充内容。

## 0. 口径与方法

### 0.1 路径口径

控制台（e-cam-web 同款）请求 `:8888/api/iam/*`，nginx 反代到 eiam `:9000/api/*`（证据：`nginx-1.25.4/conf/vhost/nginx.dev.conf:8-9`，`location ^~ /api/iam/` → `proxy_pass http://127.0.0.1:9000/api/`）。
**eiam 代码内注册路径不含 `iam` 前缀**（如 eiam 内部 `POST /api/user/list` 对外即 `/api/iam/user/list`）。本文 §1 路由表按 **eiam 内部注册路径** 录入（证据即路由注册行），§2/§4 盘点与基线按 **控制台外部路径** 录入（与 tech-design Interfaces 口径一致），两者按此映射互换。

### 0.2 证据来源形态（承接 eval 残留 #5）

| 形态 | 定义 | 适用 |
|------|------|------|
| **证据-A** | e-cam-web 生产源码文件:行号（`src/api/eiam-users.ts` / `eiam-tenant.ts` / `eiam-mapper.ts` / `stores/user.ts`） | A 档端点 |
| **证据-R** | eiam 路由注册文件:行号（`internal/web/**` + `ioc/web.go`） | 全量路由导出、B/C 档路径存在性 |
| **证据-S** | eiam 请求/响应类型定义文件:行号（`internal/web/**/vo.go`、`internal/domain/*.go`） | 形状核对 |
| **证据-P** | 真实响应样本抓包（登录态请求实测） | B/C 档升格契约的必要条件 |
| **证据-O** | owner 书面确认（eiam owner / 业务 owner） | 缺口处置、契约基线冻结、cookie 矩阵 |

### 0.3 状态词表

| 状态 | 含义 |
|------|------|
| 实核吻合 | A 档：e-cam-web 生产在用，形状与 tech-design 逐字吻合 |
| 形状分歧 | 端点存在，但真实形状与 tech-design/UF 推定不一致（差异已登记，前端 api 层归一承接） |
| 待核验假设 | B/C 档：路径已实存（证据-R），但形状未经响应样本核验（证据-P 待补），**不作为契约** |
| 缺口 | 反推全集 − 实有端点的真差集，进入 §3 处置 |
| 有意不迁移 | eiam 实有但控制台 v1 明确不用（登记理由，非缺口） |

### 0.4 eiam 鉴权分层（路由注册顺序决定，证据：`eiam/ioc/web.go:40-73`）

| 层 | 机制 | 鉴权要求 |
|----|------|----------|
| **L0 公开层** | `PublicRoutes`（web.go:41-47） | 无鉴权 |
| **L1 登录层** | `session.CheckLoginMiddleware()`（web.go:50）+ `tenancyBuilder.Build()`（web.go:52）后注册的 `IdentityRoutes`（web.go:55-58） | 仅要求登录态（共享 Redis JWT session），无细粒度权限 |
| **L2 业务鉴权层** | `middleware.CheckPermission(permSvc)`（web.go:61）后注册的 `PrivateRoutes`（web.go:64-72） | 登录 + OPA/RBAC 细粒度校验；路由经 `h.Capability("名称","动作")` 声明能力项（可叠加 `Needs(...)` 附加权限） |

### 0.5 执行人与会签位

| 角色 | 职责 | 签收 |
|------|------|------|
| 平台管理员（盘点执行人） | §1 路由导出、§2 差集求值、§3 处置建议、§4 基线起草 | 见 §7 签收表 |
| eiam owner | §3 缺口处置决策会签、§4 契约基线冻结会签、认证链路兼容矩阵（1.2） | 见 §7 签收表 |
| e-cam-web 前端 owner | A 档证据侧会签（跨仓改动归属方，RC#1 双签前置） | 见 §7 签收表 |
| 业务 owner | 「放弃」类缺口的书面确认（本盘点无「放弃」项，保留栏位） | 见 §7 签收表 |

---

## 1. eiam 路由全量导出（实有端点集）

> 来源 = eiam 仓库路由注册文件逐行扫描（`grep -rn "\.POST(\|\.GET(\|\.DELETE(\|\.PUT(" internal/web ioc pkg/web` 复跑可复现）。共 **116 条**实注册路由；2 条已注释停用（`role/assign`、`policy/attach`，见注）。证据格式：`文件:行号`。

### 1.0 路由装配（证据-R：`eiam/ioc/web.go`）

装配顺序即鉴权分层：公开路由（:41-47）→ `session.CheckLoginMiddleware()`（:50）→ `tenancyBuilder.Build()`（:52）→ Identity 路由（:55-58）→ `middleware.CheckPermission`（:61）→ Private 路由（:64-72）。

### 1.1 L0 公开层（14 条，无鉴权）

| # | 方法 | eiam 内部路径 | 处理器/说明 | 证据 |
|---|------|---------------|-------------|------|
| 1 | POST | /api/user/signup | `SignupRequest` 注册 | internal/web/user/user.go:65 |
| 2 | POST | /api/user/ldap/login | LDAP 登录 | internal/web/user/user.go:66 |
| 3 | POST | /api/user/system/login | 本地密码登录 | internal/web/user/user.go:67 |
| 4 | POST | /api/user/bind/confirm | 账号绑定确认 | internal/web/user/user.go:68 |
| 5 | GET | /api/user/oidc/render | OIDC 授权跳转 | internal/web/user/user.go:71 |
| 6 | GET | /api/user/oidc/callback | OIDC 回调（飞书回调不带登录态） | internal/web/user/user.go:73 |
| 7 | POST | /api/user/passkey/login/start | passkey 登录挑战 | internal/web/user/user.go:76 |
| 8 | POST | /api/user/passkey/login/finish | passkey 登录完成 | internal/web/user/user.go:77 |
| 9 | POST | /api/user/login/mfa/verify | 登录阶段 MFA 校验 | internal/web/user/user.go:80 |
| 10 | POST | /api/permission/check_login | SDK 鉴权接口（自带 token 逻辑内校验） | internal/web/permission/handler.go:36 |
| 11 | POST | /api/permission/check_policy | 策略判定（SDK 用） | internal/web/permission/handler.go:37 |
| 12 | GET | /api/identity_source/enabled_providers | 已启用登录方式清单 | internal/web/identity_source/handler.go:26 |
| 13 | GET | /api/invitation/verify/:code | 邀请校验 | internal/web/invitation/handler.go:34 |
| 14 | POST | /api/v1/discovery/sync | 微服务能力资产上报（**无鉴权注记**，内部仅校验 service 非空——安全面列入 1.3 核查） | internal/web/discovery/handler.go:25,28-31 |

### 1.2 L1 登录层（13 条，需登录态）

| # | 方法 | eiam 内部路径 | 处理器/说明 | 证据 |
|---|------|---------------|-------------|------|
| 15 | GET | /api/permission/menus | 当前用户可见菜单（授权菜单树） | internal/web/permission/handler.go:44 |
| 16 | POST | /api/user/password/update | 本人改密（old/new/confirm） | internal/web/user/user.go:87 |
| 17 | GET | /api/user/profile | 当前用户 profile | internal/web/user/user.go:88 |
| 18 | POST | /api/user/logout | 登出（销毁 session） | internal/web/user/user.go:89 |
| 19 | POST | /api/user/passkey/register/start | passkey **本人**注册挑战（`sess.Claims().Uid`） | internal/web/user/user.go:92,668-674 |
| 20 | POST | /api/user/passkey/register/finish | passkey 本人注册完成（header `X-Passkey-Session`） | internal/web/user/user.go:93,707 |
| 21 | GET | /api/user/mfa/totp/setup | TOTP 绑定准备 | internal/web/user/user.go:96 |
| 22 | POST | /api/user/mfa/totp/bind | TOTP 绑定 | internal/web/user/user.go:97 |
| 23 | POST | /api/user/mfa/disable | MFA 停用 | internal/web/user/user.go:98 |
| 24 | GET | /api/user/identity/list | 本人外部身份清单（passkey 绑定态查询） | internal/web/user/user.go:101 |
| 25 | GET | /api/tenant/list/mine | 我的所属租户列表（NoSync，不参与权限同步） | internal/web/tenant/handler.go:44 |
| 26 | POST | /api/tenant/switch | 切换租户（`middleware.WTS`；header `X-Active-Tenant-ID`） | internal/web/tenant/handler.go:51；pkg/web/middleware/tenant.go:15,167 |
| 27 | POST | /api/invitation/accept | 接受邀请入驻 | internal/web/invitation/handler.go:39 |

### 1.3 L2 业务鉴权层（89 条，登录 + OPA 能力校验）

**user 模块（12 条，前缀 /api/user，证据 internal/web/user/user.go:104-149）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/user/create | 创建用户 add | user.go:107 |
| POST | /api/user/list | 用户列表 view | user.go:110 |
| POST | /api/user/update | 修改用户 edit | user.go:113 |
| GET | /api/user/detail | 用户详情 get（query 传参，智能识别 id/username） | user.go:117 |
| DELETE | /api/user/delete/:id | 删除用户 delete | user.go:120 |
| POST | /api/user/batch_delete | 批量删除 batch_delete | user.go:123 |
| POST | /api/user/ldap/search | 搜索 LDAP | user.go:127 |
| POST | /api/user/ldap/sync | 同步 LDAP | user.go:130 |
| POST | /api/user/ldap/refresh_cache | 刷新 LDAP 缓存 | user.go:133 |
| POST | /api/user/identity/manage | 治理外部身份 | user.go:137 |
| POST | /api/user/identity/unbind | 解绑外部身份 | user.go:140 |
| POST | /api/user/list/attached/role | 角色关联用户列表（载荷 `role_code`） | user.go:144；vo.go:172-175 |

**role 模块（15 条，前缀 /api/role，证据 internal/web/role/handler.go:37-103）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/role/create | 创建角色 add（Needs cmdb:codebook:view） | role/handler.go:41-43 |
| POST | /api/role/update | 修改角色 edit | role/handler.go:45 |
| POST | /api/role/list | 角色列表 view | role/handler.go:48 |
| GET | /api/role/detail/:code | 角色详情 get | role/handler.go:51 |
| DELETE | /api/role/delete/:id | 删除角色 delete | role/handler.go:54 |
| POST | /api/role/batch_delete | 批量删除 batch_delete | role/handler.go:57 |
| POST | /api/role/batch_assign | 批量分配角色 batch_assign（Needs iam:role:view） | role/handler.go:66-68 |
| POST | /api/role/batch_unassign | 批量移除角色 batch_unassign | role/handler.go:71 |
| POST | /api/role/unassign | 移除角色分配 unassign | role/handler.go:74 |
| POST | /api/role/analysis/inline | 分析内联策略 analysis | role/handler.go:78 |
| POST | /api/role/add_parent | 添加父角色 add_parent（Needs iam:role:view） | role/handler.go:81-83 |
| POST | /api/role/remove_parent | 移除父角色 remove_parent | role/handler.go:85 |
| POST | /api/role/parents | 获取父角色 view_parents | role/handler.go:88 |
| GET | /api/role/mine | 查看个人角色 view_mine | role/handler.go:93 |
| POST | /api/role/list/attached/user | 查询用户角色 view_user_roles（载荷 `user_id`） | role/handler.go:98 |

> 注：`POST /api/role/assign` 已注释停用（role/handler.go:62-64），不计入实有端点。

**policy 模块（12 条，前缀 /api/policy，证据 internal/web/policy/handler.go:40-97）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/policy/create | 创建策略 add（Needs iam:permission:manifest） | policy/handler.go:43-45 |
| POST | /api/policy/update | 修改策略 edit（Needs iam:permission:manifest, iam:policy:get） | policy/handler.go:47-49 |
| POST | /api/policy/list | 策略列表 view | policy/handler.go:51 |
| GET | /api/policy/detail/:code | 策略详情 get | policy/handler.go:54 |
| POST | /api/policy/detach | 解绑策略 detach | policy/handler.go:62 |
| POST | /api/policy/batch-attach | 批量绑定策略 batch_attach（Needs iam:policy:view, iam:permission:search_subjects） | policy/handler.go:65-67 |
| POST | /api/policy/batch-detach | 批量解绑策略 batch_detach | policy/handler.go:69 |
| DELETE | /api/policy/delete/:code | 删除策略 delete | policy/handler.go:74 |
| POST | /api/policy/batch-delete | 批量删除策略 batch_delete | policy/handler.go:77 |
| POST | /api/policy/list/attached/user | 查询用户策略（载荷 `user_id`） | policy/handler.go:82 |
| POST | /api/policy/list/attached/role | 查询角色策略（载荷 `role_code`） | policy/handler.go:88 |
| POST | /api/policy/list/attached/group | 查询用户组策略 | policy/handler.go:94 |

> 注：`POST /api/policy/attach`（单体绑定）已注释停用（policy/handler.go:59-61），不计入实有端点。

**tenant 模块（13 条，前缀 /api/tenant，证据 internal/web/tenant/handler.go:57-116）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/tenant/create | 创建租户空间 add | tenant/handler.go:60 |
| POST | /api/tenant/list | 全量租户列表 view | tenant/handler.go:65 |
| POST | /api/tenant/list/by-ids | 批量查询租户 view_by_ids（NoSync） | tenant/handler.go:68 |
| POST | /api/tenant/update | 修改租户信息 edit（含 `status` 字段） | tenant/handler.go:72；vo.go:14-20 |
| DELETE | /api/tenant/delete/:id | 删除租户空间 delete | tenant/handler.go:75 |
| POST | /api/tenant/batch_delete | 批量删除 batch_delete | tenant/handler.go:78 |
| GET | /api/tenant/detail/:id | 租户详情 get | tenant/handler.go:81 |
| POST | /api/tenant/members | 查看租户成员 view_members（WithTenantOverride） | tenant/handler.go:86 |
| POST | /api/tenant/assign | 分配租户成员 assign（WithTenantOverride） | tenant/handler.go:92 |
| POST | /api/tenant/unassign | 移除租户成员 unassign（WithTenantOverride） | tenant/handler.go:96 |
| POST | /api/tenant/batch_assign | 批量分配 batch_assign | tenant/handler.go:100 |
| POST | /api/tenant/batch_unassign | 批量移除 batch_unassign | tenant/handler.go:104 |
| POST | /api/tenant/list/attached/user | 查询用户所属租户 view_user_tenants（WithTenantOverride，载荷 `user_id`） | tenant/handler.go:110 |

**department 模块（8 条，前缀 /api/department，证据 internal/web/department/handler.go:26-51）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/department/create | 创建部门 add | department/handler.go:29 |
| POST | /api/department/update | 修改部门 edit | department/handler.go:32 |
| DELETE | /api/department/delete/:id | 删除部门 delete | department/handler.go:35 |
| GET | /api/department/list | 部门树 view | department/handler.go:38 |
| GET | /api/department/detail/:id | 部门详情 get | department/handler.go:41 |
| POST | /api/department/assign | 分配部门成员 assign | department/handler.go:44 |
| POST | /api/department/remove | 移除部门成员 remove | department/handler.go:47 |
| POST | /api/department/members | 部门成员列表 members | department/handler.go:50 |

**group 模块（13 条，前缀 /api/group，证据 internal/web/group/handler.go:27-67）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/group/create | 创建用户组 add | group/handler.go:30 |
| POST | /api/group/update | 修改用户组 edit | group/handler.go:33 |
| DELETE | /api/group/delete/:id | 删除用户组 delete | group/handler.go:36 |
| POST | /api/group/list | 用户组列表 view | group/handler.go:39 |
| GET | /api/group/detail/:code | 用户组详情 get | group/handler.go:42 |
| POST | /api/group/members/assign | 分配组成员 | group/handler.go:45 |
| POST | /api/group/members/remove | 移除组成员 | group/handler.go:48 |
| POST | /api/group/members | 组成员列表 | group/handler.go:51 |
| POST | /api/group/list/attached/user | 查询用户所属分组 | group/handler.go:54 |
| POST | /api/group/list/attached/role | 查询角色关联分组 | group/handler.go:57 |
| POST | /api/group/role/assign | 用户组分配角色 | group/handler.go:60 |
| POST | /api/group/role/remove | 用户组解绑角色 | group/handler.go:63 |
| GET | /api/group/roles/:code | 查看用户组角色 | group/handler.go:66 |

**permission 模块（4 条，前缀 /api/permission，证据 internal/web/permission/handler.go:47-71）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| GET | /api/permission/manifest | 权限资产清单 manifest（Needs iam:permission:menus_by_urns） | permission/handler.go:51-53 |
| POST | /api/permission/authorizations | 授权治理列表 view_authorizations | permission/handler.go:57 |
| POST | /api/permission/subjects/search | 搜索授权主体 search_subjects | permission/handler.go:62 |
| POST | /api/permission/menus/by_urns | 批量按 URN 查菜单（NoSync） | permission/handler.go:67 |

**identity_source 模块（6 条，前缀 /api/identity_source，证据 internal/web/identity_source/handler.go:29-49）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/identity_source/save | 保存身份源 save（upsert 语义） | identity_source/handler.go:32 |
| POST | /api/identity_source/list | 身份源列表 view | identity_source/handler.go:35 |
| DELETE | /api/identity_source/delete/:id | 删除身份源 delete | identity_source/handler.go:38 |
| GET | /api/identity_source/detail/:id | 身份源详情 detail | identity_source/handler.go:41 |
| POST | /api/identity_source/test | 测试身份源连接 test | identity_source/handler.go:44 |
| POST | /api/identity_source/toggle/:id | 切换启用状态 toggle | identity_source/handler.go:47 |

**invitation 模块（6 条，前缀 /api/invitation，证据 internal/web/invitation/handler.go:42-64）**

| 方法 | 路径 | 能力项 | 证据 |
|------|------|--------|------|
| POST | /api/invitation/create | 创建邀请 add | invitation/handler.go:45 |
| POST | /api/invitation/list | 邀请列表 view | invitation/handler.go:49 |
| DELETE | /api/invitation/revoke/:code | 撤回邀请 delete | invitation/handler.go:52 |
| POST | /api/invitation/batch_revoke | 批量撤回 batch_delete | invitation/handler.go:55 |
| POST | /api/invitation/requests | 申请列表 view_requests | invitation/handler.go:60 |
| POST | /api/invitation/requests/handle | 处理申请 handle_request | invitation/handler.go:63 |

---

## 2. 差集求值：UF-1~11 反推端点全集 × eiam 实有端点

> 方法：以 prd-ui-functions.md UF-1~11 逐条反推控制台所需端点，与 §1 实有端点求差。**路径列 = 控制台外部路径**（`/api/iam/*`）。A 档证据 = e-cam-web 源文件:行号；B/C 档 = eiam 路由/类型文件:行号 + 待补响应样本。

### 2.0 共享契约

| 项 | tech-design 声明 | 实核结论 | 状态 | 证据 |
|----|------------------|----------|------|------|
| 响应信封 | `Envelope<T>{code,msg,data}`，`msg` 非 `message`，成功 `code===0` | e-cam-web 实盘（eiam-mapper 实核结论，tech-design.md:117-121 沿用） | 实核吻合 | e-cam-web/src/api/eiam-mapper.ts（A 档既有实核） |
| 列表请求 | `ListQuery{offset,limit,keyword}` | eiam `BaseListRequest` 同形 | 实核吻合 | eiam/internal/web/user/vo.go:154-158 |
| 列表响应 | `{total, users\|roles\|tenants}`，api 层归一 `Page<T>{total,items}` | eiam `RetrieveUsers[T]{total,users}` / `ListTenantRes{total,tenants}` 同形 | 实核吻合 | eiam/internal/web/user/vo.go:129-132；tenant/vo.go:43-46 |

### 2.1 UF-1 登录页

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/user/system/login | B | 待核验假设（路径实存） | 请求 `LoginRequest{username,password}` ✓；响应非 tech-design 推定 `LoginResult{mfaRequired,mfaTicket,token,lockUntil}`——见 B-1/B-2 形状分歧 | eiam/internal/web/user/user.go:67；vo.go:27-30；domain/user.go:143-155 |
| POST /api/iam/user/ldap/login | B | 待核验假设 | 请求无 `source_id`（推定多余）→ 分歧 B-3；handler 固定 `executeLogin(ctx,"ldap",...)` | eiam/internal/web/user/user.go:66,194-196 |
| POST /api/iam/user/passkey/login/start | B | 待核验假设 | 响应 `{options, session_token}`，推定字段 `ticket` 实为 `session_token` → 分歧 B-4 | eiam/internal/web/user/user.go:76；vo.go:107-110 |
| POST /api/iam/user/passkey/login/finish | B | 待核验假设 | 凭据走 HTTP body（WebAuthn protocol 解析），会话标识走 header，非推定 `{ticket, assertion}` body → 分歧 B-4 | eiam/internal/web/user/user.go:77,707 |
| POST /api/iam/user/login/mfa/verify | B | 待核验假设 | 请求 `{mfa_token, code}`，推定 `mfa_ticket` 实为 `mfa_token` → 分歧 B-1 | eiam/internal/web/user/user.go:80；vo.go:251-254 |
| GET /api/identity_source/enabled_providers | C | 待核验假设 | 登录页按启用方式渲染入口（公开无鉴权） | eiam/internal/web/identity_source/handler.go:26 |

**锁定语义实核**（E2/E9 数据源）：密码锁定=超 `MaxFailedAttempts` 后返回 `ErrUserLocked`（**纯文案** error，无 lockUntil 数据）——eiam/internal/service/user/login.go:37-52、internal/errs/error.go:22；MFA 连续 5 次 → `ErrMfaAttemptsExhausted`，失败文案携带剩余次数——internal/service/user/mfa.go:83-88。**`lockUntil` 字段不存在 → 降级中性固定文案**（tech-design B 档预置降级路径），分歧 B-2。

### 2.2 UF-2 顶部导航 + 应用切换器

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| GET /api/iam/user/profile | A | 实核吻合 | `RetrieveUser`：`user/tenants/current_tenant_id/must_select_tenant/is_admin/permissions` ✓，另含推定未列的 `must_bind/bind_token/mfa_required/mfa_token`（补充登记） | e-cam-web/src/stores/user.ts:45；eiam/internal/web/user/user.go:346-378；vo.go:134-145 |
| POST /api/iam/user/logout | A | 实核吻合 | 销毁共享 session → `null` | e-cam-web/src/stores/user.ts:83-85；eiam/internal/web/user/user.go:89,380-388 |
| GET /api/iam/permission/menus | C | 待核验假设 | 授权菜单树（`Menu{parent_id,name,path,component,meta}`），用于导航渲染 | eiam/internal/web/permission/handler.go:44；vo.go:60-76 |

**平台导航面板（task 6.1，2026-10-09）**：纯前端增强无新端点——AppSwitcher v2 面板（参照 e-cam-web PlatformNav）穷举平台前端应用 = nginx 同域形态 `/console/` + `/cam/`
（nginx-1.25.4/conf/vhost/nginx.dev.conf:57-88 实核）；云管深链 `/cam/<path>` 新标签
`window.open(_blank, noopener)`、共享 cookie SSO；收藏 localStorage 隔离键
`haven-platform-nav-favorites`。触发器位置：header 左侧 **#header-nav**（面包屑「首页」
紧邻右侧，task 6.4 由右侧 actions 迁入）。记录：tasks/records/6.1-app-switcher.md；
REAL 验证 scripts/verify-6-appnav-policyed-real.cjs（面板 7 断言）+
verify-6.4-nav-header-real.cjs（位置 4 断言）。

### 2.3 UF-3 工作台

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| GET /api/iam/user/profile（复用） | A | 实核吻合 | 我的信息 = profile 的 user 字段；eiam `User` VO 含 `nickname`（tech-design `displayName` 的映射源）、`email/avatar/job_title/phone/status/source/mfa_type/mfa_bound/identities` | eiam/internal/web/user/vo.go:40-57 |

### 2.4 UF-4 用户管理

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/user/list | A | 实核吻合 | `{offset,limit,keyword}` → `{total,users}`；eiam 实返回 `UserMemberVO`（附加 `is_member/is_system_space`，补充登记）；列表请求另支持 `usernames/ids` 过滤（推定未列） | e-cam-web/src/api/eiam-users.ts:17；eiam/internal/web/user/user.go:110；vo.go:61-70,160-164 |
| POST /api/iam/user/create | A | 实核吻合 | 载荷 `SignupRequest` 含 `password/confirm_password` → **初始密码发放经 create 承载**（UF-4「新增发放初始密码」成立） | e-cam-web/src/api/eiam-users.ts:33；eiam/internal/web/user/vo.go:20-24 |
| POST /api/iam/user/update | A | 实核吻合 | `UpdateUserReq = ID + BaseUserRequest`，**无 password 字段**（tech-design 断言实核成立：重置密码无法走 update）；`status` 启停经此承载 | e-cam-web/src/api/eiam-users.ts:38；eiam/internal/web/user/vo.go:9-17,124-127 |
| DELETE /api/iam/user/delete/:id | A | 实核吻合 | → `null` | e-cam-web/src/api/eiam-users.ts:42；eiam/internal/web/user/user.go:120 |
| GET /api/iam/user/detail | C | 待核验假设 | query 传参（id 或 username），用户详情/编辑回显用 | eiam/internal/web/user/user.go:116-119 |
| POST /api/iam/user/list/attached/role | C | 待核验假设 | 载荷 `{role_code, offset,limit,keyword}` → 角色已绑用户清单（**反向承接 C 档推定的 `role/attached/users`，路径改判非缺口**） | eiam/internal/web/user/user.go:144；vo.go:172-175 |
| POST /api/iam/role/list/attached/user | A | 实核吻合 | `{user_id, offset, limit}` → `{total, roles}`；用户编辑页角色回显 | e-cam-web/src/api/eiam-users.ts:70；eiam/internal/web/role/handler.go:98 |
| POST /api/iam/role/batch_assign | A | 实核吻合 | `{usernames, role_codes}` ✓（角色分配以 code 为准） | e-cam-web/src/api/eiam-users.ts:84；eiam/internal/web/role/handler.go:66 |
| POST /api/iam/role/batch_unassign | A | 实核吻合 | `{usernames, role_codes}` ✓ | e-cam-web/src/api/eiam-users.ts:90；eiam/internal/web/role/handler.go:71 |
| POST /api/iam/tenant/list/attached/user | A | 实核吻合 | `{user_id, offset, limit}` → `{total, tenants}` | e-cam-web/src/api/eiam-users.ts:113；eiam/internal/web/tenant/handler.go:110 |
| POST /api/iam/tenant/batch_assign | A | 实核吻合 | `{user_ids, tenant_ids}` ✓ | e-cam-web/src/api/eiam-users.ts:125；eiam/internal/web/tenant/handler.go:100 |
| POST /api/iam/tenant/batch_unassign | A | 实核吻合 | `{user_ids, tenant_ids}` ✓ | e-cam-web/src/api/eiam-users.ts:131；eiam/internal/web/tenant/handler.go:104 |
| POST /api/iam/user/reset_password | C | **缺口 G-1** | 全仓 grep `reset_password|ResetPassword` 0 命中（internal/**） | 无路由证据（缺口判定 = 反推全集 − §1 全量导出） |
| （管理员发起 passkey 注册引导） | C | **缺口 G-2** | `passkey/register/start|finish` 为 **L1 本人自服务**（取 `sess.Claims().Uid`），不支持管理员对目标用户发起 | eiam/internal/web/user/user.go:92-93,668-674,696-702 |
| DELETE /api/iam/user/passkey/:id | C | **缺口 G-3**（替代路径待核验） | 无此路由；潜在承接 = `POST /api/iam/user/identity/unbind`（`{user_id, provider, identity_id}`，provider=passkey 是否生效待响应样本核验） | eiam/internal/web/user/user.go:140；vo.go:199-203 |

### 2.5 UF-5 组织管理

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/organization/list（推定） | C | **形状分歧 D-1**（语义映射，非缺口） | eiam **无 organization** 端点；组织能力由 **department（部门）** 承载：`GET /api/department/list`（部门树）、`create/update/delete/:id/detail/:id` | eiam/internal/web/department/handler.go:26-51 |
| （其余组织 CRUD） | C | 待核验假设 | `DepartmentVO{id,parent_id,name,sort,leaders,main_leader,...}`——比 tech-design `Organization{id,name,parentId}` 多 sort/leaders 字段；`GET` list 非 `POST`；≤1000 节点/深度 ≤8 限制无 eiam 侧校验证据（前端拦截待登记） | eiam/internal/web/department/vo.go:4-9,37-41 |
| 删除被拒阻断计数（下属组织数/关联用户数，E11） | — | 待核验假设 | eiam 删除被拒的响应形态待响应样本核验 | 归 1.3 核查 |

> eiam 另有 **group（用户组）** 模型（§1.3 group 13 条），与「组织」不同概念；控制台 v1 无用户组管理页 → **有意不迁移**（能力映射归 1.4）。

> **路径前缀事故修复（2026-10-09）**：本族接入时误将 eiam 内部路径
> `/api/department/*` 直接用作控制台外部路径（违反 `/api/iam/*` 统一口径）——
> nginx/vite 仅反代 `/api/iam` 前缀，nginx 宿主下该请求落 `location /` 301 →
> 信封解析失败 → 组织页/「添加根组织」报「身份服务暂不可用」（eiam.unavailable）。
> 已修复为 `/api/iam/department/*`（records/6.3-org-prefix.md；单测路径前缀回归守卫
> organizations.test.ts；REAL 验证 verify-6.3-org-real.cjs 5/5：真实 create 200 →
> 树出现 → delete 200 自清理）。

### 2.6 UF-6 租户管理

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/tenant/list | A | 实核吻合 | `{total, tenants}`；`TenantVO` 含推定未列的 `status/ctime`（对禁用 UI 有用，补充登记） | e-cam-web/src/api/eiam-users.ts:102；eiam/internal/web/tenant/handler.go:65；vo.go:52-59 |
| POST /api/iam/tenant/create | C | 待核验假设 | `{name, code}`（`code` 必填 max32——tech-design `Tenant` 模型未列 code 唯一性约束，补充登记） | eiam/internal/web/tenant/handler.go:60；vo.go:9-12 |
| POST /api/iam/tenant/update | C | 待核验假设 | `{id,name,code,domain,status}`；`status` int 语义 1-活跃/2-禁用（**禁用经 update 承载**，无独立 toggle 端点） | eiam/internal/web/tenant/handler.go:72；vo.go:14-20；domain/tenant.go:9 |
| DELETE /api/iam/tenant/delete/:id | C | 待核验假设 | → `null` | eiam/internal/web/tenant/handler.go:75 |
| GET /api/iam/tenant/detail/:id | C | 待核验假设 | 租户详情 | eiam/internal/web/tenant/handler.go:81 |
| （禁用前置「未过期会话数=0」查询端点） | — | **缺口 G-4** | 无任何会话计数端点；tenant 服务内无活跃会话前置校验证据 | 全量导出（§1）无对应路由 |

### 2.7 UF-7 身份源管理

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/identity_source/list | C | 待核验假设（路径实存） | 存在 ✓（C 档预填 `identity_source/list` 吻合） | eiam/internal/web/identity_source/handler.go:35 |
| POST /api/iam/identity_source/save | C | 待核验假设 + 形状分歧 D-2 | **save 为 upsert**（无独立 create/update）；载荷 `{id,name,type,enabled,ldap{...}}`，与 tech-design 扁平 `ConnParams` 不一致 | eiam/internal/web/identity_source/handler.go:32；vo.go:6-17 |
| POST /api/iam/identity_source/test | C | 待核验假设 | 保存前测试连接 ✓（载荷同 save） | eiam/internal/web/identity_source/handler.go:44 |
| POST /api/iam/identity_source/toggle/:id | C | 待核验假设 | 启停 ✓（UF-7「启停」成立） | eiam/internal/web/identity_source/handler.go:47 |
| DELETE /api/iam/identity_source/delete/:id | C | 待核验假设 | ✓ | eiam/internal/web/identity_source/handler.go:38 |
| GET /api/iam/identity_source/detail/:id | C | 待核验假设 | 凭据掩码不回显行为待响应样本核验 | eiam/internal/web/identity_source/handler.go:41 |

**形状分歧 D-2 明细**：`LDAPVO{url, base_dn, bind_dn, bind_password, username_attribute, mail_attribute, display_name_attribute, title_attribute, user_filter, sync_user_filter}`（vo.go:20-36）——**无独立 `port`（含于 url）、无 `timeoutSec` 字段**；UF-7 表单字段「端口 1–65535 / 超时 1–30 秒默认 5」在 eiam 侧无对应载荷位，须登记处置。

### 2.8 UF-8 角色管理

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/role/list | A | 实核吻合 | `{total, roles}`；角色字段 `id/code/name/desc` ✓（`desc` 非 `description`） | e-cam-web/src/api/eiam-users.ts:53；e-cam-web/src/api/eiam-mapper.ts:80；eiam/internal/web/role/vo.go:54-63 |
| POST /api/iam/role/create | C | 待核验假设 | `{name, code, desc}`（tech-design 未列 `code` 必填，补充登记） | eiam/internal/web/role/handler.go:41；vo.go:5-9 |
| POST /api/iam/role/update | C | 待核验假设 | `{id, name, code, desc}` | eiam/internal/web/role/handler.go:45；vo.go:11-16 |
| DELETE /api/iam/role/delete/:id | C | 待核验假设 | 删除被拒（有用户绑定）响应形态待核验（E11） | eiam/internal/web/role/handler.go:54 |
| GET /api/iam/role/detail/:code | C | 待核验假设 | 角色详情/绑定页回显 | eiam/internal/web/role/handler.go:51 |
| POST /api/iam/role/policy/batch_assign（推定） | C | **形状分歧 D-3**（路径改判，非缺口；载荷实核） | 推定路径不存在；**绑定策略由 policy 侧承载**：`POST /api/iam/policy/batch-attach`，载荷 `{subjects:[{type, code}], policy_codes}`——SubjectItem 主体键为 **type**（非 sub_type），主体可 user/role/group 混合（实核后修正，「编辑角色」事故 @2026-10 同批） | eiam/internal/web/policy/handler.go:65；vo.go:84-89 |
| POST /api/iam/role/policy/batch_unassign（推定） | C | **形状分歧 D-3**（路径改判；载荷实核） | 解绑 = `POST /api/iam/policy/batch-detach`，载荷 `{assignments:[{sub_type, sub_code, policy_code}]}`（Assignment 扁平逐条，**无 subject 嵌套**；另有单体 `policy/detach`） | eiam/internal/web/policy/handler.go:62,69；vo.go:133-144 |
| POST /api/iam/role/attached/policies（推定） | C | **形状分歧 D-3**（路径改判；载荷实核） | 角色已绑策略清单 = `POST /api/iam/policy/list/attached/role`，载荷 `{role_code, offset, limit, keyword}`——**type 为 uint8 且 0=不过滤全类型，实际不带**；误传字符串 `type:"role"` 触发 gin JSON 绑定失败 → HTTP 400（「编辑角色 HTTP 400 Bad Request」事故根因） | eiam/internal/web/policy/handler.go:88；vo.go:57-63；repository/dao/policy.go:246 |
| POST /api/iam/role/list/attached/user | A | 实核吻合 | **user→roles 方向**（用户已绑角色），与 C 档推定的「role→users 清单」方向相反，后者由 `user/list/attached/role` 承接（见 §2.4） | e-cam-web/src/api/eiam-users.ts:70；eiam/internal/web/role/handler.go:98 |

### 2.9 UF-9 授权管理

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/permission/authorizations | C | 待核验假设 + **形状分歧 D-4（模型级）** | 授权列表 `{offset,limit,keyword,sub_type,obj_type}` → `{total, authorizations}`；`Authorization{subject, target, sub_type, obj_type, ...}`——eiam 授权模型为 **主体→目标（role\|policy）绑定关系**，非 PRD 的 subject/resource/action 三元组 | eiam/internal/web/permission/handler.go:57；vo.go:78-101 |
| POST /api/iam/permission/subjects/search | C | 待核验假设 | 主体词表 ✓（`{keyword,sub_type,offset,limit}` → `{total, subjects}`；sub_type=user/role/group） | eiam/internal/web/permission/handler.go:62；vo.go:103-113 |
| （资源/动作词表枚举端点） | — | **缺口 G-5** | 无 resources/actions 词表端点；`GET /api/iam/permission/manifest` 返回的是权限资产清单 `Manifest{actions:[Permission], services:[...]}`（action 类能力项），**不是** UF-9 语义的资源词表 | eiam/internal/web/permission/handler.go:51；vo.go:23-26 |
| （授权新增/撤销端点） | — | **缺口 G-6**（语义映射候选） | 无三元组授权 create/revoke 端点；语义映射候选 = policy `batch-attach/batch-detach`（授权 ≡ 策略绑定）与 role `batch_assign/batch_unassign`（主体→角色），**模型语义差异须业务 owner 确认** | §1.3 policy/role 模块路由 |

### 2.10 UF-10 策略管理

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| POST /api/iam/policy/list | C | 待核验假设（路径实存） | ✓ 存在 | eiam/internal/web/policy/handler.go:51 |
| POST /api/iam/policy/create | C | 待核验假设 + **形状分歧 D-5（tech-design 重点核查项已结）** + **实核补记（2026-10-09）** | **eiam 原生即 Statement 数组**（`CreatePolicyReq.Statement []Statement`）——tech-design「扁平单 effect 还是 Statement[]」核查点结论：Statement[]，归一规则无需触发；**成功响应不回传 id**：handler.go:201 `ginx.Result{Msg:"创建成功"}` 无 Data（实测 data:null）——前端 createPolicy 以信封 code==0 为成功（api/policies.ts void 契约） | eiam/internal/web/policy/handler.go:43,185-201；vo.go:5-11；scripts/verify-6-appnav-policyed-real.cjs（REAL 实测） |
| POST /api/iam/policy/update | C | 待核验假设 | `{name, code, desc, statement[]}`；Requires iam:permission:manifest + iam:policy:get（handler.go:48） | eiam/internal/web/policy/handler.go:47-48；vo.go:13-18 |
| DELETE /api/iam/policy/delete/:code | C | 待核验假设 | 按 **code** 删除（非 id）；`Policy.assignment_count` 可承载删除被拒提示（绑定角色数） | eiam/internal/web/policy/handler.go:74；vo.go:20-29 |
| GET /api/iam/policy/detail/:code | C | 待核验假设 | ✓ | eiam/internal/web/policy/handler.go:54 |
| GET /api/iam/permission/manifest | B | **升格实核（2026-10-09）** | 响应 `Manifest{actions:[PermissionVO], services:[...]}`；PermissionVO `{id,service,group,code,name,has_menu,menu_urns,access_scope_presets}`（vo.go:5-19）；Needs 双码 manifest+menus_by_urns（handler.go:51-53），**实测 dev admin profile 无 menus_by_urns 码仍 200（super 豁免）** → 控制台词表 gate = manifest 单码 + 失败降级自由输入 | eiam/internal/web/permission/handler.go:51-53；vo.go:5-19；REAL 实测：manifest 200、132 动作 + 2 services（verify-6 脚本 + 探测记录） |

**形状分歧 D-5 明细**（JSON 字段名与 UF-10 词表差异，前端 api 层归一承接）：
- 容器键为单数 `statement`（非 `statements`）——vo.go:10
- Statement 成员键为单数 `effect/action/resource/condition`（非 `actions/resources`）+ eiam 附加 `access_scope`——vo.go:31-37
- `Effect` 词表 `"Allow"/"Deny"`（首字母大写）✓ 与 UF-10 一致——domain/policy.go:27-31
- 策略实体额外字段 `code/desc/type/assignment_count`——vo.go:20-29

### 2.11 UF-11 租户选择

| 控制台外部路径 | 档 | 状态 | 形状核对 | 证据 |
|----------------|----|------|----------|------|
| GET /api/iam/tenant/list/mine | C | 待核验假设（路径实存） | 我的所属租户（L1，NoSync） | eiam/internal/web/tenant/handler.go:44 |
| POST /api/iam/tenant/switch | A | 实核吻合 | header `X-Active-Tenant-ID: <id>` ✓（eiam 侧常量 `ActiveTenantHeaderKey` 实核）；「销毁旧 session 重签 JWT」断言已复核（§6.1.6：重签 ✓ live 实核，销毁无 JWT 吊销语义 → F-3） | e-cam-web/src/api/eiam-tenant.ts:13-15；eiam/internal/web/tenant/handler.go:51；pkg/web/middleware/tenant.go:15,167 |

---

## 3. 差集清单与缺口处置（三选一）

> 处置三选一：**eiam 排期补齐 / UI 降级（仅查看或替代路径）/ 放弃（需业务 owner 书面确认）**。本盘点共 **6 项缺口（G-1~G-6）**、**5 项形状分歧（D-1~D-5）**；无「放弃」项（保留签收栏位）。任务 1.3 服务端授权核查追加授权面缺口 **G-7**（§3.1 末行；核查详情见 §6.2）。任务 1.4 能力映射追加 **G-8/G-9**（§3.1 末两行；核查详情见 §6.3.5）。处置为盘点执行人建议，**生效以 eiam owner / 业务 owner 会签为准**（决策门 = Phase 0 出口判据，PRD Goals）。

> **缺口处置决策（Phase 0 出口门 feature owner 定案，2026-09-30）**：G-1~G-9 处置口径已定案——默认 = **UI 降级/裁剪**（按 tech-design 预置降级路径在控制台侧承接，不前置 eiam 排期）；**G-7 为唯一例外（安全项）= 需 eiam owner 排期确认（安全）**，不走通用降级默认。逐项决策记录见 **§3.1.1**；§3.1 表内「建议处置」列保留盘点执行人原始建议作对照。全部决策签收/会签位（证据-O 人工门）**仍留待签，不代签**（§0.5 口径）。eiam 侧后续若排期补齐任一缺口，经 §4.2 基线追加登记流程升格即可，不受本决策阻碍。

### 3.1 缺口（反推全集 − 实有端点）

| 编号 | 缺口 | 影响 | 建议处置（三选一） | 理由 | 证据形态 | 决策签收 |
|------|------|------|--------------------|------|----------|----------|
| G-1 | `POST /api/iam/user/reset_password`（管理员重置密码，一次性初始密码 + 首登强制改密） | E1 / UF-4 | **eiam 排期补齐** | E1 为硬性验收场景且 PRD Out of Scope 明确 v1 无自助改密、唯一路径=管理员重置；无任何现役端点可承接（update 无 password 字段、password/update 为本人自服务） | 证据-R（§1 全量导出无此路由）+ 证据-S（vo.go:124-127） | eiam owner：＿＿＿＿ |
| G-2 | 管理员对目标用户发起 passkey 注册引导（15 分钟一次性会话） | UF-4（passkey 绑定） | **eiam 排期补齐**；备选降级：v1 控制台仅展示 passkey 绑定态（identity/list），绑定动作引导用户走 eiam 既有自助流程，账号恢复走 G-1 重置 | 现役 register/start|finish 为 L1 本人自服务（uid 取自 session）；管理员代发起与安全模型（私钥仅存用户设备）不冲突，缺的只是「为指定 uid 开启注册会话」的管理侧端点 | 证据-R + 证据-S（user.go:668-674） | eiam owner：＿＿＿＿ |
| G-3 | passkey 撤销（`DELETE /api/iam/user/passkey/:id` 推定路由不存在） | UF-4（设备丢失重绑前置） | **UI 降级 + 替代路径核验**：优先核验 `POST /api/iam/user/identity/unbind`（provider=passkey）是否可作撤销承载；核验不过 → 并入 G-2 由 eiam 排期补齐 | identity/unbind 语义即「解绑外部身份」，载荷含 user_id+provider+identity_id；是否覆盖 passkey 且校验完整性待响应样本 | 证据-R + 证据-P（待抓包） | eiam owner：＿＿＿＿ |
| G-4 | 租户禁用前置「未过期会话数=0」的查询端点（活跃会话计数） | UF-6 / E11 | **UI 降级**：禁用直接提交，被拒时展示 eiam 拒绝消息（tech-design Data Models 预置降级路径）；拒绝消息是否携带活跃会话数随处置核验 | eiam 无会话计数端点（§1 全量导出佐证）；硬前置无数据源 | 证据-R + 证据-P（拒绝响应样本） | eiam owner：＿＿＿＿ |
| G-5 | 授权资源/动作词表枚举端点（UF-9「选项来自 eiam 词表，不支持自由文本」） | UF-9 | **UI 降级**：静态词表（来源版本登记于本表，改动随控制台发版）；备选 eiam 补端点 | PRD UF-9 已预置降级路径；eiam `manifest` 是权限资产（action 类）而非资源词表，语义不匹配 | 证据-R + 证据-S（vo.go:23-26） | 业务 owner：＿＿＿＿ |
| G-6 | 授权三元组（subject/resource/action）新增/撤销端点 | UF-9 | **UI 语义映射**（须业务 owner 书面确认）：「授权」页改为对 eiam 原生模型的呈现——主体→策略（policy batch-attach/detach）与 主体→角色（role batch_assign/batch_unassign）两类绑定关系；若业务不认可该语义映射，转 eiam 排期补齐 | eiam 授权模型为 subject→target(role\|policy) 二元绑定（D-4），三元组为 PRD 侧推定；映射后「新增授权=绑定策略」操作完全落在实有端点上 | 证据-R + 证据-S（vo.go:78-101） | 业务 owner：＿＿＿＿ |
| G-7（任务 1.3 新增） | `POST /api/v1/discovery/sync` 微服务资产上报写面无鉴权（§1.1 #14 预告项闭环，核查详情 §6.2.3/§6.2.5） | 服务端授权面完整性（PRD Security Requirements / tech-design Security） | **eiam 排期补齐**：上报面加服务间凭证/来源绑定，或限内网 + 网络 ACL；未补齐前须确保 eiam :9000 不暴露于不可信网络 | 任意可达调用方可提交 SyncRequest 全量重写指定 service 的权限点/API 资产/code 绑定（ingestion 全清重建）——清空 code 绑定可使对应 API 降级「仅登录放行」，篡改绑定致授权面漂移；现状仅 service 非空校验。dev nginx 未转发 `/api/v1/`，控制台同源面不可达，风险面为 :9000 直连 | 证据-R（discovery/handler.go:21-26,33-50；ingestion/engine.go:54-105；permission.go:134-152） | eiam owner：＿＿＿＿ |
| G-8（任务 1.4 新增） | 首登强制改密承载机制（`must_change_password` 字段 + 登录链路强制拦截 + 改密后清除） | UF-4（初始密码发放后首登强制改密）/ E1 | **eiam 排期补齐**（建议与 G-1 重置端点一并排期：创建与重置两条初始密码发放路径都依赖该机制）；备选 = 业务 owner 书面确认 v1 放弃强制改密（仅线下提醒，安全弱化须明示） | 全仓 grep `must_change|MustChange|force_password` 0 命中（internal/**）；SignupRequest 仅 `password/confirm_password`（vo.go:20-24）；UpdateUserReq 无 password（vo.go:124-127）；`password/update` 为 L1 本人自服务需 old_password（user.go:87,416-428，§1.2 #16）→ 管理员发放初始密码后无任何「首登必须改密」的服务端标记与拦截 | 证据-R（§1 全量导出反证）+ 证据-S（vo.go:20-24,124-127；user.go:416-428） | eiam owner：＿＿＿＿ |
| G-9（任务 1.4 新增） | 用户锁定态列表可见性（UF-4 状态 enum 含「锁定」，eiam 无对应数据源） | UF-4（状态列）/ E2 | **UI 降级**：控制台用户列表状态列 = 启用/禁用/未知，锁定不展示——锁定为登录期临时态，提示由登录页错误文案承接（§3.2 B-2 口径）；若业务要求列表可见锁定态，转 eiam 排期补查询端点/字段 | 锁定态为 redis 计数（cache/user.go:68-89，lockoutKey exists + 失败计数 24h TTL；login.go:38,52,58 判定/累计/清除），不上 user 模型与 VO（domain/user.go 无 lock 字段，grep 0 命中），§1 全量导出无锁定查询路由 | 证据-R + 证据-S（login.go:37-58；cache/user.go:68-89；repository/user.go:497-503） | eiam owner：＿＿＿＿ |

#### 3.1.1 缺口处置决策记录（Phase 0 出口门 feature owner 定案，2026-09-30）

> 决策口径：默认处置 = **UI 降级/裁剪**，逐项按 tech-design 预置降级路径承接；**G-7 例外 = 需 eiam owner 排期确认（安全）**。本表为决策记录（决策来源 = Phase 0 出口门 owner 输入，1.gate 落档）；下表「会签位」与 §3.1「决策签收」列、§7.2 会签表一律**待签**（会签留位不代签）。承接落点涉及验收场景执行口径的，由对应里程碑 E2E 宿主任务（2.12/3.4/4.6/5.5）按本表落，acceptance.md §5 依赖表已同步。

| 编号 | 决策处置 | 承接口径（控制台侧） | 对应预置降级路径 |
|------|----------|----------------------|------------------|
| G-1 | UI 降级/裁剪 | v1 裁剪「管理员重置密码」流（无端点承载）；初始密码发放保留于 `user/create` 路径（A 档基线 #2，§2.4 已核成立）；E1 重置侧场景按裁剪/降级口径执行（2.12 落验收形态） | tech-design C 档 E1 承载缺口的降级默认 |
| G-2 | UI 降级/裁剪 | 按 §3.1 原备选降级执行：v1 仅展示 passkey 绑定态（`user/identity/list`，L1 基线外 L1 #24），绑定动作引导用户走 eiam 既有本人自助注册流程；不前置 eiam 排期 | §3.1 G-2 行备选降级条款 |
| G-3 | UI 降级/裁剪 | 维持替代路径核验：`user/identity/unbind`（provider=passkey）作撤销承载（证据-P 待补）；核验不过并入 G-2 降级口径，不单独立 eiam 排期前置 | §3.1 G-3 行处置条款 |
| G-4 | UI 降级 | 禁用直接提交，被拒时展示 eiam 拒绝消息；消息是否携带活跃会话数随证据-P 核验 | tech-design Data Models `Tenant` 注释预置降级 |
| G-5 | UI 降级 | 静态词表（来源版本登记于页面，改动随控制台发版） | PRD UF-9 预置降级路径 |
| G-6 | UI 语义映射 | 「授权」页按 eiam 原生二元绑定模型呈现（主体→策略 `policy/batch-attach|batch-detach` + 主体→角色 `role/batch_assign|batch_unassign`），D-4 路径归一承接 | D-4 语义映射 = 降级默认的路径归一形态 |
| G-7 | **需 eiam owner 排期确认（安全）** | 不适用 UI 降级——服务端授权面完整性问题，控制台侧无降级可承接；未补齐前部署面须确保 eiam :9000 不暴露于不可信网络（§6.2.5） | 安全项例外（本决策唯一非降级项） |
| G-8 | UI 降级（E1 降级） | v1 不做首登强制改密拦截（eiam 无载体，§6.3.5 核查结论）；初始密码一次性展示照常（create 承载），强制拦截与改密标记不做；E1 场景验收口径随 2.12 E2E 宿主落 | tech-design E1 预置降级（owner 例示） |
| G-9 | UI 降级 | 用户列表状态列 = 启用/禁用/未知，锁定不展示；锁定提示由登录页错误文案承接（中性固定文案，不做前端倒计时自算） | tech-design B 档 B-2 预置降级（owner 例示） |

### 3.2 形状分歧（端点实存，形状与 tech-design/UF 推定不一致——前端 api 层归一承接，非缺口）

| 编号 | 分歧 | tech-design/UF 推定 | eiam 实核 | 承接方式 | 证据 |
|------|------|---------------------|-----------|----------|------|
| B-1 | MFA verify 请求字段 | `{mfa_ticket, code}` | `{mfa_token, code}` | 控制台 api 层按实核字段 | eiam/internal/web/user/vo.go:251-254 |
| B-2 | LoginResult 无 `token`/`lockUntil`；锁定信息为错误文案 | `{mfaRequired, mfaTicket, token, lockUntil}` | `{mfa_required, mfa_token, must_select_tenant, must_bind, bind_token}`；session 经 Set-Cookie 颁发；锁定=error 文案（ErrUserLocked / ErrMfaAttemptsExhausted + 剩余次数文案） | 锁定倒计时降级中性固定文案（tech-design B 档预置）；token 语义改判 = cookie 会话 | eiam/internal/domain/user.go:143-155；service/user/login.go:37-52；service/user/mfa.go:83-88；errs/error.go:22 |
| B-3 | LDAP 登录请求无 `source_id` | `{username, password, source_id?}` | `{username, password}`（单 ldap 通道） | 按 `source_id` 缺省实现；多身份源选择语义归 1.2 核查 | eiam/internal/web/user/vo.go:27-30；user.go:194-196 |
| B-4 | passkey 流程会话标识 | body 字段 `{options, ticket}` / finish `{ticket, assertion}` | `{options, session_token}`；finish 凭据走 WebAuthn body 解析、会话标识走 header（登录 finish 同族） | 控制台按 `session_token` + header 实现；E2E 覆盖 | eiam/internal/web/user/vo.go:102-110；user.go:707 |
| D-1 | 组织=department（路径族整体不同） | `organization/list\|create\|update\|delete` | `department/*`（GET list 树形；字段 parent_id/sort/leaders/main_leader） | 控制台「组织管理」页直调 department 端点，类型层映射 | eiam/internal/web/department/handler.go:26-51；vo.go:4-9,37-41 |
| D-2 | 身份源载荷族 `{name,type,enabled,ldap{...}}`、upsert save、无 port/timeoutSec 独立字段 | 扁平 `ConnParams{url,port,bindDn,password,baseDn,attrMap,timeoutSec}` + 独立 create/update | 嵌套 `ldap` 对象；save 语义；属性映射字段名 `username_attribute` 等 | UF-7 表单 port/超时 → 前端仅入 url 组装与本地校验，或并入 G 类缺口处置（随 eiam owner 意见定） | eiam/internal/web/identity_source/vo.go:6-36 |
| D-3 | 角色↔策略绑定走 policy 侧 | `role/policy/batch_assign\|batch_unassign`、`role/attached/policies` | `policy/batch-attach\|batch-detach\|detach`、`policy/list/attached/role`、`user/list/attached/role` | 控制台角色详情页按实核路径调用 | eiam/internal/web/policy/handler.go:62,65,69,88；vo.go:57-63,93-98,140-144 |
| D-4 | 授权模型为 subject→target 二元绑定 | subject/resource/action 三元组 | `Authorization{subject,target,sub_type,obj_type}` | 见 G-5/G-6 处置 | eiam/internal/web/permission/vo.go:78-101 |
| D-5 | 策略容器/成员键单数 + access_scope | `statements[]`、`actions[]/resources[]` | `statement[]`、`action[]/resource[]/condition/access_scope`；`Effect ∈ {"Allow","Deny"}` ✓ | 控制台类型层按实核键名归一为规范形 `PolicyStatement[]` | eiam/internal/web/policy/vo.go:10,31-37；domain/policy.go:27-31 |

### 3.3 待核验假设清单（B/C 档，升格条件）

> 升格为契约的条件：证据-P（登录态响应样本抓包）核验通过 + eiam owner 会签（§7）。样本抓包执行归 1.2（认证链路）/1.3（服务端授权）任务窗口。

B 档 5 端点（§2.1）、C 档路径实存端点：`user/detail`、`user/list/attached/role`、`permission/menus`、`tenant/create|update|delete|detail|list/mine`、`identity_source/list|save|test|toggle|delete|detail`、`role/create|update|delete|detail`、`policy/list|create|update|delete|detail`、`permission/authorizations`、`permission/subjects/search`、`department/list|create|update|delete|detail`。

---

## 4. 契约基线（冻结快照）

> **冻结范围**：控制台 v1 依赖的端点快照（路径/方法/鉴权层/关键字段）。A 档 13 条为已实核契约；「待核验假设」端点在证据-P 核验通过前**不入基线**，升格后追加登记（§4.2）。eiam 侧变更基线内任一端点（含字段增删、语义调整）**须提前 ≥5 个工作日书面通知并附影响评估，未经评估的变更不得发布**；发布后发现的未通知变更按缺口处置登记并触发受影响场景回归（PRD Data Requirements）。

### 4.1 基线端点（A 档，实核吻合）

| # | 方法 | 路径（外部） | 鉴权层 | 关键请求字段 | 关键响应字段 | 证据 |
|---|------|--------------|--------|--------------|--------------|------|
| 1 | POST | /api/iam/user/list | L2（iam:user view） | `{offset, limit, keyword}` | `{total, users}`（users=UserMemberVO） | e-cam-web eiam-users.ts:17；eiam user.go:110 |
| 2 | POST | /api/iam/user/create | L2（iam:user add） | `SignupRequest`（含 `password/confirm_password`） | `number`（新 id） | e-cam-web eiam-users.ts:33；eiam vo.go:20-24 |
| 3 | POST | /api/iam/user/update | L2（iam:user edit） | `{id, username, email, nickname, avatar, job_title, phone, status}`（无 password） | `null` | e-cam-web eiam-users.ts:38；eiam vo.go:124-127 |
| 4 | DELETE | /api/iam/user/delete/:id | L2（iam:user delete） | — | `null` | e-cam-web eiam-users.ts:42；eiam user.go:120 |
| 5 | POST | /api/iam/role/list | L2（iam:role view） | `{offset, limit, keyword}` | `{total, roles}`（role 字段 `id/code/name/desc`） | e-cam-web eiam-users.ts:53；eiam role/vo.go:54-63 |
| 6 | POST | /api/iam/role/list/attached/user | L2（iam:role view_user_roles） | `{user_id, offset, limit}` | `{total, roles}` | e-cam-web eiam-users.ts:70；eiam role/handler.go:98 |
| 7 | POST | /api/iam/role/batch_assign | L2（iam:role batch_assign） | `{usernames[], role_codes[]}` | `null` | e-cam-web eiam-users.ts:84；eiam role/handler.go:66 |
| 8 | POST | /api/iam/role/batch_unassign | L2（iam:role batch_unassign） | `{usernames[], role_codes[]}` | `null` | e-cam-web eiam-users.ts:90；eiam role/handler.go:71 |
| 9 | POST | /api/iam/tenant/list | L2（iam:tenant view） | `{offset, limit, keyword}` | `{total, tenants}`（TenantVO 含 `status/ctime`） | e-cam-web eiam-users.ts:102；eiam tenant/vo.go:43-59 |
| 10 | POST | /api/iam/tenant/list/attached/user | L2（WithTenantOverride） | `{user_id, offset, limit}` | `{total, tenants}` | e-cam-web eiam-users.ts:113；eiam tenant/handler.go:110 |
| 11 | POST | /api/iam/tenant/batch_assign | L2（iam:tenant batch_assign） | `{user_ids[], tenant_ids[]}` | `null` | e-cam-web eiam-users.ts:125；eiam tenant/handler.go:100 |
| 12 | POST | /api/iam/tenant/batch_unassign | L2（iam:tenant batch_unassign） | `{user_ids[], tenant_ids[]}` | `null` | e-cam-web eiam-users.ts:131；eiam tenant/handler.go:104 |
| 13 | POST | /api/iam/tenant/switch | L1（WTS） | header `X-Active-Tenant-ID: <id>` | `null`；**调用方必须整页 reload**（重签 ✓ live 实核、Set-Cookie 下发，见 §6.1.6；「销毁」无 JWT 吊销语义 F-3） | e-cam-web eiam-tenant.ts:13-15；eiam tenant/handler.go:51；middleware/tenant.go:15 |
| 14 | GET | /api/iam/user/profile | L1 | — | `RetrieveUser{user, tenants, current_tenant_id, must_select_tenant, is_admin, permissions, must_bind, bind_token, mfa_required, mfa_token}` | e-cam-web stores/user.ts:45；eiam user.go:346-378；vo.go:134-145 |
| 15 | POST | /api/iam/user/logout | L1 | — | `null`（销毁共享 session） | e-cam-web stores/user.ts:83-85；eiam user.go:89 |

### 4.2 基线追加登记（待核验端点升格后填入）

| 方法 | 路径 | 鉴权层 | 关键字段 | 升格证据（证据-P + 会签） | 登记日期 |
|------|------|--------|----------|---------------------------|----------|
| GET | /api/iam/permission/manifest | L2（Needs iam:permission:manifest + menus_by_urns；super 豁免） | `{actions:[PermissionVO], services:[...]}`；成功 `{code:0,data}` | REAL 浏览器实测 200、132 动作 + 2 services（verify-6 脚本「REAL manifest 词表 200」+「词表已加载」断言；探测记录 admin profile 132 码含 manifest 无 menus_by_urns） | 2026-10-09 |

### 4.3 基线冻结签收

| 项 | 签收人 | 日期 |
|----|--------|------|
| §4.1 基线内容（路径/方法/鉴权/关键字段）核实 | 平台管理员（盘点执行人）：＿＿＿＿ | ＿＿＿＿ |
| 基线冻结 + ≥5 工作日变更通知条款确认 | eiam owner：＿＿＿＿ | ＿＿＿＿ |
| A 档证据侧复核 | e-cam-web 前端 owner：＿＿＿＿ | ＿＿＿＿ |

---

## 5. 「状态 × code → kind」错误分类对照表 + 权限码词表（P0 定稿）

### 5.0 错误分类对照表（placeholder，承接 1.2/1.3 证据-P）

> tech-design Error Handling 规定 Phase 0 产出该对照表入本文件。依赖真实错误响应样本（HTTP 状态 × 信封 code），随 1.2/1.3 的证据-P 抓包一并产出，此处保留章节位。

| HTTP 状态 | 信封 code | kind | 触发端点/场景 | 证据 |
|-----------|-----------|------|----------------|------|
| （待 1.2/1.3 补充） | | | | |

### 5.1 权限码词表（P0 定稿，2026-10-16）

> 单一事实源：代码 `haven-console/src/config/permission-codes.ts` 与本表互为投影（同口径维护，
> BIZ-auth-009）。本节定稿后供控制台三级裁剪消费：菜单（menuConfig）/ 路由守卫
> （meta.permissions）/ 按钮门控（useCan）。此前 router/menuConfig 中的「词表待 parity §5 定稿」
> 扩展点注释在本节定稿时全部落地收口。

**构造规则**（证据-S：`eiam/pkg/web/capability/api.go` normalizeCode:189-201 + `eiam/ioc` 装配）：

- service 恒为 `iam`：`capability.NewRegistry("iam", <module>, ...)`（`eiam/internal/web/{user,role,policy,tenant,department,permission,group,invitation,identity_source}/*.go`）；
- verb 含 `:` → 全码 = `iam:<verb>`；否则 = `iam:<module>:<verb>`；
- 例：`Capability("创建用户","add")` → `iam:user:add`；`Capability("授权治理列表","view_authorizations")` → `iam:permission:view_authorizations`。

**词表全量摘要**（完整清单见 permission-codes.ts 的 IAM_* 常量组）：

| 模块 | 代表码 | Scope（eiam 注册） |
|------|--------|---------------------|
| user | `iam:user:view/get/add/edit/delete/batch_delete/ldap_search/ldap_sync/manage_identity/unbind_identity/view_role_members` | tenant |
| department | `iam:department:view/get/add/edit/delete/assign/remove/members` | tenant |
| tenant | `iam:tenant:view/view_mine/get/add/edit/delete/view_members/assign/unassign/batch_assign/batch_unassign/view_user_tenants/switch` | tenant（控制台页面级另挂 requiresAdmin 双闸） |
| role | `iam:role:view/get/add/edit/delete/batch_delete/batch_assign/batch_unassign/unassign/add_parent/remove_parent/view_parents/view_mine/view_user_roles` | tenant |
| policy | `iam:policy:view/get/add/edit/delete/detach/batch_attach/batch_detach/view_user_policies/view_role_policies/view_group_policies`（attach 端点已注释停用，不入词表） | tenant（控制台页面级另挂 requiresAdmin 双闸） |
| identity_source | `iam:identity_source:view/detail/save/delete/test/toggle` | **system**（`handler.go:20` DefaultScope(ScopeSystem)） |
| permission | `iam:permission:manifest/view_authorizations/search_subjects/menus_by_urns` | tenant |
| group | `iam:group:view/get/add/edit/delete/members/assign_members/remove_members/assign_role/remove_role/roles` | tenant（P1 上页，端点现成） |
| invitation | `iam:invitation:view/add/delete/batch_delete/view_requests/handle_request` | tenant（P1 上页，端点现成） |

**三级裁剪语义（P0 落地口径，与原型 18/18 DOM 断言一致）**：

1. **菜单**：`menuConfig.visibleMenuItems({codes, isAdmin, restricted})` —— 未持 code 隐藏；
   adminOnly（租户/策略管理）双保险（权限码 + isAdmin 两道闸）；restricted（!isAdmin 且零租户）
   仅渲染无码项（工作台）。**tenant_admin 档自此由码自然得出**（原 inferRole(isAdmin/tenantsCount)
   三分支无法区分租户管理员与只读审计的缺陷已修复）；
2. **路由**：9 条路由 `meta.permissions`，守卫 `required.every(hasPermission)` 不通过 →
   `/forbidden` **原地**（BIZ-auth-003：403 本地，不跳登录）。/tenants、/policies 叠加
   requiresAdmin（Hard Rule 双闸）；
3. **按钮**：`useCan()` 派生 `can/canEvery/canSome`，页内 v-if 门控（授权管理页按 G-6 语义映射
   组合码：新增 = role:batch_assign|policy:batch_attach 任一；行级撤销按 action=assign→
   role:batch_unassign / attach→policy:batch_detach）。

**边界声明（BIZ-auth-001）**：前端三层裁剪仅体验层，可被直调接口绕过；安全边界唯一 =
eiam 服务端 OPA 强制校验（§6.2 89 条矩阵）。权限码不得作为任何安全判定凭据。

**维护口径（BIZ-auth-009）**：eiam 侧新增/变更任一端点能力码（含字段语义调整）须提前
≥5 个工作日通知，并同步更新 `permission-codes.ts` 与本表。

**本轮不消费项**：`POST /api/iam/permission/menus/by_urns`（eiam 服务端菜单树）——
P0 走本地词表；待 eiam 菜单树定义齐备后（P1）再评估切换为服务端驱动。

---

## 6. 后续任务补充章节（占位）

| 章节 | 归属任务 | 内容 |
|------|----------|------|
| 6.1 认证链路兼容矩阵（cookie 名/属性、session 结构、tenant_id 写入断言、tenant/switch 销毁重签复核、多身份源登录选择语义） | 1.2 | 已补充（2026-09-29，见下方 §6.1） |
| 6.2 服务端强制授权核查矩阵（§1 L2 各端点 OPA 校验逐项确认 + L0/L1 暴露面复核，含 discovery/sync 无鉴权上报面） | 1.3 | 已补充（2026-09-29，见下方 §6.2） |
| 6.3 `platform/users` 能力映射表（功能点 × 实现方式 × eiam 端点 × UF 落点三选一；含 group 模块「有意不迁移」确认） | 1.4 | 已补充（2026-09-29，见下方 §6.3） |
| 6.4 409 并发编辑（E6）支持性核查结论 | 1.2/1.3 | 待补充 |

### 6.1 认证链路兼容矩阵（任务 1.2 产出，2026-09-29）

> 证据源：① **证据-R/S**——eiam / ginx（v0.0.3-0.20250724125208）/ e-cam-web / e-cam-service 源码逐行，文件:行号随行标注；② **证据-P**——本地 dev 实例（localhost:8888 nginx → eiam :9000，非生产环境）live 探测（2026-09-29）：登录 / profile / tenant/switch / 登出五链路逐条跑通，探测用临时账号探测后已删除，JWT 与 token 值一律脱敏；③ **证据-O**——eiam owner 会签栏位见 §6.1.9。本节不写入任何凭据、密钥与环境敏感信息。

#### 6.1.1 Cookie 属性矩阵

| 属性 | 实测值（live Set-Cookie） | 代码证据 | 与 e-cam-web 一致性判定 |
|------|---------------------------|----------|--------------------------|
| name | `ecmdb-token-key` | live；eiam config.yaml:52 + ioc/session.go:50-52,67 | **严格一致**：e-cam-web 硬编码同名读取（cookie.ts:6） |
| domain | 配置项 `session.cookie.domain`（dev 取值 localhost；上线须改为实际登录域） | live `Domain=localhost`；config.yaml:51；ioc/session.go:53-55,71 | 部署一致性项：e-cam-web 无 domain 逻辑，同登录域即共享（tech-design 会话拓扑前提） |
| path | `/`（硬编码） | live `Path=/`；ioc/session.go:68 | **严格一致**：cookie.ts:15-19 按 `path=/`（及遗留 `/cam`）清除 |
| Max-Age | 2592000 秒（30 天，签发起算固定） | live `Max-Age=2592000`；ioc/session.go:57,66 | 前端不感知 |
| Secure | 配置项（dev=false → 属性缺失；https 上线必须置 true，否则浏览器拒存 cookie） | live 无 Secure 属性；config.yaml:53-54；ioc/session.go:61-64 | 前端不感知；部署检查项 |
| httpOnly | false（硬编码，属性缺失） | live 无 HttpOnly 属性；ioc/session.go:70 | **严格一致（硬前提）**：前端必须 JS 可读——`document.cookie` 读取后注入 Bearer（cookie.ts:9-12、request/eiam.ts:18-24） |
| SameSite | **未设置（Set-Cookie 无 SameSite 属性）** | live；ginx cookie/carrier.go:24-31 结构体无 SameSite 字段、Inject 走 gin SetCookie（gin context.go:1109-1121，sameSite 默认 0 = 属性省略）；eiam 全仓无 SetSameSite 调用 | **与 tech-design Security 前提不一致**（该章 CSRF 论证假设 Lax/Strict）→ 登记 F-4，须 owner 会签 |
| 清除语义 | `Set-Cookie: ecmdb-token-key=; Path=/; Domain=…; Max-Age=0`（登出响应下发，浏览器全局生效） | live logout 响应；ginx cookie/carrier.go:33-36 | e-cam-web 登出另做客户端补删（stores/user.ts:90 → cookie.ts:15-19），双保险 |

#### 6.1.2 Token 携带双通道

- 签发/重签/续期时 eiam 同时下发：`X-Access-Token` 响应头 + 上述 Set-Cookie（live 两处同 token；ginx header/carrier.go:48-52、mixin/carrier.go:19-25、eiam ioc/session.go:73-74）。
- 请求侧提取顺序：`Authorization: Bearer <token>` 头优先，缺失回落 cookie（mixin/carrier.go:27-34，Extract 顺序 header→cookie）。
- e-cam-web 请求层仅注入 Bearer 头（request/eiam.ts:18-24），自身从不写 session cookie（仅读/删）；nginx 显式透传 Authorization / Cookie / X-Active-Tenant-ID 并放大 buffer 承载大 JWT Set-Cookie（nginx.dev.conf:9-17；§0.1 所引 8-9 行有轻微行漂移，内容一致）。
- e-cam-service 同款 mixin 装配（e-cam-service ioc/session.go:95-105），cam 请求经 Bearer 头读取同一 Redis session（ecmdb_auth.go:46-59）。

#### 6.1.3 Session / JWT 结构（逐字段：live 解码 + 代码）

JWT（HS256）payload 实测形状：`{"data":{"Uid":<int64>,"SSID":"<uuid>","Data":{…},"Expiration":<毫秒>},"exp":<秒>,"iat":<秒>}`

| 字段 | 类型/取值 | 写入时机 | 证据 |
|------|-----------|----------|------|
| `Uid`（uid） | int64 | 签发必写 | ginx redis/provider.go:88-99；live |
| `SSID` | uuid v4，即 redis sessData 键 | 签发生成；tenant/switch 重签后实测变更 | live |
| `Data.tenant_id` | int64 的字符串形（live `"4"`） | **始终写入** | sessionclaims/claims.go:51-54；live |
| `Data.username` | string | **始终写入** | claims.go:53；live |
| `Data.is_admin` | `"true"`（仅管理员角色） | tenantID≠0 且 permSvc 查询成功才写 | claims.go:85-88；live（另见 F-2） |
| `Data.authorized_codes` | JSON 数组字符串，仅 `cert:*` 前缀子集 | 同上 | claims.go:31,78-84 |
| `Expiration` / `exp` | 签发时刻 +30 天（毫秒/秒两形态，live exp−iat=2592000s） | 固定，不随滑动续期延长 | ioc/session.go:57-58；live |

- redis sessData 另存 `uid / username / tenant_id`（user.go:407-410；tenant/handler.go:246-249），TTL 30 天仅在签发时设置（ginx redis/session.go:35-44）。
- 续期语义：`CheckLoginMiddleware` 在剩余有效期 <30 分钟时触发 `RenewAccessToken` 重签**同 claims** token（ginx global.go:51-53、redis/provider.go:63-73）——exp 不延长；该中间件仅 eiam 挂载，e-cam-service 无续期逻辑。
- **F-2（live 发现：claims 登录期降级窗口）**：登录签发时 permSvc 查询失败会静默降级为仅 `tenant_id+username`（claims.go:63-76 的降级路径）——live 实测：登录 JWT 不含 `is_admin/authorized_codes`，而数秒后 profile 实时查询返回 is_admin=true + 全量权限码；同一会话 tenant/switch 重签后的 JWT 即含 `is_admin:"true"`。影响：登录后首批依赖 claims 的下游判定（如 e-cam-service 证书域 role_middleware）可能按只读降级；控制台自身不受影响（权限以 profile 为准）。

#### 6.1.4 GET /api/iam/user/profile 响应逐字段（live）

实测 data 顶层键：`user / tenants / current_tenant_id / must_select_tenant / must_bind / mfa_required / is_admin / permissions`（`bind_token`/`mfa_token` 带 omitempty 本例缺席；VO 全集 vo.go:134-145）

| 字段 | live 值/形状 | 语义证据 |
|------|--------------|----------|
| user | 15 字段：id/username/email/nickname/avatar/job_title/phone/status/source/ctime/utime/last_login_at/mfa_type/mfa_bound/identities | vo.go:40-57 |
| tenants | `[{id, name, code, domain}]`（个人空间命名规则 live 佐证） | vo.go |
| current_tenant_id | live `4`；**读自 redis sessData**（user.go:374），非 JWT claims | user.go:374 |
| is_admin | live `true`（新建账号即 true——权限模型面归 1.3 复核） | user.go:383 |
| permissions | 全量能力码数组（live 100+ 项，cmdb:/iam: 前缀） | user.go:376 |
| must_select_tenant | live `false` → **F-1：恒 false** | Profile handler 未赋值（user.go:378-386，零值序列化）；VO 字段存在（vo.go:138），仅登录响应赋值（user.go:256；login.go:146 `len(tenants)>1`） |
| must_bind / mfa_required | live `false`（handler 同样未赋值，恒 false） | 同上 |

**F-1（形状修正，涉及 tech-design 路由守卫设计）**：profile 的 `must_select_tenant` / `must_bind` 恒 false，多租户分流判定只能取**登录响应**（handleLoginResult）字段或前端自算 `tenants.length > 1`。控制台 2.9 登录实现与路由守卫「mustSelectTenant → /tenant-select」不得依赖 profile 该字段；e-cam-web 现行即以 `tenants.length>1` 控制租户选择器呈现（TenantSelector.vue:3），佐证该语义。§2.2 的字段**存在性**登记不受影响，**语义**以本条为准。

#### 6.1.5 tenant_id 写入断言（AC-3）

- **结论：始终写入**。JWT `Data.tenant_id` 无条件写（claims.go:51-54），redis sessData 同步写（user.go:407-410）——不存在「条件写入」分支。
- **零租户用户取值：正常链路取不到 0**。登录链路 `getOrInitTenants` 对零租户用户自动 `InitPersonalTenant`（「注册即拥有」，login.go:159-172；tenant.go:23-24）——live 实测新建账号登录即得 tenant_id=4（个人空间）；多租户用��� `selectActiveTenant` 优先系统租户 1、否则首个（login.go:174-181）。
- `tenant_id=0` 的「临时凭证」语义仅存于代码注释与构建分支（claims.go:48-49、user.go:401、ecmdb_auth.go:125-126），**当前四条登录路径（local / LDAP / OIDC / passkey，含 MFA verify，均经 postLogin）不会签出 0**；0 值可见形态只剩「会话数据被销毁后的僵尸 JWT」读取（§6.1.6 F-3 中间态）。
- 下游一致性：e-cam-service 从 **JWT claims** 解析 tenant_id（字符串→int64），0 由 RequireTenant 拒绝（ecmdb_auth.go:71-82,124-134）；eiam Profile 从 **sessData** 读。两处在签发时同步写入，唯一失同步窗口即销毁后（claims 仍带旧值、sessData 读 0）。

#### 6.1.6 tenant/switch 链路复核（AC-4）

live 实测（`X-Active-Tenant-ID: 4`，同租户形态）+ 代码逐行：

1. 请求头 `X-Active-Tenant-ID: <id>`：常量与提取（middleware/tenant.go:15,168-179）；WTS 覆写执行租户（:146-163）；e-cam-web 调用侧同款（eiam-tenant.ts:12-16）。
2. 归属校验：`CheckUserTenantAccess`（tenant/handler.go:232-236）。
3. 销毁：`sess.Destroy` 删 redis ssid 键（tenant/handler.go:238-239；ginx redis/session.go:23-25）。
4. 重签：NewSessionBuilder 生成**全新 SSID + 新 JWT**，经 TokenCarrier.Inject 以 **Set-Cookie 下发**（tenant/handler.go:241-250；ginx redis/provider.go:59）——live 实核：响应携带新 Set-Cookie、新 SSID、`code:0`。**重签 cookie 到达方式 = 服务端 Set-Cookie，前端零写 cookie**（e-cam-web 无 cookie 写路径，仅读/删）。
5. 调用方语义：成功后必须整页 reload（TenantSelector.vue:47-50 `window.location.reload()`；tech-design Integration 6 同约定），reload 后路由守卫重拉 profile 获得新租户上下文。
6. 同租户 switch（header=当前租户）同样全量销毁重签（WTS 不覆写、handler 照常执行）——live 即此形态。

**F-3（安全语义复核，须 owner 会签）**：「销毁旧 session」≠ 旧 JWT 失效。认证 = JWT 签名+exp 纯校验，**无 SSID 吊销检查**（ginx internal/jwt/management.go:79-91 仅验签与 exp；redis/provider.go:86-106 Get 不查 redis 存活性）——live 实测：switch 后旧 token 请求 profile 仍 200，且读 sessData 得 `current_tenant_id=0`（僵尸态：认证通过、会话数据为空）。同浏览器内的一致性靠 Set-Cookie 替换/清除成立（tech-design「另一侧下一个请求 401 收敛」在浏览器 cookie 层成立：登出下发 Max-Age=0 对全部标签页生效；switch 下发新 cookie 全局替换）；**跨设备/带外持有的旧 JWT 副本在 30 天 exp 前持续有效**，服务端单点登出与吊销语义不成立。tech-design Security「租户切换即销毁重签」的「销毁」应按此弱化语义理解。

#### 6.1.7 登出链路（live）

- eiam `POST /api/iam/user/logout`：无会话亦返回成功；有会话则 Clear cookie（live `Max-Age=0`）+ 删 redis sessData（user.go:389-398）。
- e-cam-web 登出：调 eiam logout → resetState → `removeEcmdbToken` 客户端补删 → 跳登录页（stores/user.ts:83-93）。
- 401 收敛不动共享 cookie：redirectToLogin 不清除 `ecmdb-token-key`（request/index.ts:15-27 注释即契约）。

#### 6.1.8 多身份源登录选择语义

- 登录方式清单由公开端点 `GET /api/identity_source/enabled_providers` 提供（live 返回 `["local"]`；identity_source/handler.go:26），控制台登录页按其渲染可用入口。
- 单通道语义：LDAP 登录请求无 `source_id`（B-3 已登记，user.go:194-196）；多身份源并行启用时的选择语义 eiam 侧无现成端点/载荷位——维持 §2.1 B-3 处置；v1 若仅启用单一身份源则无新增缺口。

#### 6.1.9 发现汇总与会签

| 编号 | 发现 | 影响 | 处置 |
|------|------|------|------|
| F-1 | profile 的 must_select_tenant / must_bind 恒 false | tech-design 路由守卫「mustSelectTenant → /tenant-select」及 §2.2 字段语义登记 | 控制台 2.9 登录实现改读登录响应字段或前端自算 `tenants.length>1`；语义以本节为准 |
| F-2 | 登录签发 JWT 的 claims 降级窗口（live 复现） | 依赖 claims 的下游首请求可能只读降级 | eiam owner 会签知悉；控制台不受影响（权限以 profile 为准） |
| F-3 | 「销毁」无 JWT 吊销语义，跨设备副本 exp 前有效 | 单点登出 / 租户切换的服务端强销毁预期 | eiam owner 会签接受，或提 eiam 排期补 SSID 吊销；相关章节措辞按弱化语义修订 |
| F-4 | session cookie 无 SameSite 属性 | tech-design Security 的 CSRF 前提（Lax/Strict）不成立（浏览器默认 Lax 行为兜底，但非 eiam 显式声明） | eiam owner 会签接受，或 eiam 侧补 SetSameSite；CSP/CSRF 论证随会签结论收敛 |

> 探测残留登记：探测用临时账号（uid=3）已删除；其自动初始化的个人空间租户在 dev 库残留一行（无可用删除权限码）。本节不含任何凭据、密钥或 token 值。

| 会签 | 签名 | 日期 |
|------|------|------|
| eiam owner（§6.1 全部结论 + F-1~F-4 处置） | ＿＿＿＿ | ＿＿＿＿ |

### 6.2 服务端强制授权核查矩阵（任务 1.3 产出，2026-09-29）

> 证据源：**证据-R/S**——eiam 源码逐行（授权中间件、capability 注册、资产绑定链、路由声明、边界校验器），文件:行号随行标注；本任务未做 live 探测，403 拦截路径为**代码确凿路径**（Hard Rules 口径），必要时可按 §6.1 同法补证据-P live 样本。核查口径 = PRD「服务端强制授权、前端菜单裁剪仅为体验层非安全边界、Phase 0 核查矩阵逐端点确认、缺失端点进入缺口处置」（prd-spec.md:204）与 tech-design Security 首条（tech-design.md:339）。

#### 6.2.0 授权机制链（统一口径，逐端点矩阵的判读基础）

**鉴权方式（L1/L2 全量统一）**：共享 Redis JWT session——cookie `ecmdb-token-key` / Bearer 双通道（§6.1.1/§6.1.2），经 `session.CheckLoginMiddleware()`（`ioc/web.go:50`）校验；未登录 → **401**。L0 端点无鉴权（设计即公开，逐条复核见 §6.2.3）。

**L2 授权判定链**（全局中间件 `middleware.CheckPermission(permSvc)` 注册于全部 Private 路由之前，`ioc/web.go:61`；实现 `internal/pkg/middleware/permission.go:20-74`）：

| 步骤 | 语义 | 证据 | 失败结果 |
|------|------|------|----------|
| 1 会话校验 | sess 为空即拦 | permission.go:23-27 | **401** |
| 2 能力元数据识别 | handler 指针 → `capability.GetResourceInfo` | permission.go:36-43；capability/api.go:343-361,364-369 | 未声明 capability → **仅登录放行**（default-allow 分支；§6.2.4 核查当前无触发实例） |
| 3 资产发现 | `FindAPIByPath(service, method, path)` | permission.go:130-137 | 资产未登记 → **fail-closed 拒绝**（:134-137） |
| 4 code 绑定查找 | `FindCodesByResource(URN)` | permission.go:141 | 无绑定 → **仅登录放行**（:146-152；当前 L2 无此状态，§6.2.4） |
| 5 边界拦截 | 普通租户（tid≠1）触 `Scope=system` 权限点 | permission.go:154-157；checker/checker.go:28-51（系统租户旁路 :29-31、拦截 :43-48）；errs/error.go:36 | ErrForbidden → **403 code=403001**（permission.go:50-56） |
| 6 OPA 策略判定 | 有效策略（直接绑定+角色继承+内联）对 URN×动作裁决，**Deny 优先** | permission.go:166-183；authz/rego/policy.rego:5,9-13,50-61,71-79 | 不允许 → **403 code=403001「无权执行该操作，请联系管理员授权」**（permission.go:64-70） |

**授权粒度词表**（AC「授权粒度」判读依据）：

| 粒度 | 判据 | 覆盖模块 |
|------|------|----------|
| **租户级** | capability 默认 `ScopeTenant`（capability/api.go:124）；casbin 角色分配按租户域隔离（规则 v2=tid：permission.go:703-733,744-782），策略判定按**身份租户**解析（getEffectivePolicies permission.go:377-383；GetRolesForUser :929-938；ctxutil/context.go:68-69,83-84） | user / role / policy / department（组织承接）/ group / invitation / permission（授权治理承接） |
| **平台级** | 模块级 `DefaultScope(ScopeSystem)`——tenant（tenant/handler.go:29-30）、identity_source（identity_source/handler.go:20-21）；非系统租户触即 403（边界拦截），**与是否被授予策略无关** | tenant CRUD（L2 13 条）、identity_source（6 条） |
| **越租户阻断** | `WithTenantOverride`：目标租户≠身份租户 且 身份租户≠系统租户(1) → 403 403001「检测到跨租户越权操作，该请求已被安全拦截」（pkg/web/middleware/tenant.go:103-130，判定 :106-121；SystemTenantID=1 ctxutil/context.go:22-25）；超管放行仅覆写执行租户、身份租户保留（:123-126）。注：目标租户取自 `X-Active-Tenant-ID`（:15,168-179，与代码注释 X-Tenant-ID 措辞漂移，以常量为准） | tenant/members、tenant/assign、tenant/unassign、tenant/list/attached/user（4 条，WTO 包装见 §1.3 tenant 表） |
| 数据级隔离 | tenancyBuilder 注入执行租户 tenant_id（tenant.go:44-87，tenant_id=操作数据 / origin_tenant_id=鉴权身份，:76-83 注记）供查询层使用；非系统租户 user/list 降级返回基础信息（user/user.go:443-448） | 全部 L2 |

**前端裁剪 ↔ 后端 403 兜底对应**（「裁剪非安全边界」两侧实证，AC3 判读依据）：

- 服务端裁剪源 ①菜单：`GET /api/iam/permission/menus` = GetAuthorizedMenus——OPA 批量裁决 + 边界预过滤（permission.go:264-302；filterByBoundary :305-326）；②按钮级权限码：profile.permissions = GetAuthorizedCodes（permission.go:229-261）。前端按两者渲染导航/显隐按钮——**仅为体验层**。
- 兜底：任何绕过裁剪的直达请求 → 判定链步骤 3-6 → **403 code=403001**（普通租户触系统级另经边界拦截 403）。菜单裁剪、按钮裁剪与 403 兜底为同一权限模型（manifest/capability code）的三个出口，口径一致。

#### 6.2.1 L2 管理端点逐条核查矩阵（89 条）

> 逐条核对 9 个 handler 的 `PrivateRoutes`（user/user.go:104-149、role/handler.go:37-103、policy/handler.go:40-97、tenant/handler.go:57-116、department/handler.go:26-51、group/handler.go:27-67、permission/handler.go:47-71、identity_source/handler.go:28-49、invitation/handler.go:42-64）：**89/89 均声明 Capability**，无 default-allow 实例。能力码 = `Capability("名称","code")` 第二参经 normalizeCode 加 `iam:<module>:` 前缀（capability/api.go:189-201）；`Needs(...)` 附加项与路由行号沿用 §1.3 各模块表（1.1 已实核），不重复。「403 路径」列取值：**标准** = CheckPermission 判定 false（permission.go:64-70）；**边界** = 系统级边界拦截（checker.go:43-48 → :50-56）；**WTO** = 跨租户拦截（tenant.go:114-121）。「结论」= 服务端鉴权（session/JWT，§6.2.0）+ 授权强制 + 403 路径确凿。

**user（12 条，`iam:user:*`，租户级）**

| 端点（eiam 内部路径） | 能力码 | 403 路径 | 结论 |
|----------------------|--------|----------|------|
| POST /api/user/create | add | 标准 | 强制 ✓ |
| POST /api/user/list | view | 标准 | 强制 ✓ |
| POST /api/user/update | edit | 标准 | 强制 ✓ |
| GET /api/user/detail | get | 标准 | 强制 ✓ |
| DELETE /api/user/delete/:id | delete | 标准 | 强制 ✓ |
| POST /api/user/batch_delete | batch_delete | 标准 | 强制 ✓ |
| POST /api/user/ldap/search | ldap_search | 标准 | 强制 ✓ |
| POST /api/user/ldap/sync | ldap_sync | 标准 | 强制 ✓ |
| POST /api/user/ldap/refresh_cache | ldap_refresh | 标准 | 强制 ✓ |
| POST /api/user/identity/manage | manage_identity | 标准 | 强制 ✓ |
| POST /api/user/identity/unbind | unbind_identity | 标准 | 强制 ✓（passkey 撤销承接候选 → G-3） |
| POST /api/user/list/attached/role | view_role_members | 标准 | 强制 ✓ |

**role（15 条，`iam:role:*`，租户级）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| POST /api/role/create | add | 标准 | 强制 ✓ |
| POST /api/role/update | edit | 标准 | 强制 ✓ |
| POST /api/role/list | view | 标准 | 强制 ✓ |
| GET /api/role/detail/:code | get | 标准 | 强制 ✓ |
| DELETE /api/role/delete/:id | delete | 标准 | 强制 ✓ |
| POST /api/role/batch_delete | batch_delete | 标准 | 强制 ✓ |
| POST /api/role/batch_assign | batch_assign | 标准 | 强制 ✓ |
| POST /api/role/batch_unassign | batch_unassign | 标准 | 强制 ✓ |
| POST /api/role/unassign | unassign | 标准 | 强制 ✓ |
| POST /api/role/analysis/inline | analysis | 标准 | 强制 ✓ |
| POST /api/role/add_parent | add_parent | 标准 | 强制 ✓ |
| POST /api/role/remove_parent | remove_parent | 标准 | 强制 ✓ |
| POST /api/role/parents | view_parents | 标准 | 强制 ✓ |
| GET /api/role/mine | view_mine | 标准 | 强制 ✓ |
| POST /api/role/list/attached/user | view_user_roles | 标准 | 强制 ✓ |

**policy（12 条，`iam:policy:*`，租户级；batch-attach/batch-detach 即 G-6 语义映射的「授权新增/撤销」落点）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| POST /api/policy/create | add | 标准 | 强制 ✓ |
| POST /api/policy/update | edit | 标准 | 强制 ✓ |
| POST /api/policy/list | view | 标准 | 强制 ✓ |
| GET /api/policy/detail/:code | get | 标准 | 强制 ✓ |
| POST /api/policy/detach | detach | 标准 | 强制 ✓ |
| POST /api/policy/batch-attach | batch_attach | 标准 | 强制 ✓（G-6「新增授权」候选） |
| POST /api/policy/batch-detach | batch_detach | 标准 | 强制 ✓（G-6「撤销授权」候选） |
| DELETE /api/policy/delete/:code | delete | 标准 | 强制 ✓ |
| POST /api/policy/batch-delete | batch_delete | 标准 | 强制 ✓ |
| POST /api/policy/list/attached/user | view_user_policies | 标准 | 强制 ✓ |
| POST /api/policy/list/attached/role | view_role_policies | 标准 | 强制 ✓ |
| POST /api/policy/list/attached/group | view_group_policies | 标准 | 强制 ✓ |

**tenant（13 条，`iam:tenant:*`，平台级 ScopeSystem）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| POST /api/tenant/create | add | 边界 + 标准 | 强制 ✓ |
| POST /api/tenant/list | view | 边界 + 标准 | 强制 ✓ |
| POST /api/tenant/list/by-ids | view_by_ids | 边界 + 标准 | 强制 ✓ |
| POST /api/tenant/update | edit | 边界 + 标准 | 强制 ✓（status 承载禁用，§2.6） |
| DELETE /api/tenant/delete/:id | delete | 边界 + 标准 | 强制 ✓ |
| POST /api/tenant/batch_delete | batch_delete | 边界 + 标准 | 强制 ✓ |
| GET /api/tenant/detail/:id | get | 边界 + 标准 | 强制 ✓ |
| POST /api/tenant/members | view_members | 边界 + WTO + 标准 | 强制 ✓（跨租户阻断） |
| POST /api/tenant/assign | assign | 边界 + WTO + 标准 | 强制 ✓（跨租户阻断） |
| POST /api/tenant/unassign | unassign | 边界 + WTO + 标准 | 强制 ✓（跨租户阻断） |
| POST /api/tenant/batch_assign | batch_assign | 边界 + 标准 | 强制 ✓ |
| POST /api/tenant/batch_unassign | batch_unassign | 边界 + 标准 | 强制 ✓ |
| POST /api/tenant/list/attached/user | view_user_tenants | 边界 + WTO + 标准 | 强制 ✓（跨租户阻断） |

**department（8 条，`iam:department:*`，租户级；「组织管理」语义承接 D-1）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| POST /api/department/create | add | 标准 | 强制 ✓ |
| POST /api/department/update | edit | 标准 | 强制 ✓ |
| DELETE /api/department/delete/:id | delete | 标准 | 强制 ✓ |
| GET /api/department/list | view | 标准 | 强制 ✓ |
| GET /api/department/detail/:id | get | 标准 | 强制 ✓ |
| POST /api/department/assign | assign | 标准 | 强制 ✓ |
| POST /api/department/remove | remove | 标准 | 强制 ✓ |
| POST /api/department/members | members | 标准 | 强制 ✓ |

**group（13 条，`iam:group:*`，租户级；控制台 v1 有意不迁移，§2.5 注）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| POST /api/group/create | add | 标准 | 强制 ✓ |
| POST /api/group/update | edit | 标准 | 强制 ✓ |
| DELETE /api/group/delete/:id | delete | 标准 | 强制 ✓ |
| POST /api/group/list | view | 标准 | 强制 ✓ |
| GET /api/group/detail/:code | get | 标准 | 强制 ✓ |
| POST /api/group/members/assign | assign_members | 标准 | 强制 ✓ |
| POST /api/group/members/remove | remove_members | 标准 | 强制 ✓ |
| POST /api/group/members | members | 标准 | 强制 ✓ |
| POST /api/group/list/attached/user | view | 标准 | 强制 ✓ |
| POST /api/group/list/attached/role | view | 标准 | 强制 ✓ |
| POST /api/group/role/assign | assign_role | 标准 | 强制 ✓ |
| POST /api/group/role/remove | remove_role | 标准 | 强制 ✓ |
| GET /api/group/roles/:code | roles | 标准 | 强制 ✓ |

**permission（4 条，`iam:permission:*`，租户级；authorizations 承接 UF-9 授权治理，G-5/G-6）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| GET /api/permission/manifest | manifest | 标准 | 强制 ✓（权限资产清单，非资源词表 → G-5） |
| POST /api/permission/authorizations | view_authorizations | 标准 | 强制 ✓（D-4 二元绑定模型） |
| POST /api/permission/subjects/search | search_subjects | 标准 | 强制 ✓ |
| POST /api/permission/menus/by_urns | menus_by_urns | 标准 | 强制 ✓（NoSync） |

**identity_source（6 条，`iam:identity_source:*`，平台级 ScopeSystem）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| POST /api/identity_source/save | save（upsert，D-2） | 边界 + 标准 | 强制 ✓ |
| POST /api/identity_source/list | view | 边界 + 标准 | 强制 ✓ |
| DELETE /api/identity_source/delete/:id | delete | 边界 + 标准 | 强制 ✓ |
| GET /api/identity_source/detail/:id | detail | 边界 + 标准 | 强制 ✓ |
| POST /api/identity_source/test | test | 边界 + 标准 | 强制 ✓ |
| POST /api/identity_source/toggle/:id | toggle | 边界 + 标准 | 强制 ✓ |

**invitation（6 条，`iam:invitation:*`，租户级）**

| 端点 | 能力码 | 403 路径 | 结论 |
|------|--------|----------|------|
| POST /api/invitation/create | add | 标准 | 强制 ✓ |
| POST /api/invitation/list | view | 标准 | 强制 ✓ |
| DELETE /api/invitation/revoke/:code | delete | 标准 | 强制 ✓ |
| POST /api/invitation/batch_revoke | batch_delete | 标准 | 强制 ✓ |
| POST /api/invitation/requests | view_requests | 标准 | 强制 ✓ |
| POST /api/invitation/requests/handle | handle_request | 标准 | 强制 ✓ |

> **AC1 特殊端点闭环**：`user/reset_password` 不存在（G-1，§3.1，无授权面可核查）；管理员发起 passkey 注册不存在（G-2，现役 register 为 L1 本人自服务，见 §6.2.2）；passkey 撤销承接候选 `user/identity/unbind` 为 L2 强制 ✓（G-3 待响应样本核验）。「授权 CRUD」三元组端点不存在（G-6，语义映射落点 policy/role 绑定端点均为 L2 强制 ✓）。

#### 6.2.2 L1 登录态端点暴露面复核（13 条，无细粒度授权——自服务绑定逐条核验）

> L1 = 仅 CheckLoginMiddleware + tenancyBuilder（ioc/web.go:50-58），**无 OPA 判定**；安全性取决于 handler 是否以会话身份为准、是否可操作他人数据。逐条结论：**全部绑定当前会话身份，无跨用户写面**。

| 端点 | 自服务绑定 | 复核结论 |
|------|-----------|----------|
| GET /api/permission/menus | 以会话 username 判定（handler.go:185 → permission.go:264） | 本人菜单树，服务端裁剪源 ① ✓ |
| POST /api/user/password/update | uid 取自 sess（user.go:421-425）+ old password 校验（:428） | 本人改密 ✓ |
| GET /api/user/profile | 本人（user.go:346-378） | ✓（§6.1.4 已核） |
| POST /api/user/logout | 本人会话销毁（user.go:380-398） | ✓（§6.1.7 已核） |
| POST /api/user/passkey/register/start/finish | uid = sess.Claims().Uid（user.go:668-674,707） | 本人自服务 ✓；**管理员代发起缺失 → G-2** |
| GET /api/user/mfa/totp/setup、POST bind、POST disable | uid = sess.Claims().Uid（user.go:838-860） | 本人 MFA 自服务 ✓ |
| GET /api/user/identity/list | 本人清单（user.go:101） | ✓（passkey 绑定态查询） |
| GET /api/tenant/list/mine | 本人所属（tenant/handler.go:44，NoSync） | ✓ |
| POST /api/tenant/switch | WTS + CheckUserTenantAccess 归属校验（tenant/handler.go:51,232-236；tenant.go:146-164） | ✓（链路复核 §6.1.6；销毁弱语义 F-3） |
| POST /api/invitation/accept | 登录态 + code 语义（invitation/handler.go:39） | ✓ |

> 授权视角登记：G-2（管理员代发起 passkey 注册缺管理侧端点）与 G-3（撤销承接候选 unbind 为 L2 强制 ✓，仅待核验 passkey 完整性）已在 §3.1 处置，本矩阵不重复处置。

#### 6.2.3 L0 公开端点暴露面复核（14 条，无鉴权——逐条判定公开合理性与暴露内容）

| 端点 | 暴露内容 | 复核结论 |
|------|----------|----------|
| POST /api/user/signup、/user/ldap/login、/user/system/login、/user/bind/confirm、GET /user/oidc/render、/user/oidc/callback、POST /user/passkey/login/start、/user/passkey/login/finish、POST /user/login/mfa/verify（user.go:65-80） | 认证流程入口（user.go:65-80） | **公开合理**：登录/注册/绑定即免登语义；流程令牌前置（bind_token §6.1.4、mfa_token §2.1、session_token §2.1） |
| POST /api/permission/check_login（handler.go:36；实现 :124-137） | 自行提取并校验 token（ctxWithAuth :165-176），仅回显 uid/tenant_id | **不构成绕过**：无 token → ErrUnauthenticated |
| POST /api/permission/check_policy（handler.go:37；实现 :140-165） | 同上身份提取，仅返回判定结果 | **不构成绕过**：判定主体=请求者本人 |
| GET /api/identity_source/enabled_providers（identity_source/handler.go:26；实现 :130-136） | 已启用登录方式类型枚举（live `["local"]`，§6.1.8） | **无敏感面**：不含身份源配置/名称/凭据 |
| GET /api/invitation/verify/:code（invitation/handler.go:34） | 按 code 校验邀请 | 公开合理：枚举面 = code 熵 |
| POST /api/v1/discovery/sync（discovery/handler.go:21-26；实现 :33-50） | **微服务资产上报写面，无任何鉴权**（仅校验 service 非空 :34-36） | **缺口 → G-7**（详见 §6.2.5） |

#### 6.2.4 缺省放行风险面核查（判定链三分支逐分支核查结论）

- **default-allow 分支**（私有路由未声明 capability → 仅登录放行，permission.go:38-42）：本任务对 9 个 handler 的 PrivateRoutes 89 条路由逐条核对，**全部声明 Capability → 当前 0 触发实例**。该分支属潜在脚手架风险（新增 L2 路由漏声明即降级），登记为 eiam 侧长期注记项。
- **fail-closed 兜底**（资产未登记 → 拒绝，permission.go:134-137）：新增/漏声明路由因不进资产同步（collectAPIs 仅取 handlerRegistry 内条目，capability/api.go:467-485）而天然 fail-closed → 403，缺省安全 ✓。
- **pass-through 分支**（资产已登记但无 code 绑定 → 仅登录放行，permission.go:146-152）：当前 L2 89 条均有 code 绑定——绑定链闭环：`h.Capability().Handle()` 注册（api.go:343-361）→ collectAPIs（:467-485）→ 快照绑定 code→URN（ingestion/snapshot.go:52-60）→ BatchBindResources（engine.go:87-93）→ FindCodesByResource（permission.go:141）。**人为制造该状态的途径即 G-7 无鉴权重同步**（清空绑定后对应 API 降级）。
- **L1 自服务绑定**：13 条全部以 sess 身份为准、无可指定他人的请求参数（§6.2.2 逐条证据）✓。
- **§6.1.4 is_admin=true 疑问闭环**：新账号个人空间自动建租户并**自授 admin 角色**（tenant/tenant.go:115-141，casbin 域=个人空间 tid :133-138）——is_admin=true 为**个人空间内**管理员，非平台级；平台级能力经 ScopeSystem 边界（checker.go:43-48）与租户域 casbin 双重隔离，不构成越权。

#### 6.2.5 发现汇总与会签

| 编号 | 发现 | 影响 | 处置 |
|------|------|------|------|
| G-7 | `POST /api/v1/discovery/sync` 无鉴权资产上报写面（§1.1 #14 预告项闭环） | 任意可达调用方可提交 SyncRequest **全量重写**指定 service 的权限点 + API 资产 + code 绑定（ingestion/engine.go:58-104 全清重建）；后果 ① 清空 code 绑定 → 对应 API 降级「仅登录放行」（permission.go:146-152），② 篡改绑定/权限资产 → 授权面漂移。dev nginx 未转发 `/api/v1/`（nginx.dev.conf:8-65 无匹配 location）→ 控制台同源面不可达；风险面 = eiam :9000 直连网络面（部署边界决定） | **eiam 排期补齐**（上报面加服务间凭证/来源绑定，或内网隔离+网络 ACL）；未补齐前须确保 :9000 不暴露于不可信网络。已登记 §3.1，签收位随 §3.1 |
| 注记 | default-allow 分支（permission.go:38-42）为脚手架级潜在风险，当前 0 实例（§6.2.4） | eiam 长期注记：新增 L2 路由必须 Capability 声明（漏声明时 fail-closed 兜底为 403，资产先行同步后漏绑定才降级） | 随 G-7 一并知悉，不单列处置 |
| 闭环 | §6.1.4「新建账号即 is_admin=true」疑问：为个人空间内 admin（tenant.go:132-138 自授，casbin 域隔离），非平台级越权 | §6.1.4 该括注结论收敛 | 无需处置 |

> 本矩阵结论（§6.2.0~§6.2.4）+ G-7 处置建议，仅登记不修复（Hard Rules）；生效以 eiam owner 会签为准。

| 会签 | 签名 | 日期 |
|------|------|------|
| eiam owner（§6.2 全部结论 + G-7 处置 + default-allow 注记） | ＿＿＿＿ | ＿＿＿＿ |

### 6.3 platform/users 能力映射（任务 1.4 产出，2026-09-29）

> **定位**：RC#1（e-cam-web 移出 `platform/users`）迁移前置门的双签材料（tech-design「跨仓改动设计 RC#1」+ PRD Related Change #1 前置门补充）；本表同时是 5.4 RC#1 迁出收尾的硬前置。**Hard Rule 口径：未经本表映射的能力项不得标记「已覆盖」；迁移判定以本表为唯一口径。**
>
> **证据形态**：证据-A = e-cam-web 源码逐行（本任务全量重盘，file:line 随行标注，盘点日期 2026-09-29）；证据-R/S = eiam 侧路由与类型（复用 §1/§2.4/§4.1 已核结论，本任务增量核查 G-8/G-9 两处）；引用复盘 = 全仓 grep 重跑（§6.3.1）。落点三选一图例：**【映射】**= 映射到控制台 UF 对应项 / **【不迁移】**= 有意不迁移（登记理由）/ **【缺口】**= eiam 缺端点进缺口处置（§3.1）。
>
> 旧页代码面 = `src/views/platform/users/index.vue`（974 行）+ `src/views/platform/users/components/UserForm.vue`（325 行）+ `src/api/eiam-users.ts`（134 行）；能力点编号 C-x（旧页能力）、R 表（控制台 UF-4 需求项对照）、N-x（eiam 既有能力落点）。

#### 6.3.1 引用复盘（全仓 grep 重跑，对照 tech-design RC#1 实核结论）

| 对象 | tech-design RC#1 实核结论 | 本次 grep 复盘（2026-09-29，`e-cam-web/src/` 全仓） | 一致性 | RC#1 落点 |
|------|---------------------------|------------------------------------------------------|--------|-----------|
| `src/api/eiam-users.ts` | 仅 2 处 import：index.vue + UserForm.vue，随页面整体迁出 | `views/platform/users/index.vue:23`、`views/platform/users/components/UserForm.vue:16`，无第三处 | 一致 | **随页面迁出（删除）** |
| `src/api/eiam-mapper.ts` | 引用方 = eiam-users.ts + UserForm.vue + 自测 eiam-mapper.test.ts，RC#1 后成死代码；**不导出 mapEiamProfile**（同名函数在 stores/user-mapper.ts） | `eiam-mapper.test.ts:3`、`eiam-users.ts:2`、`UserForm.vue:17`；`mapEiamProfile` 在本文件仅出现于注释（eiam-mapper.ts:51），定义在 `stores/user-mapper.ts:66`、`stores/user.ts:4,59` 引用 | 一致 | **判死随删（连测试文件一并删）** |
| `src/api/eiam-tenant.ts` | 保留，TenantSelector 在用 | 仅 `components/TenantSelector.vue:24`（switchTenant，§4.1 基线 #13） | 一致 | **保留** |
| `stores/user-mapper.ts` 的 mapEiamProfile | 独立保留，与 eiam-mapper.ts 无关 | 定义 user-mapper.ts:66；引用 stores/user.ts:4,59 + 自测 user-mapper.test.ts | 一致 | **保留** |

本次盘点**新增发现**（设计清单未列，随 RC#1 一并处置）：

| 对象 | 复盘结论 | 证据 | 处置 |
|------|----------|------|------|
| `src/api/types/eiam.ts` | **整文件判死**——importers 恰为四个 RC#1 消亡对象：eiam-mapper.ts:2、eiam-users.ts:10、UserForm.vue:23、index.vue:24 | grep `types/eiam` 全仓 4 处，无外部消费者 | 属设计「及仅被其引用的类型」条款，**整文件随 RC#1 删除**；其 re-export 的 `EiamTenant`（types/eiam.ts:73）源头在 stores/user-mapper.ts（保留，不受影响） |
| 路由/菜单删除面 | 与设计改动清单逐项吻合 | `router/routes.ts:111-113`（/platform/users 路由 + 动态 import index.vue）；`layouts/MainLayout.vue:331`（菜单项 platform-users，requireAdmin）；`components/PlatformNav/index.vue:51`（图标映射 `'platform-users'`）+ `:141`（导航项 `go('/platform/users')`） | 按设计清单执行 |
| `src/api/request/eiam.ts`（eiamAxios） | RC#1 后仍有存活消费者 | eiam-tenant.ts:1（TenantSelector 链路）；且为 RC#2 改造对象（redirectToLogin 目标值） | **保留** |

#### 6.3.2 能力清单与三选一落点（旧页能力 → 控制台落点）

**A. 列表页能力（index.vue，依赖 eiam-users.ts）**

| # | 能力点 | 当前实现方式 | 依赖 eiam 端点（§4.1 基线号） | 落点（三选一） | 证据（证据-A） |
|---|--------|--------------|------------------------------|----------------|----------------|
| C-1 | 用户列表分页展示 | 卡片列表 + 本地分页（offset/limit，页大小 10/20/50，页码省略号算法） | POST /api/iam/user/list（#1，L2 iam:user:view） | **【映射】UF-4 列表页（/users）** | eiam-users.ts:15-29；index.vue:78-94,373-410 |
| C-2 | 关键词搜索 | 500ms 防抖 + 重置按钮，keyword 透传 | user/list `keyword`（#1） | **【映射】UF-4 搜索过滤** | index.vue:96-109；eiam-users.ts:20 |
| C-3 | 统计卡片（总用户/启用/禁用） | total 取全局，启用/禁用为**本页前端本地计数**（实现注释自认「启用/禁用取本页」） | 无独立端点（复用 #1 响应） | **【不迁移】** 控制台 UF-4 规格无统计卡元素；「本页计数」语义在服务端分页下统计不完整（非全量），迁移价值低 | index.vue:46-48,232-260 |
| C-4 | 状态列（启用/禁用/未知） | `parseEiamStatus` 把 eiam 原值收敛为 `active/disable/unknown`，非法值归 unknown | user/list `UserVO.status`（#1） | **【映射】UF-4 状态列**（「锁定」项旧页本无、eiam 亦无数据源 → 见 G-9） | eiam-mapper.ts:43-47；index.vue:178-188,322-325 |
| C-5 | 最近登录时间列 | `last_login_at`（unix 秒）本地格式化，0 值显示 — | user/list `UserVO.last_login_at`（#1） | **【映射】UF-4 列表字段** | eiam-mapper.ts:66；index.vue:190-193,326 |
| C-6 | 启用/禁用切换（确认弹窗） | `updateUser({id, status})` 翻转 lever + ElMessageBox 确认 | POST /api/iam/user/update（#3，L2 iam:user:edit） | **【映射】UF-4 状态启停** | index.vue:129-146；eiam-users.ts:37-39 |
| C-7 | 删除用户（不可恢复确认） | `deleteUser(id)` 单条 + 确认弹窗 | DELETE /api/iam/user/delete/:id（#4，L2 iam:user:delete） | **【映射】UF-4 删除需确认**（UF-4 Flow） | index.vue:148-165；eiam-users.ts:41-43 |
| C-8 | 错误态展示与重试 | ErrorDisplay 组件 + apiUrl 透出 + retry 回调 | —（纯前端） | **【映射】UF-4 错误态 + E12 重试口径** | index.vue:37,263-268 |
| C-9 | 空状态（有/无筛选双文案） | 前端条件渲染 | —（纯前端） | **【映射】UF-4 空态** | index.vue:355-369 |
| C-10 | 非系统租户警示条（系统级能力守卫） | `currentTenantId === 1` computed；非系统租户显示警示条，提示切换系统管理空间 | 无独立端点（读 profile/登录态，§6.1.4） | **【映射】UF-11/E3** 控制台租户分流承接：管理页仅平台管理员可达且运行于系统租户 tid=1（§6.2 粒度结论），页面级警示条形态不复刻 | index.vue:30-33,220-229 |

**B. 表单弹窗能力（UserForm.vue）**

| # | 能力点 | 当前实现方式 | 依赖 eiam 端点（§4.1 基线号） | 落点（三选一） | 证据（证据-A） |
|---|--------|--------------|------------------------------|----------------|----------------|
| C-11 | 创建用户（用户名/密码/确认密码/昵称/邮箱/手机号/职位/状态） | `createUser(SignupRequest)` 拿新 id 后按需续调角色/租户 assign | POST /api/iam/user/create（#2，L2 iam:user:add；`password/confirm_password` 承载初始密码发放，§2.4 已核） | **【映射】UF-4 新增弹窗**；「系统生成初始密码」为控制台新增能力（前端生成、create 载荷不变，无 eiam 缺口） | UserForm.vue:175-196；eiam-users.ts:32-34 |
| C-12 | 编辑用户资料（用户名禁改） | `updateUser`；编辑回显直接用列表行数据（**无 detail 请求**） | POST /api/iam/user/update（#3） | **【映射】UF-4 编辑弹窗**；`user/detail`（C 档，§2.4）非必需，维持待核验备选 | UserForm.vue:124-151,198-210,241-247；eiam-users.ts:37-39 |
| C-13 | 角色多选分配（选项上限 200） | `listRoles({offset:0,limit:200})` 填选项，assign 以 `username+role_code` 为准 | POST /api/iam/role/list（#5）+ role/batch_assign（#7） | **【映射】UF-4 角色绑定 ref[] 多选** | UserForm.vue:99,188-191,277-294；eiam-users.ts:47-65,82-86 |
| C-14 | 用户当前角色回显 | `listRolesForUser(user_id)` | POST /api/iam/role/list/attached/user（#6） | **【映射】UF-4 编辑回显** | UserForm.vue:137-147；eiam-users.ts:68-79 |
| C-15 | 角色 diff 增删（快照 diff 提交） | 编辑态快照 `initialRoleCodes` → `diffSets` 分流 batch_assign / batch_unassign | role/batch_assign（#7）+ batch_unassign（#8） | **【映射】UF-4 保存链路**（diff 提交模式随迁，控制台可复用） | UserForm.vue:38,213-215；eiam-mapper.ts:88-95 |
| C-16 | 租户多选分配（选项上限 200） | `listTenants({offset:0,limit:200})` 填选项，assign 以 `user_ids+tenant_ids` 为准 | POST /api/iam/tenant/list（#9）+ tenant/batch_assign（#11） | **【映射】UF-4 所属租户 ref 关联选择** | UserForm.vue:104,193-195,296-314；eiam-users.ts:96-108,123-127 |
| C-17 | 用户当前租户回显 | `listTenantsForUser(user_id)` | POST /api/iam/tenant/list/attached/user（#10，WTO 越租户阻断） | **【映射】UF-4 编辑回显** | UserForm.vue:137-147；eiam-users.ts:111-120 |
| C-18 | 租户 diff 增删（快照 diff 提交） | 编辑态快照 `initialTenantIds` → `diffSets` 分流 batch_assign / batch_unassign | tenant/batch_assign（#11）+ batch_unassign（#12） | **【映射】UF-4 保存链路**（同 C-15 模式随迁） | UserForm.vue:39,218-220；eiam-users.ts:129-133 |
| C-19 | 租户选项 403 降级（非系统租户下） | `tenantsLoadError` 静默置空 + 选择器禁用 + placeholder 提示；角色/租户加载相互独立容错 | tenant/list 403（§6.2：tenant 模块平台级，tid≠1 触即 403） | **【映射】UF-4 表单容错**（角色/租户独立加载的容错模式随迁；403 降级分支在控制台固定系统租户语境下预计不触发，保留为防御） | UserForm.vue:34-35,91-113,302-305 |

#### 6.3.3 UF-4 控制台需求项对照（旧页承载现状 × 落点）

> 方向反转的覆盖检查：PRD UF-4 每个需求项必须有落点，「已覆盖」标记必须指向 §6.3.2 的 C-x 映射行（Hard Rule）。

| UF-4 需求项（prd-ui-functions.md「UI Function 4」） | 旧页承载现状 | 落点 | 关联 |
|----------------------------------------------------|--------------|------|------|
| 用户列表（分页/搜索） | 有 | **已覆盖**（经 C-1/C-2 映射）；加载态旧页为 v-loading，控制台 States 要求 skeleton，属实现形态差非能力缺口 | — |
| 用户名（必填，3–32 字符正则前端即时拦截，唯一性由 eiam 校验） | 旧表单仅 required 校验（无格式/长度规则） | **【映射】控制台新实现**（E10 前端即时拦截）；唯一性错误反馈链路已有（create/update 失败 ElMessage 透传 eiam 错误文案） | E10；UserForm.vue:59-72 |
| 初始密码二选一（管理员设置 / 系统生成，展示一次） | 设置侧有（C-11）；生成侧无 | **【映射】控制台新实现**（前端生成随机密码 + 弹窗展示一次；eiam create 载荷不变，§2.4 已核成立） | — |
| 首次登录强制改密 | **无**（eiam 无标记字段、无强制链路） | **【缺口】G-8**（本次新增，§3.1/§6.3.5） | E1 |
| 登录方式 enum（密码/passkey�� | 旧页仅密码账号 CRUD，无 passkey 概念 | **【缺口】G-2**（管理员发起注册引导）/ **G-3**（撤销，设备丢失重绑前置）；绑定态展示候选 = `GET /api/iam/user/identity/list`（L1 #24） | UF-4 passkey 引导定义 |
| 状态 enum（启用/禁用/锁定） | 启用/禁用有（C-4/C-6）；锁定无数据源 | **【缺口】G-9**（本次新增，§3.1/§6.3.5） | E2；§3.2 B-2 |
| 所属租户/组织 ref | 租户有（C-16~C-18）；组织无 | 租户**已覆盖**（经 C-16~C-18 映射）；组织承载 = department（§3.2 D-1）归 **UF-5 页面**，用户表单 v1 不并入组织选择（有意不迁移旧页形态，是否扩展归 UF-4 页面设计决策） | D-1 / UF-5 |
| 角色绑定 ref[] 多选 | 有 | **已覆盖**（经 C-13~C-15 映射） | — |
| 删除需确认 | 有 | **已覆盖**（经 C-7 映射） | E11 |
| 分页与并发边界（E6/E7） | E7 旧页本地分页/页码算法可参照（超界回落未显式处理）；E6 409 语义 eiam 未证 | E7 **【映射】控制台实现参照旧页算法补齐边界**；E6 归 **§6.4**（1.2/1.3 均未填，待补充） | E6/E7 |

#### 6.3.4 eiam 既有能力 × 控制台 v1 落点（含 group「有意不迁移」确认）

| 编号 | eiam 既有能力（端点族） | 旧页使用 | 控制台 v1 落点 | 理由 |
|------|--------------------------|----------|----------------|------|
| N-1 | group 用户组模块（§1.3，13 条） | **未使用**（e-cam-web 全部 eiam 客户端文件无 group 引用，grep 佐证） | **【不迁移】** | §2.5 预告项在此正式确认：PRD UF-1~11 无用户组能力诉求，控制台 v1 不设用户组管理页；group 为租户级（ScopeTenant 默认，§6.2），与「组织」department 是不同概念（D-1） |
| N-2 | user/ldap/search、ldap/sync、ldap/refresh_cache（§1.3，3 条） | 未使用 | **【不迁移】** | LDAP 登录走 UF-1（单通道，§3.2 B-3）；LDAP 用户同步为运维动作而非管理页能力，v1 控制台无此页；身份源配置归 UF-7 |
| N-3 | user/batch_delete（§1.3） | 未使用（仅单条删除 C-7） | **【不迁移】** | UF-4 Flow 为单条删除确认；v1 不做批量操作（端点保留，后续版本可接入） |
| N-4 | user/identity/manage（§1.3） | 未使用 | **【不迁移】** | 外部身份治理 v1 无承载页；passkey 相关落点见 §6.3.3（G-2/G-3 + identity/list L1 #24） |
| N-5 | user/detail（C 档，§2.4） | 未使用（编辑回显用列表行数据） | **【备选保留】** | 控制台编辑如需详情回显可接入（C 档升格流程 §3.3）；非必需端点，不进基线 |

#### 6.3.5 映射新增缺口（G-8/G-9，登记 §3.1 末两行）

| 编号 | 缺口 | 核查结论 | 建议处置 | 证据 |
|------|------|----------|----------|------|
| G-8 | 首登强制改密承载机制（`must_change_password` 字段 + 登录链路强制拦截 + 改密后清除） | UF-4/E1 要求「发放初始密码 → 首登强制改密」，eiam 全链路无承载：全仓 grep `must_change|MustChange|force_password` 0 命中（internal/**）；SignupRequest 仅 `password/confirm_password`（vo.go:20-24）；UpdateUserReq 无 password（vo.go:124-127，重置/改密均无法经 update 承载）；`password/update` 为 L1 本人自服务且需 old_password（user.go:87,416-428，§1.2 #16），不构成强制机制 | **eiam 排期补齐**，建议与 G-1 一并排期（创建与重置两条初始密码发放路径都依赖该机制）；备选 = 业务 owner 书面确认 v1 放弃强制改密（安全弱化须明示） | 证据-R + 证据-S |
| G-9 | 用户锁定态列表可见性 | UF-4 状态 enum 含「锁定」；eiam 锁定态为 redis 计数（cache/user.go:68-89 lockoutKey exists + 失败计数 24h TTL；login.go:38,52,58 判定/累计/清除），不上 user 模型与 VO（domain/user.go 无 lock 字段，grep 0 命中），§1 全量导出无锁定查询路由 | **UI 降级**：控制台用户列表状态列 = 启用/禁用/未知，锁定不展示（锁定提示由登录页错误文案承接，§3.2 B-2 口径）；若业务要求列表可见，转 eiam 补查询端点/字段 | 证据-R + 证据-S（login.go:37-58；cache/user.go:68-89；repository/user.go:497-503） |

#### 6.3.6 映射结论与会签

- 旧页 19 项能力（C-1~C-19）：**17 项【映射】**（其中 C-10/C-19 的守卫/降级形态映射到控制台租户分流 UF-11/E3，不复刻页面级实现）、**1 项【不迁移】**（C-3 统计卡）、**0 项悬空**；C-4 的「锁定」延伸面落 G-9、「组织」延伸面落 UF-5/D-1。
- 控制台 UF-4 需求项对照（§6.3.3）：4 项标记**已覆盖**（均有 C-x 映射指向），3 项控制台新实现（无 eiam 依赖），2 项落**既有缺口** G-2/G-3，2 项落**本次新增缺口** G-8/G-9。
- eiam 既有能力落点（§6.3.4）：group 等 4 族**有意不迁移**（N-1~N-4），user/detail 备选保留（N-5）。
- RC#1 引用盘点结论经全仓 grep 重跑**全部一致**（§6.3.1），设计改动清单生效；新增 `types/eiam.ts` 整文件判死，并入迁出 commit。
- 证据形态边界：本表逐项已附证据-A（file:line 代码级，盘点日期 2026-09-29）；双签材料中的「截图/接口响应/用例链接」属会签执行时的逐项补充材料——随 5.4 RC#1 迁出收尾与 acceptance 场景回填产出，三方会签位（上表）随证据-O 一并落笔（与 §6.1.9/§6.2.5 会签位留待签同口径）。

| 会签 | 会签事项 | 签名 | 日期 |
|------|----------|------|------|
| 平台管理员 | §6.3 能力映射完整性（C-1~C-19 / R 表 / N-1~N-5 无悬空项） | ＿＿＿＿ | ＿＿＿＿ |
| eiam owner | §6.3 端点口径（§4.1 基线引用）+ G-8/G-9 处置建议 | ＿＿＿＿ | ＿＿＿＿ |
| e-cam-web 前端 owner | §6.3.1 引用复盘 + §6.3.2 能力清单证据-A 复核（跨仓改动归属方） | ＿＿＿＿ | ＿＿＿＿ |

---

## 7. 证据来源形态汇总与会签

### 7.1 本任务证据形态覆盖

| 核查项类别 | 证据形态 | 已覆盖 | 待补 |
|------------|----------|--------|------|
| eiam 路由全量导出（§1，116 条） | 证据-R（路由注册文件:行号） | ✓ 全覆盖 | — |
| A 档形状核对（§2/§4.1） | 证据-A（e-cam-web 源文件:行号）+ 证据-S | ✓ 全覆盖 | — |
| B/C 档形状核对 | 证据-R + 证据-S（类型定义:行号） | ✓ 路径/类型级 | 响应样本（证据-P）→ 1.2/1.3 |
| 缺口判定（G-1~G-6） | 证据-R（全量导出反证）+ 证据-S | ✓ 全覆盖 | — |
| 缺口处置决策 | 证据-O（owner 书面确认） | 决策口径已落档（§3.1.1，Phase 0 出口门 2026-09-30：默认 UI 降级/裁剪，G-7 安全项例外） | 证据-O 会签位（§3.1/§4.3/§7.2，待签） |
| 契约基线冻结 | 证据-O + 证据-A/S | 内容 ✓ | §4.3 会签位 |
| 认证链路兼容矩阵（§6.1，任务 1.2） | 证据-R/S（源码逐行）+ 证据-P（本地 dev live 五链路探测，2026-09-29） | ✓ cookie/JWT/profile/tenant_id/switch/登出全覆盖 | F-1~F-4 处置与矩阵会签（证据-O）→ §6.1.9 |
| 服务端授权核查矩阵（§6.2，任务 1.3） | 证据-R/S（授权中间件/capability 注册/资产绑定链/边界校验器/路由声明源码逐行；未做 live 探测，403 为代码确凿路径） | ✓ 机制链 + 89 条 L2 逐端点（能力码/粒度/403 路径）+ L0/L1 暴露面 27 条复核 + 缺省放行三分支核查 | G-7 处置与矩阵结论会签（证据-O）→ §6.2.5 |
| platform/users 能力映射（§6.3，任务 1.4） | 证据-A（e-cam-web 源码逐行，index.vue/UserForm.vue/eiam-users.ts/eiam-mapper.ts 全量重盘）+ 证据-R/S（G-8/G-9 增量核查：lock 计数、must_change_password、password/update）+ 全仓 grep 引用复盘 | ✓ 19 项能力三选一落点无悬空 + UF-4 需求项逐项对照 + 引用复盘与设计结论一致 | G-8/G-9 处置与映射表会签（证据-O）→ §6.3.6 |

### 7.2 会签表

| 角色 | 会签事项 | 签名 | 日期 |
|------|----------|------|------|
| 平台管理员 | §2 差集盘点、§3 处置建议、§4 基线起草 | ＿＿＿＿ | ＿＿＿＿ |
| eiam owner | §3.1 G-1~G-4 处置、§4.3 基线冻结、变更通知条款 | ＿＿＿＿ | ＿＿＿＿ |
| 业务 owner | §3.1 G-5~G-6 处置（授权语义映射/静态词表降级） | ＿＿＿＿ | ＿＿＿＿ |
| e-cam-web 前端 owner | A 档证据复核（eiam-users.ts / eiam-tenant.ts / eiam-mapper.ts / stores/user.ts 行号引用） | ＿＿＿＿ | ＿＿＿＿ |
