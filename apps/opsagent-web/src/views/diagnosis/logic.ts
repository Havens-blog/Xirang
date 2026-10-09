/**
 * 诊断详情页纯展示逻辑（任务 5.4 / 6.2 原型对位版）。
 *
 * 无 Vue / DOM 运行时依赖，可被 node 环境单测直接导入。依据：
 * docs/features/haven-opsagent/design/api-handbook.md §6 +
 * ui/prototype/diagnosis-detail.html（根因结论/处置建议/调用链）。
 *
 * 职责：处置建议风险档徽标、编排调用链 agent 中文四角色
 * （编排/日志查询/告警/诊断）、置信度格式化。
 */

import type { RiskLevel } from '@/api/opsagent'

// ==================== 风险档徽标（标题行/处置建议共用语义） ====================

export interface RiskLevelMeta {
    label: string
    tone: 'info' | 'warning' | 'danger'
}

/** 风险档徽标（read/low/high → 只读/低危/高危） */
export const RISK_LEVEL_META: Record<RiskLevel, RiskLevelMeta> = {
    read: { label: '只读', tone: 'info' },
    low: { label: '低危', tone: 'warning' },
    high: { label: '高危', tone: 'danger' },
}

// ==================== 处置建议徽标 ====================

export interface DispositionBadgeMeta {
    /** 徽标文案（原型 diagnosis-detail.html：自动执行 / 仅展示 · P3 可执行 / 人工确认） */
    text: string
    /** badge 样式档（badge-ok / badge-ghost / badge-high） */
    tone: 'ok' | 'ghost' | 'high'
}

/** 处置条目风险档 → 执行语义徽标（P1 只读：不渲染任何执行按钮） */
export function dispositionBadge(risk: string): DispositionBadgeMeta {
    switch (risk) {
        case 'read':
            return { text: '自动执行', tone: 'ok' }
        case 'low':
            return { text: '仅展示 · P3 可执行', tone: 'ghost' }
        case 'high':
            return { text: '人工确认', tone: 'high' }
        default:
            return { text: '仅展示', tone: 'ghost' }
    }
}

// ==================== 编排调用链标签 ====================

/** 调用链步骤 action → 中文 */
export const TRACE_ACTION_LABEL: Record<string, string> = {
    intent: '意图识别',
    query: '查询',
    diagnose: '诊断',
    report: '报告',
    notify: '通知',
}

/** 调用链 agent → 中文四角色（原型 agent-chip 体系：编排/日志查询/告警/诊断） */
export const AGENT_LABEL: Record<string, string> = {
    coordinator: '编排',
    log_analyst: '日志查询',
    monitor: '告警',
    inspector: '诊断',
}

/** 未知 action/agent 兜底：原样返回（容错，不吞数据） */
export function actionLabel(action: string): string {
    return TRACE_ACTION_LABEL[action] ?? action
}

export function agentLabel(agent: string): string {
    return AGENT_LABEL[agent] ?? agent
}

// ==================== 置信度 ====================

/** 置信度 0~1 → 百分比文案 */
export function formatConfidence(confidence: number): string {
    return `${Math.round(confidence * 100)}%`
}