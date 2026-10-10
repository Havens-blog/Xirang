import { beforeEach, describe, expect, it, vi } from "vitest";

/**
 * iam API 路径回归守卫（2026-10-09 平台事故口径）：
 * profile 必须走 /api/iam/user/profile（nginx / vite proxy 重写为 eiam /api/*），
 * 不得直连 /user/profile 或 /api/user/profile —— 那条路没有反代会拿到 SPA HTML。
 */

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  unwrap: vi.fn(),
}));

vi.mock("./request/eiam", () => ({
  eiamAxios: { get: mocks.get, post: vi.fn() },
  unwrapEiam: mocks.unwrap,
}));

import { fetchUserProfile } from "./iam";

describe("fetchUserProfile 路径守卫", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // 模拟真实 unwrapEiam：Promise<AxiosResponse<Envelope<T>>> → envelope.data
    mocks.unwrap.mockImplementation(
      async (p: Promise<{ data: { data?: unknown } }>) => {
        const resp = await p;
        return resp.data.data;
      },
    );
  });

  it("GET /api/iam/user/profile（统一外部前缀，非 /user/profile）", async () => {
    mocks.get.mockResolvedValue({ data: { code: 0, data: null } });
    await fetchUserProfile();
    expect(mocks.get).toHaveBeenCalledTimes(1);
    expect(mocks.get).toHaveBeenCalledWith("/api/iam/user/profile");
    expect(mocks.get).not.toHaveBeenCalledWith("/user/profile");
  });

  it("经 unwrapEiam 解包返回 data", async () => {
    const payload = { user: { username: "admin" }, is_admin: true };
    mocks.get.mockResolvedValue({ data: { code: 0, msg: "ok", data: payload } });
    const out = await fetchUserProfile();
    expect(mocks.unwrap).toHaveBeenCalled();
    expect(out).toEqual(payload);
  });
});