<template>
  <div class="opsagent-page">
    <div v-if="error" class="settings__error">{{ error }}</div>
    <el-skeleton v-else-if="loading" :rows="10" animated />

    <template v-else-if="settings">
      <!-- 页头（原型 settings.html：h1 + 副标题） -->
      <div class="page-header">
        <h1>系统配置</h1>
        <p class="subtitle">数据源、LLM 模型、通知渠道、租户与安全</p>
      </div>

      <!-- Tab 列表（原型 tab-btn 口径；预置查询为第 5 个增量页） -->
      <div class="tab-list" role="tablist" aria-label="系统配置">
        <button
          v-for="tab in TABS"
          :key="tab.key"
          type="button"
          class="tab-btn"
          :class="{ active: activeTab === tab.key }"
          role="tab"
          :aria-selected="activeTab === tab.key"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Panels -->
      <div v-if="activeTab === 'datasource'" class="tab-panel active" role="tabpanel">
        <DatasourceTab :datasources="settings.datasources" />
      </div>
      <div v-else-if="activeTab === 'llm'" class="tab-panel active" role="tabpanel">
        <LLMTab
          :providers="settings.llmProviders"
          :editable="isAdmin"
          :busy="busy"
          @draft="onDraftLLM"
        />
      </div>
      <div v-else-if="activeTab === 'notify'" class="tab-panel active" role="tabpanel">
        <NotifyTab :channels="notifyDraft" :editable="isAdmin" :busy="busy" @toggle="onToggleChannel" />
      </div>
      <div v-else-if="activeTab === 'security'" class="tab-panel active" role="tabpanel">
        <SecurityTab :whitelist="settings.riskWhitelist" />
      </div>
      <div v-else-if="activeTab === 'preset'" class="tab-panel active" role="tabpanel">
        <PresetTab :presets="presetDraft" :editable="isAdmin" :busy="busy" @draft="onDraftPreset" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import {
    getSettingsApi,
    putSettingsApi,
    type LLMProviderInput,
    type NotifyChannel,
    type PresetQuery,
    type SettingsData,
} from '@/api/opsagent'
import { useTopbar } from '@/composables/useTopbar'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import DatasourceTab from './components/DatasourceTab.vue'
import LLMTab from './components/LLMTab.vue'
import NotifyTab from './components/NotifyTab.vue'
import PresetTab from './components/PresetTab.vue'
import SecurityTab from './components/SecurityTab.vue'
import { notifyChannelLabel, validatePreset } from './logic'

/** 原型四 tab + 预置查询（增量功能，后端可存） */
const TABS = [
    { key: 'datasource', label: '数据源配置' },
    { key: 'llm', label: 'LLM 模型' },
    { key: 'notify', label: '通知渠道' },
    { key: 'security', label: '安全与租户' },
    { key: 'preset', label: '预置查询' },
] as const

const userStore = useUserStore()
const isAdmin = computed(() => userStore.isAdmin)

const settings = ref<SettingsData | null>(null)
const loading = ref(true)
const error = ref('')
const busy = ref(false)
const activeTab = ref<(typeof TABS)[number]['key']>('datasource')

// ===== 父级持有的三份可编辑草稿（不可变原则：仅保存时合并提交） =====
const notifyDraft = ref<NotifyChannel[]>([])
const llmDraft = ref<LLMProviderInput[]>([])
const presetDraft = ref<PresetQuery[]>([])

function clonePresets(list: PresetQuery[]): PresetQuery[] {
    return list.map((p) => ({ ...p, params: { ...p.params, timeframe: { ...p.params.timeframe } } }))
}

async function load(): Promise<void> {
    loading.value = true
    error.value = ''
    try {
        const data = await getSettingsApi()
        settings.value = data
        notifyDraft.value = data.notifyChannels.map((c) => ({ ...c }))
        llmDraft.value = data.llmProviders.map((p) => ({
            name: p.name,
            default: p.default,
            model: p.model,
            apiKey: '',
        }))
        presetDraft.value = clonePresets(data.presetQueries)
    } catch (err) {
        error.value = err instanceof Error ? err.message : '加载失败'
    } finally {
        loading.value = false
    }
}

// ===== 各 tab 草稿上抛 =====

/** 通知渠道开关：更新草稿 + 原型 toast（「钉钉 已关闭/已启用」） */
function onToggleChannel(channel: string, enabled: boolean): void {
    notifyDraft.value = notifyDraft.value.map((c) => (c.channel === channel ? { ...c, enabled } : c))
    ElMessage.success(`${notifyChannelLabel(channel)} ${enabled ? '已启用' : '已关闭'}`)
}

function onDraftLLM(providers: LLMProviderInput[]): void {
    llmDraft.value = providers
}

function onDraftPreset(presets: PresetQuery[]): void {
    presetDraft.value = presets
}

// ===== topbar「保存配置」全局按钮（原型 settings.html topbar-actions） =====

async function saveAll(): Promise<void> {
    if (!isAdmin.value) {
        ElMessage.warning('仅平台管理员可保存配置')
        return
    }
    const bad = presetDraft.value.find((p) => validatePreset(p) !== null)
    if (bad) {
        ElMessage.warning(validatePreset(bad) ?? '预置查询校验失败')
        return
    }
    busy.value = true
    try {
        await putSettingsApi({
            llmProviders: llmDraft.value,
            notifyChannels: notifyDraft.value,
            presetQueries: presetDraft.value,
        })
        ElMessage.success('配置已保存')
        await load()
    } catch (err) {
        ElMessage.error(err instanceof Error ? err.message : '保存失败')
    } finally {
        busy.value = false
    }
}

const topbar = useTopbar()
onMounted(() => {
    void load()
    topbar.setActions([
        { key: 'save-config', label: '保存配置', primary: true, onClick: () => void saveAll() },
    ])
})
onBeforeUnmount(() => topbar.setActions([]))
</script>

<style scoped>
.settings__error {
    padding: 24px;
    text-align: center;
    color: hsl(var(--destructive));
}
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
.tab-list {
    display: flex;
    gap: 6px;
    margin-bottom: 16px;
    border-bottom: 1px solid hsl(var(--border));
}
.tab-btn {
    padding: 8px 14px;
    font-size: 13px;
    background: transparent;
    color: hsl(var(--muted-foreground));
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    font-family: inherit;
}
.tab-btn.active {
    color: hsl(var(--primary));
    border-bottom-color: hsl(var(--primary));
    font-weight: 600;
}
.tab-panel {
    padding-top: 2px;
}
</style>