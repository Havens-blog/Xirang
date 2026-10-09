// Package domain 跨模块共享的领域类型（tech-design › Shared Types 唯一定义处）。
package domain

import "context"

// userAttrKey 请求上下文用户属性键。
// web 层在进入编排前把会话用户名挂载进请求上下文；编排/意图层用它为
// correction_log 等审计字段补全操作人（By），不承载鉴权结论（鉴权只认中间件
// 注入的会话主体）。
type userAttrKey struct{}

// WithUserAttr 挂载用户属性（当前仅 username）到上下文。
func WithUserAttr(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, userAttrKey{}, username)
}

// UserAttrFrom 从上下文取出用户属性（无则空串）。
func UserAttrFrom(ctx context.Context) string {
	v, _ := ctx.Value(userAttrKey{}).(string)
	return v
}
