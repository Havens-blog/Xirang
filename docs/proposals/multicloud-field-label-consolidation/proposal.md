---
created: "2026-09-30"
author: "Haven"
status: Approved
intent: "refactor"
---

# Proposal: 多云产品字段展示单源收口(状态/版本/产品枚举)

## Problem

15 个云产品页面把「状态」等字段的展示文案散落为 index.vue 列表页 + 详情抽屉各写一份本地 map,同一字段键 ↔ 文案之间没有单一来源约束,已出现 index↔drawer 漂移与跨产品同词异义。

### Evidence

代码审计(2026-09-30)实测:

- 状态 `status`:7 个列表页各自本地 `XXX_STATUS_LABELS`(image/eip/eni/lb/vpc/vswitch/waf)+ 6 个详情抽屉手写 `statusLabels`(image/eip/eni/lb/vpc/vswitch)+ ecs/disk 在 index 内多处内联 map。数据库/中间件/存储域 6 产品(mongodb/rds/redis/es/kafka/nas)在 index.vue 与抽屉**各写一份**内联 `getStatusText`(例:mongodb 只认 `running/stopped`,kafka 多认 `serving`、es 多认 `active`)。
- 同键漂移:`active` 在 waf 显示「正常」、ecs/assets 显示「运行中」;`stopped` 有的「已关机」有的「已停止」;WAF 抽屉已迁 `ASSET_STATUS_LABELS`(active→运行中)而其列表页仍用本地 `WAF_STATUS_LABELS`(active→正常)——同一产品的 index 与 drawer 已不一致。
- 版本 `edition`:WAF 的 `editionMap`(basic/pro/business/enterprise)在 `waf/index.vue`(还含 sparta-waf/EdgeOne)与 `WafDetailDrawer.vue` **两份、键集不等**。
- 散点产品专属枚举:磁盘类型(高效云盘/SSD/ESSD)在 `disk/index.vue` 与 `DiskDetailDrawer.vue` 双份;EIP 实例类型、ENI/LB/image 类型等各自内联。
- 对照(证明单一来源模式可行):`charge_type` 已由 `fieldLabels.ts` 的 `CHARGE_TYPE_LABELS` 收敛,eip/ecs/cmdb/rds/redis/mongodb/nas/template/provision 全部 `labelOfLenient(CHARGE_TYPE_LABELS, …)`,零残留。

### Urgency

每新增一个状态/枚举取值都需同时改 index+drawer 多处,漏一处即同页显示不一致;一笔多云资源在列表与详情里状态文案对不上,直接影响运营核对。不改则后续每个产品迭代持续放大漂移。当前无阻断,但逐月递增维护成本。

## Proposed Solution

建立「产品内单一来源」:每个产品的「状态 / 版本 / 产品专属枚举」词表各只保留**一份**共享定义,列表页、详情抽屉一律 import 同一份;产品语义差异**保留**(EIP 已绑定/未绑定、WAF 防护态、DNS 域名态不与通用运行态强行同一)。裁定规则:列表页词汇为 canonical,抽屉对齐;已知 index↔drawer 冲突按此修正。复用已建立的 `fieldLabels.ts` 单源形态(`labelOf`/`labelOfLenient`),与 charge_type 先例同型。

### Innovation Highlights

非创新,是把 `fieldLabels.ts`「单一来源」既定模式从计费类型一个域**推广铺满**到状态/版本/枚举各域;本质是诚实复用代码库里已被验证成功的先例,价值在覆盖而非新意。

## Requirements Analysis

### Key Scenarios

- 列表页与详情抽屉对同一资源的状态展示文案一致(消除 index↔drawer 漂移)。
- 新增某个状态/枚举取值时只需改一处共享定义。
- 各产品保留自身语义措辞(EIP `InUse`→已绑定、DNS `paused`→暂停、WAF `active`→正常 等不被通用词覆盖)。
- 已知冲突(index 与 drawer 值不一时)按「列表页为 canonical」裁决,且不引入后端契约变更。

### Non-Functional Requirements

- 兼容性:不改后端数据契约、不改字段键,纯前端展示层;调用方签名不变(仍 `(value) => string`)。
- 性能:无(查表 O(1),与现状同构)。

### Constraints & Dependencies

- 纯前端 Vue3 + TS,无新依赖、无后端改动。
- 依赖已存在的 `src/utils/fieldLabels.ts`(`CHARGE_TYPE_LABELS`/`ASSET_STATUS_LABELS`/`labelOf`/`labelOfLenient`)与 `src/utils/cdn.ts`(cdn 已收口,作参照)。
- 并行会话正在改的 WIP 文件不得触碰(现 `src/views/network/vpc/components/VpcDetailDrawer.vue` 已被并行会话修改,须提交前重新拉取最新或暂缓该文件)。

## Alternatives & Industry Benchmarking

### Industry Solutions

前端字段文案单一来源是通用模式:i18n 词表、常量字典、枚举→文案映射表单一导出,消费方 import 而非复制。Vue3 生态常规做法与此一致(单一 `labels.ts`/`enums.ts`)。本项目已有 `utils/fieldLabels.ts` 前半程落地。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零成本 | 漂移继续放大,index↔drawer 已不一致 | Rejected: 已出现实际漂移 |
| 全局强统一词表 | 单一 ASSET_STATUS_LABELS 覆盖所有产品 | 视觉最统一 | 覆写产品语义(EIP InUse→使用中、WAF→运行中 误导) | Rejected: Assumption Flip 证明语义冲突 |
| 各文件保留 + 加固 lint 约束 | — | 改动小 | 治标不治本,lint 难约束「文案」级漂移 | Rejected: 解决不了键↔文案漂移 |
| **产品内单一来源** | fieldLabels.ts 既有模式推广 | 消除重复与漂移、语义保留、可 grep 校验 | 治理文件多(~30 文件) | **Selected: 语义正确 + 可客观验证** |

## Feasibility Assessment

### Technical Feasibility

纯前端机械重构,技术栈完全支持,无阻拦性依赖;charge_type 先例已验证「单一来源 + labelOf 形态」在本仓可落地。

### Resource & Timeline

机械但量大(15 产品、约 30 文件),按域拆分后每块独立可验;无新技能门槛。

### Dependency Readiness

无外部依赖。唯一依赖的风险是并行会话 WIP 文件,以「提交前复核、遇冲突暂缓」处置。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 「统一 = 全产品同一张词表」 | Assumption Flip | Overturned:active/InUse 跨产品语义冲突,强统一会标错 |
| 「问题只剩状态一个字段」 | 代码审计 | Refined:版本 edition、散点产品枚举也在漂移 |
| 「要把产品语义抹平才算统一」 | 5 Whys | Overturned:统一的目标是可维护 + 无漂移,不是强求同词 |
| 「charge_type 仍未收敛」 | 代码审计 | Overturned:已全量 `labelOfLenient(CHARGE_TYPE_LABELS)`,排除出范围 |

## Scope

### In Scope

- 状态单源收口 —— 网络域 7 产品(eip/eni/lb/vpc/vswitch/waf/dns)。index+drawer 共用一个状态词表。
- 状态单源收口 —— 计算域 3 产品(ecs/image/disk)。消除 index 多内联 map 与 drawer 手写 map。
- 状态单源收口 —— 数据/中间件/存储域 6 产品(rds/redis/mongodb/es/kafka/nas)。合并 index+drawer 双份内联 `getStatusText`。
- 版本 `edition` 单源收口 —— WAF:index 与 drawer 的 `editionMap` 合一(含 sparta-waf/EdgeOne 键),消键集不等。
- 散点产品专属枚举单源收口 —— 凡 index↔drawer 或 index 内重复出现的枚举(磁盘类型、EIP 实例类型、ENI/LB/image 类型等)收口;单文件单用且无漂移的保留不动。
- 漂移裁决细则落地:列表页词汇 canonical、抽屉对齐;已知 index↔drawer 文案冲突修正。
- 全程保证:前端 typecheck 与既有单测通过;收敛后 grep 校验手写本地 map 归零。

### Out of Scope

- 计费类型 `charge_type`(已收敛;注:WAF 抽屉另有 `subscription/payasyougo` 变体键与共享 `CHARGE_TYPE_LABELS` 语义相同但键不同,留作后续单独小改)。
- `iam/users` 账号状态(非云资源语义)。
- `cdn`(已由 `cdn.ts` 收口)。
- 颜色映射(如 DNS `providerColors`/`recordTypeColors`)。
- 后端字段数据回填(层面 B:字段占位/无真实录入)。
- 跨产品强制统一措辞(保留产品语义)。
- 并行会话 WIP 文件(当前 `VpcDetailDrawer.vue` 等),复用其最新成果。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 与并行会话 WIP 冲突(如 VpcDetailDrawer.vue) | M | M | 任务开始前 `git status` 复核;遇冲突改/合并后再动;不覆盖他人改动 |
| 词汇裁决造成可见文案变化 | M | L | 列表页为 canonical;抽样人工核对受影响产品 |
| 量大易漏(15 产品) | M | M | 按域拆任务、每任务 ≤8 文件;完成标准用 grep 客观校验 |
| 破坏类型签名/导出链路 | L | M | 保持 `(value) => string` 形态,复用 labelOf/labelOfLenient,不改键 |

## Success Criteria

- [ ] 收敛后 `grep -rnE 'statusLabels|getStatusText|STATUS_LABELS' src/views/{compute,network,databases,middleware,storage}` 中手写本地 map(定义处)归零,仅余共享引用 import。
- [ ] WAF `editionMap` 定义收敛到单一文件,`waf/index.vue` 与 `WafDetailDrawer.vue` 仅 import。
- [ ] 每个产品内同一状态键在 index 与 drawer 的展示文案一致(无 index↔drawer 漂移)。
- [ ] 产品语义保留:EIP `InUse`→已绑定、DNS `paused`→暂停、WAF `active`→正常 等专属措辞不被覆盖。
- [ ] 前端 typecheck 通过、既有单测通过(纯前端重构,无后端编译要求)。

## Next Steps

- quick 模式:直接进入任务生成(Skill `forge:quick-tasks`),由 proposal 的 In Scope 派生 per-域任务。