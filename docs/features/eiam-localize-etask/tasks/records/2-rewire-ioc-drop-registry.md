---
status: "completed"
started: "2026-10-09 11:50"
completed: "2026-10-09 11:56"
time_spent: "~6m"
---

# Task Record: 2 接线替换（ioc/grpc.go + app.go + 删 registry.go）并重跑 wire

## Summary
Rewired ioc layer from etask grpc/registry to local internal/grpcx: dropped registry parameter from InitGrpcServer, changed App.Server to *grpcx.Server, deleted ioc/registry.go, removed InitRegistry from wire provider set, and regenerated wire_gen.go. All etask imports now zero; go build ./... passes; InitEtcd preserved for capability/discovery usage.

## Changes

### Files Created
无

### Files Modified
- eiam/ioc/grpc.go
- eiam/ioc/app.go
- eiam/ioc/wire.go
- eiam/ioc/wire_gen.go

### Key Decisions
- Used wire ./ioc to regenerate wire_gen.go (tool was available), which removed InitRegistry call and dropped the registry parameter from InitGrpcServer automatically
- Kept InitEtcd in BaseSet — still required by InitCapabilityRegistry and internal/service/discovery worker
- Preserved the three proto service registrations unchanged since grpcx.Server embeds *grpc.Server satisfying grpc.ServiceRegistrar

## Test Results
- **Tests Executed**: Yes
- **Passed**: 5
- **Failed**: 0
- **Coverage**: 46.3%

## Acceptance Criteria
- [x] ioc/grpc.go uses grpcx.NewServer(cfg, grpcx.WithJWTAuth(cfg.AuthToken)), drops registry param
- [x] ioc/app.go Server field changed to *grpcx.Server
- [x] ioc/registry.go (InitRegistry) deleted
- [x] ioc/wire.go removed InitRegistry from provider set; wire_gen.go regenerated with InitGrpcServer dropping registry arg
- [x] go build ./... passes
- [x] grep -rn Duke1616/etask eiam/ --include=*.go returns 0 results

## Notes
ioc/ has no test files (wire-generated code). Verification evidence: go build ./... exits 0; go test ./internal/grpcx/... passes 5 JWT tests (the auth path InitGrpcServer now wires via grpcx.WithJWTAuth, coverage 46.3%); grep -rn 'Duke1616/etask' eiam/ --include=*.go returns 0. wire_gen.go regenerated cleanly by the wire tool.
