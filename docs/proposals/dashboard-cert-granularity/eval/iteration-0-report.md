---
iteration: 0
title: "Pre-Revision (Freeform Findings)"
note: "提取命中率 24/24=100%（评审原叙见 freeform-review.md）；以下为聚合后 ATTACK_POINTS。标签 <!-- pre-revised: {severity} --> 将标于 Reviser 修改段。"
ATTACK_POINTS:
  - - **[high]** 特殊卡闭合判据自相矛盾：把"计数与卡片值一致"重写为"域名级计数 ⇔ 证书行集合"的显式关系，筛选谓词由"行徽标相等"改为"行内任一 SAN 具该状态" | quote: "风险：特殊卡的闭合判据自相矛盾——\"计数与卡片值一致\"在证书粒度主表下原理上不可达，且最差优先聚合会让低优先级状态（exempt）从行上消失、无法被筛选命中。" | improvement: 重写 SC-5 与特殊卡交互规范：diff/exempt 卡计数保持域名级语义（diffAlertCount/exemptCount 不动），点击后的行集合 = 「存在任一 SAN 探测态为该状态的证书行」；匹配谓词以行内 SAN（非行徽标）为准，杜绝 exempt 徽标坍缩不可筛
  - - **[high]** 分级卡"动态相等"×孤儿隐藏×countsByLevel 含孤儿 三方交互未定义：显式移植台账"分级筛选激活即豁免孤儿隐藏" | quote: "风险：分级卡\"动态相等\"×孤儿默认隐藏×countsByLevel 含孤儿 三者的相互作用未定义，台账已解决的\"过期筛选豁免\"语义没有被显式移植到看板侧，11=11 的快乐路径暗含了一个未声明的规则。" | improvement: 在交互规范与 SC 中显式写入：任何一级分级卡被激活（页内筛选生效）即豁免孤儿隐藏（与台账 daysLeft=expired 豁免同构，统一为"主动分级查看=查看该档全量证书"）；配 11=11 的联调断言
  - - **[high]** 隐藏执行位置未定 + 载荷缺 referenceStatus：判定单点 + 行级字段补全 | quote: "问题：隐藏的执行位置（服务端过滤 vs 客户端过滤）未定义，而 \"items……（每证一行，含孤儿）\" + \"API 返回 hiddenCount\" + 前端 \"孤儿默认隐藏 + 切换入口\" 的组合会迫使前端用 TypeScript 第三次重写隐藏判定，直接违反 proposal 自己的\"单源\"工程核心；且现行 `DashboardItem` 根本没有 `referenceStatus` 字段，前端无从判定谁是孤儿。" | improvement: 载荷契约补齐行级 referenceStatus 字段；隐藏由服务端执行（复用 ledger 三态判定单点），前端仅消费 visible 行 + hiddenCount，不派生三态
  - - **[high]** 卡片双口径需每档隐藏数，来源缺失：规定 per-level hidden 数据来源，解禁"countsByLevel 语义不动" | quote: "问题：\"卡片与表均以'可见 N · 隐藏 M'双口径呈现\"需要**每档**隐藏数，但 proposal 只规定了全局 hiddenCount，同时用 \"countsByLevel 语义不动\" 封死了扩展口——数据来源缺失。" | improvement: 明确 per-level 隐藏数来源：countsByLevel 升级为「每档 {visible, hidden}」或后端另返回 per-level hidden 数组；去除"语义不动"封口，写明兼容消费端
  - - **[high]** 多 SAN 行复用"现成单域抽屉"假设不成立：定义多 SAN 行抽屉进入语义 | quote: "风险：行点击\"仍进入现成的域名级探测详情抽屉\"的复用假设对多 SAN 行不成立——现成抽屉是单域名契约。" | improvement: 定义证书行的抽屉契约：头部证书摘录 + 该证全部 SAN 探测列表（默认最差优先排前）；行点击统一进入该多 SAN 视图而非单域抽屉
  - - **[high]** blind_spot 可见性无成功判据锁定：升格为成功判据，禁止裸 refCount 判定 | quote: "风险：blind_spot（盲区）的可见性没有被任何成功判据锁定，而它正是\"有引用的过期证书必须始终可见\"原则的另一半。" | improvement: SC 增加「blind_spot / 未知引用状态的过期证书默认可见」判据；隐藏判定严格限定 expired AND no_refs_scanned（三态明确），单测锁定 blind_spot 与 has_refs 反向可见
  - - **[medium]** 聚合徽标域级基准与证书行粒度错位：显式徽标行级语义 + 并列 tie | quote: "风险：聚合徽标的语义基准是域级\"任一归属证书一致\"（`probe_service.go` 的 `ownership[name][onlineFP]`），与证书行粒度存在错位——过期旧证行会显示\"一致\"，用户会误读为\"这张过期证在线上生效\"；同时最差优先序存在未定义的并列与部分探测场景。" | improvement: 显式定义徽标文案与语义（如「该证 SAN 现网一致（归属另一证书）」区分于「本证线上生效」）+ 并列与部分探测的 tie 规则（取该证 SAN 中最差态，未探测视为最低优先）
  - - **[medium]** wildcardSkippedCount 不在域名口径保护名单：纳入保护并防重计 | quote: "风险：wildcardSkippedCount 不在 Out of Scope 的域名口径保护名单里，证书循环直出会让共享通配符 SAN 的多张证书被重复计数。" | improvement: wildcardSkippedCount 纳入 Out of Scope 域名口径保护名单；证书循环直出时通配符跳过计数按 domain 去重，防重计单测锁定
  - - **[medium]** 三态判定访问模式未规定：规定批量访问免 N+1 | quote: "风险：三态判定的访问模式未被规定——台账只对\"过期且计数为 0\"的候选做派生，看板若为每行标注三态就可能引入逐证 N+1 查询。" | improvement: NFR 注明三态判定为服务端批量内存判定（一次 refs/snapshot 拉取），仅对过期且 refCount=0 候选做派生，逐证 N+1 列为反例
  - - **[medium]** row-key=domain 失效 + 云选项含孤儿云 + aria-live total 基准：三处可见性管线修正 | quote: "风险：`DashboardTable` 现用 `row-key=\"domain\"`，证书粒度化后域名不再唯一，行 key 失效会引发 Vue 复用错乱；连带云筛选选项会把隐藏孤儿的云也列为可选项、选中即空态。" | improvement: 行 key 改为证书 ID；云筛选选项来源过滤被隐藏孤儿；aria-live 通告 total 基准 = 可见 total（组合筛选通告基准明确定义）
  - - **[low]** NFR"维持排序"实为变更 + 并列 notAfter tie 未定义 | quote: "问题：NFR 声称\"维持排序输出（到期优先）\"，但现状是域名字典序——这不是维持而是变更；且证书粒度下并列 notAfter 的 tiebreak 未定义。" | improvement: 修正表述为「排序由域名字典序改为到期优先（notAfter asc）」；定义并列 notAfter tie（指纹字典序稳定取一）
  - - **[low]** 验收断言字面加固（排序、aria-pressed、grep 断言） | quote: "问题：两处验收表述与代码现实有字面出入，会让判据在执行时被绕过。" | improvement: SC 断言用词与实现对齐（含 grep 无 router.push 残留、aria-pressed 选中态、聚合序落地单测）
BORDERLINE_FINDINGS: []
SKIPPED_FINDINGS:
  - - "无主观偏好类发现——24 项提取均已吸收进 ATTACK_POINTS（聚合后 12 条）"
rubric:
  Problem Definition: N/A
  Solution Clarity: N/A
  Industry Benchmarking: N/A
  Requirements Completeness: N/A
  Solution Creativity: N/A
  Feasibility: N/A
  Scope Definition: N/A
  Risk Assessment: N/A
  Success Criteria: N/A
  Logical Consistency: N/A