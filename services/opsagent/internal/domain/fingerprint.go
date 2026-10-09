// Package domain 跨模块共享的领域类型（Ground Truth）。
//
// 指纹函数供 worker 告警去重与告警记账（alerts.fingerprint）复用：
// Fingerprint = sha256(tenant + serviceName + metric + level) 取前 32 位十六进制
// （tech-design › Shared Types › AlertEvent）。
package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// fingerprintHexLen 指纹十六进制长度（sha256 前 128 bit）。
const fingerprintHexLen = 32

// Fingerprint 计算告警指纹：sha256(tenant+serviceName+metric+level) 前 32 位十六进制。
// 同指纹同租户的告警在去重窗口内归并为一次排查。
func Fingerprint(tenant, serviceName, metric, level string) string {
	sum := sha256.Sum256([]byte(tenant + serviceName + metric + level))
	return hex.EncodeToString(sum[:])[:fingerprintHexLen]
}
