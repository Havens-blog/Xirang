# PRD Eval — Iteration 2 Merged Report

**Averaged Score**: 885/1000 (PM 900 / QA 870)
**Target**: 900

## Merged Attack Points (deduplicated, 13)

1. [Flow] eiam 不可用分支 + 租户分流（0/1/多租户）分支未入图；登出经中间节点不可直连 — 补图与图文一致
2. [Functional] 审计角色可见性矛盾：导航注记「可看 2–8」 vs UF-6/UF-10「仅平台管理员可见」 — 统一审计角色可见范围
3. [Functional] 四管理页（UF-6/8/9/10）States 缺错误态、与 E12 脱节；UF-3「加载足」笔误；UF-2 应用清单以「等」收尾不可枚举 — 补错误+重试态、修笔误、穷举应用清单
4. [Functional] UF-9 授权三元组的资源/动作词表来源 + 唯一性规则未定义 — 定义资源/动作词表与三元组唯一性
5. [Functional] passkey 账号注册引导主体/执行方式/设备丢失重绑路径未定义 — 定义注册引导与重绑
6. [Functional] 防枚举「不泄露用户名」与锁定倒计时文案互相牵制 — 界定非泄露规则仅限凭据错误提示，或锁定改中性文案
7. [UserStory] S1/S2 负向与生命周期 AC 缺失（登录失败/MFA 锁定/passkey 不支持）— 补失败/锁定 AC 或显式声明由 acceptance.md 的 E1–E12 映射承接
8. [Scenario] 自助改密与 passkey 设备丢失归属未声明 — 显式声明归 eiam 承接或列 Out of Scope
9. [Scenario] E3 零租户规则未豁免平台管理员（与 Who 职责冲突）；DF002 以事实口吻断言 tenant_id 在 JWT claims 未标待核查 — 补豁免 + 标注待 Phase 0 核查
10. [blindspot] 迁移门禁只测端点 parity 不测被迁页面功能 parity — Related Change 1 前置门补 platform/users 现有能力清单 + UF-4 逐项映射与缺口处置
11. [blindspot] eiam API 契约无冻结/变更控制条款 — 声明契约基线（parity-checklist 快照）+ 端点变更书面通知与影响评估
12. [blindspot] 里程碑 M1–M4 悬空引用（三份文档未定义）— 定义四里程碑内容或删除该引用
13. [blindspot] acceptance.md「≥24 条」只约束总数可凑数，且无产出阶段/owner/完成形态 — 增加映射矩阵约束（每条 E 与每个 UF 至少 1 条、负向占比声明）并指定产出阶段/owner/形态