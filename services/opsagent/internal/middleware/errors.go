package middleware

import (
	"errors"
	"net/http"

	"github.com/Havens-blog/opsagent/internal/repository/dao"
	"go.mongodb.org/mongo-driver/mongo"
)

// 内部错误码 → HTTP 状态映射（tech-design › Error Handling 错误表为唯一权威）。
//
// 跨租户/不存在一律 ERR_NOT_FOUND 404（不泄露资源存在性）；ERR_TENANT_DENIED 仅用于
// 401 鉴权失败；ERR_PERSIST_FAILED 为「HTTP 200 + code≠0」信封语义（提示结果暂未能保存 + 补偿）。

// 错误码常量（api-handbook 错误表）。
const (
	CodeParamInvalid  = "ERR_PARAM_INVALID"
	CodeIntentUnknown = "ERR_INTENT_UNKNOWN"
	CodeOutOfScope    = "ERR_OUT_OF_SCOPE"
	CodeLLMUnavail    = "ERR_LLM_UNAVAILABLE"
	CodeBaseTimeout   = "ERR_BASE_TIMEOUT"
	CodeContractDrift = "ERR_CONTRACT_DRIFT"
	// CodeTenantDenied 定义于 auth.go（ERR_TENANT_DENIED，401）。
	CodeNotFound      = "ERR_NOT_FOUND"
	CodeConflict      = "ERR_CONFLICT"
	CodeQueueFull     = "ERR_QUEUE_FULL"
	CodePersistFailed = "ERR_PERSIST_FAILED"
	// CodeInternal 错误表外兜底内部错误（500）。
	CodeInternal = "ERR_INTERNAL"
)

// ParamInvalidError 请求体非法（必填缺失/超长/互斥字段同现）→ 400。
type ParamInvalidError struct{ Msg string }

func (e *ParamInvalidError) Error() string { return "param invalid: " + e.Msg }

// PersistFailedError 落库失败（重试 2 次仍失败）→ HTTP 200 + code≠0 信封语义。
type PersistFailedError struct{ Err error }

func (e *PersistFailedError) Error() string { return "persist failed: " + e.Err.Error() }
func (e *PersistFailedError) Unwrap() error { return e.Err }

// QueueFullError 排查队列溢出（低级别告警）→ 202 半成功语义（data.dropped=true）。
type QueueFullError struct{}

func (e *QueueFullError) Error() string { return "queue full" }

// MapError 内部错误 → (HTTP 状态, 业务错误码, 用户文案)。
//
// 与 e-cam-service 装配约定一致：跨租户/不存在统一 mongo.ErrNoDocuments → 404，
// 不在错误文案中泄露资源是否真实存在。
func MapError(err error) (status int, code string, message string) {
	switch {
	case err == nil:
		return http.StatusOK, "0", "ok"
	case errors.Is(err, mongo.ErrNoDocuments):
		return http.StatusNotFound, CodeNotFound, "资源不存在"
	case isConflict(err):
		return http.StatusConflict, CodeConflict, "已被其他值班更新，请刷新后重试"
	case isParamInvalid(err):
		return http.StatusBadRequest, CodeParamInvalid, "请求参数非法"
	case isPersistFailed(err):
		return http.StatusOK, CodePersistFailed, "结果暂未能保存，请稍后重试"
	case isQueueFull(err):
		return http.StatusAccepted, CodeQueueFull, "告警已受理，队列繁忙部分低级别告警已丢弃"
	default:
		return http.StatusInternalServerError, CodeInternal, "服务内部错误"
	}
}

func isConflict(err error) bool {
	var ce *dao.ConflictError
	return errors.As(err, &ce)
}

func isParamInvalid(err error) bool {
	var pe *ParamInvalidError
	return errors.As(err, &pe)
}

func isPersistFailed(err error) bool {
	var pe *PersistFailedError
	return errors.As(err, &pe)
}

func isQueueFull(err error) bool {
	var qe *QueueFullError
	return errors.As(err, &qe)
}
