---
feature: "dead-modules-cleanup"
created: "2026-09-09"
status: completed
mode: quick
---

# Feature (Quick): dead-modules-cleanup

<!-- Status flow: tasks -> in-progress -> completed -->

## Documents

| Document | Path |
|----------|------|
| Proposal | ../../proposals/dead-modules-cleanup/proposal.md |
| 基线留档 | baseline-record.md |
| 验收留档 | acceptance-record.md |

## Tasks

| ID | Title | Status | File |
|----|-------|--------|------|
| 1 | 记录清理前基线（build/vet/test 失败集合留档） | completed | [tasks/1-record-baseline.md](tasks/1-record-baseline.md) |
| 2 | tmp 工具留档：两只入库存档 + tmp-k8s-cred 摘录留档 | completed | [tasks/2-archive-tmp-tools.md](tasks/2-archive-tmp-tools.md) |
| 3 | 删除三死模块与三只 tmp 工具目录 | completed | [tasks/3-delete-dead-modules.md](tasks/3-delete-dead-modules.md) |
| 4 | account/asset 正名：删死外壳 + 按代码事实重写 doc.go | completed | [tasks/4-rename-account-asset.md](tasks/4-rename-account-asset.md) |
| 5 | 决策文档 0001 Phase 4 补记 asset 归属决断项 | completed | [tasks/5-decision-doc-linkage.md](tasks/5-decision-doc-linkage.md) |
| 6 | 全量验收：失败集合子集比对 | completed | [tasks/6-final-acceptance.md](tasks/6-final-acceptance.md) |
| review-doc | 文档质量审查（quick 模式交付物） | completed | [tasks/review-doc.md](tasks/review-doc.md) |
| fix-1 | 修复 cert/domain K8sCredential 基线红灯（bson 标签对齐 schema.sql） | completed | [tasks/fix-1.md](tasks/fix-1.md) |

## Execution Summary

全部任务已完成（e-cam-service 提交：存档 be3bd99/c4cfd70、删除 dc3eee0、正名 0a7bcbc、基线修复 a15ebc0）。SC-5 验收通过：清理后失败集合 ⊆ 基线失败集合（5 ⊆ 6，唯一减少项即 a15ebc0 修复的 cert/domain 存量红灯），详见 acceptance-record.md。
