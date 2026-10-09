---
status: "completed"
started: "2026-09-19 11:41"
completed: "2026-09-19 11:51"
time_spent: "~10m"
---

# Task Record: 1 cdnMetricDAO 新增 BulkUpsertMetrics 批量写

## Summary
cdnMetricDAO 新增 BulkUpsertMetrics 批量写:mongo BulkWrite + UpdateOne upsert,filter 复用 UpsertMetric 唯一键 (account_id, domain, date)(抽共享 helper cdnMetricFilter/cdnMetricUpsertUpdate,单条与批量共用同一键/更新构造,防键漂移);unordered 批(沿仓内 topology edge.go UpsertMany 先例,部分失败不阻塞无关行);空批早退 nil 不触碰数据库;BulkWrite 错误原样回传。CDNMetricDAO 接口同步新增方法(供 Task 2 执行器攒批调用),executor mock 内嵌接口故编译安全。既有 UpsertMetric 保留,语义不变。单测:空批(nil+空切片)/错误注入(超时 ctx 经驱动回传)/活体(MONGO_DSN 门控):单批多条落库、同键二次批量写入幂等不重复且值覆盖、跨账号同域名行互不覆盖。

## Changes

### Files Created
- e-cam-service/internal/cam/repository/dao/cdn_metric_bulk_test.go

### Files Modified
- e-cam-service/internal/cam/repository/dao/cdn_metric.go

### Key Decisions
- filter/update 抽为 cdnMetricFilter/cdnMetricUpsertUpdate 共享 helper,UpsertMetric 与 BulkUpsertMetrics 共用,唯一键 (account_id, domain, date) 单点维护
- BulkWrite 采用 SetOrdered(false),沿仓内 topology/repository/dao/edge.go UpsertMany 先例;unordered 下部分失败不阻塞批内无关行,错误仍整批回传
- 活体测试沿包内 MONGO_DSN 门控约定(test-bulk- 前缀域名 + 结束清理);本机 Mongo(含 test env 118.145.73.93)不可达,活体用例在此环境 SKIP,多批/幂等断言在有 DSN 环境执行
- 错误注入用超时 context + 不可达 DSN 直构 cdnMetricDAO(绕开 NewCDNMetricDAO 建索引的 server selection 阻塞),0.1s 内确定性回传错误

## Test Results
- **Tests Executed**: Yes
- **Passed**: 4
- **Failed**: 0
- **Coverage**: 5.1%

## Acceptance Criteria
- [x] cdnMetricDAO.BulkUpsertMetrics(ctx, []types.CDNMetric) error 存在:BulkWrite UpdateOne+upsert,filter 用唯一键 (account_id+domain+date)
- [x] 空 metrics 直接返回 nil;BulkWrite 错误回传(含部分失败语义符合 mongo BulkWrite)
- [x] 既有 UpsertMetric 单条方法保留(兼容)
- [x] 单测覆盖:空批、单批多条、幂等(同键两次写入不重复)、错误注入回传
- [x] go test ./internal/cam/repository/dao/ 全绿;go build ./internal/cam/repository/dao/ 通过

## Notes
coverage 5.1% 为 go test -cover 报的 dao 包整包值——该包测试全为 MONGO_DSN 门控活体,本机无 Mongo 全部 SKIP;新代码 BulkUpsertMetrics/cdnMetricFilter/cdnMetricUpsertUpdate 均达 100%(go tool cover -func),UpsertMetric 等既有函数 0% 系活体 SKIP 所致非回归。go build ./... 全仓通过;go vet 干净;gofmt -l 提示的 10 个文件均为未触碰的存量文件(account.go/asset.go 等),本次两个文件格式干净。-race 因本机 CGO_ENABLED=0 不可用(单 goroutine 用例不受影响;提案 -race 验证落在 Task 2 执行器并发范围)
