---
status: "completed"
started: "2026-10-09 11:38"
completed: "2026-10-09 11:46"
time_spent: "~8m"
---

# Task Record: 1 新建 internal/grpcx（Server + JWT 拦截器 + 单测）

## Summary
新建 eiam/internal/grpcx 包：Server（内嵌 *grpc.Server 实现 ego server.Server 接口，无 etcd 注册）+ ServerConfig（Name/ListenAddr/AuthToken mapstructure 标签）+ NewServer/WithJWTAuth 选项 + 从 etask interceptors/jwt 逐字移植的 HS256 JWT 鉴权一元拦截器（Bearer 前缀、token.Valid 校验、claims 注入 via Set/Get 上下文辅助）。仅依赖 golang-jwt/jwt/v4 与 google.golang.org/grpc（拦截器部分），不引入 etcd/netx/bizid/tenant。

## Changes

### Files Created
- eiam/internal/grpcx/server.go
- eiam/internal/grpcx/jwt.go
- eiam/internal/grpcx/context.go
- eiam/internal/grpcx/jwt_test.go

### Files Modified
无

### Key Decisions
- ServerConfig 裁掉 etask 的 ServiceId/AdvertiseAddr/ServiceName，按任务要求只保留 Name/ListenAddr/AuthToken，mapstructure 标签对应 name/listen_addr/auth_token
- NewServer 当 AuthToken 非空时自动挂载 JWT 一元拦截器，AuthToken 为空则裸 grpc.Server（与 etask buildInterceptors 行为等价但去除 bizid/tenant 管道）
- Start 用 net.Listen + 异步 grpc.Server.Serve，Info().Healthy 以 listener!=nil 判定；GracefulStop 直接调 grpc.Server.GracefulStop，无 registry 注销逻辑
- JWT 拦截器 Decode/JwtAuthInterceptor 逐字移植 etask：HMAC 算法校验、Bearer 前缀 strip、token.Valid、ErrTokenExpired→中文「授权令牌已过期」、ErrTokenSignatureInvalid→「授权令牌签名无效」、其余→Unauthenticated
- Set/Get 上下文辅助改为注入完整 MapClaims（local claimsKey），不依赖 eiam/pkg/ctxutil，以满足任务「拦截器仅依赖 jwt+grpc」的硬约束；eiam 的 gRPC handler 按请求字段取数不依赖注入的强类型 ID
- Encode 一并移植用于 round-trip 单测

## Test Results
- **Tests Executed**: Yes
- **Passed**: 5
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] 新建 eiam/internal/grpcx：ServerConfig（Name/ListenAddr/AuthToken）、Server（内嵌 *grpc.Server）、NewServer(cfg, opts...)、WithJWTAuth 选项
- [x] Server 完整实现 ego server.Server 接口（Name/Init/Start/Stop/GracefulStop/PackageName/Info），Start 用 net.Listen + 异步 grpc.Server.Serve，无 etcd 注册逻辑
- [x] JWT 鉴权拦截器移植自 etask interceptors/jwt（HS256、Bearer 前缀、claims 注入、Set/Get 上下文辅助），仅依赖 golang-jwt/jwt/v4 与 google.golang.org/grpc
- [x] 单测覆盖 JWT 四路：有效 token 放行 / 无 token 拒绝 / 过期拒绝 / 签名错拒绝，go test ./internal/grpcx/... 通过

## Notes
go build ./internal/grpcx/... 与 go vet ./internal/grpcx/... 通过，gofmt 干净（goimports 本机未安装）。未修改任何 grpcx 之外的包（Hard Rule 遵守）。eiam 工作树有来自并发会话的外来脏文件（internal/web/tenant/handler.go、internal/web/user/user.go、ioc/session.go、internal/web/sessionclaims/），与本任务无关，提交时仅显式 pathspec eiam/internal/grpcx/。jwt/v4 仍为 indirect，按计划在任务 3 转 direct。
