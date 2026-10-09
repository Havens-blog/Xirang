# -*- coding: utf-8 -*-
"""Phase 0.2/0.3:腾讯云 CLS 日志集/主题探测 + 样本。"""
import io, sys, json, time
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts
from tencentcloud.common import credential
from tencentcloud.cls.v20201016 import cls_client, models as cls_models
from tencentcloud.common.exception.tencent_cloud_sdk_exception import TencentCloudSDKException

acc = accounts()['tencent']
cred = credential.Credential(acc['ak'], acc['sk'])
print('tencent account:', acc['name'])

# 腾讯常用地域:ap-guangzhou(账号主地域)+ ap-shanghai/beijing/chengdu 常见 WAF 落点
for region in ['ap-guangzhou', 'ap-shanghai', 'ap-beijing']:
    client = cls_client.ClsClient(cred, region)
    try:
        # 日志集清单
        sets = client.DescribeLogsets(cls_models.DescribeLogsetsRequest()) if hasattr(cls_models, 'DescribeLogsetsRequest') else None
    except TencentCloudSDKException as e:
        print(f'{region}: DescribeLogsets 不支持/失败: {e.code}')
        sets = None
    try:
        req = cls_models.DescribeTopicsRequest()
        req.Offset, req.Limit = 0, 100
        topics = client.DescribeTopics(req)
        print(f'\n== {region}: {topics.TotalCount} topics ==')
        for t in topics.Topics:
            print(f"  topic={t.TopicId} [{t.TopicName}] logset={t.LogsetId}")
    except TencentCloudSDKException as e:
        print(f'{region}: DescribeTopics 失败: {e.code} {e.message}')
