# -*- coding: utf-8 -*-
"""Phase 0 收尾:把各源真实日志样本固化为 testdata/ fixture 文件。"""
import io, sys, json, time, gzip, os
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts

OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'testdata')
os.makedirs(OUT, exist_ok=True)

def write_fixture(name, meta, samples):
    doc = {'meta': meta, 'sample_count': len(samples), 'samples': samples}
    path = os.path.join(OUT, name + '.json')
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(doc, f, ensure_ascii=False, indent=2)
    print(f'  -> {name}.json ({len(samples)} samples)')

now = int(time.time())

# ========== 阿里云 SLS ==========
print('== aliyun SLS ==')
from aliyun.log import LogClient, GetLogsRequest
from aliyun.log.logexception import LogException
acc = accounts()['aliyun']
ak, sk = acc['ak'], acc['sk']

ALIYUN_TARGETS = [
    ('alb-access', 'ALB 访问日志', 'cn-shenzhen', 'jlc-lb-log', 'alb-mvezktcft1d81uojx9'),
    ('alb-access-overseas', '海外 ALB 访问日志', 'eu-central-1', 'jlc-prod-overseas-log', 'alb_overseas-access-gateway_lb_access'),
    ('waf3-access', 'WAF3.0 访问日志', 'eu-central-1', 'wafnew-project-1210557380197478-eu-central-1', 'wafnew-logstore'),
    ('akamai-cdn', 'Akamai CDN(自采)', 'eu-central-1', 'jlc-prod-akamai-cdnwaf-log', 'jlc-prod-akamai-cdn-log'),
    ('akamai-waf', 'Akamai WAF(自采)', 'eu-central-1', 'jlc-prod-akamai-cdnwaf-log', 'jlc-prod-akamai-waf-log'),
    ('dcdn-edge-rtlog', 'DCDN 边缘实时日志', 'cn-shenzhen', 'dcdn-edge-rtlog-cn-42d9825f', 'dcdn-edge-rtlog'),
]
for name, desc, region, project, logstore in ALIYUN_TARGETS:
    client = LogClient(f'{region}.log.aliyuncs.com', ak, sk)
    try:
        for window in (3600, 7 * 86400, 30 * 86400):
            req = GetLogsRequest(project, logstore, now - window, now, '', '*', 3, 0, False)
            resp = client.get_logs(req)
            logs = resp.get_logs()
            if logs:
                write_fixture('aliyun-' + name,
                              {'cloud': 'aliyun', 'log_type': name, 'desc': desc, 'region': region,
                               'project': project, 'logstore': logstore,
                               'query_window_seconds': window, 'pulled_at': now},
                              [lg.get_contents() for lg in logs])
                break
        else:
            print(f'  [aliyun-{name}] {desc}: 无数据(30天)')
    except LogException as e:
        print(f'  [aliyun-{name}] {desc}: 错误 {e.get_error_code()}')

# ========== AWS S3 ==========
print('\n== AWS S3 ==')
import boto3
from botocore.config import Config
acc = accounts()['aws']
s3 = boto3.client('s3', region_name='eu-west-1',
                  aws_access_key_id=acc['ak'], aws_secret_access_key=acc['sk'],
                  config=Config(retries={'max_attempts': 2}, connect_timeout=8, read_timeout=30))

AWS_TARGETS = [
    ('aws-waf-cloudfront', 'AWS WAFv2 (CLOUDFRONT scope, Firehose->S3)', 'waf-loghub-clloggingbucket5f34e4eb-hjalewqix1db', 'AWSLogs/'),
    ('aws-cloudfront', 'CloudFront 标准日志', 'jlc-prod-cdn-log', ''),
]
for name, desc, bucket, prefix in AWS_TARGETS:
    try:
        objs = []
        paginator = s3.get_paginator('list_objects_v2')
        for page in paginator.paginate(Bucket=bucket, Prefix=prefix, MaxKeys=300):
            objs.extend(page.get('Contents', []))
            if len(objs) >= 300:
                break
        if not objs:
            print(f'  [{name}] {bucket}: 空')
            continue
        objs.sort(key=lambda o: o['LastModified'])
        latest = objs[-1]
        body = s3.get_object(Bucket=bucket, Key=latest['Key'])['Body'].read()
        if latest['Key'].endswith('.gz'):
            body = gzip.decompress(body)
        text = body.decode('utf-8', 'replace')
        lines = [l for l in text.strip().splitlines() if l][:3]
        # WAF 是 JSON Lines;CloudFront 是 TSV
        samples = []
        for l in lines:
            try:
                samples.append(json.loads(l))
            except Exception:
                samples.append({'_raw_line': l})
        write_fixture(name,
                      {'cloud': 'aws', 'log_type': name.split('-', 1)[1], 'desc': desc,
                       'bucket': bucket, 'key': latest['Key'],
                       'format': 'json-lines' if isinstance(samples[0], dict) and '_raw_line' not in samples[0] else 'tsv',
                       'object_time': str(latest['LastModified']), 'pulled_at': now},
                      samples)
    except Exception as e:
        print(f'  [{name}] 失败: {str(e)[:150]}')

# ========== 华为 LTS ==========
print('\n== 华为 LTS ==')
from huaweicloudsdkcore.auth.credentials import BasicCredentials
from huaweicloudsdklts.v2.region.lts_region import LtsRegion
from huaweicloudsdklts.v2 import LtsClient, ListLogGroupsRequest, ListLogStreamsRequest, ListLogsRequest
from huaweicloudsdklts.v2.model import QueryLtsLogParams
acc = accounts()['huawei']
cred = BasicCredentials(acc['ak'], acc['sk'])
client = LtsClient.new_builder().with_credentials(cred).with_region(LtsRegion.value_of('cn-south-1')).build()
group_ids = {g.log_group_name: g.log_group_id for g in (client.list_log_groups(ListLogGroupsRequest()).log_groups or [])}
HW_TARGETS = [
    ('waf-attack', '华为 WAF 攻击日志', 'hwyun-waf-logs', 'hwyun-waf-logs-attack'),
    ('waf-access', '华为 WAF 访问日志', 'hwyun-waf-logs', 'hwyun-waf-logs-access'),
    ('elb-access', '华为 ELB 访问日志', 'eda-prod-elb', 'eda-prod-gw-std-elb'),
]
for name, desc, gname, sname in HW_TARGETS:
    try:
        gid = group_ids.get(gname)
        streams = client.list_log_streams(ListLogStreamsRequest(log_group_name=gname)).log_streams or []
        st = next((s for s in streams if s.log_stream_name == sname), None)
        if not st:
            print(f'  [huawei-{name}] 流不存在')
            continue
        resp = client.list_logs(ListLogsRequest(
            log_group_id=gid, log_stream_id=st.log_stream_id,
            body=QueryLtsLogParams(start_time=now * 1000 - 3600_000, end_time=now * 1000,
                                   keywords='', limit=3, search_type='forwards')))
        logs = getattr(resp, 'logs', None) or []
        samples = []
        for l in logs:
            d = l.to_dict() if hasattr(l, 'to_dict') else dict(l)
            samples.append(d)
        if samples:
            write_fixture('huawei-' + name,
                          {'cloud': 'huawei', 'log_type': name, 'desc': desc,
                           'group': gname, 'stream': sname, 'region': 'cn-south-1', 'pulled_at': now},
                          samples)
        else:
            print(f'  [huawei-{name}] {desc}: 近1h无数据')
    except Exception as e:
        print(f'  [huawei-{name}] 失败: {str(e)[:150]}')

print('\n完成。输出目录:', OUT)
