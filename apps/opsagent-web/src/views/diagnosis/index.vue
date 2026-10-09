<template>
  <div class="opsagent-page">
    <!-- 面包屑（原型：风险中心 / 诊断详情） -->
    <div class="breadcrumb">
      <RouterLink to="/risk-center">风险中心</RouterLink>
      <span class="sep">/</span>
      <span>诊断详情</span>
    </div>

    <div v-if="error" class="diagnosis__error">{{ error }}</div>
    <el-skeleton v-else-if="loading" :rows="8" animated />
    <div v-else-if="diagnosis" class="diagnosis">
      <!-- 标题行（原型：h1 服务名 + 风险档/severity/时间 + 降级/截断说明） -->
      <div class="diagnosis__title-row">
        <div>
          <div class="diagnosis__title-line">
            <h1 class="diagnosis__service">{{ serviceName }}</h1>
            <span class="badge" :class="`badge--${riskTone}`">{{ riskLabel }}</span>
            <span class="badge badge-ghost diagnosis__severity">{{ diagnosis.severity }}</span>
            <span v-if="createdTime" class="diagnosis__time">{{ createdTime }}</span>
          </div>
          <div class="diagnosis__meta-line">
            <span v-if="diagnosis.degraded" class="badge badge-muted">降级：LLM 预算耗尽，纯规则结论</span>
            <span v-if="diagnosis.truncated" class="badge badge-muted">结果已截断，请缩小时间窗或服务范围</span>
          </div>
        </div>
      </div>

      <!-- 单栏主视觉（原型同构：根因 → 处置建议 → 编排调用链） -->
      <RootCauseCard :diagnosis="diagnosis" />
      <DispositionList
        :disposition="diagnosis.disposition"
        :fallback="diagnosis.conclusions.map((c) => c.text)"
      />
      <TraceView :trace="diagnosis.trace" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { getDiagnosisApi, historyApi, type Diagnosis } from '@/api/opsagent'
import { ElMessage } from 'element-plus'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { RISK_LEVEL_META } from './logic'
import DispositionList from './components/DispositionList.vue'
import RootCauseCard from './components/RootCauseCard.vue'
import TraceView from './components/TraceView.vue'

const route = useRoute()

const diagnosis = ref<Diagnosis | null>(null)
const loading = ref(true)
const error = ref('')
/**
 * 服务名/时间：Diagnosis 契约不含 title 元数据，经精确会话回查补齐
 * （GET /history?sessionId= 已支持，api-handbook §7 精确路径）。
 */
const serviceName = ref('未知服务')
const createdTime = ref('')

const riskLabel = computed(() => (diagnosis.value ? RISK_LEVEL_META[diagnosis.value.riskLevel].label : ''))
const riskTone = computed(() => (diagnosis.value ? RISK_LEVEL_META[diagnosis.value.riskLevel].tone : 'info'))

/**
 * 打开详情即触发自动已读：GET 携带 riskEntryId，后端惰性 CAS 完成
 * pending_view→viewed（api-handbook §6）；本页展示诊断、风险状态由风险中心
 * 返回时重新加载体现。
 */
async function load(): Promise<void> {
    loading.value = true
    error.value = ''
    try {
        const id = String(route.params.id)
        const riskEntryId = route.query.riskEntryId ? String(route.query.riskEntryId) : undefined
        diagnosis.value = await getDiagnosisApi(id, { riskEntryId })
        await loadSessionMeta()
    } catch (err) {
        error.value = err instanceof Error ? err.message : '加载失败'
        ElMessage.error(error.value)
    } finally {
        loading.value = false
    }
}

/** 精确会话回查：补齐服务名与发起时间（失败静默，标题行退化仍可用） */
async function loadSessionMeta(): Promise<void> {
    try {
        const sessionId = diagnosis.value?.sessionId
        if (!sessionId) return
        const history = await historyApi({ sessionId })
        const item = history.items[0]
        if (item) {
            serviceName.value = item.serviceName || '未知服务'
            const at = new Date(item.createdAt)
            if (!Number.isNaN(at.getTime())) {
                createdTime.value = `${String(at.getHours()).padStart(2, '0')}:${String(at.getMinutes()).padStart(2, '0')}`
            }
        }
    } catch {
        // 静默：标题行元数据属增强信息，不影响诊断正文
    }
}

onMounted(load)
</script>

<style scoped>
.diagnosis__error {
  padding: 24px;
  text-align: center;
  color: hsl(var(--destructive));
}
.breadcrumb {
  display: flex;
  gap: 6px;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  margin-bottom: 16px;
}
.breadcrumb a {
  color: hsl(var(--primary));
  text-decoration: none;
}
.diagnosis {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.diagnosis__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}
.diagnosis__title-line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}
.diagnosis__service {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
  letter-spacing: -0.02em;
}
.diagnosis__severity {
  font-family: monospace;
}
.diagnosis__time {
  font-family: monospace;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}
.diagnosis__meta-line {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
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
}
.badge--danger {
  background: hsl(var(--severity-high) / 0.15);
  color: hsl(var(--severity-high));
  border: 1px solid hsl(var(--severity-high) / 0.3);
}
.badge--warning {
  background: hsl(var(--severity-low) / 0.15);
  color: hsl(var(--severity-low));
  border: 1px solid hsl(var(--severity-low) / 0.3);
}
.badge--info {
  background: hsl(var(--primary) / 0.15);
  color: hsl(var(--primary));
  border: 1px solid hsl(var(--primary) / 0.3);
}
.badge-ghost {
  border: 1px solid hsl(var(--border));
  color: hsl(var(--muted-foreground));
}
.badge-muted {
  background: hsl(var(--muted));
  color: hsl(var(--muted-foreground));
}
</style>