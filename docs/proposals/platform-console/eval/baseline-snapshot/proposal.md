---
created: "2026-09-29"
author: "Haven"
status: Draft
intent: "new-feature"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: 平台控制台子项目（统一入口 + 身份治理 + 权限管理 + 工作台）

## Problem

去掉 Duke1616 的 ecmdb-web 之后，平台级的「统一入口 + 权限管理 + 身份治理 + 工作台」将失去归属：这些 UI 目前寄生在 ecmdb-web 里（其 `login`/`iam` 视图直调 eiam），而 e-cam-web 是云管业务前端，不承载平台身份治理。

### Evidence

- e-cam-web `views/iam/`（audit/compare/groups/sync/templates/users）是**云 IAM**——云账号的子账号/用户组/策略同步，与平台身份是两个领域。
- e-cam-web 对平台身份只有 `views/platform/users` 一个孤立页面，无成体系的身份治理/权限管理。
- ecmdb-web 承担 `login`（登录，调 eiam 的 `/user/{system,ldap,passkey}/login`）+ `iam`（平台用户/角色/租户/授权）+ `navigation` + `ticket`/`change`。
- eiam 决定**保留**，但它没有自己的前端；其平台管理 UI 随 ecmdb-web 下线而消失。
- 现有 `docs/proposals/` 29 个提案中没有任何一个涉及「统一入口/门户/工作台/身份治理控制台」。

### Urgency

**中**：不是数据正确性缺陷，但这是「去 ecmdb-web」重构的前置依赖——一旦下线 ecmdb-web，平台身份/权限管理 UI 直接断供。越晚建控制台，越长时间依赖 Duke1616 前端。成本是「去他依赖 / 100% 可控」目标延期。

## Proposed Solution

新建一个**独立的纯前端**「平台控制台」子项目，直调 eiam 已有 API，承载四类能力，e-cam-web 保持云管定位，两者单点登录互通：

1. **登录/登出 + 统一导航**：接管 ecmdb-web 现在的登录页（eiam 的 system/LDAP/passkey/MFA 流程）+ 顶部导航 + 应用切换（可单点跳 e-cam-web）。
2. **身份治理**：平台用户/组织/租户/身份源管理（直调 eiam，不写后端）。
3. **权限管理**：角色/授权/策略管理（直调 eiam，不写后端）。
4. **工作台**（纯 eiam 侧）：欢迎 + 我的信息 + 应用快捷导航，不跨服务聚合云管数据。
5. **边界清理**：从 e-cam-web 迁出平台身份 UI（`platform/users`），消除双处维护。

关键约束：**控制台为零后端纯前端**（除 eiam 外不新增任何服务），登录态复用 eiam 写入的共享 Redis session——与 e-cam-service 认证中间件（`EcmdbAuthMiddleware` 已在读该 session）天然兼容，后端零改动。

### Innovation Highlights

诚实说明：这是企业级**「门户/控制台与业务应用分离」**的标准架构（对标云厂商控制台、Keycloak/Casdoor 等身份服务自带管理控制台），模式本身无创新。真正的取舍点在两处克制：

1. **纯前端直调 eiam、零新增后端**——不做 BFF、不做聚合层，避免为"工作台聚合"而过早引入微前端/网关复杂度（YAGNI）。
2. **把身份治理从云管前端剥离开**——让 e-cam-web 专注云管，平台治理职责归控制台，避免单一前端无限膨胀。

## Requirements Analysis

### Key Scenarios

- 用户通过控制台登录（密码为最低要求；LDAP/passkey/MFA 沿用 eiam 已有能力并补 UI 入口）。
- 管理员对平台用户/角色/租户/授权做增删改查，能力不低于原 ecmdb-web 的 iam 管理页。
- 用户从控制台工作台/顶部导航单点跳转入 e-cam-web，共享会话、不重复登录。
- 边界情况：会话过期 → 统一跳回登录页；无权限操作 → 明确提示而非白屏。
- 错误情况：eiam 某端点缺（部分身份/权限操作无对应 API）→ Phase 0 盘点兜底，UI 降级提示。

### Non-Functional Requirements

- **安全**：登录态与 e-cam-service 认证中间件口令一致（共享 cookie/session 机制），不自造鉴权。
- **兼容**：cookie 的 domain/路径与 eiam、e-cam-web 一致（SSO 及共享 session 的前提）。
- **性能**：纯静态前端站，可 CDN 托管，首屏资源可缓存。
- **可验证性（硬性）**：登录/登出、SSO 跳转、身份治理/权限管理核心链路必须经**真实浏览器**验证（Playwright 或浏览器 MCP 自动化），不接受仅单测/接口级验证——沿用本项目既有前例（e-cam-web 经 `8888/cam` 真实浏览器联调、MCP 编程操作登录与 el-select 等）。
- 无障碍/多语言：v1 不强制，沿用 e-cam-web 现状即可。

### Constraints & Dependencies

- 依赖 eiam 已有 HTTP API（`/api/iam/user/*` 等）与 session 机制，eiam 必须保留且稳定。
- 依赖 e-cam-service 认证中间件对共享 Redis session 的读取逻辑**保持不变**（本次不动）。
- 控制台需与 e-cam-web 在同一登录域下，共享认证/租户选择语义（eiam 以 tenant_id 区分租户）。
- 约束（非选型）：控制台与 e-cam-web 需共享视觉体系与交互一致性，避免两套产品体验。

## Alternatives & Industry Benchmarking

### Industry Solutions

- 云厂商控制台（阿里云/腾讯云）：统一门户 + 分业务应用，SSO 互通。
- 开源身份服务自带管理控制台：Keycloak Admin Console、Casdoor、Authing——管理模式与本提案一致。
- 微前端门户：qiankun / wujie 把多个业务应用装入单一外壳。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing（保留 ecmdb-web 当身份/门户壳） | — | 零新增工作 | 继续依赖 Duke1616 前端，违背「100% 可控」 | Rejected: 与核心动机冲突 |
| fork 收编 ecmdb-web 的 login/iam 视图 | ecmdb-web | 快，eiam API 集成现成 | 起步仍是他人代码，且与 ecmdb 自身 API 层耦合需清理 | Rejected: 血统与解耦成本不划算 |
| 微前端门户外壳（云管装进去） | qiankun/wujie | 单入口体验统一 | 重、改造面最大，与「去依赖」目标无关 | Rejected: 过度设计（YAGNI） |
| **独立纯前端控制台（自建）** | — | 血统纯、边界清晰、无后端、可独立演进 | 页面需逐个写 | **Selected: 匹配「100% 可控」，且克制无后端** |

## Feasibility Assessment

### Technical Feasibility

可行，无 showstopper：eiam API 已存在并在生产运行；session 机制已与 e-cam-service 认证中间件解耦（共享 Redis）。唯一需验证的是 eiam API 面对身份治理/权限管理的覆盖度（Phase 0 盘点）。

### Resource & Timeline

团队已熟悉 e-cam-web 同生态（现有 26 个 UI 重写提案均在此栈落地），无新技能门槛。v1 页面量可控（登录 + 用户/角色/租户/授权 列表与详情 + 工作台首页 + 导航）。

### Dependency Readiness

eiam 稳定（已承载平台登录/鉴权）。Phase 0 需盘点「原 ecmdb-web iam 页面的后端端点清单」与「eiam 现有端点」的差集，作为任务拆解输入。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 统一入口必须把云管装进一个外壳 | Assumption Flip（反转：外壳只装治理能力） | Overturned: 控制台与云管平级 + SSO 更简单可控 |
| 工作台需要跨 eiam+e-cam-service 聚合 | Occam's Razor | Refined: v1 收敛为纯 eiam 侧（欢迎+我的信息+快捷导航） |
| 直接并入 e-cam-web 最省事 | 5 Whys（为什么云管前端要承载平台治理） | Overturned: 云管与平台治理职责不同，混合会拖累两者 |

## Scope

### In Scope

- 新建独立「平台控制台」前端子项目（纯前端、直调 eiam、无自建后端）
- 登录/登出（eiam system/LDAP/passkey/MFA 流程）
- 统一导航 + 应用切换（单点跳 e-cam-web）
- 身份治理：平台用户/组织/租户/身份源管理
- 权限管理：角色/授权/策略管理
- 工作台（纯 eiam 侧：欢迎 + 我的信息 + 应用快捷导航）
- 从 e-cam-web 迁出平台身份 UI（`platform/users`），消除重复

### Out of Scope

- 工作台跨服务聚合（云资产/告警/证书概览）——后续迭代
- eiam 本身替换（保留 eiam，含其 SDK/服务）
- 后端改造：反转 ecmdb 决策 / internal/cmdb 扶正——另立提案
- 工单（ticket）/ 变更（change）
- 自建 BFF / 聚合网关后端
- 云管业务演进（e-cam-web 保持现职）
- 共享组件库 / 设计系统沉淀——另立提案（v1 仅视觉对齐）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 登录态/SSO 对接不准（cookie/session 格式不兼容） | M | H | 照抄 ecmdb-web 现状的登录/session 写入口径，用 Playwright/浏览器 MCP 真实浏览器实测闭环 |
| eiam API 面覆盖不足（部分身份/权限操作缺端点） | M | H | Phase 0 盘点 eiam API 与 ecmdb-web iam 页面差距，缺口显式列单 |
| e-cam-web 迁出 platform/users 牵连内部引用 | M | M | 先盘点引用关系，灰度迁移，保留回退 |
| 工作台早期内容空、价值弱 | H | L | 先做轻量入口（欢迎+我的信息+快捷导航），不聚合，避免过度设计 |
| 新前端与 e-cam-web 视觉/组件不统一 | M | M | v1 先视觉对齐；共享设计资产另立提案 |

## Success Criteria

- [ ] 控制台可完成密码登录与登出；LDAP/passkey/MFA 沿用 eiam 能力并补 UI 入口
- [ ] 平台用户/角色/租户/授权可增删改查，能力不低于原 ecmdb-web 的 iam 管理页（按 Phase 0 盘点清单逐项覆盖）
- [ ] 控制台顶部导航可单点跳转 e-cam-web，共享会话、不重复登录
- [ ] 登录/登出与 SSO 跳转经真实浏览器（Playwright 或浏览器 MCP）验证通过，而非仅接口级断言
- [ ] 工作台可展示当前用户信息（来自 eiam）与应用快捷导航，且可跳转
- [ ] e-cam-web 迁出平台身份 UI 后云管功能无回退（走查通过）
- [ ] 全链路除 eiam 外不新增任何后端服务（部署清单可核验）

## Next Steps

- Proceed to `/write-prd` to formalize requirements