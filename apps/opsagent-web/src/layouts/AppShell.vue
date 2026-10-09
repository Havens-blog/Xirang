<script setup lang="ts">
/**
 * 运维 Agent 应用壳：cyan 深色导航头 + 内容 outlet。
 *
 * - 根部 .opsagent-page 命中 opsagent-theme.css 深色 tokens（与各页面同源）；
 * - 导航五项：对话排障 / 风险中心 / 历史回溯 / 系统配置 / Agent 管理
 *   （诊断详情为详情子页，经风险中心进入、不进导航）；
 * - 挂载时拉取 eiam 用户档案（settings 页 isAdmin 等字段来源），失败不影响使用。
 */
import { onMounted } from "vue";
import { useUserStore } from "@/stores/user";

const NAV_ITEMS = [
  { label: "对话排障", to: "/chat" },
  { label: "风险中心", to: "/risk-center" },
  { label: "历史回溯", to: "/history" },
  { label: "系统配置", to: "/settings" },
  { label: "Agent 管理", to: "/agents" },
];

const userStore = useUserStore();
onMounted(() => {
  void userStore.fetchProfile();
});
</script>

<template>
  <div class="opsagent-page app-shell">
    <header class="app-shell__header">
      <div class="app-shell__brand">
        <span class="app-shell__logo" aria-hidden="true">⌾</span>
        <span class="app-shell__title">运维 Agent</span>
      </div>
      <nav class="app-shell__nav" aria-label="运维 Agent 导航">
        <RouterLink
          v-for="item in NAV_ITEMS"
          :key="item.to"
          :to="item.to"
          class="app-shell__link"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
      <div class="app-shell__user">
        <span v-if="userStore.username" class="app-shell__username">
          {{ userStore.username }}
          <span v-if="userStore.isAdmin" class="app-shell__admin-badge">管理员</span>
        </span>
      </div>
    </header>

    <main class="app-shell__main">
      <RouterView />
    </main>
  </div>
</template>

<style lang="scss" scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;

  &__header {
    display: flex;
    align-items: center;
    gap: 28px;
    flex-shrink: 0;
    height: 52px;
    padding: 0 20px;
    background: hsl(var(--card));
    border-bottom: 1px solid hsl(var(--border));
  }

  &__brand {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  &__logo {
    color: hsl(var(--primary));
    font-size: 20px;
    line-height: 1;
  }

  &__title {
    font-size: 15px;
    color: hsl(var(--foreground));
  }

  &__nav {
    display: flex;
    align-items: center;
    gap: 4px;
    flex: 1 1 auto;
  }

  &__link {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 13px;
    color: hsl(var(--muted-foreground));
    text-decoration: none;
    transition:
      color 0.15s ease,
      background 0.15s ease;

    &:hover {
      color: hsl(var(--foreground));
      background: hsl(var(--secondary));
    }

    &.router-link-active {
      color: hsl(var(--primary-foreground));
      background: hsl(var(--primary));
    }
  }

  &__user {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  &__username {
    font-size: 13px;
    color: hsl(var(--muted-foreground));
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  &__admin-badge {
    padding: 1px 8px;
    border-radius: 999px;
    font-size: 11px;
    // severity/主色语义直接取主题 tokens
    background: hsl(var(--secondary));
    color: hsl(var(--primary));
    border: 1px solid hsl(var(--border));
  }

  &__main {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
  }
}
</style>