---
title: "部署与运维约定"
domains: [deployment, nginx, rollback, ops, health-check, alerting]
---

# 技术约定：部署与运维

> 跨特性可复用约定。源：feature/platform-console。

## 回切开关

### TECH-deployment-001: 回切开关三件套

**Requirement**: 回切开关为三件套，全部无 rebuild 可执行：① nginx `/console` location 注释行切换（或 upstream 双池）+ `nginx -s reload`（秒级）；② e-cam-web 401 目标以 build 期 env 注入、新旧双产物并存，切换 = 换静态目录软链；③ 过渡期旧登录入口 301 同挂 nginx location。上线前在 staging 完整演练一次回切并计时（承诺 10 分钟内），演练记录（步骤 / 耗时 / 回退验证）入部署清单。执行人：平台管理员 + eiam owner 双人复核。
**Source**: feature/platform-console tech-design §Integration Specs 5

## 探活与告警

### TECH-deployment-002: 探活健康检查与告警接入

**Requirement**: 控制台 `/console` 静态资源探活（HTTP 200 健康检查）接入告警，与 e-cam-web 同一 nginx 运维面、同一告警渠道，不新增独立 SLA。
**Source**: feature/platform-console prd-spec.md §Monitoring Requirements；proposal §NFR 可用性
