<template>
  <div class="search-bar">
    <div class="search-bar__field search-bar__field--wide">
      <label class="field-label" for="h-range">时间窗（≤24h）</label>
      <el-date-picker
        id="h-range"
        v-model="timeRange"
        class="search-bar__picker"
        type="datetimerange"
        start-placeholder="最近 24 小时"
        end-placeholder="结束时间"
        value-format="YYYY-MM-DDTHH:mm:ssZ"
        clearable
      />
    </div>
    <div class="search-bar__field">
      <label class="field-label" for="h-svc">服务名</label>
      <el-input
        id="h-svc"
        v-model="serviceName"
        class="search-bar__input"
        placeholder="如 order-service"
        clearable
        @keyup.enter="onSearch"
      />
    </div>
    <button type="button" class="btn search-bar__btn" @click="onSearch">检索</button>

    <!-- 超窗提示（原型 history.html：表单内联 error-banner，非 toast） -->
    <div v-if="rangeError" class="error-banner" role="alert">时间窗上限 24 小时，请缩小范围</div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { isTimeWindowValid } from '../logic'

/** 检索参数（父级负责转 API query；≤24h 校验在本组件内联完成） */
export interface HistoryFilter {
    serviceName: string
    /** [startTime, endTime]（RFC3339） */
    timeRange: [string, string] | null
}

const props = defineProps<{
    /** 默认时间窗（父级注入，通常 now-24h） */
    defaultRange?: [string, string]
}>()

const emit = defineEmits<{
    (e: 'search', filters: HistoryFilter): void
}>()

const serviceName = ref('')
const timeRange = ref<[string, string] | null>(props.defaultRange ?? null)
const rangeError = ref(false)

function current(): HistoryFilter {
    return { serviceName: serviceName.value.trim(), timeRange: timeRange.value }
}

function onSearch(): void {
    const f = current()
    if (f.timeRange && !isTimeWindowValid(f.timeRange[0], f.timeRange[1])) {
        rangeError.value = true
        return
    }
    rangeError.value = false
    emit('search', f)
}

/** 时间窗变更时清掉超窗提示（原型交互直觉） */
function clearError(): void {
    rangeError.value = false
}
watch(timeRange, clearError)
</script>

<style scoped>
.search-bar {
    display: flex;
    align-items: flex-end;
    gap: 12px;
    flex-wrap: wrap;
    padding: 20px;
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    margin-bottom: 16px;
}
.search-bar__field {
    display: flex;
    flex-direction: column;
    gap: 5px;
}
.search-bar__field--wide {
    flex: 1 1 220px;
}
.search-bar__field:not(.search-bar__field--wide) {
    flex: 1 1 200px;
}
.field-label {
    font-size: 12px;
    font-weight: 500;
    color: hsl(var(--muted-foreground));
}
.search-bar__picker,
.search-bar__input {
    width: 100%;
}
.search-bar__btn {
    height: 36px;
    flex-shrink: 0;
    margin-bottom: 1px;
}
.error-banner {
    flex-basis: 100%;
    background: hsl(var(--destructive) / 0.12);
    color: hsl(var(--destructive));
    border: 1px solid hsl(var(--destructive) / 0.35);
    border-radius: 8px;
    padding: 10px 14px;
    font-size: 13px;
    margin: 0;
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