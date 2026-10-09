<template>
  <div class="datasource-tab">
    <p class="datasource-tab-note">
      连接状态由底座实时探针上报（只读）；展开查看契约字段与最近一次探测结果。
    </p>
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
        <p class="ds-panel-title">契约字段</p>
        <ul>
          <li v-for="f in card.meta.fields" :key="f">· {{ f }}</li>
        </ul>

        <!-- 探针实测（后端 GET /settings 的 datasources 只读上报） -->
        <template v-if="card.probes.length > 0">
          <p class="ds-panel-title">探针实测</p>
          <ul class="ds-probes">
            <li v-for="p in card.probes" :key="p.name" class="ds-probe">
              <span class="status-dot" :class="p.ok ? 'connected' : 'disconnected'" />
              <span class="ds-probe-name">{{ dataSourceLabel(p.name) }}</span>
              <span class="ds-probe-meta">{{ p.cloud || '—' }} · {{ connectionStatusText(p.ok) }} · {{ formatClock(p.checkedAt) }}</span>
            </li>
          </ul>
        </template>
        <p v-else class="ds-panel-title ds-panel-empty">底座未上报该数据源探针</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Datasource } from '@/api/opsagent'
import { connectionStatusText, DATASOURCE_CARDS, dataSourceLabel } from '../logic'
import { formatClock } from '../../format'

const props = defineProps<{ datasources: Datasource[] }>()

const open = ref<string | null>(null)

/** 注册表卡 + 后端 probe 推导的连接态与实测明细 */
const cards = computed(() =>
    DATASOURCE_CARDS.map((meta) => {
        const probes = meta.probeKeys
            .map((k) => props.datasources.find((d) => d.name === k))
            .filter((p): p is Datasource => p !== undefined)
        const connected = probes.length > 0 && probes.every((p) => p.ok)
        return { meta, probes, connected }
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
.datasource-tab-note {
    margin: 0 0 14px;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
    line-height: 1.6;
}
.ds-panel-title {
    margin: 14px 0 4px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    color: hsl(var(--muted-foreground));
}
.ds-panel-title:first-child {
    margin-top: 0;
}
.ds-probes {
    display: flex;
    flex-direction: column;
    gap: 6px;
    line-height: 1.6;
}
.ds-probe {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
}
.ds-probe-name {
    font-weight: 500;
    color: hsl(var(--foreground));
}
.ds-probe-meta {
    color: hsl(var(--muted-foreground));
}
.ds-panel-empty {
    font-weight: 400;
    text-transform: none;
    letter-spacing: 0;
}
</style>