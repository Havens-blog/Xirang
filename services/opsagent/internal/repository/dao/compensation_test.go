package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestCompensationDAO_EnqueueAndListDue AC：写入失败项（payload/retry_count/status=retry），
// ListDue 按 status=retry 且 next_retry_at ≤ now 扫描，next_retry_at 升序。
func TestCompensationDAO_EnqueueAndListDue(t *testing.T) {
	db := newTestMongo(t)
	d := NewCompensationDAO(db)
	ctx := context.Background()
	now := time.Now()

	if err := d.Enqueue(ctx, domain.CompKindPersistFailed, map[string]any{"session_id": "sess-1", "data": "diag"}); err != nil {
		t.Fatalf("enqueue 1: %v", err)
	}
	if err := d.Enqueue(ctx, domain.CompKindPersistFailed, map[string]any{"session_id": "sess-2", "data": "alert"}); err != nil {
		t.Fatalf("enqueue 2: %v", err)
	}

	due, err := d.ListDue(ctx, now.Add(time.Hour), 10)
	if err != nil {
		t.Fatalf("list due: %v", err)
	}
	if len(due) != 2 {
		t.Fatalf("due = %d, want 2", len(due))
	}
	for _, c := range due {
		if c.Status != domain.CompStatusRetry || c.RetryCount != 0 {
			t.Fatalf("entry = %+v, want retry/0", c)
		}
		if c.Kind != domain.CompKindPersistFailed {
			t.Fatalf("kind = %q, want persist_failed", c.Kind)
		}
		if c.CreatedAt.IsZero() || c.NextRetryAt.IsZero() {
			t.Fatalf("timestamps not filled: %+v", c)
		}
	}
}

// TestCompensationDAO_ListDue_Filter 到期谓词：next_retry_at 在未来者不命中「当前到期」。
func TestCompensationDAO_ListDue_Filter(t *testing.T) {
	db := newTestMongo(t)
	d := NewCompensationDAO(db)
	ctx := context.Background()
	now := time.Now()

	_ = d.Enqueue(ctx, domain.CompKindPersistFailed, map[string]any{"a": "1"})
	_ = d.Enqueue(ctx, domain.CompKindPersistFailed, map[string]any{"b": "2"})

	// 尚未到期（now 早于 enqueue 时刻的 next_retry_at）→ 空。
	due, err := d.ListDue(ctx, now.Add(-time.Hour), 10)
	if err != nil || len(due) != 0 {
		t.Fatalf("past due = %d err=%v, want 0", len(due), err)
	}
	// 全部到期。
	all, _ := d.ListDue(ctx, now.Add(2*time.Hour), 10)
	if len(all) != 2 {
		t.Fatalf("all = %d, want 2", len(all))
	}
	// 推进其中一条的未来重试时间后，该条不再命中「当前到期」。
	if err := d.MarkRetried(ctx, all[0].ID, now.Add(time.Hour)); err != nil {
		t.Fatalf("mark retried: %v", err)
	}
	due, _ = d.ListDue(ctx, now.Add(time.Minute), 10)
	if len(due) != 1 || due[0].ID != all[1].ID {
		t.Fatalf("due now = %+v, want only the un-advanced entry", due)
	}
}

// TestCompensationDAO_DeleteAndMarkRetried AC：重放成功删除；重放失败 retry_count+1 并推进，
// 达上限标 dead 不再被扫描。
func TestCompensationDAO_DeleteAndMarkRetried(t *testing.T) {
	db := newTestMongo(t)
	d := NewCompensationDAO(db)
	ctx := context.Background()
	far := time.Now().Add(2 * time.Hour)

	_ = d.Enqueue(ctx, domain.CompKindPersistFailed, map[string]any{"session_id": "sess-del"})
	_ = d.Enqueue(ctx, domain.CompKindPersistFailed, map[string]any{"session_id": "sess-dead"})

	all, _ := d.ListDue(ctx, far, 10)
	del, dead := all[0], all[1]

	// 删除。
	if err := d.Delete(ctx, del.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := d.Delete(ctx, del.ID); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("double delete err = %v, want ErrNoDocuments", err)
	}

	// 达上限 → dead。maxCompensationRetries 次失败后不再 retry。
	for i := 0; i < maxCompensationRetries; i++ {
		if err := d.MarkRetried(ctx, dead.ID, far); err != nil {
			t.Fatalf("mark retried %d: %v", i, err)
		}
	}
	var raw struct {
		Status     string `bson:"status"`
		RetryCount int    `bson:"retry_count"`
	}
	if err := db.Collection(CompensationsCollection).FindOne(ctx, bson.M{"_id": dead.ID}).Decode(&raw); err != nil {
		t.Fatalf("read dead doc: %v", err)
	}
	if raw.Status != string(domain.CompStatusDead) || raw.RetryCount != maxCompensationRetries {
		t.Fatalf("dead = %q/%d, want dead/%d", raw.Status, raw.RetryCount, maxCompensationRetries)
	}

	// dead 不再被扫描。
	remaining, _ := d.ListDue(ctx, far, 10)
	if len(remaining) != 0 {
		t.Fatalf("remaining = %d, want 0 (dead excluded)", len(remaining))
	}

	// 对 dead 项再 MarkRetried → 已被 MarkRetried 的 status=retry 过滤拒绝。
	if err := d.MarkRetried(ctx, dead.ID, far); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("mark dead again err = %v, want ErrNoDocuments", err)
	}
}

// TestCompensationDAO_InitIndexes 索引名与 schema.mongo.js 一致，幂等。
func TestCompensationDAO_InitIndexes(t *testing.T) {
	db := newTestMongo(t)
	d := NewCompensationDAO(db)
	ctx := context.Background()

	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes: %v", err)
	}
	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes (idempotent): %v", err)
	}
	cur, err := db.Collection(CompensationsCollection).Indexes().List(ctx)
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	got := map[string]bool{}
	for cur.Next(ctx) {
		var doc struct {
			Name string `bson:"name"`
		}
		if err := cur.Decode(&doc); err != nil {
			t.Fatalf("decode index: %v", err)
		}
		got[doc.Name] = true
	}
	if !got["idx_comp_retry"] {
		t.Fatalf("index idx_comp_retry missing, got %v", got)
	}
}
