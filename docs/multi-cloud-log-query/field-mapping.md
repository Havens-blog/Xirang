# 多云日志查询 · 字段映射表(raw -> unified)

> 依据 `testdata/` 下 11 份真实样本(2026-08-27 实时拉取)整理。
> 统一模型:`CDNLogEntry` / `WAFLogEntry` / `SLBLogEntry` + `LogMeta` + `Raw`(见 plan.md §4.1)。
> **硬规则**:时间统一 Unix 毫秒 UTC;枚举归一小写;缺失容忍(UI 显 `-`);`Raw` 全量保留。

---

## 一、SLB 域(`SLBLogEntry`)

### 1. 阿里云 ALB 访问日志(SLS,topic=alb_layer7_access_log)

来源:`jlc-lb-log`(国内)/ `jlc-prod-overseas-log`(海外),schema 34 字段跨 region 一致。

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `time` | ISO8601 带时区(`2026-08-20T20:55:57+08:00`)-> ms UTC |
| ClientIP | `client_ip` | 直取 |
| ClientPort | `client_port` | 直取 |
| Method | `request_method` | 直取 |
| URL | `scheme` + `http_host` + `request_uri` | 拼接 |
| Host | `http_host` | 直取 |
| Protocol | `server_protocol` | 直取 |
| Status | `status` | int |
| RequestLength | `request_length` | int |
| BytesSent | `body_bytes_sent` | int |
| LatencyMs | `request_time` | 秒(字符串)×1000 |
| TargetIP/Port | `upstream_addr` | `172.19.152.31:6090` 拆分;多后端逗号分隔取首个 |
| UpstreamLatencyMs | `upstream_response_time` | 秒,可能逗号多个,取首个×1000 |
| UpstreamStatus | `upstream_status` | int,可能多个 |
| TLS | `ssl_protocol` / `ssl_cipher` | 直取 |
| ResourceID | `app_lb_id` | `alb-mvezktcft1d81uojx9` |

### 2. 华为云 ELB 访问日志(LTS,空格分隔文本,content 内)

来源:`eda-prod-elb` 组,`eda-prod-gw-*-elb` 流。格式(官方 25+ 列,关键位):

```
1787824562.052 <uuid> [2026-08-27T17:56:02+08:00] elb_01 10.41.185.47:49798 200 "GET https://inner.oshwhub.com/... HTTP/1.1" 295 10942 10212 0.032 "200" "0.000" "0.031" "0.032" "10.41.185.220:8080" "UA" ... loadbalancer_<id> listener_<id> ... pool_<id>*<lb-id> ...
```

| 统一字段 | 原始位置 | 转换 |
|---|---|---|
| Timestamp | 第 3 列(1-based)`[2026-08-27T17:56:02+08:00]` | ISO8601;第 1 列 msec 兜底 |
| ClientIP/Port | 第 5 列 `10.41.185.47:49798` | 拆分 |
| Status | 第 6 列 | int |
| Method/URL/Protocol | 第 7 列 `"GET https://… HTTP/1.1"` | 引号内按空格拆 3 段;Host 从绝对 URL 解析 |
| BytesSent | 第 8 列(body_bytes_sent) | int |
| RequestLength | 第 10 列 | int |
| LatencyMs | 第 11 列 request_time(总耗时,秒) | ×1000 |
| UpstreamStatus | 第 12 列 `"200"` | 引号内 int |
| UpstreamLatencyMs | 第 15 列 upstream_response_time | 秒×1000 |
| TargetIP/Port | 第 16 列 `"10.41.185.220:8080"` | 引号内拆分 |
| UserAgent | 第 17 列 | 引号内 |
| ResourceID | 19 列后 `loadbalancer_<uuid>` | 提取 uuid(列数随版本浮动,扫描而非定死) |

解析策略:正则按列位 + 引号段混合;列数可能随格式版本浮动,**解析失败必须降级为"整行进 Raw,统一字段留空"而非报错**。

---

## 二、WAF 域(`WAFLogEntry`)

### 3. 阿里云 WAF3.0 访问日志(SLS wafnew-logstore,topic=waf_access_log)

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `start_time` | unix 秒 -> ms(`time` 是 ISO 备用) |
| ClientIP | `real_client_ip`(优先)/ `src_ip` / `remote_addr` | 三级回退 |
| Host | `host` / `matched_host` | 直取 |
| URI | `request_uri` + `querystring` | 拼接 |
| Method | `request_method` | 直取 |
| Status | `status` | int |
| RuleID/RuleName | -(访问流无) | 留空 |
| Action | -(访问流) | 视作 `pass` |
| LatencyMs | `request_time_msec` | 已是 ms,可能多值取首个 |
| UpstreamLatencyMs | `upstream_response_time` | 秒×1000 |
| Geo | `region`(`int`=国际) | 直取 |
| ResourceID | `host` | 以域名为资源标识 |

### 4. 华为云 WAF 攻击日志(LTS hwyun-waf-logs-attack,content 内 JSON)

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `attack-time` | ISO8601 Z(`2026-08-27T09:56:02.000Z`)-> ms |
| ClientIP | `remote_ip` | 直取 |
| Host | `host` | 直取 |
| URI | `raw_uri` | 直取 |
| Method | `method` | 直取 |
| RuleName | `rule_name`(中文,如 `恶意爬虫2`) | 直取 |
| Severity | `level`(数字 1-5) | 映射为低/中/高 |
| Action | `action`(`block`) | 直取(已小写) |
| Geo | `geo_str`(`US:0`、`CN:HE`) | 直取 |
| ResourceID | `host` | 域名 |

访问流(hwyun-waf-logs-access)字段类似:`time_iso8601`、`remote_ip`、`x_forwarded_for`、`method`、`request_length`、`body_bytes_sent`、`upstream_response_time`、`waf-time`(ISO Z)。

### 5. Akamai WAF(SLS jlc-prod-akamai-waf-log,CEF 格式展开为字段)

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `start`(unix 秒) | ×1000;`_unixtimestamp_` 是**纳秒**,仅做校验别直接用 |
| ClientIP | `src` | 直取 |
| Host | `dhost` | 直取 |
| URI | `dpath`(路径)/ `request`(完整 URL) | 优先 request |
| Method | `requestMethod` | 直取 |
| RuleName/RuleID | `name`(如 `Activity detected`)/ `cs1`(`Rules` label,规则 ID 串) | 直取 |
| Severity | `severity`(1-10) | 映射低/中/高 |
| Action | `act`(`alert`/`block`…) | 直取 |
| Port | `dpt` | 直取 |

### 6. AWS WAFv2(S3 JSON Lines,Firehose 投递)

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `timestamp` | 已是 ms,直取 |
| ClientIP | `httpRequest.clientIp` | 直取 |
| Geo | `httpRequest.country`(`DE`) | 直取 |
| Host | `httpRequest.headers` 中 name=host | headers 是数组需查找 |
| URI | `httpRequest.uri` + `httpRequest.args` | 拼接(args 已 URL 编码) |
| Method | `httpRequest.httpMethod` | 直取 |
| RuleID | `terminatingRuleId`(`Default_Action` 亦即放行) | 直取 |
| Action | `action` | `ALLOW`->`allow`,`BLOCK`->`block` 小写化 |
| ResourceID | `httpSourceId`(distribution ID,如 `E2U51XYZHX0AP`) | 直取 |
| UA | `httpRequest.headers` 中 name=user-agent | 查找 |
| 指纹 | `ja3Fingerprint` / `ja4Fingerprint` | 附加进 Raw |

---

## 三、CDN 域(`CDNLogEntry`)

### 7. 阿里云 DCDN 边缘实时日志(SLS dcdn-edge-rtlog)

> **查询形态**:单 logstore 混装全部域名,provider 按域名扇出(100 条探查拿活跃域名 -> `domain: xxx` 逐域名并发,每域名各自凑配额,热点域名不挤占长尾)。
> **同构源**:CDN(非 DCDN)实时日志投递 `jlc-prod-cdn-log-monitor` 与本 schema 完全一致(31 字段),复用 kindDCDN mapper 与域名扇出(2026-09-04 确认)。

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `unixtime` | unix 秒 -> ms |
| Domain(ResourceID) | `domain` | 直取 |
| ClientIP | `client_ip` | 直取 |
| Method | `method` | 直取 |
| URL | `scheme` + `domain` + `uri` + `uri_param` | 拼接 |
| Status | `return_code` | int |
| BytesSent | `response_size` | int |
| RequestSize | `request_size` | int |
| LatencyMs | `request_time` | **已是 ms 直取**(曾误套 ALB 的秒->ms 放大 1000 倍;fixture 实测 2~10001,按秒解读会是小时级) |
| CacheHit | `hit_info`(`-,WS|CHARGE|NOTLAST`) | 取 `|` 分段首列做命中率判据 |
| EdgeNode | `via_info`(`cache15.l2eu95-4[...]`) | 取首个节点名 |
| Referer | `refer_uri` / `refer_domain` | 直取 |
| UserAgent | `user_agent` | 直取 |

### 8. Akamai CDN(SLS jlc-prod-akamai-cdn-log,自采,69 字段)

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `reqTimeSec`(1787225893.800,注意:名为 time 实为**请求时刻戳**) | ×1000 取整 |
| Domain | `reqHost` | 直取 |
| ClientIP | `cliIP` | 直取 |
| Method | `reqMethod` | 直取 |
| URL | `reqHost` + path 字段 | path 在样本中见 `reqPath` 类字段;以 fixture 实际为准 |
| Status | `statusCode` | int |
| BytesSent | `bytes`(=objSize) | int |
| LatencyMs | `turnAroundTimeMSec` | 已是 ms |
| CacheHit | `cacheStatus`(数字码,`0`…) | 码表映射:0->hit(按 Akamai 文档核) |
| Referer | `referer` | 直取 |
| UserAgent | `UA` | **URL 编码串**(`%20`),需 url-decode |
| EdgeNode | `cp`(POP 代码 `1928404`)/ `edgeIP` | 直取 |
| RequestID | `reqId` | 直取 |
| TLS | `tlsVersion` | 直取 |
| Geo | `serverCountry` / `city` | 直取 |

### 9. AWS CloudFront 标准日志(S3 TSV,gzip)

`#Fields:` 头声明 33 列,以头驱动解析(勿写死列位):

```
date time x-edge-location sc-bytes c-ip cs-method cs(Host) cs-uri-stem sc-status
cs(Referer) cs(User-Agent) cs-uri-query cs(Cookie) x-edge-result-type
x-edge-request-id x-host-header cs-protocol cs-bytes time-taken
x-forwarded-for ssl-protocol ssl-cipher x-edge-response-result-type
cs-protocol-version fle-status fle-encrypted-fields c-port time-to-first-byte
x-edge-detailed-result-type sc-content-type sc-content-len sc-range-start sc-range-end
```

| 统一字段 | 原始列 | 转换 |
|---|---|---|
| Timestamp | `date` + `time` | UTC 日期+时间 -> ms(无时区标记,CloudFront 均为 UTC) |
| Domain | `x-host-header` | 直取(cs(Host) 是 dist 域名) |
| ClientIP | `c-ip` | 直取 |
| Method | `cs-method` | 直取 |
| URL | `x-host-header` + `cs-uri-stem` + `cs-uri-query` | 拼接;stem 已 URL 编码可保留 |
| Status | `sc-status` | int |
| BytesSent | `sc-bytes` | int |
| LatencyMs | `time-taken` | 秒×1000 |
| CacheHit | `x-edge-result-type` / `x-edge-detailed-result-type` | `Hit`->`hit`,`Error`/`Miss`->相应枚举 |
| Referer | `cs(Referer)` | `-` -> 空 |
| UserAgent | `cs(User-Agent)` | 直取 |
| EdgeNode | `x-edge-location`(`CDG54-P2`) | 直取 |
| RequestID | `x-edge-request-id` | 直取 |
| ResourceID | 桶 key 前缀(域名) | `3d.jlcpcb.com/E1OVWLY3JDS4DE.…gz` |

---

### 10. 阿里云 CDN 离线转存(SLS jlc-prod-cdn-log,kindCDNOffline)

logstore 即转存任务(单站点),**独立 PascalCase schema**,与 DCDN 完全不同
(2026-09-03 误判同构,2026-09-04 独立 mapper 后重新纳入 catalog;项目内动态枚举,
当前 jlcfa-his=www.jlcfa.com + api_forface3d_com,metrics/internal 流已过滤):

```
ResponseSize / ContentType / RequestTime / Referer / HTTPMethod / HTTPStatus /
RequestSize / Time / ProxyIP / RemoteIP / UserAgent / RequestURL / HitInfo
```

| 统一字段 | 原始字段 | 转换 |
|---|---|---|
| Timestamp | `Time` | nginx 格式 `3/Sep/2026:11:23:01 +0800` -> ms UTC(ParseTimeMs 第六种入参) |
| Domain(ResourceID) | `RequestURL` | 解析 netloc 提取 host(无独立 domain 字段) |
| URL | `RequestURL` | 直取(完整 URL) |
| ClientIP | `RemoteIP` | 直取 |
| Method | `HTTPMethod` | 直取 |
| Status | `HTTPStatus` | int |
| BytesSent | `ResponseSize` | int |
| LatencyMs | `RequestTime` | 已是 ms 直取(实测 78~221 与 API 耗时吻合) |
| CacheHit | `HitInfo`(HIT/MISS) | 小写归一 |
| Referer | `Referer` | `-` -> 空 |
| UserAgent | `UserAgent` | 直取 |

无统一字段对应(仅留 Raw):`ProxyIP`(回源代理,常为 `-`)、`ContentType`、`RequestSize`。

## 四、腾讯云(本期预留映射,待数据源修复后校准)

CLS `SearchLog` 返回结构已知(topic 存在),但 30 天 0 数据,映射表待真实样本到位后补。
CDN 离线日志包格式为腾讯标准(TSV,时间|域名|IP|方法|…),映射骨架先按公开文档写,fixture 缺失标注 `unverified`。

## 五、火山云

TLS API 本机网络不可达,暂无映射。

---

## 单测要求

- 每个 fixture 一条 golden test:读 `testdata/*.json` -> 走映射函数 -> 断言统一字段 + Raw 完整性;
- 时间归一专项:ALB ISO+08:00 / DCDN unix 秒 / Akamai 纳秒 / AWS WAF ms / CloudFront UTC 拼接,五种入参都收敛到同一 Unix ms;
- 枚举归一专项:action(ALLOW/block/alert/intercept)、cache(TCP_HIT/Hit/0/Error);
- 解析降级专项:华为 ELB 列数不符、CloudFront 缺 `#Fields` 头时,整条进 Raw 不 panic。
