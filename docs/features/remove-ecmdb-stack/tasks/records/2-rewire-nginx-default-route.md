---
status: "completed"
started: "2026-10-08 21:25"
completed: "2026-10-08 21:31"
time_spent: "~6m"
---

# Task Record: 2 nginx 摘除 /api/cmdb 路由并把默认路由 / 改为 301 /console/

## Summary
Removed /api/cmdb -> ecmdb:8000 route from nginx.dev.conf and rewired the default `location /` (previously proxying to ecmdb-web:3333) to `return 301 /console/`, redirecting the :8888 root entry to the unified haven-console entry. Verified via `nginx -t` (config syntax OK), `nginx -s reload`, and curl checks: :8888/ returns 301 Location /console/; /api/cmdb no longer routes to a dead upstream (falls through to 301); /console/ returns 200; /cam/, /api/iam/, /api/v1/cam/ routes remain intact (502 = upstream dev services not running in executor env, not a routing regression).

## Changes

### Files Created
无

### Files Modified
- nginx-1.25.4/conf/vhost/nginx.dev.conf

### Key Decisions
- Kept /api/v1/cmdb/ route (points to e-cam-service :8001, not the removed ecmdb :8000) — only the /api/cmdb -> 127.0.0.1:8000 route was removed.
- Replaced the ecmdb-web :3333 proxy_pass default route with `return 301 /console/` per proposal decision.

## Test Results
- **Tests Executed**: Yes
- **Passed**: 1
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] nginx.dev.conf 无 /api/cmdb 路由、无对 127.0.0.1:8000 的 proxy_pass
- [x] :8888/ 返回 301 且 Location 为 /console/
- [x] /console/、/cam/、/api/iam/、/api/v1/* 路由不受影响

## Notes
Test evidence: `nginx -t` -> 'syntax is ok / test is successful' (1 test, passed). AC verification via curl after `nginx -s reload`: [AC-1] grep for /api/cmdb and 127.0.0.1:8000 in nginx.dev.conf = no matches; curl :8888/api/cmdb -> 301 (fallthrough, not dead-upstream 502). [AC-2] curl -sI :8888/ -> HTTP/1.1 301 Moved Permanently, Location: http://localhost:8888/console/. [AC-3] curl :8888/console/ -> 200; :8888/cam/ -> 502; :8888/api/iam/ -> 502; :8888/api/v1/cam/ -> 502 (routes intact — 502 = upstream dev services not running in this executor environment, confirms nginx still matches and proxies each route; a removed route would fall to `location /` and return 301, not 502). No coverage percentage: nginx vhost configs have no coverage runner; coverage omitted rather than guessed per record-format rules.
