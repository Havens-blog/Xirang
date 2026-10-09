// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { RiskWhitelistEntry } from '@/api/opsagent'
import SecurityTab from './SecurityTab.vue'

describe('SecurityTab 安全与租户（服务端强制项 + 白名单真实 CRUD）', () => {
    const base: RiskWhitelistEntry[] = [
        { tool: 'notify', riskLevel: 'low' },
        { tool: 'query_log', riskLevel: 'read' },
    ]

    it('白名单行展示中文工具名 + 风险档徽标 + 处置语义', () => {
        const w = mount(SecurityTab, { props: { whitelist: base } })
        const rows = w.findAll('.data-table tbody tr')
        expect(rows).toHaveLength(2)
        expect(rows[0]!.text()).toContain('发送通知')
        expect(rows[0]!.text()).toContain('低危')
        expect(rows[0]!.text()).toContain('白名单自动执行')
        expect(rows[1]!.text()).toContain('查询日志（logquery 联邦）')
        expect(rows[1]!.text()).toContain('自动执行')
    })

    it('空白名单 → 「暂无白名单条目」占位', () => {
        const w = mount(SecurityTab, { props: { whitelist: [] } })
        expect(w.text()).toContain('暂无白名单条目')
    })

    it('非管理员：不渲染新增区与移除列（真实写能力由 editable 门控）', () => {
        const w = mount(SecurityTab, { props: { whitelist: base } })
        expect(w.find('.whitelist-add').exists()).toBe(false)
        expect(w.findAll('.btn-remove')).toHaveLength(0)
    })

    it('管理员：工具下拉高危项禁选（optgroup「高危 · 不允许入白名单」）', () => {
        const w = mount(SecurityTab, { props: { whitelist: base, editable: true } })
        const highGroup = w.find('.whitelist-add select optgroup[label="高危 · 不允许入白名单"]')
        expect(highGroup.exists()).toBe(true)
        const options = highGroup.findAll('option')
        expect(options).toHaveLength(4)
        // happy-dom 布尔属性存在即 `''`（undefined = 不存在）
        expect(options.every((o) => o.attributes('disabled') !== undefined)).toBe(true)
        expect(options.map((o) => o.text())).toEqual([
            '重启服务（高危） · 不可入白名单',
            '回滚部署（高危） · 不可入白名单',
            '封禁 IP（高危） · 不可入白名单',
            '删除资产（高危） · 不可入白名单',
        ])
    })

    it('管理员：添加合法条目 → draft 上抛（含新条目且不可变）', async () => {
        const w = mount(SecurityTab, { props: { whitelist: base, editable: true } })
        await w.find('select[aria-label="选择工具"]').setValue('query_alert')
        await w.find('select[aria-label="风险档"]').setValue('read')
        await w.find('.whitelist-add .btn').trigger('click')

        const emitted = w.emitted('draft')
        expect(emitted).toHaveLength(1)
        expect(emitted![0]![0]).toEqual([
            ...base,
            { tool: 'query_alert', riskLevel: 'read' },
        ])
        // 原 prop 数组不可变
        expect(base).toHaveLength(2)
    })

    it('重复工具添加 → 内联错误，不触发 draft', async () => {
        const w = mount(SecurityTab, { props: { whitelist: base, editable: true } })
        await w.find('select[aria-label="选择工具"]').setValue('notify')
        await w.find('.whitelist-add .btn').trigger('click')
        expect(w.find('.whitelist-error').text()).toBe('工具 notify 已在白名单中')
        expect(w.emitted('draft')).toBeUndefined()
    })

    it('移除条目 → draft 上抛（过滤目标工具）', async () => {
        const w = mount(SecurityTab, { props: { whitelist: base, editable: true } })
        await w.find('.btn-remove[aria-label="移除 notify"]').trigger('click')
        expect(w.emitted('draft')![0]![0]).toEqual([{ tool: 'query_log', riskLevel: 'read' }])
    })

    it('多租户隔离两项：诚实呈现「服务端强制开启」徽标（无配置开关）', () => {
        const w = mount(SecurityTab, { props: { whitelist: [] } })
        const badges = w.findAll('.sec-row .badge')
        expect(badges).toHaveLength(2)
        for (const b of badges) {
            expect(b.text()).toBe('服务端强制开启')
        }
        const titles = w.findAll('.sec-row-title').map((t) => t.text())
        expect(titles).toEqual(['强制租户上下文注入', 'LLM 回复引用校验'])
    })

    it('底座接口契约四条目冻结版本徽标', () => {
        const w = mount(SecurityTab, { props: { whitelist: [] } })
        const rows = w.findAll('.contract-row')
        expect(rows).toHaveLength(4)
        expect(rows[0]!.text()).toContain('logquery')
        expect(rows[0]!.text()).toContain('已冻结 v1.2')
    })
})