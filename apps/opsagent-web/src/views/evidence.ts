/**
 * 运维 Agent 证据引用着色（跨页共享：对话排障内嵌证据 / 诊断详情根因卡引用）。
 *
 * 依据：docs/features/haven-opsagent/ui/prototype（证据行 = 色点 + 来源名 +
 * 一句陈述，色点语义色区分来源类型，分量与 opsagent-theme.css 的 --agent-*
 * tokens 同值）。source_type 依正冻结契约（api-handbook Data Contracts）
 * 为 4 类 log|metric|asset|alert。
 */

import type { CitationSourceType } from '@/api/opsagent'

export interface EvidenceSourceMeta {
    /** 来源名（原型口径：日志/指标/资产/告警） */
    label: string
    /** 色点 HSL 分量（与 --agent-* tokens 同值，模板用 hsl(...) 引用） */
    hsl: string
}

/** 证据 source_type → 语义色元数据 */
export const SOURCE_TYPE_META: Record<CitationSourceType, EvidenceSourceMeta> = {
    log: { label: '日志', hsl: '217 91% 60%' },      // --agent-log
    metric: { label: '指标', hsl: '142 71% 45%' },   // --agent-monitor
    asset: { label: '资产', hsl: '271 91% 65%' },    // --agent-inspector
    alert: { label: '告警', hsl: '199 89% 48%' },    // --agent-coord
}

/** 未知 source_type 的兜底元数据（容错，不吞数据） */
export function sourceTypeMeta(sourceType: string): EvidenceSourceMeta {
    return SOURCE_TYPE_META[sourceType as CitationSourceType] ?? {
        label: sourceType || '未知',
        hsl: '199 89% 48%',
    }
}