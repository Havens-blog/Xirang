---
title: "身份数据完整性业务规则"
domains: [identity, user, tenant, username, validation, data-integrity, concurrency]
---

# 业务规则：身份数据完整性

> 跨特性领域不变量。源：feature/platform-console。

## 标识与格式

### BIZ-identity-001: 用户名格式约束

**Rule**: 用户名 3–32 字符，正则 `^[a-z][a-z0-9_-]{2,31}$`；前端即时拦截，唯一性由 eiam 校验。
**Context**: 用户名是跨系统稳定标识，格式约束保证可读性与一致引用。
**Source**: feature/platform-console prd-spec.md E10

## 租户归属

### BIZ-identity-002: 用户-租户为 1:N（禁止 M:N）

**Rule**: eiam 禁止用户-租户多对多关系；`tenant/batch_assign` 载荷为 `{user_ids, tenant_ids}` 但语义受 1:N 约束。
**Context**: 租户是数据隔离边界，M:N 会破坏隔离；与租户边界越权判定同源。
**Source**: feature/platform-console tech-design §Interfaces A 档 tenant/batch_assign

## 删除与并发

### BIZ-identity-003: 依赖删除被拒

**Rule**: 删除被依赖对象（下属组织 / 关联用户 / 用户绑定）被拒：弹窗列出阻断计数，确认按钮置灰。
**Context**: 防止孤儿数据与悬空引用；删除前置依赖检查由 eiam 服务端强制。
**Source**: feature/platform-console prd-spec.md E11

### BIZ-identity-004: 并发编辑 last-submitter-wins

**Rule**: 同一对象并发编辑：后提交者生效；若 eiam 返回 409 版本冲突则提示「数据已被他人修改，请刷新后重试」并自动重拉当前页（Phase 0 核查 eiam 是否支持 409，不支持则降级并入 validation）。
**Context**: 无乐观锁时的显式取舍；409 是 eiam 支持时的最优解。
**Source**: feature/platform-console prd-spec.md E6；tech-design §Error Handling conflict

## 状态词表

### BIZ-identity-005: 用户状态词表仅 active / disable

**Rule**: 用户状态原值为 `active` | `disable`；未知串归一 `unknown`（写错会静默降级、启停操作失效）。
**Context**: 状态词表是启停操作的契约边界，写错值不报错但功能失效是高危静默失败。
**Source**: feature/platform-console tech-design §Data Models User.status
