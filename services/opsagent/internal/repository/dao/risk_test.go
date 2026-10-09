package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// newTestRiskEntry 构造一条合法风险中心条目（告警触发落库形态：version=1、
// status=pending_view，由 Insert 补齐默认值）。
func newTestRiskEntry(tenant, service, diagnosisID string) *domain.RiskEntry {
	return &domain.RiskEntry{
		Tenant:      tenant,
		Fingerprint: "fp-" + service,
		DiagnosisID: diagnosisID,
		ServiceName: service,
		Severity:    domain.SeverityP1,
	}
}

// mustInsertRisk 插入条目，失败即终止测试。
func mustInsertRisk(t *testing.T, d RiskEntryDAO, e *domain.RiskEntry) {
	t.Helper()
	if err := d.Insert(context.Background(), e); err != nil {
		t.Fatalf("insert risk entry: %v", err)
	}
}

// TestRiskDAO_InsertDefaults AC：条目落库默认值 —— version 初始 1（api-handbook
// expectedVersion 取值来源）、status=pending_view、时间戳补齐。
func TestRiskDAO_InsertDefaults(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	if in.ID == "" {
		t.Fatal("insert should generate entry ID")
	}
	out, err := d.GetByID(ctx, "tenant-a", in.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if out.Version != 1 {
		t.Fatalf("version = %d, want 1 (CAS 初始版本)", out.Version)
	}
	if out.Status != domain.RiskStatusPendingView {
		t.Fatalf("status = %q, want pending_view", out.Status)
	}
	if out.CreatedAt.IsZero() {
		t.Fatal("created_at should be filled on insert")
	}
	if out.Severity != domain.SeverityP1 || out.ServiceName != "order-service" {
		t.Fatalf("payload mismatch: %+v", out)
	}
}

// TestRiskDAO_GetByID_TenantIsolation Hard Rule：CAS/查询过滤必须含 tenant_id，
// 跨租户与不存在统一 ErrNoDocuments（不泄露存在性）。
func TestRiskDAO_GetByID_TenantIsolation(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	if _, err := d.GetByID(ctx, "tenant-b", in.ID); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant get err = %v, want ErrNoDocuments", err)
	}
	if _, err := d.GetByID(ctx, "tenant-a", "ffffffffffffffffffffffff"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("missing-id get err = %v, want ErrNoDocuments", err)
	}
	if _, err := d.GetByID(ctx, "tenant-a", "not-an-objectid"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("malformed-id get err = %v, want ErrNoDocuments", err)
	}
}

// TestRiskDAO_UpdateStatus_FullFlow AC-1/AC-2：合法流转 pending_view → viewed →
// done 经 CAS findAndModify，version 每次 $inc（1→2→3），status_history 逐次追加
// from/to/by，并返回含自增 version 的最新条目。
func TestRiskDAO_UpdateStatus_FullFlow(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	viewed, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusViewed, 1, "user-1")
	if err != nil {
		t.Fatalf("mark viewed: %v", err)
	}
	if viewed.Status != domain.RiskStatusViewed || viewed.Version != 2 {
		t.Fatalf("after viewed: status=%q version=%d, want viewed/2", viewed.Status, viewed.Version)
	}
	if len(viewed.StatusHistory) != 1 {
		t.Fatalf("status_history len = %d, want 1", len(viewed.StatusHistory))
	}
	h := viewed.StatusHistory[0]
	if h.From != string(domain.RiskStatusPendingView) || h.To != string(domain.RiskStatusViewed) || h.By != "user-1" {
		t.Fatalf("history[0] = %+v, want pending_view→viewed by user-1", h)
	}
	if h.At.IsZero() {
		t.Fatal("history[0].at should be set")
	}

	done, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusDone, 2, "user-2")
	if err != nil {
		t.Fatalf("mark done: %v", err)
	}
	if done.Status != domain.RiskStatusDone || done.Version != 3 {
		t.Fatalf("after done: status=%q version=%d, want done/3", done.Status, done.Version)
	}
	if len(done.StatusHistory) != 2 || done.StatusHistory[1].By != "user-2" {
		t.Fatalf("status_history = %+v, want 2 entries with by=user-2 last", done.StatusHistory)
	}

	// 持久化确认：CAS 返回条目即库内最新状态。
	persisted, err := d.GetByID(ctx, "tenant-a", in.ID)
	if err != nil {
		t.Fatalf("get persisted: %v", err)
	}
	if persisted.Version != 3 || persisted.Status != domain.RiskStatusDone {
		t.Fatalf("persisted = %q/%d, want done/3", persisted.Status, persisted.Version)
	}
}

// TestRiskDAO_UpdateStatus_VersionConflict AC-3：version 不匹配返回冲突 +
// 最新条目（供 service 层映射 409 ERR_CONFLICT 并刷新）。
func TestRiskDAO_UpdateStatus_VersionConflict(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	// 并发对手先完成 viewed 流转，版本推进到 2。
	if _, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusViewed, 1, "rival"); err != nil {
		t.Fatalf("rival update: %v", err)
	}

	// 本端仍持陈旧 version=1 提交 done → 冲突，返回最新条目。
	got, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusDone, 1, "user-1")
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("err = %v, want *ConflictError", err)
	}
	if conflict.Current.Version != 2 || conflict.Current.Status != domain.RiskStatusViewed {
		t.Fatalf("conflict.Current = %q/%d, want viewed/2", conflict.Current.Status, conflict.Current.Version)
	}
	if got.Version != 2 || got.Status != domain.RiskStatusViewed {
		t.Fatalf("returned entry = %q/%d, want latest viewed/2", got.Status, got.Version)
	}
	// 冲突不得推进版本：库内仍是 2。
	persisted, _ := d.GetByID(ctx, "tenant-a", in.ID)
	if persisted.Version != 2 {
		t.Fatalf("persisted version = %d after conflict, want 2 (no overwrite)", persisted.Version)
	}
}

// TestRiskDAO_UpdateStatus_NoVersionCheck api-handbook §4：expectedVersion 不传
// （DAO 层 0 值）则不做版本校验，仅凭状态机流转。
func TestRiskDAO_UpdateStatus_NoVersionCheck(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	got, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusDone, 0, "user-1")
	if err != nil {
		t.Fatalf("update without version check: %v", err)
	}
	if got.Status != domain.RiskStatusDone || got.Version != 2 {
		t.Fatalf("got %q/%d, want done/2", got.Status, got.Version)
	}
}

// TestRiskDAO_UpdateStatus_IdempotentSameStatus 同状态重复提交为幂等 no-op：
// 状态已是目标值时不推进 version、不追加 history。
func TestRiskDAO_UpdateStatus_IdempotentSameStatus(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	if _, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusViewed, 1, "user-1"); err != nil {
		t.Fatalf("first mark: %v", err)
	}
	got, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusViewed, 2, "user-2")
	if err != nil {
		t.Fatalf("duplicate mark: %v", err)
	}
	if got.Version != 2 || len(got.StatusHistory) != 1 {
		t.Fatalf("idempotent mark mutated entry: version=%d history=%d", got.Version, len(got.StatusHistory))
	}
}

// TestRiskDAO_UpdateStatus_InvalidTransition 状态机约束：done 不可回退为
// viewed；目标状态非法（如回退 pending_view）直接拒绝。
func TestRiskDAO_UpdateStatus_InvalidTransition(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)
	if _, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusDone, 0, "user-1"); err != nil {
		t.Fatalf("mark done: %v", err)
	}

	if _, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusViewed, 2, "user-1"); !errors.Is(err, errInvalidRiskTransition) {
		t.Fatalf("done→viewed err = %v, want errInvalidRiskTransition", err)
	}
	if _, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusPendingView, 2, "user-1"); err == nil {
		t.Fatal("target pending_view should be rejected")
	}
}

// TestRiskDAO_UpdateStatus_NotFound 不存在/跨租户统一 ErrNoDocuments。
func TestRiskDAO_UpdateStatus_NotFound(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	if _, err := d.UpdateStatus(ctx, "tenant-b", in.ID, domain.RiskStatusViewed, 0, "user-1"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant update err = %v, want ErrNoDocuments", err)
	}
	if _, err := d.UpdateStatus(ctx, "tenant-a", "ffffffffffffffffffffffff", domain.RiskStatusViewed, 0, "user-1"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("missing-id update err = %v, want ErrNoDocuments", err)
	}
}

// TestRiskDAO_MarkViewed_Lazy AC-5：诊断详情惰性自动已读 —— 过滤
// {_id, tenant_id, diagnosis_id, status:pending_view} → viewed（$inc version +
// history），幂等且不关联/不存在时不流转。
func TestRiskDAO_MarkViewed_Lazy(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "dx-1")
	mustInsertRisk(t, d, in)

	// 惰性流转：pending_view → viewed，by=system（服务端自动）。
	got, changed, err := d.MarkViewed(ctx, "tenant-a", in.ID, "dx-1", "user-1")
	if err != nil {
		t.Fatalf("mark viewed: %v", err)
	}
	if !changed {
		t.Fatal("first lazy view should change status")
	}
	if got.Status != domain.RiskStatusViewed || got.Version != 2 {
		t.Fatalf("got %q/%d, want viewed/2", got.Status, got.Version)
	}
	if len(got.StatusHistory) != 1 || got.StatusHistory[0].From != string(domain.RiskStatusPendingView) {
		t.Fatalf("history = %+v, want pending_view→viewed", got.StatusHistory)
	}

	// 幂等：重复触发不再流转。
	again, changed, err := d.MarkViewed(ctx, "tenant-a", in.ID, "dx-1", "user-1")
	if err != nil || changed {
		t.Fatalf("repeat lazy view: changed=%v err=%v, want false/nil", changed, err)
	}
	if again.Version != 2 || len(again.StatusHistory) != 1 {
		t.Fatalf("repeat mutated entry: version=%d history=%d", again.Version, len(again.StatusHistory))
	}

	// 不关联本诊断：忽略流转，按无参数处理（changed=false）。
	other := newTestRiskEntry("tenant-a", "pay-service", "dx-2")
	mustInsertRisk(t, d, other)
	if _, changed, err := d.MarkViewed(ctx, "tenant-a", other.ID, "dx-wrong", "user-1"); err != nil || changed {
		t.Fatalf("mismatched diagnosis: changed=%v err=%v, want false/nil", changed, err)
	}

	// 不存在/跨租户：ErrNoDocuments（上层按无参数处理，不泄露存在性）。
	if _, _, err := d.MarkViewed(ctx, "tenant-a", "ffffffffffffffffffffffff", "dx-1", "user-1"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("missing entry err = %v, want ErrNoDocuments", err)
	}
	if _, _, err := d.MarkViewed(ctx, "tenant-b", in.ID, "dx-1", "user-1"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("cross-tenant mark viewed err = %v, want ErrNoDocuments", err)
	}
}

// TestRiskDAO_BatchUpdateStatus AC-4：批量逐条独立 CAS，成功/失败分别聚合；
// 冲突条目不覆盖并发单标，仅记账 failed（reasonCode=conflict/not_found）。
func TestRiskDAO_BatchUpdateStatus(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	ok := newTestRiskEntry("tenant-a", "order-service", "dx-a")
	mustInsertRisk(t, d, ok)
	stale := newTestRiskEntry("tenant-a", "pay-service", "dx-b")
	mustInsertRisk(t, d, stale)
	missing := newTestRiskEntry("tenant-a", "risk-service", "dx-c")
	mustInsertRisk(t, d, missing)

	// 并发单标抢先推进 stale 到 viewed/2（批量请求仍持 version=1）。
	if _, err := d.UpdateStatus(ctx, "tenant-a", stale.ID, domain.RiskStatusViewed, 1, "rival"); err != nil {
		t.Fatalf("rival single mark: %v", err)
	}

	res, err := d.BatchUpdateStatus(ctx, "tenant-a", domain.RiskStatusDone, []BatchStatusItem{
		{ID: ok.ID, ExpectedVersion: 1},
		{ID: stale.ID, ExpectedVersion: 1},
		{ID: "ffffffffffffffffffffffff", ExpectedVersion: 1},
	})
	if err != nil {
		t.Fatalf("batch update: %v", err)
	}

	if len(res.Succeeded) != 1 || res.Succeeded[0].ID != ok.ID || res.Succeeded[0].Version != 2 {
		t.Fatalf("succeeded = %+v, want [ok/2]", res.Succeeded)
	}
	if len(res.Failed) != 2 {
		t.Fatalf("failed = %+v, want 2 entries", res.Failed)
	}
	byID := map[string]BatchStatusFailure{}
	for _, f := range res.Failed {
		byID[f.ID] = f
	}
	conf, isConf := byID[stale.ID]
	if !isConf || conf.ReasonCode != "conflict" || conf.CurrentVersion != 2 {
		t.Fatalf("stale failure = %+v, want conflict with currentVersion=2", conf)
	}
	nf, isNF := byID["ffffffffffffffffffffffff"]
	if !isNF || nf.ReasonCode != "not_found" {
		t.Fatalf("missing failure = %+v, want not_found", nf)
	}

	// 冲突不覆盖：stale 仍是 rival 单标结果 viewed/2，批量未写入 done。
	persisted, _ := d.GetByID(ctx, "tenant-a", stale.ID)
	if persisted.Status != domain.RiskStatusViewed || persisted.Version != 2 {
		t.Fatalf("conflicted entry = %q/%d, want untouched viewed/2", persisted.Status, persisted.Version)
	}
}

// TestRiskDAO_List_Filters AC（api-handbook §3 列表语义）：created_at 倒序，
// 支持 status/severity/serviceName（模糊匹配）筛选与分页 + total。
func TestRiskDAO_List_Filters(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	base := time.Now().Add(-time.Hour)
	entries := []*domain.RiskEntry{
		func() *domain.RiskEntry {
			e := newTestRiskEntry("tenant-a", "order-service", "")
			e.Severity = domain.SeverityP0
			return e
		}(),
		newTestRiskEntry("tenant-a", "pay-service", ""),
		newTestRiskEntry("tenant-a", "order-gateway", ""),
		newTestRiskEntry("tenant-b", "order-service", ""), // 他租户，不可见
	}
	for i, e := range entries {
		e.CreatedAt = base.Add(time.Duration(i) * time.Minute)
		mustInsertRisk(t, d, e)
	}

	all, total, err := d.List(ctx, "tenant-a", RiskEntryFilter{}, 0, 0)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if total != 3 || len(all) != 3 {
		t.Fatalf("total=%d len=%d, want 3/3", total, len(all))
	}
	// created_at 倒序（最新在前）。
	if all[0].ServiceName != "order-gateway" || all[2].ServiceName != "order-service" {
		t.Fatalf("sort order wrong: %s, %s, %s", all[0].ServiceName, all[1].ServiceName, all[2].ServiceName)
	}

	// severity 精确筛选。
	p0s, total, err := d.List(ctx, "tenant-a", RiskEntryFilter{Severity: domain.SeverityP0}, 0, 0)
	if err != nil || total != 1 || len(p0s) != 1 || p0s[0].Severity != domain.SeverityP0 {
		t.Fatalf("severity filter: total=%d err=%v", total, err)
	}

	// serviceName 模糊匹配（子串 order）。
	orders, total, err := d.List(ctx, "tenant-a", RiskEntryFilter{ServiceName: "order"}, 0, 0)
	if err != nil || total != 2 || len(orders) != 2 {
		t.Fatalf("fuzzy service filter: total=%d len=%d err=%v", total, len(orders), err)
	}

	// 分页：limit=2, offset=1。
	page, total, err := d.List(ctx, "tenant-a", RiskEntryFilter{}, 2, 1)
	if err != nil || total != 3 || len(page) != 2 || page[0].ServiceName != "pay-service" {
		t.Fatalf("pagination: total=%d page=%+v err=%v", total, page, err)
	}

	// status 筛选（含流转后状态）。
	if _, err := d.UpdateStatus(ctx, "tenant-a", entries[1].ID, domain.RiskStatusViewed, 0, "user-1"); err != nil {
		t.Fatalf("mark viewed: %v", err)
	}
	viewed, total, err := d.List(ctx, "tenant-a", RiskEntryFilter{Status: domain.RiskStatusViewed}, 0, 0)
	if err != nil || total != 1 || len(viewed) != 1 || viewed[0].Status != domain.RiskStatusViewed {
		t.Fatalf("status filter: total=%d err=%v", total, err)
	}
}

// TestRiskDAO_UpdatedAtPersisted er-diagram：updated_at 为 NOT NULL 存储字段，
// Insert 与每次 CAS 更新均需刷新（domain 形状不携带，经存储侧补充）。
func TestRiskDAO_UpdatedAtPersisted(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	in := newTestRiskEntry("tenant-a", "order-service", "")
	mustInsertRisk(t, d, in)

	assertUpdatedAt := func(stage string) time.Time {
		t.Helper()
		var raw bson.M
		err := db.Collection(RiskEntriesCollection).FindOne(ctx, bson.M{"_id": in.ID}).Decode(&raw)
		if err != nil {
			t.Fatalf("%s: read raw doc: %v", stage, err)
		}
		// bson.M 原始解码将 BSON 日期还原为 primitive.DateTime，经 .Time() 取值。
		dt, ok := raw["updated_at"].(primitive.DateTime)
		if !ok {
			t.Fatalf("%s: updated_at missing or wrong type (%T)", stage, raw["updated_at"])
		}
		return dt.Time()
	}
	first := assertUpdatedAt("insert")

	if _, err := d.UpdateStatus(ctx, "tenant-a", in.ID, domain.RiskStatusViewed, 1, "user-1"); err != nil {
		t.Fatalf("update: %v", err)
	}
	second := assertUpdatedAt("update")
	if !second.After(first) {
		t.Fatal("updated_at should advance after CAS update")
	}
}

// TestRiskDAO_InitIndexes 索引名与 schema.mongo.js 一致：
// idx_re_tenant_status / idx_re_tenant_service / idx_re_tenant_fingerprint / idx_re_dx。
func TestRiskDAO_InitIndexes(t *testing.T) {
	db := newTestMongo(t)
	d := NewRiskEntryDAO(db)
	ctx := context.Background()

	// 执行两次验证幂等。
	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes: %v", err)
	}
	if err := d.InitIndexes(ctx); err != nil {
		t.Fatalf("init indexes (idempotent): %v", err)
	}
	cur, err := db.Collection(RiskEntriesCollection).Indexes().List(ctx)
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
	for _, want := range []string{"idx_re_tenant_status", "idx_re_tenant_service", "idx_re_tenant_fingerprint", "idx_re_dx"} {
		if !got[want] {
			t.Fatalf("index %q missing, got %v", want, got)
		}
	}
}
