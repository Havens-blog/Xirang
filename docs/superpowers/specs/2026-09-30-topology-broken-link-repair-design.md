# 资源拓扑断链修复（第一轮）设计

- 日期：2026-09-30
- 状态：待评审
- 范围：e-cam-service `internal/topology` + `cmd/apm-push`
- 决策记录：见文末「已定取舍」

## 1. 背景与目标

资源拓扑当前由两个写入方生产，但它们互不相通，且「断链」口径错误，导致 700 节点 / 889 边里存在大量**真实却无法被正确计数与定位**的断链。

Playwright 实测（tenant=深圳嘉立创，243 域名全 tencent）：

| 指标 | 实测值 |
|---|---|
| 总节点 / 总边 / 域名 | 700 / 889 / 243 |
| `GetStats.broken_count` | 70（= pending 边数） |
| APM 子图 | 84 节点（`svc-*`）、261 边、**70 pending**（全部 target 缺失） |
| 域名 infra 链（pro.lceda.cn） | DNS→CDN→WAF→SLB→EIP→15 ENI，0 ECS / 0 服务 / bridge=0 |
| APM 服务 `service_ips` | `172.23.x / 10.201.x`（pod IP） |
| ENI `primary_private_ip` | `10.41.x`（node IP） |

**本轮目标**：①统一「断链」口径为可验证的拓扑定义；②修掉 70 条 APM 断链的根因；③加一个轻量 reconciliation 让 pending 边可自愈；④补类型清单漂移。**不做 infra↔服务桥接**（留第二轮）。

## 2. 根因（数据佐证）

1. **70 条 APM 断链**：`cmd/apm-push/declaration_gen.go:Generate` 只按 `caller` 分组生成节点声明；callee-only 服务（`svc-basic-data`、`svc-file-storage-core`、`svc-jlc-wallet-system` 等）在 `nameMapping` 中已存在但永不生成节点 → 指向它的边在 `Register` 里被标 `pending` 后永久悬空。
2. **infra↔服务不连通**：唯一桥 `bridgeELBToAPM` 用 ENI `private_ip`(10.41.x) 匹配服务 pod IP(172.23.x)，**物理上不可能相等**；且只在内存、不落库。→ 本轮不修，见 §9。
3. **「断链」口径错**：`GetStats` 用 pending 数，`GetBusinessTopology` 用 `DetectBrokenLinks` 的「双向类型只有单侧边」启发式（误报 CDN/WAF 叶子 + 漏报真正的悬空边）。
4. **类型清单漂移**：`live_builder.modelUIDToType` 产出 `nlb/eni/clb/mongodb/vpc/eip/nas/kafka/elasticsearch`，均不在 `ValidNodeTypes`；未来任何经 `Validate` 的路径都会拒绝这些合法节点。

## 3. 正确口径定义

> **一条边断链 ⟺ 其 `source_id` 或 `target_id` 在 `ecam_topo_node`（同租户）中不存在。**

- `pending` 是断链的子集（目标尚未注册），计入 `broken_count`。
- `broken_count` 是**租户级数据完整性指标**，与当前视图筛选无关。
- 移除「双向类型节点只有单侧边」启发式；`BidirectionalNodeTypes` / `TopoNode.IsBidirectional` 若无其他引用则一并删除。

## 4. 本轮范围

| 项 | 内容 | 状态 |
|---|---|---|
| P0 | 统一 broken_count 口径（backend） | 本轮 |
| P1 | apm-push 为 caller ∪ callee 都发节点声明 | 本轮 |
| P3-lite | 查询时批量激活「目标已存在」的 pending 边 | 本轮 |
| P4-lite | 补 `ValidNodeTypes` 缺失类型 + 前端 `NODE_TYPES` | 本轮 |
| ~~P2~~ | infra↔服务域名 join 桥接 | **下轮** |
| ~~死代码采集器下线 / svc-*↔k8s-* 统一 / ENI→ECS~~ | 清理类 | **下轮** |

## 5. 设计

### 5.1 P0：统一 broken_count 口径

**文件**：`internal/topology/service/dag_builder.go`、`internal/topology/service/topology.go`、`internal/topology/service/live_builder.go`

- 新增唯一实现：

```go
// 悬空边数 = 端点(source/target)不在节点集内的边数（含 pending）
func computeBrokenCount(nodes []domain.TopoNode, edges []domain.TopoEdge) int

// 租户级断链数：加载租户全部节点 ID + 全部边，独立于视图的 provider/type/source_collector 筛选
func (s *topologyService) tenantBrokenCount(ctx context.Context, tenantID int64) int
```

- `DagBuilder.DetectBrokenLinks` 改为委托 `computeBrokenCount`（保留方法名以免散落调用点，语义纠正）。`live_builder.BuildFromDNS` 内部已调用它，自动纠正。
- `GetStats`：由「`BrokenCount = CountPending`」改为调用 `tenantBrokenCount`。
- `GetBusinessTopology`：`computeStats` 不再自行算 broken_count；最终 `Stats.BrokenCount` 由 `tenantBrokenCount` 赋值（live 与纯 DB 两条路径统一，均不受视图筛选影响）。

结果：`/topology`（画布 stats 条）与 `/topology/stats` 两处返回同一数字，语义一致。

### 5.2 P1：apm-push 为 callee 也建节点

**文件**：`cmd/apm-push/declaration_gen.go`（`Generate`，`GenerateWithIPs` 复用）

- 现状：`Generate` 只遍历 `callerGroups`（按 caller 分组）。
- 改法：以 `nameMapping`（已含 caller ∪ callee ∪ trace 服务名）为**权威节点清单**，逐名生成声明：
  - 该名有出边（是 caller）→ 生成带 `Links` 的声明（逻辑不变）；
  - 该名无出边（callee-only / span-only）→ 生成 `Links` 为空的叶子声明，节点 type 仍为 `k8s_deployment`、category `container`。
- 保留对 `nameMapping[callee]` 缺失的防御性跳过（真正未映射的外部名不建节点）。
- `GenerateWithIPs` 的 `service_ips` 附着循环已按 `Node.ID` 匹配 `nameMapping`，叶子节点自动覆盖，无需改。

**预期效果**：apm-push 重新 push 后，`svc-basic-data` 等 70 个 callee-only 服务落为节点；`Register` 对每个新节点调用 `UpdatePendingEdges`，70 条 pending 边自动激活 → `broken_count` 由 70 → 0（存量为 0 会在下一轮 push 后收敛）。

### 5.3 P3-lite：查询时自愈 pending

**文件**：`internal/topology/repository/edge.go`、`internal/topology/repository/dao/edge.go`、`internal/topology/service/topology.go`

- `EdgeRepository` 新增 `ActivateResolvablePendingEdges(ctx, tenantID int64) (int64, error)`：
  批量把「target_id 在节点集中存在、status=pending」的边置为 active（DAO 用 `$lookup` 或两步：先取节点 ID 集，再 `UpdateMany({"tenant_id":…,"status":"pending","target_id":{"$in":ids}})`）。
- `GetBusinessTopology` 开头调用一次（忽略错误，仅日志）。作用：目标节点若由非 Register 路径（如 LiveBuilder）创建，pending 也能被激活。成本低、幂等。

### 5.4 P4-lite：类型清单补全

**文件**：`internal/topology/domain/node.go`、`e-cam-web/src/api/types/topology.ts`

- `node.go` 补常量并加入 `ValidNodeTypes`：`nlb/clb/eni/mongodb/vpc/eip/nas/kafka/elasticsearch`。
- 前端 `NODE_TYPES` 同步补这些常量（供渲染类型映射兜底）。
- `BidirectionalNodeTypes` 若在 5.1 后无引用则删除；`NodeTypeENI` 等常量命名沿用现有 snake_case 风格。

## 6. 测试计划

- `cmd/apm-push/declaration_gen_property_test.go`：新增「callee-only 服务生成叶子节点声明」「叶子声明 `Links` 为空」「service_ips 附着到叶子节点」。
- `internal/topology/service/dag_builder_test.go`：改写 `DetectBrokenLinks` 用例为「悬空边计数」——source 缺失、target 缺失、pending 各计 1，双向节点单侧边**不再**计断链。
- `internal/topology/service/topology_property_test.go`（若有 GetStats 用例）：断言 `GetStats.BrokenCount == computeBrokenCount(全量)`。
- `internal/topology/repository`：DAO `ActivateResolvablePendingEdges` 单测（mock/内存）。
- 全量回归：`just e-cam-service-unit-test`（现有单测范围）+ `cd e-cam-service && GOFLAGS=-p=2 go test -count=1 ./internal/topology/... ./cmd/apm-push/...`。

## 7. 风险与回滚

- **P1 改动 apm-push 输出**：新增的是节点声明（叶子），不删不改为默认已有边；`Register` 幂等。若 ARMS cron 短期未重跑，70 断链不会立刻归零，但不引入新问题。
- **broken_count 语义变化**：由「pending 数 / 启发式」变为「悬空边数」。前端仅读数字展示，无结构依赖；数字可能暂时非 0（见 §8 存量），需在发布说明标注。
- **`ActivateResolvablePendingEdges`**：只把「目标已存在」的 pending 置 active，方向单一、幂等，无副作用。
- 回滚：每项改动独立，可按 P0/P1/P3/P4 逐 commit 回退。

## 8. 存量数据迁移

- 70 条 pending 边：无需手工迁移，P1 后由 apm-push 重推 + `Register` 激活自动收敛。
- 历史上可能存在的「source 缺失」悬空边：`broken_count` 会如实报出；其剪除不属本轮（P2 轮的 reconciliation 统一处理），本轮只让数字变准。

## 9. 第二轮预告（不含在本轮）

- **P2 infra↔服务桥接**：join key 改用「域名」——APM 边已带 `attributes.domains`（来自 trace root `http.host`），把「域名→infra 末端 LB→该域名入口服务」建成 `route` 边并落库。**难点**：入口服务判定（某域名 APM 边中只作 source、其 domains 含该域名者）。
- ENI→ECS 绑定修复（`instance_id` 是 UUID，需按 `asset_id`/`servername` 回查）。
- 死代码采集器（`DNSCollector`/`K8sCollector`/`CollectorManager` 未接线）下线或接线；`svc-*` 与 `k8s-*` 语义统一（涉及已有 ID 迁移）。
- 前端 `showBrokenOnly` 与明细面板的悬空边可视化。

## 已定取舍

1. P1 修法：**apm-push 为 callee 也建节点**（源头正确），非后端占位兜底。
2. P2 桥接：**留到下一轮**，本轮收敛为「正口径 + 修断链 + 轻量 reconciliation + 类型补全」。