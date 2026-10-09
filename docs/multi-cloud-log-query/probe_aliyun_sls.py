# -*- coding: utf-8 -*-
"""Phase 0.2 探测:阿里云 SLS 各区域 Project/Logstore 清单,识别 WAF/SLB 日志投递。"""
import io, sys
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
from cam_creds import accounts
acc = accounts()['aliyun']
ak, sk = acc['ak'], acc['sk']
print('aliyun account:', acc['name'], 'ak:', ak[:8] + '...')

from aliyun.log import LogClient, ListLogstoresRequest
from aliyun.log.logexception import LogException

REGIONS = ['cn-shenzhen', 'cn-wulanchabu', 'cn-guangzhou', 'eu-central-1', 'ap-southeast-1']
for r in REGIONS:
    endpoint = f'{r}.log.aliyuncs.com'
    client = LogClient(endpoint, ak, sk)
    try:
        resp = client.list_project()
        projects = [p['projectName'] if isinstance(p, dict) else p.get_project_name() for p in resp.get_projects()]
        print(f'\n== {r}: {len(projects)} projects ==')
        for p in projects:
            try:
                ls = client.list_logstores(ListLogstoresRequest(p))
                stores = ls.get_logstores()
                print(f'  {p}: {stores}')
            except LogException as e:
                print(f'  {p}: <list_logstores 失败 {e.get_error_code()}>')
    except LogException as e:
        print(f'\n== {r}: list_project 失败 {e.get_error_code()} {e.get_error_message()} ==')
