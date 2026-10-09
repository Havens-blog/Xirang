// Package mongox 通用 MongoDB 数据库连接封装。
//
// 与 e-cam-service 的 pkg/mongox 同语义（连接封装不承载业务谓词）；
// opsagent 为独立 module，无法跨 module import，故在本仓库内提供同形实现。
// DAO（internal/repository/dao）一律经本包获取 *mongo.Collection，
// 租户谓词等业务过滤由 DAO 层负责。
package mongox

import "go.mongodb.org/mongo-driver/mongo"

// Mongo MongoDB 数据库连接封装（持连接与库名）。
type Mongo struct {
	DBClient *mongo.Client
	dbName   string
}

// NewMongo 基于已建立的 client 与库名构造封装。
func NewMongo(client *mongo.Client, dbName string) *Mongo {
	return &Mongo{DBClient: client, dbName: dbName}
}

// Database 返回当前库。
func (m *Mongo) Database() *mongo.Database {
	return m.DBClient.Database(m.dbName)
}

// Collection 返回当前库下的集合。
func (m *Mongo) Collection(collName string) *mongo.Collection {
	return m.Database().Collection(collName)
}

// DBName 返回当前数据库名称。
func (m *Mongo) DBName() string {
	return m.dbName
}
