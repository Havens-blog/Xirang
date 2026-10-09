// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { RiskStats } from '@/api/opsagent'
import StatCards from './StatCards.vue'

describe('StatCards 统计卡（原型 risk-center.html 口径）', () => {
    it('渲染三卡按原型顺序：高危待确认(highRisk)/今日诊断(todayNew)/待查看(pendingView)', () => {
        const stats: RiskStats = { pendingView: 7, todayNew: 3, highRisk: 2 }
        const w = mount(StatCards, { props: { stats } })
        const cards = w.findAll('.stat-card')
        expect(cards).toHaveLength(3)
        expect(cards[0]!.find('.stat-card__label').text()).toBe('高危待确认')
        expect(cards[0]!.find('.stat-card__value').text()).toBe('2')
        expect(cards[1]!.find('.stat-card__label').text()).toBe('今日诊断')
        expect(cards[1]!.find('.stat-card__value').text()).toBe('3')
        expect(cards[2]!.find('.stat-card__label').text()).toBe('待查看')
        expect(cards[2]!.find('.stat-card__value').text()).toBe('7')
    })

    it('数值色阶：高危红 / 待查看橙（severity tokens）', () => {
        const stats: RiskStats = { pendingView: 1, todayNew: 1, highRisk: 1 }
        const w = mount(StatCards, { props: { stats } })
        expect(w.findAll('.stat-card__value--high')).toHaveLength(1)
        expect(w.findAll('.stat-card__value--low')).toHaveLength(1)
    })
})