# -*- coding: utf-8 -*-
"""Phase 0.2:AWS 日志投递状态探测(eu-west-1):S3 日志桶 / CloudFront / ALB / WAF。"""
import io, sys, json
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts
import boto3
from botocore.config import Config

acc = accounts()['aws']
boto = lambda service, region='eu-west-1': boto3.client(
    service, region_name=region,
    aws_access_key_id=acc['ak'], aws_secret_access_key=acc['sk'],
    config=Config(retries={'max_attempts': 2}, connect_timeout=8, read_timeout=15))

print('aws account:', acc['name'])

# 1. S3 桶清单(找日志类桶)
s3 = boto('s3')
print('\n== S3 buckets ==')
for b in s3.list_buckets()['Buckets']:
    print(' ', b['Name'])

# 2. CloudFront 分发(全局服务,us-east-1 入口)及其日志配置
print('\n== CloudFront distributions ==')
try:
    cf = boto('cloudfront', 'us-east-1')
    resp = cf.list_distributions()
    for d in resp.get('DistributionList', {}).get('Items', []):
        cfg = d['DistributionConfig']
        lg = cfg.get('Logging', {})
        print(f"  {d['Id']} {d['DomainName']} ({cfg.get('Aliases',{}).get('Items') or ['no-alias']}) "
              f"logging={'ON -> ' + lg.get('Bucket','') if lg.get('Enabled') else 'OFF'}")
except Exception as e:
    print('  CloudFront 失败:', str(e)[:150])

# 3. ALB/NLB 与访问日志配置
print('\n== ELBv2 load balancers ==')
try:
    elb = boto('elbv2')
    paginator = elb.get_paginator('describe_load_balancers')
    for page in paginator.paginate():
        for lb in page['LoadBalancers']:
            attrs = elb.describe_load_balancer_attributes(LoadBalancerArn=lb['LoadBalancerArn'])['Attributes']
            al = next((a for a in attrs if a['Key'] == 'access_logs.s3.enabled'), {})
            bucket = next((a for a in attrs if a['Key'] == 'access_logs.s3.bucket'), {})
            print(f"  {lb['LoadBalancerName']} ({lb['Type']}, {lb['Scheme']}) "
                  f"access_log={al.get('Value')} bucket={bucket.get('Value','') or '-'}")
except Exception as e:
    print('  ELBv2 失败:', str(e)[:150])

# 4. WAFv2(WebACL + 日志配置;CLOUDFRONT scope 在 us-east-1)
print('\n== WAFv2 web ACLs ==')
for scope, region in (('REGIONAL', 'eu-west-1'), ('CLOUDFRONT', 'us-east-1')):
    try:
        waf = boto('wafv2', region)
        resp = waf.list_web_acls(Scope=scope, Limit=100)
        for acl in resp.get('WebACLs', []):
            try:
                lg = waf.get_logging_configuration(ResourceArn=acl['ARN'])['LoggingConfiguration']
                dest = lg.get('LogDestinationConfigs', [])
                print(f"  [{scope}] {acl['Name']} logging=ON -> {dest}")
            except waf.exceptions.WAFNonexistentItemException:
                print(f"  [{scope}] {acl['Name']} logging=OFF")
    except Exception as e:
        print(f'  [{scope}] 失败:', str(e)[:150])
