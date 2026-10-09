// @vitest-environment happy-dom
/**
 * PolicyList 策略编辑对话框双模式（图形化编排 / JSON）集成测试（task 6.2）。
 *
 * 覆盖 AC：
 * - 新增对话框默认图形化编排 tab；默认语句 Allow/cam:cert:Get/cert/*
 * - 图形化编辑（effect 切换、添加 Statement）→ JSON 文本双向同步；提交载荷由
 *   parseStatementsJson(editStatements) 产出（单一事实源口径不变）
 * - JSON → 图形化：合法 JSON 切回图形化渲染对应条数；非法 JSON 图形化禁编（warn 横幅）
 * - condition 保留：图形化不编辑但保存不丢失（切 JSON 仍在）
 * - submit：JSON 语法错 → 阻止保存（字段错误就近提示）
 *
 * 手法：mock usePolicyListStore / api/permissions.fetchPermissionManifest；
 * stub el-select（渲染 modelValue + 捕获 update:modelValue 事件 → 图形化模型变更）；
 * 真实 pinia user store 注入权限码 + 内存 router。
 */
import { createPinia, setActivePinia } from "pinia";
import piniaPluginPersistedstate from "pinia-plugin-persistedstate";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createApp, defineComponent, h } from "vue";
import { mount } from "@vue/test-utils";
import { createMemoryHistory, createRouter } from "vue-router";
import PolicyList from "./PolicyList.vue";
import { useUserStore } from "@/stores/user";
import { IAM_ALL_CODES } from "@/config/permission-codes";

// ---- fake policyList store（只读消费 + create/update/remove 桩）----
const fakeStore = {
  items: [],
  total: 0,
  page: 1,
  pageSize: 20,
  loading: false,
  error: null as string | null,
  fetch: vi.fn(),
  setKeyword: vi.fn(),
  setPage: vi.fn(),
  setPageSize: vi.fn(),
  create: vi.fn().mockResolvedValue(true),
  update: vi.fn().mockResolvedValue(true),
  remove: vi.fn().mockResolvedValue(true),
};

vi.mock("@/stores/policyList", () => ({
  usePolicyListStore: () => fakeStore,
}));

vi.mock("@/api/permissions", () => ({
  fetchPermissionManifest: vi.fn().mockResolvedValue([
    { code: "iam:user:view", name: "用户查看", group: "用户管理", service: "iam", hasMenu: true, menuUrns: [] },
    { code: "cam:cert:Get", name: "证书获取", group: "证书管理", service: "cam", hasMenu: false, menuUrns: [] },
    { code: "iam:policy:add", name: "新增策略", group: "策略管理", service: "iam", hasMenu: false, menuUrns: [] },
  ]),
}));

// ---- el-select stub：渲染 modelValue（data-model），供 vm.$emit 驱动图形变更 ----
const ElSelectStub = defineComponent({
  name: "ElSelect",
  props: {
    modelValue: { type: [String, Array, Number], default: null },
  },
  emits: ["update:modelValue"],
  setup(props, { attrs, slots }) {
    return () =>
      h(
        "div",
        {
          class: "el-select-stub",
          "data-testid": (attrs["data-testid"] as string | undefined) ?? "",
          "data-model": JSON.stringify(props.modelValue ?? null),
        },
        slots.default?.(),
      );
  },
});

// ---- 装配 ----
function freshPinia() {
  const pinia = createPinia();
  pinia.use(piniaPluginPersistedstate);
  createApp({ render: () => null }).use(pinia);
  setActivePinia(pinia);
}

async function mountPolicyList() {
  freshPinia();
  const userStore = useUserStore();
  userStore.isAdmin = true;
  (
    userStore as unknown as { permissions: string[] }
  ).permissions = [...IAM_ALL_CODES];
  (
    userStore as unknown as {
      tenants: { id: number; name: string; code: string; domain: string }[];
    }
  ).tenants = [{ id: 1, name: "t1", code: "t1", domain: "" }];

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: "/policies", component: PolicyList }],
  });
  await router.push("/policies");
  await router.isReady();
  const wrapper = mount(PolicyList, {
    global: {
      plugins: [router],
      stubs: { ElSelect: ElSelectStub, ElOption: true, ElOptionGroup: true },
    },
  });
  await wrapper.vm.$nextTick();
  return wrapper;
}

/** 已挂载 wrapper 类型（mountPolicyList 异步装配后） */
type Wrapped = Awaited<ReturnType<typeof mountPolicyList>>;

async function openCreate(wrapper: Wrapped) {
  const btn = wrapper
    .findAll("button")
    .find((b) => b.text().includes("新增策略"));
  expect(btn).toBeTruthy();
  await btn!.trigger("click");
  await wrapper.vm.$nextTick();
  return wrapper;
}

function editorTextarea(wrapper: Wrapped) {
  return wrapper.find('[data-testid="edit-policy-statements"]');
}

function jsonTab(wrapper: Wrapped) {
  return wrapper.find('[data-testid="pse-tab-json"]');
}

function graphicalTab(wrapper: Wrapped) {
  return wrapper.find('[data-testid="pse-tab-graphical"]');
}

function effectStubs(wrapper: Wrapped) {
  return wrapper
    .findAllComponents(ElSelectStub)
    .filter((w) => w.attributes("data-testid") === "pse-effect");
}

function parseTextareaJson(wrapper: Wrapped): unknown[] {
  const text = (editorTextarea(wrapper).element as HTMLTextAreaElement).value;
  return JSON.parse(text) as unknown[];
}

describe("PolicyList 图形化策略编辑器（task 6.2 双模式）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("新增对话框默认图形化编排 tab，含默认语句（Allow/cam:cert:Get/cert/*）", async () => {
    const wrapper = await openCreate(await mountPolicyList());

    expect(wrapper.find('[data-testid="policy-statements-editor"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="pse-tab-graphical"]').attributes("aria-selected")).toBe("true");
    expect(wrapper.findAll('[data-testid="pse-statement"]')).toHaveLength(1);
    // effect 默认 Allow；actions/resources 由 json 文本同步而来
    const effect = effectStubs(wrapper)[0];
    expect(effect?.attributes("data-model")).toBe('"Allow"');
    const actions = wrapper.find('[data-testid="pse-actions-el"]');
    expect(actions.attributes("data-model")).toContain("cam:cert:Get");
    const resources = wrapper.find('[data-testid="pse-resources-el"]');
    expect(resources.attributes("data-model")).toContain("cert/*");
  });

  it("图形化编辑 → JSON 双向同步 + 提交载荷经 editStatements 产出", async () => {
    const wrapper = await openCreate(await mountPolicyList());

    // fill code/name
    await wrapper.find('[data-testid="edit-policy-code"]').setValue("ci-read");
    await wrapper.find('[data-testid="edit-policy-name"]').setValue("CI 只读策略");

    // effect Allow → Deny（vm.$emit 驱动 onEffectChange）
    effectStubs(wrapper)[0]?.vm.$emit("update:modelValue", "Deny");
    await wrapper.vm.$nextTick();

    // 添加一条 Statement
    await wrapper.find('[data-testid="pse-add-statement"]').trigger("click");
    await wrapper.vm.$nextTick();
    expect(wrapper.findAll('[data-testid="pse-statement"]')).toHaveLength(2);

    // 第二条补齐 actions/resources（parseStatementsJson 要求非空数组）
    const actionsStubs = wrapper
      .findAllComponents(ElSelectStub)
      .filter((w) => w.attributes("data-testid") === "pse-actions-el");
    actionsStubs[1]?.vm.$emit("update:modelValue", ["iam:user:view"]);
    const resourcesStubs = wrapper
      .findAllComponents(ElSelectStub)
      .filter((w) => w.attributes("data-testid") === "pse-resources-el");
    resourcesStubs[1]?.vm.$emit("update:modelValue", ["user/*"]);
    await wrapper.vm.$nextTick();

    // 切 JSON 模式：两条语句，第一条 effect Deny 已同步
    await jsonTab(wrapper).trigger("click");
    const json = parseTextareaJson(wrapper);
    expect(json).toHaveLength(2);
    expect((json[0] as { effect: string }).effect).toBe("Deny");
    expect((json[0] as { actions: string[] }).actions).toContain("cam:cert:Get");
    expect(json[1]).toEqual({
      effect: "Allow",
      actions: ["iam:user:view"],
      resources: ["user/*"],
    });

    // 提交：create 载荷 = parseStatementsJson(editStatements)（单一事实源口径）
    await wrapper.find("form").trigger("submit");
    expect(fakeStore.create).toHaveBeenCalledTimes(1);
    const payload = vi.mocked(fakeStore.create).mock.calls[0]?.[0] as {
      code: string;
      name: string;
      statements: Array<{ effect: string; actions: string[]; resources: string[] }>;
    };
    expect(payload.code).toBe("ci-read");
    expect(payload.name).toBe("CI 只读策略");
    expect(payload.statements).toHaveLength(2);
    expect(payload.statements[0]?.effect).toBe("Deny");
    expect(payload.statements[1]?.actions).toEqual(["iam:user:view"]);
    expect(payload.statements[1]?.resources).toEqual(["user/*"]);
  });

  it("JSON → 图形化：合法 JSON 切回渲染对应语句条数", async () => {
    const wrapper = await openCreate(await mountPolicyList());
    await jsonTab(wrapper).trigger("click");

    const twoStatements = JSON.stringify(
      [
        { effect: "Allow", actions: ["iam:user:view"], resources: ["user/*"] },
        { effect: "Deny", actions: ["iam:user:delete"], resources: ["user/admin"] },
      ],
      null,
      2,
    );
    await editorTextarea(wrapper).setValue(twoStatements);
    await graphicalTab(wrapper).trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.findAll('[data-testid="pse-statement"]')).toHaveLength(2);
    const effects = effectStubs(wrapper);
    expect(effects[0]?.attributes("data-model")).toBe('"Allow"');
    expect(effects[1]?.attributes("data-model")).toBe('"Deny"');
  });

  it("非法 JSON → 图形化禁编（warn 横幅 + 无语句卡）", async () => {
    const wrapper = await openCreate(await mountPolicyList());
    await jsonTab(wrapper).trigger("click");
    await editorTextarea(wrapper).setValue("{bad json!!");
    await graphicalTab(wrapper).trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.find(".pse-warn").exists()).toBe(true);
    expect(wrapper.find(".pse-warn").text()).toContain("图形化编排暂不可用");
    expect(wrapper.findAll('[data-testid="pse-statement"]')).toHaveLength(0);
  });

  it("非法 JSON 提交被前端校验拦截（create 不调用，字段错误标红）", async () => {
    const wrapper = await openCreate(await mountPolicyList());
    await jsonTab(wrapper).trigger("click");
    await editorTextarea(wrapper).setValue("[]");
    await wrapper.find('[data-testid="edit-policy-code"]').setValue("ci-read");
    await wrapper.find('[data-testid="edit-policy-name"]').setValue("CI 策略");
    await wrapper.find("form").trigger("submit");
    await wrapper.vm.$nextTick();

    expect(fakeStore.create).not.toHaveBeenCalled();
    const err = wrapper.find("#pse-json-error");
    expect(err.exists()).toBe(true);
    expect(err.text()).toContain("Statement 不能为空数组");
  });

  it("condition 保留：图形化编辑不丢失 condition（切 JSON 仍在）", async () => {
    const wrapper = await openCreate(await mountPolicyList());
    await jsonTab(wrapper).trigger("click");
    const withCond = JSON.stringify(
      [
        {
          effect: "Allow",
          actions: ["iam:user:view"],
          resources: ["user/*"],
          condition: [{ key: "env", operator: "eq", values: ["prod"] }],
        },
      ],
      null,
      2,
    );
    await editorTextarea(wrapper).setValue(withCond);
    await graphicalTab(wrapper).trigger("click");
    await wrapper.vm.$nextTick();

    // condition 提示 visible；effect 编辑后 condition 不丢
    const note = wrapper.find('[data-testid="pse-condition-note"]');
    expect(note.exists()).toBe(true);
    expect(note.text()).toContain("1 项");
    effectStubs(wrapper)[0]?.vm.$emit("update:modelValue", "Deny");
    await wrapper.vm.$nextTick();

    await jsonTab(wrapper).trigger("click");
    const json = parseTextareaJson(wrapper);
    expect((json[0] as { effect: string }).effect).toBe("Deny");
    const cond = (json[0] as { condition?: unknown[] }).condition;
    expect(cond).toEqual([{ key: "env", operator: "eq", values: ["prod"] }]);
  });
});