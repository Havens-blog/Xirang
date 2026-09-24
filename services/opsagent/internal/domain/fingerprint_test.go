package domain_test

import (
	"regexp"
	"testing"

	"github.com/Havens-blog/opsagent/internal/domain"
)

func TestFingerprintIsFirst32HexOfSHA256(t *testing.T) {
	// 独立基准值：sha256("tenantA" + "order-service" + "error_rate" + "P1")
	// = 858591a7d47f035c5e923bc807980d92...(sha256sum 实测)取前 32 位
	const want = "858591a7d47f035c5e923bc807980d92"
	got := domain.Fingerprint("tenantA", "order-service", "error_rate", "P1")
	if got != want {
		t.Errorf("Fingerprint() = %q, want %q", got, want)
	}
}

func TestFingerprintLengthAndHexCharset(t *testing.T) {
	got := domain.Fingerprint("t", "s", "m", "L")
	if len(got) != 32 {
		t.Errorf("len(Fingerprint()) = %d, want 32", len(got))
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(got) {
		t.Errorf("Fingerprint() = %q, want lowercase hex of 32 chars", got)
	}
}

func TestFingerprintDeterministic(t *testing.T) {
	a := domain.Fingerprint("t1", "order-service", "error_rate", "P1")
	b := domain.Fingerprint("t1", "order-service", "error_rate", "P1")
	if a != b {
		t.Errorf("same inputs produced different fingerprints: %q vs %q", a, b)
	}
}

func TestFingerprintInputSensitivity(t *testing.T) {
	base := domain.Fingerprint("t1", "svc", "error_rate", "P1")
	variants := []struct {
		name string
		fp   string
	}{
		{"tenant", domain.Fingerprint("t2", "svc", "error_rate", "P1")},
		{"service", domain.Fingerprint("t1", "svc2", "error_rate", "P1")},
		{"metric", domain.Fingerprint("t1", "svc", "latency", "P1")},
		{"level", domain.Fingerprint("t1", "svc", "error_rate", "P2")},
	}
	seen := map[string]bool{base: true}
	for _, v := range variants {
		if v.fp == base {
			t.Errorf("%s variation collided with base fingerprint", v.name)
		}
		if seen[v.fp] {
			t.Errorf("%s variation collided with another variant", v.name)
		}
		seen[v.fp] = true
	}
}
