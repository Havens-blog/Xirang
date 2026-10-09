# 决策记录：告警平台 WatchAlert 以「方案 A 反代接入」并入 Haven，保留上游贡献能力

| 项目 | 内容 |
|---|---|
| 状态 | 已决策（2026-09-24），反代接入已完成，SSO 免密为开放项 |
| 决策人 | 项目负责人 |
| 影响范围 | nginx 路由、WatchAlert（D:\WatchAlert + D:\WatchAlert-web）、Haven 门户入口 |

## 1. 背景

现有告警能力来自上游开源项目 WatchAlert（Go 后端 + React 前端），已 fork 到 `Havens-blog/e-watch-alert`，并已向上游提交两处通用修复 PR：

- **#257** `fix(alert): prevent nil pointer panic when RecoverNotify is unset`（恢复事件处理时的 nil 指针崩溃）
- **#258** `chore(config): ignore local config.yaml, ship example file`（本地配置凭据隔离）

需求是把它接入 `D:\Haven` 运维平台，同时**不丢失向上游提 PR 的能力**。二者并非对立，关键在于区分「通用修复」与「Haven 专有改动」。

## 2. 决策

**采用方案 A —— 反代 / 独立服务接入，源码零改动。**

- WatchAlert 后端（:9001）、前端（:3001）**保持独立仓库与独立进程**，不 fork 进 Haven monorepo。
- Haven 的 nginx 增加反代 vhost，将 WatchAlert 暴露为统一入口。
- 通用修复全部走 `D:\WatchAlert`（干净 fork）→ PR 到上游；合并后 `git pull` 拉回。**不留下游补丁。**

不采用方案 B（submodule）/ C（源码拷入 monorepo）——目前对 WatchAlert 的所有改动都是通用修复，无需深改内核，避免平添 rebase 维护负担。若日后确需深改，再升级为 B/C 并存 `patches/` 纪律。

## 3. 实施

- **nginx**：新增 `nginx-1.25.4/conf/vhost/watchalert.conf`
  - 开发入口 `http://localhost:8891`（独立端口，避免与 8888 上 `/api/iam`、`/api/cmdb`、`/api/v1/*` 冲突）。
  - `location /api/` → 后端 `127.0.0.1:9001`；`location /` → 前端 `127.0.0.1:3001`（HMR ws 已配）。
  - 关键点：前端 `src/utils/http.jsx` 的 axios `baseURL` 取当前 `window.location` 的 host:port，因此经 8891 访问时 `/api/*` 自动落到本 vhost，无需改前端。
  - 生产：vhost 内附注释好的 `alert.jlcops.com` 子域名示例（静态 build + `/api` 反代）。
- **nginx.conf**：http 块增加 `map $http_upgrade $connection_upgrade`（HMR websocket 升级映射）。
- **架构文档**：本条决策即最终记录。

## 4. 保持「可提 PR」的机制（关键）

| 改动类型 | 去处 |
|---|---|
| 通用修复/功能（坑全部用户） | `D:\WatchAlert` 干净 fork 上开发 → 推 `mine` → PR 到 `opsre/WatchAlert:master` → 合并后 pull |
| Haven 专有（品牌/账密/布局） | 仅在 Haven 侧做配置或极少量补丁（`Haven/patches/`），绝不进上游 |

约束（避免退化成「再也提不了 PR」的死 fork）：
1. 不复制粘贴源码导致丢失 `.git` 历史；
2. 通用 bugfix 与 Haven 专有改动分分支、分 PR，不混在一坨 diff；
3. 定期 `git pull upstream master` 同步上游，避免分叉过远。

`D:\WatchAlert` 现有 remote：`origin = opsre/WatchAlert`（上游目标）、`upstream = w8t-io/WatchAlert`、`mine = Havens-blog/e-watch-alert`（fork）。

## 5. 风险与开放问题

- **SSO 免密单点（开放项）**：当前接入是「网关鉴权 + WatchAlert 自身 JWT 登录」，即统一入口但非免密单点。`www.jlcops.com` 的 `auth_request /api/auth → 18080` 网关只在生产存在，dev 环境未布线；eiam 是身份源管理（对接 LDAP/OIDC/飞书/企微），不直接向下游应用下发 OIDC。免密单点需后续三选一：
  1. WatchAlert 增加 OIDC/OAuth 登录（需改上游代码 → 走通用功能补丁 PR），
  2. WatchAlert 接 LDAP + eiam 作为 LDAP 源，
  3. 自制 auth 代理注入可信 `X-User` 头并改 WatchAlert 信任该头。
- **dev.ps1 未纳入管理**：WatchAlert 构建入口在 `cmd/`（非根 main.go），与 dev.ps1 的 Daemon `go build -o $exe .` 约定不符，暂未挂进 dev.ps1；当前手动启动（后端 `go build -o watchalert.exe ./cmd` 后运行，前端 `yarn start`）。如需纳入，需给 dev.ps1 服务表增加自定义 build 命令字段。
- **生产鉴权网关地址**：18080 auth gateway 的部署位置与约定需与生产环境核实后再启用 vhost 内的生产 SSO 段。

## 6. 关联

- 上游 PR：`opsre/WatchAlert#257`、`opsre/WatchAlert#258`
- 上游仓库：`w8t-io/WatchAlert` / `opsre/WatchAlert`；fork：`Havens-blog/e-watch-alert`
- 本决策与 `0001-cmdb-consolidation-keep-ecmdb.md` 同属「外部依赖接入方式」系列。