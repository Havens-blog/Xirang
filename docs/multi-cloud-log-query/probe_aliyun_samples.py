# -*- coding: utf-8 -*-
"""Phase 0.3:从关键 SLS logstore 拉真实日志样本,做映射层 fixture。"""
import io, sys, json, time
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts
from aliyun.log import LogClient, GetLogsRequest
from aliyun.log.logexception import LogException

acc = accounts()['aliyun']
ak, sk = acc['ak'], acc['sk']
now = int(time.time())

# (region, project, logstore, 说明) -- 来自 probe_aliyun_sls.py 的盘点结果
TARGETS = [
    ('cn-shenzhen', 'jlc-lb-log', 'alb-mvezktcft1d81uojx9', 'ALB 访问日志(实例名命名)'),
    ('cn-shenzhen', 'jlc-lb-log', 'jlc-lb-log', 'LB 日志(聚合)'),
    ('cn-shenzhen', 'jlc-prod-cdn-log', 'api_forface3d_com', 'CDN 域名日志'),
    ('cn-shenzhen', 'aliyun-cloudsiem-channel-1210557380197478-cn-shenzhen', 'wafng-logstore', 'WAF3.0(云安全中心渠道)'),
    ('eu-central-1', 'wafnew-project-1210557380197478-eu-central-1', 'wafnew-logstore', 'WAF 日志库'),
    ('eu-central-1', 'jlc-prod-overseas-log', 'alb_overseas-access-gateway_lb_access', '海外 ALB 访问日志'),
    ('cn-shenzhen', 'jlc-prod-akamai-cdnwaf-log', 'jlc-prod-akamai-cdn-log', 'Akamai CDN(自采)'),
    ('cn-shenzhen', 'jlc-prod-akamai-cdnwaf-log', 'jlc-prod-akamai-waf-log', 'Akamai WAF(自采)'),
    ('cn-shenzhen', 'dcdn-edge-rtlog-cn-42d9825f', 'dcdn-edge-rtlog', 'DCDN 边缘实时日志'),
]

for region, project, logstore, desc in TARGETS:
    client = LogClient(f'{region}.log.aliyuncs.com', ak, sk)
    try:
        # 先试最近 1 小时,再退到最近 7 天
        for window in (3600, 7 * 86400):
            req = GetLogsRequest(project, logstore, now - window, now, '', '*', 2, 0, False)
            resp = client.get_logs(req)
            logs = resp.get_logs()
            if logs:
                print(f'\n===== {desc} | {region}/{project}/{logstore} | 窗口{window}s =====')
                for lg in logs:
                    d = lg.get_contents()
                    print(json.dumps(d, ensure_ascii=False)[:1200])
                break
        else:
            print(f'\n===== {desc} | {region}/{project}/{logstore} | 无数据 =====')
    except LogException as e:
        print(f'\n===== {desc} | {region}/{project}/{logstore} | 错误: {e.get_error_code()} {e.get_error_message()[:100]} =====')
