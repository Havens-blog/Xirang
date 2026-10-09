<template>
  <div class="security-tab">
    <!-- 多租户隔离（原型：两个强制开关；服务端强制 → 只读展示 ON） -->
    <div class="card">
      <div class="card-header"><h3>多租户隔离</h3></div>
      <div v-for="(s, i) in SECURITY_SWITCHES" :key="s.key" class="sec-row" :class="{ 'sec-row--last': i === SECURITY_SWITCHES.length - 1 }">
        <div class="sec-row-text">
          <p class="sec-row-title">{{ s.title }}</p>
          <p class="sec-row-hint">{{ s.hint }}</p>
        </div>
        <button
          type="button"
          class="switch"
          role="switch"
          aria-checked="true"
          :aria-label="s.title"
          disabled
          :title="`${s.title}由服务端强制开启，前端不可配置`"
        />
      </div>
    </div>

    <!-- 风险分级白名单（原型：确定性白名单表 操作/风险档/处置） -->
    <div class="card">
      <div class="card-header"><h3>风险分级白名单（写操作）</h3></div>
      <p class="whitelist-note">
        风险分级由确定性代码（工具元数据注册表 + 白名单匹配）完成，LLM 输出不得改变风险档位。
      </p>
      <div class="whitelist-table">
        <table class="data-table">
          <thead>
            <tr><th>操作</th><th>风险档</th><th>处置</th></tr>
          </thead>
          <tbody>
            <tr v-for="w in whitelist" :key="w.tool">
              <td>{{ w.tool }}</td>
              <td>
                <span class="badge" :class="`badge-${RISK_LEVEL_META[w.riskLevel].tone}`">
                  {{ RISK_LEVEL_META[w.riskLevel].label }}
                </span>
              </td>
              <td>{{ dispositionForRisk(w.riskLevel) }}</td>
            </tr>
            <tr v-if="whitelist.length === 0">
              <td colspan="3" class="whitelist-empty">暂无白名单条目</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 底座接口契约（原型：静态冻结版本列表，来自设计文档） -->
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
import type { RiskWhitelistEntry } from '@/api/opsagent'
import { API_CONTRACTS, dispositionForRisk, RISK_LEVEL_META, SECURITY_SWITCHES } from '../logic'

defineProps<{ whitelist: RiskWhitelistEntry[] }>()
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
.switch {
    position: relative;
    flex-shrink: 0;
    width: 42px;
    height: 24px;
    border-radius: 9999px;
    border: 1px solid hsl(var(--primary));
    background: hsl(var(--primary));
}
.switch::after {
    content: '';
    position: absolute;
    top: 2px;
    right: 2px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: hsl(var(--card));
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.4);
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
.whitelist-empty {
    text-align: center;
    color: hsl(var(--muted-foreground));
    font-style: italic;
    padding: 14px;
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