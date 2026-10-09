---
created: "2026-10-09"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: eiam 去 etask 依赖（保留 gRPC 只留 JWT）

## Problem

eiam 仍依赖 `github.com/Duke1616/etask v0.0.7`——这是继 ecmdb（已移除）之后 eiam 侧**最后一条** Duke1616 依赖。etask 的 `pkg/grpc` 提供了 gRPC server 包装 + 服务注册 + 拦截器链，但其中大部分是为 ework-runner 多租户平台设计的，eiam 并不需要。

### Evidence

- `eiam/go.mod` 第 7 行：`github.com/Duke1616/etask v0.0.7`。
- etask 仅被 3 个文件 import：`ioc/app.go`（`grpcpkg.Server` 类型）、`ioc/grpc.go`（`grpcpkg.NewServer` + `registrysdk.Registry`）、`ioc/registry.go`（`etcd.NewRegistry`）。
- **eiam 的 gRPC 服务零消费方**：全仓（e-cam-service / haven-console / opsagent / multi-cloud-dashboard）无任何 `grpc.Dial`、`ListServices`、`:8077` 引用；eiam 主接口是 HTTP `:9000`。
- etask 被 eiam 用到的最小闭包：`server.go`(314 行) + 拦截器管道(bizid/jwt/tenant) + `registry/etcd`(262 行) + `netx`(23 行)，合计约 1000 行；其中 bizid/tenant 仅在 ework-runner 场景使用，etcd 注册无发现方。

### Urgency

- 完成 eiam 对 Duke1616 依赖的彻底归零（自持目标）；继续保留一个 ~1000 行、且大半是「无人消费」的依赖纯属负担。
- 依赖面有明确无人使用的部分（etcd 注册、bizid/tenant 拦截器），是低成本去除的窗口。

## Proposed Solution

**保留 gRPC 服务（user/tenant/department @ `:8077`），但把 etask 的 server 包装替换为原生 `google.golang.org/grpc` + 仅移植 JWT 拦截器；砍掉 etcd 服务注册与 bizid/tenant 拦截器。**

具体：
1. 新建本地包 `eiam/internal/grpcx`：实现 `Server`（内嵌 `*grpc.Server`，实现 ego `server.Server` 接口：Name/Init/Start/Stop/GracefulStop/PackageName/Info）+ `ServerConfig`（name/listen_addr/auth_token）+ `NewServer`/`WithJWTAuth` + JWT 鉴权拦截器（移植自 etask `interceptors/jwt` 的 server 侧，含 claims 上下文注入）。
2. `ioc/grpc.go` 改用本地 `grpcx`，去掉 `registry` 参数；删除 `ioc/registry.go`（`InitRegistry`，仅被 gRPC 使用）；重跑 `wire` 更新 `wire.go`/`wire_gen.go`。
3. 从 `go.mod` 删除 `etask`，`go mod tidy`；`github.com/golang-jwt/jwt/v4` 由 indirect 转为 direct（JWT 拦截器直接 import）。
4. 简化 `config/config.yaml` 的 `grpc.server.eiam`（去掉 `advertise_addr`，保留 `name`/`listen_addr`/`auth_token`）。

### Innovation Highlights

诚实说明：这不是算法/架构创新，而是**按需裁剪 + 自主化**。把「为一个更大平台（ework-runner）设计的通用 gRPC 框架」裁剪成「eiam 实际需要的：原生 gRPC + JWT 鉴权」。裁剪的依据是硬事实——etcd 注册零发现方、bizid/tenant 零消费方。去掉的代码量远大于新写的代码量。

## Requirements Analysis

### Key Scenarios

- **gRPC 服务照常可用**：`user/tenant/department` 三服务仍注册到 gRPC server，在 `:8077` 监听，带 JWT 鉴权（`auth_token` 生效）。
- **无 etcd 注册**：server 启动时不再向 etcd 注册 `eiam` 服务（原来也无发现方，属写空）。
- **JWT 鉴权行为不变**：Bearer token 校验 HMAC 签名、过期、注入 claims，与原 etask 拦截器字节级一致。

### Non-Functional Requirements

- **兼容性**：JWT 拦截器行为与原 etask 一致（HS256 + tokensecret、Bearer 前缀、claims 注入）；gRPC 服务名/端口/鉴权不变。
- **ego 框架适配**：新 `Server` 必须实现 ego `server.Server` 接口，才能继续被 `egoApp.Serve(...)` 托管。

### Constraints & Dependencies

- 保留 eiam 自身的直接依赖：`google.golang.org/grpc`、`gotomicro/ego`、`go.etcd.io/etcd/client/v3`（后两者不是 Duke1616，且分别被 HTTP/资产发现继续使用）。
- JWT 移植依赖 `github.com/golang-jwt/jwt/v4`（原为 indirect，去 etask 后转 direct）。

## Alternatives & Industry Benchmarking

### Industry Solutions

去第三方 gRPC 框架依赖的通行做法：恢复到官方 `google.golang.org/grpc` 原生 server，把真正需要的中间件（鉴权、日志、metrics）按需内联——这正是行业主流（官方 grpc + 自写 interceptor），etask 只是在原生 grpc 上包了一层带 ework-runner 色彩的外壳。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零改动 | 继续依赖 ~1000 行且大半无人消费的 Duke1616 模块 | Rejected: 违背自持目标 |
| 原样移植 etask 子集 | etask pkg/grpc | 行为字节一致 | ~1000 行、把 bizid/tenant/etcd 死代码原样搬进自己仓 | Rejected: 无谓搬运 |
| 直接删 gRPC 服务 | — | 范围最小 | 丢掉 user/tenant/department 的 gRPC API 面 | Rejected: 丢未来接口 |
| **原生 grpc + 只留 JWT** | 本方案 | 去 etask 彻底、代码最小、保留 gRPC API | 需实现 ego server 接口 + 重跑 wire | **Selected: 裁剪掉的 > 新增的，去依赖彻底** |

## Feasibility Assessment

### Technical Feasibility

完全可行。原生 `google.golang.org/grpc` 与 ego `server.Server` 接口都是 eiam 已掌握的依赖；JWT 拦截器 ~190 行可直接移植；wire 重跑是机械操作。无未知外部依赖。

### Resource & Timeline

中等规模重构，估算 1 人天内；涉及 wire 重生成与 gRPC 服务冒烟验证（有本地/远端环境可验证 :8077 与 JWT）。

### Dependency Readiness

etask v0.0.7 源码冻存在模块缓存可作移植参照；ego/grpc 均为稳定公共依赖；无外部服务阻塞。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| etask 依赖巨大、必须 full mode | 源码行数 + 消费方分析 | Overturned: eiam 只用到 ~1000 行，且 gRPC 零消费方，裁剪后仅 ~350 行新增 |
| etcd 注册必须保留 | 消费方反查 | Overturned: 全仓无 ListServices/grpc.Dial 消费 eiam 注册，注册是写空 |
| bizid/tenant 拦截器是 eiam 需要 | 场景核对 | Overturned: 属 ework-runner 多租户，auth_token 已配置时服务端只用 bizid+jwt |

## Scope

### In Scope

- 新建 `eiam/internal/grpcx`：`Server`（内嵌 `*grpc.Server` + ego `server.Server` 接口）、`ServerConfig`、`NewServer`/`WithJWTAuth`、JWT 鉴权拦截器（HS256/Bearer/claims 注入，移植 etask jwt 的 server 侧）。
- `ioc/grpc.go` 改用本地 `grpcx`、去 `registry` 参数；删除 `ioc/registry.go`；`ioc/app.go` 改 `grpcx.Server`。
- `ioc/wire.go` + `ioc/wire_gen.go` 重生成（去掉 `InitRegistry` 供给项）。
- `go.mod` 删 `etask` + `go mod tidy`；`config/config.yaml` 的 `grpc.server.eiam` 去 `advertise_addr`。
- 验证：`go build ./...`、`go vet ./...`、gRPC server 启动并注册 user/tenant/department、JWT 鉴权生效。

### Out of Scope

- `go.etcd.io/etcd/client/v3` 的 `InitEtcd` 及其在 `internal/service/discovery`、`pkg/web/capability` 的用法（eiam 自有依赖，继续用）。
- `api/proto/gen/eiam/*` 的 proto 契约变更。
- eiam 模块名 `github.com/Duke1616/eiam` 更名（大盘面，单独规划）。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| wire 重生成出错致依赖注入断裂 | M | H | 用 google/wire 重跑生成，`go build ./...` + 启动冒烟验证 |
| 去掉 etcd 注册导致某些隐含发现方失联 | L | M | 已全仓反查无 ListServices/Subscribe 消费方；保留 InitEtcd 供 discovery/capability 使用 |
| JWT 拦截器移植细节不一致导致鉴权行为漂移 | L | H | 逐字移植 etask `interceptors/jwt/server.go`+`context.go`，加 KAT/round-trip 单测 |
| 简化后 gRPC server 未实现 ego 接口致启动失败 | M | M | Server 完整实现 ego `server.Server` 六方法，启动冒烟验证 |

## Success Criteria

- [ ] SC-1: `go.mod`/`go.sum` 中不再出现 `github.com/Duke1616/etask`，`go mod tidy` 无报错
- [ ] SC-2: `grep -rn "Duke1616/etask" eiam/ --include=*.go` 返回 0 条结果
- [ ] SC-3: `go build ./...` 与 `go vet ./...` 通过
- [ ] SC-4: `eiam/internal/grpcx` 单元测试通过（JWT 鉴权：有效/过期/签名错/无 token 四路）
- [ ] SC-5: gRPC server 启动后能注册 user/tenant/department 三服务，`auth_token` 生效（冒烟）

consistency_check_result:
  status: pass
  pairs_checked: 8
  conflicts_found: 0

## Next Steps

- `/quick-tasks` 生成任务（裁剪后范围已降到 quick 级别），任务内完成 wire 重生成与 gRPC 冒烟。