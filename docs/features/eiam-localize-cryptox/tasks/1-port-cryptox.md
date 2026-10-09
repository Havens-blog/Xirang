---
id: "1"
title: "内联移植 cryptox 到 eiam/internal/cryptox（含 round-trip 单测）"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: []
surface-key: "eiam"
surface-type: "api"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 1: 内联移植 cryptox 到 eiam/internal/cryptox（含 round-trip 单测）

## Description
将已废弃的 `github.com/Duke1616/ecmdb v1.11.0` 中 `pkg/cryptox` 包原样内联移植为 eiam 本地包 `eiam/internal/cryptox`，仅依赖 Go 标准库，导出 API 与原库一致（供后续 5 个调用点仅改 import 不改业务代码）。算法为标准 AES-GCM，但必须逐字保留历史细节：V1 的 `padKey`（截断/零填充到 16 字节）+ 对明文 `json.Marshal`；V2 的 SHA256(key) 作 32 字节密钥且明文不序列化；`CryptoManager` 的 `ENC:<version>:<hex>` 前缀与 `WithLegacyAlgo` 回退语义。

## Reference Files
- `docs/proposals/eiam-localize-cryptox/proposal.md` — 方案与移植范围 (ref: Proposed Solution)
- `github.com/Duke1616/ecmdb@v1.11.0/pkg/cryptox/{types,aes_crypto,aes_crypto_v2,manager}.go`（模块缓存） — 移植参照源码 (ref: Proposed Solution)

## Acceptance Criteria
- [ ] 新建 `eiam/internal/cryptox`，移植 `Crypto` 接口、`CryptoManager`（`NewCryptoManager`/`Register`/`WithLegacyAlgo`/`Encrypt`/`Decrypt` + `ENC:` 前缀格式）、`MustNewAESCrypto`（V1：padKey+JSON+AES-GCM）、`MustNewAESCryptoV2`（V2：SHA256+AES-GCM），导出名与签名与原库完全一致
- [ ] 仅依赖 Go 标准库（`crypto/aes`、`crypto/cipher`、`crypto/rand`、`crypto/sha256`、`encoding/hex`、`encoding/json`）
- [ ] round-trip 单测：V1、V2、Manager（含 `ENC:` 前缀、legacy 回退）加密后解密还原明文，`go test ./internal/cryptox/...` 全部通过

## Hard Rules
（无）

## Implementation Notes
- 移植参照 `ecmdb@v1.11.0/pkg/cryptox` 四个文件：`types.go`（Crypto 接口 + EncryptedPrefix）、`manager.go`、`aes_crypto.go`（V1）、`aes_crypto_v2.go`（V2）。`regsiter.go`（CryptoRegistry）与 `aes.go`（EncryptAES/DecryptAES 自由函数）不被 eiam 使用，不移植。
- 注意 eiam 存在两套版本键语义，移植时保持字节/字符串一致：`ioc/crypto.go` 用默认 `"v1"`（小写）+ V2 算法；`cmd/migrate/internal/migrations/user.go` 用默认 `"V2"` + 注册 `encryptionVersion`（legacy）+ `WithLegacyAlgo(encryptionVersion)`。
- 包放 `internal/cryptox`：cryptox 是敏感工具且无需对外导出，internal 避免外部模块 import；同 module 内 `cmd/migrate` 可正常引用。
- gofmt/goimports 跑干净。

### Test Impact
- Affected test suite(s): `eiam/internal/cryptox/`（新增）
- Expected fixture changes: 无（纯新增）
- Risk level: low