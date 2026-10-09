---
id: "2"
title: "nginx 摘除 /api/cmdb 路由并把默认路由 / 改为 301 /console/"
priority: "P1"
estimated_time: "0.5h"
complexity: "low"
dependencies: []
surface-key: ""
surface-type: ""
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 2: nginx 摘除 /api/cmdb 路由并把默认路由 / 改为 301 /console/

## Description
从 `nginx-1.25.4/conf/vhost/nginx.dev.conf` 删除 `/api/cmdb → ecmdb:8000` 路由；把默认路由 `location /`（原 `proxy_pass` 到 ecmdb-web:3333）改为 `return 301 /console/`，让 `:8888` 根入口落到新统一入口 haven-console。

## Reference Files
- `docs/proposals/remove-ecmdb-stack/proposal.md` — 根入口重定向决策与成功标准（ref: Proposed Solution / Success Criteria）
- `nginx-1.25.4/conf/vhost/nginx.dev.conf` — 删 `/api/cmdb` location；`location /` 改 301 `/console/`（ref: Scope）

## Acceptance Criteria
- [ ] nginx.dev.conf 无 `/api/cmdb` 路由、无对 `127.0.0.1:8000` 的 proxy_pass
- [ ] `:8888/` 返回 301 且 Location 为 `/console/`
- [ ] `/console/`、`/cam/`、`/api/iam/`、`/api/v1/*` 路由不受影响（curl 逐一验证）

## Implementation Notes
- 目标路由：`nginx.dev.conf` 第 20-25 行（`/api/cmdb`）、第 84-90 行（`location /` 默认路由）。
- 修改后执行 `nginx -s reload`（`.\dev.ps1 restart nginx` 或直接 `nginx.exe -s reload`）再验证。
- 验证命令：`curl -sI http://localhost:8888/` 看 301；`curl -s -o /dev/null -w '%{http_code}' http://localhost:8888/console/` 看 200。