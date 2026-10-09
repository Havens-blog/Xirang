/**
 * 系统配置页纯展示逻辑（任务 5.6）。
 *
 * 无 Vue / DOM 运行时依赖，可被 node 环境单测直接导入。依据：
 * docs/features/haven-opsagent/design/api-handbook.md §8 + page-map.md「系统配置」。
 */

import type { PresetQuery, RiskLevel, SettingsData } from '@/api/opsagent'

/** 预置查询时间窗上限（与后端 24h 对齐） */
export const MAX_WINDOW_MS = 24 * 3600 * 1000

// ==================== 数据源注册表（原型 settings.html 五卡） ====================

export interface DatasourceCardMeta {
    /** 注册表键（未见后端 probe 的卡展示「未连接」） */
    key: string
    /** 卡片左上双字母图标 */
    icon: string
    /** 图标底色调（token 名） */
    gradToken: string
    name: string
    desc: string
    /** 展开明细（原型 · 字段列表，静态度量说明） */
    fields: string[]
    /** 对应后端 Datasource probe name（连接态求 all-ok；空 = 无 probe → 未连接） */
    probeKeys: string[]
}

/** 前端固定注册表（原型逐字）：连接态由后端 probes 推导，不虚构探针结果 */
export const DATASOURCE_CARDS: DatasourceCardMeta[] = [
    {
        key: 'logquery',
        icon: 'LOG',
        gradToken: '--agent-log',
        name: '多云日志查询',
        desc: 'CDN / WAF / LB 日志联邦查询 + 刷量诊断（阿里 SLS / 华为 LTS / AWS S3 / 腾讯 CLS）',
        fields: [
            '日志类型：CDN / WAF / LB',
            '云厂商：阿里云 / 华为云 / AWS / 腾讯云',
            '查询模式：联邦实时（不落库）',
        ],
        probeKeys: ['cdn', 'waf', 'lb'],
    },
    {
        key: 'alert',
        icon: 'AL',
        gradToken: '--agent-monitor',
        name: '告警链路',
        desc: '多渠道告警 + 变更检测（钉钉 / 飞书 / 企微 / 邮件）',
        fields: ['告警源：阈值规则 + 变更检测', '通知渠道：钉钉 / 飞书 / 企微 / 邮件'],
        probeKeys: ['metric'],
    },
    {
        key: 'asset',
        icon: 'CM',
        gradToken: '--primary',
        name: '资产 / CMDB',
        desc: '多云资产同步（ECS/RDS/Redis/EIP/NAS/OSS 等）+ ecmdb CMDB',
        fields: ['资源类型：ECS / RDS / Redis / EIP / NAS / OSS', '查询：MCP Tools + 内部接口'],
        probeKeys: ['asset'],
    },
    {
        key: 'topology',
        icon: 'TO',
        gradToken: '--agent-inspector',
        name: '服务拓扑',
        desc: 'dns / k8s 采集 + 图模型（依赖关系、故障传播）',
        fields: ['采集：dns_collector / k8s_collector', '模型：node / edge / graph'],
        probeKeys: [],
    },
    {
        key: 'audit',
        icon: 'AU',
        gradToken: '--severity-low',
        name: '变更与审计',
        desc: '审计留痕 + 变更单审计（change / change_order_audit）',
        fields: ['审计：操作留痕', '变更：变更单审计 + 变更检测'],
        probeKeys: [],
    },
]

// ==================== 数据源标签 ====================

/** 数据源 name → 中文（cdn/waf/lb/metric/asset） */
export const DATA_SOURCE_LABEL: Record<string, string> = {
    cdn: 'CDN',
    waf: 'WAF',
    lb: '负载均衡',
    metric: '指标',
    asset: '资产',
}

/** 未知名兜底：原样返回 */
export function dataSourceLabel(name: string): string {
    return DATA_SOURCE_LABEL[name] ?? name
}

// ==================== 通知渠道标签 ====================

/** 通知渠道 channel → 中文（4 渠道） */
export const NOTIFY_CHANNEL_LABEL: Record<string, string> = {
    dingtalk: '钉钉',
    feishu: '飞书',
    wecom: '企业微信',
    email: '邮件',
}

export function notifyChannelLabel(channel: string): string {
    return NOTIFY_CHANNEL_LABEL[channel] ?? channel
}

/** 通知渠道卡元数据（原型 settings.html 四卡：图标字 + 渐变 + 描述） */
export const NOTIFY_CHANNEL_META: Record<
    string,
    { icon: string; grad: string; desc: string }
> = {
    dingtalk: {
        icon: '钉',
        grad: 'linear-gradient(135deg, #0ea5e9, #0284c7)',
        desc: '值班群机器人 webhook',
    },
    feishu: {
        icon: '飞',
        grad: 'linear-gradient(135deg, #06b6d4, #0ea5e9)',
        desc: '告警通知机器人',
    },
    wecom: {
        icon: '企',
        grad: 'linear-gradient(135deg, #22c55e, #16a34a)',
        desc: '运维群 webhook',
    },
    email: {
        icon: '邮',
        grad: 'linear-gradient(135deg, #f59e0b, #ea580c)',
        desc: 'SMTP · 值班邮箱列表',
    },
}

// ==================== LLM 提供商标签 ====================

/** 3 LLM 提供商中文标签 */
export const PROVIDER_LABEL: Record<string, string> = {
    qwen: '通义千问',
    deepseek: 'DeepSeek',
    openai: 'OpenAI',
}

export function providerLabel(name: string): string {
    return PROVIDER_LABEL[name] ?? name
}

/** 提供商卡元数据（原型 settings.html：字母图标 + 渐变 + 站点） */
export const PROVIDER_META: Record<string, { letter: string; grad: string; site: string }> = {
    qwen: {
        letter: 'Q',
        grad: 'linear-gradient(135deg, #3b82f6, #4f46e5)',
        site: 'dashscope.aliyun.com',
    },
    deepseek: {
        letter: 'D',
        grad: 'linear-gradient(135deg, #0ea5e9, #6366f1)',
        site: 'api.deepseek.com',
    },
    openai: {
        letter: 'O',
        grad: 'linear-gradient(135deg, #10b981, #14b8a6)',
        site: 'openai.com',
    },
}

// ==================== 连接状态 ====================

/** 数据源连接状态文案（ok 布尔 → 正常/异常） */
export function connectionStatusText(ok: boolean): string {
    return ok ? '正常' : '异常'
}

// ==================== 风险档 ====================

export interface RiskLevelMeta {
    label: string
    tone: 'info' | 'warning' | 'danger'
}

/** 风险档徽标（read/low/high） */
export const RISK_LEVEL_META: Record<RiskLevel, RiskLevelMeta> = {
    read: { label: '只读', tone: 'info' },
    low: { label: '低危', tone: 'warning' },
    high: { label: '高危', tone: 'danger' },
}

// ==================== 安全与租户 ====================

/**
 * 风险档 → 处置语义（原型 settings.html 白名单表「处置」列，确定性推导：
 * read 自动执行 / low 白名单自动执行 / high 人工确认后执行（P3））。
 */
export function dispositionForRisk(riskLevel: RiskLevel): string {
    switch (riskLevel) {
        case 'read':
            return '自动执行'
        case 'low':
            return '白名单自动执行'
        case 'high':
            return '人工确认后执行（P3）'
    }
}

/** 底座接口契约清单（原型 settings.html 静态冻结版本，来自设计文档） */
export interface ApiContractItem {
    name: string
    version: string
}

export const API_CONTRACTS: ApiContractItem[] = [
    { name: 'logquery（日志+诊断）', version: 'v1.2' },
    { name: 'alert（告警）', version: 'v1.0' },
    { name: '资产 / MCP', version: 'v1.0' },
    { name: 'eiam（鉴权 + 租户）', version: 'v0.20' },
]

/** 安全开关（原型 settings.html 两个强制项；均服务端强制，前端只读展示） */
export const SECURITY_SWITCHES: { key: string; title: string; hint: string }[] = [
    {
        key: 'tenant-inject',
        title: '强制租户上下文注入',
        hint: 'tenant 一律由服务端从 eiam 会话派生并强制覆盖，LLM 输出中的租户字段丢弃',
    },
    {
        key: 'citation-check',
        title: 'LLM 回复引用校验',
        hint: '越界引用一律丢弃，不得回显未授权数据',
    },
]

// ==================== 响应规整（契约兜底） ====================

/**
 * GET /settings 响应规整：Go nil slice 序列化为 JSON null（settings 集合该 scope
 * 尚无文档时后端即返回 null），前端统一回退空数组/空对象，避免 `.map` on null。
 */
export function normalizeSettings(raw: SettingsData): SettingsData {
    return {
        datasources: raw.datasources ?? [],
        llmProviders: raw.llmProviders ?? [],
        notifyChannels: raw.notifyChannels ?? [],
        riskWhitelist: raw.riskWhitelist ?? [],
        presetQueries: raw.presetQueries ?? [],
        guidedTemplates: raw.guidedTemplates ?? {},
    }
}

// ==================== 时间展示 ====================

/** RFC3339 时间格式化（跨页共享，见 ../format） */
export { formatTime } from '../format'

// ==================== 预置查询校验 ====================

/**
 * 预置查询前端校验（服务端仍整体拒绝）：serviceName 非空、时间窗 ≤24h。
 * 返回错误文案或 null（合法）。
 */
export function validatePreset(preset: PresetQuery): string | null {
    if (!preset?.params?.serviceName?.trim()) {
        return '缺少 serviceName'
    }
    const tf = preset.params.timeframe
    if (tf?.startTime && tf?.endTime) {
        const s = Date.parse(tf.startTime)
        const e = Date.parse(tf.endTime)
        if (!Number.isNaN(s) && !Number.isNaN(e) && e - s > MAX_WINDOW_MS) {
            return '时间窗超过 24 小时'
        }
    }
    return null
}