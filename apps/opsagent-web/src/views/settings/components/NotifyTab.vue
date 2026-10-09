<template>
  <div class="notify-tab">
    <p v-if="channels.length === 0" class="notify-tab__empty">未配置通知渠道</p>
    <!-- 原型 settings.html：图标卡 + 描述 + switch（点击即上抛 toggle，父级持有状态并 toast） -->
    <div v-for="c in channels" :key="c.channel" class="card channel-card">
      <div class="channel-card-main">
        <div class="ds-icon" :style="{ background: meta(c.channel).grad }">
          {{ meta(c.channel).icon }}
        </div>
        <div class="channel-card-meta">
          <p class="channel-card-name">{{ notifyChannelLabel(c.channel) }}</p>
          <p class="channel-card-desc">{{ meta(c.channel).desc }}</p>
        </div>
      </div>
      <button
        type="button"
        class="switch"
        role="switch"
        :aria-checked="c.enabled"
        :aria-label="`${notifyChannelLabel(c.channel)} 开关`"
        :disabled="!editable || busy"
        @click="emit('toggle', c.channel, !c.enabled)"
      />
    </div>
    <p v-if="editable" class="notify-tab__note">开关即时保存到后端（PUT /settings notifyChannels）</p>
    <p v-else class="notify-tab__note">仅平台管理员可调整渠道开关</p>
  </div>
</template>

<script setup lang="ts">
import type { NotifyChannel } from '@/api/opsagent'
import { NOTIFY_CHANNEL_META, notifyChannelLabel } from '../logic'

defineProps<{
    channels: NotifyChannel[]
    editable?: boolean
    busy?: boolean
}>()

const emit = defineEmits<{
    (e: 'toggle', channel: string, enabled: boolean): void
}>()

/** 渠道卡元数据（未知渠道兜底灰渐变） */
function meta(channel: string): { icon: string; grad: string; desc: string } {
    return (
        NOTIFY_CHANNEL_META[channel] ?? {
            icon: notifyChannelLabel(channel)[0] ?? '?',
            grad: 'linear-gradient(135deg, #64748b, #475569)',
            desc: '—',
        }
    )
}
</script>

<style scoped>
.notify-tab__empty {
    margin: 0;
    padding: 16px;
    text-align: center;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
}
.notify-tab__note {
    margin: 12px 0 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
.card {
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: 12px;
    padding: 16px;
}
.channel-card {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
}
.channel-card:last-child {
    margin-bottom: 0;
}
.channel-card-main {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
    min-width: 0;
}
.ds-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 700;
    color: #fff;
}
.channel-card-meta {
    min-width: 0;
}
.channel-card-name {
    margin: 0 0 3px;
    font-size: 13px;
    font-weight: 600;
}
.channel-card-desc {
    margin: 0;
    font-size: 12px;
    color: hsl(var(--muted-foreground));
}
/* 原型 .switch：胶囊滑块 */
.switch {
    position: relative;
    flex-shrink: 0;
    width: 42px;
    height: 24px;
    border-radius: 9999px;
    border: 1px solid hsl(var(--border));
    background: hsl(var(--muted));
    cursor: pointer;
    transition: background 0.15s;
}
.switch::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: hsl(var(--card));
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.4);
    transition: transform 0.15s;
}
.switch[aria-checked='true'] {
    background: hsl(var(--primary));
    border-color: hsl(var(--primary));
}
.switch[aria-checked='true']::after {
    transform: translateX(18px);
}
.switch:disabled {
    opacity: 0.55;
    cursor: not-allowed;
}
</style>