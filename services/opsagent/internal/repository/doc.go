// Package repository Mongo DAO（DB: opsagent）。
//
// 集合：sessions / diagnoses / citations / risk_entries / alerts / llm_usage /
// settings / compensations。沿用 e-cam-service 的 pkg/mongox 封装模式
// （internal/<mod>/repository/dao）；全部查询必须携带租户谓词。索引初始化与
// 各 DAO 实现在后续任务交付。
package repository
