// Package repository Mongo DAO（DB: opsagent）。
//
// 集合：sessions / diagnoses / citations / risk_entries / alerts / llm_usage /
// settings / compensations。沿用 e-cam-service 的 pkg/mongox 封装模式
// （pkg/mongox + internal/repository/dao）；全部查询必须携带租户谓词。
// sessions / diagnoses / citations DAO 见 repository/dao 子包（任务 2.1）；
// risk_entries / alerts / settings / compensations DAO 由后续任务交付。
package repository
