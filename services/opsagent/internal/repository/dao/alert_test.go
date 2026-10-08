package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// readAlertDoc 读回 alerts 原始文档供断言（绕过 DAO，验证落库形态）。
func readAlertDoc(t *testing.T, d *mongox.Mongo, tenant, fingerprint string) (dedup string, count int64) {
	t.Helper()
	var doc struct {
		Dedup string `bson:"dedup_conclusion"`
		Count int64  `bson:"count"`
	}
	err := d.Collection(AlertsCollection).FindOne(context.Background(), bson.M{"tenant_id": tenant, "fingerprint": fingerprint}).Decode(&doc)
	if err != nil {
		t.Fatalf("read alert doc: %v", err)
	}
	return doc.Dedup, doc.Count
}

// TestAlertDAO_RecordArrival_NewThenMerged AC：首条 new/count=1，同指纹后续 $inc count
// 置 merged。
func TestAlertDAO_RecordArrival_NewThenMerged(t *testing.T) {
	db := newTestMongo(t)
	d := NewAlertDAO(db)
	ctx := context.Background()
	base := time.Now().Truncate(time.Millisecond)
	payload := map[string]any{"serviceName": "order-service", "metric": "error_rate"}

	got, count, err := d.RecordArrival(ctx, "tenant-a", "fp-order", payload, base)
	if err != nil {
		t.Fatalf("first arrival: %v", err)
	}
	if got != domain.DedupNew || count != 1 {
		t.Fatalf("first = %q/%d, want new/1", got, count)
	}

	got, count, err = d.RecordArrival(ctx, "tenant-a", "fp-order", payload, base)
	if err != nil || got != domain.DedupMerged || count != 2 {
		t.Fatalf("second = %q/%d err=%v, want merged/2", got, count, err)
	}

	got, count, err = d.RecordArrival(ctx, "tenant-a", "fp-order", payload, base)
	if err != nil || got != domain.DedupMerged || count != 3 {
		t.Fatalf("third = %q/%d err=%v, want merged/3", got, count, err)
	}

	if dedup, cnt := readAlertDoc(t, db, "tenant-a", "fp-order"); dedup != string(domain.DedupMerged) || cnt != 3 {
		t.Fatalf("persisted = %q/%d, want merged/3", dedup, cnt)
	}
}

// TestAlertDAO_RecordArrival_WindowAndTenantIsolation 窗口与租户谓词：不同 window_start
// 或不同 tenant 均独立记账（硬约束：读写带 tenant 谓词）。
func TestAlertDAO_RecordArrival_WindowAndTenantIsolation(t *testing.T) {
	db := newTestMongo(t)
	d := NewAlertDAO(db)
	ctx := context.Background()
	base := time.Now().Truncate(time.Millisecond)

	if got, _, _ := d.RecordArrival(ctx, "tenant-a", "fp-1", map[string]any{}, base); got != domain.DedupNew {
		t.Fatalf("w1 = %q, want new", got)
	}
	// 同指纹不同窗口 → 新窗口首条。
	if got, _, _ := d.RecordArrival(ctx, "tenant-a", "fp-1", map[string]any{}, base.Add(10*time.Minute)); got != domain.DedupNew {
		t.Fatalf("w2 = %q, want new", got)
	}
	// 同指纹同窗口不同租户 → 租户隔离，各自首条。
	if got, _, _ := d.RecordArrival(ctx, "tenant-b", "fp-1", map[string]any{}, base); got != domain.DedupNew {
		t.Fatalf("tenant-b = %q, want new", got)
	}
}

// TestAlertDAO_RecordOverflow AC：溢出丢弃写 overflow_dropped 且保留 raw_payload。
func TestAlertDAO_RecordOverflow(t *testing.T) {
	db := newTestMongo(t)
	d := NewAlertDAO(db)
	ctx := context.Background()
	base := time.Now().Truncate(time.Millisecond)
	payload := map[string]any{"serviceName": "cdn-edge", "dropped": true}

	if err := d.RecordOverflow(ctx, "tenant-a", "fp-cdn", payload, base); err != nil {
		t.Fatalf("record overflow: %v", err)
	}

	var doc struct {
		Dedup      string         `bson:"dedup_conclusion"`
		RawPayload map[string]any `bson:"raw_payload"`
	}
	if err := db.Collection(AlertsCollection).FindOne(ctx, bson.M{"tenant_id": "tenant-a", "fingerprint": "fp-cdn"}).Decode(&doc); err != nil {
		t.Fatalf("read overflow doc: %v", err)
	}
	if doc.Dedup != string(domain.DedupOverflowDropped) {
		t.Fatalf("dedup = %q, want overflow_dropped", doc.Dedup)
	}
	if doc.RawPayload["serviceName"] != "cdn-edge" {
		t.Fatalf("raw_payload lost: %+v", doc.RawPayload)
	}
}

// TestAlertDAO_LinkDiagnosis 回填指纹窗口关联诊断。
func TestAlertDAO_LinkDiagnosis(t *testing.T) {
	db := newTestMongo(t)
	d := NewAlertDAO(db)
	ctx := context.Background()
	base := time.Now().Truncate(time.Millisecond)

	if _, _, err := d.RecordArrival(ctx, "tenant-a", "fp-1", map[string]any{}, base); err != nil {
		t.Fatalf("arrival: %v", err)
	}
	if err := d.LinkDiagnosis(ctx, "tenant-a", "fp-1", base, "dx-1"); err != nil {
		t.Fatalf("link diagnosis: %v", err)
	}
	var doc struct {
		DiagnosisRef string `bson:"diagnosis_ref"`
	}
	if err := db.Collection(AlertsCollection).FindOne(ctx, bson.M{"tenant_id": "tenant-a", "fingerprint": "fp-1"}).Decode(&doc); err != nil {
		t.Fatalf("read doc: %v", err)
	}
	if doc.DiagnosisRef != "dx-1" {
		t.Fatalf("diagnosis_ref = %q, want dx-1", doc.DiagnosisRef)
	}

	// 不存在/跨租户 → ErrNoDocuments。
	if err := d.LinkDiagnosis(ctx, "tenant-a", "fp-1", base.Add(time.Hour), "dx-2"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("missing window err = %v, want ErrNoDocuments", err)
	}
}

// TestAlertDAO_InitIndexes 索引名与 schema.mongo.js 一致，重复执行幂等。
func TestAlertDAO_InitIndexes(t *testing.T) {
	db := newTestMongo(t)
	d := NewAlertDAO(db)
	ctx := context.Background()

	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes: %v", err)
	}
	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes (idempotent): %v", err)
	}
	cur, err := db.Collection(AlertsCollection).Indexes().List(ctx)
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
	for _, want := range []string{"idx_alerts_fingerprint_window", "idx_alerts_tenant_time"} {
		if !got[want] {
			t.Fatalf("index %q missing, got %v", want, got)
		}
	}
}
