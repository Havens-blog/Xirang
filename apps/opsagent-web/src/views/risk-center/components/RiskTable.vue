<template>
  <!-- 风险条目表（原型 risk-center.html：批量操作条 + data-table + 内嵌分页） -->
  <div class="risk-table">
    <!-- 批量操作条（原型：左批量按钮 + 右「已选 N 条」） -->
    <div class="risk-table__toolbar">
      <div class="risk-table__batch">
        <button
          type="button"
          class="btn btn-sm btn-outline"
          :disabled="busy || selected.length === 0"
          @click="emit('batch', 'viewed', selected)"
        >
          批量标记已查看
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline"
          :disabled="busy || selected.length === 0"
          @click="emit('batch', 'done', selected)"
        >
          批量标记已处理
        </button>
      </div>
      <span class="risk-table__selected">已选 {{ selected.length }} 条</span>
    </div>

    <!-- 表格卡片（原型列：□ 服务名 级别 风险分 时间 状态 降级 操作） -->
    <div class="risk-table__card">
      <el-table :data="items" @selection-change="onSelectionChange">
        <el-table-column type="selection" width="44" />
        <el-table-column label="服务名" min-width="140">
          <template #default="{ row }">
            <span class="risk-table__service">{{ row.serviceName }}</span>
          </template>
        </el-table-column>
        <el-table-column label="级别" min-width="150">
          <template #default="{ row }">
            <span class="sev-badge" :class="`sev-badge--${sevMeta(row.severity).tone}`">
              {{ sevMeta(row.severity).label }}
            </span>
            <HighRiskTag :severity="row.severity" />
          </template>
        </el-table-column>
        <el-table-column label="风险分" width="90">
          <template #default>—</template>
        </el-table-column>
        <el-table-column label="时间" width="90">
          <template #default="{ row }">
            <span class="risk-table__time">{{ formatClock(row.createdAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">{{ RISK_STATUS_LABEL[row.status as RiskEntryStatus] }}</template>
        </el-table-column>
        <el-table-column label="降级" width="80">
          <template #default>—</template>
        </el-table-column>
        <el-table-column label="操作" min-width="180">
          <template #default="{ row }">
            <div class="risk-table__actions">
              <button type="button" class="btn btn-sm btn-outline" @click="emit('openDetail', row)">
                查看详情
              </button>
              <el-dropdown
                v-if="markableActions(row.status as RiskEntryStatus).length > 0"
                trigger="click"
                @command="(status: 'viewed' | 'done') => emit('mark', row, status)"
              >
                <button type="button" class="btn btn-sm btn-outline risk-table__more" aria-label="更多操作">
                  ⋯
                </button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-for="a in markableActions(row.status as RiskEntryStatus)"
                      :key="a.status"
                      :command="a.status"
                    >
                      {{ a.label }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页（原型：共 N 条 ‹ 1 ›，内嵌卡片） -->
      <div v-if="total !== undefined" class="pagination">
        <span class="pagination__total">共 {{ total }} 条</span>
        <span class="pagination__flex" />
        <button
          type="button"
          class="page-btn"
          :disabled="page <= 1"
          aria-label="上一页"
          @click="emit('page-change', page - 1)"
        >
          ‹
        </button>
        <button type="button" class="page-btn current">{{ page }}</button>
        <button
          type="button"
          class="page-btn"
          :disabled="page * limit >= total"
          aria-label="下一页"
          @click="emit('page-change', page + 1)"
        >
          ›
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { RiskEntry, RiskEntryStatus, Severity } from '@/api/opsagent'
import { formatClock, markableActions, RISK_STATUS_LABEL, SEVERITY_META } from '../logic'
import HighRiskTag from './HighRiskTag.vue'

withDefaults(
    defineProps<{
        items: RiskEntry[]
        busy?: boolean
        total?: number
        page?: number
        limit?: number
    }>(),
    { page: 1, limit: 20 },
)

const emit = defineEmits<{
    (e: 'mark', entry: RiskEntry, status: 'viewed' | 'done'): void
    (e: 'batch', status: 'viewed' | 'done', entries: RiskEntry[]): void
    (e: 'openDetail', entry: RiskEntry): void
    (e: 'page-change', page: number): void
}>()

const selected = ref<RiskEntry[]>([])

function onSelectionChange(rows: RiskEntry[]): void {
    selected.value = rows
}

/** el-table 插槽行类型为 any，经具名助手收窄为 domain 类型 */
function sevMeta(severity: Severity): { label: string; tone: string } {
    return SEVERITY_META[severity]
}
</script>

<style scoped>
.risk-table__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
}
.risk-table__batch {
    display: flex;
    align-items: center;
    gap: 10px;
}
.risk-table__selected {
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.risk-table__card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    padding: 0;
    overflow: hidden;
}
.risk-table__service {
    font-weight: 500;
    color: hsl(var(--foreground));
}
.risk-table__time {
    font-family: monospace;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.risk-table__actions {
    display: flex;
    align-items: center;
    gap: 8px;
}
.risk-table__more {
    padding: 0 10px;
    font-weight: 700;
    letter-spacing: 1px;
}
.sev-badge {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 10px;
    border-radius: 9999px;
    white-space: nowrap;
}
.sev-badge--danger {
    background: hsl(var(--severity-high) / 0.15);
    color: hsl(var(--severity-high));
}
.sev-badge--warning {
    background: hsl(var(--severity-low) / 0.15);
    color: hsl(var(--severity-low));
}
.sev-badge--info {
    background: hsl(var(--primary) / 0.15);
    color: hsl(var(--primary));
}
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    height: 30px;
    padding: 0 12px;
    border-radius: 6px;
    border: 1px solid hsl(var(--border));
    background: transparent;
    color: hsl(var(--foreground));
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    font-family: inherit;
}
.btn:hover:not(:disabled) {
    background: hsl(var(--accent));
}
.btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
}
.pagination {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 14px 16px;
}
.pagination__total {
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.pagination__flex {
    flex: 1;
}
.page-btn {
    min-width: 30px;
    height: 30px;
    padding: 0 8px;
    border-radius: 6px;
    border: 1px solid hsl(var(--border));
    background: transparent;
    cursor: pointer;
    font-size: 12px;
    color: hsl(var(--foreground));
    font-family: inherit;
}
.page-btn:hover:not(:disabled) {
    background: hsl(var(--muted));
}
.page-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}
.page-btn.current {
    background: hsl(var(--primary) / 0.15);
    color: hsl(var(--primary));
    border-color: hsl(var(--primary) / 0.3);
}
</style>