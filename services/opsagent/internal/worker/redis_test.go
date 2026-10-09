package worker

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// fakeCmdable redisCmdable 窄接口的测试实现（免真实 Redis）。
type fakeCmdable struct {
	setNXVal    bool
	incrVal     int64
	expireCalls []string
}

func (f *fakeCmdable) SetNX(ctx context.Context, key string, _ any, _ time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	cmd.SetVal(f.setNXVal)
	return cmd
}

func (f *fakeCmdable) Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(f.incrVal)
	return cmd
}

func (f *fakeCmdable) Expire(ctx context.Context, key string, _ time.Duration) *redis.BoolCmd {
	f.expireCalls = append(f.expireCalls, key)
	cmd := redis.NewBoolCmd(ctx)
	cmd.SetVal(true)
	return cmd
}

// TestRedisDeduper_TryAcquire SET NX 占位：首次 true / 重复 false。
func TestRedisDeduper_TryAcquire(t *testing.T) {
	d := NewRedisDeduper(&fakeCmdable{setNXVal: true})
	ok, err := d.TryAcquire(context.Background(), "fp", time.Minute)
	if err != nil || !ok {
		t.Fatalf("first acquire should be true, got ok=%v err=%v", ok, err)
	}

	d2 := NewRedisDeduper(&fakeCmdable{setNXVal: false})
	ok, err = d2.TryAcquire(context.Background(), "fp", time.Minute)
	if err != nil || ok {
		t.Fatalf("dup acquire should be false, got ok=%v err=%v", ok, err)
	}
}

// TestRedisBudget_Incr 首增设过期 / 非首增不设过期。
func TestRedisBudget_Incr(t *testing.T) {
	f := &fakeCmdable{incrVal: 1}
	b := NewRedisBudget(f)
	n, err := b.Incr(context.Background(), "k", time.Hour)
	if err != nil || n != 1 {
		t.Fatalf("incr = %d, err = %v", n, err)
	}
	if len(f.expireCalls) != 1 {
		t.Fatalf("first incr should set expire, got %d calls", len(f.expireCalls))
	}

	f2 := &fakeCmdable{incrVal: 3}
	b2 := NewRedisBudget(f2)
	if _, err := b2.Incr(context.Background(), "k", time.Hour); err != nil {
		t.Fatalf("incr err: %v", err)
	}
	if len(f2.expireCalls) != 0 {
		t.Fatalf("non-first incr should not set expire, got %d calls", len(f2.expireCalls))
	}
}
