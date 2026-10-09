import { describe, expect, it } from 'vitest'
import {
    actionLabel,
    agentLabel,
    AGENT_LABEL,
    dispositionBadge,
    formatConfidence,
    RISK_LEVEL_META,
    TRACE_ACTION_LABEL,
} from './logic'

describe('处置预案风险档徽标（任务 5.4 AC 处置预案风险档）', () => {
    it('read/low/high 全量中文 + 色阶', () => {
        expect(RISK_LEVEL_META).toEqual({
            read: { label: '只读', tone: 'info' },
            low: { label: '低危', tone: 'warning' },
            high: { label: '高危', tone: 'danger' },
        })
    })
})

describe('处置建议执行语义徽标（原型 diagnosis-detail.html）', () => {
    it('read→自动执行 / low→仅展示 · P3 可执行 / high→人工确认 / 未知→仅展示', () => {
        expect(dispositionBadge('read')).toEqual({ text: '自动执行', tone: 'ok' })
        expect(dispositionBadge('low')).toEqual({ text: '仅展示 · P3 可执行', tone: 'ghost' })
        expect(dispositionBadge('high')).toEqual({ text: '人工确认', tone: 'high' })
        expect(dispositionBadge('unknown')).toEqual({ text: '仅展示', tone: 'ghost' })
    })
})

describe('编排调用链标签', () => {
    it('action 中文标签全覆盖 + 未知兜底原样', () => {
        expect(TRACE_ACTION_LABEL.diagnose).toBe('诊断')
        expect(TRACE_ACTION_LABEL.notify).toBe('通知')
        expect(actionLabel('query')).toBe('查询')
        expect(actionLabel('custom')).toBe('custom')
    })

    it('agent 中文四角色（编排/日志查询/告警/诊断）+ 未知兜底原样', () => {
        expect(AGENT_LABEL.coordinator).toBe('编排')
        expect(AGENT_LABEL.log_analyst).toBe('日志查询')
        expect(AGENT_LABEL.monitor).toBe('告警')
        expect(AGENT_LABEL.inspector).toBe('诊断')
        expect(agentLabel('coordinator')).toBe('编排')
        expect(agentLabel('unknown')).toBe('unknown')
    })
})

describe('置信度格式化', () => {
    it('0~1 → 百分比', () => {
        expect(formatConfidence(0.877)).toBe('88%')
        expect(formatConfidence(0)).toBe('0%')
    })
})