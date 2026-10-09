<template>
  <!-- 统计卡（原型 risk-center.html：高危待确认红 / 今日诊断白 / 待查看橙） -->
  <div class="stat-cards">
    <div class="card stat-card">
      <p class="stat-card__label">高危待确认</p>
      <p class="stat-card__value stat-card__value--high">{{ stats.highRisk }}</p>
    </div>
    <div class="card stat-card">
      <p class="stat-card__label">今日诊断</p>
      <p class="stat-card__value">{{ stats.todayNew }}</p>
    </div>
    <div class="card stat-card">
      <p class="stat-card__label">待查看</p>
      <p class="stat-card__value stat-card__value--low">{{ stats.pendingView }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { RiskStats } from '@/api/opsagent'

defineProps<{ stats: RiskStats }>()
</script>

<style scoped>
.stat-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 16px;
    margin-bottom: 20px;
}
.card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    padding: 20px;
    color: hsl(var(--card-foreground));
}
.stat-card__label {
    margin: 0 0 8px;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.stat-card__value {
    margin: 0;
    font-size: 28px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
}
.stat-card__value--high {
    color: hsl(var(--severity-high));
}
.stat-card__value--low {
    color: hsl(var(--severity-low));
}
</style>