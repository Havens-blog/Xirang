// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { TraceStep } from '@/api/opsagent'
import TraceView from './TraceView.vue'

describe('TraceView 编排调用链（原型 diagnosis-detail.html 口径：可折叠 + code-block）', () => {
    it('空 trace：默认折叠，展开显示空态文案', async () => {
        const w = mount(TraceView, { props: { trace: [] } })
        expect(w.find('.collapsible-header').text()).toContain('🧬 编排调用链（供 SRE 二次核实）')
        expect(w.find('.collapsible-header').attributes('aria-expanded')).toBe('false')
        expect(w.find('.code-block').exists()).toBe(false)

        await w.find('.collapsible-header').trigger('click')
        expect(w.find('.collapsible-header').attributes('aria-expanded')).toBe('true')
        expect(w.find('.code-block').text()).toBe('无编排调用链记录')
    })

    it('展开后按步输出文本链：中文四角色 + action(source) + 摘要 + 耗时 + 降级级', async () => {
        const trace: TraceStep[] = [
            { step: 1, agent: 'coordinator', action: 'intent', durationMs: 5, summary: '识别意图' },
            { step: 2, agent: 'log_analyst', action: 'query', source: 'logquery', durationMs: 120, summary: '查询日志', degradeLevel: 2 },
        ]
        const w = mount(TraceView, { props: { trace } })
        await w.find('.collapsible-header').trigger('click')
        const text = w.find('.code-block').text()
        expect(text).toContain('→ 编排.intent → 识别意图 (5ms)')
        expect(text).toContain('→ 日志查询.query(logquery) → 查询日志 [降级 L2] (120ms)')
    })
})