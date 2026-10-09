---
feature: "platform-console"
created: "2026-09-29"
status: completed
---

# Feature: platform-console

<!-- Status flow: prd → design → tasks → in-progress → completed -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 独立纯前端「平台控制台」承载 登录/身份治理/权限管理/工作台，直调 eiam、SSO 互通、零新增后端 |
| User Stories | prd/prd-user-stories.md | 6 条故事覆盖 4 类角色（平台/租户管理员、普通用户、只读审计）+ 双向单点登出 |
| UI Functions | prd/prd-ui-functions.md | 11 个 UI 函数：登录、导航/应用切换、工作台、租户选择 + 身份治理四页 + 权限管理三页 |
| UI Design | ui/ui-design.md | Linear 靛紫设计系统下的 11 个组件（登录/导航/工作台/租户选择/身份四页/权限三页）+ 文案契约 |
| Prototype | ui/prototype/ | HTML 可交互原型（17 文件：登录/租户选择/工作台/身份四页/权限三页/子页/占位） |
| Tech Design | design/tech-design.md | 独立纯前端 SPA（Vue3/Vite/Element Plus/Pinia/axios）直调 eiam，零后端，拷贝锁版本 + RC#1/#2 跨仓改动 |
| Task Breakdown | tasks/index.json | 35 业务任务（Phase 0 五个 doc + M1–M4 三十个 coding），5 阶段门 + 5 阶段摘要自动生成 |
| Consolidated Specs | specs/ | 已抽取 15 业务规则 + 39 技术约定；CROSS 项已集成至 `docs/business-rules/`（auth/identity/resilience）与 `docs/conventions/`（api/error-handling/frontend-state/data-model/security/performance/testing/deployment/cross-repo-copy/tech-stack）|

## Milestones

| 阶段 | 范围 | 业务任务 |
|------|------|---------|
| Phase 0 | 端点差集盘点/认证链路核查/服务端授权核查/能力映射/acceptance 骨架 | 1.1–1.5 |
| M1 | 脚手架 + 登录/登出/会话 + RC#2（e-cam-web 401 跳转切换） | 2.1–2.13 |
| M2 | 租户分流 + 工作台 + 应用切换 | 3.1–3.4 |
| M3 | 身份治理四页（UF-4/5/6/7） | 4.1–4.6 |
| M4 | 权限管理三页（UF-8/9/10）+ RC#1 迁出收尾 | 5.1–5.7 |

## Traceability

| PRD Section | Design Section | UI Component | Placement | Tasks |
|-------------|----------------|--------------|-----------|-------|
| UI Functions > 登录页 (UF-1) | Interfaces §登录端点 (B 档) | 登录页 (ui-design §登录页) | new-page:/login | 2.9, 2.12 |
| UI Functions > 顶部导航+应用切换 (UF-2) | Architecture §主布局 | 主布局 (ui-design §主布局) | new-page:主布局 | 2.8, 3.1 |
| UI Functions > 工作台 (UF-3) | Architecture §数据流 | 工作台 (ui-design §工作台) | new-page:/workbench | 3.3, 3.4 |
| UI Functions > 租户选择 (UF-11) | Integration Specs §租户上下文 | 租户选择 (ui-design §租户选择) | new-page:/tenant-select | 3.1, 3.2, 3.4 |
| UI Functions > 用户管理 (UF-4) | Interfaces / Data Models | 用户管理 (ui-design §用户管理) | new-page:/users | 4.1, 4.2, 4.6 |
| UI Functions > 组织管理 (UF-5) | Interfaces / Data Models | 组织管理 (ui-design §组织管理) | new-page:/organizations | 4.1, 4.3, 4.6 |
| UI Functions > 租户管理 (UF-6) | Interfaces / Data Models | 租户管理 (ui-design §租户管理) | new-page:/tenants | 4.1, 4.4, 4.6 |
| UI Functions > 身份源管理 (UF-7) | Interfaces / Data Models | 身份源管理 (ui-design §身份源管理) | new-page:/identity-sources | 4.1, 4.5, 4.6 |
| UI Functions > 角色管理 (UF-8) | Interfaces / Data Models | 角色管理 (ui-design §角色管理) | new-page:/roles | 5.1, 5.5 |
| UI Functions > 授权管理 (UF-9) | Interfaces / Data Models | 授权管理 (ui-design §授权管理) | new-page:/authorizations | 5.2, 5.5 |
| UI Functions > 策略管理 (UF-10) | Interfaces / Data Models | 策略管理 (ui-design §策略管理) | new-page:/policies | 5.3, 5.5 |
| PRD > Related Change #3 (端点差集+契约基线) | Interfaces §差集盘点 | — | — | 1.1, 1.2, 1.3 |
| PRD > Related Change #1 (platform/users 迁出) | RC#1 | — | — | 1.4, 5.4 |
| PRD > Related Change #2 (401 跳转切换) | RC#2 | — | — | 2.13 |
| PRD > 验收清单约束 | Testing §E2E 映射 | — | — | 1.5, 2.12, 3.4, 4.6, 5.5 |
| PRD > In Scope #8 下线就绪清单 | 跨仓改动设计 §过渡期 | — | — | 5.6 |
| PRD > Other Notes (性能/监控/安全) | 性能预算 / Integration Specs 5 / Security | — | — | 2.10, 2.11, 5.7 |
| PRD > Goals 可控交付 (脚手架/基础设施) | Overview / Architecture §目录树 | — | — | 2.1–2.7 |