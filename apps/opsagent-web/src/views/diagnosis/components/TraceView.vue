<template>
  <!-- 编排调用链（原型 diagnosis-detail.html：可折叠卡 + code-block 文本链，默认折叠） -->
  <div class="card">
    <div
      class="collapsible-header"
      :aria-expanded="open"
      @click="open = !open"
    >
      <span>{{ heading }}</span>
      <span class="chevron">▾</span>
    </div>
    <div v-if="open" class="collapsible-content">
      <pre class="code-block">{{ chainText }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { TraceStep } from '@/api/opsagent'
import { agentLabel } from '../logic'

const props = withDefaults(
    defineProps<{
        trace: TraceStep[]
        /** 折叠头标题（诊断页带「供 SRE 二次核实」副题；历史页为原型短标题） */
        heading?: string
    }>(),
    { heading: '🧬 编排调用链（供 SRE 二次核实）' },
)

const open = ref(false)

/** 文本调用链：→ 角色.action(source) → 摘要 (耗时ms)，降级步附 L 级 */
const chainText = computed(() => {
    if (props.trace.length === 0) return '无编排调用链记录'
    return props.trace
        .map((t) => {
            const source = t.source ? `(${t.source})` : ''
            const degrade = t.degradeLevel && t.degradeLevel > 0 ? ` [降级 L${t.degradeLevel}]` : ''
            return `→ ${agentLabel(t.agent)}.${t.action}${source} → ${t.summary}${degrade} (${t.durationMs}ms)`
        })
        .join('\n')
})
</script>

<style scoped>
.card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
  padding: 20px;
  color: hsl(var(--card-foreground));
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
}
.code-block {
  background: hsl(var(--muted) / 0.4);
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 14px;
  overflow-x: auto;
  font-family: 'JetBrains Mono', 'Cascadia Mono', Consolas, monospace;
  font-size: 12.5px;
  margin: 0;
  color: hsl(var(--foreground));
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>