<template>
  <!-- 任务队列（原型 agent-management.html 四格：待处理/处理中/已完成/失败） -->
  <div class="card task-queue">
    <div class="card-header"><h3>任务队列</h3></div>
    <div class="task-queue-grid">
      <div class="tile">
        <div class="tile-value" style="color: hsl(var(--severity-high))">{{ queue.depth }}</div>
        <div class="tile-label">待处理</div>
      </div>
      <div class="tile">
        <div class="tile-value" style="color: hsl(var(--primary))">{{ queue.inflight }}</div>
        <div class="tile-label">处理中</div>
      </div>
      <div class="tile">
        <div class="tile-value" style="color: hsl(var(--severity-ok))">{{ totals.done }}</div>
        <div class="tile-label">已完成</div>
      </div>
      <div class="tile">
        <div class="tile-value" style="color: hsl(var(--severity-low))">{{ totals.failed }}</div>
        <div class="tile-label">失败</div>
      </div>
    </div>
    <!-- 容量/并发/溢出（数据诚实：QueueStat 真实字段，非原型虚构的排队条目） -->
    <p class="task-queue-meta">
      容量 {{ queue.capacity }} · 并发 {{ queue.concurrency }} · 溢出总量 {{ queue.overflowTotal }}
    </p>
  </div>
</template>

<script setup lang="ts">
import type { QueueStat } from '@/api/opsagent'
import type { AgentTotals } from '../logic'

defineProps<{
    queue: QueueStat
    totals: AgentTotals
}>()
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
.task-queue-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 12px;
    margin-bottom: 14px;
}
.tile {
    text-align: center;
}
.tile-value {
    font-size: 26px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
}
.tile-label {
    font-size: 12px;
    color: hsl(var(--muted-foreground));
    margin-top: 2px;
}
.task-queue-meta {
    margin: 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
</style>