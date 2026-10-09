---
title: "安全约定"
domains: [security, csp, xss, cookie, token, csrf, login-page]
---

# 技术约定：安全

> 跨特性可复用约定。源：feature/platform-console。

## CSP

### TECH-security-001: CSP 指令基线（nginx 注入）

**Requirement**: nginx 注入 CSP 覆盖全部 `/console` 路由：`default-src 'self'`；`script-src 'self'`（禁 inline/eval）；`style-src 'self' 'unsafe-inline'`（Element Plus 弹层运行时内联定位样式所需，实测可去除则收紧并记录）；`connect-src 'self'`；`img-src 'self' data:`；`frame-ancestors 'none'`；`object-src 'none'`；`base-uri 'self'`。起草人 = 平台管理员，复核 = eiam owner，随认证链路兼容矩阵双签；Phase 0 按真实构建产物实测收敛。
**Source**: feature/platform-console tech-design §Security CSP 指令基线

## XSS / token 窃取

### TECH-security-002: token 窃取 / XSS 缓解配对

**Requirement**: 继承「从 cookie 读 token 注入 Bearer」⇒ session cookie 必须可被 JS 读取（非 httpOnly），任何 XSS 可直接窃取平台凭证。缓解配对：① CSP 基线收紧脚本与连接源；② 零第三方脚本 + 构建产物无内联脚本；③ Vue 模板默认转义，**禁用 `v-html` 渲染任何 eiam 返回内容**（文案契约、字段错误一律纯文本 / 组件化渲染）；④ cookie `Secure/SameSite/Path/Domain` 以 Phase 0 矩阵为准并经 eiam owner 会签；⑤ 会话固定 / 重放：凭据由 eiam 服务端签发与销毁（租户切换即销毁重签），前端不生成、不续期、凭据不落 localStorage。
**Source**: feature/platform-console tech-design §Security token 窃取 / XSS

## cookie

### TECH-security-003: cookie 域严格一致

**Requirement**: cookie `domain / path` 严格与 eiam、e-cam-web 一致（SSO 及共享 session 前提）；`SameSite / Secure` 以 Phase 0 认证链路兼容矩阵为准（eiam owner 书面会签）。
**Source**: feature/platform-console tech-design §Security 会话 / §Architecture 多方会话拓扑

## 登录页安全面

### TECH-security-004: 登录页安全面

**Requirement**: CSP / `X-Frame-Options` / HSTS 由 nginx 注入；CSRF 归属 eiam（前提：session cookie 按兼容矩阵携带 `SameSite=Lax/Strict`，跨站请求不携带凭据，前端无需 CSRF token 配合；前提不成立则重议）；零第三方脚本。登录页仅作 eiam 流程 UI 壳，凭据校验全在 eiam 侧。
**Source**: feature/platform-console prd-spec.md §Security Requirements；tech-design §Security 登录页安全面
