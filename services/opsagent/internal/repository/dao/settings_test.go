package dao

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Havens-blog/opsagent/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestSettingsDAO_UpsertGetRoundtrip AC：按 scope 读写，整体替换，updated_at 刷新。
func TestSettingsDAO_UpsertGetRoundtrip(t *testing.T) {
	db := newTestMongo(t)
	d := NewSettingsDAO(db)
	ctx := context.Background()

	// 整体替换：先写 v1，再写 v2，读回 v2。
	v1 := []map[string]any{{"name": "qwen", "default": true}, {"name": "deepseek", "default": false}}
	if err := d.Upsert(ctx, "tenant-a", domain.ScopeLLM, v1); err != nil {
		t.Fatalf("upsert v1: %v", err)
	}
	var got1 []map[string]any
	if err := d.Get(ctx, "tenant-a", domain.ScopeLLM, &got1); err != nil {
		t.Fatalf("get v1: %v", err)
	}
	if !reflect.DeepEqual(got1, v1) {
		t.Fatalf("roundtrip mismatch: got %#v want %#v", got1, v1)
	}

	v2 := []map[string]any{{"name": "openai", "default": true}}
	if err := d.Upsert(ctx, "tenant-a", domain.ScopeLLM, v2); err != nil {
		t.Fatalf("upsert v2: %v", err)
	}
	var got2 []map[string]any
	if err := d.Get(ctx, "tenant-a", domain.ScopeLLM, &got2); err != nil {
		t.Fatalf("get v2: %v", err)
	}
	if !reflect.DeepEqual(got2, v2) {
		t.Fatalf("replace mismatch: got %#v want %#v", got2, v2)
	}
}

// TestSettingsDAO_GlobalSentinel AC：tenant_id="" 为全局哨兵，与租户文档同唯一索引不冲突。
func TestSettingsDAO_GlobalSentinel(t *testing.T) {
	db := newTestMongo(t)
	d := NewSettingsDAO(db)
	ctx := context.Background()

	global := map[string]any{"channel": "dingtalk", "enabled": "true"}
	if err := d.Upsert(ctx, "", domain.ScopeNotify, global); err != nil {
		t.Fatalf("global upsert: %v", err)
	}
	tenant := map[string]any{"channel": "feishu", "enabled": "false"}
	if err := d.Upsert(ctx, "tenant-a", domain.ScopeNotify, tenant); err != nil {
		t.Fatalf("tenant upsert: %v", err)
	}

	var g map[string]any
	if err := d.Get(ctx, "", domain.ScopeNotify, &g); err != nil {
		t.Fatalf("global get: %v", err)
	}
	if g["channel"] != "dingtalk" {
		t.Fatalf("global = %#v, want dingtalk", g)
	}
	var tn map[string]any
	if err := d.Get(ctx, "tenant-a", domain.ScopeNotify, &tn); err != nil {
		t.Fatalf("tenant get: %v", err)
	}
	if tn["channel"] != "feishu" {
		t.Fatalf("tenant = %#v, want feishu", tn)
	}
}

// TestSettingsDAO_ScopeAndTenantIsolation 同 scope 不同租户、不同 scope 互不影响。
func TestSettingsDAO_ScopeAndTenantIsolation(t *testing.T) {
	db := newTestMongo(t)
	d := NewSettingsDAO(db)
	ctx := context.Background()

	if err := d.Upsert(ctx, "tenant-a", domain.ScopeRiskWhitelist, []map[string]any{{"tool": "query_log"}}); err != nil {
		t.Fatalf("upsert risk a: %v", err)
	}
	if err := d.Upsert(ctx, "tenant-b", domain.ScopeRiskWhitelist, []map[string]any{{"tool": "query_asset"}}); err != nil {
		t.Fatalf("upsert risk b: %v", err)
	}

	var a []map[string]any
	if err := d.Get(ctx, "tenant-a", domain.ScopeRiskWhitelist, &a); err != nil {
		t.Fatalf("get a: %v", err)
	}
	if a[0]["tool"] != "query_log" {
		t.Fatalf("tenant-a = %#v", a)
	}

	// 不同 scope 无数据 → ErrNoDocuments。
	var out any
	if err := d.Get(ctx, "tenant-a", domain.ScopePresetQueries, &out); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("missing scope err = %v, want ErrNoDocuments", err)
	}
}

// TestSettingsDAO_InitIndexes 唯一索引幂等（scope + tenant_id）。
func TestSettingsDAO_InitIndexes(t *testing.T) {
	db := newTestMongo(t)
	d := NewSettingsDAO(db)
	ctx := context.Background()

	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes: %v", err)
	}
	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes (idempotent): %v", err)
	}
	// 唯一约束：重复 upsert 同 (scope, tenant) 仅一行，不因唯一索引报错。
	for i := 0; i < 3; i++ {
		if err := d.Upsert(ctx, "tenant-a", domain.ScopeDatasource, []map[string]any{{"name": "cdn"}}); err != nil {
			t.Fatalf("repeated upsert: %v", err)
		}
	}
	cur, err := db.Collection(SettingsCollection).Indexes().List(ctx)
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
	if !got["idx_settings_scope_tenant"] {
		t.Fatalf("index idx_settings_scope_tenant missing, got %v", got)
	}
}
