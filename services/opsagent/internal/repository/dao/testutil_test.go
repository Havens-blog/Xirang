package dao

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Havens-blog/opsagent/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testDBName 测试专用库（与业务库 opsagent 隔离，测试结束整体 Drop）。
const testDBName = "opsagent_dao_test"

// connectTimeout / dropTimeout 连接与清理超时。
const (
	connectTimeout = 10 * time.Second
	dropTimeout    = 5 * time.Second
)

// defaultMongoURI 未显式指定 MONGO_URI 时的默认地址（与 e-cam-service 脚本约定一致）。
const defaultMongoURI = "mongodb://localhost:27017"

// mongoSkipRequested 是否显式要求跳过 Mongo 集成测试。
// 仅当 MONGO_SKIP_INTEGRATION=1/true/yes 时跳过；未设置默认不跳过——
// 否则「缺 Mongo 静默 skip」会掩盖 DAO 覆盖率空洞，虚报全绿。
func mongoSkipRequested() bool {
	switch os.Getenv("MONGO_SKIP_INTEGRATION") {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	default:
		return false
	}
}

// resolveMongoURI 返回集成测试目标地址：优先 MONGO_URI，缺省回退本机 27017。
func resolveMongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	return defaultMongoURI
}

// newTestMongo 连接 Mongo（默认本机 :27017，或 MONGO_URI 指定地址），复用 pkg/mongox 封装。
// 连接失败即 t.Fatal（给出恢复指引），而非静默 Skip——覆盖率门禁依赖 DAO 集成测试真实执行。
// 仅当显式 MONGO_SKIP_INTEGRATION=1 时才跳过（不引入 testcontainers——任务 Hard Rule）。
func newTestMongo(t *testing.T) *mongox.Mongo {
	t.Helper()
	if mongoSkipRequested() {
		t.Skip("MONGO_SKIP_INTEGRATION set: skipping Mongo integration tests")
	}
	uri := resolveMongoURI()

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("连接 Mongo(%s) 失败: %v —— 请先启动本机 mongod(:27017) 或用 MONGO_URI 指向可用实例；确需跳过请设 MONGO_SKIP_INTEGRATION=1", uri, err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		t.Fatalf("Ping Mongo(%s) 失败: %v —— 请先启动本机 mongod(:27017)；确需跳过请设 MONGO_SKIP_INTEGRATION=1", uri, err)
	}

	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), dropTimeout)
		defer shutCancel()
		_ = client.Database(testDBName).Drop(shutCtx)
		_ = client.Disconnect(shutCtx)
	})
	return mongox.NewMongo(client, testDBName)
}

// mustUTCMillis 将时间截断到毫秒（Mongo 存储精度），避免往返比较抖动。
func mustUTCMillis(t time.Time) time.Time {
	return t.Truncate(time.Millisecond).UTC()
}

// TestMongoSkipRequested 覆盖门控判定：仅显式真值时跳过。
func TestMongoSkipRequested(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"", false},
		{"0", false},
		{"false", false},
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"YES", true},
	}
	for _, tc := range cases {
		t.Setenv("MONGO_SKIP_INTEGRATION", tc.val)
		if got := mongoSkipRequested(); got != tc.want {
			t.Errorf("MONGO_SKIP_INTEGRATION=%q → skip=%v, want %v", tc.val, got, tc.want)
		}
	}
}

// TestResolveMongoURI 覆盖地址解析：环境变量优先，缺省回退本机。
func TestResolveMongoURI(t *testing.T) {
	t.Setenv("MONGO_URI", "")
	if got := resolveMongoURI(); got != defaultMongoURI {
		t.Fatalf("默认 URI = %q, want %q", got, defaultMongoURI)
	}
	t.Setenv("MONGO_URI", "mongodb://example.test:27018")
	if got := resolveMongoURI(); got != "mongodb://example.test:27018" {
		t.Fatalf("环境变量 URI = %q, want mongodb://example.test:27018", got)
	}
}
