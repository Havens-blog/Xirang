---
created: "2026-10-09"
author: "Haven"
status: Draft
intent: "new-feature"
---

# Proposal: WatchAlert 作为子项目并入运维平台（含 eiam 统一鉴权）

## Problem

告警能力现状是「反代·独立服务」接入（决策 0002），WatchAlert 后端与前端以独立目录/独立进程运行在 `D:\WatchAlert` / `D:\WatchAlert-web`，未被 Haven workspace 的编排（dev.ps1 / Justfile / forge 质量门禁）纳入，且登录仍是 WatchAlert 自身的本地 JWT，与运维平台的身份体系（eiam）割裂。

### Evidence

- `docs/decisions/0002-watchalert-integration.md`（2026-09-24）采用「方案 A 反代接入」：源码不进 workspace、独立进程、nginx 8891 反代；并把 **SSO 免密单点列为开放项**。
- `dev.ps1` 服务注册表目前只登记 e-cam-service / eiam / e-cam-web，**没有 watch-alert**；WatchAlert 需手动 `go build -o watchalert.exe ./cmd` + 前端 `yarn start` 启动（0002 已记录此缺口）。
- `Justfile` 无 `watch-alert-*` recipe，`.forge/config.yaml` 的 `surfaces` 无 watch-alert —— WatchAlert 完全脱离平台的统一测试/质量门禁。
- D:\Haven 是多仓库 workspace（eiam / e-cam-service / e-cam-web / haven-console / opsagent 各自独立 .git），子项目以「workspace 下独立仓库 + dev.ps1/Justfile/forge 登记」的形式存在；WatchAlert 尚未按此形态纳入。
- WatchAlert 自身有三种登录（本地 JWT / LDAP / OIDC），运维用户登录告警平台要单独一套账号，与 eiam 管理的账号/组织脱节。

### Urgency

WatchAlert 已承载真实告警（CDN/ECS 指标、告警规则、故障中心），且已向上游 fork 提交多轮 PR（#257/#263/#264 等）。现状代价：每次改动要跨两个独立目录手动启动/编译；告警平台账号与平台账号不一致（安全与审计隐患）；质量门禁覆盖不到这个正在变重要的子系统。量化上，当前 fork 与上游已分叉约 15-20 个 commit（含已提 PR 与本地适配），按近期节奏每延迟一月预计新增 2-4 个分叉 commit，迁移时需 rebase 的冲突面线性增长，估算每延迟一月迁移成本增加约 4-8 人时。**待作者确认**：当前已纳管告警的 CDN 域名数 / ECS 实例数 / 日均告警规则触发量 / 登录告警平台的用户数与角色分布——这些数据用于在 PRD 阶段量化「迁移期间影响面」（被影响的告警对象规模、受影响用户会话数），作者需确认具体口径后补入；未确认前不臆造数值。

## Proposed Solution

把两个 fork 直接 clone 到 workspace 下，成为与 eiam/e-cam-service 同级的**子项目仓库**（`D:\WatchAlert` → `Haven/watch-alert` 后端、`D:\WatchAlert-web` → `Haven/watch-alert-web` 前端）；每个子项目自带完整 `.git` 历史与 `mine` 远端，天然保留向上游 fork 提 PR 的能力。同时补齐 Haven 子项目「四件套」（子目录 + dev.ps1 服务表 + Justfile recipe + forge surface），并把登录鉴权委托给 eiam 统一认证，去除 WatchAlert 本地 JWT 登录入口。

### Innovation Highlights

- 关键洞察：D:\Haven 本来就是**多仓库 workspace**，无需「subtree 迁入 monorepo」这类假设存在 monorepo 的重操作——**fork 本身就是 git 仓库，直接 clone 下沉为同级子项目**即可，历史与 `mine` remote 零成本保留，提 PR 就是 `git push mine`，比 subtree 的 push/pull 纪律更简单。
- 鉴权侧选定**优先路径 + 回退决策矩阵**，而非三岔口：**优先**复用 WatchAlert 内置 OIDC 直连 eiam（若 eiam 暴露 OIDC discovery + /userinfo 端点）；**回退 1** 走 WatchAlert 内置 LDAP 直连 eiam 的 LDAP 后端（若 eiam LDAP 网络可达）；**回退 2** 由 eiam 新增一个给下游应用的 token 校验端点，WatchAlert 经可插拔鉴权接口调用（桥接）。三条通道按 tech-design 阶段对 eiam 端点的 spike 验证结果**依序裁决**（OIDC → LDAP → 桥接），同一时刻只采用一条；裁决依据与所需 eiam 最小改动列入 Next Steps 的收敛矩阵交付物。**「通道未最终选定」是刻意的**：三条通道的可行性取决于 eiam 现有端点是否暴露 OIDC discovery / LDAP 网络可达 / 是否需新建桥接端点，这些只有 spike 能裁决；提案不预设答案，避免在未验证假设上锁定协议。PRD/tech-design 交接时显式声明：PRD 负责定义鉴权的用户可见行为（SSO 登录、降级 UX、会话语义），tech-design 负责锁定具体通道与协议；交接物为「鉴权对接协议收敛矩阵」，PRD 在通道未锁定前不描述具体鉴权协议字段。

## Requirements Analysis

### Key Scenarios

- **开发启动**：`dev.ps1 start watch-alert` 一键拉起后端(:9001) + 前端(:3001)，`status`/`logs` 可观测；编译失败旧进程继续跑（沿用现有 Daemon 语义），但 `dev.ps1 status` 须对 watch-alert 编译失败给出显眼标记（区别于普通业务服务静默续跑），防止「规则修复假生效」。
- **SSO 登录**：运维用户在浏览器打开告警平台，凭 eiam 账号（或其背后 LDAP/OIDC/飞书/企微身份源）登录，无需 WatchAlert 独立账号。
- **质量门禁**：`forge quality-gate`（或 `just watch-alert-*`）对后端跑 compile/fmt/lint/unit-test，前端跑 build/test。
- **上游贡献**：在 `Haven/watch-alert`（fork 的 clone）修通用 bug → `git push mine` → fork 提 PR → 合并后 `git pull origin/mine` 同步。
- **迁移过渡（切流）**：旧 `D:\WatchAlert` 仍在承载真实告警期间，新 `Haven/watch-alert` 并行运行但不接生产流量；切流前完成 eiam 登录验证（SC-6）与回滚 runbook 演练。切流流程：停旧实例写入口 → 同步告警规则/数据源配置到新实例（配置为文件/DB 导出导入，非历史事件迁移）→ nginx 反代指向新实例 → 观察 1 个告警周期确认无遗漏 → 旧实例保留只读 2 周后退役。切流期间 live 告警由旧实例继续投递，新实例首次仅承接读流量，确认告警链路（规则→数据源→通知）全链路正常后再承接写。退役标准：新实例连续 2 周零告警丢失/误报、eiam 登录稳定、回滚 runbook 已演练。
- **边界/失败场景**：子项目仓库与上游个人 fork 分叉、鉴权服务(eiam)不可用时的降级（方向见下）、Windows 下 CompileDaemon 对 `cmd/` 结构的编译命令不支持。

**eiam 降级方向**（提案层面选定「token 缓存宽限」为降级主路径）—— 区分**延迟劣化**与**不可用**两级：

- **延迟劣化**（eiam 端点 P95 校验延迟 > 300ms 但仍可返回有效校验结果）：进入**短时宽限放行**态，已签发 token 继续放行、写操作继续允许（不触发只读），宽限窗口 15 分钟（由 tech-design 校准），窗口内探针轮询 eiam 延迟；延迟回落阈值以下连续 2 分钟即退出宽限恢复常态。此级目的是吸收 eiam 瞬时慢请求（ GC、网络抖动），避免 350ms 触发 4h 只读的失衡。
- **不可用**（eiam 端点连续探针失败达 3 次 / 60 秒内无有效响应）：进入**长时宽限 + 只读**态。已签发 token 在本地缓存并宽限有效期（默认宽限窗口 4 小时，由 tech-design 校准），不回退到本地 JWT 兜底登录。**宽限期内 UX**：页面顶部常驻告警条「鉴权服务暂不可用，已切换宽限模式（token 剩余 X 小时）」；允许的操作为告警列表/规则/历史事件**只读查看**；**禁止**的操作为编辑规则、确认/静默告警、修改通知配置、任何写操作（后端对写接口在宽限态强制返回 503）。**过期后**降级为只读免登录视图：展示告警看板 + 规则只读 + 事件历史，无操作按钮。**新/未登录用户**（无缓存 token 且本地 JWT 已移除）：直接落到该只读免登录视图，并提示「请联系管理员或等待 eiam 恢复」，不提供任何登录入口（避免重新引入被移除的本地登录路径）。
- **降级退出**（eiam 恢复后回正）：WatchAlert 周期性向 eiam 发**恢复探针**（间隔 1 分钟，校验一个 dummy token 或 health 端点），连续 3 次成功即判定 eiam 恢复；恢复后自动退出宽限态，已缓存 token 继续有效直至自然过期、新登录请求改走 eiam 实时校验；退出时发一条「鉴权已恢复正常」的告警（由 opsagent 承载，避免 WatchAlert 自告自警），并记录降级持续时长用于事后复盘。宽限窗口、只读边界、禁用清单、恢复探针参数由 tech-design 最终确定。

**鉴权切流时的活跃会话迁移**（JWT 移除瞬间）：本地 JWT 移除前已签发的 JWT 不会瞬时失效，采用**过渡期双通道并存**策略——在 SC-6（eiam 登录验证通过）后、SC-7（JWT 移除）前的过渡窗口（建议 1 个告警周期，约 tech-design 校准），WatchAlert 同时接受本地 JWT 与 eiam 下发的新 token；过渡期结束后移除本地 JWT handler，旧 JWT 自然过期；过渡期内强制已登录用户在下次写操作时重走 eiam 登录（写接口要求 eiam token，读接口容忍旧 JWT 至过期），避免强制全员重登的同时保证写路径鉴权已切到 eiam。

### Non-Functional Requirements

- **安全**：鉴权改造后不得在 WatchAlert 侧残留明文凭据/绕过本地登录的洞；账号审计走平台（eiam）体系。落地：eiam 下发的 token 仅存内存（进程内 cache，TTL 对齐宽限窗口），**不落盘、不明文缓存**；可插拔鉴权接口限定为「校验回调」语义（接口只接收 token、返回用户身份，不接触原始口令/密钥），适配层与 eiam 的调用走平台内网 + mTLS，凭据不进入 WatchAlert 配置或日志。
- **性能**：鉴权委托的登录校验延迟阈值 ≤ 300ms（eiam 端点 P95）。触发语义与降级行为对齐（见降级方向）：**延迟劣化**（P95 > 300ms 但仍返回有效校验）触发短时宽限放行（写操作继续，宽限 15min）；**不可用**（连续探针失败）触发长时宽限 + 只读（写请求直接 503，宽限 4h）。两级的触发条件不同，避免用同一重降级响应不同严重度。
- **可观测性**：eiam 集成在生产侧监控——鉴权成功率、token 校验延迟 P95/P99、降级触发次数与持续时间（区分延迟劣化级与不可用级）接入 eiam/opsagent 现有指标体系；降级触发即告警（告警本身由 opsagent 承载，避免 WatchAlert 自告自警）。**导出机制**：WatchAlert 通过 `/metrics` 端点暴露 Prometheus 文本格式指标（`watchalert_auth_check_total{result=success|fail}`、`watchalert_auth_check_duration_seconds` 直方图、`watchalert_degradation_active{level=latency|unavailable}`），由 opsagent 既有 Prometheus scrape 配置抓取；若 eiam 桥接通道采用，桥接适配层复用同一端点，不另建导出通道。
- **兼容性**：迁入后后端 `go build`、前端生产 build 必须通过；现有告警规则/数据源/通知不回归。

### Constraints & Dependencies

- 依赖 eiam 的认证能力（HTTP Gin API + gRPC；系统登录/LDAP/OIDC 登录已具备，对接协议需 tech-design 明确，可能需要 eiam 侧最小改动）。
- Windows 开发环境：WatchAlert 后端入口在 `cmd/`（非根 `main.go`），与 dev.ps1 Daemon 当前 `go build -o $exe .` 约定不符；dev.ps1 服务表当前版本**不存在**「自定义 build 命令字段」，需作为显式前置开发任务新增。
- 两个 fork 远端：后端 `Havens-blog/e-watch-alert`、前端 `Havens-blog/WatchAlert-web`。

## Alternatives & Industry Benchmarking

### Industry Solutions

- 业界对「外部开源项目 + 内部魔改」的收纳：git submodule、git subtree、独立 fork + clone。多仓库 workspace 形态下，主流是**直接把 fork clone 为 workspace 子项目**（保留上游 remote，PR 直推）。可比范式：**umbrella repo**（顶层仅含编排/文档，子项目各自独立仓库，如很多 CNCF 项目集合）、**git subtrees 目录约定**（monorepo 内以目录隔离外部依赖，push/pull 纪律重）、**Bazel / please 等单体仓库工具**（强假设单一仓库根、跨仓依赖图集中管理）。本提案的 D:\Haven 形态最接近 umbrella repo + 各子仓独立 .git，故选 fork clone（与 eiam/opsagent/e-cam-service 既有形态同构），而非引入 Bazel 这类假设存在单一根仓库的工具。
- 鉴权侧主流是「应用委托统一 IdP」（OIDC/LDAP），与 eiam 作为平台 IAM 的定位一致。具体参考：**Grafana** 内置 generic OAuth2/OIDC provider，通过 `auth.generic_oauth` 配置直连外部 IdP，不在 Grafana 侧另造协议；**Alertmanager** 本身无内置鉴权，业界惯例是反代层（nginx/oauth2-proxy）+ 平台 IAM 网关收口身份，不在告警组件内做登录。本提案的「复用 WatchAlert 内置 OIDC + eiam 作 IdP」与 Grafana 模式一致，「桥接 token 校验端点」与 Alertmanager 的反代网关模式同源。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing（0002 反代现状） | 现状 | 零成本 | 无统一编排/质量门禁；账号割裂；技术债累积 | Rejected: 用户已明确推翻 |
| 反代 + 只补编排 | 0002 + 增量 | 改动最小 | 仍是两处源码、两套生命周期，非"子项目" | Rejected: 用户要求源码并入 |
| git submodule | git 标准 | 独立 .git | detached HEAD/递归 clone/双提交，日常坑多 | Rejected: 维护负担高 |
| git subtree | git 标准 | 内联历史 | 假设存在 monorepo；D:\Haven 是 workspace 无此前提 | Rejected: 前提不成立 |
| **clone fork 作子项目仓库** | git clone | 与 eiam/e-cam-service 同构；历史+remote 零成本；PR 直推 | 鉴权对接复杂度需 tech-design 收敛（见决策矩阵）；dev.ps1 自定义 build 命令字段当前缺失需新增；需持续向上游提通用 PR 以规避死 fork | **Selected: 与现有工作区形态一致且最简** |

## Feasibility Assessment

### Technical Feasibility

`git clone` + dev.ps1/Justfile/forge 登记都是 Haven 已有成熟范式（eiam/opsagent 已在用）。eiam 已具备登录认证能力，WatchAlert 已内置 OIDC/LDAP 登录 —— 两端能力就绪，缺的是对接协议设计。整体技术可行。

### Resource & Timeline

单人推进，分四阶段（工作量估算为单人日，含调试缓冲）：
- **P1 阶段化编排纳入**（约 2-3 人日）：clone fork + dev.ps1 服务表（含自定义 build 字段新增）+ Justfile/forge 登记 + nginx dev 入口。
- **P2 鉴权对接 spike**（约 3-5 人日）：对 eiam 端点做 OIDC/LDAP/桥接三通道 spike，据结果裁决通道、锁定协议。**若裁决为桥接通道**（需 eiam 新增给下游应用的 token 校验端点），eiam 侧工作量估算约 2-3 人日（端点设计 + 实现 + 对 WatchAlert 的联调），视为**本提案内部依赖**（按 Scope 判定线归入本提案），计入 P2/P3 总工期；eiam 侧开发若由不同人承担，需在 spike 裁决后立即与 eiam 维护者对齐排期，否则 P3 阻塞。
- **P3 鉴权改造与 JWT 移除**（约 4-6 人日）：可插拔鉴权接口 + 适配层 + 本地 JWT handler/密钥/前端入口移除。
- **P4 切流与退役**（约 2-3 人日 + 2 周观察期）：并行运行、配置迁移、nginx 切流、旧实例退役。
风险集中在 P2 的 eiam 端点不确定性，属 tech-design 可收敛的技术不确定性，非资源瓶颈。

### Dependency Readiness

- **eiam 认证 API：未就绪**。需确认是否有可复用的登录校验端点、是否需扩一个给下游应用用的验证接口；按 Scope 判定线，该最小改动可能归入本提案，功能性增强另立 feature。此为关键依赖，阻塞 P2/P3。
- **dev.ps1 自定义 build 命令字段：未就绪**。当前版本不存在该字段，需作为 P1 显式前置开发任务新增，影响 `cmd/` 结构的后端编译。
- 两个 fork 远端已就绪（`Havens-blog/e-watch-alert`、`Havens-blog/WatchAlert-web`）。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "合并子项目 = 迁入 git monorepo" | XY Detection | Overturned：D:\Haven 是多仓 workspace，子项目 = workspace 下独立仓库 + 编排登记 |
| "源码进了 workspace 就保不住向上游提 PR" | Assumption Flip | Overturned：fork 直接 clone 下沉，`.git` + `mine` remote 自带，PR 即 `git push mine` |
| "eiam 不直接向下游下发 OIDC，所以 SSO 难做" | Stress Test | Refined（改为待验证假设）：eiam 作统一认证源，WatchAlert 复用已有 OIDC/LDAP 登录对接之；「不要求 eiam 成为完整 OIDC IdP」是预期但需 tech-design 验证，若 eiam 不暴露 OIDC 兼容端点则改走 LDAP 直连或新增桥接（见 tech-design 交付物「鉴权对接协议收敛矩阵」） |
| "鉴权只是配置项，不改 WatchAlert 代码" | 5 Whys | Refined：改动分两类——①可提上游的通用改动（鉴权接口抽象、配置化），按「通用功能可提 PR」纪律同步上游；②平台专属适配层（eiam 对接代码）通过 WatchAlert 内可插拔鉴权接口隔离，不硬编码 eiam 调用，留在 Haven 本地不分叉上游 |

## Scope

### In Scope

- 将 fork clone 到 workspace：后端 `D:\WatchAlert` → `Haven/watch-alert`、前端 `D:\WatchAlert-web` → `Haven/watch-alert-web`（各自独立仓库，保留 `.git` + `mine` remote）
- dev.ps1 服务表登记 watch-alert（后端 :9001）、watch-alert-web（前端 :3001），含对 `cmd/` 结构的自定义 build 命令支持
- Justfile 增加 `watch-alert-{compile,fmt,lint,unit-test,regression,test}` recipe
- `.forge/config.yaml` `surfaces` 新增 watch-alert（api）/ watch-alert-web（web）
- 鉴权统一到 eiam：WatchAlert 登录委托 eiam 认证，移除本地 JWT 登录入口（删除后端登录 handler、清除签名密钥配置、删除前端登录入口）
- 更新/作废决策 0002，新增子项目接入决策记录，内容含：①选型理由（fork clone 作子项目 vs submodule/subtree 的否决依据）；②子项目仓库与远端约定（`mine` remote 指向个人 fork、`origin` 指向上游、push 纪律）；③鉴权通道裁决结果（P2 spike 后锁定 OIDC/LDAP/桥接之一及 eiam 端点清单）；④与 eiam 的边界（哪些改动归本提案、哪些属 eiam 独立 feature）；⑤dev.ps1 自定义 build 字段的引入记录
- **迁移/切流/退役交付物（P4）纳入 In Scope**：①切流 runbook（停旧写→同步配置→nginx 切流→观察 1 周期的逐步操作清单与回滚触发条件）；②回滚 runbook（已在 Rollback & Upstream 章节描述，正式纳入 In Scope 作为交付物）；③退役 checklist（连续 2 周零告警丢失/误报、eiam 登录稳定、回滚 runbook 已演练 三项退役标准 + 旧实例只读保留期与下线步骤）
- 鉴权降级设计文档（延迟劣化级/不可用级两级触发条件、宽限窗口、只读边界、恢复探针、会话过渡策略）作为 tech-design 交付物
- nginx dev 入口调整：将 8891 反代的 `proxy_pass` 从旧 `D:\WatchAlert`(:9001 旧实例) 指向 workspace 新实例 `127.0.0.1:9001`，新增前端 dev 入口 `:3001` 反代条目；「dev.ps1 直接起」指 dev.ps1 在 `start watch-alert` 时同时拉起新实例进程并验证 nginx 反代可达（`status` 显示反代健康），消除当前「手动 go build + yarn start + 手动对 nginx」的流程。具体 nginx 配置 diff 列入 tech-design。
- dev.ps1 status 对 watch-alert 编译失败须给出显眼标记（如红色/COMPILE-FAILED 标签），区别于普通业务服务静默续跑

### Out of Scope

- 告警规则/事件/故障中心的历史数据迁移
- WatchAlert 深度内核改造（除鉴权衔接所需的最小通用改动外）
- 前端页面嵌入 haven-console 布局（保持独立前端）
- eiam 本体功能性/能力增强

**Scope 判定线**（消除 Dependency Readiness 与 Out of Scope 的矛盾）：eiam 为「下游应用鉴权」所需的最小改动——即桥接通道所需的 token 校验端点（`POST /api/v1/token/verify`，接收 token 返回用户身份）——**归入本提案内建**（若 P2 spike 裁决采用桥接通道，则该端点的实现属 P2/P3 交付物，eiam 侧工作量计入工期，见 Resource & Timeline）。该端点在 spike 裁决前为「待验证的外部依赖未就绪」，裁决后转为「本提案内建」并进入 In Scope。eiam 自身的能力增强（如成为完整 OIDC IdP、新增身份源、组织树改造）另立 feature，属 Out of Scope。此判定线消除 Dependency Readiness 中「可能归入本提案」的含糊：端点要么 spike 否决（不建），要么裁决通过（本提案建）。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 子项目 clone 与上游个人 fork 分叉 → 忘同步/丢改动 | M | H | 文档化「改动前先 pull 上游、通用修复直推 fork」纪律，dev.ps1 不吞 `.git`；**自动化护栏**：在 `Haven/watch-alert` 加 pre-push hook（或 CI 检查）校验本地 `mine` remote 存在且 `git log origin..mine` 的未推送 commit 数 ≤ 阈值，超阈则阻断 push 并提示先同步上游 |
| eiam 认证对接协议不明，可能需 eiam 侧改动 | M | H | tech-design 阶段先摸清 eiam 认证端点，锁定协议再动手 |
| eiam 不可用导致告警平台登录中断 | M | H | 选定「token 缓存宽限 + 只读免登录」降级路径，不回退本地 JWT；宽限窗口由 tech-design 定 |
| dev.ps1 Daemon 不支持 `cmd/` 编译结构（自定义 build 命令字段当前不存在，需新增） | H | M | 作为显式前置任务在 dev.ps1 服务表新增自定义 build 命令字段；非「已解决」，0002 仅预见未实现 |
| 鉴权改造波及 WatchAlert 通用代码，死 fork 风险 | M | M | 拆分「可提上游通用改动」与「平台专属适配层」：前者提 PR 上游，后者通过可插拔鉴权接口隔离在 Haven 本地，不硬编码 eiam |
| 前端 labels/构建（React CRA）与 Haven 前端门禁不一致 | M | L | 前端沿用自身 build/test 方式接 forge surface，不强行套 vue 门禁 |
| watch-alert 编译失败但旧进程继续跑，造成规则修复假生效 | M | M | dev.ps1 status 对 watch-alert 编译失败显眼标记，区别于普通服务静默续跑 |
| 上游 WatchAlert 改动鉴权接口契约，打破本地可插拔适配层 | L | M | 版本 pinning：track 上游具体 tag 而非 main HEAD；rebase 前先跑鉴权接口契约回归用例；上游发版节奏分析见 Rollback & Upstream 章节 |
| 切流期间迁移失败需回退到旧实例 | L | H | 回滚 runbook 见 Rollback & Upstream 章节 |
| 鉴权统一根本失败（OIDC/LDAP/桥接三通道 spike 全部证伪，eiam 无法以任何方式支持下游鉴权） | L | H | 作为显式风险入表：触发回滚 runbook（恢复本地 JWT），鉴权统一转为后续独立 feature 重新立项；提案主体（workspace 编排纳入 + 四件套）不阻断，单独交付；本风险在 P2 spike 结束时即可判定，不拖到 P3 |

## Rollback & Upstream Cadence

### 回滚 Runbook（基础设施迁移）

**触发条件**（满足任一即回滚）——指标口径已定义：新实例登录链路连续失败率 > 5%（**测量口径**：滚动 5 分钟窗口内 eiam 登录校验返回失败/超时数 ÷ 登录请求总数，由 WatchAlert `/metrics` 的 `watchalert_auth_check_total{result=fail}` 与总数计算，opsagent 聚合；连续 5 分钟窗口均值 > 5% 即触发）；告警投递出现丢失/误报且 1 个告警周期内未恢复；eiam 集成出现无法在宽限窗口内修复的阻断。

**回滚步骤**：
1. **鉴权回退（后端 + 前端双仓）**：后端——重新启用 WatchAlert 本地 JWT（恢复后端登录 handler、重新装入签名密钥），本地 JWT handler 与密钥配置在移除时打 `revert-jwt-backend` tag 保留，回滚即 `git revert` 该 tag 后重新 `go build` 部署后端。前端——登录入口移除在独立仓库 `Haven/watch-alert-web`，移除时同步打 `revert-jwt-frontend` tag，回滚即在该仓库 `git revert` 该 tag → `yarn build` 重新构建 → 重新部署前端静态资源；前端 revert 不依赖后端 tag，双仓可独立回滚。
2. **workspace 回退**：dev.ps1/nginx 反代指回旧 `D:\WatchAlert` 实例（:9001 旧端口），`dev.ps1 start watch-alert` 改回拉起旧实例或直接停用新实例登记。
3. **数据/状态保留**：告警规则、数据源配置、通知配置为文件/DB 导出导入物，回滚不删数据；新实例期间产生的事件历史保留在库，旧实例继续以既有状态服务。
4. **验证**：旧实例登录 + 告警投递链路回归，确认 eiam 侧无残留 token 校验调用。

### 上游发版节奏与 Pinning/Rebasing

- **节奏分析**：WatchAlert 上游（个人 fork 的源头）发版不规律，鉴权相关接口变动属低频但高影响。`cmd/` 结构与鉴权 handler 是 rebasing 时的高风险面。
- **Pinning 策略**：`Haven/watch-alert` 锁定上游某个稳定 tag/commit（非 main HEAD），新实例部署基于该 pin；升级时显式 rebase。
- **Rebasing 计划**：rebase 前必跑「鉴权接口契约回归用例」（SC-6 + 可插拔鉴权接口的单元用例）；若上游改动鉴权接口契约，先更新本地的可插拔适配层（不触碰 eiam 专属代码），通过后再合入。鉴权接口契约变更优先作为「可提上游通用改动」提 PR，减少本地分叉。

## Success Criteria

- [ ] `Haven/watch-alert`、`Haven/watch-alert-web` 为独立仓库且 `git log` 可追溯上游 fork 提交历史、`git remote` 含 `mine`
- [ ] `git push mine` 能直接提交 PR 回 fork（后端+前端各验证一次），无"死 fork"
- [ ] `dev.ps1 start watch-alert` 一键拉起后端(:9001)+前端(:3001)，`status`/`logs` 可见
- [ ] `just watch-alert-compile`、`just watch-alert-unit-test` 存在且通过（后端 `go build`、前端 production build 通过）
- [ ] `.forge/config.yaml` surfaces 新增 watch-alert，`forge quality-gate` 覆盖其 compile/fmt/lint/unit-test
- [ ] nginx dev 入口：`dev.ps1 start watch-alert` 后 `:9001`（后端）/`:3001`（前端）经 nginx 反代可达，`dev.ps1 status` 显示反代健康
- [ ] `dev.ps1 status` 对 watch-alert 编译失败显示显眼标记（红色/COMPILE-FAILED），区别于普通服务静默续跑
- [ ] **SC-6（门控前置）**：用 eiam 账号登录 WatchAlert 成功（SSO），且鉴权委托延迟 P95 ≤ 300ms
- [ ] **SC-7（须在 SC-6 验证通过后执行）**：旧用户名/密码 POST 本地 JWT 登录端点返回 401/404，不得签发 token；前端无本地登录入口残留。门控目的：避免在唯一登录路径（eiam）验证成功前砍掉本地 JWT，导致系统无登录路径
- [ ] 决策 0002 标记作废，新增子项目接入决策记录（含 fork 同步/PR 约定）
- [ ] **SC-8（切流后告警链路完整性）**：切流完成后 1 个告警周期内，新实例「规则触发 → 数据源查询 → 通知投递」全链路无丢失/误报；退役前连续 2 周零告警丢失/误报、eiam 登录稳定、回滚 runbook 已演练

## Next Steps

- Proceed to `/write-prd` to formalize requirements（全流程，非 quick）
- tech-design 阶段交付「鉴权对接协议收敛矩阵」：逐条配对 OIDC 通道 / LDAP 直连通道 与 eiam 现有能力（系统登录、LDAP、OIDC 登录、下游验证接口），标注每条通道的 eiam 端点就绪状态与所需最小改动，据以收敛「不要求 eiam 成为完整 OIDC IdP」假设
- Spike（PRD 前验证）：评估 dev.ps1 服务表新增「自定义 build 命令字段」的改动面，含对 eiam/e-cam-service 现有编译回归的影响，确认不破坏既有服务启动