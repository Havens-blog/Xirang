---
id: "2"
title: "tmp 工具留档：两只入库存档 + tmp-k8s-cred 摘录留档"
priority: "P0"
estimated_time: "1h"
complexity: "medium"
dependencies: [1]
surface-key: ""
surface-type: "api"
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 2: tmp 工具留档：两只入库存档 + tmp-k8s-cred 摘录留档

## Description
三只 tmp-* 诊断工具均未入 git（`git ls-files` 为空），直接删除即永久丢失。按提案 Risk 4/Risk 5 分档处置：`cmd/tmp-cas-diag`、`cmd/tmp-k8s-diag` 确认无硬编码凭证后整目录提交入库（存档 commit）；`cmd/tmp-k8s-cred` 是凭证处理工具（main.go:55 解密 AccessKey 调云 API），不得整目录入库，仅摘录诊断逻辑（Mongo 查询结构与调用链说明）留档至提案目录。

## Reference Files
- `docs/proposals/dead-modules-cleanup/proposal.md` — In Scope; Risks & Mitigations (Risk 4/Risk 5); Success Criteria (SC-1)

## Acceptance Criteria
- [ ] `cmd/tmp-cas-diag`、`cmd/tmp-k8s-diag` 各有一个存档 commit（`git log --oneline -- cmd/<dir>` 非空，提交信息注明存档用途）
- [ ] 入库前对两只工具 grep 硬编码凭证并记录结果（提案 Evidence 已核实均经 viper `v.GetString` 读配置）
- [ ] `cmd/tmp-k8s-cred` 诊断逻辑摘录文件存在于 `D:\Haven\docs\proposals\dead-modules-cleanup\`，含 Mongo 查询结构与调用链说明、不含凭证读取代码

## Hard Rules
- `cmd/tmp-k8s-cred` 不得整目录入库（涉敏调用链永久写入 git 历史不可逆，Risk 5）

## Implementation Notes
- 存档 commit 与后续删除 commit 分离，便于 revert 定位
- 摘录文件命名建议 `archive-tmp-k8s-cred-notes.md`
