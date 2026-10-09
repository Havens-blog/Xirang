// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { Diagnosis, SessionSummary } from '@/api/opsagent'
import SessionView from './SessionView.vue'

const SESSION: SessionSummary = {
    id: 's1',
    type: 'chat',
    query: 'order-service 最近一小时错误率上升，帮我看看',
    serviceName: 'order-service',
    intentType: 'diagnose',
    status: 'done',
    diagnosisId: 'd1',
    createdAt: '2026-10-08T09:45:00',
}

const DIAGNOSIS: Diagnosis = {
    id: 'd1',
    sessionId: 's1',
    tenant: 't1',
    rootCause: '数据库连接池耗尽',
    confidence: 0.9,
    severity: 'P3',
    riskLevel: 'low',
    conclusions: [{ text: '扩容 order-service 连接池至 300', citation: [] }],
    disposition: [],
    citations: [{ sourceType: 'metric', sourceKey: 'p99', snippet: 'postgres-order 活跃连接触顶 200/200' }],
    degraded: false,
    truncated: false,
    trace: [{ step: 1, agent: 'coordinator', action: 'intent', durationMs: 5, summary: '识别意图' }],
}

const mountView = (over: Partial<{ session: SessionSummary; diagnosis: Diagnosis | null; loading: boolean }> = {}) =>
    mount(SessionView, {
        props: {
            session: SESSION,
            diagnosis: DIAGNOSIS,
            loading: false,
            ...over,
        },
        global: {
            stubs: {
                RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
                TraceView: { props: ['trace', 'heading'], template: '<div class="trace-stub">{{ heading }}</div>' },
            },
        },
    })

describe('SessionView 会话内容视图（原型 history.html 口径）', () => {
    it('面包屑返回 + 标题行：服务名/排障 chip/来源/时间 + 查看详情链接', async () => {
        const w = mountView()
        expect(w.find('.breadcrumb').text()).toContain('↩ 返回会话列表')
        expect(w.find('.session-view__service').text()).toBe('order-service')
        expect(w.find('.agent-chip.diagnose').text()).toBe('排障')
        expect(w.find('.agent-chip.log').text()).toBe('对话')
        expect(w.find('.session-view__time').text()).toBe('09:45')
        const detail = w.find('a[href="/diagnosis/d1"]')
        expect(detail.exists()).toBe(true)
        expect(detail.text()).toBe('查看详情')

        await w.find('.breadcrumb-back').trigger('click')
        expect(w.emitted('back')).toBeTruthy()
    })

    it('🔎 诊断报告卡：合并根因/置信度 + 折叠数据源引用', async () => {
        const w = mountView()
        expect(w.find('.report-label').text()).toBe('🔎 诊断报告')
        expect(w.find('.report-cause').text()).toContain('数据库连接池耗尽')
        expect(w.find('.report-conf').text()).toBe('（置信度 90%）')
        expect(w.find('.report-cites-label').text()).toBe('数据源引用（1）')
        await w.find('.collapsible-header').trigger('click')
        expect(w.find('.evidence-card').text()).toBe('指标：postgres-order 活跃连接触顶 200/200')
    })

    it('💬 对话记录卡：用户提问原文 + Agent 回复摘录（由诊断数据合成）', () => {
        const w = mountView()
        expect(w.find('.card-title').text()).toBe('💬 对话记录')
        expect(w.find('.bubble--user').text()).toBe('order-service 最近一小时错误率上升，帮我看看')
        expect(w.find('.bubble--agent').text()).toContain('系统诊断回复：根因「数据库连接池耗尽」')
        expect(w.find('.bubble--agent').text()).toContain('建议扩容 order-service 连接池至 300')
        expect(w.find('.bubble--agent').text()).toContain('（完整内容见诊断报告）')
    })

    it('编排调用链卡使用历史页短标题', () => {
        const w = mountView()
        expect(w.find('.trace-stub').text()).toBe('🧬 编排调用链')
    })

    it('无诊断内容时显示占位说明', () => {
        const w = mountView({ diagnosis: null })
        expect(w.find('.session-view__nodata').text()).toContain('暂无诊断内容')
        expect(w.find('.report-label').exists()).toBe(false)
    })
})