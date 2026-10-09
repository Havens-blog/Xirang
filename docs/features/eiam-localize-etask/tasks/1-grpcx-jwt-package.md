---
id: "1"
title: "新建 internal/grpcx（Server + JWT 拦截器 + 单测）"
priority: "P0"
estimated_time: "1.5h"
complexity: "medium"
dependencies: []
surface-key: "eiam"
surface-type: "api"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: 新建 internal/grpcx（Server + JWT 拦截器 + 单测）

## Description
新建本地包 `eiam/internal/grpcx`，用原生 `google.golang.org/grpc` 替换 etask 的 server 包装：实现 `Server`（内嵌 `*grpc.Server`，实现 ego `server.Server` 接口，无 etcd 注册逻辑）+ `ServerConfig`（Name/ListenAddr/AuthToken）+ `NewServer`/`WithJWTAuth`，以及从 etask `interceptors/jwt` 移植的 JWT 鉴权拦截器（仅 server 侧）。这是去 etask 的核心交付物。

## Reference Files
- `docs/proposals/eiam-localize-etask/proposal.md` — 方案与裁剪范围 (ref: Proposed Solution)
- `etask@v0.0.7/pkg/grpc/{server.go,interceptors/jwt/server.go,interceptors/jwt/context.go}`（模块缓存） — 移植参照 (ref: Proposed Solution)

## Acceptance Criteria
- [ ] 新建 `eiam/internal/grpcx`：`ServerConfig`（Name/ListenAddr/AuthToken）、`Server`（内嵌 `*grpc.Server`）、`NewServer(cfg, opts...)`、`WithJWTAuth` 选项
- [ ] `Server` 完整实现 ego `server.Server` 接口（Name/Init/Start/Stop/GracefulStop/PackageName/Info），`Start` 用 net.Listen + 异步 `grpc.Server.Serve`，**无** etcd 注册逻辑
- [ ] JWT 鉴权拦截器移植自 etask `interceptors/jwt`（HS256、Bearer 前缀、claims 注入、`Set`/`Get` 上下文辅助），仅依赖 `github.com/golang-jwt/jwt/v4` 与 `google.golang.org/grpc`
- [ ] 单测覆盖 JWT 四路：有效 token 放行 / 无 token 拒绝 / 过期拒绝 / 签名错拒绝，`go test ./internal/grpcx/...` 通过

## Hard Rules
- 仅新建 `eiam/internal/grpcx/` 下的文件，不修改其他包

## Implementation Notes
- 移植参照 module cache `etask@v0.0.7/pkg/grpc`：`server.go`（去 registry/netx/advertise 逻辑）、`interceptors/jwt/server.go` + `interceptors/jwt/context.go`。**不移植** bizid/tenant/`registry/etcd`/`netx`（ework-runner 专属、eiam 无人用）。
- ego `server.Server` 接口六方法签名以 etask server.go 的 Name/Init/Start/Stop/GracefulStop/PackageName/Info 为参照（`Start` 返回 error、`GracefulStop(ctx)` 返回 error、`Info()` 返回 `*server.ServiceInfo`）。
- 不引入 `go.etcd.io/etcd`（grpcx 与 etcd 无关）；`golang-jwt/jwt/v4` 在任务 3 由 indirect 转 direct。
- gofmt/goimports 干净。

### Test Impact
- Affected test suite(s): `eiam/internal/grpcx/`（新增）
- Expected fixture changes: 无（纯新增）
- Risk level: low