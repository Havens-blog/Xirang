package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// ---- fakes ----

type fakeDeduper struct {
	acquired bool
	err      error
	seen     []string
}

func (f *fakeDeduper) TryAcquire(_ context.Context, fp string, _ time.Duration) (bool, error) {
	f.seen = append(f.seen, fp)
	return f.acquired, f.err
}

type fakeBudget struct {
	n    int64
	err  error
	keys []string
}

func (f *fakeBudget) Incr(_ context.Context, key string, _ time.Duration) (int64, error) {
	f.keys = append(f.keys, key)
	return f.n, f.err
}

type fakeLedger struct {
	arrivals       int
	overflowCount  int
	overflowErr    error
	overflowTenant string
}

func (f *fakeLedger) RecordArrival(_ context.Context, _, _ string, _ map[string]any, _ time.Time) (domain.DedupConclusion, int64, error) {
	f.arrivals++
	return domain.DedupMerged, 1, nil
}

func (f *fakeLedger) RecordOverflow(_ context.Context, tenantID, _ string, _ map[string]any, _ time.Time) error {
	f.overflowCount++
	f.overflowTenant = tenantID
	return f.overflowErr
}

type fakeInvest struct {
	mu           sync.Mutex
	alerts       []domain.AlertEvent
	degradeCalls []bool
	entry        *domain.RiskEntry
	err          error
}

func (f *fakeInvest) Investigate(_ context.Context, a domain.AlertEvent, degrade bool) (*domain.RiskEntry, error) {
	f.mu.Lock()
	f.alerts = append(f.alerts, a)
	f.degradeCalls = append(f.degradeCalls, degrade)
	f.mu.Unlock()
	return f.entry, f.err
}

func (f *fakeInvest) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.alerts)
}

func (f *fakeInvest) lastDegrade() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.degradeCalls) == 0 {
		return false
	}
	return f.degradeCalls[len(f.degradeCalls)-1]
}

func testAlert(level string) domain.AlertEvent {
	return domain.AlertEvent{
		Tenant:      "t1",
		ServiceName: "order-service",
		Metric:      "错误率",
		Level:       level,
		Timeframe:   domain.Timeframe{StartTime: time.Now().Add(-time.Hour), EndTime: time.Now()},
		Raw:         map[string]any{"x": 1},
	}
}

// ---- tests ----

// TestProcess_DedupMerged 指纹去重命中（窗口内已排查）→ 归并记账 merge，不重复排查。
func TestProcess_DedupMerged(t *testing.T) {
	ledger := &fakeLedger{}
	invest := &fakeInvest{}
	dedup := &fakeDeduper{acquired: false}
	w := New(dedup, &fakeBudget{n: 1}, ledger, invest, Config{BudgetLimit: 100}, nil)

	w.process(context.Background(), testAlert("P3"))

	if ledger.arrivals != 1 {
		t.Fatalf("merged should record arrival, got %d", ledger.arrivals)
	}
	if invest.count() != 0 {
		t.Fatalf("merged should not investigate, got %d", invest.count())
	}
}

// TestProcess_DedupAcquiredUnderBudget 窗口首条且预算未耗尽 → 正常排查（degrade=false）。
func TestProcess_DedupAcquiredUnderBudget(t *testing.T) {
	invest := &fakeInvest{}
	w := New(&fakeDeduper{acquired: true}, &fakeBudget{n: 50}, &fakeLedger{}, invest, Config{BudgetLimit: 100}, nil)

	w.process(context.Background(), testAlert("P3"))

	if invest.count() != 1 {
		t.Fatalf("should investigate once, got %d", invest.count())
	}
	if invest.lastDegrade() {
		t.Fatal("under budget should not degrade")
	}
}

// TestProcess_BudgetExhausted 预算耗尽（>limit）→ 排查降级纯规则结论（degrade=true）。
func TestProcess_BudgetExhausted(t *testing.T) {
	invest := &fakeInvest{}
	w := New(&fakeDeduper{acquired: true}, &fakeBudget{n: 101}, &fakeLedger{}, invest, Config{BudgetLimit: 100}, nil)

	w.process(context.Background(), testAlert("P3"))

	if invest.count() != 1 || !invest.lastDegrade() {
		t.Fatalf("budget exhausted should degrade, calls=%d degrade=%v", invest.count(), invest.lastDegrade())
	}
}

// TestProcess_DedupErrorFailOpen 去重 Redis 故障 → fail-open 仍排查（宁多不漏）。
func TestProcess_DedupErrorFailOpen(t *testing.T) {
	invest := &fakeInvest{}
	w := New(&fakeDeduper{err: errors.New("redis down")}, &fakeBudget{n: 1}, &fakeLedger{}, invest, Config{}, nil)

	w.process(context.Background(), testAlert("P3"))

	if invest.count() != 1 {
		t.Fatalf("dedup error should fail-open and investigate, got %d", invest.count())
	}
}

// TestEnqueue_LowOverflow 低级别队列满 → 溢出丢弃 + 留档 overflow_dropped。
func TestEnqueue_LowOverflow(t *testing.T) {
	ledger := &fakeLedger{}
	w := New(&fakeDeduper{}, &fakeBudget{}, ledger, &fakeInvest{}, Config{QueueCapacity: 1}, nil)
	w.queue <- testAlert("P3") // 填满

	res := w.Enqueue(context.Background(), testAlert("P3"))
	if !res.Dropped || res.Accepted {
		t.Fatalf("low enqueue on full queue should drop, got %+v", res)
	}
	if ledger.overflowCount != 1 {
		t.Fatalf("overflow should be recorded, got %d", ledger.overflowCount)
	}
}

// TestEnqueue_OverflowRecordError 溢出留档失败 → 仍 dropped，但 Err 上抛。
func TestEnqueue_OverflowRecordError(t *testing.T) {
	ledger := &fakeLedger{overflowErr: errors.New("mongo down")}
	w := New(&fakeDeduper{}, &fakeBudget{}, ledger, &fakeInvest{}, Config{QueueCapacity: 1}, nil)
	w.queue <- testAlert("P3")

	res := w.Enqueue(context.Background(), testAlert("P3"))
	if !res.Dropped || res.Err == nil {
		t.Fatalf("overflow record error should surface, got %+v", res)
	}
}

// TestEnqueue_HighNeverDropped 高危队列满 → 阻塞入队（背压），绝不溢出丢弃。
func TestEnqueue_HighNeverDropped(t *testing.T) {
	w := New(&fakeDeduper{}, &fakeBudget{}, &fakeLedger{}, &fakeInvest{}, Config{QueueCapacity: 1}, nil)
	w.queue <- testAlert("P3") // 填满

	done := make(chan EnqueueResult, 1)
	go func() { done <- w.Enqueue(context.Background(), testAlert("P0")) }()

	<-w.queue // 释放空间

	select {
	case res := <-done:
		if !res.Accepted || res.Dropped {
			t.Fatalf("high severity should never drop, got %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("high severity enqueue blocked forever")
	}
}

// TestWorker_ConcurrencyLimit 并发上限生效：最多 Concurrency 个排查同时在跑。
func TestWorker_ConcurrencyLimit(t *testing.T) {
	const concurrency = 2
	const total = 6

	gate := make(chan struct{})
	var mu sync.Mutex
	inFlight, maxInFlight := 0, 0

	invest := &gateInvest{
		onEnter: func() {
			mu.Lock()
			inFlight++
			if inFlight > maxInFlight {
				maxInFlight = inFlight
			}
			mu.Unlock()
		},
		onLeave: func() {
			mu.Lock()
			inFlight--
			mu.Unlock()
		},
		gate: gate,
	}

	w := New(&fakeDeduper{acquired: true}, &fakeBudget{n: 1}, &fakeLedger{}, invest, Config{Concurrency: concurrency, QueueCapacity: total}, nil)
	ctx := context.Background()
	w.Start(ctx)

	for i := 0; i < total; i++ {
		w.Enqueue(ctx, testAlert("P3"))
	}

	// 等待 concurrency 个排查进入（阻塞在 gate）。
	if !waitFor(t, 2*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return inFlight >= concurrency
	}) {
		t.Fatal("timeout waiting for investigations to start")
	}

	mu.Lock()
	got := maxInFlight
	mu.Unlock()
	if got > concurrency {
		t.Fatalf("max concurrent = %d, want <= %d", got, concurrency)
	}

	close(gate) // 释放全部
	w.Stop()

	mu.Lock()
	finalMax := maxInFlight
	mu.Unlock()
	if finalMax != concurrency {
		t.Fatalf("final max concurrent = %d, want %d", finalMax, concurrency)
	}
}

// gateInvest 阻塞式 Investigator：进入时通知并阻塞在 gate，用于并发上限断言。
type gateInvest struct {
	onEnter func()
	onLeave func()
	gate    chan struct{}
}

func (g *gateInvest) Investigate(_ context.Context, _ domain.AlertEvent, _ bool) (*domain.RiskEntry, error) {
	g.onEnter()
	<-g.gate
	g.onLeave()
	return &domain.RiskEntry{}, nil
}

// waitFor 轮询 until 直到为真或超时。
func waitFor(t *testing.T, timeout time.Duration, until func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if until() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return until()
}

// ---- helpers ----

func TestIsHigh(t *testing.T) {
	for _, l := range []string{"P0", "P1", "CRITICAL", "HIGH", "critical"} {
		if !isHigh(l) {
			t.Fatalf("isHigh(%q) should be true", l)
		}
	}
	for _, l := range []string{"P2", "P3", "MEDIUM", "low", ""} {
		if isHigh(l) {
			t.Fatalf("isHigh(%q) should be false", l)
		}
	}
}

func TestBudgetKey(t *testing.T) {
	now := time.Date(2026, 10, 8, 13, 45, 0, 0, time.UTC)
	got := budgetKey("t1", now)
	want := fmt.Sprintf("opsagent:llm:budget:t1:%d", now.Truncate(time.Hour).Unix())
	if got != want {
		t.Fatalf("budgetKey = %q, want %q", got, want)
	}
}

func TestWindowStartOf(t *testing.T) {
	now := time.Now()
	a := domain.AlertEvent{Timeframe: domain.Timeframe{StartTime: now.Add(-2 * time.Hour)}}
	if !windowStartOf(a).Equal(now.Add(-2 * time.Hour)) {
		t.Fatal("windowStartOf should use timeframe start")
	}
	b := domain.AlertEvent{}
	if windowStartOf(b).IsZero() {
		t.Fatal("windowStartOf should fall back to now-1h")
	}
}
