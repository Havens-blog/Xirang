// @vitest-environment happy-dom
import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createApp } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";
import AppShell from "./AppShell.vue";

/**
 * 应用壳结构测试 —— 锁定原型 ui/prototype（app.js sidebarHTML + styles.css）口径：
 * - 品牌区「Haven 运维 Agent」+ 分组侧边栏（核心/数据视图/管理）
 * - 底栏：新建诊断按钮 + 用户块 + 主题切换按钮
 * - topbar：路由标题映射（含历史诊断回溯 / 编排层 chip）
 * P2 页面（RCA/拓扑）不进导航；诊断详情为子页不进导航。
 */

vi.mock("@/api/iam", () => ({
  fetchUserProfile: vi.fn().mockResolvedValue({
    // eiam RetrieveUser 嵌套结构（POST 契约实核口径）
    user: { username: "havens", nickname: "值班 havens" },
    is_admin: false,
  }),
}));

const freshPinia = () => {
  const pinia = createPinia();
  createApp({ render: () => null }).use(pinia);
  setActivePinia(pinia);
  return pinia;
};

async function mountShell(path: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/chat", name: "ops-chat", component: { render: () => null } },
      { path: "/risk-center", name: "ops-risk-center", component: { render: () => null } },
      { path: "/history", name: "ops-history", component: { render: () => null } },
      { path: "/settings", name: "ops-settings", component: { render: () => null } },
      { path: "/agents", name: "ops-agents", component: { render: () => null } },
    ],
  });
  await router.push(path);
  await router.isReady();
  const wrapper = mount(AppShell, {
    global: { plugins: [router, freshPinia()] },
  });
  await flushPromises();
  return { wrapper, router };
}

describe("AppShell 原型结构", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("品牌区 + 分组侧边栏：核心/数据视图/管理 5 项（P1 裁剪）", async () => {
    const { wrapper } = await mountShell("/chat");
    expect(wrapper.find(".sidebar-brand h1").text()).toBe("Haven 运维 Agent");
    const groups = wrapper.findAll(".nav-group");
    const labels = groups.map((g) => g.find(".nav-group-label").text());
    expect(labels).toEqual(["核心", "数据视图", "管理"]);
    const items = wrapper.findAll(".nav-group .nav-item");
    expect(items.map((i) => i.text())).toEqual([
      "对话排障",
      "风险中心",
      "历史回溯",
      "Agent 管理",
      "系统配置",
    ]);
    // 完整 href 相对路径落点在历史导航（含 base /opsagent/）
    expect(items[0]!.attributes("href")).toContain("/chat");
  });

  it("底栏：新建诊断按钮 + 用户块（havens/值班运维）+ 主题切换", async () => {
    const { wrapper } = await mountShell("/chat");
    const footer = wrapper.find(".sidebar-footer");
    expect(footer.find(".footer-new").text()).toContain("新建诊断");
    expect(footer.find(".u-name").text()).toBe("havens");
    expect(footer.find(".u-role").text()).toBe("值班运维");
    expect(footer.find(".theme-btn").attributes("aria-label")).toBe("切换主题");
  });

  it("topbar 标题随路由：/history → 历史诊断回溯；/chat → 对话排障 + 编排层 chip", async () => {
    const chat = await mountShell("/chat");
    expect(chat.wrapper.find(".topbar-title").text()).toContain("对话排障");
    expect(chat.wrapper.find(".topbar-title").text()).toContain("编排层");
    chat.wrapper.unmount();

    const history = await mountShell("/history");
    expect(history.wrapper.find(".topbar-title").text()).toBe("历史诊断回溯");
  });

  it("主题切换：html.dark 与 localStorage haven-theme 联动（原型同键）", async () => {
    const { wrapper } = await mountShell("/chat");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    await wrapper.find(".theme-btn").trigger("click");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(localStorage.getItem("haven-theme")).toBe("light");
  });

  it("对话排障页 topbar 含「🕘 历史诊断」入口；风险中心同", async () => {
    const chat = await mountShell("/chat");
    const chatActions = chat.wrapper.findAll(".topbar-actions a");
    expect(chatActions.some((a) => a.text().includes("历史诊断"))).toBe(true);
    chat.wrapper.unmount();

    const risk = await mountShell("/risk-center");
    const riskActions = risk.wrapper.findAll(".topbar-actions a");
    expect(riskActions.some((a) => a.text().includes("历史诊断"))).toBe(true);
  });
});