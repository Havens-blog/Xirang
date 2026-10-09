# eiam owner 收尾交办清单

> 平台控制台（platform-console）已构建完成（35 任务 + 验收 E2E 61/61）。以下三块需 **eiam owner（Duke1616）** 排期/会签，其余缺口已在控制台侧按「UI 降级/裁剪」落地。本清单一页交办，evidence 全部 file:line 落在 `parity-checklist.md`。

## 一、会签（eiam owner 待签位）

| 待签材料 | 位置 | 会签内容 |
|----------|------|---------|
| 端点差集盘点 + 契约基线冻结 | parity §7.2 / §4.3 | 15 端点契约基线、G-1~G-4 处置、「变更 ≥5 工作日书面通知」条款 |
| 认证链路兼容矩阵 | parity §6.1.9 | cookie 属性/Session JWT 结构/tenant_id 写入断言/tenant·switch 语义 |
| 服务端授权核查矩阵 | parity §6.2.5 | 89 条 L2 逐端点 + G-7 处置建议 + default-allow 注记 |
| platform/users 能力映射 | parity §6.3.6 | 19 项三选一落点 + G-8/G-9 处置 |

> **平台管理员 + e-cam-web 前端 owner 两方已由项目 owner（Havens-blog）2026-10 授权签署**（RC#1 任务 5.4 记录在案）。剩 **eiam owner** 一方待签。

## 二、安全缺口 G-7（最高优先，需排期）

**`POST /api/v1/discovery/sync` 无鉴权资产上报写面**（discovery/handler.go:21-26，实现 :33-50 仅校验 service 非空）。

- **后果**：任意可达调用方可提交 SyncRequest **全量重写**指定 service 的权限点 + API 资产 + code 绑定（ingestion/engine.go:58-104 全清重建）→ ① 清空 code 绑定使对应 API 降级「仅登录放行」（permission.go:146-152），② 篡改绑定/权限资产致授权面漂移。
- **当前暴露面**：dev nginx 未转发 `/api/v1/`，控制台同源面不可达；风险 = eiam `:9000` 直连网络面。
- **建议处置**：上报面加服务间凭证/来源绑定，或内网隔离 + 网络 ACL。未补齐前须确保 `:9000` 不暴露于不可信网络。

## 三、端点缺口（控制台阻塞场景的 eiam 侧依赖）

| 缺口 | 阻塞场景 | eiam 侧需补 | evidence |
|------|----------|-------------|----------|
| **G-1** | S-M1-11 管理员重置密码 | `user/reset_password`（一次性初始密码） | §1.2 全量导出反证；user/update 无 password（user/vo.go:124-127） |
| **G-2** | S-M3-04 passkey 注册引导 | 管理员代发起 passkey 注册（现 register 为 L1 本人自服务，user.go:668-707） | §6.2.2 #24 |
| **G-3** | S-M3-05 passkey 撤销重绑 | 核验 `user/identity/unbind` 是否覆盖 passkey 完整性 + 补响应样本 | §6.2.1（unbind 为 L2 强制 ✓，仅差核验） |
| **G-8** | E1 首登强制改密（连带 G-1） | `must_change_password` 字段 + 登录链路强制拦截 + 改密后清除 | §6.3.5（全仓 grep must_change 0 命中） |

> G-4（租户禁用需会话数）、G-5（资源词表）、G-6（三元组授权）、G-9（锁定态列表）已在控制台侧按 UI 降级/静态词表/语义映射落地，**不阻塞交付**；若业务要求原生能力，可按 parity §3.1 处置建议单独立项。

## 收尾确认路径

1. eiam owner 签署 §7.2 会签表（四张表）。
2. G-7 安全修复排期（建议加急）。
3. G-1/G-2/G-3/G-8 端点排期 → 完成后回填 acceptance.md 的 S-M1-11 / S-M3-04·05 / E1 场景为 REAL 档。