# 决策记录：CMDB 双实现收敛 —— 保留 ecmdb，e-cam-service 的 CMDB 退位

> **已废弃（2026-10-08）** —— 本决策「保留 ecmdb 为唯一 CMDB」的结论已被
> [`0003-remove-ecmdb.md`](./0003-remove-ecmdb.md) 推翻：ecmdb 与 ecmdb-web 的全部承载能力
> 已废弃，两个 Duke1616 上游 checkout 已物理删除。原文保留以供历史审计，**不再作为有效决策依据**，
> 后续以 0003 为准。

| 项目 | 内容 |
|---|---|
| 状态 | 已废弃（2026-10-08，被 [0003-remove-ecmdb.md](./0003-remove-ecmdb.md) 取代）；原记录日期 2026-09-08，迁移未开始 |
| 决策人 | 项目负责人 |
| 影响范围 | ecmdb、e-cam-service、e-cam-web、ecmdb-web、nginx 路由 |

## 1. 背景

平台存在两套 CMDB 实现：

1. **ecmdb**（独立服务 :8000）——动态模型/属性/实例/关系的权威实现，mongox 已内置租户自动注入（`ecmdb/pkg/mongox/collection.go:197-238`，`tenant_id` 来自 eiam `ctxutil`），通过 nginx `/api/cmdb` 对外服务，ecmdb-web 直接消费。
2. **e-cam-service/internal/cmdb**——ecmdb fork 后并入单体的副本，挂 `/api/v1/cam/cmdb/...` 路由，供 e-cam-web `views/cmdb` 消费。云资产同步（`e-cam-service/internal/cam/service/asset_sync.go`）**绕过 ecmdb API 直写同一 Mongo 的 `c_instance` 集合**，靠"约定 model_uid 命名（aliyun_ecs 等）"与 ecmdb 隐式耦合。

双写共库导致模型/属性定义必然漂移，是全平台最大架构风险。

## 2. 决策

**CMDB 的逻辑统一归属 ecmdb 模块。** e-cam-service 退位为"云资产采集与多云执行层"，不再内嵌 CMDB 实现；资产数据写入必须经过 ecmdb。

## 3. 分阶段迁移路径

### Phase 0 — 冻结与盘点（低成本，先行）
- e-cam-service `internal/cmdb` 冻结：不再新增功能，仅修 P0 缺陷。
- 盘点 e-cam-web `views/cmdb` 对 `/api/v1/cam/cmdb` 的实际 API 面与 ecmdb 现有 API 的差距（动态列、加密字段、关系图谱、拓扑）。

> **Phase 0 盘点结论（2026-09-14 实测完成，据 e-cam-web/src/api/cmdb.ts 与 ecmdb/internal/web/*）：**
>
> **总体：ecmdb 核心能力是超集**（模型/属性/关系/拓扑/加密字段/自定义列/导入导出均为 e-cam-web 缺失的增益项），但**不能零改造承接**。前端在用的后端端点约 28 个；`/cam/cmdb/instances` 写入套件前端已不直连（batch/upsert 走内部通道）。
>
> **阻塞项（按影响排序，即 Phase 1 必须补齐的缺口）：**
> 1. **实例列表查询（高）**：前端 `GET /cam/instances` 带 15+ 属性过滤参数 + 标签过滤（实例页 + service-tree 页共用）；ecmdb `POST /api/resource/list` 仅 model_uid+分页、无属性/关键字/标签过滤。需扩展 ListResourceReq 或新增查询端点。
> 2. **租户隔离谓词（高）**：ecmdb 全链路无租户过滤——`tenant_id` 仅残留字段（`dao/resource.go:433`），baseFilter 无租户谓词。e-cam 侧 RequireTenant + DAO 谓词语义必须补齐，否则多租户串库。
> 3. **批量写入 HTTP 端点（中高）**：ecmdb 无 bulk HTTP 端点（仅单条 create/update）；批量能力在内部 dataio 链路（`dao/resource.go:354 BatchCreateOrUpdate`）。云资产同步若走 HTTP 需新增 bulk upsert 端点或暴露内部批量接口。
> 4. **API 形状/鉴权全面不适配（中）**：REST+`/api/v1`+code 信封 vs POST-action+EIAM capability；uid vs id 寻址；`asset_id/account_id` 业务字段需下沉 data 或建模；模型删除依赖检查（501002/501003）语义更严；attribute 强制 group_id>0。需 BFF 适配层或前端 API 层重写（后者建议，前端 CMDB 页面本就要重做）。
> 5. **小缺口（低）**：field-types 枚举端点、model-groups /init、通用 group update、平铺模型列表路由（`model/handler.go:285` ListModels 有实现未注册）。
>
> **零调用死代码（可顺带清）**：`createCmdbInstanceBatchApi`/`upsertCmdbInstance*Api`/实例关系全套/`getRelatedInstancesApi` 等（`src/api/cmdb.ts`）。

### Phase 1 — ecmdb 补齐能力缺口
- 按 Phase 0 盘点结果补 ecmdb API；重点验证：实例批量写入性能（资产同步吞吐）、动态自定义列、关系/拓扑查询是否满足 e-cam-web。
- 若资产同步需要无用户会话的服务身份写入，明确 ecmdb 的 service-to-service 鉴权方式（eiam 客户端凭证或专用 ingest 端点）。

### Phase 2 — 写入路径切换（关键步）
- `asset_sync` 改为调用 ecmdb（HTTP API 或新增 gRPC ingest 接口），携带租户上下文；停止直写 `c_instance`。
- 上线前做存量数据对账：`c_instance` 的 `tenant_id`/`model_uid` 两侧一致性校验脚本。

### Phase 3 — 前端切换
- e-cam-web `views/cmdb` 改调 ecmdb API（nginx `/api/cmdb` 路由已存在）；权限口径对齐 ecmdb 的 eiam 能力码体系。

### Phase 4 — 退位删除
- 删除 `e-cam-service/internal/cmdb`、其路由注册与 collection 直写代码；nginx 保留 `/api/v1/cam/cmdb` 短期重定向后移除。

> **补记（2026-09-09，dead-modules-cleanup 提案联动）：`internal/asset` 归属决断项。**
> `e-cam-service/internal/asset` 继承了与 `asset_sync` 同源的历史包袱——直写 `c_instance` 集合（绕过 ecmdb API），现仅存 cert 域覆盖率分母的只读聚合消费（`cert/service/coverage_meta.go`、`ioc/cert.go`）；其零引用外壳 `module.go` 的删除与 `doc.go` 的按代码事实正名**已由** dead-modules-cleanup 任务 4（`docs/features/dead-modules-cleanup/tasks/4-rename-account-asset.md`）执行（e-cam-service 提交 `0a7bcbc`，2026-09-09），`internal/asset/doc.go` 现已链接本决策文件作为唯一事实源。**该目录的最终归属（收编进 cert 域 / 随 CMDB 收敛删除直写路径后重新定位）随本决策 Phase 4 一并决断**，避免与 Phase 2（写入路径切换、停止直写 `c_instance`）及 Phase 4（退位删除）的删除面重叠造成二次返工。
> 坐标系说明：本决策文档位于**工作区根** `D:\Haven\docs\decisions\`，不在 e-cam-service 仓库内；任务 4 已落地，`e-cam-service/internal/asset/doc.go` 现反向链接本条，双向锚点成立。本补记仅追加联动项，不改变上述既有决策结论。

## 4. 风险与开放问题

- **模型命名契约**：`aliyun_ecs` 等 model_uid 必须成为显式契约（建议 ecmdb bootstrap YAML 固化），不再靠注释约定。
- **同步吞吐**：ecmdb API 单条写入若不满足资产同步吞吐，需设计批量 ingest 接口，而非回退直写。
- **服务身份与租户**：同步任务无用户会话，ecmdb 需定义服务身份携带租户的方式（当前 mongox 依赖 `ctxutil.GetTenantID`）。
- **切换窗口**：Phase 2~3 期间前端与写入路径需兼容双读或灰度切换。

## 5. 关联

- 2026-09-08 架构审查报告（对话内）；gRPC 死链路已拆除（e-cam-service 的 ecmdb Policy/Endpoint gRPC 客户端与空 server 移除），为后续 ecmdb 正式 gRPC 接口腾出干净起点。
