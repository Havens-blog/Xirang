---
tags: [go, mongodb, bson, serialization-contract]
date: 2026-09-24
---

# Gotcha: Go 结构体只有 `json:` tag 没有 `bson:` tag，经 `map[string]interface{}` 落 Mongo 后字段名被压成小写

## Problem

运营侧「采集异常」warn 态永远不亮：任务 `Result.failures` 在数据库里出现了，但前端按 `account_id` / `error_count` / `last_error` 读不到值（`undefined`），过滤条件 `typeof error_count === 'number'` 恒 false。

## Root Cause

`nasProviderFailure` 结构体只写了 `json:` tag（`account_id` / `error_count` / `last_error`），没有 `bson:` tag。该结构体被塞进 `t.Result = map[string]interface{}{"failures": failures, ...}`，再由 `MongoRepository.Update` / `InsertOne` 经 mongo-driver 编码落库。mongo-driver 的 BSON 编码器**只认 `bson:` tag**，对无 tag 的字段按「Go 字段名小写」生成键：`AccountID` → `accountid`、`ErrorCount` → `errorcount`、`LastError` → `lasterror`。`json:` tag 在 BSON 编码路径完全不起作用。

链路（全在 e-cam-service 仓库内）：

```
sync_nas_metrics.go: t.Result = map[string]any{... "failures": []nasProviderFailure ...}
pkg/taskx/task.go:28:   Result map[string]interface{} `json:"result,omitempty" bson:"result"`
pkg/taskx/repository_mongo.go: Update → collection.UpdateOne(ctx, bson.M{"$set": task})
   → BSON 编码 Result 内部的 []nasProviderFailure → key 变成 accountid/errorcount/lasterror
```

## Solution

给进 `map[string]interface{}` 落库的结构体同时加 `bson:` tag，与 `json:` tag 保持一致：

```go
type nasProviderFailure struct {
	Provider   string `json:"provider" bson:"provider"`       // 云厂商
	AccountID  int64  `json:"account_id" bson:"account_id"`   // 云账号 ID
	ErrorCount int    `json:"error_count" bson:"error_count"` // 本次任务内该账号累计失败次数
	LastError  string `json:"last_error" bson:"last_error"`   // 末次错误信息
}
```

`diskProviderFailure` / `ossProviderFailure` / NAS 回填的 `failures` 均以 `type X = nasProviderFailure` 别名复用它，一处收口即全修。

## Reusable Pattern

- **凡结构体会经 `map[string]interface{}`（或任意 interface 字段）进入 Mongo/BSON 编码路径，`json:` tag 与 `bson:` tag 必须成对出现**。只有 `json:` 是不够的——BSON 编码器不读 json tag。
- 判定「是否会踩坑」的探针：该结构体的值最终是否落入 `*x.Repository` 的 `InsertOne/UpdateOne/ReplaceOne` 参数、或其上级结构体带 `bson:"result"` 之类嵌套。落在纯 JSON 路径（HTTP 响应、`json.Marshal`）则不踩。
- 校验手段：写一个 `bson.Marshal(map[string]any{"failures": []nasProviderFailure{...}})` → `bson.Unmarshal` 回 `bson.M` 的 round-trip 测试，断言蛇形键存在（见 `nas_provider_failure_bson_test.go`）。该测试对「还原 bug」会红：无 bson tag 时 key 变成 `accountid` 等。
- 排查「数据落库了但字段读不到」先查 BSON key 是否存在（`bson.D`/`bson.M` 读出真实键），再查前端契约。

## Related Files

- internal/cam/task/executor/sync_nas_metrics.go（`nasProviderFailure` 定义，已加 bson tag）
- internal/cam/task/executor/sync_disk_metrics.go / sync_oss_metrics.go / sync_nas_metrics_backfill.go（别名复用）
- pkg/taskx/task.go（`Result map[string]interface{} bson:"result"`）
- pkg/taskx/repository_mongo.go（Mongo 落地路径）
- internal/cam/task/executor/nas_provider_failure_bson_test.go（round-trip 回归测试）