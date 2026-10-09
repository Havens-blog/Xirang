# -*- coding: utf-8 -*-
"""Phase 0.3:腾讯 CLS 关键 topic 拉样本。"""
import io, sys, json, time
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts
from tencentcloud.common import credential
from tencentcloud.cls.v20201016 import cls_client, models as cls_models
from tencentcloud.common.exception.tencent_cloud_sdk_exception import TencentCloudSDKException

acc = accounts()['tencent']
cred = credential.Credential(acc['ak'], acc['sk'])
now = int(time.time())

TARGETS = [
    ('ap-shanghai', 'a0bfd8ed-d7b1-480a-879b-3c143f7302b8', 'WAF 访问日志'),
    ('ap-guangzhou', 'f9a20d7f-11cc-4e3c-bd42-7bff6bb20874', 'EdgeOne 日志'),
]
for region, topic, desc in TARGETS:
    client = cls_client.ClsClient(cred, region)
    for window, wname in ((3600, '1h'), (7 * 86400, '7d')):
        try:
            req = cls_models.SearchLogRequest()
            req.TopicId = topic
            req.From, req.To = now - window, now
            req.Query = ''
            req.Limit = 2
            resp = client.SearchLog(req)
            if resp.Results:
                print(f'\n===== {desc} | {region} | 窗口{wname} =====')
                for r in resp.Results:
                    print(json.dumps(json.loads(r.Log), ensure_ascii=False)[:1500])
                break
            else:
                print(f'===== {desc} | {region} | {wname} 无数据 =====')
        except TencentCloudSDKException as e:
            print(f'===== {desc} | {region} | {wname} 错误: {e.code} {e.message} =====')
            break
