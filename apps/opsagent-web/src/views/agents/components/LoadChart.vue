<template>
  <div class="card load-chart">
    <div class="card-header">
      <h3>Agent 负载历史（24 小时）</h3>
      <!-- 时间范围下拉：后端观测仅返回自身窗口段（LoadPoint 无区间参数）→ 省略，不渲染无效控件 -->
    </div>

    <p v-if="loadHistory.length < 2" class="load-chart-empty">暂无足够负载记录，无法绘制曲线</p>
    <template v-else>
      <svg viewBox="0 0 800 180" class="load-chart-svg" role="img" aria-label="Agent 负载历史">
        <line v-for="y in [30, 70, 110, 150]" :key="y" x1="40" :y1="y" x2="780" :y2="y" stroke="hsl(var(--border))" stroke-width="1" />
        <!-- 数据诚实：depth/inflight 双折线（LoadPoint 契约仅两序列） -->
        <polyline :points="depthPoints" fill="none" stroke="hsl(var(--agent-log))" stroke-width="2" />
        <polyline :points="inflightPoints" fill="none" stroke="hsl(var(--agent-monitor))" stroke-width="2" />
      </svg>
      <div class="load-chart-legend">
        <span class="legend-item"><span class="legend-swatch" style="background: hsl(var(--agent-log))" />队列深度</span>
        <span class="legend-item"><span class="legend-swatch" style="background: hsl(var(--agent-monitor))" />在途任务</span>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { LoadPoint } from '@/api/opsagent'

const props = defineProps<{ loadHistory: LoadPoint[] }>()

const CHART_TOP = 30
const CHART_BOTTOM = 150
const X0 = 40
const X1 = 780

/** 序列 → SVG polyline points（横向等距铺开，纵向按 0..max 归一） */
function toPoints(values: number[]): string {
    if (values.length === 0) return ''
    const max = Math.max(...values, 1)
    const span = X1 - X0
    return values
        .map((v, i) => {
            const x = values.length === 1 ? X0 : X0 + (span * i) / (values.length - 1)
            const y = CHART_BOTTOM - ((CHART_BOTTOM - CHART_TOP) * v) / max
            return `${Math.round(x)},${Math.round(y)}`
        })
        .join(' ')
}

const depthPoints = computed(() => toPoints(props.loadHistory.map((p) => p.depth)))
const inflightPoints = computed(() => toPoints(props.loadHistory.map((p) => p.inflight)))
</script>

<style scoped>
.card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    padding: 20px;
    color: hsl(var(--card-foreground));
}
.card-header {
    margin-bottom: 14px;
}
.card-header h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
}
.load-chart-empty {
    margin: 0;
    padding: 20px 0;
    text-align: center;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
    font-style: italic;
}
.load-chart-svg {
    width: 100%;
    height: 180px;
    display: block;
}
.load-chart-legend {
    display: flex;
    justify-content: center;
    gap: 16px;
    margin-top: 8px;
    font-size: 12px;
}
.legend-item {
    display: flex;
    align-items: center;
    gap: 5px;
    color: hsl(var(--muted-foreground));
}
.legend-swatch {
    width: 10px;
    height: 10px;
    border-radius: 2px;
    display: inline-block;
}
</style>