// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { NotifyChannel } from '@/api/opsagent'
import NotifyTab from './NotifyTab.vue'

const CHANNELS: NotifyChannel[] = [
    { channel: 'dingtalk', enabled: true },
    { channel: 'email', enabled: false },
]

describe('NotifyTab 通知渠道（原型 settings.html 四卡 + switch）', () => {
    it('渲染渠道图标卡：中文名 + 描述 + role=switch 状态', () => {
        const w = mount(NotifyTab, { props: { channels: CHANNELS, editable: true } })
        const rows = w.findAll('.channel-card')
        expect(rows).toHaveLength(2)
        expect(rows[0]!.find('.channel-card-name').text()).toBe('钉钉')
        expect(rows[0]!.find('.channel-card-desc').text()).toBe('值班群机器人 webhook')
        expect(rows[0]!.find('.switch').attributes('aria-checked')).toBe('true')
        expect(rows[1]!.find('.switch').attributes('aria-checked')).toBe('false')
        expect(w.find('.switch').attributes('role')).toBe('switch')
    })

    it('点击开关上抛 toggle(channel, enabled)', async () => {
        const w = mount(NotifyTab, { props: { channels: CHANNELS, editable: true } })
        await w.findAll('.switch')[1]!.trigger('click')
        expect(w.emitted('toggle')).toEqual([['email', true]])
    })

    it('不可编辑时开关禁用；未配置渠道显示空态', () => {
        const w = mount(NotifyTab, { props: { channels: CHANNELS, editable: false } })
        expect(w.find('.switch').attributes('disabled')).toBeDefined()
        const empty = mount(NotifyTab, { props: { channels: [], editable: true } })
        expect(empty.find('.notify-tab__empty').text()).toBe('未配置通知渠道')
    })

    it('说明条：管理员→开关即时保存；非管理员→仅管理员可调整', () => {
        const admin = mount(NotifyTab, { props: { channels: CHANNELS, editable: true } })
        expect(admin.find('.notify-tab__note').text()).toContain('即时保存到后端')
        const viewer = mount(NotifyTab, { props: { channels: CHANNELS, editable: false } })
        expect(viewer.find('.notify-tab__note').text()).toContain('仅平台管理员')
    })
})