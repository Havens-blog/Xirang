# -*- coding: utf-8 -*-
"""补充探测:AWS WAF Firehose 落桶样本 / 腾讯30天窗口+CDN日志包 / 华为 LTS。"""
import io, sys, json, gzip, time
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts

# ---------- AWS:Firehose -> S3 落桶 + WAF 日志样本 ----------
print('===== AWS WAF Firehose =====')
try:
    import boto3
    from botocore.config import Config
    acc = accounts()['aws']
    def boto(svc, region='us-east-1'):
        return boto3.client(svc, region_name=region,
            aws_access_key_id=acc['ak'], aws_secret_access_key=acc['sk'],
            config=Config(retries={'max_attempts': 2}, connect_timeout=8, read_timeout=20))
    fh = boto('firehose')
    streams = fh.list_delivery_streams(DeliveryStreamType='DirectPut').get('DeliveryStreamNames', [])
    for name in streams:
        if 'waf' not in name.lower():
            continue
        ds = fh.describe_delivery_stream(DeliveryStreamName=name)['DeliveryStreamDescription']
        dest = ds['Destinations'][0].get('S3DestinationDescription') or ds['Destinations'][0].get('ExtendedS3DestinationDescription')
        if dest:
            print(f"  {name} -> s3://{dest['BucketARN'].split(':')[-1]}/{dest.get('Prefix','')}")
    # 从第一个 waf 流的桶里抓样本
    s3 = boto('s3', 'eu-central-1')
    for name in streams:
        if 'waf' not in name.lower():
            continue
        ds = fh.describe_delivery_stream(DeliveryStreamName=name)['DeliveryStreamDescription']
        dest = ds['Destinations'][0].get('S3DestinationDescription') or ds['Destinations'][0].get('ExtendedS3DestinationDescription')
        if not dest:
            continue
        bucket = dest['BucketARN'].split(':')[-1]
        prefix = dest.get('Prefix', '')
        objs = []
        for page in boto('s3').get_paginator('list_objects_v2').paginate(Bucket=bucket, Prefix=prefix, MaxKeys=100):
            objs.extend(page.get('Contents', []))
        if not objs:
            print(f"  [{bucket}/{prefix}] 空")
            continue
        objs.sort(key=lambda o: o['LastModified'])
        latest = objs[-1]
        print(f"  [{bucket}] latest: {latest['Key']} @ {latest['LastModified']}")
        body = boto('s3').get_object(Bucket=bucket, Key=latest['Key'])['Body'].read()
        if latest['Key'].endswith('.gz'):
            body = gzip.decompress(body)
        lines = [l for l in body.decode('utf-8', 'replace').strip().splitlines() if l.strip()]
        print(f'    {len(lines)} 行,样本: {lines[0][:1200]}')
        break
except Exception as e:
    print('  失败:', str(e)[:200])

# ---------- 腾讯:30 天窗口重试 + CDN 日志包 API ----------
print('\n===== 腾讯 30 天窗口 =====')
try:
    from tencentcloud.common import credential
    from tencentcloud.cls.v20201016 import cls_client, models as cls_models
    acc = accounts()['tencent']
    cred = credential.Credential(acc['ak'], acc['sk'])
    now = int(time.time())
    for region, topic, desc in [('ap-shanghai', 'a0bfd8ed-d7b1-480a-879b-3c143f7302b8', 'WAF'),
                                ('ap-guangzhou', 'f9a20d7f-11cc-4e3c-bd42-7bff6bb20874', 'EdgeOne')]:
        client = cls_client.ClsClient(cred, region)
        req = cls_models.SearchLogRequest()
        req.TopicId, req.From, req.To, req.Query, req.Limit = topic, now - 30*86400, now, '', 1
        resp = client.SearchLog(req)
        print(f"  {desc}: 30天 {len(resp.Results)} 条" + (' | ' + json.dumps(json.loads(resp.Results[0].Log), ensure_ascii=False)[:800] if resp.Results else ''))
except Exception as e:
    print('  失败:', str(e)[:200])

# ---------- 华为:LTS 日志组/流 ----------
print('\n===== 华为 LTS =====')
try:
    from huaweicloudsdkcore.auth.credentials import BasicCredentials
    from huaweicloudsdklts.v2.region.lts_region import LtsRegion
    from huaweicloudsdklts.v2 import LtsClient, ListLogGroupsRequest
    acc = accounts()['huawei']
    cred = BasicCredentials(acc['ak'], acc['sk'])
    client = LtsClient.new_builder().with_credentials(cred).with_region(LtsRegion.value_of('cn-south-1')).build()
    groups = client.list_log_groups(ListLogGroupsRequest())
    for g in groups.log_groups or []:
        print(f"  group: {g.log_group_name} ({g.log_group_id})")
except Exception as e:
    print('  失败(需要 huaweicloudsdklts 或项目ID):', str(e)[:200])
