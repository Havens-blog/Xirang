package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// 告警异步排查 worker（tech-design › Architecture › 告警风暴防护）：
// 指纹聚合去重（Redis SET NX + TTL 10min）+ 并发上限 5（进程内池）+ 等待队列 100
// （有界 channel）+ LLM 滚动 1h 预算（Redis INCR，耗尽降级纯规则结论）。
// P1 单实例约束：队列/并发为进程局部资源，多实例语义失效（横向扩容列入 P2）。

// 缺省阈值（与 config.Thresholds 默认对齐）。
const (
	defaultConcurrency   = 5
	defaultQueueCapacity = 100
	defaultDedupWindow   = 10 * time.Minute
	defaultBudgetLimit   = 100
)

// budgetTTL LLM 预算窗口 TTL（滚动 1h，键按整小时分桶）。
const budgetTTL = time.Hour

// Config 告警 worker 阈值配置。
type Config struct {
	Concurrency   int           // 并发上限（≤0 取 5）
	QueueCapacity int           // 队列容量（≤0 取 100）
	DedupWindow   time.Duration // 指纹去重窗口（≤0 取 10min）
	BudgetLimit   int64         // LLM 1h 预算（≤0 取 100）
}

// Deduper 指纹去重窄接口（Redis SET NX + TTL）：返回是否抢占到（true=窗口首条）。
type Deduper interface {
	// TryAcquire 抢占指纹去重窗口；true 表示窗口首条（应排查），false 表示窗口内已排查。
	TryAcquire(ctx context.Context, fingerprint string, ttl time.Duration) (bool, error)
}

// Budget LLM 滚动 1h 预算窄接口（Redis INCR + EXPIRE）：返回窗口累计调用次数。
type Budget interface {
	// Incr 递增预算计数并返回累计值（首次创建键时按 ttl 设置过期）。
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

// AlertLedger 告警记账窄接口（去重留档 merged / 溢出丢弃留档 overflow_dropped）。
type AlertLedger interface {
	// RecordArrival 窗口内首条落地（new）或同指纹 $inc count 置 merged。
	RecordArrival(ctx context.Context, tenantID, fingerprint string, payload map[string]any, windowStart time.Time) (domain.DedupConclusion, int64, error)
	// RecordOverflow 队列溢出丢弃的低级别告警留档（overflow_dropped，保留 raw_payload）。
	RecordOverflow(ctx context.Context, tenantID, fingerprint string, payload map[string]any, windowStart time.Time) error
}

// Investigator 排查执行窄接口（编排层 Orchestrator.RunAlert 适配）。
// degrade=true 表示 LLM 滚动 1h 预算耗尽、应降级纯规则结论；排查/落库/推送不中断。
type Investigator interface {
	Investigate(ctx context.Context, alert domain.AlertEvent, degrade bool) (*domain.RiskEntry, error)
}

// EnqueueResult 入队结果（半成功语义：溢出丢弃仍视为受理，dropped=true 标志低级别放弃排查）。
type EnqueueResult struct {
	Accepted bool
	Dropped  bool
	Err      error // 非致命错误（溢出处留档失败等），供上层埋点
}

// Worker 告警异步排查 worker。
type Worker struct {
	deduper Deduper
	budget  Budget
	ledger  AlertLedger
	invest  Investigator
	cfg     Config
	onError func(error) // 后台任务错误回调（日志/监控）；nil 忽略

	queue chan domain.AlertEvent
	stop  chan struct{}
	wg    sync.WaitGroup
}

// New 构造 Worker（阈值非法回退默认）。
func New(deduper Deduper, budget Budget, ledger AlertLedger, invest Investigator, cfg Config, onError func(error)) *Worker {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = defaultConcurrency
	}
	if cfg.QueueCapacity <= 0 {
		cfg.QueueCapacity = defaultQueueCapacity
	}
	if cfg.DedupWindow <= 0 {
		cfg.DedupWindow = defaultDedupWindow
	}
	if cfg.BudgetLimit <= 0 {
		cfg.BudgetLimit = defaultBudgetLimit
	}
	return &Worker{
		deduper: deduper,
		budget:  budget,
		ledger:  ledger,
		invest:  invest,
		cfg:     cfg,
		onError: onError,
		queue:   make(chan domain.AlertEvent, cfg.QueueCapacity),
		stop:    make(chan struct{}),
	}
}

// Start 启动并发池（Concurrency 个 goroutine 消费队列）。
func (w *Worker) Start(ctx context.Context) {
	for i := 0; i < w.cfg.Concurrency; i++ {
		w.wg.Add(1)
		go w.loop(ctx)
	}
}

// Stop 关闭并等待在途任务结束（关闭后入队不再受理）。
func (w *Worker) Stop() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	w.wg.Wait()
}

// Enqueue 入队：低级别队列满时溢出丢弃（留档 overflow_dropped）；高危阻塞入队永不丢弃。
func (w *Worker) Enqueue(ctx context.Context, alert domain.AlertEvent) EnqueueResult {
	if isHigh(alert.Level) {
		return w.enqueueHigh(ctx, alert)
	}
	select {
	case w.queue <- alert:
		return EnqueueResult{Accepted: true}
	default:
		return w.dropOverflow(ctx, alert)
	}
}

// enqueueHigh 高危告警：阻塞入队（背压），绝不溢出丢弃。
func (w *Worker) enqueueHigh(ctx context.Context, alert domain.AlertEvent) EnqueueResult {
	select {
	case w.queue <- alert:
		return EnqueueResult{Accepted: true}
	case <-w.stop:
		return EnqueueResult{}
	case <-ctx.Done():
		return EnqueueResult{}
	}
}

// dropOverflow 低级别告警队列满：写 alerts overflow_dropped 留档后返回 dropped。
func (w *Worker) dropOverflow(ctx context.Context, alert domain.AlertEvent) EnqueueResult {
	if err := w.ledger.RecordOverflow(ctx, alert.Tenant, fingerprintOf(alert), alert.Raw, windowStartOf(alert)); err != nil {
		return EnqueueResult{Dropped: true, Err: err}
	}
	return EnqueueResult{Dropped: true}
}

// loop 单个消费 goroutine：抢队列 → 去重 → 预算 → 排查。
func (w *Worker) loop(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case alert := <-w.queue:
			w.process(ctx, alert)
		}
	}
}

// process 单条告警排查：指纹去重 + LLM 预算 + 排查执行。
func (w *Worker) process(ctx context.Context, alert domain.AlertEvent) {
	fp := fingerprintOf(alert)
	ws := windowStartOf(alert)

	// ① 指纹去重（Redis SET NX + TTL）。Redis 故障 fail-open：宁多排查不漏告警。
	acquired, err := w.deduper.TryAcquire(ctx, fp, w.cfg.DedupWindow)
	if err != nil {
		w.reportErr(err)
		acquired = true
	}
	if !acquired {
		// 窗口内已排查：归并 $inc count + merged（记账留档），不重复排查。
		if _, _, err := w.ledger.RecordArrival(ctx, alert.Tenant, fp, alert.Raw, ws); err != nil {
			w.reportErr(err)
		}
		return
	}

	// ② LLM 滚动 1h 预算（Redis INCR）。预算耗尽 → degrade=true（纯规则结论）。
	degrade := false
	if n, err := w.budget.Incr(ctx, budgetKey(alert.Tenant, time.Now()), budgetTTL); err != nil {
		w.reportErr(err)
	} else {
		degrade = n > w.cfg.BudgetLimit
	}

	// ③ 排查执行（编排层 RunAlert；失败不中断其它告警）。
	if _, err := w.invest.Investigate(ctx, alert, degrade); err != nil {
		w.reportErr(err)
	}
}

// reportErr 上报后台任务错误（onError 非 nil 时）。
func (w *Worker) reportErr(err error) {
	if err != nil && w.onError != nil {
		w.onError(err)
	}
}

// fingerprintOf 告警指纹（与 domain.Fingerprint 一致）。
func fingerprintOf(alert domain.AlertEvent) string {
	return domain.Fingerprint(alert.Tenant, alert.ServiceName, alert.Metric, alert.Level)
}

// windowStartOf 去重窗口起点（与编排层 RunAlert 一致：缺省近 1h）。
func windowStartOf(alert domain.AlertEvent) time.Time {
	if alert.Timeframe.StartTime.IsZero() {
		return time.Now().Add(-time.Hour)
	}
	return alert.Timeframe.StartTime
}

// isHigh 高危级别判定（P0/P1/CRITICAL/HIGH 永不丢弃）。
func isHigh(level string) bool {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "P0", "P1", "CRITICAL", "HIGH":
		return true
	default:
		return false
	}
}

// budgetKey LLM 预算键（按租户 + 整小时分桶）。
func budgetKey(tenant string, now time.Time) string {
	return fmt.Sprintf("opsagent:llm:budget:%s:%d", tenant, now.Truncate(time.Hour).Unix())
}
