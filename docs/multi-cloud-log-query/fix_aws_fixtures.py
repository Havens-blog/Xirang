# -*- coding: utf-8 -*-
"""修正 AWS fixture:CloudFront 拿头+数据行;WAF 找真实 WAF JSON 记录(非 Firehose 错误日志)。"""
import io, sys, json, time, gzip, os
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts
import boto3
from botocore.config import Config

OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'testdata')
now = int(time.time())
acc = accounts()['aws']
s3 = boto3.client('s3', region_name='eu-west-1',
                  aws_access_key_id=acc['ak'], aws_secret_access_key=acc['sk'],
                  config=Config(retries={'max_attempts': 2}, connect_timeout=8, read_timeout=30))

def latest_objects(bucket, prefix, n=20):
    objs = []
    paginator = s3.get_paginator('list_objects_v2')
    for page in paginator.paginate(Bucket=bucket, Prefix=prefix, MaxKeys=500):
        objs.extend(page.get('Contents', []))
        if len(objs) >= 500:
            break
    objs.sort(key=lambda o: o['LastModified'])
    return objs[-n:]

# ---- CloudFront 标准日志:头两行(#Version/#Fields)+ 数据行 ----
print('== CloudFront 标准日志 ==')
try:
    objs = latest_objects('jlc-prod-cdn-log', '')
    # 找一个 .gz 且有真实数据的
    for o in reversed(objs):
        body = s3.get_object(Bucket='jlc-prod-cdn-log', Key=o['Key'])['Body'].read()
        if o['Key'].endswith('.gz'):
            body = gzip.decompress(body)
        lines = body.decode('utf-8', 'replace').strip().splitlines()
        data_lines = [l for l in lines if not l.startswith('#')]
        if len(lines) >= 2 and data_lines:
            header = [l for l in lines if l.startswith('#')]
            doc = {
                'meta': {'cloud': 'aws', 'log_type': 'cloudfront', 'desc': 'CloudFront 标准日志(TSV)',
                         'bucket': 'jlc-prod-cdn-log', 'key': o['Key'], 'format': 'tsv',
                         'header_lines': header, 'object_time': str(o['LastModified']), 'pulled_at': now},
                'sample_count': len(data_lines[:3]),
                'samples': [{'_raw_line': l} for l in data_lines[:3]],
            }
            with open(os.path.join(OUT, 'aws-cloudfront.json'), 'w', encoding='utf-8') as f:
                json.dump(doc, f, ensure_ascii=False, indent=2)
            print(f"  -> aws-cloudfront.json key={o['Key']}")
            print('  header:', header[1][:200] if len(header) > 1 else header)
            print('  data  :', data_lines[0][:200])
            break
except Exception as e:
    print('  失败:', str(e)[:200])

# ---- AWS WAF:遍历最新对象找真实 WAF 记录(含 webaclId/timestamp) ----
print('\n== AWS WAF ==')
for bucket in ['waf-loghub-clloggingbucket5f34e4eb-hjalewqix1db',
               'loghub-clloggingbucket5f34e4eb-asng9spoysri']:
    try:
        objs = latest_objects(bucket, 'AWSLogs/', 40)
        found = None
        for o in reversed(objs):
            if not o['Key'].endswith('.gz'):
                continue
            body = gzip.decompress(s3.get_object(Bucket=bucket, Key=o['Key'])['Body'].read())
            for l in body.decode('utf-8', 'replace').strip().splitlines():
                try:
                    rec = json.loads(l)
                except Exception:
                    continue
                if 'webaclId' in rec and 'httpRequest' in rec:
                    found = (o, rec)
                    break
            if found:
                break
        if found:
            o, rec = found
            # 再收集同文件更多样本
            body = gzip.decompress(s3.get_object(Bucket=bucket, Key=o['Key'])['Body'].read())
            samples = []
            for l in body.decode('utf-8', 'replace').strip().splitlines():
                try:
                    r = json.loads(l)
                    if 'webaclId' in r:
                        samples.append(r)
                        if len(samples) >= 3:
                            break
                except Exception:
                    pass
            doc = {
                'meta': {'cloud': 'aws', 'log_type': 'waf-cloudfront', 'desc': 'AWS WAFv2 CLOUDFRONT scope (Firehose->S3 JSON)',
                         'bucket': bucket, 'key': o['Key'], 'format': 'json-lines',
                         'object_time': str(o['LastModified']), 'pulled_at': now},
                'sample_count': len(samples), 'samples': samples,
            }
            with open(os.path.join(OUT, 'aws-waf-cloudfront.json'), 'w', encoding='utf-8') as f:
                json.dump(doc, f, ensure_ascii=False, indent=2)
            print(f"  -> aws-waf-cloudfront.json bucket={bucket} key={o['Key']} ({len(samples)} samples)")
            print('  keys:', list(samples[0].keys()))
            break
        else:
            print(f'  [{bucket}] 40 个最新对象中无 WAF 记录')
    except Exception as e:
        print(f'  [{bucket}] 失败:', str(e)[:200])
