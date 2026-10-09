# Eval Report — Iteration 0 (Pre-Revision: Freeform Findings)

- **iteration**: 0
- **title**: Pre-Revision (Freeform Findings)
- **source**: `eval/freeform-review.md`（专家：go-refactor-architect，复用）
- **extraction**: 13 条提取 / 13 个标记段落，hit_rate = 1.0（100%）
- **triage**: 6 个唯一攻击点（13 条提取去重合并）；0 主观偏好项；0 边界待裁项

## ATTACK_POINTS

- **[high]** Risk 4 的"可从 git 历史找回"对 tmp 工具不成立（三只均未入库，含盘点遗漏的 tmp-k8s-diag），删除即永久丢失诊断逻辑 | quote: "风险：Risk 4 的缓解措施\"需要时可从 git 历史找回\"对它要保护的对象是假的——两只 tmp 工具根本没进过 git 历史。" | improvement: In Scope 增补 `cmd/tmp-k8s-diag`；Risk 4 缓解改为"先提交、再删除"（三只 tmp 工具先入库或摘录留档）
- **[medium]** 测试基线非全绿且未钉住，SC-5"全部通过"在纯删除前提下不可满足 | quote: "风险：基线\"当前全部通过\"不成立——实测当前测试套件就是红的，SC-5 按原文写下\"全部通过\"在纯删除的前提下无法达成；基线同时未被钉住，测量发生在 79 项未提交变更的脏工作区上。" | improvement: SC-5 改为"通过/失败集合与清理前基线一致"，并写明先记录基线失败集合（如 internal/cert/domain 存量红灯）作为证据
- **[medium]** account 正名文案与代码事实不符：只读定性被 mcp SyncAccount（nil 队列 panic 坏路径）反驳，消费方清单漏 internal/cam/task | quote: "风险：account 的\"正名\"文案与代码事实不符，新 `doc.go` 将出生即不准确——与本次清理要治的病是同一个病。" | improvement: Solution 中 doc.go 文案要求改为"按代码事实重写：消费方列全（含 cam/cost、cam/task、cert、logquery、mcp、ioc），如实登记 SyncAccount nil 队列坏路径"
- **[medium]** 注释级悬空引用盘点不完整（live 代码约 10 处"internal/task 框架"注释），提案自己的收益（文档与事实一致）只兑现一半 | quote: "风险：Evidence 节对注释级引用的盘点不完整，删除后将在 live 代码里留下 10 处指向已删包的悬空注释，提案自己定义的收益（文档与事实一致）只兑现了一半。" | improvement: Out of Scope 显式登记 follow-up（清理 live 代码中指向已删包的注释），写明不纳入本阶段的理由（纯删除边界，不动 live 代码）
- **[low]** 决策文档路径锚点是工作区根而非仓库根，与其余路径坐标系不一致 | quote: "问题：决策文档的路径锚点是工作区根，不是仓库根，与其余所有路径的坐标系不一致。" | improvement: In Scope 第 5 项补明确路径锚点（`D:\Haven\docs\decisions\0001-...`，工作区根，非 e-cam-service 仓库内）
- **[high]** SC-3"描述与消费方事实一致"缺机械验收判据 | quote: "建议：给 SC-3 补一个可机械执行的验收判据，避免\"描述与消费方事实一致\"沦为自由心证。" | improvement: SC-3 补机械判据——doc.go 中消费方清单与 import 语句级 grep 结果一一对应（多一个少一个都算失败）

## BORDERLINE_FINDINGS

（无——全部发现均分类为可直接编辑的事实修正或可验证不一致的结构性建议）

## SKIPPED_FINDINGS

（无主观偏好项）

## rubric

（all dimensions）: N/A
