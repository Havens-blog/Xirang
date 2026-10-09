# Eval Proposal Complete — dashboard-cert-granularity

**Final Score**: 924/1000 (target: 900) ✅ **Target reached**
**Iterations Used**: 3/3 rubric rounds + freeform pre-revision (iteration 0) + 1 user-authorized re-evaluation (iteration 4)

## Score Progression

| Iteration | Score | Delta | Note |
|-----------|-------|-------|------|
| Baseline (pre-revision, informational) | 815 | — | CTO scorer, codebase verified (lenient) |
| 1 | 734 | −81 | CTO scorer, code access forbidden (strict verification stance) |
| 2 | 889 | +155 | after revision round 1 |
| 3 | 878 | −11 | after revision round 2 |
| 4 | 924 | +46 | after hand-applied裁决 (10 residual gaps closed) |

> **方法论口径说明（诚实记录）**：iteration 1 的 734 与 baseline 815 不可直接比较——baseline 评分者被允许读源码把断言核为「已证实」，iteration 1+ 采用严格 verification stance（禁读码、未自证断言一律扣分）。734 反映的是早稿对「实现符号/部署序/提交 SHA」缺乏文内证据的真实弱点，非文档退化。iteration 4 与 iteration 3 同口径（禁读码），+46 分真实反映最后一轮裁决的改善。

## Dimension Breakdown (final, 924)

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 103 | 110 |
| Solution Clarity | 114 | 120 |
| Industry Benchmarking | 111 | 120 |
| Requirements Completeness | 101 | 110 |
| Solution Creativity | 90 | 100 |
| Feasibility | 92 | 100 |
| Scope Definition | 77 | 80 |
| Risk Assessment | 87 | 90 |
| Success Criteria | 69 | 80 |
| Logical Consistency | 80 | 90 |

## Outcome

**Target reached — 924 ≥ 900。**

最后一轮手工裁决闭合了 iteration-3 的全部 10 条：豁免边界（仅到期风险维度搜索触发，云/托管属性筛选不豁免）、fetch 接线（豁免→includeHidden=true 重拉）、localStorage 定案、countsByLevel {total,visible,hidden} 契约、具名 benchmark + 孤儿处理对比、8 组门禁并入任务估算、分页首期不做改观测阈值、referenceStatus 安全裁决、抽屉 SAN 上限、回滚窗口兼容。

## Residual Issues（达标后的残余，供实现阶段注意，不再迭代）

合法性小矛盾（建议生成任务前 1 句裁决即可）：
1. 「总规则」措辞：diff/exempt 是探测维度卡，却触发「到期风险维度」豁免——需在措辞上把传入维度统一为「到期/探测风险维度」
2. 隐藏范围语义二选一：云过滤下孤儿是可见还是隐藏，需统一「不豁免=仅当前可见行集再过滤」与「仅无筛选默认隐藏」两处字面
3. SC5 徽标排序尾部缺「未探测」与「排序尾」，需与 Key Scenario 对齐为单一顺序
4. SC 用「SAN>20」魔法阈值，需锚定到 NFR/需求并说明依据
5. Proposed Solution 与 In Scope 对 countsByLevel 键名不一（visible/hidden vs total,visible,hidden），以 In Scope 契约为准

实现阶段注意事项：
6. 工作量「2~3 任务」与 9 组 SC 门禁 + 10 风险缓解的粒度不符，建议按门禁聚类拆任务
7. 错误路径仅一行，需补 API 失败 / 加载 / 部分探测降级三态
8. 危害量化：仅 94 证/93 域 +1 行，需统计实际受掩盖证书数支撑「风险盲区」表述
9. 只读角色看板→台账路径中断（有权限角色靠菜单入口，需验证只读角色是否有台账可见入口）
10. 隐藏谓词缺扫描新鲜度契约：最近获得引用的证可能仍显示 no_refs_scanned 被静默隐藏——需复用台账的新鲜度阈值 + 展示 lastScan 时间
11. 「用户点击卡片」意图主张基于作者 5 Whys 自反而非用户研究，应标注为假设
12. localStorage 持久化缺键名命名空间与回滚交互（前端回退到旧版本时）

其中 #10（扫描新鲜度契约）是唯一需要在实现阶段**实质处理**的项，其余为措辞/粒度一致性问题。