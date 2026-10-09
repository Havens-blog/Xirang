---
id: "3"
title: "删 etask 依赖、简化 config 并全量验证"
priority: "P1"
estimated_time: "0.5h"
complexity: "medium"
dependencies: ["2"]
surface-key: "eiam"
surface-type: "api"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 3: 删 etask 依赖、简化 config 并全量验证

## Description
从 `go.mod` 删除 `github.com/Duke1616/etask` 并 `go mod tidy`，简化 `config/config.yaml` 的 `grpc.server.eiam`（去 `advertise_addr`），最后全量验证 build/vet/test/冒烟收口，彻底摘除 etask 依赖。

## Reference Files
- `docs/proposals/eiam-localize-etask/proposal.md` — Success Criteria 与依赖清理 (ref: Success Criteria)
- `eiam/go.mod`、`eiam/config/config.yaml` — 删依赖与简化配置 (ref: Scope)

## Acceptance Criteria
- [ ] `go.mod` 删除 `github.com/Duke1616/etask`，`go mod tidy` 后 `go.sum` 不再含该依赖、无报错
- [ ] `config/config.yaml` 的 `grpc.server.eiam` 去掉 `advertise_addr`（保留 `name`/`listen_addr`/`auth_token`）
- [ ] `go build ./...` 与 `go vet ./...` 通过
- [ ] `go test ./internal/grpcx/...` 通过
- [ ] gRPC 冒烟：构建产物可编译、三 proto 服务注册链路完整（环境允许则启动验证 `:8077` 与 JWT auth_token）

## Hard Rules
（无）

## Implementation Notes
- `github.com/golang-jwt/jwt/v4` 会由 indirect 转 direct（grpcx 直接 import），tidy 会去掉 `// indirect` 标记，属预期。
- `go.etcd.io/etcd/client/v3` 保留（`InitEtcd` 仍被 discovery/capability 使用），不要顺手删。
- 冒烟验证若受环境（远端 etcd/DB/端口）限制无法启动，则以 `go build ./...` + grpcx 单测作为回归基线，并在记录里如实说明冒烟方式与结果。

### Test Impact
- Affected test suite(s): `eiam/internal/grpcx/`、eiam 全量 `go build ./...`
- Expected fixture changes: 无
- Risk level: low