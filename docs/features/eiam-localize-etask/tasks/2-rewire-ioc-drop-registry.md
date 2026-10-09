---
id: "2"
title: "接线替换（ioc/grpc.go + app.go + 删 registry.go）并重跑 wire"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: ["1"]
surface-key: "eiam"
surface-type: "api"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 2: 接线替换（ioc/grpc.go + app.go + 删 registry.go）并重跑 wire

## Description
把 ioc 层的 etask 接线改为本地 `grpcx`：`ioc/grpc.go` 去 `registry` 参数，`ioc/app.go` 的 `Server` 类型改 `*grpcx.Server`，删除 `ioc/registry.go`（`InitRegistry`，仅被 gRPC 使用），并重跑 wire 更新 `wire.go`/`wire_gen.go`。此步完成后所有 etask import 归零，为删依赖铺路。

## Reference Files
- `docs/proposals/eiam-localize-etask/proposal.md` — In Scope 与依赖清理 (ref: Scope)
- `eiam/ioc/{grpc,app,registry,wire,wire_gen}.go` — 接线改动点 (ref: Scope)

## Acceptance Criteria
- [ ] `ioc/grpc.go` 改用 `grpcx.NewServer(cfg, grpcx.WithJWTAuth(cfg.AuthToken))`，去掉 `registry registrysdk.Registry` 参数
- [ ] `ioc/app.go` 的 `Server *grpcpkg.Server` 改为 `*grpcx.Server`
- [ ] 删除 `ioc/registry.go`（`InitRegistry`）
- [ ] `ioc/wire.go` 移除 `InitRegistry` 供给项并重跑 `wire` 更新 `wire_gen.go`（`InitGrpcServer` 调用链去掉 registry 入参）；`go build ./...` 通过
- [ ] `grep -rn "Duke1616/etask" eiam/ --include=*.go` 返回 0 条结果

## Hard Rules
- 仅修改 `ioc/` 下的 grpc.go/app.go/registry.go/wire.go/wire_gen.go，勿动其他包

## Implementation Notes
- `InitEtcd`（`ioc/etcd.go`）**保留**——仍被 `ioc/capability.go` 的 `InitCapability` 与 `internal/service/discovery` 使用；只删 `InitRegistry`（etcd→etask Registry 的桥）。
- wire 重生成：在 `eiam` 目录下执行 `wire ./ioc`（google/wire）；若无 wire 工具，则手改 `wire_gen.go`：删 `registryRegistry := InitRegistry(clientv3Client)` 一行、把 `InitGrpcServer(registryRegistry, ...)` 改为 `InitGrpcServer(...)`（去首个参数），并同步删 `wire.go` 的 `InitRegistry` 供给项。
- 三个 proto 服务注册（`userv1.RegisterUserServiceServer` 等）继续作用于 `grpcx.Server`（内嵌 `*grpc.Server` 满足 `grpc.ServiceRegistrar`），无需改。

### Test Impact
- Affected test suite(s): `eiam/ioc/`（wire_gen.go 重生成）
- Expected fixture changes: wire_gen.go 由工具重生成
- Risk level: medium