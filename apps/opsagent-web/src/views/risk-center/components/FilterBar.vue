<template>
  <!-- 筛选条（原型 risk-center.html：卡片内联表单，按钮文案「筛选」） -->
  <div class="filter-bar">
    <div class="filter-bar__field filter-bar__field--wide">
      <label class="field-label" for="f-range">时间范围</label>
      <el-date-picker
        id="f-range"
        v-model="timeRange"
        class="filter-bar__control"
        type="datetimerange"
        start-placeholder="最近 24 小时"
        end-placeholder="结束时间"
        value-format="YYYY-MM-DDTHH:mm:ssZ"
        clearable
      />
    </div>
    <div class="filter-bar__field">
      <label class="field-label" for="f-svc">服务名</label>
      <el-input
        id="f-svc"
        v-model="serviceName"
        class="filter-bar__control"
        placeholder="如 order-service"
        clearable
        @keyup.enter="onSearch"
      />
    </div>
    <div class="filter-bar__field filter-bar__field--short">
      <label class="field-label" for="f-status">状态</label>
      <el-select id="f-status" v-model="status" class="filter-bar__control" placeholder="全部" clearable>
        <el-option v-for="(label, val) in STATUS_OPTIONS" :key="val" :label="label" :value="val" />
      </el-select>
    </div>
    <div class="filter-bar__field filter-bar__field--short">
      <label class="field-label" for="f-sev">级别</label>
      <el-select id="f-sev" v-model="severity" class="filter-bar__control" placeholder="全部" clearable>
        <el-option v-for="s in SEVERITY_OPTIONS" :key="s" :label="s" :value="s" />
      </el-select>
    </div>
    <button type="button" class="btn filter-bar__btn" @click="onSearch">筛选</button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { RiskEntryStatus, Severity } from '@/api/opsagent'
import { RISK_STATUS_LABEL } from '../logic'

/** 筛选参数（父级负责转成 API query，空值不参与过滤） */
export interface RiskFilter {
    serviceName: string
    status: RiskEntryStatus | ''
    severity: Severity | ''
    /** [startTime, endTime]（RFC3339），空 = 不过滤 */
    timeRange: [string, string] | null
}

const emit = defineEmits<{
    (e: 'search', filters: RiskFilter): void
}>()

const STATUS_OPTIONS = RISK_STATUS_LABEL
const SEVERITY_OPTIONS: Severity[] = ['P0', 'P1', 'P2', 'P3']

const serviceName = ref('')
const status = ref<RiskEntryStatus | ''>('')
const severity = ref<Severity | ''>('')
const timeRange = ref<[string, string] | null>(null)

function current(): RiskFilter {
    return {
        serviceName: serviceName.value.trim(),
        status: status.value,
        severity: severity.value,
        timeRange: timeRange.value,
    }
}

function onSearch(): void {
    emit('search', current())
}
</script>

<style scoped>
.filter-bar {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    align-items: flex-end;
    padding: 20px;
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    margin-bottom: 16px;
}
.filter-bar__field {
    display: flex;
    flex-direction: column;
    gap: 5px;
}
.filter-bar__field--wide {
    flex: 1 1 220px;
}
.filter-bar__field:not(.filter-bar__field--wide) {
    flex: 1 1 160px;
}
.filter-bar__control {
    width: 100%;
}
.filter-bar__btn {
    flex-shrink: 0;
    margin-bottom: 1px;
}
.field-label {
    font-size: 12px;
    font-weight: 500;
    color: hsl(var(--muted-foreground));
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