---
title: "数据模型归一约定"
domains: [data-model, normalization, status, policy, role, tenant, passkey, conn-params]
---

# 技术约定：数据模型归一

> 跨特性可复用约定。源：feature/platform-console。

## 状态归一

### TECH-data-model-001: 用户状态归一词表

**Requirement**: eiam 原值 `'active' | 'disable'`，`ParseStatus` 未知串归一 `'unknown'`；写错会静默降级、启停操作失效。
**Source**: feature/platform-console tech-design §Data Models User.status

## 角色与租户

### TECH-data-model-002: 角色分配以 code 为准

**Requirement**: eiam 角色分配以 `code` 为准（实盘：`role/batch_assign` 载荷为 `{usernames, role_codes}`）。
**Source**: feature/platform-console tech-design §Interfaces A 档 role/batch_assign

### TECH-data-model-003: tenant/switch 销毁旧 session 重签 JWT

**Requirement**: `POST /api/iam/tenant/switch`（header `X-Active-Tenant-ID`）→ eiam 销毁旧 session 并重签 JWT，调用方必须随后整页 reload。
**Source**: feature/platform-console tech-design §Interfaces A 档 tenant/switch

## 策略

### TECH-data-model-004: Policy Statement[] 归一规则

**Requirement**: 前端规范形策略 = `PolicyStatement[]`（一条策略可同时含 Allow 与 Deny）。归一规则：若实核 eiam 为「扁平单 effect」存储（顶层 effect + actions/resources），扁平结构 ≡ 恰含一条 Statement 的数组；扁平结构无法表达多 Statement 策略时，该形态上报 validation（不静默截断）。
**Source**: feature/platform-console tech-design §Data Models Policy / PolicyStatement

## 身份源与 passkey

### TECH-data-model-005: 身份源 ConnParams 密码写后不回显

**Requirement**: 身份源连接参数 `ConnParams.password` 仅写入、响应掩码不回显；编辑态占位 `'******'`。`ConnParams` 六字段：`url`（ldap:// 或 ldaps://host:port）、`port`(1–65535)、`bindDn`(必填)、`password`、`baseDn`(必填)、`attrMap.username`(必填)、`timeoutSec`(1–30，默认 5)。
**Source**: feature/platform-console tech-design §Data Models ConnParams

### TECH-data-model-006: PasskeyRegistrationSession

**Requirement**: passkey 注册引导会话 `{ ticket, expiresAt, used }`：15 分钟有效、单次使用。
**Source**: feature/platform-console tech-design §Data Models PasskeyRegistrationSession
