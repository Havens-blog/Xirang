package risk

import (
	"errors"
	"testing"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// TestRegistry_Classify AC：注册表静态声明 read/low/high 三档，Classify 查表判定。
func TestRegistry_Classify(t *testing.T) {
	r := NewRegistry()
	cases := []struct {
		tool string
		lvl  domain.RiskLevel
	}{
		{"query_log", domain.RiskRead},
		{"diagnose", domain.RiskRead},
		{"query_alert", domain.RiskRead},
		{"query_asset", domain.RiskRead},
		{"notify", domain.RiskLow},
		{"restart_service", domain.RiskHigh},
	}
	for _, c := range cases {
		meta, err := r.Classify(c.tool)
		if err != nil {
			t.Fatalf("Classify(%q) err = %v", c.tool, err)
		}
		if meta.Name != c.tool || meta.RiskLevel != c.lvl {
			t.Fatalf("Classify(%q) = %+v, want name/level %q/%q", c.tool, meta, c.tool, c.lvl)
		}
		if meta.Allowlist {
			t.Fatalf("Classify(%q) Allowlist = true, want false (no whitelist yet)", c.tool)
		}
	}
}

// TestRegistry_Classify_Unregistered AC：未注册工具返回错误（不默认放行）。
func TestRegistry_Classify_Unregistered(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Classify("totally_unknown"); !errors.Is(err, ErrUnregistered) {
		t.Fatalf("err = %v, want ErrUnregistered", err)
	}
}

// TestRegistry_ApplyWhitelist AC：白名单匹配生效，Classify 反映 Allowlist=true。
func TestRegistry_ApplyWhitelist(t *testing.T) {
	r := NewRegistry()
	if err := r.ApplyWhitelist([]domain.RiskWhitelistEntry{
		{Tool: "notify", RiskLevel: domain.RiskLow},
	}); err != nil {
		t.Fatalf("apply whitelist: %v", err)
	}

	meta, err := r.Classify("notify")
	if err != nil || !meta.Allowlist {
		t.Fatalf("notify = %+v err=%v, want Allowlist=true", meta, err)
	}
	// 未入白名单的工具仍为 false。
	if meta, _ := r.Classify("query_log"); meta.Allowlist {
		t.Fatal("query_log should not be allowlisted")
	}
}

// TestRegistry_ApplyWhitelist_RejectUnregistered AC：白名单项必须命中注册表。
func TestRegistry_ApplyWhitelist_RejectUnregistered(t *testing.T) {
	r := NewRegistry()
	err := r.ApplyWhitelist([]domain.RiskWhitelistEntry{
		{Tool: "ghost_tool", RiskLevel: domain.RiskLow},
	})
	if !errors.Is(err, ErrUnregistered) {
		t.Fatalf("err = %v, want ErrUnregistered", err)
	}
}

// TestRegistry_ApplyWhitelist_RejectHigh AC：高危工具不允许入白名单。
func TestRegistry_ApplyWhitelist_RejectHigh(t *testing.T) {
	r := NewRegistry()
	err := r.ApplyWhitelist([]domain.RiskWhitelistEntry{
		{Tool: "restart_service", RiskLevel: domain.RiskRead},
	})
	if !errors.Is(err, ErrHighNotAllowlisted) {
		t.Fatalf("err = %v, want ErrHighNotAllowlisted", err)
	}
}

// TestRegistry_ApplyWhitelist_RejectHighLevel AC：条目 riskLevel=high 不允许。
func TestRegistry_ApplyWhitelist_RejectHighLevel(t *testing.T) {
	r := NewRegistry()
	err := r.ApplyWhitelist([]domain.RiskWhitelistEntry{
		{Tool: "query_log", RiskLevel: domain.RiskHigh},
	})
	if !errors.Is(err, ErrInvalidWhitelistLevel) {
		t.Fatalf("err = %v, want ErrInvalidWhitelistLevel", err)
	}
}

// TestRegistry_ApplyWhitelist_ReplaceAndAtomic 整体替换语义：成功替换旧条目；失败不部分生效。
func TestRegistry_ApplyWhitelist_ReplaceAndAtomic(t *testing.T) {
	r := NewRegistry()
	_ = r.ApplyWhitelist([]domain.RiskWhitelistEntry{{Tool: "notify", RiskLevel: domain.RiskLow}})

	// 替换：移除 notify，加入 query_log。
	if err := r.ApplyWhitelist([]domain.RiskWhitelistEntry{{Tool: "query_log", RiskLevel: domain.RiskRead}}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if meta, _ := r.Classify("notify"); meta.Allowlist {
		t.Fatal("notify should be removed after replace")
	}
	if meta, _ := r.Classify("query_log"); !meta.Allowlist {
		t.Fatal("query_log should be allowlisted after replace")
	}

	// 原子性：含非法条目的整体替换失败，旧白名单保持不变。
	err := r.ApplyWhitelist([]domain.RiskWhitelistEntry{
		{Tool: "query_log", RiskLevel: domain.RiskRead},
		{Tool: "restart_service", RiskLevel: domain.RiskRead}, // 非法 → 整体拒绝
	})
	if err == nil {
		t.Fatal("invalid batch should return error")
	}
	if meta, _ := r.Classify("query_log"); !meta.Allowlist {
		t.Fatal("query_log whitelist should survive failed replace")
	}
}
