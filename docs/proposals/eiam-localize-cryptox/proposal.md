---
created: "2026-10-09"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: eiam 内联 cryptox 去除 ecmdb 依赖

## Problem

eiam 仍依赖已废弃的第三方模块 `github.com/Duke1616/ecmdb v1.11.0`，且该依赖**唯一用途**是 `pkg/cryptox`（AES-GCM 加解密工具）；这是 ecmdb 被废弃（决策 0003）后 eiam 侧遗留的最后一条 ecmdb 依赖。

### Evidence

- `grep -rn "ecmdb/pkg/cryptox" eiam/` 命中 5 个文件：`ioc/crypto.go`、`ioc/identity_source.go`、`internal/service/user/service.go`、`internal/service/identity_source/service.go`、`cmd/migrate/internal/migrations/user.go`。
- `eiam/go.mod` 第 6 行：`github.com/Duke1616/ecmdb v1.11.0`。
- 该依赖仅 `pkg/cryptox` 一个包被 import（其余 ecmdb 能力均未使用），已在本次会话逐一核对源码。

### Urgency

- ecmdb 仓库目录（ecmdb/、ecmdb-web/）已删除、决策 0003 已登记废弃；继续依赖一个已废弃且被删除的第三方库，存在上游断更/供应链风险，也与「Duke1616 组件系统性自持」目标相悖。
- 依赖面极小（~300 行纯标准库），是低成本、低风险的自持化机会，越早做越好。

## Proposed Solution

把 ecmdb `pkg/cryptox`（~300 行，仅依赖 Go 标准库 crypto）**原样内联**到 `eiam/internal/cryptox`，保持导出 API 与算法完全一致（字节级兼容存量密文），随后从 `go.mod` 删除 ecmdb 依赖、清理 `go.sum`，并把 5 个文件的 import 路径从 `github.com/Duke1616/ecmdb/pkg/cryptox` 改为本地包。

### Innovation Highlights

诚实说明：这不是算法创新，而是**忠实移植**。算法为标准 AES-GCM；真正需要逐字保留的是三处历史细节——(a) V1 的 `padKey`（截断/零填充到 16 字节）并对明文做 `json.Marshal`；(b) V2 用 SHA256(key) 作 32 字节密钥、明文不序列化；(c) `CryptoManager` 的 `ENC:<version>:<hex>` 前缀与 legacy 回退语义（`WithLegacyAlgo`）。这些细节决定存量密文能否解开，是本移植的正确性核心。

## Requirements Analysis

### Key Scenarios

- **运行时常规加解密**：MFA secret（`mfa.go`）、LDAP bind password、OIDC client secret（`identity_source/service.go`）以 V2 算法加解密，格式 `ENC:v1:<hex>`。
- **历史密码迁移解密**：`cmd/migrate` 用 legacy AES V1 解密 ecmdb 时代 MongoDB 里的密码，再转 bcrypt（`migrations/user.go`），失败降级初始密码。
- **边界**：密文无 `ENC:` 前缀且注册了 legacy 算法时按 legacy 解密；无 legacy 算法时按明文回退；版本键未注册时报错。

### Non-Functional Requirements

- **兼容性（关键）**：本地实现与原库 `ecmdb@v1.11.0` 对相同输入的加解密结果必须字节级一致。
- **并发安全**：`AEAD.Seal`/`Open` 本身并发安全，移植后保持。
- **安全**：仅标准库、不引入自研算法；密钥仅由调用方配置注入，包内不落盘。

### Constraints & Dependencies

- 仅 Go 标准库（`crypto/aes`、`crypto/cipher`、`crypto/rand`、`crypto/sha256`、`encoding/hex`、`encoding/json`），无第三方依赖。
- 移除 ecmdb 前需先用原库生成已知密文向量（KAT）固化进测试，作为字节兼容的客观证明。

## Alternatives & Industry Benchmarking

### Industry Solutions

去第三方依赖的通用做法：`go mod vendor`（代码进 vendor 但 module path 不变）、`replace` 指向本地 fork（保留独立模块）、或**原样内联**（把已用部分移植进自己的包）。行业上，当依赖面小且算法需与历史数据兼容时，内联是主流做法。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零改动 | 继续依赖已废弃库，供应链/断更风险 | Rejected: 违背自持目标 |
| `go mod vendor` | Go module | 代码进树 | module path 仍是 `Duke1616/ecmdb`，未真正去依赖 | Rejected: 不满足目标 |
| `replace` 本地 fork | Go module | 可控 | 仍需独立模块、额外维护，未彻底自持 | Rejected: 治标不治本 |
| **原样内联到 `internal/cryptox`** | 本方案 | 彻底去依赖、字节兼容、无新算法 | 需逐字核对移植细节 | **Selected: 依赖面 ~300 行纯标准库，内联成本低、收益彻底** |

## Feasibility Assessment

### Technical Feasibility

完全可行。cryptox 源码已在模块缓存中完整核对（7 文件 396 行，其中 2 个文件 `regsiter.go`/`aes.go` 的 `CryptoRegistry`/`EncryptAES` 不被 eiam 使用，可不移植），实际移植约 300 行；所依赖均为 Go 标准库，无未知外部依赖。

### Resource & Timeline

单一小组件，纯机械移植 + 兼容性测试，估算 0.5 人天内可完成；无跨团队依赖。

### Dependency Readiness

上游 ecmdb v1.11.0 源码已冻存在模块缓存，可直接作为移植参照与 KAT 生成源；无外部 API/服务依赖。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| ecmdb 依赖只被 cryptox 使用 | 全量 grep 核对 | Confirmed: 仅 `pkg/cryptox`，且 5 文件内只用 CryptoManager + AES v1/v2 |
| 存量密文可随意重加密 | 反事实推演 | Overturned: MFA/LDAP/OIDC 密文仍在库中,必须字节兼容,不能换算法 |
| 依赖面可能很大 | 源码行数 | Overturned: 实际仅 ~300 行标准库,quick 可承载 |

## Scope

### In Scope

- 新建 `eiam/internal/cryptox`，内联移植：`Crypto` 接口、`CryptoManager`（`NewCryptoManager`/`Register`/`WithLegacyAlgo`/`Encrypt`/`Decrypt` + `ENC:` 前缀格式）、`MustNewAESCrypto`（V1：padKey+JSON+AES-GCM）、`MustNewAESCryptoV2`（V2：SHA256+AES-GCM）。导出名与签名与原库一致，使 5 个调用点仅改 import 不改业务代码。
- 修改 5 个文件的 import 路径为本地包。
- 从 `go.mod` 删除 `github.com/Duke1616/ecmdb`，`go mod tidy` 清理 `go.sum`。
- 兼容性测试：加密→解密 round-trip + 原库预生成已知密文向量（KAT）断言字节兼容。

### Out of Scope

- `github.com/Duke1616/etask v0.0.7`（gRPC 服务注册/etcd registry）的移除——单独作为 full-mode feature 处理。
- 任何加密算法、密钥派生、密文格式的变更。
- 存量数据迁移或批量重加密。
- eiam 模块名 `github.com/Duke1616/eiam` 本身的更名（更大的盘面，单独规划）。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 移植细节不一致导致原库能解本地不能解（存量密文失效） | M | H | 逐行对照移植；KAT 已知密文向量双向断言（原库密文→本地解、本地密文→原库解） |
| legacy AES v1 版本键大小写/取值不一致，migrate 解密历史密码失败 | M | M | 保留 Register/WithLegacyAlgo 原语义，`encryptionVersion` 从配置透传不改 |
| 密码学代码被改坏 | L | H | 只用标准库、不引入自研算法；完成后跑 gosec/security-reviewer |

## Success Criteria

- [ ] SC-1: `go.mod` 与 `go.sum` 中不再出现 `github.com/Duke1616/ecmdb`，`go mod tidy` 无报错
- [ ] SC-2: `grep -rn "ecmdb/pkg/cryptox" eiam/ --include=*.go` 返回 0 条结果
- [ ] SC-3: `go build ./...` 与 `go vet ./...` 通过，无新增告警
- [ ] SC-4: `eiam/internal/cryptox` 单元测试全部通过，含 round-trip 与 KAT 兼容断言
- [ ] SC-5: 用原库 `ecmdb@v1.11.0` 生成的已知密文向量，本地 `Decrypt` 输出与明文一致（字节兼容客观证明）

consistency_check_result:
  status: pass
  pairs_checked: 6
  conflicts_found: 0

## Next Steps

- 在 `/quick-tasks` 生成任务前，确认 KAT 生成时序（删依赖前用原库固化密文向量）。