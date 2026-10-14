# Xirang — 运维平台编排仓库（orchestration-only）

> 本仓库**只做编排**:负责把平台全部模块从各自独立的代码仓库拉取/构建/运行起来。
> **应用源代码不在本仓库** —— 各模块代码留在各自原仓库(见下方模块清单),继续在原仓库开发。
> 本仓库仅保留:开发脚本、nginx 反代配置、模块清单。

## 设计原则

- **代码不回仓**:eiam / e-cam-service / opsagent / haven-console / opsagent-web / e-cam-web 的源码
  一律保留在原仓库,本仓库不复制、不 fork,杜绝双线漂移。
- **引用式统一管理**:本仓库作为唯一入口,一个 `dev-up.ps1` 拉全平台、一个清单看清所有模块。
- **运行 = 从原仓库构建**:所有二进制/dist 均从原仓库工作区构建,与开发线完全一致。

## 模块清单(源代码所在)

| 模块 | 原仓库 remote | 本地路径 | 端口/路由 |
|---|---|---|---|
| eiam(身份认证) | github.com/Havens-blog/e-iam | `D:\Haven\eiam` | :9000 → `/api/iam/*` |
| e-cam-service(云管后端) | github.com/Havens-blog/e-cam-service | `D:\Haven\e-cam-service` | :8001 → `/api/v1/{cam,cmdb,certs}` |
| opsagent(运维 Agent 后端) | github.com/Havens-blog/e-ops-agent-web | `D:\Haven\opsagent` | :8081 → `/api/v1/opsagent/` |
| haven-console(控制台) | github.com/Havens-blog/e-console-web | `D:\Haven\haven-console` | `/console/`(静态 dist) |
| opsagent-web(Agent 前端) | github.com/Havens-blog/e-ops-agent-web | `D:\Haven\opsagent-web` | `/opsagent/`(静态 dist) |
| e-cam-web(云管前端) | github.com/Havens-blog/e-cam-web | `D:\Haven\e-cam-web` | `/cam/`(vite dev :5173) |
| 共享库 e-common-go | github.com/Havens-blog/e-common-go | `D:\Haven\e-common-go` | go.work 引用 |
| 共享库 e-cloudx-sdk | github.com/Havens-blog/e-cloudx-sdk | `D:\Haven\e-cloudx-sdk` | go.work 引用 |
| 共享库 e-cmdb-service | (随 e-cam-service) | `D:\Haven\e-cmdb-service` | go.work 引用 |

## 快速开始(本地,无 Docker)

```powershell
# 1. 确保上面模块清单里的原仓库都已 clone 到 D:\Haven\
# 2. 一键构建 + 启动全部后端 + 前端 + reload nginx
.\scripts\dev-up.ps1

# 3. 若要同时启动 e-cam-web 的 vite dev(:5173,供 nginx /cam/)
.\scripts\dev-up.ps1 -WithCamWeb

# 4. 停止 dev-up 启动的服务(不动其它实例/nginx)
.\scripts\dev-down.ps1
```

统一入口:`http://localhost:8888/`(nginx 反代,指向上述各模块)。

## Go 工作区约定

`D:\Haven\go.work` 引用 `eiam` / `e-cam-service` / `e-common-go` / `e-cloudx-sdk` / `e-cmdb-service`。
**opsagent 不在 go.work 内**,独立解析依赖 —— `dev-up.ps1` 构建 opsagent 时会自动 `GOWORK=off`。

## nginx

`deploy/nginx/conf/vhost/nginx.dev.conf` 是本仓库维护的反代配置副本(实际生效位置:
`D:\Haven\nginx-1.25.4\conf\vhost\nginx.dev.conf`)。修改后需同步回该位置并
`nginx -p D:\Haven\nginx-1.25.4\ -s reload`。

## 目录

```
Xirang/
  scripts/        开发编排脚本(dev-up.ps1 / dev-down.ps1)
  deploy/nginx/   nginx 反代配置副本
  README.md       本清单
```
