# Iteration 1 Merged Report

**PM**: 618/1000 | **QA**: 652/1000 | **Averaged**: 635/1000 (target 900)

## Merged Attack Points（去重后 14 条）

1. [Functional Specs] 宽限期时长未定义（UF-2「剩余 X 小时」、Story 2、NFR 多处出现却无具体值）— "宽限剩余时长 | 数值 | 后端状态 | 横幅展示「剩余 X 小时」" — 定义宽限期具体时长（如 2 小时），否则无法写边界测试
2. [User Stories] Story 1 只覆盖正常流程，缺错误 AC — "Then 认证成功后我进入告警看板" — 补 eiam 认证失败、eiam 不可用、会话过期的 Given/When/Then
3. [Scenario Completeness] 会话过期场景缺失（违反 BIZ-auth-008）— 流程图仅 CheckAuth 初始检查 — 补「会话使用中过期 → 跳登录 → 回原操作路径」场景
4. [Scenario Completeness] 单点登出未覆盖（违反 BIZ-auth-005）— In-Scope 只提"移除本地登录入口" — 补登出/共享会话销毁/下次请求 401 收敛行为，并说明 WatchAlert 是否加入共享 session
5. [Flow Diagrams] 迁移切流主流程缺 Mermaid 图 — "主流程 3 —— 迁移切流" 仅文字 — 补图，含阶段门控 + 回滚分支
6. [Functional Specs] 403 授权拒绝场景缺失 + 401/404 语义不一致 — 登录图 EiamOK→Session→Dash 把"认证成功"当"授权成功" — 补「认证通过但无 WatchAlert 内部角色 → 403 原地提示」分支（BIZ-auth-003）；旧账密移除端点统一为 404（401=未认证语义，不该用于已移除端点）
7. [User Stories] 迁移切流在 In-Scope 但无对应故事 — "迁移切流：并行运行…退役" — 补迁移/切流用户故事（运维/管理员视角）
8. [Edge Case] 迁移回滚步骤与验证门控缺失 — "观察一个告警周期""含回滚方式"只提未述 — 定义回滚触发/步骤/数据一致性、切流门控（"告警周期"具体时长、并行/切流/退役各阶段 pass/fail 标准）
9. [Functional Specs] eiam→WatchAlert 角色映射机制空白 — "eiam 只负责认证…授权留在 WatchAlert 内部" — 说明 eiam 用户如何获得 WatchAlert 内部角色（预置/首登/claim）
10. [Functional Specs] 回跳地址开放重定向风险 — "回跳地址 | 字符串 | 当前页 URL" 未清洗 — 补 URL 白名单/清洗校验规则
11. [Edge Case] token 多进程存储/验签机制矛盾 — "token 仅存内存，不落盘" 在多进程架构下歧义（谁的内存？无共享存储如何验签）— 明确 token 签发/存储/验证机制
12. [Edge Case] eiam 抖动（快速可用/不可用切换）无防抖/滞回 — 补防抖或滞回机制，避免横幅快速闪烁
13. [Functional Specs] 降级探针机制未定义 — "鉴权服务状态 | 布尔 | 后端探针" — 定义探针端点/轮询频率/恢复判定
14. [Background & Goals] 目标未量化 + 质量清单自证与正文矛盾 — "启动耗时 ≤ 编译 + 数十秒"、"上游贡献不中断"无指标、清单勾选"无含糊措辞"与"数十秒"矛盾 — 量化所有阈值（超时给具体秒数、「上游贡献」给可测指标），移除/修正自证勾选