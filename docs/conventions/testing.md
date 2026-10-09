---
title: "测试约定"
domains: [testing, vitest, playwright, e2e, coverage, happy-dom, nginx-host]
---

# 技术约定：测试

> 跨特性可复用约定。源：feature/platform-console。

## 单元测试

### TECH-testing-001: 单元测试 Vitest + happy-dom

**Requirement**: Vitest ^4.1.5（DOM 环境具名 happy-dom ^20.11.6，与 e-cam-web devDeps 同版本）；覆盖 ≥ 80%；`pnpm typecheck`（vue-tsc）+ `pnpm build` 通过。选 happy-dom 而非 jsdom：theme store 的 `localStorage` 持久化断言与 cookie utils 需真实 DOM 存储实现，且与现仓测试环境一致、拷贝来的 utils 测试可直接复跑。
**Source**: feature/platform-console tech-design §Testing

## 组件测试

### TECH-testing-002: 组件测试范围收敛

**Requirement**: @vue/test-utils 范围收敛到三个契约敏感组件——登录表单（校验规则 / 提交 / 锁定倒计时态）、文案契约渲染组件（code → 文案逐字断言，防二次映射回归）、StatusBadge（状态归一词表渲染）。页面级交互不写组件测试、统一由 E2E 承接，避免同一路径双重维护。
**Source**: feature/platform-console tech-design §Testing 组件层

## E2E

### TECH-testing-003: Playwright E2E 硬性

**Requirement**: Playwright ^1.55（真实浏览器 E2E；浏览器矩阵仅 chromium，与内网统一浏览器一致）。覆盖登录 / 登出、SSO 跳转、双向单点登出、身份 / 权限 CRUD 的正向 + 负向 + 边界。验收场景映射 `acceptance.md` 的 E1–E12 与 UF-1~11（每条至少 1 个映射，负向 + 边界 ≥ 1/3），逐条附 Playwright / MCP 执行证据。
**Source**: feature/platform-console tech-design §Testing 浏览器 E2E

## 验收宿主

### TECH-testing-004: 验收宿主必同域 nginx 反代

**Requirement**: 验收宿主必为最终同域 nginx 反代形态（沿用 `:8888/cam` 前例），**不得**在独立端口 dev server 上判定通过（vite 直连缺 `/api/iam` 代理 + 无共享 cookie）。
**Source**: feature/platform-console tech-design §Testing 验收宿主
