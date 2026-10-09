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
- eiam 决定**保留**，但它没有自己的前端；其平台管理 UI 随 ecmdb-web 下线而消失。该决策目前为口头结论、无书面出处，列为**待确认假设 A-1**：由提案作者在提案评审会上向架构负责人书面确认（决策出处、时间、授权方）；确认结果为否则本提案立即终止。
- 现有 `docs/proposals/` 29 个提案中没有任何一个涉及「统一入口/门户/工作台/身份治理控制台」。

### Urgency

**中**：不是数据正确性缺陷，但这是「去 ecmdb-web」重构的前置依赖——一旦下线 ecmdb-web，平台身份/权限管理 UI 直接断供。越晚建控制台，越长时间依赖 Duke1616 前端。成本是「去他依赖 / 100% 可控」目标延期。时间锚点：ecmdb-web 前端下线的触发条件即本提案的「下线就绪」判据（见 Success Criteria）——控制台每顺延一个里程碑（约 2–4 周），下线同步顺延。

## Proposed Solution

新建一个**独立的纯前端**「平台控制台」子项目，直调 eiam 已有 API，承载四类能力，e-cam-web 保持云管定位，两者单点登录互通：

1. **登录/登出 + 统一导航**：接管 ecmdb-web 现在的登录页（eiam 的 system/LDAP/passkey/MFA 流程）+ 顶部导航 + 应用切换（可单点跳 e-cam-web）。
2. **身份治理**：平台用户/组织/租户/身份源管理（直调 eiam，不写后端）。
3. **权限管理**：角色/授权/策略管理（直调 eiam，不写后端）。
4. **工作台**（纯 eiam 侧）：欢迎 + 我的信息 + 应用快捷导航，不跨服务聚合云管数据。
5. **边界清理**：从 e-cam-web 迁出平台身份 UI（`platform/users`），消除双处维护。

关键约束：**控制台为零后端纯前端**（除 eiam 外不新增任何服务），登录态复用 eiam 写入的共享 Redis session。「与 e-cam-service 认证中间件（`EcmdbAuthMiddleware` 已在读该 session）兼容、后端零改动」是基于现状的推断，尚非源码核验结论——Phase 0 增设**认证链路事实核查**前置��务，产出 cookie/session 读写兼容矩阵（写入方/读取方、字段、domain/path、SameSite、过期与续期行为），核查通过前不将该承诺视为成立。

### Innovation Highlights

诚实说明：这是企业级**「门户/控制台与业务应用分离」**的标准架构（对标云厂商控制台、Keycloak/Casdoor 等身份服务自带管理控制台），模式本身无创新。真正的取舍点在两处克制：

1. **纯前端直调 eiam、零新增后端**——不做 BFF、不做聚合层，避免为"工作台聚合"而过早引入微前端/网关复杂度（YAGNI）。
2. **把身份治理从云管前端剥离开**——让 e-cam-web 专注云管，平台治理职责归控制台，避免单一前端无限膨胀。

## Requirements Analysis

### Key Scenarios

- 用户通过控制台登录（密码为最低要求）。登录页按身份源提供入口：system 密码表单、LDAP 账号密码（eiam 侧校验）、passkey 按钮（WebAuthn 流程沿用 eiam，注册/管理入口在「我的信息-安全设置」）；MFA 为密码通过后 eiam 返回挑战 → 控制台渲染挑战输入页并提交 eiam 判定，挑战形式随 eiam 现状。
- 管理员对平台用户/角色/租户/授权做增删改查，能力不低于原 ecmdb-web 的 iam 管理页。
- 用户从控制台工作台/顶部导航单点跳转入 e-cam-web，共享会话、不重复登录；任一侧登出，另一侧会话同步失效（双向单点登出）。
- 多标签/并发会话：任一标签登出或 session 失效后，其余标签不做实时广播（v1 不引入 WebSocket/BroadcastChannel），其下一次任意请求或路由跳转即 401 → 统一跳回登录页。
- eiam 不可用/降级：控制台为静态站，页面资源仍可加载；登录与数据操作呈现明确错误提示（含重试指引），不白屏、不无限 loading；已登录的 e-cam-web 会话不受影响。
- 边界情况：会话过期 → 统一跳回登录页；无权限操作（403）与 API 失败由**统一的 API 拦截器 + 路由级错误页**承载，渲染明确提示（403 无权限、网络错误可重试、其余透出 eiam 错误信息），不白屏。
- 错误情况：eiam 某端点缺（部分身份/权限操作无对应 API）→ Phase 0 缺口报告逐项决策三类处置之一（UI 降级 / 扩 eiam 端点 / 放弃能力），其中属 parity 域（iam 管理 CRUD）的缺口不适用「UI 降级/放弃能力」，只能实现或扩 eiam 端点后实现（与处置门、SC-2 同一口径），不做静默缺失。

### Non-Functional Requirements

- **安全**：登录态与 e-cam-service 认证中间件口令一致（共享 cookie/session 机制），不自造鉴权。
- **登录页安全面**：登录页仅作 eiam 已有流程的 UI 壳，凭据校验全部在 eiam 侧；CSP、`X-Frame-Options: DENY`、HSTS 由 nginx 反代统一注入；CSRF 防护归属 eiam 登录接口现有机制，控制台不自造；登录页不引入任何第三方脚本。核验口径：部署后经浏览器 DevTools 检查响应头与页面脚本来源，纳入迁出走查清单。
- **兼容**：cookie 的 domain/路径与 eiam、e-cam-web 一致（SSO 及共享 session 的前提）。浏览器矩阵：Chrome/Edge 最近两个大版本 + Firefox ESR，与 e-cam-web 支持口径一致，不支持 IE。
- **性能（可测）**：登录页首屏静态资源 gzip 后 ≤ 300 KB，同域内网 LCP ≤ 2.5 s；构建产物文件名含内容哈希可长缓存，二次访问 LCP ≤ 1 s。宿主形态钉死为**与 e-cam-web 同域的 nginx 反代路径**——同域是共享认证 cookie 的前提；跨域 CDN 直调不受支持（cookie SameSite 与 eiam CORS 限制），仅当未来变更部署形态并专项核对 CORS/SameSite 后才可重议。
- **可用性**：401 切换后控制台登录页成为全员登录链路入口（见 Key Risks），可用性基线为「与 e-cam-web 静态资源同一 nginx 运维面、同一可用性现状」，不新增独立 SLA；部署清单须包含 `/console` 静态资源探活（HTTP 200 健康检查）与告警接入（沿用 e-cam-web 现有告警渠道）；上线前完成一次 401 目标回切 ecmdb-web 登录页的开关演练，回切操作须在 10 分钟内完成，验证故障退路可用。
- **可验证性（硬性）**：登录/登出、SSO 跳转、身份治理/权限管理核心链路必须经**真实浏览器**验证（Playwright 或浏览器 MCP 自动化），不接受仅单测/接口级验证——沿用本项目既有前例（e-cam-web 经 `8888/cam` 真实浏览器联调、MCP 编程操作登录与 el-select 等）。SSO/共享会话闭环必须在**与生产一致的同域宿主形态（nginx 反代路径）**下验证，不得以独立端口 dev server 的通过作为判定依据，避免验证环境宿主口径不一致导致假通过。
- 无障碍/多语言：v1 不强制，沿用 e-cam-web 现状即可。

### Constraints & Dependencies

- 依赖 eiam 已有 HTTP API（`/api/iam/user/*` 等）与 session 机制。eiam 必须保留；「稳定」定义为：现有 API 不做不兼容变更，可用性沿用 eiam 现状（与 e-cam-service 登录链路同级），本提案不新增 SLA 承诺。
- 依赖 e-cam-service 认证中间件对共享 Redis session 的读取逻辑**保持不变**（本次不动）。
- **宿主形态（钉死）**：控制台与 e-cam-web 处于同一登录域——部署为同一域名下 nginx 反代路径（如 `/console`），不采用独立域名/端口；认证 cookie 的 domain/path 必须与 eiam、e-cam-web 写入口径一致，SameSite/CORS 行为列入 Phase 0 认证链路事实核查；反代配置改动由本提案提交 PR、经宿主 nginx 现有维护流程评审合入，不单方变更。
- **租户选择语义**：eiam 以 tenant_id 区分租户；但「是否需要选租户交互、tenant_id 存放在 session 还是 cookie、跳转 e-cam-web 时租户上下文从何而来」以 Phase 0 事实核查（读 eiam session 结构与 ecmdb-web 现行为）结论为准，本提案不预设——核查结论落地前，按「登录成功但看到错租户数据」为最高优先风险对待。
- 约束（非选型）：控制台与 e-cam-web 需共享视觉体系与交互一致性，避免两套产品体验；落地口径为「与 e-cam-web 同一组件库且锁同一大版本 + 主题 token 以共享文件形式引用（非复制粘贴）」。技术栈随 e-cam-web 现状锁定：Vue 3 + Vite + Element Plus + Pinia，测试 Vitest + Playwright（Phase 0 盘点时固化具体版本）。

## Alternatives & Industry Benchmarking

### Industry Solutions

- 云厂商控制台（阿里云/腾讯云）：统一门户 + 分业务应用，SSO 互通。
- 开源身份服务自带管理控制台：Keycloak Admin Console、Casdoor、Authing——管理模式同源（控制台承载 IdP 的治理能力），部署形态分歧见下。
- 微前端门户：qiankun / wujie 把多个业务应用装入单一外壳。

机制级借鉴：云厂商控制台的 SSO 依赖「统一登录域写 cookie + 各业务域同域反代共享会话」，本提案的同域 nginx 反代路径宿主正是该机制的落地。与所引基准的**部署分歧**：Keycloak/Casdoor 的管理控制台由 IdP 自身托管、随服务一起部署升级；本提案的控制台是独立静态站、独立仓，部署与升级解耦于 eiam——取舍是 eiam 升级不牵连前端、前端可独立演进，代价是 UI 与 eiam 的版本一致性只能靠 API 兼容约定与 Phase 0 端点清单维护。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing（保留 ecmdb-web 当身份/门户壳） | — | 零新增工作 | 继续依赖 Duke1616 前端，违背「100% 可控」 | Rejected: 与核心动机冲突 |
| fork 收编 ecmdb-web 的 login/iam 视图 | ecmdb-web | 快，eiam API 集成现成 | 起步仍是他人代码，且与 ecmdb 自身 API 层耦合需清理 | Rejected: 血统与解耦成本不划算 |
| 复用开源控制台 UI 壳（fork Casdoor 管理界面 / vue-element-admin 类模板作骨架，指向 eiam API） | 开源项目 | 布局/列表/表单壳现成，起步快，正面回应自建方案「页面需逐个写」的缺点 | 骨架自带的路由/权限/请求层约定需剥离改造才能对齐 eiam API 形态与共享 session 认证；引入外部血统与上游跟随负担，与「100% 可控」冲突 | Rejected: 省的是壳、费的是剥离与长期跟随，可控性目标受损 |
| 并入 e-cam-web（平台治理页作为其模块） | e-cam-web | 无新仓，复用现有登录/构建/发布 | 云管前端承载平台治理职责（5 Whys 已否），平台身份发布节奏被云管绑死 | Rejected: 职责混合 |
| 微前端门户外壳（云管装进去） | qiankun/wujie | 单入口体验统一 | 重、改造面最大，与「去依赖」目标无关 | Rejected: 过度设计（YAGNI） |
| **独立纯前端控制台（自建）** | — | 血统纯、边界清晰、无后端、可独立演进 | 页面需逐个写 | **Selected: 匹配「100% 可控」，且克制无后端** |

## Feasibility Assessment

### Technical Feasibility

初步可行，**待 Phase 0 盘点确认**——两项前置核查未完成前不下「无 showstopper」结论：eiam API 已存在并在生产运行；session 机制与 e-cam-service 认证中间件解耦（共享 Redis）。需验证的有两项：①eiam API 面对身份治理/权限管理的覆盖度（端点差集盘点）；②认证链路事实核查（cookie/session 读写兼容矩阵）。

### Resource & Timeline

团队已熟悉 e-cam-web 同生态（现有 26 个 UI 重写提案均在此栈落地），无新技能门槛。估算（1 名前端）：Phase 0 约 1 周；v1 开发 5.5–6.5 人周——登录/SSO 同域闭环 1.5、身份/权限 CRUD 2–3、工作台+导航 0.5、迁出与 401 切换 1、真实浏览器验收与部署 0.5（总项即分项之和）。里程碑：M1 Phase 0 结论 → M2 登录+SSO 闭环 → M3 iam CRUD parity → M4 迁出+下线就绪。M1→M2 关键路径含两段外部等待，计入里程碑日历排期、不计入上述人周估算：nginx 反代配置 PR 的评审合入周期（外部维护流程）、扩端点排期的 1 周书面确认窗口；获确认的扩端点项按项追加估算并顺延对应里程碑，不占用基线人周。v1 页面量可控（登录 + 用户/角色/租户/授权 列表与详情 + 工作台首页 + 导航）。

### Dependency Readiness

eiam 稳定（已承载平台登录/鉴权）。Phase 0 需盘点「原 ecmdb-web iam 页面的后端端点清单」与「eiam 现有端点」的差集，并对每项缺口做处置决策，作为任务拆解输入：
- **认证链路事实核查**：读 eiam session 写入逻辑与 `EcmdbAuthMiddleware` 读取逻辑源码，产出 cookie/session 读写兼容矩阵（字段、domain/path、SameSite、过期与续期行为），作为「后端零改动」承诺与同域 cookie 前提的核验依据。
- **e-cam-web 登出链路核查**：读 e-cam-web 登出逻辑源码，确认其登出时是否删除共享 Redis session——这是「e-cam-web 登出后控制台会话同样失效」这条 SC 的机制前提，不在前两处源码范围内。
- **缺口处置决策门**：缺口报告不得止步于列缺口，每项缺口须归入三类处置之一——**UI 降级**（控制台内明确提示能力受限）、**扩 eiam 端点**（属 eiam 自身演进，与「控制台零新增后端」不冲突，但推翻「eiam 完全不动」假设：由提案作者在缺口报告产出后 1 周内与 eiam 后端 owner 书面确认排期，未确认排期的项不计入 v1 承诺）、**放弃能力**（该操作不迁移，登记为功能损失）。与 SC-2 的对齐口径：属 parity 域（iam 管理 CRUD）的缺口不适用「UI 降级/放弃能力」，只能实现或扩端点后实现；这两个分支仅适用于清单内非 parity 域与清单外能力。

**Phase 0 执行计划**（全文各处下游承诺的共同前置）：
- Owner：提案作者（核查与产出）；eiam 后端 owner（session 写入侧答疑、扩端点排期确认、产出物①③会签）。两者在提案评审会上认领。
- 时间盒：5 个工作日产出报告；逾期未收敛升级至提案评审会决策，不允许无限悬置。
- 产出物：① cookie/session 读写兼容矩阵；② eiam 端点差集与缺口处置单；③ 租户语义结论；④ e-cam-web 登出链路核查结论。
- 退出判据：四项产出物齐备，逐项标注「兼容 / 不兼容 / 需扩端点」，且产出物①（兼容矩阵）③（租户语义）经 eiam 后端 owner 书面会签（或以评审会独立评审门替代）——未会签不进入 M2。
- 失败分支：兼容矩阵出现不兼容项且无法以同域 cookie 口径解决 → 本提案重定范围为「含最小后端适配层」并重新评审；eiam API 覆盖度低 → v1 收缩为已覆盖端点的能力子集，缺口按处置门决策；「扩 eiam 端点」排期确认失败 → v1 范围重议，parity 承诺随之重谈；③ 租户语义核查失败（不改后端无法可靠传递租户上下文）→ v1 不做租户选择交互、按 eiam session 现状传递，租户交互登记为后续迭代，若错租户风险无法规避则并入「重定范围并重新评审」路径；④ e-cam-web 登出链路核查失败（登出不删共享 session）→ 修复属第三类 e-cam-web 改动、超出 Out of Scope 封闭清单（仅迁出与 401 切换两类例外）→ 触发范围重议并重新评审，重议前「双向单点登出」SC 不视为可满足。

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
- 从 e-cam-web 迁出平台身份 UI（`platform/users`），消除重复；**前置门：Phase 0 端点清单逐项「处置完成」**（parity 项已实现，清单外项处置决策落地并登记）方可执行迁出，迁出 PR 附引用盘点、回归走查与回退方案，避免平台身份管理出现比现状更糟的真空 
- 产出「ecmdb-web 前端下线就绪清单」：平台级入口（login/iam/navigation）逐项处置完成（与下线就绪 SC 同一口径：parity 项已实现、清单外项处置决策落地登记）+ 401 切换与迁出回归通过即视为就绪（执行下线本身不属本提案，见 Out of Scope）；清单须登记 ticket/change 退役前置条件（近 90 天使用数据与使用方确认结论），该前置动作由下线提案在 ecmdb-web 仍在线的窗口内执行
- e-cam-web 会话过期/401 跳转目标切换为控制台登录页（现指向 ecmdb-web 登录页；属共享登录体系后的必然联动改动，显式纳入范围） 

### Out of Scope

- 工作台跨服务聚合（云资产/告警/证书概览）——后续迭代
- eiam 本身替换（保留 eiam，含其 SDK/服务）
- 后端改造：反转 ecmdb 决策 / internal/cmdb 扶正——另立提案
- 工单（ticket）/ 变更（change）：**有意退役**——随 ecmdb-web 下线一并消失，控制台不重建、本提案不为其归宿兜底；若日后业务需要，另立提案申请。退役前置动作：拉取 ticket/change 近 90 天使用数据并向现有使用方确认——存在活跃使用则退役决定须使用方 owner 显式签字，无使用证据则按计划随下线消失。前置动作执行归属：由下线提案在 ecmdb-web 仍在线期间执行，本提案经「下线就绪清单」登记该前置条件使其进入交接视线（见 In Scope） 
- 自建 BFF / 聚合网关后端
- 云管业务演进（e-cam-web 不承载平台治理职责、不因本提案新增业务功能；本提案内的迁出 `platform/users` 与 401 跳转切换两类边界清理改动除外）
- 共享组件库抽包 / 设计系统规范文档沉淀——另立提案；v1 的「视觉对齐」限定为：同一组件库锁同一大版本 + 主题 token 以共享文件形式引用（见 Constraints），不含组件库抽包与规范文档
- 执行 ecmdb-web 前端下线本身——由其归属方按「下线就绪清单」另立下线提案，本提案只交付就绪判据

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 登录态/SSO 对接不准（cookie/session 格式不兼容） | H | H | 先完成 Phase 0 认证链路事实核查（兼容矩阵），再照 ecmdb-web 现状口径实现；用 Playwright/浏览器 MCP 在同域宿主形态下实测登入 + 双向登出闭环  |
| 租户上下文错乱（登录成功但看到错租户数据，Constraints 列为最高优先风险） | M | H | 实现唯一依据为 Phase 0 产出物③租户语义结论（eiam session 结构 + ecmdb-web 现行为），不预设选租户交互；核查失败走 Phase 0 失败分支（错租户风险无法规避则范围重议）；验收走查含多租户账号登录后的租户上下文正确性（owner：提案作者） |
| eiam API 面覆盖不足（部分身份/权限操作缺端点） | M | H | Phase 0 盘点 eiam API 与 ecmdb-web iam 页面差距，缺口显式列单 |
| e-cam-web 迁出 platform/users 牵连内部引用 | M | M | 前置门：Phase 0 端点清单逐项处置完成（parity 项已实现）；先盘点引用关系，灰度迁移，迁出 PR 附回归走查与回退方案  |
| 工作台早期内容空、价值弱 | H | L | 先做轻量入口（欢迎+我的信息+快捷导航），不聚合，避免过度设计 |
| 新前端与 e-cam-web 视觉/组件不统一 | M | M | v1 视觉对齐的操作定义：同一组件库大版本 + 共享主题 token 文件渲染，登录/列表/详情/工作台首页四类关键页与 e-cam-web 截图走查无样式断裂；共享设计资产另立提案 |
| eiam 团队带宽不足（扩端点排期无法承诺） | M | H | Phase 0 处置单标出全部需扩端点项，报告产出后 1 周内与 eiam 后端 owner 书面确认；未确认项不计入 v1 承诺并触发范围重议 |
| ticket/change 现有使用方反弹（无证据即退役） | M | M | 退役前拉取近 90 天使用数据并向使用方确认；有活跃使用须使用方 owner 签字，无则按计划随下线消失 |
| 401 切换后控制台成为登录链路强依赖（静态站不可用 = 全员无法登录） | L | H | 控制台为独立静态资源，与 e-cam-web 同一 nginx 运维面；登录页零 eiam 依赖、故障时可渲染错误提示；nginx 保留 401 目标一键回切 ecmdb-web 登录页的配置开关（ecmdb-web 下线前有效） |
| 新登录页扩大钓鱼/凭据攻击面 | L | H | 登录页仅作 eiam 流程 UI 壳，凭据校验全在 eiam 侧；CSP/X-Frame-Options/HSTS 由 nginx 注入、不引入第三方脚本（见 NFR 登录页安全面） |

## Success Criteria

- [ ] 控制台可完成密码登录与登出（真实浏览器全闭环）；LDAP/passkey/MFA 沿用 eiam 能力并补 UI 入口。验证口径分两档：密码流程全闭环，LDAP/passkey/MFA 各至少一次真实浏览器冒烟；eiam 环境暂不具备的身份源登记为待验项，不计入通过。待验项退出机制：由提案作者在 Phase 0 结束时登记入缺口处置单，v1 验收前必须清零——补齐环境完成冒烟，或触发范围重议；LDAP/passkey/MFA 属 In Scope 交付物，不得经处置门「放弃能力」退出（该处置仅适用于清单内非 parity 域与清单外能力），不允许无声缩水或无期限存续 
- [ ] 平台用户/角色/租户/授权可增删改查：Phase 0 盘点清单中属 parity 承诺的项（iam 管理 CRUD）一律以实现或「扩 eiam 端点后实现」闭环，**不适用「UI 降级/放弃能力」处置**——这两个处置分支仅适用于清单内非 parity 域与清单外能力；验收基线为「经处置后保留的能力集，逐项不低于原 ecmdb-web 的 iam 管理页对应能力」
- [ ] 控制台顶部导航可单点跳转 e-cam-web，共享会话、不重复登录
- [ ] 双向单点登出：控制台登出后，e-cam-web 已登录会话的下次请求即失效并跳回控制台登录页；e-cam-web 登出后控制台会话同样失效；经真实浏览器双标签页闭环验证 
- [ ] 登录/登出与 SSO 跳转经真实浏览器（Playwright 或浏览器 MCP）验证通过，而非仅接口级断言；验证宿主须与生产一致（同域 nginx 反代路径），独立端口 dev server 的通过不作为判定依据 
- [ ] 工作台可展示当前用户信息（来自 eiam）与应用快捷导航，且可跳转
- [ ] e-cam-web 迁出平台身份 UI 后云管功能无回退：按走查清单逐页核验（登录、SSO 跳入跳出、云 IAM 五个子域各一次增删改查、401 过期跳转），走查清单随迁出 PR 提交
- [ ] e-cam-web 会话过期/401 路径独立验证（区别于主动登出链路）：过期后任意请求 → 跳转控制台登录页 → 重新登录后回到原操作路径；真实浏览器验证
- [ ] 无权限操作（403）与 API 失败经统一承载（API 拦截器 + 路由级错误页）显示明确提示：403 显示无权限提示、网络错误可重试、不白屏；真实浏览器验证
- [ ] 性能 NFR 达标：同域 nginx 宿主下实测登录页首屏静态资源 gzip 后 ≤ 300 KB（附构建产物体积报告）、LCP ≤ 2.5 s、二次访问 LCP ≤ 1 s（Playwright 或浏览器性能测量）
- [ ] eiam 不可用降级验证：阻断 eiam 可达性后，登录与数据操作呈现明确错误提示（含重试指引），页面不白屏、不无限 loading，已登录的 e-cam-web 会话不受影响；真实浏览器验证
- [ ] 下线就绪：Phase 0 端点清单逐项处置完成（parity 项全实现、清单外项处置决策全部落地登记），且迁出回归与 401 切换验证通过 → 产出「ecmdb-web 前端下线就绪清单」，平台级入口（login/iam/navigation）具备退役条件
- [ ] 全链路除 eiam 外不新增任何后端服务（部署清单可核验）

## Next Steps

- Proceed to `/write-prd` to formalize requirements