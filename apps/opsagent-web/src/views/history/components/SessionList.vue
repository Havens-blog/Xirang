<template>
  <div class="session-list">
    <p v-if="sessions.length === 0" class="session-list__empty">暂无会话记录</p>
    <div v-else class="session-list__card">
      <!-- 原型 history.html：data-table + 分页，列 = 服务名/意图/触发来源/时间/降级/操作 -->
      <table class="data-table">
        <thead>
          <tr>
            <th>服务名</th>
            <th>意图</th>
            <th>触发来源</th>
            <th>时间</th>
            <th>降级</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in sessions"
            :key="s.id"
            class="clickable"
            :class="{ 'session-row--active': s.id === selectedId }"
            @click="emit('select', s)"
          >
            <td class="session-row__service">{{ s.serviceName || '未指定服务' }}</td>
            <td>
              <span class="agent-chip" :class="intentChip(s.intentType).style">
                {{ intentChip(s.intentType).text }}
              </span>
            </td>
            <td>
              <span v-if="s.type === 'chat'" class="agent-chip log">对话</span>
              <span v-else class="badge badge-ghost">告警</span>
            </td>
            <td class="session-row__time">{{ formatClock(s.createdAt) }}</td>
            <td>—</td>
            <td>
              <button type="button" class="btn btn-sm btn-outline" @click.stop="emit('select', s)">查看</button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页（原型：共 N 条 ‹ 1 ›） -->
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
import type { SessionSummary } from '@/api/opsagent'
import { formatClock, intentChip } from '../logic'

withDefaults(
    defineProps<{
        sessions: SessionSummary[]
        selectedId?: string
        /** 分页信息（缺省整体隐藏分页条） */
        total?: number
        page?: number
        limit?: number
    }>(),
    { page: 1, limit: 20 },
)

const emit = defineEmits<{
    (e: 'select', session: SessionSummary): void
    (e: 'page-change', page: number): void
}>()
</script>

<style scoped>
.session-list__empty {
    margin: 0;
    padding: 20px;
    text-align: center;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
}
.session-list__card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    padding: 0;
    overflow: hidden;
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
    padding: 13px 14px;
    border-bottom: 1px solid hsl(var(--border));
    vertical-align: middle;
}
.data-table tbody tr.clickable {
    cursor: pointer;
}
.data-table tbody tr:hover {
    background: hsl(var(--accent) / 0.4);
}
.session-row--active {
    background: hsl(var(--primary) / 0.08);
}
.session-row__service {
    font-weight: 500;
    color: hsl(var(--foreground));
}
.session-row__time {
    font-family: monospace;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.agent-chip {
    font-size: 11px;
    padding: 2px 10px;
    border-radius: 9999px;
    font-weight: 500;
}
.agent-chip.diagnose {
    background: hsl(var(--agent-inspector) / 0.15);
    color: hsl(var(--agent-inspector));
    border: 1px solid hsl(var(--agent-inspector) / 0.3);
}
.agent-chip.log {
    background: hsl(var(--agent-log) / 0.15);
    color: hsl(var(--agent-log));
    border: 1px solid hsl(var(--agent-log) / 0.3);
}
.agent-chip.ghost {
    border: 1px solid hsl(var(--border));
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
}
.badge-ghost {
    border: 1px solid hsl(var(--border));
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
.btn-outline {
    background: transparent;
    border-color: hsl(var(--border));
    color: hsl(var(--foreground));
}
.btn-outline:hover {
    background: hsl(var(--accent));
}
.btn-sm {
    height: 30px;
    padding: 0 12px;
    font-size: 12px;
    border-radius: 6px;
}
.pagination {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 14px 16px;
    font-size: 13px;
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