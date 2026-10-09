---
feature: "platform-console"
created: "2026-10-08"
status: Draft
task: "5.7 部署与运维加固（探活监控 + 回切演练 + CSP 收敛）"
---

# deploy-hardening — 部署与运维加固

> PRD `prd-spec.md` §Monitoring Requirements（探活接入告警、与 e-cam-web 同运维面；上线前回切开关演练 10 分钟内）+ §Security Requirements（CSP / X-Frame-Options / HSTS 由 nginx 注入）的交付工件。
> tech-design `tech-design.md` §Integration Specs 5（回切三件套 + staging 演练计时入部署清单）+ §Security（CSP 指令基线、style-src 实测收敛）。
> 交付形态：nginx 配置改动（`haven-console/nginx/haven-console.conf`）+ 本 runbook/记录。本提案仅产工件与配置，**不在本提案内执行上线/演练**（见 §0.2 边界）。

## 0. 口径与边界

### 0.1 与既有工件的关系

| 工件 | 位置 | 本任务关系 |
|------|------|-----------|
| nginx 同域入口（三件套 location + CSP 基线） | `haven-console/nginx/haven-console.conf`（任务 2.10 落地） | **本任务修改**：加 `/console/health` 探活 + 收紧 `style-src`（移除 `'unsafe-inline'`） |
| 部署与运维约定（TECH-deployment-001/002） | `docs/conventions/deployment.md` | 已抽取的跨特性约定，本任务是其落地产出之一 |
| WatchAlert 接入决策 | `docs/decisions/0002-watchalert-integration.md` | 「同运维面、同告警渠道」的权威来源（反代 :8891） |
| e-cam-web nginx 探活形态 | `e-cam-web/nginx.conf:110` / `nginx.cam.conf:57` | 「同运维面」的镜像基线（`location /health` 形态） |
| 下线就绪清单 | `teardown-checklist.md`（任务 5.6） | 下线侧工件；本任务为上线侧工件，二者互补 |

### 0.2 二义性声明

- **本工件是交付件，不是执行动作**：§2 回切演练 runbook 给出步骤与计时口径，**staging 完整演练本身**（PRD 承诺 10 分钟内）须由平台管理员在上线前执行一次并回填实际耗时 / 回退验证证据；本提案内 runbook 列「待演练」态，未演练不得上线（Hard Rules）。
- **探活 / CSP 改动不得绕过「同运维面 + 双签」约束**（Hard Rules）：探活接入告警须走 e-cam-web 同面（WatchAlert 同渠道，不新增独立 SLA）；CSP 收敛须平台管理员起草 + eiam owner 复核双签，随认证链路兼容矩阵（parity-checklist F-4 / 1.2）双签生效。

### 0.3 执行人与会签位

| 角色 | 职责 | 来源 |
|------|------|------|
| 平台管理员 | nginx 配置起草（探活 + CSP 收敛）、回切演练执行主体、演练记录回填 | tech-design `:283` 回切三件套执行人 |
| eiam owner | CSP 收敛复核会签（同认证链路兼容矩阵双签口径） | tech-design `:343` CSP 双签 |
| WatchAlert 运维 owner | 探活探测器配置（GET /console/health 期望 200）+ 告警规则接入 | ADR-0002 同运维面 |

---

## 1. 探活健康检查 + 监控告警接入（与 e-cam-web 同运维面）

### 1.1 nginx /console/health 探活端点

**已落地**（`haven-console/nginx/haven-console.conf` §`location = /console/health`）：

```nginx
location = /console/health {
    access_log off;
    add_header Content-Type text/plain;
    return 200 "ok\n";
}
```

形态镜像 e-cam-web `nginx.conf:110` 的 `location /health`（`access_log off` + `200` 纯文本），与 e-cam-web 同一 nginx 运维面、同一告警渠道。设计取舍：

- **独立 exact-match location**（`location =`）：不继承 `/console/` 前缀 location 的 CSP / 缓存头——探活端点无页面语义，与 e-cam-web 同口径（e-cam-web 的 `/health` 亦不带 CSP）。
- **nginx 层不产应用状态**（纯静态站）：探活 = 静态站可达 + 反代 upstream 双层可达性。返回 200 仅证「nginx 入口活 + 静态池可读」。
- **eiam 上游不可达** 由 `/api/iam/*` 反代 5xx 体现，由同面告警规则覆盖——**不单设 eiam 探活**（eiam 自有运维，跨服务边界不重复布探针）。
- **body `ok\n`**：与 e-cam-web `nginx.cam.conf:59` 的 `ok\n` 同形，便于同一条 WatchAlert 探测器规则复用（期望 200 + body 含 `ok`）。

### 1.2 监控告警接入（WatchAlert 同渠道）

**告警平台**：WatchAlert（反代 `:8891`，后端 `:9001` / 前端 `:3001`，ADR-0002 决策方案 A「反代接入、源码零改动」）。控制台探活接入与 e-cam-web **同一告警渠道**，不新增独立 SLA（PRD Monitoring Requirements 硬约束）。

**探测器配置**（WatchAlert 运维 owner 落地，本任务只登记口径）：

| 字段 | 值 | 说明 |
|------|----|----|
| 探测目标 | `GET http://<console-host>:<port>/console/health` | 与 e-cam-web `GET /health` 同面登记，host/port 同 vhost |
| 期望 | HTTP 200 + body 含 `ok` | 与 e-cam-web 同一条规则可复用 |
| 探测频率 | 与 e-cam-web 同档（建议 30s，沿用现役 e-cam-web 探测周期） | 不引入独立 SLA |
| 告警阈值 | 连续 N 次失败触发（N 与 e-cam-web 同档） | 同面同口径 |
| 告警渠道 | WatchAlert 现役通知渠道（与 e-cam-web 同一收件人组） | 不新建独立通知链 |

**不设独立 SLA 的边界**：控制台与 e-cam-web 共用 nginx 入口（同 vhost，`:8888` 生产形态按宿主调整），任一侧 nginx 层不可达会同时影响双侧探活——同面告警即覆盖，无需为控制台单设 SLA。控制台静态池独立（`/console/*` alias 独立目录），静态产物损坏仅影响 `/console/health` 200 输出，与 `/cam/*` 互不影响。

### 1.3 活体验证（本任务执行）

独立 nginx 1.25.4 实例（端口 `:8899`，不动现役 `:8888` dev vhost）活体冒烟：

| # | 断言 | 结果 |
|---|------|------|
| 1 | `GET /console/health` → `200 OK` | PASS |
| 2 | `Content-Type: text/plain` 命中 | PASS |
| 3 | body = `ok\n`（3 字节） | PASS |
| 4 | 响应**不含** CSP / X-Frame-Options / HSTS / Cache-Control（探活端点无页面语义，与 e-cam-web 同口径） | PASS |

> 探活接入 WatchAlert 告警规则本身（探测器条目 + 通知渠道绑定）属 WatchAlert 运维 owner 在生产/预发环境的配置动作，不在本提案代码面内；本任务交付 nginx 探活端点 + 登记同面口径，探测器落地由 §0.3 WatchAlert 运维 owner 按本表执行。

---

## 2. 回切开关三件套演练 runbook（PRD 承诺 10 分钟内）

### 2.1 三件套总览

| # | 开关 | nginx location | 切换机制 | 切换粒度 | 影响面 |
|---|------|----------------|----------|----------|--------|
| ① | `/console` 静态池（控制台自身） | `location ^~ /console/` | 注释行切换 `alias` + `nginx -s reload`（或软链重指，免 reload） | 秒级 | 控制台静态产物回上一版，e-cam-web / eiam API 不受影响 |
| ② | e-cam-web 401 目标（新旧双产物） | `location ^~ /cam/` | 注释行切换 `alias`（`dist` 现役池 ↔ `dist-old` 回切池） + `nginx -s reload` | 秒级 | e-cam-web 401/登出跳转目标在 `/console/login`（新）↔ ecmdb-web `/login`（旧）间切换 |
| ③ | 过渡期旧登录入口 301 | `location = /login` | 注释/启用整 location + `nginx -s reload` | 秒级 | 旧登录入口是否 301 收敛至 `/console/login`（ecmdb-web 下线时移除本 location） |

全部**无 rebuild 可执行**（build 期 env 注入、新旧双产物并存；tech-design §Integration Specs 5）。执行人：平台管理员 + eiam owner 双人复核（与 Related Change #1 双签同名单，tech-design `:283`）。

### 2.2 演练前置条件（staging）

- [ ] staging 环境已部署现役控制台产物（`haven-console/dist/`）+ 上一版产物（`dist-previous/` 或 `dist-old/` 作为回切池）
- [ ] staging e-cam-web 已部署双产物（`dist/` 现役 = 401 跳 `/console/login`；`dist-old/` 回切 = 401 跳 ecmdb-web `/login`）
- [ ] staging nginx 已并入 `haven-console.conf`（含本任务新增的 `/console/health` 探活）
- [ ] staging eiam `:9000` 可达（演练需真实登录态验证 401 跳转）
- [ ] 平台管理员 + eiam owner 双人在场（演练计时 + 复核）
- [ ] 演练记录表（§2.5）就绪

### 2.3 逐件演练步骤 + 计时口径

**计时口径**：T0 = 第一个 `nginx -s reload` 命令回车时刻；T_end = 三件套全部回切完成且 §2.4 回退验证全过时刻。PRD 承诺 T_end − T0 ≤ 10 分钟。**不含前置准备时间**（产物就位 / 双人到场），仅计切换动作 + reload + 验证。

#### ① /console 静态池回切（现役 → 回切池）

| 步骤 | 动作 | 预期 | 累计耗时（目标） |
|------|------|------|------------------|
| 1.1 | 编辑 `haven-console.conf` `/console/` location：注释现役 `alias .../dist/;` 行，启用 `# alias .../dist-previous/;` 行 | 配置 diff 仅一行切换 | ~30s |
| 1.2 | `nginx -t` 验证语法 | `syntax is ok` / `test is successful` | ~5s |
| 1.3 | `nginx -s reload` | 平滑 reload，无中断 | ~3s |
| 1.4 | 验证：`GET /console/` 出回切池标记（如版本号 / 构建时间戳回退至上一版） | 200，产物为 previous | ~10s |
| 1.5 | 验证：`GET /console/health` 仍 `200 ok`（探活不受池切换影响） | 200 + body `ok` | ~5s |
| 1.6 | 验证：`GET /console/login` 仍可加载（SPA 回退完整） | 200，login 页渲染 | ~10s |

小计目标 ≤ 1 分钟。

#### ② e-cam-web 401 目标回切（新产物 → 旧产物）

| 步骤 | 动作 | 预期 | 累计耗时（目标） |
|------|------|------|------------------|
| 2.1 | 编辑 `/cam/` location：注释现役 `alias .../dist/;`，启用 `# alias .../dist-old/;` | 配置 diff 仅一行切换 | ~30s |
| 2.2 | `nginx -t` | `syntax is ok` | ~5s |
| 2.3 | `nginx -s reload` | 平滑 reload | ~3s |
| 2.4 | 验证：e-cam-web 触发 401（如清除共享 session 后访问 `/cam/` 受保护路由）→ 跳转目标为 ecmdb-web `/login`（旧形态），**不是** `/console/login` | 401 → `/login`（query 保留） | ~30s |
| 2.5 | 验证：`GET /cam/` 仍 200 出 e-cam-web 静态 | 200 | ~10s |

小计目标 ≤ 1.5 分钟。受 e-cam-web 双产物并存的 build 期 env 注入保障（RC#2 / 任务 2.13）。

#### ③ 过渡期旧登录入口 301 回切（启用 → 注释整 location）

| 步骤 | 动作 | 预期 | 累计耗时（目标） |
|------|------|------|------------------|
| 3.1 | 编辑：注释整 `location = /login { ... }` 块 | 该 location 不再生效 | ~30s |
| 3.2 | `nginx -t` | `syntax is ok` | ~5s |
| 3.3 | `nginx -s reload` | 平滑 reload | ~3s |
| 3.4 | 验证：`GET /login?redirect=/cam/xxx` 不再 301，落至默认 location（独立运行态 404 / 并入现役 vhost 时由 vhost 既有 location 承接） | 不再 301 → `/console/login` | ~10s |
| 3.5 | 反向验证（恢复 301）：取消注释整 location + reload → `GET /login?redirect=/cam/xxx` → `301` → `/console/login?redirect=/cam/xxx`（query 保留） | 301 + query 保留 | ~15s |

小计目标 ≤ 1 分钟。三件套正向（回切）+ 反向（恢复）双向验证，确认切换可逆。

### 2.4 回退验证清单（演练放行门）

每件回切完成后须全过：

- [ ] 切换后 nginx `-t` 通过、reload 无报错
- [ ] 回切池产物可访问（版本标记回退至上一版）
- [ ] 探活端点 `/console/health` 全程 `200 ok`（探活不因池切换中断）
- [ ] SPA 路由回退完整（`/console/login` 等 SPA 路径 200，非白屏）
- [ ] eiam API 反代不受影响（`GET /api/iam/user/profile` 仍返回同口径响应）
- [ ] 三件套切换可逆（反向恢复后行为回到现役态）
- [ ] 总耗时 T_end − T0 ≤ 10 分钟（PRD 承诺）

### 2.5 演练记录表（staging 执行后回填）

> 本表为 staging 演练的回填槽，**未演练不得上线**（Hard Rules）。本提案交付时为「待演练」态，由平台管理员在上线前 staging 执行后回填实际值。

| 字段 | 值（待回填） |
|------|-------------|
| 演练日期 | ____ |
| 环境（staging host:port） | ____ |
| 平台管理员（执行） | ____ |
| eiam owner（复核） | ____ |
| ① /console 池切换耗时 | ____ |
| ② /cam/ 401 目标切换耗时 | ____ |
| ③ 旧登录 301 切换耗时 | ____ |
| 总耗时 T_end − T0 | ____ 分钟（目标 ≤ 10） |
| 回退验证 §2.4 全过 | ☐ 是 / ☐ 否（否 = 阻塞上线） |
| 异常与处置 | ____ |
| 结论 | ☐ 放行上线 / ☐ 阻塞（须修复后重演练） |

---

## 3. CSP 收敛实测（按真实构建产物逐指令收敛）

### 3.1 基线与目标

tech-design §Security CSP 指令基线表（`tech-design.md:343-353`）：

| 指令 | 基线 | 收敛目标 | 说明 |
|------|------|----------|------|
| `default-src` | `'self'` | 保持 | 未列指令兜底 |
| `script-src` | `'self'` | 保持 | 禁 inline / eval，vite 产物外链 content-hash |
| `style-src` | `'self'` `'unsafe-inline'` | `'self'`（**已收紧**） | Element Plus 弹层运行时内联定位样式所需；实测可去除则收紧 |
| `connect-src` | `'self'` | 保持 | API 仅同源 `/api/iam` |
| `img-src` | `'self'` data: | 保持 | 头像 / 图标 |
| `frame-ancestors` / `object-src` / `base-uri` | `'none'` / `'none'` / `'self'` | 保持 | XFO 现代等价 + 防基址劫持 |

收敛焦点 = `style-src` 的 `'unsafe-inline'`（任务 2.10 落地基线时保留，挂账本任务实测）。

### 3.2 真实构建产物测量（已执行）

**测量对象**：`haven-console/dist/`（task 5.5 验收态 61/61 E2E 通过所对应的构建产物；RC#1 done per 5.4）。

**测量方法**：逐产物（index.html + 全部 JS chunk）扫描受 CSP `style-src` 辖制的两类形态——① parser-inserted `style=""` 属性（HTML 标记 / Vue 模板静态 style）② `<style>` 元素标签（含 JS 注入字面量）。CSSOM 写入（`.style.X =` / `patchStyle` / `.style.cssText =`）**不**受 CSP `style-src` 辖制，单独统计但不计入收紧阻塞项。

| 产物 | `style=""` 属性 | `<style>` 元素 | `setAttribute('style',…)` | CSSOM `.style.X` | 结论 |
|------|----------------|----------------|---------------------------|------------------|------|
| `dist/index.html` | 0 | 0 | N/A | N/A | 干净 |
| `dist/js/*.js`（全部 chunk） | 0 | 0（零 `<style` 字面量） | 0 | 3（`/.style.display`，应用层 transition） | 干净 |
| `dist/js/element-plus-*.js`（按需 tree-shake 后） | 0 | 0 | 0 | 0 | 干净 |

**关键证据**：
- `index.html`：grep `<style` = 0，grep `style="` = 0——产物入口零 parser-inserted 内联样式。
- JS chunk：grep `setAttribute\(['"]style` = 0——零程序化 style 属性写入；grep `<style` 字面量 = 0——零运行时注入 `<style>` 标签。
- Vue 3 渲染器（`vue-vendor-*.js`）`patchStyle` 经 CSSOM（`.style.prop =` / `.style.cssText =`）应用样式——**CSP `style-src` 不辖 CSSOM 写入**（规范：style-src 仅辖 `style` 内容属性与 `<style>` 元素）。`:style` 绑定、Element Plus 弹层定位（`.style.top/.left`）均走 CSSOM，CSP-exempt。
- Element Plus chunk 按需 tree-shake 后（`unplugin-vue-components` 自动引入仅含已用组件）零 style-attr prop 绑定。

### 3.3 收敛结论

**`style-src 'unsafe-inline'` 可去除，已收紧为 `style-src 'self'`。**

依据：真实构建产物在 CSP `style-src` 辖制面（parser-inserted `style` 属性 + `<style>` 元素 + `setAttribute('style')`）零命中；全部样式应用经 CSSOM，CSP-exempt。Element Plus 弹层运行时内联定位样式（tech-design 基线注记的保留理由）实为 CSSOM `.style.top/.left` 写入，不触发 CSP——原保留 `'unsafe-inline'` 属过度宽放。

**已落地**：`haven-console/nginx/haven-console.conf` §`location ^~ /console/` 的 `Content-Security-Policy` 头已改为：

```
default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; object-src 'none'; base-uri 'self'
```

### 3.4 活体验证（已执行）+ staging 放行门

**本任务活体验证**（独立 nginx 1.25.4 实例 :8899，不动现役 :8888）：

| # | 断言 | 结果 |
|---|------|------|
| 1 | `nginx -t`（并入 http 块 wrap）通过 | PASS |
| 2 | `GET /console/` 响应头 `Content-Security-Policy` 不含 `'unsafe-inline'` | PASS（`style-src 'self'`） |
| 3 | 其余指令（default/script/connect/img/frame-ancestors/object/base-uri）逐项保持基线 | PASS |
| 4 | `X-Frame-Options: DENY` / `Strict-Transport-Security` / `Cache-Control` 保持 | PASS |
| 5 | `GET /console/health` 不携带 CSP（探活端点独立 location，同 e-cam-web 口径） | PASS |

**staging 放行门（上线前必跑，本提案交付时为「待执行」）**：

静态产物测量证明 parser-inserted 面零命中，但 CSP 收敛的最终判据须在 staging 真实浏览器活体验证（DevTools Console 跨全页面零 CSP 违规报告）。门：

- [ ] staging 同域 nginx 形态（`:8888/cam/` 前例，非 vite dev server）下，Chrome DevTools Console 跨 11 页（登录 / 工作台 / 用户 / 组织 / 租户 / 身份源 / 角色 / 授权 / 策略 / 租户选择 / 主布局）零 `Refused to apply inline style` / `Content Security Policy` 违规
- [ ] 重点覆盖弹层交互：el-dialog 打开、el-select 下拉、el-popover/tip 定位、主题切换（深/浅）
- [ ] 若任一页面出现 CSP 违规 → 回退 `style-src 'self' 'unsafe-inline'`，记违规组件 + 触发点原因，回到基线值，挂账下一轮收敛（不强行上线）
- [ ] 双签：平台管理员起草 + eiam owner 复核（随认证链路兼容矩阵双签口径）

> 静态测量与活体验证互补：静态测量证明产物面干净，活体验证兜底运行时动态行为（Element Plus 弹层 CSSOM 定位、主题切换注入 CSS 变量等）。两者均过方可声明 CSP 收敛完成。

---

## 4. 双人复核登记（部署清单入卷）

PRD / tech-design 要求「回切演练记录 + CSP 收敛结论入部署清单，平台管理员 + eiam owner 双签」（AC4）。本节为双签登记槽。

| 交付项 | 状态 | 平台管理员（起草/执行） | eiam owner（复核会签） | 备注 |
|--------|------|--------------------------|------------------------|------|
| nginx `/console/health` 探活端点 | 已落地（配置面） | ____（签名/日期） | ____（签名/日期） | 探活接入告警须走同面 WatchAlert 渠道（§1.2） |
| 探活 → WatchAlert 告警规则接入 | 待 staging 落地 | ____ | ____（不涉，eiam 非告警面） | WatchAlert 运维 owner 落地探测器条目（§0.3） |
| 回切三件套 staging 演练（计时 ≤ 10 min + 回退验证） | 待 staging 演练 | ____（执行/计时） | ____（复核） | §2.5 记录表回填 |
| CSP `style-src` 收紧（移除 `'unsafe-inline'`） | 已落地（配置面） | ____（起草） | ____（复核） | §3 收敛结论 + staging 活体验证门 |
| CSP staging 活体验证（DevTools Console 零违规） | 待 staging 执行 | ____（执行） | ____（复核） | §3.4 放行门 |

> 「已落地」= 配置 / 记录件已交付；「待 staging」= 须在 staging 环境执行并回填证据后方可声明完成。本提案交付工件 + 配置，**不替代 staging 执行**（§0.2 边界）。

---

## 5. 关联文档与追溯

| 关联 | 位置 | 关系 |
|------|------|------|
| PRD Monitoring Requirements | `prd/prd-spec.md:197-200` | §1 探活 + §2 回切演练 10 分钟承诺来源 |
| PRD Security Requirements | `prd/prd-spec.md:201-205` | §3 CSP nginx 注入来源 |
| tech-design §Integration Specs 5（部署与回切开关） | `design/tech-design.md:283` | §2 三件套 + staging 演练计时入部署清单来源 |
| tech-design §Security（CSP 指令基线表） | `design/tech-design.md:343-353` | §3 CSP 收敛基线与目标来源 |
| tech-design §跨仓改动设计 RC#2 | `design/tech-design.md:306-314` | §2 ② e-cam-web 401 目标双产物来源 |
| nginx 同域入口配置（任务 2.10） | `haven-console/nginx/haven-console.conf` | 本任务修改对象（加探活 + 收紧 style-src） |
| 部署与运维约定 | `docs/conventions/deployment.md` | TECH-deployment-001（三件套）/ 002（探活告警）跨特性约定 |
| WatchAlert 接入决策 | `docs/decisions/0002-watchalert-integration.md` | §1.2 同运维面 / 同告警渠道来源 |
| e-cam-web nginx 探活基线 | `e-cam-web/nginx.conf:110` / `nginx.cam.conf:57` | §1.1 镜像形态来源 |
| 下线就绪清单（任务 5.6） | `teardown-checklist.md` | 互补工件（本 = 上线侧加固，彼 = 下线侧判据） |
| 任务 5.5 验收态 | acceptance.md 61/61 | §3.2 构建产物测量对象所对应的验收态 |

---

## 6. 变更记录

| 日期 | 任务 | 变更 | 作者 |
|------|------|------|------|
| 2026-10-08 | 5.7 | 建立 deploy-hardening runbook：§1 探活 + 监控告警接入（nginx /console/health 已落地，WatchAlert 同渠道登记）；§2 回切三件套演练 runbook（步骤/耗时/回退验证/演练记录表）；§3 CSP 收敛实测（真实构建产物逐指令测量，style-src 'unsafe-inline' 已收紧为 'self' + staging 活体验证门）；§4 双签登记槽；nginx 配置改动 = 加 /console/health + 移除 style-src 'unsafe-inline' | forge-task-executor |
