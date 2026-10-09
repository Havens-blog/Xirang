# Iteration 0 Report (Pre-Revision — Freeform Findings)

> 本轮由自由评审（领域专家）驱动，非 rubric 评分。以下 8 项发现需在 proposal 中修订。

## Attack Points

1. **[high]** refCount 为 null/未扫描的过期证书可能被误判为「无引用」而误藏，丢失漂移风险信号
   - quote: "如果 `refCount IS NULL` 被当作 0 处理，这些过期证书会在默认视图静默消失——而这恰恰可能是一张刚过期、引用扫描还没跑完、线上仍在用的证书。"
   - improvement: 给 refCount 判定加保守兜底——仅「扫描已完成且显式为 0」才判为无引用，null/未知一律视为有引用（显示）；后端单测覆盖「refCount=null 的过期证默认可见」「fingerprint_only 且 refCount=0」两个用例。

2. **[high]** 组合过滤下隐藏计数 N 的口径未定义，可能产生新的「11=11」式口径漂移
   - quote: "如果 N 恒为全局值（当前 1），而用户搜索的关键词命中结果里根本没有被隐藏的行，提示「已隐藏 1 张」就是错的——用户点「查看」后行数不变，提示与行为对不上。"
   - improvement: 隐藏计数 = 满足当前其余过滤条件（search/hostingStatus）但被隐藏的行数，由服务端在同一响应返回两个口径（visibleTotal + hiddenCount），禁止前端推算；三方闭合限定在同一过滤上下文。

3. **[medium]** daysLeft=expired 豁免收敛在哪一层未定死，URL 直达存在绕过路径
   - quote: "这一层不定死，实现者很可能选最省事的前端豁免，把风险留给每一个入口。"
   - improvement: 豁免由后端强制执行——exclude 参数不改变 daysLeft=expired 语义，前端无豁免逻辑；澄清约束为明确决策。

4. **[medium]** 切换状态与筛选状态机、分页行为未定义，恢复语义存在自洽缺口
   - quote: "若用户停在非第一页（例如第 2 页），切换后该页可能超出新的 total 范围，是重置到第 1 页还是显示空页？"
   - improvement: 定义状态机：用户切换态仅作用于默认视图；daysLeft=expired 时挂起（显示全部）；清除筛选后回到切换态；切换导致 total 变化时重置到第 1 页。

5. **[medium]** 「无新增查询」性能声明与响应携带隐藏计数存在实现矛盾
   - quote: "实现者照着「无新增查询」做，最可能的偷懒是让前端用某种方式推算隐藏计数——那就又回到了前端口径漂移。"
   - improvement: 改写性能声明为「复用 listCertsApi 主查询，仅新增一次同条件统计口径（或单查询双 total），不新增独立接口」。

6. **[medium]** 空态下「已隐藏 N 张」提示条是否可见未定义，残余静默吞行路径
   - quote: "如果此时提示条不在空态视图中，「有证书被隐藏」这个事实只剩一行小字，静默吞行以最彻底的形式复现。"
   - improvement: 空态视图中仍显示「已隐藏 N 张…」提示与「查看全部」入口（提示条不随表格主体消失）；SC 第二条补空态场景。

7. **[medium]** 被隐藏的无引用过期证书恰恰是「可删除」的可操作对象，隐藏使清理入口变深
   - quote: "隐藏它等于把「可清理的对象」藏了起来——用户失去了发现「这里有一张过期证可以顺手删掉」的机会。"
   - improvement: 行内提示文案体现「可删除」属性（如「已隐藏 N 张无引用过期证书（可删除）· 查看全部」），或在 Key Scenarios 记录该取舍。

8. **[medium]** 验收标准硬编码当前快照数字 11，换证推进后将变成假断言
   - quote: "把字面 11 写进 Success Criteria，意味着今天验收通过、两周后换证完成数量变成 8 时，这条验收标准就变成了一条永远无法满足的假断言。"
   - improvement: Success Criteria 第四条改为动态相等断言：daysLeft=expired 筛选行数 = 看板「已过期」卡计数（不写死数值）。
---

# Baseline Scorer Attacks (762/1000) — 合并入预修订

9. **[行业基准]** 全节无具体产品/开源项目引用 — quote: "「默认隐藏已过期」是常见做法（如证书管理工具的过期证书归档/折叠）" — improvement: 补具体行业方案名（cert-manager、商业证书管理产品过期处理行为）并说明取舍。
10. **[风险评估]** 风险表无可能性×影响评级 — quote: "| # | Risk | Mitigation |" — improvement: 补可能性×影响评级列。
11. **[成功标准]** SC 硬编码当期快照数字 — quote: "11 行 = 看板「已过期」卡计数 11" — improvement: 改为动态相等断言（行数=卡计数），不写死数值。（→ 攻击#8 同源）
12. **[解决方案清晰度]** daysLeft=expired 豁免收敛层未定死（前端不传参 vs 后端强制）— quote: "新增「排除无引用过期」参数需与 daysLeft/expired 联动的语义" — improvement: 定死豁免落点 + 定义切换态×筛选态状态机与分页重置。（→ 攻击#3/#4 同源）
13. **[需求完整性]** NFR「无新增查询」与 In Scope「响应返回隐藏计数」矛盾，诱导前端推算计数 — quote: "性能：无新增查询（复用 listCertsApi，新增一个布尔过滤参数）。" — improvement: 改写为诚实声明（同条件双统计口径 visibleTotal + hiddenCount），计数按当前过滤上下文重算。（→ 攻击#2/#5 同源）
14. **[逻辑一致性]** 过滤谓词 "expired AND refCount=0" 无法保证「被引用过期始终显示」承诺 — quote: "被线上引用的过期证书始终显示（风险信号）" — improvement: 谓词改为 `expired AND referenceStatus=no_refs_scanned`，blind_spot/未知保守显示并补单测。
15. **[blindspot]** 核心谓词与代码库已固化三态语义冲突：reference_status.go 明文「未发现引用」≠「无引用」……blind_spot 表示无成功快照或扫描范围未覆盖，禁止据此放行删除；refCount=0 同时覆盖 no_refs_scanned 与 blind_spot ⇒ 信息最少、最可能藏线上在用过期证。— quote: "过滤条件严格限定 `expired AND refCount=0`；单测锁定「有引用过期可见」" — improvement: 改用引用三态（no_refs_scanned）作判定，blind_spot 保守显示，补 refCount=null/blind_spot 单测用例与风险条目。（→ 攻击#1 深化）
16. **[blindspot]** 无回滚/降级方案 — quote: "后端 `GET /certs` 新增过滤：排除「已过期且 refCount=0」的行" — improvement: 补关闭开关或阈值自动降级缓解，风险表中登记。
17. **[blindspot]** Evidence 内部矛盾——被定义为不可操作的行恰是唯一可操作清理对象 — quote: "不可操作、无风险信号" vs "非保护期，可删除" — improvement: Key Scenarios 承认取舍，或提示文案带「可删除」属性。（→ 攻击#7 深化）
18. **[blindspot]** 文案三处漂移（"点击查看"/"查看全部"/"再次点击隐藏"）— quote: "工具栏显示行内提示「已隐藏 N 张无引用过期证书 · 点击查看」" — improvement: 统一 canonical 文案（默认态/展开态两个字符串）。

---

## Reviser Edit Results (Phase 0.5 Pre-Revision)

修订完成：11 个段落实质修订（Proposed Solution/Innovation/Key Scenarios/NFR/Constraints/Industry/In Scope/Risks/SC/Problem 等），11 处 `<!-- pre-revised -->` 标注（high 8 / medium 3）。18 攻击点全部按处置表处理（同源合并：#1/#15、#2/#5/#13、#3/#12、#8/#11、#7/#17）。修订后迭代 1 评分 902/1000（≥900 通过）。
