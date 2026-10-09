---
id: "2"
title: "用原库生成 KAT 已知密文向量并固化字节兼容测试"
priority: "P0"
estimated_time: "0.5h"
complexity: "medium"
dependencies: ["1"]
surface-key: "eiam"
surface-type: "api"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 2: 用原库生成 KAT 已知密文向量并固化字节兼容测试

## Description
在删除 ecmdb 依赖**之前**，用原库 `github.com/Duke1616/ecmdb/pkg/cryptox` 生成已知密文向量（Known-Answer Test，KAT），固化为 `internal/cryptox/compat_test.go` 的字面常量；随后断言本地实现 `Decrypt(原库密文)` 还原明文。这是「字节级兼容存量密文」的客观证明，防止移植细节不一致导致存量密文解不开。

## Reference Files
- `docs/proposals/eiam-localize-cryptox/proposal.md` — 字节兼容要求与 KAT 策略 (ref: Requirements Analysis)
- `github.com/Duke1616/ecmdb@v1.11.0/pkg/cryptox`（模块缓存） — KAT 向量生成源 (ref: Constraints & Dependencies)

## Acceptance Criteria
- [ ] 删依赖前，用一次性程序（临时 test 或 `go run`）import 原库 cryptox，以固定 key 加密固定明文，导出 hex 密文及 `ENC:v1:` / legacy 前缀形式
- [ ] 将导出的向量固化为 `internal/cryptox/compat_test.go` 字面常量
- [ ] 断言本地 `Decrypt(原库密文)` 还原明文，覆盖 V1（非前缀 legacy）、V2、Manager（`ENC:v1:` 前缀）三路
- [ ] `go test ./internal/cryptox/...` 全部通过，兼容测试作为字节兼容证据落地

## Hard Rules
（无）

## Implementation Notes
- AES-GCM 使用随机 nonce，密文非确定 —— 兼容方向必须用「原库加密 → 本地解密」的单向验证，**不得**比对两次加密的输出。
- KAT 生成与任务 3（删依赖）有严格时序：向量必须先固化，任务 3 才能摘除 ecmdb 依赖。
- 一次性向量生成程序用后即弃；KAT 常量里的 key 与明文须固定，避免测试随环境漂移。

### Test Impact
- Affected test suite(s): `eiam/internal/cryptox/`（新增 compat_test.go）
- Expected fixture changes: KAT 字面常量（由原库一次性生成，非人工推算）
- Risk level: low