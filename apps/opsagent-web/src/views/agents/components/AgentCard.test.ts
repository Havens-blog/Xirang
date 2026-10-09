// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { AgentStat } from '@/api/opsagent'
import AgentCard from './AgentCard.vue'

const AGENT: AgentStat = { name: 'log_analyst', status: 'ok', tasksTotal: 42, tasksFailed: 3, avgDurationMs: 120 }

describe('AgentCard（原型 agent-management.html 卡片口径）', () => {
    it('英文名 + 角色行 + 运行中 pulse 徽标（无告警前缀）', () => {
        const w = mount(AgentCard, {
            props: { agent: AGENT },
            global: { stubs: { teleport: true, transition: false } },
        })
        expect(w.find('.agent-card-title h3').text()).toBe('LogAnalyst')
        expect(w.find('.agent-card-title p').text()).toBe('日志查询 · 日志分析')
        expect(w.find('.agent-icon').text()).toBe('📋')
        expect(w.find('.badge-ok').text()).toContain('运行中')
        expect(w.find('.pulse-dot').exists()).toBe(true)
    })

    it('异常状态 → ⚠️ 告警中 badge-low', () => {
        const w = mount(AgentCard, {
            props: { agent: { ...AGENT, status: 'error' } },
            global: { stubs: { teleport: true, transition: false } },
        })
        expect(w.find('.badge-low').text()).toContain('⚠️ 告警中')
        expect(w.find('.pulse-dot').exists()).toBe(false)
    })

    it('六行指标：活跃任务/正常运行时长/版本 无契约 → —；完成任务/平均延迟/成功率真实值', () => {
        const w = mount(AgentCard, {
            props: { agent: AGENT },
            global: { stubs: { teleport: true, transition: false } },
        })
        const values = w.findAll('.m-value').map((v) => v.text())
        expect(values).toEqual(['—', '42', '120ms', '92.9%', '—', '—'])
    })

    it('仅渲染 查看详情 按钮（无停止/重启假控件）', () => {
        const w = mount(AgentCard, {
            props: { agent: AGENT },
            global: { stubs: { teleport: true, transition: false } },
        })
        const btns = w.findAll('.agent-card-actions button')
        expect(btns).toHaveLength(1)
        expect(btns[0]!.text()).toBe('查看详情')
    })
})