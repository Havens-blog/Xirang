<template>
  <div class="opsagent-page">
    <!-- 面包屑（原型 history.html：对话排障 / 历史诊断回溯） -->
    <div class="breadcrumb">
      <RouterLink to="/chat">对话排障</RouterLink>
      <span class="sep">/</span>
      <span>历史诊断回溯</span>
    </div>

    <!-- 列表视图（原型：检索卡 + 会话表格 + 分页） -->
    <template v-if="!selectedSession">
      <SearchBar :default-range="defaultRange" @search="onSearch" />
      <SessionList
        :sessions="items"
        :total="total"
        :page="page"
        :limit="limit"
        @select="onSelect"
        @page-change="onPageChange"
      />
    </template>

    <!-- 会话内容视图（原型：页内替换，↩ 返回会话列表） -->
    <SessionView
      v-else
      :session="selectedSession"
      :diagnosis="selectedDiagnosis"
      :loading="loadingDiagnosis"
      @back="onBackToList"
    />
  </div>
</template>

<script setup lang="ts">
import {
    getDiagnosisApi,
    historyApi,
    type Diagnosis,
    type HistoryParams,
    type SessionSummary,
} from '@/api/opsagent'
import { ref } from 'vue'
import SearchBar, { type HistoryFilter } from './components/SearchBar.vue'
import SessionList from './components/SessionList.vue'
import SessionView from './components/SessionView.vue'
import { defaultTimeWindow } from './logic'

const items = ref<SessionSummary[]>([])
const filters = ref<HistoryFilter | null>(null)
const page = ref(1)
const limit = 20
const total = ref(0)
const loadingList = ref(false)

const selectedSession = ref<SessionSummary | null>(null)
const selectedDiagnosis = ref<Diagnosis | null>(null)
const loadingDiagnosis = ref(false)

/** 默认时间窗 now-24h（page-map Query Parameters） */
const defaultRange = ref<[string, string]>(defaultTimeWindow())

function buildParams(): HistoryParams {
    const p: HistoryParams = { page: page.value, limit }
    const f = filters.value
    if (f?.serviceName) p.serviceName = f.serviceName
    const range = f?.timeRange ?? defaultRange.value
    p.startTime = range[0]
    p.endTime = range[1]
    return p
}

async function load(): Promise<void> {
    loadingList.value = true
    try {
        const data = await historyApi(buildParams())
        items.value = data.items
        total.value = data.total
    } catch {
        // 跨租户/服务异常由请求层 401/404 收敛；此处静默避免打断浏览
        items.value = []
        total.value = 0
    } finally {
        loadingList.value = false
    }
}

/** 检索：≤24h 校验已由 SearchBar 内联完成（超窗时不会到达此分支） */
function onSearch(f: HistoryFilter): void {
    filters.value = f
    page.value = 1
    load()
}

function onPageChange(next: number): void {
    page.value = next
    load()
}

async function onSelect(session: SessionSummary): Promise<void> {
    selectedSession.value = session
    selectedDiagnosis.value = null
    if (!session.diagnosisId) return
    loadingDiagnosis.value = true
    try {
        selectedDiagnosis.value = await getDiagnosisApi(session.diagnosisId)
    } catch {
        selectedDiagnosis.value = null
    } finally {
        loadingDiagnosis.value = false
    }
}

function onBackToList(): void {
    selectedSession.value = null
    load()
}
</script>

<style scoped>
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
</style>