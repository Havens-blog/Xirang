<template>
  <div class="chat-panel">
    <div class="chat-scroll" aria-live="polite">
      <!-- 空态（原型 chat.html：💬 + h3 + 副题 + 3 枚示例 chip） -->
      <div v-if="messages.length === 0 && !busy" class="empty-state">
        <div class="icon" aria-hidden="true">💬</div>
        <h3>有什么故障需要排查？</h3>
        <p>试试下面的示例，或直接输入你的问题</p>
        <div class="suggestion-chips">
          <button
            v-for="chip in SUGGESTIONS"
            :key="chip.label"
            type="button"
            class="chip"
            :data-q="chip.query"
            @click="fillSuggestion(chip.query)"
          >
            {{ chip.label }}
          </button>
        </div>
      </div>

      <!-- 对话流 -->
      <div v-for="(m, i) in messages" :key="i" class="chat-row" :class="m.role === 'user' ? 'user' : 'agent'">
        <div v-if="m.role === 'user'" class="who-icon">{{ avatarLetter }}</div>
        <div v-else class="who-icon" aria-hidden="true">🎯</div>
        <div class="bubble-wrap">
          <div class="bubble-head">
            <span class="name">{{ m.role === 'user' ? userName : 'Haven 运维 Agent' }}</span>
            <span v-if="m.role === 'assistant'" class="agent-chip agent-chip-coord">编排</span>
          </div>
          <div v-if="m.role === 'user'" class="bubble">{{ m.text }}</div>
          <div v-else class="bubble bubble--agent">{{ m.text }}</div>
        </div>
      </div>

      <!-- 报告卡（原型 agentRow：诊断结论 + 处置建议 + 截断句 + 折叠引用 + 查看详情） -->
      <div v-if="diagnosis" class="chat-row agent">
        <div class="who-icon" aria-hidden="true">🎯</div>
        <div class="bubble-wrap">
          <div class="bubble-head">
            <span class="name">Haven 运维 Agent</span>
            <span class="agent-chip agent-chip-coord">编排</span>
          </div>
          <div class="report-card">
            <h3>诊断结论：{{ diagnosis.rootCause || '（无明确根因）' }}（置信度 {{ confidencePct }}%）</h3>

            <div v-if="degradeBadgeMeta" class="report-card__badges">
              <span class="badge badge-warning">{{ degradeBadgeMeta.text }}</span>
            </div>

            <div v-if="diagnosis.conclusions.length" class="report-card__advice">
              <strong>处置建议</strong>
              <ul>
                <li v-for="(c, i) in diagnosis.conclusions" :key="i">{{ c.text }}</li>
              </ul>
            </div>

            <div v-if="diagnosis.truncated" class="report-card__truncation">
              <span class="badge badge-muted" role="status">结果已截断，请缩小时间窗或服务范围</span>
            </div>

            <div class="collapsible-header" :aria-expanded="citationsOpen" @click="citationsOpen = !citationsOpen">
              <span>数据源引用（{{ diagnosis.citations.length }}）</span>
              <span class="chevron">▾</span>
            </div>
            <div v-if="citationsOpen" class="collapsible-content">
              <div
                v-for="(c, i) in diagnosis.citations"
                :key="`${c.sourceKey}-${i}`"
                class="evidence-card"
              >
                <div class="evidence-card-header">
                  <span class="dot" :style="{ background: `hsl(${sourceColor(c.sourceType)})` }" aria-hidden="true" />
                  <span>{{ sourceLabel(c.sourceType) }}：{{ c.snippet }}</span>
                </div>
              </div>
              <p v-if="diagnosis.citations.length === 0" class="evidence-empty">无数据源引用</p>
            </div>

            <div class="report-card__footer">
              <RouterLink class="btn btn-sm btn-outline" :to="`/diagnosis/${diagnosis.id}`">查看详情 →</RouterLink>
            </div>
          </div>
        </div>
      </div>

      <!-- 思考行（原型 thinkingRow：🤖 + 编排中 + 折叠思考块） -->
      <div v-if="busy" class="chat-row agent">
        <div class="who-icon" aria-hidden="true">🤖</div>
        <div class="bubble-wrap">
          <div class="bubble-head">
            <span class="name">Haven 运维 Agent</span>
            <span class="agent-chip agent-chip-coord">编排中</span>
          </div>
          <ThinkingBlock :active="busy" />
        </div>
      </div>
    </div>

    <!-- 输入条（原型 .chat-input-bar：sticky + 字数 + 当前模式） -->
    <form class="chat-input-bar" @submit.prevent="onSubmit">
      <div class="chat-input-inner">
        <label for="chatInput" class="visually-hidden">提问输入框</label>
        <input
          id="chatInput"
          v-model="draft"
          type="text"
          class="input"
          aria-label="提问输入框"
          placeholder="描述故障现象，如：某服务最近一小时错误率上升…"
          maxlength="500"
          :disabled="busy"
        />
        <button type="submit" class="btn chat-send" :disabled="busy || !draft.trim()">
          发送
        </button>
      </div>
      <div class="chat-hint">
        <span :class="{ 'chat-hint--full': draft.length >= 500 }">{{ draft.length }}/500</span>
        · 当前模式：<span>{{ modeLabel }}</span>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { CitationSourceType, Diagnosis } from '@/api/opsagent'
import { degradeBadge } from '../logic'
import { sourceTypeMeta } from '@/views/evidence'
import ThinkingBlock from './ThinkingBlock.vue'

/** 对话流单条消息 */
export interface ChatMessage {
    role: 'user' | 'assistant'
    text: string
}

const props = defineProps<{
    messages: ChatMessage[]
    diagnosis: Diagnosis | null
    /** degradeLevel（0 正常 / 1 模板 / 2 预置入口 / 3 明示降级） */
    degradeLevel: number
    busy: boolean
    /** 当前用户名（用户行署名，回退 havens 与原型一致） */
    username?: string
}>()

const emit = defineEmits<{ (e: 'submit', message: string): void }>()

const SUGGESTIONS = [
    { label: 'order-service 错误率上升', query: 'order-service 最近一小时错误率上升，帮我看看' },
    { label: 'CDN 是否被刷', query: '最近 CDN 有没有被刷？' },
    { label: 'CDN 流量突增', query: 'cdn-edge 服务最近 30 分钟流量突增' },
] as const

const draft = ref('')
const citationsOpen = ref(false)

const userName = computed(() => props.username || 'havens')
const avatarLetter = computed(() => (userName.value[0] || 'H').toUpperCase())
const confidencePct = computed(() => Math.round((props.diagnosis?.confidence ?? 0) * 100))
const degradeBadgeMeta = computed(() => degradeBadge(props.degradeLevel))
const modeLabel = computed(() => (props.degradeLevel === 0 ? '正常' : '降级'))

function sourceLabel(sourceType: CitationSourceType): string {
    return sourceTypeMeta(sourceType).label
}
function sourceColor(sourceType: CitationSourceType): string {
    return sourceTypeMeta(sourceType).hsl
}

function fillSuggestion(query: string): void {
    if (props.busy) return
    draft.value = query
}

function onSubmit(): void {
    const text = draft.value.trim()
    if (!text || props.busy) return
    emit('submit', text)
    draft.value = ''
}
</script>

<style scoped>
.visually-hidden {
  position: absolute;
  left: -9999px;
}
.chat-panel {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 56px);
}
.chat-scroll {
  display: flex;
  flex-direction: column;
  gap: 24px;
  flex: 1;
}
.empty-state {
  text-align: center;
  padding: 48px 24px;
  color: hsl(var(--muted-foreground));
}
.empty-state .icon {
  font-size: 40px;
  margin-bottom: 12px;
}
.empty-state h3 {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: hsl(var(--foreground));
}
.empty-state p {
  margin: 0 0 16px;
  font-size: 13px;
}
.suggestion-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}
.chip {
  border: 1px solid hsl(var(--border));
  border-radius: 9999px;
  padding: 7px 16px;
  font-size: 13px;
  cursor: pointer;
  background: transparent;
  color: hsl(var(--foreground));
  font-family: inherit;
  transition:
    background 0.15s,
    border-color 0.15s;
}
.chip:hover {
  background: hsl(var(--accent));
  border-color: hsl(var(--primary) / 0.4);
}
.chat-row {
  display: flex;
  gap: 14px;
}
.who-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  flex-shrink: 0;
  color: hsl(var(--primary-foreground));
}
.chat-row.user .who-icon {
  border-radius: 50%;
  background: linear-gradient(135deg, hsl(var(--primary)), hsl(271 91% 65%));
  font-size: 12px;
}
.chat-row.agent .who-icon {
  background: hsl(var(--primary) / 0.18);
  border: 1px solid hsl(var(--primary) / 0.35);
}
.bubble-wrap {
  flex: 1;
  min-width: 0;
}
.bubble-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.bubble-head .name {
  font-size: 13px;
  font-weight: 600;
}
.bubble {
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  border-radius: 16px 16px 4px 16px;
  padding: 12px 16px;
  font-size: 13px;
  display: inline-block;
  max-width: 100%;
  white-space: pre-wrap;
  word-break: break-word;
}
.bubble--agent {
  background: hsl(var(--accent));
  color: hsl(var(--foreground));
  border-radius: 16px 16px 16px 4px;
}
.agent-chip {
  font-size: 11px;
  padding: 2px 10px;
  border-radius: 9999px;
  font-weight: 500;
}
.agent-chip-coord {
  background: hsl(var(--agent-coord) / 0.15);
  color: hsl(var(--agent-coord));
  border: 1px solid hsl(var(--agent-coord) / 0.3);
}
.report-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
  padding: 16px;
  color: hsl(var(--card-foreground));
}
.report-card h3 {
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 6px;
}
.report-card__badges {
  display: flex;
  gap: 6px;
  margin: 6px 0;
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
.badge-warning {
  background: hsl(var(--severity-low) / 0.15);
  color: hsl(var(--severity-low));
  border: 1px solid hsl(var(--severity-low) / 0.3);
}
.badge-muted {
  background: hsl(var(--muted));
  color: hsl(var(--muted-foreground));
}
.report-card__advice {
  font-size: 13px;
  margin-bottom: 10px;
}
.report-card__advice ul {
  margin: 6px 0 0 18px;
}
.report-card__advice li {
  margin-bottom: 4px;
}
.report-card__truncation {
  margin: 10px 0;
}
.collapsible-header {
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 500;
  padding: 6px 0;
}
.collapsible-content {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.chevron {
  transition: transform 0.15s;
  font-size: 11px;
  color: hsl(var(--muted-foreground));
}
.evidence-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 12px 14px;
  margin: 0;
}
.evidence-card-header {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: hsl(var(--foreground));
}
.evidence-card-header .dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 4px;
}
.evidence-empty {
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  font-style: italic;
  margin: 0;
}
.report-card__footer {
  margin-top: 12px;
  text-align: right;
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
.chat-input-bar {
  position: sticky;
  bottom: 0;
  background: hsl(var(--background));
  border-top: 1px solid hsl(var(--border));
  padding: 16px 28px;
  margin: 32px -28px -28px;
  z-index: 40;
}
.chat-input-inner {
  display: flex;
  gap: 10px;
  align-items: center;
}
.chat-input-inner .input {
  flex: 1;
  height: 44px;
}
.input {
  height: 36px;
  padding: 0 12px;
  border-radius: 8px;
  border: 1px solid hsl(var(--input));
  background: hsl(var(--background) / 0.5);
  font-size: 13px;
  width: 100%;
  color: hsl(var(--foreground));
  font-family: inherit;
}
.input::placeholder {
  color: hsl(var(--muted-foreground) / 0.7);
}
.input:focus {
  border-color: hsl(var(--ring));
  box-shadow: 0 0 0 3px hsl(var(--ring) / 0.15);
  outline: none;
}
.chat-send {
  height: 44px;
  flex-shrink: 0;
}
.chat-send:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.chat-hint {
  text-align: center;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  margin-top: 8px;
}
.chat-hint--full {
  color: hsl(var(--severity-high));
}
</style>