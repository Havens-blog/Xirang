---
id: "3"
title: "切换 5 件 import、删除 go.mod ecmdb 依赖并全量验证"
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

# 3: 切换 5 件 import、删除 go.mod ecmdb 依赖并全量验证

## Description
把 5 个引用 `github.com/Duke1616/ecmdb/pkg/cryptox` 的文件改为引用本地 `github.com/Duke1616/eiam/internal/cryptox`（业务代码不改），从 `go.mod` 删除 `github.com/Duke1616/ecmdb` 并 `go mod tidy` 清理，最后全量验证 build/vet/test/grep 收口，彻底摘除 ecmdb 依赖。

## Reference Files
- `docs/proposals/eiam-localize-cryptox/proposal.md` — In Scope 与 Success Criteria (ref: Scope / Success Criteria)
- `eiam/go.mod` — 删除 ecmdb 依赖 (ref: Scope)

## Acceptance Criteria
- [ ] 5 个文件的 import 从 `github.com/Duke1616/ecmdb/pkg/cryptox` 改为 `github.com/Duke1616/eiam/internal/cryptox`，且 `grep -rn "ecmdb/pkg/cryptox" eiam/ --include=*.go` 返回 0 条结果
- [ ] `go.mod` 删除 `github.com/Duke1616/ecmdb`，`go mod tidy` 后 `go.sum` 不再含该依赖、无报错
- [ ] `go build ./...` 与 `go vet ./...` 通过，无新增告警
- [ ] `go test ./internal/cryptox/...` 全部通过

## Hard Rules
（无）

## Implementation Notes
- 5 个文件清单（勿遗漏）：`ioc/crypto.go`、`ioc/identity_source.go`、`internal/service/user/service.go`、`internal/service/identity_source/service.go`、`cmd/migrate/internal/migrations/user.go`。
- `ioc/wire_gen.go` 只引用 `InitCryptoManager()` 函数、不直接 import cryptox 包，无需改动；若执行时发现 wire 生成物引用了旧包类型，需重跑 `wire` 或修正 import。
- 业务代码（加解密调用、`Register`/`WithLegacyAlgo` 链、迁移逻辑）一律不改，只改 import 路径。
- 在 eiam 目录下执行 build/vet/test（eiam 是独立 Go module）。

### Test Impact
- Affected test suite(s): `eiam/internal/cryptox/`、eiam 全量 `go build ./...`
- Expected fixture changes: 无
- Risk level: low