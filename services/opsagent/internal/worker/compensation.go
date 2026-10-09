package worker

import (
	"context"
	"time"

	"github.com/Havens-blog/opsagent/internal/repository/dao"
)

// 落库补偿重试循环（tech-design › Error Handling › 落库补偿）：
// 每 30s 扫描 compensations（status=retry 且 next_retry_at ≤ now）逐条重放，
// 成功即删除、重试达上限 5 次标 dead 并推运维告警。该循环是 compensations 集合
// 唯一读写属主——落库失败方只 Enqueue 写入，不重放。
// 与 4.1 告警 worker 同进程池但互相隔离（独立 goroutine + 独立依赖）。

// maxCompensationRetries 补偿重试上限（达上限标 dead + 运维告警）。
const maxCompensationRetries = 5

// defaultScanInterval 补偿循环扫描周期。
const defaultScanInterval = 30 * time.Second

// defaultCompBatch 单轮扫描批量上限（防单轮过量）。
const defaultCompBatch = 100

// maxBackoff 重试退避上限（防指数退避无限膨胀）。
const maxBackoff = time.Hour

// CompensationStore 补偿队列窄接口（签名与 dao.CompensationDAO 对齐，组合根直接注入）。
type CompensationStore interface {
	// ListDue 扫描到期待重试项（status=retry 且 next_retry_at ≤ now），next_retry_at 升序。
	ListDue(ctx context.Context, now time.Time, limit int) ([]dao.Compensation, error)
	// Delete 重放成功后删除补偿项。
	Delete(ctx context.Context, id string) error
	// MarkRetried 重放失败：retry_count+1 并推进 next_retry_at；已达上限标 dead。
	MarkRetried(ctx context.Context, id string, nextRetryAt time.Time) error
}

// Replayer 补偿重放器。幂等约束：按 payload 内业务唯一 ID（session/diagnosis ID）
// 幂等重放（upsert），成功后可重复调用无害；实现由组合根注入（落库方逻辑）。
type Replayer interface {
	Replay(ctx context.Context, c dao.Compensation) error
}

// CompensationConfig 补偿循环配置。
type CompensationConfig struct {
	Interval time.Duration // 扫描周期（≤0 取 30s）
	Batch    int           // 单轮批量上限（≤0 取 100）
}

// CompensationLoop 补偿重试循环。
type CompensationLoop struct {
	store    CompensationStore
	replay   Replayer
	interval time.Duration
	batch    int
	now      func() time.Time
	backoff  func(retryCount int) time.Duration
	onDead   func(c dao.Compensation) // dead 运维告警回调；nil 忽略
}

// NewCompensationLoop 构造补偿循环。onDead 为 dead 告警回调（记录/推送运维告警）。
func NewCompensationLoop(store CompensationStore, replay Replayer, cfg CompensationConfig, onDead func(dao.Compensation)) *CompensationLoop {
	if cfg.Interval <= 0 {
		cfg.Interval = defaultScanInterval
	}
	if cfg.Batch <= 0 {
		cfg.Batch = defaultCompBatch
	}
	return &CompensationLoop{
		store:    store,
		replay:   replay,
		interval: cfg.Interval,
		batch:    cfg.Batch,
		now:      time.Now,
		backoff:  defaultBackoff,
		onDead:   onDead,
	}
}

// Run 启动循环：立即扫一轮后按周期重复，阻塞直至 ctx 取消。
func (l *CompensationLoop) Run(ctx context.Context) {
	l.tick(ctx)
	t := time.NewTicker(l.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.tick(ctx)
		}
	}
}

// Tick 扫描并重放一轮到期补偿项（导出供单测与手动触发）。
func (l *CompensationLoop) Tick(ctx context.Context) {
	l.tick(ctx)
}

// tick 单轮：ListDue → 逐条重放 → 成功删除 / 失败 MarkRetried（达上限 dead + onDead）。
// 单轮 ListDue 失败仅跳过本轮，不中断循环。
func (l *CompensationLoop) tick(ctx context.Context) {
	now := l.now()
	items, err := l.store.ListDue(ctx, now, l.batch)
	if err != nil {
		return
	}
	for _, c := range items {
		if err := l.replay.Replay(ctx, c); err == nil {
			_ = l.store.Delete(ctx, c.ID)
			continue
		}
		_ = l.store.MarkRetried(ctx, c.ID, now.Add(l.backoff(c.RetryCount)))
		if c.RetryCount+1 >= maxCompensationRetries {
			l.alertDead(c)
		}
	}
}

// alertDead dead 补偿项上报（onDead 非 nil 时）。
func (l *CompensationLoop) alertDead(c dao.Compensation) {
	if l.onDead != nil {
		l.onDead(c)
	}
}

// defaultBackoff 指数退避：30s · 2^retryCount，封顶 1h。
func defaultBackoff(retryCount int) time.Duration {
	d := defaultScanInterval * time.Duration(1<<retryCount)
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}
