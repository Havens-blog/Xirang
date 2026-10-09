# -*- coding: utf-8 -*-
"""Phase 0.3 深挖:AWS 日志桶内容取样 + CloudFront 日志配置 + 多区域 ELB。"""
import io, sys, json, gzip
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts
import boto3
from botocore.config import Config

acc = accounts()['aws']
def boto(service, region):
    return boto3.client(service, region_name=region,
        aws_access_key_id=acc['ak'], aws_secret_access_key=acc['sk'],
        config=Config(retries={'max_attempts': 2}, connect_timeout=8, read_timeout=20))

# 1. CloudFront:逐个 get_distribution 拿日志配置
print('== CloudFront distributions(逐个取配置)==')
cf = boto('cloudfront', 'us-east-1')
ids = [d['Id'] for d in cf.list_distributions().get('DistributionList', {}).get('Items', [])]
for did in ids:
    d = cf.get_distribution(Id=did)['Distribution']
    cfg, dom = d['DistributionConfig'], d['DomainName']
    aliases = cfg.get('Aliases', {}).get('Items') or ['-']
    lg = cfg.get('Logging', {})
    # 2025 新版可能用 Origins/DefaultCacheBehavior 之外的字段;兼容无 Logging 键
    status = f"ON->{lg.get('Bucket','')}/{lg.get('Prefix','')}" if lg.get('Enabled') else 'OFF/新版默认'
    print(f"  {did} {dom} aliases={aliases} logging={status}")

# 2. 多区域 ELB
print('\n== ELBv2 across regions ==')
for region in ['eu-west-1', 'eu-central-1', 'eu-south-1', 'us-east-1']:
    try:
        elb = boto('elbv2', region)
        lbs = elb.describe_load_balancers().get('LoadBalancers', [])
        for lb in lbs:
            attrs = elb.describe_load_balancer_attributes(LoadBalancerArn=lb['LoadBalancerArn'])['Attributes']
            al = next((a['Value'] for a in attrs if a['Key'] == 'access_logs.s3.enabled'), '?')
            bucket = next((a['Value'] for a in attrs if a['Key'] == 'access_logs.s3.bucket'), '-')
            print(f"  [{region}] {lb['LoadBalancerName']} ({lb['Type']}) access_log={al} bucket={bucket}")
        if not lbs:
            print(f'  [{region}] 无 LB')
    except Exception as e:
        print(f'  [{region}] 失败: {str(e)[:100]}')

# 3. 日志桶最新对象 + 取样
print('\n== 日志桶内容 ==')
s3 = boto('s3', 'eu-west-1')
for bucket in ['aws-lb-logs-jlc', 'aws-waf-logs-jlc', 'jlc-prod-cdn-log', 'prod-jlc-cloudfront-log']:
    try:
        # 列最近对象(按字典序倒序拿不到时间序,先拿首页再排序)
        objs = []
        paginator = s3.get_paginator('list_objects_v2')
        for page in paginator.paginate(Bucket=bucket, MaxKeys=200):
            objs.extend(page.get('Contents', []))
            if len(objs) >= 200:
                break
        if not objs:
            print(f'  [{bucket}] 空')
            continue
        objs.sort(key=lambda o: o['LastModified'])
        latest = objs[-1]
        print(f"  [{bucket}] {len(objs)}+ objects, latest: {latest['Key']} @ {latest['LastModified']} ({latest['Size']}B)")
        # 下载并解一行样本
        key = latest['Key']
        body = s3.get_object(Bucket=bucket, Key=key)['Body'].read()
        if key.endswith('.gz'):
            body = gzip.decompress(body)
        lines = body.decode('utf-8', 'replace').strip().splitlines()
        print(f'    共 {len(lines)} 行,样本: {lines[0][:900]}')
    except Exception as e:
        print(f'  [{bucket}] 失败: {str(e)[:150]}')
