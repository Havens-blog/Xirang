// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { Diagnosis } from '@/api/opsagent'
import RootCauseCard from './RootCauseCard.vue'

const DIAGNOSIS: Diagnosis = {
    id: 'd1',
    sessionId: 's1',
    tenant: 't1',
    rootCause: '数据库连接池耗尽',
    confidence: 0.87,
    severity: 'P3',
    riskLevel: 'high',
    conclusions: [{ text: '连接池配置偏小', citation: ['metric-p99'] }],
    disposition: [],
    citations: [
        { sourceType: 'log', sourceKey: 'log-0', snippet: "order-service-prod 10:10 起 'connection pool exhausted'" },
        { sourceType: 'metric', sourceKey: 'metric-1', snippet: 'postgres-order 活跃连接触顶' },
    ],
    degraded: false,
    truncated: true,
    trace: [],
}

describe('RootCauseCard 根因结论（原型 diagnosis-detail.html 口径）', () => {
    it('card-rca 主视觉：🔎 根因结论 + h2 根因 + 合并置信度', () => {
        const w = mount(RootCauseCard, { props: { diagnosis: DIAGNOSIS } })
        expect(w.find('.card').classes()).toContain('card-rca')
        expect(w.find('.rca-label').text()).toBe('🔎 根因结论')
        expect(w.find('.rca-cause').text()).toContain('数据库连接池耗尽')
        expect(w.find('.rca-confidence').text()).toBe('（置信度 87%）')
    })

    it('卡内折叠数据源引用（计数），展开后逐条：色点 + 来源名 + 陈述', async () => {
        const w = mount(RootCauseCard, { props: { diagnosis: DIAGNOSIS } })
        expect(w.find('.rca-citations-label').text()).toBe('数据源引用（2）')
        expect(w.findAll('.evidence-card')).toHaveLength(0)

        await w.find('.collapsible-header').trigger('click')
        expect(w.find('.collapsible-header').attributes('aria-expanded')).toBe('true')
        const evidences = w.findAll('.evidence-card')
        expect(evidences).toHaveLength(2)
        expect(evidences[0]!.find('.evidence-card-header').text()).toContain('日志')
        expect(evidences[0]!.find('.dot').attributes('style')).toContain('217 91% 60%')
        expect(evidences[0]!.find('.evidence-card-body').text()).toContain('connection pool exhausted')
        expect(evidences[1]!.find('.evidence-card-header').text()).toContain('指标')
    })

    it('空引用展开显示空态文案', async () => {
        const w = mount(RootCauseCard, { props: { diagnosis: { ...DIAGNOSIS, citations: [] } } })
        await w.find('.collapsible-header').trigger('click')
        expect(w.find('.evidence-empty').text()).toBe('无数据源引用')
    })
})