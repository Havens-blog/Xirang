---
created: "2026-09-21"
author: "Haven"
status: Approved
intent: "enhancement"
---

# Proposal: 到期看板分级卡页内下钻（证书粒度主表，弃跳台账）

## Problem

到期看板分级卡的点击交互是"跳转台账按档过滤"，而非页面内呈现该档证书——根因是看板存在粒度矛盾：**分级卡计数是证书粒度，主表却是域名粒度**（同域多张证书仅展示 `notAfter` 最新者独占一行），导致过期/临期旧证被同名新证掩盖，只能跳台账看证书明细。

### Evidence

- `e-cam-web/src/views/cert/dashboard/index.vue:158-165`（7cdd488 注释自陈）：*"看板表格为域名粒度，过期/临期证书被同名新证掩盖无法页内表达"*——跳台账是当时对粒度矛盾的妥协，非设计目标。
- `OverviewCards.vue:13` 卡片静态文案"点击跳转台账按档过滤"；用户点击卡片的第一反应是**就地看到该档具体证书**（依据见 Assumptions Challenged 表），跳转造成页面断裂感。
- 用户此前实测反馈："到期看板有的点击按分级过滤为空"——即域名粒度下行内筛选与卡片计数对不上。
- 相邻特性 `ledger-hide-expired-orphans` 刚交付"台账默认隐藏过期且无引用证书"，看板不跟随将与其**口径分裂**。

### Urgency

- 看板是**只读角色的默认页与唯一菜单入口**（`index.vue:75`），其粒度矛盾直接影响最大众的用户群。
- 域名粒度造成**风险盲区**：域名行显示健康（最新证有效），但同域已存在过期旧证（可能仍被引用或裸奔），运营误判"无风险"。
- 若现在不做，看板与刚验收的台账"已过期"计数立即对不上——不一致窗口越久，用户信任损耗越大。

## Proposed Solution

将到期看板主表从**域名粒度重铸为证书粒度**，分级卡点击改为**页内筛选**（复用现有 filter 状态源），移除跳台账；探测以**行级聚合徽标**附着在证书行（行点击进入探测抽屉扩展的"证书多 SAN 视图"）；"已过期且 referenceStatus=no_refs_scanned"孤儿证书**与台账同构由服务端默认隐藏**（复用三态判定单点），API 返回双口径计数（全局 hiddenCount + countsByLevel 每档 total/visible/hidden）。


### Innovation Highlights

- 主表对齐真实业务实体（证书）粒度，分级计数与明细天然闭合——cert-manager 逐证呈现到期，AWS ACM 控制台同理，域名仅作字段。
- 探测状态**聚合为行级徽标**（借鉴 CI 与监控面板的"最差优先"reducer），解决探测按域名、证书按张的 1:N 附着。
- 与台账共用同一三态判定，隐藏口径单源；豁免规则与开关强制开启是同一机制（单一"分级查看=全量"不变量）。差异点：探测态与到期态并置于一张证书行——业界证书列表无此合并。


## Requirements Analysis

### Key Scenarios

- **快乐路径**：见"已过期 N"卡，点击 → 主表就地筛出 N 张过期证书行（含同名多证并立），行上徽标即判优先级，行点击看探测详情，全程未离开看板。N=N 依赖**总规则（豁免仅由风险维度搜索触发）**：到期分级卡或 diff/exempt 探测卡激活即豁免孤儿隐藏；云/托管/issuer/关键字搜索等属性维度筛选**不**豁免（孤儿保持隐藏，仅在当前可见行集上再过滤）。豁免生效时前端以 `includeHidden=true` 重新拉取全量，隐藏判定始终在服务端执行、前端不派生三态。countsByLevel 含孤儿，否则行数必小于卡计数；与台账 `daysLeft=expired` 豁免同构。

- **边界：同名多证**：`*.jlc-robot.com` 过期旧证与有效新证同时在主表成两行（不再互相掩盖），daysLeft 独立分档。
- **边界：探测态冲突**：行徽标按最差优先序 `diff > change_linked_diff > unreachable > wildcard_skipped > exempt > 健康`（未探测 SAN 排序尾）聚合；并列同态不细分。徽标仅作快速分级，tooltip 注明冲突 SAN 明细；筛选匹配以行内 SAN 明细为准。

- **边界：特殊卡筛选（diff/exempt）**：diffAlertCount/exemptCount 保持域名级计数；行集合 = 「存在任一 SAN 探测态为该状态的证书行」（逐 SAN 谓词），exempt 等低优先态被更差聚合掩盖时仍可筛中；按总规则豁免孤儿隐藏。特殊卡副标题标注「域名计数」，卡值小于筛出行数（一个 diff 域名挂 2 张证 → 卡 1 行 2）为预期而非缺陷。徽标"一致"判定基准是域级 `ownership[name][onlineFP]`（`probe_service.go`）；过期旧证行上到期状态视觉优先于徽标，副标签读作「生效归属另一证书」——运营不会误读为"本证线上生效"。

- **边界：孤儿**：仅"过期且 `referenceStatus=no_refs_scanned`"由**服务端**隐藏（复用台账三态判定单点；`has_refs` 与 `blind_spot` 恒可见；默认隐藏，仅风险维度卡激活豁免——见总规则）；隐藏谓词复配台账的**扫描新鲜度阈值**：仅依据新鲜度阈值内的成功扫描快照判定，超阈值（快照陈旧可能漏掉新引用）降级为保守显示、不隐藏，并展示最近扫描时间；卡片与表以双口径呈现（每档隐藏数来自 countsByLevel per-level hidden 分量）；前端只消费 visible 行 + hiddenCount，不派生三态。

- **错误路径**：探测数据缺失 → 行徽标显示"未探测"；载荷变更无第三方调用者（见兼容性）。

### Non-Functional Requirements

- **性能**：主表行数由"去重域名数"增至"证书数"；三态判定为**服务端批量内存判定**（一次拉取 refs + 最新扫描快照后内存归约，仅对"过期且 refCount=0"候选派生；逐证 N+1 为反例）；上线后观测阈值：主表 >5000 行（当前 jlc 94 证远低于阈值）即行数日志告警、届时加装分页/虚拟化，首期不实现分页。SAN 折叠阈值 20：证书 SAN 超 20 时抽屉默认折叠（最差优先前 20，其余可展开）。

- **排序**：由域名字典序**改为** notAfter asc（变更而非维持）；并列按证书指纹字典序稳定取序。

- **可访问性**：沿用现有 `aria-live` 筛选通告与骨架屏，通告 total 基准明确为**可见行数**（不含被隐藏孤儿）。

- **兼容性**：`GET /certs/dashboard` items 为破坏性变更；单消费端（e-cam-web dashboard 页）同步改造，无外部消费者；筛选状态同步 URL query 保留深链（弥合移除 `router.push`）。

- **只读边界与数据可见性**：看板保持无任何变更/扫描/发起更换入口（Hard Rule 不破）；只读角色可见字段即 In Scope 所列行内新字段；证书指纹为公开哈希且台账已展示，看板行沿用同一可见性；`referenceStatus` 已存在于台账详情同一权限位，看板行下发不扩大最小权限（对齐台账详情可见性，不额外暴露）。

### Constraints & Dependencies

- 依赖 ledger-hide 的引用三态判定（`referenceStatus∈{has_refs,no_refs_scanned,blind_spot}`）与 `daysLeft` expired 豁免语义——看板隐藏谓词复用**台账同一服务端三态判定源**（提取共享 helper 时以编译/单测锁定无 import cycle）；**必须复用同一实现，不重写**。

- 复用：`probes.LatestPerDomain`（探测数据源）、探测详情抽屉组件（扩展为证书多 SAN 视图）、`WildcardProbeOverrides`（通配符跳过判定）、`countsByLevel` 证书粒度计数（升级为每档 {visible, hidden}）；前端 filter 单一状态源机制沿用。


## Alternatives & Industry Benchmarking

### Industry Solutions

证书到期仪表盘主流形态：证书行主表 + 分级计数卡 + 到期筛选；探测/合规状态以行级标签呈现。具名参照：cert-manager（Kubernetes）`kubectl get certificate` 逐证列出到期日，AWS ACM 控制台逐证展示到期与续期状态——均以证书为行主体。业界孤儿处理多为到期/状态分级排序或归档，罕见"无引用过期证默认隐藏"；本方案取与台账口径对齐的单源三态判定，代价是引入隐藏状态机。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing（维持跳台账） | 现状 7cdd488 | 零成本 | 过期旧证在域名表持续被同名新证掩盖；台账"已过期"计数与看板卡计数继续分裂 | Rejected: 不作为的代价即上述盲区与口径分裂 |
| 域名主表 + 就地证书抽屉（方案 B） | 画布备选 | 纯前端小改 | 不消除"同名新证掩盖"根因；卡计数与抽屉行数两来源 | Rejected: 修标不修本 |
| **证书粒度主表 + 页内筛选 + 探测徽标 + 孤儿同构隐藏（方案 A）** | 本 proposal（依据见 Assumptions Challenged 表） | 计数⇔表行数闭合、消除掩盖、与台账口径单源 | 后端+前端中等改造（改动面见 Feasibility） | **Selected**: 唯一消除同名掩盖根因的方案（方案 B 保留该根因） |
| 证书粒度主表 + 全量显示（不做孤儿隐藏） | 方案 A 变体 | 同样消除掩盖，无隐藏状态机 | 与台账口径分裂（台账隐藏孤儿、看板全显，"已过期"计数再度对不上） | Rejected: 口径分裂代价高于状态机成本 |

## Feasibility Assessment

### Technical Feasibility

- 后端：`dashboard_service.go`（`CountsByLevel` 循环改证书直出、探测经 `probeByDomain[san]` 聚合、孤儿经三态判定过滤）+ `dashboard_handler.go`（载荷）+ 提取共享三态 helper——改动面即此三处。
- 前端：`DashboardTable.vue` 行渲染改造 + `OverviewCards.vue` 卡点击改页内 filter（`router.push` 移除）+ `format.ts` + 两侧测试更新；可无障碍复用现有抽屉/徽标/通告组件。
- 无新依赖、无 showstopper。

### Resource & Timeline

团队已完成两轮看板工程；估算约 1 个 feature 周期（2~3 任务：后端载荷+前端交互+联调验证；8 组测试门禁随对应任务内建，不另立项）。

### Dependency Readiness

ledger-hide 已合入 main（1f4ecd5、6393454、fbf31aa），三态判定可直接复用；探测管线已有产出（见项目记录）。


## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 点击分级卡"就应该跳台账" | 5 Whys | **Overturned**: 真实意图是"就地看到该档证书"；跳转只是粒度矛盾下的妥协产物 |
| 看板表格必须是域名粒度（才能表达探测） | Assumption Flip | **Refined**: 探测可聚合为证书行徽标，行点击下钻域名明细——探测洞察不丢 |
| 卡片计数与表行数"不可能一致，所以必须跳走" | Occam's Razor | **Overturned**: 证书粒度化后两者天然一致；根因就是粒度错配 |
| 台账隐藏孤儿、看板可全量显示 | Stress Test（一致性） | **Refined**: 会造成两页面计数口径分裂；判定同构复用 |
| diff/exempt 卡语义必须筛域名行 | Assumption Flip | **Refined**: 计数保持域名级；筛选迁移为"行内任一 SAN 具该状态的证书行"（逐 SAN 谓词），随主表粒度迁移且映射关系显式闭合  |

## Scope

### In Scope

- 后端 `GET /certs/dashboard`：items 改为证书行（每证一行，行级下发 `referenceStatus`；新增 `includeHidden` 查询参数：默认 false=服务端隐藏生效，true=返回全部证书行含孤儿）；探测徽标行内聚合（最差优先序+tie）；孤儿隐藏由**服务端**执行（复用台账三态判定单点，谓词=expired AND `no_refs_scanned` 且扫描快照在新鲜度阈值内；`has_refs`/`blind_spot` 恒可见；默认隐藏、仅风险维度卡激活豁免，见总规则）；countsByLevel 升级为每档 {total, visible, hidden}（total=visible+hidden）+ 全局 hiddenCount。

- 前端 `DashboardTable`：证书行渲染（CN、SANS 折叠计数、issuer、daysLeft、托管、引用云、探测徽标）；行 key 改证书 ID；孤儿默认隐藏 + 切换入口（工具栏横幅+开关，同台账模式；开关状态按用户持久化于 localStorage，对账工作流刷新不丢）+ 空态处理；行点击 → 探测抽屉扩展的证书多 SAN 视图。

- 行排序改为 notAfter asc（由域名字典序变更而来，见 NFR/SC）；`wildcardSkippedCount` 证书循环直出时按 domain 去重（共享通配符 SAN 的多证不重复计数）。

- `OverviewCards` 分级卡：点击改为页内筛选（选中高亮、再点取消、`aria-pressed` 两态），移除跳台账逻辑与静态文案；激活即按总规则豁免孤儿隐藏 → 行数 == 该档 countsByLevel 总计数；未激活显示"可见 N · 隐藏 M"。

- 特殊卡 diff/exempt：计数保持域名级（副标题标注「域名计数」）；行集合 = 「存在任一 SAN 为该状态的证书行」（逐 SAN 谓词），按总规则豁免孤儿隐藏。

- 配套测试更新：`dashboard_service_test.go` 与前端 format/组件测试随新语义重建。

### Out of Scope

- 台账 ledger 端隐藏逻辑本身（ledger-hide 已交付，只读复用不改动）。
- 探测能力 / 巡检流程 / 通配符语义（仅呈现方式调整）。
- 3 张覆盖率卡（`registrationRate`/`replaceableRate`/`fingerprintOnlyRate`）与 2 张特殊卡的**域名级计数语义**（`diffAlertCount`/`exemptCount`/`wildcardSkippedCount`，去重实现见 In Scope）；覆盖率卡为纯展示、非筛选触发器（点击行为不变）。

- 看板只读边界（不增加任何变更/扫描/更换入口）。
- 变更单、发起更换等处置流程（仍由台账承载）。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 同名多证使主表行数增多、拖长页面 | M | L | 行数增幅 = 证书数 − 去重域名数（当前 jlc 实测 94 证 / 93 域，+1 行）；行内 SAN 折叠为数量徽标；notAfter asc 排序使高风险行恒在首屏  |
| 多 SAN 探测态聚合歧义（"这一行健康吗"存疑，"一致"基准为域级归属时过期旧证行会显示"一致"） | H | M | 显式最差优先序+tie 写入实现+单测；徽标以稳定键区分归属语义，tooltip 注明冲突 SAN 明细；抽屉保留该证全部 SAN 探测真相  |
| 既有 dashboard 测试断言（域名行语义）大面积失效 | M | L | 按新语义重建 `dashboard_service_test.go` 与前端 format 测试 |
| 看板与台账隐藏判定实现漂移（两处同时改判断逻辑） | M | H | 复用**同一个** ledger 三态判定 helper，不许看板另写判定；服务端单测锁定两处一致性 |
| `GET /certs/dashboard` 破坏性变更漏网调用者 | M | M | grep 全仓 `getCertDashboardApi` 核验单消费端（dashboard 页）；加载荷契约回归测试持续断言 items 形状；**发布方式**：破坏性变更以前后端同一 release/PR 协同上线，无多阶段部署窗口；回滚=前后端各自 revert（无数据迁移）；前端静态资源缓存/CDN 与后端版本可能异步，回退窗口内看板短暂显示旧域名行视图，可接受；如需无缝则以载荷版本字段过桥  |
| 空态/全隐藏边界（全部证书被隐藏时看板表空） | L | M | 与台账同构：空态提示 + "查看全部"入口保留，不得静默吞行 |
| 行 key 与筛选选项失效：现用 `row-key="domain"`，证书粒度化后域名不唯一，Vue 行复用错乱；云筛选选项若含被隐藏孤儿的云则选中即空态 | H | M | 行 key 改证书 ID；云选项来源改为「当前可见行」派生；两者各加组件测试（同域两证不串行、隐藏态下云选项不含孤儿云）  |
| `wildcardSkippedCount` 重复计数（共享通配符 SAN 的多证直出时各计一次） | M | M | 按 domain 去重累加（保持域名级口径）；单测断言两证计数 1  |
| 非 1:1 租户（cert:domain 比例爆炸）行数失控 | L | M | 上线前核查各租户比例；行数日志告警，超 5000 行阈值（见 NFR）届时加装分页/虚拟化  |
| 服务端隐藏集合上线后不可观测（静默少行难发现） | M | M | 服务端日志/指标记录 hidden set 大小与 hiddenCount，上线后核对  |

## Success Criteria

- [ ] `GET /certs/dashboard` 返回证书行：`includeHidden=true` 时 items 长度 == 全部证书数（含同域多证，不再域名去重）；单测断言同域新旧两证同时成行、独立分档，且每行携带 `referenceStatus`
- [ ] 分级卡点击在**页内筛选**：激活即豁免孤儿隐藏（总规则）→ 主表行数 == 该档 countsByLevel 总计数（纯关系断言 `rows.length === countsByLevel[tier].total`）；再点取消恢复默认隐藏视图；grep `src/views/cert/dashboard/` 无 `/certs` 跳转残留；分级卡选中/未选两态 `aria-pressed` 各有 vitest 断言

- [ ] 孤儿隐藏由服务端执行且不静默：默认视图（无筛选）不含"过期且 `no_refs_scanned`"行——总规则唯一生效点；countsByLevel 每档 {visible, hidden} 与全局 hiddenCount 同源（与 items 同一快照计算；单测断言 includeHidden=true 下各档 total 之和 == items.length）；"查看全部"开关状态机：分级卡激活时强制开启并禁用，清除筛选恢复原开关态（状态持久化，刷新保留），开启后行数 == visible + hidden；服务端单测断言看板与台账三态判定单点对同一证书集合隐藏一致

- [ ] 风险信号恒可见：`has_refs` 与 `blind_spot`（引用状态未知）的过期证书默认视图均可见；单测以裸 `refCount=0` 反例锁定——判定必须走 `referenceStatus`，`blind_spot` 不得隐藏

- [ ] 探测聚合与文案：构造多 SAN 混合态样例 → 行徽标按 `diff > change_linked_diff > unreachable > wildcard_skipped > exempt > 健康`，并列同态不细分、未探测 SAN 排序尾；过期旧证行"一致"徽标以稳定 test-id 断言副标签「生效归属另一证书」且到期状态视觉优先；行点击进入证书多 SAN 探测视图（SAN >20 折叠，默认最差优先前 20，其余可展开）

- [ ] 特殊卡映射闭合（非数值相等）：`diffAlertCount`/`exemptCount` 保持域名级计数（副标题标注「域名计数」）；点击后主表行集合 == 「存在任一 SAN 为该状态的证书行」集合，且按总规则豁免孤儿隐藏——单测断言集合映射（含"一个 diff 域名挂 2 张证 → 卡 1 行 2"正例）及 exempt 行被更差态掩盖时仍可筛中

- [ ] 回归门禁：wildcardSkippedCount 域名去重单测（共享 `*.example.com` 两证计数 1）；"查看全部"开关状态机行为测试；云筛选选项由当前可见行派生（隐藏态不含孤儿云）、行 key=证书 ID，各由组件测试锁定；Out-of-Scope 不变量（3 张覆盖率卡 + 三类域名级计数）重构后口径不变的断言

- [ ] 排序可复现：items 按 notAfter asc，并列按证书指纹字典序稳定取序；单测断言同 notAfter 两证顺序固定

- [ ] 全部门禁绿：e-cam-service `go test ./...`（含新的 dashboard_service 单测）、e-cam-web `vitest` + `vue-tsc` 通过

## Next Steps

- 先获明确批准，再经 `/quick` 管线生成任务并执行（intent=enhancement，自动串行）；或先进 `/write-prd` 正式化需求。
