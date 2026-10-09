---
status: "living"
updated: "2026-10-10"
---

# 息壤运维平台 · 本地部署运行手册（本机 Windows · 无 Docker 裸进程形态）

> 目标：在本机把整套平台跑起来（nginx :8888 统一入口）。本机没有 Docker，
> 数据面长期直连远程共享库；线上形态（容器 + CI 构建推送）另见 `.github/workflows/deploy-*.yml`。

## 0. 当前正在跑的实例（基线核对）

| 组件 | 二进制 / 源 | 端口 | 说明 |
|---|---|---|---|
| nginx 1.25.4 | `D:\Haven\nginx-1.25.4\nginx.exe` | 8888 | 统一入口，vhost = `conf/vhost/nginx.dev.conf` |
| eiam | `D:\Haven\logs\dev\bin\eiam.exe server`（源码 `D:\Haven\eiam`） | 9000 | 配置按**工作目录**解析 `config/config.yaml` |
| e-cam-service | `D:\Haven\logs\dev\bin\e-cam-service.exe start`（源码 `D:\Haven\e-cam-service`） | 8001 | EGO 框架；读 config/prod.yaml（远程库+云凭据） |
| opsagent-api | `F:\OpsAgent\build\opsagent-api.exe`（源码 `F:\OpsAgent`） | **8080** | ⚠️ 见 §4 坑一：nginx 指向 8081 |
| opsagent 旧前端(Next) | `F:\OpsAgent\web\frontend`（`npm run dev`/`next start`） | 3000 | 直连使用，不经 nginx |
| e-cam-web | `D:\Haven\e-cam-web` vite dev server | 5173 | nginx `/cam/` → `127.0.0.1:5173/cam/` |
| haven-console | `D:\Haven\haven-console\dist`（构建产物） | — | nginx `/console/` 静态托管 |
| opsagent-web(Vue) | `D:\Haven\opsagent-web\dist`（构建产物） | — | nginx `/opsagent/` 静态托管 |

## 1. 依赖口径（本机没有 DB）

- 本机**不跑** mongo/redis/etcd（无 Docker；端口 27017/6379/2379 均未监听）。
- e-cam-service 按 `EGO_CONFIG_PATH` 直连远程共享库：
  - `config/test.yaml` → 测试库（远程 IP，免云凭据场景）
  - `config/prod.yaml` → 生产库 + 云凭据（**默认口径**，端口 8001）
- eiam 的 `config/config.yaml` 同理（远程库）。redis/etcd 相关配置按模块现状指向远程/占位。

## 2. 标准启动顺序（冷启动全流程）

```powershell
# ① 三个 Go 后端
# eiam —— 必须在含 config/config.yaml 的目录下启动
cd D:\Haven\eiam
go build -o D:\Haven\logs\dev\bin\eiam.exe .
D:\Haven\logs\dev\bin\eiam.exe server          # :9000

# e-cam-service —— 官方入口：.\deploy\build.ps1 dev（wire 生成 + go run main.go start）
cd D:\Haven\e-cam-service
$env:EGO_CONFIG_PATH = "config\prod.yaml"
go run main.go start                            # :8001
# 或构建后跑：.\deploy\build.ps1 build → build\e-cam-service.exe start

# opsagent-api —— 两条线二选一（端口错位看 §4 坑一）
cd F:\OpsAgent
.\run.ps1                                       # 旧版走法：api :8080 + Next 前端 :3000
# 或单仓新版（默认 8081，与 nginx 对齐）：
#   cd D:\Haven\_mono-build\Houtu\services\opsagent
#   go build -o opsagent.exe ./cmd/opsagent && .\opsagent.exe   # :8081

# ② 前端
cd D:\Haven\e-cam-web && pnpm dev               # :5173（/cam/ 经 nginx 反代）
cd D:\Haven\haven-console && pnpm build:gz      # 产物 dist/ → nginx /console/
cd D:\Haven\opsagent-web && pnpm build:gz       # 产物 dist/ → nginx /opsagent/

# ③ nginx（已在跑时只需 reload）
D:\Haven\nginx-1.25.4\nginx.exe -s reload
```

## 3. 访问入口

- 根入口 `http://localhost:8888/` → 301 → `/console/`（haven-console 控制台）
- `/cam/` e-cam-web 云管前台 · `/opsagent/` Agent 前端
- API 路由（nginx.dev.conf）：
  - `/api/iam/*` → eiam :9000 `/api/*`（大 Set-Cookie 已放宽 proxy_buffer 32k）
  - `/api/v1/cmdb/` `/api/v1/cam/` `/api/v1/certs` → e-cam-service :8001（certs 块**不带尾斜杠**，勿改）
  - `/api/v1/opsagent/` → opsagent :8081（联邦查询 cam/certs 已放宽 300s 超时）

## 4. 三个已知坑（动手前必读）

1. **opsagent 端口错位（当前是断的）**：nginx 指向 `127.0.0.1:8081`，但
   F: 盘旧 exe 监听 8080（其 compose/env 为 `OPSAGENT_PORT=8080`）。
   → 两种修法：a) 用单仓新版（默认 8081）；b) 把 `nginx.dev.conf` 的 8081 改 8080 后
   `nginx -s reload`。改完 `curl localhost:8888/api/v1/opsagent/health` 验证。
2. **e-cam-service NAS 同步 panic（真 Bug）**：`logs\dev\manual-start.log` 记录两次
   `panic: assignment to entry in nil map`，栈在
   `internal/shared/cloudx/aliyun/nas_metrics.go:155 createCMSClient`（client 缓存
   map 未初始化即写入）。触发条件：NAS 指标同步任务 + 云凭据在库。**当前跑的实例是
   重启后的产物；下次跑 NAS 指标任务还会炸。**建议列入下一批修复（候选 7.2）。
3. **OpsAgent 双源分叉**：真身开发区在 `F:\OpsAgent`（Next 15 前端 + 4 个二进制，
   8080/9090）；单仓 `services/opsagent` 是历史镜像快照（默认端口 8081）；
   `apps/opsagent-web`（Vue 重写）镜像自 `D:\Haven\opsagent-web`，而 F: 的 Next
   前端**未入单仓**。开工前选定一条开发线，避免两处分叉。

## 5. 单仓 ⇔ 旧运行源对应

| 旧运行源（当前进程） | 单仓路径 | 备注 |
|---|---|---|
| D:\Haven\eiam | services/eiam | config.yaml 两边都有（单仓 gitignored，不提交凭据） |
| D:\Haven\e-cam-service | services/e-cam-service | 旧仓含未入单仓的 WIP；切单仓慎重 |
| D:\Haven\e-cam-web | apps/e-cam-web | 本机 dev server 用旧仓 |
| D:\Haven\haven-console | apps/haven-console | dist 静态托管 |
| D:\Haven\opsagent-web | apps/opsagent-web | Vue 重写 → dist |
| F:\OpsAgent | services/opsagent（后端 API） | ⚠️ 快照滞后；Next 前端未收录 |

## 6. 与 CI / 线上部署的关系

- **本地（Windows）无法跑 `-race`**（无 gcc，CGO_ENABLED=0）：竞态验证一律走
  services-ci 的 Linux job（-race + mongo/redis 容器，详见 records 6.6 / 7.1）。
- **线上部署不依赖本机**：根级 workflows（deploy-e-cam-service / deploy-e-cam-web /
  deploy-eiam）在 secrets 配齐后按模块路径触发真实构建推送 + ssh compose。
- 本地=裸进程+nginx、线上=容器，两者业务配置同口径（prod.yaml / 云凭据）。