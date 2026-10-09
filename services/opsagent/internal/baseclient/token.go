package baseclient

import "context"

// userTokenKey 请求上下文用户令牌挂载键。
// 编排层以 HTTP 契约调用底座（契约 B1 logquery 等）时，底座按 eiam 会话鉴权；
// web 层在进入编排前将调用方令牌（Authorization: Bearer x）挂载进请求上下文，
// 本包在发出底座 HTTP 请求时透传该令牌。告警 worker 链路无用户会话，令牌为空，
// 底座返回 401/403 时按契约降级「无数据」，不中断主链路。
type userTokenKey struct{}

// WithUserToken 挂载用户会话令牌到上下文。
func WithUserToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, userTokenKey{}, token)
}

// userTokenFrom 从上下文取出用户会话令牌（无则空串）。
func userTokenFrom(ctx context.Context) string {
	t, _ := ctx.Value(userTokenKey{}).(string)
	return t
}
