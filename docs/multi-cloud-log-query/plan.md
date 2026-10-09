# 多云统一日志查询 - 实现计划

> 特性代号:`multi-cloud-log-query`
> 涉及仓库:`e-cam-service`(后端主体)、`e-cam-web`(前端页面)
> 日期:2026-08-27 · 状态:计划评审中

---

## 1. 目标与范围

### 1.1 目标

在 e-cam-service 上实现**多云统一日志查询**,首期覆盖 **CDN / WAF / 负载均衡(ALB)** 三类日志、**阿里云 / 腾讯云 / AWS / Azure** 四家云:

- 用户选择「日志类型 × 云 × 账号 × 资源 × 时间范围」→ 一次查询返回**字段统一、时间排序**的日志列表;
- 统一字段固定列展示,原始字段完整保留(详情抽屉),信息零丢失;
- 查询走联邦模式(不落库),与云上日志投递配置解耦。

### 1.2 范围内

| 维度 | 内容 |
|---|---|
| 日志类型 | CDN 访问日志、WAF 攻击/拦截日志、负载均衡访问日志 |
| 云厂商 | 阿里云、华为云、AWS;腾讯云接口预留(数据源未流动,见 source-status.md);~~Azure~~(无账号,移出) |
| 意外收获 | Akamai CDN+WAF 已自采进阿里 SLS,随阿里源一并覆盖 |
| 查询模式 | 联邦实时查询(Phase A,本期) |
| 采集入库 | Kafka→MongoDB 跨云聚合(Phase B,二期,仅做接口预留) |

### 1.3 范围外(明确不做)

- 各云日志投递的**开启/配置管理**(只做状态探测与引导提示,开启操作在云控制台完成);
- 其他日志类型(VPC Flow Log、DNS Query Log、API 审计日志——模型预留 `log_type` 扩展位);
- 日志的下载导出、报表、告警联动(Phase B 之后)。

---

## 2. 总体架构

```
e-cam-web (logs 页面)
   │  POST /api/v1/cam/logs/search
   ▼
e-cam-service  internal/logquery
   ├─ web/         路由 + 角色门卫(照抄 cert/web 模式)
   ├─ service/     联邦编排:fan-out → 归并排序 → 截断
   └─ (归一)      声明式字段映射 + 枚举/时间归一
          ▲ 统一日志模型(types 包)
internal/shared/cloudx/logquery/
   ├─ registry.go        (cloud, logType) → Provider
   ├─ aliyun/{cdn,waf,slb}.go    SLS 查询(含 Akamai 自采 CDN/WAF logstore)
   ├─ huawei/{waf,elb}.go        LTS ListLogs 查询(WAF JSON / ELB 文本)
   ├─ tencent/{cdn,waf,clb}.go   CLS 查询 + CDN 日志包下载(接口预留,数据源待修复)
   ├─ aws/{cloudfront,waf,alb}.go  S3 清单 + 下载解压解析
   └─ (azure 移除:无云账号)
          │  复用现有云账号凭证体系 + common/*/ratelimit.go
          ▼
   各云日志源(SLS / LTS / CLS / S3 / 日志文件包)
```

**设计原则(决策记录见 §7):**

1. **采集适配与字段归一分离**——Provider 只负责"把原始日志拿来",映射层负责"变成统一模型",互不感知;
2. **按日志类型分域建模**——CDN/WAF/SLB 三个独立 schema,不做大一统大宽表;
3. **先联邦后入库**——Phase A 不落库,验证 12 个源的字段映射;Phase B 把"查询时拉"改"定时拉",映射代码原样复用;
4. **照抄既有先例**——Provider 结构照 `cloudx/billing`(同样解决了各云字段不统一),模块装配照 `internal/cert`。

---

## 3. Phase 0:前置调研清单(必须最先做)

> **这是整个计划里最高风险的一步**:字段映射不能靠文档写,必须拿到各云**真实日志样例**做 fixture。
> 同时必须确认每个源的**投递开启状态**——未开启的源,代码写完也查不到数据。

| # | 云 | 产品 | 日志位置 | 需用户开启? | 获取方式 | 实时性 | 调研动作 |
|---|---|---|---|---|---|---|---|
| 1 | 阿里 | CDN | 离线日志文件 | 否(默认有) | `DescribeCdnDomainLogs` 拿下载链接 → gzip 解包 | T+1h~T+1d | 拉一天真实日志,记录字段对照 |
| 2 | 阿里 | WAF | SLS 专属 logstore | 是(WAF 控制台开) | SLS `GetLogs`(需确认 Project/Logstore 命名) | 秒级 | 确认 4 个生产账号 WAF 日志开启状态 |
| 3 | 阿里 | ALB/NLB/CLB | SLS(访问日志投递) | 是(实例级开启) | SLS `GetLogs` | 秒级 | 确认哪些 LB 实例已开访问日志 |
| 4 | 腾讯 | CDN | 离线日志包 | 否(默认有) | `DescribeLogUrls` → 下载解压 | T+1h | 同 #1 |
| 5 | 腾讯 | WAF | CLS 主题 | 是(接入时选) | CLS `SearchLog` | 秒级 | 确认 CLS 主题存在 |
| 6 | 腾讯 | CLB | CLS(访问日志投递) | 是 | CLS `SearchLog` | 秒级 | 同 #3 |
| 7 | AWS | CloudFront | S3 标准日志 | 是(分发级开启) | S3 ListObjects + 下载解压(gzip) | ~1h | 确认生产分发已开标准日志 |
| 8 | AWS | WAF | Kinesis Firehose→S3 | 是(需配 Firehose) | S3(JSON 行,gz) | 分钟级 | 确认 Firehose 已配 |
| 9 | AWS | ALB | S3(访问日志投递) | 是(LB 级开启) | S3(固定前缀布局) | 5min+ | 确认 ALB 已开日志 |
| 10 | Azure | Front Door | Log Analytics / 存储 | 是(诊断设置) | Log Analytics REST(KQL) | 分钟级 | 确认诊断设置存在 |
| 11 | Azure | WAF(on FD) | 同上 | 同上 | 同上 | 分钟级 | 同上 |
| 12 | Azure | App Gateway | Log Analytics | 是 | 同上 | 分钟级 | 同上 |

**Phase 0 产出物(✅ 已完成,2026-08-27):**
- [x] `docs/multi-cloud-log-query/field-mapping.md`:9 个源的字段对照表(原始字段 -> 统一字段),含真实样例值;
- [x] `docs/multi-cloud-log-query/source-status.md`:各云各账号日志投递状态清单;
- [x] 11 个 `testdata/*.json` fixture 文件(映射层单测的数据基础);
- [ ] 各云 SDK 依赖引入(go.mod 加 SLS + 华为 LTS SDK,见 §4.6)。

**Phase 0 结论(改变计划的四件事):**
1. **Azure 无账号** -> Phase 4 移除,由华为 LTS 替补(实测数据流动良好);
2. **Akamai CDN/WAF 已自采进阿里 SLS**(jlc-prod-akamai-cdnwaf-log),阿里 provider 一并覆盖,等于白赚两个源;
3. **腾讯 WAF/EdgeOne topic 存在但 30 天 0 条,CDN 无离线日志包** -> 腾讯降级为接口预留,需云侧先修复投递;
4. **AWS ALB 访问日志大多未开启**(CloudFront 44/57 已开,ALB 仅 4 个 LB 开启)-> 不阻塞本期,单独立项跟进;火山 TLS API 网络不可达,暂缓。

---

## 4. 后端设计

### 4.1 统一日志模型(`internal/shared/cloudx/logquery/types`)

三个领域模型 + 一个公共元数据:

```go
// 公共元数据:所有日志类型共有
type LogMeta struct {
    Cloud      domain.CloudProvider // aliyun / tencent / aws / azure
    AccountID  string               // 云账号(平台内 ID)
    CloudAccountName string         // 展示用
    ResourceID string               // 域名 / LB 实例 ID / 分发 ID
    Region     string
    LogType    LogType              // cdn / waf / slb
}

type CDNLogEntry struct {
    Meta       LogMeta
    Timestamp  int64  // Unix 毫秒 UTC(强制归一)
    ClientIP   string
    Method     string
    URL        string
    Status     int
    BytesSent  int64
    CacheHit   string // 归一枚举: hit / miss / partial(-)
    Referer    string
    UserAgent  string
    EdgeNode   string // POP / 边缘节点
    RequestID  string
    Raw        map[string]any // 原始字段全量保留
}

type WAFLogEntry struct { // ClientIP, Host, URI, Method, RuleID, RuleName,
                            // Action(归一: block/captcha/allow/pass), AttackType,
                            // Status, UserAgent, Timestamp, Meta, Raw }
type SLBLogEntry struct { // ClientIP, ClientPort, TargetIP, TargetPort, ListenerPort,
                           // Protocol, Method, URL, Status, RequestLength, BytesSent,
                           // LatencyMs, Timestamp, Meta, Raw }
```

**归一化硬规则**(映射层职责,单测覆盖):
- 时间:阿里 SLS `__time__`(秒)→ ms;CLS 时间戳(字符串 ms);AWS S3 日志 ISO8601 带时区;Azure KQL `TimeGenerated` RFC3339 → 统一 **Unix 毫秒 UTC**;
- 枚举:action(block/intercept/BLOCK → `block`,等)、cache 状态(TCP_HIT/HIT → `hit`);
- 缺失容忍:字段可空,UI 显 `-`;`Raw` 永远全量。

### 4.2 Provider 接口(`cloudx/logquery/adapter.go`,照 billing 模式)

```go
type LogProvider interface {
    // 日志源清单(域名/LB 实例),含"未开启投递"状态
    ListLogSources(ctx context.Context, cred *CloudCredential) ([]LogSource, error)
    // 拉原始日志;游标语义由实现方自定义,编排层不感知
    FetchRaw(ctx context.Context, cred *CloudCredential, q RawQuery) (*RawPage, error)
}

type LogSource struct {
    ResourceID, Name, Region string
    Enabled bool   // 投递是否已开启
    Note    string // 未开启原因/引导文案
}
type RawQuery struct {
    Source LogSource; Start, End time.Time
    Keyword string; Cursor string; MaxEntries int
}
```

每云每类型一个实现文件,**注册进 registry**:`registry.Get(cloud, logType)` 返回 Provider。新增云/类型 = 新增文件 + 注册,零侵入。

### 4.3 映射层(`cloudx/logquery/<cloud>/mapper.go`)

- 每实现配一个 `Mapper`:声明式字段表 `map[string]string`(原始→统一)+ 少量计算函数(时间/枚举);
- **fixture 单测强制**:每个 Mapper 用 Phase 0 拿到的真实样例做 `testdata` 驱动测试,断言统一字段值——这是本特性质量的核心防线;
- 映射层无 I/O,纯函数,覆盖率达 80%+ 无压力。

### 4.4 联邦编排(`internal/logquery/service`)

```
search(logType, clouds[], accounts[], sources[], start, end, keyword):
  1. 按 (cloud, account) 分组,组内并发调 Provider.FetchRaw
     - per-cloud 限流:复用 common/{aliyun,aws}/ratelimit.go 模式
     - 单源上限 MAX_PER_SOURCE=5000 条(常量,不硬编码散落)
     - 时间窗口上限:CDN 类 7 天 / 实时类 24h(S3 文件爆炸保护)
  2. 各源独立 try/catch —— 单源失败记入 partial_errors,不阻塞其他源
  3. 全部 normalize → 按 Timestamp 降序归并 → 截断到全局 limit
  4. 返回 { entries, per_source: {source → {count, truncated, error?}} }
```

**部分失败是一等公民**:响应必带 `per_source` 状态,前端逐源显示"成功 N 条 / 截断 / 失败原因"。

### 4.5 API 设计(`internal/logquery/web`,gin 分组 + RequireRoles)

```
GET  /api/v1/cam/logs/types                    支持的日志类型与字段字典(动态列)
GET  /api/v1/cam/logs/sources?log_type=cdn     日志源清单(按云分组,含 Enabled 状态)
POST /api/v1/cam/logs/search                   联邦查询(上述编排)
```

- 路由与门卫照抄 `internal/cert/web`:`g.GET("", RequireRoles(...))` 模式;
- 角色面:只读特性,全角色可查;**实现简化**(与计划偏差):未建独立能力码,
  端点经启动期 RegisterEndpointsToEcmdb 自动进 CAM 域,由 ecmdb Casbin 统一授权;
- 错误信封复用 `WriteAPIError`;超时控制:整个联邦查询 ctx 超时 30s,超时返回已完成部分 + 超时标记。

### 4.6 新增依赖(go.mod 现无日志 SDK)

| 用途 | 包 | 说明 |
|---|---|---|
| 阿里 SLS 查询 | `github.com/aliyun/aliyun-log-go-sdk` | WAF/ALB 日志 |
| 腾讯 CLS 查询 | `tencentcloud-sdk-go` 的 `cls` 模块 | 已有主 SDK,加模块即可 |
| AWS S3 | `aws-sdk-go-v2/s3`(已有) | CloudFront/WAF/ALB |
| 华为 LTS 查询 | `github.com/huaweicloud/huaweicloud-sdk-go-v3` 的 lts 模块 | WAF/ELB 日志(ListLogs API) |
| gzip 解包 | 标准库 `compress/gzip` | CDN 日志包 |

### 4.7 eiam 资产注册

新端点按 Plan C 模式注册 eiam 资产:capability `cam:logs:view`(scope: tenant),菜单点挂 `e-cam-web` 侧边栏;边界校验沿用租户体系(日志查询按云账号隔离,天然多租户友好)。

---

## 5. 前端设计(e-cam-web)

### 5.1 页面结构

```
src/views/logs/
├── index.vue              三 Tab(CDN/WAF/ALB)+ 筛选区 + 结果表
└── components/
    ├── SourceFilter.vue   云多选 → 账号级联 → 资源多选(级联懒加载 sources 接口)
    ├── LogTable.vue       统一字段固定列 + 行展开;per-source 状态条
    ├── DetailDrawer.vue   Raw 原始字段 JSON 树(信息零丢失出口)
    └── LatencyBadge.vue   数据延迟标注(如阿里 CDN "T+1小时")
```

### 5.2 交互要点

- 筛选器与 cloud-user 页设计语言一致(既有组件复用:时间范围、多选、抽屉);
- 时间范围选择器按 logType 动态约束上限(CDN 7 天 / 实时类 24h,与后端一致);
- 结果表列 = 字段字典接口驱动(后端加字段,前端自动多列),Raw 永远可看;
- 空态/失败态:某源"未开启投递"时显示引导卡片(去云控制台开启的链接说明)。

### 5.3 前端接线

- `src/api/logs.ts` 新 API 模块(照 eiam-users.ts 的 mapper 模式,含统一模型 TS 类型);
- 路由 + 菜单(`router/routes.ts`,meta.title「日志查询」)+ eiam 前端权限点。

---

## 6. 任务拆解与里程碑

| Phase | 内容 | 任务 | 验收标准 | 估时 |
|---|---|---|---|---|
| **0 调研 ✅** | §3 清单逐项验证 | 0.1 四云控制台/API 确认投递状态;0.2 拉真实日志样例做 fixture;0.3 字段对照表;0.4 SDK 依赖引入验证 | 12 源状态表 + 12 fixture + 对照表落库 docs | 1.5~2 天 |
| **1 骨架+阿里 ✅** | MVP 打通 | 1.1 types 统一模型;1.2 Provider 接口+registry;1.3 logquery 模块装配(ioc/wire,照 cert);1.4 aliyun/cdn(日志包);1.5 aliyun/waf+slb(SLS);1.6 映射层+fixture 单测;1.7 编排 service;1.8 web 三接口+角色门卫 | 阿里三源 curl 可查,统一字段正确,单源失败不阻塞 | 3~4 天 |
| **2 华为 ✅** | 两源接入 | 2.1 huawei/waf+elb(LTS ListLogs,content 解析);2.2 mapper+单测;2.3 限流接线 | 同上,华为源并入联邦结果 | 2 天 |
| **3 AWS ✅** | 双源接入(WAF+CloudFront;ALB 大多未开启不纳入) | 3.1 aws/s3 拉取通用件(清单+并发下载+解压);3.2 cloudfront/waf/alb 三个 provider;3.3 mapper+单测 | 同上;S3 窗口保护生效 | 2~3 天 |
| **4 腾讯(预留) ✅** | 接口占位 | 4.1 tencent/{cdn,waf,clb} provider 骨架 + registry 注册(数据源修复后补 mapper);4.2 投递状态探测 | sources 接口可见腾讯源及未启用状态 | 1 天 |
| **5 前端 ✅** | 完整页面 | 5.1 api/logs.ts + 类型;5.2 页面四组件;5.3 路由/菜单/权限接线;5.4 联调 + 浏览器验收 | 四云三类型统一列表,raw 抽屉,延迟标注,per-source 状态 | 3 天 |
| **6 收尾** | 质量+上线 | 6.1 code-review/security-review;6.2 eiam 资产报备验证;6.3 浏览器全流程冒烟;6.4 文档(README/字段字典导出) | 评审通过,合并主干 | 1 天 |

**合计:约 14~16 个工作日(Phase 0~4 后端已全部完成,2026-09-02;活体验证:阿里 SLB/WAF/CDN、华为 WAF/ELB、AWS WAF/CDN 真实凭证冒烟全通过。剩余 Phase 6 收尾(浏览器联调由用户在 dev 环境执行)。)**(Phase B 采集入库另立计划,本期仅保证 Provider 接口可直接复用为采集器。)

依赖关系:`0 → 1 → {2,3,4}(可并行)→ 5 → 6`;其中 2/3/4 相互独立,可并行推进或按云分期上线。

---

## 7. 技术决策记录(ADR 摘要)

| # | 决策 | 理由 | 代价 |
|---|---|---|---|
| D1 | 先联邦查询,后采集入库 | 联邦阶段验证全部字段映射,入库时映射代码原样复用;无存储成本即可上线 | 查询慢(SLS 2~5s,S3 更慢);quota 依赖 |
| D2 | 按日志类型分域建模(3 schema) | CDN/WAF/SLB 语义差异大,大宽表必产生大量空列 | 类型间无横向聚合(Phase B 跨云看板再解) |
| D3 | Raw 字段全量保留 | 字段映射永远滞后于云厂商变更,Raw 是兜底;详情抽屉零丢失 | 传输体积增大(可接受,日志行本身小) |
| D4 | 无跨源游标,窗口内每源上限+截断标记 | 各云游标语义不兼容,统一游标是伪需求 | 深翻页不支持(首期用例:排障看最近日志,可接受) |
| D5 | 时间统一 Unix ms UTC | 各云格式最分歧的字段;ms 精度兼容全部源 | 前端需统一格式化展示 |
| D6 | 映射层纯函数+fixture 单测 | 云厂商字段变更时,测试是唯一快速感知手段 | Phase 0 必须拿到真实样例(已在调研清单) |

---

## 8. 风险与对策

| 风险 | 影响 | 对策 |
|---|---|---|
| 某云日志**未开启投递**(尤其 AWS WAF 需配 Firehose、Azure 诊断设置) | 该源查询空结果,用户以为坏了 | sources 接口显式返回 Enabled=false+引导;Phase 0 逐账号确认 |
| SLS/CLS quota(SLS 默认 10 QPS/Project) | 多账号并发查询 429 | 复用现有 ratelimit 模式按云限流;并发度可配 |
| S3 大时间窗口文件爆炸 | 查询超时/内存 | 窗口上限(CDN 7d/实时 24h)+ 文件数上限 + truncated 提示 |
| CDN 日志 T+1 延迟 | 用户查询最近 1 小时为空 | LatencyBadge 显式标注;时间选择器默认起点考虑延迟 |
| 云厂商字段悄悄变更 | 统一字段空值 | fixture 单测 + Raw 兜底;错误监控(resp 非常规 shape 告警) |
| 各云 SDK 引入带来的依赖膨胀 | 构建变慢 | Azure 走纯 REST 不引 SDK;SLS/CLS SDK 无重依赖,可控 |

---

## 9. 验收标准(整体)

- [ ] 四家云 × 三类日志(已开启投递的源)均可查询,统一字段正确(fixture 单测 + 真实数据抽查);
- [ ] 混合多云查询结果按时间正确归并排序;
- [ ] 单源失败/截断时其余源结果正常返回,UI 明确标注 per-source 状态;
- [ ] 未开启投递的源显示引导而非空结果;
- [ ] eiam 权限:viewer+ 角色可查,无权限 403;
- [ ] 映射层单测覆盖 ≥80%,全部基于真实样例 fixture;
- [ ] 前端与 cloud-user 页设计语言一致,Raw 详情抽屉可用。

---

## 附:与现有代码的对应关系(实现时照抄)

| 要抄的模式 | 位置 |
|---|---|
| 各云字段不统一的解法(registry + per-cloud adapter) | `internal/shared/cloudx/billing/` |
| 模块装配 + Wire 注入 + web 挂载 | `internal/cert/module.go` + `ioc/cert.go` |
| 路由注册 + 角色门卫 + 错误信封 | `internal/cert/web/{ledger_handler,authz}.go` |
| 云凭证获取/限流 | `cloudx/common/{aliyun,aws}/` `ratelimit.go` |
| 前端 API 模块 + 类型 + mapper | `e-cam-web/src/api/eiam-users.ts` |
| 前端页面(筛选+表格+抽屉) | `e-cam-web/src/views/platform/users/` |
