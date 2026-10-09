<script setup lang="ts">
/**
 * 运维 Agent 应用壳 —— 原型同构版（ui/prototype/ app.js sidebarHTML + styles.css）。
 *
 * - 左侧 240px 分组侧边栏：品牌（Haven 运维 Agent）+ 核心/数据视图/管理 三组导航
 *   + 底栏（新建诊断按钮 + 用户块 + 主题切换）；lucide 线框图标与原型逐路径一致。
 * - 顶部 topbar（56px sticky）：页面标题 + 编排层 chip + 右侧动作（对话排障 = 历史诊断入口）。
 * - 主题：html.dark class 切换（localStorage haven-theme，与原型同键），tokens 恒深色。
 *
 * P1 范围 = 原型侧边栏裁剪版：P2 页面（RCA 分析/拓扑视图）不进导航；诊断详情为
 * 详情子页不进导航（自风险中心进入，与原型同）。
 */
import { computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "@/stores/user";
import { useTopbar, type TopbarAction } from "@/composables/useTopbar";

// lucide 线框 SVG path（与原型 app.js ICONS 逐字一致）
const ICON_PATHS: Record<string, string> = {
  activity:
    '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M22 12h-4l-3 9L9 3l-3 9H2"/>',
  message:
    '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>',
  alert:
    '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0zM12 9v4m0 4h.01"/>',
  clock:
    '<circle cx="12" cy="12" r="10"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6l4 2"/>',
  box: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>',
  settings:
    '<circle cx="12" cy="12" r="3"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>',
  plus: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m-7-7h14"/>',
  moon: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>',
}

/** 分组导航（原型 sidebarHTML 的 P1 裁剪版：核心/数据视图/管理） */
const NAV_GROUPS: { label: string; items: { key: string; label: string; to: string; icon: string }[] }[] = [
  {
    label: "核心",
    items: [
      { key: "chat", label: "对话排障", to: "/chat", icon: "message" },
      { key: "risk", label: "风险中心", to: "/risk-center", icon: "alert" },
    ],
  },
  {
    label: "数据视图",
    items: [{ key: "history", label: "历史回溯", to: "/history", icon: "clock" }],
  },
  {
    label: "管理",
    items: [
      { key: "agents", label: "Agent 管理", to: "/agents", icon: "box" },
      { key: "settings", label: "系统配置", to: "/settings", icon: "settings" },
    ],
  },
]

/** 页面标题（topbar 左，原型逐页对齐）+ 对话排障的编排层 chip 与历史诊断入口 */
const route = useRoute()
const PAGE_TITLES: Record<string, string> = {
  "ops-chat": "对话排障",
  "ops-risk-center": "风险中心",
  "ops-diagnosis": "诊断详情",
  "ops-history": "历史诊断回溯",
  "ops-settings": "系统配置",
  "ops-agents": "Agent 管理",
}
const pageTitle = computed(() => PAGE_TITLES[String(route.name)] ?? "运维 Agent")
const isChatPage = computed(() => route.name === "ops-chat")
const isRiskPage = computed(() => route.name === "ops-risk-center")
const isDiagnosisPage = computed(() => route.name === "ops-diagnosis")
const isSettingsPage = computed(() => route.name === "ops-settings")

const router = useRouter()
const topbar = useTopbar()
const pageActions = computed<TopbarAction[]>(() => topbar.actions.value)
function runAction(action: TopbarAction): void {
  action.onClick()
}

const userStore = useUserStore()

/** 用户块：用户名（未拉取回退 havens，与原型一致）+ 职责行 */
const displayName = computed(() => userStore.username || "havens")
const displayRole = computed(() => (userStore.isAdmin ? "平台管理员" : "值班运维"))
const avatarLetter = computed(() => (displayName.value[0] || "H").toUpperCase())

/** 主题切换（html.dark + localStorage haven-theme，与原型同键） */
function applySavedTheme(): void {
  const saved = localStorage.getItem("haven-theme")
  // 原型口径：saved === 'light' 时移除 dark；默认深色
  if (saved === "light") document.documentElement.classList.remove("dark")
  else document.documentElement.classList.add("dark")
}
function toggleTheme(): void {
  document.documentElement.classList.toggle("dark")
  localStorage.setItem(
    "haven-theme",
    document.documentElement.classList.contains("dark") ? "dark" : "light",
  )
}
onMounted(() => {
  applySavedTheme()
  void userStore.fetchProfile()
})
</script>

<template>
  <div class="opsagent-page app-shell">
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="logo-icon" aria-hidden="true">
          <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" v-html="ICON_PATHS.activity" />
        </div>
        <h1>Haven 运维 Agent</h1>
      </div>

      <nav class="sidebar-nav" aria-label="主导航">
        <div v-for="group in NAV_GROUPS" :key="group.label" class="nav-group">
          <div class="nav-group-label">{{ group.label }}</div>
          <RouterLink
            v-for="item in group.items"
            :key="item.key"
            :to="item.to"
            class="nav-item"
          >
            <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" v-html="ICON_PATHS[item.icon]" />
            {{ item.label }}
          </RouterLink>
        </div>
      </nav>

      <div class="sidebar-footer">
        <RouterLink to="/chat" class="btn footer-new">
          <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" v-html="ICON_PATHS.plus" />
          新建诊断
        </RouterLink>
        <div class="sidebar-user">
          <div class="avatar" aria-hidden="true">{{ avatarLetter }}</div>
          <div class="sidebar-user-meta">
            <p class="u-name">{{ displayName }}</p>
            <p class="u-role">{{ displayRole }}</p>
          </div>
          <button
            class="btn btn-ghost btn-sm theme-btn"
            aria-label="切换主题"
            title="切换主题"
            @click="toggleTheme"
          >
            <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" v-html="ICON_PATHS.moon" />
          </button>
        </div>
      </div>
    </aside>

    <div class="main-area">
      <header class="topbar">
        <div class="topbar-title">
          {{ pageTitle }}
          <span v-if="isChatPage" class="agent-chip agent-chip-coord">编排层</span>
        </div>
        <div class="topbar-actions">
          <button
            v-for="action in pageActions"
            :key="action.key"
            :class="['btn', action.primary ? 'btn-action-primary' : 'btn-outline btn-sm']"
            @click="runAction(action)"
          >
            {{ action.label }}
          </button>
          <RouterLink v-if="isChatPage" to="/history" class="btn btn-ghost btn-sm">
            🕘 历史诊断
          </RouterLink>
          <RouterLink v-if="isRiskPage" to="/history" class="btn btn-outline btn-sm">
            🕘 历史诊断
          </RouterLink>
          <button v-if="isDiagnosisPage" class="btn btn-outline btn-sm" @click="router.back()">
            ← 返回
          </button>
        </div>
      </header>

      <div class="page" :class="{ 'page-chat': isChatPage, 'page-settings': isSettingsPage }">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
// ===== 布局骨架（原型 styles.css 同构） =====
.app-shell {
  display: flex;
  min-height: 100vh;
}

// ===== Sidebar（分组导航） =====
.sidebar {
  width: 240px;
  border-right: 1px solid hsl(var(--border));
  background: hsl(var(--card) / 0.5);
  backdrop-filter: blur(8px);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  height: 100vh;
  position: sticky;
  top: 0;
  overflow-y: auto;
}

.sidebar-brand {
  height: 64px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid hsl(var(--border));
  padding: 0 20px;
  flex-shrink: 0;

  .logo-icon {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    flex-shrink: 0;
    background: hsl(var(--primary));
    display: flex;
    align-items: center;
    justify-content: center;

    svg {
      width: 20px;
      height: 20px;
      color: hsl(var(--primary-foreground));
    }
  }

  h1 {
    font-size: 17px;
    font-weight: 700;
    letter-spacing: -0.02em;
    margin: 0;
    color: hsl(var(--foreground));
  }
}

.sidebar-nav {
  flex: 1;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
}

.nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-group-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: hsl(var(--muted-foreground) / 0.7);
  padding: 0 12px 6px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  color: hsl(var(--muted-foreground));
  text-decoration: none;
  transition:
    background 0.15s,
    color 0.15s;

  svg {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
  }

  &:hover {
    background: hsl(var(--accent));
    color: hsl(var(--foreground));
  }

  &.router-link-active {
    background: hsl(var(--primary) / 0.15);
    color: hsl(var(--primary));
  }

  &:focus-visible {
    outline: 2px solid hsl(var(--ring));
    outline-offset: 2px;
  }
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid hsl(var(--border));
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
}

.footer-new {
  width: 100%;
  text-decoration: none;
}

.sidebar-user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px;

  .avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    flex-shrink: 0;
    background: linear-gradient(135deg, hsl(var(--primary)), hsl(271 91% 65%));
    color: hsl(var(--primary-foreground));
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 600;
  }

  &-meta {
    flex: 1;
    min-width: 0;
  }

  .u-name {
    font-size: 13px;
    font-weight: 600;
    margin: 0;
    color: hsl(var(--foreground));
  }

  .u-role {
    font-size: 12px;
    color: hsl(var(--muted-foreground));
    margin: 0;
  }

  .theme-btn {
    flex-shrink: 0;

    svg {
      width: 16px;
      height: 16px;
    }
  }
}

// ===== 主区（原型 styles.css 同构） =====
.main-area {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.topbar {
  height: 56px;
  border-bottom: 1px solid hsl(var(--border));
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  background: hsl(var(--card) / 0.4);
  backdrop-filter: blur(8px);
  position: sticky;
  top: 0;
  z-index: 40;
  flex-shrink: 0;
}

.topbar-title {
  font-size: 15px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
  color: hsl(var(--foreground));
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

// ===== 内容页（原型 .page：28px 内边距 + 1360px 上限；chat 收紧 1080、settings 900） =====
.page {
  padding: 28px;
  max-width: 1360px;
  width: 100%;
  margin: 0 auto;
}

.page-chat {
  max-width: 1080px;
}

.page-settings {
  max-width: 900px;
}

// ===== 微型按钮与 chip（原型 styles.css 缩略语义） =====
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
  transition:
    background 0.15s,
    opacity 0.15s;

  svg {
    width: 16px;
    height: 16px;
  }

  &:hover {
    background: hsl(var(--primary) / 0.9);
  }
}

.btn-ghost {
  background: transparent;
  color: hsl(var(--muted-foreground));

  &:hover {
    background: hsl(var(--accent));
    color: hsl(var(--foreground));
  }
}

.btn-outline {
  background: transparent;
  border: 1px solid hsl(var(--border));
  color: hsl(var(--foreground));

  &:hover {
    border-color: hsl(var(--primary));
    color: hsl(var(--primary));
    background: hsl(var(--primary) / 0.08);
  }
}

.btn-action-primary {
  // 页面注册的主动作（如「保存配置」）：原型 btn primary 同款
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
}

.btn-sm {
  height: 30px;
  padding: 0 12px;
  font-size: 12px;
  border-radius: 6px;
}

.agent-chip {
  font-size: 11px;
  padding: 2px 10px;
  border-radius: 9999px;
  font-weight: 500;
}

.agent-chip-coord {
  background: hsl(var(--agent-coord) / 0.15);
  color: hsl(var(--agent-coord));
  border: 1px solid hsl(var(--agent-coord) / 0.3);
}
</style>