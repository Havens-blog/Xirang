---
created: "2026-09-22"
author: "Haven"
status: Approved
intent: "fix"
---

# Proposal: 服务树绑定环境按资产推断落库，纠正 691 条环境错标

## Problem

规则引擎把资产绑定到规则写死的单一 `env_id`，而资产实际横跨多个环境，导致跨环境资产（如 `rod-cpp-prod-*`）被统一绑到规则的 dev 环境。

### Evidence

嘉立创租户（tenant 3）活体实测：

- 根子树 1063 条规则绑定中，**691 条**被环境疑异检测判为错标（prod 命名绑 dev 环境）。
- CPP 规则 `env_id=5(dev)` 绑了 839 条 `rod-cpp` 资产，其中 577 条命名含 `-prod-`。
- 根因双重确认：① 规则落库 `env_id` 恒用 `rule.EnvID`；② 已有推断能力但从未接入绑定——`GetNodeAssetSummary` 里的 `matchNameEnvCode`/`normalizeEnvCode` 能从命名 `-prod-`/`-uat-`/`-test-`/`-dev-` 判定真实环境，却只用于"标红提示"不改绑。
- 额外发现：tag 环境信号是死的——`extractTagEnv` 读 `tags["env"]`，但实测标签键是 `tags["environment"]`（192 条，值 prod 162/dev 14/uat 8/fat 8），`env` 键 0 条。

### Urgency

691 条错标会让「按环境分布」的业务线资产概览失真，运维据此判断生产/预发资源分布会做出错误决策。推断能力已存在、绑定只差"接入"一步，是典型的低成本高价值修复；继续拖延只会随资产同步让错标数量继续膨胀。

## Proposed Solution

绑定时，目标环境改为「资产推断优先、规则 `env_id` 兜底」：

1. **公共环境推断**：把 `node_asset.go` 里的 `matchNameEnvCode`/`normalizeEnvCode`/`isStandardEnvCode` 抽成公共方法，新增 `inferEnvCode(instance)`——命名模式优先，其次 `tag.environment`（并修复 `env` 字段名读取错误），返回标准环境码 `dev`/`test`/`staging`/`prod` 或空串。
2. **环境码对齐**：推断出的 canonical 码（如 `-uat-` → `staging`）与租户环境表 code（`dev`/`test`/`uat`/`prod`）对齐——双方都经 `normalizeEnvCode`（`uat`≡`staging`）后再匹配；无匹配的环境码（如 `fat`）视为无信号，走规则兜底。
3. **绑定落库接入**：`ExecuteRules` 创建绑定时，`env_id = resolve(推断码, rule.EnvID)`；推断不出沿用 `rule.EnvID`。`ExecuteRules` 的 create-only 幂等与 manual 锁定语义不变。
4. **历史纠正**：`PreviewRebind`/`ApplyRebind` 的重算口径追加环境——rule 绑定资产的目标环境也用 `resolve(推断码, …)` 而非规则写死 env；候选判定从「节点不同」扩为「节点或环境不同」。复用现有「规则改绑」预览→确认入口，一次性收敛历史 691 条。

### Innovation Highlights

无新造轮子——复用 task 6 既有的环境命名模式与归一化逻辑，把"只读标红"升级为"绑定落库"同一套判断，保证概览统计与绑定记录口径一致（不再出现"概览说什么、绑定就什么"的分裂）。方案简洁处在于：改绑原语 `UpdateTarget(id, nodeID, envID, ruleID)` 本就带 env 参数，前端预览/确认弹窗都已存在，改动集中在后端计算 env 这一处。

## Requirements Analysis

### Key Scenarios

- 命名含 `-prod-` 的资产绑到含 prod 语义的规则 → 落 prod 环境（不再落 dev）。
- `tag.environment=uat` 的资产 → 落租户 `uat` 环境（对齐 canonical `staging`）。
- 既无命名模式也无 environment 标签的资产 → 规则 `env_id` 兜底，行为同现状。
- 手动绑定的资产 → 永不参与推断与改绑（manual 永锁）。
- 历史 691 条 → 改绑预览列出「节点或环境变化」候选，确认后一次性纠正。

### Non-Functional Requirements

- 幂等：`ExecuteRules` create-only 语义不变，重复执行零重复绑定。
- 性能：环境推断是纯内存字符串匹配，无额外 IO；env→id 映射沿用 `loadEnvCodeMap` 一次加载。

### Constraints & Dependencies

- 复用 `internal/cam/servicetree/service/node_asset.go` 既有常量/方法；不改 binding 表结构。
- 依赖任务 6 的 `GetNodeAssetSummary` 推断逻辑与三期 `UpdateTarget` 原语（均已落地）。

## Alternatives & Industry Benchmarking

### Industry Solutions

云资源按环境分组是 CMDB 常见需求；行业通行做法是"标签优先 + 命名约定兜底"（AWS Resource Tag / 阿里云 tag 分组），与本方案一致。差别在于多数系统要求人工打标签，本方案利用嘉立创已形成的 `-prod-/-uat-/-test-/-dev-` 命名约定 + `environment` 标签双信号，零人工成本。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零成本零风险 | 691 条持续标红，概览失真，随同步膨胀 | Rejected: 用户已拍板纠正 |
| 规则 env 变过滤条件、按环境拆多条规则 | 显式建模 | 语义最显式 | 规则数暴涨，与现有粗规则冲突 | Rejected: 违反一条规则覆盖一业务线的现状 |
| **资产推断优先 + 规则兜底** | 命名+标签双信号 | 零人工、改动集中、概览与绑定同口径 | 推断有边界（无信号走兜底） | **Selected: 复用既有推断，最小改动纠正错标** |

## Feasibility Assessment

### Technical Feasibility

推断与 mapping 逻辑全部已存在于代码库，只需抽取复用 + 改绑定落库一处。无新依赖、无 schema 变更、无新中间件。

### Resource & Timeline

单仓（e-cam-service）后端改动 + 测试，规模约 2–4 个文件，无前端改动（规则改绑预览弹窗已就绪）。

### Dependency Readiness

`UpdateTarget`（三期 282e256）、`detectEnvMismatch` 推断套件（task 6 594b892）均已落 main 并活体验证。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "环境只能靠规则写死" | Assumption Flip | Overturned——资产自身命名/标签已含环境信号，推断能力已存在于 task 6 |
| "tag 环境信号用 `tags.env`" | Occam's Razor | Overturned——实测数据用 `tags.environment`，`env` 键 0 条，现有读取是死代码 |
| "改绑只需纠正节点" | 5 Whys | Refined——719 错标的根是环境，纠正节点不纠正环境是治标 |

## Scope

### In Scope

- 抽取公共环境推断（命名模式 + `tag.environment`，修正标签键），供绑定/改绑/概览共用。
- `ExecuteRules` 绑定环境按推断优先、规则兜底。
- `PreviewRebind`/`ApplyRebind` 重算口径追加环境，候选含「环境变化」。
- 环境码对齐（`uat`≡`staging`）；`fat` 等未知码走兜底。
- 单测：推断命中/未命中、code 对齐、绑定兜底、改绑含环境纠正。

### Out of Scope

- 规则 env 变过滤条件（多规则拆分）。
- 新增前端交互（复用现有规则改绑预览/确认弹窗）。
- 手动绑定资产的环境推断（manual 永锁不改）。
- CMDB 侧环境字段的异构采集治理。

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 命名模式误判环境（如名称含 `-prod-` 但非生产语义） | M | M | 仅用双信号（命名+标签），标签优先于命名；无信号走兜底不强行改 |
| 环境码对齐错（`uat`/`staging`/`fat`） | M | M | 双方经 normalizeEnvCode 归一后再匹配；未知码 fallback 规则 env |
| 一次性纠正 691 条误伤 | M | H | 走既有预览→确认两步，非自动批量；改绑可再被规则重算对冲 |
| 推断逻辑与概览口径漂移 | L | M | 抽成公共方法供绑定/改绑/概览三方复用，杜绝第二份实现 |

## Success Criteria

- [ ] `ExecuteRules` 绑定的资产其 `env_id` 与资产命名/tag 推断环境码一致（推断命中时）；推断未命中时等于规则 `env_id`。
- [ ] 改绑预览对「环境但非节点」变化的 rule 绑定也产生候选；`ApplyRebind` 后该资产环境纠正。
- [ ] 嘉立创租户改绑执行后，根子树环境疑异数从 691 显著下降（目标 ≤ 无信号资产数），且无新增节点错绑。
- [ ] 环境推断逻辑单份复用（binding/rebind/summary 三方同源），无第二份复制实现。
- [ ] 单测覆盖推断命中/未命中、`uat`/`staging`/`fat` 对齐、manual 锁定不变；`go build` + `go test ./internal/cam/servicetree/...` 全绿。

## Next Steps

- Proceed to `/write-prd` to formalize requirements