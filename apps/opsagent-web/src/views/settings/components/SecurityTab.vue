<template>
  <div class="security-tab">
    <!-- 多租户隔离（无配置契约：两项均为服务端强制策略，诚实呈现「服务端强制开启」） -->
    <div class="card">
      <div class="card-header"><h3>多租户隔离</h3></div>
      <div
        v-for="(s, i) in SECURITY_SWITCHES"
        :key="s.key"
        class="sec-row"
        :class="{ 'sec-row--last': i === SECURITY_SWITCHES.length - 1 }"
      >
        <div class="sec-row-text">
          <p class="sec-row-title">{{ s.title }}</p>
          <p class="sec-row-hint">{{ s.hint }}</p>
        </div>
        <span class="badge badge-primary" :title="'由服务端强制，前端无配置开关可写'">服务端强制开启</span>
      </div>
    </div>

    <!-- 风险分级白名单（真实 CRUD：管理员增删 → 顶栏「保存配置」PUT riskWhitelist） -->
    <div class="card">
      <div class="card-header"><h3>风险分级白名单（写操作）</h3></div>
      <p class="whitelist-note">
        风险分级由确定性代码（工具元数据注册表 + 白名单匹配）完成，LLM 输出不得改变风险档位；
        高危工具不允许入白名单（服务端兜底校验）。
      </p>

      <div class="whitelist-table">
        <table class="data-table">
          <thead>
            <tr><th>操作</th><th>风险档</th><th>处置</th><th v-if="editable" class="col-op">操作</th></tr>
          </thead>
          <tbody>
            <tr v-for="w in whitelist" :key="w.tool">
              <td>{{ riskToolLabel(w.tool) }}</td>
              <td>
                <span class="badge" :class="`badge-${RISK_LEVEL_META[w.riskLevel].tone}`">
                  {{ RISK_LEVEL_META[w.riskLevel].label }}
                </span>
              </td>
              <td>{{ dispositionForRisk(w.riskLevel) }}</td>
              <td v-if="editable" class="col-op">
                <button
                  type="button"
                  class="btn-remove"
                  :disabled="busy"
                  :aria-label="`移除 ${w.tool}`"
                  @click="remove(w.tool)"
                >
                  移除
                </button>
              </td>
            </tr>
            <tr v-if="whitelist.length === 0">
              <td :colspan="editable ? 4 : 3" class="whitelist-empty">暂无白名单条目</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 新增条目（管理员可写；高危工具选项禁用） -->
      <div v-if="editable" class="whitelist-add">
        <select
          v-model="addTool"
          class="input"
          aria-label="选择工具"
          :disabled="busy"
          @change="clearError"
        >
          <option value="" disabled>选择工具</option>
          <optgroup label="可入白名单（read / low）">
            <option v-for="t in candidates" :key="t.tool" :value="t.tool">
              {{ t.label }}（{{ RISK_LEVEL_META[t.level].label }}）
            </option>
          </optgroup>
          <optgroup label="高危 · 不允许入白名单">
            <option v-for="t in highTools" :key="t.tool" :value="t.tool" disabled>
              {{ t.label }} · 不可入白名单
            </option>
          </optgroup>
        </select>
        <select
          v-model="addLevel"
          class="input input--level"
          aria-label="风险档"
          :disabled="busy || !addTool"
        >
          <option value="read">只读 · 自动执行</option>
          <option value="low">低危 · 白名单自动执行</option>
        </select>
        <button
          type="button"
          class="btn"
          :disabled="busy || !addTool"
          @click="add"
        >
          添加
        </button>
        <p v-if="addError" class="whitelist-error" role="alert">{{ addError }}</p>
      </div>
    </div>

    <!-- 底座接口契约（设计文档静态事实） -->
    <div class="card">
      <div class="card-header"><h3>底座接口契约</h3></div>
      <p class="contract-note">编排层仅依赖契约文档所列接口；契约测试在 CI 守护接口漂移（P1 显式交付物）。</p>
      <div class="contract-list">
        <div v-for="c in API_CONTRACTS" :key="c.name" class="contract-row">
          <span>{{ c.name }}</span>
          <span class="badge badge-primary">已冻结 {{ c.version }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { RiskLevel, RiskWhitelistEntry } from '@/api/opsagent'
import {
    addWhitelistEntry,
    API_CONTRACTS,
    dispositionForRisk,
    RISK_LEVEL_META,
    RISK_TOOL_CATALOG,
    riskToolLabel,
    SECURITY_SWITCHES,
} from '../logic'

const props = defineProps<{
    /** 父级持有的白名单草稿（增删经 draft 上抛，顶栏保存统一提交） */
    whitelist: RiskWhitelistEntry[]
    editable?: boolean
    busy?: boolean
}>()

const emit = defineEmits<{ (e: 'draft', whitelist: RiskWhitelistEntry[]): void }>()

/** 可入白名单工具（read/low）与高危（禁选展示） */
const candidates = computed(() => RISK_TOOL_CATALOG.filter((t) => t.level !== 'high'))
const highTools = computed(() => RISK_TOOL_CATALOG.filter((t) => t.level === 'high'))

const addTool = ref('')
const addLevel = ref<RiskLevel>('read')
const addError = ref('')

function clearError(): void {
    addError.value = ''
}

function add(): void {
    const result = addWhitelistEntry(props.whitelist, {
        tool: addTool.value,
        riskLevel: addLevel.value,
    })
    if (result.error) {
        addError.value = result.error
        return
    }
    addError.value = ''
    addTool.value = ''
    addLevel.value = 'read'
    emit('draft', result.list)
}

function remove(tool: string): void {
    emit('draft', props.whitelist.filter((w) => w.tool !== tool))
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
.card:last-child {
    margin-bottom: 0;
}
.card-header {
    margin-bottom: 14px;
}
.card-header h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
}
.sec-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 12px;
}
.sec-row--last {
    margin-bottom: 0;
}
.sec-row-text {
    min-width: 0;
}
.sec-row-title {
    margin: 0 0 4px;
    font-size: 13px;
    font-weight: 500;
}
.sec-row-hint {
    margin: 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
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
    flex-shrink: 0;
}
.badge-info {
    background: hsl(var(--primary) / 0.15);
    color: hsl(var(--primary));
    border: 1px solid hsl(var(--primary) / 0.3);
}
.badge-warning {
    background: hsl(var(--severity-low) / 0.15);
    color: hsl(var(--severity-low));
    border: 1px solid hsl(var(--severity-low) / 0.3);
}
.badge-danger {
    background: hsl(var(--severity-high) / 0.15);
    color: hsl(var(--severity-high));
    border: 1px solid hsl(var(--severity-high) / 0.3);
}
.badge-primary {
    background: hsl(var(--primary) / 0.15);
    color: hsl(var(--primary));
    border: 1px solid hsl(var(--primary) / 0.3);
}
.whitelist-note,
.contract-note {
    margin: 0 0 12px;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
    line-height: 1.6;
}
.whitelist-table {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    overflow: hidden;
    margin-bottom: 14px;
}
.data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
}
.data-table thead th {
    background: hsl(var(--muted) / 0.4);
    text-align: left;
    padding: 10px 14px;
    font-weight: 500;
    color: hsl(var(--muted-foreground));
    border-bottom: 1px solid hsl(var(--border));
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.03em;
}
.data-table tbody td {
    padding: 11px 14px;
    border-bottom: 1px solid hsl(var(--border));
    vertical-align: middle;
}
.data-table tbody tr:last-child td {
    border-bottom: none;
}
.col-op {
    text-align: right;
    width: 90px;
}
.whitelist-empty {
    text-align: center;
    color: hsl(var(--muted-foreground));
    font-style: italic;
    padding: 14px;
}
.whitelist-add {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
}
.input {
    padding: 8px 12px;
    background: hsl(var(--input));
    color: hsl(var(--foreground));
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    font-size: 13px;
    font-family: inherit;
    min-width: 240px;
}
.input--level {
    min-width: 180px;
}
.input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 34px;
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
.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
.btn-remove {
    padding: 4px 10px;
    font-size: 12px;
    background: hsl(var(--destructive) / 0.12);
    color: hsl(var(--destructive));
    border: 1px solid hsl(var(--destructive) / 0.4);
    border-radius: 6px;
    cursor: pointer;
    font-family: inherit;
}
.btn-remove:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
.whitelist-error {
    flex-basis: 100%;
    margin: 0;
    font-size: 12px;
    color: hsl(var(--destructive));
}
.contract-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    font-size: 13px;
}
.contract-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
}
</style>