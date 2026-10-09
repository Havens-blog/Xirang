package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Havens-blog/e-cam-service/internal/audit/domain"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitForAuditLogs 轮询等待异步审计写入落袋（中间件经 goroutine 写入）。
// 通过锁保护的 count() 观测，创建 happens-before 关系；
// 调用方随后用 snapshot() 取锁下副本再断言。
func waitForAuditLogs(t *testing.T, dao *captureAuditDAO, n int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if dao.count() >= n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("audit logs not written within deadline: got %d want %d", dao.count(), n)
}

// captureAuditDAO 捕获中间件写入的审计日志（Create 落内存）。
// 中间件是 fire-and-forget：Create 在请求 goroutine 之外的 goroutine 里被调用，
// 必须加锁（无锁 append + 测试读切片头会被竞态检测器标记）。
type captureAuditDAO struct {
	mu   sync.Mutex
	logs []domain.AuditLog
}

func (d *captureAuditDAO) Create(_ context.Context, l domain.AuditLog) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logs = append(d.logs, l)
	return int64(len(d.logs)), nil
}

// count 锁保护的长度观测（轮询等待用）。
func (d *captureAuditDAO) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.logs)
}

// snapshot 锁下的副本（断言用）。
func (d *captureAuditDAO) snapshot() []domain.AuditLog {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]domain.AuditLog(nil), d.logs...)
}

func (d *captureAuditDAO) List(context.Context, domain.AuditLogFilter) ([]domain.AuditLog, error) {
	return nil, nil
}

// 其余查询接口与断言无关（中间件仅消费 Create）。

func (d *captureAuditDAO) Count(context.Context, domain.AuditLogFilter) (int64, error) {
	return 0, nil
}
func (d *captureAuditDAO) CountByResult(context.Context, domain.AuditLogFilter) (map[string]int64, error) {
	return nil, nil
}
func (d *captureAuditDAO) CountByOperationType(context.Context, domain.AuditLogFilter) (map[string]int64, error) {
	return nil, nil
}
func (d *captureAuditDAO) CountByHTTPMethod(context.Context, domain.AuditLogFilter) (map[string]int64, error) {
	return nil, nil
}
func (d *captureAuditDAO) ListTopEndpoints(context.Context, domain.AuditLogFilter, int) ([]domain.EndpointStats, error) {
	return nil, nil
}
func (d *captureAuditDAO) ListTopOperators(context.Context, domain.AuditLogFilter, int) ([]domain.OperatorStats, error) {
	return nil, nil
}
func (d *captureAuditDAO) InitIndexes(context.Context) error { return nil }

// TestInferOperationType_CertPaths cert 域操作类型标签（7.2：导入/删除/扫描/
// 配置面写入 ecam_audit_log 的 operation_type 可辨识）。
func TestInferOperationType_CertPaths(t *testing.T) {
	cases := []struct {
		path, method string
		want         domain.AuditOperationType
	}{
		// cam 域既有口径不受影响
		{"/api/v1/cam/assets", http.MethodPost, "api_asset_create"},
		{"/api/v1/cam/assets/sync", http.MethodPost, "api_asset_sync"},
		{"/api/v1/cam/tasks", http.MethodDelete, "api_task_delete"},
		// cert 域：根路径导入/参数段归 cert，静态子资源段取段名
		{"/api/v1/certs", http.MethodPost, "api_cert_create"},
		{"/api/v1/certs/6590aabbccdd000000000001", http.MethodDelete, "api_cert_delete"},
		{"/api/v1/certs/6590aabbccdd000000000001/key", http.MethodPost, "api_cert_create"},
		{"/api/v1/certs/6590aabbccdd000000000001/scan", http.MethodPost, "api_cert_create"},
		{"/api/v1/certs/settings/exemptions", http.MethodPost, "api_settings_create"},
		{"/api/v1/certs/settings/exemptions/a.example.com", http.MethodDelete, "api_settings_delete"},
		{"/api/v1/certs/settings/test", http.MethodPost, "api_settings_create"},
		{"/api/v1/certs/settings/crds/6590aabbccdd000000000001", http.MethodDelete, "api_settings_delete"},
		{"/api/v1/certs/changes", http.MethodPost, "api_changes_create"},
		{"/api/v1/certs/changes/6590aabbccdd000000000001/cancel", http.MethodPost, "api_changes_create"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, inferOperationType(tc.path, tc.method), "%s %s", tc.method, tc.path)
	}
}

// TestAuditMiddleware_MultipartBodyOmitted multipart 请求体（证书/私钥上传）
// 不落审计日志（渗透式自查口径：日志无明文私钥），仅记录占位符。
func TestAuditMiddleware_MultipartBodyOmitted(t *testing.T) {
	dao := &captureAuditDAO{}
	mdl := NewAuditMiddleware(dao, elog.DefaultLogger)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(mdl.Build())
	engine.POST("/api/v1/certs", func(c *gin.Context) {
		// handler 侧仍可读取 body（multipart 边界在，文件可解析）
		_, err := io.Copy(io.Discard, c.Request.Body)
		assert.NoError(t, err)
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	body := "------boundary\r\nContent-Disposition: form-data; name=\"keyFile\"; filename=\"a.key\"\r\n\r\n-----BEGIN PRIVATE KEY-----\r\nsecretmaterial\r\n-----END PRIVATE KEY-----\r\n------boundary--\r\n"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/certs", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----boundary")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	waitForAuditLogs(t, dao, 1)
	logs := dao.snapshot()
	log := logs[0]
	assert.Equal(t, "api_cert_create", string(log.OperationType))
	assert.Equal(t, "[multipart/form-data body omitted]", log.RequestBody)
	assert.NotContains(t, log.RequestBody, "PRIVATE KEY")
	assert.Equal(t, "success", string(log.Result))
}

// TestAuditMiddleware_JSONBodySanitized JSON body 脱敏（既有行为回归：
// password/secret 类字段掩码后入审计）。
func TestAuditMiddleware_JSONBodySanitized(t *testing.T) {
	dao := &captureAuditDAO{}
	mdl := NewAuditMiddleware(dao, elog.DefaultLogger)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(mdl.Build())
	engine.PUT("/api/v1/certs/settings", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	payload, _ := json.Marshal(map[string]string{"webhookUrls": "https://hook.example.com/x", "password": "p@ss"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/certs/settings", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	waitForAuditLogs(t, dao, 1)
	logs := dao.snapshot()
	log := logs[0]
	assert.Contains(t, log.RequestBody, "***")
	assert.NotContains(t, log.RequestBody, "p@ss")
	assert.Equal(t, "api_settings_update", string(log.OperationType))
}
