import { describe, expect, it } from 'vitest'
import {
    agentPresent,
    agentRoleLabel,
    AGENT_ROLE_LABEL,
    agentStatusMeta,
    agentSuccessRate,
    agentTotals,
    budgetUsagePercent,
    budgetUsageRatio,
    overallErrorRate,
    sumTasksFailed,
    sumTasksTotal,
    weightedAvgDuration,
} from './logic'

describe('Agent 角色标签（四角色口径 编排/日志查询/告警/诊断）', () => {
    it('4 逻辑 Agent 中文名 + 未知兜底', () => {
        expect(AGENT_ROLE_LABEL).toEqual({
            coordinator: '编排',
            log_analyst: '日志查询',
            monitor: '告警',
            inspector: '诊断',
        })
        expect(agentRoleLabel('coordinator')).toBe('编排')
        expect(agentRoleLabel('unknown')).toBe('unknown')
    })
})

describe('Agent 卡片呈现元数据（原型 agent-management.html）', () => {
    it('四卡英文名 + 角色行 + emoji', () => {
        expect(agentPresent('coordinator')).toMatchObject({ display: 'Coordinator', roleLine: '编排 · 任务协调', emoji: '🎯' })
        expect(agentPresent('log_analyst').roleLine).toBe('日志查询 · 日志分析')
        expect(agentPresent('monitor').roleLine).toBe('告警监测 · 指标')
        expect(agentPresent('inspector').roleLine).toBe('诊断 · 巡检')
        expect(agentPresent('inspector').display).toBe('Inspector')
    })

    it('未知名兜底：display 原样', () => {
        expect(agentPresent('unknown').display).toBe('unknown')
    })
})

describe('Agent 状态（运行中 pulse / ⚠️ 告警中）', () => {
    it('ok/running → 运行中；busy/error/未知 → 告警中', () => {
        expect(agentStatusMeta('ok')).toEqual({ label: '运行中', tone: 'success', alarm: false })
        expect(agentStatusMeta('running')).toEqual({ label: '运行中', tone: 'success', alarm: false })
        expect(agentStatusMeta('busy').alarm).toBe(true)
        expect(agentStatusMeta('error').label).toBe('告警中')
        expect(agentStatusMeta('weird').label).toBe('告警中')
    })
})

describe('指标推导（数据诚实）', () => {
    const agents = [
        { name: 'a', status: 'ok', tasksTotal: 100, tasksFailed: 2, avgDurationMs: 100 },
        { name: 'b', status: 'ok', tasksTotal: 50, tasksFailed: 8, avgDurationMs: 300 },
    ]

    it('成功率 / 合计 / 加权平均 / 错误率', () => {
        expect(agentSuccessRate(100, 2)).toBe('98.0%')
        expect(agentSuccessRate(0, 0)).toBe('—')
        expect(sumTasksTotal(agents)).toBe(150)
        expect(sumTasksFailed(agents)).toBe(10)
        expect(agentTotals(agents)).toEqual({ done: 150, failed: 10 })
        expect(weightedAvgDuration(agents)).toBe('167ms')
        expect(overallErrorRate(agents)).toBe('6.7%')
    })

    it('无任务时加权平均 / 错误率兜底 —', () => {
        expect(weightedAvgDuration([])).toBe('—')
        expect(overallErrorRate([])).toBe('—')
    })
})

describe('LLM 预算用量', () => {
    it('使用率 0~1 夹取；budgetLimit≤0 兜底 0', () => {
        expect(budgetUsageRatio(40, 100)).toBe(0.4)
        expect(budgetUsageRatio(150, 100)).toBe(1)
        expect(budgetUsageRatio(-1, 100)).toBe(0)
        expect(budgetUsageRatio(10, 0)).toBe(0)
    })

    it('百分比文案', () => {
        expect(budgetUsagePercent(40, 100)).toBe('40%')
        expect(budgetUsagePercent(150, 100)).toBe('100%')
    })
})