<template>
  <!-- 思考过程块：原型 chat.html thinkingRow（对话流内折叠块，~700ms 点亮一步） -->
  <details class="thinking-block" open>
    <summary>思考过程 · 多工具编排</summary>
    <div class="steps">
      <div
        v-for="(text, i) in THINKING_STEP_TEXTS"
        :key="i"
        class="progress-step"
      >
        <span
          class="progress-step-dot"
          :class="{
            'progress-step-dot--completed': currentStep > i,
            'progress-step-dot--active': currentStep === i,
          }"
          aria-hidden="true"
        />
        <span class="progress-step-text">{{ text }}</span>
      </div>
    </div>
  </details>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { THINKING_STEP_TEXTS } from '../logic'

/**
 * props.active=true 时开始 4 步点亮循环（原型节奏 ~700ms/步），
 * 步进完成后维持全亮；组件卸载清理定时器。
 */
const props = defineProps<{ active: boolean }>()

const currentStep = ref(0)
let timer: number | undefined

function start(): void {
  stop()
  currentStep.value = 0
  timer = window.setInterval(() => {
    if (currentStep.value < THINKING_STEP_TEXTS.length - 1) {
      currentStep.value += 1
    }
  }, 700)
}

function stop(): void {
  if (timer !== undefined) {
    window.clearInterval(timer)
    timer = undefined
  }
}

watch(
  () => props.active,
  (active) => {
    if (active) start()
    else stop()
  },
  { immediate: true },
)

onMounted(() => {
  if (props.active) start()
})
onBeforeUnmount(stop)
</script>

<style scoped>
.thinking-block {
  background: hsl(var(--muted) / 0.3);
  border: 1px solid hsl(var(--border));
  border-left: 3px solid hsl(var(--primary));
  padding: 14px 16px;
  margin: 12px 0;
  border-radius: 0 8px 8px 0;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
}
.thinking-block summary {
  cursor: pointer;
  user-select: none;
  font-weight: 600;
  color: hsl(var(--foreground));
  font-size: 13px;
}
.progress-step {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 0;
}
.progress-step-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 2px solid hsl(var(--border));
}
.progress-step-dot--completed {
  background: hsl(var(--severity-ok));
  border-color: hsl(var(--severity-ok));
}
.progress-step-dot--active {
  border-color: hsl(var(--primary));
  background: hsl(var(--primary));
  box-shadow: 0 0 0 3px hsl(var(--primary) / 0.2);
  animation: pulse2 1.2s infinite;
}
@keyframes pulse2 {
  0%,
  100% {
    box-shadow: 0 0 0 3px hsl(var(--primary) / 0.2);
  }
  50% {
    box-shadow: 0 0 0 6px hsl(var(--primary) / 0.1);
  }
}
</style>