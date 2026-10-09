<template>
  <div class="session-view">
    <!-- 页内切换面包屑（原型 history.html：↩ 返回会话列表 / 会话） -->
    <div class="breadcrumb">
      <button type="button" class="breadcrumb-back" @click="emit('back')">↩ 返回会话列表</button>
      <span class="sep">/</span>
      <span>会话</span>
    </div>

    <!-- 标题行（原型：h1 服务名 + 排障 chip + 来源 + 时间 + 查看详情） -->
    <div class="session-view__title-row">
      <div class="session-view__title-line">
        <h1 class="session-view__service">{{ session.serviceName || '未指定服务' }}</h1>
        <span class="agent-chip diagnose">排障</span>
        <span v-if="session.type === 'chat'" class="agent-chip log">对话</span>
        <span v-else class="badge badge-ghost">告警</span>
        <span class="session-view__time">{{ formatClock(session.createdAt) }}</span>
      </div>
      <RouterLink
        v-if="session.diagnosisId"
        class="btn btn-outline btn-sm"
        :to="`/diagnosis/${session.diagnosisId}`"
      >
        查看详情
      </RouterLink>
    </div>

    <el-skeleton v-if="loading" :rows="8" animated />
    <template v-else-if="diagnosis">
      <!-- 诊断报告（原型：🔎 诊断报告 + 合并根因/置信度 + 折叠引用） -->
      <div class="card card-rca">
        <p class="report-label">🔎 诊断报告</p>
        <h2 class="report-cause">
          {{ diagnosis.rootCause || '（无明确根因结论）' }}
          <span class="report-conf">（置信度 {{ formatConfidence(diagnosis.confidence) }}）</span>
        </h2>

        <div
          class="collapsible-header"
          :aria-expanded="citationsOpen"
          @click="citationsOpen = !citationsOpen"
        >
          <span class="report-cites-label">数据源引用（{{ diagnosis.citations.length }}）</span>
          <span class="chevron">▾</span>
        </div>
        <div v-if="citationsOpen" class="collapsible-content">
          <div
            v-for="(c, i) in diagnosis.citations"
            :key="`${c.sourceKey}-${i}`"
            class="evidence-card"
          >
            {{ sourceLabel(c.sourceType) }}：{{ c.snippet }}
          </div>
          <p v-if="diagnosis.citations.length === 0" class="evidence-empty">无数据源引用</p>
        </div>
      </div>

      <!-- 对话记录（原型：💬 用户原文 + Agent 回复摘录） -->
      <div class="card">
        <h3 class="card-title">💬 对话记录</h3>
        <div class="conv-row">
          <div class="avatar" aria-hidden="true">H</div>
          <div class="bubble bubble--user">{{ session.query }}</div>
        </div>
        <div class="conv-row">
          <div class="who-icon" aria-hidden="true">🎯</div>
          <div class="bubble bubble--agent">{{ agentReplyExcerpt }}</div>
        </div>
      </div>

      <!-- 编排调用链（原型历史页：折叠 + code-block） -->
      <TraceView :trace="diagnosis.trace" heading="🧬 编排调用链" />
    </template>
    <p v-else class="session-view__nodata">该会话暂无诊断内容（可能仍在运行或已失败）。</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { CitationSourceType, Diagnosis, SessionSummary } from '@/api/opsagent'
import { sourceTypeMeta } from '@/views/evidence'
import { formatConfidence } from '../../diagnosis/logic'
import { formatClock } from '../logic'
import TraceView from '../../diagnosis/components/TraceView.vue'

const props = defineProps<{
    session: SessionSummary
    diagnosis: Diagnosis | null
    loading?: boolean
}>()

const emit = defineEmits<{ (e: 'back'): void }>()

const citationsOpen = ref(false)

/** Agent 回复摘录：由真实诊断数据合成（原型同样式文案） */
const agentReplyExcerpt = computed(() => {
    const d = props.diagnosis
    if (!d) return ''
    const root = d.rootCause || '（无明确根因）'
    const first = d.conclusions[0]?.text
    const advice = first ? `建议${first}` : ''
    return `系统诊断回复：根因「${root}」${advice ? `，${advice}` : ''}…（完整内容见诊断报告）`
})

function sourceLabel(sourceType: CitationSourceType): string {
    return sourceTypeMeta(sourceType).label
}
</script>

<style scoped>
.breadcrumb {
  display: flex;
  gap: 6px;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  margin-bottom: 16px;
}
.breadcrumb-back {
  background: none;
  border: none;
  padding: 0;
  font-size: 13px;
  color: hsl(var(--primary));
  cursor: pointer;
  font-family: inherit;
}
.session-view__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.session-view__title-line {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.session-view__service {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
  letter-spacing: -0.02em;
}
.session-view__time {
  font-family: monospace;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}
.agent-chip {
  font-size: 11px;
  padding: 2px 10px;
  border-radius: 9999px;
  font-weight: 500;
}
.agent-chip.diagnose {
  background: hsl(var(--agent-inspector) / 0.15);
  color: hsl(var(--agent-inspector));
  border: 1px solid hsl(var(--agent-inspector) / 0.3);
}
.agent-chip.log {
  background: hsl(var(--agent-log) / 0.15);
  color: hsl(var(--agent-log));
  border: 1px solid hsl(var(--agent-log) / 0.3);
}
.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border-radius: 9999px;
  padding: 3px 12px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}
.badge-ghost {
  border: 1px solid hsl(var(--border));
  color: hsl(var(--muted-foreground));
}
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 36px;
  padding: 0 16px;
  border-radius: 8px;
  border: 1px solid transparent;
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  text-decoration: none;
}
.btn-outline {
  background: transparent;
  border-color: hsl(var(--border));
  color: hsl(var(--foreground));
}
.btn-outline:hover {
  background: hsl(var(--accent));
}
.btn-sm {
  height: 30px;
  padding: 0 12px;
  font-size: 12px;
  border-radius: 6px;
}
.card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
  padding: 20px;
  color: hsl(var(--card-foreground));
  margin-bottom: 16px;
}
.card-rca {
  border-left: 3px solid hsl(var(--primary));
}
.report-label {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  margin: 0 0 6px;
}
.report-cause {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 10px;
  color: hsl(var(--foreground));
}
.report-conf {
  font-size: 13px;
  font-weight: 500;
  color: hsl(var(--muted-foreground));
}
.collapsible-header {
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0;
}
.report-cites-label {
  font-size: 13px;
  color: hsl(var(--muted-foreground));
}
.chevron {
  transition: transform 0.15s;
  font-size: 11px;
  color: hsl(var(--muted-foreground));
}
.collapsible-header[aria-expanded='true'] .chevron {
  transform: rotate(180deg);
}
.collapsible-content {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.evidence-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 12px 14px;
  font-size: 13px;
  color: hsl(var(--foreground));
  line-height: 1.6;
}
.evidence-empty {
  margin: 0;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  font-style: italic;
}
.card-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 12px;
}
.conv-row {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}
.conv-row:last-child {
  margin-bottom: 0;
}
.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  background: linear-gradient(135deg, hsl(var(--primary)), hsl(271 91% 65%));
  color: hsl(var(--primary-foreground));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
}
.who-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: hsl(var(--primary) / 0.18);
  border: 1px solid hsl(var(--primary) / 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  flex-shrink: 0;
}
.bubble {
  max-width: 100%;
  padding: 12px 16px;
  font-size: 13px;
  word-break: break-word;
}
.bubble--user {
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  border-radius: 16px 16px 4px 16px;
}
.bubble--agent {
  background: hsl(var(--muted) / 0.4);
  border-radius: 8px;
  color: hsl(var(--muted-foreground));
}
.session-view__nodata {
  margin: 0;
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
}
</style>