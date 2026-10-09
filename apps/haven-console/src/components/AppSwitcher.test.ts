// @vitest-environment happy-dom
/**
 * AppSwitcher 平台导航面板 v2 集成测试（task 6.1，参照 e-cam-web PlatformNav）。
 *
 * 覆盖 AC：
 * - 零租户受限态门控（!isAdmin && tenants.length===0 隐藏；admin 豁免）
 * - 触发开合面板；Esc 关闭；遮罩组结构
 * - 面板分组：收藏列（默认 4 条）+ 控制台列 + 云管分区块
 * - 收藏：☆→★ toggle、localStorage（haven-platform-nav-favorites）持久化、
 *   收藏列点击导航、✕ 移除、空态
 * - 权限码裁剪：控制台列入口按码过滤（策略/租户 requiresAdmin 同口径双闸）
 * - 导航：控制台入口 router.push 站内跳转；云管深链 window.open 新标签（noopener）
 *
 * 手法：真实 pinia user store 注入 permissions/tenants/isAdmin；
 * 内存 router 观测 push；Teleport 真实渲染至 document.body（happy-dom）供面板断言。
 */
import { createPinia, setActivePinia } from "pinia";
import piniaPluginPersistedstate from "pinia-plugin-persistedstate";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import { createApp } from "vue";
import { mount } from "@vue/test-utils";
import {
  createMemoryHistory,
  createRouter,
  type Router,
} from "vue-router";
import AppSwitcher from "./AppSwitcher.vue";
import { useUserStore } from "@/stores/user";
import {
  IAM_ALL_CODES,
  IAM_POLICY,
  IAM_TENANT,
} from "@/config/permission-codes";

const FAV_KEY = "haven-platform-nav-favorites";

type TenantLike = { id: number; name: string; code: string; domain: string };

function freshPinia() {
  const pinia = createPinia();
  pinia.use(piniaPluginPersistedstate);
  createApp({ render: () => null }).use(pinia);
  setActivePinia(pinia);
}

function setupUser(opts: {
  isAdmin?: boolean;
  codes?: readonly string[];
  tenants?: TenantLike[];
}) {
  freshPinia();
  const userStore = useUserStore();
  userStore.isAdmin = opts.isAdmin ?? false;
  (
    userStore as unknown as {
      permissions: string[];
    }
  ).permissions = [...(opts.codes ?? [])];
  (
    userStore as unknown as { tenants: TenantLike[] }
  ).tenants = [
    ...(opts.tenants ?? [{ id: 1, name: "t1", code: "t1", domain: "" }]),
  ];
  return userStore;
}

async function makeRouter(): Promise<Router> {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/workbench", component: { template: "<div />" } },
      { path: "/users", component: { template: "<div />" } },
      { path: "/policies", component: { template: "<div />" } },
      { path: "/", component: { template: "<div />" } },
    ],
  });
  await router.push("/workbench");
  await router.isReady();
  return router;
}

type Wrapper = ReturnType<typeof mount<typeof AppSwitcher>>;

/** 每个测试的挂载例（afterEach 统一卸载 + 清 body，防多实例 Teleport/监听叠加） */
const mounted: Wrapper[] = [];

function mountSwitcher(router: Router): Wrapper {
  const wrapper = mount(AppSwitcher, { global: { plugins: [router] } });
  mounted.push(wrapper);
  return wrapper;
}

/** Teleport 面板根（真实渲染至 body） */
function panelEl(): HTMLElement | null {
  return document.body.querySelector<HTMLElement>(".pnav-panel");
}

/** 面板内入口节点（.pnav-it，收藏列 .pnav-fi 除外） */
function navItems(): HTMLElement[] {
  return [...document.body.querySelectorAll<HTMLElement>(".pnav-it")];
}

/** 收藏列条目（.pnav-fi） */
function favItems(): HTMLElement[] {
  return [
    ...document.body.querySelectorAll<HTMLElement>('[data-testid="platform-nav-fav"]'),
  ];
}

describe("AppSwitcher 零租户受限态门控（UF-2 受限视图，v2 口径不变）", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    localStorage.clear();
  });
  afterEach(() => {
    for (const w of mounted) w.unmount();
    mounted.length = 0;
    document.body.innerHTML = "";
    vi.restoreAllMocks();
  });

  it("非 admin + 无租户 → 隐藏（不渲染）", async () => {
    setupUser({ isAdmin: false, codes: IAM_ALL_CODES, tenants: [] });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    expect(wrapper.find(".app-switcher").exists()).toBe(false);
  });

  it("平台管理员（零租户豁免）→ 可见", async () => {
    setupUser({ isAdmin: true, codes: IAM_ALL_CODES, tenants: [] });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    expect(wrapper.find(".app-switcher").exists()).toBe(true);
  });

  it("非 admin + 有租户 → 可见", async () => {
    setupUser({ isAdmin: false, codes: IAM_ALL_CODES, tenants: [
      { id: 1, name: "t", code: "t", domain: "" },
    ] });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    expect(wrapper.find(".app-switcher").exists()).toBe(true);
  });
});

describe("AppSwitcher 平台导航面板（分组 + 收藏 + 导航）", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    localStorage.clear();
  });
  afterEach(() => {
    for (const w of mounted) w.unmount();
    mounted.length = 0;
    document.body.innerHTML = "";
    vi.restoreAllMocks();
  });

  it("触发开合面板；结构 = 收藏列 + 控制台列 + 云管分区块", async () => {
    setupUser({ isAdmin: true, codes: IAM_ALL_CODES });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);

    const trigger = wrapper.find('[data-testid="platform-nav-trigger"]');
    expect(trigger.exists()).toBe(true);
    expect(trigger.text()).toContain("平台导航");
    expect(trigger.attributes("aria-haspopup")).toBe("dialog");

    await trigger.trigger("click");
    expect(panelEl()).toBeTruthy();
    const text = panelEl()?.textContent ?? "";
    expect(text).toContain("我的收藏");
    expect(text).toContain("控制台 · 统一身份与访问");
    expect(text).toContain("资源 · 主机");
    expect(text).toContain("资源 · 网络");
    expect(text).toContain("FINOPS");
    expect(text).toContain("平台管理");

    // 再点关闭
    await trigger.trigger("click");
    expect(panelEl()).toBeFalsy();
  });

  it("默认收藏 4 条（工作台/策略管理/虚拟机/成本概览）", async () => {
    setupUser({ isAdmin: true, codes: IAM_ALL_CODES });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    await wrapper.find('[data-testid="platform-nav-trigger"]').trigger("click");

    const favs = favItems();
    expect(favs).toHaveLength(4);
    const titles = favs.map((el) => el.textContent ?? "").join("|");
    expect(titles).toContain("工作台");
    expect(titles).toContain("策略管理");
    expect(titles).toContain("虚拟机");
    expect(titles).toContain("成本概览");
  });

  it("☆ toggle → ★ + localStorage 持久化（key haven-platform-nav-favorites）", async () => {
    setupUser({ isAdmin: true, codes: IAM_ALL_CODES });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    await wrapper.find('[data-testid="platform-nav-trigger"]').trigger("click");

    const star = document.body.querySelector<HTMLElement>(
      '[data-testid="platform-nav-star-console-roles"]',
    );
    expect(star).toBeTruthy();
    expect(star?.textContent).toBe("☆");

    star?.click();
    await wrapper.vm.$nextTick();

    // 收藏列新增「角色管理」+ ★ 变
    expect(favItems().some((el) => el.textContent?.includes("角色管理"))).toBe(true);
    expect(star?.textContent).toBe("★");
    const stored = JSON.parse(localStorage.getItem(FAV_KEY) ?? "[]") as Array<{
      key: string;
    }>;
    expect(stored.some((f) => f.key === "console-roles")).toBe(true);

    // 再点 → 取消收藏
    star?.click();
    await wrapper.vm.$nextTick();
    expect(favItems().some((el) => el.textContent?.includes("角色管理"))).toBe(false);
    const storedAfter = JSON.parse(
      localStorage.getItem(FAV_KEY) ?? "[]",
    ) as Array<{ key: string }>;
    expect(storedAfter.some((f) => f.key === "console-roles")).toBe(false);
  });

  it("收藏列 ✕ 逐条移除 → 空态 + 存储清空", async () => {
    setupUser({ isAdmin: true, codes: IAM_ALL_CODES });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    await wrapper.find('[data-testid="platform-nav-trigger"]').trigger("click");

    const removes = document.body.querySelectorAll<HTMLElement>(
      '[data-testid="platform-nav-fav-remove"]',
    );
    expect(removes.length).toBe(4);
    for (const b of [...removes]) {
      b.click();
      await wrapper.vm.$nextTick();
    }
    expect(favItems()).toHaveLength(0);
    expect(
      document.body.querySelector(".pnav-fi-empty")?.textContent,
    ).toContain("收藏常用入口");
    expect(JSON.parse(localStorage.getItem(FAV_KEY) ?? "[]")).toEqual([]);
  });

  it("控制台入口 → router.push 站内；云管深链 → window.open 新标签 noopener", async () => {
    setupUser({ isAdmin: true, codes: IAM_ALL_CODES });
    const router = await makeRouter();
    const push = vi.spyOn(router, "push").mockResolvedValue();
    const openStub = vi.fn();
    vi.stubGlobal("open", openStub);
    const wrapper = mountSwitcher(router);
    await wrapper.find('[data-testid="platform-nav-trigger"]').trigger("click");

    const usersItem = navItems().find((el) => el.textContent?.includes("用户管理"));
    expect(usersItem).toBeTruthy();
    usersItem?.click();
    await wrapper.vm.$nextTick();
    expect(push).toHaveBeenCalledWith("/users");
    expect(panelEl()).toBeFalsy(); // 导航后面板关闭

    // 云管入口
    await wrapper.find('[data-testid="platform-nav-trigger"]').trigger("click");
    const camItem = navItems().find((el) => el.textContent?.includes("虚拟机"));
    expect(camItem).toBeTruthy();
    camItem?.click();
    expect(openStub).toHaveBeenCalledWith("/cam/compute/ecs", "_blank", "noopener");
    vi.unstubAllGlobals();
  });

  it("权限码裁剪：无策略/租户码（非 admin）→ 控制台列不出策略/租户入口", async () => {
    setupUser({
      isAdmin: false,
      codes: [...IAM_ALL_CODES].filter(
        (c) => c !== IAM_POLICY.VIEW && c !== IAM_TENANT.VIEW,
      ),
    });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    await wrapper.find('[data-testid="platform-nav-trigger"]').trigger("click");

    const consoleTitles = navItems()
      .map((el) => el.textContent ?? "")
      .join("|");
    expect(consoleTitles).toContain("用户管理");
    expect(consoleTitles).toContain("工作台");
    expect(consoleTitles).not.toContain("策略管理");
    expect(consoleTitles).not.toContain("租户管理");
    // 云管侧不受控制台码裁剪（跨应用，服务端校验兜底）
    expect(consoleTitles).toContain("虚拟机");
  });

  it("Esc 关闭面板", async () => {
    setupUser({ isAdmin: true, codes: IAM_ALL_CODES });
    const router = await makeRouter();
    const wrapper = mountSwitcher(router);
    await wrapper.find('[data-testid="platform-nav-trigger"]').trigger("click");
    expect(panelEl()).toBeTruthy();
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    await wrapper.vm.$nextTick();
    expect(panelEl()).toBeFalsy();
  });
});