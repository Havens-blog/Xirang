---
id: "T-review-doc"
title: "Review Documentation Quality"
priority: "P1"
estimated_time: "30min"
dependencies: ["1.3", "5.6", "1.1", "1.2", "1.4", "5.7", "1.5"]
type: "doc.review"
surface-key: ""
surface-type: ""
---

Review documentation quality for the platform-console feature (breakdown mode).

## Acceptance Criteria Summary

The following acceptance criteria are pre-extracted from doc tasks. Use these as the review baseline.

### 1.1-endpoint-parity
- [ ] eiam 路由全量导出：从 eiam 仓库路由注册文件导出全部 HTTP 端点清单（含方法），作为差集求值的「实有端点」集；证据附文件:行号
- [ ] 差集求值：UF-1~11 反推的端点全集 − eiam 实有端点 = 差集清单；A 档（e-cam-web 生产在用）逐条核对形状差异，B/C 档逐项标注「待核验假设」状态
- [ ] 每条缺口登记处置（三选一）：eiam 排期补齐 / UI 降级仅查看 / 放弃需业务 owner 书面确认并记理由
- [ ] 契约基线：冻结本控制台依赖的端点快照（路径/方法/鉴权要求/关键字段），注明「变更须 ≥5 工作日书面通知 + 影响评估」
- [ ] 为每类核查项填写证据来源形态（路由注册文件:行号 / 响应样本 / owner 书面确认）与执行人（含 eiam owner 会签位）


### 1.2-auth-chain-audit
- [ ] cookie 矩阵：逐项记录 name/domain/path/SameSite/Secure/httpOnly 实测值，与 e-cam-web 严格一致的项显式标注
- [ ] session 结构：JWT claims（uid/username/tenant_id）与 `GET /api/iam/user/profile` 返回结构（user/permissions/tenants/current_tenant_id/is_admin/must_select_tenant）逐字段核实
- [ ] tenant_id 写入断言：是「始终写入」还是「条件写入」，零租户用户取值为何，登记结论
- [ ] tenant/switch 行为：核实 X-Active-Tenant-ID 头 + 销毁旧 session 重签 JWT + 调用方须 reload 的语义；cookie 重签到达客户端方式（Set-Cookie vs 前端写 cookie）登记
- [ ] 矩阵经 eiam owner 书面会签（签名/批注留档在 parity-checklist）


### 1.3-authz-audit
- [ ] 覆盖 1.1 反推出的全部管理端点（user/role/tenant/organization/identity_source/policy/authorization CRUD + reset_password + passkey 注册/撤销）
- [ ] 逐端点登记：是否服务端鉴权、鉴权方式（session/JWT）、授权粒度（平台级/租户级/越租户阻断）
- [ ] 前端裁剪与 403 后端兜底的对应关系逐端点标注（越权请求返回 403 的路径确凿）
- [ ] 缺失授权保护的端点 → 登记缺口处置（eiam 补齐 / 该流裁剪），不得静默放行


### 1.4-capability-map
- [ ] 能力清单：逐功能点列出当前实现方式与依赖的 eiam 端点（含 file:line 证据）
- [ ] 三选一落点：每项标注 映射到 UF-4 / 有意不迁移（登记理由）/ eiam 缺口（进缺口处置），无「未标注」残留
- [ ] 引用复盘：重跑全仓 grep 复核 eiam-users.ts / eiam-mapper.ts / eiam-tenant.ts 引用方（与设计实核结论一致性）
- [ ] 双签（+e-cam-web 前端 owner 会签）材料齐备：截图/接口响应/用例链接逐项附 evidence


### 1.5-acceptance-matrix
- [ ] E1–E12 每条 ≥1 场景，UF-1~11 每 UF ≥1 条，逐条标注来源（E#/UF#）与类型（正向/负向/边界）
- [ ] 负向 + 边界场景合计 ≥ 总数 1/3（按 ≥24 条计即 ≥8 条）
- [ ] 每条场景挂所属里程碑（M1/M2/M3/M4），骨架期只填「场景项 + 来源映射」，执行期补 Playwright 用例链接/MCP 会话记录
- [ ] owner 标注：前端 owner + 测试 owner 双署，涉 eiam 断言标 eiam owner 会签位


### 5.6-teardown-checklist
- [ ] 判据项：nginx 旧登录 301 location 移除、RC#2 个人设置跳转基址（VITE_ECMDB_WEB_BASE）处置、共享 Redis session 的 ecmdb-web 读取方下线、工单/变更退役确认签名
- [ ] 每项标注消隐时机（ecmdb-web 离线窗口）与执行人/会签位
- [ ] 二义性明确：执行下线本身不属本提案，清单为交付工件


### 5.7-deploy-hardening
- [ ] 探活：nginx /console/health 探活接入告警，与 e-cam-web 同运维面
- [ ] 回切演练 runbook：三件套逐个演练（注释行切换/双池 + 静态软链 + 301），记录步骤/耗时/回退验证，≥10 分钟内可回切（PRD 承诺）
- [ ] CSP 收敛实测：按真实构建产物逐指令收敛，style-src 'unsafe-inline' 可去除则收紧并记录（不可去除则记原因）
- [ ] 双人复核登记：回切演练记录 + CSP 收敛结论入部署清单，平台管理员 + eiam owner 双签


## Discovery Strategy

Scan ONLY the following allowlist of directories for target documents:
- docs/features/platform-console/ (prd/, design/, testing/, and any subdirectories)
- docs/proposals/platform-console/

EXCLUDE the following from scanning — do NOT read or process these:
- tasks/ directory (task definitions are not deliverables)
- tasks/records/ directory (execution records are not deliverables)
- manifest.md (build artifact)
- index.json (build artifact)

Only .md files under the allowlist directories are target deliverables.

## Acceptance Criteria

- [ ] All acceptance criteria met
