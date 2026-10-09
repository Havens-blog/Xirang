// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { Diagnosis } from '@/api/opsagent'
import ChatPanel, { type ChatMessage } from './ChatPanel.vue'

/**
 * ChatPanel 对话排障面板（原型 chat.html 口径）：
 * 空态（💬 + 标题 + 示例 chips）/ 用户与 Agent 角色行 / 报告卡（合并结论标题 +
 * 处置建议 + 截断句 + 折叠数据源引用 + 查看详情）/ 粘底输入条（placeholder +
 * 500 字上限 + 字数/模式提示）。
 */

const DIAGNOSIS: Diagnosis = {
    id: 'd-1',
    sessionId: 's-1',
    tenant: 'tenant-1',
    rootCause: '数据库连接池耗尽',
    confidence: 0.85,
    severity: 'P3',
    riskLevel: 'low',
    conclusions: [{ text: '立即扩容连接池', citation: ['log-0'] }],
    disposition: [],
    citations: [
        { sourceType: 'log', sourceKey: 'log-0', snippet: 'connection pool exhausted' },
        { sourceType: 'metric', sourceKey: 'metric-1', snippet: '活跃连接触顶' },
    ],
    degraded: false,
    truncated: true,
    trace: [],
}

function mountPanel(over: Partial<{ messages: ChatMessage[]; diagnosis: Diagnosis | null; degradeLevel: number; busy: boolean; username: string }> = {}) {
    return mount(ChatPanel, {
        props: {
            messages: [],
            diagnosis: null,
            degradeLevel: 0,
            busy: false,
            username: 'havens',
            ...over,
        },
        global: {
            stubs: {
                RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
                ThinkingBlock: { props: ['active'], template: '<details class="thinking-block"><summary>思考过程 · 多工具编排</summary></details>' },
            },
        },
    })
}

describe('ChatPanel（原型 chat.html 口径）', () => {
    it('空态：💬 + 「有什么故障需要排查？」+ 3 枚示例 chip', () => {
        const w = mountPanel()
        expect(w.find('.empty-state .icon').text()).toBe('💬')
        expect(w.find('.empty-state h3').text()).toBe('有什么故障需要排查？')
        expect(w.find('.empty-state p').text()).toBe('试试下面的示例，或直接输入你的问题')
        const chips = w.findAll('.chip')
        expect(chips).toHaveLength(3)
        expect(chips.map((c) => c.text())).toEqual(['order-service 错误率上升', 'CDN 是否被刷', 'CDN 流量突增'])
    })

    it('chip 点击回填输入框（不触发提交）', async () => {
        const w = mountPanel()
        await w.findAll('.chip')[0]!.trigger('click')
        expect((w.find('#chatInput').element as HTMLInputElement).value).toBe(
            'order-service 最近一小时错误率上升，帮我看看',
        )
    })

    it('用户行：首字母头像 + havens + 气泡；Agent 行：🎯 + Haven 运维 Agent + 编排 chip', () => {
        const w = mountPanel({
            messages: [
                { role: 'user', text: 'order-service 错误率上升' },
                { role: 'assistant', text: '正在为您排查…' },
            ],
        })
        const userRow = w.findAll('.chat-row')[0]!
        expect(userRow.classes()).toContain('user')
        expect(userRow.find('.who-icon').text()).toBe('H')
        expect(userRow.find('.name').text()).toBe('havens')
        expect(userRow.find('.bubble').text()).toBe('order-service 错误率上升')

        const agentRow = w.findAll('.chat-row')[1]!
        expect(agentRow.classes()).toContain('agent')
        expect(agentRow.find('.who-icon').text()).toBe('🎯')
        expect(agentRow.find('.name').text()).toBe('Haven 运维 Agent')
        expect(agentRow.find('.agent-chip').text()).toBe('编排')
    })

    it('报告卡：合并结论标题 + 处置建议 + 截断句 + 折叠数据源引用（N）+ 查看详情链接', async () => {
        const w = mountPanel({ diagnosis: DIAGNOSIS })
        const card = w.find('.report-card')
        expect(card.find('h3').text()).toBe('诊断结论：数据库连接池耗尽（置信度 85%）')
        expect(card.find('.report-card__advice strong').text()).toBe('处置建议')
        expect(card.find('.report-card__advice li').text()).toBe('立即扩容连接池')
        expect(card.find('.report-card__truncation').text()).toContain('结果已截断')
        // 引用折叠头带计数；展开后 2 条证据（色点 + 来源名：陈述）
        expect(card.find('.collapsible-header').text()).toContain('数据源引用（2）')
        expect(card.findAll('.evidence-card')).toHaveLength(0)
        await card.find('.collapsible-header').trigger('click')
        const evidences = w.findAll('.evidence-card')
        expect(evidences).toHaveLength(2)
        expect(evidences[0]!.text()).toContain('日志：connection pool exhausted')
        expect(evidences[0]!.find('.dot').exists()).toBe(true)
        expect(card.find('.report-card__footer a').text()).toBe('查看详情 →')
        expect(card.find('.report-card__footer a').attributes('href')).toBe('/diagnosis/d-1')
    })

    it('输入条：placeholder / 500 上限 / 字数与模式提示（正常）', async () => {
        const w = mountPanel()
        const input = w.find('#chatInput')
        expect(input.attributes('placeholder')).toBe('描述故障现象，如：某服务最近一小时错误率上升…')
        expect(input.attributes('maxlength')).toBe('500')
        expect(w.find('.chat-hint').text()).toContain('0/500')
        expect(w.find('.chat-hint').text()).toContain('当前模式：')
        expect(w.find('.chat-hint').text()).toContain('正常')

        await input.setValue('abc')
        expect(w.find('.chat-hint').text()).toContain('3/500')
    })

    it('降级模式：提示行显示「降级」，busy 时出现思考行（🤖 + 编排中）', () => {
        const w = mountPanel({ degradeLevel: 3, busy: true })
        expect(w.find('.chat-hint').text()).toContain('降级')
        const thinking = w.findAll('.chat-row.agent')
        const thinkRow = thinking[thinking.length - 1]!
        expect(thinkRow.find('.who-icon').text()).toBe('🤖')
        expect(thinkRow.find('.agent-chip').text()).toBe('编排中')
        expect(thinkRow.find('summary').text()).toBe('思考过程 · 多工具编排')
    })
})