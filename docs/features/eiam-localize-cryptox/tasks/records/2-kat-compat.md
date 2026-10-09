---
status: "completed"
started: "2026-10-09 09:53"
completed: "2026-10-09 09:58"
time_spent: "~5m"
---

# Task Record: 2 用原库生成 KAT 已知密文向量并固化字节兼容测试

## Summary
Generated KAT (Known-Answer Test) vectors using the ORIGINAL github.com/Duke1616/ecmdb@v1.11.0/pkg/cryptox with fixed keys+plaintext, then hardened them as frozen string literals in internal/cryptox/compat_test.go. The local port's Decrypt recovers the original plaintext across all three compatibility paths: V1 legacy (non-prefixed, JSON-serialized), V2 (raw bytes), and Manager ENC:v1: prefix routing. Proves byte-level compatibility of existing ciphertext without comparing non-deterministic AES-GCM encryption outputs.

## Changes

### Files Created
- eiam/internal/cryptox/compat_test.go

### Files Modified
无

### Key Decisions
- Used one-way 'original-lib encrypt -> local-lib decrypt' verification because AES-GCM uses random nonce; never compared two encryption outputs
- Fixed KAT material kept in sync between the throwaway generator and compat_test.go literals: v1Key=kat-v1-fixed-key, v2Key=kat-v2-fixed-key, plain=KAT fixed plaintext 1234
- Added a 4th test (LegacyViaManagerFallback) covering the real-world mixed-store path where unprefixed V1 ciphertext is recovered through Manager legacy-fallback, beyond the 3 paths strictly required by acceptance criteria
- Throwaway generator (eiam/cmd/katgen/main.go) imported original ecmdb cryptox, was run once to harvest vectors, then deleted

## Test Results
- **Tests Executed**: Yes
- **Passed**: 17
- **Failed**: 0
- **Coverage**: 83.0%

## Acceptance Criteria
- [x] 删依赖前用一次性程序 import 原库 cryptox，固定 key 加密固定明文，导出 hex 密文及 ENC:v1: / legacy 前缀形式
- [x] 将导出的向量固化为 internal/cryptox/compat_test.go 字面常量
- [x] 断言本地 Decrypt(原库密文) 还原明文，覆盖 V1（非前缀 legacy）、V2、Manager（ENC:v1: 前缀）三路
- [x] go test ./internal/cryptox/... 全部通过，兼容测试作为字节兼容证据落地

## Notes
Vectors are frozen one-shot artifacts; do NOT regenerate. Task 3 (remove ecmdb dependency from go.mod) can now proceed since KAT compatibility evidence is landed. go vet ./internal/cryptox/... clean.
