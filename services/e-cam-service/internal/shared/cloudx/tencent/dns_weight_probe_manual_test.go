// Package tencent_test DNSPod 记录 weight/status 实盘探测(manual probe,只读)。
// 目的:确认普通记录(未设加权)的 Weight 是 nil 还是 0,以及暂停记录如何表达。
package tencent_test

import (
	"context"
	"testing"
	"time"

	"github.com/Havens-blog/e-cam-service/internal/shared/cloudx/tencent"
	"github.com/gotomicro/ego/core/elog"
)

func TestProbeTencentDNSWeight(t *testing.T) {
	setupProbeEnv(t)
	ak, sk, _ := loadTencentCreds(t)
	if ak == "" {
		t.Skip("无 tencent 凭证,跳过")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	adapter := tencent.NewDNSAdapter(ak, sk, "ap-guangzhou", elog.DefaultLogger)
	domains, err := adapter.ListDomains(ctx)
	if err != nil {
		t.Fatalf("ListDomains 失败: %v", err)
	}
	probed := 0
	for _, d := range domains {
		if probed >= 3 {
			break
		}
		records, err := adapter.ListRecords(ctx, d.DomainName)
		if err != nil {
			t.Logf("[%s] ListRecords 失败: %v", d.DomainName, err)
			continue
		}
		t.Logf("=== domain %s (%d records) ===", d.DomainName, len(records))
		for _, r := range records {
			if r.Type != "A" && r.Type != "CNAME" {
				continue
			}
			t.Logf("  rr=%-20s type=%s status=%-8s weight=%d value=%s", r.RR, r.Type, r.Status, r.Weight, r.Value)
		}
		probed++
	}
	if probed == 0 {
		t.Log("无可用域名")
	}
}