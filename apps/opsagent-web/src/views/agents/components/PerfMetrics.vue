<template>
  <!-- Agent 性能指标（原型 agent-management.html 四格；无「资源使用率」契约 → 以 LLM 预算替代，诚实呈现） -->
  <div class="card perf-metrics">
    <div class="card-header"><h3>Agent 性能指标</h3></div>
    <div class="perf-grid">
      <div class="perf-tile">
        <div class="perf-tile-label">今日总任务数</div>
        <div class="perf-tile-value">{{ sumTasksTotal(agents) }}</div>
      </div>
      <div class="perf-tile">
        <div class="perf-tile-label">平均响应时间</div>
        <div class="perf-tile-value">{{ weightedAvgDuration(agents) }}</div>
      </div>
      <div class="perf-tile">
        <div class="perf-tile-label">错误率</div>
        <div class="perf-tile-value" :class="{ 'perf-tile-value--high': errorRate !== '—' }">
          {{ errorRate }}
        </div>
      </div>
      <div class="perf-tile">
        <div class="perf-tile-label">LLM 预算用量</div>
        <div
          class="perf-tile-value"
          :class="{ 'perf-tile-value--high': budget.exceeded }"
          :title="`${budget.calls} / ${budget.budgetLimit} 次`"
        >
          {{ budgetUsagePercent(budget.calls, budget.budgetLimit) }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AgentStat, LLMBudgetStat } from '@/api/opsagent'
import { budgetUsagePercent, overallErrorRate, sumTasksTotal, weightedAvgDuration } from '../logic'

const props = defineProps<{
    agents: AgentStat[]
    budget: LLMBudgetStat
}>()

const errorRate = computed(() => overallErrorRate(props.agents))
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
.perf-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
}
.perf-tile {
    padding: 12px;
    border-radius: 8px;
    background: hsl(var(--muted) / 0.4);
}
.perf-tile-label {
    font-size: 12px;
    color: hsl(var(--muted-foreground));
    margin-bottom: 4px;
}
.perf-tile-value {
    font-size: 22px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
}
.perf-tile-value--high {
    color: hsl(var(--severity-high));
}
</style>