---
feature: "platform-console"
created: "2026-09-30"
status: Draft
task: "5.6 下线就绪清单（ecmdb-web 下线判据）"
---

# teardown-checklist — ecmdb-web 下线就绪清单

> PRD In Scope #8 的交付工件（`prd-spec.md:67`）：产出「下线就绪清单」（ecmdb-web 下线判据，**执行下线本身不属本提案**）。
> 本清单只登记判据与处置项，不在本提案内执行下线（任务 Hard Rules）；未处置项不得自动标「已就绪」。
> 判据来源 = tech-design `跨仓改动设计` RC#2（`tech-design.md:306-318`）+ 多方会话拓扑「过渡期三方」（`tech-design.md:85-105`）+ PRD Out of Scope 工单/变更有意退役（`prd-spec.md:73`）。

## 0. 口径与边界

### 0.1 二义性声明（Hard Rules 承接）

- **本清单是交付工件，不是执行动作**：所有判据项的「处置」列仅描述下线提案须做的事，本提案不执行任何下线操作。
- **未处置 ≠ 已就绪**：状态列默认 `未就绪`；升格为 `已就绪` 须同时满足 ①处置动作在下线提案中完成 ②evidence 留痕 ③规定会签位签字。本提案内任何判据项均保持 `未就绪`。
- **ecmdb-web 离线窗口**：指下线提案执行 ecmdb-web 静态/入口下线的时间窗口；本清单所有判据项的消隐时机均锚定该窗口（PRD：工单/变更有意退役须在「ecmdb-web 在线窗口内执行」）。

### 0.2 状态词表

| 状态 | 含义 |
|------|------|
| 未就绪 | 默认态；判据项尚未在下线提案中处置，或 evidence/会签未齐 |
| 待处置 | 已登记判据与处置路径，等待下线提案执行 |
| 已就绪 | 处置动作完成 + evidence 留痕 + 会签位签字齐备（本提案内不出现） |

### 0.3 执行人与会签位

| 角色 | 职责 | 来源 |
|------|------|------|
| 平台管理员 | 下线执行主体（nginx location 移除、个人设置基址切换登记、退役登记） | tech-design `:283` 回切三件套执行人 |
| eiam owner | 共享 Redis session 读取方收敛会签（session 真相源） | tech-design `:20/:104` cookie 矩阵会签 |
| e-cam-web 前端 owner | RC#2 个人设置跳转基址处置（跨仓改动归属方）会签 | tech-design `:304` RC#1/#2 会签位 |
| e-cam-service 后端 owner | EcmdbAuthMiddleware 读取面收敛确认 | tech-design `:97` 拓扑 |
| 业务 owner | 工单/变更有意退役的近 90 天使用数据确认 + 活跃项签字 | prd-spec `:73` |

---

## 1. 下线判据项

### 1.1 nginx 旧登录 301 location 移除

| 字段 | 内容 |
|------|------|
| **判据** | nginx 中「旧登录入口 301 → `/console/login`」的 location 已移除 |
| **来源** | tech-design `:318`「ecmdb-web 下线时移除该 location 即完成『下线就绪』部署项」；拓扑 `:92/:105` |
| **处置项** | 删除 nginx vhost 中该 301 location（ecmdb-web 仓库零改动）；`nginx -t && nginx -s reload` 生效 |
| **消隐时机** | ecmdb-web 离线窗口内（与静态入口下线同步或紧随） |
| **执行人** | 平台管理员执行 + eiam owner 双人复核（与 tech-design `:283` 回切三件套同名单） |
| **就绪 evidence** | ① nginx 配置 diff 显示该 location 已删 ② `nginx -t` 通过 ③ 旧登录入口直访不再返回 301 → `/console/login`，落至 ecmdb-web 离线后预期态 |
| **状态** | 未就绪 |

### 1.2 RC#2 个人设置跳转基址（VITE_ECMDB_WEB_BASE）处置

| 字段 | 内容 |
|------|------|
| **判据** | e-cam-web `src/layouts/MainLayout.vue:58`「个人设置」跳转基址不再指向已下线的 ecmdb-web 根，无死链 |
| **来源** | tech-design `:311-312` RC#2 第三处隐式耦合；`:312`「ecmdb-web 下线时个人设置跳转随『下线就绪清单』一并处置」 |
| **处置项** | `VITE_ECMDB_WEB_BASE`（个人中心基址）的引用方 MainLayout.vue:58 在 ecmdb-web 下线后重新落点：① 个人中心迁移至控制台 `/console/profile` 并改读控制台基址，或 ② 移除「个人设置」入口。不得保留指向 ecmdb-web 根的旧基址（会落 `/profile/index` 死链）。机制 = build 期 env 注入（`tech-design.md:312`） |
| **消隐时机** | ecmdb-web 离线窗口内，**须与 ecmdb-web 下线同步或前置完成**——否则下线瞬间个人设置入口产生死链 |
| **执行人** | e-cam-web 前端 owner 执行 + 平台管理员会签（跨仓改动归属方，对齐 RC#1/#2 会签位 `tech-design.md:304`） |
| **就绪 evidence** | ① MainLayout.vue:58 引用已改读控制台基址或入口已删 ② build 期 env 不再指向 ecmdb-web 根 ③ Playwright 断言「个人设置」点击不再打开/跳转已下线站点 |
| **状态** | 未就绪 |

### 1.3 共享 Redis session 的 ecmdb-web 读取方下线

| 字段 | 内容 |
|------|------|
| **判据** | ecmdb-web 读取共享 Redis session 的链路随 ecmdb-web 下线而不再被引用；读取方收敛至控制台 + e-cam-web 两方 |
| **来源** | tech-design `:105`「ecmdb-web 自身读同一 Redis session」；`:97` e-cam-service `EcmdbAuthMiddleware` 读同一份；`:20` 控制台复用与 EcmdbAuthMiddleware 同源 session |
| **处置项** | 确认 ecmdb-web 的 session 读取面（经 e-cam-service `EcmdbAuthMiddleware` 或 ecmdb-web 自身中间件）随 ecmdb-web 静态/入口下线消失；e-cam-service 侧 `EcmdbAuthMiddleware` **不退役**（仍服务在线的 e-cam-web），仅确认「ecmdb-web 这一方读取面」消失，无残留 ecmdb-web 专用读取路径或适配分支 |
| **消隐时机** | ecmdb-web 离线窗口内（静态/入口下线后其 session 读取面自然消失；判据 = 确认无残留） |
| **执行人** | eiam owner（Redis session 真相源）+ e-cam-service 后端 owner 会签；平台管理员登记 |
| **就绪 evidence** | ① 路由/grep 核查确认无 ecmdb-web 请求落到 e-cam-service session 读取中间件 ② 共享 Redis session 仍服务控制台 + e-cam-web（未受影响，登证） ③ eiam owner 书面确认读取方收敛至两方 |
| **状态** | 未就绪 |

### 1.4 工单（ticket）/ 变更（change）退役确认签名

| 字段 | 内容 |
|------|------|
| **判据** | ecmdb-web 承载的工单/变更能力按 PRD「有意退役」处置完成：近 90 天使用数据 + 使用方确认齐备，活跃使用项 owner 签字齐备 |
| **来源** | prd-spec `:73`「工单（ticket）/ 变更（change）→ 有意退役（近 90 天使用数据 + 使用方确认后不重建；活跃使用需 owner 签字，由下线提案在 ecmdb-web 在线窗口内执行）」 |
| **处置项** | ① 拉取工单/变更近 90 天使用数据并归档 ② 使用方确认（不重建）书面留痕 ③ 活跃使用项逐项由业务 owner 签字方可退役 ④ 退役登记入下线提案执行记录。**活跃项未签字 → 阻塞 ecmdb-web 下线**（不得绕过签字标已就绪） |
| **消隐时机** | ecmdb-web **在线窗口内**执行（PRD 明示）；须在下线执行前完成，是下线前置门 |
| **执行人** | 业务 owner（工单/变更使用方）签字 + 平台管理员登记；eiam owner 不涉（非鉴权面） |
| **就绪 evidence** | ① 近 90 天使用数据归档件 ② 使用方确认书面留痕 ③ 活跃项 owner 签字清单齐备 ④ 退役登记表 |
| **状态** | 未就绪 |

---

## 2. 下线前置门汇总

下线提案执行 ecmdb-web 下线前，以下判据项须全部 `已就绪`（本提案内均保持 `未就绪`，仅登记门）：

| # | 判据项 | 是否阻塞下线 | 阻塞理由 |
|---|--------|--------------|----------|
| 1.1 | nginx 旧登录 301 location 移除 | 是 | 移除即「下线就绪」部署项本身（tech-design `:318`） |
| 1.2 | VITE_ECMDB_WEB_BASE 处置 | 是 | 不同步则下线瞬间个人设置入口死链 |
| 1.3 | 共享 Redis session 读取方收敛 | 是 | 读取面未收敛 = session 暴露面未消除（残余风险登记 `tech-design.md:289`） |
| 1.4 | 工单/变更退役确认签名 | 是 | PRD 明示须在线窗口内执行；活跃项未签字不得下线 |

> 任何一项未达 `已就绪`，ecmdb-web 不得下线（Hard Rules：未处置项不得自动标「已就绪」）。

---

## 3. 关联文档与追溯

| 关联 | 位置 | 关系 |
|------|------|------|
| PRD In Scope #8 | `prd/prd-spec.md:67` | 本清单的 PRD 授权条目 |
| PRD Out of Scope 工单/变更有意退役 | `prd/prd-spec.md:73` | 判据项 1.4 来源 |
| tech-design 跨仓改动设计 RC#2 | `design/tech-design.md:306-315` | 判据项 1.2 来源 |
| tech-design 过渡期 ecmdb-web 登录重定向 | `design/tech-design.md:316-318` | 判据项 1.1 来源 |
| tech-design 多方会话拓扑「过渡期三方」 | `design/tech-design.md:85-105` | 判据项 1.3 来源 |
| tech-design 显式接受边界（残余风险登记） | `design/tech-design.md:289` | 判据项 1.3 阻塞理由（ecmdb-web 下线即消除第三方暴露面） |
| tech-design 回切三件套执行人 | `design/tech-design.md:283` | §0.3 执行人会签位来源 |

---

## 4. 变更记录

| 日期 | 任务 | 变更 | 作者 |
|------|------|------|------|
| 2026-09-30 | 5.6 | 建立清单骨架与 4 项判据，承接 eval 残留 #3（把「下线就绪」从被动引用升级为显式交付工件） | forge-task-executor |
