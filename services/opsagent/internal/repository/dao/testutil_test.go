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

// connectTimeout / dropTimeout CI 服务容器连接与清理超时。
const (
	connectTimeout = 10 * time.Second
	dropTimeout    = 5 * time.Second
)

// newTestMongo 连接 CI 注入的 Mongo 服务容器（mongo:7，MONGO_URI 环境变量注入），
// 复用 pkg/mongox 连接封装。本地未配置 MONGO_URI 时跳过
// （不引入 testcontainers —— 任务 Hard Rule；与 tech-design › Testing Strategy 一致）。
func newTestMongo(t *testing.T) *mongox.Mongo {
	t.Helper()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		t.Skip("MONGO_URI not set: skipping Mongo integration test (CI injects mongo:7 service container)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping mongo: %v", err)
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
