/**
 * 运维 Agent 跨页共享展示工具（DRY：RFC3339 时间格式化）。
 *
 * 风险中心 / 历史回溯 / 系统配置三页此前各自重复 3 行 formatTime，
 * 收敛至此统一实现，各 logic.ts 重导出保持既有 import 不变。
 * formatClock（HH:mm）为原型列表/标题行 mono 时间口径（history/risk 共用）。
 */

/** RFC3339 → 「YYYY-MM-DD HH:mm」；空值回退占位 */
export function formatTime(rfc3339: string): string {
    if (!rfc3339) return '—'
    return rfc3339.slice(0, 16).replace('T', ' ')
}

/**
 * RFC3339 → 本地时区 HH:mm（原型 history.html / risk-center.html 的 mono 时间口径）；
 * 空值/不可解析回退占位 —。
 */
export function formatClock(rfc3339: string): string {
    if (!rfc3339) return '—'
    const at = new Date(rfc3339)
    if (Number.isNaN(at.getTime())) return '—'
    return `${String(at.getHours()).padStart(2, '0')}:${String(at.getMinutes()).padStart(2, '0')}`
}