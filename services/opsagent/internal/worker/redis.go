package worker

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisCmdable Redis 命令窄子集（*redis.Client 结构满足；测试 fake 仅需 3 方法）。
// 只用 SET NX / INCR / EXPIRE，即指纹去重与 LLM 预算两类触发信号，不承担数据存储。
type redisCmdable interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

// RedisDeduper 指纹去重 Redis 实现（SET NX + TTL：首次返回 true，窗口内重复返回 false）。
type RedisDeduper struct {
	rdb redisCmdable
}

// NewRedisDeduper 构造。
func NewRedisDeduper(rdb redisCmdable) *RedisDeduper {
	return &RedisDeduper{rdb: rdb}
}

// TryAcquire 抢占去重窗口。
func (d *RedisDeduper) TryAcquire(ctx context.Context, fingerprint string, ttl time.Duration) (bool, error) {
	return d.rdb.SetNX(ctx, fingerprint, 1, ttl).Result()
}

// RedisBudget LLM 滚动 1h 预算 Redis 实现（INCR + 首次建键 EXPIRE）。
type RedisBudget struct {
	rdb redisCmdable
}

// NewRedisBudget 构造。
func NewRedisBudget(rdb redisCmdable) *RedisBudget {
	return &RedisBudget{rdb: rdb}
}

// Incr 递增预算计数并返回累计值；首次建键（n==1）时按 ttl 设置过期，维持整小时窗口。
func (b *RedisBudget) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	n, err := b.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_ = b.rdb.Expire(ctx, key, ttl)
	}
	return n, nil
}
