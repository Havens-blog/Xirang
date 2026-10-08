package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/repository/dao"
)

// ---- fakes ----

type fakeCompStore struct {
	items   []dao.Compensation
	listErr error

	deleted   []string
	marked    []string
	nextRetry []time.Time
}

func (f *fakeCompStore) ListDue(_ context.Context, _ time.Time, _ int) ([]dao.Compensation, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.items, nil
}

func (f *fakeCompStore) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	// 模拟真实 DAO：删除后集合不再返回该项。
	for i, it := range f.items {
		if it.ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			break
		}
	}
	return nil
}

func (f *fakeCompStore) MarkRetried(_ context.Context, id string, nextRetryAt time.Time) error {
	f.marked = append(f.marked, id)
	f.nextRetry = append(f.nextRetry, nextRetryAt)
	return nil
}

type fakeReplayer struct {
	err      error
	replayed []string
}

func (f *fakeReplayer) Replay(_ context.Context, c dao.Compensation) error {
	f.replayed = append(f.replayed, c.ID)
	return f.err
}

func comp(id string, retryCount int) dao.Compensation {
	return dao.Compensation{ID: id, RetryCount: retryCount, Payload: map[string]any{"diagnosis_id": id}}
}

// ---- tests ----

// TestCompensationLoop_SuccessDeletes 重放成功 → 删除补偿项（不回退、不留存）。
func TestCompensationLoop_SuccessDeletes(t *testing.T) {
	store := &fakeCompStore{items: []dao.Compensation{comp("c1", 0)}}
	replay := &fakeReplayer{}
	var dead []string
	l := NewCompensationLoop(store, replay, CompensationConfig{}, func(c dao.Compensation) { dead = append(dead, c.ID) })

	l.Tick(context.Background())

	if len(replay.replayed) != 1 || replay.replayed[0] != "c1" {
		t.Fatalf("replay = %v", replay.replayed)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "c1" {
		t.Fatalf("deleted = %v", store.deleted)
	}
	if len(store.marked) != 0 {
		t.Fatalf("success should not mark retried: %v", store.marked)
	}
	if len(dead) != 0 {
		t.Fatalf("success should not alert dead: %v", dead)
	}
}

// TestCompensationLoop_RetryNotDead 重放失败且未达上限 → MarkRetried、不告警。
func TestCompensationLoop_RetryNotDead(t *testing.T) {
	store := &fakeCompStore{items: []dao.Compensation{comp("c1", 2)}}
	replay := &fakeReplayer{err: errors.New("replay failed")}
	var dead []string
	l := NewCompensationLoop(store, replay, CompensationConfig{}, func(c dao.Compensation) { dead = append(dead, c.ID) })

	l.Tick(context.Background())

	if len(store.marked) != 1 || store.marked[0] != "c1" {
		t.Fatalf("marked = %v", store.marked)
	}
	if len(store.deleted) != 0 {
		t.Fatalf("failed replay should not delete: %v", store.deleted)
	}
	if len(dead) != 0 {
		t.Fatalf("retry < limit should not alert dead: %v", dead)
	}
}

// TestCompensationLoop_DeadAlerts 重试达上限（第 5 次失败）→ MarkRetried + dead 告警。
func TestCompensationLoop_DeadAlerts(t *testing.T) {
	store := &fakeCompStore{items: []dao.Compensation{comp("c1", 4)}}
	replay := &fakeReplayer{err: errors.New("replay failed")}
	var dead []string
	l := NewCompensationLoop(store, replay, CompensationConfig{}, func(c dao.Compensation) { dead = append(dead, c.ID) })

	l.Tick(context.Background())

	if len(store.marked) != 1 {
		t.Fatalf("should mark retried: %v", store.marked)
	}
	if len(dead) != 1 || dead[0] != "c1" {
		t.Fatalf("5th failure should alert dead, got %v", dead)
	}
}

// TestCompensationLoop_ListErrorNoCrash ListDue 失败 → 本轮跳过、不 panic。
func TestCompensationLoop_ListErrorNoCrash(t *testing.T) {
	store := &fakeCompStore{listErr: errors.New("mongo down")}
	replay := &fakeReplayer{}
	l := NewCompensationLoop(store, replay, CompensationConfig{}, nil)

	l.Tick(context.Background()) // 不应 panic

	if len(replay.replayed) != 0 || len(store.deleted) != 0 || len(store.marked) != 0 {
		t.Fatalf("list error should skip all, got replay=%v del=%v mark=%v", replay.replayed, store.deleted, store.marked)
	}
}

// TestCompensationLoop_IdempotentDelete 幂等性：成功即删除，同 ID 不二次重放。
func TestCompensationLoop_IdempotentDelete(t *testing.T) {
	store := &fakeCompStore{items: []dao.Compensation{comp("c1", 0)}}
	replay := &fakeReplayer{}
	l := NewCompensationLoop(store, replay, CompensationConfig{}, nil)

	l.Tick(context.Background())
	l.Tick(context.Background())

	if len(replay.replayed) != 1 {
		t.Fatalf("after delete, item should not replay again, replayed=%v", replay.replayed)
	}
}

// TestDefaultBackoff 指数退避封顶 1h。
func TestDefaultBackoff(t *testing.T) {
	if defaultBackoff(0) != defaultScanInterval {
		t.Fatalf("backoff(0) = %v", defaultBackoff(0))
	}
	if defaultBackoff(10) != maxBackoff {
		t.Fatalf("backoff(10) should cap at 1h, got %v", defaultBackoff(10))
	}
}
