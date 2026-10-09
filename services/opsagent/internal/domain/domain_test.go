package domain_test

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/Havens-blog/opsagent/internal/domain"
)

// jsonKeys 序列化 v 后返回顶层 JSON key 集合（排序后）。
func jsonKeys(t *testing.T, v any) []string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal %T: %v", v, err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func assertKeys(t *testing.T, name string, got, want []string) {
	t.Helper()
	sort.Strings(want)
	if len(got) != len(want) {
		t.Errorf("%s: key count = %d, want %d\ngot:  %v\nwant: %v", name, len(got), len(want), got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s: key[%d] = %q, want %q\ngot:  %v\nwant: %v", name, i, got[i], want[i], got, want)
		}
	}
}

// ---- api-handbook Data Contracts / tech-design Shared Types 逐字段 json tag 校验 ----

func TestJSONTagsDiagnosis(t *testing.T) {
	d := domain.Diagnosis{
		ID:          "dx1",
		SessionID:   "s1",
		Tenant:      "t1",
		RootCause:   "rc",
		Confidence:  0.9,
		Severity:    domain.SeverityP1,
		RiskLevel:   domain.RiskRead,
		Conclusions: []domain.Conclusion{{Text: "c", Citation: []string{"k1"}}},
		Disposition: []domain.DispositionStep{{Step: "s", Risk: "read", Action: "a", Executed: false}},
		Citations:   []domain.Citation{{SourceType: "log", SourceKey: "k1", Snippet: "sn"}},
		Degraded:    false,
		Truncated:   false,
		Trace:       []domain.TraceStep{{Step: 1, Agent: "coordinator", Action: "intent", Source: "llm", DurationMs: 10, Summary: "ok", DegradeLevel: 1}},
	}
	assertKeys(t, "Diagnosis", jsonKeys(t, d), []string{
		"id", "sessionId", "tenant", "rootCause", "confidence", "severity", "riskLevel",
		"conclusions", "disposition", "citations", "degraded", "truncated", "trace",
	})
	assertKeys(t, "Diagnosis.Conclusion", jsonKeys(t, d.Conclusions[0]), []string{"text", "citation"})
	assertKeys(t, "Diagnosis.DispositionStep", jsonKeys(t, d.Disposition[0]), []string{"step", "risk", "action", "executed"})
	assertKeys(t, "Diagnosis.Citation", jsonKeys(t, d.Citations[0]), []string{"sourceType", "sourceKey", "snippet"})
	assertKeys(t, "Diagnosis.TraceStep", jsonKeys(t, d.Trace[0]), []string{"step", "agent", "action", "source", "durationMs", "summary", "degradeLevel"})
}

func TestJSONTagsRiskEntry(t *testing.T) {
	re := domain.RiskEntry{
		ID:            "r1",
		Tenant:        "t1",
		Fingerprint:   "fp",
		DiagnosisID:   "dx1",
		ServiceName:   "order-service",
		Severity:      domain.SeverityP0,
		Status:        domain.RiskStatusPendingView,
		Version:       1,
		StatusHistory: []domain.StatusChange{{From: "pending_view", To: "viewed", By: "u1", At: timeNow()}},
		Notified:      domain.NotifyResult{Channel: "dingtalk", OK: true, Attempts: 1, LastError: "timeout", SentAt: timeNow()},
		CreatedAt:     timeNow(),
	}
	assertKeys(t, "RiskEntry", jsonKeys(t, re), []string{
		"id", "tenant", "fingerprint", "diagnosisId", "serviceName", "severity",
		"status", "version", "statusHistory", "notified", "createdAt",
	})
	assertKeys(t, "RiskEntry.StatusChange", jsonKeys(t, re.StatusHistory[0]), []string{"from", "to", "by", "at"})
	assertKeys(t, "RiskEntry.Notified", jsonKeys(t, re.Notified), []string{"channel", "ok", "attempts", "lastError", "sentAt"})
}

func TestJSONTagsIntentTypes(t *testing.T) {
	dp := domain.DiagnoseParams{
		ServiceName: "order-service",
		Metric:      "error_rate",
		Timeframe:   domain.Timeframe{StartTime: timeNow(), EndTime: timeNow()},
		LogTypes:    []string{"cdn"},
	}
	assertKeys(t, "DiagnoseParams", jsonKeys(t, dp), []string{"serviceName", "metric", "timeframe", "logTypes"})
	assertKeys(t, "Timeframe", jsonKeys(t, dp.Timeframe), []string{"startTime", "endTime"})

	ic := domain.IntentCandidate{Type: domain.IntentDiagnose, Confidence: 0.8, Params: dp}
	assertKeys(t, "IntentCandidate", jsonKeys(t, ic), []string{"type", "confidence", "params"})
}

func TestJSONTagsBaseContractTypes(t *testing.T) {
	le := domain.LogEntry{
		Timestamp: timeNow(),
		Cloud:     "aliyun",
		LogType:   "cdn",
		Resource:  "order-service",
		Level:     "error",
		Message:   "m",
		Fields:    map[string]any{"k": "v"},
	}
	assertKeys(t, "LogEntry", jsonKeys(t, le), []string{"timestamp", "cloud", "logType", "resource", "level", "message", "fields"})

	ar := domain.AssetRecord{AssetID: "a1", Type: "ecs", Tenant: "t1", Name: "n", Region: "cn", Status: "ok", Raw: map[string]any{"k": "v"}}
	assertKeys(t, "AssetRecord", jsonKeys(t, ar), []string{"assetId", "type", "tenant", "name", "region", "status", "raw"})

	id := domain.Identity{UserID: "u1", Tenant: "t1", Roles: []string{"admin"}}
	assertKeys(t, "Identity", jsonKeys(t, id), []string{"userId", "tenant", "roles"})

	n := domain.Notification{Tenant: "t1", Channel: "dingtalk", Title: "ti", Summary: "su", DiagLinkURL: "http://x"}
	assertKeys(t, "Notification", jsonKeys(t, n), []string{"tenant", "channel", "title", "summary", "diagLinkUrl"})

	nr := domain.NotifyResult{Channel: "email", OK: false, Attempts: 3, LastError: "e", SentAt: timeNow()}
	assertKeys(t, "NotifyResult", jsonKeys(t, nr), []string{"channel", "ok", "attempts", "lastError", "sentAt"})

	ae := domain.AlertEvent{Tenant: "t1", ServiceName: "s", Metric: "m", Level: "P1", Timeframe: domain.Timeframe{StartTime: timeNow(), EndTime: timeNow()}, Raw: map[string]any{"k": "v"}}
	assertKeys(t, "AlertEvent", jsonKeys(t, ae), []string{"tenant", "serviceName", "metric", "level", "timeframe", "raw"})
}

func TestJSONTagsLLMTypes(t *testing.T) {
	lm := domain.LLMMessage{Role: "user", Text: "hi"}
	assertKeys(t, "LLMMessage", jsonKeys(t, lm), []string{"role", "text"})

	lr := domain.LLMRequest{System: "s", Messages: []domain.LLMMessage{lm}, MaxTokens: 100}
	assertKeys(t, "LLMRequest", jsonKeys(t, lr), []string{"system", "messages", "maxTokens"})
}

func TestJSONTagsDegradeAndSettingsTypes(t *testing.T) {
	pq := domain.PresetQuery{ID: "p1", Label: "l", Params: domain.DiagnoseParams{ServiceName: "s"}}
	assertKeys(t, "PresetQuery", jsonKeys(t, pq), []string{"id", "label", "params"})

	ds := domain.Datasource{Name: "cdn", Cloud: "aliyun", OK: true, CheckedAt: timeNow()}
	assertKeys(t, "Datasource", jsonKeys(t, ds), []string{"name", "cloud", "ok", "checkedAt"})

	lp := domain.LLMProviderConfig{Name: "qwen", Default: true, Model: "m", KeyMasked: "sk-****abcd"}
	assertKeys(t, "LLMProviderConfig", jsonKeys(t, lp), []string{"name", "default", "model", "keyMasked"})

	nc := domain.NotifyChannel{Channel: "feishu", Enabled: true}
	assertKeys(t, "NotifyChannel", jsonKeys(t, nc), []string{"channel", "enabled"})

	rw := domain.RiskWhitelistEntry{Tool: "query_ecs", RiskLevel: domain.RiskLow}
	assertKeys(t, "RiskWhitelistEntry", jsonKeys(t, rw), []string{"tool", "riskLevel"})

	ss := domain.SessionSummary{ID: "s1", Type: "chat", Query: "q", ServiceName: "s", IntentType: "diagnose", Status: "done", DiagnosisID: "dx1", CreatedAt: timeNow()}
	assertKeys(t, "SessionSummary", jsonKeys(t, ss), []string{"id", "type", "query", "serviceName", "intentType", "status", "diagnosisId", "createdAt"})
}

func TestJSONTagsBaseClientRequestTypes(t *testing.T) {
	req := domain.LogQueryRequest{
		Tenant:    "t1",
		LogType:   "cdn",
		Cloud:     "aliyun",
		Resource:  "order-service",
		StartTime: timeNow(),
		EndTime:   timeNow(),
		Limit:     100,
	}
	assertKeys(t, "LogQueryRequest", jsonKeys(t, req), []string{"tenant", "logType", "cloud", "resource", "startTime", "endTime", "limit"})

	score := 0.5
	res := domain.LogQueryResult{Total: 10, Truncated: false, Entries: []domain.LogEntry{}, RiskScore: &score}
	assertKeys(t, "LogQueryResult", jsonKeys(t, res), []string{"total", "truncated", "entries", "riskScore"})
}

// ---- 枚举合法性校验 ----

func TestSeverityIsValid(t *testing.T) {
	valid := map[domain.Severity]bool{
		domain.SeverityP0: true,
		domain.SeverityP1: true,
		domain.SeverityP2: true,
		domain.SeverityP3: true,
		"":                false,
		"p0":              false,
		"P4":              false,
		"critical":        false,
	}
	for s, want := range valid {
		if got := s.IsValid(); got != want {
			t.Errorf("Severity(%q).IsValid() = %v, want %v", s, got, want)
		}
	}
}

func TestIntentTypeIsValid(t *testing.T) {
	valid := map[domain.IntentType]bool{
		domain.IntentDiagnose:   true,
		domain.IntentResource:   true,
		domain.IntentOutOfScope: true,
		"":                      false,
		"chat":                  false,
		"Diagnose":              false,
	}
	for it, want := range valid {
		if got := it.IsValid(); got != want {
			t.Errorf("IntentType(%q).IsValid() = %v, want %v", it, got, want)
		}
	}
}

func TestRiskLevelIsValid(t *testing.T) {
	valid := map[domain.RiskLevel]bool{
		domain.RiskRead: true,
		domain.RiskLow:  true,
		domain.RiskHigh: true,
		"":              false,
		"medium":        false,
		"READ":          false,
	}
	for rl, want := range valid {
		if got := rl.IsValid(); got != want {
			t.Errorf("RiskLevel(%q).IsValid() = %v, want %v", rl, got, want)
		}
	}
}

func TestSessionStatusIsValid(t *testing.T) {
	valid := map[domain.SessionStatus]bool{
		domain.SessionStatusRunning: true,
		domain.SessionStatusDone:    true,
		domain.SessionStatusFailed:  true,
		"":                          false,
		"pending":                   false,
	}
	for s, want := range valid {
		if got := s.IsValid(); got != want {
			t.Errorf("SessionStatus(%q).IsValid() = %v, want %v", s, got, want)
		}
	}
}

func TestRiskEntryStatusIsValid(t *testing.T) {
	valid := map[domain.RiskEntryStatus]bool{
		domain.RiskStatusPendingView: true,
		domain.RiskStatusViewed:      true,
		domain.RiskStatusDone:        true,
		"":                           false,
		"pending":                    false,
	}
	for s, want := range valid {
		if got := s.IsValid(); got != want {
			t.Errorf("RiskEntryStatus(%q).IsValid() = %v, want %v", s, got, want)
		}
	}
}

// ---- Session 字段覆盖 er-diagram sessions 集合 ----

func TestSessionCoversERSessionsFields(t *testing.T) {
	s := domain.Session{
		ID:                 "s1",
		TenantID:           "t1",
		Type:               domain.SessionTypeChat,
		UserID:             "u1",
		Query:              "order-service 最近一小时错误率上升",
		ServiceName:        "order-service",
		IntentType:         domain.IntentDiagnose,
		IntentConfidence:   0.95,
		CorrectionLog:      []domain.CorrectionRecord{{From: domain.IntentResource, To: domain.IntentDiagnose, By: "u1", At: timeNow(), Message: "不对，我要排障"}},
		OrchestrationTrace: []domain.TraceStep{{Step: 1, Agent: "coordinator", Action: "intent", DurationMs: 5, Summary: "ok"}},
		Status:             domain.SessionStatusRunning,
		CreatedAt:          timeNow(),
		UpdatedAt:          timeNow(),
	}
	assertKeys(t, "Session", jsonKeys(t, s), []string{
		"id", "tenantId", "type", "userId", "query", "serviceName",
		"intentType", "intentConfidence", "correctionLog", "orchestrationTrace",
		"status", "createdAt", "updatedAt",
	})
	assertKeys(t, "Session.CorrectionRecord", jsonKeys(t, s.CorrectionLog[0]), []string{"from", "to", "by", "at", "message"})

	if domain.SessionTypeChat != "chat" || domain.SessionTypeAlert != "alert" {
		t.Errorf("session type constants drift: %q %q", domain.SessionTypeChat, domain.SessionTypeAlert)
	}
}
