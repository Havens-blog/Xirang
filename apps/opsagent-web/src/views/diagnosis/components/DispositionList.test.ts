// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { DispositionStep } from '@/api/opsagent'
import DispositionList from './DispositionList.vue'

const DISPOSITION: DispositionStep[] = [
    { step: '扩容连接池', risk: 'low', action: '立即扩容 order-service 连接池至 300', executed: false },
    { step: '回滚最近变更', risk: 'high', action: '核对最近一次部署是否引入连接泄漏', executed: false },
    { step: '观察回落', risk: 'read', action: '观察连接等待队列是否回落', executed: false },
]

describe('DispositionList 处置建议（原型 diagnosis-detail.html 口径，只读）', () => {
    it('空预案且无回退显示空态', () => {
        const w = mount(DispositionList, { props: { disposition: [], fallback: [] } })
        expect(w.find('.disposition-title').text()).toBe('处置建议')
        expect(w.find('.disposition-empty').text()).toBe('无处置建议')
    })

    it('逐条渲染动作文案 + 按风险档的执行语义徽标（P3 可执行/人工确认/自动执行）', () => {
        const w = mount(DispositionList, { props: { disposition: DISPOSITION, fallback: [] } })
        const items = w.findAll('li')
        expect(items).toHaveLength(3)
        expect(items[0]!.text()).toContain('立即扩容 order-service 连接池至 300')
        expect(items[0]!.text()).toContain('仅展示 · P3 可执行')
        expect(items[1]!.text()).toContain('人工确认')
        expect(items[2]!.text()).toContain('自动执行')
        // 只读：不渲染任何执行按钮（Hard Rule）
        expect(w.findAll('button')).toHaveLength(0)
    })

    it('disposition 为空时回退结论列表（无徽标）', () => {
        const w = mount(DispositionList, {
            props: { disposition: [], fallback: ['扩容连接池至 300', '核对部署变更'] },
        })
        const items = w.findAll('li')
        expect(items.map((i) => i.text())).toEqual(['扩容连接池至 300', '核对部署变更'])
        expect(w.findAll('.badge')).toHaveLength(0)
    })
})