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

WatchAlert 已承载真实告警（CDN/ECS 指标、告警规则、故障中心），且已向上游 fork 提交多轮 PR（#257/#263/#264 等）。现状代价：每次改动要跨两个独立目录手动启动/编译；告警平台账号与平台账号不一致（安全与审计隐患）；质量门禁覆盖不到这个正在变重要的子系统。愈晚迁移，与上游 fork 的历史分叉愈深、迁移成本愈高。

## Proposed Solution

把两个 fork 直接 clone 到 workspace 下，成为与 eiam/e-cam-service 同级的**子项目仓库**（`D:\WatchAlert` → `Haven/watch-alert` 后端、`D:\WatchAlert-web` → `Haven/watch-alert-web` 前端）；每个子项目自带完整 `.git` 历史与 `mine` 远端，天然保留向上游 fork 提 PR 的能力。同时补齐 Haven 子项目「四件套」（子目录 + dev.ps1 服务表 + Justfile recipe + forge surface），并把登录鉴权委托给 eiam 统一认证，去除 WatchAlert 本地 JWT 登录入口。

### Innovation Highlights

- 关键洞察：D:\Haven 本来就是**多仓库 workspace**，无需「subtree 迁入 monorepo」这类假设存在 monorepo 的重操作——**fork 本身就是 git 仓库，直接 clone 下沉为同级子项目**即可，历史与 `mine` remote 零成本保留，提 PR 就是 `git push mine`，比 subtree 的 push/pull 纪律更简单。
- 鉴权侧复用 WatchAlert 已内置的 OIDC/LDAP 登录能力，对接 eiam 作为统一认证源，不另造鉴权协议。

## Requirements Analysis

### Key Scenarios

- **开发启动**：`dev.ps1 start watch-alert` 一键拉起后端(:9001) + 前端(:3001)，`status`/`logs` 可观测；编译失败旧进程继续跑（沿用现有 Daemon 语义）。
- **SSO 登录**：运维用户在浏览器打开告警平台，凭 eiam 账号（或其背后 LDAP/OIDC/飞书/企微身份源）登录，无需 WatchAlert 独立账号。
- **质量门禁**：`forge quality-gate`（或 `just watch-alert-*`）对后端跑 compile/fmt/lint/unit-test，前端跑 build/test。
- **上游贡献**：在 `Haven/watch-alert`（fork 的 clone）修通用 bug → `git push mine` → fork 提 PR → 合并后 `git pull origin/mine` 同步。
- **边界/失败场景**：子项目仓库与上游个人 fork 分叉、鉴权服务(eiam)不可用时的降级、Windows 下 CompileDaemon 对 `cmd/` 结构的编译命令不支持。

### Non-Functional Requirements

- **安全**：鉴权改造后不得在 WatchAlert 侧残留明文凭据/绕过本地登录的洞；账号审计走平台（eiam）体系。
- **兼容性**：迁入后后端 `go build`、前端生产 build 必须通过；现有告警规则/数据源/通知不回归。

### Constraints & Dependencies

- 依赖 eiam 的认证能力（HTTP Gin API + gRPC；系统登录/LDAP/OIDC 登录已具备，对接协议需 tech-design 明确，可能需要 eiam 侧最小改动）。
- Windows 开发环境：WatchAlert 后端入口在 `cmd/`（非根 `main.go`），与 dev.ps1 Daemon 的 `go build -o $exe .` 约定不符，需服务表支持自定义 build 命令。
- 两个 fork 远端：后端 `Havens-blog/e-watch-alert`、前端 `Havens-blog/WatchAlert-web`。

## Alternatives & Industry Benchmarking

### Industry Solutions

- 业界对「外部开源项目 + 内部魔改」的收纳：git submodule、git subtree、独立 fork + clone。多仓库 workspace 形态下，主流是**直接把 fork clone 为 workspace 子项目**（保留上游 remote，PR 直推）。
- 鉴权侧主流是「应用委托统一 IdP」（OIDC/LDAP），与 eiam 作为平台 IAM 的定位一致。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing（0002 反代现状） | 现状 | 零成本 | 无统一编排/质量门禁；账号割裂；技术债累积 | Rejected: 用户已明确推翻 |
| 反代 + 只补编排 | 0002 + 增量 | 改动最小 | 仍是两处源码、两套生命周期，非"子项目" | Rejected: 用户要求源码并入 |
| git submodule | git 标准 | 独立 .git | detached HEAD/递归 clone/双提交，日常坑多 | Rejected: 维护负担高 |
| git subtree | git 标准 | 内联历史 | 假设存在 monorepo；D:\Haven 是 workspace 无此前提 | Rejected: 前提不成立 |
| **clone fork 作子项目仓库** | git clone | 与 eiam/e-cam-service 同构；历史+remote 零成本；PR 直推 | 需遵守向 fork 同步纪律 | **Selected: 与现有工作区形态一致且最简** |

## Feasibility Assessment

### Technical Feasibility

`git clone` + dev.ps1/Justfile/forge 登记都是 Haven 已有成熟范式（eiam/opsagent 已在用）。eiam 已具备登录认证能力，WatchAlert 已内置 OIDC/LDAP 登录 —— 两端能力就绪，缺的是对接协议设计。整体技术可行。

### Resource & Timeline

单人可完成，分阶段（见 Next Steps）。风险集中在「eiam 保真对接」需要先摸清 eiam 认证端点协议，属 tech-design 阶段可收敛的技术不确定性，非资源瓶颈。

### Dependency Readiness

- eiam 认证 API 需确认（是否有可复用的登录校验端点、是否需扩一个给下游应用用的验证接口）。
- 两个 fork 远端已就绪（`Havens-blog/e-watch-alert`、`Havens-blog/WatchAlert-web`）。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "合并子项目 = 迁入 git monorepo" | XY Detection | Overturned：D:\Haven 是多仓 workspace，子项目 = workspace 下独立仓库 + 编排登记 |
| "源码进了 workspace 就保不住向上游提 PR" | Assumption Flip | Overturned：fork 直接 clone 下沉，`.git` + `mine` remote 自带，PR 即 `git push mine` |
| "eiam 不直接向下游下发 OIDC，所以 SSO 难做" | Stress Test | Refined：eiam 作统一认证源，WatchAlert 复用已有 OIDC/LDAP 登录对接之，不要求 eiam 成为完整 OIDC IdP |
| "鉴权只是配置项，不改 WatchAlert 代码" | 5 Whys | Refined：可能需要 WatchAlert 少量通用改动（调用 eiam 认证），需遵守「通用功能可提 PR」纪律 |

## Scope

### In Scope

- 将 fork clone 到 workspace：后端 `D:\WatchAlert` → `Haven/watch-alert`、前端 `D:\WatchAlert-web` → `Haven/watch-alert-web`（各自独立仓库，保留 `.git` + `mine` remote）
- dev.ps1 服务表登记 watch-alert（后端 :9001）、watch-alert-web（前端 :3001），含对 `cmd/` 结构的自定义 build 命令支持
- Justfile 增加 `watch-alert-{compile,fmt,lint,unit-test,regression,test}` recipe
- `.forge/config.yaml` `surfaces` 新增 watch-alert（api）/ watch-alert-web（web）
- 鉴权统一到 eiam：WatchAlert 登录委托 eiam 认证，移除/隐藏本地 JWT 登录入口
- 更新/作废决策 0002，新增子项目接入决策记录（含向 fork 同步/提 PR 的约定）
- nginx dev 入口调整（dev.ps1 直接起，消除手动启动流程）

### Out of Scope

- 告警规则/事件/故障中心的历史数据迁移
- WatchAlert 深度内核改造（除鉴权衔接所需的最小通用改动外）
- 前端页面嵌入 haven-console 布局（保持独立前端）
- eiam 本体功能增强（若对接需要 eiam 改动，超出最小必要的另立 feature）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 子项目 clone 与上游个人 fork 分叉 → 忘同步/丢改动 | M | H | 文档化「改动前先 pull 上游、通用修复直推 fork」纪律，dev.ps1 不吞 `.git` |
| eiam 认证对接协议不明，可能需 eiam 侧改动 | M | H | tech-design 阶段先摸清 eiam 认证端点，锁定协议再动手 |
| dev.ps1 Daemon 不支持 `cmd/` 编译结构 | H | M | 服务表增加自定义 build 命令字段（0002 已预见） |
| 鉴权改造波及 WatchAlert 通用代码，违背"死 fork"风险 | M | M | 鉴权相关改动按「通用功能」开发，同步提 PR 上游 |
| 前端 labels/构建（React CRA）与 Haven 前端门禁不一致 | M | L | 前端沿用自身 build/test 方式接 forge surface，不强行套 vue 门禁 |

## Success Criteria

- [ ] `Haven/watch-alert`、`Haven/watch-alert-web` 为独立仓库且 `git log` 可追溯上游 fork 提交历史、`git remote` 含 `mine`
- [ ] `git push mine` 能直接提交 PR 回 fork（后端+前端各验证一次），无"死 fork"
- [ ] `dev.ps1 start watch-alert` 一键拉起后端(:9001)+前端(:3001)，`status`/`logs` 可见
- [ ] `just watch-alert-compile`、`just watch-alert-unit-test` 存在且通过（后端 `go build`、前端 production build 通过）
- [ ] `.forge/config.yaml` surfaces 新增 watch-alert，`forge quality-gate` 覆盖其 compile/fmt/lint/unit-test
- [ ] 用 eiam 账号登录 WatchAlert 成功（SSO），本地 JWT 登录入口移除/隐藏
- [ ] 决策 0002 标记作废，新增子项目接入决策记录（含 fork 同步/PR 约定）

## Next Steps

- Proceed to `/write-prd` to formalize requirements（全流程，非 quick）