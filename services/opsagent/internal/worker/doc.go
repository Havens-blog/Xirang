// Package worker 告警异步排查 worker（进程内池，P1 单实例约束）。
//
// 并发 5 + 队列 100 + Redis 指纹去重（SET NX + TTL 10min）+ LLM 滚动 1h 预算；
// 兼为 compensations 补偿重试循环的唯一读写属主（每 30s 扫描重放）。
// 队列/并发/去重窗口等阈值由 internal/config 注入。具体实现在后续任务交付。
package worker
