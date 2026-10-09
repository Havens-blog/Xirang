import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";

/**
 * 运维 Agent 独立前端路由（/opsagent/* ，nginx 同域托管 + vite base）。
 *
 * 鉴权口径：本应用无本地登录页——会话 cookie（ecmdb-token-key）+ Bearer 由
 * 请求层承载，401 统一跳运维平台登录（/console/login）并带 redirect 回跳；
 * 租户边界由后端 EiamAuth（会话校验 + RequireTenant）承接。
 * 路由全部 import() 懒加载（六页路由级代码分割）。
 */

export const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: () => import("@/layouts/AppShell.vue"),
    children: [
      { path: "", redirect: "/chat" },
      {
        path: "chat",
        name: "ops-chat",
        component: () => import("@/views/chat/index.vue"),
      },
      {
        path: "risk-center",
        name: "ops-risk-center",
        component: () => import("@/views/risk-center/index.vue"),
      },
      {
        path: "diagnosis/:id",
        name: "ops-diagnosis",
        component: () => import("@/views/diagnosis/index.vue"),
      },
      {
        path: "history",
        name: "ops-history",
        component: () => import("@/views/history/index.vue"),
      },
      {
        path: "settings",
        name: "ops-settings",
        component: () => import("@/views/settings/index.vue"),
      },
      {
        path: "agents",
        name: "ops-agents",
        component: () => import("@/views/agents/index.vue"),
      },
      { path: ":pathMatch(.*)*", redirect: "/chat" },
    ],
  },
];

export const router = createRouter({
  // BASE_URL 来自 vite base '/opsagent/'，与 nginx 同域反代形态一致
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});