// @vitest-environment happy-dom
// useCan 组合式函数（P0 按键级裁剪判定源）契约测试：
// - can / canEvery / canSome 三判定与 user store.permissions 集合语义一致
// - 响应式：permissions 变化（登录/切租户/登出）后判定即时翻转
// - 空集合判定安全（不抛错）
import { createPinia, setActivePinia } from "pinia";
import piniaPluginPersistedstate from "pinia-plugin-persistedstate";
import { beforeEach, describe, expect, it } from "vitest";
import { createApp } from "vue";
import { useUserStore } from "@/stores/user";
import { useCan } from "./useCan";
import { IAM_POLICY, IAM_ROLE, IAM_USER } from "@/config/permission-codes";

let userStore: ReturnType<typeof useUserStore>;

beforeEach(() => {
  const pinia = createPinia();
  pinia.use(piniaPluginPersistedstate);
  createApp({ render: () => null }).use(pinia);
  setActivePinia(pinia);
  userStore = useUserStore();
  (
    userStore as unknown as {
      permissions: string[];
    }
  ).permissions = [IAM_USER.VIEW, IAM_USER.ADD, IAM_ROLE.EDIT];
});

describe("useCan 判定", () => {
  it("can：集合成员语义", () => {
    const { can } = useCan();
    expect(can(IAM_USER.VIEW)).toBe(true);
    expect(can(IAM_USER.ADD)).toBe(true);
    expect(can(IAM_ROLE.EDIT)).toBe(true);
    expect(can(IAM_USER.DELETE)).toBe(false);
    expect(can("")).toBe(false);
  });

  it("canEvery：全部持有才通过", () => {
    const { canEvery } = useCan();
    expect(canEvery([IAM_USER.VIEW, IAM_USER.ADD])).toBe(true);
    expect(canEvery([IAM_USER.VIEW, IAM_USER.DELETE])).toBe(false);
    expect(canEvery([])).toBe(true); // 空集恒真（与 router 守卫 every 语义一致）
  });

  it("canSome：任一持有即通过（授权管理组合码形态）", () => {
    const { canSome } = useCan();
    expect(
      canSome([IAM_ROLE.BATCH_ASSIGN, IAM_POLICY.BATCH_ATTACH]),
    ).toBe(false);
    expect(canSome([IAM_ROLE.EDIT, IAM_POLICY.BATCH_ATTACH])).toBe(true);
    expect(canSome([])).toBe(false);
  });

  it("响应式：permissions 变化后判定即时翻转", () => {
    const { can } = useCan();
    expect(can(IAM_USER.DELETE)).toBe(false);
    (
      userStore as unknown as {
        permissions: string[];
      }
    ).permissions = [IAM_USER.VIEW, IAM_USER.DELETE];
    expect(can(IAM_USER.DELETE)).toBe(true);
    expect(can(IAM_USER.ADD)).toBe(false);
  });
});