# 多云日志查询 · Phase 0 源状态盘点

> 探测时间:2026-08-27(实时 API 探测,非离线推断)
> 账号清单来自 MongoDB `ecam.ecam_cloud_account`(5 个账号,全部 active,tenant_id=3/jlc)

## 总览

| 云 | 账号 | 区域 | WAF 日志 | LB/CDN 日志 | 结论 |
|---|---|---|---|---|---|
| 阿里云 | 集团-阿里云 (id=1) | cn-shenzhen / cn-wulanchabu / cn-guangzhou / eu-central-1 / ap-southeast-1 | ✅ 流动中 | ✅ ALB/DCDN/Akamai 全部流动 | **首选数据源,已高度集中** |
| 华为云 | 集团-华为云 (id=2) | cn-south-1 | ✅ 流动中 | ✅ ELB 流动中 | **LTS 可直接接入** |
| AWS | 集团-AWS (id=3) | eu-west-1 / eu-central-1 (LB) | ✅ 流动中(今日仍 ACTIVE) | ⚠️ CloudFront ON;ALB 大多 OFF | CloudFront+WAF 可用;ALB 需开开关 |
| 腾讯云 | 集团-腾讯云 (id=4) | ap-guangzhou / ap-shanghai | ❌ topic 存在但 30 天 0 条 | ❌ CDN 无离线日志包 | **暂无可用日志,需排查投递** |
| 火山云 | 集团-火山云 (id=5) | cn-guangzhou | ❓ TLS API 网络不可达 | ❓ 同左 | 暂缓,列为后续 |
| Azure | — | — | — | — | **无账号,移出范围** |

## 分云明细

### 阿里云(数据最全,均已入 SLS)

| 日志类型 | 位置(region/project/logstore) | 状态 |
|---|---|---|
| ALB 访问日志(国内) | cn-shenzhen / jlc-lb-log / alb-mvezktcft1d81uojx9(topic=alb_layer7_access_log) | ✅ 实时 |
| ALB 访问日志(海外) | eu-central-1 / jlc-prod-overseas-log / alb_overseas-access-gateway_lb_access | ✅ 实时 |
| WAF3.0 访问日志 | eu-central-1 / wafnew-project-1210557380197478-eu-central-1 / wafnew-logstore | ✅ 实时 |
| Akamai CDN(自采) | eu-central-1 / jlc-prod-akamai-cdnwaf-log / jlc-prod-akamai-cdn-log | ✅ 流动 |
| Akamai WAF(自采,CEF) | eu-central-1 / jlc-prod-akamai-cdnwaf-log / jlc-prod-akamai-waf-log | ✅ 流动 |
| DCDN 边缘实时日志 | cn-shenzhen / dcdn-edge-rtlog-cn-42d9825f / dcdn-edge-rtlog | ✅ 实时 |

国内 WAF 还有云安全中心渠道:cn-shenzhen / aliyun-cloudsiem-channel-…/ wafng-logstore。

**要点:绝大多数日志已汇聚到 SLS, federation 查询直接走 SLS GetLogs 即可,无需碰 Kafka。**

### 华为云(LTS 就绪)

| 日志类型 | 位置(group/stream) | 状态 |
|---|---|---|
| WAF 攻击日志 | hwyun-waf-logs / hwyun-waf-logs-attack | ✅ 实时(JSON content) |
| WAF 访问日志 | hwyun-waf-logs / hwyun-waf-logs-access | ✅ 实时(JSON content) |
| ELB 访问日志 | eda-prod-elb / eda-prod-gw-{std,u,asdw,pro}-elb | ✅ 实时(空格分隔文本) |

另有 eda-uat-elb(UAT)、k8s-log-*、CTS、apm 等日志组,与本次范围无关。

### AWS

| 日志类型 | 位置 | 状态 |
|---|---|---|
| WAFv2(CLOUDFRONT scope) | Firehose `aws-waf-logs-CL-s3-*` → s3://waf-loghub-clloggingbucket…(及 loghub-… 第二桶,均在 eu-central-1),路径 `AWSLogs/414660254120/WAFLogs/cloudfront/<acl>/` | ✅ 实测 8 个 ACL 源,JSON Lines,当日流动 |
| CloudFront 标准日志 | s3://jlc-prod-cdn-log(eu-central-1),按域名前缀,TSV(#Version/#Fields 头) | ✅ 实测 48 个域名目录,当日数据流动 |
| ALB 访问日志 | eu-central-1,仅 4 个 LB 开启(prod-proxy-access-to-ali-lb[02]、SDPS-A-Appli、CL-Pro-LoadB) | ⚠️ 大多 access_log=false |
| 空桶 | aws-waf-logs-jlc、prod-jlc-cloudfront-log | 空,忽略 |

**ALB 日志若要统一查询,需要在 AWS 控制台为各 LB 开启 access log(不影响本期 federation 先行)。**

### 腾讯云(topic 建了,数据没来)

| 日志类型 | 位置 | 状态 |
|---|---|---|
| WAF 访问日志 | ap-shanghai CLS topic `waf_access_logtopic` (a0bfd8ed-…) | ❌ 30 天 0 条 |
| EdgeOne | ap-guangzhou CLS topic `eo-log` (f9a20d7f-…) | ❌ 30 天 0 条 |
| CDN 离线日志包 | api.jlc.com 等 6 个在线域名,DescribeCdnDomainLogs 多天 | ❌ TotalCount=0 |

在线 CDN 域名:trade.jlc.com、m.jlc.com、helper.jlc.com、api.jlc.com 等 6 个(test* 已下线)。
**行动项:联系腾讯云侧负责人确认 WAF 日志服务是否开通、EdgeOne 日志推送是否启用。本期先不实现腾讯 adapter,接口留位。**

### 火山云

TLS 端点 `tls.cn-guangzhou.volces.com` / `tls.volces.com` 均请求超时(本机网络不可达,或服务未开通)。本期不实现,列为后续验证项。

### Azure

无云账号存在,原计划中的 Azure Phase 移除,由华为云 LTS 替补(实际数据情况比预期更好)。

## 对实施计划的影响

1. **范围调整**:Azure OUT;华为 LTS IN(Phase 2);腾讯降级为"接口预留,数据源待修复";火山暂缓。
2. **Akamai 是意外收获**:CDN+WAF 日志已自采进 SLS,一次 SLS 查询可同时覆盖"阿里系+Akamai"。
3. **federation 优先策略完全成立**:阿里(SLS)+ 华为(LTS)+ AWS(WAF S3 JSON/CloudFront TSV)均可在查询时实时拉取,无需先建 ingestion 管道。
4. **AWS ALB 部分缺失**是云上配置问题,不是平台问题,单独跟进,不阻塞本期。
