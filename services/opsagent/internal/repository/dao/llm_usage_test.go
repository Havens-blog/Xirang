package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// TestLLMUsageDAO_IncrCall_UpsertAndAccumulate AC：按 (tenant_id, window_start) 聚合
// calls，窗口不存在时 upsert，跨窗口/跨租户独立计数。
func TestLLMUsageDAO_IncrCall_UpsertAndAccumulate(t *testing.T) {
	db := newTestMongo(t)
	d := NewLLMUsageDAO(db)
	ctx := context.Background()
	w1 := time.Now().Truncate(time.Millisecond)
	w2 := w1.Add(time.Hour)

	if got, err := d.IncrCall(ctx, "tenant-a", w1); err != nil || got != 1 {
		t.Fatalf("first = %d err=%v, want 1", got, err)
	}
	if got, _ := d.IncrCall(ctx, "tenant-a", w1); got != 2 {
		t.Fatalf("second = %d, want 2", got)
	}
	if got, _ := d.IncrCall(ctx, "tenant-a", w1); got != 3 {
		t.Fatalf("third = %d, want 3", got)
	}
	// 不同窗口 → 独立计数。
	if got, _ := d.IncrCall(ctx, "tenant-a", w2); got != 1 {
		t.Fatalf("w2 = %d, want 1", got)
	}
	// 不同租户 → 独立计数。
	if got, _ := d.IncrCall(ctx, "tenant-b", w1); got != 1 {
		t.Fatalf("tenant-b = %d, want 1", got)
	}
}

// TestLLMUsageDAO_GetWindow 读窗口账本；不存在 → ErrNoDocuments。
func TestLLMUsageDAO_GetWindow(t *testing.T) {
	db := newTestMongo(t)
	d := NewLLMUsageDAO(db)
	ctx := context.Background()
	w := time.Now().Truncate(time.Millisecond)

	if _, err := d.GetWindow(ctx, "tenant-a", w); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("missing window err = %v, want ErrNoDocuments", err)
	}

	if _, err := d.IncrCall(ctx, "tenant-a", w); err != nil {
		t.Fatalf("incr: %v", err)
	}
	got, err := d.GetWindow(ctx, "tenant-a", w)
	if err != nil {
		t.Fatalf("get window: %v", err)
	}
	if got.Calls != 1 || got.BudgetExceeded {
		t.Fatalf("get = %+v, want calls=1 exceeded=false", got)
	}
}

// TestLLMUsageDAO_MarkBudgetExceeded 预算耗尽位可判（配合 calls 做预算判定）。
func TestLLMUsageDAO_MarkBudgetExceeded(t *testing.T) {
	db := newTestMongo(t)
	d := NewLLMUsageDAO(db)
	ctx := context.Background()
	w := time.Now().Truncate(time.Millisecond)

	if err := d.MarkBudgetExceeded(ctx, "tenant-a", w); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("mark missing err = %v, want ErrNoDocuments", err)
	}

	if _, err := d.IncrCall(ctx, "tenant-a", w); err != nil {
		t.Fatalf("incr: %v", err)
	}
	if err := d.MarkBudgetExceeded(ctx, "tenant-a", w); err != nil {
		t.Fatalf("mark exceeded: %v", err)
	}
	got, err := d.GetWindow(ctx, "tenant-a", w)
	if err != nil || !got.BudgetExceeded {
		t.Fatalf("get = %+v err=%v, want budget_exceeded=true", got, err)
	}
}

// TestLLMUsageDAO_InitIndexes 索引名与 schema.mongo.js 一致，幂等。
func TestLLMUsageDAO_InitIndexes(t *testing.T) {
	db := newTestMongo(t)
	d := NewLLMUsageDAO(db)
	ctx := context.Background()

	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes: %v", err)
	}
	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes (idempotent): %v", err)
	}
	cur, err := db.Collection(LLMUsageCollection).Indexes().List(ctx)
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
	if !got["idx_llm_tenant_window"] {
		t.Fatalf("index idx_llm_tenant_window missing, got %v", got)
	}
}
