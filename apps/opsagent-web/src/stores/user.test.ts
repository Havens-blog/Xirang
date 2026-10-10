// @vitest-environment happy-dom
import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createApp } from "vue";
import { fetchUserProfile } from "@/api/iam";
import { useUserStore } from "./user";

/**
 * 用户档案 store（settings 页 isAdmin 门控的数据源）：
 * - eiam RetrieveUser 嵌套结构映射（user.username / is_admin）
 * - 拉取失败静默降级（保持未登录态，后端权威鉴权）
 */

vi.mock("@/api/iam", () => ({
  fetchUserProfile: vi.fn(),
}));

const mockProfile = vi.mocked(fetchUserProfile);

const freshPinia = () => {
  const pinia = createPinia();
  createApp({ render: () => null }).use(pinia);
  setActivePinia(pinia);
  return pinia;
};

describe("useUserStore", () => {
  beforeEach(() => {
    freshPinia();
    vi.clearAllMocks();
  });

  it("嵌套结构映射：user.username → username；is_admin=true → 平台管理员门控放行", async () => {
    mockProfile.mockResolvedValue({
      user: { username: "admin", nickname: "系统管理员" },
      is_admin: true,
      current_tenant_id: 1,
    });
    const store = useUserStore();
    await store.fetchProfile();
    expect(store.username).toBe("admin");
    expect(store.isAdmin).toBe(true);
  });

  it("is_admin=false / 字段缺失 → 非管理员（编辑门控关闭）", async () => {
    mockProfile.mockResolvedValue({ user: { username: "havens" }, is_admin: false });
    const store = useUserStore();
    await store.fetchProfile();
    expect(store.username).toBe("havens");
    expect(store.isAdmin).toBe(false);
  });

  it("拉取失败静默降级且可重试（loaded 不置位）", async () => {
    mockProfile.mockRejectedValueOnce(new Error("network"));
    const store = useUserStore();
    await store.fetchProfile();
    expect(store.username).toBe("");
    expect(store.isAdmin).toBe(false);
    expect(store.loaded).toBe(false);

    mockProfile.mockResolvedValueOnce({ user: { username: "admin" }, is_admin: true });
    await store.fetchProfile();
    expect(store.isAdmin).toBe(true);
  });

  it("成功后重复 fetchProfile 不再发请求（loaded 幂等）", async () => {
    mockProfile.mockResolvedValue({ user: { username: "a" }, is_admin: true });
    const store = useUserStore();
    await store.fetchProfile();
    await store.fetchProfile();
    expect(mockProfile).toHaveBeenCalledTimes(1);
  });
});