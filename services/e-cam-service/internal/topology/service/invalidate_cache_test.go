package service

import (
	"context"
	"testing"

	"github.com/Havens-blog/e-cam-service/internal/topology/domain"
	"github.com/stretchr/testify/assert"
)

// TestInvalidateTenantCache_DeletesOnlyLiveBuilderNodes 资产同步后的缓存失效
// 只清除 LiveBuilder 产生的 dns_api / cloud_api 来源节点，APM 服务与人工声明节点保留。
func TestInvalidateTenantCache_DeletesOnlyLiveBuilderNodes(t *testing.T) {
	nodeRepo := newMockNodeRepo()
	nodeRepo.nodes = map[string]domain.TopoNode{
		"dns-1":  {ID: "dns-1", SourceCollector: domain.SourceDNSAPI},
		"inst-1": {ID: "inst-1", SourceCollector: domain.SourceCloudAPI},
		"svc-1":  {ID: "svc-1", SourceCollector: domain.SourceAPM},
		"decl-1": {ID: "decl-1", SourceCollector: domain.SourceDeclaration},
	}
	edgeRepo := newMockEdgeRepo()
	svc := &topologyService{nodeRepo: nodeRepo, edgeRepo: edgeRepo}

	svc.InvalidateTenantCache(context.Background(), 42)

	// LiveBuilder 来源被清除
	assert.NotContains(t, nodeRepo.nodes, "dns-1")
	assert.NotContains(t, nodeRepo.nodes, "inst-1")
	// APM 与声明来源保留
	assert.Contains(t, nodeRepo.nodes, "svc-1")
	assert.Contains(t, nodeRepo.nodes, "decl-1")
}