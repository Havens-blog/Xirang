<template>
  <!-- 处置建议（原型 diagnosis-detail.html：h3 + 逐条 list，首条按风险档带执行语义徽标） -->
  <div class="card">
    <h3 class="disposition-title">处置建议</h3>
    <p v-if="items.length === 0" class="disposition-empty">无处置建议</p>
    <ul v-else class="disposition-list">
      <li v-for="(d, i) in items" :key="i">
        {{ d.text }}
        <span v-if="d.badge" class="badge" :class="`badge--${d.badge.tone}`">{{ d.badge.text }}</span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { DispositionStep } from '@/api/opsagent'
import { dispositionBadge } from '../logic'

const props = defineProps<{
    disposition: DispositionStep[]
    /** disposition 为空时的结论回退（不带头徽标） */
    fallback?: string[]
}>()

interface Item {
    text: string
    badge: ReturnType<typeof dispositionBadge> | null
}

const items = computed<Item[]>(() => {
    if (props.disposition.length) {
        return props.disposition.map((d) => ({
            text: d.action || d.step,
            badge: dispositionBadge(d.risk),
        }))
    }
    return (props.fallback ?? []).map((text) => ({ text, badge: null }))
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
.disposition-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 12px;
}
.disposition-empty {
  margin: 0;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  font-style: italic;
}
.disposition-list {
  margin: 0 0 0 18px;
  font-size: 13px;
  line-height: 2;
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
  line-height: 1.4;
  vertical-align: middle;
  margin-left: 6px;
}
.badge--ok {
  background: hsl(var(--severity-ok) / 0.15);
  color: hsl(var(--severity-ok));
  border: 1px solid hsl(var(--severity-ok) / 0.3);
}
.badge--ghost {
  border: 1px solid hsl(var(--border));
  color: hsl(var(--muted-foreground));
}
.badge--high {
  background: hsl(var(--severity-high) / 0.15);
  color: hsl(var(--severity-high));
  border: 1px solid hsl(var(--severity-high) / 0.3);
}
</style>