<template>
  <div class="datasource-tab">
    <!-- 原型 settings.html：五张 ds-card（前端注册表），连接态由后端 probes 推导 -->
    <div v-for="card in cards" :key="card.meta.key" class="ds-card">
      <div
        class="ds-card-head"
        role="button"
        :aria-expanded="open === card.meta.key"
        @click="open = open === card.meta.key ? null : card.meta.key"
      >
        <div class="ds-card-head-main">
          <div class="ds-icon" :style="{ background: `hsl(var(${card.meta.gradToken}))` }">
            {{ card.meta.icon }}
          </div>
          <div class="ds-meta">
            <h3>{{ card.meta.name }}</h3>
            <p>{{ card.meta.desc }}</p>
          </div>
        </div>
        <div class="ds-card-head-side">
          <span class="ds-status">
            <span class="status-dot" :class="card.connected ? 'connected' : 'disconnected'" />
            {{ card.connected ? '已连接' : '未连接' }}
          </span>
          <span class="chevron">▾</span>
        </div>
      </div>
      <div v-show="open === card.meta.key" class="ds-panel">
        <ul>
          <li v-for="f in card.meta.fields" :key="f">· {{ f }}</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Datasource } from '@/api/opsagent'
import { DATASOURCE_CARDS } from '../logic'

const props = defineProps<{ datasources: Datasource[] }>()

const open = ref<string | null>(null)

/** 注册表卡 + 后端 probe 推导的连接态（无 probe 的卡恒未连接，不虚构在线） */
const cards = computed(() =>
    DATASOURCE_CARDS.map((meta) => {
        const probes = meta.probeKeys.map((k) => props.datasources.find((d) => d.name === k))
        const connected = probes.length > 0 && probes.every((p) => p?.ok === true)
        return { meta, connected }
    }),
)
</script>

<style scoped>
.ds-card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    margin-bottom: 12px;
    overflow: hidden;
}
.ds-card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 16px;
    cursor: pointer;
}
.ds-card-head:hover {
    background: hsl(var(--accent) / 0.5);
}
.ds-card-head-main {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
}
.ds-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.02em;
    color: hsl(var(--card) / 1);
    filter: saturate(1.1);
}
.ds-meta h3 {
    margin: 0 0 4px;
    font-size: 14px;
    font-weight: 600;
}
.ds-meta p {
    margin: 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
    line-height: 1.5;
}
.ds-card-head-side {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
}
.ds-status {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
    white-space: nowrap;
}
.status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
}
.status-dot.connected {
    background: hsl(var(--severity-ok));
    box-shadow: 0 0 0 3px hsl(var(--severity-ok) / 0.2);
}
.status-dot.disconnected {
    background: hsl(var(--muted-foreground) / 0.5);
}
.chevron {
    transition: transform 0.15s;
    font-size: 11px;
    color: hsl(var(--muted-foreground));
}
.ds-card-head[aria-expanded='true'] .chevron {
    transform: rotate(180deg);
}
.ds-panel {
    border-top: 1px solid hsl(var(--border));
    padding: 12px 16px 12px 68px;
}
.ds-panel ul {
    list-style: none;
    margin: 0;
    padding: 0;
    font-size: 13px;
    line-height: 2;
    color: hsl(var(--muted-foreground));
}
</style>