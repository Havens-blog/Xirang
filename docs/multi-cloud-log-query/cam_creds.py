# -*- coding: utf-8 -*-
"""共用:从 Mongo 读账号并用 e-cam-service 同款 AES-GCM 解密凭证。"""
import base64
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from pymongo import MongoClient

# 与 e-cam-service config/prod.yaml security.encryption_key 一致(仅本地探测用,勿提交)
KEY = b'48CpLbg+M(HhZa~R'

def decrypt(ciphertext_b64: str) -> str:
    data = base64.b64decode(ciphertext_b64)
    nonce, body = data[:12], data[12:]
    return AESGCM(KEY).decrypt(nonce, body, None).decode()

def accounts():
    c = MongoClient('mongodb://ecmdb:123456@106.52.187.69:27017/ecam?authSource=admin',
                    serverSelectionTimeoutMS=8000)
    out = {}
    for acc in c['ecam']['ecam_cloud_account'].find({}):
        out[acc['provider']] = {
            'id': acc['id'], 'name': acc['name'],
            'ak': acc['access_key_id'], 'sk': decrypt(acc['access_key_secret']),
            'regions': acc.get('regions', []),
        }
    return out

if __name__ == '__main__':
    import io, sys
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
    for p, a in accounts().items():
        print(p, a['name'], 'ak=' + a['ak'][:8] + '... sk_len=' + str(len(a['sk'])))
