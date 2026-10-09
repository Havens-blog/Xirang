# -*- coding: utf-8 -*-
"""末轮探测:华为 LTS 日志流+样本 / 腾讯 CDN 日志包 / 火山 TLS。"""
import io, sys, json, time
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts

# ---------- 华为:日志组下的流 + 样本 ----------
print('===== 华为 LTS 流与样本 =====')
try:
    from huaweicloudsdkcore.auth.credentials import BasicCredentials
    from huaweicloudsdklts.v2.region.lts_region import LtsRegion
    from huaweicloudsdklts.v2 import LtsClient, ListLogGroupsRequest, ListLogStreamsRequest, ListLogsRequest
    from huaweicloudsdklts.v2.model import QueryLtsLogParams
    acc = accounts()['huawei']
    cred = BasicCredentials(acc['ak'], acc['sk'])
    client = LtsClient.new_builder().with_credentials(cred).with_region(LtsRegion.value_of('cn-south-1')).build()
    group_ids = {}
    for g in (client.list_log_groups(ListLogGroupsRequest()).log_groups or []):
        group_ids[g.log_group_name] = g.log_group_id
    for gname in ['hwyun-waf-logs', 'eda-prod-elb']:
        gid = group_ids.get(gname)
        streams = client.list_log_streams(ListLogStreamsRequest(log_group_name=gname))
        sl = streams.log_streams or []
        print(f"  group {gname} ({gid}): {len(sl)} streams -> {[s.log_stream_name for s in sl][:10]}")
        if not sl:
            continue
        s = sl[0]
        now = int(time.time() * 1000)
        try:
            resp = client.list_logs(ListLogsRequest(
                log_group_id=gid, log_stream_id=s.log_stream_id,
                body=QueryLtsLogParams(start_time=now - 3600_000, end_time=now,
                                       keywords='', limit=5, search_type='forwards')))
            logs = getattr(resp, 'logs', None) or []
            print(f"    stream {s.log_stream_name} 近1h日志条数: {len(logs)}")
            for l in logs[:2]:
                print('    样本:', str(l.to_dict() if hasattr(l, 'to_dict') else l)[:300])
        except Exception as qe:
            print('    查询失败:', str(qe)[:180])
except Exception as e:
    print('  失败:', str(e)[:200])

# ---------- 腾讯:CDN 域名 + 日志包下载链接 ----------
print('\n===== 腾讯 CDN =====')
try:
    from tencentcloud.common import credential
    from tencentcloud.cdn.v20180606 import cdn_client, models as cdn_models
    acc = accounts()['tencent']
    cred = credential.Credential(acc['ak'], acc['sk'])
    client = cdn_client.CdnClient(cred, 'ap-guangzhou')
    req = cdn_models.DescribeDomainsRequest()
    req.Offset, req.Limit = 0, 100
    doms = client.DescribeDomains(req)
    print(f"  CDN 域名数: {doms.TotalNumber}")
    for d in (doms.Domains or [])[:10]:
        print(f"    {d.Domain} ({d.Status})")
    if doms.Domains:
        # 拿第一个在线域名昨天的日志包
        import datetime
        online = [d for d in doms.Domains if d.Status == 'online']
        if online:
            yesterday = datetime.date.today() - datetime.timedelta(days=1)
            lr = cdn_models.DescribeCdnDomainLogsRequest()
            lr.Domain = online[0].Domain
            lr.StartTime = yesterday.strftime('%Y-%m-%d 00:00:00')
            lr.EndTime = yesterday.strftime('%Y-%m-%d 23:59:59')
            lr.Offset, lr.Limit = 0, 5
            try:
                logs = client.DescribeCdnDomainLogs(lr)
                for dl in (logs.DomainLogs or [])[:2]:
                    print(f"    {online[0].Domain} 日志包: {str(dl.LogPath)[:150]}")
            except Exception as e2:
                print('    DescribeCdnDomainLogs 失败:', str(e2)[:150])
except Exception as e:
    print('  失败:', str(e)[:200])

# ---------- 火山:TLS 项目/主题 ----------
print('\n===== 火山 TLS =====')
try:
    from volcengine.tls.TLSService import TLSService
    acc = accounts()['volcengine']
    tls = TLSService('https://tls.cn-guangzhou.volces.com', acc['ak'], acc['sk'], 'cn-guangzhou')
    projects = tls.describe_projects()
    print('  describe_projects keys:', str(projects)[:400])
except Exception as e:
    print('  失败:', str(e)[:200])
