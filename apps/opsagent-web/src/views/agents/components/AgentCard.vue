<template>
  <!-- Agent 卡片（原型 agent-management.html：emoji 图标 + 英文名 + 角色行 + 6 行指标 + 查看详情） -->
  <div class="card stat-card">
    <div class="agent-card-head">
      <div class="agent-card-head-left">
        <div class="agent-icon" :style="{ background: `hsl(var(${meta.gradToken}))` }">
          {{ meta.emoji }}
        </div>
        <div class="agent-card-title">
          <h3>{{ meta.display }}</h3>
          <p>{{ meta.roleLine }}</p>
        </div>
      </div>
      <span class="badge" :class="status.alarm ? 'badge-low' : 'badge-ok'">
        <span v-if="!status.alarm" class="pulse-dot" />
        {{ status.alarm ? '⚠️ ' : '' }}{{ status.label }}
      </span>
    </div>

    <ul class="agent-metrics">
      <li><span class="m-label">活跃任务</span><span class="m-value">—</span></li>
      <li><span class="m-label">完成任务</span><span class="m-value">{{ agent.tasksTotal }}</span></li>
      <li><span class="m-label">平均延迟</span><span class="m-value">{{ agent.avgDurationMs }}ms</span></li>
      <li>
        <span class="m-label">成功率</span>
        <span class="m-value" :class="{ 'm-value--ok': success !== '—' }">{{ success }}</span>
      </li>
      <li><span class="m-label">正常运行时长</span><span class="m-value">—</span></li>
      <li><span class="m-label">版本</span><span class="m-value m-value--mono">—</span></li>
    </ul>

    <div class="agent-card-actions">
      <button type="button" class="btn btn-secondary btn-sm" style="flex: 1" @click="dialogOpen = true">
        查看详情
      </button>
      <!-- 停止/重启无 API（Hard Rule 无控制面）→ 省略，不渲染假按钮 -->
    </div>

    <el-dialog v-model="dialogOpen" :title="meta.display" width="420px" append-to-body>
      <p class="dialog-role">{{ meta.roleLine }}</p>
      <pre class="code-block">{{ infoLines }}</pre>
      <template #footer>
        <button type="button" class="btn" @click="dialogOpen = false">关闭</button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { AgentStat } from '@/api/opsagent'
import { agentPresent, agentStatusMeta, agentSuccessRate } from '../logic'

const props = defineProps<{ agent: AgentStat }>()

const dialogOpen = ref(false)

const meta = computed(() => agentPresent(props.agent.name))
const status = computed(() => agentStatusMeta(props.agent.status))
const success = computed(() => agentSuccessRate(props.agent.tasksTotal, props.agent.tasksFailed))

/** 详情 code-block（仅真实字段；未知字段以 — 呈现） */
const infoLines = computed(
    () =>
        `完成 ${props.agent.tasksTotal} · 失败 ${props.agent.tasksFailed}\n` +
        `平均延迟 ${props.agent.avgDurationMs}ms · 成功率 ${success.value}\n` +
        `正常运行时长 — · 版本 —\n` +
        `状态 ${status.value.label}`,
)
</script>

<style scoped>
.card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    padding: 20px;
    color: hsl(var(--card-foreground));
    display: flex;
    flex-direction: column;
}
.agent-card-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 14px;
}
.agent-card-head-left {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
}
.agent-icon {
    width: 44px;
    height: 44px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    flex-shrink: 0;
    filter: saturate(1.1);
}
.agent-card-title h3 {
    margin: 0 0 2px;
    font-size: 15px;
    font-weight: 600;
}
.agent-card-title p {
    margin: 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    border-radius: 9999px;
    padding: 3px 12px;
    font-size: 12px;
    font-weight: 500;
    white-space: nowrap;
}
.badge-ok {
    background: hsl(var(--severity-ok) / 0.15);
    color: hsl(var(--severity-ok));
    border: 1px solid hsl(var(--severity-ok) / 0.3);
}
.badge-low {
    background: hsl(var(--severity-low) / 0.15);
    color: hsl(var(--severity-low));
    border: 1px solid hsl(var(--severity-low) / 0.3);
}
.pulse-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: hsl(var(--severity-ok));
    box-shadow: 0 0 0 3px hsl(var(--severity-ok) / 0.25);
}
.agent-metrics {
    list-style: none;
    margin: 0 0 14px;
    padding: 0;
    font-size: 13px;
    line-height: 2;
}
.agent-metrics li {
    display: flex;
    justify-content: space-between;
    gap: 12px;
}
.m-label {
    color: hsl(var(--muted-foreground));
}
.m-value {
    font-variant-numeric: tabular-nums;
    color: hsl(var(--foreground));
}
.m-value--ok {
    color: hsl(var(--severity-ok));
}
.m-value--mono {
    font-family: monospace;
    font-size: 12px;
}
.agent-card-actions {
    display: flex;
    gap: 8px;
    margin-top: auto;
    padding-top: 14px;
    border-top: 1px solid hsl(var(--border));
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
}
.btn-sm {
    height: 30px;
    padding: 0 12px;
    font-size: 12px;
}
.btn-secondary {
    background: hsl(var(--muted) / 0.5);
    color: hsl(var(--foreground));
    border-color: hsl(var(--border));
}
.dialog-role {
    margin: 0 0 12px;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
}
.code-block {
    margin: 0;
    background: hsl(var(--muted) / 0.45);
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    padding: 12px 14px;
    font-family: monospace;
    font-size: 12px;
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-all;
    color: hsl(var(--foreground));
}
</style>