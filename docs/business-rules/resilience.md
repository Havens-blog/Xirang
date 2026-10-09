---
title: "降级与韧性业务规则"
domains: [resilience, degradation, availability, eiam, passkey, browser-compat]
---

# 业务规则：降级与韧性

> 跨特性领域不变量。源：feature/platform-console。

## eiam 不可用

### BIZ-resilience-001: eiam 不可用降级

**Rule**: eiam 不可用时，管理页错误态可重试、工作台静态内容仍渲染、不白屏不无限 loading；已登录的 e-cam-web 会话不受影响。
**Context**: 控制台为纯静态站，资源加载与 eiam 可达性解耦；降级不能波及共享 session 的其他应用。
**Source**: feature/platform-console prd-spec.md E12 / §Business Flow #8；proposal §Key Scenarios

## 浏览器能力

### BIZ-resilience-002: passkey 浏览器不支持隐藏入口

**Rule**: WebAuthn 不可用时隐藏 passkey 入口，提示改用密码 / LDAP，无 JS 报错。
**Context**: 浏览器能力探测失败不应崩溃登录页。
**Source**: feature/platform-console prd-spec.md E8
