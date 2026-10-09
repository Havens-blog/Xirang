---
created: "2026-09-14"
author: "Haven"
status: Draft
intent: "refactor"
---

> **执行进度（2026-09-14，Phase 2 全部完成并推送 origin/main）**：S3c/S3d/S4 已完成（`583d7c9`，-2570 行）；**S5 diffAndUpsert 抽公共差集删除**（`ca33088`，-629 行，18 个 syncRegion 收敛到 syncItem/diffAndUpsert，签名不变、3000 行测试零改动全绿；sync_dns 写裸 mongo、sync_oss 全局服务无删除语义，保留原实现）；**S3a ChangeTracker 审计迁入 executor**（`80d015f`：挂钩 diffAndUpsert 单点覆盖 18 类型、ChangeTracker 接口 nil 关闭、ioc 显式接线——发现 WireChangeTracker 此前从未被调用=审计从未生效；单测覆盖追踪/默认关闭）；**S5 表驱动化**（`b544114`，+99/-352：syncRegionAssets 20-case switch 收敛为 sync_registry.go 两张派发表 assetSyncFns/cloudxSyncFns，别名路由同一函数、19 个失败日志标签逐字保留、dns 显式跳过、未注册类型 default 告警且不触碰工厂，行为与旧 switch 一致）。legacy `/assets/sync` 已转队列（S4 采纳"转队列后删"），关系同步已删（S3c 采纳"删"）。推送：`80d015f`/`b665b14`/`b544114` 均已到 origin/main。

# Proposal: 同步逻辑收敛 Phase 2——asset_sync 死服务清退 + executor 内部表驱动重构

## Problem

2026-09-14 已完成同步收敛第一阶段（96c94d3）：删除 cam/sync 原型、account.go 死块、task/{queue,repository,domain} 死代码，修 MCP nil 队列 panic。权威同步实现已收敛到 `internal/cam/task/executor`。

仍有两块未动（均为行为相关，需谨慎）：

1. **asset_sync 死服务清退**：`internal/cam/service/asset_sync_*.go`（17 文件 + asset_dns.go + asset_relations.go，约 2321 行）零调用（唯一触碰是 `ioc/audit.go:26 SetChangeTracker` 给死服务注入审计器）。但其承载的 **ChangeTracker 审计变更追踪只在死路径存在**——live executor 完全没有审计能力，清退前需搬迁。
2. **executor 内部 19 份复制**：`sync_assets.go` 的 20-case switch 每 case 16 行懒加载重复；`syncRegion<X>` 内联重复差集删除（`sync_ecs.go:30-56` ×19）。另 ECS 转换逻辑全仓共 4 份。

### Evidence

- 实测：`AssetSyncService.SyncAssets/SyncAccountAssets/SyncRelations` 全仓仅定义处（2026-09-14 grep）；`SetDNSCollections` 无人调用（DNS 同步会静默跳过）；`asset_relations.go:78` 用 `ModelUID: "cloud_vm"` 查 ECS，与两代实际写入的 `%s_ecs` 永不匹配。
- executor 测试齐全（`sync_region_test.go` 65KB 等），是清退与重构的安全网。
- legacy 第四份（`cam/service/asset.go:284 SyncAssets` → 写旧集合 `ecam_cloud_asset`）为 live 但语义陈旧，需产品决策。

## Proposed Solution

### S3a：ChangeTracker 搬迁（先做，行为新增）
- 给 `SyncAssetsExecutor` 加可选 changeTracker 字段（`ioc/audit.go` 的 `WireChangeTracker` 改注入 executor）；单测断言 Upsert 前后 `TrackChanges` 被调用。
- 验证：`go test ./internal/cam/task/executor/...` + 行为对比。

### S3b：清理语义对齐
- 将 `cleanupStaleInstances` 的"按模型全局清理"与 DNS upsert 语义对照进 executor（diff 确认字段一致后合并）。

### S3c：关系同步决策
- `SyncRelations` 无人调用且 ModelUID 已错：直接删；若产品要关系同步，作为新任务在 executor 侧重建（修正为 `%s_ecs`）。

### S3d：清退 asset_sync 死服务
- 删 `asset_sync.go` 接口+实现、`cam/wire.go:111`、`module.go:32` 字段，`ioc/audit.go:23-26` 改指 executor。
- 验证：`go build ./...` + `go vet ./...` + 全量 `go test ./...` + `grep AssetSyncService|assetSyncService|asset_sync_` 归零。

### S4：legacy `/assets/sync` 决策（需产品确认）
- `asset.go:284` 写旧集合 `ecam_cloud_asset`：建议改投队列（复用 SyncAccount 模式）后删 `asset.go:281-548` 同步段。验证：API 集成测试（task_id 返回、ecam_task 记录、ecam_instance 增长、ecam_cloud_asset 停增）。

### S5：executor 表驱动重构
- 20-case switch → `map[string]syncRegionFn`；抽 `diffAndUpsert` 通用差集删除供 19 个 syncRegion 复用。
- 验证纪律：沿用 split-sync-assets-executor 的"测试文件零改动"（`docs/proposals/split-sync-assets-executor/proposal.md:223-248`）。

### S6 收尾（已部分完成）
- 统一 `TaskTypeSyncAssets` 常量到 `internal/cam/task/types.go`（删 3 处副本）；统一默认类型清单（`asset_types.go` 漂移，注释已登记）。

## Success Criteria

- [ ] `AssetSyncService` 全仓零引用，`asset_sync_*.go` 目录删除
- [ ] ChangeTracker 审计在 executor 路径生效（Upsert 触发 TrackChanges），单测覆盖
- [ ] executor 单测（含 -race）全绿且文件零改动
- [ ] legacy 同步路径按 S4 决策收敛或显式保留
- [ ] `grep TaskTypeSyncAssets` 单一定义

## Risks

| Risk | Mitigation |
|---|---|
| ChangeTracker 搬迁改变审计行为 | 先 S3a 单测锁定行为再清退 |
| 清理语义对齐不精确 | diff 逐字段核对，executor 测试兜底 |
| S4 影响前端同步返回值 | 先确认前端无 `total_synced` 依赖（vo.go:91） |
| executor 重构破坏 3000 行测试 | 沿用零改动纪律 + `-race` 全绿门 |
