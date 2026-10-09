// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ThinkingBlock from './ThinkingBlock.vue'

/**
 * ThinkingBlock 思考块（原型 chat.html thinkingRow 口径）：
 * 折叠块「思考过程 · 多工具编排」+ 4 步文案（①-④）逐条点亮。
 */

describe('ThinkingBlock 思考块（原型口径）', () => {
    beforeEach(() => {
        vi.useFakeTimers()
    })
    afterEach(() => {
        vi.useRealTimers()
    })

    it('渲染折叠块标题与 4 步原型文案（① 意图识别中… 等）', () => {
        const w = mount(ThinkingBlock, { props: { active: false } })
        expect(w.find('summary').text()).toBe('思考过程 · 多工具编排')
        const steps = w.findAll('.progress-step')
        expect(steps).toHaveLength(4)
        expect(steps.map((s) => s.text())).toEqual([
            '① 意图识别中…',
            '② 查询日志 / 资产 / 告警中…',
            '③ 诊断计算中…',
            '④ 生成报告中…',
        ])
    })

    it('active 时首步高亮，~700ms 逐条点亮（原型节奏）', async () => {
        const w = mount(ThinkingBlock, { props: { active: true } })
        const dots = () => w.findAll('.progress-step-dot')

        expect(dots()[0]!.classes()).toContain('progress-step-dot--active')
        await vi.advanceTimersByTimeAsync(700)
        expect(dots()[0]!.classes()).toContain('progress-step-dot--completed')
        expect(dots()[1]!.classes()).toContain('progress-step-dot--active')

        await vi.advanceTimersByTimeAsync(700 * 2)
        expect(dots()[2]!.classes()).toContain('progress-step-dot--completed')
        expect(dots()[0]!.classes()).not.toContain('progress-step-dot--active')
    })

    it('active=false 不启动计时', async () => {
        const w = mount(ThinkingBlock, { props: { active: false } })
        await vi.advanceTimersByTimeAsync(1500)
        const dots = w.findAll('.progress-step-dot')
        expect(dots.every((d) => d.classes().includes('progress-step-dot--active'))).toBe(false)
    })
})