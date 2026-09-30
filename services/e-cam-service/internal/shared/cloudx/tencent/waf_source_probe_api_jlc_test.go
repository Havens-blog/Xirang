// Package tencent_test 腾讯 WAF 源站回捞实盘探测(manual probe,只读)。
//
// 目的:实锤 api.jlc.com 的 WAF source_ips=null 根因——是列表接口 SrcList/UpstreamDomainList 空、
// 还是详情 DescribeDomainDetailsSaas 的 DomainsPartInfo 无 UpstreamDomain/SrcList、还是详情接口失败。
// 只读:只打印,不动生产表。
package tencent_test

import (
	"context"
	"testing"
	"time"

	"github.com/Havens-blog/e-cam-service/internal/shared/cloudx/tencent"
	"github.com/gotomicro/ego/core/elog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	waf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf/v20180125"
)

func TestProbeTencentWAFSourceapi_jlc(t *testing.T) {
	setupProbeEnv(t)
	ak, sk, _ := loadTencentCreds(t)
	if ak == "" {
		t.Skip("无 tencent 凭证,跳过")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 1. 走适配器,看 api.jlc.com 处理后 SourceIPs 是否为空
	adapter := tencent.NewWAFAdapter(ak, sk, "ap-guangzhou", elog.DefaultLogger)
	instances, err := adapter.ListInstances(ctx, "ap-guangzhou")
	if err != nil {
		t.Fatalf("ListInstances 失败: %v", err)
	}
	for _, ins := range instances {
		if ins.InstanceName == "api.jlc.com" {
			t.Logf("[adapter] id=%s edition=%s cname=%s sourceIPs=%v desc=%q state_guard=%v",
				ins.InstanceID, ins.Edition, ins.Cname, ins.SourceIPs, ins.Description, ins.Status)
		}
	}

	// 2. 直连 SDK,分别 dump 列表 + 详情原始返回
	credential := common.NewCredential(ak, sk)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "waf.tencentcloudapi.com"
	client, err := waf.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		t.Fatalf("创建 WAF client 失败: %v", err)
	}

	// 2a. DescribeDomains 列表,找 api.jlc.com
	var target *waf.DomainInfo
	offset := uint64(0)
	for {
		req := waf.NewDescribeDomainsRequest()
		req.Offset = &offset
		req.Limit = func() *uint64 { v := uint64(100); return &v }()
		resp, err := client.DescribeDomains(req)
		if err != nil {
			t.Fatalf("DescribeDomains 失败: %v", err)
		}
		if resp.Response == nil || resp.Response.Domains == nil {
			break
		}
		for _, d := range resp.Response.Domains {
			if d.Domain != nil && *d.Domain == "api.jlc.com" {
				target = d
				break
			}
		}
		if target != nil || len(resp.Response.Domains) < 100 {
			break
		}
		offset += 100
	}
	if target == nil {
		t.Fatalf("列表接口未找到 api.jlc.com")
	}

	t.Logf("[list] domain=%s id=%s edition=%s instanceId=%s",
		sd(target.Domain), sd(target.DomainId), sd(target.Edition), sd(target.InstanceId))
	var srcList, upList []string
	for _, s := range target.SrcList {
		srcList = append(srcList, sd(s))
	}
	for _, s := range target.UpstreamDomainList {
		upList = append(upList, sd(s))
	}
	t.Logf("[list] srcList=%v upstreamDomainList=%v cname=%s state=%v",
		srcList, upList, sd(target.Cname), target.State)

	// 2b. DescribeDomainDetailsSaas 详情
	detReq := waf.NewDescribeDomainDetailsSaasRequest()
	detReq.Domain = target.Domain
	detReq.DomainId = target.DomainId
	if target.InstanceId != nil {
		detReq.InstanceId = target.InstanceId
	}
	detResp, err := client.DescribeDomainDetailsSaas(detReq)
	if err != nil {
		t.Fatalf("DescribeDomainDetailsSaas 失败: %v", err)
	}
	if detResp == nil || detResp.Response == nil || detResp.Response.DomainsPartInfo == nil {
		t.Fatalf("详情返回空 DomainsPartInfo")
	}
	di := detResp.Response.DomainsPartInfo
	var dSrc []string
	for _, s := range di.SrcList {
		dSrc = append(dSrc, sd(s))
	}
	t.Logf("[detail] srcList=%v upstreamDomain=%q upstreamType=%v cname=%s mode=%v createTime=%q",
		dSrc, sd(di.UpstreamDomain), di.UpstreamType, sd(di.Cname), di.Mode, sd(di.CreateTime))
}

func sd(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}