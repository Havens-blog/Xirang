---
created: "2026-09-21"
author: "Haven"
status: Approved
intent: "new-feature"  # valid values: new-feature | enhancement | refactor | cleanup | fix | doc
---

# Proposal: 服务树功能重构一期+二期——规则引擎运转与业务线资产概览

## Problem

服务树（组织维度 × 资产归属的映射系统）两条核心能力全部断线，实测确认：

1. **归属自动化失效**：规则引擎「建而未运转」——现有 2 条规则目标节点显示「-」、未启用、从未执行（匹配数「-」），资产归属退化为纯手动，10 台 ECS 全部堆在根节点，业务线节点（SMT订单等）**0 资产**，树是空架子。
2. **可见性缺失**：无业务线维度资产概览——任一节点看不到其子树聚合的资产全景（数量/环境分布/云平台分布），根节点「关联资产」tab 为空（聚合语义未实现）。
3. **治理无护栏**：生产资产 `ali-jlc-prod-ops-datasync` 被手动绑在「开发环境」，无人发现。

### Evidence（chrome-devtools 实测 + 代码验证，2026-09-20）

- 绑定规则页 2 条规则均 `node_name: "-"`：后端 `ListRules`（servicetree/web/handler.go:628）不回填 node_name，前端 `row.node_name || '-'`（rules/index.vue:63）恒空。
- 规则条件 UI 暗坑：`RuleFormDialog.vue:82` 提供「地域 region」选项，但后端 `rule_engine.go:234 getFieldValue` 无 region case（仅 name/asset_id/model_uid/attributes.*/tag.*）→ **region 条件永不匹配**。
- 标签条件硬编码 3 个（tag.env/service/team，RuleFormDialog.vue:85-87），不能自由输入；无资源组条件——而资源组 ID 已采集（aliyun `ResourceGroupId` → 资产 `ProjectID` 属性，asset.go:198）。
- 执行无调度：`ExecuteRules` 仅有手动 POST `/rules/execute`（handler.go:70），无同步后自动触发、无定时。
- 规则启用态：现有 2 条 Enabled=false（列表状态列空）。
- 后端底子完备：`InstanceFilter` 原生 `TagFilter`；`BindType manual/rule` 区分；规则域校验 NodeID 必填。

### Urgency

- **HIGH**：服务树是「按业务线看资产」的核心动线；当前归属靠手点、概览无聚合，随资产规模增长治理成本线性恶化。用户已确认归属策略第一性取舍：**规则自动为主 + 手动兜底**。

## Proposed Solution

### 一期：规则引擎真正运转（归属自动化）

1. **修显示与暗坑**：`ListRules` 回填 `node_name`；`getFieldValue` 补 `region` case（对齐 UI）；规则列表状态/匹配数真实呈现。
2. **条件编辑器升级**：自由 tag key 输入（保留常用建议项）+ `attributes.xxx` + 资源组条件（`attributes.project_id`，资源组 ID 已采集）+ region 对齐后端。
3. **试运行 dry-run**：新 API「按规则预览匹配集」（不落库，返回命中资产清单），规则保存前可验证条件写对没有。
4. **执行与反馈**：资产同步完成后自动执行规则（事件驱动挂入同步链路）+ 保留手动按钮；执行结果落库（匹配数/执行时间），规则列表真实呈现；`BindType=rule` 的绑定记录来源规则 ID（既有字段）。
5. **手动兜底语义**：手动绑定优先级高于规则（规则执行跳过已手动绑定的资产，既有行为确认保留）；解绑手动绑定不被规则重新绑架（unpin 语义明确化）。

### 二期：业务线资产概览（可见性）

6. **子树聚合接口**：`GET /nodes/:id/asset-summary`——返回该节点子树聚合的资产总数、按环境分布、按云平台分布、按类型分布、（一期后）规则/手动来源分布。
7. **树节点徽标**：服务树总览树每个节点显示子树资产数徽标，业务线空/非空一目了然。
8. **关联资产 tab 聚合语义**：根节点/中间节点展示子树聚合资产（现状空），叶子节点展示直属绑定。
9. **环境错绑护栏**：概览/绑定列表标红「环境疑异」资产（资产命名或 `tag.env` 与绑定环境矛盾，如 prod 资产绑 dev 环境），不做自动改绑仅提醒。

### 三期（另排，不在本提案）：资源组名称解析（ProjectName 补齐需额外云 API + 缓存）、资源组维度分组视图。

## Requirements Analysis

### Key Scenarios

- 管理员创建规则「tag.service contains smt → SMT订单节点/生产环境」，dry-run 预览命中 12 台 → 保存启用 → 资产同步完成后自动绑定，SMT订单节点出现资产。
- 总览树：SMT订单节点徽标显示 12；点开节点概览显示 12 台中 8 生产/4 预发、10 阿里/2 腾讯。
- 生产资产误绑开发环境 → 概览标红提醒，管理员解绑。
- 既有手动绑定不被规则覆盖（manual 优先）。

### Non-Functional Requirements

- 规则执行幂等：重复执行不产生重复绑定（按 instance+node+env 去重，既有唯一键确认）。
- 同步后自动执行不阻塞同步主链路（异步/失败容忍）。
- dry-run 只读不落库。

### Constraints & Dependencies

- 后端 `internal/cam/servicetree`（handler/service/domain）+ 资产同步链路挂点（`internal/cam/task/executor` 同步完成后回调或事件）。
- 前端 `views/service-tree/`（rules/index + RuleFormDialog + index.vue 总览）。
- 不动 `accounts/index.vue`（并发脏）、logs；资源组名称解析（三期）不做。

## Alternatives & Industry Benchmarking

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| 一期规则运转 + 二期概览 | 本提案 | 服务树达到可用闭环；符合「规则为主」第一性 | 体量较大（前后端 + 调度挂点） | Adopt（用户已确认分期） |
| 三期全量（含资源组名称解析） | 备选 | 一步到位 | 资源组名称需额外云 API+缓存，拖长周期 | Deferred |
| 只做一期 | 备选 | 最小 | 「业务线资产概览」核心诉求延后 | Rejected |
| 规则建议+人工确认制 | 备选 | 最稳健 | 每步人工，规模化后成本高 | Rejected：用户已选自动为主 |

## Scope

### In Scope

- 后端：node_name 回填、getFieldValue 补 region、dry-run API、同步后自动执行挂点、执行结果落库、子树聚合 API、环境错绑检测字段。
- 前端：RuleFormDialog 条件编辑器升级（自由 tag/attributes/资源组/region 对齐）、规则列表真实状态、dry-run 预览交互、树节点资产数徽标、关联资产聚合语义、环境疑异标红。
- 测试：规则引擎 dry-run/执行幂等/聚合统计单测。

### Out of Scope

- 三期资源组名称解析（ProjectName）与资源组分组视图。
- 自动改绑/自动修复环境错绑（仅提醒）。
- accounts/logs（并发在途）、成本维度概览（FINOPS 另有入口）。
- 规则的多条件 OR 关系（现有 AND 语义保持）。

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 同步后自动执行拖慢同步主链路 | Medium | 异步 goroutine + 超时 + 失败容忍（对齐 CDN 指标采集的教训：失败不阻塞） |
| 规则误匹配大规模错绑（条件写宽了） | Medium | dry-run 预览必经交互 + 执行结果反馈匹配数，异常匹配数一眼可见 |
| 聚合接口大子树性能（根节点全租户） | Medium | 聚合走绑定表聚合查询（group by），非逐节点递归；必要时缓存 |
| region 等字段对齐后旧规则语义变化 | Low | 现有 2 条规则均未启用未执行，无存量语义包袱 |
| 并发会话同仓工作冲突 | Medium | 后端文件不与 cert/logs/ui-unify 在途重叠；提交显式 pathspec；e-cam-web HEAD 在 feature/ui-unify 期间提交需确认分支 |

## Success Criteria

- [ ] 规则列表 node_name 真实回填；状态/匹配数/执行时间真实呈现
- [ ] getFieldValue 支持 region；条件编辑器支持自由 tag key / attributes.xxx / 资源组（project_id）条件
- [ ] dry-run API 返回命中资产清单且不落库；前端保存前可预览
- [ ] 资产同步完成后自动执行规则（异步不阻塞），BindType=rule 绑定落库带 rule_id；幂等（重复执行零重复绑定，单测覆盖）
- [ ] 子树聚合 API 返回总数/环境/云平台/类型/来源分布；树节点显示资产数徽标；关联资产 tab 根/中间节点展示子树聚合
- [ ] 环境错绑资产在概览标红（命名或 tag.env 与绑定环境矛盾检测）
- [ ] `go build`+`go test ./internal/cam/servicetree/...` 全绿；`pnpm typecheck`+vitest 全绿（并发在途错误除外）
