# Houtu — 运维平台（Haven 单仓）

掌管大地山川的女神。本仓库统一收录运维平台全部子模块（前端 / 后端服务 / 部署配置 / 文档 / 验证脚本），各模块完整保留原仓库 git 历史。

## 模块地图

| 目录 | 模块 | 技术栈 | 说明 |
|---|---|---|---|
| `apps/haven-console` | 平台控制台 | Vue 3 + Vite 7 + TS + Element Plus + Pinia | 统一控制台（/console/），CI 含 typecheck/build/gz、vitest、gzip 预算、e-cam-web 拷贝面漂移比对 |
| `apps/e-cam-web` | 旧版云管前台 | Vue + Vite | 控制台主题令牌 / 模式文件的拷贝来源（shared-hashes 漂移防线基准） |
| `apps/opsagent-web` | Agent 独立前端 | Vue 3 + Vite | AIOps 编排控制面 |
| `apps/multi-cloud-dashboard` | 多云拓扑原型 | 静态 HTML | 原型阶段 |
| `services/eiam` | 身份认证服务 | Go | 控制台 /api/iam/* 后端 |
| `services/e-cam-service` | 云管后端 | Go | `replace e-iam => ../eiam`（本地工作区，见 go.work） |
| `services/opsagent` | 运维 Agent 后端 | Go | |
| `deploy/nginx` | nginx 反代配置 | conf | conf/ + vhost/（仅配置，nginx 二进制不入库） |
| `docs` | 平台全部文档 | Markdown | 设计/记录/验收 |
| `scripts` | 平台级脚本 | Node/PowerShell | REAL 验证脚本 + 仓库卫生检查 |
| `patches` | 上游补丁 | patch | 0001-fix-iam.patch |

## 构建

- 前端（各 app）：`pnpm i && pnpm build`（控制台加 `pnpm build:gz`）
- Go 服务：仓库根 `go.work` 已含 3 个服务；`cd services/<x> && go build ./...`
- 平台级检查：`node scripts/repo-hygiene.mjs`

## CI

`.github/workflows/ci.yml`：按模块路径过滤——console 4 项门禁（typecheck+build+gz / vitest / gzip 预算 / 漂移比对）；3 个 Go 服务 test+vet+build；e-cam-web / opsagent-web 前端构建；仓库卫生门禁（大文件/垃圾目录零容忍）。