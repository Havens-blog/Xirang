import { describe, expect, it } from 'vitest'
import type { ChatData } from '@/api/opsagent'
import {
    chatTypeLabel,
    degradeBadge,
    hasDiagnosisReport,
    resolveCitation,
    SOURCE_TYPE_META,
    sourceTypeMeta,
    THINKING_STEP_TEXTS,
} from './logic'

function chat(over: Partial<ChatData>): ChatData {
    return {
        sessionId: 's1',
        diagnosis: null,
        report: '',
        type: 'report',
        needsClarify: false,
        candidates: [],
        degradeLevel: 0,
        presets: [],
        ...over,
    }
}

describe('chat 响应 type 判别（任务 5.2 AC#4）', () => {
    it('各 type 用户侧徽标文案', () => {
        expect(chatTypeLabel('report')).toBe('诊断报告')
        expect(chatTypeLabel('preset_entries')).toBe('预置查询')
        expect(chatTypeLabel('guided')).toBe('超能力引导')
        expect(chatTypeLabel('clarify')).toBe('待澄清')
        expect(chatTypeLabel('degraded_notice')).toBe('降级模式')
    })

    it('仅 type=report 且 diagnosis 非空视为携带报告正文', () => {
        expect(hasDiagnosisReport(chat({ type: 'report', diagnosis: { report: true } as never }))).toBe(true)
        expect(hasDiagnosisReport(chat({ type: 'clarify', needsClarify: true }))).toBe(false)
        expect(hasDiagnosisReport(chat({ type: 'preset_entries', degradeLevel: 2 }))).toBe(false)
        expect(hasDiagnosisReport(chat({ type: 'report', diagnosis: null }))).toBe(false)
    })
})

describe('降级徽标（api-handbook degradeLevel）', () => {
    it('0=正常无徽标；1/2/3 对应模板/预置入口/明示降级', () => {
        expect(degradeBadge(0)).toBeNull()
        expect(degradeBadge(1)).toEqual({ text: '模板降级', tone: 'warning' })
        expect(degradeBadge(2)).toEqual({ text: '预置查询入口', tone: 'warning' })
        expect(degradeBadge(3)).toEqual({ text: '降级模式', tone: 'danger' })
    })
})

describe('证据 source_type → 语义色（原型：色点 + 来源名）', () => {
    it('四类 source_type 均有中文来源名 + HSL 语义色分量', () => {
        expect(Object.keys(SOURCE_TYPE_META).sort()).toEqual(['alert', 'asset', 'log', 'metric'])
        expect(SOURCE_TYPE_META.log.label).toBe('日志')
        expect(SOURCE_TYPE_META.metric.label).toBe('指标')
        expect(SOURCE_TYPE_META.asset.label).toBe('资产')
        expect(SOURCE_TYPE_META.alert.label).toBe('告警')
        for (const m of Object.values(SOURCE_TYPE_META)) {
            expect(m.hsl).toMatch(/^\d+ \d+% \d+%$/)
        }
    })

    it('未知 source_type 走兜底元数据（不吞数据）', () => {
        const m = sourceTypeMeta('unknown')
        expect(m.label).toBe('unknown')
        expect(m.hsl).toBe('199 89% 48%')
    })
})

describe('思考块 4 步文案（原型 chat.html stageSteps）', () => {
    it('①-④ 文案逐字对齐原型', () => {
        expect(THINKING_STEP_TEXTS).toEqual([
            '① 意图识别中…',
            '② 查询日志 / 资产 / 告警中…',
            '③ 诊断计算中…',
            '④ 生成报告中…',
        ])
    })
})

describe('引用回指解析', () => {
    it('sourceKey 命中返回引用条目，未命中返回 undefined', () => {
        const citations = [{ sourceKey: 'log-0' }, { sourceKey: 'metric-1' }]
        expect(resolveCitation('metric-1', citations)?.sourceKey).toBe('metric-1')
        expect(resolveCitation('nope', citations)).toBeUndefined()
    })
})