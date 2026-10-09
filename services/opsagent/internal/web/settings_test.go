package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/internal/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

// ---- fakes ----

type fakeSettingsStore struct {
	data      map[domain.SettingScope]any
	upsertErr error
}

func newFakeSettingsStore() *fakeSettingsStore {
	return &fakeSettingsStore{data: map[domain.SettingScope]any{}}
}

func (f *fakeSettingsStore) Get(_ context.Context, _ string, scope domain.SettingScope, out any) error {
	if v, ok := f.data[scope]; ok {
		b, _ := json.Marshal(v)
		return json.Unmarshal(b, out)
	}
	return mongo.ErrNoDocuments
}

func (f *fakeSettingsStore) Upsert(_ context.Context, _ string, scope domain.SettingScope, value any) error {
	f.data[scope] = value
	return f.upsertErr
}

type fakeRegistry struct {
	meta domain.ToolMeta
	err  error
}

func (f *fakeRegistry) Classify(string) (domain.ToolMeta, error) { return f.meta, f.err }

func newSettingsRouter(h *SettingsHandler, admin bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.TenantIDKey, int64(1))
		c.Set(middleware.CtxUidKey, int64(99))
		c.Next()
	})
	h.isAdmin = func(*gin.Context) bool { return admin }
	h.RegisterRoutes(r.Group("/api/v1/opsagent"))
	return r
}

// ---- tests ----

func TestSettings_Get(t *testing.T) {
	store := newFakeSettingsStore()
	store.data[domain.ScopePresetQueries] = []domain.PresetQuery{{ID: "p1", Label: "order-service"}}
	h := NewSettingsHandler(store, &fakeRegistry{}, nil)
	r := newSettingsRouter(h, false)

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/settings", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var d settingsData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(d.PresetQueries) != 1 || d.PresetQueries[0].ID != "p1" {
		t.Fatalf("presets = %+v", d.PresetQueries)
	}
}

func TestSettings_Put_NotAdmin(t *testing.T) {
	h := NewSettingsHandler(newFakeSettingsStore(), &fakeRegistry{}, nil)
	r := newSettingsRouter(h, false)

	rec, _ := doRequest(t, r, http.MethodPut, "/api/v1/opsagent/settings", `{"presetQueries":[]}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestSettings_Put_WhitelistUnregistered(t *testing.T) {
	reg := &fakeRegistry{err: errors.New("unregistered")}
	h := NewSettingsHandler(newFakeSettingsStore(), reg, nil)
	r := newSettingsRouter(h, true)

	rec, _ := doRequest(t, r, http.MethodPut, "/api/v1/opsagent/settings", `{"riskWhitelist":[{"tool":"hack","riskLevel":"low"}]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSettings_Put_WhitelistHighRisk(t *testing.T) {
	reg := &fakeRegistry{meta: domain.ToolMeta{RiskLevel: domain.RiskHigh}}
	h := NewSettingsHandler(newFakeSettingsStore(), reg, nil)
	r := newSettingsRouter(h, true)

	rec, _ := doRequest(t, r, http.MethodPut, "/api/v1/opsagent/settings", `{"riskWhitelist":[{"tool":"x","riskLevel":"low"}]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (high risk 不允许入白名单)", rec.Code)
	}
}

func TestSettings_Put_ApiKeyMasked(t *testing.T) {
	store := newFakeSettingsStore()
	h := NewSettingsHandler(store, &fakeRegistry{}, nil)
	r := newSettingsRouter(h, true)

	rec, _ := doRequest(t, r, http.MethodPut, "/api/v1/opsagent/settings", `{"llmProviders":[{"name":"deepseek","model":"v3","apiKey":"sk-abcdefgh12345678"}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	// 落库应只存掩码，非明文 apiKey。
	providers := store.data[domain.ScopeLLM].([]domain.LLMProviderConfig)
	if len(providers) != 1 || providers[0].KeyMasked != "sk-****5678" {
		t.Fatalf("providers = %+v, want masked key", providers)
	}
}

func TestSettings_Put_PresetInvalid(t *testing.T) {
	h := NewSettingsHandler(newFakeSettingsStore(), &fakeRegistry{}, nil)
	r := newSettingsRouter(h, true)

	rec, _ := doRequest(t, r, http.MethodPut, "/api/v1/opsagent/settings", `{"presetQueries":[{"label":"x","params":{}}]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (缺 serviceName)", rec.Code)
	}
}

func TestMaskAPIKey(t *testing.T) {
	if maskAPIKey("") != "" {
		t.Fatal("empty key should mask to empty")
	}
	if maskAPIKey("abc") != "sk-****" {
		t.Fatalf("short key = %q", maskAPIKey("abc"))
	}
	if maskAPIKey("sk-abcdefgh12345678") != "sk-****5678" {
		t.Fatalf("long key = %q", maskAPIKey("sk-abcdefgh12345678"))
	}
}

// ---- observability ----

type fakeObservability struct{}

func (f *fakeObservability) Agents() []agentStat {
	return []agentStat{{Name: "coordinator", Status: "ok", TasksTotal: 10}}
}
func (f *fakeObservability) Queue() queueStat {
	return queueStat{Depth: 3, Capacity: 100, Inflight: 2, Concurrency: 5, OverflowTotal: 7}
}
func (f *fakeObservability) LLMBudget() llmBudgetStat {
	return llmBudgetStat{Calls: 40, BudgetLimit: 100}
}
func (f *fakeObservability) LoadHistory() []loadPoint {
	return []loadPoint{{Ts: 1000, Depth: 3, Inflight: 2}}
}

func TestObservability(t *testing.T) {
	h := NewObservabilityHandler(&fakeObservability{})
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1/opsagent"))

	rec, env := doRequest(t, r, http.MethodGet, "/api/v1/opsagent/agents/observability", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var d observabilityData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(d.Agents) != 1 || d.Agents[0].Name != "coordinator" {
		t.Fatalf("agents = %+v", d.Agents)
	}
	if d.Queue.Depth != 3 || d.Queue.OverflowTotal != 7 {
		t.Fatalf("queue = %+v", d.Queue)
	}
	if d.LLMBudget.Calls != 40 || d.LLMBudget.BudgetLimit != 100 {
		t.Fatalf("llmBudget = %+v", d.LLMBudget)
	}
	if len(d.LoadHistory) != 1 {
		t.Fatalf("loadHistory = %+v", d.LoadHistory)
	}
}
