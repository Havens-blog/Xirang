# -*- coding: utf-8 -*-
"""收尾:火山 TLS 项目+主题 / 腾讯 CDN 日志包下载链接。"""
import io, sys, time
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts

print('===== 火山 TLS =====')
try:
    from volcengine.tls.TLSService import TLSService
    from volcengine.tls.tls_requests import DescribeProjectsRequest, DescribeTopicsRequest
    acc = accounts()['volcengine']
    tls = TLSService('https://tls.cn-guangzhou.volces.com', acc['ak'], acc['sk'], 'cn-guangzhou')
    projects = tls.describe_projects(DescribeProjectsRequest(page_number=1, page_size=50))
    pj = projects.get_projects() if hasattr(projects, 'get_projects') else []
    print(f'  项目数: {len(pj)}')
    for p in pj[:10]:
        print(f"    {p.get('ProjectName')} ({p.get('ProjectId')})")
        try:
            topics = tls.describe_topics(DescribeTopicsRequest(project_id=p.get('ProjectId'), page_number=1, page_size=50))
            tl = topics.get_topics() if hasattr(topics, 'get_topics') else []
            for t in tl[:8]:
                print(f"      topic: {t.get('TopicName')}")
        except Exception as te:
            print('      topic查询失败:', str(te)[:120])
except Exception as e:
    print('  失败:', str(e)[:250])

print('\n===== 腾讯 CDN 日志包 =====')
try:
    from tencentcloud.common import credential
    from tencentcloud.cdn.v20180606 import cdn_client, models as cdn_models
    import datetime
    acc = accounts()['tencent']
    cred = credential.Credential(acc['ak'], acc['sk'])
    client = cdn_client.CdnClient(cred, 'ap-guangzhou')
    req = cdn_models.DescribeDomainsRequest()
    req.Offset, req.Limit = 0, 100
    doms = client.DescribeDomains(req)
    online = [d for d in (doms.Domains or []) if d.Status == 'online']
    for d in online[:3]:
        yesterday = datetime.date.today() - datetime.timedelta(days=1)
        lr = cdn_models.DescribeCdnDomainLogsRequest()
        lr.Domain = d.Domain
        lr.StartTime = yesterday.strftime('%Y-%m-%d 00:00:00')
        lr.EndTime = yesterday.strftime('%Y-%m-%d 23:59:59')
        lr.Offset, lr.Limit = 0, 3
        try:
            logs = client.DescribeCdnDomainLogs(lr)
            total = getattr(logs, 'TotalCount', None)
            dls = logs.DomainLogs or []
            print(f"  {d.Domain}: TotalCount={total}, DomainLogs={len(dls)}")
            for dl in dls[:2]:
                print(f"    包: {str(dl.LogPath)[:140]}")
        except Exception as e2:
            print(f"  {d.Domain} 失败:", str(e2)[:160])
except Exception as e:
    print('  失败:', str(e)[:200])
