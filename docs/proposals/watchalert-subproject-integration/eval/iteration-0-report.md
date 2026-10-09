---
iteration: 0
title: "Pre-Revision (Freeform Findings)"
---

# Pre-Revision Report (Freeform Findings → Reviser)

## ATTACK_POINTS

- **[high]** 本地 JWT 登录入口「移除/隐藏」语义暧昧，验收可能放行仅隐藏前端的留洞实现 | quote: "Success Criteria 用「移除/隐藏」描述本地 JWT 登录入口，但「隐藏」不等于「移除」，隐藏的端点仍可被直接 API 调用绕过 eiam 鉴权" | improvement: 把 Success Criteria 从「移除/隐藏」收敛为「移除后端登录 handler 与签名密钥配置、删除前端入口」，并新增一条可测验收项「用旧用户名密码 POST 登录端点须返回 401/404，不得签发 token」。
- **[medium]** OIDC 协议桥接中间环节被「复用」掩盖：eiam 若不暴露 OIDC 端点，「复用 OIDC 登录」不成立 | quote: "提案一边声明 eiam 不是完整 OIDC IdP，一边又说复用 OIDC 登录对接 eiam，协议桥接的中间环节被「复用」二字掩盖" | improvement: 明确鉴权通道选择（LDAP 直连 vs eiam 暴露 OIDC 兼容层），把「不另造鉴权协议 / 不要求 eiam 成为完整 OIDC IdP」从已定结论改为待 tech-design 验证的假设，并新增一条 tech-design 交付物「鉴权对接协议收敛矩阵」。
- **[medium]** eiam 验证接口 scope 边界自相矛盾（Dependency Readiness 暗示可在本提案内，Out of Scope 又将 eiam 改动整体划出） | quote: "同一份提案里，前一句把「扩下游验证接口」列为待确认的依赖项（暗示可能在本提案内解决），后一句把「eiam 改动」整体划出 scope" | improvement: 在 Scope 里显式画一条判定线——「eiam 为下游鉴权所需的最小改动归入本提案，功能性/能力增强另立 feature」，消除歧义。
- **[medium]** eiam 不可用时的降级策略未定义，鉴权单点化引入告警平台可用性回退 | quote: "鉴权服务(eiam)不可用时的降级…是整份提案中唯一提及降级的地方，没有任何展开" | improvement: 在提案层面声明一个降级方向（受控本地 JWT 兜底 / token 缓存宽限 / 只读免登录 三选一），并把决定写进 Risk 的 mitigation，不要推到 tech-design 留白。
- **[medium]** dev.ps1「自定义 build 命令字段」被当作已有能力，实为新开发前置任务（likelihood=H 却写成已解决） | quote: "「0002 已预见」不等于「dev.ps1 已实现」…提案把一个 H 概率的编排改造写成了一个括号注" | improvement: 明确该字段在 dev.ps1 当前版本不存在、需新增为显式前置任务，并在 Next Steps 列为一个 PRD 前要验证的 spike（含对 eiam/e-cam-service 编译回归面的评估）。
- **[medium]** eiam 专属鉴权代码无法作为通用功能提 PR 上游，死 fork 风险在改动量最大的鉴权点被最大化 | quote: "eiam 不是开源世界的通用 IdP，它是 Haven 内部服务。WatchAlert 调用 eiam 的代码…绑定的是 Haven 内部端点" | improvement: 把改动显式分为「可提上游的通用改动」与「平台专属适配层」两类，声明后者通过 WatchAlert 内可插拔鉴权接口隔离（不硬编码 eiam 调用），重新定义死 fork 纪律的适用边界。
- **[medium]** Daemon「编译失败旧进程继续跑」对告警平台制造「规则修复假生效」的运维假状态 | quote: "如果你改了告警规则判定逻辑、编译失败、旧进程继续跑，你会得到一个「看起来规则生效了但实际还在用旧判定」的假状态" | improvement: 在 Scope 或 Risk 增加「watch-alert 编译失败时须在 dev.ps1 status 给出显眼标记、区别于普通业务服务静默续跑」的要求。

## BORDERLINE_FINDINGS

（无）

## SKIPPED_FINDINGS

（无 —— 全部 7 项均属可执行修正，6 条改进建议已并入对应攻击点）

## Classification Audit

| 层 | 数量 | 明细 |
|----|------|------|
| Factual correction（可验证缺陷，直接改） | 3 | 「移除/隐藏」留洞、eiam scope 矛盾、dev.ps1 字段不存在 |
| Structural/architectural（内在不一致，改） | 4 | OIDC 桥接、降级未定义、死 fork 矛盾、Daemon 假生效 |
| Subjective（不可执行，跳过） | 0 | — |

extraction hit-rate: 13/13 = 1.0（无低命中率告警）

## rubric

（所有维度在本轮不可参与 —— 本轮为自由评审发现的 pre-revision）