import { describe, expect, it } from 'vitest'
import type { PresetQuery, SettingsData } from '@/api/opsagent'
import {
    API_CONTRACTS,
    connectionStatusText,
    DATASOURCE_CARDS,
    dataSourceLabel,
    DATA_SOURCE_LABEL,
    dispositionForRisk,
    formatTime,
    normalizeSettings,
    NOTIFY_CHANNEL_LABEL,
    NOTIFY_CHANNEL_META,
    PROVIDER_LABEL,
    PROVIDER_META,
    RISK_LEVEL_META,
    SECURITY_SWITCHES,
    notifyChannelLabel,
    providerLabel,
    validatePreset,
} from './logic'

describe('数据源标签（任务 5.6 AC 数据源配置）', () => {
    it('5 数据源全覆盖 + 未知兜底', () => {
        expect(Object.keys(DATA_SOURCE_LABEL).sort()).toEqual(['asset', 'cdn', 'lb', 'metric', 'waf'])
        expect(dataSourceLabel('cdn')).toBe('CDN')
        expect(dataSourceLabel('unknown')).toBe('unknown')
    })
})

describe('通知渠道 / LLM 提供商标签', () => {
    it('4 渠道中文标签', () => {
        expect(NOTIFY_CHANNEL_LABEL).toEqual({ dingtalk: '钉钉', feishu: '飞书', wecom: '企业微信', email: '邮件' })
        expect(notifyChannelLabel('wecom')).toBe('企业微信')
    })

    it('3 提供商中文标签 + 未知兜底', () => {
        expect(PROVIDER_LABEL).toEqual({ qwen: '通义千问', deepseek: 'DeepSeek', openai: 'OpenAI' })
        expect(providerLabel('deepseek')).toBe('DeepSeek')
        expect(providerLabel('x')).toBe('x')
    })
})

describe('连接状态 / 风险档', () => {
    it('连接状态文案', () => {
        expect(connectionStatusText(true)).toBe('正常')
        expect(connectionStatusText(false)).toBe('异常')
    })

    it('风险档三态徽标（read/low/high）', () => {
        expect(RISK_LEVEL_META).toEqual({
            read: { label: '只读', tone: 'info' },
            low: { label: '低危', tone: 'warning' },
            high: { label: '高危', tone: 'danger' },
        })
    })
})

describe('时间展示', () => {
    it('RFC3339 截断；空值占位', () => {
        expect(formatTime('2026-10-08T15:04:05Z')).toBe('2026-10-08 15:04')
        expect(formatTime('')).toBe('—')
    })
})

describe('响应规整（Go nil slice → JSON null 契约兜底）', () => {
    it('各数组字段为 null 时回退空数组，guidedTemplates null 回退空对象', () => {
        const raw = {
            datasources: null,
            llmProviders: null,
            notifyChannels: null,
            riskWhitelist: null,
            presetQueries: null,
            guidedTemplates: null,
        } as unknown as SettingsData
        expect(normalizeSettings(raw)).toEqual({
            datasources: [],
            llmProviders: [],
            notifyChannels: [],
            riskWhitelist: [],
            presetQueries: [],
            guidedTemplates: {},
        })
    })

    it('已配置字段原样透传', () => {
        const raw: SettingsData = {
            datasources: [{ name: 'cdn', cloud: 'aliyun', ok: true, checkedAt: 't' }],
            llmProviders: [],
            notifyChannels: [{ channel: 'dingtalk', enabled: true }],
            riskWhitelist: [],
            presetQueries: [],
            guidedTemplates: { x: { template: 'tpl', channel: 'chat' } },
        }
        const out = normalizeSettings(raw)
        expect(out.datasources).toHaveLength(1)
        expect(out.notifyChannels[0]!.channel).toBe('dingtalk')
        expect(out.guidedTemplates.x!.template).toBe('tpl')
    })
})

describe('预置查询校验（任务 5.6 AC 预置查询可配置）', () => {
    function preset(serviceName: string, start: string, end: string): PresetQuery {
        return { id: '', label: 'x', params: { serviceName, timeframe: { startTime: start, endTime: end } } }
    }

    it('serviceName 非空 + 时间窗 ≤24h 合法；缺 serviceName / 超窗非法', () => {
        expect(validatePreset(preset('order-service', '2026-10-08T00:00:00Z', '2026-10-08T12:00:00Z'))).toBeNull()
        expect(validatePreset(preset('', '', ''))).toBe('缺少 serviceName')
        expect(validatePreset(preset('svc', '2026-10-08T00:00:00Z', '2026-10-10T00:00:00Z'))).toBe('时间窗超过 24 小时')
    })
})

describe('数据源注册表 / 通知渠道 / 提供商元数据（原型 settings.html）', () => {
    it('五卡注册表键序与探针组（logquery 三探针；topology/audit 无探针）', () => {
        expect(DATASOURCE_CARDS.map((c) => c.key)).toEqual(['logquery', 'alert', 'asset', 'topology', 'audit'])
        expect(DATASOURCE_CARDS[0]!.probeKeys).toEqual(['cdn', 'waf', 'lb'])
        expect(DATASOURCE_CARDS[3]!.probeKeys).toEqual([])
        expect(DATASOURCE_CARDS[4]!.probeKeys).toEqual([])
    })

    it('风险档 → 处置语义（read 自动执行 / low 白名单自动执行 / high 人工确认 P3）', () => {
        expect(dispositionForRisk('read')).toBe('自动执行')
        expect(dispositionForRisk('low')).toBe('白名单自动执行')
        expect(dispositionForRisk('high')).toBe('人工确认后执行（P3）')
    })

    it('底座接口契约四条目冻结版本', () => {
        expect(API_CONTRACTS.map((c) => c.name)).toEqual([
            'logquery（日志+诊断）',
            'alert（告警）',
            '资产 / MCP',
            'eiam（鉴权 + 租户）',
        ])
        expect(API_CONTRACTS[0]!.version).toBe('v1.2')
    })

    it('安全开关两强制项 + 渠道/提供商元数据存在', () => {
        expect(SECURITY_SWITCHES.map((s) => s.title)).toEqual(['强制租户上下文注入', 'LLM 回复引用校验'])
        expect(NOTIFY_CHANNEL_META.dingtalk!.desc).toBe('值班群机器人 webhook')
        expect(PROVIDER_META.qwen!.site).toBe('dashscope.aliyun.com')
    })
})