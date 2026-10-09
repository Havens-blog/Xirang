# Freeform Expert Review

评审视角：证书生命周期管理 / 列表过滤与分页口径一致性 / 企业级运维台账 UI / 服务端查询语义设计。评审前已对照代码核实 proposal 引用的关键事实：`e-cam-service/internal/cert/service/dashboard_service.go`（`CountsByLevel` 证书粒度循环、`owner` 域名收敛、`probeByDomain` 域级探测、通配符计数）、`e-cam-web/src/views/cert/dashboard/index.vue:158-165`（跳台账决策注释）、`OverviewCards.vue`（分级卡无 `aria-pressed`、特殊卡有）、`format.ts`（`filterDashboardItems` 精确匹配 `probeStatus`、`filterAnnouncement` 文案）、`DashboardTable.vue`（`row-key="domain"`、豁免列 `isExemptRow`）、`ProbeDetailDrawer.vue`（单域名契约）、`ledger_service.go`（三态判定 `deriveRefStatusFor`、`daysLeft=expired` 强制豁免、`HiddenCount` 同上下文闭合）、`probe_service.go`（`probeOne` 六态判定，`ownership[name][onlineFP]` 为"任一归属证书一致"基准）、`api/cert.ts` 的 `DashboardItem`（无 `referenceStatus`/`sans`/CN/issuer 字段）、`getCertDashboardApi` 单消费端（仅 dashboard 页引用）。

## Background Assessment

这份 proposal 要解决的问题是真实且我亲历过的那一类：看板分级卡计数是证书粒度（`CountsByLevel` 对全部证书逐张计数，代码属实），主表却是域名粒度（`owner` map 按 notAfter 最新者独占一行），于是"已过期 11"卡点进去要么跳台账、要么页内筛出与 11 对不上的域名行。proposal 选择把主表重铸为证书粒度、分级卡改页内筛选、探测聚合成行级徽标、孤儿证书与台账同构默认隐藏——方向正确，且与刚交付的 ledger-hide-expired-orphans 共用同一三态判定函数的想法（"两页面隐藏口径单源不漂移——这是本次设计的工程核心"）正是这类改造唯一正确的做法。

proposal 的证据链扎实：跳台账是妥协而非设计目标（index.vue 注释原文属实）、单消费端属实、`CountsByLevel` 本就是证书粒度属实、台账侧 `daysLeft=expired` 强制豁免与 `includeExpiredNoRefs` 双口径机制属实。Alternatives 表对方案 B 的否决理由（"卡计数与抽屉行数两来源"）说明作者理解闭合问题。Out of Scope 对台账隐藏逻辑、只读边界、处置流程的克制也到位，没有借机扩张归档生命周期之类的语义。

但作为对"过滤后 total 与行数漂移"有事故经验的人，我必须指出：这份 proposal 在**最需要闭合断言的地方恰好没有闭合断言**。它反复宣称"动态相等""双口径""同构隐藏"，却没有定义过滤态矩阵下（默认视图 / 分级卡激活 / 查看全部 / 特殊卡激活 / 组合工具栏筛选）可见行数、隐藏数、卡片数、API hiddenCount 四方谁等于谁、由谁计算。特殊卡的闭合判据甚至自相矛盾（下详）。此外它对探测聚合的语义基准（域级"任一归属证书一致"）与证书行粒度之间的错位完全没有着墨。以下按严重度展开。

## Key Risks

风险：特殊卡的闭合判据自相矛盾——"计数与卡片值一致"在证书粒度主表下原理上不可达，且最差优先聚合会让低优先级状态（exempt）从行上消失、无法被筛选命中。【CRITICAL】
> "diff/exempt 卡点击 → 主表筛出含对应探测徽标的证书行，计数与卡片值一致" 与 Out of Scope "diffAlertCount/exemptCount 保持域名级计数不动" —— 两处叠加构成一个不可能三角：diffAlertCount 数的是"最新探测为 diff 的域名数"，主表行是证书。一个 diff 域名的 SAN 若同时挂在过期旧证与有效新证上（这正是本 proposal 的核心场景"同名多证"），筛选出 2 行而卡片显示 1，判据必挂。更糟的是最差优先序 `diff > change_linked_diff > unreachable > wildcard_skipped > exempt` 把聚合徽标坍缩成单值：一张证书含一个 diff SAN 和一个 exempt SAN 时行徽标只显示 diff——现行前端 `filterDashboardItems` 的特殊卡筛选是 `it.probeStatus !== 'exempt'` 的精确匹配，迁移到"含对应探测徽标"后这张证书**永远不会被 exempt 卡筛中**，尽管它确实包含豁免 SAN。这是本 proposal 要消灭的那一类"点击按分级过滤为空"的新变体。若不修正判据定义，联调验收时要么判据被悄悄放松、要么实现被写成"行数必须等于卡片数"的错误断言。

风险：分级卡"动态相等"×孤儿默认隐藏×countsByLevel 含孤儿 三者的相互作用未定义，台账已解决的"过期筛选豁免"语义没有被显式移植到看板侧，11=11 的快乐路径暗含了一个未声明的规则。【CRITICAL】
> "见'已过期 11'卡，点击 → 主表就地筛出 11 张过期证书行" 与 "孤儿：过期且 referenceStatus=no_refs_scanned 默认隐藏" 与 "保持 summary/countsByLevel 语义不动" —— countsByLevel 对**全部**证书计数（含孤儿，代码 `view.Summary.CountsByLevel[levelIdxByTier[levelOf(c.NotAfter, now)]]++` 在证书循环内无条件自增），所以"已过期 11"包含 M 张孤儿。默认视图隐藏孤儿后，点击过期卡若仍执行隐藏，表中只有 11−M 行，"主表行数 == 该档卡片计数（动态相等）"即告失守。台账侧对此已有成熟答案——`ledger_service.go` 明写 "daysLeft=expired 后端强制豁免（看板「已过期」卡口径一致）"——proposal 在 Constraints 里说 "daysLeft expired 豁免语义——必须复用同一实现，不重写"，但通篇没有把这条规则落到看板自己的交互规范里：点击过期卡算不算"主动筛选过期"、孤儿是否随之显现、此时卡片"可见 N · 隐藏 M"显示什么（此时应显示 M=0）。三处（看板卡、看板表、台账 expired 视图）任何一处实现者理解不一致，就会复刻 proposal 自己引用的事故："用户刚验收的台账'已过期'数与看板卡片数不一致"。这是全篇最高优先级必须补的规范缺口。

问题：隐藏的执行位置（服务端过滤 vs 客户端过滤）未定义，而 "items……（每证一行，含孤儿）" + "API 返回 hiddenCount" + 前端 "孤儿默认隐藏 + 切换入口" 的组合会迫使前端用 TypeScript 第三次重写隐藏判定，直接违反 proposal 自己的"单源"工程核心；且现行 `DashboardItem` 根本没有 `referenceStatus` 字段，前端无从判定谁是孤儿。【HIGH】
> "items 改为证书行（每证一行，含孤儿）" 与 "API 返回 hiddenCount，切换后行数 == 可见+隐藏" —— 若 items 含孤儿且隐藏由前端执行，隐藏谓词（expired && no_refs_scanned，且 blind_spot 豁免、has_refs 豁免）就要在 Go（ledger）、Go（dashboard，若标注）、TS（format.ts）出现第三份表达；台账特意把判定"收敛于本方法一处"的工程成果被无声拆散。若隐藏由服务端执行，则"含孤儿"与"切换查看全部"需要第二次请求或 `showAll` 参数，proposal 又没写。两条路都通，但必须选一条并写死。同时前端行上要渲染"引用徽标"（快乐路径："从行上引用/托管/探测徽标即可判断处置优先级"）、要派生每档隐藏数，都依赖行级三态——现行 `DashboardItem`（`api/cert.ts:358-373`）只有 `referencedClouds`（引用云 chips），没有 `referenceStatus`。In Scope 的字段清单（"CN、SANS 折叠计数、issuer、daysLeft、托管、引用云、探测徽标"）里也没有它。载荷契约缺字段，实现时必然有人现场发明。

问题："卡片与表均以'可见 N · 隐藏 M'双口径呈现"需要**每档**隐藏数，但 proposal 只规定了全局 hiddenCount，同时用 "countsByLevel 语义不动" 封死了扩展口——数据来源缺失。【HIGH】
> "卡片与表均以'可见 N · 隐藏 M'双口径呈现" 与 "保持 summary/countsByLevel 语义不动（本就是证书粒度）" 与 "API 返回 hiddenCount" —— 五张分级卡若都要显示各自档位的隐藏数，单个全局 hiddenCount 推不出来（隐藏只发生在 expired 档时可以偷懒，但规范不能建立在"现在只有 expired 档会隐藏"这种偶然上）；要么把 countsByLevel 扩展为 `{total, hidden}` 对（这改动口径，需在 Out of Scope 里解禁并说明与台账 `visibleTotal/HiddenCount` 的对应关系），要么行级带 `referenceStatus` 让前端派生并用单测断言"前端派生值 == API hiddenCount"闭合。二者必择其一，现在是一个都不成立。

风险：行点击"仍进入现成的域名级探测详情抽屉"的复用假设对多 SAN 行不成立——现成抽屉是单域名契约。【HIGH】
> "行点击仍进入现成的域名级探测详情抽屉" 与 "徽标带 tooltip 注明冲突 SAN 明细，抽屉保留域名级真相" —— `ProbeDetailDrawer` 的模板直接渲染 `item.domain`、单个 `probeStatus`、单个线上/台账指纹（代码核实）。证书行聚合 N 个 SAN、N 个探测结果时，行点击该打开哪个域名？proposal 承诺了 tooltip 里的 SAN 明细，却没定义抽屉的进入语义：是默认打开最差 SAN、还是抽屉内增加 SAN 切换列表、还是行内 SAN 徽标可点分别打开。"现成的抽屉"四个字掩盖了一个必须设计的交互决策，而这恰恰是只读用户从"看到异常"到"看清异常"的最后一跳。

风险：blind_spot（盲区）的可见性没有被任何成功判据锁定，而它正是"有引用的过期证书必须始终可见"原则的另一半。【HIGH】
> "孤儿隐藏双口径（visible + hiddenCount，复用 ledger 三态判定，过期豁免恒显示）" —— proposal 全文没有出现 blind_spot 的处置。台账实现是保守的：无成功快照（found=false）一律 blind_spot 保守显示、`deriveRefStatusFor` 禁止裸 refCount 判定（"refCount=0 混淆 no_refs_scanned 与 blind_spot"）。复用三态函数固然正确，但"复用"是实现手段不是行为承诺——成功判据里只有 "默认视图不含'过期且 no_refs_scanned'行"，没有 "过期且 blind_spot 恒显示"。将来任何人把判定"简化"成 refCount==0 或把无快照当作无引用，盲区里的过期证书就会静默消失——而 blind_spot 的语义是"引用状态未知"，未知不等于安全（台账删除拦截对 blind_spot 一律 409，正是同一逻辑）。这类"判定放宽导致风险证据消失"的事故在证书台账场景里是真实发生过的形态。

风险：聚合徽标的语义基准是域级"任一归属证书一致"（`probe_service.go` 的 `ownership[name][onlineFP]`），与证书行粒度存在错位——过期旧证行会显示"一致"，用户会误读为"这张过期证在线上生效"；同时最差优先序存在未定义的并列与部分探测场景。【MEDIUM】
> "一张证书多个 SAN 探测结果不一（一 diff 一 consistent）→ 聚合取最差优先序" —— 探测的 consistent 判定基准是该 SAN 在台账中的**任一**归属证书指纹集合，不是行上这张证书。同名多证并存时，线上生效新证 → 域级 consistent → 过期旧证行也挂"一致 ✓"。对处置过期证书的操作员来说，这是危险的误读源（旧证行看起来"健康"）。另外序表把 "健康/未探测" 并列在末位：一张证书一个 SAN 一致、另一个 SAN 从未探测，徽标显示什么？未探测按 Key Scenarios 是空徽标"未探测"，那部分探测的行是"一致"还是"未探测"？tie 未定义，实现者会随手选一个，两处实现（徽标聚合 vs 单测样例）就会各选各的。

风险：wildcardSkippedCount 不在 Out of Scope 的域名口径保护名单里，证书循环直出会让共享通配符 SAN 的多张证书被重复计数。【MEDIUM】
> "3 张覆盖率卡与 2 张特殊卡的计数语义（diffAlertCount/exemptCount 保持域名级计数不动）" —— 按代码注释，summary 是 "5 分级卡 + 2 计数卡 + 3 覆盖率卡"，通配符跳过计数（`WildcardSkippedCount`）不在"2 张特殊卡"之列，保护名单恰好漏了它。现行实现对去重后的域名行逐个判定 `isWildcardSAN(name) && !overridden` 才计数；改成"证书循环直出"后，若顺手在证书循环里自增，`*.jlc-robot.com` 的新旧两证会把它数成 2。工具栏那句"通配符跳过拨测 N"是用户核对探测覆盖缺口 direct 的数字，翻倍即失真。

风险：三态判定的访问模式未被规定——台账只对"过期且计数为 0"的候选做派生，看板若为每行标注三态就可能引入逐证 N+1 查询。【MEDIUM】
> "将 owner 域名收敛改为证书循环直出（每证一行）……无新数据源" —— 台账的 `listCertsHideExpiredNoRefs` 特意只对候选行调用 `deriveRefStatusFor`（候选=已过期 && refCount==0 && 有快照），且快照与引用一次性拉取共享。看板要为行上"引用徽标"标注三态的话，若逐证调用 `deriveReferenceStatus`（内部每次 `LatestDone` + `ListByFingerprint`），94 张证就是 94 次往返。"无新数据源"是真的，但访问模式决定这页是不是仍然轻——而看板是只读角色的默认页。

风险：`DashboardTable` 现用 `row-key="domain"`，证书粒度化后域名不再唯一，行 key 失效会引发 Vue 复用错乱；连带云筛选选项会把隐藏孤儿的云也列为可选项、选中即空态。【MEDIUM】
> "前端 DashboardTable：证书行渲染（CN、SANS 折叠计数、issuer、daysLeft、托管、引用云、探测徽标）" —— proposal 未提 row-key 迁移（应迁 `certId`）。`row-key="domain"` 在同名多证下产生重复 key，Element Plus 表格的行状态/展开/复用会出错。另外 `cloudFilterOptions(items)` 基于全部 items 构建——若 items 含孤儿，孤儿证书的云 chip 会出现在下拉里，默认视图下选中它得到 0 行空态，用户会以为"这个云没有证书"。

风险：aria-live 通告在隐藏态下失真——"已取消筛选，显示全部 N 条"的"全部"与 total 基准在孤儿隐藏后不再成立；组合筛选下的通告计数基准也未定义。【MEDIUM】
> "沿用现有 aria-live 筛选通告与骨架屏" —— 现行 `filterAnnouncement` 在无筛选时输出 `已取消筛选，显示全部 ${total} 条`，total=items.length。items 含孤儿时，默认视图 8/11 也会念成"显示全部 11 条"，屏幕阅读器用户被告知的状态与实际渲染不符；有筛选时 shown/total 的分母是"含孤儿总数"还是"默认可见总数"，规范未定。可访问性不是一个可以事后补的装饰，它和视觉用户看到的是同一份口径。

问题：NFR 声称"维持排序输出（到期优先）"，但现状是域名字典序——这不是维持而是变更；且证书粒度下并列 notAfter 的 tiebreak 未定义。【LOW】
> "后端维持排序输出（到期优先）" 与 Key Risks "后端按 notAfter asc 排序到期优先" —— 代码现状 `Items 按 domain 字典序`（`sort.Strings(domains)`）。"维持"一词会让任务实现者以为排序无关紧要而不写测试。证书粒度下同 notAfter 的多证（同域同批签发的确会发生）需要确定性 tiebreak���如指纹字典序，同 `preferOwner` 的做法），否则列表翻页/重渲染顺序漂移。

问题：两处验收表述与代码现实有字面出入，会让判据在执行时被绕过。【LOW】
> "卡片选中态保留 aria-pressed" 与 "代码中无 router.push('/certs'...) 残留" —— 分级卡（`OverviewCards.vue` 的 level 按钮）今天并没有 aria-pressed，只有特殊卡有；对分级卡这是**新增**而非"保留"，写成保留会让实现者漏加。同理现行跳转是对象语法 `router.push({ path: '/certs', query: { daysLeft: tier } })`，按 proposal 里的字面串 grep 是 grep 不到的；断言应写成"dashboard 目录内无 `router.push` 且 `useRouter` 导入移除"。

## Improvement Suggestions

建议：把特殊卡闭合判据重写为"域计数 ⇔ 行集合"的显式关系，并把特殊卡筛选的匹配谓词从"行徽标相等"改为"行内任一 SAN 具有该状态"。【CRITICAL】Addresses: 特殊卡闭合判据自相矛盾（第一条风险）。
> What changes: 成功判据改为可断言的三句：(1) diff 卡：筛选结果中每个 diffAlertCount 计数的域名至少对应一行含 diff SAN 的证书行；(2) 行数 ≥ 卡片值（同名多证下允许大于），并单测锁定"1 diff 域名 × 2 证书 → 卡 1 / 行 2"的样例作为**预期行为**而非失败；(3) exempt 卡按"行含 exempt SAN"匹配（而非行徽标==exempt），保证最差优先坍缩不吞掉豁免行的可发现性。实现上 `filterDashboardItems` 需要接收行级 per-SAN 状态集合（载荷里的 `sans` + 各 SAN 状态或预计算的状态集合字段），这也顺带解决徽标 tooltip 的明细来源。

建议：把"分级筛选激活即豁免孤儿隐藏"写成显式规则与三态矩阵，配 11=11 联调断言。【CRITICAL】Addresses: 分级卡动态相等 × 孤儿隐藏 × countsByLevel 含孤儿的未定义交互。
> What changes: 在 In Scope 的前端条目下加一张矩阵：默认视图（隐藏孤儿，卡显"可见 N · 隐藏 M"）/ level=expired 激活（与台账 `daysLeft=expired` 强制豁免同构，孤儿显现且带"已隐藏候选"样式或徽标，此时 M=0）/ 查看全部开关（全量显现）/ 特殊卡与工具栏组合筛选（声明豁免是否随任一筛选激活而生效——建议与台账对齐：仅 daysLeft=expired 豁免，其余筛选不改变隐藏）。成功判据追加："过期卡激活时行数 == countsByLevel[expired]（含孤儿，11=11）；与台账 daysLeft=expired 视图行数一致"。这条断言是本 proposal 存在的理由，必须落在测试里而不是会议纪要里。

建议：在载荷契约里补齐行级字���清单并规定判定单点。【HIGH】Addresses: 隐藏执行位置未定义、前端第三处重写判定、DashboardItem 缺 referenceStatus。
> What changes: In Scope 后端条目明确新增字段：`referenceStatus`（三态，服务端派生）、`sans`（含每 SAN 聚合探测状态，供折叠计数、tooltip 明细、特殊卡谓词）、`cn`/`issuer`；明确隐藏执行位置（推荐：items 含孤儿 + 行级三态 + API `hiddenCount`，前端只做"按标记隐藏/显示"，不得自行推导三态），并在测试断言"前端按标记隐藏后的可见行数 + API hiddenCount == items 总数"。

建议：规定每档隐藏数的来源，解禁或明确拒绝扩展 countsByLevel。【HIGH】Addresses: 卡片级"可见 N · 隐藏 M"缺数据来源。
> What changes: 二选一并写入 In Scope：(a) `countsByLevel` 扩展为 `{total, hiddenPerLevel}`（同步修订 Out of Scope 的"语义不动"为"total 语义不动，新增 hidden 维度"），或 (b) 卡片只显示 total + 全局 hiddenCount 徽标，每档隐藏数由前端按行级 `referenceStatus` 派生，测试断言派生值与 API hiddenCount 闭合。无论哪种，"卡计数⇔表行数动态相等"判据必须注明其生效条件（ expired 档需在豁免规则生效的前提下成立）。

建议：定义多 SAN 行的抽屉进入语义。【HIGH】Addresses: 现成单域名抽屉对多 SAN 行不成立。
> What changes: 推荐最小改动：抽屉内增加 SAN 列表（每 SAN 行显示各自徽标/探测时间/线上指纹，点击切换主区），行点击默认定位到最差优先 SAN；若坚持"现成抽屉"，则退而求其次——行内 SAN 徽标明细 tooltip 中每个 SAN 可键盘激活并携带对应域名打开抽屉，并把该交互写入 In Scope 前端条目，同时在无障碍判据里加"键盘可达每个 SAN 的详情"。

建议：把 blind_spot 恒显示升格为成功判据。【HIGH】Addresses: 盲区可见性无判据锁定。
> What changes: 成功判据追加两条："过期且 referenceStatus=blind_spot 的证书在默认视图恒显示（含无成功快照场景），单测覆盖'无快照 + 过期'样例"；"隐藏判定禁止裸 refCount 判定，引用徽标必须渲染三态之一而非引用计数布尔"。这与台账删除拦截的保守语义同构，是"风险证据不被静默吞掉"的兜底。

建议：显式定义聚合徽标的行级语义文案与并列 tie。【MEDIUM】Addresses: 域级"任一归属证一致"基准与证书行语义错位、健康/未探测并列未定义。
> What changes: 徽标 tooltip 与文案明确为 SAN 语义："一致=该 SAN 线上生效证书为台账归属证之一"，避免过期旧证行被误读为"本证在线上生效"；最差优先序拆开"健康/未探测"：同证书内存在未探测 SAN 且无更高优先级状态时显示"未探测（部分）"或在 tooltip 列明；为 `exempt+wildcard_skipped`、`consistent+未探测` 等并列样例各写一行单测断言。顺序本身（diff 最高）合理，保留。

建议：把 wildcardSkippedCount 纳入域名口径保护名单并加防重计单测。【MEDIUM】Addresses: 证书循环直出重复计数通配符跳过。
> What changes: Out of Scope 条目改为"diffAlertCount/exemptCount/wildcardSkippedCount 保持域名（去重 SAN）级计数不动"；实现上通配符计数留在去重域名集合上的独立循环（或以 SAN 集合去重），单测断言"同一通配符 SAN 挂两张证 → WildcardSkippedCount==1"。

建议：规定三态判定的批量访问模式。【MEDIUM】Addresses: 逐证派生的 N+1 风险。
> What changes: In Scope 后端条目写明：快照与引用上下文一次拉取（复用 `cloudsByFingerprint` 已有的 `LatestDone` + 引用列表），三态派生仅对候选行（过期 && 快照计数 0）调用共享 `deriveRefStatusFor`，其余行按 `has_refs`（计数>0）/`blind_spot`（无快照）直接标注——与台账 `listCertsHideExpiredNoRefs` 的访问模式逐字对齐，并加一条"派生查询次数不随证书数线性增长"的性能注释或测试桩断言。

建议：修正客户端可见性管线的三处细节并纳入测试清单。【MEDIUM】Addresses: row-key 失效、云选项含隐藏孤儿、aria-live 通告失真。
> What changes: row-key 迁 `certId`（In Scope 明写）；`cloudFilterOptions` 改为基于"默认可见行"构建（或选中含隐藏行的云时展示"含 N 张已隐藏"提示）；`filterAnnouncement` 的 total 基准改为"默认可见总数"，无筛选时文案改为"显示 N 条，另有 M 张已隐藏"（M=0 时维持现文案），组合筛选下分母口径写入 format 测试。

建议：修正三处表述精度并加固验收断言字面。【LOW】Addresses: "维持排序"与现状矛盾、"保留 aria-pressed"措辞、grep 断言字面失配。
> What changes: NFR 改为"排序变更为 notAfter asc、同 notAfter 按指纹字典序稳定 tiebreak，单测覆盖并列样例"；可访问性条目改为"分级卡新增 aria-pressed（特殊卡已有，行为对齐）"；`router.push` 残留判据改为"dashboard 目录无 `router.push` 调用且无 `useRouter` 导入残留"，使断言可被 grep 真实执行。

---

总结：方向、克制与复用纪律都是对的，"与台账共用三态判定"更是唯一正确路径。这份 proposal 现在缺的不是方案，而是**闭合规范**：四类计数（countsByLevel、表行数、hiddenCount、diff/exempt 域计数）在五种过滤态下的相等关系没有一张矩阵，特殊卡判据甚至写了不可达的等式。按上面前两条（CRITICAL）补齐判据矩阵与 11=11 断言后，这个改造的验收才是可执行的；其余条目是实现前就该写进载荷契约与测试清单的细节，成本都很低。
