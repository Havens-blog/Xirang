<template>
  <div class="opsagent-page">
    <!-- 页头（原型 agent-management.html：Multi-Agent 系统 + 副标题） -->
    <div class="page-header">
      <h1>Multi-Agent 系统</h1>
      <p class="subtitle">编排层 + 4 个子智能体：任务协调、日志分析、告警监测、诊断巡检</p>
    </div>

    <div v-if="error" class="agents__error">{{ error }}</div>
    <el-skeleton v-else-if="loading" :rows="10" animated />
    <template v-else-if="data">
      <!-- Agent 卡片网格（原型 auto-fill minmax 280） -->
      <div class="agents-grid">
        <AgentCard v-for="a in data.agents" :key="a.name" :agent="a" />
      </div>

      <!-- 负载历史（全宽，原型 grid-column: span 2） -->
      <div class="agents-bottom">
        <LoadChart class="agents-load-wide" :load-history="data.loadHistory" />
        <TaskQueue :queue="data.queue" :totals="totals" />
        <PerfMetrics :agents="data.agents" :budget="data.llmBudget" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { getObservabilityApi, type ObservabilityData } from '@/api/opsagent'
import { computed, onMounted, ref } from 'vue'
import AgentCard from './components/AgentCard.vue'
import LoadChart from './components/LoadChart.vue'
import PerfMetrics from './components/PerfMetrics.vue'
import TaskQueue from './components/TaskQueue.vue'
import { agentTotals } from './logic'

const data = ref<ObservabilityData | null>(null)
const loading = ref(true)
const error = ref('')

/** 任务队列「已完成/失败」= 跨 Agent 合计（原型四格口径） */
const totals = computed(() => agentTotals(data.value?.agents ?? []))

async function load(): Promise<void> {
    loading.value = true
    error.value = ''
    try {
        data.value = await getObservabilityApi()
    } catch (err) {
        error.value = err instanceof Error ? err.message : '加载失败'
    } finally {
        loading.value = false
    }
}

onMounted(load)
</script>

<style scoped>
.page-header {
    margin-bottom: 20px;
}
.page-header h1 {
    margin: 0 0 4px;
    font-size: 30px;
    font-weight: 700;
    letter-spacing: -0.025em;
}
.page-header .subtitle {
    margin: 0;
    font-size: 14px;
    color: hsl(var(--muted-foreground));
}
.agents__error {
    padding: 24px;
    text-align: center;
    color: hsl(var(--destructive));
}
.agents-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
    margin-bottom: 24px;
}
.agents-bottom {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
}
.agents-load-wide {
    grid-column: span 2;
}
</style>