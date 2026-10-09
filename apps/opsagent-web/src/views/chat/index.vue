<template>
  <div class="opsagent-page chat-page">
    <!-- 对话流 + 报告卡 + 粘底输入条（原型 chat.html 同构） -->
    <ChatPanel
      :messages="messages"
      :diagnosis="diagnosis"
      :degrade-level="degradeLevel"
      :busy="busy"
      :username="userStore.username"
      @submit="onSubmit"
    />

    <!-- 澄清追问（type=clarify）：候选意图一键纠正 / 自然语言纠正 -->
    <div v-if="chatData?.type === 'clarify'" class="chat-page__clarify">
      <IntentCorrection
        :candidates="chatData.candidates"
        :busy="busy"
        @correct="onCorrectIntent"
        @correct-message="onCorrectMessage"
      />
    </div>

    <!-- L2 预置查询目录（type=preset_entries）：点击确定性重入 -->
    <div v-if="chatData?.type === 'preset_entries'" class="chat-page__presets">
      <p class="presets__label">预置查询</p>
      <button
        v-for="p in chatData.presets"
        :key="p.id"
        type="button"
        class="presets__item"
        :disabled="busy"
        @click="onPreset(p)"
      >
        {{ p.label }}
      </button>
    </div>

    <!-- 明示降级对话框（原型 degradeDialog：说明 + 知道了） -->
    <div
      v-if="degradeDialogOpen"
      class="dialog-overlay open"
      role="dialog"
      aria-modal="true"
      aria-label="降级模式"
      @click.self="degradeDialogOpen = false"
    >
      <div class="dialog">
        <h3>⚠️ 进入降级模式</h3>
        <p class="dialog__desc">
          LLM 意图识别不可用，已退化为模板/正则匹配 + 预置查询入口。以下能力当前不可用：自由转述提问、复杂多工具编排。
        </p>
        <div class="dialog-actions">
          <button class="btn" @click="degradeDialogOpen = false">知道了</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
    chatApi,
    correctApi,
    OpsagentRequestError,
    type ChatData,
    type DiagnoseParams,
    type IntentType,
    type PresetQuery,
} from '@/api/opsagent'
import { computed, ref, watch } from 'vue'
import { useUserStore } from '@/stores/user'
import ChatPanel, { type ChatMessage } from './components/ChatPanel.vue'
import IntentCorrection from './components/IntentCorrection.vue'

const userStore = useUserStore()

const busy = ref(false)
const messages = ref<ChatMessage[]>([])
const chatData = ref<ChatData | null>(null)
const degradeDialogOpen = ref(false)

const diagnosis = computed(() => chatData.value?.diagnosis ?? null)
const degradeLevel = computed(() => chatData.value?.degradeLevel ?? 0)

/** 落库失败（ERR_PERSIST_FAILED）仍携带可用 report，不应丢弃诊断正文 */
const CODE_PERSIST_FAILED = 'ERR_PERSIST_FAILED'

function applyChatData(data: ChatData): void {
    chatData.value = data
    // type=report 时诊断以报告卡呈现（ChatPanel 渲 diagnosis），正文不重复入气泡
    if (data.report && data.type !== 'report') {
        messages.value.push({ role: 'assistant', text: data.report })
    }
    if (data.status === 'running') {
        messages.value.push({ role: 'assistant', text: '诊断转入后台继续，请稍后从「历史回溯」查看结果。' })
    }
}

/** 明示降级 → 降级模式对话框（原型 degradeDialog） */
watch(
    () => chatData.value?.type,
    (type) => {
        if (type === 'degraded_notice') {
            degradeDialogOpen.value = true
        }
    },
)

/** 统一错误收敛：persist-failed 保留正文，其余以文本入对话流 */
function onError(err: unknown): void {
    if (err instanceof OpsagentRequestError && err.code === CODE_PERSIST_FAILED && err.data) {
        applyChatData(err.data as ChatData)
        return
    }
    messages.value.push({ role: 'assistant', text: err instanceof Error ? `请求失败：${err.message}` : '请求失败' })
}

async function onSubmit(message: string): Promise<void> {
    messages.value.push({ role: 'user', text: message })
    busy.value = true
    try {
        applyChatData(await chatApi({ message }))
    } catch (err) {
        onError(err)
    } finally {
        busy.value = false
    }
}

async function onCorrectIntent(targetIntent: IntentType, params: DiagnoseParams): Promise<void> {
    if (!chatData.value?.sessionId) return
    busy.value = true
    try {
        applyChatData(await correctApi(chatData.value.sessionId, { targetIntent, params }))
    } catch (err) {
        onError(err)
    } finally {
        busy.value = false
    }
}

async function onCorrectMessage(message: string): Promise<void> {
    if (!chatData.value?.sessionId) return
    busy.value = true
    try {
        applyChatData(await correctApi(chatData.value.sessionId, { message }))
    } catch (err) {
        onError(err)
    } finally {
        busy.value = false
    }
}

/** L2 预置查询条目点击：以 diagnose + 完整 params 确定性重入（不调用 LLM） */
async function onPreset(preset: PresetQuery): Promise<void> {
    await onCorrectIntent('diagnose', preset.params)
}
</script>

<style scoped>
.chat-page {
  display: flex;
  flex-direction: column;
}
.chat-page__clarify {
  padding: 12px;
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
  margin-top: 16px;
}
.chat-page__presets {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 16px;
}
.presets__label {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  color: hsl(var(--muted-foreground));
}
.presets__item {
  padding: 10px 14px;
  text-align: left;
  background: hsl(var(--card));
  color: hsl(var(--foreground));
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
  cursor: pointer;
  font-family: inherit;
}
.presets__item:hover:not(:disabled) {
  border-color: hsl(var(--primary));
}
.presets__item:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

/* 原型 Dialog 同构 */
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(2px);
  z-index: 90;
  display: flex;
  align-items: center;
  justify-content: center;
}
.dialog {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
  max-width: 480px;
  width: 90%;
  padding: 24px;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6);
}
.dialog h3 {
  margin: 0 0 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
}
.dialog__desc {
  margin: 0 0 8px;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  line-height: 1.6;
}
.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 20px;
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
</style>