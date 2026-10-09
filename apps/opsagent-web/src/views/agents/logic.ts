/**
 * Agent 管理页纯展示逻辑（任务 5.6 / 原型 agent-management.html 口径）。
 *
 * 无 Vue / DOM 运行时依赖，可被 node 环境单测直接导入。依据：
 * docs/features/haven-opsagent/design/api-handbook.md §10 + 原型 agent-management.html。
 *
 * 数据诚实原则：AgentStat 仅 {name,status,tasksTotal,tasksFailed,avgDurationMs}，
 * 活跃任务/正常运行时长/版本无契约字段 → 渲染「—」；成功率由完成任务推导。
 */

// ==================== Agent 呈现元数据 ====================

export interface AgentPresentMeta {
    /** 卡片英文显示名（原型 h3） */
    display: string
    /** 卡片副角色行（原型 h3 下小字） */
    roleLine: string
    /** 44px 图标 emoji */
    emoji: string
    /** 图标底色调 token */
    gradToken: string
    /** 角色中文短名（trace/证据等跨页复用口径：编排/日志查询/告警/诊断） */
    label: string
}

const AGENT_PRESENT: Record<string, AgentPresentMeta> = {
    coordinator: {
        display: 'Coordinator',
        roleLine: '编排 · 任务协调',
        emoji: '🎯',
        gradToken: '--agent-coord',
        label: '编排',
    },
    log_analyst: {
        display: 'LogAnalyst',
        roleLine: '日志查询 · 日志分析',
        emoji: '📋',
        gradToken: '--agent-log',
        label: '日志查询',
    },
    monitor: {
        display: 'Monitor',
        roleLine: '告警监测 · 指标',
        emoji: '📊',
        gradToken: '--agent-monitor',
        label: '告警',
    },
    inspector: {
        display: 'Inspector',
        roleLine: '诊断 · 巡检',
        emoji: '🔍',
        gradToken: '--agent-inspector',
        label: '诊断',
    },
}

/** 未知名兜底（原样 + 灰底） */
const FALLBACK_PRESENT: AgentPresentMeta = {
    display: 'Agent',
    roleLine: '—',
    emoji: '🤖',
    gradToken: '--muted',
    label: 'Agent',
}

export function agentPresent(name: string): AgentPresentMeta {
    return AGENT_PRESENT[name] ?? { ...FALLBACK_PRESENT, display: name }
}

// ==================== Agent 角色标签 ====================

/** 4 逻辑 Agent 角色中文名（coordinator/log_analyst/monitor/inspector，四角色口径） */
export const AGENT_ROLE_LABEL: Record<string, string> = {
    coordinator: '编排',
    log_analyst: '日志查询',
    monitor: '告警',
    inspector: '诊断',
}

export function agentRoleLabel(name: string): string {
    return AGENT_ROLE_LABEL[name] ?? name
}

// ==================== Agent 状态 ====================

export type AgentTone = 'success' | 'warning' | 'danger' | 'info'

export interface AgentStatusMeta {
    label: string
    tone: AgentTone
    /** 是否强调「告警中」徽标（原型 badge-low ⚠️） */
    alarm: boolean
}

/** agent.status → 中文 + 色阶（原型：运行中 pulse / ⚠️ 告警中） */
export function agentStatusMeta(status: string): AgentStatusMeta {
    if (status === 'ok' || status === 'running') {
        return { label: '运行中', tone: 'success', alarm: false }
    }
    // busy/error/failed/其余异常态 → 原型「⚠️ 告警中」
    return { label: '告警中', tone: 'warning', alarm: true }
}

// ==================== 指标推导（数据诚实，不虚构） ====================

/** 成功率（1 - failed/total），total=0 兜底 — */
export function agentSuccessRate(total: number, failed: number): string {
    if (total <= 0) return '—'
    const rate = ((total - failed) / total) * 100
    return `${rate.toFixed(1)}%`
}

/** 全体重叠合计的完成数（Σ tasksTotal） */
export function sumTasksTotal(agents: { tasksTotal: number }[]): number {
    return agents.reduce((acc, a) => acc + a.tasksTotal, 0)
}

/** 全体重叠合计的失败数（Σ tasksFailed） */
export function sumTasksFailed(agents: { tasksFailed: number }[]): number {
    return agents.reduce((acc, a) => acc + a.tasksFailed, 0)
}

/** 任务队列「已完成/失败」跨 Agent 合计（原型四格后两格） */
export interface AgentTotals {
    done: number
    failed: number
}

export function agentTotals(agents: { tasksTotal: number; tasksFailed: number }[]): AgentTotals {
    return { done: sumTasksTotal(agents), failed: sumTasksFailed(agents) }
}

/** 加权平均响应时间（Σ durationMs / Σ total；无任务兜底 —） */
export function weightedAvgDuration(agents: { tasksTotal: number; avgDurationMs: number }[]): string {
    const totalTasks = sumTasksTotal(agents)
    if (totalTasks <= 0) return '—'
    const totalMs = agents.reduce((acc, a) => acc + a.tasksTotal * a.avgDurationMs, 0)
    return `${Math.round(totalMs / totalTasks)}ms`
}

/** 整体错误率（Σ failed / Σ total）；无任务兜底 — */
export function overallErrorRate(agents: { tasksTotal: number; tasksFailed: number }[]): string {
    const total = sumTasksTotal(agents)
    if (total <= 0) return '—'
    return `${((sumTasksFailed(agents) / total) * 100).toFixed(1)}%`
}

// ==================== LLM 预算用量 ====================

/** 预算使用率 0~1（budgetLimit=0 时兜底 0，避免除零） */
export function budgetUsageRatio(calls: number, budgetLimit: number): number {
    if (budgetLimit <= 0) return 0
    const r = calls / budgetLimit
    return r < 0 ? 0 : r > 1 ? 1 : r
}

/** 预算使用率 → 百分比文案 */
export function budgetUsagePercent(calls: number, budgetLimit: number): string {
    return `${Math.round(budgetUsageRatio(calls, budgetLimit) * 100)}%`
}