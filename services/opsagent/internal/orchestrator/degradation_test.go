package orchestrator

import (
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// confidentIntent 构造高置信、含服务名的意图结果（L1 判定的前提）。
func confidentIntent() domain.IntentResult {
	return domain.IntentResult{
		Type:       domain.IntentDiagnose,
		Confidence: 0.9,
		Params:     domain.DiagnoseParams{ServiceName: "order-service"},
	}
}

func lowConfIntent() domain.IntentResult {
	return domain.IntentResult{
		Type:       domain.IntentDiagnose,
		Confidence: 0.4,
		Params:     domain.DiagnoseParams{ServiceName: ""},
	}
}

func samplePresets() []domain.PresetQuery {
	return []domain.PresetQuery{
		{ID: "p1", Label: "order-service 近1h错误率", Params: domain.DiagnoseParams{ServiceName: "order-service"}},
		{ID: "p2", Label: "cdn_edge 延迟", Params: domain.DiagnoseParams{ServiceName: "cdn_edge"}},
	}
}

func newTestController(now func() time.Time) *DegradationController {
	return NewDegradationController(5*time.Minute, 0.6, now)
}

// TestDecide_L0 LLM 探活成功 → L0 正常，不降级。
func TestDecide_L0(t *testing.T) {
	c := newTestController(nil)
	d := c.Decide("s1", func() bool { return true }, confidentIntent(), nil)
	if d.Level != LevelNormal || d.Degraded || d.Type != respTypeReport {
		t.Fatalf("decision = %+v, want L0 normal", d)
	}
}

// TestDecide_L1 LLM 不可用 + 高置信服务名 → L1 模板报告。
func TestDecide_L1(t *testing.T) {
	c := newTestController(nil)
	d := c.Decide("s1", func() bool { return false }, confidentIntent(), nil)
	if d.Level != LevelTemplate || !d.Degraded || d.Type != respTypeReport {
		t.Fatalf("decision = %+v, want L1 degraded report", d)
	}
}

// TestDecide_L2_Confidence 低置信 + 预置目录非空 → L2 预置查询入口。
func TestDecide_L2_Confidence(t *testing.T) {
	c := newTestController(nil)
	d := c.Decide("s1", func() bool { return false }, lowConfIntent(), samplePresets())
	if d.Level != LevelPreset || !d.Degraded || d.Type != respTypePresetEntries {
		t.Fatalf("decision = %+v, want L2 preset_entries", d)
	}
}

// TestDecide_L2_NoServiceName 高置信但无服务名 + 预置目录 → L2。
func TestDecide_L2_NoServiceName(t *testing.T) {
	c := newTestController(nil)
	intent := domain.IntentResult{Type: domain.IntentDiagnose, Confidence: 0.8, Params: domain.DiagnoseParams{ServiceName: ""}}
	d := c.Decide("s1", func() bool { return false }, intent, samplePresets())
	if d.Level != LevelPreset || d.Type != respTypePresetEntries {
		t.Fatalf("decision = %+v, want L2 (no service name)", d)
	}
}

// TestDecide_L3 低置信 + 预置目录空 → L3 明示降级。
func TestDecide_L3(t *testing.T) {
	c := newTestController(nil)
	d := c.Decide("s1", func() bool { return false }, lowConfIntent(), nil)
	if d.Level != LevelNotice || !d.Degraded || d.Type != respTypeDegradedNotice {
		t.Fatalf("decision = %+v, want L3 degraded_notice", d)
	}
}

// TestDecide_CooldownSkipsProbe 冷却窗口内不再探活：第二次 Decide 不触发 probe。
func TestDecide_CooldownSkipsProbe(t *testing.T) {
	base := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	now := base
	c := newTestController(func() time.Time { return now })

	probes := 0
	// 首次：探活失败 → 进入降级。
	d1 := c.Decide("s1", func() bool { probes++; return false }, lowConfIntent(), samplePresets())
	if d1.Level != LevelPreset {
		t.Fatalf("first decision = %+v, want L2", d1)
	}
	// 冷却期内：probe 应不被调用，仍维持 L2。
	d2 := c.Decide("s1", func() bool { probes++; return true }, lowConfIntent(), samplePresets())
	if d2.Level != LevelPreset {
		t.Fatalf("in-cooldown decision = %+v, want L2 (no upgrade)", d2)
	}
	if probes != 1 {
		t.Fatalf("probe calls = %d, want 1 (cooldown must skip probe)", probes)
	}
}

// TestDecide_RecoveryAfterCooldown 冷却过期后探活成功 → 回升 L0。
func TestDecide_RecoveryAfterCooldown(t *testing.T) {
	base := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	now := base
	c := newTestController(func() time.Time { return now })

	// 进入降级。
	if d := c.Decide("s1", func() bool { return false }, lowConfIntent(), nil); d.Level != LevelNotice {
		t.Fatalf("first decision = %+v, want L3", d)
	}
	// 冷却过期。
	now = base.Add(6 * time.Minute)
	d := c.Decide("s1", func() bool { return true }, confidentIntent(), nil)
	if d.Level != LevelNormal || d.Degraded {
		t.Fatalf("post-cooldown decision = %+v, want L0 recovered", d)
	}
}

// TestDecide_ResetCooldownOnRepeatedFailure 冷却过期后探活仍失败 → 维持降级并重置冷却。
func TestDecide_ResetCooldownOnRepeatedFailure(t *testing.T) {
	base := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	now := base
	c := newTestController(func() time.Time { return now })

	c.Decide("s1", func() bool { return false }, lowConfIntent(), samplePresets()) // L2
	now = base.Add(6 * time.Minute)                                                // 冷却过期
	probes := 0
	d := c.Decide("s1", func() bool { probes++; return false }, lowConfIntent(), samplePresets())
	if d.Level != LevelPreset || probes != 1 {
		t.Fatalf("decision = %+v (probes=%d), want L2 with fresh probe", d, probes)
	}
	// 仍在新的冷却窗口内：不应再次探活。
	d2 := c.Decide("s1", func() bool { probes++; return true }, lowConfIntent(), samplePresets())
	if d2.Level != LevelPreset || probes != 1 {
		t.Fatalf("decision = %+v (probes=%d), want cooldown reset so no re-probe", d2, probes)
	}
}

// TestDegradeDecision_Stamp 降级级别计入 TraceStep.DegradeLevel。
func TestDegradeDecision_Stamp(t *testing.T) {
	step := domain.TraceStep{Step: 5, Agent: agentCoordinator, Action: "report"}
	DegradeDecision{Level: LevelPreset, Degraded: true, Type: respTypePresetEntries}.Stamp(&step)
	if step.DegradeLevel != int(LevelPreset) {
		t.Fatalf("DegradeLevel = %d, want %d", step.DegradeLevel, int(LevelPreset))
	}
}

// TestNewDegradationController_Defaults 非法参数（冷却/阈值越界、now 缺失）回退默认值。
func TestNewDegradationController_Defaults(t *testing.T) {
	c := NewDegradationController(0, 0, nil)
	if c.cooldown != defaultCooldown || c.threshold != defaultThreshold {
		t.Fatalf("defaults = cooldown %v / threshold %v, want %v / %v",
			c.cooldown, c.threshold, defaultCooldown, defaultThreshold)
	}
	if c.now == nil {
		t.Fatal("now should default to time.Now")
	}
	if len(c.downgrade) != 0 {
		t.Fatal("downgrade map should be initialized empty")
	}
	// 阈值越界 >1 回退默认。
	if c2 := NewDegradationController(0, 1.5, nil); c2.threshold != defaultThreshold {
		t.Fatalf("threshold = %v, want default %v", c2.threshold, defaultThreshold)
	}
}

// TestDegradeLevel_Ordering 级别常量维持 0..3 递增语义（只降不升依赖）。
func TestDegradeLevel_Ordering(t *testing.T) {
	if LevelNormal >= LevelTemplate || LevelTemplate >= LevelPreset || LevelPreset >= LevelNotice {
		t.Fatalf("degrade levels must be strictly increasing 0..3, got %d/%d/%d/%d",
			LevelNormal, LevelTemplate, LevelPreset, LevelNotice)
	}
}
