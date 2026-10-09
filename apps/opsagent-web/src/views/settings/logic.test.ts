import { describe, expect, it } from 'vitest'
import type { PresetQuery, SettingsData } from '@/api/opsagent'
import {
    addWhitelistEntry,
    API_CONTRACTS,
    connectionStatusText,
    DATASOURCE_CARDS,
    dataSourceLabel,
    DATA_SOURCE_LABEL,
    dispositionForRisk,
    formatTime,
    mergeLLMProviders,
    mergeNotifyChannels,
    normalizeSettings,
    NOTIFY_CHANNEL_LABEL,
    NOTIFY_CHANNEL_META,
    PROVIDER_LABEL,
    PROVIDER_META,
    RISK_LEVEL_META,
    RISK_TOOL_CATALOG,
    riskToolLabel,
    savableLLMProviders,
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

describe('白名单工具目录与新增校验（真实 CRUD 支撑）', () => {
    it('九工具与后端风险注册表对齐：5 read-ish 4 read + notify low + 4 high', () => {
        expect(RISK_TOOL_CATALOG.map((t) => t.tool)).toEqual([
            'query_log',
            'diagnose',
            'query_alert',
            'query_asset',
            'notify',
            'restart_service',
            'rollback_deploy',
            'ban_ip',
            'delete_asset',
        ])
        expect(RISK_TOOL_CATALOG.filter((t) => t.level === 'read').map((t) => t.tool)).toEqual([
            'query_log',
            'diagnose',
            'query_alert',
            'query_asset',
        ])
        expect(RISK_TOOL_CATALOG.filter((t) => t.level === 'low').map((t) => t.tool)).toEqual(['notify'])
        expect(RISK_TOOL_CATALOG.filter((t) => t.level === 'high')).toHaveLength(4)
        expect(riskToolLabel('query_log')).toBe('查询日志（logquery 联邦）')
        expect(riskToolLabel('unknown_tool')).toBe('unknown_tool')
    })

    it('新增：合法追加 / 高危拒绝 / 非法档拒绝 / 重复拒绝', () => {
        const base = [{ tool: 'notify', riskLevel: 'low' as const }]
        const ok = addWhitelistEntry(base, { tool: 'query_log', riskLevel: 'read' })
        expect(ok.error).toBeUndefined()
        expect(ok.list).toEqual([
            { tool: 'notify', riskLevel: 'low' },
            { tool: 'query_log', riskLevel: 'read' },
        ])
        expect(base).toHaveLength(1) // 不可变

        expect(addWhitelistEntry(base, { tool: 'restart_service', riskLevel: 'low' }).error).toBe(
            '高危工具不允许入白名单',
        )
        expect(addWhitelistEntry([], { tool: 'x', riskLevel: 'high' }).error).toBe('高危工具不允许入白名单')
        expect(addWhitelistEntry([], { tool: 'x', riskLevel: 'nope' as never }).error).toBe('风险档仅支持 只读 / 低危')
        expect(addWhitelistEntry(base, { tool: 'notify', riskLevel: 'read' }).error).toBe('工具 notify 已在白名单中')
    })
})

describe('草稿骨架合并（空库初始化：注册表为基 / 后端覆盖 / 占位过滤）', () => {
    it('通知渠道：后端缺失渠道默认关闭补齐四卡，后端值覆盖状态', () => {
        const merged = mergeNotifyChannels([{ channel: 'dingtalk', enabled: true }])
        expect(merged).toEqual([
            { channel: 'dingtalk', enabled: true },
            { channel: 'feishu', enabled: false },
            { channel: 'wecom', enabled: false },
            { channel: 'email', enabled: false },
        ])
        expect(mergeNotifyChannels([])).toEqual([
            { channel: 'dingtalk', enabled: false },
            { channel: 'feishu', enabled: false },
            { channel: 'wecom', enabled: false },
            { channel: 'email', enabled: false },
        ])
    })

    it('LLM：后端缺失提供商空模型占位，既有值覆盖；savable 过滤空模型', () => {
        const merged = mergeLLMProviders([
            { name: 'qwen', default: true, model: 'qwen-plus', keyMasked: 'sk-****abcd' },
        ])
        expect(merged).toEqual([
            { name: 'qwen', default: true, model: 'qwen-plus', apiKey: '' },
            { name: 'deepseek', default: false, model: '', apiKey: '' },
            { name: 'openai', default: false, model: '', apiKey: '' },
        ])
        const savable = savableLLMProviders([
            { name: 'qwen', default: true, model: 'qwen-plus', apiKey: 'sk-x' },
            { name: 'deepseek', default: false, model: '  ', apiKey: '' },
        ])
        expect(savable.map((p) => p.name)).toEqual(['qwen'])
        expect(savable[0]!.apiKey).toBe('sk-x')
    })

    it('LLM：注册表外自定义提供商不丢失（保留后端条目 + 补注册表占位）', () => {
        const merged = mergeLLMProviders([
            { name: 'ollama-local', default: false, model: 'llama3:8b', keyMasked: '' },
        ])
        expect(merged.map((p) => p.name)).toEqual(['ollama-local', 'qwen', 'deepseek', 'openai'])
        expect(merged[0]).toEqual({ name: 'ollama-local', default: false, model: 'llama3:8b', apiKey: '' })
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