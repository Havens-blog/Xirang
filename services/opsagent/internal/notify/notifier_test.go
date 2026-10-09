package notify

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// ---- fakes ----

// fakeSender 可控失败次数的单次发送器。
type fakeSender struct {
	failsRemaining int
	err            error
	calls          int
}

func (f *fakeSender) Send(_ context.Context, _ string, _ any) error {
	f.calls++
	if f.failsRemaining > 0 {
		f.failsRemaining--
		return f.err
	}
	return nil
}

// fakeMetrics 埋点记录器（success/failure 按渠道计数）。
type fakeMetrics struct {
	success map[string]int
	failure map[string]int
}

func newFakeMetrics() *fakeMetrics {
	return &fakeMetrics{success: map[string]int{}, failure: map[string]int{}}
}

func (m *fakeMetrics) IncSuccess(c string) { m.success[c]++ }
func (m *fakeMetrics) IncFailure(c string) { m.failure[c]++ }

func testPayload(channel string) domain.Notification {
	return domain.Notification{
		Tenant:      "tenant-A",
		Channel:     channel,
		Title:       "运维告警：order-service",
		Summary:     "时间窗内检测到异常日志",
		DiagLinkURL: "/opsagent/diagnosis/dx-1",
	}
}

// ---- Notifier tests ----

// TestSend_SuccessFirstAttempt 首试成功 → OK + attempts=1 + sentAt + 成功埋点。
func TestSend_SuccessFirstAttempt(t *testing.T) {
	m := newFakeMetrics()
	n := New(&fakeSender{}, m)
	res := n.Send(context.Background(), testPayload(ChannelDingTalk))
	if !res.OK || res.Attempts != 1 {
		t.Fatalf("want OK on first attempt, got %+v", res)
	}
	if res.SentAt.IsZero() {
		t.Fatal("sentAt should be set on success")
	}
	if res.LastError != "" {
		t.Fatalf("lastError should be empty on success, got %q", res.LastError)
	}
	if m.success[ChannelDingTalk] != 1 || m.failure[ChannelDingTalk] != 0 {
		t.Fatalf("metrics = %+v", m)
	}
}

// TestSend_SuccessAfterRetry 重试 2 次后成功 → attempts=3，仅记一次成功。
func TestSend_SuccessAfterRetry(t *testing.T) {
	m := newFakeMetrics()
	s := &fakeSender{failsRemaining: 2, err: errors.New("boom")}
	n := New(s, m)
	res := n.Send(context.Background(), testPayload(ChannelFeishu))
	if !res.OK || res.Attempts != 3 {
		t.Fatalf("want OK on 3rd attempt, got %+v", res)
	}
	if s.calls != 3 {
		t.Fatalf("calls = %d, want 3", s.calls)
	}
	if m.success[ChannelFeishu] != 1 || m.failure[ChannelFeishu] != 0 {
		t.Fatalf("metrics = %+v", m)
	}
}

// TestSend_AllFail 三次全败 → OK=false + attempts=max + lastError + 失败埋点（仅一次）。
func TestSend_AllFail(t *testing.T) {
	m := newFakeMetrics()
	s := &fakeSender{failsRemaining: 100, err: errors.New("timeout")}
	n := New(s, m)
	res := n.Send(context.Background(), testPayload(ChannelWecom))
	if res.OK {
		t.Fatal("want failure after max attempts")
	}
	if res.Attempts != maxAttempts {
		t.Fatalf("attempts = %d, want %d", res.Attempts, maxAttempts)
	}
	if res.LastError == "" {
		t.Fatal("lastError should be set on failure")
	}
	if !res.SentAt.IsZero() {
		t.Fatal("sentAt should stay zero on failure")
	}
	if m.failure[ChannelWecom] != 1 || m.success[ChannelWecom] != 0 {
		t.Fatalf("metrics = %+v", m)
	}
}

// TestSend_NilMetrics metrics 为 nil 时禁用埋点、不 panic。
func TestSend_NilMetrics(t *testing.T) {
	n := New(&fakeSender{}, nil)
	res := n.Send(context.Background(), testPayload(ChannelEmail))
	if !res.OK {
		t.Fatalf("want OK, got %+v", res)
	}
}

// TestSend_NilSender sender 为 nil 时终失败、不 panic。
func TestSend_NilSender(t *testing.T) {
	n := New(nil, newFakeMetrics())
	res := n.Send(context.Background(), testPayload(ChannelDingTalk))
	if res.OK || res.Attempts != 0 {
		t.Fatalf("nil sender should fail fast, got %+v", res)
	}
}

// ---- buildBody tests ----

// TestBuildBody_Channels 四渠道载荷形状 + 内容只含标题/摘要/链接、不含跨租户明细。
func TestBuildBody_Channels(t *testing.T) {
	cases := []struct {
		channel string
		check   func(t *testing.T, b map[string]any)
	}{
		{ChannelDingTalk, func(t *testing.T, b map[string]any) {
			if b["msgtype"] != "markdown" {
				t.Fatalf("dingtalk shape = %+v", b)
			}
		}},
		{ChannelWecom, func(t *testing.T, b map[string]any) {
			if b["msgtype"] != "markdown" {
				t.Fatalf("wecom shape = %+v", b)
			}
		}},
		{ChannelFeishu, func(t *testing.T, b map[string]any) {
			if b["msg_type"] != "text" {
				t.Fatalf("feishu shape = %+v", b)
			}
		}},
		{ChannelEmail, func(t *testing.T, b map[string]any) {
			if _, ok := b["subject"]; !ok {
				t.Fatalf("email shape = %+v", b)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.channel, func(t *testing.T) {
			b := buildBody(testPayload(tc.channel)).(map[string]any)
			tc.check(t, b)
			text := bodyText(b)
			for _, want := range []string{"运维告警：order-service", "时间窗内检测到异常日志", "/opsagent/diagnosis/dx-1"} {
				if !strings.Contains(text, want) {
					t.Fatalf("body missing %q: %s", want, text)
				}
			}
			if strings.Contains(text, "tenant-A") {
				t.Fatalf("body 泄跨租户明细: %s", text)
			}
		})
	}
}

// bodyText 从各渠道载荷中递归收集全部叶子字符串（跨 msgtype/title/text/subject/body 等键）。
func bodyText(b map[string]any) string {
	var sb strings.Builder
	var walk func(m map[string]any)
	walk = func(m map[string]any) {
		for _, v := range m {
			switch x := v.(type) {
			case string:
				sb.WriteString(x)
				sb.WriteString("\n")
			case map[string]any:
				walk(x)
			}
		}
	}
	walk(b)
	return sb.String()
}
