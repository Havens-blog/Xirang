# 资源拓扑断链修复（第一轮）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 统一「断链」口径为可验证的拓扑定义，修掉 70 条 APM 永久 pending 断链的根因，加自愈 reconciliation，补类型清单漂移。

**Architecture:** 分四个独立任务。P0 把 broken_count 收敛为单一悬空边定义并统一 `/topology` 与 `/topology/stats` 两处返回；P1 改 apm-push 生成器为 callee-only 服务补叶子节点声明；P3 加 DAO 级批量激活 + 服务层接线；P4 补全节点类型清单。每项独立可提交、可回滚。

**Tech Stack:** Go 1.x + MongoDB（`go.mongodb.org/mongo-driver`）+ `pkg/mongox`；测试 `testify/assert` + `pgregory.net/rapid`（property）。前端仅改一个 TS 常量文件（Vue3）。

**Spec:** `docs/superpowers/specs/2026-09-30-topology-broken-link-repair-design.md`

## Global Constraints

- 只改 spec §4「本轮范围」内的文件；**不碰** `internal/servicetree`、`internal/logquery`、`internal/shared/cloudx`（它们当前有未提交的在途改动，本计划的 commit 只 `git add` 本计划列出的文件）。
- 断链定义唯一口径：`broken_count = 悬空边数（source_id 或 target_id 不在同租户节点集内，含 pending）`；移除「双向类型节点只有单侧边」启发式。
- `broken_count` 是租户级数据完整性指标，与视图 provider/type/source_collector 筛选无关。
- APM 节点 ID 保持 `svc-*` 现状（本轮**不**做 `svc-*`→`k8s-*` 迁移）。
- 不落库也不改 `bridgeELBToAPM`（P2 桥接留第二轮）。
- Go 单测 `gofmt` 通过；commit message 用 conventional commits，末尾附 `Co-Authored-By: Claude Code <noreply@anthropic.com>`。

## Review Focus

（spec 隐含但单个任务测试未覆盖、最可能咬人的输入/失效模式，逐一在所属任务补测试）

1. **source 缺失的边也要计断链**（不只是 target 缺失）→ Task 1 补 source-missing 用例。
2. **两个 ARMS 服务名映射到同一节点 ID**（`prod_x`/`staging_x` 都→`svc-x`）时，叶子声明不得重复生成同一节点 → Task 2 补去重用例。
3. **broken_count 不受视图筛选影响**（provider/type/source_collector 过滤下仍返回全租户悬空边）→ Task 1 的 `tenantBrokenCount` 用无筛选 `Find`，补测试。
4. **空图不 panic、broken=0** → Task 1 补空切片用例。
5. **callee-only 且其 own 出边全未映射**的服务，仍应作为叶子节点落地（不再因为 `len(links)==0` 被丢弃）→ Task 2 补测。

---

## Task 0: 建分支

- [ ] **Step 1: 从 main 建功能分支（不提交在途改动）**

```bash
cd D:/Haven/e-cam-service
git checkout -b fix/topology-broken-links
```

Expected: 分支切换成功；`git status --short` 仍显示 servicetree/logquery 的在途未提交改动（它们不参与后续 commit）。

---

## Task 1: P0 统一 broken_count 口径

**Files:**
- Modify: `internal/topology/service/dag_builder.go`
- Modify: `internal/topology/service/topology.go`
- Modify: `internal/topology/domain/node.go`（删除 `BidirectionalNodeTypes`）
- Modify: `internal/topology/domain/node_test.go`（删除 `TestTopoNode_IsBidirectional`）
- Modify: `internal/topology/service/dag_builder_test.go`（重写 `TestDagBuilder_DetectBrokenLinks`）
- Create: `internal/topology/service/broken_count_test.go`

**Interfaces:**
- Produces: `func computeBrokenCount(nodes []domain.TopoNode, edges []domain.TopoEdge) int`（同包可见）
- Produces: `func (s *topologyService) tenantBrokenCount(ctx context.Context, tenantID int64) int`
- Produces: `func (s *topologyService) computeStats(nodes []domain.TopoNode, edges []domain.TopoEdge, brokenCount int) domain.TopoStats`（签名变化，Task 3 不依赖；同包两处调用点随本任务同步改）

- [ ] **Step 1: 写失败测试（悬空边计数 + 空图 + source 缺失）**

Create `internal/topology/service/broken_count_test.go`:

```go
package service

import (
	"context"
	"testing"

	"github.com/Havens-blog/e-cam-service/internal/topology/domain"
	"github.com/stretchr/testify/assert"
)

func TestComputeBrokenCount(t *testing.T) {
	t.Run("no broken links", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "a"}, {ID: "b"}, {ID: "c"}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "a", TargetID: "b"}, {ID: "e2", SourceID: "b", TargetID: "c"}}
		assert.Equal(t, 0, computeBrokenCount(nodes, edges))
	})

	t.Run("target missing counts as broken", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "a"}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "a", TargetID: "missing", Status: domain.EdgeStatusActive}}
		assert.Equal(t, 1, computeBrokenCount(nodes, edges))
	})

	t.Run("source missing counts as broken", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "b"}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "missing", TargetID: "b", Status: domain.EdgeStatusActive}}
		assert.Equal(t, 1, computeBrokenCount(nodes, edges))
	})

	t.Run("pending edge counts as broken", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "a"}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "a", TargetID: "x", Status: domain.EdgeStatusPending}}
		assert.Equal(t, 1, computeBrokenCount(nodes, edges))
	})

	t.Run("both endpoints missing counts once", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "a"}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "x", TargetID: "y", Status: domain.EdgeStatusActive}}
		assert.Equal(t, 1, computeBrokenCount(nodes, edges))
	})

	t.Run("empty graph is zero", func(t *testing.T) {
		assert.Equal(t, 0, computeBrokenCount(nil, nil))
	})
}
```

并在 `internal/topology/service/broken_count_test.go` 追加 `tenantBrokenCount` 过滤无关性测试（依赖 Task 1 Step 4 的 mock 增强后可用；先写，Step 2 会因未定义而失败）：

```go
func TestTenantBrokenCount_IgnoresViewFilters(t *testing.T) {
	nodeRepo := newMockNodeRepo()
	nodeRepo.nodes["a"] = domain.TopoNode{ID: "a"}
	nodeRepo.nodes["b"] = domain.TopoNode{ID: "b"}
	edgeRepo := newMockEdgeRepo()
	edgeRepo.edges = []domain.TopoEdge{
		{ID: "e1", SourceID: "a", TargetID: "b", SourceCollector: domain.SourceAPM},
		{ID: "e2", SourceID: "a", TargetID: "ghost", SourceCollector: domain.SourceDNSAPI},
	}
	svc := &topologyService{nodeRepo: nodeRepo, edgeRepo: edgeRepo}
	assert.Equal(t, 1, svc.tenantBrokenCount(context.Background(), 1))
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./internal/topology/service/ -run 'TestComputeBrokenCount|TestTenantBrokenCount_IgnoresViewFilters'`
Expected: FAIL `undefined: computeBrokenCount` / `undefined: tenantBrokenCount`。

- [ ] **Step 3: 实现 computeBrokenCount + 重写 DetectBrokenLinks**

Modify `internal/topology/service/dag_builder.go`：替换 `DetectBrokenLinks` 全文（第 104-136 行）为：

```go
// DetectBrokenLinks 统计断链边数：source 或 target 端点不在节点集内的边即断链。
// pending 边计入（其 target 尚未注册）。不再使用「双向类型只有单侧边」启发式。
func (b *DagBuilder) DetectBrokenLinks(nodes []domain.TopoNode, edges []domain.TopoEdge) int {
	return computeBrokenCount(nodes, edges)
}

// computeBrokenCount 悬空边计数：source_id 或 target_id 不在 nodes 内的边数（含 pending）。
func computeBrokenCount(nodes []domain.TopoNode, edges []domain.TopoEdge) int {
	nodeIDs := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		nodeIDs[n.ID] = struct{}{}
	}
	broken := 0
	for _, e := range edges {
		if _, ok := nodeIDs[e.SourceID]; !ok {
			broken++
			continue
		}
		if _, ok := nodeIDs[e.TargetID]; !ok {
			broken++
		}
	}
	return broken
}
```

- [ ] **Step 4: 实现 tenantBrokenCount + mock Find 增强 + GetStats/computeStats 改造**

Modify `internal/topology/service/topology.go`：

在 `topologyService` 方法区新增（如 `GetStats` 后）：

```go
// tenantBrokenCount 计算租户级断链边数（与视图 provider/type/source_collector 筛选无关）。
func (s *topologyService) tenantBrokenCount(ctx context.Context, tenantID int64) int {
	nodes, err := s.nodeRepo.Find(ctx, domain.NodeFilter{TenantID: tenantID})
	if err != nil {
		return 0
	}
	edges, err := s.edgeRepo.Find(ctx, domain.EdgeFilter{TenantID: tenantID})
	if err != nil {
		return 0
	}
	return computeBrokenCount(nodes, edges)
}
```

`GetStats`：将「`pendingCount, err := s.edgeRepo.CountPending(...)` + `BrokenCount: int(pendingCount)`」替换为 `BrokenCount: s.tenantBrokenCount(ctx, tenantID)`，并删除 `pendingCount` 变量。

`computeStats` 签名改为接收 brokenCount 并去掉内部 DetectBrokenLinks：

```go
func (s *topologyService) computeStats(nodes []domain.TopoNode, edges []domain.TopoEdge, brokenCount int) domain.TopoStats {
	stats := domain.TopoStats{
		NodeCount:   len(nodes),
		EdgeCount:   len(edges),
		BrokenCount: brokenCount,
	}
	for _, n := range nodes {
		if n.Type == domain.NodeTypeDNSRecord {
			stats.DomainCount++
		}
		if n.DagDepth > stats.MaxDepth {
			stats.MaxDepth = n.DagDepth
		}
	}
	return stats
}
```

`GetBusinessTopology`：在 dbNodes 查询后（原第 79 行下方）追加一行：

```go
brokenCount := s.tenantBrokenCount(ctx, params.TenantID)
```

将两处 `s.computeStats(dbNodes, edges)`（LiveBuilder 路径）与 `s.computeStats(nodes, edges)`（纯 DB 路径）改为三参调用：`s.computeStats(..., brokenCount)`。空结果早期 return（原第 166-172 行）的 `Stats: domain.TopoStats{}` 改为 `Stats: domain.TopoStats{BrokenCount: brokenCount}`。

Modify `internal/topology/service/declaration_property_test.go`：把 `mockNodeRepo.Find` 由返回 `nil, nil` 改为返回已存节点：

```go
func (m *mockNodeRepo) Find(_ context.Context, _ domain.NodeFilter) ([]domain.TopoNode, error) {
	result := make([]domain.TopoNode, 0, len(m.nodes))
	for _, n := range m.nodes {
		result = append(result, n)
	}
	return result, nil
}
```

（`mockEdgeRepo.Find` 已返回 `m.edges`，无需改。）

- [ ] **Step 5: 删除死代码 + 重写 DetectBrokenLinks 测试**

Modify `internal/topology/domain/node.go`：删除 `BidirectionalNodeTypes`（第 109-114 行）与 `IsBidirectional`（第 158-161 行）。

Modify `internal/topology/domain/node_test.go`：删除 `TestTopoNode_IsBidirectional`（第 81-88 行）。

Modify `internal/topology/service/dag_builder_test.go`：将 `TestDagBuilder_DetectBrokenLinks`（第 159-214 行）整体替换为：

```go
func TestDagBuilder_DetectBrokenLinks_DanglingEdges(t *testing.T) {
	builder := NewDagBuilder()

	t.Run("no dangling edges", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "dns-1", Type: domain.NodeTypeDNSRecord}, {ID: "slb-1", Type: domain.NodeTypeSLB}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "dns-1", TargetID: "slb-1", Status: domain.EdgeStatusActive}}
		assert.Equal(t, 0, builder.DetectBrokenLinks(nodes, edges))
	})

	t.Run("unidirectional leaf is NOT broken", func(t *testing.T) {
		// SLB 只有入边、RDS 只有入边：不再是断链（原启发式误报，现已移除）
		nodes := []domain.TopoNode{{ID: "dns-1", Type: domain.NodeTypeDNSRecord}, {ID: "slb-1", Type: domain.NodeTypeSLB}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "dns-1", TargetID: "slb-1", Status: domain.EdgeStatusActive}}
		assert.Equal(t, 0, builder.DetectBrokenLinks(nodes, edges))
	})

	t.Run("pending edge with missing target is broken", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "a", Type: domain.NodeTypeK8sDeployment}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "a", TargetID: "b", Status: domain.EdgeStatusPending}}
		assert.Equal(t, 1, builder.DetectBrokenLinks(nodes, edges))
	})

	t.Run("active edge with missing target is broken", func(t *testing.T) {
		nodes := []domain.TopoNode{{ID: "a", Type: domain.NodeTypeK8sDeployment}}
		edges := []domain.TopoEdge{{ID: "e1", SourceID: "a", TargetID: "b", Status: domain.EdgeStatusActive}}
		assert.Equal(t, 1, builder.DetectBrokenLinks(nodes, edges))
	})
}
```

- [ ] **Step 6: 跑测试确认通过**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./internal/topology/...`
Expected: PASS（原 property 测试亦通过）。

- [ ] **Step 7: gofmt + 提交（只 add 本计划文件）**

```bash
cd D:/Haven/e-cam-service
gofmt -l internal/topology/
git add internal/topology/service/dag_builder.go internal/topology/service/topology.go internal/topology/domain/node.go internal/topology/domain/node_test.go internal/topology/service/dag_builder_test.go internal/topology/service/broken_count_test.go internal/topology/service/declaration_property_test.go
git commit -m "refactor(topology): unified broken_count to dangling-edge definition

Replace the 'bidirectional node with one-sided edges' heuristic with a
single dangling-edge count (endpoint missing). /topology and /topology/stats
now both return the tenant-level broken_count. Removes dead BidirectionalNodeTypes.

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

## Task 2: P1 apm-push 为 callee 也建节点

**Files:**
- Modify: `cmd/apm-push/declaration_gen.go`
- Modify: `cmd/apm-push/declaration_gen_property_test.go`

**Interfaces:**
- Consumes: `nameMapping map[string]string`（arms 服务名 → 节点 ID，含 caller ∪ callee ∪ trace 服务名）
- Produces: `Generate`/`GenerateWithIPs` 现在保证 `nameMapping` 中每个节点 ID 至少有（且仅有）一条节点声明。

- [ ] **Step 1: 写失败测试**

在 `cmd/apm-push/declaration_gen_property_test.go` 追加：

```go
func TestGenerate_CalleeOnlyServicesGetLeafNodes(t *testing.T) {
	g := NewDefaultDeclarationGenerator("tenant-x")
	nameMapping := map[string]string{
		"prod_a": "svc-a",
		"prod_b": "svc-b", // callee-only
		"prod_c": "svc-c", // callee-only, 无任何出边
	}
	deps := []ServiceDependency{
		{CallerServiceName: "prod_a", CalleeServiceName: "prod_b", QPS: 1, LatencyP99: 2, ErrorRate: 0},
		{CallerServiceName: "prod_a", CalleeServiceName: "prod_c", QPS: 1, LatencyP99: 2, ErrorRate: 0},
	}

	decls := g.Generate(deps, nameMapping, nil, nil)

	nodeLinks := map[string]int{}
	seen := map[string]int{}
	for _, d := range decls {
		nodeLinks[d.Node.ID] = len(d.Links)
		seen[d.Node.ID]++
	}
	if !hasKey(nodeLinks, "svc-b") || !hasKey(nodeLinks, "svc-c") {
		t.Fatalf("callee-only services must each get a node declaration; got %v", nodeLinks)
	}
	if nodeLinks["svc-b"] != 0 || nodeLinks["svc-c"] != 0 {
		t.Fatalf("leaf declarations must have zero links; got %v", nodeLinks)
	}
	if nodeLinks["svc-a"] != 2 {
		t.Fatalf("caller svc-a must keep its %d outgoing links; got %d", 2, nodeLinks["svc-a"])
	}
}

func TestGenerate_SameNodeIDNotDuplicated(t *testing.T) {
	g := NewDefaultDeclarationGenerator("tenant-x")
	// prod_x 与 staging_x 映射到同一节点 ID（环境前缀被剥掉）
	nameMapping := map[string]string{
		"prod_a":    "svc-a",
		"staging_a": "svc-a",
		"prod_b":    "svc-b",
	}
	deps := []ServiceDependency{
		{CallerServiceName: "prod_a", CalleeServiceName: "staging_a", QPS: 1, LatencyP99: 2, ErrorRate: 0},
	}

	decls := g.Generate(deps, nameMapping, nil, nil)

	count := 0
	for _, d := range decls {
		if d.Node.ID == "svc-a" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("node svc-a must appear in exactly one declaration, got %d", count)
	}
}

func hasKey(m map[string]int, k string) bool { _, ok := m[k]; return ok }
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./cmd/apm-push/ -run 'TestGenerate_CalleeOnlyServicesGetLeafNodes|TestGenerate_SameNodeIDNotDuplicated'`
Expected: FAIL（callee-only 服务缺节点声明 / svc-b 无声明）。

- [ ] **Step 3: 改造 Generate**

Modify `cmd/apm-push/declaration_gen.go` 的 `Generate`：在现有 caller 分组循环内，生成声明后标记 `emitted[callerNodeID] = true`；循环结束后追加叶子声明块。完整替换函数体（保留 `Name`/Type 常量与现有 link 组装逻辑）：

```go
func (g *DefaultDeclarationGenerator) Generate(
	deps []ServiceDependency,
	nameMapping map[string]string,
	edgeDomains map[string][]string,
	domainMetrics map[string]map[string]DomainMetric,
) []LinkDeclaration {
	callerGroups := make(map[string][]ServiceDependency)
	for _, dep := range deps {
		callerGroups[dep.CallerServiceName] = append(callerGroups[dep.CallerServiceName], dep)
	}

	now := time.Now().Format(time.RFC3339)
	var declarations []LinkDeclaration
	emitted := make(map[string]bool, len(nameMapping)) // nodeID → 已生成节点声明

	for callerName, callees := range callerGroups {
		callerNodeID, ok := nameMapping[callerName]
		if !ok {
			log.Printf("WARN: unmapped ARMS service (caller), skipping: %s", callerName)
			continue
		}

		links := make([]DeclarationLink, 0, len(callees))
		for _, dep := range callees {
			calleeNodeID, ok := nameMapping[dep.CalleeServiceName]
			if !ok {
				log.Printf("WARN: unmapped ARMS service (callee), skipping: %s", dep.CalleeServiceName)
				continue
			}

			edgeKey := EdgeKey(dep.CallerServiceName, dep.CalleeServiceName)
			attrs := map[string]interface{}{
				"qps":          dep.QPS,
				"latency_p99":  dep.LatencyP99,
				"error_rate":   dep.ErrorRate,
				"last_seen_at": now,
			}
			if domains, ok := edgeDomains[edgeKey]; ok && len(domains) > 0 {
				attrs["domains"] = domains
			}
			if dm, ok := domainMetrics[edgeKey]; ok && len(dm) > 0 {
				dmMap := make(map[string]interface{})
				for domain, metric := range dm {
					dmMap[domain] = map[string]interface{}{
						"qps": metric.QPS, "latency_p99": metric.LatencyP99, "error_rate": metric.ErrorRate,
					}
				}
				attrs["domain_metrics"] = dmMap
			}

			links = append(links, DeclarationLink{
				Target:     calleeNodeID,
				TargetType: "k8s_deployment",
				Relation:   "calls",
				Direction:  "outbound",
				Attributes: attrs,
			})
		}

		emitted[callerNodeID] = true
		declarations = append(declarations, LinkDeclaration{
			Source:    "arms-apm",
			Collector: "api",
			Node: DeclarationNode{
				ID:       callerNodeID,
				Name:     extractDeploymentName(callerNodeID),
				Type:     "k8s_deployment",
				Category: "container",
			},
			Links:    links,
			TenantID: g.tenantID,
		})
	}

	// 为 callee-only（从未作为 caller 出现，或无任何出边）的映射服务补叶子节点声明，
	// 使其落地为节点，避免指向它的边永久 pending 悬空。
	for _, nodeID := range nameMapping {
		if emitted[nodeID] {
			continue
		}
		emitted[nodeID] = true
		declarations = append(declarations, LinkDeclaration{
			Source:    "arms-apm",
			Collector: "api",
			Node: DeclarationNode{
				ID:       nodeID,
				Name:     extractDeploymentName(nodeID),
				Type:     "k8s_deployment",
				Category: "container",
			},
			TenantID: g.tenantID,
		})
	}

	return declarations
}
```

注意：原实现里 `if len(links) == 0 { continue }` 会导致 caller 无出边时不发声明，新实现里 caller 无论有无出边都 `emitted` 并发声明（叶子语义一致，且 caller 也被保留为节点）。

- [ ] **Step 4: 跑测试确认通过**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./cmd/apm-push/`
Expected: PASS（含既有 `TestProperty3_*`；它们只断言 links 数量与 per-declaration 字段，叶子声明不破坏断言）。

- [ ] **Step 5: gofmt + 提交**

```bash
cd D:/Haven/e-cam-service
gofmt -l cmd/apm-push/
git add cmd/apm-push/declaration_gen.go cmd/apm-push/declaration_gen_property_test.go
git commit -m "fix(apm-push): emit leaf node declarations for callee-only services

Generate now emits a node declaration for every mapped service name
(caller ∪ callee), not just callers. Callee-only services like svc-basic-data
become nodes, so edges targeting them resolve instead of staying pending
forever (fixes 70 broken APM edges at the source).

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

## Task 3: P3 查询时自愈 pending 边

**Files:**
- Modify: `internal/topology/repository/edge.go`
- Modify: `internal/topology/repository/dao/edge.go`
- Modify: `internal/topology/service/topology.go`
- Modify: `internal/topology/service/declaration_property_test.go`（mock 增补）
- Create: `internal/topology/service/topology_reconcile_test.go`

**Interfaces:**
- Produces: `EdgeRepository.ActivateResolvablePendingEdges(ctx context.Context, tenantID int64) (int64, error)`
- Produces: `func (s *topologyService) activateResolvablePending(ctx context.Context, tenantID int64)`

- [ ] **Step 1: 写失败测试**

Create `internal/topology/service/topology_reconcile_test.go`:

```go
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActivateResolvablePending_TargetsExisting(t *testing.T) {
	mock := newMockEdgeRepo()
	svc := &topologyService{edgeRepo: mock}

	svc.activateResolvablePending(context.Background(), 42)

	assert.Equal(t, []int64{42}, mock.activateResolvableCalls)
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./internal/topology/service/ -run TestActivateResolvablePending`
Expected: FAIL `mock.activateResolvableCalls undefined` / `svc.activateResolvablePending undefined`。

- [ ] **Step 3: DAO + Repository 实现**

Modify `internal/topology/repository/dao/edge.go` 末尾新增：

```go
// ActivateResolvablePendingEdges 将所有目标节点已存在的 pending 边激活为 active。
// 分两步：先取租户节点 ID 集，再按 target_id 批量更新。返回激活的边数。
func (d *EdgeDAO) ActivateResolvablePendingEdges(ctx context.Context, tenantID int64) (int64, error) {
	nodeIDs, err := d.collectNodeIDs(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	if len(nodeIDs) == 0 {
		return 0, nil
	}

	var total int64
	for _, chunk := range chunkStrings(nodeIDs, 1000) {
		res, err := d.col().UpdateMany(ctx,
			bson.M{"tenant_id": tenantID, "status": domain.EdgeStatusPending, "target_id": bson.M{"$in": chunk}},
			bson.M{"$set": bson.M{"status": domain.EdgeStatusActive, "updated_at": time.Now()}},
		)
		if err != nil {
			return total, err
		}
		total += res.ModifiedCount
	}
	return total, nil
}

// collectNodeIDs 收集租户全部节点 ID（仅投影 _id）。
func (d *EdgeDAO) collectNodeIDs(ctx context.Context, tenantID int64) ([]string, error) {
	cursor, err := d.db.Collection(TopoNodesCollection).Find(ctx,
		bson.M{"tenant_id": tenantID},
		options.Find().SetProjection(bson.M{"_id": 1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []string
	for cursor.Next(ctx) {
		var doc struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		ids = append(ids, doc.ID)
	}
	return ids, cursor.Err()
}

// chunkStrings 将字符串切片按 size 分块，避免单个 $in 过大。
func chunkStrings(items []string, size int) [][]string {
	if len(items) == 0 {
		return nil
	}
	var chunks [][]string
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}
```

（`TopoNodesCollection` 常量来自同包 `dao/node.go`；`time` `bson` `options` 已在文件顶部 import。）

Modify `internal/topology/repository/edge.go`：
- 接口 `EdgeRepository` 增加一行：`ActivateResolvablePendingEdges(ctx context.Context, tenantID int64) (int64, error)`。
- `edgeRepository` 增加委托方法：

```go
func (r *edgeRepository) ActivateResolvablePendingEdges(ctx context.Context, tenantID int64) (int64, error) {
	return r.dao.ActivateResolvablePendingEdges(ctx, tenantID)
}
```

- [ ] **Step 4: 服务层接线 + mock 增补**

Modify `internal/topology/service/topology.go`：
- 新增方法：

```go
// activateResolvablePending 激活所有目标节点已存在的 pending 边（幂等自愈）。
// 失败不阻塞拓扑查询。
func (s *topologyService) activateResolvablePending(ctx context.Context, tenantID int64) {
	if s.edgeRepo == nil {
		return
	}
	if _, err := s.edgeRepo.ActivateResolvablePendingEdges(ctx, tenantID); err != nil {
		_ = err // 激活失败仅忽略，不影响返回
	}
}
```

- 在 `GetBusinessTopology` 开头（第 49 行 `if params.Refresh && ...` 之前）加：
```go
s.activateResolvablePending(ctx, params.TenantID)
```

Modify `internal/topology/service/declaration_property_test.go`：给 `mockEdgeRepo` 加字段与方法：

```go
type mockEdgeRepo struct {
	edges                 []domain.TopoEdge
	activateResolvableCalls []int64
}

func (m *mockEdgeRepo) ActivateResolvablePendingEdges(_ context.Context, tenantID int64) (int64, error) {
	m.activateResolvableCalls = append(m.activateResolvableCalls, tenantID)
	return 0, nil
}
```

（`newMockEdgeRepo` 无需改，切片零值即可。）

- [ ] **Step 5: 跑测试确认通过**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./internal/topology/...`
Expected: PASS。

- [ ] **Step 6: gofmt + 提交**

```bash
cd D:/Haven/e-cam-service
gofmt -l internal/topology/
git add internal/topology/repository/edge.go internal/topology/repository/dao/edge.go internal/topology/service/topology.go internal/topology/service/declaration_property_test.go internal/topology/service/topology_reconcile_test.go
git commit -m "feat(topology): activate resolvable pending edges on query

Add EdgeRepository.ActivateResolvablePendingEdges and invoke it at the start
of GetBusinessTopology, so pending edges whose target node was created by a
non-Register path (e.g. LiveBuilder) self-heal into active.

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

## Task 4: P4 补全节点类型清单

**Files:**
- Modify: `internal/topology/domain/node.go`
- Modify: `e-cam-web/src/api/types/topology.ts`
- Modify: `internal/topology/domain/node_test.go`（追加 Validate 用例）

- [ ] **Step 1: 写失败测试**

在 `internal/topology/domain/node_test.go` 追加：

```go
func TestTopoNode_Validate_ExtendedTypes(t *testing.T) {
	extended := []string{"nlb", "clb", "eni", "mongodb", "vpc", "eip", "nas", "kafka", "elasticsearch"}
	for _, typ := range extended {
		n := TopoNode{ID: "n-" + typ, Name: "node", Type: typ, Category: CategoryNetwork, TenantID: 1}
		assert.NoError(t, n.Validate(), "type %s should validate", typ)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./internal/topology/domain/ -run TestTopoNode_Validate_ExtendedTypes`
Expected: FAIL `invalid node type`（这些类型尚不在 `ValidNodeTypes`）。

- [ ] **Step 3: 补常量 + ValidNodeTypes**

Modify `internal/topology/domain/node.go`：在类型常量块（`NodeTypeS3` 之后）补：

```go
	NodeTypeNLB           = "nlb"
	NodeTypeCLB           = "clb"
	NodeTypeENI           = "eni"
	NodeTypeMongoDB       = "mongodb"
	NodeTypeVPC           = "vpc"
	NodeTypeEIP           = "eip"
	NodeTypeNAS           = "nas"
	NodeTypeKafka         = "kafka"
	NodeTypeElasticsearch = "elasticsearch"
```

`ValidNodeTypes` 补：

```go
	NodeTypeNLB: true, NodeTypeCLB: true, NodeTypeENI: true, NodeTypeMongoDB: true,
	NodeTypeVPC: true, NodeTypeEIP: true, NodeTypeNAS: true, NodeTypeKafka: true,
	NodeTypeElasticsearch: true,
```

Modify `e-cam-web/src/api/types/topology.ts` 的 `NODE_TYPES`（`UNKNOWN` 之后）补：

```ts
    NLB: 'nlb',
    CLB: 'clb',
    ENI: 'eni',
    MONGODB: 'mongodb',
    VPC: 'vpc',
    EIP: 'eip',
    NAS: 'nas',
    KAFKA: 'kafka',
    ELASTICSEARCH: 'elasticsearch',
```

- [ ] **Step 4: 跑测试确认通过**

Run: `cd D:/Haven/e-cam-service && GOFLAGS=-p=2 go test -count=1 ./internal/topology/domain/`
Expected: PASS。

- [ ] **Step 5: gofmt + 提交**

```bash
cd D:/Haven/e-cam-service
gofmt -l internal/topology/domain/
git add internal/topology/domain/node.go internal/topology/domain/node_test.go e-cam-web/src/api/types/topology.ts
git commit -m "fix(topology): add nlb/eni/clb/mongodb/vpc/eip/nas/kafka/es to valid node types

LiveBuilder already emits these types; they were missing from ValidNodeTypes
(and the frontend NODE_TYPES), so any future Validate path would reject them.

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

## Task 5: 全量回归

- [ ] **Step 1: 编译 + 全量单测**

```bash
cd D:/Haven/e-cam-service
GOFLAGS=-p=2 go build ./...
GOFLAGS=-p=2 go test -count=1 ./internal/topology/... ./cmd/apm-push/...
```

Expected: build PASS，test PASS。

- [ ] **Step 2: 前端类型检查**

```bash
cd D:/Haven/e-cam-web && npx vue-tsc -b
```

Expected: PASS（NODE_TYPES 常量新增不产生类型错误）。

- [ ] **Step 3: 提交（如有剩余改动）**

若 Step 1/2 无新改动则跳过。否则 `git add` + conventional commit。