<template>
  <!-- 根因结论（原型 diagnosis-detail.html：card-rca 主视觉 + 卡内折叠数据源引用） -->
  <div class="card card-rca">
    <p class="rca-label">🔎 根因结论</p>
    <h2 class="rca-cause">
      {{ diagnosis.rootCause || '（无明确根因结论）' }}
      <span class="rca-confidence">（置信度 {{ formatConfidence(diagnosis.confidence) }}）</span>
    </h2>

    <div
      class="collapsible-header"
      :aria-expanded="citationsOpen"
      @click="citationsOpen = !citationsOpen"
    >
      <span class="rca-citations-label">数据源引用（{{ diagnosis.citations.length }}）</span>
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
          {{ sourceLabel(c.sourceType) }}
        </div>
        <p class="evidence-card-body">{{ c.snippet }}</p>
      </div>
      <p v-if="diagnosis.citations.length === 0" class="evidence-empty">无数据源引用</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { CitationSourceType, Diagnosis } from '@/api/opsagent'
import { sourceTypeMeta } from '@/views/evidence'
import { formatConfidence } from '../logic'

defineProps<{ diagnosis: Diagnosis }>()

const citationsOpen = ref(false)

function sourceLabel(sourceType: CitationSourceType): string {
    return sourceTypeMeta(sourceType).label
}
function sourceColor(sourceType: CitationSourceType): string {
    return sourceTypeMeta(sourceType).hsl
}
</script>

<style scoped>
.card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
  padding: 24px;
  color: hsl(var(--card-foreground));
}
.card-rca {
  border-left: 3px solid hsl(var(--primary));
}
.rca-label {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  margin: 0 0 6px;
}
.rca-cause {
  font-size: 20px;
  font-weight: 700;
  margin: 0 0 12px;
  color: hsl(var(--foreground));
}
.rca-confidence {
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
.rca-citations-label {
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
}
.evidence-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  font-weight: 600;
}
.evidence-card-header .dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}
.evidence-card-body {
  margin: 0;
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
</style>