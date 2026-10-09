<template>
  <div class="llm-tab">
    <!-- 模型提供商（原型 settings.html：provider-card 单选） -->
    <div class="card">
      <div class="card-header"><h3>模型提供商</h3></div>
      <div v-if="providers.length === 0" class="llm-tab__empty">未配置 LLM 提供商</div>
      <div
        v-for="p in providers"
        :key="p.name"
        class="provider-card"
        :class="{ selected: p.default }"
        role="radio"
        :aria-checked="p.default"
        tabindex="0"
        @click="selectDefault(p.name)"
      >
        <div
          class="ds-icon"
          :style="{ background: PROVIDER_META[p.name]?.grad ?? 'hsl(var(--primary))' }"
        >
          {{ PROVIDER_META[p.name]?.letter ?? providerLabel(p.name)[0] ?? '?' }}
        </div>
        <div class="provider-card-meta">
          <p class="provider-card-name">{{ providerLabel(p.name) }}</p>
          <p class="provider-card-site">{{ PROVIDER_META[p.name]?.site ?? '—' }}</p>
        </div>
        <span class="status-dot" :class="p.default ? 'connected' : 'disconnected'" />
      </div>
    </div>

    <!-- API 配置（原型：Key + Endpoint + 默认模型；作用于当前默认提供商） -->
    <div class="card">
      <div class="card-header"><h3>API 配置</h3></div>
      <div v-if="selected" class="form">
        <div class="form-row">
          <label for="apiKey">API Key <span class="req">*</span></label>
          <div class="form-row-main">
            <input
              id="apiKey"
              v-model="keyInput"
              type="password"
              class="input"
              :placeholder="maskedOrEmpty"
              :disabled="!editable || busy"
              @input="emitDraft"
            />
            <div class="form-hint">密钥不落代码/配置仓，从环境变量注入（对应 Haven `LOGQUERY_LLM_API_KEY`）</div>
          </div>
        </div>
        <div class="form-row">
          <label for="apiBase">API Endpoint（可选）</label>
          <div class="form-row-main">
            <input
              id="apiBase"
              class="input"
              placeholder="默认：https://dashscope.aliyuncs.com/compatible-mode/v1"
              disabled
              title="由服务端环境变量注入，前端不提供编辑"
            />
            <div class="form-hint">自定义代理地址可留空（由服务端配置）</div>
          </div>
        </div>
        <div class="form-row form-row--end">
          <label for="defModel">默认模型</label>
          <div class="form-row-main">
            <input
              id="defModel"
              v-model="modelInput"
              class="input"
              placeholder="如 qwen-plus（推荐）"
              :disabled="!editable || busy"
              @input="emitDraft"
            />
          </div>
        </div>
      </div>
      <p v-else class="llm-tab__empty">未配置 LLM 提供商</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { LLMProviderConfig, LLMProviderInput } from '@/api/opsagent'
import { PROVIDER_META, providerLabel } from '../logic'

const props = defineProps<{
    providers: LLMProviderConfig[]
    editable?: boolean
    busy?: boolean
}>()

const emit = defineEmits<{ (e: 'draft', providers: LLMProviderInput[]): void }>()

/** 当前默认提供商（未配置回退第一个） */
const selected = computed<LLMProviderConfig | undefined>(
    () => props.providers.find((p) => p.default) ?? props.providers[0],
)

const keyInput = ref('')
const modelInput = ref(selected.value?.model ?? '')

// 默认提供商切换时同步模型输入框
watch(
    () => selected.value?.name,
    () => {
        modelInput.value = selected.value?.model ?? ''
    },
)

const maskedOrEmpty = computed(() => selected.value?.keyMasked || '未配置 Key')

/** 变更 → 上抛完整草稿（apiKey 只写；默认由单选决定） */
function emitDraft(): void {
    const out: LLMProviderInput[] = props.providers.map((p) => ({
        name: p.name,
        default: p.default,
        model: p.name === selected.value?.name ? modelInput.value : p.model,
        apiKey: p.name === selected.value?.name ? keyInput.value : '',
    }))
    emit('draft', out)
}

function selectDefault(name: string): void {
    const out: LLMProviderInput[] = props.providers.map((p) => ({
        name: p.name,
        default: p.name === name,
        model: p.name === name ? (p.name === selected.value?.name ? modelInput.value : p.model) : p.model,
        apiKey: '',
    }))
    emit('draft', out)
}
</script>

<style scoped>
.card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 16px;
}
.card-header {
    margin-bottom: 14px;
}
.card-header h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
}
.llm-tab__empty {
    margin: 0;
    padding: 12px 0;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
    font-style: italic;
}
.provider-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    border: 1px solid hsl(var(--border));
    border-radius: 10px;
    margin-bottom: 10px;
    cursor: pointer;
    transition: border-color 0.15s;
}
.provider-card:hover {
    border-color: hsl(var(--primary) / 0.6);
}
.provider-card.selected {
    border-color: hsl(var(--primary));
    background: hsl(var(--primary) / 0.06);
}
.provider-card:last-child {
    margin-bottom: 0;
}
.ds-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 15px;
    font-weight: 700;
    color: #fff;
}
.provider-card-meta {
    flex: 1;
    min-width: 0;
}
.provider-card-name {
    margin: 0 0 3px;
    font-size: 13px;
    font-weight: 600;
}
.provider-card-site {
    margin: 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
}
.status-dot.connected {
    background: hsl(var(--severity-ok));
    box-shadow: 0 0 0 3px hsl(var(--severity-ok) / 0.2);
}
.status-dot.disconnected {
    background: transparent;
    border: 1px solid hsl(var(--border));
}
.form-row {
    display: flex;
    gap: 16px;
    margin-bottom: 16px;
    align-items: flex-start;
}
.form-row--end {
    margin-bottom: 0;
}
.form-row > label {
    width: 170px;
    flex-shrink: 0;
    font-size: 13px;
    font-weight: 500;
    padding-top: 8px;
}
.req {
    color: hsl(var(--destructive));
}
.form-row-main {
    flex: 1;
    min-width: 0;
}
.input {
    width: 100%;
    box-sizing: border-box;
    padding: 8px 12px;
    background: hsl(var(--input));
    color: hsl(var(--foreground));
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    font-size: 13px;
    font-family: inherit;
}
.input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
.form-hint {
    margin: 5px 0 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
</style>