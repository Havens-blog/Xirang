---
status: "completed"
started: "2026-10-09 09:33"
completed: "2026-10-09 09:38"
time_spent: "~5m"
---

# Task Record: 1 内联移植 cryptox 到 eiam/internal/cryptox（含 round-trip 单测）

## Summary
Ported ecmdb v1.11.0 pkg/cryptox byte-for-byte to eiam/internal/cryptox (4 source files + round-trip unit tests). Crypto interface, CryptoManager (NewCryptoManager/Register/WithLegacyAlgo/WithMigrationHandler/Encrypt/Decrypt + ENC: prefix), MustNewAESCrypto (V1: padKey+JSON+AES-GCM), MustNewAESCryptoV2 (V2: SHA256+AES-GCM) all preserve exact exported names/signatures. Only Go stdlib dependencies. 11 tests covering V1/V2/Manager round-trip, ENC: prefix, legacy fallback, migration handler, plaintext fallback, and error paths — all green at 83% coverage.

## Changes

### Files Created
- eiam/internal/cryptox/types.go
- eiam/internal/cryptox/manager.go
- eiam/internal/cryptox/aes_crypto.go
- eiam/internal/cryptox/aes_crypto_v2.go
- eiam/internal/cryptox/cryptox_test.go

### Files Modified
无

### Key Decisions
- Byte-for-byte port of 4 source files (types/manager/aes_crypto/aes_crypto_v2) from ecmdb@v1.11.0 module cache, package name unchanged — enables 5 caller files to switch imports without business-code changes in a later task
- Excluded regsiter.go (CryptoRegistry) and aes.go (EncryptAES/DecryptAES free functions) per task Implementation Notes — not used by eiam
- Preserved V1 historical details: padKey (truncate/zero-pad to 16 bytes) + json.Marshal of plaintext; V2: SHA256(key) as 32-byte key, no serialization; CryptoManager ENC:<version>:<hex> prefix and WithLegacyAlgo fallback semantics
- Round-trip tests only (no KAT vectors) — task AC specifies round-trip; KAT with original library is feasible but out of scope for this task

## Test Results
- **Tests Executed**: Yes
- **Passed**: 11
- **Failed**: 0
- **Coverage**: 83.0%

## Acceptance Criteria
- [x] eiam/internal/cryptox created with Crypto interface, CryptoManager (NewCryptoManager/Register/WithLegacyAlgo/Encrypt/Decrypt + ENC: prefix), MustNewAESCrypto (V1), MustNewAESCryptoV2 (V2) — exports/signatures match original
- [x] Only Go stdlib dependencies (crypto/aes, crypto/cipher, crypto/rand, crypto/sha256, encoding/hex, encoding/json)
- [x] Round-trip unit tests: V1, V2, Manager (ENC: prefix, legacy fallback) encrypt-then-decrypt restores plaintext; go test ./internal/cryptox/... all pass

## Notes
go build ok, go vet ok, gofmt clean. git status had unrelated dirty files from concurrent sessions (internal/web/tenant/handler.go, internal/web/user/user.go, ioc/session.go) — this task is pure new-file addition so safe rollback is unaffected. Branch is main. The 5 caller import-path changes and go.mod ecmdb removal are explicitly a later task per proposal scope.
