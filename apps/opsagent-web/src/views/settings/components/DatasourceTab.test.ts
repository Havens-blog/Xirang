// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { Datasource } from '@/api/opsagent'
import DatasourceTab from './DatasourceTab.vue'

describe('DatasourceTab 数据源配置（原型 settings.html 五卡注册表口径）', () => {
    it('渲染前端注册表五卡：多云日志查询/告警链路/资产 CMDB/服务拓扑/变更与审计', () => {
        const w = mount(DatasourceTab, { props: { datasources: [] } })
        const cards = w.findAll('.ds-card')
        expect(cards).toHaveLength(5)
        const names = w.findAll('.ds-meta h3').map((h) => h.text())
        expect(names).toEqual(['多云日志查询', '告警链路', '资产 / CMDB', '服务拓扑', '变更与审计'])
    })

    it('连接态自后端 probes 推导：logquery 三探针全绿→已连接；无探针卡→未连接', () => {
        const datasources: Datasource[] = [
            { name: 'cdn', cloud: 'aliyun', ok: true, checkedAt: '2026-10-08T10:00:00Z' },
            { name: 'waf', cloud: 'aliyun', ok: true, checkedAt: '2026-10-08T10:00:00Z' },
            { name: 'lb', cloud: 'aliyun', ok: true, checkedAt: '2026-10-08T10:00:00Z' },
            { name: 'metric', cloud: 'aliyun', ok: false, checkedAt: '2026-10-08T10:00:00Z' },
        ]
        const w = mount(DatasourceTab, { props: { datasources } })
        const cards = w.findAll('.ds-card')
        // logquery（0 卡）已连接；alert（1 卡，metric 异常）未连接；topology/audit（3/4）未连接
        expect(cards[0]!.find('.ds-status').text()).toContain('已连接')
        expect(cards[1]!.find('.ds-status').text()).toContain('未连接')
        expect(cards[3]!.find('.ds-status').text()).toContain('未连接')
        expect(cards[4]!.find('.ds-status').text()).toContain('未连接')
    })

    it('点击卡头展开明细字段列表（aria-expanded 翻转）', async () => {
        const w = mount(DatasourceTab, { props: { datasources: [] } })
        const head = w.find('.ds-card-head')
        expect(head.attributes('aria-expanded')).toBe('false')
        await head.trigger('click')
        expect(w.find('.ds-card-head').attributes('aria-expanded')).toBe('true')
        const panelText = w.find('.ds-panel').text()
        expect(panelText).toContain('日志类型：CDN / WAF / LB')
        expect(panelText).toContain('查询模式：联邦实时（不落库）')
    })
})