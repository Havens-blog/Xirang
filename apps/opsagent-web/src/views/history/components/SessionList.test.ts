// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { SessionSummary } from '@/api/opsagent'
import SessionList from './SessionList.vue'

const SESSIONS: SessionSummary[] = [
    { id: 's1', type: 'chat', query: 'order-service 错误率上升', serviceName: 'order-service', intentType: 'diagnose', status: 'done', diagnosisId: 'd1', createdAt: '2026-10-08T09:45:00' },
    { id: 's2', type: 'alert', query: 'CPU 告警', serviceName: 'gateway', intentType: 'diagnose', status: 'running', createdAt: '2026-10-08T10:00:00' },
]

describe('SessionList 会话列表（原型 history.html 表格口径）', () => {
    it('空列表显示空态；不渲染表格与分页', () => {
        const w = mount(SessionList, { props: { sessions: [], total: 0, page: 1, limit: 20 } })
        expect(w.find('.session-list__empty').text()).toBe('暂无会话记录')
        expect(w.find('table').exists()).toBe(false)
        expect(w.find('.pagination').exists()).toBe(false)
    })

    it('表头列 = 服务名/意图/触发来源/时间/降级/操作；行内意图=排障 chip', () => {
        const w = mount(SessionList, { props: { sessions: SESSIONS } })
        const headers = w.findAll('thead th').map((h) => h.text())
        expect(headers).toEqual(['服务名', '意图', '触发来源', '时间', '降级', '操作'])
        const rows = w.findAll('tbody tr')
        expect(rows).toHaveLength(2)
        expect(rows[0]!.find('.session-row__service').text()).toBe('order-service')
        expect(rows[0]!.find('.agent-chip').text()).toBe('排障')
        // 触发来源：chat→对话 chip；alert→告警 badge
        expect(rows[0]!.text()).toContain('对话')
        expect(rows[1]!.text()).toContain('告警')
        // 时间 HH:mm
        expect(rows[0]!.find('.session-row__time').text()).toBe('09:45')
        // 降级列无数据 → 占位
        expect(rows[0]!.findAll('td')[4]!.text()).toBe('—')
    })

    it('查看按钮点击上抛 select（stop 冒泡）', async () => {
        const w = mount(SessionList, { props: { sessions: SESSIONS } })
        await w.findAll('tbody tr')[0]!.find('button').trigger('click')
        expect(w.emitted('select')).toBeTruthy()
        expect(w.emitted('select')![0]![0]).toEqual(SESSIONS[0])
    })

    it('分页条：共 N 条 + 当前页；点击翻页上抛 page-change', async () => {
        const w = mount(SessionList, { props: { sessions: SESSIONS, total: 25, page: 1, limit: 20 } })
        expect(w.find('.pagination__total').text()).toBe('共 25 条')
        expect(w.find('.page-btn.current').text()).toBe('1')
        const next = w.find('[aria-label="下一页"]')
        expect(next.attributes('disabled')).toBeUndefined()
        await next.trigger('click')
        expect(w.emitted('page-change')![0]![0]).toBe(2)
    })

    it('末页禁用下一页且禁用上一页于第 1 页', () => {
        const last = mount(SessionList, { props: { sessions: SESSIONS, total: 25, page: 2, limit: 20 } })
        expect(last.find('[aria-label="下一页"]').attributes('disabled')).toBeDefined()
        const first = mount(SessionList, { props: { sessions: SESSIONS, total: 25, page: 1, limit: 20 } })
        expect(first.find('[aria-label="上一页"]').attributes('disabled')).toBeDefined()
    })
})