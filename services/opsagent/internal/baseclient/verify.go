package baseclient

import (
	"context"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// TokenVerifier eiam 本地令牌校验抽象（契约 B3：共享密钥 HMAC + 共享 Redis 会话本地
// 校验，非 OIDC introspection / 非 HTTP 验证端点）。实际实现复用了 ginx session.Provider
// （与 internal/middleware 同源），由装配层注入。
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (domain.Identity, error)
}

const eiamModule = "eiam"

// VerifyToken 校验 eiam 令牌，返回 Identity（契约 B3）。令牌无效/过期 → 上游映射 401；
// 未注入 TokenVerifier（未配置）→ 返回类型化 ModuleError，不抛裸错误。
func (c *Client) VerifyToken(ctx context.Context, token string) (domain.Identity, error) {
	if c.verify == nil {
		return domain.Identity{}, &ModuleError{Module: eiamModule, Msg: "token verifier not configured"}
	}
	return c.verify.Verify(ctx, token)
}
