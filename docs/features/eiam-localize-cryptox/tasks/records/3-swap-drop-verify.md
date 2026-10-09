---
status: "completed"
started: "2026-10-09 10:01"
completed: "2026-10-09 10:07"
time_spent: "~6m"
---

# Task Record: 3 切换 5 件 import、删除 go.mod ecmdb 依赖并全量验证

## Summary
Swapped 5 files' import from github.com/Duke1616/ecmdb/pkg/cryptox to github.com/Duke1616/eiam/internal/cryptox (business logic unchanged), removed github.com/Duke1616/ecmdb from go.mod and ran go mod tidy so go.sum dropped it. Verified: grep ecmdb/pkg/cryptox returns 0; go build ./..., go vet ./... pass; go test ./internal/cryptox/... passes. wire_gen.go does not import cryptox directly, no changes needed.

## Changes

### Files Created
无

### Files Modified
- eiam/ioc/crypto.go
- eiam/ioc/identity_source.go
- eiam/internal/service/user/service.go
- eiam/internal/service/identity_source/service.go
- eiam/cmd/migrate/internal/migrations/user.go
- eiam/go.mod
- eiam/go.sum

### Key Decisions
- Only changed import paths, no business-logic edits (Register/WithLegacyAlgo/migration logic untouched)
- go mod tidy cleaned go.sum automatically after removing ecmdb require line
- ioc/wire_gen.go confirmed to not import cryptox directly, no wire regen needed

## Test Results
- **Tests Executed**: Yes
- **Passed**: 1
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 5 files import switched to eiam/internal/cryptox; grep ecmdb/pkg/cryptox returns 0
- [x] go.mod drops github.com/Duke1616/ecmdb; go mod tidy leaves go.sum without ecmdb and no errors
- [x] go build ./... and go vet ./... pass with no new warnings
- [x] go test ./internal/cryptox/... all pass

## Notes
eiam is an independent Go module at D:/Haven/eiam; all go commands run from there. Concurrent session left unrelated modified files (internal/web/tenant/handler.go, internal/web/user/user.go, ioc/session.go, internal/web/sessionclaims/) which were intentionally NOT staged in this commit.
