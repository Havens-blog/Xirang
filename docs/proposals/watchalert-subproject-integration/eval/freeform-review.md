# Freeform Expert Review

## Background Assessment

这份提案要解决的问题是清晰的：WatchAlert（Go 后端 + React 前端的告警平台）目前以「反代·独立服务」形态游离在 Haven workspace 的编排体系之外，源码在 `D:\WatchAlert` / `D:\WatchAlert-web` 两个独立目录，未被 dev.ps1 / Justfile / forge 质量门禁纳入，登录仍是 WatchAlert 自身的本地 JWT，与平台身份体系 eiam 割裂。提案的核心动作有两个：一是把两个 fork 直接 clone 进 workspace 成为与 eiam/e-cam-service 同级的子项目仓库，保留 `.git` 与 `mine` remote 以维持向上游提 PR 的能力；二是把鉴权委托给 eiam，移除本地 JWT 登录入口。

我对 workspace 形态前提的判断与提案一致——D:\Haven 是多仓库 workspace 而非 monorepo，subtree 的「内联历史」前提确实不成立，submodule 的 detached HEAD/递归 clone 在日常确实坑多，「clone fork 下沉为同级子项目」是当前形态下最自然的选型。这一步的前提错配（XY Detection）做得对，不需要质疑。

但提案真正的不确定性全部压在了鉴权侧，而恰恰是这一侧，提案用「复用」「不另造」「最小改动」三个词把最难的部分收了起来。编排侧（四件套）虽然也有坑，但那是 Haven 已经在 eiam/opsagent 上跑通过的模式，风险可控。我的评审因此会重心偏向鉴权委托链路——因为这是我亲历过协议收敛、知道它有多容易在「看似就绪、实则缺一个接口」的地方卡死的地方。

## Key Risks

提案在鉴权侧的核心论断是 eiam 不需要成为完整 OIDC IdP，WatchAlert 复用自身已有的 OIDC/LDAP 登录对接 eiam 即可。这个论断有一个被跳过的中间环节：如果 eiam 不直接下发 OIDC，那 WatchAlert 内置的 OIDC 登录流程指向谁？

风险：WatchAlert 的 OIDC 登录能力需要一个真正的 OIDC Provider 端点，而提案一边声明 eiam 不是完整 OIDC IdP，一边又说复用 OIDC 登录对接 eiam，协议桥接的中间环节被「复用」二字掩盖。
> "eiam 作统一认证源，WatchAlert 复用已有 OIDC/LDAP 登录对接之，不要求 eiam 成为完整 OIDC IdP" — WatchAlert 的 OIDC 登录在代码层面是去命中一个 OIDC discovery endpoint（issuer /.well-known/openid-configuration、token endpoint、userinfo endpoint）的。如果 eiam 不暴露这些，那「复用 OIDC 登录」就不成立，实际走的大概率是 LDAP 通道（WatchAlert LDAP → eiam 的 LDAP 源）或者 eiam 需要新暴露一个 OIDC 兼容层。提案把这条链路的具体协议留给了 tech-design，但它已经在 Innovation Highlights 和 Assumptions Challenged 里把结论先定了（「不另造鉴权协议」「不要求 eiam 成为完整 OIDC IdP」）。先定结论再留协议设计空间，意味着 tech-design 阶段一旦发现 eiam 必须扩 OIDC 兼容层，整个「最小改动」叙事就要回退。这是我踩过的坑：协议收敛的真正成本不在写代码，而在确认 IdP 到底能下发什么。

这条风险引出一个更深的边界问题。提案在 Dependency Readiness 里自己问出了正确的问题，却没有给出答案。

问题：eiam 是否需要为下游应用新扩一个验证接口，这件事提案同时归入「可能需要的最小改动」和「超出最小必要的另立 feature」，scope 边界自相矛盾。
> "是否有可复用的登录校验端点、是否需扩一个给下游应用用的验证接口" 与 "eiam 本体功能增强（若对接需要 eiam 改动，超出最小必要的另立 feature）" — 同一份提案里，前一句把「扩下游验证接口」列为待确认的依赖项（暗示可能在本提案内解决），后一句把「eiam 改动」整体划出 scope。如果 tech-design 阶段确认 eiam 必须新增一个 token 校验或 userinfo 端点给 WatchAlert 用，那这到底是本提案的最小改动，还是另立 feature？提案没有画这条线。后果是：scope 在最关键的依赖上是不确定的，评审者无法判断本提案是否自洽可交付，实施者会在 tech-design 阶段陷入「这算不算 scope 蔓延」的拉锯。

接下来是本地 JWT 移除的安全残留。提案在 Non-Functional 里立了正确的 flag（不得残留明文凭据/绕过洞），但 Success Criteria 用了一个暧昧的词。

风险：Success Criteria 用「移除/隐藏」描述本地 JWT 登录入口，但「隐藏」不等于「移除」，隐藏的端点仍可被直接 API 调用绕过 eiam 鉴权。
> "用 eiam 账号登录 WatchAlert 成功（SSO），本地 JWT 登录入口移除/隐藏" — 「移除/隐藏」是两种强度截然不同的处置。如果只是在前端隐藏登录入口而后端 `/api/auth/login`（或等价端点）仍然接受用户名密码并签发本地 JWT，那任何知道端点的人都能绕过 eiam 直接拿 token，SSO 形同虚设。提案自己在 Non-Functional 里已经意识到「不得残留明文凭据/绕过本地登录的洞」，但验收标准却没有要求「移除」而是放行了「隐藏」。这意味着验收可能通过一个实际留洞的实现。更具体地说，WatchAlert 的本地 JWT 通常依赖一个配置里的签名密钥（secret），如果只是隐藏入口而密钥、用户表、登录 handler 仍在，那就是一个待激活的后门。

与本地 JWT 移除直接绑定的，是 eiam 不可用时的降级。提案把这条列进了「边界/失败场景」，然后就没了。

风险：eiam 不可用时的降级策略完全未定义，而 WatchAlert 是承载真实告警的运维关键系统，鉴权单点化会引入可用性回退。
> "鉴权服务(eiam)不可用时的降级" — 这句话出现在 Key Scenarios 的边界场景列表里，是整份提案中唯一提及降级的地方，没有任何展开。现状下 WatchAlert 用自己的本地 JWT，eiam 挂了不影响告警平台访问。迁入后如果 eiam 是唯一登录通道，eiam 宕机就意味着告警平台无法登录——对一个「已承载真实告警（CDN/ECS 指标、告警规则、故障中心）」的系统，这是可用性回退。提案在 Urgency 里强调 WatchAlert 越来越重要，却没把这个重要性传导到降级设计上。至少需要回答：eiam 不可用时是否能用本地 JWT 兜底（与「移除本地 JWT」直接冲突）、是否能用缓存 token、是否接受只读模式。这不是配置问题，是架构取舍。

转到编排侧。cmd/ 结构的编译坑，提案把它当作一个已有 mitigation 的已知风险，但 mitigation 的前提是 dev.ps1 服务表已经支持自定义 build 命令字段。

问题：提案把「服务表增加自定义 build 命令字段」当作 dev.ps1 已有能力来用，但没有说明这个字段是已存在的还是需要新开发的，likelihood 标 H 却写成已解决的 mitigation。
> "dev.ps1 Daemon 的 `go build -o $exe .` 约定不符，需服务表支持自定义 build 命令" 与 Risk 表 mitigation "服务表增加自定义 build 命令字段（0002 已预见）" — 「0002 已预见」不等于「dev.ps1 已实现」。如果 dev.ps1 的服务注册表当前没有 build 命令字段（只有 exe 名、启动参数、端口），那本提案隐含了一个 dev.ps1 的功能开发任务：给 Daemon 的服务表加 build 字段、改 compile 流程读这个字段。这件事的工作量和回归面（eiam/e-cam-service 的编译会不会被影响）没有被评估。提案把一个 H 概率的编排改造写成了一个括号注，这会让人低估它的实施成本。

最后是「死 fork」纪律与 eiam 鉴权改动的内在矛盾。这是提案逻辑里我认为最值得当面戳破的一处。

问题：提案要求「鉴权相关改动按通用功能开发、同步提 PR 上游」，但 eiam 是 Haven 平台内部 IAM，调用 eiam 认证的代码本质上是平台专属的，无法作为通用功能提回上游 WatchAlert fork。
> "可能需要 WatchAlert 少量通用改动（调用 eiam 认证），需遵守「通用功能可提 PR」纪律" 与 "鉴权相关改动按「通用功能」开发，同步提 PR 上游" — 这两句话同时出现在 Assumptions Challenged 和 Risk 表里。问题是：eiam 不是开源世界的通用 IdP，它是 Haven 内部服务。WatchAlert 调用 eiam 的代码（不管是 HTTP 校验还是 gRPC）绑定的是 Haven 内部端点，这种代码提回上游 WatchAlert fork 没有任何意义——上游不会接受一个调用某个私有 IAM 的 PR。这意味着提案设想的「鉴权改动走通用 PR 纪律」在这里是落空的：要么鉴权改动留在本地成为死 fork 的一部分，要么得把 eiam 对接抽象成一个可配置的通用 OIDC/LDAP 适配层（这又回到第一个风险——eiam 要不要暴露 OIDC）。提案在 Assumptions Challenged 用 5 Whys 得出「需要少量通用改动」的结论，但「通用」二字在这里是未经检验的假设。后果是：死 fork 的风险恰恰在鉴权这个改动量最大、最需要提上游的点上最大化了。

另外有一条编排语义值得提一句，虽不致命但与告警平台的安全语义相关。

风险：提案沿用「编译失败旧进程继续跑」的 Daemon 语义，但对告警平台来说，静默跑旧代码意味着你以为部署了规则修复但其实没有。
> "编译失败旧进程继续跑（沿用现有 Daemon 语义）" — 对 e-cam-service 这类业务服务，旧进程继续跑只是功能滞后。对 WatchAlert 这种告警引擎，如果你改了告警规则判定逻辑、编译失败、旧进程继续跑，你会得到一个「看起来规则生效了但实际还在用旧判定」的假状态。提案没有区分这两类子项目对编译失败的容忍度。这不是要求改 Daemon 语义，而是要求 WatchAlert 的服务表登记至少在编译失败时有一个明显的 status 标记或告警，而不是和其它服务一样静默续跑。

## Improvement Suggestions

基于上述风险，我按影响优先级提出以下建议。核心思路是：鉴权侧的协议不确定性是本提案能否自洽交付的唯一关键路径，必须在本提案阶段（而非推到 tech-design）就把边界收敛清楚；编排侧的坑可控，但需要把假设变成已验证事实。

建议：在本提案内补一个「鉴权对接协议收敛矩阵」，把 WatchAlert 的三种登录通道（本地 JWT / LDAP / OIDC）与 eiam 的三种身份源能力（系统登录 / LDAP / OIDC）逐一配对，标出每一对的协议可行性与 eiam 侧所需改动。
Addresses: 风险（OIDC 复用掩盖协议桥接缺口）与 问题（eiam 验证接口 scope 边界矛盾）
> What changes: 提案目前用「复用」「不另造」这类定性结论把协议留空。补一个配对矩阵后，能直接看出哪条链路是零改动（如 WatchAlert LDAP → eiam LDAP 源直连）、哪条链路要求 eiam 新增端点（如 WatchAlert OIDC → eiam 需暴露 OIDC 兼容层）。把要求 eiam 新增端点的那几条显式标为本提案的 scope 或明确划出 scope，就能消解当前「最小改动」与「另立 feature」的边界矛盾。提案在这之后才能真正声明 scope 自洽。

建议：把 Success Criteria 里「本地 JWT 登录入口移除/隐藏」收敛为「移除后端登录 handler 与签名密钥配置，前端入口删除」，明确不留兜底端点。
Addresses: 风险（移除/隐藏暧昧导致验收留洞）
> What changes: 验收标准从「移除/隐藏」改成强制的「移除」。同时新增一条验收项：用旧的用户名密码直接 POST WatchAlert 登录端点必须返回 401/404，而不能签发 token。这样实施者无法用「前端隐藏」糊弄验收。如果确实需要本地 JWT 作为 eiam 宕机时的降级通道（见下条），那必须显式定义为受控的降级路径，而不是默认留着的后门。

建议：明确 eiam 不可用时的降级策略，至少在提案层面声明一个取舍方向（本地 JWT 兜底 / token 缓存宽限 / 只读免登录），不要把降级推到 tech-design。
Addresses: 风险（eiam 不可用降级未定义导致告警平台可用性回退）
> What changes: 提案在 Key Scenarios 列了降级场景就停了。建议在 Feasibility 或 Key Risks 里直接声明一个降级方向。我的倾向是：对告警平台这种关键系统，保留一个受 eiam 委托签发的短时效本地 token 缓存（eiam 正常时由 eiam 签发并周期刷新，eiam 宕机时允许在宽限期内用缓存 token 登录），而不是要么完全依赖 eiam 要么留一个不受控的本地 JWT。这与上一条「移除本地 JWT」不冲突——关键区别是这个缓存 token 的签发权仍归 eiam，不是 WatchAlert 自主签发。

建议：把 dev.ps1「服务表自定义 build 命令字段」从 mitigation 文字升级为一个显式的前置任务项，并标注该字段当前是否已存在；若不存在，把它列为 Next Steps 里 PRD 之前要验证的 spike。
Addresses: 问题（cmd/ 编译 mitigation 假设 dev.ps1 已支持自定义 build）
> What changes: 提案现在把这件事写成「0002 已预见」的括号注。改成显式任务后，实施者会在 PRD 之前先打开 dev.ps1 看服务表结构，确认 build 字段是否存在、eiam/e-cam-service 的编译流程是否会被波及。如果字段不存在，就把「dev.ps1 Daemon 支持 per-service build 命令」列为本提案的一个子交付物，给它独立的工作量评估和回归测试范围，而不是淹没在「四件套登记」里。

建议：承认 eiam 鉴权对接代码是平台专属的、不提上游，并据此重新定义「死 fork」纪律的适用边界——把改动分成「可提上游的通用改动」与「平台专属适配层」两类，后者通过在 WatchAlert 内预留的可插拔鉴权接口（而非硬编码 eiam 调用）来隔离。
Addresses: 问题（eiam 专属代码无法作为通用功能提 PR 的矛盾）
> What changes: 提案现在的叙事是「鉴权改动按通用功能开发、提 PR 上游」，这对 eiam 对接是不成立的。改成：WatchAlert 侧的改动拆成两层——一层是把鉴权做成可插拔接口（auth provider 抽象），这是通用的、可提上游 PR；另一层是实现一个 eiam provider 适配器，这是平台专属的、留在 fork 本地不提上游。这样死 fork 的面被压缩到一个 adapter 文件，而不是散落在 WatchAlert 鉴权逻辑各处。提案的 Assumptions Challenged 里「通用功能可提 PR」的结论需要相应 refine：不是所有鉴权改动都通用，只有接口抽象那部分是。

建议：对 WatchAlert 的 Daemon 登记，在编译失败时与普通业务服务区分处理——至少在 `dev.ps1 status` 里对编译失败的 watch-alert 给出显眼标记，而不是静默续跑。
Addresses: 风险（告警平台静默跑旧代码导致规则修复假生效）
> What changes: 这不需要改 Daemon 的核心语义（旧进程继续跑这个默认行为对其它服务保留），只需在 watch-alert 的服务表登记里加一个「编译失败告警标记」字段，或者在 Justfile 的 `watch-alert-compile` recipe 里加一个失败后输出醒目提示。提案的 Scope 里 dev.ps1 服务表本来就要改，顺带把这个语义加上成本极低，但能避免告警平台「规则改了却没生效」的运维事故。
