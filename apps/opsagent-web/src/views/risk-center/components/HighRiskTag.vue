<template>
  <span v-if="high" class="high-risk-tag" role="status">待人工确认</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Severity } from '@/api/opsagent'
import { isHighRisk } from '../logic'

const props = defineProps<{ severity: Severity }>()

/** 高危（severity P0|P1）→ ghost 徽标「待人工确认」（原型 risk-center.html 行内 ghost badge） */
const high = computed(() => isHighRisk(props.severity))
</script>

<style scoped>
.high-risk-tag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    font-weight: 500;
    padding: 2px 10px;
    border-radius: 9999px;
    border: 1px solid hsl(var(--border));
    color: hsl(var(--muted-foreground));
    white-space: nowrap;
    margin-left: 8px;
}
</style>