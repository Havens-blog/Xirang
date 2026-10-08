package dao

import (
	"context"
	"time"

	"github.com/Havens-blog/opsagent/internal/domain"
	"github.com/Havens-blog/opsagent/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SettingsCollection settings 集合名（schema.mongo.js）。
const SettingsCollection = "settings"

// settingsDoc settings 集合存储文档。value 为 scope 对应配置内容（object）：不同
// scope 形状各异（datasource→[]Datasource、preset_queries→[]PresetQuery、
// guided_templates→map 等），DAO 层按 scope 做泛化读写，不强绑具体类型。
type settingsDoc struct {
	ID        string    `bson:"_id"`
	TenantID  string    `bson:"tenant_id"`
	Scope     string    `bson:"scope"`
	Value     any       `bson:"value"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// SettingsDAO 系统配置 DAO（settings 集合；按 (scope, tenant_id) 唯一读写）。
// tenant_id="" 为全局配置哨兵值（er-diagram settings 约定）。
type SettingsDAO interface {
	// Get 按 (tenant_id, scope) 读出配置内容到 out（指向目标类型的指针）；不存在 →
	// mongo.ErrNoDocuments。
	Get(ctx context.Context, tenantID string, scope domain.SettingScope, out any) error
	// Upsert 按 (tenant_id, scope) 创建或整体替换配置内容，刷新 updated_at。
	Upsert(ctx context.Context, tenantID string, scope domain.SettingScope, value any) error
	// InitIndexes 初始化 schema.mongo.js 定义的 settings 索引（幂等）。
	InitIndexes(ctx context.Context) error
}

type settingsDAO struct {
	db *mongox.Mongo
}

// NewSettingsDAO 构造 SettingsDAO。
func NewSettingsDAO(db *mongox.Mongo) SettingsDAO {
	return &settingsDAO{db: db}
}

// InitIndexes 创建 idx_settings_scope_tenant 唯一索引（scope + tenant_id）。
func (d *settingsDAO) InitIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "scope", Value: 1}, {Key: "tenant_id", Value: 1}},
			Options: options.Index().SetName("idx_settings_scope_tenant").SetUnique(true),
		},
	}
	_, err := d.db.Collection(SettingsCollection).Indexes().CreateMany(ctx, models)
	return err
}

// Get 读配置内容并解码到 out。
func (d *settingsDAO) Get(ctx context.Context, tenantID string, scope domain.SettingScope, out any) error {
	filter := bson.M{"tenant_id": tenantID, "scope": string(scope)}
	// value 以 bson.RawValue 捕获（可为 document 或 array），再反序列化到 out 的强类型。
	var doc struct {
		Value bson.RawValue `bson:"value"`
	}
	if err := d.db.Collection(SettingsCollection).FindOne(ctx, filter).Decode(&doc); err != nil {
		return err
	}
	return doc.Value.Unmarshal(out)
}

// Upsert 创建或整体替换 scope 配置。
func (d *settingsDAO) Upsert(ctx context.Context, tenantID string, scope domain.SettingScope, value any) error {
	filter := bson.M{"tenant_id": tenantID, "scope": string(scope)}
	update := bson.M{"$set": bson.M{
		"tenant_id":  tenantID,
		"scope":      string(scope),
		"value":      value,
		"updated_at": time.Now(),
	}}
	_, err := d.db.Collection(SettingsCollection).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}
