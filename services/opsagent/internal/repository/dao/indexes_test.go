package dao

import (
	"context"
	"testing"
)

// TestDAOInitIndexes 三个集合的索引初始化（幂等），索引名与 schema.mongo.js 一致：
// sessions: idx_sessions_tenant_time / idx_sessions_tenant_service / idx_sessions_tenant_type
// diagnoses: idx_dx_session / idx_dx_tenant_service / idx_dx_tenant_time
// citations: idx_cit_dx / idx_cit_tenant
func TestDAOInitIndexes(t *testing.T) {
	db := newTestMongo(t)
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		coll string
		init func(context.Context) error
		want []string
	}{
		{
			name: "sessions",
			coll: SessionsCollection,
			init: NewSessionDAO(db).InitIndexes,
			want: []string{"idx_sessions_tenant_time", "idx_sessions_tenant_service", "idx_sessions_tenant_type"},
		},
		{
			name: "diagnoses",
			coll: DiagnosesCollection,
			init: NewDiagnosisDAO(db).InitIndexes,
			want: []string{"idx_dx_session", "idx_dx_tenant_service", "idx_dx_tenant_time"},
		},
		{
			name: "citations",
			coll: CitationsCollection,
			init: NewCitationDAO(db).InitIndexes,
			want: []string{"idx_cit_dx", "idx_cit_tenant"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 执行两次验证幂等（已存在同名索引不报错）
			if err := tc.init(ctx); err != nil {
				t.Fatalf("init indexes: %v", err)
			}
			if err := tc.init(ctx); err != nil {
				t.Fatalf("init indexes (idempotent): %v", err)
			}
			cur, err := db.Collection(tc.coll).Indexes().List(ctx)
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
			for _, want := range tc.want {
				if !got[want] {
					t.Fatalf("index %q missing, got %v", want, got)
				}
			}
		})
	}
}
