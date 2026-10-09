---
iteration: 0
title: "Pre-Revision (Freeform Findings)"
rubric: "N/A (all dimensions)"
---

# Pre-Revision Report (Iteration 0)

自由评审产生的 21 条结构化发现（命中率 1.0），按分级全部为结构性/事实性问题，无主观偏好、无边界项。合并为 12 个 ATTACK_POINTS 路由给 Reviser。

## ATTACK_POINTS

- **[high]** CDN 托管与「同一登录域」共享 cookie 未调和 | quote: "纯静态前端站，可 CDN 托管" 与 "控制台需与 e-cam-web 在同一登录域下" | improvement: 在 Constraints 钉死宿主形态为同域 nginx 反代路径，声明跨域 CDN 直调 eiam 不支持，补 CORS/SameSite 核对说明
- **[high]** 「天然兼容、后端零改动」是未经源码核验的断言 | quote: "与 e-cam-service 认证中间件（EcmdbAuthMiddleware 已在读该 session）天然兼容，后端零改动" | improvement: 增加「认证链路事实核查」前置任务，产出 cookie/session 读写兼容矩阵，取代「天然」断言
- **[high]** 单点登出语义缺席，成功标准只覆盖登入 | quote: "共享会话意味着双向责任" | improvement: 成功标准补充控制台↔e-cam-web 双向单点登出，纳入真实浏览器双标签页闭环
- **[medium]** 验证环境宿主口径未定义，真实浏览器验证易假通过 | quote: "若控制台在开发/验证环境的宿主…不在提案里钉死" | improvement: 在验证要求中明确 SSO 闭环须在最终同域宿主形态执行，不得在独立端口 dev server 判定通过
- **[medium]** Phase 0 只列缺口、无处置决策路径，扩 eiam 端点与零后端相抵 | quote: "而补端点意味着改 eiam" | improvement: Phase 0 产出升级为带三类处置（UI 降级/扩 eiam 端点/放弃能力）决策门的缺口报告，澄清「扩展 eiam 端点」归属
- **[medium]** platform/users 迁出无能力对齐顺序门 | quote: "平台身份管理就会出现比现状更糟的真空" | improvement: 为迁出设置「控制台对 Phase 0 清单逐项覆盖完成」前置门，迁出 PR 附引用盘点+回归+回退
- **[medium]** 「无 showstopper」用未执行的 Phase 0 佐证自己 | quote: "在 Phase 0 跑完之前给出「无 showstopper」的定论" | improvement: 将可行性表述改为「待 Phase 0 盘点确认」，不预支未执行的验证结论
- **[medium]** v1 含四登录流程但 SC 仅硬性要求密码，验证口径三层落差 | quote: "两条口径不一致" | improvement: 对齐范围与成功标准，明确四流程各自验证口径（密码全闭环、LDAP/passkey/MFA 冒烟）
- **[high]** ticket/change 排除范围却未交代归宿 | quote: "去 ecmdb-web 的目标完成之日就是工单/变更功能无声消失之时" | improvement: 在 Out of Scope 显式声明 ticket/change 为「有意退役」或「挂起另立提案」
- **[medium]** tenant_id 租户选择语义仅一句括号 | quote: "用户会「登录成功但看到错租户的数据」" | improvement: 补租户选择语义说明（选租户交互、tenant_id 存放位置、跳转上下文来源），结论源自事实核查
- **[medium]** e-cam-web 过期/401 跳转目标改造未入范围 | quote: "这是一处必然发生的 e-cam-web 前端改动，游离在范围之外" | improvement: 将「e-cam-web 401/过期跳转目标切换到控制台登录页」列为 In Scope 显式任务
- **[low]** 视觉对齐未落到可执行组件库版本/token 共享 | quote: "主题 token…以文件形式共享而非复制粘贴" | improvement: 在 Constraints 将视觉对齐具体化为「同组件库同版本 + token 文件共享」

## BORDERLINE_FINDINGS

（无）

## SKIPPED_FINDINGS

（无，自由评审未表达主观偏好）